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

package goalesce_test

import (
	"fmt"
	"github.com/adutra/goalesce"
	"reflect"
)

func Example() {
	// Coalescing scalars
	coalesced, _ := goalesce.Coalesce("abc", "def")
	fmt.Printf("Coalesce(abc, def) = %v\n", coalesced)
	coalesced, _ = goalesce.Coalesce(1, 0)
	fmt.Printf("Coalesce(1, 0) = %v\n", coalesced)

	// Coalescing structs
	type Foo struct {
		Field1 int
		Field2 string
	}
	coalesced, _ = goalesce.Coalesce(&Foo{Field1: 1}, &Foo{Field2: "abc"})
	fmt.Printf("Coalesce(&{1}, &{abc}) = %v\n", coalesced)

	// Coalescing maps
	coalesced, _ = goalesce.Coalesce(map[int]string{1: "a", 2: "b"}, map[int]string{2: "c", 3: "d"})
	fmt.Printf("Coalesce(map[1:a 2:b], map[2:c 3:d]) = %v\n", coalesced)

	// Coalescing slices
	coalesced, _ = goalesce.Coalesce([]int{1, 2}, []int{2, 3})
	fmt.Printf("Coalesce([1 2], [2 3]) = %v\n", coalesced)
	sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultSetUnion())
	coalesced, _ = goalesce.Coalesce([]int{1, 2}, []int{2, 3}, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce([1 2], [2 3], SetUnion) = %v\n", coalesced)
	// output:
	// Coalesce(abc, def) = def
	// Coalesce(1, 0) = 1
	// Coalesce(&{1}, &{abc}) = &{1 abc}
	// Coalesce(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]
	// Coalesce([1 2], [2 3]) = [2 3]
	// Coalesce([1 2], [2 3], SetUnion) = [1 2 3]
}

func ExampleSliceMergeKeyFunc() {
	type Foo struct {
		Field1 int
	}
	mergeKeyFunc := func(v reflect.Value) reflect.Value {
		return v.FieldByName("Field1")
	}
	sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithMergeByKey(reflect.TypeOf(Foo{}), mergeKeyFunc))
	coalesced, _ := goalesce.Coalesce([]Foo{{1}, {2}}, []Foo{{2}, {3}}, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce([1 2], [2 3], MergeByKey) = %v\n", coalesced)
	// output:
	// Coalesce([1 2], [2 3], MergeByKey) = [{1} {2} {3}]
}
