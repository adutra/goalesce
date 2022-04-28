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

// MainCoalescerOption is an option that can be passed the main Coalesce function to customize its coalescing behavior.
type MainCoalescerOption func(c *mainCoalescer)

// WithAtomicType causes the given type to be coalesced atomically, that is, with  "atomic" semantics, instead of its
// default coalesce semantics. When 2 non-zero-values of this type are coalesced, the second value is returned as is.
func WithAtomicType(t reflect.Type) MainCoalescerOption {
	return WithTypeCoalescer(t, &atomicCoalescer{})
}

// WithTrileans causes all boolean pointers to be coalesced using a three-valued logic, instead of their default
// coalesce semantics. When this is enabled, boolean pointers will behave as if they were "trileans", that is, a type
// with 3 possible values: nil (its zero-value), false and true (contrary to booleans, with trileans false is NOT a
// zero-value).
// The coalescing of trileans obeys the following rules:
//   v1    v2    coalesced
//   nil   nil   nil
//   nil   false false
//   nil   true  true
//   false nil   false
//   false false false
//   false true  true
//   true  nil   true
//   true  false false
//   true  true  true
// The biggest difference with regular boolean pointers is that Coalesce(&true, &false) will return &true, while with
// trileans, it will return &false.
func WithTrileans() MainCoalescerOption {
	return WithAtomicType(reflect.PtrTo(reflect.TypeOf(false)))
}

// WithTypeCoalescer will defer the coalescing of the given type to the given custom coalescer.
func WithTypeCoalescer(t reflect.Type, coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		if c.typeCoalescers == nil {
			c.typeCoalescers = make(map[reflect.Type]Coalescer)
		}
		c.typeCoalescers[t] = coalescer
	}
}

// WithDefaultCoalescer changes the coalescer used for scalar types and types other than map, slice, struct and pointer.
// This should rarely be necessary.
func WithDefaultCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.defaultCoalescer = coalescer
	}
}

// WithStructCoalescer changes the coalescer used for structs. Use this to customize the coalescing of structs. See
// NewStructCoalescer for more information.
func WithStructCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.structCoalescer = coalescer
	}
}

// WithMapCoalescer changes the coalescer used for maps. Use this to customize the coalescing of maps. See
// NewMapCoalescer for more information.
func WithMapCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.mapCoalescer = coalescer
	}
}

// WithSliceCoalescer changes the coalescer used for slices. Use this to customize the coalescing of slices. See
// NewSliceCoalescer for more information.
func WithSliceCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.sliceCoalescer = coalescer
	}
}

// WithPointerCoalescer changes the coalescer used for pointers. Use this to customize the coalescing of pointers. See
// NewPointerCoalescer for more information.
func WithPointerCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.pointerCoalescer = coalescer
	}
}

// WithInterfaceCoalescer changes the coalescer used for interfaces. Use this to customize the coalescing of interfaces.
// See NewInterfaceCoalescer for more information.
func WithInterfaceCoalescer(coalescer Coalescer) MainCoalescerOption {
	return func(c *mainCoalescer) {
		c.interfaceCoalescer = coalescer
	}
}

// NewMainCoalescer creates a new main coalescer with the given options. This is the Coalescer used by the Coalesce
// function. The main coalescer always sets itself as the fallback coalescer for all its delegating coalescers.
func NewMainCoalescer(opts ...MainCoalescerOption) *mainCoalescer {
	c := &mainCoalescer{
		defaultCoalescer:   NewAtomicCoalescer(),
		sliceCoalescer:     NewSliceCoalescer(),
		mapCoalescer:       NewMapCoalescer(),
		structCoalescer:    NewStructCoalescer(),
		pointerCoalescer:   NewPointerCoalescer(),
		interfaceCoalescer: NewInterfaceCoalescer(),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.WithFallback(c)
	return c
}

type mainCoalescer struct {
	defaultCoalescer   Coalescer
	pointerCoalescer   Coalescer
	interfaceCoalescer Coalescer
	mapCoalescer       Coalescer
	structCoalescer    Coalescer
	sliceCoalescer     Coalescer
	typeCoalescers     map[reflect.Type]Coalescer
}

func (c *mainCoalescer) WithFallback(Coalescer) {
	c.defaultCoalescer.WithFallback(c)
	c.sliceCoalescer.WithFallback(c)
	c.mapCoalescer.WithFallback(c)
	c.structCoalescer.WithFallback(c)
	c.pointerCoalescer.WithFallback(c)
	c.interfaceCoalescer.WithFallback(c)
	for _, coalescer := range c.typeCoalescers {
		coalescer.WithFallback(c)
	}
}

func (c *mainCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if err := checkTypesMatch(v1, v2); err != nil {
		return reflect.Value{}, err
	}
	if coalescer, found := c.typeCoalescers[v1.Type()]; found {
		return coalescer.Coalesce(v1, v2)
	}
	if value, done := checkZero(v1, v2); done {
		return value, nil
	}
	switch v1.Type().Kind() {
	case reflect.Interface:
		return c.interfaceCoalescer.Coalesce(v1, v2)
	case reflect.Ptr:
		return c.pointerCoalescer.Coalesce(v1, v2)
	case reflect.Map:
		return c.mapCoalescer.Coalesce(v1, v2)
	case reflect.Struct:
		return c.structCoalescer.Coalesce(v1, v2)
	case reflect.Slice:
		return c.sliceCoalescer.Coalesce(v1, v2)
	}
	return c.defaultCoalescer.Coalesce(v1, v2)
}
