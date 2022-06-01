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

// DeepMergeOption is an option that can be passed to DeepMerge to customize the function's merge behavior.
type DeepMergeOption func(c *coalescer)

// DeepCopyOption is an option that can be passed to DeepCopy to customize the function's merge behavior.
type DeepCopyOption func(c *coalescer)

// COMMON OPTIONS

// WithErrorOnCycle instructs the operation to return an error when a cycle is detected. By default, cycles are replaced
// with a nil pointer.
func WithErrorOnCycle() func(c *coalescer) {
	return func(c *coalescer) {
		c.errorOnCycle = true
	}
}

// DEEP COPY ONLY OPTIONS

// WithAtomicType causes the given type to be merged atomically, that is, with  "atomic" semantics, instead of its
// default merge semantics. When 2 non-zero-values of this type are merged, the second value is returned as is.
func WithAtomicType(t reflect.Type) DeepMergeOption {
	return func(c *coalescer) {
		if c.typeMergers == nil {
			c.typeMergers = make(map[reflect.Type]DeepMergeFunc)
		}
		c.typeMergers[t] = c.deepMergeAtomic
	}
}

// WithTrileans causes all boolean pointers to be merged using a three-valued logic, instead of their default merge
// semantics. When this is enabled, boolean pointers will behave as if they were "trileans", that is, a type with 3
// possible values: nil (its zero-value), false and true (contrary to booleans, with trileans false is NOT a
// zero-value).
// The merge of trileans obeys the following rules:
//   v1    v2    merged
//   nil   nil   nil
//   nil   false false
//   nil   true  true
//   false nil   false
//   false false false
//   false true  true
//   true  nil   true
//   true  false false
//   true  true  true
// The biggest difference with regular boolean pointers is that DeepMerge(&true, &false) will return &true for boolean
// pointers, while with trileans, it will return &false.
func WithTrileans() DeepMergeOption {
	return WithAtomicType(reflect.PtrTo(reflect.TypeOf(false)))
}

// WithTypeMerger will defer the merge of the given type to the given custom merger. This option does not allow the type
// merger to access the parent DeepMergeFunc instance being created. For that, use WithTypeMergerProvider instead.
func WithTypeMerger(t reflect.Type, merger DeepMergeFunc) DeepMergeOption {
	return WithTypeMergerProvider(t, func(parent DeepMergeFunc) DeepMergeFunc {
		return merger
	})
}

// WithTypeMergerProvider will defer the merge of the given type to a custom merger that will be obtained by calling the
// given provider function with the parent DeepMergeFunc instance being created.
func WithTypeMergerProvider(t reflect.Type, provider func(parent DeepMergeFunc) DeepMergeFunc) DeepMergeOption {
	return func(c *coalescer) {
		if c.typeMergers == nil {
			c.typeMergers = make(map[reflect.Type]DeepMergeFunc)
		}
		c.typeMergers[t] = provider(c.deepMerge)
	}
}

// WithZeroEmptySlice instructs the merger to consider empty slices as zero (nil) slices. This changes the default
// behavior: when merge a non-empty slice with an empty slice, normally the empty slice is returned, but with this
// option, the non-empty slice is returned.
func WithZeroEmptySlice() DeepMergeOption {
	return func(c *coalescer) {
		c.zeroEmptySlice = true
	}
}

// WithDefaultListAppend applies list-append semantics to all slices to be merged.
func WithDefaultListAppend() DeepMergeOption {
	return func(c *coalescer) {
		c.sliceMerger = c.deepSliceAppend
	}
}

// WithDefaultSetUnion applies set-union semantics to all slices to be merged. When the slice elements are pointers,
// this strategy dereferences the pointers and compare their targets. This strategy is fine for slices of scalars and
// pointers thereof, but it is not recommended for slices of complex types as the elements may not be fully comparable.
func WithDefaultSetUnion() DeepMergeOption {
	return func(c *coalescer) {
		c.sliceMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepSliceMerge(v1, v2, SliceUnion)
		}
	}
}

// WithDefaultMergeByIndex applies merge-by-index semantics to all slices to be merged.
func WithDefaultMergeByIndex() DeepMergeOption {
	return func(c *coalescer) {
		c.sliceMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepSliceMerge(v1, v2, SliceIndex)
		}
	}
}

// WithSetUnion applies set-union semantics to the given slice type. When the slice elements are of a pointer type, this
// strategy dereferences the pointers and compare their targets. This strategy is fine for slices of scalars and
// pointers thereof, but it is not recommended for slices of complex types as the elements may not be fully comparable.
func WithSetUnion(sliceType reflect.Type) DeepMergeOption {
	return WithMergeByKeyFunc(sliceType, SliceUnion)
}

// WithListAppend applies list-append semantics to the given slice type.
func WithListAppend(sliceType reflect.Type) DeepMergeOption {
	return func(c *coalescer) {
		if c.sliceMergers == nil {
			c.sliceMergers = make(map[reflect.Type]DeepMergeFunc)
		}
		c.sliceMergers[sliceType] = c.deepSliceAppend
	}
}

