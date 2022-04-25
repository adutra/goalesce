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

// InterfaceCoalescerOption is an option to be passed to NewInterfaceCoalescer. There is currently no available built-in
// option, but this could change in the future.
type InterfaceCoalescerOption func(c *interfaceCoalescer)

// NewInterfaceCoalescer creates a new Coalescer for interface types.
func NewInterfaceCoalescer(opts ...InterfaceCoalescerOption) Coalescer {
	c := &interfaceCoalescer{
		fallback: &defaultCoalescer{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type interfaceCoalescer struct {
	fallback Coalescer
}

func (c *interfaceCoalescer) WithFallback(fallback Coalescer) {
	c.fallback = fallback
}

func (c *interfaceCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatchWithKind(v1, v2, reflect.Interface); err != nil {
		return reflect.Value{}, err
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	if v1.Elem().Type() != v2.Elem().Type() {
		return v2, nil
	}
	coalesced := reflect.New(v1.Type())
	coalescedTarget, err := c.fallback.Coalesce(v1.Elem(), v2.Elem())
	if err != nil {
		return reflect.Value{}, err
	}
	coalesced.Elem().Set(coalescedTarget)
	return coalesced.Elem(), nil
}
