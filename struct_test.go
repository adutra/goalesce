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
	"github.com/stretchr/testify/require"
)

func Test_coalescer_deepMergeStruct(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type foo struct {
			FieldInt int
		}
		type bar struct {
			FieldInt       int
			FieldFoo       foo
			FieldIntPtr    *int
			FieldBarPtr    *bar
			unexported     int
			FieldInterface interface{}
			FieldMap       map[int]string
			FieldMapAtomic map[int]string `goalesce:"atomic"`
		}
		tests := []struct {
			name string
			v1   bar
			v2   bar
			want bar
		}{
			{
				"zeroes",
				bar{},
				bar{},
				bar{},
			},
			{
				"non zeroes int",
				bar{FieldInt: 1},
				bar{FieldInt: 2},
				bar{FieldInt: 2},
			},
			{
				"non zeroes bar",
				bar{FieldFoo: foo{FieldInt: 1}},
				bar{FieldFoo: foo{FieldInt: 2}},
				bar{FieldFoo: foo{FieldInt: 2}},
			},
			{
				"mixed zeroes intptr1",
				bar{FieldIntPtr: intPtr(1)},
				bar{FieldIntPtr: nil},
				bar{FieldIntPtr: intPtr(1)},
			},
			{
				"mixed zeroes intptr2",
				bar{FieldIntPtr: nil},
				bar{FieldIntPtr: intPtr(2)},
				bar{FieldIntPtr: intPtr(2)},
			},
			{
				"mixed zeroes barptr1",
				bar{FieldBarPtr: &bar{FieldInt: 1}},
				bar{FieldBarPtr: nil},
				bar{FieldBarPtr: &bar{FieldInt: 1}},
			},
			{
				"mixed zeroes barptr 2",
				bar{FieldBarPtr: nil},
				bar{FieldBarPtr: &bar{FieldInt: 2}},
				bar{FieldBarPtr: &bar{FieldInt: 2}},
			},
			{
				"non zeroes intptr",
				bar{FieldIntPtr: intPtr(1)},
				bar{FieldIntPtr: intPtr(2)},
				bar{FieldIntPtr: intPtr(2)},
			},
			{
				"non zeroes barptr",
				bar{FieldBarPtr: &bar{FieldInt: 1}},
				bar{FieldBarPtr: &bar{FieldInt: 2}},
				bar{FieldBarPtr: &bar{FieldInt: 2}},
			},
			{
				"non zeroes unexported",
				bar{unexported: 1},
				bar{unexported: 2},
				bar{},
			},
			{
				"field interface nil 1",
				bar{FieldInterface: &foo{FieldInt: 1}},
				bar{FieldInterface: nil},
				bar{FieldInterface: &foo{FieldInt: 1}},
			},
			{
				"field interface nil 2",
				bar{FieldInterface: nil},
				bar{FieldInterface: &foo{FieldInt: 1}},
				bar{FieldInterface: &foo{FieldInt: 1}},
			},
			{
				"field interface same types",
				bar{FieldInterface: &bar{FieldInt: 1, FieldMap: map[int]string{1: "a"}}},
				bar{FieldInterface: &bar{FieldInt: 0, FieldMap: map[int]string{2: "b"}}},
				bar{FieldInterface: &bar{FieldInt: 1, FieldMap: map[int]string{1: "a", 2: "b"}}},
			},
			{
				"field map",
				bar{FieldMap: map[int]string{1: "a", 2: "b"}},
				bar{FieldMap: map[int]string{2: "a", 3: "b"}},
				bar{FieldMap: map[int]string{1: "a", 2: "a", 3: "b"}},
			},
			{
				"field map atomic",
				bar{FieldMapAtomic: map[int]string{1: "a", 2: "b"}},
				bar{FieldMapAtomic: map[int]string{2: "a", 3: "b"}},
				bar{FieldMapAtomic: map[int]string{2: "a", 3: "b"}},
			},
			{
				"non zeroes unexported",
				bar{unexported: 1},
				bar{unexported: 2},
				bar{},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := newCoalescer()
				got, err := c.deepMergeStruct(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
				assertNotSame(t, tt.v1, got.Interface())
				assertNotSame(t, tt.v2, got.Interface())
			})
		}
	})
	t.Run("slice fields", func(t *testing.T) {
		type foo struct {
			FieldInt    int
			FieldIntPtr *int
		}
		type nested struct {
			FieldKey    int
			FieldNonKey string
			FieldInts   []int
		}
		type bar struct {
			FieldInts            []int
			FieldIntsAtomic      []int  `goalesce:"atomic"`
			FieldIntsUnion       []int  `goalesce:"union"`
			FieldIntsAppend      []int  `goalesce:"append"`
			FieldIntsIndex       []int  `goalesce:"index"`
			FieldIntsIndexArray  [3]int `goalesce:"index"`
			FieldFoos            []foo
			FieldFoosAtomic      []foo    `goalesce:"atomic"`
			FieldFoosUnion       []foo    `goalesce:"union"`
			FieldFoosAppend      []foo    `goalesce:"append"`
			FieldFoosIndex       []foo    `goalesce:"index"`
			FieldFoosIndexArray  [3]foo   `goalesce:"index"`
			FieldFoosMergeKey    []foo    `goalesce:"id:FieldInt"`
			FieldFooPtrsMergeKey []*foo   `goalesce:"id:FieldIntPtr"`
			FieldNestedSlice     []nested `goalesce:"id:FieldKey"`
		}
		tests := []struct {
			name string
			v1   bar
			v2   bar
			want bar
		}{
			{
				"slice ints default",
				bar{FieldInts: []int{1, 2}},
				bar{FieldInts: []int{2, 3}},
				bar{FieldInts: []int{2, 3}},
			},
			{
				"slice ints atomic",
				bar{FieldIntsAtomic: []int{1, 2}},
				bar{FieldIntsAtomic: []int{2, 3}},
				bar{FieldIntsAtomic: []int{2, 3}},
			},
			{
				"slice ints union",
				bar{FieldIntsUnion: []int{1, 2}},
				bar{FieldIntsUnion: []int{2, 3}},
				bar{FieldIntsUnion: []int{1, 2, 3}},
			},
			{
				"slice ints append",
				bar{FieldIntsAppend: []int{1, 2}},
				bar{FieldIntsAppend: []int{2, 3}},
				bar{FieldIntsAppend: []int{1, 2, 2, 3}},
			},
			{
				"slice ints index",
				bar{FieldIntsIndex: []int{1, 2, 3}},
				bar{FieldIntsIndex: []int{-1, -2}},
				bar{FieldIntsIndex: []int{-1, -2, 3}},
			},
			{
				"array ints index",
				bar{FieldIntsIndexArray: [3]int{1, 2, 3}},
				bar{FieldIntsIndexArray: [3]int{-1, -2}},
				bar{FieldIntsIndexArray: [3]int{-1, -2, 3}},
			},
			{
				"slice foos default",
				bar{FieldFoos: []foo{{FieldInt: 1}, {FieldInt: 2}}},
				bar{FieldFoos: []foo{{FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoos: []foo{{FieldInt: 2}, {FieldInt: 3}}},
			},
			{
				"slice foos atomic",
				bar{FieldFoosAtomic: []foo{{FieldInt: 1}, {FieldInt: 2}}},
				bar{FieldFoosAtomic: []foo{{FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosAtomic: []foo{{FieldInt: 2}, {FieldInt: 3}}},
			},
			{
				"slice foos union",
				bar{FieldFoosUnion: []foo{{FieldInt: 1}, {FieldInt: 2}}},
				bar{FieldFoosUnion: []foo{{FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosUnion: []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}},
			},
			{
				"slice foos append",
				bar{FieldFoosAppend: []foo{{FieldInt: 1}, {FieldInt: 2}}},
				bar{FieldFoosAppend: []foo{{FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosAppend: []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 2}, {FieldInt: 3}}},
			},
			{
				"slice foos index",
				bar{FieldFoosIndex: []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosIndex: []foo{{FieldInt: -1}, {FieldInt: -2}}},
				bar{FieldFoosIndex: []foo{{FieldInt: -1}, {FieldInt: -2}, {FieldInt: 3}}},
			},
			{
				"array foos index",
				bar{FieldFoosIndexArray: [3]foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosIndexArray: [3]foo{{FieldInt: -1}, {FieldInt: -2}}},
				bar{FieldFoosIndexArray: [3]foo{{FieldInt: -1}, {FieldInt: -2}, {FieldInt: 3}}},
			},
			{
				"slice foos merge key",
				bar{FieldFoosMergeKey: []foo{{FieldInt: 1}, {FieldInt: 2}}},
				bar{FieldFoosMergeKey: []foo{{FieldInt: 2}, {FieldInt: 3}}},
				bar{FieldFoosMergeKey: []foo{{FieldInt: 1}, {FieldInt: 2}, {FieldInt: 3}}},
			},
			{
				"slice foo ptrs merge key",
				bar{FieldFooPtrsMergeKey: []*foo{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}}},
				bar{FieldFooPtrsMergeKey: []*foo{{FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}}},
				bar{FieldFooPtrsMergeKey: []*foo{{FieldIntPtr: intPtr(1)}, {FieldIntPtr: intPtr(2)}, {FieldIntPtr: intPtr(3)}}},
			},
			{
				"nested slice",
				bar{FieldNestedSlice: []nested{{FieldKey: 1, FieldNonKey: "abc", FieldInts: []int{1, 2}}}},
				bar{FieldNestedSlice: []nested{{FieldKey: 1, FieldNonKey: "def", FieldInts: []int{2, 3}}}},
				bar{FieldNestedSlice: []nested{{FieldKey: 1, FieldNonKey: "def", FieldInts: []int{2, 3}}}},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := newCoalescer()
				got, err := c.deepMergeStruct(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
				assertNotSame(t, tt.v1, got.Interface())
				assertNotSame(t, tt.v2, got.Interface())
			})
		}
	})
	t.Run("options", func(t *testing.T) {
		t.Run("field merger", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			fieldMerger := func(v1, v2 reflect.Value) (reflect.Value, error) {
				return reflect.ValueOf([]int{1, 2, 3}), nil
			}
			c := newCoalescer(WithFieldMerger(reflect.TypeOf(foo{}), "FieldInts", fieldMerger))
			got, err := c.deepMergeStruct(reflect.ValueOf(foo{FieldInts: []int{1, 2}}), reflect.ValueOf(foo{FieldInts: []int{2, 3}}))
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 3}}, got.Interface())
		})
		t.Run("field merger provider", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			fieldMerger := func(v1, v2 reflect.Value) (reflect.Value, error) {
				return reflect.ValueOf([]int{1, 2, 3}), nil
			}
			c := newCoalescer(WithFieldMergerProvider(reflect.TypeOf(foo{}), "FieldInts", func(DeepMergeFunc, DeepCopyFunc) DeepMergeFunc {
				return fieldMerger
			}))
			got, err := c.deepMergeStruct(reflect.ValueOf(foo{FieldInts: []int{1, 2}}), reflect.ValueOf(foo{FieldInts: []int{2, 3}}))
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 3}}, got.Interface())
		})
		t.Run("atomic field", func(t *testing.T) {
			type foo struct {
				FieldInts map[int]string
			}
			c := newCoalescer(WithAtomicFieldMerge(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldInts: map[int]string{1: "abc"}}),
				reflect.ValueOf(foo{FieldInts: map[int]string{1: "def"}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: map[int]string{1: "def"}}, got.Interface())
		})
		t.Run("field set-union", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			c := newCoalescer(WithFieldSetUnionMerge(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldInts: []int{1, 2}}),
				reflect.ValueOf(foo{FieldInts: []int{2, 3}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 3}}, got.Interface())
		})
		t.Run("field list-append", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			c := newCoalescer(WithFieldListAppendMerge(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldInts: []int{1, 2}}),
				reflect.ValueOf(foo{FieldInts: []int{2, 3}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 2, 3}}, got.Interface())
		})
		t.Run("field merge-by-index", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			c := newCoalescer(WithFieldMergeByIndex(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldInts: []int{1, 2}}),
				reflect.ValueOf(foo{FieldInts: []int{-1}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{-1, 2}}, got.Interface())
		})
		t.Run("field merge-by-id", func(t *testing.T) {
			type bar struct {
				Name string
			}
			type foo struct {
				FieldBars []bar
			}
			c := newCoalescer(WithFieldMergeByID(reflect.TypeOf(foo{}), "FieldBars", "Name"))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldBars: []bar{{"a"}, {"b"}}}),
				reflect.ValueOf(foo{FieldBars: []bar{{"b"}, {"a"}}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldBars: []bar{{"a"}, {"b"}}}, got.Interface())
		})
		t.Run("field merge-by-key-func", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			c := newCoalescer(WithFieldMergeByKeyFunc(reflect.TypeOf(foo{}), "FieldInts", SliceUnion))
			got, err := c.deepMergeStruct(
				reflect.ValueOf(foo{FieldInts: []int{1, 2}}),
				reflect.ValueOf(foo{FieldInts: []int{2, 3}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 3}}, got.Interface())
		})
	})
	t.Run("tag errors", func(t *testing.T) {
		type foo struct {
			FieldInt int
		}
		type unknownStrategy struct {
			FieldInts []int `goalesce:"unknown"`
		}
		type invalidAppend struct {
			FieldInt int `goalesce:"append"`
		}
		type invalidUnion struct {
			FieldInt int `goalesce:"union"`
		}
		type invalidIndex struct {
			FieldInt int `goalesce:"index"`
		}
		type invalidMerge struct {
			FieldInt int `goalesce:"id"`
		}
		type missingKey struct {
			FieldInts []int `goalesce:"id"`
		}
		type missingKey2 struct {
			FieldInts []int `goalesce:"id:"`
		}
		type missingKey3 struct {
			FieldInts []int `goalesce:"id key"`
		}
		type wrongElemType struct {
			FieldInts []int `goalesce:"id:irrelevant"`
		}
		type unknownField struct {
			FieldFoos    []foo  `goalesce:"id:unknown"`
			FieldFooPtrs []*foo `goalesce:"id:unknown"`
		}
		tests := []struct {
			name string
			v1   interface{}
			v2   interface{}
			want string
		}{
			{
				"unknown strategy",
				unknownStrategy{FieldInts: []int{1, 2}},
				unknownStrategy{FieldInts: []int{2, 3}},
				"field goalesce.unknownStrategy.FieldInts: unknown merge strategy: unknown",
			},
			{
				"invalid append",
				invalidAppend{FieldInt: 1},
				invalidAppend{FieldInt: 2},
				"field goalesce.invalidAppend.FieldInt: append strategy is only supported for slices",
			},
			{
				"invalid union",
				invalidUnion{FieldInt: 1},
				invalidUnion{FieldInt: 2},
				"field goalesce.invalidUnion.FieldInt: union strategy is only supported for slices",
			},
			{
				"invalid index",
				invalidIndex{FieldInt: 1},
				invalidIndex{FieldInt: 2},
				"field goalesce.invalidIndex.FieldInt: index strategy is only supported for slices and arrays",
			},
			{
				"invalid merge",
				invalidMerge{FieldInt: 1},
				invalidMerge{FieldInt: 2},
				"field goalesce.invalidMerge.FieldInt: id strategy is only supported for slices",
			},
			{
				"missing merge key",
				missingKey{FieldInts: []int{1}},
				missingKey{FieldInts: []int{2}},
				"field goalesce.missingKey.FieldInts: id strategy must be followed by a colon and the merge key",
			},
			{
				"missing merge key 2",
				missingKey2{FieldInts: []int{1}},
				missingKey2{FieldInts: []int{2}},
				"field goalesce.missingKey2.FieldInts: id strategy must be followed by a colon and the merge key",
			},
			{
				"missing merge key 3",
				missingKey3{FieldInts: []int{1}},
				missingKey3{FieldInts: []int{2}},
				"field goalesce.missingKey3.FieldInts: id strategy must be followed by a colon and the merge key",
			},
			{
				"wrong element type",
				wrongElemType{FieldInts: []int{1}},
				wrongElemType{FieldInts: []int{2}},
				"field goalesce.wrongElemType.FieldInts: expecting slice of struct or pointer thereto, got: []int",
			},
			{
				"unknown field",
				unknownField{FieldFoos: []foo{{FieldInt: 1}}},
				unknownField{FieldFoos: []foo{{FieldInt: 2}}},
				"field goalesce.unknownField.FieldFoos: slice element type goalesce.foo has no field named unknown",
			},
			{
				"unknown field ptr",
				unknownField{FieldFooPtrs: []*foo{{FieldInt: 1}}},
				unknownField{FieldFooPtrs: []*foo{{FieldInt: 2}}},
				"field goalesce.unknownField.FieldFoos: slice element type goalesce.foo has no field named unknown",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				c := newCoalescer()
				_, err := c.deepMergeStruct(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				assert.EqualError(t, err, tt.want)
			})
		}
	})
	t.Run("field merge errors", func(t *testing.T) {
		type foo struct {
			FieldInterface interface{}
		}
		c := newCoalescer()
		_, err := c.deepMergeStruct(reflect.ValueOf(foo{FieldInterface: 123}), reflect.ValueOf(foo{FieldInterface: "abc"}))
		assert.EqualError(t, err, "types do not match: int != string")
	})
	t.Run("generic error", func(t *testing.T) {
		type foo struct {
			FieldInt int
		}
		c := newCoalescer(withMockDeepMergeError)
		_, err := c.deepMergeStruct(reflect.ValueOf(foo{FieldInt: 1}), reflect.ValueOf(foo{FieldInt: 2}))
		assert.EqualError(t, err, "mock DeepMerge error")
	})
}

