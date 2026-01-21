# Unmarshaling in kdl-go

Because KDL is such a flexible language that can be useful for such a wide variety of use-cases, kdl-go's unmarshaler
tries to be similarly flexible in the data structures you can use to represent your KDL documents.



## Basic Unmarshaling

Unmarshal() unmarshals KDL to a Go `map` or `struct`. The `kdl` tag can be used to map KDL node names to struct fields
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
// output:
Person{Name:"Bob", Age:76, Active:true}
``` 


## Multiple Arguments

Nodes with multiple arguments can be unmarshaled into slices:

```go
type Things struct {
    Vegetables   []string      `kdl:"vegetables"`
    Fruits       []interface{} `kdl:"fruits"`
    MagicNumbers []int         `kdl:"magic-numbers"`
}

data := `
    vegetables "broccoli" "carrot" "cucumber"
    fruits "apple" "orange" "watermelon"
    magic-numbers 4 8 16 32
`

var things Things
if err := kdl.Unmarshal(data, &things); err == nil {
    fmt.Printf("%+v\n",things)
}
```
```go
// output:
Things{
    Vegetables: []string{"broccoli", "carrot", "cucumber"},
    Fruits: []interface{}{"apple", "orange", "watermelon"},
    MagicNumbers: []int{4, 8, 16, 32}
}
```

## Properties

Nodes with properties can be unmarshaled into maps:

```go
type Things struct {
    Car       map[string]interface{} `kdl:"car"`
    Truck     map[string]string      `kdl:"truck"`
    Inventory map[string]int         `kdl:"inventory"`
}

data := `
    car make="ford" model="mustang" color="red" year=1967
    truck make="toyota" model="tacoma" color="black" year=2022
    inventory widgets=32 frobnobs=17
`

var things Things
if err := kdl.Unmarshal([]byte(data), &things); err == nil {
    fmt.Printf("%#v\n", things)
}
```
```go
// output:
Things{
    Car:map[string]interface{}{"color":"red", "make":"ford", "model":"mustang", "year":1967},
    Truck:map[string]string{"color":"black", "make":"toyota", "model":"tacoma", "year":"2022"},
    Inventory:map[string]int{"frobnobs":17, "widgets":32}
}
```

## Both Arguments and Properties

Nodes with both arguments and properties (or any combination thereof) can be unmarshaled into a variety of Go types;
kdl-go will always try to do something reasonable with each node.

When unmarshaling into a struct, kdl-go unmarshals a node's arguments into a slice struct field tagged `",args"`. It
unmarshals a node's properties into a map struct field tagged `",props"`.

```go
type Staff struct {
    CEO struct {
        Args  []interface{}          `kdl:",args"`
        Props map[string]interface{} `kdl:",props"`
    } `kdl:"ceo"`
}

var staff Staff
if err := kdl.Unmarshal([]byte(data), &staff); err == nil {
    fmt.Printf("%#v\n", staff)
}
```
```go
// output:
Staff{
    CEO: {
        Args:[]interface{}{"Bob", "Smith"},
        Props:{"age":76}}
    }
}
```

Alternately, arguments can be unmarshaled individually, in order, into struct fields of any type tagged with `",arg"`.
Similarly, properties can be unmarshaled individually into struct fields of any type tagged with the property's name:

```go
type Staff struct {
    CEO struct {
        First string `kdl:",arg"`
        Last  string `kdl:",arg"`
        Age   int    `kdl:"age"`
    } `kdl:"ceo"`
}

var staff Staff
if err := kdl.Unmarshal([]byte(data), &staff); err == nil {
    fmt.Printf("%#v\n", staff)
}
```
```go
// output:
Staff{
    CEO: {
        First:"Bob", 
        Last:"Smith", 
        Age:76
    }
}
```

A node can also be unmarshaled into a `map`. Arguments are keyed by their index in the argument list (0, 1, ...) and
properties are keyed by their property name:

```go
type Staff struct {
    CEO map[string]interface{} `kdl:"ceo"`
}

var staff Staff
if err := kdl.Unmarshal([]byte(data), &staff); err == nil {
    fmt.Printf("%#v\n", staff)
}
```
```go
// output:
map[string]interface{}{
    "0": "Bob",
    "1": "Smith",
    "age": 76
}
```

A node can be unmarshaled into a slice. Arguments are added to the slice first, in order, followed by properties.
If the slice is of type interface{}, each property is a []interface{}{"key", value}. If the slice is of type string,
each property is a string in the format "key=value".

```go
type Staff struct {
    CEO []interface{} `kdl:"ceo"`
}

