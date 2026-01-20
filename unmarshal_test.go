package kdl

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/relaxed"
)

const kdlRelaxedSuffixes = `
one-thousand 1k
two-point-three-thousand 2.3k
onek-string "1k"
two-point-threek-string "2.3k"

one-kibi 1kb
two-point-three-kibi 2.3kb

non-numeric-bare 2.3k
non-numeric-quoted "3.2k"
`

type testRelaxedSuffixes struct {
	OneThousand           int     `kdl:"one-thousand"`
	TwoPointThreeThousand int     `kdl:"two-point-three-thousand"`
	OneKString            int     `kdl:"onek-string"`
	TwoPointThreeKString  int     `kdl:"two-point-threek-string"`
	OneKibi               int     `kdl:"one-kibi"`
	TwoPointThreeKibi     float64 `kdl:"two-point-three-kibi"`
	NonNumericBare        string  `kdl:"non-numeric-bare"`
	NonNumericQuoted      string  `kdl:"non-numeric-quoted"`
}

var expectRelaxedSuffixes = testRelaxedSuffixes{
	OneThousand:           1000,
	TwoPointThreeThousand: 2300,
	OneKString:            1000,
	TwoPointThreeKString:  2300,
	OneKibi:               1024,
	TwoPointThreeKibi:     2355.2,
	NonNumericBare:        "2.3k",
	NonNumericQuoted:      "3.2k",
}

func TestRelaxedSuffixes(t *testing.T) {
	into := &testRelaxedSuffixes{}
	want := &expectRelaxedSuffixes
	dec := NewDecoder(bytes.NewReader([]byte(kdlRelaxedSuffixes)))
	dec.Options.RelaxedNonCompliant |= relaxed.MultiplierSuffixes

	if err := dec.Decode(into); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	} else {
		fmt.Printf("%#v\n", into)
		if !reflect.DeepEqual(into, want) {
			t.Fatalf("Unmarshal():\ngot : %#v\nwant: %#v", into, want)
		}
	}
}

const kdlSingleArg = `
name "Bob"
age 76
active true
temperature 98.6
secret 42
divides-by null
`

type testSingleArg struct {
	Name        string      `kdl:"name"`
	Age         int         `kdl:"age"`
	Active      bool        `kdl:"active"`
	Temperature float64     `kdl:"temperature"`
	Secret      interface{} `kdl:"secret"`
	DividesBy   interface{} `kdl:"divides-by"`
}

var expectSingleArg = testSingleArg{
	Name:        "Bob",
	Age:         76,
	Active:      true,
	Temperature: 98.6,
	Secret:      int64(42),
	DividesBy:   nil,
}

var expectSingleArgMap = map[string]interface{}{
	"name":        "Bob",
	"age":         int64(76),
	"active":      true,
	"temperature": float64(98.6),
	"secret":      int64(42),
	"divides-by":  nil,
}

type testSingleArgPtr struct {
	Name        *string     `kdl:"name"`
	Age         *int        `kdl:"age"`
	Active      *bool       `kdl:"active"`
	Temperature *float64    `kdl:"temperature"`
	Secret      interface{} `kdl:"secret"`
	DividesBy   interface{} `kdl:"divides-by"`
}

var (
	expectSingleArgPtrName       = "Bob"
	expectSingleArgPtrAge        = 76
	expectSingleArgPtrActive     = true
	expectSingleArgPtrTempeature = 98.6
	expectSingleArgPtr           = testSingleArgPtr{
		Name:        &expectSingleArgPtrName,
		Age:         &expectSingleArgPtrAge,
		Active:      &expectSingleArgPtrActive,
		Temperature: &expectSingleArgPtrTempeature,
		Secret:      int64(42),
		DividesBy:   nil,
	}
)

type testSingleArgEmbedded struct {
	Age         int     `kdl:"age"`
	Active      bool    `kdl:"active"`
	Temperature float64 `kdl:"temperature"`
}

type testSingleArgEmbed struct {
	testSingleArgEmbedded
	Name      string      `kdl:"name"`
	Secret    interface{} `kdl:"secret"`
	DividesBy interface{} `kdl:"divides-by"`
}

var expectSingleArgEmbed = testSingleArgEmbed{
	testSingleArgEmbedded: testSingleArgEmbedded{
		Age:         76,
		Active:      true,
		Temperature: 98.6,
	},
	Name:      "Bob",
	Secret:    int64(42),
	DividesBy: nil,
}

const kdlMultipleArgs = `
vegetables "broccoli" "carrot" "cucumber"
fruits "apple" "orange" "watermelon"
magic-numbers 4 8 16 32
`

type testMultipleArgs struct {
	Vegetables   []string      `kdl:"vegetables"`
	Fruits       []interface{} `kdl:"fruits"`
	MagicNumbers []int         `kdl:"magic-numbers"`
}

var expectMultipleArgs = testMultipleArgs{
	Fruits:       []interface{}{"apple", "orange", "watermelon"},
	Vegetables:   []string{"broccoli", "carrot", "cucumber"},
	MagicNumbers: []int{4, 8, 16, 32},
}

type testMultipleArgsPtr struct {
	Vegetables   *[]string      `kdl:"vegetables"`
	Fruits       *[]interface{} `kdl:"fruits"`
	MagicNumbers *[]int         `kdl:"magic-numbers"`
}

var expectMultipleArgsPtr = testMultipleArgsPtr{
	Fruits:       &[]interface{}{"apple", "orange", "watermelon"},
	Vegetables:   &[]string{"broccoli", "carrot", "cucumber"},
	MagicNumbers: &[]int{4, 8, 16, 32},
}

