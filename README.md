# Automapper

Fast, reflection-based object mapper for Go.

Automapper maps values between different Go types using pre-built codecs. It supports structs, slices, arrays, maps, pointers, type conversions, field renaming, ignored fields, and custom codecs.

## Features

- Fast runtime mapping
- Codec-based architecture
- Automatic struct field mapping by field name
- Struct field renaming with `mapper` tags
- Struct field ignoring with `mapper:"-"`
- Nested struct mapping
- Assignable type support
- Convertible type support
- Slice and array mapping
- Map key and value mapping
- Pointer-aware mapping
- Custom codecs
- Atomic, concurrency-safe custom codec configuration
- Generic API
- Codec creation separated from value mapping
- Reusable codecs for repeated mappings
- No runtime locking for custom codec reads

## Installation

~~~bash
go get github.com/vedadiyan/automapper
~~~

## Basic Usage

~~~go
package main

import (
	"fmt"

	mapper "github.com/vedadiyan/automapper"
)

type User struct {
	ID   int
	Name string
}

type UserDTO struct {
	ID   int
	Name string
}

func main() {
	mapUser := mapper.CreateCodecFor[User, UserDTO]()

	user := User{
		ID:   42,
		Name: "John",
	}

	result, err := mapUser(&user)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", *result)
}
~~~

`CreateCodecFor` performs type analysis once and returns a reusable mapping function.

## How It Works

Automapper separates codec creation from value mapping.

A codec has the following form:

~~~go
type Codec func(sourceValue RValue, targetValue RValue) error
~~~

Codec factories determine whether they can handle a particular source/target type pair:

~~~go
type CodecFactory func(sourceField RType, targetField RType) Codec
~~~

When a codec is created, Automapper recursively builds the codecs required for nested fields, collections, and maps.

The resulting codec can then be reused without repeating codec discovery.

### Codec Resolution Order

Codecs are checked in this order:

1. Custom codec
2. Assignable type
3. Convertible type
4. Struct
5. Array / slice
6. Map

The first matching codec is used.

## Struct Mapping

Struct fields are matched by name.

~~~go
type Source struct {
	ID    int
	Name  string
	Email string
}

type Target struct {
	ID   int
	Name string
}
~~~

`ID` and `Name` are mapped automatically.

`Email` is ignored because the target has no matching field.

Nested structs are handled recursively.

~~~go
type Source struct {
	User User
}

type Target struct {
	User UserDTO
}
~~~

## Struct Tags

Automapper supports the `mapper` struct tag for renaming and ignoring fields.

### Rename a Field

Use the target field name as the tag value:

~~~go
type Source struct {
	UserID int `mapper:"ID"`
	Name   string
}

type Target struct {
	ID   int
	Name string
}
~~~

`UserID` is mapped to `ID`.

### Ignore a Field

Use `mapper:"-"`:

~~~go
type Source struct {
	ID     int
	Name   string
	Secret string `mapper:"-"`
}

type Target struct {
	ID   int
	Name string
}
~~~

`Secret` is skipped during codec creation.

### No Tag

Without a tag, the field name is used:

~~~go
type Source struct {
	ID   int
	Name string
}
~~~

This is equivalent to:

~~~go
type Source struct {
	ID   int    `mapper:"ID"`
	Name string `mapper:"Name"`
}
~~~

The tag only affects the source field. The target field itself does not require a tag.

## Type Conversion

Assignable and convertible Go types are handled automatically.

For example:

~~~go
type Source struct {
	ID int
}

type Target struct {
	ID int64
}
~~~

If the source type is convertible to the target type, Automapper uses Go's `reflect.Value.Convert`.

## Slices and Arrays

Slices and arrays are mapped recursively when their element types are compatible.

~~~go
type Source struct {
	Users []User
}

type Target struct {
	Users []UserDTO
}
~~~

The element codec is created once and reused for every element.

