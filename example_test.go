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
	"reflect"
	"strings"

	"github.com/adutra/goalesce"
)

type User struct {
	ID   int
	Name string
	Age  int
}

type Actor struct {
	ID   int
	Name string
}

type Movie struct {
	Name        string
	Description string
	Actors      []Actor           `goalesce:"id:ID"`
	Tags        []string          `goalesce:"union"`
	Labels      map[string]string `goalesce:"atomic"`
}

func Example() {

	var v, copied interface{}

	// Copying immutable values: the "copy" operation is a no-op
	v = "abc"
	copied, _ = goalesce.DeepCopy(v)
	fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)

	// Copying structs: fields are deep-copied
	v = User{ID: 1, Name: "Alice"}
	copied, _ = goalesce.DeepCopy(v)
	fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)

	// Copying pointers: the values pointed to are deep-copied, the pointers have different addresses
	v = &User{ID: 1, Name: "Alice"}
	copied, _ = goalesce.DeepCopy(v)
	fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)

	// Copying maps: keys and values are deep-copied, the maps point to different addresses
	v = map[int]string{1: "a", 2: "b"}
	copied, _ = goalesce.DeepCopy(v)
	fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)

	// Copying slices: elements are deep-copied, the slices point to different addresses
	v = []int{1, 2}
	copied, _ = goalesce.DeepCopy(v)
	fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)

	var v1, v2, merged interface{}

	// Merging immutable values: the "merge" operation returns v2 if it is non-zero, otherwise v1
	v1 = "abc"
	v2 = "def"
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	v1 = 1
	v2 = 0 // 0 is the zero-value for ints: v1 will be returned
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	// Merging structs
	v1 = User{ID: 1, Name: "Alice"}
	v2 = User{ID: 1, Age: 20}
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	// Merging pointers
	v1 = &User{ID: 1, Name: "Alice"}
	v2 = &User{ID: 1, Age: 20}
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	// Merging maps
	v1 = map[int]string{1: "a", 2: "b"}
	v2 = map[int]string{2: "c", 3: "d"}
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	// Merging slices with default atomic semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	v1 = []int{1, 2}
	v2 = []int{} // empty slice is NOT a zero-value!
	merged, _ = goalesce.DeepMerge(v1, v2)
	fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)

	// Merging slices with empty slices treated as zero-value slices
	v1 = []int{1, 2}
	v2 = []int{} // empty slice will be considered zero-value
	merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithZeroEmptySliceMerge())
	fmt.Printf("DeepMerge(%+v, %+v, ZeroEmptySlice) = %+v\n", v1, v2, merged)

	// Merging slices with set-union semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceSetUnionMerge())
	fmt.Printf("DeepMerge(%+v, %+v, SetUnion) = %+v\n", v1, v2, merged)

	// Merging slices with list-append semantics
	v1 = []int{1, 2}
	v2 = []int{2, 3}
	merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceListAppendMerge())
	fmt.Printf("DeepMerge(%+v, %+v, ListAppend) = %+v\n", v1, v2, merged)

	// Merging slices with merge-by-index semantics
	v1 = []int{1, 2, 3}
	v2 = []int{-1, -2}
	merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceMergeByIndex())
	fmt.Printf("DeepMerge(%+v, %+v, MergeByIndex) = %+v\n", v1, v2, merged)

	// Merging slices with merge-by-id semantics, merge key = field User.ID
	v1 = []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	v2 = []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
	merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithSliceMergeByID(reflect.TypeOf([]User{}), "ID"))
	fmt.Printf("DeepMerge(%+v, %+v, MergeByID) = %+v\n", v1, v2, merged)

	// Merging structs with custom field strategies
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
	merged, _ = goalesce.DeepMerge(v1, v2)
	jsn, _ := json.MarshalIndent(merged, "", "  ")
	fmt.Printf("Merged movie:\n%+v\n", string(jsn))
	// output:
	// DeepCopy(abc) = abc
	// DeepCopy({ID:1 Name:Alice Age:0}) = {ID:1 Name:Alice Age:0}
	// DeepCopy(&{ID:1 Name:Alice Age:0}) = &{ID:1 Name:Alice Age:0}
	// DeepCopy(map[1:a 2:b]) = map[1:a 2:b]
	// DeepCopy([1 2]) = [1 2]
	// DeepMerge(abc, def) = def
	// DeepMerge(1, 0) = 1
	// DeepMerge({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}) = {ID:1 Name:Alice Age:20}
	// DeepMerge(&{ID:1 Name:Alice Age:0}, &{ID:1 Name: Age:20}) = &{ID:1 Name:Alice Age:20}
	// DeepMerge(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]
	// DeepMerge([1 2], [2 3]) = [2 3]
	// DeepMerge([1 2], []) = []
	// DeepMerge([1 2], [], ZeroEmptySlice) = [1 2]
	// DeepMerge([1 2], [2 3], SetUnion) = [1 2 3]
	// DeepMerge([1 2], [2 3], ListAppend) = [1 2 2 3]
	// DeepMerge([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]
	// DeepMerge([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByID) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// Merged movie:
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