const kdlProps = `
car make=ford model=mustang color=red year=1967
truck make=toyota model=tacoma color=black year=2022
inventory widgets=32 frobnobs=17
`

type testProps struct {
	Car       map[string]interface{} `kdl:"car"`
	Truck     map[string]string      `kdl:"truck"`
	Inventory map[string]int         `kdl:"inventory"`
}

var expectProps = testProps{
	Car:       map[string]interface{}{"make": "ford", "model": "mustang", "color": "red", "year": int64(1967)},
	Truck:     map[string]string{"make": "toyota", "model": "tacoma", "color": "black", "year": "2022"},
	Inventory: map[string]int{"widgets": 32, "frobnobs": 17},
}

const kdlArgProps = `
old-man "Bob" "Smith" age=76
young-man "Tim" "Jones" age=21
old-woman "Ethel" "Smith" age=72
young-woman "Sue" "Jones" age=22
ugly-man "Carl" "Smith" age=42
ugly-woman "Anna" "Jones" age=32
crazy-man "Stu" "Jones" age=52
`

type testArgProps struct {
	OldMan struct {
		Args  []interface{}          `kdl:",args"`  // []{"Bob","Smith"}
		Props map[string]interface{} `kdl:",props"` // {"age":76}
	} `kdl:"old-man"`
	YoungMan   map[string]interface{} `kdl:"young-man"`   // {"1":"Tim", "2":"Jones", "age":21}
	OldWoman   []interface{}          `kdl:"old-woman"`   // []{"Ethel","Smith",[]interface{}{"age",72}}
	YoungWoman []string               `kdl:"young-woman"` // []{"Sue","Jones","age=22"}
	UglyMan    struct {
		Args []interface{} `kdl:",args"` // []{"Carl","Smith"}
		Age  int           `kdl:"age"`   // 42
	} `kdl:"ugly-man"`
	UglyWoman struct {
		First string `kdl:",arg"` // Anna
		Last  string `kdl:",arg"` // Jones
		Age   int    `kdl:"age"`  // 32
	} `kdl:"ugly-woman"`
	CrazyMan *struct {
		Args []interface{} `kdl:",args"` // []{"Stu","Jones"}
		Age  int           `kdl:"age"`   // 52
	} `kdl:"crazy-man"`
}

var expectArgProps = testArgProps{
	OldMan: struct {
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	}{
		Args:  []interface{}{"Bob", "Smith"},
		Props: map[string]interface{}{"age": int64(76)},
	},
	YoungMan: map[string]interface{}{
		"0":   "Tim",
		"1":   "Jones",
		"age": int64(21),
	},
	OldWoman: []interface{}{
		"Ethel",
		"Smith",
		[]interface{}{"age", int64(72)},
	},
	YoungWoman: []string{
		"Sue",
		"Jones",
		"age=22",
	},
	UglyMan: struct {
		Args []interface{} `kdl:",args"`
		Age  int           `kdl:"age"`
	}{
		Args: []interface{}{"Carl", "Smith"},
		Age:  42,
	},
	UglyWoman: struct {
		First string `kdl:",arg"`
		Last  string `kdl:",arg"`
		Age   int    `kdl:"age"`
	}{
		First: "Anna",
		Last:  "Jones",
		Age:   32,
	},
	CrazyMan: &struct {
		Args []interface{} `kdl:",args"`
		Age  int           `kdl:"age"`
	}{
		Args: []interface{}{"Stu", "Jones"},
		Age:  52,
	},
}

type testChildrenPerson struct {
	Nationality string          `kdl:"nationality,child"`
	Age         int             `kdl:"age,child"`
	Language    map[string]bool `kdl:"language"`
}

const kdlChildren = `
bob {
	nationality "Canadian"
	age 27
	language English=true French=false
}

klaus {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testChildren struct {
	Bob   map[string]interface{} `kdl:"bob"`
	Klaus testChildrenPerson     `kdl:"klaus"`
}

var expectChildren = testChildren{
	Bob: map[string]interface{}{
		"nationality": "Canadian",
		"age":         int64(27),
		"language":    map[string]interface{}{"English": true, "French": false},
	},
	Klaus: testChildrenPerson{
		Nationality: "German",
		Age:         32,
		Language:    map[string]bool{"English": false, "German": true},
	},
}

const kdlArgsChildren = `
bob "Johnson" {
	nationality "Canadian"
	age 27
	language English=true French=false
}

