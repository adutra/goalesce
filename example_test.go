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
	"fmt"
	"github.com/adutra/goalesce"
	"reflect"
)

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
	type User struct {
		Id   int
		Name string
		Age  int
	}
	v1 = User{Id: 1, Name: "Alice"}
	v2 = User{Id: 1, Age: 20}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing pointers
	v1 = &User{Id: 1, Name: "Alice"}
	v2 = &User{Id: 1, Age: 20}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing maps
	v1 = map[int]string{1: "a", 2: "b"}
	v2 = map[int]string{2: "c", 3: "d"}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with default replace semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	v1 = []int{1, 2}
	v2 = []int{} // empty slice is NOT a zero-value!
	coalesced, _ = goalesce.Coalesce(v1, v2)
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with set-union semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultSetUnion())
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v, SetUnion) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with list-append semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	sliceCoalescer = goalesce.NewSliceCoalescer(goalesce.WithDefaultListAppend())
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v, ListAppend) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with merge-by-index semantics
	v1 = []int{1, 2, 3}
	v2 = []int{-1, -2}
	sliceCoalescer = goalesce.NewSliceCoalescer(goalesce.WithDefaultMergeByIndex())
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v, MergeByIndex) = %+v\n", v1, v2, coalesced)

	// Coalescing slices with merge-by-key semantics, merge key = field User.Id
	v1 = []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
	v2 = []User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
	sliceCoalescer = goalesce.NewSliceCoalescer(goalesce.WithMergeByField(reflect.TypeOf(User{}), "Id"))
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", v1, v2, coalesced)

	// Coalescing structs with custom field strategies
	type Actor struct {
		Id   int
		Name string
	}
	type Movie struct {
		Name        string
		Description string
		Actors      []Actor           `coalesceStrategy:"merge" coalesceMergeKey:"Id"`
		Tags        []string          `coalesceStrategy:"union"`
		Labels      map[string]string `coalesceStrategy:"replace"`
	}
	v1 = Movie{
		Name:        "The Matrix",
		Description: "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
		Actors: []Actor{
			{Id: 1, Name: "Keanu Reeves"},
			{Id: 2, Name: "Laurence Fishburne"},
			{Id: 3, Name: "Carrie-Anne Moss"},
		},
		Tags: []string{"sci-fi", "action"},
		Labels: map[string]string{
			"producer": "Wachowski Brothers",
		},
	}
	v2 = Movie{
		Name: "The Matrix",
		Actors: []Actor{
			{Id: 2, Name: "Laurence Fishburne"},
			{Id: 3, Name: "Carrie-Anne Moss"},
			{Id: 4, Name: "Hugo Weaving"},
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
	// Coalesce({Id:1 Name:Alice Age:0}, {Id:1 Name: Age:20}) = {Id:1 Name:Alice Age:20}
	// Coalesce(&{Id:1 Name:Alice Age:0}, &{Id:1 Name: Age:20}) = &{Id:1 Name:Alice Age:20}
	// Coalesce(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]
	// Coalesce([1 2], [2 3]) = [2 3]
	// Coalesce([1 2], []) = []
	// Coalesce([1 2], [2 3], SetUnion) = [1 2 3]
	// Coalesce([1 2], [2 3], ListAppend) = [1 2 2 3]
	// Coalesce([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]
	// Coalesce([{Id:1 Name:Alice Age:0} {Id:2 Name:Bob Age:0}], [{Id:2 Name: Age:30} {Id:1 Name: Age:20}], MergeByField) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]
	// Coalesced movie:
	// {
	//   "Name": "The Matrix",
	//   "Description": "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
	//   "Actors": [
	//     {
	//       "Id": 1,
	//       "Name": "Keanu Reeves"
	//     },
	//     {
	//       "Id": 2,
	//       "Name": "Laurence Fishburne"
	//     },
	//     {
	//       "Id": 3,
	//       "Name": "Carrie-Anne Moss"
	//     },
	//     {
	//       "Id": 4,
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

func ExampleWithMergeByField() {
	type User struct {
		Id   int
		Name string
		Age  int
	}
	var v1, v2, coalesced interface{}
	sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithMergeByField(reflect.TypeOf(User{}), "Id"))

	v1 = []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
	v2 = []User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)

	// also works on slices of *User:
	v1 = []*User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
	v2 = []*User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
	coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	jsn, _ := json.MarshalIndent(coalesced, "", "  ")
	fmt.Printf("Coalesced users:\n%+v\n", string(jsn))
	// output:
	// Coalesce([{Id:1 Name:Alice Age:0} {Id:2 Name:Bob Age:0}], [{Id:2 Name: Age:30} {Id:1 Name: Age:20}]) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]
	// Coalesced users:
	// [
	//   {
	//     "Id": 1,
	//     "Name": "Alice",
	//     "Age": 20
	//   },
	//   {
	//     "Id": 2,
	//     "Name": "Bob",
	//     "Age": 30
	//   }
	// ]
}

func ExampleSliceMergeKeyFunc() {
	type User struct {
		Id   int
		Name string
		Age  int
	}
	mergeKeyFunc := func(_ int, v reflect.Value) reflect.Value {
		return v.FieldByName("Id")
	}
	v1 := []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
	v2 := []User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
	sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithMergeByKey(reflect.TypeOf(User{}), mergeKeyFunc))
	coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
	fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)
	// output:
	// Coalesce([{Id:1 Name:Alice Age:0} {Id:2 Name:Bob Age:0}], [{Id:2 Name: Age:30} {Id:1 Name: Age:20}]) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]
}
