# Go Generics API Design Patterns Research

## Executive Summary

This document summarizes patterns from Go's standard library (1.18+) and popular frameworks for designing user-friendly generic APIs. The key insight is that **Go generics minimize user friction through type inference**, allowing generic APIs to be called just as cleanly as non-generic ones, even though declarations are more verbose.

## 1. Go Standard Library Patterns

### 1.1 slices Package

The `slices` package (introduced in Go 1.18) provides the clearest examples of idiomatic generic API design.

**Key Design Patterns:**

1. **Type Parameter Naming:**
   - `S` for slice types (e.g., `S ~[]E`)
   - `E` for element types
   - Convention: Single uppercase letters (T, E, K, V, S)

2. **Constraint Patterns:**
   - Use `~[]E` to allow any type whose underlying type is a slice
   - `E comparable` for functions that need equality checks (Index, Contains)
   - `constraints.Ordered` for sorting operations

3. **Minimal User-Facing Verbosity:**
   ```go
   // Declaration (verbose):
   func Index[S ~[]E, E comparable](s S, v E) int

   // Usage (clean, type parameters inferred):
   idx := slices.Index(mySlice, "value")
   ```

4. **Return Value Patterns:**
   - Functions that change slice length return a new slice: `Compact`, `Delete`, `Insert`, `Replace`, `Grow`
   - Functions that only reorder elements don't return: `Sort`, `Reverse`
   - This pattern prevents memory leaks (before Go 1.22, `Delete` didn't zero elements)

**Example Functions:**
- `slices.Sort(s []E)` - works with any ordered type
- `slices.Index[S ~[]E, E comparable](s S, v E) int`
- `slices.Clone[S ~[]E, E any](s S) S`

### 1.2 maps Package

The `maps` package (experimental in x/exp/maps, standard library in Go 1.21) follows similar patterns.

**Key Design Patterns:**

1. **Type Parameter Naming:**
   - `K` for key types (must be `comparable`)
   - `V` for value types
   - `M` for map types (e.g., `M ~map[K]V`)

2. **Signature Example:**
   ```go
   func Clone[M ~map[K]V, K comparable, V any](m M) M
   ```

3. **Constraint Pattern:**
   - Map keys always require `comparable` constraint
   - Values typically use `any` unless specific operations needed
   - Use tilde (`~`) to allow user-defined types with map underlying types

**Example Functions:**
- `maps.Clone[M ~map[K]V, K comparable, V any](m M) M`
- `maps.EqualFunc` - custom comparison for values
- `maps.DeleteFunc` - predicate-based deletion

### 1.3 sync/atomic Package

Go 1.19 introduced generic types in `sync/atomic`:

**New Generic Types:**
- `atomic.Pointer[T any]` - type-safe atomic pointer operations
- `atomic.Int64`, `atomic.Uint32`, etc. - type-safe atomic numeric operations

**Design Pattern:**
- Generic types eliminate need for unsafe.Pointer casts
- Type parameter on the type itself, not on methods
- Methods operate on the constrained type

**Note:** The main `sync` package (sync.Map, sync.Pool) does NOT have generic versions due to backward compatibility. There are proposals for a future `sync/v2` package.

## 2. Popular Go Libraries with Generics

### 2.1 Framework Trends (2024-2025)

**Most Popular Frameworks:**
1. **Gin** (48% adoption) - leading framework, added generic support in recent versions
2. **Fiber** (11% adoption) - speed-focused, uses generics for type-safe middleware
3. **Echo** (16% adoption)
4. **Hertz** - CloudWeGo's high-performance framework with generics

### 2.2 Generic-First Libraries

Several libraries built specifically around generics:

1. **Type-Safe Collections:**
   - Generic thread-safe doubly linked list (replacement for container/list)
   - Swiss Map implementation with generics
   - Generic ordered maps with safety guarantees

2. **Configuration Libraries:**
   - `enflag` - container-oriented config using generics without reflection or struct tags

3. **Functional Libraries:**
   - Generic map/filter/reduce operations
   - Generic utility functions for common patterns

## 3. Key Questions Answered

### 3.1 Do libraries use type aliases for user convenience?

**Yes, but with important caveats:**

- **Go 1.24** introduced full support for generic type aliases
- **Primary Use Case:** Package refactoring and API evolution
  - When moving generic types between packages
  - Maintaining backward compatibility
  - Providing simpler names for complex generic types

**Example Pattern:**
```go
// Complex generic type
type Result[T any, E error] struct { ... }

// Convenience alias for common case
type StringResult = Result[string, error]
```

**Limitation Before 1.24:**
- Aliases couldn't have their own type parameters
- This limited their usefulness for generic APIs

### 3.2 How verbose are typical user-facing APIs?

**Minimal verbosity due to type inference:**

**Two Inference Mechanisms:**

1. **Function Argument Inference:**
   ```go
   // Declaration
   func Map[T any, U any](slice []T, fn func(T) U) []U

   // Usage - types inferred from arguments
   results := Map(numbers, strconv.Itoa)
   ```

2. **Constraint Type Inference:**
   ```go
   // Declaration
   func Sort[S ~[]E, E constraints.Ordered](s S)

   // Usage - both S and E inferred
   Sort(myInts)
   ```

**Key Principle:** "Calls to generic functions could be as clean as ordinary functions, even if generic function declarations are more verbose."

**Important Rules:**
- Type arguments can be omitted entirely OR specified entirely
- Cannot specify partial type arguments
- Inferred arguments must be a suffix of the type parameter list

### 3.3 What naming conventions are used?

**Standard Conventions:**

| Letter | Meaning | Usage Context |
|--------|---------|---------------|
| `T` | General type | Generic functions/types |
| `E` | Element | Slice/collection elements |
| `K` | Key | Map keys |
| `V` | Value | Map values |
| `S` | Slice | Slice type parameters |
| `M` | Map | Map type parameters |
| `R` | Result | Return type parameters |

**Best Practices:**
- Single uppercase letter is idiomatic
- Multiple type parameters: `T`, `U`, `V` or descriptive single letters
- Longer names acceptable if clarifying (e.g., `K comparable`)

### 3.4 How do they handle user's own concrete types?

**Three Key Patterns:**

1. **Use Tilde (~) for Underlying Types:**
   ```go
   // Without ~: only works with []int
   func Process[T []int](t T)

   // With ~: works with any type whose underlying type is []int
   func Process[T ~[]int](t T)

   // User can now use:
   type MyInts []int
   Process(MyInts{1, 2, 3}) // Works!
   ```

2. **Constraint Interfaces:**
   ```go
   type Number interface {
       ~int | ~int64 | ~float64 | ~float32
   }

   func Sum[T Number](values []T) T {
       // Works with int, MyInt, float64, MyFloat, etc.
   }
   ```

3. **Method-Based Constraints:**
   ```go
   type Stringer interface {
       String() string
   }

   func PrintAll[T Stringer](items []T) {
       // Works with any user type implementing String()
   }
   ```

**The `comparable` Constraint:**
- Predeclared in Go
- Required for map keys
- Allows any type usable with `==` and `!=`
- Works with user-defined types automatically

**Type Inference Benefits:**
```go
// Library provides:
func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V

// User with custom types:
type MyKey string
type MyValue int64

myMap := map[MyKey]MyValue{"a": 1, "b": 2}
result := SumIntsOrFloats(myMap) // Just works! Types inferred.
```

## 4. Practical Recommendations for API Design

### 4.1 Minimize User-Facing Complexity

1. **Rely on Type Inference:**
   - Design functions so type parameters can be inferred
   - Place inferrable parameters early in type parameter list
   - Test that explicit type arguments aren't needed

2. **Use Standard Constraints:**
   - Prefer `comparable`, `any`, `constraints.Ordered`
   - Define reusable constraint interfaces
   - Document what operations your constraints enable

3. **Follow Naming Conventions:**
   - Use single-letter type parameters (T, E, K, V)
   - Be consistent across your package
   - Match stdlib patterns where possible

### 4.2 Design for User Types

1. **Always Use Tilde (~):**
   ```go
   // Good: works with user-defined types
   func Process[S ~[]E, E any](s S)

   // Bad: only works with exact []E type
   func Process[S []E, E any](s S)
   ```

2. **Return Concrete Types When Possible:**
   - Helps with type inference
   - Reduces cascading type parameters
   - Example: `slices.Clone` returns same type as input

3. **Consider Functional Options:**
   ```go
   // Instead of multiple type parameters:
   func Sort[T any](slice []T, less func(T, T) bool)

   // Users provide behavior via functions
   ```

### 4.3 Compile-Time Safety

**Key Advantage of Generics:** Type checking at compile time

**Patterns:**
- Older APIs used `interface{}` → runtime panics
- Generic APIs catch type errors at compile time
- Example: `slices.Sort` vs old `sort.Interface`

```go
// Old way: runtime error if comparison fails
sort.Slice(items, func(i, j int) bool {
    return items[i].(int) < items[j].(string) // Compiles, panics at runtime!
})

// Generic way: compile error
slices.SortFunc(items, func(a int, b string) int { // Won't compile
    return cmp.Compare(a, b)
})
```

### 4.4 API Evolution

**Go 1.24 Generic Type Aliases Enable:**
- Moving types between packages while maintaining compatibility
- Creating convenience aliases for common instantiations
- Gradual migration paths

**Example:**
```go
// old/package
type Container[T any] struct { value T }

// new/package
type Container[T any] = old.Container[T] // Fully compatible
```

## 5. Anti-Patterns to Avoid

### 5.1 Over-Parameterization

```go
// Bad: too many type parameters
func Process[T, U, V, W, X any](t T, u U, v V, w W) X

// Good: use concrete types where possible
func Process[T any](t T, config Config) Result
```

### 5.2 Forcing Type Arguments

```go
// Bad: user must specify types
result := Process[string, int]("hello", 42)

// Good: design for inference
result := Process("hello", 42)
```

### 5.3 Ignoring Underlying Types

```go
// Bad: doesn't work with user's custom int type
func Sum[T int | int64](values []T) T

// Good: works with any underlying int
func Sum[T ~int | ~int64](values []T) T
```

## 6. Performance Considerations

**Go 1.19 Improvements:**
- Up to 20% performance improvements for generic programs
- Continued refinement of edge cases
- Generic code now competitive with hand-written type-specific code

**Best Practices:**
- Generics don't add runtime overhead (monomorphization)
- Use generics freely where type safety matters
- Profile if performance-critical

## Sources

### Go Official Resources
- [Robust generic functions on slices](https://go.dev/blog/generic-slice-functions)
- [Tutorial: Getting started with generics](https://go.dev/doc/tutorial/generics)
- [An Introduction To Generics](https://go.dev/blog/intro-generics)
- [Everything You Always Wanted to Know About Type Inference](https://go.dev/blog/type-inference)
- [Go 1.19 is released!](https://go.dev/blog/go1.19)
- [What's in an (Alias) Name?](https://go.dev/blog/alias-names)
- [Generic interfaces](https://go.dev/blog/generic-interfaces)

### GitHub Discussions & Issues
- [slices: new package to provide generic slice functions · Issue #45955](https://github.com/golang/go/issues/45955)
- [proposal: slices: new package · Discussion #47203](https://github.com/golang/go/discussions/47203)
- [proposal: maps: new package · Discussion #47330](https://github.com/golang/go/discussions/47330)
- [how to update APIs for generics · Discussion #48287](https://github.com/golang/go/discussions/48287)
- [spec: generics: permit type parameters on aliases · Issue #46477](https://github.com/golang/go/issues/46477)

### Community Resources
- [Generic Instantiations and Type Argument Inferences - Go 101](https://go101.org/generics/666-generic-instantiations-and-type-argument-inferences.html)
- [Understanding Go 1.21 generics type inference – Encore Blog](https://encore.dev/blog/go1.21-generics)
- [Sort a slice of any type using Generics in Go](https://gosamples.dev/generics-sort-slice/)
- [Generic Type Aliases in Go 1.24 | Medium](https://medium.com/@okoanton/generic-type-aliases-in-go-1-24-what-they-are-and-why-we-need-them-07ca05539500)
- [Top 5 Popular Frameworks and Libraries for Go in 2025](https://dev.to/empiree/top-5-popular-frameworks-and-libraries-for-go-in-2024-c6n)
- [The Go Ecosystem in 2025 | The GoLand Blog](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)

### Package Documentation
- [slices package](https://pkg.go.dev/slices)
- [maps package](https://pkg.go.dev/maps)
- [sync/atomic package](https://pkg.go.dev/sync/atomic)

## Conclusion

**Key Takeaways for Generic API Design:**

1. **Type inference is king** - Design APIs so users rarely specify type arguments
2. **Use tilde (~) liberally** - Support user-defined types automatically
3. **Follow stdlib conventions** - T, E, K, V naming; comparable for equality
4. **Keep declarations verbose, calls clean** - The burden is on the library author
5. **Leverage Go 1.24 type aliases** - For API evolution and convenience
6. **Compile-time safety wins** - Generics catch errors that interface{} misses

The Go community has converged on patterns that make generic APIs feel natural and idiomatic while maintaining Go's simplicity and clarity.
