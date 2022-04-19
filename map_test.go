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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestNewMapCoalescer(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		got := NewMapCoalescer()
		assert.Equal(t, &mapCoalescer{fallback: &defaultCoalescer{}}, got)
	})
	t.Run("with generic option", func(t *testing.T) {
		var passed *mapCoalescer
		opt := func(c *mapCoalescer) {
			passed = c
		}
		returned := NewMapCoalescer(opt)
		assert.Equal(t, &mapCoalescer{fallback: &defaultCoalescer{}}, returned)
		assert.Equal(t, returned, passed)
	})
}

func Test_mapCoalescer_Coalesce(t *testing.T) {
	type foo struct {
		Int int
	}
	type bar struct {
		Int *int
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
			map[string]*bar{"a": {nil}, "b": {intPtr(2)}},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewMapCoalescer()
			got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("type errors", func(t *testing.T) {
		_, err := NewMapCoalescer().Coalesce(reflect.ValueOf(map[string]int{"a": 2}), reflect.ValueOf(map[string]string{"a": "b"}))
		assert.EqualError(t, err, "types do not match: map[string]int != map[string]string")
		_, err = NewMapCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected map, got int")
	})
	t.Run("fallback error", func(t *testing.T) {
		coalescer := NewMapCoalescer()
		m := new(mockCoalescer)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		m.Test(t)
		coalescer.WithFallback(m)
		_, err := coalescer.Coalesce(reflect.ValueOf(map[string]int{"a": 2}), reflect.ValueOf(map[string]int{"a": 2}))
		assert.EqualError(t, err, "fake")
	})
}
