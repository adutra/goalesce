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
	"github.com/stretchr/testify/require"
)

func Test_mainCoalescer_coalesceStruct(t *testing.T) {
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
				"field interface different types",
				bar{FieldInterface: 1},
				bar{FieldInterface: "abc"},
				bar{FieldInterface: "abc"},
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
				coalescer := NewCoalescer()
				got, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
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
			FieldIntsAtomic      []int `goalesce:"atomic"`
			FieldIntsUnion       []int `goalesce:"union"`
			FieldIntsAppend      []int `goalesce:"append"`
			FieldIntsIndex       []int `goalesce:"index"`
			FieldFoos            []foo
			FieldFoosAtomic      []foo    `goalesce:"atomic"`
			FieldFoosUnion       []foo    `goalesce:"union"`
			FieldFoosAppend      []foo    `goalesce:"append"`
			FieldFoosIndex       []foo    `goalesce:"index"`
			FieldFoosMergeKey    []foo    `goalesce:"merge,FieldInt"`
			FieldFooPtrsMergeKey []*foo   `goalesce:"merge,FieldIntPtr"`
			FieldNestedSlice     []nested `goalesce:"merge,FieldKey"`
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
				coalescer := NewCoalescer()
				got, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
			})
		}
	})
	t.Run("options", func(t *testing.T) {
		t.Run("field coalescer", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			coalescer := NewCoalescer(WithFieldCoalescer(reflect.TypeOf(foo{}), "FieldInts", coalesceSliceAppend))
			got, err := coalescer(reflect.ValueOf(foo{FieldInts: []int{1, 2}}), reflect.ValueOf(foo{FieldInts: []int{2, 3}}))
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 2, 3}}, got.Interface())
		})
		t.Run("field coalescer provider", func(t *testing.T) {
			type foo struct {
				FieldInts []int
			}
			coalescer := NewCoalescer(WithFieldCoalescerProvider(reflect.TypeOf(foo{}), "FieldInts", func(parent Coalescer) Coalescer {
				return coalesceSliceAppend
			}))
			got, err := coalescer(reflect.ValueOf(foo{FieldInts: []int{1, 2}}), reflect.ValueOf(foo{FieldInts: []int{2, 3}}))
			require.NoError(t, err)
			assert.Equal(t, foo{FieldInts: []int{1, 2, 2, 3}}, got.Interface())
		})
		t.Run("atomic field", func(t *testing.T) {
			type foo struct {
				FieldInts map[int]string
			}
			coalescer := NewCoalescer(WithAtomicField(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := coalescer(
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
			coalescer := NewCoalescer(WithFieldSetUnion(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := coalescer(
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
			coalescer := NewCoalescer(WithFieldListAppend(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := coalescer(
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
			coalescer := NewCoalescer(WithFieldMergeByIndex(reflect.TypeOf(foo{}), "FieldInts"))
			got, err := coalescer(
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
			coalescer := NewCoalescer(WithFieldMergeByID(reflect.TypeOf(foo{}), "FieldBars", "Name"))
			got, err := coalescer(
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
			coalescer := NewCoalescer(WithFieldMergeByKeyFunc(reflect.TypeOf(foo{}), "FieldInts", SliceUnion))
			got, err := coalescer(
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
			FieldInt int `goalesce:"merge"`
		}
		type missingKey struct {
			FieldInts []int `goalesce:"merge"`
		}
		type missingKey2 struct {
			FieldInts []int `goalesce:"merge,"`
		}
		type missingKey3 struct {
			FieldInts []int `goalesce:"merge key"`
		}
		type wrongElemType struct {
			FieldInts []int `goalesce:"merge,irrelevant"`
		}
		type unknownField struct {
			FieldFoos    []foo  `goalesce:"merge,unknown"`
			FieldFooPtrs []*foo `goalesce:"merge,unknown"`
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
				"field goalesce.unknownStrategy.FieldInts: unknown coalesce strategy: unknown",
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
				"field goalesce.invalidIndex.FieldInt: index strategy is only supported for slices",
			},
			{
				"invalid merge",
				invalidMerge{FieldInt: 1},
				invalidMerge{FieldInt: 2},
				"field goalesce.invalidMerge.FieldInt: merge strategy is only supported for slices",
			},
			{
				"missing merge key",
				missingKey{FieldInts: []int{1}},
				missingKey{FieldInts: []int{2}},
				"field goalesce.missingKey.FieldInts: merge strategy must be followed by a comma and the merge key",
			},
			{
				"missing merge key 2",
				missingKey2{FieldInts: []int{1}},
				missingKey2{FieldInts: []int{2}},
				"field goalesce.missingKey2.FieldInts: merge strategy must be followed by a comma and the merge key",
			},
			{
				"missing merge key 3",
				missingKey3{FieldInts: []int{1}},
				missingKey3{FieldInts: []int{2}},
				"field goalesce.missingKey3.FieldInts: merge strategy must be followed by a comma and the merge key",
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
				coalescer := NewCoalescer()
				_, err := coalescer(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				assert.EqualError(t, err, tt.want)
			})
		}
	})
	t.Run("fallback error", func(t *testing.T) {
		type foo struct {
			FieldInt int
		}
		coalescer := NewCoalescer(WithTypeCoalescer(reflect.TypeOf(0), func(v1, v2 reflect.Value) (reflect.Value, error) {
			return reflect.Value{}, errors.New("fake")
		}))
		_, err := coalescer(reflect.ValueOf(foo{FieldInt: 1}), reflect.ValueOf(foo{FieldInt: 2}))
		assert.EqualError(t, err, "fake")
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
