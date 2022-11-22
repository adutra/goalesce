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

// deepMergeAtomic merges two values with atomic semantics, that is, it assumes the values are
// immutable and indivisible and therefore, that there is nothing to be merged. To comply with the
// general contract of DeepMergeFunc, it returns a deep copy of the first value if the second value
// is the zero-value; otherwise, it returns a deep copy of the second value. By default, this
// function is used to "merge" all immutable value types (int, string, etc.), and also to merge
// slices and arrays.
func (c *coalescer) deepMergeAtomic(v1, v2 reflect.Value) (reflect.Value, error) {
	if v2.IsZero() {
		return c.deepCopy(v1)
	}
	return c.deepCopy(v2)
}

// deepCopyAtomic copies the value with atomic semantics, that is, it assumes the value is immutable
// and indivisible, and that the value is a copy of itself. Therefore, it simply returns the value
// as is. By default, this function is used to "copy" all immutable value types (int, string, etc.).
func (c *coalescer) deepCopyAtomic(v reflect.Value) (reflect.Value, error) {
	return v, nil
}
