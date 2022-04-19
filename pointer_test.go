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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestNewPointerCoalescer(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		got := NewPointerCoalescer()
		assert.Equal(t, &pointerCoalescer{fallback: &defaultCoalescer{}}, got)
	})
	t.Run("with generic option", func(t *testing.T) {
		var passed *pointerCoalescer
		opt := func(c *pointerCoalescer) {
			passed = c
		}
		returned := NewPointerCoalescer(opt)
		assert.Equal(t, &pointerCoalescer{fallback: &defaultCoalescer{}}, returned)
		assert.Equal(t, returned, passed)
	})
	t.Run("with error on cycle", func(t *testing.T) {
		expected := &pointerCoalescer{
			onCycleReturnError: true,
			fallback:           &defaultCoalescer{},
		}
		actual := NewPointerCoalescer(WithOnCycleReturnError())
		assert.Equal(t, expected, actual)
	})
}

func Test_pointerCoalescer_Coalesce(t *testing.T) {
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
			coalescer := NewPointerCoalescer()
			coalescer.WithFallback(NewMainCoalescer())
			got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("type errors", func(t *testing.T) {
		_, err := NewPointerCoalescer().Coalesce(reflect.ValueOf(intPtr(1)), reflect.ValueOf(stringPtr("a")))
		assert.EqualError(t, err, "types do not match: *int != *string")
		_, err = NewPointerCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected ptr, got int")
	})
	t.Run("cycle errors", func(t *testing.T) {
		withCycle := reflect.ValueOf(simpleCycle())
		withoutCycle := reflect.ValueOf(&cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{Cycle: &cycle{}}}}})
		coalescer := NewPointerCoalescer(WithOnCycleReturnError())
		coalescer.WithFallback(NewMainCoalescer(WithPointerCoalescer(coalescer)))
		_, err := coalescer.Coalesce(withCycle, withoutCycle)
		assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
		coalescer = NewPointerCoalescer(WithOnCycleReturnError())
		coalescer.WithFallback(NewMainCoalescer(WithPointerCoalescer(coalescer)))
		_, err = coalescer.Coalesce(withoutCycle, withCycle)
		assert.EqualError(t, err, "*goalesce.cycle: cycle detected")
	})
	t.Run("fallback error", func(t *testing.T) {
		m := new(mockCoalescer)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		m.Test(t)
		coalescer := NewPointerCoalescer()
		coalescer.WithFallback(m)
		_, err := coalescer.Coalesce(reflect.ValueOf(intPtr(1)), reflect.ValueOf(intPtr(2)))
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