klaus "Werner" {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testArgsChildren struct {
	Bob   map[string]interface{} `kdl:"bob"`
	Klaus struct {
		Args        []interface{}   `kdl:",args"`
		Nationality string          `kdl:"nationality,child"`
		Age         int             `kdl:"age,child"`
		Language    map[string]bool `kdl:"language,child"`
	} `kdl:"klaus"`
}

var expectArgsChildren = testArgsChildren{
	Bob: map[string]interface{}{
		"0":           "Johnson",
		"nationality": "Canadian",
		"age":         int64(27),
		"language":    map[string]interface{}{"English": true, "French": false},
	},
	Klaus: struct {
		Args        []interface{}   `kdl:",args"`
		Nationality string          `kdl:"nationality,child"`
		Age         int             `kdl:"age,child"`
		Language    map[string]bool `kdl:"language,child"`
	}{
		Args:        []interface{}{"Werner"},
		Nationality: "German",
		Age:         32,
		Language:    map[string]bool{"English": false, "German": true},
	},
}

type testArgsChildrenField struct {
	Bob   map[string]interface{} `kdl:"bob"`
	Klaus struct {
		Args     []interface{}          `kdl:",args"`
		Children map[string]interface{} `kdl:",children"`
	} `kdl:"klaus"`
}

var expectArgsChildrenField = testArgsChildrenField{
	Bob: map[string]interface{}{
		"0":           "Johnson",
		"nationality": "Canadian",
		"age":         int64(27),
		"language":    map[string]interface{}{"English": true, "French": false},
	},
	Klaus: struct {
		Args     []interface{}          `kdl:",args"`
		Children map[string]interface{} `kdl:",children"`
	}{
		Args: []interface{}{"Werner"},
		Children: map[string]interface{}{
			"nationality": "German",
			"age":         int64(32),
			"language":    map[string]interface{}{"English": false, "German": true},
		},
	},
}

const kdlArgsPropsChildren = `
bob "Johnson" active=true {
	nationality "Canadian"
	age 27
	language English=true French=false
}

klaus "Werner" active=false {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testArgsPropsChildren struct {
	Bob   map[string]interface{} `kdl:"bob"`
	Klaus struct {
		Args        []interface{}          `kdl:",args"`
		Props       map[string]interface{} `kdl:",props"`
		Nationality string                 `kdl:"nationality,child"`
		Age         int                    `kdl:"age,child"`
		Language    map[string]bool        `kdl:"language,child"`
	} `kdl:"klaus"`
}

var expectArgsPropsChildren = testArgsPropsChildren{
	Bob: map[string]interface{}{
		"0":           "Johnson",
		"active":      true,
		"nationality": "Canadian",
		"age":         int64(27),
		"language":    map[string]interface{}{"English": true, "French": false},
	},
	Klaus: struct {
		Args        []interface{}          `kdl:",args"`
		Props       map[string]interface{} `kdl:",props"`
		Nationality string                 `kdl:"nationality,child"`
		Age         int                    `kdl:"age,child"`
		Language    map[string]bool        `kdl:"language,child"`
	}{
		Args:        []interface{}{"Werner"},
		Props:       map[string]interface{}{"active": false},
		Nationality: "German",
		Age:         32,
		Language:    map[string]bool{"English": false, "German": true},
	},
}

const kdlMultiChildrenSlice = `
person {
	nationality "Canadian"
	age 27
	language English=true French=false
}

person {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testMultiChildrenSlice struct {
	Person []testChildrenPerson `kdl:"person,multiple"`
}

var expectMultiChildrenSlice = testMultiChildrenSlice{
	Person: []testChildrenPerson{
		{
			Nationality: "Canadian",
			Age:         27,
			Language:    map[string]bool{"English": true, "French": false},
		},
		{
			Nationality: "German",
			Age:         32,
			Language:    map[string]bool{"English": false, "German": true},
		},
	},
}

type testMultiChildrenPtrSlice struct {
	Person []*testChildrenPerson `kdl:"person,multiple"`
}

var expectMultiChildrenPtrSlice = testMultiChildrenPtrSlice{
	Person: []*testChildrenPerson{
		{
			Nationality: "Canadian",
			Age:         27,
			Language:    map[string]bool{"English": true, "French": false},
		},
		{
			Nationality: "German",
			Age:         32,
			Language:    map[string]bool{"English": false, "German": true},
		},
	},
}

const kdlMultiChildrenMap = `
person "Bob" {
	nationality "Canadian"
	age 27
	language English=true French=false
}

person "Klaus" {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testMultiChildrenMap struct {
	Person map[string]testChildrenPerson `kdl:"person,multiple"`
}

var expectMultiChildrenMap = testMultiChildrenMap{
	Person: map[string]testChildrenPerson{
		"Bob": {
			Nationality: "Canadian",
			Age:         27,
			Language:    map[string]bool{"English": true, "French": false},
		},
		"Klaus": {
			Nationality: "German",
			Age:         32,
			Language:    map[string]bool{"English": false, "German": true},
		},
	},
}

const kdlTwoDimMultiChildrenMap = `
person "Johnson" "Bob" {
	nationality "Canadian"
	age 27
	language English=true French=false
}

person "Johnson" "Jim" {
	nationality "Canadian"
	age 35
	language English=true French=false
}

person "Werner" "Klaus" {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testTwoDimMultiChildrenMap struct {
	Person map[string]map[string]testChildrenPerson `kdl:"person,multiple"`
}

var expectTwoDimMultiChildrenMap = testTwoDimMultiChildrenMap{
	Person: map[string]map[string]testChildrenPerson{
		"Johnson": {
			"Bob": {
				Nationality: "Canadian",
				Age:         27,
				Language:    map[string]bool{"English": true, "French": false},
			},
			"Jim": {
				Nationality: "Canadian",
				Age:         35,
				Language:    map[string]bool{"English": true, "French": false},
			},
		},
		"Werner": {
			"Klaus": {
				Nationality: "German",
				Age:         32,
				Language:    map[string]bool{"English": false, "German": true},
			},
		},
	},
}

const kdlArgsPropsTwoDimMultiChildren = `
person "Johnson" "Bob" "leprechaun" active=true {
	nationality "Canadian"
	age 27
	language English=true French=false
}

person "Johnson" "Jim" "chupacabra" active=true {
	nationality "Canadian"
	age 35
	language English=true French=false
}

person "Werner" "Klaus" "sasquatch" active=false {
	nationality "German"
	age 32
	language English=false German=true
}
`

type testArgsPropsChildrenPerson struct {
	Args        []interface{}          `kdl:",args"`
	Props       map[string]interface{} `kdl:",props"`
	Nationality string                 `kdl:"nationality,child"`
	Age         int                    `kdl:"age,child"`
	Language    map[string]bool        `kdl:"language,child"`
}

type testArgsPropsTwoDimMultiChildrenMap struct {
	Person map[string]map[string]testArgsPropsChildrenPerson `kdl:"person,multiple"`
}

var expectArgsPropsTwoDimMultiChildrenMap = testArgsPropsTwoDimMultiChildrenMap{
	Person: map[string]map[string]testArgsPropsChildrenPerson{
		"Johnson": {
			"Bob": {
				Args:        []interface{}{"leprechaun"},
				Props:       map[string]interface{}{"active": true},
				Nationality: "Canadian",
				Age:         27,
				Language:    map[string]bool{"English": true, "French": false},
			},
			"Jim": {
				Args:        []interface{}{"chupacabra"},
				Props:       map[string]interface{}{"active": true},
				Nationality: "Canadian",
				Age:         35,
				Language:    map[string]bool{"English": true, "French": false},
			},
		},
		"Werner": {
			"Klaus": {
				Args:        []interface{}{"sasquatch"},
				Props:       map[string]interface{}{"active": false},
				Nationality: "German",
				Age:         32,
				Language:    map[string]bool{"English": false, "German": true},
			},
		},
	},
}

type testArgsPropsTwoDimMultiChildrenPtrMap struct {
	Person map[string]map[string]*testArgsPropsChildrenPerson `kdl:"person,multiple"`
}

var expectArgsPropsTwoDimMultiChildrenPtrMap = testArgsPropsTwoDimMultiChildrenPtrMap{
	Person: map[string]map[string]*testArgsPropsChildrenPerson{
		"Johnson": {
			"Bob": {
				Args:        []interface{}{"leprechaun"},
				Props:       map[string]interface{}{"active": true},
				Nationality: "Canadian",
				Age:         27,
				Language:    map[string]bool{"English": true, "French": false},
			},
			"Jim": {
				Args:        []interface{}{"chupacabra"},
				Props:       map[string]interface{}{"active": true},
				Nationality: "Canadian",
				Age:         35,
				Language:    map[string]bool{"English": true, "French": false},
			},
		},
		"Werner": {
			"Klaus": {
				Args:        []interface{}{"sasquatch"},
				Props:       map[string]interface{}{"active": false},
				Nationality: "German",
				Age:         32,
				Language:    map[string]bool{"English": false, "German": true},
			},
		},
	},
}

var kdlArgNullPtr = `
person "Bob" "Smith"
`

type testArgNullPtr struct {
	Person struct {
		First *string `kdl:",arg"`
		Last  *string `kdl:",arg"`
	} `kdl:"person"`
}

var (
	expectArgNullPtrFirst = "Bob"
	expectArgNullPtrLast  = "Smith"

	expectArgNullPtr = testArgNullPtr{
		Person: struct {
			First *string `kdl:",arg"`
			Last  *string `kdl:",arg"`
		}{
			First: &expectArgNullPtrFirst,
			Last:  &expectArgNullPtrLast,
		},
	}
)

const kdlDuplicateNodes = `
map-person "Bob" age=32 occupation=Lawyer
map-person "Joe" age=52 occupation=Plumber

struct-person "Bob" age=32 occupation=Lawyer
struct-person "Joe" age=52 occupation=Plumber

slice-person "Bob" age=32 occupation=Lawyer
slice-person "Joe" age=52 occupation=Plumber

overwrite-person "Bob" age=32 occupation=Lawyer
overwrite-person "Joe" age=52 occupation=Plumber

intf-person "Bob" age=32 occupation=Lawyer
intf-person "Joe" age=52 occupation=Plumber
`

type testDuplicateNodes struct {
	MapPerson    map[string]map[string]interface{} `kdl:"map-person,multiple"` // {"Bob":{"age":32,"occupation":"Lawyer"},...}
	StructPerson map[string]struct {               // {"Bob":{...}, "Joe": {...}}
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	} `kdl:"struct-person,multiple"`
	SlicePerson []struct { // []{ {Args:[]{"Bob"},Props: {...}},  {Args:[]{"Joe"},Props: {...}} }
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	} `kdl:"slice-person,multiple"`
	OverwritePerson struct { // {Args:[]{"Joe"},Props: {...}},
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	} `kdl:"overwrite-person"`
	IntfPerson map[string]interface{} `kdl:"intf-person,multiple"` // {"Bob":{"age":32,"occupation":"Lawyer"},...}
}

var expectDuplicateNodes = testDuplicateNodes{
	MapPerson: map[string]map[string]interface{}{
		"Bob": {"age": 32, "occupation": "Lawyer"},
		"Joe": {"age": 52, "occupation": "Plumber"},
	},

	StructPerson: map[string]struct {
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	}{
		"Bob": {Args: nil, Props: map[string]interface{}{"age": 32, "occupation": "Lawyer"}},
		"Joe": {Args: nil, Props: map[string]interface{}{"age": 52, "occupation": "Plumber"}},
	},

	SlicePerson: []struct {
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	}{
		{Args: []interface{}{"Bob"}, Props: map[string]interface{}{"age": 32, "occupation": "Lawyer"}},
		{Args: []interface{}{"Joe"}, Props: map[string]interface{}{"age": 52, "occupation": "Plumber"}},
	},

	OverwritePerson: struct {
		Args  []interface{}          `kdl:",args"`
		Props map[string]interface{} `kdl:",props"`
	}{
		Args:  []interface{}{"Joe"},
		Props: map[string]interface{}{"age": 52, "occupation": "Plumber"},
	},

	IntfPerson: map[string]interface{}{
		"Bob": map[string]interface{}{"age": 32, "occupation": "Lawyer"},
		"Joe": map[string]interface{}{"age": 52, "occupation": "Plumber"},
	},
}

const kdlUnmarshalKDLNode = `
father "Bob" "Johnson" age=32 active=true
mother "Jane" "Johnson" age=28 active=true
`

type testUnmarshalTextPerson struct {
	FirstName  string
	LastName   string
	CurrentAge int
	IsActive   bool
}

func (t *testUnmarshalTextPerson) UnmarshalText(b []byte) error {
	parts := bytes.Split(b, []byte{' '})
	t.FirstName, _ = strconv.Unquote(string(parts[0]))
	t.LastName, _ = strconv.Unquote(string(parts[1]))
	parts = parts[2:]
	for _, part := range parts {
		k, v, _ := strings.Cut(string(part), "=")
		switch k {
		case "age":
			var age int64
			age, _ = strconv.ParseInt(v, 10, 64)
			t.CurrentAge = int(age)
		case "active":
			active, _ := strconv.ParseBool(v)
			t.IsActive = active
		}
	}
	return nil
}

func (t *testUnmarshalTextPerson) MarshalText() ([]byte, error) {
	b := make([]byte, 0, len(t.FirstName)+3+len(t.LastName)+3+len("age=nn ")+len("active=false "))
	b = append(b, strconv.Quote(t.FirstName)...)
	b = append(b, ' ')
	b = append(b, strconv.Quote(t.LastName)...)
	b = append(b, " age="...)
	b = strconv.AppendInt(b, int64(t.CurrentAge), 10)
	b = append(b, " active="...)
	b = strconv.AppendBool(b, t.IsActive)
	return b, nil
}

type testUnmarshalTextNode struct {
	Father *testUnmarshalTextPerson `kdl:"father"`
	Mother testUnmarshalTextPerson  `kdl:"mother"`
}

var expectUnmarshalTextNode = testUnmarshalTextNode{
	Father: &testUnmarshalTextPerson{
		FirstName:  "Bob",
		LastName:   "Johnson",
		CurrentAge: 32,
		IsActive:   true,
	},
	Mother: testUnmarshalTextPerson{
		FirstName:  "Jane",
		LastName:   "Johnson",
		CurrentAge: 28,
		IsActive:   true,
	},
}

type testUnmarshalKDLPerson struct {
	FirstName  string
	LastName   string
	CurrentAge int
	IsActive   bool
}

func (t *testUnmarshalKDLPerson) UnmarshalKDL(node *document.Node) error {
	if len(node.Arguments) != 2 {
		return errors.New("exactly 2 arguments required")
	}
	t.FirstName = strings.ToUpper(node.Arguments[0].ValueString())
	t.LastName = strings.ToUpper(node.Arguments[1].ValueString())

	if age, ok := node.Properties.Unordered()["age"].ResolvedValue().(int64); ok {
		t.CurrentAge = int(age)
	} else {
		return errors.New("age must be an int")
	}

	if active, ok := node.Properties.Unordered()["active"].ResolvedValue().(bool); ok {
		t.IsActive = active
	} else {
		return errors.New("active must be a bool")
	}
	return nil
}

func (t *testUnmarshalKDLPerson) MarshalKDL(node *document.Node) error {
	node.AddArgument(t.FirstName, "")
	node.AddArgument(t.LastName, "")
	node.AddProperty("age", t.CurrentAge, "")
	node.AddProperty("active", t.IsActive, "")
	return nil
}

type testUnmarshalKDLNode struct {
	Father *testUnmarshalKDLPerson `kdl:"father"`
	Mother testUnmarshalKDLPerson  `kdl:"mother"`
}

var expectUnmarshalKDLNode = testUnmarshalKDLNode{
	Father: &testUnmarshalKDLPerson{
		FirstName:  "BOB",
		LastName:   "JOHNSON",
		CurrentAge: 32,
		IsActive:   true,
	},
	Mother: testUnmarshalKDLPerson{
		FirstName:  "JANE",
		LastName:   "JOHNSON",
		CurrentAge: 28,
		IsActive:   true,
	},
}

const kdlUnmarshalTextValue = `
father firstname="Bob" lastname="Johnson"
mother firstname="Jane" lastname="Johnson"
`

type testUnmarshalTextValueName string

func (t *testUnmarshalTextValueName) UnmarshalText(b []byte) error {
	*t = testUnmarshalTextValueName(strings.ToUpper(string(b)))
	return nil
}

func (t *testUnmarshalTextValueName) MarshalText() ([]byte, error) {
	return []byte(*t), nil
}

type testUnmarshalTextValuePerson struct {
	FirstName testUnmarshalTextValueName
	LastName  *testUnmarshalTextValueName
}

type testUnmarshalTextValue struct {
	Father *testUnmarshalTextValuePerson `kdl:"father"`
	Mother testUnmarshalTextValuePerson  `kdl:"mother"`
}

var testUnmarshalTextValueJohnson = testUnmarshalTextValueName("JOHNSON")

var expectUnmarshalTextValue = testUnmarshalTextValue{
	Father: &testUnmarshalTextValuePerson{
		FirstName: "BOB",
		LastName:  &testUnmarshalTextValueJohnson,
	},
	Mother: testUnmarshalTextValuePerson{
		FirstName: "JANE",
		LastName:  &testUnmarshalTextValueJohnson,
	},
}

type testUnmarshalKDLValueName string

func (t *testUnmarshalKDLValueName) UnmarshalKDLValue(v *document.Value) error {
	*t = testUnmarshalKDLValueName(strings.ToUpper(v.ValueString()))
	return nil
}

func (t *testUnmarshalKDLValueName) MarshalKDLValue(v *document.Value) error {
	v.Value = strings.ToLower(string(*t))
	return nil
}

type testUnmarshalKDLValuePerson struct {
	FirstName testUnmarshalKDLValueName
	LastName  *testUnmarshalKDLValueName
}

type testUnmarshalKDLValue struct {
	Father *testUnmarshalKDLValuePerson `kdl:"father"`
	Mother testUnmarshalKDLValuePerson  `kdl:"mother"`
}

var testUnmarshalKDLValueJohnson = testUnmarshalKDLValueName("JOHNSON")

var expectUnmarshalKDLValue = testUnmarshalKDLValue{
	Father: &testUnmarshalKDLValuePerson{
		FirstName: "BOB",
		LastName:  &testUnmarshalKDLValueJohnson,
	},
	Mother: testUnmarshalKDLValuePerson{
		FirstName: "JANE",
		LastName:  &testUnmarshalKDLValueJohnson,
	},
}

const kdlTimeDuration = `
time-unix 1696805603
time-rfc3339 "2023-10-08T15:54:13-07:00"
time-rfc822z "08 Oct 23 15:54 -0700" 
time-date "2023-10-08"
duration "1h32m8s"
unpacked "2023-10-08T15:54:13-07:00" "1h32m8s"
map-times {
	test "2023-10-08T15:54:13-07:00"
}
multi-map-times "woo" {
	testy "2023-10-08T15:54:13-07:00"
	testo "2023-10-08T15:54:13-07:00"
}
`

type testTimeDuration struct {
	TimeUnix    time.Time     `kdl:"time-unix,format:unix"`
	TimeRFC3339 time.Time     `kdl:"time-rfc3339,format:RFC3339"`
	TimeRFC822Z time.Time     `kdl:"time-rfc822z,format:RFC822Z"`
	TimeDate    time.Time     `kdl:"time-date,format:'2006-01-02'"`
	Duration    time.Duration `kdl:"duration"`
	Unpacked    struct {
		First time.Time     `kdl:",arg,format:RFC3339"`
		Last  time.Duration `kdl:",arg"`
	} `kdl:"unpacked"`
	MapTimes      map[string]time.Time            `kdl:"map-times,format:RFC3339"`
	MultiMapTimes map[string]map[string]time.Time `kdl:"multi-map-times,multiple,format:RFC3339"`
}

var (
	expectTimeDurationRFC3339, _  = time.Parse(time.RFC3339, "2023-10-08T15:54:13-07:00")
	expectTimeDurationRFC822Z, _  = time.Parse(time.RFC822Z, "08 Oct 23 15:54 -0700")
	expectTimeDurationDate, _     = time.Parse("2006-01-02", "2023-10-08")
	expectTimeDurationDuration, _ = time.ParseDuration("1h32m8s")
	expectTimeDuration            = testTimeDuration{
		TimeUnix:    time.Unix(1696805603, 0),
		TimeRFC3339: expectTimeDurationRFC3339,
		TimeRFC822Z: expectTimeDurationRFC822Z,
		TimeDate:    expectTimeDurationDate,
		Duration:    expectTimeDurationDuration,
		Unpacked: struct {
			First time.Time     `kdl:",arg,format:RFC3339"`
			Last  time.Duration `kdl:",arg"`
		}(struct {
			First time.Time
			Last  time.Duration
		}{
			First: expectTimeDurationRFC3339,
			Last:  expectTimeDurationDuration,
		}),
		MapTimes: map[string]time.Time{
			"test": expectTimeDurationRFC3339,
		},
		MultiMapTimes: map[string]map[string]time.Time{
			"woo": {
				"testy": expectTimeDurationRFC3339,
				"testo": expectTimeDurationRFC3339,
			},
		},
	}
)

const kdlFormat = `
bytes-b64 "aGVsbG8="
bytes-b64url "dGVzdGluZw=="
bytes-b32 "ORSXG5DJNZTQ===="
bytes-b32hex "EHIN6T39DPJG===="
bytes-b16 "74657374696e67"
bytes-hex "74657374696e67"
bytes-array 84 69 83 84 73 78 71
bytes-string "this is a test"
float64posinf "+Inf"
float64neginf "-Inf"
float64inf "+Inf"
float64nan "NaN"
float32nan "NaN"
float64 "+Inf"
float32 "NaN"
`

type testFormat struct {
	Base64Bytes    []byte  `kdl:"bytes-b64,format:base64"`
	Base64URLBytes []byte  `kdl:"bytes-b64url,format:base64url"`
	Base32Bytes    []byte  `kdl:"bytes-b32,format:base32"`
	Base32HexBytes []byte  `kdl:"bytes-b32hex,format:base32hex"`
	Base16Bytes    []byte  `kdl:"bytes-b16,format:base16"`
	HexBytes       []byte  `kdl:"bytes-hex,format:hex"`
	Array          []byte  `kdl:"bytes-array,format:array"`
	StringBytes    []byte  `kdl:"bytes-string,format:string"`
	Float64PosInf  float64 `kdl:"float64posinf,format:nonfinite"`
	Float64NegInf  float64 `kdl:"float64neginf,format:nonfinite"`
	Float64Inf     float64 `kdl:"float64inf,format:nonfinite"`
	Float64NaN     float64 `kdl:"float64nan,format:nonfinite"`
	Float32NaN     float32 `kdl:"float32nan,format:nonfinite"`
	Float64        float64 `kdl:"float64"`
	Float32        float32 `kdl:"float32"`
}

var expectFormat = testFormat{
	Base64Bytes:    []byte("hello"),
	Base64URLBytes: []byte("testing"),
	Base32Bytes:    []byte("testing"),
	Base32HexBytes: []byte("testing"),
	Base16Bytes:    []byte("testing"),
	HexBytes:       []byte("testing"),
	Array:          []byte("TESTING"),
	StringBytes:    []byte("this is a test"),
	Float64PosInf:  math.Inf(1),
	Float64NegInf:  math.Inf(-1),
	Float64Inf:     math.Inf(1),
	Float64NaN:     math.NaN(),
	Float32NaN:     float32(math.NaN()),
	Float64:        0.0,
	Float32:        0.0,
}

const kdlIgnoreField = `
autoname "this is a test"
explicit-name "another test"
ignored "totally ignored"
`

var expectIgnoreField = testIgnoreField{
	AutoName:     "automatic name",
	ExplicitName: "explicit name",
	Ignored:      "",
}

const kdlChildIntf = `
location "@maintenance" {
	add_header "Content-Type" "text/html"
}
`

type testChildIntfLoc struct {
	Args     []interface{}          `kdl:",args"`
	Props    map[string]interface{} `kdl:",props"`
	Children map[string]interface{} `kdl:",children"`
}

type testChildIntf struct {
	Locations map[string]testChildIntfLoc `kdl:"location,multiple"`
}

var expectChildIntf = testChildIntf{
	Locations: map[string]testChildIntfLoc{
		"@maintenance": {
			Args:  nil,
			Props: nil,
			Children: map[string]interface{}{
				"add_header": []interface{}{"Content-Type", "text/html"},
			},
		},
	},
}

const kdlChildPtrVal = `
location "@maintenance" "fruits" apple="orange" {
	add_header "Content-Type"
}
`

// this is of course a bit of an extreme corner case of pointer usage and still, we haven't tested all possible
// combinations of unmarshaling into pointer values but we at least cover some common cases between this and the other
// tests
type testChildPtrValLoc struct {
	Args     *[]*string          `kdl:",args"`
	Props    *map[string]*string `kdl:",props"`
	Children *map[string]*string `kdl:",children"`
}

type testChildPtrVal struct {
	Locations *map[string]*testChildPtrValLoc `kdl:"location,multiple"`
}

var (
	sfruits           = "fruits"
	sorange           = "orange"
	scontenttype      = "Content-Type"
	expectChildPtrVal = testChildPtrVal{
		Locations: &map[string]*testChildPtrValLoc{
			"@maintenance": {
				Args:  &[]*string{&sfruits},
				Props: &map[string]*string{"apple": &sorange},
				Children: &map[string]*string{
					"add_header": &scontenttype,
				},
			},
		},
	}
)

const kdlChildPtrIntf = `
location "@maintenance" "fruits" apple="orange" {
	add_header "Content-Type"
}
`

// this is of course a bit of an extreme corner case of pointer usage and still, we haven't tested all possible
// combinations of unmarshaling into pointer values but we at least cover some common cases between this and the other
// tests
type testChildPtrIntfLoc struct {
	Args     *[]*interface{}          `kdl:",args"`
	Props    *map[string]*interface{} `kdl:",props"`
	Children *map[string]*interface{} `kdl:",children"`
}

type testChildPtrIntf struct {
	Locations *map[string]*testChildPtrIntfLoc `kdl:"location,multiple"`
}

var (
	ifruits            interface{} = "fruits"
	iorange            interface{} = "orange"
	icontenttype       interface{} = "Content-Type"
	expectChildPtrIntf             = testChildPtrIntf{
		Locations: &map[string]*testChildPtrIntfLoc{
			"@maintenance": {
				Args:  &[]*interface{}{&ifruits},
				Props: &map[string]*interface{}{"apple": &iorange},
				Children: &map[string]*interface{}{
					"add_header": &icontenttype,
				},
			},
		},
	}
)

// TestUnmarshalSuite should be run with `-tags kdldeterministic` to avoid false failures due to nondeterministic map order
func TestUnmarshalSuite(t *testing.T) {
	var (
		intf                   interface{}
		expectSingleArgMapIntf interface{} = expectSingleArgMap
	)
	tests := []struct {
		name string
		kdl  string
		into interface{}
		want interface{}
	}{
		{"singleArg", kdlSingleArg, &testSingleArg{}, &expectSingleArg},
		{"singleArgMap", kdlSingleArg, make(map[string]interface{}), expectSingleArgMap},
		{"singleArgIntf", kdlSingleArg, &intf, &expectSingleArgMapIntf},
		{"singleArgPtr", kdlSingleArg, &testSingleArgPtr{}, &expectSingleArgPtr},
		{"singleArgEmbed", kdlSingleArg, &testSingleArgEmbed{}, &expectSingleArgEmbed},
		{"multipleArgs", kdlMultipleArgs, &testMultipleArgs{}, &expectMultipleArgs},
		{"multipleArgsPtr", kdlMultipleArgs, &testMultipleArgsPtr{}, &expectMultipleArgsPtr},
		{"props", kdlProps, &testProps{}, &expectProps},
		{"argProps", kdlArgProps, &testArgProps{}, &expectArgProps},
		{"children", kdlChildren, &testChildren{}, &expectChildren},
		{"argsChildren", kdlArgsChildren, &testArgsChildren{}, &expectArgsChildren},
		{"argsChildrenField", kdlArgsChildren, &testArgsChildrenField{}, &expectArgsChildrenField},
		{"argsPropsChildren", kdlArgsPropsChildren, &testArgsPropsChildren{}, &expectArgsPropsChildren},
		{"multiChildrenSlice", kdlMultiChildrenSlice, &testMultiChildrenSlice{}, &expectMultiChildrenSlice},
		{"multiChildrenPtrSlice", kdlMultiChildrenSlice, &testMultiChildrenPtrSlice{}, &expectMultiChildrenPtrSlice},
		{"multiChildrenMap", kdlMultiChildrenMap, &testMultiChildrenMap{}, &expectMultiChildrenMap},
		{"twoDimMultiChildrenMap", kdlTwoDimMultiChildrenMap, &testTwoDimMultiChildrenMap{}, &expectTwoDimMultiChildrenMap},
		{"argsPropsTwoDimMultiChildrenMap", kdlArgsPropsTwoDimMultiChildren, &testArgsPropsTwoDimMultiChildrenMap{}, &expectArgsPropsTwoDimMultiChildrenMap},
		{"argsPropsTwoDimMultiChildrenPtrMap", kdlArgsPropsTwoDimMultiChildren, &testArgsPropsTwoDimMultiChildrenPtrMap{}, &expectArgsPropsTwoDimMultiChildrenPtrMap},
		{"argNullPtr", kdlArgNullPtr, &testArgNullPtr{}, &expectArgNullPtr},
		{"unmarshalTextValue", kdlUnmarshalTextValue, &testUnmarshalTextValue{}, &expectUnmarshalTextValue},
		{"unmarshalKDLNode", kdlUnmarshalKDLNode, &testUnmarshalKDLNode{}, &expectUnmarshalKDLNode},
		{"unmarshalKDLValue", kdlUnmarshalTextValue, &testUnmarshalKDLValue{}, &expectUnmarshalKDLValue},
		{"timeDuration", kdlTimeDuration, &testTimeDuration{}, &expectTimeDuration},
		{"format", kdlFormat, &testFormat{}, &expectFormat},
		{"ignoreField", kdlIgnoreField, &testIgnoreField{}, nil},
		{"childIntf", kdlChildIntf, &testChildIntf{}, &expectChildIntf},
		{"childPtrVal", kdlChildPtrVal, &testChildPtrVal{}, &expectChildPtrVal},
		{"childPtrIntf", kdlChildPtrIntf, &testChildPtrIntf{}, &expectChildPtrIntf},
		// {"duplicateNodes", kdlDuplicateNodes, &testDuplicateNodes{}, &expectDuplicateNodes},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal([]byte(tt.kdl), tt.into); err != nil {
				if tt.want != nil {
					t.Fatalf("Unmarshal() error = %v", err)
				}
			} else {

				if tt.name == "format" {
					into := tt.into.(*testFormat)
					want := tt.want.(*testFormat)
					// NaN never equals another NaN so we have to zero these out
					into.Float32NaN = 0
					want.Float32NaN = 0
					into.Float64NaN = 0
					want.Float64NaN = 0
				}

				// into := tt.into.(*testUnmarshalTextValue)
				// want := tt.want.(*testUnmarshalTextValue)
				// fmt.Printf("expect %#v, got %#v\n", *want.Father, *into.Father)

				// fmt.Printf("%#v\n", tt.into)

				if !reflect.DeepEqual(tt.into, tt.want) {
					t.Fatalf("Unmarshal():\ngot : %#v\nwant: %#v", tt.into, tt.want)
				}
			}
		})
	}

}