var staff Staff
if err := kdl.Unmarshal([]byte(data), &staff); err == nil {
    fmt.Printf("%#v\n", staff)
}
```
```go
// output:
Staff{
    CEO: []interface{}{
        "Bob",
        "Smith", 
        []interface{}{"age", 76}
    }
}
```

## Child nodes

Children can also be unmarshaled into a variety of Go data types, such as structs:

```go
data := `
    bob {
        nationality "Canadian"
        language English=true French=false
    }
`

type Person struct {
    Nationality string          `kdl:"nationality"`
    Language    map[string]bool `kdl:"language"`
}
type People struct {
    Bob Person     `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob: Person{
        Nationality: "Canadian",
        Language: map[string]bool{"English":true, "French":false}
    }
}
```

Or maps:
```go
data := `
    bob {
        nationality "Canadian"
        language English=true French=false
    }
`

type People struct {
    Bob map[string]interface{} `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob: {
        "language": map[string]interface{}{
            "English": true,
            "French": false
        }, 
        "nationality": "Canadian"
    }
}
```

## Arguments, Properties, and Children combined

If a node has arguments, properties, and children, the above techniques can be combined to decode the node.

When decoding into a map, arguments are added to the map first, keyed by their argument index. Properties are added
next, keyed by the property name. Finally, children are added, keyed by their node name.

```go
data := `
    bob "Johnson" active=true {
        nationality "Canadian"
        language English=true French=false
    }
`

type People struct {
    Bob map[string]interface{} `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob: map[string]interface{}{
        "0": "Johnson",
        "active": true,
        "language": map[string]interface{}{
            "English": true,
            "French": false
        },
        "nationality": "Canadian"
    }
}
```

When decoding into a map, arguments are added to the map first (keyed by their index), properties are added next (keyed
by their names) and children are added last (keyed by their names):

```go
data := `
    bob "Johnson" active=true {
        nationality "Canadian"
        language English=true French=false
    }
`

type People struct {
    Bob map[string]interface{} `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob:map[string]interface{}{
        "0": "Johnson",
        "active": true,
        "language": map[string]interface{}{
            "English": true,
            "French": false
        },
        "nationality": "Canadian"
    }
}
```

When unmarshaling into a struct, the `",arg"` tag can be used to capture arguments on-by-one, and/or the `",args"` tag
can be used to capture all remaining arguments in a slice. The `",props"` tag can be used to capture properties (or the
properties can be tagged by name). And finally, children can be tagged either using the `",children"` tag (similar to
`",props"`, but for children) or tagged by name.

```go
data := `
    bob "Johnson" active=true {
        nationality "Canadian"
        language English=true French=false
    }
`

type Person struct {
    Args        []interface{}          `kdl:",args"`
    Props       map[string]interface{} `kdl:",props"`
    Nationality string                 `kdl:"nationality"`
    Language    map[string]bool        `kdl:"language"`
}

type People struct {
    Bob Person `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob: Person{
        Args: []interface{}{"Johnson"},
        Props: map[string]interface{}{
            "active": true
        },
        Nationality: "Canadian",
        Language: map[string]bool{
            "English": true,
            "French": false
        }
    }
}
```

Capturing child nodes using the `",children"` tag:

```go
data := `
    bob "Johnson" active=true {
        nationality "Canadian"
        language English=true French=false
    }
`

type Person struct {
    Args     []interface{}          `kdl:",args"`
    Props    map[string]interface{} `kdl:",props"`
    Children map[string]interface{} `kdl:",children"`
}
type People struct {
    Bob Person `kdl:"bob"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Bob: Person{
        Args: []interface{}{"Johnson"},
        Props: map[string]interface{}{"active":true},
        Children: map[string]interface{}{
            "language": map[string]interface{}{
                "English":true, "French":false
            },
            "nationality": "Canadian"
        }
    }
}
```
## Multiple nodes with the same name

It's not uncommon to need multiple instances of a node in a document; for example, you might want to parse multiple
`location` directives in an nginx-like configuration file.

Add the `",multiple"` tag to a struct field to use the node's first argument as the map key, for example:

```go
data := `
    location "/" {
        root "/var/www/html";
    }
    
    location "/missing" {
        return 404;
    }
`

type NginxServer struct {
    Locations map[string]interface{} `kdl:"location,multiple"`
}

