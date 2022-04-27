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

// Coalescer is the main interface for coalescing objects. Simple usages of this package do not need to implement this
// interface. Implementing this interface is considered an advanced usage.
type Coalescer interface {

	// Coalesce coalesces the 2 values into a single value. All built-in implementations of this interface favor v2 over
	// v1 in case of conflicts. Note that the passed values can be zero-values, but will never be invalid values; it is
	// also expected that implementors will never return invalid values if the returned error is nil.
	Coalesce(v1, v2 reflect.Value) (reflect.Value, error)

	// WithFallback sets the fallback Coalescer for this coalescer. This is used in case this coalescer only handles one
	// specific type, and needs to call another coalescer to handle other types encountered during merge. Any Coalescer
	// implementation passed to the Coalesce function will have its WithFallback method called with the main Coalescer
	// before the coalescing operation begins. Simple Coalescers can implement this method as a no-op.
	WithFallback(fallback Coalescer)
}
