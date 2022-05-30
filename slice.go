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
	"fmt"
	"reflect"
)

// SliceMergeKeyFunc is a function that extracts a merge key from a slice element's index and value. The passed element
// may be the zero-value for the slice element type, but it will never be an invalid value. The returned merge key can
// be a zero-value, but cannot be invalid; moreover, it must be comparable as it will be stored internally in a
// temporary map during the merge.
type SliceMergeKeyFunc func(index int, element reflect.Value) (key reflect.Value)

// SliceUnion is a merge key func that returns the elements themselves as keys, thus achieving set-union semantics. If
// the elements are pointers, they are dereferenced, which means that the set-union semantics will apply to the pointer
// targets, not to the pointers themselves. When using this func to do slice merges, the resulting slices will have no
// duplicate items (that is, items having the same merge key).
var SliceUnion SliceMergeKeyFunc = func(index int, element reflect.Value) (key reflect.Value) {
	return safeIndirect(element)
}

// SliceIndex is a merge key func that returns the element indices as keys, thus achieving merge-by-index semantics.
// When using this func to do slice merges, the resulting slices will have their elements coalesced index by index.
var SliceIndex SliceMergeKeyFunc = func(index int, element reflect.Value) (key reflect.Value) {
	return reflect.ValueOf(index)
}

func (c *mainCoalescer) coalesceSlice(v1, v2 reflect.Value) (reflect.Value, error) {
	if v1.Len() == 0 && v2.Len() == 0 {
		return v2, nil
	}
	if c.zeroEmptySlice {
		if v1.Len() == 0 {
			v1 = reflect.Zero(v1.Type())
		}
		if v2.Len() == 0 {
			v2 = reflect.Zero(v2.Type())
		}
	}
	if coalescer, found := c.sliceCoalescers[v1.Type()]; found {
		return coalescer(v1, v2)
	}
	if c.sliceCoalescer != nil {
		return c.sliceCoalescer(v1, v2)
	}
	return coalesceAtomic(v1, v2)
}

func coalesceSliceAppend(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	l := v1.Len() + v2.Len()
	coalesced := reflect.MakeSlice(v1.Type(), l, l)
	for i := 0; i < v1.Len(); i++ {
		coalesced.Index(i).Set(v1.Index(i))
	}
	for i := 0; i < v2.Len(); i++ {
		coalesced.Index(v1.Len() + i).Set(v2.Index(i))
	}
	return coalesced, nil
}

var typeOfInterface = reflect.TypeOf((*interface{})(nil)).Elem()

func (c *mainCoalescer) coalesceSliceMerge(v1, v2 reflect.Value, mergeKeyFunc SliceMergeKeyFunc) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	// the keys slice allows to keep a deterministic element order in the resulting slice
	keys := reflect.MakeSlice(reflect.SliceOf(typeOfInterface), 0, 0)
	m1 := reflect.MakeMap(reflect.MapOf(typeOfInterface, v1.Type().Elem()))
	for i := 0; i < v1.Len(); i++ {
		v := v1.Index(i)
		k := mergeKeyFunc(i, v)
		if err := checkMergeKey(k); err != nil {
			return reflect.Value{}, err
		}
		if !m1.MapIndex(k).IsValid() {
			keys = reflect.Append(keys, k)
		}
		m1.SetMapIndex(k, v)
	}
	m2 := reflect.MakeMap(reflect.MapOf(typeOfInterface, v2.Type().Elem()))
	for i := 0; i < v2.Len(); i++ {
		v := v2.Index(i)
		k := mergeKeyFunc(i, v)
		if err := checkMergeKey(k); err != nil {
			return reflect.Value{}, err
		}
		if !m1.MapIndex(k).IsValid() && !m2.MapIndex(k).IsValid() {
			keys = reflect.Append(keys, k)
		}
		m2.SetMapIndex(k, v)
	}
	m, err := c.coalesce(m1, m2)
	if err != nil {
		return reflect.Value{}, err
	}
	coalesced := reflect.MakeSlice(v1.Type(), 0, 0)
	for i := 0; i < keys.Len(); i++ {
		k := keys.Index(i)
		if m.MapIndex(k).IsValid() {
			coalesced = reflect.Append(coalesced, m.MapIndex(k))
		}
	}
	return coalesced, nil
}

func checkMergeKey(k reflect.Value) error {
	if !k.IsValid() {
		return fmt.Errorf("slice merge key func returned nil")
	} else if !k.Type().Comparable() {
		return fmt.Errorf("slice merge key %v of type %T is not comparable", k.Interface(), k.Interface())
	}
	return nil
}