func ExampleWithDefaultSliceSetUnionMerge() {
	{
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceSetUnionMerge())
		fmt.Printf("DeepMerge(%+v, %+v, SetUnion) = %+v\n", v1, v2, merged)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{new(int), intPtr(0)} // new(int) and intPtr(0) are equal and point both to 0
		v2 := []*int{nil, intPtr(1)}      // nil will be merged as the zero-value (0)
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceSetUnionMerge())
		fmt.Printf("DeepMerge(%+v, %+v, SetUnion) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(merged))
	}
	// output:
	// DeepMerge([1 2], [2 3], SetUnion) = [1 2 3]
	// DeepMerge([&0 &0], [*int(nil) &1], SetUnion) = [&0 &1]
}

func ExampleWithDefaultSliceListAppendMerge() {
	{
		v1 := []int{1, 2}
		v2 := []int{2, 3}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceListAppendMerge())
		fmt.Printf("DeepMerge(%+v, %+v, ListAppend) = %+v\n", v1, v2, merged)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{new(int), intPtr(0)}
		v2 := []*int{(*int)(nil), intPtr(1)}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceListAppendMerge())
		fmt.Printf("DeepMerge(%+v, %+v, ListAppend) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(merged))
	}
	// output:
	// DeepMerge([1 2], [2 3], ListAppend) = [1 2 2 3]
	// DeepMerge([&0 &0], [*int(nil) &1], ListAppend) = [&0 &0 *int(nil) &1]
}

func ExampleWithDefaultSliceMergeByIndex() {
	{
		v1 := []int{1, 2, 3}
		v2 := []int{-1, -2}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceMergeByIndex())
		fmt.Printf("DeepMerge(%+v, %+v, MergeByIndex) = %+v\n", v1, v2, merged)
	}
	{
		// slice of pointers
		intPtr := func(i int) *int { return &i }
		v1 := []*int{intPtr(1), intPtr(2), intPtr(3)}
		v2 := []*int{nil, intPtr(-2)}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceMergeByIndex())
		fmt.Printf("DeepMerge(%+v, %+v, MergeByIndex) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(merged))
	}
	// output:
	// DeepMerge([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]
	// DeepMerge([&1 &2 &3], [*int(nil) &-2], MergeByIndex) = [&1 &-2 &3]
}

func ExampleWithSliceMergeByID() {
	{
		v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithSliceMergeByID(reflect.TypeOf([]User{}), "ID"))
		fmt.Printf("DeepMerge(%+v, %+v, MergeByID) = %+v\n", v1, v2, merged)
	}
	{
		// slice of pointers
		v1 := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []*User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithSliceMergeByID(reflect.TypeOf([]*User{}), "ID"))
		fmt.Printf("DeepMerge(%+v, %+v, MergeByID) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(merged))
	}
	// output:
	// DeepMerge([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByID) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// DeepMerge([&{ID:1 Name:Alice Age:0} &{ID:2 Name:Bob Age:0}], [&{ID:2 Name: Age:30} &{ID:1 Name: Age:20}], MergeByID) = [&{ID:1 Name:Alice Age:20} &{ID:2 Name:Bob Age:30}]
}

func ExampleWithSliceMergeByKeyFunc() {
	{
		v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		mergeKeyFunc := func(_ int, v reflect.Value) (reflect.Value, error) {
			return v.FieldByName("ID"), nil
		}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithSliceMergeByKeyFunc(reflect.TypeOf([]User{}), mergeKeyFunc))
		fmt.Printf("DeepMerge(%+v, %+v, MergeByKeyFunc) = %+v\n", v1, v2, merged)
	}
	{
		v1 := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
		v2 := []*User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
		mergeKeyFunc := func(_ int, v reflect.Value) (reflect.Value, error) {
			return v.Elem().FieldByName("ID"), nil
		}
		merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithSliceMergeByKeyFunc(reflect.TypeOf([]*User{}), mergeKeyFunc))
		fmt.Printf("DeepMerge(%+v, %+v, MergeByKeyFunc) = %+v\n", printPtrSlice(v1), printPtrSlice(v2), printPtrSlice(merged))
	}
	// output:
	// DeepMerge([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}], MergeByKeyFunc) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]
	// DeepMerge([&{ID:1 Name:Alice Age:0} &{ID:2 Name:Bob Age:0}], [&{ID:2 Name: Age:30} &{ID:1 Name: Age:20}], MergeByKeyFunc) = [&{ID:1 Name:Alice Age:20} &{ID:2 Name:Bob Age:30}]
}

