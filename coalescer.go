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

// Coalescer is the main function for coalescing objects. Simple usages of this package do not need to implement this
// function. Implementing this function is considered an advanced usage.
// A coalescer function coalesces the 2 values into a single value, favoring v2 over v1 in case of conflicts. Note that
// the passed values can be zero-values, but will never be invalid values.
// When a coalescer function returns an invalid value and a nil error, it is assumed that the function is delegating the
// coalescing to its parent, if any.
type Coalescer func(v1, v2 reflect.Value) (reflect.Value, error)

// NewCoalescer creates a new coalescer with the given options.
func NewCoalescer(opts ...CoalescerOption) Coalescer {
	c := &mainCoalescer{}
	for _, opt := range opts {
		opt(c)
	}
	return c.coalesce
}

type mainCoalescer struct {
	typeCoalescers  map[reflect.Type]Coalescer
	sliceCoalescer  Coalescer
	sliceCoalescers map[ /* slice type */ reflect.Type]Coalescer
	fieldCoalescers map[ /* struct type */ reflect.Type]map[ /* field name */ string]Coalescer
	zeroEmptySlice  bool
	errorOnCycle    bool
	seen            map[uintptr]bool
}

func (c *mainCoalescer) coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatch(v1, v2); err != nil {
		return reflect.Value{}, err
	}
	if coalescer, found := c.typeCoalescers[v1.Type()]; found {
		value, err := coalescer(v1, v2)
		if value.IsValid() || err != nil {
			return value, err
		}
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	switch v1.Type().Kind() {
	case reflect.Interface:
		return c.coalesceInterface(v1, v2)
	case reflect.Ptr:
		return c.coalescePointer(v1, v2)
	case reflect.Map:
		return c.coalesceMap(v1, v2)
	case reflect.Struct:
		return c.coalesceStruct(v1, v2)
	case reflect.Slice:
		return c.coalesceSlice(v1, v2)
	}
	return coalesceAtomic(v1, v2)
}
