// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate_series

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

type Kase[T Number] struct {
	start T
	end   T
	step  T
	res   []T
	err   bool
}

func TestDoGenerateInt32(t *testing.T) {
	kases := []Kase[int32]{
		{
			start: 1,
			end:   10,
			step:  1,
			res:   []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			start: 1,
			end:   10,
			step:  2,
			res:   []int32{1, 3, 5, 7, 9},
		},
		{
			start: 1,
			end:   10,
			step:  -1,
			err:   true,
		},
		{
			start: 10,
			end:   1,
			step:  -1,
			res:   []int32{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			start: 10,
			end:   1,
			step:  -2,
			res:   []int32{10, 8, 6, 4, 2},
		},
		{
			start: 10,
			end:   1,
			step:  1,
			err:   true,
		},
		{
			start: 1,
			end:   10,
			step:  0,
			err:   true,
		},
		{
			start: 1,
			end:   10,
			step:  -1,
			err:   true,
		},
		{
			start: 1,
			end:   1,
			step:  0,
			res:   []int32{1},
		},
		{
			start: 1,
			end:   1,
			step:  1,
			res:   []int32{1},
		},
		{
			start: math.MaxInt32 - 1,
			end:   math.MaxInt32,
			step:  1,
			res:   []int32{math.MaxInt32 - 1, math.MaxInt32},
		},
		{
			start: math.MaxInt32,
			end:   math.MaxInt32 - 1,
			step:  -1,
			res:   []int32{math.MaxInt32, math.MaxInt32 - 1},
		},
		{
			start: math.MinInt32,
			end:   math.MinInt32 + 100,
			step:  19,
			res:   []int32{math.MinInt32, math.MinInt32 + 19, math.MinInt32 + 38, math.MinInt32 + 57, math.MinInt32 + 76, math.MinInt32 + 95},
		},
		{
			start: math.MinInt32 + 100,
			end:   math.MinInt32,
			step:  -19,
			res:   []int32{math.MinInt32 + 100, math.MinInt32 + 81, math.MinInt32 + 62, math.MinInt32 + 43, math.MinInt32 + 24, math.MinInt32 + 5},
		},
	}
	for _, kase := range kases {
		res, err := generateInt32(kase.start, kase.end, kase.step)
		if kase.err {
			require.NotNil(t, err)
			continue
		}
		require.Nil(t, err)
		require.Equal(t, kase.res, res)
	}
}

func TestDoGenerateInt64(t *testing.T) {
	kases := []Kase[int64]{
		{
			start: 1,
			end:   10,
			step:  1,
			res:   []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			start: 1,
			end:   10,
			step:  2,
			res:   []int64{1, 3, 5, 7, 9},
		},
		{
			start: 1,
			end:   10,
			step:  -1,
			err:   true,
		},
		{
			start: 10,
			end:   1,
			step:  -1,
			res:   []int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		},
		{
			start: 10,
			end:   1,
			step:  -2,
			res:   []int64{10, 8, 6, 4, 2},
		},
		{
			start: 10,
			end:   1,
			step:  1,
			err:   true,
		},
		{
			start: 1,
			end:   10,
			step:  0,
			err:   true,
		},
		{
			start: 1,
			end:   10,
			step:  -1,
			err:   true,
		},
		{
			start: 1,
			end:   1,
			step:  0,
			res:   []int64{1},
		},
		{
			start: 1,
			end:   1,
			step:  1,
			res:   []int64{1},
		},
		{
			start: math.MaxInt32 - 1,
			end:   math.MaxInt32,
			step:  1,
			res:   []int64{math.MaxInt32 - 1, math.MaxInt32},
		},
		{
			start: math.MaxInt32,
			end:   math.MaxInt32 - 1,
			step:  -1,
			res:   []int64{math.MaxInt32, math.MaxInt32 - 1},
		},
		{
			start: math.MinInt32,
			end:   math.MinInt32 + 100,
			step:  19,
			res:   []int64{math.MinInt32, math.MinInt32 + 19, math.MinInt32 + 38, math.MinInt32 + 57, math.MinInt32 + 76, math.MinInt32 + 95},
		},
		{
			start: math.MinInt32 + 100,
			end:   math.MinInt32,
			step:  -19,
			res:   []int64{math.MinInt32 + 100, math.MinInt32 + 81, math.MinInt32 + 62, math.MinInt32 + 43, math.MinInt32 + 24, math.MinInt32 + 5},
		},
		// int64
		{
			start: math.MaxInt32,
			end:   math.MaxInt32 + 100,
			step:  19,
			res:   []int64{math.MaxInt32, math.MaxInt32 + 19, math.MaxInt32 + 38, math.MaxInt32 + 57, math.MaxInt32 + 76, math.MaxInt32 + 95},
		},
		{
			start: math.MaxInt32 + 100,
			end:   math.MaxInt32,
			step:  -19,
			res:   []int64{math.MaxInt32 + 100, math.MaxInt32 + 81, math.MaxInt32 + 62, math.MaxInt32 + 43, math.MaxInt32 + 24, math.MaxInt32 + 5},
		},
		{
			start: math.MinInt32,
			end:   math.MinInt32 - 100,
			step:  -19,
			res:   []int64{math.MinInt32, math.MinInt32 - 19, math.MinInt32 - 38, math.MinInt32 - 57, math.MinInt32 - 76, math.MinInt32 - 95},
		},
		{
			start: math.MinInt32 - 100,
			end:   math.MinInt32,
			step:  19,
			res:   []int64{math.MinInt32 - 100, math.MinInt32 - 81, math.MinInt32 - 62, math.MinInt32 - 43, math.MinInt32 - 24, math.MinInt32 - 5},
		},
		{
			start: math.MaxInt64 - 1,
			end:   math.MaxInt64,
			step:  1,
			res:   []int64{math.MaxInt64 - 1, math.MaxInt64},
		},
		{
			start: math.MaxInt64,
			end:   math.MaxInt64 - 1,
			step:  -1,
			res:   []int64{math.MaxInt64, math.MaxInt64 - 1},
		},
		{
			start:math.MaxInt64-100,
			end:math.MaxInt64,
			step:19,
			res:[]int64{math.MaxInt64-100,math.MaxInt64-81,math.MaxInt64-62,math.MaxInt64-43,math.MaxInt64-24,math.MaxInt64-5},
		},
		{
			start:math.MaxInt64,
			end:math.MaxInt64-100,
			step:-19,
			res:[]int64{math.MaxInt64,math.MaxInt64-19,math.MaxInt64-38,math.MaxInt64-57,math.MaxInt64-76,math.MaxInt64-95},
		},
		{
			start: math.MinInt64,
			end:   math.MinInt64 + 100,
			step:  19,
			res:   []int64{math.MinInt64, math.MinInt64 + 19, math.MinInt64 + 38, math.MinInt64 + 57, math.MinInt64 + 76, math.MinInt64 + 95},
		},
		{
			start: math.MinInt64 + 100,
			end:   math.MinInt64,
			step:  -19,
			res:   []int64{math.MinInt64 + 100, math.MinInt64 + 81, math.MinInt64 + 62, math.MinInt64 + 43, math.MinInt64 + 24, math.MinInt64 + 5},
		},
	}
	for _, kase := range kases {
		res, err := generateInt64(kase.start, kase.end, kase.step)
		if kase.err {
			require.NotNil(t, err)
			continue
		}
		require.Nil(t, err)
		require.Equal(t, kase.res, res)
	}
}
