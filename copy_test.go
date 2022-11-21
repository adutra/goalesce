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

func TestDeepCopy(t *testing.T) {
	// Note: we don't need to test all the types and corner cases here, as the underlying copy
	// functions are thoroughly tested.
	t.Run("untyped nil", func(t *testing.T) {
		var x interface{}
		got, err := DeepCopy(x)
		assert.Nil(t, got)
		assert.NoError(t, err)
	})
	t.Run("typed nil", func(t *testing.T) {
		var x *int
		got, err := DeepCopy(x)
		assert.Nil(t, got)
		assert.NoError(t, err)
	})
	t.Run("int", func(t *testing.T) {
		x := 123
		got, err := DeepCopy(x)
		assert.Equal(t, 123, got)
		assert.NoError(t, err)
	})
	t.Run("string", func(t *testing.T) {
		x := "abc"
		got, err := DeepCopy(x)
		assert.Equal(t, "abc", got)
		assert.NoError(t, err)
	})
	t.Run("[]int", func(t *testing.T) {
		x := []int{1, 2, 3}
		got, err := DeepCopy(x)
		assert.Equal(t, []int{1, 2, 3}, got)
		assert.NoError(t, err)
	})
	t.Run("[]*int", func(t *testing.T) {
		x := []*int{intPtr(1), intPtr(2), intPtr(3)}
		got, err := DeepCopy(x)
		assert.Equal(t, []*int{intPtr(1), intPtr(2), intPtr(3)}, got)
		assertNotSame(t, x, got)
		assert.NoError(t, err)
	})
	t.Run("[3]int", func(t *testing.T) {
		x := [3]int{1, 2, 3}
		got, err := DeepCopy(x)
		assert.Equal(t, [3]int{1, 2, 3}, got)
		assert.NoError(t, err)
	})
	t.Run("[3]*int", func(t *testing.T) {
		x := [3]*int{intPtr(1), intPtr(2), intPtr(3)}
		got, err := DeepCopy(x)
		assert.Equal(t, [3]*int{intPtr(1), intPtr(2), intPtr(3)}, got)
		assertNotSame(t, x, got)
		assert.NoError(t, err)
	})
	t.Run("map[string]int", func(t *testing.T) {
		x := map[string]int{"a": 1, "b": 2, "c": 3}
		got, err := DeepCopy(x)
		assert.Equal(t, map[string]int{"a": 1, "b": 2, "c": 3}, got)
		assert.NoError(t, err)
	})
	t.Run("map[string]*int", func(t *testing.T) {
		x := map[string]*int{"a": intPtr(1), "b": intPtr(2), "c": intPtr(3)}
		got, err := DeepCopy(x)
		assert.Equal(t, map[string]*int{"a": intPtr(1), "b": intPtr(2), "c": intPtr(3)}, got)
		assertNotSame(t, x, got)
		assert.NoError(t, err)
	})
	t.Run("struct", func(t *testing.T) {
		type foo struct {
			FieldInt    int
			FieldIntPtr *int
			FieldString string
			FieldSlice  []int
			FieldMap    map[string]int
		}
		x := &foo{
			FieldInt:    123,
			FieldIntPtr: intPtr(123),
			FieldString: "abc",
			FieldSlice:  []int{1, 2, 3},
			FieldMap:    map[string]int{"a": 1, "b": 2, "c": 3},
		}
		got, err := DeepCopy(x)
		assert.Equal(t, &foo{
			FieldInt:    123,
			FieldIntPtr: intPtr(123),
			FieldString: "abc",
			FieldSlice:  []int{1, 2, 3},
			FieldMap:    map[string]int{"a": 1, "b": 2, "c": 3},
		}, got)
		assertNotSame(t, x, got)
		assert.NoError(t, err)
	})
	t.Run("with type copier", func(t *testing.T) {
		got, err := DeepCopy("abc", WithTypeCopier(reflect.TypeOf(""), func(v reflect.Value) (reflect.Value, error) {
			return reflect.ValueOf("def"), nil
		}))
		assert.Equal(t, "def", got)
		assert.NoError(t, err)
	})
	t.Run("type mismatch", func(t *testing.T) {
		got, err := DeepCopy("abc", WithTypeCopier(reflect.TypeOf(""), func(v reflect.Value) (reflect.Value, error) {
			return reflect.ValueOf(123), nil
		}))
		assert.Equal(t, "", got)
		assert.EqualError(t, err, "cannot convert int to string")
	})
	t.Run("generic error", func(t *testing.T) {
		got, err := DeepCopy("abc", withMockDeepCopyError)
		assert.Equal(t, "", got)
		assert.EqualError(t, err, "mock DeepCopy error")
	})
}

func TestMustDeepCopy(t *testing.T) {
	input := stringPtr("abc")
	copied := MustDeepCopy(input)
	assert.Equal(t, input, copied)
	assert.NotSame(t, input, copied)
	assert.PanicsWithError(t, "mock DeepCopy error", func() {
		MustDeepCopy("abc", withMockDeepCopyError)
	})
}
