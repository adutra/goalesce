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
	"reflect"
)

// DeepMerge merges the 2 values and returns the merged value.
//
// When called with no options, the function uses the following default algorithm:
//
//  - If both values are nil, return nil.
//  - If one value is nil, return the other value.
//  - If both values are zero-values for the type, return the type's zero-value.
//  - If one value is a zero-value for the type, return the other value.
//  - Otherwise, the values are merged using the following rules:
//    - If both values are interfaces of same underlying types, merge the underlying values.
//    - If both values are pointers, merge the values pointed to.
//    - If both values are maps, merge the maps recursively, key by key.
//    - If both values are structs, merge the structs recursively, field by field.
//    - For other types (including slices), return the second value ("atomic" semantics)
//
// This function never modifies its inputs. It always returns an entirely newly-allocated value that
// shares no references with the inputs.
//
// Note that by default, slices are merged with atomic semantics, that is, the second slice
// overwrites the first one completely. It is possible to change this behavior and use list-append,
// set-union, or merge-by semantics. See Option.
//
// This function returns an error if the values are not of the same type, or if the merge encounters
// an error.
func DeepMerge[T any](o1, o2 T, opts ...Option) (T, error) {
	v1 := reflect.ValueOf(o1)
	v2 := reflect.ValueOf(o2)
	coalescer := newCoalescer(opts...)
	result, err := coalescer.deepMerge(v1, v2)
	if !result.IsValid() || err != nil {
		return zero[T](), err
	}
	return cast[T](result)
}

// MustDeepMerge is like DeepMerge, but panics if the merge returns an error.
func MustDeepMerge[T any](o1, o2 T, opts ...Option) T {
	merged, err := DeepMerge(o1, o2, opts...)
	if err != nil {
		panic(err)
	}
	return merged
}
