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

func (c *coalescer) deepMergePointer(v1, v2 reflect.Value) (reflect.Value, error) {
	if value, done := checkZero(v1, v2); done {
		return c.deepCopy(value)
	}
	if c.checkCycle(v1) {
		if c.errorOnCycle {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v1.Type().String())
		}
		return c.deepCopy(v2)
	}
	if c.checkCycle(v2) {
		if c.errorOnCycle {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v2.Type().String())
		}
		c.unsee(v1) // because checkCycle(v1) was called
		return c.deepCopy(v1)
	}
	mergedTarget, err := c.deepMerge(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	merged := reflect.New(v1.Type().Elem())
	merged.Elem().Set(mergedTarget)
	return merged, nil
}

func (c *coalescer) deepCopyPointer(v reflect.Value) (reflect.Value, error) {
	if v.IsZero() {
		return reflect.Zero(v.Type()), nil
	}
	if c.checkCycle(v) {
		if c.errorOnCycle {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v.Type().String())
		}
		return reflect.Zero(v.Type()), nil
	}
	copiedTarget, err := c.deepCopy(v.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	copied := reflect.New(v.Type().Elem())
	copied.Elem().Set(copiedTarget)
	return copied, nil
}

func (c *coalescer) checkCycle(v reflect.Value) bool {
	if v.CanAddr() {
		addr := v.UnsafeAddr()
		if _, found := c.seen[addr]; found {
			return true
		}
		c.seen[addr] = true
	}
	return false
}

func (c *coalescer) unsee(v reflect.Value) {
	if v.CanAddr() {
		addr := v.UnsafeAddr()
		delete(c.seen, addr)
	}
}
