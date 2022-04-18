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
  - If both values are zero values for the type, return the type's zero value.
  - If one value is a zero value for the type, return the other value.
  - If both values are non-zero values, the values are coalesced using the following rules:
    - If both values are pointers, coalesce the values pointed to.
    - If both values are structs, coalesce the structs recursively, field by field.
    - If both values are maps, coalesce the maps recursively, key by key.
    - Otherwise, return the second value.

The `Coalesce` function can be called with a list of options to modify its default coalescing behavior. See the
documentation of each option for details.

## Advanced usage

The `Coalescer` interface allows for custom coalescing algorithms to be implemented. By passing custom coalescers to
the `Coalesce` function, its behavior can be modified in any way.

## Examples 

### Coalescing maps

When both maps are non-zero, the default behavior is to coalesce the two maps key by key, recursively coalescing the
values.

```go
v1 := map[int]string{1: "a", 2: "b"}
v2 := map[int]string{2: "c", 3: "d"}
coalesced, _ = goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]

### Coalescing slices

When both slices are non-zero, the default behavior is to _replace_ the first slice with the second one: 

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%v, %v) = %v\n", v1, v2, coalesced)
```

Output:

    Coalesce([1 2], [2 3]) = [2 3]

This is indeed the safest choice when coalescing slices, but other coalescing strategies can be used.

Note: an empty slice is _not_ a zero value for a slice. Therefore, when the second slice is an empty slice, an empty 
slice is returned:

```go
v1 := []int{1, 2}
v2 := []int{} // empty slice
coalesced, _ := goalesce.Coalesce(v1, v2)
fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)
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

When the slice elements are pointers, this strategy dereferences the pointers and compare their targets.

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

The resulting slice's element order is deterministic: the two slices are simply concatenated together.

#### Using "merge-by" strategy

The "merge-by" strategy can be used to coalesce two slices together using a merge key. The most common usage for 
this strategy is to coalesce slices of structs, where the merge key is the name of a primary key field:

```go
v1 := []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []User{{Id: 1, Age: 20}      , {Id: 2, Age: 30}}
sliceCoalescer := goalesce.NewSliceCoalescer(goalesce.WithMergeByField(reflect.TypeOf(User{}), "Id"))
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithSliceCoalescer(sliceCoalescer))
fmt.Printf("Coalesce(%+v, %+v, MergeByField) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce([{Id:1 Name:Alice} {Id:2 Name:Bob}], [{Id:1 Age:20} {Id:2 Age:30}], MergeByField) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]

This strategy is similar to Kubernetes' [strategic merge patch].

[strategic merge patch]:https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#notes-on-the-strategic-merge-patch

### Coalescing structs

When both structs are non-zero, the default behavior is to coalesce the two structs field by field, recursively
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

When the default behavior is not desired or sufficient, per-field coalescing strategies can be used. 

The struct tag `coalesceStrategy` allows to specify the following per-field strategies:

| Strategy  | Valid on     | effect                                                        |
|-----------|--------------|---------------------------------------------------------------|
| `replace` | Any field    | Applies "replace" semantics.                                  |
| `union`   | Slice fields | Applies "set-union" semantics.                                |
| `append`  | Slice fields | Applies "list-append" semantics.                              |   
| `merge`   | Slice fields | Applies "merge" semantics. A merge key must also be provided. |   

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
println(string(jsn))
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