func ExampleWithTypeCopier() {
	userCopier := func(v reflect.Value) (reflect.Value, error) {
		if v.FieldByName("ID").Int() == 1 {
			return reflect.Value{}, errors.New("user 1 has been deleted")
		}
		return reflect.Value{}, nil // delegate to parent coalescer
	}
	{
		v := User{ID: 1, Name: "Alice"}
		copied, err := goalesce.DeepCopy(v, goalesce.WithTypeCopier(reflect.TypeOf(User{}), userCopier))
		fmt.Printf("DeepCopy(%+v, WithTypeCopier) = %+v, %v\n", v, copied, err)
	}
	{
		v := User{ID: 2, Name: "Bob"}
		copied, err := goalesce.DeepCopy(v, goalesce.WithTypeCopier(reflect.TypeOf(User{}), userCopier))
		fmt.Printf("DeepCopy(%+v, WithTypeCopier) = %+v, %v\n", v, copied, err)
	}
	// output:
	// DeepCopy({ID:1 Name:Alice Age:0}, WithTypeCopier) = <nil>, user 1 has been deleted
	// DeepCopy({ID:2 Name:Bob Age:0}, WithTypeCopier) = {ID:2 Name:Bob Age:0}, <nil>
}

func ExampleWithTypeMerger() {
	userMerger := func(v1, v2 reflect.Value) (reflect.Value, error) {
		if v1.FieldByName("ID").Int() == 1 {
			return reflect.Value{}, errors.New("user 1 has been deleted")
		}
		return reflect.Value{}, nil // delegate to parent coalescer
	}
	{
		v1 := User{ID: 1, Name: "Alice"}
		v2 := User{ID: 1, Age: 20}
		merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithTypeMerger(reflect.TypeOf(User{}), userMerger))
		fmt.Printf("DeepMerge(%+v, %+v, WithTypeMerger) = %+v, %v\n", v1, v2, merged, err)
	}
	{
		v1 := User{ID: 2, Name: "Bob"}
		v2 := User{ID: 2, Age: 30}
		merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithTypeMerger(reflect.TypeOf(User{}), userMerger))
		fmt.Printf("DeepMerge(%+v, %+v, WithTypeMerger) = %+v, %v\n", v1, v2, merged, err)
	}
	// output:
	// DeepMerge({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}, WithTypeMerger) = <nil>, user 1 has been deleted
	// DeepMerge({ID:2 Name:Bob Age:0}, {ID:2 Name: Age:30}, WithTypeMerger) = {ID:2 Name:Bob Age:30}, <nil>
}

func ExampleWithFieldMergerProvider() {
	userMergerProvider := func(parent goalesce.DeepMergeFunc) goalesce.DeepMergeFunc {
		return func(v1, v2 reflect.Value) (reflect.Value, error) {
			if v1.Int() == 1 {
				return reflect.Value{}, errors.New("user 1 has been deleted")
			}
			return parent(v1, v2) // call parent coalescer explicitly
		}
	}
	{
		v1 := User{ID: 1, Name: "Alice"}
		v2 := User{ID: 1, Age: 20}
		merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithFieldMergerProvider(reflect.TypeOf(User{}), "ID", userMergerProvider))
		fmt.Printf("DeepMerge(%+v, %+v, WithFieldMergerProvider) = %+v, %v\n", v1, v2, merged, err)
	}
	{
		v1 := User{ID: 2, Name: "Bob"}
		v2 := User{ID: 2, Age: 30}
		merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithFieldMergerProvider(reflect.TypeOf(User{}), "ID", userMergerProvider))
		fmt.Printf("DeepMerge(%+v, %+v, WithFieldMergerProvider) = %+v, %v\n", v1, v2, merged, err)
	}
	// output:
	// DeepMerge({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}, WithFieldMergerProvider) = <nil>, user 1 has been deleted
	// DeepMerge({ID:2 Name:Bob Age:0}, {ID:2 Name: Age:30}, WithFieldMergerProvider) = {ID:2 Name:Bob Age:30}, <nil>
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
