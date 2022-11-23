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
func DeepCopy[T any](o T, opts ...Option) (T, error) {
	coalescer := newCoalescer(opts...)
	v := reflect.ValueOf(o)
	result, err := coalescer.deepCopy(v)
	if !result.IsValid() || err != nil {
		return zero[T](), err
	}
	return cast[T](result)
}

// MustDeepCopy is like DeepCopy, but panics if the copy returns an error.
func MustDeepCopy[T any](o T, opts ...Option) T {
	copied, err := DeepCopy(o, opts...)
	if err != nil {
		panic(err)
	}
	return copied
}
