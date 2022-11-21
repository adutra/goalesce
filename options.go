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

// Option is an option that can be passed to DeepCopy or DeepMerge to customize the function
// behavior.
type Option func(c *coalescer)

// DeepCopyFunc is a function for copying objects. A deep copy function is expected to abide by the
// general contract of DeepCopy and to copy the given value to a newly-allocated value, avoiding
// retaining references to passed objects. Note that the passed values can be zero-values, but will
// never be invalid values. The returned value must be of same type as the passed value. The
// returned value should only be invalid when the copy failed and the error is not nil.
type DeepCopyFunc func(v reflect.Value) (reflect.Value, error)

// DeepMergeFunc is a function for merging objects. A deep merge function is expected to abide by
// the general contract of DeepMerge and to merge the 2 values into a single value, favoring v2 over
// v1 in case of conflicts. Note that the passed values can be zero-values, but will never be
// invalid values. The passed values are guaranteed to be of the same type; the returned value must
// also be of that same type. The returned value should only be invalid when the merge failed and
// the error is not nil.
type DeepMergeFunc func(v1, v2 reflect.Value) (reflect.Value, error)

// COMMON OPTIONS

// WithErrorOnCycle instructs the operation to return an error when a cycle is detected. By default,
// cycles are replaced with a nil pointer.
func WithErrorOnCycle() Option {
	return func(c *coalescer) {
		c.errorOnCycle = true
	}
}

// DEEP COPY OPTIONS

// WithTypeCopier will defer the copy of the given type to the given custom copier. This option does
// not allow the type copier to access the global DeepCopyFunc instance. For that, use
// WithTypeCopierProvider instead.
func WithTypeCopier(t reflect.Type, copier DeepCopyFunc) Option {
	return WithTypeCopierProvider(t, func(DeepCopyFunc) DeepCopyFunc {
		return copier
	})
}

// WithTypeCopierProvider will defer the copy of the given type to a custom copier that will be
// obtained by calling the given provider function with the global DeepCopyFunc instance. This
// option allows the type copier to access this instance in order to delegate the copy of nested
// objects. See ExampleWithTypeCopierProvider.
func WithTypeCopierProvider(t reflect.Type, provider func(global DeepCopyFunc) DeepCopyFunc) Option {
	return func(c *coalescer) {
		c.typeCopiers[t] = provider(c.deepCopy)
	}
}

// DEEP MERGE OPTIONS

// WithAtomicMerge causes the given type to be merged atomically, that is, with  "atomic" semantics,
// instead of its default merge semantics. When 2 non-zero-values of this type are merged, the
// second value is returned as is.
func WithAtomicMerge(t reflect.Type) Option {
	return func(c *coalescer) {
		c.typeMergers[t] = c.deepMergeAtomic
	}
}

// WithTrileanMerge causes all boolean pointers to be merged using a three-valued logic, instead of
// their default merge semantics. When this is enabled, boolean pointers will behave as if they were
// "trileans", that is, a type with 3 possible values: nil (its zero-value), false and true
// (contrary to booleans, with trileans false is NOT a zero-value).
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
// The biggest difference with regular boolean pointers is that DeepMerge(&true, &false) will return
// &true for boolean pointers, while with trileans, it will return &false.
func WithTrileanMerge() Option {
	return WithAtomicMerge(reflect.PtrTo(reflect.TypeOf(false)))
}

// WithTypeMerger will defer the merge of the given type to the given custom merger. This option
// does not allow the type merger to access the global DeepMergeFunc instance. For
// that, use WithTypeMergerProvider instead.
func WithTypeMerger(t reflect.Type, merger DeepMergeFunc) Option {
	return WithTypeMergerProvider(t, func(DeepMergeFunc, DeepCopyFunc) DeepMergeFunc {
		return merger
	})
}

// WithTypeMergerProvider will defer the merge of the given type to a custom merger that will be
// obtained by calling the given provider function with the global DeepMergeFunc and DeepCopyFunc
// instances. This option allows the type merger to access those instances in order to delegate the
// merge and copy of nested objects. See ExampleWithTypeMergerProvider.
func WithTypeMergerProvider(t reflect.Type, provider func(globalMerger DeepMergeFunc, globalCopier DeepCopyFunc) DeepMergeFunc) Option {
	return func(c *coalescer) {
		c.typeMergers[t] = provider(c.deepMerge, c.deepCopy)
	}
}

