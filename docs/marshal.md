# Marshaling in kdl-go

Because KDL is such a flexible language that can be useful for such a wide variety of use-cases, kdl-go's marshaler
tries to be similarly flexible in the data structures you can use to represent your KDL documents.


## Basic Marshaling

Marshal() marshals a Go `map` or `struct` to KDL. The `kdl` tag can be used to map struct fields to KDL node names
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
## Multiple Arguments

Slice struct fields can be marshaled into node arguments:

```go
type Things struct {
    Vegetables   []string      `kdl:"vegetables"`
    Fruits       []interface{} `kdl:"fruits"`
    MagicNumbers []int         `kdl:"magic-numbers"`
}

things := Things{
    Vegetables:   []string{"broccoli", "carrot", "cucumber"},
    Fruits:       []interface{}{"apple", "orange", "watermelon"},
    MagicNumbers: []int{4, 8, 16, 32},
}

if data, err := kdl.Marshal(things); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
vegetables "broccoli" "carrot" "cucumber"
fruits "apple" "orange" "watermelon"
magic-numbers 4 8 16 32
```

## Properties

Map struct fields can be marshaled into node properties:

```go
type Things struct {
    Car       map[string]interface{} `kdl:"car"`
    Truck     map[string]string      `kdl:"truck"`
    Inventory map[string]int         `kdl:"inventory"`
}

things := Things{
    Car:       map[string]interface{}{"color": "red", "make": "ford", "model": "mustang", "year": 1967},
    Truck:     map[string]string{"color": "black", "make": "toyota", "model": "tacoma", "year": "2022"},
    Inventory: map[string]int{"frobnobs": 17, "widgets": 32},
}

if data, err := kdl.Marshal(things); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
car make="ford" model="mustang" year=1967 color="red"
truck color="black" make="toyota" model="tacoma" year="2022"
inventory frobnobs=17 widgets=32
```

## Both Arguments and Properties

A struct with special tags can specify custom marshaling preferences.

When marshaling a node from a struct, a slice field tagged `",args"` can be used for the node's arguments. A map field
tagged `",props"` can be used for the node's properties.

```go
type Person struct {
    Args  []interface{}          `kdl:",args"`
    Props map[string]interface{} `kdl:",props"`
}
type Staff struct {
    CEO Person `kdl:"ceo"`
}

staff := Staff{
    CEO: Person{
        Args:  []interface{}{"Bob", "Smith"},
        Props: map[string]interface{}{"age": 76},
    },
}

if data, err := kdl.Marshal(staff); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
ceo "Bob" "Smith" age=76
```

Alternately, arguments can be marshaled individually, in order, from struct fields of any type tagged with `",arg"`.
Similarly, properties can be marshaled individually from struct fields of any type tagged with the property's name:

```go
type Person struct {
    First string `kdl:",arg"`
    Last  string `kdl:",arg"`
    Age   int    `kdl:"age"`
}
type Staff struct {
    CEO Person `kdl:"ceo"`
}

staff := Staff{
    CEO: Person{
        First: "Bob",
        Last:  "Smith",
        Age:   76,
    },
}

if data, err := kdl.Marshal(staff); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
ceo "Bob" "Smith" age=76
```

A `map` can also be marshaled into a node. Numeric map keys (0, 1, ... or "0", "1", ...) are marshaled as arguments;
other keys are marshaled as properties:

```go
type Staff struct {
    CEO map[string]interface{} `kdl:"ceo"`
}

staff := Staff{
    CEO: map[string]interface{}{
        "0":   "Bob",
        "1":   "Smith",
        "age": 76,
    },
}
if data, err := kdl.Marshal(staff); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
ceo "Bob" "Smith" age=76
```

A slice can also be marshaled into a node.

If the slice is of type `[]interface{}`, elements of type `[]interface{}{"key", value}` are treated as properties, and
all other elements are treated as arguments.

If the slice is of type `[]string`, elements containing an equal sign (`=`) are treated as properties by splitting them
at the equal sign (eg: in `name=John`, `name` becomes the key and `John` becomes the value), and all other elements are
treated as arguments.

```go
type Staff struct {
    CEO []interface{} `kdl:"ceo"`
}

staff := Staff{
    CEO: []interface{}{
        "Bob",
        "Smith",
        []interface{}{"age", 76},
    },
}

if data, err := kdl.Marshal(staff); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
ceo "Bob" "Smith" age=76
```

## Child nodes

kdl-go will automatically marshal struct fields into child nodes if they cannot be represented as properties:

```go
type Person struct {
    Nationality string          `kdl:"nationality"`
    Language    map[string]bool `kdl:"language"`
}
type People struct {
    Bob Person `kdl:"bob"`
}

people := People{
    Bob: Person{
        Nationality: "Canadian",
        Language:    map[string]bool{"English": true, "French": false},
    },
}

if data, err := kdl.Marshal(people); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
bob nationality="Canadian" {
    language French=false English=true
}
```

