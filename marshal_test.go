package kdl

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var (
	kdlOutputSingleArg = `
name "Bob"
age 76
active true
temperature 98.6
secret 42
divides-by null
`

	kdlOutputSingleArgMap = `
active true
age 76
divides-by null
name "Bob"
secret 42
temperature 98.6
`

	kdlOutputSingleArgIntf = `
active true
age 76
divides-by null
name "Bob"
secret 42
temperature 98.6
`

	kdlOutputSingleArgPtr = `
name "Bob"
age 76
active true
temperature 98.6
secret 42
divides-by null
`

	kdlOutputSingleArgEmbed = `
age 76
active true
temperature 98.6
name "Bob"
secret 42
divides-by null
`

	kdlOutputMultipleArgs = `
vegetables "broccoli" "carrot" "cucumber"
fruits "apple" "orange" "watermelon"
magic-numbers 4 8 16 32
`

	kdlOutputMultipleArgsPtr = `
vegetables "broccoli" "carrot" "cucumber"
fruits "apple" "orange" "watermelon"
magic-numbers 4 8 16 32
`

	kdlOutputProps = `
car color="red" make="ford" model="mustang" year=1967
truck color="black" make="toyota" model="tacoma" year="2022"
inventory frobnobs=17 widgets=32
`

	kdlOutputArgProps = `
old-man "Bob" "Smith" age=76
young-man "Tim" "Jones" age=21
old-woman "Ethel" "Smith" age=72
young-woman "Sue" "Jones" age=22
ugly-man "Carl" "Smith" age=42
ugly-woman "Anna" "Jones" age=32
crazy-man "Stu" "Jones" age=52
`

	kdlOutputChildren = `
bob age=27 nationality="Canadian" {
	language English=true French=false
}
klaus {
	nationality "German"
	age 32
	language English=false German=true
}
`

	kdlOutputArgsChildren = `
bob "Johnson" age=27 nationality="Canadian" {
	language English=true French=false
}
klaus "Werner" {
	nationality "German"
	age 32
	language English=false German=true
}
`

	kdlOutputArgsChildrenField = `
bob "Johnson" age=27 nationality="Canadian" {
	language English=true French=false
}
klaus "Werner" {
	age 32
	language English=false German=true
	nationality "German"
}
`

	kdlOutputArgsPropsChildren = `
bob "Johnson" active=true age=27 nationality="Canadian" {
	language English=true French=false
}
klaus "Werner" active=false {
	nationality "German"
	age 32
	language English=false German=true
}
`

	kdlOutputMultiChildrenSlice = `
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

	kdlOutputMultiChildrenMap = `
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

	kdlOutputTwoDimMultiChildrenMap = `
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
}`

	kdlOutputArgsPropsTwoDimMultiChildrenMap = `
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

	kdlOutputArgsPropsTwoDimMultiChildrenPtrMap = `
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

	kdlOutputArgNullPtr = `
person "Bob" "Smith"
`

	kdlOutputMarshalTextNode = `
father "\"Bob\""
mother "\"Jane\""
`

	kdlOutputMarshalKDLNode = `
father "BOB" "JOHNSON" age=32 active=true
mother "JANE" "JOHNSON" age=28 active=true
`

	kdlOutputMarshalTextValue = `
father firstname="BOB" lastname="JOHNSON"
mother firstname="JANE" lastname="JOHNSON"
`

	kdlOutputMarshalKDLValue = `
father firstname="bob" lastname="johnson"
mother firstname="jane" lastname="johnson"
`

	kdlOutputTimeDuration = `
time-unix 1696805603
time-rfc3339 "2023-10-08T15:54:13-07:00"
time-rfc822z "08 Oct 23 15:54 -0700"
time-date "2023-10-08"
duration "1h32m8s"
unpacked "2023-10-08T15:54:13-07:00" "1h32m8s"
map-times test="2023-10-08T15:54:13-07:00"
multi-map-times "woo" "testo" "2023-10-08T15:54:13-07:00"
multi-map-times "woo" "testy" "2023-10-08T15:54:13-07:00"
`

	kdlOutputFormat = `
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
float64 0.0
float32 0.0
`

	kdlOutputIgnoreField = `
autoname "this is a test"
explicit-name "another test"
`
)

type testIgnoreField struct {
	AutoName     string
	ExplicitName string `kdl:"explicit-name"`
	Ignored      string `kdl:"-"`
}

