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
type SliceMergeKeyFunc func(index int, element reflect.Value) (key reflect.Value, err error)

// SliceUnion is a merge key func that returns the elements themselves as keys, thus achieving set-union semantics. If
// the elements are pointers, they are dereferenced, which means that the set-union semantics will apply to the pointer
// targets, not to the pointers themselves. When using this func to do slice merges, the resulting slices will have no
// duplicate items (that is, items having the same merge key).
var SliceUnion SliceMergeKeyFunc = func(index int, element reflect.Value) (key reflect.Value, err error) {
	return safeIndirect(element), nil
}

// SliceIndex is a merge key func that returns the element indices as keys, thus achieving merge-by-index semantics.
// When using this func to do slice merges, the resulting slices will have their elements coalesced index by index.
var SliceIndex SliceMergeKeyFunc = func(index int, element reflect.Value) (key reflect.Value, err error) {
	return reflect.ValueOf(index), nil
}

// deepMergeSlice is the default slice merger. It first checks if there is a custom slice merger
// registered for the slice type. If there is, it uses it. Otherwise, it uses the default slice
// merge strategy, which is atomic.
func (c *coalescer) deepMergeSlice(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	if v1.Len() == 0 && v2.Len() == 0 {
		return c.deepCopy(v2)
	}
	if c.zeroEmptySlice {
		if v1.Len() == 0 {
			v1 = reflect.Zero(v1.Type())
		}
		if v2.Len() == 0 {
			v2 = reflect.Zero(v2.Type())
		}
		if value, done := checkZero(v1, v2); done {
			return c.deepCopy(value)
		}
	}
	if sliceMerger, found := c.sliceMergers[v1.Type()]; found {
		return sliceMerger(v1, v2)
	}
	if c.sliceMerger != nil {
		return c.sliceMerger(v1, v2)
	}
	return c.deepMergeAtomic(v1, v2)
}

// deepMergeSliceWithListAppend is an alternate slice merger that appends the elements of the second
// slice to the first slice. It is not the default merge strategy for slices; it is only activated
// if a slice merger has been registered through one of the options:
// WithDefaultSliceListAppendMerge, WithSliceListAppendMerge or WithFieldListAppendMerge.
func (c *coalescer) deepMergeSliceWithListAppend(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	if v1.Len() == 0 && v2.Len() == 0 {
		return c.deepCopy(v2)
	}
	l := v1.Len() + v2.Len()
	merged := reflect.MakeSlice(v1.Type(), l, l)
	for i := 0; i < v1.Len(); i++ {
		elem, err := c.deepCopy(v1.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		merged.Index(i).Set(elem)
	}
	for i := 0; i < v2.Len(); i++ {
		elem, err := c.deepCopy(v2.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		merged.Index(v1.Len() + i).Set(elem)
	}
	return merged, nil
}

var typeOfInterface = reflect.TypeOf((*interface{})(nil)).Elem()

// deepMergeSliceWithMergeKey is an alternate slice merger that merges the elements of the two
// slices using a merge key function. It is not the default merge strategy for slices; it is only
// activated if a slice merger has been registered through one of the options:
// WithDefaultSliceSetUnionMerge, WithDefaultSliceMergeByIndex, WithSliceSetUnionMerge,
// WithSliceMergeByIndex, WithSliceMergeByID, WithSliceMergeByKeyFunc, WithFieldMergeByIndex,
// WithFieldMergeByID, WithFieldMergeByKeyFunc.
func (c *coalescer) deepMergeSliceWithMergeKey(v1, v2 reflect.Value, mergeKeyFunc SliceMergeKeyFunc) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	if v1.Len() == 0 && v2.Len() == 0 {
		return c.deepCopy(v2)
	}
	// The "keys" slice allows to keep a deterministic element order in the resulting slice.
	keys := reflect.MakeSlice(reflect.SliceOf(typeOfInterface), 0, 0)
	m1 := reflect.MakeMap(reflect.MapOf(typeOfInterface, v1.Type().Elem()))
	for i := 0; i < v1.Len(); i++ {
		v := v1.Index(i)
		k, err := mergeKeyFunc(i, v)
		if err != nil {
			return reflect.Value{}, err
		} else if err := checkMergeKey(k); err != nil {
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
		k, err := mergeKeyFunc(i, v)
		if err != nil {
			return reflect.Value{}, err
		} else if err := checkMergeKey(k); err != nil {
			return reflect.Value{}, err
		}
		if !m1.MapIndex(k).IsValid() && !m2.MapIndex(k).IsValid() {
			keys = reflect.Append(keys, k)
		}
		m2.SetMapIndex(k, v)
	}
	// Note: we can't call deepMergeMap here because it is important to NOT copy the merge keys
	m := reflect.MakeMap(m1.Type())
	for _, k := range m1.MapKeys() {
		if !m2.MapIndex(k).IsValid() {
			copiedValue, err := c.deepCopy(m1.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			m.SetMapIndex(k, copiedValue)
		}
	}
	for _, k := range m2.MapKeys() {
		if m1.MapIndex(k).IsValid() {
			mergedValue, err := c.deepMerge(m1.MapIndex(k), m2.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			m.SetMapIndex(k, mergedValue)
		} else {
			copiedValue, err := c.deepCopy(m2.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			m.SetMapIndex(k, copiedValue)
		}
	}
	merged := reflect.MakeSlice(v1.Type(), 0, 0)
	for i := 0; i < keys.Len(); i++ {
		k := keys.Index(i)
		if m.MapIndex(k).IsValid() {
			merged = reflect.Append(merged, m.MapIndex(k))
		}
	}
	return merged, nil
}

func checkMergeKey(k reflect.Value) error {
	if !k.IsValid() {
		return fmt.Errorf("slice merge key func returned nil")
	} else if !k.Type().Comparable() {
		return fmt.Errorf("slice merge key %v of type %T is not comparable", k.Interface(), k.Interface())
	}
	return nil
}

func (c *coalescer) deepCopySlice(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	copied := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	for i := 0; i < v.Len(); i++ {
		elem, err := c.deepCopy(v.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		copied.Index(i).Set(elem)
	}
	return copied, nil
}