But it's also possible to explicitly tag struct fields with `,child` to force them into child nodes:

```go
type Person struct {
    Nationality string          `kdl:"nationality,child"`
    Language    map[string]bool `kdl:"language"`
}
type People struct {
    Bob Person `kdl:"bob"`
}

people := People{
    Bob: Person{
        Nationality: "Canadian",
        Language:    map[string]bool{"English": true, "French": false},
    },
}

if data, err := kdl.Marshal(people); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
bob {
    nationality "Canadian"
    language French=false English=true
}
```

Child nodes can also be sourced from a map struct field using the `",children"` tag, similar to the `",args"` tag for
arguments:

```go
type Person struct {
    Args     []interface{}          `kdl:",args"`
    Props    map[string]interface{} `kdl:",props"`
    Children map[string]interface{} `kdl:",children"`
}
type People struct {
    Bob Person `kdl:"bob"`
}

people := People{
    Bob: Person{
        Args:  []interface{}{"Johnson"},
        Props: map[string]interface{}{"active": true},
        Children: map[string]interface{}{
            "language": map[string]interface{}{
                "English": true, "French": false,
            },
            "nationality": "Canadian",
        },
    },
}

if data, err := kdl.Marshal(people); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
bob "Johnson" active=true {
    language English=true French=false
    nationality "Canadian"
}		
```


Map values are also marshaled into child nodes if they cannot be represented as properties:

```go
type People struct {
    Bob map[string]interface{} `kdl:"bob"`
}
people := People{
    Bob: map[string]interface{}{
        "language": map[string]interface{}{
            "English": true,
            "French":  false,
        },
        "nationality": "Canadian",
    },
}

if data, err := kdl.Marshal(people); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
bob nationality="Canadian" {
    language English=true French=false
}		
```


## Multiple nodes with the same name

In some cases you may want to generate independent nodes for each map key, rather than a single node with properties or
children.

Add the `",multiple"` tag to a map struct field to emit separate nodes for each map key:

```go
type Location struct {
    Root   string `kdl:"root,omitempty,child"`
    Return int    `kdl:"return,omitempty,child"`
}
type NginxServer struct {
    Locations map[string]Location `kdl:"location,multiple"`
}

ngx := NginxServer{
    Locations: map[string]Location{
        "/": {
            Root: "/var/www/html",
        },
        "/missing": {
              Return: 404,
        },
    },
}

if data, err := kdl.Marshal(ngx); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
location "/" {
    root "/var/www/html"
}
location "/missing" {
    return 404
}		
```

By contrast, without the `",multiple"` tag, kdl-go would generate the following KDL instead:

```kdl
location {
    "/" {
        root "/var/www/html"
    }
    "/missing" {
        return 404
    }
}
```

This effect can stack multiple times when a struct field is a multiply-nested map: 

```go
type LatLon struct {
    Latitude  float64 `kdl:"latitude,child"`
    Longitude float64 `kdl:"longitude,child"`
}
type Cities struct {
    City map[string]map[string]map[string]LatLon `kdl:"city,multiple"`
}

cities := Cities{
    City: map[string]map[string]map[string]LatLon{
        "Canada": {
            "BC": {
                "Vancouver": {Latitude: 49.24966, Longitude: -123.11934},
                "Whistler":  {Latitude: 50.11632, Longitude: -122.95736},
            },
        },
    },
}

if data, err := kdl.Marshal(cities); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
city "Canada" "BC" "Vancouver" {
    latitude 49.24966
    longitude -123.11934
}
city "Canada" "BC" "Whistler" {
    latitude 50.11632
    longitude -122.95736
}	
```


The `",multiple"` tag can also be used with slices:

```go
type People struct {
    Person []map[string]interface{} `kdl:"person,multiple"`
}

people := People{
    Person: []map[string]interface{}{
        {"0": "Bob", "active": true},
        {"0": "Jane", "active": true},
    },
}

if data, err := kdl.Marshal(people); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
person "Bob" active=true
person "Jane" active=true		
```


## Selecting Struct Fields

Struct field tags can be used to control which fields are marshaled into KDL nodes and which names are used by the
nodes:

```go
type Info struct {
    Address   string                            // marshaled to a node named "address"
    Birthdate time.Time `kdl:"dob"`             // marshaled to a node named "dob"
    Phone     string    `kdl:"phone,omitempty"` // omitted if empty, otherwise marshaled to a node named "phone"
    Password  string    `kdl:"-"`               // not included in the marshaled output
}
```

A field without a `kdl:"..."` tag name is marshaled using the lowercase name of the Go struct field. A field with a tag
name of `-` is never marshaled. A field tagged `,omitempty` is omitted when its value is equal to the zero value for its 
type.


