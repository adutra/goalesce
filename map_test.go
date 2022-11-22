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

func Test_coalescer_deepMergeMap(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldIntPtr *int
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
			name: "map[string]int zero",
			v1:   (map[string]int)(nil),
			v2:   (map[string]int)(nil),
			want: (map[string]int)(nil),
		},
		{
			name: "map[string]int mixed zero",
			v1:   map[string]int{"a": 1},
			v2:   (map[string]int)(nil),
			want: map[string]int{"a": 1},
		},
		{
			name: "map[string]int mixed zero2",
			v1:   (map[string]int)(nil),
			v2:   map[string]int{"b": 2},
			want: map[string]int{"b": 2},
		},
		{
			name: "map[string]int empty",
			v1:   map[string]int{},
			v2:   map[string]int{},
			want: map[string]int{},
		},
		{
			name: "map[string]int mixed empty",
			v1:   map[string]int{"a": 1},
			v2:   map[string]int{},
			want: map[string]int{"a": 1},
		},
		{
			name: "map[string]int mixed empty 2",
			v1:   map[string]int{},
			v2:   map[string]int{"a": 2},
			want: map[string]int{"a": 2},
		},
		{
			name: "map[string]int non empty",
			v1:   map[string]int{"a": 1, "b": 1},
			v2:   map[string]int{"a": 2, "c": 2},
			want: map[string]int{"a": 2, "b": 1, "c": 2},
		},
		{
			name: "map[string]foo zero",
			v1:   (map[string]foo)(nil),
			v2:   (map[string]foo)(nil),
			want: (map[string]foo)(nil),
		},
		{
			name: "map[string]foo empty",
			v1:   map[string]foo{},
			v2:   map[string]foo{},
			want: map[string]foo{},
		},
		{
			name: "map[string]foo mixed empty",
			v1:   map[string]foo{"a": {1}},
			v2:   map[string]foo{},
			want: map[string]foo{"a": {1}},
		},
		{
			name: "map[string]foo mixed empty 2",
			v1:   map[string]foo{},
			v2:   map[string]foo{"b": {2}},
			want: map[string]foo{"b": {2}},
		},
		{
			name: "map[string]foo non empty",
			v1:   map[string]foo{"a": {1}, "b": {2}},
			v2:   map[string]foo{"a": {2}, "c": {2}},
			want: map[string]foo{"a": {2}, "b": {2}, "c": {2}},
		},
		{
			name: "map[string]*int zero",
			v1:   (map[string]*int)(nil),
			v2:   (map[string]*int)(nil),
			want: (map[string]*int)(nil),
		},
		{
			name: "map[string]*int mixed zero",
			v1:   map[string]*int{"a": nil},
			v2:   (map[string]*int)(nil),
			want: map[string]*int{"a": nil},
		},
		{
			name: "map[string]*int mixed zero2",
			v1:   (map[string]*int)(nil),
			v2:   map[string]*int{"b": nil},
			want: map[string]*int{"b": nil},
		},
		{
			name: "map[string]*int empty",
			v1:   map[string]*int{},
			v2:   map[string]*int{},
			want: map[string]*int{},
		},
		{
			name: "map[string]*int mixed empty",
			v1:   map[string]*int{"a": intPtr(1)},
			v2:   map[string]*int{},
			want: map[string]*int{"a": intPtr(1)},
		},
		{
			name: "map[string]*int mixed empty 2",
			v1:   map[string]*int{},
			v2:   map[string]*int{"a": intPtr(2)},
			want: map[string]*int{"a": intPtr(2)},
		},
		{
			name: "map[string]*int mixed empty 3",
			v1:   map[string]*int{"a": intPtr(1)},
			v2:   map[string]*int{"a": nil},
			want: map[string]*int{"a": intPtr(1)},
		},
		{
			name: "map[string]*int mixed empty 4",
			v1:   map[string]*int{"a": nil},
			v2:   map[string]*int{"a": intPtr(2)},
			want: map[string]*int{"a": intPtr(2)},
		},
		{
			name: "map[string]*int non empty",
			v1:   map[string]*int{"a": intPtr(1), "b": intPtr(1)},
			v2:   map[string]*int{"a": intPtr(2), "c": nil},
			want: map[string]*int{"a": intPtr(2), "b": intPtr(1), "c": nil},
		},
		{
			name: "map[string][]int empty",
			v1:   map[string][]int{},
			v2:   map[string][]int{},
			want: map[string][]int{},
		},
		{
			name: "map[string][]int mixed empty",
			v1:   map[string][]int{"a": {1}},
			v2:   map[string][]int{},
			want: map[string][]int{"a": {1}},
		},
		{
			name: "map[string][]int mixed empty 2",
			v1:   map[string][]int{},
			v2:   map[string][]int{"a": {2}},
			want: map[string][]int{"a": {2}},
		},
		{
			name: "map[string][]int mixed empty 3",
			v1:   map[string][]int{"a": {1}},
			v2:   map[string][]int{"a": nil},
			want: map[string][]int{"a": {1}},
		},
		{
			name: "map[string][]int mixed empty 4",
			v1:   map[string][]int{"a": nil},
			v2:   map[string][]int{"a": {2}},
			want: map[string][]int{"a": {2}},
		},
		{
			name: "map[string][]int non empty",
			v1:   map[string][]int{"a": {1}, "b": {1}},
			v2:   map[string][]int{"a": {1, 2}, "c": nil},
			want: map[string][]int{"a": {1, 2}, "b": {1}, "c": nil},
		},
		{
			name: "map[string]*bar empty",
			v1:   map[string]*bar{},
			v2:   map[string]*bar{},
			want: map[string]*bar{},
		},
		{
			name: "map[string]*bar mixed empty",
			v1:   map[string]*bar{"a": {intPtr(1)}},
			v2:   map[string]*bar{},
			want: map[string]*bar{"a": {intPtr(1)}},
		},
		{
			name: "map[string]*bar mixed empty 2",
			v1:   map[string]*bar{},
			v2:   map[string]*bar{"b": {intPtr(2)}},
			want: map[string]*bar{"b": {intPtr(2)}},
		},
		{
			name: "map[string]*bar nested nils",
			v1:   map[string]*bar{"a": {nil}, "b": nil},
			v2:   map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			want: map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
		},
		{
			name: "map[string]*bar nested nils 2",
			v1:   map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			v2:   map[string]*bar{"a": {nil}, "b": nil},
			want: map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
		},
		{
			name: "map[string]*bar non empty",
			v1:   map[string]*bar{"a": {intPtr(1)}, "b": {intPtr(2)}},
			v2:   map[string]*bar{"a": {intPtr(2)}, "c": {intPtr(2)}},
			want: map[string]*bar{"a": {intPtr(2)}, "b": {intPtr(2)}, "c": {intPtr(2)}},
		},
		{
			name: "map[string][]*int empty",
			v1:   map[string][]*int{},
			v2:   map[string][]*int{},
			want: map[string][]*int{},
		},
		{
			name: "map[string][]*int mixed empty",
			v1:   map[string][]*int{"a": {intPtr(1)}},
			v2:   map[string][]*int{},
			want: map[string][]*int{"a": {intPtr(1)}},
		},
		{
			name: "map[string][]*int mixed empty 2",
			v1:   map[string][]*int{},
			v2:   map[string][]*int{"a": {intPtr(2)}},
			want: map[string][]*int{"a": {intPtr(2)}},
		},
		{
			name: "map[string][]*int mixed empty 3",
			v1:   map[string][]*int{"a": {intPtr(1)}},
			v2:   map[string][]*int{"a": nil},
			want: map[string][]*int{"a": {intPtr(1)}},
		},
		{
			name: "map[string][]*int mixed empty 4",
			v1:   map[string][]*int{"a": nil},
			v2:   map[string][]*int{"a": {intPtr(2)}},
			want: map[string][]*int{"a": {intPtr(2)}},
		},
		{
			name: "map[string][]*int non empty",
			v1:   map[string][]*int{"a": {intPtr(1)}, "b": {intPtr(1)}},
			v2:   map[string][]*int{"a": {intPtr(1), intPtr(2)}, "c": nil},
			want: map[string][]*int{"a": {intPtr(1), intPtr(2)}, "b": {intPtr(1)}, "c": nil},
		},
		{
			name: "map[string]interface{} empty",
			v1:   map[string]interface{}{},
			v2:   map[string]interface{}{},
			want: map[string]interface{}{},
		},
		{
			name: "map[string]interface{} mixed empty",
			v1:   map[string]interface{}{"a": &bar{intPtr(1)}},
			v2:   map[string]interface{}{},
			want: map[string]interface{}{"a": &bar{intPtr(1)}},
		},
		{
			name: "map[string]interface{} mixed empty 2",
			v1:   map[string]interface{}{},
			v2:   map[string]interface{}{"b": &bar{intPtr(2)}},
			want: map[string]interface{}{"b": &bar{intPtr(2)}},
		},
		{
			name: "map[string]interface{} nested nils",
			v1:   map[string]interface{}{"a": &bar{nil}, "b": nil},
			v2:   map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			want: map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
		},
		{
			name: "map[string]interface{} nested nils 2",
			v1:   map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			v2:   map[string]interface{}{"a": &bar{nil}, "b": nil},
			want: map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
		},
		{
			name: "map[string]interface{} non empty",
			v1:   map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
			v2:   map[string]interface{}{"a": &bar{intPtr(2)}, "c": &bar{intPtr(2)}},
			want: map[string]interface{}{"a": &bar{intPtr(2)}, "b": &bar{intPtr(2)}, "c": &bar{intPtr(2)}},
		},
		{
			name:    "generic error copy key 1",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"b": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen("a")},
		},
		{
			name:    "generic error copy value 1",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"b": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(1)},
		},
		{
			name:    "generic error copy key 1",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"b": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen("b")},
		},
		{
			name:    "generic error copy value 1",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"b": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(2)},
		},
		{
			name:    "generic error copy common entry key",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"a": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen("a")},
		},
		{
			name:    "generic error merge common entry value",
			v1:      map[string]int{"a": 1},
			v2:      map[string]int{"a": 2},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepMergeErrorWhen(1, 2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeMap(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
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

func Test_coalescer_deepCopyMap(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldIntPtr *int
	}
	tests := []struct {
		name    string
		v       interface{}
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "map[string]int zero",
			v:    (map[string]int)(nil),
		},
		{
			name: "map[string]int empty",
			v:    map[string]int{},
		},
		{
			name: "map[string]int non empty",
			v:    map[string]int{"a": 1, "b": 1},
		},
		{
			name: "map[string]foo zero",
			v:    (map[string]foo)(nil),
		},
		{
			name: "map[string]foo empty",
			v:    map[string]foo{},
		},
		{
			name: "map[string]foo non empty",
			v:    map[string]foo{"a": {1}, "b": {2}},
		},
		{
			name: "map[string]*int zero",
			v:    (map[string]*int)(nil),
		},
		{
			name: "map[string]*int empty",
			v:    map[string]*int{},
		},
		{
			name: "map[string]*int non empty",
			v:    map[string]*int{"a": intPtr(1), "b": intPtr(1)},
		},
		{
			name: "map[string][]int empty",
			v:    map[string][]int{},
		},
		{
			name: "map[string][]int non empty",
			v:    map[string][]int{"a": {1}, "b": {1}},
		},
		{
			name: "map[string][]*int empty",
			v:    map[string][]*int{},
		},
		{
			name: "map[string][]*int non empty",
			v:    map[string][]*int{"a": {intPtr(1)}, "b": {intPtr(1)}},
		},
		{
			name: "map[string]interface{} empty",
			v:    map[string]interface{}{},
		},
		{
			name: "map[string]interface{} non empty",
			v:    map[string]interface{}{"a": &bar{intPtr(1)}, "b": &bar{intPtr(2)}},
		},
		{
			name:    "generic error copy key",
			v:       map[string]int{"a": 1},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen("a")},
		},
		{
			name:    "generic error copy value",
			v:       map[string]int{"a": 1},
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(1)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepCopyMap(reflect.ValueOf(tt.v))
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
