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

func Test_mainCoalescer_coalesceSlice(t *testing.T) {
	type foo struct {
		FieldInt int
	}
	type bar struct {
		FieldIntPtr *int
	}
	fooMergeFunc := func(_ int, value reflect.Value) (reflect.Value, error) {
		elem := value.Interface().(foo)
		return reflect.ValueOf(elem.FieldInt), nil
	}
	barPtrMergeFunc := func(_ int, value reflect.Value) (reflect.Value, error) {
		elem := value.Interface().(*bar)
		if elem == nil {
			elem = &bar{}
		}
		if elem.FieldIntPtr == nil {
			elem.FieldIntPtr = new(int)
		}
		key := *elem.FieldIntPtr
		return reflect.ValueOf(key), nil
	}
	barPtrInterfaceMergeFunc := func(i int, value reflect.Value) (reflect.Value, error) {
		var elem *bar
		if !value.IsNil() {
			elem = value.Interface().(*bar)
		}
		if elem == nil {
			elem = &bar{}
		}
		if elem.FieldIntPtr == nil {
			elem.FieldIntPtr = new(int)
		}
		key := *elem.FieldIntPtr
		return reflect.ValueOf(key), nil
	}
	tests := []struct {
		name string
		v1   interface{}
		v2   interface{}
		opt  []CoalescerOption
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
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int(nil),
		},
		{
			"[]int union mixed zero",
			[]int{},
			[]int(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed zero 2",
			[]int(nil),
			[]int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union empty",
			[]int{},
			[]int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed empty",
			[]int{1},
			[]int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{1},
		},
		{
			"[]int union mixed empty 2",
			[]int{},
			[]int{2},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{2},
		},
		{
			"[]int union non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"[]int append zero",
			[]int(nil),
			[]int(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]int(nil),
		},
		{
			"[]int append mixed zero",
			[]int{},
			[]int(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed zero 2",
			[]int(nil),
			[]int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append empty",
			[]int{},
			[]int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed empty",
			[]int{1},
			[]int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{1},
		},
		{
			"[]int append mixed empty 2",
			[]int{},
			[]int{2},
			[]CoalescerOption{WithDefaultListAppend()},
			[]int{2},
		},
		{
			"[]int append non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]CoalescerOption{WithDefaultListAppend()},
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
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int(nil),
		},
		{
			"[]*int union mixed zero",
			[]*int{},
			[]*int(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union empty",
			[]*int{},
			[]*int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int union mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int union non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*int{intPtr(1), intPtr(2), nil, intPtr(4), intPtr(5)},
		},
		{
			"[]*int append zero",
			[]*int(nil),
			[]*int(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int(nil),
		},
		{
			"[]*int append mixed zero",
			[]*int{},
			[]*int(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append empty",
			[]*int{},
			[]*int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int append mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int append non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]CoalescerOption{WithDefaultListAppend()},
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
			[]foo{{FieldInt: 1}},
			[]foo{},
			nil,
			[]foo{},
		},
		{
			"[]foo mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			nil,
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			nil,
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo union zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo(nil),
		},
		{
			"[]foo union mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom union zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom union mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo field zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo(nil),
		},
		{
			"[]foo field mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo field mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo field non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo merge key zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo(nil),
		},
		{
			"[]foo merge key mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge key mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge key non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo default merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo(nil),
		},
		{
			"[]foo default merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo default merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo default merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo default merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo append zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo(nil),
		},
		{
			"[]foo append mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom append zero",
			[]foo(nil),
			[]foo(nil),
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom append mixed zero",
			[]foo{},
			[]foo(nil),
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append empty",
			[]foo{},
			[]foo{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
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
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			nil,
			[]*bar{},
		},
		{
			"[]*bar mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			nil,
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			nil,
			[]*bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]*bar union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar(nil),
		},
		{
			"[]*bar union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar custom union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar field zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar(nil),
		},
		{
			"[]*bar field mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar field mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar field non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]CoalescerOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(4)}},
		},
		{
			"[]*bar merge key zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar(nil),
		},
		{
			"[]*bar merge key mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar merge key mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar merge key non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(4)}},
		},
		{
			"[]interface{} zero",
			[]interface{}(nil),
			[]interface{}(nil),
			nil,
			[]interface{}(nil),
		},
		{
			"[]interface{} mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			nil,
			[]interface{}{},
		},
		{
			"[]interface{} mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			nil,
			[]interface{}{},
		},
		{
			"[]interface{} empty",
			[]interface{}{},
			[]interface{}{},
			nil,
			[]interface{}{},
		},
		{
			"[]interface{} mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			nil,
			[]interface{}{},
		},
		{
			"[]interface{} mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			nil,
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: nil}, nil},
			nil,
			[]interface{}{&bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: nil}, nil},
		},
		{
			"[]interface{} union zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}(nil),
		},
		{
			"[]interface{} union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union empty",
			[]interface{}{},
			[]interface{}{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]CoalescerOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} custom union zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}(nil),
		},
		{
			"[]interface{} custom union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union empty",
			[]interface{}{},
			[]interface{}{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} custom union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} custom union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]CoalescerOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} merge key zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}(nil),
		},
		{
			"[]interface{} merge key mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key empty",
			[]interface{}{},
			[]interface{}{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} merge key mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} merge key non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: nil}, nil},
			[]CoalescerOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(4)}},
		},
		{
			"[]bar default merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar(nil),
		},
		{
			"[]bar default merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index empty",
			[]bar{},
			[]bar{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar default merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar default merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar default merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]CoalescerOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar(nil),
		},
		{
			"[]bar merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index empty",
			[]bar{},
			[]bar{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]CoalescerOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},

		{
			"[]*bar append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar(nil),
		},
		{
			"[]*bar append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]CoalescerOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]*bar custom append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append empty",
			[]*bar{},
			[]*bar{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]CoalescerOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]int nil + nil w/ zero empty option",
			[]int(nil),
			[]int(nil),
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int(nil),
		},
		{
			"[]int nil + empty w/ zero empty option",
			[]int(nil),
			[]int{},
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + nil w/ zero empty option",
			[]int{},
			[]int(nil),
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + empty w/ zero empty option",
			[]int{},
			[]int{},
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + non-empty w/ zero empty option",
			[]int{},
			[]int{1, 2, 3},
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int{1, 2, 3},
		},
		{
			"[]int non-empty + empty w/ zero empty option",
			[]int{1, 2, 3},
			[]int{},
			[]CoalescerOption{WithZeroEmptySlice()},
			[]int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coalescer := NewCoalescer(tt.opt...)
			got, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("fallback error", func(t *testing.T) {
		coalescer := NewCoalescer(WithDefaultSetUnion(), WithTypeCoalescer(reflect.TypeOf(map[interface{}]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		_, err := coalescer(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{2}))
		assert.EqualError(t, err, "fake")
	})
	t.Run("merge key func errors", func(t *testing.T) {
		coalescer := NewCoalescer(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, nil
		}))
		_, err := coalescer(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		_, err = coalescer(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		coalescer = NewCoalescer(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.ValueOf([]int{1, 2, 3}), nil
		}))
		_, err = coalescer(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		_, err = coalescer(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		coalescer = NewCoalescer(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("merge key func error")
		}))
		_, err = coalescer(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "merge key func error")
		_, err = coalescer(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "merge key func error")
	})
}
