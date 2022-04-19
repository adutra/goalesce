[![GoVersion][GoVersionImg]][GoVersionLink]
[![GoDoc][GoDocImg]][GoDocLink]
[![GoReport][GoReportImg]][GoReportLink]
[![codecov][CodeCovImg]][CodeCovLink]

# Package goalesce

Package goalesce is a library for coalescing (a.k.a. merging) objects in Go. It can coalesce any type of object,
including structs, maps, and slices, even nested ones.

## Introduction

The main entry point is the `Coalesce` function:

    func Coalesce(o1, o2 interface{}, opts ...MainCoalescerOption) (coalesced interface{}, err error)

It coalesces the 2 values into a single value and returns that value. 

When called with no options, the function uses the following coalescing algorithm:

- If both values are untyped nils, return nil.
- If one value is untyped nil, return the other value.
- If both values are [zero-values] for the type, return the type's zero-value.
- If one value is a zero-value for the type, return the other value.
- Otherwise, the values are coalesced using the following rules:
  - If both values are pointers, coalesce the values pointed to.
  - If both values are maps, coalesce the maps recursively, key by key.
  - If both values are structs, coalesce the structs recursively, field by field.
  - For other types (including slices), return the second value ("replace" semantics).

Note that by default, slices are coalesced with replace semantics, that is, the second slice overwrites the first one
completely. It is possible to change this behavior, see examples below.

The `Coalesce` function can be called with a list of options to modify its default coalescing behavior. See the
documentation of each option for details.

## Examples 

### Coalescing scalars

Scalars are always coalesced with replace semantics when both values are non-zero-values:

```go
v1 := "abc"
v2 := "def"
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce(abc, def) = def

### Coalescing pointers

Pointers are coalesced by coalescing the values they point to (which could be nil):

```go
stringPtr := func(s string) *string { return &s }
v1 := stringPtr("abc")
v2 := stringPtr("def")
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", *v1, *v2, *(coalesced.(*string)))
```

Output:

    Coalesce(abc, def) = def

### Coalescing maps

When both maps are non-zero-values, the default behavior is to coalesce the two maps key by key, recursively coalescing
the values.

```go
v1 := map[int]string{1: "a", 2: "b"}
v2 := map[int]string{2: "c", 3: "d"}
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]

### Coalescing slices

When both slices are non-zero-values, the default behavior is to _replace_ the first slice with the second one: 

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce([1 2], [2 3]) = [2 3]

This is indeed the safest choice when coalescing slices, but other coalescing strategies can be used (see below).

Note: an empty slice is _not_ a zero-value for a slice. Therefore, when the second slice is an empty slice, an empty 
slice is returned:

```go
v1 := []int{1, 2}
v2 := []int{} // empty slice
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce([1 2], []) = []

#### Using "set-union" strategy

The "set-union" strategy can be used to coalesce the two slices together by creating a resulting slice that contains all
elements from both slices, but no duplicates:

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultSetUnion())
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
fmt.Printf("Coalesce(%v, %v, SetUnion) = %v\n", coalesced)
```

Output:

    Coalesce([1 2], [2 3], SetUnion) = [1 2 3]

When the slice elements are pointers, this strategy dereferences the pointers and compare their targets. If the 
resulting value is nil, the zero-value is used instead. _This means that two nil pointers are considered equal, and 
equal to a non-nil pointer to the zero-value_:

```go
intPtr := func(i int) *int { return &i }
v1 := []*int{new(int), intPtr(0)} // new(int) and intPtr(0) are equal and point both to the zero-value (0)
v2 := []*int{nil, intPtr(1)}      // nil will be coalesced as the zero-value (0)
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultSetUnion())
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
for i, elem := range coalesced.([]*int) {
    fmt.Printf("%v: %T %v\n", i, elem, *elem)
}
```

Output:

    0: *int 0
    1: *int 1

This strategy is fine for slices of scalars and pointers thereof, but it is not recommended for slices of complex 
types as the elements may not be fully comparable.

The resulting slice's element order is deterministic: each element appears in the order it was first encountered when 
iterating over the two slices.

#### Using "list-append" strategy

The "list-append" strategy appends the second slice to the first one (possibly resulting in duplicates):

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultListAppend())
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
fmt.Printf("Coalesce(%v, %v, ListAppend) = %v\n", coalesced)
```

Output

    Coalesce([1 2], [2 3], ListAppend) = [1 2 2 3]

The resulting slice's element order is deterministic.

#### Using "merge-by-index" strategy

The "merge-by-index" strategy can be used to coalesce two slices together using their indices as the merge key:

```go
v1 := []int{1, 2, 3}
v2 := []int{-1, -2}
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithDefaultMergeByIndex())
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
fmt.Printf("Coalesce(%v, %v, MergeByIndex) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]

#### Using "merge-by-key" strategy

The "merge-by-key" strategy can be used to coalesce two slices together using an arbitrary merge key:

```go
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
```

Output:

    Coalesce([{Id:1 Name:Alice Age:0} {Id:2 Name:Bob Age:0}], [{Id:2 Name: Age:30} {Id:1 Name: Age:20}]) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]

