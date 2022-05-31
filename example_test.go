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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adutra/goalesce"
	"reflect"
	"strings"
)

type User struct {
	ID   int
	Name string
	Age  int
}

func Example() {
	var v1, v2, coalesced interface{}

	// Coalescing scalars
	v1 = "abc"
	v2 = "def"
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	v1 = 1
	v2 = 0 // zero-value for ints
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing structs
	v1 = User{ID: 1, Name: "Alice"}
	v2 = User{ID: 1, Age: 20}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing pointers
	v1 = &User{ID: 1, Name: "Alice"}
	v2 = &User{ID: 1, Age: 20}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing maps
	v1 = map[int]string{1: "a", 2: "b"}
	v2 = map[int]string{2: "c", 3: "d"}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with default atomic semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	v1 = []int{1, 2}
	v2 = []int{} // empty slice is NOT a zero-value!
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with empty slices treated as zero-value slices
	v1 = []int{1, 2}
	v2 = []int{} // empty slice will be considered zero-value
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithZeroEmptySlice())
	fmt.Printf("Coalesce(%+v, %+v, ZeroEmptySlice) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with set-union semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithDefaultSetUnion())
	fmt.Printf("Coalesce(%+v, %+v, SetUnion) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with list-append semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithDefaultListAppend())
	fmt.Printf("Coalesce(%+v, %+v, ListAppend) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with merge-by-index semantics
	v1 = []int{1, 2, 3}
	v2 = []int{-1, -2}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithDefaultMergeByIndex())
	fmt.Printf("Coalesce(%+v, %+v, MergeByIndex) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with merge-by-key semantics, merge key = field User.ID
	v1 = []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	v2 = []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithMergeByID(reflect.TypeOf([]User{}), "ID"))
	fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", v1, v2, coalesced)

	// Coalescing structs with custom field strategies
	type Actor struct {
		ID   int
		Name string
	}
	type Movie struct {
		Name        string
		Description string
		Actors      []Actor           `goalesce:"merge,ID"`
		Tags        []string          `goalesce:"union"`
		Labels      map[string]string `goalesce:"atomic"`
	}
	v1 = Movie{
		Name:        "The Matrix",
		Description: "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
		Actors: []Actor{
			{ID: 1, Name: "Keanu Reeves"},
			{ID: 2, Name: "Laurence Fishburne"},
			{ID: 3, Name: "Carrie-Anne Moss"},
		},
		Tags: []string{"sci-fi", "action"},
		Labels: map[string]string{
			"producer": "Wachowski Brothers",
		},
	}
	v2 = Movie{
		Name: "The Matrix",
		Actors: []Actor{
			{ID: 2, Name: "Laurence Fishburne"},
			{ID: 3, Name: "Carrie-Anne Moss"},
			{ID: 4, Name: "Hugo Weaving"},
		},
		Tags: []string{"action", "fantasy"},
		Labels: map[string]string{
			"director": "Wachowski Brothers",
		},
	}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	jsn, _ := json.MarshalIndent(coalesced, "", "  ")
	fmt.Printf("Coalesced movie:\n%+v\n", string(jsn))
	// output:
	// Coalesce(abc, def) = def
	// Coalesce(1, 0) = 1
	// Coalesce({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}) = {ID:1 Name:Alice Age:20}
	// Coalesce(&{ID:1 Name:Alice Age:0}, &{ID:1 Name: Age:20}) = &{ID:1 Name:Alice Age:20}
	// Coalesce(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]
	// Coalesce([1 2], [2 3]) = [2 3]
	// Coalesce([1 2], []) = []
	// Coalesce([1 2], [], ZeroEmptySlice) = [1 2]
	// Coalesce([1 2], [2 3], SetUnion) = [1 2 3]
	// Coalesce([1 2], [2 3], ListAppend) = [1 2 2 3]
	// Coalesce([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]
	// Coalesce([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByField) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// Coalesced movie:
	// {
	//   "Name": "The Matrix",
	//   "Description": "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
	//   "Actors": [
	//     {
	//       "ID": 1,
	//       "Name": "Keanu Reeves"
	//     },
	//     {
	//       "ID": 2,
	//       "Name": "Laurence Fishburne"
	//     },
	//     {
	//       "ID": 3,
	//       "Name": "Carrie-Anne Moss"
	//     },
	//     {
	//       "ID": 4,
	//       "Name": "Hugo Weaving"
	//     }
	//   ],
	//   "Tags": [
	//     "sci-fi",
	//     "action",
	//     "fantasy"
	//   ],
	//   "Labels": {
	//     "director": "Wachowski Brothers"
	//   }
	// }
}

func ExampleWithDefaultSetUnion() {
	{
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultSetUnion())
		fmt.Printf("Coalesce(%+v, %+v, SetUnion) = %+v\n", v1, v2, coalesced)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{new(int), intPtr(0)} // new(int) and intPtr(0) are equal and point both to 0
		v2 := []*int{nil, intPtr(1)}      // nil will be coalesced as the zero-value (0)
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultSetUnion())
		fmt.Printf("Coalesce(%+v, %+v, SetUnion) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(coalesced))
	}
	// output:
	// Coalesce([1 2], [2 3], SetUnion) = [1 2 3]
	// Coalesce([&0 &0], [*int(nil) &1], SetUnion) = [&0 &1]
}

