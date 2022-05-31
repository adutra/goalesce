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

func TestWithAtomicType(t *testing.T) {
	c := &mainCoalescer{}
	WithAtomicType(reflect.TypeOf(map[string]int{}))(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(map[string]int{})])
	got, err := c.coalesce(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int{"b": 2}))
	assert.Equal(t, map[string]int{"b": 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithTrileans(t *testing.T) {
	c := &mainCoalescer{}
	WithTrileans()(c)
	assert.NotNil(t, c.typeCoalescers[reflect.PtrTo(reflect.TypeOf(false))])
	got, err := c.coalesce(reflect.ValueOf(boolPtr(true)), reflect.ValueOf(boolPtr(false)))
	assert.Equal(t, boolPtr(false), got.Interface())
	assert.NoError(t, err)
}

func TestWithTypeCoalescer(t *testing.T) {
	c := &mainCoalescer{}
	called := false
	WithTypeCoalescer(reflect.TypeOf(map[string]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
		called = true
		return coalesceAtomic(v1, v2)
	})(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(map[string]int{})])
	got, err := c.coalesce(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int{"b": 2}))
	assert.Equal(t, map[string]int{"b": 2}, got.Interface())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithTypeCoalescerProvider(t *testing.T) {
	c := &mainCoalescer{}
	called := 0
	WithTypeCoalescerProvider(reflect.TypeOf(map[string]int{}), func(parent Coalescer) Coalescer {
		called++
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			called++
			return coalesceAtomic(v1, v2)
		}
	})(c)
	assert.NotNil(t, c.typeCoalescers[reflect.TypeOf(map[string]int{})])
	got, err := c.coalesce(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int{"b": 2}))
	assert.Equal(t, map[string]int{"b": 2}, got.Interface())
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithFieldCoalescer(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	called := false
	WithFieldCoalescer(reflect.TypeOf(User{}), "ID", func(v1, v2 reflect.Value) (reflect.Value, error) {
		called = true
		return coalesceAtomic(v1, v2)
	})(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
	got, err := c.coalesce(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
	assert.Equal(t, User{"Bob"}, got.Interface())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithFieldCoalescerProvider(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	called := 0
	WithFieldCoalescerProvider(reflect.TypeOf(User{}), "ID", func(parent Coalescer) Coalescer {
		called++
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			called++
			return coalesceAtomic(v1, v2)
		}
	})(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
	got, err := c.coalesce(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
	assert.Equal(t, User{"Bob"}, got.Interface())
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithAtomicField(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	WithAtomicField(reflect.TypeOf(User{}), "ID")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["ID"])
	got, err := c.coalesce(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
	assert.Equal(t, User{"Bob"}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultListAppend(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultListAppend()(c)
	assert.NotNil(t, c.sliceCoalescer)
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultMergeByIndex(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultMergeByIndex()(c)
	assert.NotNil(t, c.sliceCoalescer)
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{-1}))
	assert.Equal(t, []int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultSetUnion(t *testing.T) {
	c := &mainCoalescer{}
	WithDefaultSetUnion()(c)
	assert.NotNil(t, c.sliceCoalescer)
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithErrorOnCycle(t *testing.T) {
	c := &mainCoalescer{}
	WithErrorOnCycle()(c)
	assert.Equal(t, true, c.errorOnCycle)
}

func TestWithListAppend(t *testing.T) {
	c := &mainCoalescer{}
	WithListAppend(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithSetUnion(t *testing.T) {
	c := &mainCoalescer{}
	WithSetUnion(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithMergeByIndex(t *testing.T) {
	c := &mainCoalescer{}
	WithMergeByIndex(reflect.TypeOf([]int{}))(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]int{})])
	got, err := c.coalesce(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{-1}))
	assert.Equal(t, []int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithMergeByKey(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	called := false
	mergeKeyFunc := func(index int, element reflect.Value) (key reflect.Value, err error) {
		called = true
		return element.FieldByName("ID"), nil
	}
	WithMergeByKeyFunc(reflect.TypeOf([]User{}), mergeKeyFunc)(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]User{})])
	got, err := c.coalesce(reflect.ValueOf([]User{{"Alice"}, {"Bob"}}), reflect.ValueOf([]User{{"Bob"}, {"Alice"}}))
	assert.Equal(t, []User{{"Alice"}, {"Bob"}}, got.Interface())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithMergeByField(t *testing.T) {
	type User struct {
		ID string
	}
	c := &mainCoalescer{}
	WithMergeByID(reflect.TypeOf([]User{}), "ID")(c)
	assert.NotNil(t, c.sliceCoalescers[reflect.TypeOf([]User{})])
	got, err := c.coalesce(reflect.ValueOf([]User{{"Alice"}, {"Bob"}}), reflect.ValueOf([]User{{"Bob"}, {"Alice"}}))
	assert.Equal(t, []User{{"Alice"}, {"Bob"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithZeroEmptySlice(t *testing.T) {
	c := &mainCoalescer{}
	WithZeroEmptySlice()(c)
	assert.Equal(t, true, c.zeroEmptySlice)
}

func TestWithFieldListAppend(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := &mainCoalescer{}
	WithFieldListAppend(reflect.TypeOf(User{}), "Tags")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.coalesce(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldSetUnion(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := &mainCoalescer{}
	WithFieldSetUnion(reflect.TypeOf(User{}), "Tags")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.coalesce(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldMergeByIndex(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := &mainCoalescer{}
	WithFieldMergeByIndex(reflect.TypeOf(User{}), "Tags")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.coalesce(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag1a"}}))
	assert.Equal(t, User{Tags: []string{"tag1a", "tag2"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldMergeByID(t *testing.T) {
	type Tag struct {
		Name string
	}
	type User struct {
		Tags []Tag
	}
	c := &mainCoalescer{}
	WithFieldMergeByID(reflect.TypeOf(User{}), "Tags", "Name")(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.coalesce(reflect.ValueOf(User{Tags: []Tag{{"tag1"}, {"tag2"}}}), reflect.ValueOf(User{Tags: []Tag{{"tag2"}, {"tag3"}}}))
	assert.Equal(t, User{Tags: []Tag{{"tag1"}, {"tag2"}, {"tag3"}}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldMergeByKeyFunc(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := &mainCoalescer{}
	WithFieldMergeByKeyFunc(reflect.TypeOf(User{}), "Tags", SliceUnion)(c)
	assert.NotNil(t, c.fieldCoalescers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.coalesce(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}
