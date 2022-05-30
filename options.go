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

// CoalescerOption is an option that can be passed the Coalesce function to customize its coalescing behavior.
type CoalescerOption func(c *mainCoalescer)

// WithAtomicType causes the given type to be coalesced atomically, that is, with  "atomic" semantics, instead of its
// default coalesce semantics. When 2 non-zero-values of this type are coalesced, the second value is returned as is.
func WithAtomicType(t reflect.Type) CoalescerOption {
	return WithTypeCoalescer(t, coalesceAtomic)
}

// WithTrileans causes all boolean pointers to be coalesced using a three-valued logic, instead of their default
// coalesce semantics. When this is enabled, boolean pointers will behave as if they were "trileans", that is, a type
// with 3 possible values: nil (its zero-value), false and true (contrary to booleans, with trileans false is NOT a
// zero-value).
// The coalescing of trileans obeys the following rules:
//   v1    v2    coalesced
//   nil   nil   nil
//   nil   false false
//   nil   true  true
//   false nil   false
//   false false false
//   false true  true
//   true  nil   true
//   true  false false
//   true  true  true
// The biggest difference with regular boolean pointers is that Coalesce(&true, &false) will return &true for boolean
// pointers, while with trileans, it will return &false.
func WithTrileans() CoalescerOption {
	return WithAtomicType(reflect.PtrTo(reflect.TypeOf(false)))
}

// WithTypeCoalescer will defer the coalescing of the given type to the given custom coalescer. This option does not
// allow the type coalescer to access the parent Coalescer instance being created. For that, use
// WithTypeCoalescerProvider instead.
func WithTypeCoalescer(t reflect.Type, coalescer Coalescer) CoalescerOption {
	return WithTypeCoalescerProvider(t, func(parent Coalescer) Coalescer {
		return coalescer
	})
}

// WithTypeCoalescerProvider will defer the coalescing of the given type to a custom coalescer that will be obtained by
// calling the given provider function with the parent Coalescer instance being created.
func WithTypeCoalescerProvider(t reflect.Type, provider func(parent Coalescer) Coalescer) CoalescerOption {
	return func(c *mainCoalescer) {
		if c.typeCoalescers == nil {
			c.typeCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.typeCoalescers[t] = provider(c.coalesce)
	}
}

// WithErrorOnCycle instructs the coalescer to return an error when a cycle is detected. By default, the coalescer
// replaces cycles with a nil pointer.
func WithErrorOnCycle() CoalescerOption {
	return func(c *mainCoalescer) {
		c.errorOnCycle = true
	}
}

// WithDefaultSetUnion applies set-union semantics to all slices to be coalesced. When the slice elements are pointers,
// this strategy dereferences the pointers and compare their targets. This strategy is fine for slices of scalars and
// pointers thereof, but it is not recommended for slices of complex types as the elements may not be fully comparable.
func WithDefaultSetUnion() CoalescerOption {
	return func(c *mainCoalescer) {
		c.sliceCoalescer = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.coalesceSliceMerge(v1, v2, SliceUnion)
		}
	}
}

// WithDefaultListAppend applies list-append semantics to all slices to be coalesced.
func WithDefaultListAppend() CoalescerOption {
	return func(c *mainCoalescer) {
		c.sliceCoalescer = coalesceSliceAppend
	}
}

// WithDefaultMergeByIndex applies merge-by-index semantics to all slices to be coalesced.
func WithDefaultMergeByIndex() CoalescerOption {
	return func(c *mainCoalescer) {
		c.sliceCoalescer = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.coalesceSliceMerge(v1, v2, SliceIndex)
		}
	}
}

// WithSetUnion applies set-union semantics to the given slice type. When the slice elements are of a pointer type, this
// strategy dereferences the pointers and compare their targets. This strategy is fine for slices of scalars and
// pointers thereof, but it is not recommended for slices of complex types as the elements may not be fully comparable.
func WithSetUnion(sliceType reflect.Type) CoalescerOption {
	return WithMergeByKey(sliceType, SliceUnion)
}

// WithListAppend applies list-append semantics to the given slice type.
func WithListAppend(sliceType reflect.Type) CoalescerOption {
	return func(c *mainCoalescer) {
		if c.sliceCoalescers == nil {
			c.sliceCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.sliceCoalescers[sliceType] = coalesceSliceAppend
	}
}

// WithMergeByKey applies merge-by-key semantics to the given slice type. The given mergeKeyFunc will be used to extract
// the element merge key.
func WithMergeByKey(sliceType reflect.Type, f SliceMergeKeyFunc) CoalescerOption {
	return func(c *mainCoalescer) {
		if c.sliceCoalescers == nil {
			c.sliceCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.sliceCoalescers[sliceType] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.coalesceSliceMerge(v1, v2, f)
		}
	}
}

// WithMergeByIndex applies merge-by-index semantics to the given slice type. The given mergeKeyFunc will be used to
// extract the element merge key.
func WithMergeByIndex(sliceType reflect.Type) CoalescerOption {
	return WithMergeByKey(sliceType, SliceIndex)
}

// WithZeroEmptySlice instructs the coalescer to consider empty slices as zero (nil) slices. This changes the default
// behavior: when coalescing a non-empty slice with an empty slice, normally the empty slice is returned, but with this
// option, the non-empty slice is returned.
func WithZeroEmptySlice() CoalescerOption {
	return func(c *mainCoalescer) {
		c.zeroEmptySlice = true
	}
}

// WithMergeByField applies merge-by-key semantics to slices whose elements are of some struct type, or a pointer
// thereto. The passed field name will be used to extract the element's merge key; therefore, the field should generally
// be a unique identifier or primary key for objects of this type.
func WithMergeByField(sliceOfStructType reflect.Type, field string) CoalescerOption {
	return func(c *mainCoalescer) {
		WithMergeByKey(sliceOfStructType, newMergeByField(field))(c)
	}
}

// WithFieldCoalescer coalesces the given struct field with the given custom coalescer. This option does not allow the
// type coalescer to access the parent Coalescer instance being created. For that, use WithFieldCoalescerProvider
// instead.
func WithFieldCoalescer(structType reflect.Type, field string, coalescer Coalescer) CoalescerOption {
	return WithFieldCoalescerProvider(structType, field, func(parent Coalescer) Coalescer {
		return coalescer
	})
}

// WithFieldCoalescerProvider coalesces the given struct field with a custom coalescer that will be obtained by calling
// the given provider function with the Coalescer instance being created.
func WithFieldCoalescerProvider(structType reflect.Type, field string, provider func(parent Coalescer) Coalescer) CoalescerOption {
	return func(c *mainCoalescer) {
		if c.fieldCoalescers == nil {
			c.fieldCoalescers = make(map[reflect.Type]map[string]Coalescer)
		}
		if c.fieldCoalescers[structType] == nil {
			c.fieldCoalescers[structType] = make(map[string]Coalescer)
		}
		c.fieldCoalescers[structType][field] = provider(c.coalesce)
	}
}

// WithAtomicField causes the given field to be coalesced atomically, that is, with  "atomic" semantics, instead of its
// default coalesce semantics. When 2 non-zero-values of this field are coalesced, the second value is returned as is.
func WithAtomicField(structType reflect.Type, field string) CoalescerOption {
	return WithFieldCoalescer(structType, field, coalesceAtomic)
}
