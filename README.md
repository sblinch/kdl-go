# KDLv1 Library for Go

kdl-go is a Go library for [version 1](https://kdl.dev/spec-v1/) of the [KDL Document Language](https://kdl.dev/). It supports encoding and decoding KDL
documents, marshaling and unmarshaling them into Go structs. 


# Features

- supports all KDLv1 language features and passes all of the official KDLv1 test cases as of time of writing
- designed with performance and usability in mind
- familiar API and tag syntax, similar to `encoding/json`
- supports marshaling/unmarshaling into Go structures with support for `encoding.Text(Un)Marshaler` and its own custom
  marshal/unmarshal interfaces
- support for `encoding/json/v2`-style `format` options for `time.Time`, `time.Duration`, `[]byte`, and `float32/64`
- contextual errors, including the line and column of each error and a sample line displaying the error location


# Import

```go
import "github.com/sblinch/kdl-go"
````


# Decoding

`Parse()` decodes KDL to a `*document.Document`:

```go
data := `
    name "Bob"
    age 76
    active true
`

if doc, err := kdl.Parse(strings.NewReader(data)); err == nil {
    // print the top-level nodes
    for _, node := range doc.Nodes {
        fmt.Println(node.Name.String())
    }
}
```
```go
// output
name
age
active
```


# Encoding

`Generate()` generates KDL from a `*document.Document`:

```go
data := `
    name "Bob"
    age 76
    active true
`

if doc, err := Parse(strings.NewReader(data)); err == nil {
    // output the KDL representation of doc to stdout
    if err := Generate(doc, os.Stdout); err != nil {
        panic(err)
    }
}
```
```kdl
// output:
name "Bob"
age 76
active true
```


# Unmarshaling

## via Unmarshal

`Unmarshal()` unmarshals KDL to a Go `map` or `struct`. The `kdl` tag can be used to map KDL node names to struct fields
or otherwise change unmarshaling behavior:

```go
type Person struct {
    Name        string      `kdl:"name"`
    Age         int         `kdl:"age"`
    Active      bool        `kdl:"active"`
}

data := `
    name "Bob"
    age 76
    active true
`

var person Person
if err := kdl.Unmarshal(data, &person); err == nil {
    fmt.Printf("%+v\n",person)
}
```
```go
// output
Person{
    Name: "Bob",
    Age: 76,
    Active: true
}
```

kdl-go's unmarshaler is described in detail in [Unmarshaling in kdl-go](docs/unmarshal.md).


## via Decoder

Use `kdl.NewDecoder()` to create a new KDL decoder whose options can be customized to your needs:

```go
type Person struct {
    Name        string      `kdl:"name"`
    Age         int         `kdl:"age"`
    Active      bool        `kdl:"active"`
}

data := `
    name "Bob"
    age 76
    active true
    geriatric true
`

var person Person
dec := kdl.NewDecoder(strings.NewReader(data))

// ignore the unhandled "geriatric" node 
dec.Options.AllowUnhandledNodes = true

if err := dec.Decode(&person); err == nil {
    fmt.Printf("%+v\n", person)
}
```
```go
// output
Person{
    Name: "Bob",
    Age: 76, 
    Active: true
}
```


# Marshaling

## via Marshal

`Marshal()` marshals a Go `map` or `struct` into KDL. The `kdl` tag can be used to map struct fields to KDL node names
or otherwise change marshaling behavior:

```go
type Person struct {
    Name   string `kdl:"name"`
    Age    int    `kdl:"age"`
    Active bool   `kdl:"active"`
}

person := Person{
    Name:   "Bob",
    Age:    32,
    Active: true,
}

if data, err := kdl.Marshal(person); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
name "Bob"
age 32
active true
```

kdl-go's marshaler is described in detail in [Marshaling in kdl-go](docs/marshal.md).


## via Encoder

Use `kdl.NewEncoder()` to create a new KDL encoder whose options can be customized to your needs:

```go
type Person struct {
    Name   string `kdl:"name"`
    Age    int    `kdl:"age"`
    Active bool   `kdl:"active"`
}

person := Person{
    Name:   "Bob Jones",
    Age:    32,
    Active: true,
}

enc := kdl.NewEncoder(os.Stdout)
if err := enc.Encode(person); err != nil {
	panic(err)
}
```
```kdl
//output
name "Bob Jones"
age 32
active true
```

# nginx-style Syntax Mode

kdl-go can also parse nginx-style configuration files using its `relaxed.NGINXSyntax` mode:

```go
data := `
    # web root
    location / {
        root /var/www/html;
    }

    # a missing location
    location /missing {
        return 404;
    }
`

type Location struct {
    Root   string `kdl:"root,omitempty,child"`
    Return int    `kdl:"return,omitempty,child"`
}
type NginxServer struct {
    Locations map[string]Location `kdl:"location,multiple"`
}

var ngx NginxServer
dec := kdl.NewDecoder(strings.NewReader(data))
dec.Options.RelaxedNonCompliant |= relaxed.NGINXSyntax

if err := dec.Decode(&ngx); err == nil {
    fmt.Printf("%#v\n", ngx)
}
```
```go
// output:
NginxServer{
    Locations: {
        "/": { Root:"/var/www/html", Return:0 }, 
        "/missing": { Root:"", Return:404 }
    }
}
```

See the [unmarshaling docs](docs/unmarshal.md) for further information.


# Verifying Spec Compliance

To download and test against all [Full Document Test Cases](https://github.com/kdl-org/kdl/tree/main/tests/test_cases) from the
[kdl.org repository](https://github.com/kdl-org/kdl), run:

```bash
git clone https://github.com/sblinch/kdl-go
cd kdl-go
git clone --branch release/v1 https://github.com/kdl-org/kdl kdl-org
cd internal/parser
go test -v -run TestKDLOrgTestCases -tags kdldeterministic
```
As of October 2023, kdl-go passes all of the available test cases.


# Development Status

kdl-go is actively maintained and is has been used as a configuration unmarshaler in a number of production applications
for several years now. It is considered stable at this time.

Issue reports and pull requests are welcome.


# On KDLv2

kdl-go implements the [KDLv1 specification](https://kdl.dev/spec-v1/) only. 

The [KDLv2 specification](https://kdl.dev/spec/) introduces some changes that this maintainer is not comfortable
implementing as explained in [issue #6](https://github.com/sblinch/kdl-go/issues/6#issuecomment-2561391289). Similar
concerns have been raised by other users as well (eg: [here](https://github.com/kdl-org/kdl/issues/537) and
[here](https://github.com/kdl-org/kdl/issues/512)). As such, support for KDLv2 is not planned at this time.


# License

kdl-go is released under the MIT license. See [LICENSE](LICENSE) for details.
