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

func (c *coalescer) deepMergeMap(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	coalesced := reflect.MakeMap(v1.Type())
	for _, k := range v1.MapKeys() {
		if !v2.MapIndex(k).IsValid() {
			copiedKey, err := c.deepCopy(k)
			if err != nil {
				return reflect.Value{}, err
			}
			copiedValue, err := c.deepCopy(v1.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			coalesced.SetMapIndex(copiedKey, copiedValue)
		}
	}
	for _, k := range v2.MapKeys() {
		copiedKey, err := c.deepCopy(k)
		if err != nil {
			return reflect.Value{}, err
		}
		if v1.MapIndex(k).IsValid() {
			coalescedValue, err := c.deepMerge(v1.MapIndex(k), v2.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			coalesced.SetMapIndex(copiedKey, coalescedValue)
		} else {
			copiedValue, err := c.deepCopy(v2.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			coalesced.SetMapIndex(copiedKey, copiedValue)
		}
	}
	return coalesced, nil
}

func (c *coalescer) deepCopyMap(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	copied := reflect.MakeMapWithSize(v.Type(), v.Len())
	for _, k := range v.MapKeys() {
		copiedKey, err := c.deepCopy(k)
		if err != nil {
			return reflect.Value{}, err
		}
		copiedValue, err := c.deepCopy(v.MapIndex(k))
		if err != nil {
			return reflect.Value{}, err
		}
		copied.SetMapIndex(copiedKey, copiedValue)
	}
	return copied, nil
}
