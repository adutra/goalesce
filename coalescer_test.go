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

func Test_coalescer_defaultDeepMerge(t *testing.T) {
	// Note: we don't need to test all the types and corner cases here, as the underlying merge
	// functions are thoroughly tested.
	type Foo struct {
		A int
	}
	tests := []struct {
		name    string
		v1      reflect.Value
		v2      reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "zero",
			v1:   reflect.ValueOf(0),
			v2:   reflect.ValueOf(0),
			want: reflect.ValueOf(0),
		},
		{
			name: "mixed zero 1",
			v1:   reflect.ValueOf(1),
			v2:   reflect.ValueOf(0),
			want: reflect.ValueOf(1),
		},
		{
			name: "mixed zero 2",
			v1:   reflect.ValueOf(0),
			v2:   reflect.ValueOf(1),
			want: reflect.ValueOf(1),
		},
		{
			name: "non zero atomic",
			v1:   reflect.ValueOf(1),
			v2:   reflect.ValueOf(2),
			want: reflect.ValueOf(2),
		},
		{
			name: "non zero pointer",
			v1:   reflect.ValueOf(intPtr(1)),
			v2:   reflect.ValueOf(intPtr(2)),
			want: reflect.ValueOf(intPtr(2)),
		},
		{
			name: "non zero interface",
			v1:   reflect.ValueOf(interfacePtr(1)).Elem(),
			v2:   reflect.ValueOf(interfacePtr(2)).Elem(),
			want: reflect.ValueOf(interfacePtr(2)).Elem(),
		},
		{
			name: "non zero map",
			v1:   reflect.ValueOf(map[string]int{"a": 1}),
			v2:   reflect.ValueOf(map[string]int{"b": 2}),
			want: reflect.ValueOf(map[string]int{"a": 1, "b": 2}),
		},
		{
			name: "non zero slice",
			v1:   reflect.ValueOf([]int{1}),
			v2:   reflect.ValueOf([]int{2}),
			want: reflect.ValueOf([]int{2}),
		},
		{
			name: "non zero array",
			v1:   reflect.ValueOf([1]int{1}),
			v2:   reflect.ValueOf([1]int{2}),
			want: reflect.ValueOf([1]int{2}),
		},
		{
			name: "non zero struct",
			v1:   reflect.ValueOf(Foo{1}),
			v2:   reflect.ValueOf(Foo{2}),
			want: reflect.ValueOf(Foo{2}),
		},
		{
			name: "type merger",
			v1:   reflect.ValueOf(Foo{1}),
			v2:   reflect.ValueOf(Foo{2}),
			want: reflect.ValueOf(Foo{3}),
			opts: []Option{WithTypeMerger(reflect.TypeOf(Foo{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
				return reflect.ValueOf(Foo{v1.Interface().(Foo).A + v2.Interface().(Foo).A}), nil
			})},
		},
		{
			name:    "generic error",
			v1:      reflect.ValueOf(1),
			v2:      reflect.ValueOf(2),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
		{
			name:    "type mismatch",
			v1:      reflect.ValueOf(123),
			v2:      reflect.ValueOf("abc"),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.defaultDeepMerge(tt.v1, tt.v2)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v1.Interface(), got.Interface())
				assertNotSame(t, tt.v2.Interface(), got.Interface())
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

func Test_coalescer_defaultDeepCopy(t *testing.T) {
	// Note: we don't need to test all the types and corner cases here, as the underlying copy
	// functions are thoroughly tested.
	type Foo struct {
		A int
	}
	tests := []struct {
		name    string
		v       reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "zero",
			v:    reflect.ValueOf(0),
			want: reflect.ValueOf(0),
		},
		{
			name: "non zero atomic",
			v:    reflect.ValueOf(1),
			want: reflect.ValueOf(1),
		},
		{
			name: "non zero pointer",
			v:    reflect.ValueOf(intPtr(1)),
			want: reflect.ValueOf(intPtr(1)),
		},
		{
			name: "non zero interface",
			v:    reflect.ValueOf(interfacePtr(1)).Elem(),
			want: reflect.ValueOf(interfacePtr(1)).Elem(),
		},
		{
			name: "non zero map",
			v:    reflect.ValueOf(map[string]int{"a": 1}),
			want: reflect.ValueOf(map[string]int{"a": 1}),
		},
		{
			name: "non zero slice",
			v:    reflect.ValueOf([]int{1}),
			want: reflect.ValueOf([]int{1}),
		},
		{
			name: "non zero array",
			v:    reflect.ValueOf([1]int{1}),
			want: reflect.ValueOf([1]int{1}),
		},
		{
			name: "non zero struct",
			v:    reflect.ValueOf(Foo{1}),
			want: reflect.ValueOf(Foo{1}),
		},
		{
			name: "type copier",
			v:    reflect.ValueOf(Foo{1}),
			want: reflect.ValueOf(Foo{2}),
			opts: []Option{WithTypeCopier(reflect.TypeOf(Foo{}), func(v reflect.Value) (reflect.Value, error) {
				return reflect.ValueOf(Foo{v.Interface().(Foo).A + 1}), nil
			})},
		},
		{
			name:    "generic error",
			v:       reflect.ValueOf(Foo{1}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.defaultDeepCopy(tt.v)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v.Interface(), got.Interface())
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
