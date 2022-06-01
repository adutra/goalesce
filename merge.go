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

// DeepMerge merges the 2 values into a single value.
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
// The function never modifies its inputs. It always returns an entirely newly-allocated value that shares no references
// with the inputs.
//
// Note that by default, slices are merged with atomic semantics, that is, the second slice overwrites the first one
// completely. It is possible to change this behavior and use list-append, set-union, or merge-by semantics.
// See DeepMergeOption.
//
// The function returns an error if the values are not of the same type, or if the merge encounters an error.
func DeepMerge(o1, o2 interface{}, opts ...DeepMergeOption) (coalesced interface{}, err error) {
	if o1 == nil {
		return o2, nil
	} else if o2 == nil {
		return o1, nil
	}
	deepMerge := NewDeepMergeFunc(opts...)
	v1 := reflect.ValueOf(o1)
	v2 := reflect.ValueOf(o2)
	result, err := deepMerge(v1, v2)
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

// MustDeepMerge is like DeepMerge, but panics if the merge returns an error.
func MustDeepMerge(o1, o2 interface{}, opts ...DeepMergeOption) interface{} {
	merged, err := DeepMerge(o1, o2, opts...)
	if err != nil {
		panic(err)
	}
	return merged
}
