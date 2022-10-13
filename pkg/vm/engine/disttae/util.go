// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package disttae

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/hashtable"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	plan2 "github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/plan/function"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/index"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func checkExprIsMonotonical(expr *plan.Expr) bool {
	switch exprImpl := expr.Expr.(type) {
	case *plan.Expr_F:
		for _, arg := range exprImpl.F.Args {
			isMonotonical := checkExprIsMonotonical(arg)
			if !isMonotonical {
				return false
			}
		}

		isMonotonical, _ := function.GetFunctionIsMonotonicalById(exprImpl.F.Func.GetObj())
		if !isMonotonical {
			return false
		}

		return true
	default:
		return true
	}
}

func _getColumnMapByExpr(expr *plan.Expr, columnMap map[int]struct{}) {
	switch exprImpl := expr.Expr.(type) {
	case *plan.Expr_F:
		for _, arg := range exprImpl.F.Args {
			_getColumnMapByExpr(arg, columnMap)
		}
	case *plan.Expr_Col:
		idx := exprImpl.Col.ColPos
		columnMap[int(idx)] = struct{}{}
	}
}

func getColumnsByExpr(expr *plan.Expr) []int {
	columnMap := make(map[int]struct{})
	_getColumnMapByExpr(expr, columnMap)

	columns := make([]int, len(columnMap))
	i := 0
	for k := range columnMap {
		columns[i] = k
		i++
	}
	sort.Ints(columns)
	return columns
}

func getIndexDataFromVec(idx uint16, vec *vector.Vector) (objectio.IndexData, objectio.IndexData, error) {
	var bloomFilter, zoneMap objectio.IndexData

	// get min/max from  vector
	if vec.Length() > 0 {
		cvec := containers.NewVectorWithSharedMemory(vec, true)

		// create zone map
		zm := index.NewZoneMap(vec.Typ)
		ctx := new(index.KeysCtx)
		ctx.Keys = cvec
		ctx.Count = vec.Length()
		defer ctx.Keys.Close()
		err := zm.BatchUpdate(ctx)
		if err != nil {
			return nil, nil, err
		}
		buf, err := zm.Marshal()
		if err != nil {
			return nil, nil, err
		}
		zoneMap, err = objectio.NewZoneMap(idx, buf)
		if err != nil {
			return nil, nil, err
		}

		// create bloomfilter
		sf, err := index.NewBinaryFuseFilter(cvec)
		if err != nil {
			return nil, nil, err
		}
		bf, err := sf.Marshal()
		if err != nil {
			return nil, nil, err
		}
		alg := uint8(0)
		bloomFilter = objectio.NewBloomFilter(idx, alg, bf)
	}

	return bloomFilter, zoneMap, nil
}

func getZonemapDataFromMeta(columns []int, meta BlockMeta, tableDef *plan.TableDef) ([][2]any, []uint8, error) {
	var columnMeta *ColumnMeta

	getIdx := func(idx int) int {
		return int(tableDef.Name2ColIndex[tableDef.Cols[columns[idx]].Name])
	}

	dataLength := len(columns)

	datas := make([][2]any, dataLength)
	dataTypes := make([]uint8, dataLength)

	for i := 0; i < dataLength; i++ {
		columnMeta = meta.columns[getIdx(i)]
		dataTypes[i] = columnMeta.typ
		typ := types.T(columnMeta.typ).ToType()

		czm, _ := columnMeta.zoneMap.(*objectio.ZoneMap)
		buf := czm.GetData()
		zm := index.NewZoneMap(typ)
		err := zm.Unmarshal(buf)
		if err != nil {
			return nil, nil, err
		}

		datas[i] = [2]any{
			zm.GetMin(),
			zm.GetMax(),
		}
	}

	return datas, dataTypes, nil
}

func evalFilterExpr(expr *plan.Expr, bat *batch.Batch, proc *process.Process) (bool, error) {
	if len(bat.Vecs) == 0 { //that's constant expr
		e, err := plan2.ConstantFold(bat, expr)
		if err != nil {
			return false, err
		}

		if cExpr, ok := e.Expr.(*plan.Expr_C); ok {
			if bVal, bOk := cExpr.C.Value.(*plan.Const_Bval); bOk {
				return bVal.Bval, nil
			}
		}
		return false, moerr.NewInternalError("cannot eval filter expr")
	} else {
		vec, err := colexec.EvalExprByZonemapBat(bat, proc, expr)
		if err != nil {
			return false, err
		}
		if vec.Typ.Oid != types.T_bool {
			return false, moerr.NewInternalError("cannot eval filter expr")
		}
		cols := vector.MustTCols[bool](vec)
		for _, isNeed := range cols {
			if isNeed {
				return true, nil
			}
		}
		return false, nil
	}
}