// WithZeroEmptySliceMerge instructs the merger to consider empty slices as zero (nil) slices. This
// changes the default behavior: when merging a non-empty slice with an empty slice, normally the
// empty slice is returned, but with this option, the non-empty slice is returned.
func WithZeroEmptySliceMerge() Option {
	return func(c *coalescer) {
		c.zeroEmptySlice = true
	}
}

// WithDefaultSliceListAppendMerge applies list-append merge semantics to all slices to be merged.
func WithDefaultSliceListAppendMerge() Option {
	return func(c *coalescer) {
		c.sliceMerger = c.deepMergeSliceWithListAppend
	}
}

// WithDefaultSliceSetUnionMerge applies set-union merge semantics to all slices to be merged. When
// the slice elements are pointers, this strategy dereferences the pointers and compare their
// targets. This strategy is fine for slices of simple types and pointers thereof, but it is not
// recommended for slices of complex types as the elements may not be fully comparable.
func WithDefaultSliceSetUnionMerge() Option {
	return func(c *coalescer) {
		c.sliceMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeSliceWithMergeKey(v1, v2, SliceUnion)
		}
	}
}

// WithDefaultSliceMergeByIndex applies merge-by-index semantics to all slices to be merged.
func WithDefaultSliceMergeByIndex() Option {
	return func(c *coalescer) {
		c.sliceMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeSliceWithMergeKey(v1, v2, SliceIndex)
		}
	}
}

// WithDefaultArrayMergeByIndex applies merge-by-index semantics to all arrays to be merged.
func WithDefaultArrayMergeByIndex() Option {
	return func(c *coalescer) {
		c.arrayMerger = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeArrayByIndex(v1, v2)
		}
	}
}

// WithSliceSetUnionMerge applies set-union merge semantics to the given slice type. When the slice
// elements are of a pointer type, this strategy dereferences the pointers and compare their
// targets. This strategy is fine for slices of simple types and pointers thereof, but it is not
// recommended for slices of complex types as the elements may not be fully comparable.
func WithSliceSetUnionMerge(sliceType reflect.Type) Option {
	return WithSliceMergeByKeyFunc(sliceType, SliceUnion)
}

// WithSliceListAppendMerge applies list-append merge semantics to the given slice type.
func WithSliceListAppendMerge(sliceType reflect.Type) Option {
	return func(c *coalescer) {
		c.sliceMergers[sliceType] = c.deepMergeSliceWithListAppend
	}
}

// WithSliceMergeByIndex applies merge-by-index semantics to the given slice type. The given
// mergeKeyFunc will be used to extract the element merge key.
func WithSliceMergeByIndex(sliceType reflect.Type) Option {
	return WithSliceMergeByKeyFunc(sliceType, SliceIndex)
}

// WithArrayMergeByIndex applies merge-by-index semantics to the given slice type. The given
// mergeKeyFunc will be used to extract the element merge key.
func WithArrayMergeByIndex(arrayType reflect.Type) Option {
	return func(c *coalescer) {
		c.arrayMergers[arrayType] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeArrayByIndex(v1, v2)
		}
	}
}

// WithSliceMergeByID applies merge-by-key semantics to slices whose elements are of some struct
// type, or a pointer thereto. The passed field name will be used to extract the element's merge
// key; therefore, the field should generally be a unique identifier or primary key for objects of
// this type.
func WithSliceMergeByID(sliceOfStructType reflect.Type, elemField string) Option {
	return func(c *coalescer) {
		WithSliceMergeByKeyFunc(sliceOfStructType, newMergeByField(elemField))(c)
	}
}

// WithSliceMergeByKeyFunc applies merge-by-key semantics to the given slice type. The given
// SliceMergeKeyFunc will be used to extract the element merge key.
func WithSliceMergeByKeyFunc(sliceType reflect.Type, mergeKeyFunc SliceMergeKeyFunc) Option {
	return func(c *coalescer) {
		c.sliceMergers[sliceType] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeSliceWithMergeKey(v1, v2, mergeKeyFunc)
		}
	}
}