## The `format` Option 

kdl-go implements the `format` tag option for `[]byte`, `time.Time`, `time.Duration`, `float32`, and `float64` values,
as described in the spec for Go's upcoming [encoding/json/v2](https://github.com/golang/go/discussions/63397)
implementation.


### time.Time Formats

For `time.Time` fields, the options (per the `json/v2` spec) are defined as follows:

> The time.Time type accepts a "format" value which may either be a Go identifier for one of the format constants (e.g.,
> "RFC3339") or the format string itself to use with time.Time.Format or time.Parse. It can also be "unix", "unixmilli",
> "unixmicro", or "unixnano" to be represented as a decimal number reporting the number of seconds (or milliseconds,
> etc.) since the Unix epoch.

If no `format` value is specified, RFC3339 is assumed.

In kdl-go, this looks like the following:

```go
type TimeFormats struct {
    TimeUnix    time.Time `kdl:"time-unix,format:unix"`         // use time.Unix(..., 0)
    TimeRFC3339 time.Time `kdl:"time-rfc3339,format:RFC3339"`   // use time.Parse(time.RFC3339, ...)
    TimeRFC822Z time.Time `kdl:"time-rfc822z,format:RFC822Z"`   // use time.Parse(time.RFC822Z, ...)
    TimeDate    time.Time `kdl:"time-date,format:'2006-01-02'"` // use time.Parse("2006-01-02", ...)
}

tf := TimeFormats{
    TimeUnix:    time.Date(2023, time.October, 8, 15, 53, 23, 0, time.Local),
    TimeRFC3339: time.Date(2023, time.October, 8, 15, 54, 13, 0, time.Local),
    TimeRFC822Z: time.Date(2023, time.October, 8, 15, 54, 0, 0, time.Local),
    TimeDate:    time.Date(2023, time.October, 8, 0, 0, 0, 0, time.UTC),
}

if data, err := kdl.Marshal(tf); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
time-unix 1696805603
time-rfc3339 "2023-10-08T15:54:13-07:00"
time-rfc822z "08 Oct 23 15:54 -0700"
time-date "2023-10-08"
```


### time.Duration Formats

For `time.Duration` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> The time.Duration type accepts a "format" value of "sec", "milli", "micro", or "nano" to represent it as the number of
> seconds (or milliseconds, etc.) formatted as a [KDL] number. ... If the format is "base60", it is encoded as a [KDL]
> string using the "H:MM:SS.SSSSSSSSS" representation.

If no `format` value is specified for a string value, `time.Duration.String()` format (eg: `"1h32m7s"`) is assumed.

In kdl-go, this looks like the following:

```go
type DurationFormats struct {
    Duration time.Duration `kdl:"duration"`
    HMS      time.Duration `kdl:"hms,format:base60"`
    Seconds  time.Duration `kdl:"seconds,format:sec"`
}

refTime := 2*time.Hour + 32*time.Minute + 7*time.Second
df := DurationFormats{
    Duration: refTime,
    HMS:      refTime,
    Seconds:  refTime,
}

if data, err := kdl.Marshal(df); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
duration "2h32m7s"
hms "02:32:07.0"
seconds 9127		
```


### []byte Formats

For `[]byte` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> []byte and [N]byte types accept "format" values of either "base64", "base64url", "base32", "base32hex", "base16", or
> "hex", where it represents the binary bytes as a [KDL] string encoded using the specified format in RFC 4648. It may
> also be "array" to treat the slice or array as a [KDL] array of numbers.

Additionally, kdl-go implements `format=string` which marshals a byte slice into a single string argument.

If no `format` value is specified, `format=base64` is assumed when marshaling.

In kdl-go, this looks like the following:

```go
type ByteSliceFormats struct {
    Base64Bytes    []byte `kdl:"bytes-b64,format:base64"`
    Base64URLBytes []byte `kdl:"bytes-b64url,format:base64url"`
    Base32Bytes    []byte `kdl:"bytes-b32,format:base32"`
    Base32HexBytes []byte `kdl:"bytes-b32hex,format:base32hex"`
    HexBytes       []byte `kdl:"bytes-hex,format:hex"` // same as `format:base16`
    Array          []byte `kdl:"bytes-array,format:array"`
    StringBytes    []byte `kdl:"bytes-string,format:string"`
}

f := ByteSliceFormats{
    Base64Bytes:    []byte("hello"),
    Base64URLBytes: []byte("testing"),
    Base32Bytes:    []byte("testing"),
    Base32HexBytes: []byte("testing"),
    HexBytes:       []byte("testing"),
    Array:          []byte("TESTING"),
    StringBytes:    []byte("this is a test"),
}

if data, err := kdl.Marshal(f); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
bytes-b64 "aGVsbG8="
bytes-b64url "dGVzdGluZw=="
bytes-b32 "ORSXG5DJNZTQ===="
bytes-b32hex "EHIN6T39DPJG===="
bytes-hex "74657374696e67"
bytes-array 84 69 83 84 73 78 71
bytes-string "this is a test"		
```


