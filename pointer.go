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
	"fmt"
	"reflect"
)

func (c *mainCoalescer) coalescePointer(v1, v2 reflect.Value) (reflect.Value, error) {
	if c.checkCycle(v1) {
		if c.errorOnCycle {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v1.Type().String())
		}
		v1 = reflect.Zero(v1.Type())
	}
	if c.checkCycle(v2) {
		if c.errorOnCycle {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v2.Type().String())
		}
		v2 = reflect.Zero(v2.Type())
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	coalesced := reflect.New(v1.Elem().Type())
	coalescedTarget, err := c.coalesce(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	coalesced.Elem().Set(coalescedTarget)
	return coalesced, nil
}

func (c *mainCoalescer) checkCycle(v reflect.Value) bool {
	if v.CanAddr() {
		if c.seen == nil {
			c.seen = make(map[uintptr]bool)
		}
		addr := v.UnsafeAddr()
		if _, found := c.seen[addr]; found {
			return true
		}
		c.seen[addr] = true
	}
	return false
}
