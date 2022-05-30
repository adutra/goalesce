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

func (c *mainCoalescer) coalesceMap(v1, v2 reflect.Value) (reflect.Value, error) {
	coalesced := reflect.MakeMap(v1.Type())
	for _, k := range v1.MapKeys() {
		coalesced.SetMapIndex(k, v1.MapIndex(k))
	}
	for _, k := range v2.MapKeys() {
		if v1.MapIndex(k).IsValid() {
			coalescedValue, err := c.coalesce(v1.MapIndex(k), v2.MapIndex(k))
			if err != nil {
				return reflect.Value{}, err
			}
			coalesced.SetMapIndex(k, coalescedValue)
		} else {
			coalesced.SetMapIndex(k, v2.MapIndex(k))
		}
	}
	return coalesced, nil
}
