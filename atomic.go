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

// NewAtomicCoalescer returns a coalescer that always applies "atomic" semantics when coalescing. In other words, this
// coalescer does not coalesce anything, it simply returns v1 if v2 is the zero-value for its type, and v2 otherwise. It
// is suitable for coalescing scalar types mostly.
func NewAtomicCoalescer() Coalescer {
	return &atomicCoalescer{}
}

type atomicCoalescer struct{}

func (c *atomicCoalescer) WithFallback(Coalescer) {}

func (c *atomicCoalescer) Coalesce(v1, v2 reflect.Value) (reflect.Value, error) {
	if v2.IsZero() {
		return v1, nil
	}
	return v2, nil
}