func exchangeVectors(datas [][2]any, depth int, tmpResult []any, result *[]*vector.Vector, mp *mpool.MPool) {
	for i := 0; i < len(datas[depth]); i++ {
		tmpResult[depth] = datas[depth][i]
		if depth != len(datas)-1 {
			exchangeVectors(datas, depth+1, tmpResult, result, mp)
		} else {
			for j, val := range tmpResult {
				(*result)[j].Append(val, false, mp)
			}
		}
	}
}

func buildVectorsByData(datas [][2]any, dataTypes []uint8, mp *mpool.MPool) []*vector.Vector {
	vectors := make([]*vector.Vector, len(dataTypes))
	for i, typ := range dataTypes {
		vectors[i] = vector.New(types.T(typ).ToType())
	}

	tmpResult := make([]any, len(datas))
	exchangeVectors(datas, 0, tmpResult, &vectors, mp)

	return vectors
}

// getNameFromMeta  TODO change later
func getNameFromMeta(blkInfo BlockMeta) string {
	return fmt.Sprintf("%s:%d_%d_%d.blk", "local", blkInfo.header.blockId, blkInfo.header.segmentId, blkInfo.header.tableId)
}

// getExtentFromMeta  TODO change later
func getExtentFromMeta(blkInfo BlockMeta) objectio.Extent {
	return objectio.NewExtent(blkInfo.localExtent.offset, blkInfo.localExtent.length, blkInfo.localExtent.originSize)
}

// computeRange compute primaryKey range by Expr
// only under the following conditions：
// 1、function named ["and", "or", ">", "<", ">=", "<=", "="]
// 2、if function name is not "and", "or".  then one arg is column, the other is constant
func computeRange(expr *plan.Expr, pkIdx int32) (bool, [][2]int64) {
	type argType int
	var typeConstant argType = 0
	var typeColumn argType = 1
	var leftArg argType
	var leftConstant, rightConstat int64
	var ok bool

	getConstant := func(e *plan.Expr_C) (bool, int64) {
		switch val := e.C.Value.(type) {
		case *plan.Const_Ival:
			return true, val.Ival
		case *plan.Const_Uval:
			if val.Uval > uint64(math.MaxInt64) {
				return false, 0
			}
			return true, int64(val.Uval)
		}
		return false, 0
	}

	switch exprImpl := expr.Expr.(type) {
	case *plan.Expr_F:
		funName := exprImpl.F.Func.ObjName
		switch funName {
		case "and", "or":
			canCompute, leftRange := computeRange(exprImpl.F.Args[0], pkIdx)
			if !canCompute {
				return canCompute, nil
			}

			canCompute, rightRange := computeRange(exprImpl.F.Args[1], pkIdx)
			if !canCompute {
				return canCompute, nil
			}

			if funName == "and" {
				return true, _computeAnd(leftRange, rightRange)
			} else {
				return true, _computeOr(leftRange, rightRange)
			}

		case ">", "<", ">=", "<=", "=":
			switch subExpr := exprImpl.F.Args[0].Expr.(type) {
			case *plan.Expr_C:
				ok, leftConstant = getConstant(subExpr)
				if !ok {
					return false, nil
				}
				leftArg = typeConstant

			case *plan.Expr_Col:
				if subExpr.Col.ColPos != pkIdx {
					return false, nil
				}
				leftArg = typeColumn

			default:
				return false, nil
			}

			switch subExpr := exprImpl.F.Args[1].Expr.(type) {
			case *plan.Expr_C:
				if leftArg == typeColumn {
					ok, rightConstat = getConstant(subExpr)
					if !ok {
						return false, nil
					}
					switch funName {
					case ">":
						return true, [][2]int64{{rightConstat + 1, math.MaxInt64}}
					case ">=":
						return true, [][2]int64{{rightConstat, math.MaxInt64}}
					case "<":
						return true, [][2]int64{{math.MinInt64, rightConstat - 1}}
					case "<=":
						return true, [][2]int64{{math.MinInt64, rightConstat}}
					case "=":
						return true, [][2]int64{{rightConstat, rightConstat}}
					}
					return false, nil
				}
			case *plan.Expr_Col:
				if subExpr.Col.ColPos != pkIdx {
					return false, nil
				}

				if leftArg == typeConstant {
					switch funName {
					case ">":
						return true, [][2]int64{{math.MinInt64, leftConstant - 1}}
					case ">=":
						return true, [][2]int64{{math.MinInt64, leftConstant}}
					case "<":
						return true, [][2]int64{{leftConstant + 1, math.MaxInt64}}
					case "<=":
						return true, [][2]int64{{leftConstant, math.MaxInt64}}
					case "=":
						return true, [][2]int64{{leftConstant, leftConstant}}
					}
					return false, nil
				}
			}
		}
	}

	return false, nil
}

