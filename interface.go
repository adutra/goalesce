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

func (c *coalescer) deepMergeInterface(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	mergedTarget, err := c.deepMerge(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	merged := reflect.New(v1.Type())
	merged.Elem().Set(mergedTarget)
	return merged.Elem(), nil
}

func (c *coalescer) deepCopyInterface(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	copied := reflect.New(v.Type())
	copiedTarget, err := c.deepCopy(v.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	copied.Elem().Set(copiedTarget)
	return copied.Elem(), nil
}
