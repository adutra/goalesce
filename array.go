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

// TODO deepMergeArray

func (c *coalescer) deepCopyArray(v reflect.Value) (reflect.Value, error) {
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
