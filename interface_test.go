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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewInterfaceCoalescer(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		got := NewInterfaceCoalescer()
		assert.Equal(t, &interfaceCoalescer{fallback: &atomicCoalescer{}}, got)
	})
	t.Run("with generic option", func(t *testing.T) {
		var passed *interfaceCoalescer
		opt := func(c *interfaceCoalescer) {
			passed = c
		}
		returned := NewInterfaceCoalescer(opt)
		assert.Equal(t, &interfaceCoalescer{fallback: &atomicCoalescer{}}, returned)
		assert.Equal(t, returned, passed)
	})
}

func Test_interfaceCoalescer_Coalesce(t *testing.T) {
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
			"zeroes",
			nil,
			nil,
			nil,
		},
		{
			"mixed zeroes 1",
			123,
			nil,
			123,
		},
		{
			"mixed zeroes 2",
			nil,
			123,
			123,
		},
		{
			"ints zeroes",
			0,
			0,
			0,
		},
		{
			"ints",
			1,
			2,
			2,
		},
		{
			"structs zeroes",
			foo{},
			foo{},
			foo{},
		},
		{
			"structs",
			foo{Int: 1},
			foo{Int: 0},
			foo{Int: 1},
		},
		{
			"maps",
			map[string]int{"a": 1},
			map[string]int{"b": 2},
			map[string]int{"a": 1, "b": 2},
		},
		{
			"*ints zeroes",
			(*int)(nil),
			(*int)(nil),
			(*int)(nil),
		},
		{
			"*ints",
			intPtr(1),
			intPtr(2),
			intPtr(2),
		},
		{
			"*structs zeroes",
			(*foo)(nil),
			(*foo)(nil),
			(*foo)(nil),
		},
		{
			"*structs",
			&foo{Int: 1},
			&foo{Int: 2},
			&foo{Int: 2},
		},
		{
			"different types",
			123,
			12.34,
			12.34,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewInterfaceCoalescer()
			coalescer.WithFallback(NewMainCoalescer())
			got, err := coalescer.Coalesce(reflect.ValueOf(&tt.v1).Elem(), reflect.ValueOf(&tt.v2).Elem())
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("type errors", func(t *testing.T) {
		_, err := NewInterfaceCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf("a"))
		assert.EqualError(t, err, "types do not match: int != string")
		_, err = NewInterfaceCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected interface, got int")
	})
	t.Run("fallback error", func(t *testing.T) {
		m := newMockCoalescer(t)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		coalescer := NewInterfaceCoalescer()
		coalescer.WithFallback(m)
		var v1 interface{} = 1
		var v2 interface{} = 2
		_, err := coalescer.Coalesce(reflect.ValueOf(&v1).Elem(), reflect.ValueOf(&v2).Elem())
		assert.EqualError(t, err, "fake")
	})
}