var ngx NginxServer
if err := kdl.Unmarshal([]byte(data), &ngx); err == nil {
    fmt.Printf("%#v\n", ngx)
}
```
```go
// output:
NginxServer{
    Locations: map[string]interface{}{
        "/": map[string]interface{}{
            "root": "/var/www/html"
        },
        "/missing": map[string]interface{}{
            "return": 404
        }
    }
}
```

(Without the `",multiple"` tag, kdl-go would instead unmarshal the first `location` directive into the `Locations` map,
then immediately overwrite it with the second `location` directive.)

This effect can stack multiple times when nodes with multiple aguments are unmarshaled into multiply-nested maps:

```go
data := `
    city "Canada" "BC" "Vancouver" {
        latitude 49.24966
        longitude -123.11934
    }
    city "Canada" "BC" "Whistler" {
        latitude 50.11632
        longitude -122.95736 
    }
`

type LatLon struct {
    Latitude  float64 `kdl:"latitude"`
    Longitude float64 `kdl:"longitude"`
}
type Cities struct {
    City map[string]map[string]map[string]LatLon `kdl:"city,multiple"`
}

var cities Cities
if err := kdl.Unmarshal([]byte(data), &cities); err == nil {
    fmt.Printf("%#v\n", cities)
}
```
```go
// output:
Cities{
    City: {
        "Canada": {
            "BC": {
                "Vancouver": {Latitude:49.24966, Longitude:-123.11934},
                "Whistler": {Latitude:50.11632, Longitude:-122.95736}
            }
        }
    }
}
```

The `",multiple"` tag can also be used with slices:

```go
data := `
    person "Bob" active=true
    person "Jane" active=true
`
type People struct {
    Person []map[string]interface{} `kdl:"person,multiple"`
}

var people People
if err := kdl.Unmarshal([]byte(data), &people); err == nil {
    fmt.Printf("%#v\n", people)
}
```
```go
// output:
People{
    Person: []map[string]interface{}{
        map[string]interface{}{"0":"Bob", "active":true},
        map[string]interface{}{"0":"Jane", "active":true}
    }
}
```

## Selecting Struct Fields

Struct field tags can be used to control which KDL nodes are unmarshaled into which struct fields:

```go
type Info struct {
    Address   string                            // unmarshaled from a node named "address"
    Birthdate time.Time `kdl:"dob"`             // unmarshaled from node named "dob"
    Phone     string    `kdl:"phone,omitempty"` // unmarshaled from a node named "phone"
    Password  string    `kdl:"-"`               // never unmarshaled into
}
```

A field without a `kdl:"..."` tag name is unmarshaled from a node with the lowercase name of the Go struct field. A
field with a tag name of `-` is never unmarshaled into. The `,omitempty` tag is used only when marshaling and is ignored
during unmarshaling.


## The `format` Option

kdl-go implements the `format` tag option for `[]byte`, `time.Time`, `time.Duration`, `float32`, and `float64` values,
as described in the spec for Go's upcoming [encoding/json/v2](https://github.com/golang/go/discussions/63397)
implementation.


### time.Time formats

For `time.Time` fields, the options (per the `json/v2` spec) are defined as follows:

> The time.Time type accepts a "format" value which may either be a Go identifier for one of the format constants (e.g.,
> "RFC3339") or the format string itself to use with time.Time.Format or time.Parse (#21990). It can also be "unix", 
> "unixmilli", "unixmicro", or "unixnano" to be represented as a decimal number reporting the number of seconds (or
> milliseconds, etc.) since the Unix epoch.

If no `format` value is specified for a string value, RFC3339 is assumed.

In kdl-go, this looks like the following:

```go
data := `
    time-unix 1696805603
    time-rfc3339 "2023-10-08T15:54:13-07:00"
    time-rfc822z "08 Oct 23 15:54 -0700"
    time-date "2023-10-08"
`

type TimeFormats struct {
    TimeUnix    time.Time     `kdl:"time-unix,format:unix"`         // use time.Unix(..., 0)
    TimeRFC3339 time.Time     `kdl:"time-rfc3339,format:RFC3339"`   // use time.Parse(time.RFC3339, ...)
    TimeRFC822Z time.Time     `kdl:"time-rfc822z,format:RFC822Z"`   // use time.Parse(time.RFC822Z, ...)
    TimeDate    time.Time     `kdl:"time-date,format:'2006-01-02'"` // use time.Parse("2006-01-02", ...) 
}

