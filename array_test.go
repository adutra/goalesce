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

func Test_coalescer_deepMergeArray(t *testing.T) {
	tests := []struct {
		name    string
		v1      reflect.Value
		v2      reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "empty",
			v1:   reflect.ValueOf([0]int{}),
			v2:   reflect.ValueOf([0]int{}),
			want: reflect.ValueOf([0]int{}),
		},
		{
			name: "zero",
			v1:   reflect.ValueOf([3]int{}),
			v2:   reflect.ValueOf([3]int{}),
			want: reflect.ValueOf([3]int{0, 0, 0}),
		},
		{
			name: "non zero",
			v1:   reflect.ValueOf([3]int{1, 2, 3}),
			v2:   reflect.ValueOf([3]int{4, 5, 6}),
			want: reflect.ValueOf([3]int{4, 5, 6}),
		},
		{
			name: "non zero, zero elements",
			v1:   reflect.ValueOf([3]int{1, 2, 0}),
			v2:   reflect.ValueOf([3]int{0, 0, 3}),
			want: reflect.ValueOf([3]int{0, 0, 3}),
		},
		{
			name: "non zero pointers",
			v1:   reflect.ValueOf([3]*int{intPtr(1), intPtr(2), nil}),
			v2:   reflect.ValueOf([3]*int{nil, nil, intPtr(3)}),
			want: reflect.ValueOf([3]*int{nil, nil, intPtr(3)}),
		},
		{
			name: "array merge by index",
			v1:   reflect.ValueOf([3]*int{intPtr(1), intPtr(2), nil}),
			v2:   reflect.ValueOf([3]*int{nil, nil, intPtr(3)}),
			want: reflect.ValueOf([3]*int{intPtr(1), intPtr(2), intPtr(3)}),
			opts: []Option{WithDefaultArrayMergeByIndex()},
		},
		{
			name: "array merge by index 2",
			v1:   reflect.ValueOf([3]*int{intPtr(1), intPtr(2), nil}),
			v2:   reflect.ValueOf([3]*int{nil, nil, intPtr(3)}),
			want: reflect.ValueOf([3]*int{intPtr(1), intPtr(2), intPtr(3)}),
			opts: []Option{WithArrayMergeByIndex(reflect.TypeOf([3]*int{}))},
		},
		{
			name:    "generic error",
			v1:      reflect.ValueOf([3]int{1, 2, 3}),
			v2:      reflect.ValueOf([3]int{4, 5, 6}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeArray(tt.v1, tt.v2)
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

func Test_coalescer_deepMergeArrayByIndex(t *testing.T) {
	tests := []struct {
		name    string
		v1      reflect.Value
		v2      reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "empty",
			v1:   reflect.ValueOf([0]int{}),
			v2:   reflect.ValueOf([0]int{}),
			want: reflect.ValueOf([0]int{}),
		},
		{
			name: "zero",
			v1:   reflect.ValueOf([3]int{}),
			v2:   reflect.ValueOf([3]int{}),
			want: reflect.ValueOf([3]int{0, 0, 0}),
		},
		{
			name: "non zero",
			v1:   reflect.ValueOf([3]int{1, 2, 3}),
			v2:   reflect.ValueOf([3]int{4, 5, 6}),
			want: reflect.ValueOf([3]int{4, 5, 6}),
		},
		{
			name: "non zero, zero elements",
			v1:   reflect.ValueOf([3]int{1, 2, 0}),
			v2:   reflect.ValueOf([3]int{0, 0, 3}),
			want: reflect.ValueOf([3]int{1, 2, 3}),
		},
		{
			name: "non zero pointers",
			v1:   reflect.ValueOf([3]*int{intPtr(1), intPtr(2), nil}),
			v2:   reflect.ValueOf([3]*int{nil, nil, intPtr(3)}),
			want: reflect.ValueOf([3]*int{intPtr(1), intPtr(2), intPtr(3)}),
		},
		{
			name:    "generic error",
			v1:      reflect.ValueOf([3]int{1, 2, 3}),
			v2:      reflect.ValueOf([3]int{4, 5, 6}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepMergeError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeArrayByIndex(tt.v1, tt.v2)
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

func Test_coalescer_deepCopyArray(t *testing.T) {
	tests := []struct {
		name    string
		v       reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "empty",
			v:    reflect.ValueOf([0]int{}),
		},
		{
			name: "zero",
			v:    reflect.ValueOf([3]int{}),
		},
		{
			name: "non zero",
			v:    reflect.ValueOf([3]int{1, 2, 3}),
		},
		{
			name: "non zero pointers",
			v:    reflect.ValueOf([3]*int{intPtr(1), intPtr(2), nil}),
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
			got, err := c.deepCopyArray(tt.v)
			if err == nil {
				assert.Equal(t, tt.v.Interface(), got.Interface())
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
