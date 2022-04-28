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

// PointerCoalescerOption is an option to be passed to NewPointerCoalescer. There is currently no available built-in
// option, but this could change in the future.
type PointerCoalescerOption func(c *pointerCoalescer)

// NewPointerCoalescer creates a new Coalescer for pointer types.
func NewPointerCoalescer(opts ...PointerCoalescerOption) Coalescer {
	c := &pointerCoalescer{
		fallback: &atomicCoalescer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithOnCycleReturnError instructs the coalescer to return an error when a cycle is detected. By default, the coalescer
// replaces cycles with a nil pointer.
func WithOnCycleReturnError() PointerCoalescerOption {
	return func(c *pointerCoalescer) {
		c.onCycleReturnError = true
	}
}

type pointerCoalescer struct {
	fallback           Coalescer
	seen               map[uintptr]bool
	onCycleReturnError bool
}

func (c *pointerCoalescer) WithFallback(fallback Coalescer) {
	c.fallback = fallback
}

func (c *pointerCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Ptr); err != nil {
		return reflect.Value{}, err
	}
	if c.checkCycle(v1) {
		if c.onCycleReturnError {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v1.Type().String())
		}
		v1 = reflect.Zero(v1.Type())
	}
	if c.checkCycle(v2) {
		if c.onCycleReturnError {
			return reflect.Value{}, fmt.Errorf("%s: cycle detected", v2.Type().String())
		}
		v2 = reflect.Zero(v2.Type())
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	coalesced := reflect.New(v1.Elem().Type())
	coalescedTarget, err := c.fallback.Coalesce(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	coalesced.Elem().Set(coalescedTarget)
	return coalesced, nil
}

func (c *pointerCoalescer) checkCycle(v reflect.Value) bool {
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