var t TimeFormats
if err := kdl.Unmarshal([]byte(data), &t); err == nil {
    fmt.Printf("%#v\n", t)
}
```
```go
// output:
TimeFormats{
  TimeUnix:    time.Date(2023, time.October, 8, 15, 53, 23, 0, time.Local),
  TimeRFC3339: time.Date(2023, time.October, 8, 15, 54, 13, 0, time.Local),
  TimeRFC822Z: time.Date(2023, time.October, 8, 15, 54,  0, 0, time.Local),
  TimeDate:    time.Date(2023, time.October, 8,  0,  0,  0, 0, time.UTC),
}
```

### time.Duration Formats


A `time.Duration` field can be unmarshaled from a numeric value representing the number of seconds, or from a duration
string configured using the `format` option.

For `time.Duration` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> The time.Duration type accepts a "format" value of "sec", "milli", "micro", or "nano" to represent it as the number of
> seconds (or milliseconds, etc.) formatted as a ... number. ... If the format is "base60", it is encoded as a ...
> string using the "H:MM:SS.SSSSSSSSS" representation.

If no `format` value is specified for a string value, `time.Duration.String()` format (eg: `"1h32m7s"`) is assumed.

In kdl-go, this looks like the following:

```go
data := `
    duration "1h32m8s"
	hms "01:32:08.0"
    seconds 5528
`

type DurationFormats struct {
    Duration time.Duration `kdl:"duration"`
    HMS      time.Duration `kdl:"hms,format:base60"`
    Seconds  time.Duration `kdl:"seconds,format:sec"`
}

var d DurationFormats
if err := kdl.Unmarshal([]byte(data), &d); err == nil {
    fmt.Printf("%#v\n", d)
}
```
```go
// output:
DurationFormats{
   Duration: 5528000000000,
   HMS: 5528000000000,
   Seconds: 5528000000000
}

```
### []byte Formats

For `[]byte` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> []byte and [N]byte types accept "format" values of either "base64", "base64url", "base32", "base32hex", "base16", or
> "hex", where it represents the binary bytes as a [KDL] string encoded using the specified format in RFC 4648. It may
> also be "array" to treat the slice or array as a [KDL] array of numbers.

Additionally, kdl-go implements `format=string` which unmarshals a single string argument into the byte slice.

If no `format` value is specified for a string value, `format=base64` is assumed if a single string argument exists;
`format=array` is assumed if multiple arguments exist.

In kdl-go, this looks like the following:

```go
data := `
    bytes-b64 "aGVsbG8="
    bytes-b64url "dGVzdGluZw=="
    bytes-b32 "ORSXG5DJNZTQ===="
    bytes-b32hex "EHIN6T39DPJG===="
    bytes-hex "74657374696e67"
    bytes-array 84 69 83 84 73 78 71
    bytes-string "testing"
`

type ByteSliceFormats struct {
    Base64Bytes    []byte  `kdl:"bytes-b64,format:base64"`
    Base64URLBytes []byte  `kdl:"bytes-b64url,format:base64url"`
    Base32Bytes    []byte  `kdl:"bytes-b32,format:base32"`
    Base32HexBytes []byte  `kdl:"bytes-b32hex,format:base32hex"`
    HexBytes       []byte  `kdl:"bytes-hex,format:hex"` // same as `format:base16`
    Array          []byte  `kdl:"bytes-array,format:array"`
    StringBytes    []byte  `kdl:"bytes-string,format:string"`
}

var d ByteSliceFormats
if err := kdl.Unmarshal([]byte(data), &d); err == nil {
    fmt.Printf("%#v\n", d)
}
```
```go
// output:
ByteSliceFormats{
    Base64Bytes: []uint8{0x68, 0x65, 0x6c, 0x6c, 0x6f}, // "hello"
    Base64URLBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
    Base32Bytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
    Base32HexBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
    HexBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
    Array: []uint8{0x54, 0x45, 0x53, 0x54, 0x49, 0x4e, 0x47}, // "TESTING"
    BytesString: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
}
```

### float32/float64 Formats

For `float32` and `float64` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> float32 and float64 types accept a "format" value of "nonfinite", where NaN and infinity are represented as [KDL]
> strings.

If no `format` value is specified for a floating point field, `NaN`, `+Inf`, and `-Inf` values are unmarshaled as `0.0`.

In kdl-go, this looks like the following:

```go
data := `
    float64posinf "+Inf"
    float64neginf "-Inf"
    float64inf "+Inf"
    float64nan "NaN"
    float32nan "NaN"
    float64 "+Inf"
    float32 "NaN"
`

