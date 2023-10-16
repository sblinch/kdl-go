package kdl

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/relaxed"
)

func TestParse(t *testing.T) {
	data := `
name "Bob"
age 76
active true
`

	doc, err := Parse(strings.NewReader(data))
	if err != nil {
		// ...handle error
	} else {
		for _, node := range doc.Nodes {
			println(node.Name.String())
		}
	}
}

func TestGenerate(t *testing.T) {
	data := `
name "Bob"
age 76
active true
`

	doc, err := Parse(strings.NewReader(data))
	if err != nil {
		// ...handle error
	} else if err := Generate(doc, os.Stderr); err != nil {
		println(err.Error())
	}
}

func TestExampleUnmarshal1(t *testing.T) {
	type Person struct {
		Name   string `kdl:"name"`
		Age    int    `kdl:"age"`
		Active bool   `kdl:"active"`
	}

	data := `
    name "Bob"
    age 76
    active true
`

	var person Person
	if err := Unmarshal([]byte(data), &person); err == nil {
		fmt.Printf("%#v\n", person)
	}
}

func TestExampleUnmarshal2(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &things); err == nil {
		fmt.Printf("%#v\n", things)
	}
}

func TestExampleUnmarshal3(t *testing.T) {
	type Things struct {
		Car       map[string]interface{} `kdl:"car"`
		Truck     map[string]string      `kdl:"truck"`
		Inventory map[string]int         `kdl:"inventory"`
	}

	data := `
	car make=ford model=mustang color=red year=1967
	truck make=toyota model=tacoma color=black year=2022
	inventory widgets=32 frobnobs=17
`

	var things Things
	if err := Unmarshal([]byte(data), &things); err == nil {
		fmt.Printf("%#v\n", things)
	}
}

func TestExampleUnmarshal4(t *testing.T) {
	data := `
ceo "Bob" "Smith" age=76
`

	func() {
		type Staff struct {
			CEO struct {
				Args  []interface{}          `kdl:",args"`  // []{"Bob","Smith"}
				Props map[string]interface{} `kdl:",props"` // {"age":76}
			} `kdl:"ceo"`
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO map[string]interface{} `kdl:"ceo"` // {"1":"Tim", "2":"Jones", "age":21}
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO []interface{} `kdl:"ceo"` // []{"Ethel","Smith",[]interface{}{"age",72}}
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO []string `kdl:"ceo"` // []{"Sue","Jones","age=22"}
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO struct {
				Args []interface{} `kdl:",args"` // []{"Carl","Smith"}
				Age  int           `kdl:"age"`   // 42
			} `kdl:"ceo"`
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO struct {
				First string `kdl:",arg"` // Anna
				Last  string `kdl:",arg"` // Jones
				Age   int    `kdl:"age"`  // 32
			} `kdl:"ceo"`
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()
	func() {
		type Staff struct {
			CEO *struct {
				Args []interface{} `kdl:",args"` // []{"Stu","Jones"}
				Age  int           `kdl:"age"`   // 52
			} `kdl:"ceo"`
		}
		var staff Staff
		if err := Unmarshal([]byte(data), &staff); err == nil {
			fmt.Printf("%#v\n", staff)
		}
	}()

}

func TestExampleUnmarshal5(t *testing.T) {
	data := `
    bob {
        nationality "Canadian"
        language English=true French=false
    }
    
    klaus {
        nationality "German"
        language English=false German=true
    }
`

	type Person struct {
		Nationality string          `kdl:"nationality"`
		Language    map[string]bool `kdl:"language"`
	}
	type People struct {
		Bob   map[string]interface{} `kdl:"bob"`
		Klaus Person                 `kdl:"klaus"`
	}

	var people People
	if err := Unmarshal([]byte(data), &people); err == nil {
		fmt.Printf("%#v\n", people)
	}
}

func TestExampleUnmarshal6(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &people); err == nil {
		fmt.Printf("%#v\n", people)
	}
}

func TestExampleUnmarshal7(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &people); err == nil {
		fmt.Printf("%#v\n", people)
	}
}

func TestExampleUnmarshal8(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &people); err == nil {
		fmt.Printf("%#v\n", people)
	}

}

func TestExampleUnmarshal9(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &ngx); err == nil {
		fmt.Printf("%#v\n", ngx)
	} else {
		println(err.Error())
	}
}

