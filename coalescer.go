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

import "reflect"

// coalescer is the engine for merging and copying objets. It has two methods that satisfy DeepMergeFunc and
// DeepCopyFunc: deepMerge and deepCopy respectively.
type coalescer struct {
	typeMergers    map[reflect.Type]DeepMergeFunc
	sliceMerger    DeepMergeFunc
	sliceMergers   map[ /* slice type */ reflect.Type]DeepMergeFunc
	fieldMergers   map[ /* struct type */ reflect.Type]map[ /* field name */ string]DeepMergeFunc
	zeroEmptySlice bool
	errorOnCycle   bool
	seen           map[uintptr]bool
}

func (c *coalescer) deepMerge(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatch(v1, v2); err != nil {
		return reflect.Value{}, err
	}
	if merger, found := c.typeMergers[v1.Type()]; found {
		value, err := merger(v1, v2)
		if value.IsValid() || err != nil {
			return value, err
		}
	}
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	switch v1.Type().Kind() {
	case reflect.Interface:
		return c.deepMergeInterface(v1, v2)
	case reflect.Ptr:
		return c.deepMergePointer(v1, v2)
	case reflect.Map:
		return c.deepMergeMap(v1, v2)
	case reflect.Struct:
		return c.deepMergeStruct(v1, v2)
	case reflect.Slice:
		return c.deepMergeSlice(v1, v2)
	}
	return c.deepMergeAtomic(v1, v2)
}

func (c *coalescer) deepCopy(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	switch v.Type().Kind() {
	case reflect.Interface:
		return c.deepCopyInterface(v)
	case reflect.Ptr:
		return c.deepCopyPointer(v)
	case reflect.Map:
		return c.deepCopyMap(v)
	case reflect.Struct:
		return c.deepCopyStruct(v)
	case reflect.Slice:
		return c.deepCopySlice(v)
	case reflect.Array:
		return c.deepCopyArray(v)
	}
	return v, nil
}