type FloatFormats struct {
    Float64PosInf  float64 `kdl:"float64posinf,format:nonfinite"`
    Float64NegInf  float64 `kdl:"float64neginf,format:nonfinite"`
    Float64Inf     float64 `kdl:"float64inf,format:nonfinite"`
    Float64NaN     float64 `kdl:"float64nan,format:nonfinite"`
    Float32NaN     float32 `kdl:"float32nan,format:nonfinite"`
    Float64        float64 `kdl:"float64"`
    Float32        float32 `kdl:"float32"`
}

var d FloatFormats
if err := kdl.Unmarshal([]byte(data), &d); err == nil {
    fmt.Printf("%#v\n", d)
}
```
```go
// output:
FloatFormats{
    Float64PosInf: +Inf,
    Float64NegInf: -Inf,
    Float64Inf: +Inf,
    Float64NaN: NaN,
    Float32NaN: NaN
    Float64: 0
    Float32: 0
}
```


## Custom unmarshaling

kdl-go supports three mechanisms for custom unmarshaling of KDL markup:

- the `encoding.TextUnmarshaler` interface, which many types already implement
- the `kdl.Unmarshaler` interface (and the `kdl.ValueUnmarshaler` interface)
- the `AddCustomUnmarshaler` function (and the `AddCustomValueUnmarshaler` function)

Each is documented below.


### Using encoding.TextUnmarshaler

`UnmarshalText` is used to unmarshal a single value into an argument or property value. When invoked during
unmarshaling, `UnmarshalText` is passed the best available string representation of the value to be unmarshaled; for
example:
 - for a string value: `[]byte("my string"),`
 - for a numeric value: `[]byte("1234")`
 - for a boolean value: `[]byte("true")`,
 - for a null value: `[]byte("null")`.

`UnmarshalText` cannot be used to unmarshal an entire node, and is ignored if implemented on a value into which a node
must be unmarshaled. (Use `UnmarshalKDL` to unmarshal an entire KDL node.)

In this example, `*PersonName` has an `UnmarshalText` method that converts the value to uppercase:

```go
type PersonName string
func (n *PersonName) UnmarshalText(b []byte) error {
    *n = PersonName(strings.ToUpper(string(b)))
    return nil
}

type Person struct {
    FirstName PersonName
    LastName  PersonName
}
type People struct {
    Father Person `kdl:"father"`
}

data := `
    father firstname="Bob" lastname="Johnson"
`

var p People
if err := kdl.Unmarshal([]byte(data), &p); err == nil {
    fmt.Printf("%#v\n", p)
}
```
```go
// output:
People{
    Father: Person{
        FirstName: "BOB",
        LastName: "JOHNSON"
    }
}
```


### Using kdl.Unmarshaler

`UnmarshalKDL` allows unmarshaling an entire node into a Go value, and provides direct access to the KDL document node
from which the KDL is being unmarshaled:

```go
data := `
    person "Bob" "Johnson" age=32 active=true
`

type Person struct {
	FirstName  string
	LastName   string
	CurrentAge int
	IsActive   bool
}

func (p *Person) UnmarshalKDL(node *document.Node) error {
    if len(node.Arguments) != 2 {
        return errors.New("exactly 2 arguments required")
    }
    t.FirstName = strings.ToUpper(node.Arguments[0].ValueString())
    t.LastName = strings.ToUpper(node.Arguments[1].ValueString())
    
    if age, ok := node.Properties["age"].ResolvedValue().(int64); ok {
        t.CurrentAge = int(age)
    } else {
        return errors.New("age must be an int")
    }
    
    if active, ok := node.Properties["active"].ResolvedValue().(bool); ok {
        t.IsActive = active
    } else {
        return errors.New("active must be a bool")
    }
    return nil
}

type People struct {
	Person *Person `kdl:"person"`
}

var p People
if err := kdl.Unmarshal([]byte(data), &p); err == nil {
    fmt.Printf("%#v\n", p)
}
```
```go
// output:
People{
	Person: Person{
		FirstName: "BOB",
		LastName: "JOHNSON",
		CurrentAge: 32,
		IsActive:true
	}
}
```

Note that `UnmarshalKDL` is only invoked when unmarshaling an entire node. If custom unmarshaling is required only for
individual values within the node (arguments and property values) use `UnmarshalText` or `UnmarshalKDLValue` instead.


### Using kdl.ValueUnmarshaler

`UnmarshalKDLValue` is used to unmarshal a single Go value into an argument or property value. When invoked during
unmarshaling, `UnmarshalKDLValue` is passed the `*document.Value` from which the value must be unmarshaled. This can
be preferable to `UnmarshalText` both because it is more efficient and it preserves the type information about the
source value.

`UnmarshalKDLValue` cannot be used to unmarshal an entire node. (Use `UnmarshalKDL` to unmarshal an entire KDL node.)

In this example, `*PersonName` has an `UnmarshalKDLValue` method that converts the value to uppercase:

```go
type PersonName string
func (n *PersonName) UnmarshalKDLValue(value *document.Value) error {
	*n = PersonName(strings.ToUpper(value.ValueString()))
    return nil
}

