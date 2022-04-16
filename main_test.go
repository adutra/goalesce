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
	"reflect"
	"testing"
)

func TestNewMainCoalescer(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		expected := &mainCoalescer{}
		expected.defaultCoalescer = &defaultCoalescer{}
		expected.pointerCoalescer = &pointerCoalescer{fallback: expected}
		expected.mapCoalescer = &mapCoalescer{fallback: expected}
		expected.structCoalescer = &structCoalescer{fallback: expected}
		expected.sliceCoalescer = &sliceCoalescer{
			defaultCoalescer: &defaultCoalescer{},
		}
		actual := NewMainCoalescer()
		assert.Equal(t, expected, actual)
	})
	t.Run("with option", func(t *testing.T) {
		var passed *mainCoalescer
		opt := func(c *mainCoalescer) {
			passed = c
		}
		returned := NewMainCoalescer(opt)
		assert.Equal(t, passed, returned)
	})
}

func TestWithAtomicType(t *testing.T) {
	c := &mainCoalescer{}
	WithAtomicType(reflect.TypeOf(0))(c)
	expected := map[reflect.Type]Coalescer{reflect.TypeOf(0): &defaultCoalescer{}}
	assert.Equal(t, expected, c.typeCoalescers)
}

func TestWithTypeCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithTypeCoalescer(reflect.TypeOf(0), mock)(c)
	expected := map[reflect.Type]Coalescer{reflect.TypeOf(0): mock}
	assert.Equal(t, expected, c.typeCoalescers)
}

func TestWithDefaultCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithDefaultCoalescer(mock)(c)
	assert.Equal(t, mock, c.defaultCoalescer)
}

func TestWithMapCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithMapCoalescer(mock)(c)
	assert.Equal(t, mock, c.mapCoalescer)
}

func TestWithPointerCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithPointerCoalescer(mock)(c)
	assert.Equal(t, mock, c.pointerCoalescer)
}

func TestWithSliceCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithSliceCoalescer(mock)(c)
	assert.Equal(t, mock, c.sliceCoalescer)
}

func TestWithStructCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	mock := &mockCoalescer{}
	WithStructCoalescer(mock)(c)
	assert.Equal(t, mock, c.structCoalescer)
}
