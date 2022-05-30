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

func TestWithAtomicType(t *testing.T) {
	c := &mainCoalescer{}
	WithAtomicType(reflect.TypeOf(0))(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(0)])
}

func TestWithTrileans(t *testing.T) {
	c := &mainCoalescer{}
	WithTrileans()(c)
	assert.NotNil(t, c.typeCoalescers[reflect.PtrTo(reflect.TypeOf(false))])
}

func TestWithTypeCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	WithTypeCoalescer(reflect.TypeOf(0), coalesceAtomic)(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(0)])
}

func TestWithTypeCoalescerProvider(t *testing.T) {
	c := &mainCoalescer{}
	WithTypeCoalescerProvider(reflect.TypeOf(0), func(parent Coalescer) Coalescer {
		return coalesceAtomic
	})(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(0)])
}

func TestWithFieldCoalescer(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	WithFieldCoalescer(reflect.TypeOf(User{}), "ID", coalesceAtomic)(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
}

func TestWithFieldCoalescerProvider(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	WithFieldCoalescerProvider(reflect.TypeOf(User{}), "ID", func(parent Coalescer) Coalescer {
		return coalesceAtomic
	})(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
}

func TestWithAtomicField(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	WithAtomicField(reflect.TypeOf(User{}), "ID")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
}

func TestWithDefaultListAppend(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultListAppend()(c)
	assert.NotNil(t, c.sliceCoalescer)
}

func TestWithDefaultMergeByIndex(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultMergeByIndex()(c)
	assert.NotNil(t, c.sliceCoalescer)
}

func TestWithDefaultSetUnion(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultSetUnion()(c)
	assert.NotNil(t, c.sliceCoalescer)
}

func TestWithErrorOnCycle(t *testing.T) {
	c := &mainCoalescer{}
	WithErrorOnCycle()(c)
	assert.Equal(t, true, c.errorOnCycle)
}

func TestWithListAppend(t *testing.T) {
	c := &mainCoalescer{}
	WithListAppend(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
}

func TestWithSetUnion(t *testing.T) {
	c := &mainCoalescer{}
	WithSetUnion(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
}

func TestWithMergeByIndex(t *testing.T) {
	c := &mainCoalescer{}
	WithMergeByIndex(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
}

func TestWithMergeByKey(t *testing.T) {
	c := &mainCoalescer{}
	mergeKeyFunc := func(index int, element reflect.Value) (key reflect.Value) { return reflect.ValueOf("whatever") }
	WithMergeByKey(reflect.TypeOf([]int{}), mergeKeyFunc)(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
}

func TestWithMergeByField(t *testing.T) {
	type foo struct{ ID string }
	c := &mainCoalescer{}
	WithMergeByField(reflect.TypeOf([]foo{}), "ID")(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]foo{})])
}

func TestWithZeroEmptySlice(t *testing.T) {
	c := &mainCoalescer{}
	WithZeroEmptySlice()(c)
	assert.Equal(t, true, c.zeroEmptySlice)
}