This strategy is similar to Kubernetes' [strategic merge patch].

The function `mergeKeyFunc` must be of type `SliceMergeKeyFunc`. It will be invoked with the index and value of the
slice element to extract a merge key from.

The most common usage for this strategy is to coalesce slices of structs, where the merge key is the name of a primary
key field. In this case, we can use the `WithMergeByField` option to specify the field name to use as merge key, and 
simplify the example above as follows:

```go
v1 := []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []User{{Id: 1, Age: 20}      , {Id: 2, Age: 30}}
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithMergeByField(reflect.TypeOf(User{}), "Id"))
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce([{Id:1 Name:Alice} {Id:2 Name:Bob}], [{Id:1 Age:20} {Id:2 Age:30}], MergeByField) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]

The option `WithMergeByField` also works out-of-the-box on slices of pointers to structs:

```go
v1 := []*User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []*User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
jsn, _ := json.MarshalIndent(coalesced, "", "  ")
fmt.Println(string(jsn))
```

Output:

```json
[
  {
    "Id": 1,
    "Name": "Alice",
    "Age": 20
  },
  {
    "Id": 2,
    "Name": "Bob",
    "Age": 30
  }
]
```    

### Coalescing structs

When both structs are non-zero-values, the default behavior is to coalesce the two structs field by field, recursively
coalescing their values.

```go
type User struct {
    Id   int
    Name string
    Age  int
}
v1 := User{Id: 1, Name: "Alice"}
v2 := User{Id: 1, Age: 20}
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce({Id:1 Name:Alice}, {Id:1 Age:20}) = {Id:1 Name:Alice Age:20}

#### Per-field coalescing strategies

When the default struct coalescing behavior is not desired or sufficient, per-field coalescing strategies can be used. 

The struct tag `coalesceStrategy` allows to specify the following per-field strategies:

| Strategy  | Valid on               | Effect                              |
|-----------|------------------------|-------------------------------------|
| `replace` | Any field              | Applies "replace" semantics.        |
| `union`   | Slice fields           | Applies "set-union" semantics.      |
| `append`  | Slice fields           | Applies "list-append" semantics.    |   
| `index`   | Slice fields           | Applies "merge-by-index" semantics. |   
| `merge`   | Slice of struct fields | Applies "merge-by-key" semantics.   |   

With `merge`, a merge key must also be provided in another struct tag: `coalesceMergeKey`. The merge key _must_ be the
name of a field in the slice's struct element type.

Example:

```go
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
fmt.Println(string(jsn))
```

Output:

```json
{
  "Name": "The Matrix",
  "Description": "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
  "Actors": [
    {
      "Id": 1,
      "Name": "Keanu Reeves"
    },
    {
      "Id": 2,
      "Name": "Laurence Fishburne"
    },
    {
      "Id": 3,
      "Name": "Carrie-Anne Moss"
    },
    {
      "Id": 4,
      "Name": "Hugo Weaving"
    }
  ],
  "Tags": [
    "sci-fi",
    "action",
    "fantasy"
  ],
  "Labels": {
    "director": "Wachowski Brothers"
  }
}
```

See the [online documentation](https://pkg.go.dev/github.com/adutra/goalesce?tab=doc) for more examples.

## Advanced usage

The `Coalescer` interface allows for custom coalescing algorithms to be implemented. By passing custom coalescers to
the `Coalesce` function, its behavior can be modified in any way.

[GoDocImg]: https://img.shields.io/badge/docs-golang-blue.svg
[GoDocLink]: https://godoc.org/github.com/adutra/goalesce
[GoVersionImg]: https://img.shields.io/github/go-mod/go-version/adutra/goalesce.svg
[GoVersionLink]: https://github.com/adutra/goalesce
[GoReportImg]: https://goreportcard.com/badge/github.com/adutra/goalesce
[GoReportLink]: https://goreportcard.com/report/github.com/adutra/goalesce
[CodeCovImg]: https://codecov.io/gh/adutra/goalesce/branch/main/graph/badge.svg?token=REA57AEQZ6
[CodeCovLink]: https://codecov.io/gh/adutra/goalesce
[zero-values]:https://go.dev/ref/spec#The_zero_value
[strategic merge patch]:https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#notes-on-the-strategic-merge-patch