type Person struct {
    FirstName PersonName
    LastName  PersonName
}
type People struct {
    Father Person `kdl:"father"`
}

data := `
    father firstname="Bob" lastname="Johnson"
`

var p People
if err := kdl.Unmarshal([]byte(data), &p); err == nil {
    fmt.Printf("%#v\n", p)
}
```
```go
// output:
People{
    Father: Person{
        FirstName: "BOB",
        LastName: "JOHNSON"
    }
}
```


### Using kdl.AddCustomUnmarshaler

When unmarshaling types that are not under your control (eg: types from the Go standard library or third-party code) it
may be desirable to configure a custom unmarshaler for a type without modifying the type itself to implement an
UnmarshalKDL method.

For these cases, the `AddCustomUnmarshaler` function allows registering unmarshalers for arbitrary types.

All calls to `AddCustomUnmarshaler` must be made before any marshal/unmarshal operations are performed, otherwise it
will panic.

In this example, `Person` has an unmarshaler registered via AddCustomUnmarshaler that performs custom validation before
assigning values to the node.

```go
data := `
    person "Bob" "Johnson" age=32 active=true
`

type Person struct {
	FirstName  string
	LastName   string
	CurrentAge int
	IsActive   bool
}

kdl.AddCustomUnmarshaler[Person](func(node *document.Node, v reflect.Value) error {
    if len(node.Arguments) != 2 {
        return errors.New("exactly 2 arguments required")
    }

    t := v.Interface().(Person)
    t.FirstName = strings.ToUpper(node.Arguments[0].ValueString())
    t.LastName = strings.ToUpper(node.Arguments[1].ValueString())
    
    if age, ok := node.Properties["age"].ResolvedValue().(int64); ok {
        t.CurrentAge = int(age)
    } else {
        return errors.New("age must be an int")
    }
    
    if active, ok := node.Properties["active"].ResolvedValue().(bool); ok {
        t.IsActive = active
    } else {
        return errors.New("active must be a bool")
    }
    return nil	
})

type People struct {
	Person *Person `kdl:"person"`
}

var p People
if err := kdl.Unmarshal([]byte(data), &p); err == nil {
    fmt.Printf("%#v\n", p)
}
```
```go
// output:
People{
	Person: Person{
		FirstName: "BOB",
		LastName: "JOHNSON",
		CurrentAge: 32,
		IsActive:true
	}
}
```

Note that the custom unmarshaler is only invoked when unmarshaling an entire node. If custom unmarshaling is required
only for individual values within the node (arguments and property values) use `UnmarshalText` instead.


### Using kdl.AddCustomValueUnmarshaler

`AddCustomValueUnmarshaler` is used to unmarshal a single Go value into an argument or property value. When invoked
during unmarshaling, the unmarshaler is passed the `*document.Value` from which the value must be unmarshaled. 

`AddCustomValueUnmarshaler` cannot be used to unmarshal an entire node. (Use `AddCustomUnmarshaler` to unmarshal an
entire KDL node.)

All calls to `AddCustomValueUnmarshaler` must be made before any marshal/unmarshal operations are performed, otherwise
it will panic.

In this example, `PersonName` has an unmarshaler registered via `AddCustomValueUnmarshaler` that converts the value to
uppercase:

```go
type PersonName string

type Person struct {
    FirstName PersonName
    LastName  PersonName
}
type People struct {
    Father Person `kdl:"father"`
}

kdl.AddCustomValueUnmarshaler[PersonName](func(value *document.Value, v reflect.Value, format string) error {
    v.SetString(strings.ToUpper(value.ValueString()))
    return nil
})


data := `
    father firstname="Bob" lastname="Johnson"
`

