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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestNewMainCoalescer(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		expected := &mainCoalescer{}
		expected.defaultCoalescer = &atomicCoalescer{}
		expected.interfaceCoalescer = &interfaceCoalescer{fallback: expected}
		expected.pointerCoalescer = &pointerCoalescer{fallback: expected}
		expected.mapCoalescer = &mapCoalescer{fallback: expected}
		expected.structCoalescer = &structCoalescer{fallback: expected}
		expected.sliceCoalescer = &sliceCoalescer{
			defaultCoalescer: &atomicCoalescer{},
		}
		actual := NewMainCoalescer()
		assert.Equal(t, expected, actual)
	})
	t.Run("with generic option", func(t *testing.T) {
		var passed *mainCoalescer
		opt := func(c *mainCoalescer) {
			passed = c
		}
		returned := NewMainCoalescer(opt)
		assert.Equal(t, passed, returned)
	})
	t.Run("with type coalescer", func(t *testing.T) {
		type foo struct {
			Int int
		}
		m := newMockCoalescer(t)
		m.On("WithFallback", mock.Anything).Return()
		expected := &mainCoalescer{}
		expected.defaultCoalescer = &atomicCoalescer{}
		expected.interfaceCoalescer = &interfaceCoalescer{fallback: expected}
		expected.pointerCoalescer = &pointerCoalescer{fallback: expected}
		expected.mapCoalescer = &mapCoalescer{fallback: expected}
		expected.structCoalescer = &structCoalescer{fallback: expected}
		expected.sliceCoalescer = &sliceCoalescer{
			defaultCoalescer: &atomicCoalescer{},
		}
		expected.typeCoalescers = map[reflect.Type]Coalescer{
			reflect.TypeOf(foo{}): m,
		}
		actual := NewMainCoalescer(WithTypeCoalescer(reflect.TypeOf(foo{}), m))
		assert.Equal(t, expected, actual)
		actual.WithFallback(actual)
		m.AssertCalled(t, "WithFallback", actual)
	})
}

func TestWithAtomicType(t *testing.T) {
	c := &mainCoalescer{}
	WithAtomicType(reflect.TypeOf(0))(c)
	expected := map[reflect.Type]Coalescer{reflect.TypeOf(0): &atomicCoalescer{}}
	assert.Equal(t, expected, c.typeCoalescers)
}

func TestWithTrileans(t *testing.T) {
	c := &mainCoalescer{}
	WithTrileans()(c)
	expected := map[reflect.Type]Coalescer{reflect.PtrTo(reflect.TypeOf(false)): &atomicCoalescer{}}
	assert.Equal(t, expected, c.typeCoalescers)
}

func TestWithTypeCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithTypeCoalescer(reflect.TypeOf(0), m)(c)
	expected := map[reflect.Type]Coalescer{reflect.TypeOf(0): m}
	assert.Equal(t, expected, c.typeCoalescers)
}

func TestWithDefaultCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithDefaultCoalescer(m)(c)
	assert.Equal(t, m, c.defaultCoalescer)
}

func TestWithInterfaceCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithInterfaceCoalescer(m)(c)
	assert.Equal(t, m, c.interfaceCoalescer)
}

func TestWithMapCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithMapCoalescer(m)(c)
	assert.Equal(t, m, c.mapCoalescer)
}

func TestWithPointerCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithPointerCoalescer(m)(c)
	assert.Equal(t, m, c.pointerCoalescer)
}

func TestWithSliceCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithSliceCoalescer(m)(c)
	assert.Equal(t, m, c.sliceCoalescer)
}

func TestWithStructCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	m := &mockCoalescer{}
	WithStructCoalescer(m)(c)
	assert.Equal(t, m, c.structCoalescer)
}
