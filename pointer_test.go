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

func Test_coalescer_deepMergePointer(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type cycle struct {
		Cycle *cycle
	}
	simpleCycle := func() *cycle {
		c := &cycle{}
		c.Cycle = c
		return c
	}
	complexCycle := func() *cycle {
		c1 := &cycle{}
		c2 := &cycle{}
		c3 := &cycle{}
		c1.Cycle = c2
		c2.Cycle = c3
		c3.Cycle = c1
		return c1
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
			name: "ints zeroes",
			v1:   (*int)(nil),
			v2:   (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "ints",
			v1:   intPtr(1),
			v2:   intPtr(2),
			want: intPtr(2),
		},
		{
			name: "structs zeroes",
			v1:   (*foo)(nil),
			v2:   (*foo)(nil),
			want: (*foo)(nil),
		},
		{
			name: "structs",
			v1:   &foo{FieldInt: 1},
			v2:   &foo{FieldInt: 2},
			want: &foo{FieldInt: 2},
		},
		{
			name: "cycle 1",
			v1:   simpleCycle(),
			v2:   (*cycle)(nil),
			want: &cycle{Cycle: &cycle{}},
		},
		{
			name: "cycle 2",
			v1:   (*cycle)(nil),
			v2:   simpleCycle(),
			want: &cycle{Cycle: &cycle{}},
		},
		{
			name: "cycle 3",
			v1:   simpleCycle(),
			v2:   simpleCycle(),
			want: &cycle{Cycle: &cycle{}},
		},
		{
			name: "cycle 4",
			v1:   &cycle{Cycle: &cycle{Cycle: &cycle{}}},
			v2:   simpleCycle(),
			want: &cycle{Cycle: &cycle{Cycle: &cycle{}}},
		},
		{
			name: "cycle 5",
			v1:   simpleCycle(),
			v2:   &cycle{Cycle: &cycle{Cycle: &cycle{}}},
			want: &cycle{Cycle: &cycle{Cycle: &cycle{}}},
		},
		{
			name: "cycle 6",
			v1:   simpleCycle(),
			v2:   &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
			want: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
		},
		{
			name: "cycle 7",
			v1:   complexCycle(),
			v2:   &cycle{Cycle: &cycle{}},
			want: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}},
		},
		{
			name: "cycle 8",
			v1:   complexCycle(),
			v2:   complexCycle(),
			want: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}},
		},
		{
			name: "cycle 9",
			v1:   complexCycle(),
			v2:   &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}}},
			want: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}}},
		},
		{
			name: "cycle error 1",
			v1:   simpleCycle(),
			v2:   &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
			},
			opts: []Option{WithErrorOnCycle()},
		},
		{
			name: "cycle error 2",
			v1:   &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
			v2:   simpleCycle(),
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
			},
			opts: []Option{WithErrorOnCycle()},
		},
		{
			name:    "generic error",
			v1:      intPtr(1),
			v2:      intPtr(2),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepMergeError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergePointer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
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

func Test_coalescer_deepCopyPointer(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type cycle struct {
		Cycle *cycle
	}
	simpleCycle := func() *cycle {
		c := &cycle{}
		c.Cycle = c
		return c
	}
	complexCycle := func() *cycle {
		c1 := &cycle{}
		c2 := &cycle{}
		c3 := &cycle{}
		c1.Cycle = c2
		c2.Cycle = c3
		c3.Cycle = c1
		return c1
	}
	tests := []struct {
		name    string
		v       interface{}
		want    interface{}
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "ints zeroes",
			v:    (*int)(nil),
			want: (*int)(nil),
		},
		{
			name: "ints",
			v:    intPtr(1),
			want: intPtr(1),
		},
		{
			name: "structs zeroes",
			v:    (*foo)(nil),
			want: (*foo)(nil),
		},
		{
			name: "structs",
			v:    &foo{FieldInt: 1},
			want: &foo{FieldInt: 1},
		},
		{
			name: "cycle simple",
			v:    simpleCycle(),
			want: &cycle{Cycle: &cycle{}},
		},
		{
			name: "cycle complex",
			v:    complexCycle(),
			want: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}},
		},
		{
			name: "cycle error simple",
			v:    simpleCycle(),
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
			},
			opts: []Option{WithErrorOnCycle()},
		},
		{
			name: "cycle error complex",
			v:    complexCycle(),
			wantErr: func(t assert.TestingT, err error, _ ...interface{}) bool {
				return assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
			},
			opts: []Option{WithErrorOnCycle()},
		},
		{
			name:    "generic error",
			v:       intPtr(1),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepCopyPointer(reflect.ValueOf(tt.v))
			if err == nil {
				assert.Equal(t, tt.want, got.Interface())
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

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func interfacePtr(i interface{}) *interface{} {
	return &i
}
