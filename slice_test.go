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

func Test_coalescer_deepMergeSlice(t *testing.T) {
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
		opts []Option
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
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int(nil),
		},
		{
			"[]int union mixed zero",
			[]int{},
			[]int(nil),
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{},
		},
		{
			"[]int union mixed zero 2",
			[]int(nil),
			[]int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{},
		},
		{
			"[]int union empty",
			[]int{},
			[]int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{},
		},
		{
			"[]int union mixed empty",
			[]int{1},
			[]int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{1},
		},
		{
			"[]int union mixed empty 2",
			[]int{},
			[]int{2},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{2},
		},
		{
			"[]int union non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"[]int append zero",
			[]int(nil),
			[]int(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int(nil),
		},
		{
			"[]int append mixed zero",
			[]int{},
			[]int(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int{},
		},
		{
			"[]int append mixed zero 2",
			[]int(nil),
			[]int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int{},
		},
		{
			"[]int append empty",
			[]int{},
			[]int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int{},
		},
		{
			"[]int append mixed empty",
			[]int{1},
			[]int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int{1},
		},
		{
			"[]int append mixed empty 2",
			[]int{},
			[]int{2},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]int{2},
		},
		{
			"[]int append non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]Option{WithDefaultSliceListAppendMerge()},
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
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int(nil),
		},
		{
			"[]*int union mixed zero",
			[]*int{},
			[]*int(nil),
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{},
		},
		{
			"[]*int union mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{},
		},
		{
			"[]*int union empty",
			[]*int{},
			[]*int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{},
		},
		{
			"[]*int union mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int union mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int union non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*int{intPtr(1), intPtr(2), nil, intPtr(4), intPtr(5)},
		},
		{
			"[]*int append zero",
			[]*int(nil),
			[]*int(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int(nil),
		},
		{
			"[]*int append mixed zero",
			[]*int{},
			[]*int(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int{},
		},
		{
			"[]*int append mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int{},
		},
		{
			"[]*int append empty",
			[]*int{},
			[]*int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int{},
		},
		{
			"[]*int append mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int append mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int append non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]Option{WithDefaultSliceListAppendMerge()},
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
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo(nil),
		},
		{
			"[]foo union mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{},
		},
		{
			"[]foo union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{},
		},
		{
			"[]foo union empty",
			[]foo{},
			[]foo{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{},
		},
		{
			"[]foo union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom union zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom union mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union empty",
			[]foo{},
			[]foo{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo field zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo(nil),
		},
		{
			"[]foo field mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field empty",
			[]foo{},
			[]foo{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo field mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo field non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo merge key zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo(nil),
		},
		{
			"[]foo merge key mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key empty",
			[]foo{},
			[]foo{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge key mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge key non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]int default merge by index",
			[]int{1, 2, 0},
			[]int{0, 0, 3},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]int{1, 2, 3},
		},
		{
			"[]foo default merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo(nil),
		},
		{
			"[]foo default merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index empty",
			[]foo{},
			[]foo{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo default merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo default merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo default merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index empty",
			[]foo{},
			[]foo{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo append zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo(nil),
		},
		{
			"[]foo append mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{},
		},
		{
			"[]foo append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{},
		},
		{
			"[]foo append empty",
			[]foo{},
			[]foo{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{},
		},
		{
			"[]foo append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom append zero",
			[]foo(nil),
			[]foo(nil),
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom append mixed zero",
			[]foo{},
			[]foo(nil),
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append empty",
			[]foo{},
			[]foo{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]foo{}))},
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
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar(nil),
		},
		{
			"[]*bar union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{},
		},
		{
			"[]*bar union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{},
		},
		{
			"[]*bar union empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{},
		},
		{
			"[]*bar union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar custom union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar field zero",
			[]*bar(nil),
			[]*bar(nil),
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar(nil),
		},
		{
			"[]*bar field mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar field mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar field non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]Option{WithSliceMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(4)}},
		},
		{
			"[]*bar merge key zero",
			[]*bar(nil),
			[]*bar(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar(nil),
		},
		{
			"[]*bar merge key mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar merge key mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar merge key non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
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
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}(nil),
		},
		{
			"[]interface{} union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{},
		},
		{
			"[]interface{} union empty",
			[]interface{}{},
			[]interface{}{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]Option{WithDefaultSliceSetUnionMerge()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} custom union zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}(nil),
		},
		{
			"[]interface{} custom union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union empty",
			[]interface{}{},
			[]interface{}{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} custom union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} custom union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]Option{WithSliceSetUnionMerge(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} merge key zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}(nil),
		},
		{
			"[]interface{} merge key mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key empty",
			[]interface{}{},
			[]interface{}{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} merge key mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} merge key non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: nil}, nil},
			[]Option{WithSliceMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(4)}},
		},
		{
			"[]bar default merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar(nil),
		},
		{
			"[]bar default merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index empty",
			[]bar{},
			[]bar{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar default merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar default merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar default merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]Option{WithDefaultSliceMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar(nil),
		},
		{
			"[]bar merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index empty",
			[]bar{},
			[]bar{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]Option{WithSliceMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},

		{
			"[]*bar append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar(nil),
		},
		{
			"[]*bar append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{},
		},
		{
			"[]*bar append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{},
		},
		{
			"[]*bar append empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{},
		},
		{
			"[]*bar append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]Option{WithDefaultSliceListAppendMerge()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]*bar custom append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append empty",
			[]*bar{},
			[]*bar{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]Option{WithSliceListAppendMerge(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]int nil + nil w/ zero empty option",
			[]int(nil),
			[]int(nil),
			[]Option{WithZeroEmptySliceMerge()},
			[]int(nil),
		},
		{
			"[]int nil + empty w/ zero empty option",
			[]int(nil),
			[]int{},
			[]Option{WithZeroEmptySliceMerge()},
			[]int{},
		},
		{
			"[]int empty + nil w/ zero empty option",
			[]int{},
			[]int(nil),
			[]Option{WithZeroEmptySliceMerge()},
			[]int{},
		},
		{
			"[]int empty + empty w/ zero empty option",
			[]int{},
			[]int{},
			[]Option{WithZeroEmptySliceMerge()},
			[]int{},
		},
		{
			"[]int empty + non-empty w/ zero empty option",
			[]int{},
			[]int{1, 2, 3},
			[]Option{WithZeroEmptySliceMerge()},
			[]int{1, 2, 3},
		},
		{
			"[]int non-empty + empty w/ zero empty option",
			[]int{1, 2, 3},
			[]int{},
			[]Option{WithZeroEmptySliceMerge()},
			[]int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeSlice(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
			assertNotSame(t, tt.v1, got.Interface())
			assertNotSame(t, tt.v2, got.Interface())
		})
	}
	t.Run("generic error", func(t *testing.T) {
		c := newCoalescer(withMockDeepCopyError)
		_, err := c.deepMergeSlice(reflect.ValueOf([]int{}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "mock DeepCopy error")
	})
	t.Run("append errors", func(t *testing.T) {
		c := newCoalescer(WithDefaultSliceListAppendMerge(), withMockDeepCopyError)
		_, err := c.deepMergeSlice(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "mock DeepCopy error")
	})
	t.Run("merge key errors", func(t *testing.T) {
		mergeKeyFunc := func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, nil
		}
		c := newCoalescer(WithSliceMergeByKeyFunc(reflect.TypeOf([]int{}), mergeKeyFunc))
		_, err := c.deepMergeSlice(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		_, err = c.deepMergeSlice(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		c = newCoalescer(WithSliceMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.ValueOf([]int{1, 2, 3}), nil
		}))
		_, err = c.deepMergeSlice(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		_, err = c.deepMergeSlice(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		c = newCoalescer(WithSliceMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("merge key func error")
		}))
		_, err = c.deepMergeSlice(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "merge key func error")
		_, err = c.deepMergeSlice(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "merge key func error")
	})
}

func Test_coalescer_deepMergeSliceWithAppend(t *testing.T) {
	// Note: we don't need to test all the corner cases here, as these are thoroughly tested in
	// Test_coalescer_deepMergeSlice.
	tests := []struct {
		name    string
		v1      reflect.Value
		v2      reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "v1 nil",
			v1:   reflect.ValueOf([]int(nil)),
			v2:   reflect.ValueOf([]int{3, 4, 5}),
			want: reflect.ValueOf([]int{3, 4, 5}),
		},
		{
			name: "v2 nil",
			v1:   reflect.ValueOf([]int{1, 2, 3}),
			v2:   reflect.ValueOf([]int(nil)),
			want: reflect.ValueOf([]int{1, 2, 3}),
		},
		{
			name: "empty",
			v1:   reflect.ValueOf([]int{}),
			v2:   reflect.ValueOf([]int{}),
			want: reflect.ValueOf([]int{}),
		},
		{
			name: "simple",
			v1:   reflect.ValueOf([]int{1, 2, 3}),
			v2:   reflect.ValueOf([]int{4, 5, 6}),
			want: reflect.ValueOf([]int{1, 2, 3, 4, 5, 6}),
		},
		{
			name:    "error copy v1",
			v1:      reflect.ValueOf([]int{1, 2, 3}),
			v2:      reflect.ValueOf([]int{4, 5, 6}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(1)},
		},
		{
			name:    "error copy v2",
			v1:      reflect.ValueOf([]int{1, 2, 3}),
			v2:      reflect.ValueOf([]int{4, 5, 6}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(4)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeSliceWithListAppend(tt.v1, tt.v2)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v1.Interface(), got.Interface())
				assertNotSame(t, tt.v2.Interface(), got.Interface())
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

func Test_coalescer_deepMergeSliceWithMergeKey(t *testing.T) {
	// Note: we don't need to test all the corner cases here, as these are thoroughly tested in
	// Test_coalescer_deepMergeSlice.
	tests := []struct {
		name    string
		v1      reflect.Value
		v2      reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "v1 nil",
			v1:   reflect.ValueOf([]int(nil)),
			v2:   reflect.ValueOf([]int{3, 4, 5}),
			want: reflect.ValueOf([]int{3, 4, 5}),
		},
		{
			name: "v2 nil",
			v1:   reflect.ValueOf([]int{1, 2, 3}),
			v2:   reflect.ValueOf([]int(nil)),
			want: reflect.ValueOf([]int{1, 2, 3}),
		},
		{
			name: "empty",
			v1:   reflect.ValueOf([]int{}),
			v2:   reflect.ValueOf([]int{}),
			want: reflect.ValueOf([]int{}),
		},
		{
			name: "simple",
			v1:   reflect.ValueOf([]int{1, 2, 3}),
			v2:   reflect.ValueOf([]int{3, 4, 5}),
			want: reflect.ValueOf([]int{1, 2, 3, 4, 5}),
		},
		{
			name:    "error copy v1",
			v1:      reflect.ValueOf([]int{1, 2, 3}),
			v2:      reflect.ValueOf([]int{3, 4, 5}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(1)},
		},
		{
			name:    "error copy v2",
			v1:      reflect.ValueOf([]int{1, 2, 3}),
			v2:      reflect.ValueOf([]int{3, 4, 5}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyErrorWhen(4)},
		},
		{
			name:    "error merge",
			v1:      reflect.ValueOf([]int{1, 2, 3}),
			v2:      reflect.ValueOf([]int{3, 4, 5}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepMergeError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepMergeSliceWithMergeKey(tt.v1, tt.v2, SliceUnion)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v1.Interface(), got.Interface())
				assertNotSame(t, tt.v2.Interface(), got.Interface())
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

func Test_coalescer_deepCopySlice(t *testing.T) {
	tests := []struct {
		name    string
		v       reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "nil",
			v:    reflect.ValueOf([]int(nil)),
			want: reflect.ValueOf([]int(nil)),
		},
		{
			name: "empty",
			v:    reflect.ValueOf([]int{}),
			want: reflect.ValueOf([]int{}),
		},
		{
			name: "non empty",
			v:    reflect.ValueOf([]int{1, 2, 3}),
			want: reflect.ValueOf([]int{1, 2, 3}),
		},
		{
			name:    "error",
			v:       reflect.ValueOf([]int{1, 2, 3}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepCopySlice(tt.v)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v.Interface(), got.Interface())
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
