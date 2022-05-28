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

// DeepCopy deep-copies the value and returns the copied value.
//
// This function never modifies its inputs. It always returns an entirely newly-allocated value that
// shares no references with the inputs.
func DeepCopy(o interface{}, opts ...Option) (copied interface{}, err error) {
	if o == nil {
		return nil, nil
	}
	deepCopy := NewDeepCopyFunc(opts...)
	v := reflect.ValueOf(o)
	result, err := deepCopy(v)
	if err != nil {
		return nil, err
	}
	return result.Interface(), nil
}

// MustDeepCopy is like DeepCopy, but panics if the copy returns an error.
func MustDeepCopy(o interface{}, opts ...Option) interface{} {
	copied, err := DeepCopy(o, opts...)
	if err != nil {
		panic(err)
	}
	return copied
}

// DeepCopyFunc is the main function for copying objects. Simple usages of this package do not need
// to implement this function. Implementing this function is considered an advanced usage. A deep
// copy function copies the given value to a newly-allocated value, avoiding retaining references to
// passed objects. Note that the passed values can be zero-values, but will never be invalid values.
// When a deep copy function returns an invalid value and a nil error, it is assumed that the
// function is delegating the copy to its parent, if any.
type DeepCopyFunc func(v reflect.Value) (reflect.Value, error)

// NewDeepCopyFunc creates a new DeepCopyFunc with the given options.
func NewDeepCopyFunc(opts ...Option) DeepCopyFunc {
	return newCoalescer(opts...).deepCopy
}