func ExampleWithDefaultListAppend() {
	{
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultListAppend())
		fmt.Printf("Coalesce(%+v, %+v, ListAppend) = %+v\n", v1, v2, coalesced)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{new(int), intPtr(0)}
		v2 := []*int{(*int)(nil), intPtr(1)}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultListAppend())
		fmt.Printf("Coalesce(%+v, %+v, ListAppend) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(coalesced))
	}
	// output:
	// Coalesce([1 2], [2 3], ListAppend) = [1 2 2 3]
	// Coalesce([&0 &0], [*int(nil) &1], ListAppend) = [&0 &0 *int(nil) &1]
}

func ExampleWithDefaultMergeByIndex() {
	{
		v1 := []int{1, 2, 3}
		v2 := []int{-1, -2}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultMergeByIndex())
		fmt.Printf("Coalesce(%+v, %+v, MergeByIndex) = %+v\n", v1, v2, coalesced)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{intPtr(1), intPtr(2), intPtr(3)}
		v2 := []*int{nil, intPtr(-2)}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultMergeByIndex())
		fmt.Printf("Coalesce(%+v, %+v, MergeByIndex) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(coalesced))
	}
	// output:
	// Coalesce([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]
	// Coalesce([&1 &2 &3], [*int(nil) &-2], MergeByIndex) = [&1 &-2 &3]
}

func ExampleWithMergeByID() {
	{
		v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByID(reflect.TypeOf([]User{}), "ID"))
		fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", v1, v2, coalesced)
	}
	{
		// slice of pointers
		v1 := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []*User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByID(reflect.TypeOf([]*User{}), "ID"))
		fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(coalesced))
	}
	// output:
	// Coalesce([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByField) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// Coalesce([&{ID:1 Name:Alice Age:0} &{ID:2 Name:Bob Age:0}], [&{ID:2 Name: Age:30} &{ID:1 Name: Age:20}], MergeByField) = [&{ID:1 Name:Alice Age:20} &{ID:2 Name:Bob Age:30}]
}

func ExampleWithMergeByKeyFunc() {
	{
		v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		mergeKeyFunc := func(_ int, v reflect.Value) reflect.Value {
			return v.FieldByName("ID")
		}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByKeyFunc(reflect.TypeOf([]User{}), mergeKeyFunc))
		fmt.Printf("Coalesce(%+v, %+v, MergeByKey) = %+v\n", v1, v2, coalesced)
	}
	{
		v1 := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []*User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		mergeKeyFunc := func(_ int, v reflect.Value) reflect.Value {
			return v.Elem().FieldByName("ID")
		}
		coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByKeyFunc(reflect.TypeOf([]*User{}), mergeKeyFunc))
		fmt.Printf("Coalesce(%+v, %+v, MergeByKey) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(coalesced))
	}
	// output:
	// Coalesce([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByKey) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// Coalesce([&{ID:1 Name:Alice Age:0} &{ID:2 Name:Bob Age:0}], [&{ID:2 Name: Age:30} &{ID:1 Name: Age:20}], MergeByKey) = [&{ID:1 Name:Alice Age:20} &{ID:2 Name:Bob Age:30}]
}

