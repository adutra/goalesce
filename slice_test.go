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

func TestNewSliceCoalescer(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		got := NewSliceCoalescer()
		assert.Equal(t, &sliceCoalescer{defaultCoalescer: &defaultCoalescer{}}, got)
	})
	t.Run("with default union", func(t *testing.T) {
		got := NewSliceCoalescer(WithDefaultSetUnion())
		assert.IsType(t, &sliceMergeCoalescer{}, got.(*sliceCoalescer).defaultCoalescer)
	})
	t.Run("with default append", func(t *testing.T) {
		got := NewSliceCoalescer(WithDefaultListAppend())
		assert.Equal(t, &sliceCoalescer{defaultCoalescer: &sliceAppendCoalescer{}}, got)
	})
	t.Run("with append", func(t *testing.T) {
		got := NewSliceCoalescer(WithListAppend(reflect.TypeOf("")))
		assert.Equal(t, &defaultCoalescer{}, got.(*sliceCoalescer).defaultCoalescer)
		assert.Equal(t, &sliceAppendCoalescer{}, got.(*sliceCoalescer).elemCoalescers[reflect.TypeOf("")])
	})
	t.Run("with merge by union", func(t *testing.T) {
		got := NewSliceCoalescer(WithSetUnion(reflect.TypeOf(0)))
		assert.Equal(t, &defaultCoalescer{}, got.(*sliceCoalescer).defaultCoalescer)
		assert.IsType(t, &sliceMergeCoalescer{}, got.(*sliceCoalescer).elemCoalescers[reflect.TypeOf(0)])
	})
	t.Run("with merge by field", func(t *testing.T) {
		type foo struct {
			Int int
		}
		got := NewSliceCoalescer(WithMergeByField(reflect.TypeOf(foo{}), "Int"))
		assert.Equal(t, &defaultCoalescer{}, got.(*sliceCoalescer).defaultCoalescer)
		assert.IsType(t, &sliceMergeCoalescer{}, got.(*sliceCoalescer).elemCoalescers[reflect.TypeOf(foo{})])
	})
	t.Run("with merge by key", func(t *testing.T) {
		type foo struct {
			Int int
		}
		got := NewSliceCoalescer(WithMergeByKey(reflect.TypeOf(foo{}), func(value reflect.Value) reflect.Value {
			return reflect.ValueOf(value.Interface().(foo).Int)
		}))
		assert.Equal(t, &defaultCoalescer{}, got.(*sliceCoalescer).defaultCoalescer)
		assert.IsType(t, &sliceMergeCoalescer{}, got.(*sliceCoalescer).elemCoalescers[reflect.TypeOf(foo{})])
	})
}

