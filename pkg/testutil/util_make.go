// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
)

// All vectors generated by the Make Function, their memory is not allocated through the memory pool
// if you want to generate a vector in memory pool, use NewFunction to instead of MakeFunction.
var (
	TestUtilMp = mpool.MustNewZeroNoFixed()

	MakeBoolVector = func(values []bool) *vector.Vector {
		return makeVector(values, nil, boolType)
	}

	MakeBooleanlVector = func(values []bool, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, boolType)
	}

	MakeInt64Vector = func(values []int64, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, int64Type)
	}

	MakeInt32Vector = func(values []int32, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, int32Type)
	}

	MakeInt16Vector = func(values []int16, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, int16Type)
	}

	MakeInt8Vector = func(values []int8, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, int8Type)
	}

	MakeUint64Vector = func(values []uint64, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, uint64Type)
	}

	MakeUint32Vector = func(values []uint32, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, uint32Type)
	}

	MakeUint16Vector = func(values []uint16, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, uint16Type)
	}

	MakeUint8Vector = func(values []uint8, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, uint8Type)
	}

	MakeFloat32Vector = func(values []float32, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, float32Type)
	}

	MakeFloat64Vector = func(values []float64, nsp []uint64) *vector.Vector {
		return makeVector(values, nsp, float64Type)
	}

	MakeCharVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, charType)
	}

	MakeVarcharVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, varcharType)
	}

	MakeBinaryVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, binaryType)
	}

	MakeVarbinaryVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, varbinaryType)
	}

	MakeBlobVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, blobType)
	}
	MakeTextVector = func(values []string, nsp []uint64) *vector.Vector {
		return makeStringVector(values, nsp, textType)
	}
	MakeDecimal64Vector = func(values []int64, nsp []uint64, _ types.Type) *vector.Vector {
		cols := make([]types.Decimal64, len(values))
		for i, v := range values {
			d, _ := types.InitDecimal64(v, 64, 0)
			cols[i] = d
		}
		return makeVector(cols, nsp, decimal64Type)
	}

	MakeDecimal128Vector = func(values []int64, nsp []uint64, _ types.Type) *vector.Vector {
		cols := make([]types.Decimal128, len(values))
		for i, v := range values {
			d, _ := types.InitDecimal128(v, 64, 0)
			cols[i] = d
		}
		return makeVector(cols, nsp, decimal128Type)
	}

	MakeDateVector = func(values []string, nsp []uint64) *vector.Vector {
		ds := make([]types.Date, len(values))
		ns := nulls.Build(len(values), nsp...)
		for i, s := range values {
			if nulls.Contains(ns, uint64(i)) {
				continue
			}
			d, err := types.ParseDateCast(s)
			if err != nil {
				panic(err)
			}
			ds[i] = d
		}
		vec := vector.NewVec(types.T_date.ToType())
		vector.AppendFixedList(vec, ds, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}

	MakeTimeVector = func(values []string, nsp []uint64) *vector.Vector {
		ds := make([]types.Time, len(values))
		ns := nulls.Build(len(values), nsp...)
		for i, s := range values {
			if nulls.Contains(ns, uint64(i)) {
				continue
			}
			d, err := types.ParseTime(s, 6)
			if err != nil {
				panic(err)
			}
			ds[i] = d
		}
		vec := vector.NewVec(types.T_time.ToType())
		vector.AppendFixedList(vec, ds, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}

	MakeDateTimeVector = func(values []string, nsp []uint64) *vector.Vector {
		ds := make([]types.Datetime, len(values))
		ns := nulls.Build(len(values), nsp...)
		for i, s := range values {
			if nulls.Contains(ns, uint64(i)) {
				continue
			}
			d, err := types.ParseDatetime(s, 6)
			if err != nil {
				panic(err)
			}
			ds[i] = d
		}
		vec := vector.NewVec(types.T_datetime.ToType())
		vector.AppendFixedList(vec, ds, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}

	MakeTimeStampVector = func(values []string, nsp []uint64) *vector.Vector {
		ds := make([]types.Timestamp, len(values))
		ns := nulls.Build(len(values), nsp...)
		for i, s := range values {
			if nulls.Contains(ns, uint64(i)) {
				continue
			}
			d, err := types.ParseTimestamp(time.Local, s, 6)
			if err != nil {
				panic(err)
			}
			ds[i] = d
		}
		vec := vector.NewVec(types.T_timestamp.ToType())
		vector.AppendFixedList(vec, ds, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}

	MakeUuidVector = func(values []types.Uuid, nsp []uint64) *vector.Vector {
		ns := nulls.Build(len(values), nsp...)
		vec := vector.NewVec(uuidType)
		vector.AppendFixedList(vec, values, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}

	MakeUuidVectorByString = func(values []string, nsp []uint64) *vector.Vector {
		ds := make([]types.Uuid, len(values))
		ns := nulls.Build(len(values), nsp...)
		for i, s := range values {
			if nulls.Contains(ns, uint64(i)) {
				continue
			}
			d, err := types.ParseUuid(s)
			if err != nil {
				panic(err)
			}
			ds[i] = d
		}
		vec := vector.NewVec(types.T_uuid.ToType())
		vector.AppendFixedList(vec, ds, nil, TestUtilMp)
		vec.SetNulls(ns)
		return vec
	}
)

// functions to make a scalar vector for test.
var (
	MakeScalarNull = func(typ types.T, length int) *vector.Vector {
		return vector.NewConstNull(typ.ToType(), length, NewProc().Mp())
	}

	MakeScalarBool = func(v bool, length int) *vector.Vector {
		return makeScalar(v, length, boolType)
	}

	MakeScalarInt64 = func(v int64, length int) *vector.Vector {
		return makeScalar(v, length, int64Type)
	}

	MakeScalarInt32 = func(v int32, length int) *vector.Vector {
		return makeScalar(v, length, int32Type)
	}

	MakeScalarInt16 = func(v int16, length int) *vector.Vector {
		return makeScalar(v, length, int16Type)
	}

	MakeScalarInt8 = func(v int8, length int) *vector.Vector {
		return makeScalar(v, length, int8Type)
	}

	MakeScalarUint64 = func(v uint64, length int) *vector.Vector {
		return makeScalar(v, length, uint64Type)
	}

	MakeScalarUint3 = func(v uint32, length int) *vector.Vector {
		return makeScalar(v, length, uint32Type)
	}

	MakeScalarUint16 = func(v uint16, length int) *vector.Vector {
		return makeScalar(v, length, uint16Type)
	}

	MakeScalarUint8 = func(v uint8, length int) *vector.Vector {
		return makeScalar(v, length, uint8Type)
	}

	MakeScalarFloat32 = func(v float32, length int) *vector.Vector {
		return makeScalar(v, length, float32Type)
	}

	MakeScalarFloat64 = func(v float64, length int) *vector.Vector {
		return makeScalar(v, length, float64Type)
	}

	MakeScalarChar = func(value string, length int) *vector.Vector {
		return makeScalarString(value, length, charType)
	}

	MakeScalarVarchar = func(value string, length int) *vector.Vector {
		return makeScalarString(value, length, varcharType)
	}

	MakeTextVarchar = func(value string, length int) *vector.Vector {
		return makeScalarString(value, length, textType)
	}

	MakeScalarDate = func(value string, length int) *vector.Vector {
		d, err := types.ParseDateCast(value)
		if err != nil {
			panic(err)
		}
		return makeScalar(d, length, dateType)
	}

	MakeScalarTime = func(value string, length int) *vector.Vector {
		d, err := types.ParseTime(value, 6)
		if err != nil {
			panic(err)
		}
		return makeScalar(d, length, timeType)
	}

	MakeScalarDateTime = func(value string, length int) *vector.Vector {
		d, err := types.ParseDatetime(value, 6)
		if err != nil {
			panic(err)
		}
		return makeScalar(d, length, datetimeType)
	}

	MakeScalarTimeStamp = func(value string, length int) *vector.Vector {
		d, err := types.ParseTimestamp(time.Local, value, 6)
		if err != nil {
			panic(err)
		}
		return makeScalar(d, length, timestampType)
	}

	MakeScalarDecimal64 = func(v int64, length int, _ types.Type) *vector.Vector {
		d, _ := types.InitDecimal64(v, 64, 0)
		return makeScalar(d, length, decimal64Type)
	}

	MakeScalarDecimal128 = func(v uint64, length int, _ types.Type) *vector.Vector {
		d, _ := types.InitDecimal128UsingUint(v, 64, 0)
		return makeScalar(d, length, decimal128Type)
	}

	MakeScalarDecimal128ByFloat64 = func(v float64, length int, _ types.Type) *vector.Vector {
		val := fmt.Sprintf("%f", v)
		_, scale, err := types.ParseStringToDecimal128WithoutTable(val)
		if err != nil {
			panic(err)
		}
		dec128Val, err := types.ParseStringToDecimal128(val, 34, scale, false)
		if err != nil {
			panic(err)
		}
		return makeScalar(dec128Val, length, decimal128Type)
	}
)

func makeVector[T types.FixedSizeT](values []T, nsp []uint64, typ types.Type) *vector.Vector {
	vec := vector.NewVec(typ)
	vector.AppendFixedList(vec, values, nil, TestUtilMp)
	vec.SetNulls(nulls.Build(len(values), nsp...))
	return vec
}

func makeStringVector(values []string, nsp []uint64, typ types.Type) *vector.Vector {
	if nsp == nil {
		vec := vector.NewVec(typ)
		vector.AppendStringList(vec, values, nil, TestUtilMp)
		return vec
	} else {
		vec := vector.NewVec(typ)
		vector.AppendStringList(vec, values, nil, TestUtilMp)
		vec.SetNulls(nulls.Build(len(values), nsp...))
		return vec
	}
}

func makeScalar[T types.FixedSizeT](value T, length int, typ types.Type) *vector.Vector {
	return vector.NewConstFixed(typ, value, length, TestUtilMp)
}

func makeScalarString(value string, length int, typ types.Type) *vector.Vector {
	return vector.NewConstBytes(typ, []byte(value), length, TestUtilMp)
}

func MakeDecimal64ArrByInt64Arr(input []int64) []types.Decimal64 {
	ret := make([]types.Decimal64, len(input))
	for i, v := range input {
		d, _ := types.InitDecimal64(v, 64, 0)
		ret[i] = d
	}

	return ret
}

func MakeDecimal64ArrByFloat64Arr(input []float64) []types.Decimal64 {
	ret := make([]types.Decimal64, len(input))
	for i, v := range input {
		d, _ := types.Decimal64FromFloat64(v, 64, 10)
		ret[i] = d
	}

	return ret
}

func MakeDecimal128ArrByInt64Arr(input []int64) []types.Decimal128 {
	ret := make([]types.Decimal128, len(input))
	for i, v := range input {
		d, _ := types.InitDecimal128(v, 64, 0)
		ret[i] = d
	}

	return ret
}

func MakeDecimal128ArrByFloat64Arr(input []float64) []types.Decimal128 {
	ret := make([]types.Decimal128, len(input))
	for i, v := range input {
		d, _ := types.Decimal128FromFloat64(v, 64, 10)
		ret[i] = d
	}

	return ret
}

func MakeRandomStrings(cardinality, targetRows int) []string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	charsLen := len(chars)

	// mock random strings
	dataset := make([][]byte, cardinality)
	for i := range dataset {
		randLen := rand.Intn(charsLen) + 1
		dataset[i] = make([]byte, randLen)
		for j := range dataset[i] {
			dataset[i][j] = chars[rand.Intn(charsLen)]
		}
	}

	data := make([]string, 0)
	for {
		for i := range dataset {
			n := rand.Intn(10) + 1
			for j := 0; j < n; j++ {
				data = append(data, string(append([]byte{}, dataset[i]...)))
			}
		}
		if len(data) >= targetRows {
			break
		}
	}
	return data
}