// WithMergeByIndex applies merge-by-index semantics to the given slice type. The given mergeKeyFunc will be used to
// extract the element merge key.
func WithMergeByIndex(sliceType reflect.Type) DeepMergeOption {
	return WithMergeByKeyFunc(sliceType, SliceIndex)
}

// WithMergeByID applies merge-by-key semantics to slices whose elements are of some struct type, or a pointer thereto.
// The passed field name will be used to extract the element's merge key; therefore, the field should generally be a
// unique identifier or primary key for objects of this type.
func WithMergeByID(sliceOfStructType reflect.Type, elemField string) DeepMergeOption {
	return func(c *coalescer) {
		WithMergeByKeyFunc(sliceOfStructType, newMergeByField(elemField))(c)
	}
}

// WithMergeByKeyFunc applies merge-by-key semantics to the given slice type. The given SliceMergeKeyFunc will be used
// to extract the element merge key.
func WithMergeByKeyFunc(sliceType reflect.Type, mergeKeyFunc SliceMergeKeyFunc) DeepMergeOption {
	return func(c *coalescer) {
		if c.sliceMergers == nil {
			c.sliceMergers = make(map[reflect.Type]DeepMergeFunc)
		}
		c.sliceMergers[sliceType] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepSliceMerge(v1, v2, mergeKeyFunc)
		}
	}
}

// WithFieldMerger merges the given struct field with the given custom merger. This option does not allow the type
// merger to access the parent DeepMergeFunc instance being created. For that, use WithFieldMergerProvider instead.
func WithFieldMerger(structType reflect.Type, field string, merger DeepMergeFunc) DeepMergeOption {
	return WithFieldMergerProvider(structType, field, func(parent DeepMergeFunc) DeepMergeFunc {
		return merger
	})
}

// WithFieldMergerProvider merges the given struct field with a custom merger that will be obtained by calling the given
// provider function with the DeepMergeFunc instance being created.
func WithFieldMergerProvider(structType reflect.Type, field string, provider func(parent DeepMergeFunc) DeepMergeFunc) DeepMergeOption {
	return func(c *coalescer) {
		if c.fieldMergers == nil {
			c.fieldMergers = make(map[reflect.Type]map[string]DeepMergeFunc)
		}
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = provider(c.deepMerge)
	}
}

// WithFieldListAppend merges the given struct field with list-append semantics. The field must be of slice type. This
// is the programmatic equivalent of adding a `goalesce:append` struct tag to that field.
func WithFieldListAppend(structType reflect.Type, field string) DeepMergeOption {
	return func(c *coalescer) {
		if c.fieldMergers == nil {
			c.fieldMergers = make(map[reflect.Type]map[string]DeepMergeFunc)
		}
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = c.deepSliceAppend
	}
}

// WithFieldSetUnion merges the given struct field with set-union semantics. The field must be of slice type. This is
// the programmatic equivalent of adding a `goalesce:union` struct tag to that field.
func WithFieldSetUnion(structType reflect.Type, field string) DeepMergeOption {
	return WithFieldMergeByKeyFunc(structType, field, SliceUnion)
}

// WithFieldMergeByIndex merges the given struct field with merge-by-index semantics. The field must be of slice
// type. This is the programmatic equivalent of adding a `goalesce:index` struct tag to that field.
func WithFieldMergeByIndex(structType reflect.Type, field string) DeepMergeOption {
	return WithFieldMergeByKeyFunc(structType, field, SliceIndex)
}

// WithFieldMergeByID merges the given struct field with merge-by-key semantics. The field must be of slice type.
// The slice element type must be of some other struct type, or a pointer thereto. The passed key must be a valid field
// name for that struct type and will be used to extract the slice element's merge key; therefore, that field should
// generally be a unique identifier or primary key for objects of this type. This is the programmatic equivalent of
// adding a `goalesce:id:key` struct tag to the struct field.
func WithFieldMergeByID(structType reflect.Type, field string, key string) DeepMergeOption {
	return WithFieldMergeByKeyFunc(structType, field, newMergeByField(key))
}

// WithFieldMergeByKeyFunc merges the given struct field with merge-by-key semantics. The field must be of slice
// type. The slice element type must be of some other struct type, or a pointer thereto. The given SliceMergeKeyFunc
// will be used to extract the slice element's merge key; therefore, the field should generally be a unique identifier
// or primary key for objects of this type.
func WithFieldMergeByKeyFunc(structType reflect.Type, field string, mergeKeyFunc SliceMergeKeyFunc) DeepMergeOption {
	return func(c *coalescer) {
		if c.fieldMergers == nil {
			c.fieldMergers = make(map[reflect.Type]map[string]DeepMergeFunc)
		}
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepSliceMerge(v1, v2, mergeKeyFunc)
		}
	}
}

// WithAtomicField causes the given field to be merged atomically, that is, with  "atomic" semantics, instead of its
// default merge semantics. When 2 non-zero-values of this field are merged, the second value is returned as is.
// This is the programmatic equivalent of adding a `goalesce:atomic` struct tag to that field.
func WithAtomicField(structType reflect.Type, field string) DeepMergeOption {
	return func(c *coalescer) {
		if c.fieldMergers == nil {
			c.fieldMergers = make(map[reflect.Type]map[string]DeepMergeFunc)
		}
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = c.deepMergeAtomic
	}
}