func Test_sliceCoalescer_Coalesce(t *testing.T) {
	type foo struct {
		Int int
	}
	type bar struct {
		Int *int
	}
	fooMergeFunc := func(value reflect.Value) reflect.Value {
		elem := value.Interface().(foo)
		return reflect.ValueOf(elem.Int)
	}
	barPtrMergeFunc := func(value reflect.Value) reflect.Value {
		elem := value.Interface().(*bar)
		if elem == nil {
			elem = &bar{}
		}
		if elem.Int == nil {
			elem.Int = new(int)
		}
		key := *elem.Int
		return reflect.ValueOf(key)
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		opt  []SliceCoalescerOption
		want interface{}
	}{
		{
			"[]int zero",
			[]int(nil),
			[]int(nil),
			nil,
			[]int(nil),
		},
		{
			"[]int mixed zero",
			[]int{},
			[]int(nil),
			nil,
			[]int{},
		},
		{
			"[]int mixed zero 2",
			[]int(nil),
			[]int{},
			nil,
			[]int{},
		},
		{
			"[]int empty",
			[]int{},
			[]int{},
			nil,
			[]int{},
		},
		{
			"[]int mixed empty",
			[]int{1},
			[]int{},
			nil,
			[]int{},
		},
		{
			"[]int mixed empty 2",
			[]int{},
			[]int{2},
			nil,
			[]int{2},
		},
		{
			"[]int non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			nil,
			[]int{3, 4, 5},
		},
		{
			"[]int union zero",
			[]int(nil),
			[]int(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int(nil),
		},
		{
			"[]int union mixed zero",
			[]int{},
			[]int(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed zero 2",
			[]int(nil),
			[]int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union empty",
			[]int{},
			[]int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed empty",
			[]int{1},
			[]int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{1},
		},
		{
			"[]int union mixed empty 2",
			[]int{},
			[]int{2},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{2},
		},
		{
			"[]int union non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"[]int append zero",
			[]int(nil),
			[]int(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int(nil),
		},
		{
			"[]int append mixed zero",
			[]int{},
			[]int(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed zero 2",
			[]int(nil),
			[]int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append empty",
			[]int{},
			[]int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed empty",
			[]int{1},
			[]int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{1},
		},
		{
			"[]int append mixed empty 2",
			[]int{},
			[]int{2},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{2},
		},
		{
			"[]int append non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]int{1, 2, 3, 3, 4, 5},
		},
		{
			"[]*int zero",
			[]*int(nil),
			[]*int(nil),
			nil,
			[]*int(nil),
		},
		{
			"[]*int mixed zero",
			[]*int{},
			[]*int(nil),
			nil,
			[]*int{},
		},
		{
			"[]*int mixed zero 2",
			[]*int(nil),
			[]*int{},
			nil,
			[]*int{},
		},
		{
			"[]*int empty",
			[]*int{},
			[]*int{},
			nil,
			[]*int{},
		},
		{
			"[]*int mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			nil,
			[]*int{},
		},
		{
			"[]*int mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			nil,
			[]*int{intPtr(2)},
		},
		{
			"[]*int non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(4), intPtr(5), nil},
			nil,
			[]*int{intPtr(4), intPtr(5), nil},
		},
		{
			"[]*int union zero",
			[]*int(nil),
			[]*int(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int(nil),
		},
		{
			"[]*int union mixed zero",
			[]*int{},
			[]*int(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union empty",
			[]*int{},
			[]*int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int union mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int union non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(1), intPtr(2), nil, intPtr(4), intPtr(5)},
		},
		{
			"[]*int append zero",
			[]*int(nil),
			[]*int(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int(nil),
		},
		{
			"[]*int append mixed zero",
			[]*int{},
			[]*int(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append empty",
			[]*int{},
			[]*int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int append mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int append non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*int{intPtr(1), intPtr(2), nil, intPtr(2), intPtr(4), intPtr(5), nil},
		},

		{
			"[]foo zero",
			[]foo(nil),
			[]foo(nil),
			nil,
			[]foo(nil),
		},
		{
			"[]foo mixed zero",
			[]foo{},
			[]foo(nil),
			nil,
			[]foo{},
		},
		{
			"[]foo mixed zero 2",
			[]foo(nil),
			[]foo{},
			nil,
			[]foo{},
		},
		{
			"[]foo empty",
			[]foo{},
			[]foo{},
			nil,
			[]foo{},
		},
		{
			"[]foo mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			nil,
			[]foo{},
		},
		{
			"[]foo mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			nil,
			[]foo{{Int: 2}},
		},
		{
			"[]foo non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			nil,
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo union zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo(nil),
		},
		{
			"[]foo union mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{{Int: 1}},
		},
		{
			"[]foo union mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{{Int: 2}},
		},
		{
			"[]foo union non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo custom union zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom union mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{{Int: 1}},
		},
		{
			"[]foo custom union mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{{Int: 2}},
		},
		{
			"[]foo custom union non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithSetUnion(reflect.TypeOf(foo{}))},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo field zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo(nil),
		},
		{
			"[]foo field mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{},
		},
		{
			"[]foo field mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{},
		},
		{
			"[]foo field empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{},
		},
		{
			"[]foo field mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{{Int: 1}},
		},
		{
			"[]foo field mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{{Int: 2}},
		},
		{
			"[]foo field non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(foo{}), "Int")},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo merge key zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo(nil),
		},
		{
			"[]foo merge key mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{{Int: 1}},
		},
		{
			"[]foo merge key mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{{Int: 2}},
		},
		{
			"[]foo merge key non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithMergeByKey(reflect.TypeOf(foo{}), fooMergeFunc)},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo append zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo(nil),
		},
		{
			"[]foo append mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{{Int: 1}},
		},
		{
			"[]foo append mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{{Int: 2}},
		},
		{
			"[]foo append non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]foo custom append zero",
			[]foo(nil),
			[]foo(nil),
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom append mixed zero",
			[]foo{},
			[]foo(nil),
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append empty",
			[]foo{},
			[]foo{},
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed empty",
			[]foo{{Int: 1}},
			[]foo{},
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{{Int: 1}},
		},
		{
			"[]foo custom append mixed empty 2",
			[]foo{},
			[]foo{{Int: 2}},
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{{Int: 2}},
		},
		{
			"[]foo custom append non empty",
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}},
			[]foo{{Int: 3}, {Int: 4}, {Int: 5}},
			[]SliceCoalescerOption{WithListAppend(reflect.TypeOf(foo{}))},
			[]foo{{Int: 1}, {Int: 2}, {Int: 3}, {Int: 3}, {Int: 4}, {Int: 5}},
		},
		{
			"[]*bar zero",
			[]*bar(nil),
			[]*bar(nil),
			nil,
			[]*bar(nil),
		},
		{
			"[]*bar mixed zero",
			[]*bar{},
			[]*bar(nil),
			nil,
			[]*bar{},
		},
		{
			"[]*bar mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			nil,
			[]*bar{},
		},
		{
			"[]*bar empty",
			[]*bar{},
			[]*bar{},
			nil,
			[]*bar{},
		},
		{
			"[]*bar mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			nil,
			[]*bar{},
		},
		{
			"[]*bar mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			nil,
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(4)}, {Int: nil}, nil},
			nil,
			[]*bar{{Int: intPtr(4)}, {Int: nil}, nil},
		},
		{
			"[]*bar union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar(nil),
		},
		{
			"[]*bar union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar union mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar union non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: intPtr(5)}, nil},
			[]SliceCoalescerOption{WithDefaultSetUnion()},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(2)}, {Int: intPtr(4)}, {Int: intPtr(5)}},
		},
		{
			"[]*bar custom union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar(nil),
		},
		{
			"[]*bar custom union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom union empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar custom union mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar custom union non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: intPtr(5)}, nil},
			[]SliceCoalescerOption{WithSetUnion(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(2)}, {Int: intPtr(4)}, {Int: intPtr(5)}},
		},
		{
			"[]*bar field zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar(nil),
		},
		{
			"[]*bar field mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{},
		},
		{
			"[]*bar field mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{},
		},
		{
			"[]*bar field empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{},
		},
		{
			"[]*bar field mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar field mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar field non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
			[]SliceCoalescerOption{WithMergeByField(reflect.TypeOf(bar{}), "Int")},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(4)}},
		},
		{
			"[]*bar merge key zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar(nil),
		},
		{
			"[]*bar merge key mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar merge key mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar merge key non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
			[]SliceCoalescerOption{WithMergeByKey(reflect.PtrTo(reflect.TypeOf(bar{})), barPtrMergeFunc)},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(4)}},
		},
		{
			"[]*bar append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar(nil),
		},
		{
			"[]*bar append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar append mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar append non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
			[]SliceCoalescerOption{WithDefaultListAppend()},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
		},
		{
			"[]*bar custom append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar(nil),
		},
		{
			"[]*bar custom append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom append empty",
			[]*bar{},
			[]*bar{},
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed empty",
			[]*bar{{Int: intPtr(1)}},
			[]*bar{},
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(1)}},
		},
		{
			"[]*bar custom append mixed empty 2",
			[]*bar{},
			[]*bar{{Int: intPtr(2)}},
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(2)}},
		},
		{
			"[]*bar custom append non empty",
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil},
			[]*bar{{Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
			[]SliceCoalescerOption{WithListAppend(reflect.PtrTo(reflect.TypeOf(bar{})))},
			[]*bar{{Int: intPtr(1)}, {Int: intPtr(2)}, nil, {Int: intPtr(2)}, {Int: intPtr(4)}, {Int: nil}, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewSliceCoalescer(tt.opt...)
			coalescer.WithFallback(NewMainCoalescer())
			got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, got.Interface())
		})
	}
	t.Run("type errors", func(t *testing.T) {
		_, err := (&sliceCoalescer{}).Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = (&sliceCoalescer{}).Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected slice, got int")
		_, err = (&sliceMergeCoalescer{}).Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = (&sliceMergeCoalescer{}).Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected slice, got int")
		_, err = (&sliceAppendCoalescer{}).Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = (&sliceAppendCoalescer{}).Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected slice, got int")
	})
	t.Run("fallback error", func(t *testing.T) {
		m := new(mockCoalescer)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		m.Test(t)
		coalescer := NewSliceCoalescer(WithDefaultSetUnion())
		coalescer.WithFallback(m)
		_, err := coalescer.Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{2}))
		assert.EqualError(t, err, "fake")
	})
	t.Run("merge key func error", func(t *testing.T) {
		coalescer := NewSliceCoalescer(WithMergeByKey(reflect.TypeOf(0), func(value reflect.Value) reflect.Value {
			return reflect.Value{}
		}))
		_, err := coalescer.Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		_, err = coalescer.Coalesce(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		coalescer = NewSliceCoalescer(WithMergeByKey(reflect.TypeOf(0), func(value reflect.Value) reflect.Value {
			return reflect.ValueOf([]int{1, 2, 3})
		}))
		_, err = coalescer.Coalesce(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		_, err = coalescer.Coalesce(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
	})
}
