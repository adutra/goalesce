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

// DeepCopyFunc is the main function for copying objects. Simple usages of this package do not need to implement
// this function. Implementing this function is considered an advanced usage.
type DeepCopyFunc func(v reflect.Value) (reflect.Value, error)

// NewDeepCopyFunc creates a new DeepCopyFunc with the given options.
func NewDeepCopyFunc(opts ...DeepCopyOption) DeepCopyFunc {
	c := &coalescer{}
	for _, opt := range opts {
		opt(c)
	}
	return c.deepCopy
}
