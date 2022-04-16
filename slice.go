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

// SliceCoalescerOption is an option to be passed to NewSliceCoalescer.
type SliceCoalescerOption func(c *sliceCoalescer)

// NewSliceCoalescer creates a new Coalescer for slices.
func NewSliceCoalescer(opts ...SliceCoalescerOption) Coalescer {
	sc := &sliceCoalescer{
		defaultCoalescer: &defaultCoalescer{},
	}
	for _, opt := range opts {
		opt(sc)
	}
	return sc
}

// WithDefaultSetUnion applies set-union semantics to all slices to be coalesced.
func WithDefaultSetUnion() SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		c.defaultCoalescer = &sliceMergeCoalescer{
			mergeKeyFunc: SliceUnion,
		}
	}
}

// WithDefaultListAppend applies list-append semantics to all slices to be coalesced.
func WithDefaultListAppend() SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		c.defaultCoalescer = &sliceAppendCoalescer{}
	}
}

// WithSetUnion applies set-union semantics to the given slice element type. If the slice element type is a pointer
// type, the passed argument must be that pointer type, not its target type.
func WithSetUnion(elemType reflect.Type) SliceCoalescerOption {
	return WithMergeByKey(elemType, SliceUnion)
}

// WithListAppend applies list-append semantics to the given slice element type. If the slice element type is a pointer
// type, the passed argument must be that pointer type, not its target type.
func WithListAppend(elem reflect.Type) SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		if c.elemCoalescers == nil {
			c.elemCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.elemCoalescers[elem] = &sliceAppendCoalescer{}
	}
}

// WithMergeByField applies merge-by semantics to slices whose elements are of the passed struct type, or a pointer
// thereto. The passed field name will be used to extract the element merge key; therefore, the field should generally
// be a unique identifier or primary key for objects of this type.
func WithMergeByField(structType reflect.Type, field string) SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		f := newMergeByField(field)
		WithMergeByKey(structType, f)(c)
		WithMergeByKey(reflect.PtrTo(structType), f)(c)
	}
}

// WithMergeByKey applies merge-by semantics to the given slice element type. The given mergeKeyFunc will be used to
// extract the element merge key. If the slice element type is a pointer type, the passed type argument must be that
// pointer type, not its target type.
func WithMergeByKey(elemType reflect.Type, f SliceMergeKeyFunc) SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		if c.elemCoalescers == nil {
			c.elemCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.elemCoalescers[elemType] = &sliceMergeCoalescer{
			mergeKeyFunc: f,
		}
	}
}

type sliceCoalescer struct {
	defaultCoalescer Coalescer
	elemCoalescers   map[reflect.Type]Coalescer
}

func (c *sliceCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Slice); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	elemType := v1.Type().Elem()
	if coalescer, found := c.elemCoalescers[elemType]; found {
		return coalescer.Coalesce(v1, v2)
	}
	return c.defaultCoalescer.Coalesce(v1, v2)
}

func (c *sliceCoalescer) WithFallback(fallback Coalescer) {
	for _, delegate := range c.elemCoalescers {
		delegate.WithFallback(fallback)
	}
	c.defaultCoalescer.WithFallback(fallback)
}

type sliceAppendCoalescer struct{}

func (c *sliceAppendCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Slice); err != nil {
		return reflect.Value{}, err
	}
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

func (c *sliceAppendCoalescer) WithFallback(Coalescer) {}

// SliceMergeKeyFunc is a function that extracts a merge key from a slice element. The passed element may be the zero
// value for the slice element type, but it will never be an invalid value. The returned merge key can be a zero value,
// but cannot be invalid; moreover, it must be comparable as it will be stored internally in a temporary map during the
// merge.
type SliceMergeKeyFunc func(reflect.Value) reflect.Value

// SliceUnion is a merge key func that returns the elements themselves as keys, thus achieving set-union semantics. If
// the elements are pointers, they are dereferenced, which means that the set-union semantics will apply to the pointer
// targets, not to the pointers themselves. When using this func to do slice merges, the resulting slices will have no
// duplicate items (that is, items having the same merge key).
var SliceUnion SliceMergeKeyFunc = safeIndirect

// newMergeByField returns a SliceMergeKeyFunc that returns the value of the given struct field for each slice element.
// This function is designed to work on slices of structs, and slices of pointers to structs. When this function
// encounters a pointer while extracting the merge key, it dereferences the pointer; if the pointer was nil, a zero
// value will be used instead, but beware that this may result in nondeterministic merge results.
func newMergeByField(field string) SliceMergeKeyFunc {
	return func(elem reflect.Value) reflect.Value {
		// the slice element itself may be a pointer; we want to dereference it and return a zero value if it's nil.
		elem = safeIndirect(elem)
		// the slice element's field may also be a pointer; again, we want to dereference it and return a zero value
		// if it's nil.
		return safeIndirect(elem.FieldByName(field))
	}
}

type sliceMergeCoalescer struct {
	fallback     Coalescer
	mergeKeyFunc SliceMergeKeyFunc
}

func (c *sliceMergeCoalescer) WithFallback(fallback Coalescer) {
	c.fallback = fallback
}

var typeOfInterface = reflect.TypeOf((*interface{})(nil)).Elem()

func (c *sliceMergeCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Slice); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	// the keys slice allows to keep a deterministic element order in the resulting slice
	keys := reflect.MakeSlice(reflect.SliceOf(typeOfInterface), 0, 0)
	m1 := reflect.MakeMap(reflect.MapOf(typeOfInterface, v1.Type().Elem()))
	for i := 0; i < v1.Len(); i++ {
		v := v1.Index(i)
		k := c.mergeKeyFunc(v)
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
		k := c.mergeKeyFunc(v)
		if err := checkMergeKey(k); err != nil {
			return reflect.Value{}, err
		}
		if !m1.MapIndex(k).IsValid() && !m2.MapIndex(k).IsValid() {
			keys = reflect.Append(keys, k)
		}
		m2.SetMapIndex(k, v)
	}
	m, err := c.fallback.Coalesce(m1, m2)
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
