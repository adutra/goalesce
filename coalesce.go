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

// Coalesce merges the 2 values into a single value.
//
// When called with no options, the function uses the following default algorithm:
//
//   - If both values are nil, return nil.
//   - If one value is nil, return the other value.
//   - If both values are zero values for the type, return the type's zero value.
//   - If one value is a zero value for the type, return the other value.
//   - If both values are non-zero values, the values are coalesced using the following rules:
//     - If both values are pointers, coalesce the values pointed to.
//     - If both values are structs, coalesce the structs recursively, field by field.
//     - If both values are maps, coalesce the maps recursively, key by key.
//     - Otherwise, return the second value.
//
// The function never modifies its inputs. It may return one of the (unmodified) inputs as the coalesced value, but
// whenever changes to the inputs are required, the function always returns an entirely newly-allocated value.
//
// Note that by default, slices are coalesced with replace semantics, that is, the second slice overwrites the first one
// completely. It is possible to change this behavior and use list-append, set-union, or merge-by semantics.
// See SliceCoalescerOption.
//
// The function returns an error if the values are not of the same type.
func Coalesce(o1, o2 interface{}, opts ...MainCoalescerOption) (coalesced interface{}, err error) {
	if o1 == nil {
		return o2, nil
	} else if o2 == nil {
		return o1, nil
	}
	coalescer := NewMainCoalescer(opts...)
	v1 := reflect.ValueOf(o1)
	v2 := reflect.ValueOf(o2)
	result, err := coalescer.Coalesce(v1, v2)
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}
