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
		defaultCoalescer: &atomicCoalescer{},
	}
	for _, opt := range opts {
		opt(sc)
	}
	return sc
}

// WithDefaultSetUnion applies set-union semantics to all slices to be coalesced. When the slice elements are pointers,
// this strategy dereferences the pointers and compare their targets. This strategy is fine for slices of scalars and
// pointers thereof, but it is not recommended for slices of complex types as the elements may not be fully comparable.
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

// WithDefaultMergeByIndex applies merge-by-index semantics to all slices to be coalesced.
func WithDefaultMergeByIndex() SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		c.defaultCoalescer = &sliceMergeCoalescer{
			mergeKeyFunc: SliceIndex,
		}
	}
}

// WithSetUnion applies set-union semantics to the given slice element type. When the slice elements are of a pointer
// type, this strategy dereferences the pointers and compare their targets; also, the passed argument must be that
// pointer type, not its target type. This strategy is fine for slices of scalars and pointers thereof, but it is not
// recommended for slices of complex types as the elements may not be fully comparable.
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

// WithMergeByField applies merge-by-key semantics to slices whose elements are of the passed struct type, or a pointer
// thereto. The passed field name will be used to extract the element merge key; therefore, the field should generally
// be a unique identifier or primary key for objects of this type.
func WithMergeByField(structType reflect.Type, field string) SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		f := newMergeByField(field)
		WithMergeByKey(structType, f)(c)
		WithMergeByKey(reflect.PtrTo(structType), f)(c)
	}
}

// WithMergeByKey applies merge-by-key semantics to the given slice element type. The given mergeKeyFunc will be used to
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

// WithMergeByIndex applies merge-by-index semantics to the given slice element type. The given mergeKeyFunc will be
// used to extract the element merge key. If the slice element type is a pointer type, the passed type argument must be
// that pointer type, not its target type.
func WithMergeByIndex(elemType reflect.Type) SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		if c.elemCoalescers == nil {
			c.elemCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.elemCoalescers[elemType] = &sliceMergeCoalescer{
			mergeKeyFunc: SliceIndex,
		}
	}
}

// WithZeroEmptySlice instructs the coalescer to consider empty slices as zero (nil) slices. This changes the default
// behavior: when coalescing a non-empty slice with an empty slice, normally the empty slice is returned, but with this
// option, the non-empty slice is returned.
func WithZeroEmptySlice() SliceCoalescerOption {
	return func(c *sliceCoalescer) {
		c.zeroEmptySlice = true
	}
}

type sliceCoalescer struct {
	defaultCoalescer Coalescer
	elemCoalescers   map[reflect.Type]Coalescer
	zeroEmptySlice   bool
}

func (c *sliceCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Slice); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
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

// newMergeByField returns a SliceMergeKeyFunc that returns the value of the given struct field for each slice element.
// This function is designed to work on slices of structs, and slices of pointers to structs. When this function
// encounters a pointer while extracting the merge key, it dereferences the pointer; if the pointer was nil, a zero
// value will be used instead, but beware that this may result in nondeterministic merge results.
func newMergeByField(field string) SliceMergeKeyFunc {
	return func(_ int, elem reflect.Value) reflect.Value {
		// the slice element itself may be a pointer; we want to dereference it and return a zero-value if it's nil.
		elem = safeIndirect(elem)
		// the slice element's field may also be a pointer; again, we want to dereference it and return a zero-value
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
		k := c.mergeKeyFunc(i, v)
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
		k := c.mergeKeyFunc(i, v)
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