func TestExampleUnmarshal10(t *testing.T) {
	data := `
person "Bob" active=true
person "Jane" active=true
`
	type People struct {
		Person []map[string]interface{} `kdl:"person,multiple"`
	}

	var people People
	if err := Unmarshal([]byte(data), &people); err == nil {
		fmt.Printf("%#v\n", people)
	} else {
		println(err.Error())
	}
}

func TestExampleUnmarshal11(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &cities); err == nil {
		fmt.Printf("%#v\n", cities)
	} else {
		println(err.Error())
	}
}

func TestExampleUnmarshal12(t *testing.T) {
	type Person struct {
		Name   string `kdl:"name"`
		Age    int    `kdl:"age"`
		Active bool   `kdl:"active"`
	}

	data := `
    name "Bob"
    age 76
    active true
`

	var person Person
	dec := NewDecoder(strings.NewReader(data))
	dec.Options.AllowUnhandledArgs = true
	if err := dec.Decode(&person); err == nil {
		fmt.Printf("%+v\n", person)
		// Person{Name:"Bob", Age:76, Active:true}
	}
}

func TestExampleUnmarshal13(t *testing.T) {
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
	dec := NewDecoder(strings.NewReader(data))
	dec.Options.RelaxedNonCompliant = relaxed.NGINXSyntax | relaxed.YAMLTOMLAssignments
	if err := dec.Decode(&ngx); err == nil {
		fmt.Printf("%#v\n", ngx)
	} else {
		println(err.Error())
	}
}

func TestExampleUnmarshal14(t *testing.T) {
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
	if err := Unmarshal([]byte(data), &d); err == nil {
		fmt.Printf("%#v\n", d)
		// kdl.DurationFormats{Duration:5528000000000, HMS:5528000000000, Seconds:5528000000000}
	}
}

func TestExampleUnmarshal15(t *testing.T) {
	data := `
bytes-b64 "aGVsbG8="
bytes-b64url "dGVzdGluZw=="
bytes-b32 "ORSXG5DJNZTQ===="
bytes-b32hex "EHIN6T39DPJG===="
bytes-b16 "74657374696e67"
bytes-hex "74657374696e67"
bytes-array 84 69 83 84 73 78 71
float64posinf "+Inf"
float64neginf "-Inf"
float64inf "+Inf"
float64nan "NaN"
float32nan "NaN"
`

	type ByteSliceFormats struct {
		Base64Bytes    []byte  `kdl:"bytes-b64,format:base64"`
		Base64URLBytes []byte  `kdl:"bytes-b64url,format:base64url"`
		Base32Bytes    []byte  `kdl:"bytes-b32,format:base32"`
		Base32HexBytes []byte  `kdl:"bytes-b32hex,format:base32hex"`
		Base16Bytes    []byte  `kdl:"bytes-b16,format:base16"`
		HexBytes       []byte  `kdl:"bytes-hex,format:hex"`
		Array          []byte  `kdl:"bytes-array,format:array"`
		Float64PosInf  float64 `kdl:"float64posinf,format:nonfinite"`
		Float64NegInf  float64 `kdl:"float64neginf,format:nonfinite"`
		Float64Inf     float64 `kdl:"float64inf,format:nonfinite"`
		Float64NaN     float64 `kdl:"float64nan,format:nonfinite"`
		Float32NaN     float32 `kdl:"float32nan,format:nonfinite"`
	}

	var d ByteSliceFormats
	if err := Unmarshal([]byte(data), &d); err == nil {
		fmt.Printf("%#v\n", d)
		// ByteSliceFormats{
		// 	Base64Bytes: []uint8{0x68, 0x65, 0x6c, 0x6c, 0x6f}, // "hello"
		// 	Base64URLBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
		// 	Base32Bytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
		// 	Base32HexBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
		// 	Base16Bytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
		// 	HexBytes: []uint8{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67}, // "testing"
		// 	Array: []uint8{0x54, 0x45, 0x53, 0x54, 0x49, 0x4e, 0x47}, // "TESTING"
		// 	Float64PosInf:+Inf,
		// 	Float64NegInf:-Inf,
		// 	Float64Inf:+Inf,
		// 	Float64NaN:NaN,
		// 	Float32NaN:NaN
		// }
	}
}

type PersonName string

func (n *PersonName) UnmarshalText(b []byte) error {
	*n = PersonName(strings.ToUpper(string(b)))
	return nil
}

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

func TestExampleUnmarshal16(t *testing.T) {

	data := `
father firstname="Bob" lastname="Johnson"
`

	var p People
	if err := Unmarshal([]byte(data), &p); err == nil {
		fmt.Printf("%#v\n", p)
		// People{
		// 	Father: Person{
		// 		FirstName:"BOB",
		// 		LastName:"JOHNSON"
		// 	}
		// }
	}
}

