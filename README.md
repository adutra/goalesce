[![GoVersion][GoVersionImg]][GoVersionLink]
[![GoDoc][GoDocImg]][GoDocLink]
[![GoReport][GoReportImg]][GoReportLink]
[![CodeCov][CodeCovImg]][CodeCovLink]

# Package goalesce

Package goalesce is a library for copying and merging objects in Go. It can merge and copy any type
of object, including structs, maps, arrays and slices, even nested ones.

## Introduction

The main entry points are the `DeepCopy` and `DeepMerge` functions:

    func DeepCopy  [T any](     o T, opts ...Option) (T, error)
    func DeepMerge [T any](o1, o2 T, opts ...Option) (T, error)

`DeepCopy`, as the name implies, copies the given object and returns the copy. The copy is "deep" in
the sense that it copies all the fields and elements of the object recursively.

`DeepMerge` merges the two values into a single value and returns that value. Again, the merge is
"deep" and will merge all the fields and elements of the object recursively.

When called with no options, `DeepMerge` uses the following merge algorithm:

- If both values are untyped nils, return nil.
- If one value is untyped nil, return the other value.
- If both values are [zero-values] for the type, return the type's zero-value.
- If one value is a zero-value for the type, return the other value.
- Otherwise, the values are merged using the following rules:
  - If both values are interfaces of same underlying types, merge the underlying values.
  - If both values are pointers, merge the values pointed to.
  - If both values are maps, merge the maps recursively, key by key.
  - If both values are structs, merge the structs recursively, field by field.
  - For other types (including slices), return the second value ("atomic" semantics).

Note that by default, slices and arrays are merged with atomic semantics, that is, the second slice
or array overwrites the first one completely if it is non-zero. It is possible to change this
behavior, see examples below.

Both `DeepCopy` and `DeepMerge` can be called with a list of options to modify its default merging
and copying behavior. See the documentation of each option for details.

## Using DeepCopy

Using `DeepCopy` is extremely simple:

### Copying atomic values

Immutable values are always copied with atomic semantics, that is, the returned "copy" is actually
the value itself. This is OK since only immutable types that are typically passed by value (int,
string, etc.) are copied with this strategy.

```go
v = "abc"
copied, _ = goalesce.DeepCopy(v)
fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)
```

Output:

    DeepCopy(abc) = abc

### Copying structs

The copied struct is a newly-allocated object; the struct fields are deep-copied:

```go
type User struct {
    ID   int
    Name string
}
v = User{ID: 1, Name: "Alice"}
copied, _ = goalesce.DeepCopy(v)
fmt.Printf("DeepCopy(%+v) = %+v\n", v, copied)
```

Output:

    DeepCopy({ID:1 Name:Alice}) = {ID:1 Name:Alice}

### Copying pointers

The copied pointer never points to the same memory address; the pointer target is deep-copied:

```go
v = &User{ID: 1, Name: "Alice"}
copied, _ = goalesce.DeepCopy(v)
fmt.Printf("DeepCopy(%+v) = %+v, %p != %p\n", v, copied, v, copied)
```

Output:

    DeepCopy(&{ID:1 Name:Alice}) = &{ID:1 Name:Alice}, 0xc00000e2a0 != 0xc00000e2d0

### Copying maps

The copied map never points to the same memory address; the map entries are deep-copied:

```go
v = map[int]string{1: "a", 2: "b"}
copied, _ = goalesce.DeepCopy(v)
fmt.Printf("DeepCopy(%+v) = %+v, %p != %p\n", v, copied, v, copied)
```

Output:

    DeepCopy(map[1:a 2:b]) = map[1:a 2:b], 0xc000101470 != 0xc0001015c0

### Copying slices

The copied slice never points to the same memory address; the slice elements are deep-copied:

```go
v = []int{1, 2}
copied, _ = goalesce.DeepCopy(v)
fmt.Printf("DeepCopy(%+v) = %+v, %p != %p\n", v, copied, v, copied)
```

Output:

    DeepCopy([1 2]) = [1 2], 0xc000018b90 != 0xc000018ba0

### Custom copiers

The option `WithTypeCopier` can be used to delegate the copying of a given type to a custom
function:

```go
negatingCopier := func(v reflect.Value) (reflect.Value, error) {
    result := reflect.New(v.Type()).Elem()
    result.SetInt(-v.Int())
    return result, nil
}
v := 1
copied, err := goalesce.DeepCopy(v, goalesce.WithTypeCopier(reflect.TypeOf(v), negatingCopier))
fmt.Printf("DeepCopy(%+v, WithTypeCopier) = %+v, %v\n", v, copied, err)
```

Output:

    DeepCopy(1, WithTypeCopier) = -1, <nil>

## Using DeepMerge 

### Merging atomic values

Immutable values are always merged with atomic semantics: the "merged" value is actually the second
value if it is non-zero, and the first value otherwise. This is OK since values of types like int,
string, etc. are immutable.

```go
v1 := "abc"
v2 := "def"
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%v, %v) = %v\n", v1, v2, merged)
```

Output:

    DeepMerge(abc, def) = def

### Merging pointers

Pointers are merged by merging the values they point to (which could be nil):

```go
stringPtr := func(s string) *string { return &s }
v1 := stringPtr("abc")
v2 := stringPtr("def")
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%v, %v) = %v\n", *v1, *v2, *(merged.(*string)))
```

Output:

    DeepMerge(abc, def) = def

### Merging maps

When both maps are non-zero-values, the default behavior is to merge the two maps key by key,
recursively merging the values.

```go
v1 := map[int]string{1: "a", 2: "b"}
v2 := map[int]string{2: "c", 3: "d"}
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%v, %v) = %v\n", v1, v2, merged)
```

Output:

    DeepMerge(map[1:a 2:b], map[2:c 3:d]) = map[1:a 2:c 3:d]

### Merging slices and arrays

When both slices or arrays are non-zero-values, the default behavior is to apply atomic semantics,
that is, to _replace_ the first slice or array with the second one:

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%v, %v) = %v\n", v1, v2, merged)
```

Output:

    DeepMerge([1 2], [2 3]) = [2 3]

This is indeed the safest choice when merging slices and arrays, but other merging strategies can be
used (see below).

#### Treating empty slices as zero-values

An empty slice is _not_ a zero-value for a slice. Therefore, when the second slice is an empty
slice, an empty slice is returned:

```go
v1 := []int{1, 2}
v2 := []int{} // empty slice
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%v, %v) = %v\n", v1, v2, merged)
```

Output:

    DeepMerge([1 2], []) = []

To consider empty slices as zero-values, use the `WithZeroEmptySlice` option. This changes the
default behavior: when merging a non-empty slice with an empty slice, normally the empty slice is
returned as in the example above; but with this option, the non-empty slice is returned.

```go
v1 = []int{1, 2}
v2 = []int{} // empty slice will be considered zero-value
merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithZeroEmptySliceMerge())
fmt.Printf("DeepMerge(%+v, %+v, ZeroEmptySlice) = %+v\n", v1, v2, merged)
```

Output:

    DeepMerge([1 2], [], ZeroEmptySlice) = [1 2]

#### Using "set-union" strategy

The "set-union" strategy can be used to merge the two slices together by creating a resulting slice
that contains all elements from both slices, but no duplicates:

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceSetUnionMerge())
fmt.Printf("DeepMerge(%v, %v, SetUnion) = %v\n", merged)
```

Output:

    DeepMerge([1 2], [2 3], SetUnion) = [1 2 3]

When the slice elements are pointers, this strategy dereferences the pointers and compare their
targets. If the resulting value is nil, the zero-value is used instead. _This means that two nil
pointers are considered equal, and equal to a non-nil pointer to the zero-value_:

```go
intPtr := func(i int) *int { return &i }
v1 := []*int{new(int), intPtr(0)} // new(int) and intPtr(0) are equal and point both to the zero-value (0)
v2 := []*int{nil, intPtr(1)}      // nil will be merged as the zero-value (0)
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceSetUnionMerge())
for i, elem := range merged.([]*int) {
    fmt.Printf("%v: %T %v\n", i, elem, *elem)
}
```

Output:

    0: *int 0
    1: *int 1

This strategy is fine for slices of simple types and pointers thereof, but it is not recommended for
slices of complex types as the elements may not be fully comparable. Also, it is not suitable for
slices of double pointers.

The resulting slice's element order is deterministic: each element appears in the order it was first
encountered when iterating over the two slices.

This strategy is available with two options:

* `WithDefaultSliceSetUnionMerge`: applies this strategy to all slices;
* `WithSliceSetUnionMerge`: applies this strategy to slices of a given type.