// WithFieldMerger merges the given struct field with the given custom merger. This option does not
// allow the type merger to access the parent DeepMergeFunc instance being created. For that, use
// WithFieldMergerProvider instead.
func WithFieldMerger(structType reflect.Type, field string, merger DeepMergeFunc) Option {
	return WithFieldMergerProvider(structType, field, func(DeepMergeFunc, DeepCopyFunc) DeepMergeFunc {
		return merger
	})
}

// WithFieldMergerProvider merges the given struct field with a custom merger that will be obtained
// by calling the given provider function with the global DeepMergeFunc and DeepCopyFunc instances.
// This option allows the type merger to access those instances in order to delegate the merge and
// copy of nested objects. See ExampleWithFieldMergerProvider.
func WithFieldMergerProvider(structType reflect.Type, field string, provider func(globalMerger DeepMergeFunc, globalCopier DeepCopyFunc) DeepMergeFunc) Option {
	return func(c *coalescer) {
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = provider(c.deepMerge, c.deepCopy)
	}
}

// WithFieldListAppendMerge merges the given struct field with list-append semantics. The field must
// be of slice type. This is the programmatic equivalent of adding a `goalesce:append` struct tag to
// that field.
func WithFieldListAppendMerge(structType reflect.Type, field string) Option {
	return func(c *coalescer) {
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = c.deepMergeSliceWithListAppend
	}
}

// WithFieldSetUnionMerge merges the given struct field with set-union semantics. The field must be
// of slice type. This is the programmatic equivalent of adding a `goalesce:union` struct tag to
// that field.
func WithFieldSetUnionMerge(structType reflect.Type, field string) Option {
	return WithFieldMergeByKeyFunc(structType, field, SliceUnion)
}

// WithFieldMergeByIndex merges the given struct field with merge-by-index semantics. The field must
// be of slice type. This is the programmatic equivalent of adding a `goalesce:index` struct tag to
// that field.
func WithFieldMergeByIndex(structType reflect.Type, field string) Option {
	return WithFieldMergeByKeyFunc(structType, field, SliceIndex)
}

// WithFieldMergeByID merges the given struct field with merge-by-key semantics. The field must be
// of slice type. The slice element type must be of some other struct type, or a pointer thereto.
// The passed key must be a valid field name for that struct type and will be used to extract the
// slice element's merge key; therefore, that field should generally be a unique identifier or
// primary key for objects of this type. This is the programmatic equivalent of adding a
// `goalesce:id:key` struct tag to the struct field.
func WithFieldMergeByID(structType reflect.Type, field string, key string) Option {
	return WithFieldMergeByKeyFunc(structType, field, newMergeByField(key))
}

// WithFieldMergeByKeyFunc merges the given struct field with merge-by-key semantics. The field must
// be of slice type. The slice element type must be of some other struct type, or a pointer thereto.
// The given SliceMergeKeyFunc will be used to extract the slice element's merge key; therefore, the
// field should generally be a unique identifier or primary key for objects of this type.
func WithFieldMergeByKeyFunc(structType reflect.Type, field string, mergeKeyFunc SliceMergeKeyFunc) Option {
	return func(c *coalescer) {
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return c.deepMergeSliceWithMergeKey(v1, v2, mergeKeyFunc)
		}
	}
}

// WithAtomicFieldMerge causes the given field to be merged atomically, that is, with "atomic"
// semantics, instead of its default merge semantics. When 2 non-zero-values of this field are
// merged, the second value is returned as is. This is the programmatic equivalent of adding a
// `goalesce:atomic` struct tag to that field.
func WithAtomicFieldMerge(structType reflect.Type, field string) Option {
	return func(c *coalescer) {
		if c.fieldMergers[structType] == nil {
			c.fieldMergers[structType] = make(map[string]DeepMergeFunc)
		}
		c.fieldMergers[structType][field] = c.deepMergeAtomic
	}
}
