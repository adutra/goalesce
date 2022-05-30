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

func Test_mainCoalescer_coalesceMap(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldIntPtr *int
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want interface{}
	}{
		{
			"map[string]int zero",
			(map[string]int)(nil),
			(map[string]int)(nil),
			(map[string]int)(nil),
		},
		{
			"map[string]int mixed zero",
			map[string]int{"a": 1},
			(map[string]int)(nil),
			map[string]int{"a": 1},
		},
		{
			"map[string]int mixed zero2",
			(map[string]int)(nil),
			map[string]int{"b": 2},
			map[string]int{"b": 2},
		},
		{
			"map[string]int empty",
			map[string]int{},
			map[string]int{},
			map[string]int{},
		},
		{
			"map[string]int mixed empty",
			map[string]int{"a": 1},
			map[string]int{},
			map[string]int{"a": 1},
		},
		{
			"map[string]int mixed empty 2",
			map[string]int{},
			map[string]int{"a": 2},
			map[string]int{"a": 2},
		},
		{
			"map[string]int non empty",
			map[string]int{"a": 1, "b": 1},
			map[string]int{"a": 2, "c": 2},
			map[string]int{"a": 2, "b": 1, "c": 2},
		},
		{
			"map[string]foo zero",
			(map[string]foo)(nil),
			(map[string]foo)(nil),
			(map[string]foo)(nil),
		},
		{
			"map[string]foo empty",
			map[string]foo{},
			map[string]foo{},
			map[string]foo{},
		},
		{
			"map[string]foo mixed empty",
			map[string]foo{"a": {1}},
			map[string]foo{},
			map[string]foo{"a": {1}},
		},
		{
			"map[string]foo mixed empty 2",
			map[string]foo{},
			map[string]foo{"b": {2}},
			map[string]foo{"b": {2}},
		},
		{
			"map[string]foo non empty",
			map[string]foo{"a": {1}, "b": {2}},
			map[string]foo{"a": {2}, "c": {2}},
			map[string]foo{"a": {2}, "b": {2}, "c": {2}},
		},
		{
			"map[string]*int zero",
			(map[string]*int)(nil),
			(map[string]*int)(nil),
			(map[string]*int)(nil),
		},
		{
			"map[string]*int mixed zero",
			map[string]*int{"a": nil},
			(map[string]*int)(nil),
			map[string]*int{"a": nil},
		},
		{
			"map[string]*int mixed zero2",
			(map[string]*int)(nil),
			map[string]*int{"b": nil},
			map[string]*int{"b": nil},
		},
		{
			"map[string]*int empty",
			map[string]*int{},
			map[string]*int{},
			map[string]*int{},
		},
		{
			"map[string]*int mixed empty",
			map[string]*int{"a": intPtr(1)},
			map[string]*int{},
			map[string]*int{"a": intPtr(1)},
		},
		{
			"map[string]*int mixed empty 2",
			map[string]*int{},
			map[string]*int{"a": intPtr(2)},
			map[string]*int{"a": intPtr(2)},
		},
		{
			"map[string]*int mixed empty 3",
			map[string]*int{"a": intPtr(1)},
			map[string]*int{"a": nil},
			map[string]*int{"a": intPtr(1)},
		},
		{
			"map[string]*int mixed empty 4",
			map[string]*int{"a": nil},
			map[string]*int{"a": intPtr(2)},
			map[string]*int{"a": intPtr(2)},
		},
		{
			"map[string]*int non empty",
			map[string]*int{"a": intPtr(1), "b": intPtr(1)},
			map[string]*int{"a": intPtr(2), "c": nil},
			map[string]*int{"a": intPtr(2), "b": intPtr(1), "c": nil},
		},
		{
			"map[string][]int empty",
			map[string][]int{},
			map[string][]int{},
			map[string][]int{},
		},
		{
			"map[string][]int mixed empty",
			map[string][]int{"a": {1}},
			map[string][]int{},
			map[string][]int{"a": {1}},
		},
		{
			"map[string][]int mixed empty 2",
			map[string][]int{},
			map[string][]int{"a": {2}},
			map[string][]int{"a": {2}},
		},
		{
			"map[string][]int mixed empty 3",
			map[string][]int{"a": {1}},
			map[string][]int{"a": nil},
			map[string][]int{"a": {1}},
		},
		{
			"map[string][]int mixed empty 4",
			map[string][]int{"a": nil},
			map[string][]int{"a": {2}},
			map[string][]int{"a": {2}},
		},
		{
			"map[string][]int non empty",
			map[string][]int{"a": {1}, "b": {1}},
			map[string][]int{"a": {1, 2}, "c": nil},
			map[string][]int{"a": {1, 2}, "b": {1}, "c": nil},
		},
		{
			"map[string]*bar empty",
			map[string]*bar{},
			map[string]*bar{},
			map[string]*bar{},
		},
		{
			"map[string]*bar mixed empty",
			map[string]*bar{"a": {intPtr(1)}},
			map[string]*bar{},
			map[string]*bar{"a": {intPtr(1)}},
		},
		{
			"map[string]*bar mixed empty 2",
			map[string]*bar{},
			map[string]*bar{"b": {intPtr(2)}},
			map[string]*bar{"b": {intPtr(2)}},
		},
		{
			"map[string]*bar nested nils",
			map[string]*bar{"a": {nil}, "b": nil},
			map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
		},
		{
			"map[string]*bar nested nils 2",
			map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			map[string]*bar{"a": {nil}, "b": nil},
			map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
		},
		{
			"map[string]*bar non empty",
			map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			map[string]*bar{"a": {intPtr(2)}, "c": {intPtr(2)}},
			map[string]*bar{"a": {intPtr(2)}, "b": {intPtr(2)}, "c": {intPtr(2)}},
		},
		{
			"map[string][]*int empty",
			map[string][]*int{},
			map[string][]*int{},
			map[string][]*int{},
		},
		{
			"map[string][]*int mixed empty",
			map[string][]*int{"a": {intPtr(1)}},
			map[string][]*int{},
			map[string][]*int{"a": {intPtr(1)}},
		},
		{
			"map[string][]*int mixed empty 2",
			map[string][]*int{},
			map[string][]*int{"a": {intPtr(2)}},
			map[string][]*int{"a": {intPtr(2)}},
		},
		{
			"map[string][]*int mixed empty 3",
			map[string][]*int{"a": {intPtr(1)}},
			map[string][]*int{"a": nil},
			map[string][]*int{"a": {intPtr(1)}},
		},
		{
			"map[string][]*int mixed empty 4",
			map[string][]*int{"a": nil},
			map[string][]*int{"a": {intPtr(2)}},
			map[string][]*int{"a": {intPtr(2)}},
		},
		{
			"map[string][]*int non empty",
			map[string][]*int{"a": {intPtr(1)}, "b": {intPtr(1)}},
			map[string][]*int{"a": {intPtr(1), intPtr(2)}, "c": nil},
			map[string][]*int{"a": {intPtr(1), intPtr(2)}, "b": {intPtr(1)}, "c": nil},
		},
		{
			"map[string]interface{} empty",
			map[string]interface{}{},
			map[string]interface{}{},
			map[string]interface{}{},
		},
		{
			"map[string]interface{} mixed empty",
			map[string]interface{}{"a": &bar{intPtr(1)}},
			map[string]interface{}{},
			map[string]interface{}{"a": &bar{intPtr(1)}},
		},
		{
			"map[string]interface{} mixed empty 2",
			map[string]interface{}{},
			map[string]interface{}{"b": &bar{intPtr(2)}},
			map[string]interface{}{"b": &bar{intPtr(2)}},
		},
		{
			"map[string]interface{} nested nils",
			map[string]interface{}{"a": &bar{nil}, "b": nil},
			map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
		},
		{
			"map[string]interface{} nested nils 2",
			map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			map[string]interface{}{"a": &bar{nil}, "b": nil},
			map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
		},
		{
			"map[string]interface{} non empty",
			map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			map[string]interface{}{"a": &bar{intPtr(2)}, "c": &bar{intPtr(2)}},
			map[string]interface{}{"a": &bar{intPtr(2)}, "b": &bar{intPtr(2)}, "c": &bar{intPtr(2)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewCoalescer()
			got, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("fallback error", func(t *testing.T) {
		coalescer := NewCoalescer(WithTypeCoalescer(reflect.TypeOf(0), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		_, err := coalescer(reflect.ValueOf(map[string]int{"a": 2}), reflect.ValueOf(map[string]int{"a": 2}))
		assert.EqualError(t, err, "fake")
	})
}