var srcIgnoreField = testIgnoreField{
	AutoName:     "this is a test",
	ExplicitName: "another test",
	Ignored:      "omit me, please",
}

// TestMarshalSuite should be run with `-tags kdldeterministic` to avoid false failures due to nondeterministic map order
func TestMarshalSuite(t *testing.T) {
	var (
		expectSingleArgMapIntf interface{} = expectSingleArgMap
	)
	tests := []struct {
		name string
		intf interface{}
		want string
	}{
		{"singleArg", expectSingleArg, kdlOutputSingleArg},
		{"singleArgMap", expectSingleArgMap, kdlOutputSingleArgMap},
		{"singleArgIntf", &expectSingleArgMapIntf, kdlOutputSingleArgIntf},
		{"singleArgPtr", expectSingleArgPtr, kdlOutputSingleArgPtr},
		{"singleArgEmbed", expectSingleArgEmbed, kdlOutputSingleArgEmbed},
		{"multipleArgs", expectMultipleArgs, kdlOutputMultipleArgs},
		{"multipleArgsPtr", expectMultipleArgsPtr, kdlOutputMultipleArgsPtr},
		{"props", expectProps, kdlOutputProps},
		{"argProps", expectArgProps, kdlOutputArgProps},
		{"children", expectChildren, kdlOutputChildren},
		{"argsChildren", expectArgsChildren, kdlOutputArgsChildren},
		{"argsChildrenField", expectArgsChildrenField, kdlOutputArgsChildrenField},
		{"argsPropsChildren", expectArgsPropsChildren, kdlOutputArgsPropsChildren},
		{"multiChildrenSlice", expectMultiChildrenSlice, kdlOutputMultiChildrenSlice},
		{"multiChildrenMap", expectMultiChildrenMap, kdlOutputMultiChildrenMap},
		{"twoDimMultiChildrenMap", expectTwoDimMultiChildrenMap, kdlOutputTwoDimMultiChildrenMap},
		{"argsPropsTwoDimMultiChildrenMap", expectArgsPropsTwoDimMultiChildrenMap, kdlOutputArgsPropsTwoDimMultiChildrenMap},
		{"argsPropsTwoDimMultiChildrenPtrMap", expectArgsPropsTwoDimMultiChildrenPtrMap, kdlOutputArgsPropsTwoDimMultiChildrenPtrMap},
		{"argNullPtr", expectArgNullPtr, kdlOutputArgNullPtr},
		{"marshalTextNode", expectUnmarshalTextNode, kdlOutputMarshalTextNode},
		{"marshalTextValue", expectUnmarshalTextValue, kdlOutputMarshalTextValue},
		{"marshalKDLNode", expectUnmarshalKDLNode, kdlOutputMarshalKDLNode},
		{"marshalKDLValue", expectUnmarshalKDLValue, kdlOutputMarshalKDLValue},
		{"timeDuration", expectTimeDuration, kdlOutputTimeDuration},
		{"format", expectFormat, kdlOutputFormat},
		{"ignoreField", srcIgnoreField, kdlOutputIgnoreField},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if b, err := Marshal(tt.intf); err != nil {
				t.Fatalf("Marshal() error = %v", err)
			} else {
				got := string(bytes.TrimSpace(b))
				want := strings.TrimSpace(tt.want)
				if got != want {
					t.Fatalf("Marshal():\ngot :\n%s\nwant:\n%s", got, want)
				}
			}
		})
	}

}

func TestBug1(t *testing.T) {
	type Foo struct {
		T    time.Time `kdl:"time,child"`
		Name string    `kdl:"name,child"`
	}
	type Bar struct {
		Foos []Foo `kdl:"foo,multiple"`
	}

	tt := time.Date(2023, 11, 12, 1, 2, 3, 0, time.UTC)

	foo := Bar{
		Foos: []Foo{
			{T: tt, Name: "dan"},
			{T: tt, Name: "eve"},
		},
	}

	got, _ := Marshal(foo)
	want := `foo {
	time "2023-11-12T01:02:03Z"
	name "dan"
}
foo {
	time "2023-11-12T01:02:03Z"
	name "eve"
}
`

	if string(got) != want {
		t.Errorf("TestBug1: want:\n%s\n\ngot:\n%s\n", want, string(got))
	}
}
