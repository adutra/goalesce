// Copyright 2022 Alexandre Dutra
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

package goalesce

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mainCoalescer_coalesceInterface(t *testing.T) {
	type foo struct {
		Int int
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want interface{}
	}{
		{
			"zeroes",
			nil,
			nil,
			nil,
		},
		{
			"mixed zeroes 1",
			123,
			nil,
			123,
		},
		{
			"mixed zeroes 2",
			nil,
			123,
			123,
		},
		{
			"ints zeroes",
			0,
			0,
			0,
		},
		{
			"ints",
			1,
			2,
			2,
		},
		{
			"structs zeroes",
			foo{},
			foo{},
			foo{},
		},
		{
			"structs",
			foo{Int: 1},
			foo{Int: 0},
			foo{Int: 1},
		},
		{
			"structs and empty structs",
			foo{Int: 1},
			foo{},
			foo{Int: 1},
		},
		{
			"maps",
			map[string]int{"a": 1},
			map[string]int{"b": 2},
			map[string]int{"a": 1, "b": 2},
		},
		{
			"*ints zeroes",
			(*int)(nil),
			(*int)(nil),
			(*int)(nil),
		},
		{
			"*ints",
			intPtr(1),
			intPtr(2),
			intPtr(2),
		},
		{
			"*structs zeroes",
			(*foo)(nil),
			(*foo)(nil),
			(*foo)(nil),
		},
		{
			"*structs",
			&foo{Int: 1},
			&foo{Int: 2},
			&foo{Int: 2},
		},
		{
			"different types",
			123,
			12.34,
			12.34,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewCoalescer()
			got, err := coalescer(reflect.ValueOf(&tt.v1).Elem(), reflect.ValueOf(&tt.v2).Elem())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("fallback error", func(t *testing.T) {
		coalescer := NewCoalescer(WithTypeCoalescer(reflect.TypeOf(0), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		var v1 interface{} = 1
		var v2 interface{} = 2
		_, err := coalescer(reflect.ValueOf(&v1).Elem(), reflect.ValueOf(&v2).Elem())
		assert.EqualError(t, err, "fake")
	})
}
