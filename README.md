# Automapper

Fast, reflection-based object mapper for Go.

Automapper maps values between different Go types using pre-built codecs. It supports structs, slices, arrays, maps, pointers, type conversions, and custom codecs.

## Features

- Fast runtime mapping
- Automatic struct field mapping by field name
- Convertible type support
- Slice and array mapping
- Map mapping
- Pointer-aware mapping
- Custom codecs
- Concurrency-safe custom codec configuration
- Generic API
- Codec creation separated from value mapping

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

## How It Works

Automapper builds a `Codec` from the source and target types.

~~~go
type Codec func(sourceValue RValue, targetValue RValue) error
~~~

A codec is created once and can then be reused:

~~~go
codec := mapper.CreateCodecFor[User, UserDTO]()

for _, user := range users {
	dto, err := codec(&user)
	if err != nil {
		return err
	}

	// use dto
}
~~~

The mapper resolves codecs in this order:

1. Custom codec
2. Assignable type
3. Convertible type
4. Struct
5. Array / slice
6. Map

## Struct Mapping

Fields are matched by name.

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

`Email` is ignored because the target does not contain a matching field.

Nested structs are handled automatically.

## Slices and Arrays

Slices and arrays are supported when their element types can be mapped.

~~~go
type Source struct {
	Users []User
}

type Target struct {
	Users []UserDTO
}
~~~

The element codec is created once and reused for every element.

## Maps

Maps are mapped recursively, including their keys and values.

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

Pointer levels are handled automatically.

~~~go
type Source struct {
	Name *string
}

type Target struct {
	Name **string
}
~~~

## Custom Codecs

For cases where automatic mapping is not sufficient, create a custom codec:

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

Custom codecs take precedence over the built-in codecs.

### Replacing Custom Codecs

`SetCustomCodecs` replaces the complete custom-codec set.

~~~go
mapper.SetCustomCodecs([]mapper.CodecFactory{
	codec1,
	codec2,
})
~~~

The supplied slice is copied before being published.

The configuration is published atomically, so replacing custom codecs is safe while mappings are running concurrently.

## Reusing a Codec

For performance-sensitive code, create the codec once and reuse it:

~~~go
codec := mapper.CreateCodecFor[Source, Target]()

for i := range sources {
	target, err := codec(&sources[i])
	if err != nil {
		return err
	}

	// use target
}
~~~

This avoids repeatedly performing codec discovery for the same source/target pair.

## Performance

Example benchmark results on an AMD EPYC 7763:

~~~text
BenchmarkAutomapper_Struct-2          1434258    825.6 ns/op     80 B/op     3 allocs/op
BenchmarkAutomapper_NestedStruct-2   1000000   1091   ns/op    104 B/op     4 allocs/op
BenchmarkAutomapper_Slice-2           663115   1945   ns/op    216 B/op    14 allocs/op
BenchmarkAutomapper_Map-2             515337   2413   ns/op    448 B/op    20 allocs/op
~~~

Compared with the benchmarked `copier` implementation:

~~~text
                         Automapper       Copier

Struct                   825.6 ns/op      3844 ns/op
Nested struct           1091   ns/op      5623 ns/op
Slice                     1945 ns/op      2675 ns/op
Map                       2413 ns/op      3770 ns/op
~~~

Benchmarks are workload-dependent. Benchmark your own types and mapping patterns before making performance assumptions.

## Design

Automapper separates type inspection from value mapping.

~~~text
Source/Target Types
        │
        ▼
   CodecFactory
        │
        ▼
      Codec
        │
        ▼
Source Value ──────► Target Value
~~~

A `CodecFactory` determines whether it can handle a source/target type pair.

~~~go
type CodecFactory func(
	sourceField RType,
	targetField RType,
) Codec
~~~

Once created, the resulting `Codec` performs the actual mapping.

This means type analysis happens during codec creation rather than repeatedly during every field mapping operation.

## Supported Mapping Rules

| Source | Target | Support |
|---|---|---|
| Assignable types | Same/compatible type | Yes |
| Convertible types | Convertible type | Yes |
| Struct | Struct | Yes |
| Slice | Slice | Yes |
| Array | Array | Yes |
| Slice | Array | Yes |
| Array | Slice | Yes |
| Map | Map | Yes |
| Pointer types | Pointer types | Yes |
| Custom codec | Explicitly registered pair | Yes |

## Limitations

Automapper uses field names for automatic struct mapping.

It does not guess semantic relationships between differently named fields.

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

Use a custom codec when explicit transformation logic is required.

## Requirements

Go 1.19+

## License

See the repository for license information.