func TestExampleUnmarshal17(t *testing.T) {
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
	dec := NewDecoder(strings.NewReader(data))
	dec.Options.RelaxedNonCompliant |= relaxed.NGINXSyntax

	if err := dec.Decode(&ngx); err == nil {
		fmt.Printf("%#v\n", ngx)
	}

}

func TestExampleMarshal1(t *testing.T) {
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

	if data, err := Marshal(person); err == nil {
		fmt.Println(string(data))
		// name "Bob Jones"
		// age 32
		// active true
	}
}

func TestExampleMarshal2(t *testing.T) {
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

	enc := NewEncoder(os.Stdout)
	_ = enc.Encode(person)
	// name "Bob Jones"
	// age 32
	// active true
}

func TestExampleMarshal3(t *testing.T) {
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

	if data, err := Marshal(things); err == nil {
		fmt.Println(string(data))
		// vegetables "broccoli" "carrot" "cucumber"
		// fruits "apple" "orange" "watermelon"
		// magic-numbers 4 8 16 32
	}
}

func TestExampleMarshal4(t *testing.T) {
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

	if data, err := Marshal(things); err == nil {
		fmt.Println(string(data))
		// car make=ford model=mustang year=1967 color=red
		// truck color=black make=toyota model=tacoma year="2022"
		// inventory frobnobs=17 widgets=32
	}
}

func TestExampleMarshal5(t *testing.T) {
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

	if data, err := Marshal(staff); err == nil {
		fmt.Println(string(data))
		// ceo "Bob" "Smith" age=76
	}

}

func TestExampleMarshal6(t *testing.T) {
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

	if data, err := Marshal(staff); err == nil {
		fmt.Println(string(data))
		// ceo "Bob" "Smith" age=76
	}

}

func TestExampleMarshal7(t *testing.T) {
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
	if data, err := Marshal(staff); err == nil {
		fmt.Println(string(data))
		// ceo "Bob" "Smith" age=76
	}

}

func TestExampleMarshal8(t *testing.T) {
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

	if data, err := Marshal(people); err == nil {
		fmt.Println(string(data))
		// bob nationality=Canadian {
		// 	language French=false English=true
		// }

	}
}

func TestExampleMarshal9(t *testing.T) {
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

	if data, err := Marshal(people); err == nil {
		fmt.Println(string(data))
		// bob nationality=Canadian {
		// 	language English=true French=false
		// }
	}
}

func TestExampleMarshal10(t *testing.T) {
	type People struct {
		Bob map[string]interface{} `kdl:"bob"`
	}

	people := People{
		Bob: map[string]interface{}{
			"0":      "Johnson",
			"active": true,
			"language": map[string]interface{}{
				"English": true,
				"French":  false,
			},
			"nationality": "Canadian",
		},
	}

	if data, err := Marshal(people); err == nil {
		fmt.Println(string(data))
		// bob "Johnson" nationality=Canadian active=true {
		// 	language English=true French=false
		// }
	}
}

func TestExampleMarshal11(t *testing.T) {
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

	if data, err := Marshal(people); err == nil {
		fmt.Println(string(data))
		// bob "Johnson" active=true {
		// 	language English=true French=false
		// 	nationality "Canadian"
		// }
	}
}

func TestExampleMarshal12(t *testing.T) {
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

	if data, err := Marshal(ngx); err == nil {
		fmt.Println(string(data))
		// location "/" {
		// 	root "/var/www/html"
		// }
		// location "/missing" {
		// 	return 404
		// }
	}
}

func TestExampleMarshal13(t *testing.T) {
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

	if data, err := Marshal(cities); err == nil {
		fmt.Println(string(data))
	}
}

func TestExampleMarshal14(t *testing.T) {
	type People struct {
		Person []map[string]interface{} `kdl:"person,multiple"`
	}

	people := People{
		Person: []map[string]interface{}{
			{"0": "Bob", "active": true},
			{"0": "Jane", "active": true},
		},
	}

	if data, err := Marshal(people); err == nil {
		fmt.Println(string(data))
		// person "Bob" active=true
		// person "Jane" active=true
	}
}

func TestExampleMarshal15(t *testing.T) {
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
	if data, err := Marshal(tf); err == nil {
		fmt.Println(string(data))
		// time-unix 1696805603
		// time-rfc3339 "2023-10-08T15:54:13-07:00"
		// time-rfc822z "08 Oct 23 15:54 -0700"
		// time-date "2023-10-08"
	}
}