func ExampleWithTypeCoalescer() {
	userCoalescer := func(v1, v2 reflect.Value) (reflect.Value, error) {
		if v1.FieldByName("ID").Int() == 1 {
			return reflect.Value{}, errors.New("user 1 has been deleted")
		}
		return reflect.Value{}, nil // delegate to parent coalescer
	}
	{
		v1 := User{ID: 1, Name: "Alice"}
		v2 := User{ID: 1, Age: 20}
		mainCoalescer := goalesce.NewCoalescer(goalesce.WithTypeCoalescer(reflect.TypeOf(User{}), userCoalescer))
		coalesced, err := mainCoalescer(reflect.ValueOf(v1), reflect.ValueOf(v2))
		fmt.Printf("Coalesce(%+v, %+v, WithTypeCoalescer) = %+v, %v\n", v1, v2, coalesced, err)
	}
	{
		v1 := User{ID: 2, Name: "Bob"}
		v2 := User{ID: 2, Age: 30}
		mainCoalescer := goalesce.NewCoalescer(goalesce.WithTypeCoalescer(reflect.TypeOf(User{}), userCoalescer))
		coalesced, err := mainCoalescer(reflect.ValueOf(v1), reflect.ValueOf(v2))
		fmt.Printf("Coalesce(%+v, %+v, WithTypeCoalescer) = %+v, %v\n", v1, v2, coalesced, err)
	}
	// output:
	// Coalesce({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}, WithTypeCoalescer) = <invalid reflect.Value>, user 1 has been deleted
	// Coalesce({ID:2 Name:Bob Age:0}, {ID:2 Name: Age:30}, WithTypeCoalescer) = {ID:2 Name:Bob Age:30}, <nil>
}

func ExampleWithFieldCoalescerProvider() {
	userCoalescerProvider := func(parent goalesce.Coalescer) goalesce.Coalescer {
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			if v1.Int() == 1 {
				return reflect.Value{}, errors.New("user 1 has been deleted")
			}
			return parent(v1, v2) // use parent coalescer
		}
	}
	{
		v1 := User{ID: 1, Name: "Alice"}
		v2 := User{ID: 1, Age: 20}
		mainCoalescer := goalesce.NewCoalescer(goalesce.WithFieldCoalescerProvider(reflect.TypeOf(User{}), "ID", userCoalescerProvider))
		coalesced, err := mainCoalescer(reflect.ValueOf(v1), reflect.ValueOf(v2))
		fmt.Printf("Coalesce(%+v, %+v, WithFieldCoalescerProvider) = %+v, %v\n", v1, v2, coalesced, err)
	}
	{
		v1 := User{ID: 2, Name: "Bob"}
		v2 := User{ID: 2, Age: 30}
		mainCoalescer := goalesce.NewCoalescer(goalesce.WithFieldCoalescerProvider(reflect.TypeOf(User{}), "ID", userCoalescerProvider))
		coalesced, err := mainCoalescer(reflect.ValueOf(v1), reflect.ValueOf(v2))
		fmt.Printf("Coalesce(%+v, %+v, WithFieldCoalescerProvider) = %+v, %v\n", v1, v2, coalesced, err)
	}
	// output:
	// Coalesce({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}, WithFieldCoalescerProvider) = <invalid reflect.Value>, user 1 has been deleted
	// Coalesce({ID:2 Name:Bob Age:0}, {ID:2 Name: Age:30}, WithFieldCoalescerProvider) = {ID:2 Name:Bob Age:30}, <nil>
}

func printPtrSlice(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.IsNil() {
		return fmt.Sprintf("%T(nil)", i)
	}
	s := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		s[i] = printPtr(v.Index(i).Interface())
	}
	return fmt.Sprintf("[%v]", strings.Join(s, " "))
}

func printPtr(i interface{}) string {
	v := reflect.ValueOf(i)
	if v.IsNil() {
		return fmt.Sprintf("%T(nil)", i)
	}
	return fmt.Sprintf("&%+v", v.Elem().Interface())
}
