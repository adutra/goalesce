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
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertNotSame is a deep version of assert.NotSame and is used to ensure that two pointers, even
// nested ones, do not point to the same object.
func assertNotSame(t *testing.T, o1, o2 interface{}) bool {
	if o1 == nil || o2 == nil {
		return true
	}
	v1 := reflect.ValueOf(o1)
	v2 := reflect.ValueOf(o2)
	switch v1.Type().Kind() {
	case reflect.Array:
		for i := 0; i < v1.Len() && i < v2.Len(); i++ {
			if !assertNotSame(t, v1.Index(i), v2.Index(i)) {
				return false
			}
		}
	case reflect.Slice:
		if !v1.IsNil() {
			for i := 0; i < v1.Len() && i < v2.Len(); i++ {
				if !assertNotSame(t, v1.Index(i), v2.Index(i)) {
					return false
				}
			}
		}
	case reflect.Map:
		if !v1.IsNil() {
			for i, k := range v1.MapKeys() {
				if v2.MapIndex(k).IsValid() {
					if !assertNotSame(t, k, v2.MapKeys()[i]) {
						return false
					}
					if !assertNotSame(t, v1.MapIndex(k), v2.MapIndex(k)) {
						return false
					}
				}
			}
		}
	case reflect.Struct:
		for i := 0; i < v1.NumField(); i++ {
			if v1.Type().Field(i).IsExported() {
				if !assertNotSame(t, v1.Field(i), v2.Field(i)) {
					return false
				}
			}
		}
	case reflect.Interface:
		if !v1.IsNil() {
			if !assertNotSame(t, v1.Elem(), v2.Elem()) {
				return false
			}
		}
	case reflect.Ptr:
		if !v1.IsNil() {
			if !assert.NotSame(t, v1.Interface(), v2.Interface()) {
				return false
			}
			if !assertNotSame(t, v1.Elem(), v2.Elem()) {
				return false
			}
		}
	}
	return true
}