func TestExampleMarshal16(t *testing.T) {
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

	if data, err := Marshal(df); err == nil {
		fmt.Println(string(data))
		// duration "2h32m7s"
		// hms "02:32:07.0"
		// seconds 9127
	}
}

func TestExampleMarshal17(t *testing.T) {
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

	if data, err := Marshal(f); err == nil {
		fmt.Println(string(data))
		// bytes-b64 "aGVsbG8="
		// bytes-b64url "dGVzdGluZw=="
		// bytes-b32 "ORSXG5DJNZTQ===="
		// bytes-b32hex "EHIN6T39DPJG===="
		// bytes-hex "74657374696e67"
		// bytes-array 84 69 83 84 73 78 71
		// bytes-string "this is a test"
	}
}

func TestExampleMarshal18(t *testing.T) {
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
	if data, err := Marshal(f); err == nil {
		fmt.Println(string(data))
		// float64posinf "+Inf"
		// float64neginf "-Inf"
		// float64inf "+Inf"
		// float64nan "NaN"
		// float32nan "NaN"
		// float64 0.0
		// float32 0.0
	}
}

func TestExampleMarshal19(t *testing.T) {

	p := People{
		Father: Person{
			FirstName: "BOB",
			LastName:  "JOHNSON",
		},
	}

	if data, err := Marshal(p); err == nil {
		fmt.Println(string(data))
	}
}

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

func TestExampleMarshal20(t *testing.T) {
	p := Family{
		Father: Relative{
			FirstName:  "Bob",
			LastName:   "Johnson",
			CurrentAge: 32,
			IsParent:   true,
		},
	}
	if data, err := Marshal(p); err == nil {
		fmt.Println(string(data))
		// father "Bob" "Johnson" age=32 parent=true
	}

}

func TestUnmarshalNginx(t *testing.T) {
	// home, _ := os.UserHomeDir()
	b, err := os.ReadFile("m:/nginx-sample.conf")
	if err != nil {
		t.Fatalf("failed to read nginx config: %v", err)
	}

	data := string(b)

	type Location struct {
		Args     []interface{}          `kdl:",args"`
		Props    map[string]interface{} `kdl:",props"`
		Children map[string]interface{} `kdl:",children"`
	}
	type Server struct {
		Locations map[string]Location `kdl:"location,multiple"`
	}
	type HTTP struct {
		Server []*Server `kdl:"server,multiple"`
	}
	type NginxServer struct {
		HTTP *HTTP `kdl:"http"`
	}

	var ngx NginxServer
	dec := NewDecoder(strings.NewReader(data))
	dec.Options.RelaxedNonCompliant |= relaxed.NGINXSyntax
	dec.Options.AllowUnhandledNodes = true

	if err := dec.Decode(&ngx); err == nil {
		// dump := spew.NewDefaultConfig()
		// dump.Indent = "    "
		// dump.DisablePointerAddresses = true
		// dump.DisableMethods = true
		// dump.DisablePointerMethods = true
		// dump.Dump(ngx)
	} else {
		t.Fatalf("failed: %v\n", err)
	}

}

func TestUnmarshalTOMLYAMLAssn(t *testing.T) {
	data := `
name: "Bob"
age = 32
`
	type Person struct {
		Name string `kdl:"name"`
		Age  int    `kdl:"age"`
	}

	var p Person

	dec := NewDecoder(strings.NewReader(data))
	dec.Options.RelaxedNonCompliant |= relaxed.YAMLTOMLAssignments
	dec.Options.AllowUnhandledNodes = true

	if err := dec.Decode(&p); err == nil {
		fmt.Printf("%#v\n", p)
		// kdl.Person{Name:"Bob", Age:32}
	} else {
		t.Fatalf("failed: %v\n", err)
	}
}

func TestUnmarshalSuffixes(t *testing.T) {
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
		DiskQuota        int     `kdl:"disk-quota"`
		DiskUsage        int     `kdl:"disk-usage"`
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
		// kdl.Measurements{MinMemory:1000, MaxMemory:2300000, DiskQuota:32000000, DiskUsage:3500000, StorageUsed:33554432, StorageAvail:1.349517312e+08, NotANumberBare:"2.3k", NotANumberQuoted:"3.2k"}
		fmt.Printf("%.3f\n", m.StorageAvail)
	} else {
		println(err.Error())
	}
}
