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

// MapCoalescerOption is an option to be passed to NewMapCoalescer. There is currently no available built-in option, but
// this could change in the future.
type MapCoalescerOption func(c *mapCoalescer)

// NewMapCoalescer creates a new Coalescer for maps.
func NewMapCoalescer(opts ...MapCoalescerOption) Coalescer {
	c := &mapCoalescer{
		fallback: &atomicCoalescer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type mapCoalescer struct {
	fallback Coalescer
}

func (c *mapCoalescer) WithFallback(fallback Coalescer) {
	c.fallback = fallback
}

func (c *mapCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Map); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	coalesced := reflect.MakeMap(v1.Type())
	for _, k := range v1.MapKeys() {
		coalesced.SetMapIndex(k, v1.MapIndex(k))
	}
	for _, k := range v2.MapKeys() {
		if v1.MapIndex(k).IsValid() {
			coalescedValue, err := c.fallback.Coalesce(v1.MapIndex(k), v2.MapIndex(k))
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
