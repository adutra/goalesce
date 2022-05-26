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

func Test_defaultCoalescer_Coalesce(t *testing.T) {
	type foo struct {
		Int int
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		want interface{}
	}{
		{
			name: "int both zero",
			v1:   0,
			v2:   0,
			want: 0,
		},
		{
			name: "int v1 zero",
			v1:   0,
			v2:   1,
			want: 1,
		},
		{
			name: "int v2 zero",
			v1:   1,
			v2:   0,
			want: 1,
		},
		{
			name: "int none zero",
			v1:   1,
			v2:   2,
			want: 2,
		},
		{
			name: "*int both zero",
			v1:   (*int)(nil),
			v2:   (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "*int v1 zero",
			v1:   (*int)(nil),
			v2:   intPtr(0),
			want: intPtr(0),
		},
		{
			name: "*int v2 zero",
			v1:   intPtr(0),
			v2:   (*int)(nil),
			want: intPtr(0),
		},
		{
			name: "*int none zero",
			v1:   intPtr(1),
			v2:   intPtr(0),
			want: intPtr(0),
		},
		{
			name: "string both empty",
			v1:   "",
			v2:   "",
			want: "",
		},
		{
			name: "string v1 empty",
			v1:   "",
			v2:   "a",
			want: "a",
		},
		{
			name: "string v2 empty",
			v1:   "a",
			v2:   "",
			want: "a",
		},
		{
			name: "string none empty",
			v1:   "a",
			v2:   "b",
			want: "b",
		},
		{
			name: "*string both empty",
			v1:   (*string)(nil),
			v2:   (*string)(nil),
			want: (*string)(nil),
		},
		{
			name: "*string v1 empty",
			v1:   (*string)(nil),
			v2:   stringPtr(""),
			want: stringPtr(""),
		},
		{
			name: "*string v2 empty",
			v1:   stringPtr(""),
			v2:   (*string)(nil),
			want: stringPtr(""),
		},
		{
			name: "*string none empty",
			v1:   stringPtr("a"),
			v2:   stringPtr(""),
			want: stringPtr(""),
		},
		{
			name: "bool both false",
			v1:   false,
			v2:   false,
			want: false,
		},
		{
			name: "bool v1 false",
			v1:   false,
			v2:   true,
			want: true,
		},
		{
			name: "bool v2 false",
			v1:   true,
			v2:   false,
			want: true,
		},
		{
			name: "bool none false",
			v1:   true,
			v2:   true,
			want: true,
		},
		{
			name: "*bool both nil",
			v1:   (*bool)(nil),
			v2:   (*bool)(nil),
			want: (*bool)(nil),
		},
		{
			name: "*bool v1 nil",
			v1:   (*bool)(nil),
			v2:   boolPtr(false),
			want: boolPtr(false),
		},
		{
			name: "*bool v2 nil",
			v1:   boolPtr(false),
			v2:   (*bool)(nil),
			want: boolPtr(false),
		},
		{
			name: "*bool none nil",
			v1:   boolPtr(true),
			v2:   boolPtr(false),
			want: boolPtr(false), // trilean semantics
		},
		{
			name: "foo both zero",
			v1:   foo{},
			v2:   foo{},
			want: foo{},
		},
		{
			name: "foo v1 zero",
			v1:   foo{},
			v2:   foo{1},
			want: foo{1},
		},
		{
			name: "foo v2 zero",
			v1:   foo{1},
			v2:   foo{},
			want: foo{1},
		},
		{
			name: "foo none zero",
			v1:   foo{1},
			v2:   foo{2},
			want: foo{2},
		},
		{
			name: "*foo both zero",
			v1:   (*foo)(nil),
			v2:   (*foo)(nil),
			want: (*foo)(nil),
		},
		{
			name: "*foo v1 zero",
			v1:   (*foo)(nil),
			v2:   &foo{},
			want: &foo{},
		},
		{
			name: "*foo v2 zero",
			v1:   &foo{},
			v2:   (*foo)(nil),
			want: &foo{},
		},
		{
			name: "*foo none zero",
			v1:   &foo{1},
			v2:   &foo{},
			want: &foo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &atomicCoalescer{}
			got, err := c.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
}