This strategy is not available for arrays.

#### Using "list-append" strategy

The "list-append" strategy appends the second slice to the first one (possibly resulting in
duplicates):

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceListAppendMerge())
fmt.Printf("DeepMerge(%v, %v, ListAppend) = %v\n", merged)
```

Output

    DeepMerge([1 2], [2 3], ListAppend) = [1 2 2 3]

The resulting slice's element order is deterministic.

This strategy is available with two options:

* `WithDefaultSliceListAppendMerge`: applies this strategy to all slices;
* `WithSliceListAppendMerge`: applies this strategy to slices of a given type.

This strategy is not available for arrays.

#### Using "merge-by-index" strategy

The "merge-by-index" strategy can be used to merge two slices together using their indices as the
merge key:

```go
v1 := []int{1, 2, 3}
v2 := []int{-1, -2}
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithDefaultSliceMergeByIndex())
fmt.Printf("DeepMerge(%v, %v, MergeByIndex) = %v\n", v1, v2, merged)
```

Output:

    DeepMerge([1 2 3], [-1 -2], MergeByIndex) = [-1 -2 3]

This strategy is available for slices with two options:

* `WithDefaultSliceMergeByIndex`: applies this strategy to all slices;
* `WithSliceMergeByIndex`: applies this strategy to slices of a given type.

This strategy is available for arrays with two options:

* `WithDefaultArrayMergeByIndex`: applies this strategy to all arrays;
* `WithArrayMergeByIndex`: applies this strategy to arrays of a given type.

#### Using "merge-by-key" strategy

The "merge-by-key" strategy can be used to merge two slices together using an arbitrary merge key:

```go
type User struct {
    ID   int
    Name string
    Age  int
}
mergeKeyFunc := func(_ int, v reflect.Value) (reflect.Value, error) {
    return v.FieldByName("ID"), nil
}
v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
v2 := []User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithMergeByKeyFunc(reflect.TypeOf(User{}), mergeKeyFunc))
fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)
```

Output:

    DeepMerge([{ID:1 Name:Alice Age:0} {ID:2 Name:Bob Age:0}], [{ID:2 Name: Age:30} {ID:1 Name: Age:20}]) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]

This strategy is similar to Kubernetes' [strategic merge patch].

The function `mergeKeyFunc` must be of type `SliceMergeKeyFunc`. It will be invoked with the index
and value of the slice element to extract a merge key from.

The most common usage for this strategy is to merge slices of structs, where the merge key is the
name of a primary key field. In this case, we can use the `WithMergeByID` option to specify the
field name to use as merge key, and simplify the example above as follows:

```go
v1 := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
v2 := []User{{ID: 1, Age: 20}      , {ID: 2, Age: 30}}
merged, _ := goalesce.DeepMerge(v1, v2, goalesce.WithMergeByID(reflect.TypeOf(User{}), "ID"))
fmt.Printf("DeepMerge(%+v, %+v, MergeByID) = %+v\n", v1, v2, merged)
```

Output:

    DeepMerge([{ID:1 Name:Alice} {ID:2 Name:Bob}], [{ID:1 Age:20} {ID:2 Age:30}], MergeByID) = [{ID:1 Name:Alice Age:20} {ID:2 Name:Bob Age:30}]

The option `WithMergeByID` also works out-of-the-box on slices of pointers to structs:

```go
v1 := []*User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
v2 := []*User{{ID: 2, Age: 30}, {ID: 1, Age: 20}}
merged, _ = goalesce.DeepMerge(v1, v2, goalesce.WithMergeByID(reflect.TypeOf(User{}), "ID"))
jsn, _ := json.MarshalIndent(merged, "", "  ")
fmt.Println(string(jsn))
```

Output:

```json
[
  {
    "ID": 1,
    "Name": "Alice",
    "Age": 20
  },
  {
    "ID": 2,
    "Name": "Bob",
    "Age": 30
  }
]
```    

This strategy is not available for arrays.

### Merging structs

When both structs are non-zero-values, the default behavior is to merge the two structs field by
field, recursively merging their values.

```go
type User struct {
    ID   int
    Name string
    Age  int
}
v1 := User{ID: 1, Name: "Alice"}
v2 := User{ID: 1, Age: 20}
merged, _ := goalesce.DeepMerge(v1, v2)
fmt.Printf("DeepMerge(%+v, %+v) = %+v\n", v1, v2, merged)
```

Output:

    DeepMerge({ID:1 Name:Alice}, {ID:1 Age:20}) = {ID:1 Name:Alice Age:20}

#### Per-field merging strategies

When the default struct merging behavior is not desired or sufficient, per-field merging strategies
can be used.

The struct tag `goalesce` allows to specify the following per-field strategies:

| Strategy | Valid on               | Effect                              |
|----------|------------------------|-------------------------------------|
| `atomic` | Any field              | Applies "atomic" semantics.         |
| `union`  | Slice fields           | Applies "set-union" semantics.      |
| `append` | Slice fields           | Applies "list-append" semantics.    |   
| `index`  | Slice fields           | Applies "merge-by-index" semantics. |   
| `id`     | Slice of struct fields | Applies "merge-by-id" semantics.    |   

With the `id` strategy, a merge key must also be provided, separated by a colon from the strategy
name itself, e.g. `goalesce:"id:ID"`. The merge key _must_ be the name of an exported field in the
slice's struct element type.

Example:

```go
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
fmt.Println(string(jsn))
```

Output:

```json
{
  "Name": "The Matrix",
  "Description": "A computer hacker learns from mysterious rebels about the true nature of his reality and his role in the war against its controllers.",
  "Actors": [
    {
      "ID": 1,
      "Name": "Keanu Reeves"
    },
    {
      "ID": 2,
      "Name": "Laurence Fishburne"
    },
    {
      "ID": 3,
      "Name": "Carrie-Anne Moss"
    },
    {
      "ID": 4,
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

If you cannot annotate your struct with a `goalesce` tag, you can use the following options to
specify per-field strategies programmatically:

* `WithFieldListAppendMerge`
* `WithFieldListUnionMerge`
* `WithFieldMergeByIndex`
* `WithFieldMergeByID`
* `WithFieldMergeByKeyFunc`

See the [online documentation](https://pkg.go.dev/github.com/adutra/goalesce?tab=doc) for more examples.

### Custom mergers

The following options allow to pass a custom merger to the `DeepMerge` function:

* The `WithTypeMerger` option can be used to merge a given type with a custom merger.
* The `WithFieldMerger` option can be used to merge a given struct field with a custom merger.

Here is an example showcasing `WithTypeMerger`:

```go
summingMerger := func(v1, v2 reflect.Value) (reflect.Value, error) {
    result := reflect.New(v1.Type()).Elem()
    result.SetInt(v1.Int() + v2.Int())
    return result, nil
}
v1 := 1
v2 := 2
merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithTypeMerger(reflect.TypeOf(v1), summingMerger))
fmt.Printf("DeepMerge(%+v, %+v, WithTypeMerger) = %+v, %v\n", v1, v2, merged, err)
```

Output:

    DeepMerge(1, 2, WithTypeMerger) = 3, <nil>

It gets a bit more involved when the custom merger needs to access the global merge function, for
example, to delegate the merging of child values.

For these cases, there are 2 other options:

* The `WithTypeMergerProvider` option can be used to merge a given type with a custom
  `DeepMergeFunc`.
* The `WithFieldMergerProvider` option can be used to merge a given struct field with a custom
  `DeepMergeFunc`.

The above options give the custom merger access to the parent merger and the parent copier. 

Here is an example showcasing `WithFieldMergerProvider`:

```go
userMergerProvider := func(globalMerger goalesce.DeepMergeFunc, globalCopier goalesce.DeepCopyFunc) goalesce.DeepMergeFunc {
    return func(v1, v2 reflect.Value) (reflect.Value, error) {
        if v1.Int() == 1 {
            return reflect.Value{}, errors.New("user 1 has been deleted")
        }
        return globalMerger(v1, v2) // use global merger for other values
    }
}
v1 := User{ID: 1, Name: "Alice"}
v2 := User{ID: 1, Age: 20}
merged, err := goalesce.DeepMerge(v1, v2, goalesce.WithFieldMergerProvider(reflect.TypeOf(User{}), "ID", userMergerProvider))
fmt.Printf("DeepMerge(%+v, %+v, WithFieldMergerProvider) = %+v, %v\n", v1, v2, merged, err)
```

Output:

    DeepMerge({ID:1 Name:Alice Age:0}, {ID:1 Name: Age:20}, WithFieldMergerProvider) = {ID:0 Name: Age:0}, user 1 has been deleted


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