func Test_newMergeByField(t *testing.T) {
	type User struct {
		ID   int
		Name *string
	}
	u := User{ID: 1, Name: stringPtr("Alice")}
	t.Run("on struct", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("ID")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(u))
		assert.Equal(t, 1, mergeKey.Interface())
		assert.NoError(t, err)
	})
	t.Run("on pointer to struct", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("ID")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(u))
		assert.Equal(t, 1, mergeKey.Interface())
		assert.NoError(t, err)
	})
	t.Run("on pointer field", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("Name")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(u))
		assert.Equal(t, "Alice", mergeKey.String())
		assert.NoError(t, err)
	})
	t.Run("on pointer to struct and pointer field", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("Name")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(u))
		assert.Equal(t, "Alice", mergeKey.String())
		assert.NoError(t, err)
	})
	t.Run("not a struct", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("Name")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(123))
		assert.False(t, mergeKey.IsValid())
		assert.ErrorContains(t, err, "expecting struct or pointer thereto, got: int")
	})
	t.Run("not a struct, pointer", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("Name")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(intPtr(123)))
		assert.False(t, mergeKey.IsValid())
		assert.ErrorContains(t, err, "expecting struct or pointer thereto, got: *int")
	})
	t.Run("invalid field", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("NonExistent")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(u))
		assert.False(t, mergeKey.IsValid())
		assert.ErrorContains(t, err, "struct type goalesce.User has no field named NonExistent")
	})
	t.Run("invalid field on pointer to struct", func(t *testing.T) {
		mergeKeyFunc := newMergeByField("NonExistent")
		mergeKey, err := mergeKeyFunc(-1, reflect.ValueOf(&u))
		assert.False(t, mergeKey.IsValid())
		assert.ErrorContains(t, err, "struct type goalesce.User has no field named NonExistent")
	})
}