var p People
if err := kdl.Unmarshal([]byte(data), &p); err == nil {
    fmt.Printf("%#v\n", p)
}
```
```go
// output:
People{
    Father: Person{
        FirstName: "BOB",
        LastName: "JOHNSON"
    }
}
```


## Breaking the standard

kdl-go also offers a set of relaxed modes that are not fully compliant with the KDL specification but allow for parsing
documents with looser syntax requirements.  Note that documents using these modes will **not** be readable by other
standard KDL parsers.

Relaxed modes can be enabled by creating a new `kdl.Decoder` and setting flag bits in `Options.RelaxedNonCompliant`
before calling the `Decode()` method.

Note that the functionality of relaxed modes may change as needed and without notice from version to version if new
features or changes are introduced to the KDL spec which conflict with relaxed mode features.

Also note that relaxed modes only affects decoding/unmarshaling; documents are always encoded/marshaled using KDL syntax.


### Multiplier Suffixes mode

Multiplier suffixes mode is enabled by setting `RelaxedNonCompliant |= relaxed.MultiplierSuffixes`, which makes the
following changes to allow a multiplier suffix at the end of a numeric value:

 - Numeric values may end in any of the following multiplier suffixes:
   - `k` or `K` for "kilo", to multiply the numeric value by 1e3, `kb` or `Kb` for "kibi", to multiply it by 2e10
   - `m` or `M` for "mega", to multiply the numeric value by 1e6, `mb` or `Mb` for "mibi", to multiply it by 2e20
   - `g` or `G` for "giga", to multiply the numeric value by 1e9, `gb` or `Gb` for "gibi", to multiply it by 2e30
   - `t` or `T` for "tera", to multiply the numeric value by 1e12, `tb` or `Tb` for "tibi", to multiply it by 2e40
   - `p` or `P` for "peta", to multiply the numeric value by 1e15, `pb` or `Pb` for "tibi", to multiply it by 2e50
 - Duration values such as `1h5m20s` are permitted as bare identifiers

```go
data := `
min-memory 1K
max-memory 2.3M
disk-usage "3.5M"
disk-quota "32M"
storage-used 32Mb
storage-avail 128.7Mb
not-a-number-bare 2.3k
not-a-number-quoted "3.2k"
`

type Measurements struct {
      MinMemory        int     `kdl:"min-memory"`
      MaxMemory        int     `kdl:"max-memory"`
      DiskUsage        int     `kdl:"disk-usage"`
      DiskQuota        int     `kdl:"disk-quota"`
      StorageUsed      int     `kdl:"storage-used"`
      StorageAvail     float64 `kdl:"storage-avail"`
      NotANumberBare   string  `kdl:"not-a-number-bare"`
      NotANumberQuoted string  `kdl:"not-a-number-quoted"`
}

dec := NewDecoder(strings.NewReader(data))
dec.Options.RelaxedNonCompliant |= relaxed.MultiplierSuffixes

var m Measurements
if err := dec.Decode(&m); err == nil {
        fmt.Printf("%#v\n", m)
}
```
```go
// output:
Measurements{
      MinMemory: 1000,           // 1 * 1000
      MaxMemory: 2300000,        // 2.3 * 1000000
      DiskUsage: 3500000,        // 3.5 * 1000000
      DiskQuota: 32000000,       // 32 * 1000000
      StorageUsed: 33554432,     // 32 * 1024 * 1024
      StorageAvail: 134951731.2, // 128.7 * 1024 * 1024
      NotANumberBare: "2.3k",    // bare value is used unchanged when unmarshaled into a string 
      NotANumberQuoted: "3.2k"   // quoted value is used unchanged when unmarshaled into a string
}
```


### NGINX syntax mode

NGINX syntax mode is enabled by setting `RelaxedNonCompliant |= relaxed.NGINXSyntax`, which makes the following changes
to allow parsing NGINX-style configuration syntax:

  - Arguments may be bare identifiers
  - Bare identifiers may contain additional characters (`()/.\:`) and start with additional characters (`()/.\_?`)
  - Quoted strings may be single-quoted (eg: `'foo bar'`)
  - Hash marks (`#`) are interpreted as the beginning of a line comment, identical to `//`
  - Type annotations and continuations are disallowed due to ambiguities introduced by allowing bare identifiers to
    start with `(` and `\`

Typically this is used in combination with `relaxed.MultiplierSuffixes` as NGINX-style configurations also allow
multiplier suffixes.

> Note: NGINX syntax mode was designed with the goal of parsing *NGINX-style* syntax -- which the author uses
> extensively as a configuration format in his own applications -- and such support was implemented with as few changes
> as possible to the KDL parser.
> 
> It may not accommodate every possible quirk of an *actual* NGINX configuration file -- in particular, certain regular
> expressions in `location` blocks can be tricky to parse as bare identifiers without fundamental changes to a KDL
> parser. kdl-go remains a KDL parser first, and fully supporting every NGINX corner case is a non-goal at this time.
> 
> Anecdotally, however, kdl-go does correctly parse all of the author's current development and production NGINX
> configuration files as well.


#### Example

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

type NginxServer struct {
    Locations map[string]interface{} `kdl:"location,multiple"`
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
    Locations:map[string]interface{}{
        "/": map[string]interface{}{
            "root": "/var/www/html"
        },
        "/missing": map[string]interface{}{
            "return": 404
        }
    }
}
```


