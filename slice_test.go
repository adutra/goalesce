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
		opt  []DeepMergeOption
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
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int(nil),
		},
		{
			"[]int union mixed zero",
			[]int{},
			[]int(nil),
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed zero 2",
			[]int(nil),
			[]int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union empty",
			[]int{},
			[]int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{},
		},
		{
			"[]int union mixed empty",
			[]int{1},
			[]int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{1},
		},
		{
			"[]int union mixed empty 2",
			[]int{},
			[]int{2},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{2},
		},
		{
			"[]int union non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]int{1, 2, 3, 4, 5},
		},
		{
			"[]int append zero",
			[]int(nil),
			[]int(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int(nil),
		},
		{
			"[]int append mixed zero",
			[]int{},
			[]int(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed zero 2",
			[]int(nil),
			[]int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append empty",
			[]int{},
			[]int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int{},
		},
		{
			"[]int append mixed empty",
			[]int{1},
			[]int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int{1},
		},
		{
			"[]int append mixed empty 2",
			[]int{},
			[]int{2},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]int{2},
		},
		{
			"[]int append non empty",
			[]int{1, 2, 3},
			[]int{3, 4, 5},
			[]DeepMergeOption{WithDefaultListAppend()},
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
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int(nil),
		},
		{
			"[]*int union mixed zero",
			[]*int{},
			[]*int(nil),
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union empty",
			[]*int{},
			[]*int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{},
		},
		{
			"[]*int union mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int union mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int union non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*int{intPtr(1), intPtr(2), nil, intPtr(4), intPtr(5)},
		},
		{
			"[]*int append zero",
			[]*int(nil),
			[]*int(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int(nil),
		},
		{
			"[]*int append mixed zero",
			[]*int{},
			[]*int(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed zero 2",
			[]*int(nil),
			[]*int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append empty",
			[]*int{},
			[]*int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int{},
		},
		{
			"[]*int append mixed empty",
			[]*int{intPtr(1)},
			[]*int{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int{intPtr(1)},
		},
		{
			"[]*int append mixed empty 2",
			[]*int{},
			[]*int{intPtr(2)},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*int{intPtr(2)},
		},
		{
			"[]*int append non empty",
			[]*int{intPtr(1), intPtr(2), nil},
			[]*int{intPtr(2), intPtr(4), intPtr(5), nil},
			[]DeepMergeOption{WithDefaultListAppend()},
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
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo(nil),
		},
		{
			"[]foo union mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{},
		},
		{
			"[]foo union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom union zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom union mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom union mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom union mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom union non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo field zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo(nil),
		},
		{
			"[]foo field mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{},
		},
		{
			"[]foo field mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo field mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo field non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]foo{}), "FieldInt")},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo merge key zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo(nil),
		},
		{
			"[]foo merge key mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{},
		},
		{
			"[]foo merge key mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge key mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge key non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]foo{}), fooMergeFunc)},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo default merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo(nil),
		},
		{
			"[]foo default merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{},
		},
		{
			"[]foo default merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo default merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo default merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo default merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo merge by index mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo merge by index mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo merge by index mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo merge by index non empty 1",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 4}, {FieldInt: 5}, {FieldInt: 3}},
		},
		{
			"[]foo merge by index non empty 2",
			[]foo{{FieldInt: 4}, {FieldInt: 5}},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
		},
		{
			"[]foo append zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo(nil),
		},
		{
			"[]foo append mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{},
		},
		{
			"[]foo append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}, {FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
		},
		{
			"[]foo custom append zero",
			[]foo(nil),
			[]foo(nil),
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo(nil),
		},
		{
			"[]foo custom append mixed zero",
			[]foo{},
			[]foo(nil),
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed zero 2",
			[]foo(nil),
			[]foo{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append empty",
			[]foo{},
			[]foo{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{},
		},
		{
			"[]foo custom append mixed empty",
			[]foo{{FieldInt: 1}},
			[]foo{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 1}},
		},
		{
			"[]foo custom append mixed empty 2",
			[]foo{},
			[]foo{{FieldInt: 2}},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
			[]foo{{FieldInt: 2}},
		},
		{
			"[]foo custom append non empty",
			[]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}},
			[]foo{{FieldInt: 3}, {FieldInt: 4}, {FieldInt: 5}},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]foo{}))},
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
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar(nil),
		},
		{
			"[]*bar union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{},
		},
		{
			"[]*bar union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar custom union zero",
			[]*bar(nil),
			[]*bar(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom union mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom union mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom union mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom union non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, nil},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
		},
		{
			"[]*bar field zero",
			[]*bar(nil),
			[]*bar(nil),
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar(nil),
		},
		{
			"[]*bar field mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{},
		},
		{
			"[]*bar field mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar field mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar field non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]DeepMergeOption{WithMergeByID(reflect.TypeOf([]*bar{}), "FieldIntPtr")},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(4)}},
		},
		{
			"[]*bar merge key zero",
			[]*bar(nil),
			[]*bar(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar(nil),
		},
		{
			"[]*bar merge key mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{},
		},
		{
			"[]*bar merge key mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar merge key mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar merge key non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]*bar{}), barPtrMergeFunc)},
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
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}(nil),
		},
		{
			"[]interface{} union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union empty",
			[]interface{}{},
			[]interface{}{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{},
		},
		{
			"[]interface{} union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]DeepMergeOption{WithDefaultSetUnion()},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} custom union zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}(nil),
		},
		{
			"[]interface{} custom union mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union empty",
			[]interface{}{},
			[]interface{}{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{},
		},
		{
			"[]interface{} custom union mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} custom union mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} custom union non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}, nil},
			[]DeepMergeOption{WithSetUnion(reflect.TypeOf([]interface{}{}))},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: intPtr(5)}},
		},
		{
			"[]interface{} merge key zero",
			[]interface{}(nil),
			[]interface{}(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}(nil),
		},
		{
			"[]interface{} merge key mixed zero",
			[]interface{}{},
			[]interface{}(nil),
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed zero 2",
			[]interface{}(nil),
			[]interface{}{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key empty",
			[]interface{}{},
			[]interface{}{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{},
		},
		{
			"[]interface{} merge key mixed empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
			[]interface{}{},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]interface{} merge key mixed empty 2",
			[]interface{}{},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]interface{} merge key non empty",
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil},
			[]interface{}{&bar{FieldIntPtr: intPtr(2)}, &bar{FieldIntPtr: intPtr(4)}, &bar{FieldIntPtr: nil}, nil},
			[]DeepMergeOption{WithMergeByKeyFunc(reflect.TypeOf([]interface{}{}), barPtrInterfaceMergeFunc)},
			[]interface{}{&bar{FieldIntPtr: intPtr(1)}, &bar{FieldIntPtr: intPtr(2)}, nil, &bar{FieldIntPtr: intPtr(4)}},
		},
		{
			"[]bar default merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar(nil),
		},
		{
			"[]bar default merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index empty",
			[]bar{},
			[]bar{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{},
		},
		{
			"[]bar default merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar default merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar default merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar default merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]DeepMergeOption{WithDefaultMergeByIndex()},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index zero",
			[]bar(nil),
			[]bar(nil),
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar(nil),
		},
		{
			"[]bar merge by index mixed zero",
			[]bar{},
			[]bar(nil),
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed zero 2",
			[]bar(nil),
			[]bar{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index empty",
			[]bar{},
			[]bar{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{},
		},
		{
			"[]bar merge by index mixed empty",
			[]bar{{FieldIntPtr: intPtr(1)}},
			[]bar{},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]bar merge by index mixed empty 2",
			[]bar{},
			[]bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]bar merge by index non empty 1",
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}, {FieldIntPtr: intPtr(3)}},
		},
		{
			"[]bar merge by index non empty 2",
			[]bar{{FieldIntPtr: intPtr(4)}, {FieldIntPtr: intPtr(5)}},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
			[]DeepMergeOption{WithMergeByIndex(reflect.TypeOf([]bar{}))},
			[]bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}},
		},

		{
			"[]*bar append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar(nil),
		},
		{
			"[]*bar append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{},
		},
		{
			"[]*bar append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]DeepMergeOption{WithDefaultListAppend()},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]*bar custom append zero",
			[]*bar(nil),
			[]*bar(nil),
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar(nil),
		},
		{
			"[]*bar custom append mixed zero",
			[]*bar{},
			[]*bar(nil),
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed zero 2",
			[]*bar(nil),
			[]*bar{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append empty",
			[]*bar{},
			[]*bar{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{},
		},
		{
			"[]*bar custom append mixed empty",
			[]*bar{{FieldIntPtr: intPtr(1)}},
			[]*bar{},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}},
		},
		{
			"[]*bar custom append mixed empty 2",
			[]*bar{},
			[]*bar{{FieldIntPtr: intPtr(2)}},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(2)}},
		},
		{
			"[]*bar custom append non empty",
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil},
			[]*bar{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
			[]DeepMergeOption{WithListAppend(reflect.TypeOf([]*bar{}))},
			[]*bar{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, nil, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(4)}, {FieldIntPtr: nil}, nil},
		},
		{
			"[]int nil + nil w/ zero empty option",
			[]int(nil),
			[]int(nil),
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int(nil),
		},
		{
			"[]int nil + empty w/ zero empty option",
			[]int(nil),
			[]int{},
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + nil w/ zero empty option",
			[]int{},
			[]int(nil),
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + empty w/ zero empty option",
			[]int{},
			[]int{},
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int{},
		},
		{
			"[]int empty + non-empty w/ zero empty option",
			[]int{},
			[]int{1, 2, 3},
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int{1, 2, 3},
		},
		{
			"[]int non-empty + empty w/ zero empty option",
			[]int{1, 2, 3},
			[]int{},
			[]DeepMergeOption{WithZeroEmptySlice()},
			[]int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deepMerge := NewDeepMergeFunc(tt.opt...)
			got, err := deepMerge(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Interface())
		})
	}
	t.Run("fallback error", func(t *testing.T) {
		deepMerge := NewDeepMergeFunc(WithTypeMerger(reflect.TypeOf([]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		_, err := deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{2}))
		assert.EqualError(t, err, "fake")
	})
	t.Run("merge key func errors", func(t *testing.T) {
		deepMerge := NewDeepMergeFunc(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, nil
		}))
		_, err := deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		_, err = deepMerge(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key func returned nil")
		deepMerge = NewDeepMergeFunc(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.ValueOf([]int{1, 2, 3}), nil
		}))
		_, err = deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		_, err = deepMerge(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "slice merge key [1 2 3] of type []int is not comparable")
		deepMerge = NewDeepMergeFunc(WithMergeByKeyFunc(reflect.TypeOf([]int{}), func(int, reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("merge key func error")
		}))
		_, err = deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{}))
		assert.EqualError(t, err, "merge key func error")
		_, err = deepMerge(reflect.ValueOf([]int{}), reflect.ValueOf([]int{1}))
		assert.EqualError(t, err, "merge key func error")
	})
}
