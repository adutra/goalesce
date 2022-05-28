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

func Test_coalescer_deepMergeInterface(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	tests := []struct {
		name    string
		v1      interface{}
		v2      interface{}
		want    interface{}
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "zeroes",
			v1:   nil,
			v2:   nil,
			want: nil,
		},
		{
			name: "mixed zeroes 1",
			v1:   123,
			v2:   nil,
			want: 123,
		},
		{
			name: "mixed zeroes 2",
			v1:   nil,
			v2:   123,
			want: 123,
		},
		{
			name: "ints zeroes",
			v1:   0,
			v2:   0,
			want: 0,
		},
		{
			name: "ints",
			v1:   1,
			v2:   2,
			want: 2,
		},
		{
			name: "structs zeroes",
			v1:   foo{},
			v2:   foo{},
			want: foo{},
		},
		{
			name: "structs",
			v1:   foo{FieldInt: 1},
			v2:   foo{FieldInt: 0},
			want: foo{FieldInt: 1},
		},
		{
			name: "structs and empty structs",
			v1:   foo{FieldInt: 1},
			v2:   foo{},
			want: foo{FieldInt: 1},
		},
		{
			name: "maps",
			v1:   map[string]int{"a": 1},
			v2:   map[string]int{"b": 2},
			want: map[string]int{"a": 1, "b": 2},
		},
		{
			name: "*ints zeroes",
			v1:   (*int)(nil),
			v2:   (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "*ints",
			v1:   intPtr(1),
			v2:   intPtr(2),
			want: intPtr(2),
		},
		{
			name: "*structs zeroes",
			v1:   (*foo)(nil),
			v2:   (*foo)(nil),
			want: (*foo)(nil),
		},
		{
			name: "*structs",
			v1:   &foo{FieldInt: 1},
			v2:   &foo{FieldInt: 2},
			want: &foo{FieldInt: 2},
		},
		{
			name:    "different types",
			v1:      123,
			v2:      12.34,
			wantErr: assert.Error,
		},
		{
			name:    "generic error",
			v1:      123,
			v2:      456,
			wantErr: assert.Error,
			opts:    []Option{withMockDeepMergeError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeInterface(reflect.ValueOf(&tt.v1).Elem(), reflect.ValueOf(&tt.v2).Elem())
			if err == nil {
				assert.Equal(t, tt.want, got.Interface())
				assertNotSame(t, tt.v1, got.Interface())
				assertNotSame(t, tt.v2, got.Interface())
			} else {
				assert.False(t, got.IsValid())
			}
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_coalescer_deepCopyInterface(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{}
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "untyped nil",
			v:    nil,
		},
		{
			name: "typed nil",
			v:    (*int)(nil),
		},
		{
			name: "zero",
			v:    0,
		},
		{
			name: "non zero",
			v:    123,
		},
		{
			name:    "error",
			v:       reflect.ValueOf([3]int{1, 2, 3}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepCopyInterface(reflect.ValueOf(&tt.v).Elem())
			if err == nil {
				assert.Equal(t, tt.v, got.Interface())
				assertNotSame(t, tt.v, got.Interface())
			} else {
				assert.False(t, got.IsValid())
			}
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
