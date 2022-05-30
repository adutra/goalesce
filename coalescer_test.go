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

func TestNewCoalescer(t *testing.T) {
	t.Run("with generic option", func(t *testing.T) {
		var actual int
		opt := func(c *mainCoalescer) {
			actual = 1
		}
		NewCoalescer(opt)
		assert.Equal(t, 1, actual)
	})
	t.Run("type errors", func(t *testing.T) {
		_, err := NewCoalescer()(reflect.ValueOf(1), reflect.ValueOf("a"))
		assert.EqualError(t, err, "types do not match: int != string")
		_, err = NewCoalescer()(reflect.ValueOf(map[string]int{"a": 2}), reflect.ValueOf(map[string]string{"a": "b"}))
		assert.EqualError(t, err, "types do not match: map[string]int != map[string]string")
		_, err = NewCoalescer()(reflect.ValueOf(intPtr(1)), reflect.ValueOf(stringPtr("a")))
		assert.EqualError(t, err, "types do not match: *int != *string")
		_, err = NewCoalescer()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = NewCoalescer()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		_, err = NewCoalescer()(reflect.ValueOf([]int{1}), reflect.ValueOf([]string{"a"}))
		assert.EqualError(t, err, "types do not match: []int != []string")
		type foo struct {
			Int int
		}
		type bar struct {
			Int int
		}
		_, err = NewCoalescer()(reflect.ValueOf(foo{}), reflect.ValueOf(bar{}))
		assert.EqualError(t, err, "types do not match: goalesce.foo != goalesce.bar")
	})
}
