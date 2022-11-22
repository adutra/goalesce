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
	type foo struct {
		FieldInt    int
		FieldIntPtr *int
		FieldString string
		FieldSlice  []int
		FieldMap    map[string]int
	}
	tests := []struct {
		name    string
		v       interface{}
		want    interface{}
		opts    []Option
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "untyped nil",
			v:    nil,
			want: nil,
		},
		{
			name: "typed nil",
			v:    (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "string",
			v:    "abc",
			want: "abc",
		},
		{
			name: "int",
			v:    123,
			want: 123,
		},
		{
			name: "*int",
			v:    intPtr(123),
			want: intPtr(123),
		},
		{
			name: "[]int",
			v:    []int{1, 2, 3},
			want: []int{1, 2, 3},
		},
		{
			name: "[]*int",
			v:    []*int{intPtr(1), intPtr(2), intPtr(3)},
			want: []*int{intPtr(1), intPtr(2), intPtr(3)},
		},
		{
			name: "[3]int",
			v:    [3]int{1, 2, 3},
			want: [3]int{1, 2, 3},
		},
		{
			name: "[3]*int",
			v:    [3]*int{intPtr(1), intPtr(2), intPtr(3)},
			want: [3]*int{intPtr(1), intPtr(2), intPtr(3)},
		},
		{
			name: "map[string]int",
			v:    map[string]int{"a": 1, "b": 2, "c": 3},
			want: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name: "map[string]*int",
			v:    map[string]*int{"a": intPtr(1), "b": intPtr(2), "c": intPtr(3)},
			want: map[string]*int{"a": intPtr(1), "b": intPtr(2), "c": intPtr(3)},
		},
		{
			name: "struct",
			v: &foo{
				FieldInt:    123,
				FieldIntPtr: intPtr(123),
				FieldString: "abc",
				FieldSlice:  []int{1, 2, 3},
				FieldMap:    map[string]int{"a": 1, "b": 2, "c": 3},
			},
			want: &foo{
				FieldInt:    123,
				FieldIntPtr: intPtr(123),
				FieldString: "abc",
				FieldSlice:  []int{1, 2, 3},
				FieldMap:    map[string]int{"a": 1, "b": 2, "c": 3},
			},
		},
		{
			name: "with type copier",
			v:    "abc",
			want: "def",
			opts: []Option{
				WithTypeCopier(reflect.TypeOf(""), func(v reflect.Value) (reflect.Value, error) {
					return reflect.ValueOf("def"), nil
				}),
			},
		},
		{
			name:    "error",
			v:       123,
			opts:    []Option{withMockDeepCopyError},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepCopy(tt.v, tt.opts...)
			if err == nil {
				assert.Equal(t, tt.want, got)
				assertNotSame(t, tt.v, got)
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

func TestMustDeepCopy(t *testing.T) {
	input := stringPtr("abc")
	copied := MustDeepCopy(input)
	assert.Equal(t, input, copied)
	assert.NotSame(t, input, copied)
	assert.PanicsWithError(t, "mock DeepCopy error", func() {
		MustDeepCopy("abc", withMockDeepCopyError)
	})
}

func TestNewDeepCopyFunc(t *testing.T) {
	t.Run("with generic option", func(t *testing.T) {
		called := false
		opt := func(c *coalescer) {
			called = true
		}
		NewDeepCopyFunc(opt)
		assert.True(t, called)
	})
	t.Run("generic error", func(t *testing.T) {
		_, err := NewDeepCopyFunc(withMockDeepCopyError)(reflect.ValueOf(1))
		assert.EqualError(t, err, "mock DeepCopy error")
	})
}
