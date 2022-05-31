[![GoVersion][GoVersionImg]][GoVersionLink]
[![GoDoc][GoDocImg]][GoDocLink]
[![GoReport][GoReportImg]][GoReportLink]
[![CodeCov][CodeCovImg]][CodeCovLink]

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
  - If both values are interfaces of same underlying types, coalesce the underlying values.
  - If both values are pointers, coalesce the values pointed to.
  - If both values are maps, coalesce the maps recursively, key by key.
  - If both values are structs, coalesce the structs recursively, field by field.
  - For other types (including slices), return the second value ("atomic" semantics).

Note that by default, slices are coalesced with atomic semantics, that is, the second slice overwrites the first one
completely. It is possible to change this behavior, see examples below.

The `Coalesce` function can be called with a list of options to modify its default coalescing behavior. See the
documentation of each option for details.

## Examples 

### Coalescing scalars

Scalars are always coalesced with atomic semantics when both values are non-zero-values:

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

When both slices are non-zero-values, the default behavior is to apply atomic semantics, that is, to _replace_ the first
slice with the second one:

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

#### Treating empty slices as zero-values

To consider empty slices as zero-values, use the `WithZeroEmptySlice` option. This changes the default behavior: when
coalescing a non-empty slice with an empty slice, normally the empty slice is returned as in the example above; but with
this option, the non-empty slice is returned.

```go
v1 = []int{1, 2}
v2 = []int{} // empty slice will be considered zero-value
coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithZeroEmptySlice())
fmt.Printf("Coalesce(%+v, %+v, ZeroEmptySlice) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce([1 2], [], ZeroEmptySlice) = [1 2]

#### Using "set-union" strategy

The "set-union" strategy can be used to coalesce the two slices together by creating a resulting slice that contains all
elements from both slices, but no duplicates:

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultSetUnion())
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
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultSetUnion())
for i, elem := range coalesced.([]*int) {
    fmt.Printf("%v: %T %v\n", i, elem, *elem)
}
```

Output:

    0: *int 0
    1: *int 1

This strategy is fine for slices of scalars and pointers thereof, but it is not recommended for slices of complex 
types as the elements may not be fully comparable. Also, it is not suitable for slices of double pointers.

The resulting slice's element order is deterministic: each element appears in the order it was first encountered when 
iterating over the two slices.

#### Using "list-append" strategy

The "list-append" strategy appends the second slice to the first one (possibly resulting in duplicates):

```go
v1 := []int{1, 2}
v2 := []int{2, 3}
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultListAppend())
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
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithDefaultMergeByIndex())
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
mergeKeyFunc := func(_ int, v reflect.Value) (reflect.Value, error) {
    return v.FieldByName("Id"), nil
}
v1 := []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByKeyFunc(reflect.TypeOf(User{}), mergeKeyFunc))
fmt.Printf("Coalesce(%+v, %+v) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce([{Id:1 Name:Alice Age:0} {Id:2 Name:Bob Age:0}], [{Id:2 Name: Age:30} {Id:1 Name: Age:20}]) = [{Id:1 Name:Alice Age:20} {Id:2 Name:Bob Age:30}]

This strategy is similar to Kubernetes' [strategic merge patch].

The function `mergeKeyFunc` must be of type `SliceMergeKeyFunc`. It will be invoked with the index and value of the
slice element to extract a merge key from.

The most common usage for this strategy is to coalesce slices of structs, where the merge key is the name of a primary
key field. In this case, we can use the `WithMergeByID` option to specify the field name to use as merge key, and
simplify the example above as follows:

```go
v1 := []User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []User{{Id: 1, Age: 20}      , {Id: 2, Age: 30}}
coalesced, _ := goalesce.Coalesce(v1, v2, goalesce.WithMergeByID(reflect.TypeOf(User{}), "Id"))
fmt.Printf("Coalesce(%+v, %+v, MergeByID) = %+v\n", v1, v2, coalesced)
```

Output:

    Coalesce([{Id:1 Name:Alice} {Id:2 Name:Bob}], [{Id:1 Age:20} {Id:2 Age:30}], MergeByID) = [{Id:1 Name:Alice 
Age:20} {Id:2 Name:Bob Age:30}]

The option `WithMergeByID` also works out-of-the-box on slices of pointers to structs:

```go
v1 := []*User{{Id: 1, Name: "Alice"}, {Id: 2, Name: "Bob"}}
v2 := []*User{{Id: 2, Age: 30}, {Id: 1, Age: 20}}
coalesced, _ = goalesce.Coalesce(v1, v2, goalesce.WithMergeByID(reflect.TypeOf(User{}), "Id"))
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

The struct tag `goalesce` allows to specify the following per-field strategies:

| Strategy   | Valid on               | Effect                              |
|------------|------------------------|-------------------------------------|
| `atomic`   | Any field              | Applies "atomic" semantics.         |
| `union`    | Slice fields           | Applies "set-union" semantics.      |
| `append`   | Slice fields           | Applies "list-append" semantics.    |   
| `index`    | Slice fields           | Applies "merge-by-index" semantics. |   
| `merge`    | Slice of struct fields | Applies "merge-by-key" semantics.   |   

With `merge`, a merge key must also be provided, separated by a comma from the strategy name itself, e.g.
`goalesce:"merge,Id"`.  The merge key _must_ be the name of a field in the slice's struct element type.

Example:

```go
type Actor struct {
    Id   int
    Name string
}
type Movie struct {
    Name        string
    Description string
    Actors      []Actor           `goalesce:"merge,Id"`
    Tags        []string          `goalesce:"union"`
    Labels      map[string]string `goalesce:"atomic"`
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

If you cannot annotate your struct with a `goalesce` tag, you can use the following options to specify per-field 
strategies programmatically:

* `WithFieldListAppend`
* `WithFieldListUnion`
* `WithFieldMergeByIndex`
* `WithFieldMergeByID`
* `WithFieldMergeByKeyFunc`

See the [online documentation](https://pkg.go.dev/github.com/adutra/goalesce?tab=doc) for more examples.

## Advanced usage

The `Coalescer` function allows for custom coalescing algorithms to be implemented. By passing custom coalescers as
options to the `Coalesce` function, its behavior can be modified in any way.

The following options allow to pass a custom coalescer to the `Coalesce` function:

* The `WithTypeCoalescer` option can be used to coalesce a given type with a custom `Coalescer`.
* The `WithFieldCoalescer` option can be used to coalesce a given struct field with a custom `Coalescer`.

* Here is an example showcasing `WithTypeCoalescer`:

```go
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
```

It gets a bit more involved when the custom coalescer needs to access its parent coalescer, for example, to delegate the
coalescing of child values.

For these cases, there are 2 other options:

* The `WithTypeCoalescerProvider` option can be used to coalesce a given type with a custom `Coalescer`.
* The `WithFieldCoalescerProvider` option can be used to coalesce a given struct field with a custom `Coalescer`.

The above options give the custom coalescer access to the parent coalescer. Here is an example showcasing
`WithFieldCoalescerProvider`:

```go
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
```

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
