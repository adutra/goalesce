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
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestWithTypeCopier(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(map[string]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(map[string]int{})])
		got, err := c.deepCopy(reflect.ValueOf(map[string]int{"a": 1}))
		assert.Equal(t, map[string]int{"a": 1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("struct", func(t *testing.T) {
		called := false
		type User struct {
			ID string
		}
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(User{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(User{})])
		got, err := c.deepCopy(reflect.ValueOf(User{"Alice"}))
		assert.Equal(t, User{"Alice"}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("pointer", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(intPtr(0)), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(intPtr(0))])
		got, err := c.deepCopy(reflect.ValueOf(intPtr(0)))
		assert.Equal(t, intPtr(0), got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("slice", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf([]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf([]int{})])
		got, err := c.deepCopy(reflect.ValueOf([]int{1}))
		assert.Equal(t, []int{1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)

	})
	t.Run("array", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf([2]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf([2]int{})])
		got, err := c.deepCopy(reflect.ValueOf([2]int{1, 2}))
		assert.Equal(t, [2]int{1, 2}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWithTypeCopierProvider(t *testing.T) {
	called := 0
	c := newCoalescer(
		WithTypeCopierProvider(reflect.TypeOf(map[string]int{}), func(DeepCopyFunc) DeepCopyFunc {
			called++
			return func(v reflect.Value) (reflect.Value, error) {
				called++
				return v, nil
			}
		}))
	assert.NotNil(t, c.typeCopiers[reflect.TypeOf(map[string]int{})])
	got, err := c.deepCopy(reflect.ValueOf(map[string]int{"a": 1}))
	assert.Equal(t, map[string]int{"a": 1}, got.Interface())
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithAtomicCopy(t *testing.T) {
	v := intPtr(1)
	c := newCoalescer(WithAtomicCopy(reflect.TypeOf(v)))
	assert.NotNil(t, c.typeCopiers[reflect.TypeOf(v)])
	got, err := c.deepCopy(reflect.ValueOf(v))
	assert.Same(t, v, got.Interface())
	assert.NoError(t, err)
}

func TestWithAtomicMerge(t *testing.T) {
	v1 := intPtr(1)
	v2 := intPtr(0)
	c := newCoalescer(WithAtomicMerge(reflect.TypeOf(v1)))
	assert.NotNil(t, c.typeMergers[reflect.TypeOf(v1)])
	got, err := c.deepMerge(reflect.ValueOf(v1), reflect.ValueOf(v2))
	assert.Equal(t, v2, got.Interface())
	assert.NotSame(t, v2, got.Interface())
	assert.NoError(t, err)
}

func TestWithTrileanMerge(t *testing.T) {
	c := newCoalescer(WithTrileanMerge())
	assert.NotNil(t, c.typeMergers[reflect.PtrTo(reflect.TypeOf(false))])
	got, err := c.deepMerge(reflect.ValueOf(boolPtr(true)), reflect.ValueOf(boolPtr(false)))
	assert.Equal(t, boolPtr(false), got.Interface())
	assert.NoError(t, err)
}

func TestWithTypeMerger(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeMerger(reflect.TypeOf(map[string]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
				called = true
				return v2, nil
			}))
		assert.NotNil(t, c.typeMergers[reflect.TypeOf(map[string]int{})])
		got, err := c.deepMerge(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int{"b": 2}))
		assert.Equal(t, map[string]int{"b": 2}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("struct", func(t *testing.T) {
		called := false
		type User struct {
			ID string
		}
		c := newCoalescer(
			WithTypeMerger(reflect.TypeOf(User{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
				called = true
				return v2, nil
			}))
		assert.NotNil(t, c.typeMergers[reflect.TypeOf(User{})])
		got, err := c.deepMerge(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
		assert.Equal(t, User{"Bob"}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("pointer", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeMerger(reflect.TypeOf(intPtr(0)), func(v1, v2 reflect.Value) (reflect.Value, error) {
				called = true
				return v2, nil
			}))
		assert.NotNil(t, c.typeMergers[reflect.TypeOf(intPtr(0))])
		got, err := c.deepMerge(reflect.ValueOf(intPtr(1)), reflect.ValueOf(intPtr(0)))
		assert.Equal(t, intPtr(0), got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("slice", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeMerger(reflect.TypeOf([]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
				called = true
				return v2, nil
			}))
		assert.NotNil(t, c.typeMergers[reflect.TypeOf([]int{})])
		got, err := c.deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int{2}))
		assert.Equal(t, []int{2}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("array", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeMerger(reflect.TypeOf([2]int{}), func(v1, v2 reflect.Value) (reflect.Value, error) {
				called = true
				return v2, nil
			}))
		assert.NotNil(t, c.typeMergers[reflect.TypeOf([2]int{})])
		got, err := c.deepMerge(reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{2, 3}))
		assert.Equal(t, [2]int{2, 3}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestMergeZeroValueWithTypeCopier(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(map[string]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(map[string]int{})])
		got, err := c.deepMerge(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int(nil)))
		assert.Equal(t, map[string]int{"a": 1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf(map[string]int(nil)), reflect.ValueOf(map[string]int{"a": 1}))
		assert.Equal(t, map[string]int{"a": 1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("slice", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf([]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf([]int{})])
		got, err := c.deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int(nil)))
		assert.Equal(t, []int{1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf([]int(nil)), reflect.ValueOf([]int{1}))
		assert.Equal(t, []int{1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("struct", func(t *testing.T) {
		called := false
		type User struct {
			ID string
		}
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(User{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(User{})])
		got, err := c.deepMerge(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{}))
		assert.Equal(t, User{"Alice"}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf(User{}), reflect.ValueOf(User{"Alice"}))
		assert.Equal(t, User{"Alice"}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("struct with field merger", func(t *testing.T) {
		// edge case: struct has field merger, but there is also a type copier for it:
		// type copier will not be called when merging zero with non-zero
		called := false
		type User struct {
			ID string `goalesce:"atomic"`
		}
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(User{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}))
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(User{})])
		got, err := c.deepMerge(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{}))
		assert.Equal(t, User{"Alice"}, got.Interface())
		assert.NoError(t, err)
		assert.False(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf(User{}), reflect.ValueOf(User{"Alice"}))
		assert.Equal(t, User{"Alice"}, got.Interface())
		assert.NoError(t, err)
		assert.False(t, called)
	})
	t.Run("pointer", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf(intPtr(0)), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}),
		)
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf(intPtr(0))])
		got, err := c.deepMerge(reflect.ValueOf((*int)(nil)), reflect.ValueOf(intPtr(1)))
		assert.Equal(t, intPtr(1), got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf(intPtr(1)), reflect.ValueOf((*int)(nil)))
		assert.Equal(t, intPtr(1), got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("slice", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf([]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}),
		)
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf([]int{})])
		got, err := c.deepMerge(reflect.ValueOf([]int{1}), reflect.ValueOf([]int(nil)))
		assert.Equal(t, []int{1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf([]int(nil)), reflect.ValueOf([]int{1}))
		assert.Equal(t, []int{1}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
	t.Run("array", func(t *testing.T) {
		called := false
		c := newCoalescer(
			WithTypeCopier(reflect.TypeOf([2]int{}), func(v reflect.Value) (reflect.Value, error) {
				called = true
				return v, nil
			}),
		)
		assert.NotNil(t, c.typeCopiers[reflect.TypeOf([2]int{})])
		got, err := c.deepMerge(reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{}))
		assert.Equal(t, [2]int{1, 2}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
		called = false
		got, err = c.deepMerge(reflect.ValueOf([2]int{}), reflect.ValueOf([2]int{1, 2}))
		assert.Equal(t, [2]int{1, 2}, got.Interface())
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestWithTypeMergerProvider(t *testing.T) {
	called := 0
	c := newCoalescer(
		WithTypeMergerProvider(reflect.TypeOf(map[string]int{}), func(DeepMergeFunc, DeepCopyFunc) DeepMergeFunc {
			called++
			return func(v1, v2 reflect.Value) (reflect.Value, error) {
				called++
				return v2, nil
			}
		}))
	assert.NotNil(t, c.typeMergers[reflect.TypeOf(map[string]int{})])
	got, err := c.deepMerge(reflect.ValueOf(map[string]int{"a": 1}), reflect.ValueOf(map[string]int{"b": 2}))
	assert.Equal(t, map[string]int{"b": 2}, got.Interface())
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithFieldMerger(t *testing.T) {
	type User struct {
		ID string
	}
	called := false
	c := newCoalescer(
		WithFieldMerger(reflect.TypeOf(User{}), "ID", func(v1, v2 reflect.Value) (reflect.Value, error) {
			called = true
			return v2, nil
		}))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["ID"])
	got, err := c.deepMerge(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
	assert.Equal(t, User{"Bob"}, got.Interface())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithFieldMergerProvider(t *testing.T) {
	type User struct {
		ID string
	}
	called := 0
	c := newCoalescer(
		WithFieldMergerProvider(reflect.TypeOf(User{}), "ID", func(DeepMergeFunc, DeepCopyFunc) DeepMergeFunc {
			called++
			return func(v1, v2 reflect.Value) (reflect.Value, error) {
				called++
				return v2, nil
			}
		}))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["ID"])
	got, err := c.deepMerge(reflect.ValueOf(User{"Alice"}), reflect.ValueOf(User{"Bob"}))
	assert.Equal(t, User{"Bob"}, got.Interface())
	assert.NoError(t, err)
	assert.Equal(t, 2, called)
}

func TestWithAtomicFieldMerge(t *testing.T) {
	t.Run("struct field", func(t *testing.T) {
		type Uuid struct {
			Msb int64
			Lsb int64
		}
		type User struct {
			ID Uuid
		}
		c := newCoalescer(WithAtomicFieldMerge(reflect.TypeOf(User{}), "ID"))
		assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["ID"])
		got, err := c.deepMerge(reflect.ValueOf(User{ID: Uuid{Msb: 123}}), reflect.ValueOf(User{ID: Uuid{Lsb: 456}}))
		assert.Equal(t, User{ID: Uuid{Lsb: 456}}, got.Interface())
		assert.NoError(t, err)
	})
	t.Run("pointer field", func(t *testing.T) {
		type User struct {
			ID *int
		}
		c := newCoalescer(WithAtomicFieldMerge(reflect.TypeOf(User{}), "ID"))
		assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["ID"])
		u1 := User{intPtr(1)}
		u2 := User{intPtr(0)}
		got, err := c.deepMerge(reflect.ValueOf(u1), reflect.ValueOf(u2))
		assert.Equal(t, u2, got.Interface())
		assert.NotSame(t, u2.ID, (got.Interface().(User)).ID)
		assert.NoError(t, err)
	})
}

func TestWithDefaultSliceListAppendMerge(t *testing.T) {
	c := newCoalescer(WithDefaultSliceListAppendMerge())
	assert.NotNil(t, c.sliceMerger)
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultSliceMergeByIndex(t *testing.T) {
	c := newCoalescer(WithDefaultSliceMergeByIndex())
	assert.NotNil(t, c.sliceMerger)
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{-1}))
	assert.Equal(t, []int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultArrayMergeByIndex(t *testing.T) {
	c := newCoalescer(WithDefaultArrayMergeByIndex())
	assert.NotNil(t, c.arrayMerger)
	got, err := c.deepMerge(reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{-1}))
	assert.Equal(t, [2]int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithDefaultSliceSetUnionMerge(t *testing.T) {
	c := newCoalescer(WithDefaultSliceSetUnionMerge())
	assert.NotNil(t, c.sliceMerger)
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithErrorOnCycle(t *testing.T) {
	c := newCoalescer(WithErrorOnCycle())
	assert.Equal(t, true, c.errorOnCycle)
}

func TestWithSliceListAppendMerge(t *testing.T) {
	c := newCoalescer(WithSliceListAppendMerge(reflect.TypeOf([]int{})))
	assert.NotNil(t, c.sliceMergers[reflect.TypeOf([]int{})])
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithSliceSetUnionMerge(t *testing.T) {
	c := newCoalescer(WithSliceSetUnionMerge(reflect.TypeOf([]int{})))
	assert.NotNil(t, c.sliceMergers[reflect.TypeOf([]int{})])
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{2, 3}))
	assert.Equal(t, []int{1, 2, 3}, got.Interface())
	assert.NoError(t, err)
}

func TestWithSliceMergeByIndex(t *testing.T) {
	c := newCoalescer(WithSliceMergeByIndex(reflect.TypeOf([]int{})))
	assert.NotNil(t, c.sliceMergers[reflect.TypeOf([]int{})])
	got, err := c.deepMerge(reflect.ValueOf([]int{1, 2}), reflect.ValueOf([]int{-1}))
	assert.Equal(t, []int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithArrayMergeByIndex(t *testing.T) {
	c := newCoalescer(WithArrayMergeByIndex(reflect.TypeOf([2]int{})))
	assert.NotNil(t, c.arrayMergers[reflect.TypeOf([2]int{})])
	got, err := c.deepMerge(reflect.ValueOf([2]int{1, 2}), reflect.ValueOf([2]int{-1}))
	assert.Equal(t, [2]int{-1, 2}, got.Interface())
	assert.NoError(t, err)
}

func TestWithSliceMergeByKeyFunc(t *testing.T) {
	type User struct {
		ID string
	}
	called := false
	mergeKeyFunc := func(index int, element reflect.Value) (key reflect.Value, err error) {
		called = true
		return element.FieldByName("ID"), nil
	}
	c := newCoalescer(WithSliceMergeByKeyFunc(reflect.TypeOf([]User{}), mergeKeyFunc))
	assert.NotNil(t, c.sliceMergers[reflect.TypeOf([]User{})])
	got, err := c.deepMerge(reflect.ValueOf([]User{{"Alice"}, {"Bob"}}), reflect.ValueOf([]User{{"Bob"}, {"Alice"}}))
	assert.Equal(t, []User{{"Alice"}, {"Bob"}}, got.Interface())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestWithSliceMergeByID(t *testing.T) {
	type User struct {
		ID string
	}
	c := newCoalescer(WithSliceMergeByID(reflect.TypeOf([]User{}), "ID"))
	assert.NotNil(t, c.sliceMergers[reflect.TypeOf([]User{})])
	got, err := c.deepMerge(reflect.ValueOf([]User{{"Alice"}, {"Bob"}}), reflect.ValueOf([]User{{"Bob"}, {"Alice"}}))
	assert.Equal(t, []User{{"Alice"}, {"Bob"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithZeroEmptySliceMerge(t *testing.T) {
	c := newCoalescer(WithZeroEmptySliceMerge())
	assert.Equal(t, true, c.zeroEmptySlice)
}

func TestWithFieldListAppendMerge(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := newCoalescer(WithFieldListAppendMerge(reflect.TypeOf(User{}), "Tags"))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.deepMerge(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldSetUnionMerge(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := newCoalescer(WithFieldSetUnionMerge(reflect.TypeOf(User{}), "Tags"))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.deepMerge(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldMergeByIndex(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := newCoalescer(WithFieldMergeByIndex(reflect.TypeOf(User{}), "Tags"))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.deepMerge(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag1a"}}))
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
	c := newCoalescer(WithFieldMergeByID(reflect.TypeOf(User{}), "Tags", "Name"))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.deepMerge(reflect.ValueOf(User{Tags: []Tag{{"tag1"}, {"tag2"}}}), reflect.ValueOf(User{Tags: []Tag{{"tag2"}, {"tag3"}}}))
	assert.Equal(t, User{Tags: []Tag{{"tag1"}, {"tag2"}, {"tag3"}}}, got.Interface())
	assert.NoError(t, err)
}

func TestWithFieldMergeByKeyFunc(t *testing.T) {
	type User struct {
		Tags []string
	}
	c := newCoalescer(WithFieldMergeByKeyFunc(reflect.TypeOf(User{}), "Tags", SliceUnion))
	assert.NotNil(t, c.fieldMergers[reflect.TypeOf(User{})]["Tags"])
	got, err := c.deepMerge(reflect.ValueOf(User{Tags: []string{"tag1", "tag2"}}), reflect.ValueOf(User{Tags: []string{"tag2", "tag3"}}))
	assert.Equal(t, User{Tags: []string{"tag1", "tag2", "tag3"}}, got.Interface())
	assert.NoError(t, err)
}
