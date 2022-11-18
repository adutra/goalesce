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

// deepMergeArray is the default array merger. It first checks if there is a custom array merger
// registered for the array type. If there is, it uses it. Otherwise, it uses the default array
// merge strategy, which is atomic.
func (c *coalescer) deepMergeArray(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	if arrayMerger, found := c.arrayMergers[v1.Type()]; found {
		return arrayMerger(v1, v2)
	}
	if c.arrayMerger != nil {
		return c.arrayMerger(v1, v2)
	}
	return c.deepMergeAtomic(v1, v2)
}

// deepMergeArrayByIndex is an alternate array merger that merges arrays with by-index semantics. It
// is not the default merge strategy for arrays; it is only activated if an array merger has been
// registered through one of the options: WithDefaultArrayMergeByIndex, WithArrayMergeByIndex.
func (c *coalescer) deepMergeArrayByIndex(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	merged := reflect.New(v1.Type())
	for i := 0; i < v1.Len(); i++ {
		elem, err := c.deepMerge(v1.Index(i), v2.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		merged.Elem().Index(i).Set(elem)
	}
	return merged.Elem(), nil
}

func (c *coalescer) deepCopyArray(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	copied := reflect.New(v.Type())
	for i := 0; i < v.Len(); i++ {
		elem, err := c.deepCopy(v.Index(i))
		if err != nil {
			return reflect.Value{}, err
		}
		copied.Elem().Index(i).Set(elem)
	}
	return copied.Elem(), nil
}
