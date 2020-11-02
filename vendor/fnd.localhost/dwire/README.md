
# dwire

A library for encoding and decoding things using Footnote's wire format.

## Usage

To encode a single value:

```go
import (
	"fnd.localhost/dwire"
	"fnd"
)

value := "a string value"
var w bytes.Buffer
if err := dwire.EncodeField(&w, value); err != nil {
	log.Fatal(err)
}
```

To encode a struct:

```go
import (
	"fnd.localhost/dwire"
	"bytes"
)

type Foo struct {
	StrField string
	Uint8Field uint8
}

foo := &Foo{
	StrField: "some string",
	Uint8Field: 1
}

var w bytes.Buffer
if err := dwire.EncodeFields(&w, foo.StrField, foo.Uint8Field); err != nil {
	log.Fatal(err)
}
```

Decoding works on pointer values. For example, to decode a single value:

```go
import (
	"fnd.localhost/dwire"
	"bytes"
)

var uint8Value uint8
r := bytes.NewReader([]byte{ 0x01 })
if err := dwire.DecodeField(r, &uint8Value); err != nil {
	log.Fatal(err)
}
```

To decode a struct:

```go
import (
	"fnd.localhost/dwire"
	"bytes"
)

type Foo struct {
	Uint8Field uint8
}

var foo Foo
r := bytes.NewReader([]byte{0x01})

if err := dwire.DecodeFields(r, &foo.Uint8Field); err != nil {
	log.Fatal(err)
}
```

For convenience, objects that implement the `Encoder` and `Decoder` interfaces can be passed directly to `dwire`'s encoding/decodig methods. For example, the first struct example above could be implemented as follows:

```go
import (
	"fnd.localhost/dwire"
	"bytes"
)

type Foo struct {
	StrField string
	Uint8Field uint8
}

func (f *Foo) Encode(w io.Writer) error {
	return dwire.EncodeFields(
		w,
		f.StrField,
		f.Uint8Field,
	)
}

func (f *Foo) Decode(r io.Reader) error {
	return dwire.DecodeFields(
		r,
		&f.StrField,
		&f.Uint8Field,
	)
}

foo := &Foo{
	StrField: "some string",
	Uint8Field: 1
}

var w bytes.Buffer
if err := dwire.EncodeField(&w, foo); err != nil {
	log.Fatal(err)
}
```

## Wire Format Definition

This definition is lifted from the canonical specification described in [PIP-1](https://fnd.network/docs/spec/pip-1.html). For clarity, we will use the function `Encode(t) = b`, where `t` is the inputted field and `b` represents the outputted bytes, to provide example encodings where necessary.

`dwire` defines encodings for the following types:

1. `bool`: Encoded as `0x01` or `0x00` if the value is `true` or `false`, respectively.
2. `uint8`, `uint16`, `uint32`, `uint64`: Encoded as big-endian unsigned integers.
3. `byte`: Encoded as `uint8`.
4. `[N]<T>` (i.e., fixed-length arrays of element `T`): Encoded as the concatenation of `Encode(<T>)` for each array element.
5. `[]<T>` (i.e., variable-length arrays of element `T`): Encoded as the concatenation of a `Uvarint` length prefix and `Encode(<T>)` for each array element.
6. `string`: Encoded as `[]byte`.

### Well-Known Types

Certain complex types are considered "well known," and are encoded/decoded directly in this package without implementing the `Encoder`/`Decoder` interfaces. These types are:

1. `time.Time`: Encoded as a `uint64` Unix timestamp.

## Benchmarks

Benchmarks recorded on a mid-2018 MacBook Pro with the following specifications:

1. 2.2GHz 6-Core Intel i7
2. 32GB 2400MHz DDR4

```
goos: darwin
goarch: amd64
pkg: fnd.localhost/dwire
BenchmarkUint8Encoding-12                 	  711883	      1469 ns/op
BenchmarkUint16Encoding-12                	  867793	      1472 ns/op
BenchmarkUint32Encoding-12                	  841219	      1446 ns/op
BenchmarkUint64Encoding-12                	  845859	      1544 ns/op
BenchmarkByteSliceEncoding1024-12         	  744835	      1623 ns/op
BenchmarkStringEncoding1024-12            	  710882	      1845 ns/op
BenchmarkByteArrayEncoding32-12           	  748132	      1551 ns/op
BenchmarkByteArrayEncodingReflect-12      	  577855	      2178 ns/op
BenchmarkUint16ArrayEncodingReflect-12    	   30091	     41010 ns/op
BenchmarkStringArrayEncodingReflect-12    	   10000	    121458 ns/op
BenchmarkWellKnownTimeEncoding-12         	  701935	      1845 ns/op
```

## Acknowledgements

Much of this library was directly inspired by `lnwire`, the wire encoding library used by the Lightning Network. As such, we would like to extend our deepest thanks to the Lightning team for their trailblazing work.  