The following combinations are supported:

- Slice → Slice
- Array → Array
- Slice → Array
- Array → Slice

## Maps

Maps are mapped recursively, including both keys and values.

~~~go
type Source struct {
	Users map[int]User
}

type Target struct {
	Users map[int]UserDTO
}
~~~

Both the key and value types must have a compatible codec.

## Pointer Support

Pointer levels are tracked automatically.

~~~go
type Source struct {
	Name *string
}

type Target struct {
	Name **string
}
~~~

Automapper handles the required pointer levels when assigning mapped values.

## Custom Codecs

Use a custom codec when automatic mapping is not sufficient.

Create one with `CreateCustomCodec`:

~~~go
custom := mapper.CreateCustomCodec[User, UserDTO](
	func(value mapper.RValue) (mapper.RValue, error) {
		user := value.ConcreteValue().Interface().(User)

		return mapper.ValueOf(UserDTO{
			ID:   user.ID,
			Name: user.Name,
		}), nil
	},
)
~~~

Register it:

~~~go
mapper.SetCustomCodecs([]mapper.CodecFactory{
	custom,
})
~~~

Custom codecs take precedence over all built-in codecs.

### Why `RValue`?

The custom codec receives an `RValue` instead of `*T`.

This keeps the custom-codec path inside Automapper's reflection/value abstraction and avoids converting the value into an intermediate Go value solely for the callback.

For example:

~~~go
func(value mapper.RValue) (mapper.RValue, error) {
	user := value.ConcreteValue().Interface().(User)

	return mapper.ValueOf(UserDTO{
		ID:   user.ID,
		Name: user.Name,
	}), nil
}
~~~

## Custom Codec Configuration

Custom codecs are configured as a complete set:

~~~go
mapper.SetCustomCodecs([]mapper.CodecFactory{
	codec1,
	codec2,
	codec3,
})
~~~

Calling `SetCustomCodecs` replaces the previous configuration.

The supplied slice is copied before being published, so later modifications to the caller's slice do not affect Automapper.

### Concurrency

Custom codec configuration uses `atomic.Pointer` rather than a mutex.

Reads therefore require no locking:

~~~go
func CustomCodec(sourceField RType, targetField RType) Codec {
	for _, codec := range *customCodecs.Load() {
		if fn := codec(sourceField, targetField); fn != nil {
			return fn
		}
	}

	return nil
}
~~~
Replacing the custom codec configuration is atomic, making concurrent codec creation and mapping safe.

`SetCustomCodecs` uses replace-all semantics rather than add/remove semantics.

## Reusing a Codec

For performance-sensitive code, create the codec once and reuse it:

~~~go
codec := mapper.CreateCodecFor[Source, Target]()
if codec == nil {
	// No compatible mapping exists.
	return
}

for i := range sources {
	target, err := codec(&sources[i])
	if err != nil {
		return err
	}

	// use target
}
~~~

This avoids repeatedly performing codec discovery for the same source/target pair.

For high-throughput workloads, this is the recommended usage pattern.

## Direct Codec Creation

Automapper also exposes the lower-level API:

~~~go
codec := mapper.CreateCodec(
	mapper.TypeFor[Source]().ConcreteType(),
	mapper.TypeFor[Target]().ConcreteType(),
)
~~~

This is useful when working directly with `RType` and `RValue`.

For normal generic usage, prefer:

~~~go
mapper.CreateCodecFor[Source, Target]()
~~~

## Performance

Example benchmark results on an AMD EPYC 7763:

~~~text
BenchmarkAutomapper_Struct-2             1434258    825.6 ns/op     80 B/op     3 allocs/op
BenchmarkAutomapper_NestedStruct-2       1000000   1091   ns/op    104 B/op     4 allocs/op
BenchmarkAutomapper_Slice-2               663115   1945   ns/op    216 B/op    14 allocs/op
BenchmarkAutomapper_Map-2                 515337   2413   ns/op    448 B/op    20 allocs/op
~~~

