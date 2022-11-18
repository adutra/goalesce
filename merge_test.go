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

func TestDeepMerge(t *testing.T) {
	// Note: we don't need to test all the types and corner cases here, as the underlying merge
	// functions are thoroughly tested.
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldInt int
		FieldFoo foo
	}
	tests := []struct {
		name    string
		v1      interface{}
		v2      interface{}
		opts    []Option
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "untyped nil",
			v1:   nil,
			v2:   nil,
			want: nil,
		},
		{
			name: "untyped nil partial 1",
			v1:   1,
			v2:   nil,
			want: 1,
		},
		{
			name: "untyped nil partial 2",
			v1:   nil,
			v2:   1,
			want: 1,
		},
		{
			name: "int",
			v1:   1,
			v2:   2,
			want: 2,
		},
		{
			name: "int zero partial 1",
			v1:   1,
			v2:   0,
			want: 1,
		},
		{
			name: "int zero partial 2",
			v1:   0,
			v2:   1,
			want: 1,
		},
		{
			name: "*int",
			v1:   intPtr(1),
			v2:   intPtr(2),
			want: intPtr(2),
		},
		{
			name: "*int zero",
			v1:   (*int)(nil),
			v2:   (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "*int zero partial 1",
			v1:   intPtr(1),
			v2:   (*int)(nil),
			want: intPtr(1),
		},
		{
			name: "nil *int partial 2",
			v1:   (*int)(nil),
			v2:   intPtr(1),
			want: intPtr(1),
		},
		{
			name: "string",
			v1:   "a",
			v2:   "b",
			want: "b",
		},
		{
			name: "string zero",
			v1:   "a",
			v2:   "",
			want: "a",
		},
		{
			name: "bool",
			v1:   true,
			v2:   false,
			want: true,
		},
		{
			name: "struct",
			v1:   foo{FieldInt: 1},
			v2:   foo{FieldInt: 2},
			want: foo{FieldInt: 2},
		},
		{
			name: "struct zero",
			v1:   foo{},
			v2:   foo{FieldInt: 0},
			want: foo{},
		},
		{
			name: "struct zero partial 1",
			v1:   foo{FieldInt: 1},
			v2:   foo{FieldInt: 0},
			want: foo{FieldInt: 1},
		},
		{
			name: "struct zero partial 2",
			v1:   foo{FieldInt: 0},
			v2:   foo{FieldInt: 1},
			want: foo{FieldInt: 1},
		},
		{
			name: "struct non zero",
			v1:   bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}},
			v2:   bar{FieldInt: 1},
			want: bar{FieldInt: 1, FieldFoo: foo{FieldInt: 1}},
		},
		{
			name: "struct non zero custom coalescer",
			v1:   bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}},
			v2:   bar{FieldInt: 1},
			opts: []Option{WithAtomicMerge(reflect.TypeOf(bar{}))},
			want: bar{FieldInt: 1},
		},
		{
			name: "map[int]int",
			v1:   map[int]int{1: 1, 3: 1},
			v2:   map[int]int{1: 2, 2: 2},
			want: map[int]int{1: 2, 2: 2, 3: 1},
		},
		{
			name: "map[int]foo",
			v1:   map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}},
			v2:   map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
			want: map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}, 3: {FieldInt: 3}},
		},
		{
			name: "map[int]foo custom coalescer",
			v1:   map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}},
			v2:   map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
			opts: []Option{WithAtomicMerge(reflect.TypeOf(map[int]foo{}))},
			want: map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}},
		},
		{
			name: "[]int",
			v1:   []int{1, 3},
			v2:   []int{1, 2},
			want: []int{1, 2},
		},
		{
			name: "[]foo",
			v1:   []foo{{FieldInt: 1}, {FieldInt: 2}},
			v2:   []foo{{FieldInt: 3}, {FieldInt: 4}},
			want: []foo{{FieldInt: 3}, {FieldInt: 4}},
		},
		{
			name: "[2]foo",
			v1:   [2]foo{{FieldInt: 1}, {FieldInt: 2}},
			v2:   [2]foo{{FieldInt: 3}, {FieldInt: 4}},
			want: [2]foo{{FieldInt: 3}, {FieldInt: 4}},
		},
		{
			name: "[]int union",
			v1:   []int{1, 3},
			v2:   []int{1, 2},
			opts: []Option{WithDefaultSliceSetUnionMerge()},
			want: []int{1, 3, 2},
		},
		{
			name: "[]int append",
			v1:   []int{1, 3},
			v2:   []int{1, 2},
			opts: []Option{WithDefaultSliceListAppendMerge()},
			want: []int{1, 3, 1, 2},
		},
		{
			name: "[]foo custom",
			v1:   []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			v2:   []foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			opts: []Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			want: []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			name: "[]*int custom",
			v1:   []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			v2:   []*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			opts: []Option{
				WithSliceMergeByKeyFunc(
					reflect.TypeOf([]*foo{}),
					func(_ int, v reflect.Value) (reflect.Value, error) {
						i := v.Interface()
						return reflect.ValueOf(i.(*foo).FieldInt), nil
					},
				)},
			want: []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			name: "[]*int type merger",
			v1:   []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			v2:   []*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			opts: []Option{
				WithSliceMergeByID(reflect.TypeOf([]*foo{}), "FieldInt"),
				WithSliceListAppendMerge(reflect.TypeOf([]*foo{}))}, // will prevail
			want: []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			name: "trilean nil nil",
			v1:   (*bool)(nil),
			v2:   (*bool)(nil),
			opts: []Option{WithTrileanMerge()},
			want: (*bool)(nil),
		},
		{
			name: "trilean nil false",
			v1:   (*bool)(nil),
			v2:   boolPtr(false),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(false),
		},
		{
			name: "trilean nil true",
			v1:   (*bool)(nil),
			v2:   boolPtr(true),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(true),
		},
		{
			name: "trilean false nil",
			v1:   boolPtr(false),
			v2:   (*bool)(nil),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(false),
		},
		{
			name: "trilean false false",
			v1:   boolPtr(false),
			v2:   boolPtr(false),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(false),
		},
		{
			name: "trilean false true",
			v1:   boolPtr(false),
			v2:   boolPtr(true),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(true),
		},
		{
			name: "trilean true nil",
			v1:   boolPtr(true),
			v2:   (*bool)(nil),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(true),
		},
		// with trileans: deepMerge(true, false) = false
		{
			name: "trilean true false",
			v1:   boolPtr(true),
			v2:   boolPtr(false),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(false),
		},
		// without trileans: deepMerge(true, false) = true
		{
			name: "trilean true false",
			v1:   boolPtr(true),
			v2:   boolPtr(false),
			want: boolPtr(true),
		},
		{
			name: "trilean true true",
			v1:   boolPtr(true),
			v2:   boolPtr(true),
			opts: []Option{WithTrileanMerge()},
			want: boolPtr(true),
		},
		{
			name: "type merger zero values",
			v1:   "",
			v2:   "",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringDeepMerge)},
			want: "ZERO!ZERO!",
		},
		{
			name: "type merger zero value 1",
			v1:   "abc",
			v2:   "",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringDeepMerge)},
			want: "abcZERO!",
		},
		{
			name: "type merger zero value 2",
			v1:   "",
			v2:   "def",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringDeepMerge)},
			want: "ZERO!def",
		},
		{
			name: "type merger non-zero values",
			v1:   "abc",
			v2:   "def",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringDeepMerge)},
			want: "abcdef",
		},
		{
			name: "type mismatch",
			v1:   1,
			v2:   "a",
			wantErr: func(t assert.TestingT, err error, args ...interface{}) bool {
				return assert.EqualError(t, err, "types do not match: int != string")
			},
		},
		{
			name:    "generic error",
			v1:      map[int]int{1: 1},
			v2:      map[int]int{1: 2},
			opts:    []Option{withMockDeepMergeError},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepMerge(tt.v1, tt.v2, tt.opts...)
			if err == nil {
				assert.Equal(t, tt.want, got)
				assertNotSame(t, tt.v1, got)
				assertNotSame(t, tt.v2, got)
			} else {
				assert.Nil(t, got)
			}
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMustDeepMerge(t *testing.T) {
	v1 := stringPtr("abc")
	v2 := stringPtr("def")
	merged := MustDeepMerge(v1, v2)
	assert.Equal(t, v2, merged)
	assert.NotSame(t, v1, merged)
	assert.NotSame(t, v2, merged)
	assert.PanicsWithError(t, "mock DeepMerge error", func() {
		MustDeepMerge("abc", "def", withMockDeepMergeError)
	})
}

func TestNewDeepMergeFunc(t *testing.T) {
	t.Run("with generic option", func(t *testing.T) {
		called := false
		opt := func(c *coalescer) {
			called = true
		}
		NewDeepMergeFunc(opt)
		assert.True(t, called)
	})
	t.Run("generic error", func(t *testing.T) {
		_, err := NewDeepMergeFunc(withMockDeepMergeError)(reflect.ValueOf(1), reflect.ValueOf("a"))
		assert.EqualError(t, err, "mock DeepMerge error")
	})
	t.Run("type mismatches", func(t *testing.T) {
		_, err := NewDeepMergeFunc()(reflect.ValueOf(1), reflect.ValueOf("a"))
		assert.EqualError(t, err, "types do not match: int != string")
		_, err = NewDeepMergeFunc()(reflect.ValueOf(map[string]int{"a": 2}), reflect.ValueOf(map[string]string{"a": "b"}))
		assert.EqualError(t, err, "types do not match: map[string]int != map[string]string")
		_, err = NewDeepMergeFunc()(reflect.ValueOf(intPtr(1)), reflect.ValueOf(stringPtr("a")))
		assert.EqualError(t, err, "types do not match: *int != *string")
		_, err = NewDeepMergeFunc()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = NewDeepMergeFunc()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = NewDeepMergeFunc()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		type foo struct {
			Int int
		}
		type bar struct {
			Int int
		}
		_, err = NewDeepMergeFunc()(reflect.ValueOf(foo{}), reflect.ValueOf(bar{}))
		assert.EqualError(t, err, "types do not match: goalesce.foo != goalesce.bar")
	})
}

func weirdStringDeepMerge(v1, v2 reflect.Value) (reflect.Value, error) {
	if v1.IsZero() {
		v1 = reflect.ValueOf("ZERO!")
	}
	if v2.IsZero() {
		v2 = reflect.ValueOf("ZERO!")
	}
	return reflect.ValueOf(v1.Interface().(string) + v2.Interface().(string)), nil
}