### YAML/TOML assignment mode

YAML/TOML assignment mode is enabled by setting `RelaxedNonCompliant |= relaxed.YAMLTOMLAssignments`, which makes the
following changes to be more forgiving to common assignment mistakes from other markup languages:

  - A colon (`:`) is permitted after a node name, for users who think they are in YAML-land
  - An equal sign (`=`) is permitted after a node name, for users who think they are in TOML-land

Combined with NGINX syntax mode (above), this also allows a configuration syntax similar to a subset of
[UCL](https://github.com/vstakhov/libucl).


#### Example

```go
data := `
    name: "Bob"
    age = 32
`
type Person struct {
    Name string `kdl:"name"`
    Age  int    `kdl:"age"`
}

var p Person

dec := kdl.NewDecoder(strings.NewReader(data))
dec.Options.RelaxedNonCompliant |= relaxed.YAMLTOMLAssignments
dec.Options.AllowUnhandledNodes = true

if err := dec.Decode(&p); err == nil {
  fmt.Printf("%#v\n", p)
}
```
```go
Person{ Name: "Bob", Age: 32 }
```


### Preserving Comments

Limited support is available for preserving comments from an input document during unmarshaling, and restoring the
comments in the appropriate places in the output document during marshaling. This is performed on a best-effort basis
and there are numerous edge cases where this will not work -- and indeed, may be impossible -- based on the layout of
the marshaled structs and the nature of the modifications made between unmarshaling and remarshaling.

Comments are preserved by adding a special `interface{}` field tagged with `,structure` to the struct you are
unmarshaling, and then calling the unmarshaler with the `ParseComments` option:

```go
type Things struct {
    Vegetables   []string      `kdl:"vegetables"`
    Fruits       []interface{} `kdl:"fruits"`
    MagicNumbers []int         `kdl:"magic-numbers"`
    Structure    interface{}   `kdl:",structure"` // note the preceding comma
}

data := `
    // Vegetables are healthy
    vegetables "broccoli" "carrot" "cucumber"
    // Fruits are healthy 
    fruits "apple" "orange" "watermelon"
    // Magic numbers may be carcinogenic 
    magic-numbers 4 8 16 32
`

var things Things
var opts = kdl.UnmarshalOptions{ParseComments: true}
if err = kdl.UnmarshalWithOptions(data, &things, opts); err != nil {
	...
}

things.Vegetables[0] = "cabbage"
things.MagicNumbers[0] = 99

if output, err := kdl.Marshal(things); err == nil {
    fmt.Println(string(output))
}	
```
```kdl
// output:

// Vegetables are healthy
vegetables "cabbage" "carrot" "cucumber"
// Fruits are healthy 
fruits "apple" "orange" "watermelon"
// Magic numbers may be carcinogenic 
magic-numbers 99 8 16 32
```

If a struct contains nested struct fields, each child struct must have its own `interface{}` field tagged with
`,structure` as well, to preserve the respective struct's comments:

```go
type Baz struct {
    Name      string      `kdl:"name"`
    Structure interface{} `kdl:",structure"`
}
type Bar struct {
    Name      string      `kdl:"name"`
    Structure interface{} `kdl:",structure"`
}
type Foo struct {
    Bar       Bar         `kdl:"bar"`
    Bazzes    []Baz       `kdl:"bazzes,multiple"`
    Structure interface{} `kdl:",structure"`
}
```

Comments may be dropped if:

- a value in a slice or map of structs has been replaced
- a node's name, arguments, or properties have changed, or the proper location for the comment in the output document is
  otherwise ambiguous
- the comment does not appear between two complete node declarations, or at the beginning or end of the document or a
  block of child nodes; the KDL spec allows for comments in some unusual places (such as between a node name and
  argument) and no effort is made to preserve such comments
- there is no enclosing struct in which to store the comment data, such as in a map of maps, map of slices, slice of
  maps, etc.
- various other corner cases are encountered; again, this feature is implemented on a best-effort basis only