### float32/float64 Formats

For `float32` and `float64` fields, the `format` options (per the `json/v2` spec) are defined as follows:

> float32 and float64 types accept a "format" value of "nonfinite", where NaN and infinity are represented as [KDL]
> strings.

If no `format` value is specified for a floating point field, `NaN`, `+Inf`, and `-Inf` values are unmarshaled as `0.0`.

In kdl-go, this looks like the following:

```go
type ByteSliceFormats struct {
    Float64PosInf float64 `kdl:"float64posinf,format:nonfinite"`
    Float64NegInf float64 `kdl:"float64neginf,format:nonfinite"`
    Float64Inf    float64 `kdl:"float64inf,format:nonfinite"`
    Float64NaN    float64 `kdl:"float64nan,format:nonfinite"`
    Float32NaN    float32 `kdl:"float32nan,format:nonfinite"`
    Float64       float64 `kdl:"float64"`
    Float32       float32 `kdl:"float32"`
}

f := ByteSliceFormats{
    Float64PosInf: math.Inf(1),
    Float64NegInf: math.Inf(-1),
    Float64Inf:    math.Inf(1),
    Float64NaN:    math.NaN(),
    Float32NaN:    float32(math.NaN()),
    Float64:       math.Inf(1),
    Float32:       float32(math.Inf(1)),
}

if data, err := kdl.Marshal(f); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
float64posinf "+Inf"
float64neginf "-Inf"
float64inf "+Inf"
float64nan "NaN"
float32nan "NaN"
float64 0.0
float32 0.0		
```


## Custom marshaling

kdl-go supports both the `encoding.TextMarshaler` interface and its own `kdl.Marshaler` interface for custom
marshaling of KDL markup.


### Using encoding.TextMarshaler

`MarshalText` is used to marshal a single value (from an argument or property value) into its string representation.

`MarshalText` cannot be used to marshal an entire node, and is ignored if implemented on a value from which a node must
be marshaled. (Use `MarshalKDL` to marshal an entire KDL node.)


In this example, `*PersonName` has a `MarshalText` method that converts the value to lowercase:

```go
type PersonName string
func (n *PersonName) MarshalText() ([]byte, error) {
    return []byte(strings.ToLower(string(*n))), nil
}

type Person struct {
    FirstName PersonName
    LastName  PersonName
}
type People struct {
    Father Person `kdl:"father"`
}

p := People{
    Father: Person{
        FirstName: "Bob",
        LastName:  "Johnson",
    },
}

if data, err := kdl.Marshal(p); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
father firstname="bob" lastname="johnson"
```


### Using kdl.Marshaler

`MarshalKDL` allows marshaling a Go value directly into a KDL document node:

```go
type Relative struct {
    FirstName  string
    LastName   string
    CurrentAge int
    IsParent   bool
}

func (t *Relative) MarshalKDL(node *document.Node) error {
    node.AddArgument(t.FirstName, "")
    node.AddArgument(t.LastName, "")
    node.AddProperty("age", t.CurrentAge, "")
    node.AddProperty("parent", t.IsParent, "")
    return nil
}

type Family struct {
    Father Relative `kdl:"father"`
}

p := Family{
    Father: Relative{
        FirstName:  "Bob",
        LastName:   "Johnson",
        CurrentAge: 32,
        IsParent:   true,
    },
}

if data, err := kdl.Marshal(p); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
father "Bob" "Johnson" age=32 parent=true
```

Note that `MarshalKDL` is only invoked when marshaling an entire node. If custom marshaling is required only for
individual values within the node (such as arguments, property values, etc.) use `MarshalKDLValue` instead.


### Using kdl.ValueMarshaler

`MarshalKDLValue` is used to marshal a single Go value into a `*document.Value` (for an argument or property).

`MarshalKDLValue` cannot be used to marshal an entire node, and is ignored if implemented on a value from which a node
must be marshaled. (Use `MarshalKDL` to marshal an entire KDL node.)

In this example, `*PersonName` has a `MarshalKDLValue` method that converts the value to lowercase:

```go
type PersonName string
func (n *PersonName) MarshalKDLValue(value *document.Value) error {
	value.Value = strings.ToLower(string(*n))
	return nil
}

type Person struct {
    FirstName PersonName
    LastName  PersonName
}
type People struct {
    Father Person `kdl:"father"`
}

p := People{
    Father: Person{
        FirstName: "Bob",
        LastName:  "Johnson",
    },
}

if data, err := kdl.Marshal(p); err == nil {
    fmt.Println(string(data))
}
```
```kdl
// output:
father firstname="bob" lastname="johnson"
```
