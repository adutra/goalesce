# Package goalesce

Package goalesce is a library for coalescing (a.k.a. merging) objects in Go. It can coalesce any type of object,
including structs, maps, and slices, even nested ones.

The main entry point is the `Coalesce` function:

    func Coalesce(o1, o2 interface{}, opts ...MainCoalescerOption) (coalesced interface{}, err error)

Simple usage:

```go
type Foo struct {
    Field1 int
    Field2 string
}
coalesced, _ = goalesce.Coalesce(&Foo{Field1: 1}, &Foo{Field2: "abc"})
fmt.Printf("Coalesce(&{1}, &{abc}) = %v\n", coalesced)
```

Output:

	Coalesce(&{1}, &{abc}) = &{1 abc}

See the [online documentation](https://pkg.go.dev/github.com/adutra/goalesce?tab=doc) for more details.
