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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestNewStructCoalescer(t *testing.T) {
	t.Run("no opts", func(t *testing.T) {
		got := NewStructCoalescer()
		assert.Equal(t, &structCoalescer{fallback: &atomicCoalescer{}}, got)
	})
	t.Run("with generic option", func(t *testing.T) {
		var passed *structCoalescer
		opt := func(c *structCoalescer) {
			passed = c
		}
		returned := NewStructCoalescer(opt)
		assert.Equal(t, &structCoalescer{fallback: &atomicCoalescer{}}, returned)
		assert.Equal(t, returned, passed)
	})
	t.Run("with field coalescer", func(t *testing.T) {
		type foo struct {
			Int int
		}
		m := newMockCoalescer(t)
		m.On("WithFallback", mock.Anything).Return()
		actual := NewStructCoalescer(WithFieldCoalescer(reflect.TypeOf(foo{}), "Int", m))
		assert.Equal(
			t,
			&structCoalescer{
				fallback: &atomicCoalescer{},
				fieldCoalescers: map[reflect.Type]map[string]Coalescer{
					reflect.TypeOf(foo{}): {"Int": m},
				},
			},
			actual)
		main := NewMainCoalescer()
		actual.WithFallback(main)
		m.AssertCalled(t, "WithFallback", main)
	})
}

func TestWithFieldCoalescer(t *testing.T) {
	type User struct {
		Id string
	}
	c := &structCoalescer{}
	m := &mockCoalescer{}
	WithFieldCoalescer(reflect.TypeOf(User{}), "Id", m)(c)
	expected := map[reflect.Type]map[string]Coalescer{reflect.TypeOf(User{}): {"Id": m}}
	assert.Equal(t, expected, c.fieldCoalescers)
}

func TestWithAtomicField(t *testing.T) {
	type User struct {
		Id string
	}
	c := &structCoalescer{}
	WithAtomicField(reflect.TypeOf(User{}), "Id")(c)
	expected := map[reflect.Type]map[string]Coalescer{reflect.TypeOf(User{}): {"Id": &atomicCoalescer{}}}
	assert.Equal(t, expected, c.fieldCoalescers)
}