func TestUnmarshalProfile(t *testing.T) {
	cpuf, err := os.Create("cpu.pprof")
	if err != nil {
		panic(fmt.Sprintf("Failed to create CPU profile: %v", err))
	}
	_ = pprof.StartCPUProfile(cpuf)
	defer func() {
		pprof.StopCPUProfile()
		_ = cpuf.Close()
	}()

	memf, err := os.Create("mem.pprof")
	if err != nil {
		panic(fmt.Sprintf("Failed to create memory profile: %v", err))
	}
	defer func() {
		runtime.GC()

		_ = pprof.Lookup("heap").WriteTo(memf, 0)
		_ = memf.Close()
	}()

	for i := 0; i < 2000; i++ {
		TestUnmarshalSuite(t)
	}

}

func TestBug4(t *testing.T) {
	b := []byte(`
map "skipped"
map key="skipped" key="value"
`)
	var got map[string]interface{}
	err := Unmarshal(b, &got)
	if err != nil {
		t.Errorf("Failed: %v\n", err)
	}
	want := map[string]interface{}{"map": map[string]interface{}{"key": "value"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TestBug4: got %#v, want %#v", got, want)
	}
}

/*
These tests cannot be run as part of a larger batch as the AddCustomUnmarshaler calls will panic given that other tests
have already created Indexers. Uncomment and run individually to test custom unmarshaling.

func TestCustomUnmarshaler(t *testing.T) {
	type coocooKachoo struct {
		S string
	}
	type snackbar struct {
		Chugga coocooKachoo `kdl:"chugga"`
	}

	AddCustomUnmarshaler[coocooKachoo](func(node *document.Node, v reflect.Value) error {
		if len(node.Arguments) == 0 {
			return errors.New("no arguments on this node")
		}
		v.Field(0).SetString("custom " + node.Arguments[0].ValueString())
		return nil
	})

	v := &snackbar{}

	err := Unmarshal([]byte(`chugga "choo choo"`), v)
	if err != nil {
		t.Fatal(err)
	}

	got := v.Chugga.S
	want := `custom choo choo`

	if got != want {
		t.Fatalf("want: %s\n got: %s\n", want, got)
	}
}

func TestCustomValueUnmarshaler(t *testing.T) {
	type coocooKachoo struct {
		S string
	}
	type snackbar struct {
		Chugga coocooKachoo `kdl:"chugga"`
	}

	AddCustomValueUnmarshaler[coocooKachoo](func(value *document.Value, v reflect.Value, format string) error {
		v.Field(0).SetString("custom " + value.ValueString())
		return nil
	})

	v := &snackbar{}

	err := Unmarshal([]byte(`chugga "choo choo"`), v)
	if err != nil {
		t.Fatal(err)
	}

	got := v.Chugga.S
	want := `custom choo choo`

	if got != want {
		t.Fatalf("want: %s\n got: %s\n", want, got)
	}
}
*/