func Test_coalescer_deepCopyStruct(t *testing.T) {
	type Foo struct {
		FieldInt int
		FieldStr string
		FieldPtr *string
	}
	tests := []struct {
		name    string
		v       reflect.Value
		want    reflect.Value
		wantErr assert.ErrorAssertionFunc
		opts    []Option
	}{
		{
			name: "zero",
			v:    reflect.ValueOf(Foo{}),
			want: reflect.ValueOf(Foo{}),
		},
		{
			name: "non zero",
			v: reflect.ValueOf(Foo{
				FieldInt: 123,
				FieldStr: "abc",
				FieldPtr: stringPtr("def"),
			}),
			want: reflect.ValueOf(Foo{
				FieldInt: 123,
				FieldStr: "abc",
				FieldPtr: stringPtr("def"),
			}),
		},
		{
			name:    "error",
			v:       reflect.ValueOf(Foo{FieldInt: 123}),
			wantErr: assert.Error,
			opts:    []Option{withMockDeepCopyError},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCoalescer(tt.opts...)
			got, err := c.deepCopyStruct(tt.v)
			if err == nil {
				assert.Equal(t, tt.want.Interface(), got.Interface())
				assertNotSame(t, tt.v.Interface(), got.Interface())
			} else {
				assert.False(t, got.IsValid())
			}
			if tt.wantErr != nil {
				tt.wantErr(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