func Test_structCoalescer_Coalesce(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		type foo struct {
			Int int
		}
		type bar struct {
			Int        int
			Foo        foo
			IntPtr     *int
			BarPtr     *bar
			unexported int
			Interface  interface{}
			Map        map[int]string
			MapAtomic  map[int]string `goalesce:"atomic"`
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
				bar{Int: 1},
				bar{Int: 2},
				bar{Int: 2},
			},
			{
				"non zeroes bar",
				bar{Foo: foo{Int: 1}},
				bar{Foo: foo{Int: 2}},
				bar{Foo: foo{Int: 2}},
			},
			{
				"mixed zeroes intptr1",
				bar{IntPtr: intPtr(1)},
				bar{IntPtr: nil},
				bar{IntPtr: intPtr(1)},
			},
			{
				"mixed zeroes intptr2",
				bar{IntPtr: nil},
				bar{IntPtr: intPtr(2)},
				bar{IntPtr: intPtr(2)},
			},
			{
				"mixed zeroes barptr1",
				bar{BarPtr: &bar{Int: 1}},
				bar{BarPtr: nil},
				bar{BarPtr: &bar{Int: 1}},
			},
			{
				"mixed zeroes barptr 2",
				bar{BarPtr: nil},
				bar{BarPtr: &bar{Int: 2}},
				bar{BarPtr: &bar{Int: 2}},
			},
			{
				"non zeroes intptr",
				bar{IntPtr: intPtr(1)},
				bar{IntPtr: intPtr(2)},
				bar{IntPtr: intPtr(2)},
			},
			{
				"non zeroes barptr",
				bar{BarPtr: &bar{Int: 1}},
				bar{BarPtr: &bar{Int: 2}},
				bar{BarPtr: &bar{Int: 2}},
			},
			{
				"non zeroes unexported",
				bar{unexported: 1},
				bar{unexported: 2},
				bar{},
			},
			{
				"field interface different types",
				bar{Interface: 1},
				bar{Interface: "abc"},
				bar{Interface: "abc"},
			},
			{
				"field interface nil 1",
				bar{Interface: &foo{Int: 1}},
				bar{Interface: nil},
				bar{Interface: &foo{Int: 1}},
			},
			{
				"field interface nil 2",
				bar{Interface: nil},
				bar{Interface: &foo{Int: 1}},
				bar{Interface: &foo{Int: 1}},
			},
			{
				"field interface same types",
				bar{Interface: &bar{Int: 1, Map: map[int]string{1: "a"}}},
				bar{Interface: &bar{Int: 0, Map: map[int]string{2: "b"}}},
				bar{Interface: &bar{Int: 1, Map: map[int]string{1: "a", 2: "b"}}},
			},
			{
				"field map",
				bar{Map: map[int]string{1: "a", 2: "b"}},
				bar{Map: map[int]string{2: "a", 3: "b"}},
				bar{Map: map[int]string{1: "a", 2: "a", 3: "b"}},
			},
			{
				"field map atomic",
				bar{MapAtomic: map[int]string{1: "a", 2: "b"}},
				bar{MapAtomic: map[int]string{2: "a", 3: "b"}},
				bar{MapAtomic: map[int]string{2: "a", 3: "b"}},
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
				coalescer := NewStructCoalescer()
				coalescer.WithFallback(NewMainCoalescer())
				got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
			})
		}
	})
	t.Run("slice fields", func(t *testing.T) {
		type foo struct {
			Int    int
			IntPtr *int
		}
		type nested struct {
			Key    int
			NonKey string
			Ints   []int
		}
		type bar struct {
			Ints            []int
			IntsAtomic      []int `goalesce:"atomic"`
			IntsUnion       []int `goalesce:"union"`
			IntsAppend      []int `goalesce:"append"`
			IntsIndex       []int `goalesce:"index"`
			Foos            []foo
			FoosAtomic      []foo    `goalesce:"atomic"`
			FoosUnion       []foo    `goalesce:"union"`
			FoosAppend      []foo    `goalesce:"append"`
			FoosIndex       []foo    `goalesce:"index"`
			FoosMergeKey    []foo    `goalesce:"merge,Int"`
			FooPtrsMergeKey []*foo   `goalesce:"merge,IntPtr"`
			NestedSlice     []nested `goalesce:"merge,Key"`
		}
		tests := []struct {
			name string
			v1   bar
			v2   bar
			want bar
		}{
			{
				"slice ints default",
				bar{Ints: []int{1, 2}},
				bar{Ints: []int{2, 3}},
				bar{Ints: []int{2, 3}},
			},
			{
				"slice ints atomic",
				bar{IntsAtomic: []int{1, 2}},
				bar{IntsAtomic: []int{2, 3}},
				bar{IntsAtomic: []int{2, 3}},
			},
			{
				"slice ints union",
				bar{IntsUnion: []int{1, 2}},
				bar{IntsUnion: []int{2, 3}},
				bar{IntsUnion: []int{1, 2, 3}},
			},
			{
				"slice ints append",
				bar{IntsAppend: []int{1, 2}},
				bar{IntsAppend: []int{2, 3}},
				bar{IntsAppend: []int{1, 2, 2, 3}},
			},
			{
				"slice ints index",
				bar{IntsIndex: []int{1, 2, 3}},
				bar{IntsIndex: []int{-1, -2}},
				bar{IntsIndex: []int{-1, -2, 3}},
			},
			{
				"slice foos default",
				bar{Foos: []foo{{Int: 1}, {Int: 2}}},
				bar{Foos: []foo{{Int: 2}, {Int: 3}}},
				bar{Foos: []foo{{Int: 2}, {Int: 3}}},
			},
			{
				"slice foos atomic",
				bar{FoosAtomic: []foo{{Int: 1}, {Int: 2}}},
				bar{FoosAtomic: []foo{{Int: 2}, {Int: 3}}},
				bar{FoosAtomic: []foo{{Int: 2}, {Int: 3}}},
			},
			{
				"slice foos union",
				bar{FoosUnion: []foo{{Int: 1}, {Int: 2}}},
				bar{FoosUnion: []foo{{Int: 2}, {Int: 3}}},
				bar{FoosUnion: []foo{{Int: 1}, {Int: 2}, {Int: 3}}},
			},
			{
				"slice foos append",
				bar{FoosAppend: []foo{{Int: 1}, {Int: 2}}},
				bar{FoosAppend: []foo{{Int: 2}, {Int: 3}}},
				bar{FoosAppend: []foo{{Int: 1}, {Int: 2}, {Int: 2}, {Int: 3}}},
			},
			{
				"slice foos index",
				bar{FoosIndex: []foo{{Int: 1}, {Int: 2}, {Int: 3}}},
				bar{FoosIndex: []foo{{Int: -1}, {Int: -2}}},
				bar{FoosIndex: []foo{{Int: -1}, {Int: -2}, {Int: 3}}},
			},
			{
				"slice foos merge key",
				bar{FoosMergeKey: []foo{{Int: 1}, {Int: 2}}},
				bar{FoosMergeKey: []foo{{Int: 2}, {Int: 3}}},
				bar{FoosMergeKey: []foo{{Int: 1}, {Int: 2}, {Int: 3}}},
			},
			{
				"slice foo ptrs merge key",
				bar{FooPtrsMergeKey: []*foo{{IntPtr: intPtr(1)}, {IntPtr: intPtr(2)}}},
				bar{FooPtrsMergeKey: []*foo{{IntPtr: intPtr(2)}, {IntPtr: intPtr(3)}}},
				bar{FooPtrsMergeKey: []*foo{{IntPtr: intPtr(1)}, {IntPtr: intPtr(2)}, {IntPtr: intPtr(3)}}},
			},
			{
				"nested slice",
				bar{NestedSlice: []nested{{Key: 1, NonKey: "abc", Ints: []int{1, 2}}}},
				bar{NestedSlice: []nested{{Key: 1, NonKey: "def", Ints: []int{2, 3}}}},
				bar{NestedSlice: []nested{{Key: 1, NonKey: "def", Ints: []int{2, 3}}}},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				coalescer := NewStructCoalescer()
				coalescer.WithFallback(NewMainCoalescer())
				got, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got.Interface())
			})
		}
	})
	t.Run("options", func(t *testing.T) {
		t.Run("field coalescer", func(t *testing.T) {
			type foo struct {
				Ints []int
			}
			coalescer := NewStructCoalescer(WithFieldCoalescer(reflect.TypeOf(foo{}), "Ints", &sliceAppendCoalescer{}))
			got, err := coalescer.Coalesce(reflect.ValueOf(foo{Ints: []int{1, 2}}), reflect.ValueOf(foo{Ints: []int{2, 3}}))
			require.NoError(t, err)
			assert.Equal(t, foo{Ints: []int{1, 2, 2, 3}}, got.Interface())
		})
		t.Run("atomic field", func(t *testing.T) {
			type foo struct {
				Ints map[int]string
			}
			coalescer := NewStructCoalescer(WithAtomicField(reflect.TypeOf(foo{}), "Ints"))
			got, err := coalescer.Coalesce(
				reflect.ValueOf(foo{Ints: map[int]string{1: "abc"}}),
				reflect.ValueOf(foo{Ints: map[int]string{1: "def"}}),
			)
			require.NoError(t, err)
			assert.Equal(t, foo{Ints: map[int]string{1: "def"}}, got.Interface())
		})
	})
	t.Run("tag errors", func(t *testing.T) {
		type foo struct {
			Int int
		}
		type unknownStrategy struct {
			Ints []int `goalesce:"unknown"`
		}
		type invalidAppend struct {
			Int int `goalesce:"append"`
		}
		type invalidUnion struct {
			Int int `goalesce:"union"`
		}
		type invalidIndex struct {
			Int int `goalesce:"index"`
		}
		type invalidMerge struct {
			Int int `goalesce:"merge"`
		}
		type missingKey struct {
			Ints []int `goalesce:"merge"`
		}
		type missingKey2 struct {
			Ints []int `goalesce:"merge,"`
		}
		type missingKey3 struct {
			Ints []int `goalesce:"merge key"`
		}
		type wrongElemType struct {
			Ints []int `goalesce:"merge,irrelevant"`
		}
		type unknownField struct {
			Foos    []foo  `goalesce:"merge,unknown"`
			FooPtrs []*foo `goalesce:"merge,unknown"`
		}
		tests := []struct {
			name string
			v1   interface{}
			v2   interface{}
			want string
		}{
			{
				"unknown strategy",
				unknownStrategy{Ints: []int{1, 2}},
				unknownStrategy{Ints: []int{2, 3}},
				"field goalesce.unknownStrategy.Ints: unknown coalesce strategy: unknown",
			},
			{
				"invalid append",
				invalidAppend{Int: 1},
				invalidAppend{Int: 2},
				"field goalesce.invalidAppend.Int: append strategy is only supported for slices",
			},
			{
				"invalid union",
				invalidUnion{Int: 1},
				invalidUnion{Int: 2},
				"field goalesce.invalidUnion.Int: union strategy is only supported for slices",
			},
			{
				"invalid index",
				invalidIndex{Int: 1},
				invalidIndex{Int: 2},
				"field goalesce.invalidIndex.Int: index strategy is only supported for slices",
			},
			{
				"invalid merge",
				invalidMerge{Int: 1},
				invalidMerge{Int: 2},
				"field goalesce.invalidMerge.Int: merge strategy is only supported for slices",
			},
			{
				"missing merge key",
				missingKey{Ints: []int{1}},
				missingKey{Ints: []int{2}},
				"field goalesce.missingKey.Ints: merge strategy must be followed by a comma and the merge key",
			},
			{
				"missing merge key 2",
				missingKey2{Ints: []int{1}},
				missingKey2{Ints: []int{2}},
				"field goalesce.missingKey2.Ints: merge strategy must be followed by a comma and the merge key",
			},
			{
				"missing merge key 3",
				missingKey3{Ints: []int{1}},
				missingKey3{Ints: []int{2}},
				"field goalesce.missingKey3.Ints: merge strategy must be followed by a comma and the merge key",
			},
			{
				"wrong element type",
				wrongElemType{Ints: []int{1}},
				wrongElemType{Ints: []int{2}},
				"field goalesce.wrongElemType.Ints: expecting slice of struct or pointer thereto, got: []int",
			},
			{
				"unknown field",
				unknownField{Foos: []foo{{Int: 1}}},
				unknownField{Foos: []foo{{Int: 2}}},
				"field goalesce.unknownField.Foos: slice element type goalesce.foo has no field named unknown",
			},
			{
				"unknown field ptr",
				unknownField{FooPtrs: []*foo{{Int: 1}}},
				unknownField{FooPtrs: []*foo{{Int: 2}}},
				"field goalesce.unknownField.Foos: slice element type goalesce.foo has no field named unknown",
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				coalescer := NewStructCoalescer()
				_, err := coalescer.Coalesce(reflect.ValueOf(tt.v1), reflect.ValueOf(tt.v2))
				assert.EqualError(t, err, tt.want)
			})
		}
	})
	t.Run("type errors", func(t *testing.T) {
		type foo struct {
			Int int
		}
		type bar struct {
			Int int
		}
		_, err := NewStructCoalescer().Coalesce(reflect.ValueOf(foo{}), reflect.ValueOf(bar{}))
		assert.EqualError(t, err, "types do not match: goalesce.foo != goalesce.bar")
		_, err = NewStructCoalescer().Coalesce(reflect.ValueOf(1), reflect.ValueOf(2))
		assert.EqualError(t, err, "values have wrong kind: expected struct, got int")
	})
	t.Run("fallback error", func(t *testing.T) {
		type foo struct {
			Int int
		}
		m := newMockCoalescer(t)
		m.On("Coalesce", mock.Anything, mock.Anything).Return(reflect.Value{}, errors.New("fake"))
		coalescer := NewStructCoalescer()
		coalescer.WithFallback(m)
		_, err := coalescer.Coalesce(reflect.ValueOf(foo{Int: 1}), reflect.ValueOf(foo{Int: 2}))
		assert.EqualError(t, err, "fake")
	})
}
