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
	"errors"
	"reflect"
)

var (
	withMockDeepCopyError Option = func(c *coalescer) {
		c.deepCopy = func(v reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("mock DeepCopy error")
		}
	}
	withMockDeepMergeError Option = func(c *coalescer) {
		c.deepMerge = func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("mock DeepMerge error")
		}
	}
)

func withMockDeepCopyErrorWhen(expected interface{}) Option {
	return func(c *coalescer) {
		c.deepCopy = func(v reflect.Value) (reflect.Value, error) {
			if expected == v.Interface() {
				return reflect.Value{}, errors.New("mock DeepCopy error")
			}
			return c.defaultDeepCopy(v)
		}
	}
}

func withMockDeepMergeErrorWhen(expected1, expected2 interface{}) Option {
	return func(c *coalescer) {
		c.deepMerge = func(v1, v2 reflect.Value) (reflect.Value, error) {
			if expected1 == v1.Interface() && expected2 == v2.Interface() {
				return reflect.Value{}, errors.New("mock DeepMerge error")
			}
			return c.defaultDeepMerge(v1, v2)
		}
	}
}
