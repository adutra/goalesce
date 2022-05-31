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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoalesce(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldInt int
		FieldFoo foo
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		opts []CoalescerOption
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
			foo{FieldInt: 1},
			foo{FieldInt: 2},
			nil,
			foo{FieldInt: 2},
		},
		{
			"struct zero",
			foo{},
			foo{FieldInt: 0},
			nil,
			foo{},
		},
		{
			"struct zero partial 1",
			foo{FieldInt: 1},
			foo{FieldInt: 0},
			nil,
			foo{FieldInt: 1},
		},
		{
			"struct zero partial 2",
			foo{FieldInt: 0},
			foo{FieldInt: 1},
			nil,
			foo{FieldInt: 1},
		},
		{
			"struct non zero",
			bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}},
			bar{FieldInt: 1},
			nil,
			bar{FieldInt: 1, FieldFoo: foo{FieldInt: 1}},
		},
		{
			"struct non zero custom coalescer",
			bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}},
			bar{FieldInt: 1},
			[]CoalescerOption{WithAtomicType(reflect.TypeOf(bar{}))},
			bar{FieldInt: 1},
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
			map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}},
			map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
			nil,
			map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}, 3: {FieldInt: 3}},
		},
		{
			"map[int]foo custom coalescer",
			map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}},
			map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
			[]CoalescerOption{WithAtomicType(reflect.TypeOf(map[int]foo{}))},
			map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
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
			[]foo{{FieldInt: 1}, {FieldInt: 2}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}},
			nil,
			[]foo{{FieldInt: 3}, {FieldInt: 4}},
		},
		{
			"[2]foo",
			[2]foo{{FieldInt: 1}, {FieldInt: 2}},
			[2]foo{{FieldInt: 3}, {FieldInt: 4}},
			nil,
			[2]foo{{FieldInt: 3}, {FieldInt: 4}},
		},
		{
			"[]int union",
			[]int{1, 3},
			[]int{1, 2},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{1, 3, 2},
		},
		{
			"[]int append",
			[]int{1, 3},
			[]int{1, 2},
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{1, 3, 1, 2},
		},
		{
			"[]foo custom",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]*int custom",
			[]*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{
				WithMergeByKeyFunc(
					reflect.TypeOf([]*foo{}),
					func(_ int, v reflect.Value) reflect.Value {
						i := v.Interface()
						return reflect.ValueOf(i.(*foo).FieldInt)
					},
				)},
			[]*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]*int type coalescer",
			[]*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{
				WithMergeByID(reflect.TypeOf([]*foo{}), "FieldInt"),
				WithTypeCoalescer(reflect.TypeOf([]*foo{}), coalesceAtomic)}, // will prevail
			[]*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{"trilean nil nil", (*bool)(nil), (*bool)(nil), []CoalescerOption{WithTrileans()}, (*bool)(nil)},
		{"trilean nil false", (*bool)(nil), boolPtr(false), []CoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean nil true", (*bool)(nil), boolPtr(true), []CoalescerOption{WithTrileans()}, boolPtr(true)},
		{"trilean false nil", boolPtr(false), (*bool)(nil), []CoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean false false", boolPtr(false), boolPtr(false), []CoalescerOption{WithTrileans()}, boolPtr(false)},
		{"trilean false true", boolPtr(false), boolPtr(true), []CoalescerOption{WithTrileans()}, boolPtr(true)},
		{"trilean true nil", boolPtr(true), (*bool)(nil), []CoalescerOption{WithTrileans()}, boolPtr(true)},
		// with trileans: coalesce(true, false) = false
		{"trilean true false", boolPtr(true), boolPtr(false), []CoalescerOption{WithTrileans()}, boolPtr(false)},
		// without trileans: coalesce(true, false) = true
		{"trilean true false", boolPtr(true), boolPtr(false), nil, boolPtr(true)},
		{"trilean true true", boolPtr(true), boolPtr(true), []CoalescerOption{WithTrileans()}, boolPtr(true)},
		{"type coalescer zero values", "", "", []CoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer)}, "ZERO!ZERO!"},
		{"type coalescer zero value 1", "abc", "", []CoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer)}, "abcZERO!"},
		{"type coalescer zero value 2", "", "def", []CoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer)}, "ZERO!def"},
		{"type coalescer non-zero values", "abc", "def", []CoalescerOption{WithTypeCoalescer(reflect.TypeOf(""), weirdStringCoalescer)}, "abcdef"},
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

func weirdStringCoalescer(v1, v2 reflect.Value) (reflect.Value, error) {
	if v1.IsZero() {
		v1 = reflect.ValueOf("ZERO!")
	}
	if v2.IsZero() {
		v2 = reflect.ValueOf("ZERO!")
	}
	return reflect.ValueOf(v1.Interface().(string) + v2.Interface().(string)), nil
}