func _computeAnd(leftRange [][2]int64, rightRange [][2]int64) [][2]int64 {
	if len(leftRange) == 0 {
		return rightRange
	} else if len(rightRange) == 0 {
		return leftRange
	}

	compute := func(left [2]int64, right [2]int64) (bool, [2]int64) {
		min := left[0]
		max := left[1]

		if min > right[1] {
			// eg: a > 10 and a < 2
			return false, left
		} else if max < right[0] {
			// eg: a < 2 and a > 10
			return false, left
		} else {
			// eg: a > 2 and a < 10
			// a > 2 and a > 10
			// a > 2 and a = -2
			if right[0] > min {
				min = right[0]
			}
			if right[1] < max {
				max = right[1]
			}
			return true, [2]int64{min, max}
		}

	}

	// eg: (a >3 or a=1) and (a < 10 or a =11)
	var newRange [][2]int64
	for _, left := range leftRange {
		for _, right := range rightRange {
			ok, tmp := compute(left, right)
			if ok {
				newRange = append(newRange, tmp)
			}
		}
	}

	return newRange
}

func _computeOr(leftRange [][2]int64, rightRange [][2]int64) [][2]int64 {
	if len(leftRange) == 0 {
		return rightRange
	} else if len(rightRange) == 0 {
		return leftRange
	}

	compute := func(left [2]int64, right [2]int64) [][2]int64 {
		min := left[0]
		max := left[1]
		if min > right[1] {
			// eg: a > 10 or a < 2
			return [][2]int64{left, right}
		} else if max < right[0] {
			// eg: a < 2 or a > 10
			return [][2]int64{left, right}
		} else {
			// eg: a > 2 or a < 10
			// a > 2 or a > 10
			// a > 2 or a = -2
			if right[0] < min {
				min = right[0]
			}
			if right[1] > max {
				max = right[1]
			}
			return [][2]int64{{min, max}}
		}
	}

	// eg: (a>10 or a=1) or (a<5 or a=6)
	var newRange [][2]int64
	for _, left := range leftRange {
		for _, right := range rightRange {
			tmp := compute(left, right)
			newRange = append(newRange, tmp...)
		}
	}

	return newRange
}

func getListByRange[T DNStore](list []T, pkRange [][2]int64) []T {
	listLen := uint64(len(list))
	if listLen == 1 || len(pkRange) == 0 {
		return list
	}

	keys := make([]byte, 8)
	var states [3]uint64

	getHashValue := func(input int64) uint64 {
		binary.LittleEndian.PutUint64(keys, uint64(input))
		buf := append([]byte{0}, keys...)
		if l := len(buf); l < 16 {
			buf = append(buf, hashtable.StrKeyPadding[l:]...)
		}
		hashtable.BytesBatchGenHashStates(&buf, &states, 1)
		return states[0]
	}

	listMap := make(map[uint64]struct{})
	for _, r := range pkRange {
		for i := r[0]; i <= r[1]; i++ {
			val := getHashValue(i)
			modVal := val % listLen
			listMap[modVal] = struct{}{}
			if len(listMap) == int(listLen) {
				return list
			}
		}
	}

	returnList := make([]T, len(listMap))
	var i = 0
	for idx := range listMap {
		returnList[i] = list[idx]
		i = i + 1
	}

	return returnList
}
