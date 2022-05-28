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

// DeepCopy deep-copies the value and returns the copied value.
func DeepCopy(o interface{}, opts ...DeepCopyOption) (copied interface{}, err error) {
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
func MustDeepCopy(o interface{}, opts ...DeepCopyOption) interface{} {
	copied, err := DeepCopy(o, opts...)
	if err != nil {
		panic(err)
	}
	return copied
}