Compared with the benchmarked `copier` implementation:

~~~text
                         Automapper       Copier

Struct                   825.6 ns/op      3844 ns/op
Nested struct           1091   ns/op      5623 ns/op
Slice                     1945 ns/op      2675 ns/op
Map                       2413 ns/op      3770 ns/op
~~~

These benchmarks are workload-dependent. Actual performance depends on the structures being mapped, nesting depth, collection sizes, pointer levels, and custom codecs.

Benchmark your own workload before making performance assumptions.

## Architecture

Automapper has three main concepts.

~~~text
                 Source / Target Types
                          │
                          ▼
                    CodecFactory
                          │
                          ▼
                        Codec
                          │
                          ▼
                 Source RValue
                          │
                          ▼
                 Target RValue
~~~

### `RType`

`RType` wraps `reflect.Type` and caches information used repeatedly during mapping:

- Original Go type
- Concrete/non-pointer type
- Pointer count

### `RValue`

`RValue` wraps `reflect.Value` and similarly tracks:

- Original value
- Concrete value
- Pointer count

This keeps pointer handling and reflection operations centralized rather than duplicating them throughout every codec.

### Codec Factory

A `CodecFactory` answers one question:

> Can this factory map this source type to this target type?

If yes, it returns a `Codec`.

If not, it returns `nil`.

This makes adding new mapping strategies straightforward.

## Supported Mapping Rules

| Source | Target | Support |
|---|---|---|
| Assignable type | Compatible type | Yes |
| Convertible type | Convertible type | Yes |
| Struct | Struct | Yes |
| Slice | Slice | Yes |
| Array | Array | Yes |
| Slice | Array | Yes |
| Array | Slice | Yes |
| Map | Map | Yes |
| Pointer | Pointer | Yes |
| Custom codec | Registered pair | Yes |
| Nested structs | Nested structs | Yes |
| Nested collections | Nested collections | Yes |
| Map keys | Mappable keys | Yes |
| Map values | Mappable values | Yes |
| Field rename | `mapper:"FieldName"` | Yes |
| Field ignore | `mapper:"-"` | Yes |

## Limitations

Automapper matches struct fields by name unless a `mapper` tag specifies another source field name.

For example:

~~~go
type Source struct {
	UserID int
}

type Target struct {
	ID int
}
~~~

`UserID` will not automatically map to `ID`.

Use:

~~~go
type Source struct {
	UserID int `mapper:"ID"`
}
~~~

or a custom codec when explicit transformation logic is required.

Automapper does not attempt to perform arbitrary business-logic transformations automatically. It focuses on structural and type-compatible mapping.

## Cycle Loops

Automapper does not support cyclic type definitions.

If the target type contains a cycle, `CreateCodecFor` detects it and returns a codec that reports an error when used:

~~~go
type Node struct {
	Value int
	Next  *Node
}

codec := mapper.CreateCodecFor[Node, Node]()

_, err := codec(&Node{Value: 1})
// err: cycle loop detected
~~~

Cycle loops are intentionally rejected to prevent infinite recursive codec construction and mapping.

Automapper supports recursive/nested structures only when the type graph is acyclic.

## Design Goals

Automapper is designed around a few principles:

- **Build once, execute many times** — type analysis happens when the codec is created.
- **Keep the hot path small** — mapping uses already-created codecs.
- **Avoid unnecessary synchronization** — custom codec reads use atomic publication rather than locks.
- **Composable codecs** — structs, arrays, slices, and maps recursively reuse codecs.
- **Explicit customization** — custom codecs handle transformations that cannot be inferred safely.
- **Use Go's type system** — assignability and convertibility follow Go reflection semantics.
- **Keep configuration predictable** — custom codec configuration uses atomic replace-all semantics.

## Requirements

Go 1.19+

## License

See the repository for license information.