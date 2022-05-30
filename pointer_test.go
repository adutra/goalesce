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
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_mainCoalescer_coalescePointer(t *testing.T) {
	type foo struct {
		Int int
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
		name string
		v1   interface{}
		v2   interface{}
		want interface{}
	}{
		{
			"ints zeroes",
			(*int)(nil),
			(*int)(nil),
			(*int)(nil),
		},
		{
			"ints",
			intPtr(1),
			intPtr(2),
			intPtr(2),
		},
		{
			"structs zeroes",
			(*foo)(nil),
			(*foo)(nil),
			(*foo)(nil),
		},
		{
			"structs",
			&foo{Int: 1},
			&foo{Int: 2},
			&foo{Int: 2},
		},
		{
			"cycle 1",
			simpleCycle(),
			(*cycle)(nil),
			simpleCycle(), // unnoticed
		},
		{
			"cycle 2",
			(*cycle)(nil),
			simpleCycle(),
			simpleCycle(), // unnoticed
		},
		{
			"cycle 3",
			simpleCycle(),
			simpleCycle(),
			&cycle{Cycle: &cycle{}},
		},
		{
			"cycle 4",
			&cycle{Cycle: &cycle{Cycle: &cycle{}}},
			simpleCycle(),
			&cycle{Cycle: &cycle{Cycle: &cycle{}}},
		},
		{
			"cycle 5",
			simpleCycle(),
			&cycle{Cycle: &cycle{Cycle: &cycle{}}},
			&cycle{Cycle: &cycle{Cycle: &cycle{}}},
		},
		{
			"cycle 6",
			simpleCycle(),
			&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
			&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}},
		},
		{
			"cycle 7",
			complexCycle(),
			&cycle{Cycle: &cycle{}},
			complexCycle(), // unnoticed
		},
		{
			"cycle 8",
			complexCycle(),
			complexCycle(),
			&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}},
		},
		{
			"cycle 9",
			complexCycle(),
			&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}}},
			&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewCoalescer()
			got, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("cycle errors", func(t *testing.T) {
		withCycle := reflect.ValueOf(simpleCycle())
		withoutCycle := reflect.ValueOf(&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}})
		coalescer := NewCoalescer(WithErrorOnCycle())
		_, err := coalescer(withCycle, withoutCycle)
		assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
		coalescer = NewCoalescer(WithErrorOnCycle())
		_, err = coalescer(withoutCycle, withCycle)
		assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
	})
	t.Run("fallback error", func(t *testing.T) {
		coalescer := NewCoalescer(WithTypeCoalescer(reflect.TypeOf(0), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		_, err := coalescer(reflect.ValueOf(intPtr(1)), reflect.ValueOf(intPtr(2)))
		assert.EqualError(t, err, "fake")
	})
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
