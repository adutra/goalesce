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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestCoalesce(t *testing.T) {
	type foo struct {
		Int int
	}
	type bar struct {
		Int int
		Foo foo
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		opts []MainCoalescerOption
		want interface{}
	}{
		{
			"untyped nil",
			nil,
			nil,
			nil,
			nil,
		},
		{
			"untyped nil partial 1",
			1,
			nil,
			nil,
			1,
		},
		{
			"untyped nil partial 2",
			nil,
			1,
			nil,
			1,
		},
		{
			"int",
			1,
			2,
			nil,
			2,
		},
		{
			"int zero partial 1",
			1,
			0,
			nil,
			1,
		},
		{
			"int zero partial 2",
			0,
			1,
			nil,
			1,
		},
		{
			"*int",
			intPtr(1),
			intPtr(2),
			nil,
			intPtr(2),
		},
		{
			"*int zero",
			(*int)(nil),
			(*int)(nil),
			nil,
			(*int)(nil),
		},
		{
			"*int zero partial 1",
			intPtr(1),
			(*int)(nil),
			nil,
			intPtr(1),
		},
		{
			"nil *int partial 2",
			(*int)(nil),
			intPtr(1),
			nil,
			intPtr(1),
		},
		{
			"*int custom coalescer",
			intPtr(1),
			intPtr(2),
			[]MainCoalescerOption{WithPointerCoalescer(func() Coalescer {
				c := &mockCoalescer{}
				c.On("WithFallback", mock.Anything).Return()
				c.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.ValueOf(intPtr(3)), (error)(nil))
				return c
			}())},
			intPtr(3),
		},
		{
			"string",
			"a",
			"b",
			nil,
			"b",
		},
		{
			"string zero",
			"a",
			"",
			nil,
			"a",
		},
		{
			"bool",
			true,
			false,
			nil,
			true,
		},
		{
			"struct",
			foo{Int: 1},
			foo{Int: 2},
			nil,
			foo{Int: 2},
		},
		{
			"struct zero",
			foo{},
			foo{Int: 0},
			nil,
			foo{},
		},
		{
			"struct zero partial 1",
			foo{Int: 1},
			foo{Int: 0},
			nil,
			foo{Int: 1},
		},
		{
			"struct zero partial 2",
			foo{Int: 0},
			foo{Int: 1},
			nil,
			foo{Int: 1},
		},
		{
			"struct non zero",
			bar{Int: 0, Foo: foo{Int: 1}},
			bar{Int: 1},
			nil,
			bar{Int: 1, Foo: foo{Int: 1}},
		},
		{
			"struct non zero custom coalescer",
			bar{Int: 0, Foo: foo{Int: 1}},
			bar{Int: 1},
			[]MainCoalescerOption{WithStructCoalescer(NewDefaultCoalescer())},
			bar{Int: 1},
		},
		{
			"map[int]int",
			map[int]int{1: 1, 3: 1},
			map[int]int{1: 2, 2: 2},
			nil,
			map[int]int{1: 2, 2: 2, 3: 1},
		},
		{
			"map[int]foo",
			map[int]foo{1: {Int: 1}, 3: {Int: 3}},
			map[int]foo{1: {Int: 2}, 2: {Int: 2}},
			nil,
			map[int]foo{1: {Int: 2}, 2: {Int: 2}, 3: {Int: 3}},
		},
		{
			"map[int]foo custom coalescer",
			map[int]foo{1: {Int: 1}, 3: {Int: 3}},
			map[int]foo{1: {Int: 2}, 2: {Int: 2}},
			[]MainCoalescerOption{WithMapCoalescer(NewDefaultCoalescer())},
			map[int]foo{1: {Int: 2}, 2: {Int: 2}},
		},
		{
			"[]int",
			[]int{1, 3},
			[]int{1, 2},
			nil,
			[]int{1, 2},
		},
		{
			"[]foo",
			[]foo{{Int: 1}, {Int: 2}},
			[]foo{{Int: 3}, {Int: 4}},
			nil,
			[]foo{{Int: 3}, {Int: 4}},
		},
		{
			"[2]foo",
			[2]foo{{Int: 1}, {Int: 2}},
			[2]foo{{Int: 3}, {Int: 4}},
			nil,
			[2]foo{{Int: 3}, {Int: 4}},
		},
		{
			"[]int union",
			[]int{1, 3},
			[]int{1, 2},
			[]MainCoalescerOption{WithSliceCoalescer(NewSliceCoalescer(WithDefaultSetUnion()))},
			[]int{1, 3, 2},
		},
		{
			"[]int append",
			[]int{1, 3},
			[]int{1, 2},
			[]MainCoalescerOption{WithSliceCoalescer(NewSliceCoalescer(WithDefaultListAppend()))},
			[]int{1, 3, 1, 2},
		},
		{
			"[]foo custom",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]MainCoalescerOption{WithSliceCoalescer(NewSliceCoalescer(WithMergeByField(reflect.TypeOf(foo{}), "Int")))},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]*int custom",
			[]*foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]*foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]MainCoalescerOption{
				WithSliceCoalescer(
					NewSliceCoalescer(
						WithMergeByKey(
							reflect.PtrTo(reflect.TypeOf(foo{})),
							func(_ int, v reflect.Value) reflect.Value {
								i := v.Interface()
								return reflect.ValueOf(i.(*foo).Int)
							},
						)))},
			[]*foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]*int type coalescer",
			[]*foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]*foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]MainCoalescerOption{
				WithSliceCoalescer(NewSliceCoalescer(WithMergeByField(reflect.TypeOf(foo{}), "Int"))),
				WithTypeCoalescer(reflect.TypeOf([]*foo{}), &defaultCoalescer{})}, // will prevail
			[]*foo{{Int: 3}, {Int: 4}, {Int: 5}},
		},
		{"trilean nil nil", (*bool)(nil), (*bool)(nil), []MainCoalescerOption{WithTrileans()}, (*bool)(nil)},
		{"trilean nil false", (*bool)(nil), boolPtr(false), []MainCoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean nil true", (*bool)(nil), boolPtr(true), []MainCoalescerOption{WithTrileans()}, boolPtr(true)},
		{"trilean false nil", boolPtr(false), (*bool)(nil), []MainCoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean false false", boolPtr(false), boolPtr(false), []MainCoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean false true", boolPtr(false), boolPtr(true), []MainCoalescerOption{WithTrileans()}, boolPtr(true)},
		{"trilean true nil", boolPtr(true), (*bool)(nil), []MainCoalescerOption{WithTrileans()}, boolPtr(true)},
		// with trileans: Coalesce(true, false) = false
		{"trilean true false", boolPtr(true), boolPtr(false), []MainCoalescerOption{WithTrileans()}, boolPtr(false)},
		// without trileans: Coalesce(true, false) = true
		{"trilean true false", boolPtr(true), boolPtr(false), nil, boolPtr(true)},
		{"trilean true true", boolPtr(true), boolPtr(true), []MainCoalescerOption{WithTrileans()}, boolPtr(true)},
		{"type coalescer zero values", "", "", []MainCoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer{})}, "ZERO!ZERO!"},
		{"type coalescer zero value 1", "abc", "", []MainCoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer{})}, "abcZERO!"},
		{"type coalescer zero value 2", "", "def", []MainCoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer{})}, "ZERO!def"},
		{"type coalescer non-zero values", "abc", "def", []MainCoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer{})}, "abcdef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Coalesce(tt.v1, tt.v2, tt.opts...)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, err)
		})
	}
	t.Run("errors", func(t *testing.T) {
		_, err := Coalesce(1, "a")
		assert.EqualError(t, err, "types do not match: int != string")
	})
}

func TestMustCoalesce(t *testing.T) {
	assert.Equal(t, "def", MustCoalesce("abc", "def"))
	assert.PanicsWithError(t, "types do not match: int != string", func() {
		MustCoalesce(1, " abc")
	})
}

type weirdStringCoalescer struct{}

func (w weirdStringCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if v1.IsZero() {
		v1 = reflect.ValueOf("ZERO!")
	}
	if v2.IsZero() {
		v2 = reflect.ValueOf("ZERO!")
	}
	return reflect.ValueOf(v1.Interface().(string) + v2.Interface().(string)), nil
}

func (w weirdStringCoalescer) WithFallback(_ Coalescer) {}
