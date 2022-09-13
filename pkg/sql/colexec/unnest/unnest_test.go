// Copyright 2022 Matrix Origin
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

package unnest

import (
	"bytes"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/guest"
	"github.com/matrixorigin/matrixone/pkg/vm/mmu/host"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type unnestTestCase struct {
	arg        *Argument
	proc       *process.Process
	reg        *process.WaitRegister
	isCol      bool
	jsons      []string
	inputTimes int
}

var (
	utc            []unnestTestCase
	defaultAttrs   = []string{"col", "seq", "key", "path", "index", "value", "this"}
	defaultColDefs = []*plan.ColDef{
		{
			Name: "col",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    4,
			},
		},
		{
			Name: "seq",
			Typ: &plan.Type{
				Id:       int32(types.T_int32),
				Nullable: true,
			},
		},
		{
			Name: "key",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    256,
			},
		},
		{
			Name: "path",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    256,
			},
		},
		{
			Name: "index",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    4,
			},
		},
		{
			Name: "value",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    1024,
			},
		},
		{
			Name: "this",
			Typ: &plan.Type{
				Id:       int32(types.T_varchar),
				Nullable: true,
				Width:    1024,
			},
		},
	}
)

func init() {
	hm := host.New(1 << 30)
	gm := guest.New(1<<30, hm)
	utc = []unnestTestCase{
		newTestCase(mheap.New(gm), defaultAttrs, defaultColDefs, `{"a":1}`, "$", false, false, nil, 0),
		newTestCase(mheap.New(gm), defaultAttrs, defaultColDefs, tree.SetUnresolvedName("t1", "a"), "$", false, true, []string{`{"a":1}`}, 3),
	}
}

func newTestCase(m *mheap.Mheap, attrs []string, colDefs []*plan.ColDef, origin interface{}, path string, outer, isCol bool, jsons []string, inputTimes int) unnestTestCase {
	proc := testutil.NewProcessWithMheap(m)
	var reg *process.WaitRegister
	if isCol {
		proc.Reg.MergeReceivers = []*process.WaitRegister{
			{
				Ctx: proc.Ctx,
				Ch:  make(chan *batch.Batch, 1),
			},
		}
		reg = proc.Reg.MergeReceivers[0]
	}
	return unnestTestCase{
		proc: proc,
		arg: &Argument{
			Es: &Param{
				Attrs: attrs,
				Cols:  colDefs,
				Extern: &tree.UnnestParam{
					Origin: origin,
					Path:   path,
					Outer:  outer,
					IsCol:  isCol,
				},
			},
		},
		reg:        reg,
		isCol:      isCol,
		jsons:      jsons,
		inputTimes: inputTimes,
	}
}

func TestString(t *testing.T) {
	buf := new(bytes.Buffer)
	for _, ut := range utc {
		String(ut.arg, buf)
	}
}

func TestUnnest(t *testing.T) {
	for _, ut := range utc {
		err := Prepare(ut.proc, ut.arg)
		require.Nil(t, err)
		if !ut.isCol {
			end, err := Call(0, ut.proc, ut.arg)
			require.Nil(t, err)
			require.False(t, end)
			require.True(t, ut.arg.Es.end)
			require.NotNil(t, ut.proc.InputBatch())
			ut.proc.SetInputBatch(nil)
			end, err = Call(0, ut.proc, ut.arg)
			require.Nil(t, err)
			require.True(t, end)
			require.True(t, ut.arg.Es.end)
			require.Nil(t, ut.proc.InputBatch())
			continue
		}
		wg := sync.WaitGroup{}
		wg.Add(ut.inputTimes + 1)
		go func(t *testing.T, proc *process.Process, arg *Argument, wg *sync.WaitGroup) {
			end := false
			for !end {
				var err error
				end, err = Call(0, proc, arg)
				require.Nil(t, err)
				require.False(t, arg.Es.end)
				if end {
					require.Nil(t, proc.InputBatch())
				} else {
					require.NotNil(t, proc.InputBatch())
				}
				wg.Done()
			}
		}(t, ut.proc, ut.arg, &wg)
		for i := 0; i < ut.inputTimes; i++ {
			bat, err := makeTestBatch(ut.jsons, ut.proc)
			require.Nil(t, err)
			require.NotNil(t, bat)
			ut.reg.Ch <- bat
		}
		ut.reg.Ch <- nil // terminate
		wg.Wait()
	}
}
func makeTestBatch(jsons []string, proc *process.Process) (*batch.Batch, error) {
	bat := batch.New(true, []string{"a"})
	for i := range bat.Vecs {
		bat.Vecs[i] = vector.New(types.Type{
			Oid:   types.T_json,
			Width: 256,
		})
	}
	for _, json := range jsons {
		bj, err := types.ParseStringToByteJson(json)
		if err != nil {
			return nil, err
		}
		bjBytes, err := types.EncodeJson(bj)
		if err != nil {
			return nil, err
		}
		err = bat.GetVector(0).Append(bjBytes, false, proc.Mp())
	}
	return bat, nil
}
