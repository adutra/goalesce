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
	t.Run("untyped nil", func(t *testing.T) {
		var v1, v2 interface{}
		got, err := DeepMerge(v1, v2)
		assert.Nil(t, got)
		assert.NoError(t, err)
	})
	t.Run("untyped nil partial 1", func(t *testing.T) {
		var v1 interface{} = 1
		var v2 interface{}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, 1, got)
		assert.NoError(t, err)
	})
	t.Run("untyped nil partial 2", func(t *testing.T) {
		var v1 interface{}
		var v2 interface{} = 1
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, 1, got)
		assert.NoError(t, err)
	})
	t.Run("int", func(t *testing.T) {
		v1 := 1
		v2 := 2
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, 2, got)
		assert.NoError(t, err)
	})
	t.Run("int mixed zero 1", func(t *testing.T) {
		v1 := 0
		v2 := 1
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, 1, got)
		assert.NoError(t, err)
	})
	t.Run("int mixed zero 2", func(t *testing.T) {
		v1 := 1
		v2 := 0
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, 1, got)
		assert.NoError(t, err)
	})
	t.Run("*int", func(t *testing.T) {
		v1 := intPtr(1)
		v2 := intPtr(2)
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, intPtr(2), got)
		assertNotSame(t, v1, got)
		assertNotSame(t, v2, got)
		assert.NoError(t, err)
	})
	t.Run("*int zero", func(t *testing.T) {
		v1 := (*int)(nil)
		v2 := (*int)(nil)
		got, err := DeepMerge(v1, v2)
		assert.Nil(t, got)
		assert.NoError(t, err)
	})
	t.Run("*int mixed zero 1", func(t *testing.T) {
		v1 := (*int)(nil)
		v2 := intPtr(1)
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, intPtr(1), got)
		assertNotSame(t, v2, got)
		assert.NoError(t, err)
	})
	t.Run("*int mixed zero 2", func(t *testing.T) {
		v1 := intPtr(1)
		v2 := (*int)(nil)
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, intPtr(1), got)
		assertNotSame(t, v1, got)
		assert.NoError(t, err)
	})
	t.Run("string", func(t *testing.T) {
		v1 := "abc"
		v2 := "def"
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, "def", got)
		assert.NoError(t, err)
	})
	t.Run("string zero", func(t *testing.T) {
		v1 := "abc"
		v2 := ""
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, "abc", got)
		assert.NoError(t, err)
	})
	t.Run("bool", func(t *testing.T) {
		v1 := true
		v2 := false
		got, err := DeepMerge(v1, v2)
		assert.True(t, got)
		assert.NoError(t, err)
	})
	type foo struct {
		FieldInt int
	}
	t.Run("struct", func(t *testing.T) {
		v1 := foo{FieldInt: 1}
		v2 := foo{FieldInt: 2}
		want := foo{FieldInt: 2}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("struct zero", func(t *testing.T) {
		v1 := foo{}
		v2 := foo{FieldInt: 0}
		want := foo{}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("struct zero mixed 1", func(t *testing.T) {
		v1 := foo{FieldInt: 1}
		v2 := foo{FieldInt: 0}
		want := foo{FieldInt: 1}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("struct zero mixed 2", func(t *testing.T) {
		v1 := foo{FieldInt: 0}
		v2 := foo{FieldInt: 1}
		want := foo{FieldInt: 1}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	type bar struct {
		FieldInt int
		FieldFoo foo
	}
	t.Run("struct complex", func(t *testing.T) {
		v1 := bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}}
		v2 := bar{FieldInt: 1}
		want := bar{FieldInt: 1, FieldFoo: foo{FieldInt: 1}}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("struct complex atomic", func(t *testing.T) {
		v1 := bar{FieldInt: 0, FieldFoo: foo{FieldInt: 1}}
		v2 := bar{FieldInt: 1}
		want := bar{FieldInt: 1}
		got, err := DeepMerge(v1, v2, WithAtomicMerge(reflect.TypeOf(bar{})))
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("map[int]int", func(t *testing.T) {
		v1 := map[int]int{1: 1, 2: 2}
		v2 := map[int]int{1: 2, 3: 3}
		want := map[int]int{1: 2, 2: 2, 3: 3}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("map[int]foo", func(t *testing.T) {
		v1 := map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}}
		v2 := map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}}
		want := map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}, 3: {FieldInt: 3}}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("map[int]foo atomic", func(t *testing.T) {
		v1 := map[int]foo{1: {FieldInt: 1}, 3: {FieldInt: 3}}
		v2 := map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}}
		want := map[int]foo{1: {FieldInt: 2}, 2: {FieldInt: 2}}
		got, err := DeepMerge(v1, v2, WithAtomicMerge(reflect.TypeOf(map[int]foo{})))
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]int", func(t *testing.T) {
		v1 := []int{1, 2}
		v2 := []int{3, 4}
		want := []int{3, 4}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]foo", func(t *testing.T) {
		v1 := []foo{{FieldInt: 1}, {FieldInt: 2}}
		v2 := []foo{{FieldInt: 3}, {FieldInt: 4}}
		want := []foo{{FieldInt: 3}, {FieldInt: 4}}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[2]foo", func(t *testing.T) {
		v1 := [2]foo{{FieldInt: 1}, {FieldInt: 2}}
		v2 := [2]foo{{FieldInt: 3}, {FieldInt: 4}}
		want := [2]foo{{FieldInt: 3}, {FieldInt: 4}}
		got, err := DeepMerge(v1, v2)
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]int union", func(t *testing.T) {
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		want := []int{1, 2, 3}
		got, err := DeepMerge(v1, v2, WithDefaultSliceSetUnionMerge())
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]int append", func(t *testing.T) {
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		want := []int{1, 2, 2, 3}
		got, err := DeepMerge(v1, v2, WithDefaultSliceListAppendMerge())
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]foo custom", func(t *testing.T) {
		v1 := []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}
		v2 := []foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		want := []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		got, err := DeepMerge(v1, v2, WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt"))
		assert.Equal(t, want, got)
		assert.NoError(t, err)
	})
	t.Run("[]*int custom", func(t *testing.T) {
		v1 := []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}
		v2 := []*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		want := []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		got, err := DeepMerge(v1, v2, WithSliceMergeByKeyFunc(
			reflect.TypeOf([]*foo{}),
			func(_ int, v reflect.Value) (reflect.Value, error) {
				i := v.Interface()
				return reflect.ValueOf(i.(*foo).FieldInt), nil
			},
		))
		assert.Equal(t, want, got)
		assertNotSame(t, v1, got)
		assertNotSame(t, v2, got)
		assert.NoError(t, err)
	})
	t.Run("[]*int type merger conflict", func(t *testing.T) {
		v1 := []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}
		v2 := []*foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		want := []*foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}}
		got, err := DeepMerge(v1, v2,
			WithSliceMergeByID(reflect.TypeOf([]*foo{}), "FieldInt"),
			WithSliceListAppendMerge(reflect.TypeOf([]*foo{})), // will prevail
		)
		assert.Equal(t, want, got)
		assertNotSame(t, v1, got)
		assertNotSame(t, v2, got)
		assert.NoError(t, err)
	})
	t.Run("interface", func(t *testing.T) {
		var v1 Bird = &Duck{"Donald"}
		var v2 Bird = &Duck{"Scrooge"}
		called := false
		// reflect.TypeOf(v1) is *goalesce.Duck, not Bird
		got, err := DeepMerge(v1, v2, WithTypeMerger(reflect.TypeOf(v1), func(v1, v2 reflect.Value) (reflect.Value, error) {
			called = true
			return reflect.Value{}, nil
		}))
		assert.Equal(t, &Duck{"Scrooge"}, got)
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("interface type mismatch", func(t *testing.T) {
		var v1 Bird = &Duck{"Donald"}
		var v2 Bird = &Goose{"Scrooge"}
		got, err := DeepMerge(v1, v2)
		assert.Zero(t, got)
		assert.EqualError(t, err, "types do not match: *goalesce.Duck != *goalesce.Goose")
	})
	t.Run("interface pointer", func(t *testing.T) {
		var v1 Bird = &Duck{"Donald"}
		var v2 Bird = &Duck{"Scrooge"}
		called := false
		// reflect.TypeOf(&v1).Elem() is Bird, not *goalesce.Duck
		got, err := DeepMerge(&v1, &v2, WithTypeMerger(reflect.TypeOf(&v1).Elem(), func(v1, v2 reflect.Value) (reflect.Value, error) {
			called = true
			return reflect.Value{}, nil
		}))
		assert.Equal(t, &Duck{"Scrooge"}, *got)
		assert.NoError(t, err)
		assert.True(t, called)
	})
	trileanTests := []struct {
		name string
		v1   *bool
		v2   *bool
		opts []Option
		want *bool
	}{
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
	}
	for _, tt := range trileanTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepMerge(tt.v1, tt.v2, tt.opts...)
			assert.Equal(t, tt.want, got)
			assertNotSame(t, tt.v1, got)
			assertNotSame(t, tt.v2, got)
			assert.NoError(t, err)
		})
	}
	weirdStringMerger := func(v1, v2 reflect.Value) (reflect.Value, error) {
		if v1.IsZero() {
			v1 = reflect.ValueOf("ZERO!")
		}
		if v2.IsZero() {
			v2 = reflect.ValueOf("ZERO!")
		}
		return reflect.ValueOf(v1.Interface().(string) + v2.Interface().(string)), nil
	}
	typeMergerTests := []struct {
		name string
		v1   string
		v2   string
		opts []Option
		want string
	}{
		{
			name: "type merger zero values",
			v1:   "",
			v2:   "",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringMerger)},
			want: "ZERO!ZERO!",
		},
		{
			name: "type merger zero value 1",
			v1:   "abc",
			v2:   "",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringMerger)},
			want: "abcZERO!",
		},
		{
			name: "type merger zero value 2",
			v1:   "",
			v2:   "def",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringMerger)},
			want: "ZERO!def",
		},
		{
			name: "type merger non-zero values",
			v1:   "abc",
			v2:   "def",
			opts: []Option{WithTypeMerger(reflect.TypeOf(""), weirdStringMerger)},
			want: "abcdef",
		},
	}
	for _, tt := range typeMergerTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepMerge(tt.v1, tt.v2, tt.opts...)
			if err == nil {
				assert.Equal(t, tt.want, got)
				assertNotSame(t, tt.v1, got)
				assertNotSame(t, tt.v2, got)
			} else {
				assert.Nil(t, got)
			}
			assert.NoError(t, err)
		})
	}
	t.Run("generic error", func(t *testing.T) {
		got, err := DeepMerge("abc", "def", withMockDeepMergeError)
		assert.Equal(t, "", got)
		assert.EqualError(t, err, "mock DeepMerge error")
	})
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
