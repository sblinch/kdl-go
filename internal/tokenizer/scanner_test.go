package tokenizer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/pprof"
	"testing"

	"github.com/sblinch/kdl-go/relaxed"
)

const kdlSchema = `document {
	tests {
		small-integer-unsigned 3
		large-integer-unsigned 314159
		small-integer-signed -3
		large-integer-signed -314159
		small-float-unsigned 3.14159
		large-float-unsigned 31415.9
		small-float-signed -3.14159
		large-float-signed -31415.9
		exp-unsigned-unsigned 3.14159E3
		exp-signed-unsigned -3.14159E4
		exp-unsigned-signed 3.14159E-10
		exp-signed-signed -3.14159E-20
		hex 0xdeadbeef
		octal 0o1755
		binary 0b11011011
		int64 (i64)1234
		float (f64)1234.56
		url (url)"http://www.google.ca"
		custom (custom)"this is my custom type"
		boolean true
		nullean null
		quoted-with-escape-seqs "this\tis a test\nnice, right?"
		multiline-comment-at-end 42 /* comment
with multiple
lines */
		-tricky-name 3
		interrupted /* this is a comment */ 42
		comment-at-end 42 // this is a comment
	}
	spec-examples {
		foo 1 key="val" 3 {
			bar
			(role)baz 1 2
		}
		my-node 1 2 \  // comments are ok after \
        		3 4    // This is the actual end of the Node.
		node a=1 a=2
		my-node 1 2 3 "a" "b" "c"
		
		parent {
			child1
			child2
		}
		
		parent { child1; child2; }

		node (u8)123
		node prop=(regex)".*"
		(published)date "1970-01-01"
		(contributor)person name="Foo McBar"

		just-escapes r"\n will be literal"
		quotes-and-escapes r#"hello\n\r\asd"world"#

		my-node true value=false
		my-node null key=null
		foo {
			bar
		}
		baz
	}
	semicolons {
		title "Yep";
		value 42;
	}
    info {
        title "KDL Schema" lang="en"
        description "KDL Schema KDL schema in KDL" lang="en"
        author "Kat Marchán" {
            link "https://github.com/zkat" rel="self"
        }
        contributor "Lars Willighagen" {
            link "https://github.com/larsgw" rel="self"
        }
        link "https://github.com/zkat/kdl" rel="documentation"
        license "Creative Commons Attribution-ShareAlike 4.0 International License" spdx="CC-BY-SA-4.0" {
            link "https://creativecommons.org/licenses/by-sa/4.0/" lang="en"
        }
        published "2021-08-31"
        modified "2021-09-01"
    }
    node "document" {
        min 1
        max 1
        children id="node-children" {
            node "node-names" id="node-names-node" description="Validations to apply specifically to arbitrary node names" {
                children ref=r#"[id="validations"]"#
            }
            node "other-nodes-allowed" id="other-nodes-allowed-node" description="Whether to allow child nodes other than the ones explicitly listed. Defaults to 'false'." {
                max 1
                value {
                    min 1
                    max 1
                    type "boolean"
                }
            }
            node "tag-names" description="Validations to apply specifically to arbitrary type tag names" {
                children ref=r#"[id="validations"]"#
            }
            node "other-tags-allowed" description="Whether to allow child node tags other than the ones explicitly listed. Defaults to 'false'." {
                max 1
                value {
                    min 1
                    max 1
                    type "boolean"
                }
            }
            node "info" description="A child node that describes the schema itself." {
                children {
                    node "title" description="The title of the schema or the format it describes" {
                        value description="The title text" {
                            type "string"
                            min 1
                            max 1
                        }
                        prop "lang" id="info-lang" description="The language of the text" {
                            type "string"
                        }
                    }
                    node "description" description="A description of the schema or the format it describes" {
                        value description="The description text" {
                            type "string"
                            min 1
                            max 1
                        }
                        prop ref=r#"[id="info-lang"]"#
                    }
                    node "author" description="Author of the schema" {
                        value id="info-person-name" description="Person name" {
                            type "string"
                            min 1
                            max 1
                        }
                        prop "orcid" id="info-orcid" description="The ORCID of the person" {
                            type "string"
                            pattern r"\d{4}-\d{4}-\d{4}-\d{4}"
                        }
                        children {
                            node ref=r#"[id="info-link"]"#
                        }
                    }
                    node "contributor" description="Contributor to the schema" {
                        value ref=r#"[id="info-person-name"]"#
                        prop ref=r#"[id="info-orcid"]"#
                        children {
                            node ref=r#"[id="info-link"]"#
                        }
                    }
                    node "link" id="info-link" description="Links to itself, and to sources describing it" {
                        value description="A URL that the link points to" {
                            type "string"
                            format "url" "irl"
                            min 1
                            max 1
                        }
                        prop "rel" description="The relation between the current entity and the URL" {
                            type "string"
                            enum "self" "documentation"
                        }
                        prop ref=r#"[id="info-lang"]"#
                    }
                    node "license" description="The license(s) that the schema is licensed under" {
                        value description="Name of the used license" {
                            type "string"
                            min 1
                            max 1
                        }
                        prop "spdx" description="An SPDX license identifier" {
                            type "string"
                        }
                        children {
                            node ref=r#"[id="info-link"]"#
                        }
                    }
                    node "published" description="When the schema was published" {
                        value description="Publication date" {
                            type "string"
                            format "date"
                            min 1
                            max 1
                        }
                        prop "time" id="info-time" description="A time to accompany the date" {
                            type "string"
                            format "time"
                        }
                    }
                    node "modified" description="When the schema was last modified" {
                        value description="Modification date" {
                            type "string"
                            format "date"
                            min 1
                            max 1
                        }
                        prop ref=r#"[id="info-time"]"#
                    }
                    node "version" description="The version number of this version of the schema" {
                        value description="Semver version number" {
                            type "string"
                            pattern r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$"
                            min 1
                            max 1
                        }
                    }
                }
            }
            node "tag" id="tag-node" description="A tag belonging to a child node of document or another node." {
                value description="The name of the tag. If a tag name is not supplied, the node rules apply to _all_ nodes belonging to the parent." {
                    type "string"
                    max 1
                }
                prop "description" description="A description of this node's purpose." {
                    type "string"
                }
                prop "id" description="A globally-unique ID for this node." {
                    type "string"
                }
                prop "ref" description="A globally unique reference to another node." {
                    type "string"
                    format "kdl-query"
                }
                children {
                    node ref=r#"[id="node-names-node"]"#
                    node ref=r#"[id="other-nodes-allowed-node"]"#
                    node ref=r#"[id="node-node"]"#
                }
            }
            node "node" id="node-node" description="A child node belonging either to document or to another node. Nodes may be anonymous." {
                value description="The name of the node. If a node name is not supplied, the node rules apply to _all_ nodes belonging to the parent." {
                    type "string"
                    max 1
                }
                prop "description" description="A description of this node's purpose." {
                    type "string"
                }
                prop "id" description="A globally-unique ID for this node." {
                    type "string"
                }
                prop "ref" description="A globally unique reference to another node." {
                    type "string"
                    format "kdl-query"
                }
                children {
                    node "prop-names" description="Validations to apply specifically to arbitrary property names" {
                        children ref=r#"[id="validations"]"#
                    }
                    node "other-props-allowed" description="Whether to allow properties other than the ones explicitly listed. Defaults to 'false'." {
                        max 1
                        value {
                            min 1
                            max 1
                            type "boolean"
                        }
                    }
                    node "min" description="minimum number of instances of this node in its parent's children." {
                        max 1
                        value {
                            min 1
                            max 1
                            type "number"
                        }
                    }
                    node "max" description="maximum number of instances of this node in its parent's children." {
                        max 1
                        value {
                            min 1
                            max 1
                            type "number"
                        }
                    }
                    node ref=r#"[id="value-tag-node"]"#
                    node "prop" id="prop-node" description="A node property key/value pair." {
                        value description="The property key." {
                            type "string"
                        }
                        prop "id" description="A globally-unique ID of this property." {
                            type "string"
                        }
                        prop "ref" description="A globally unique reference to another property node." {
                            type "string"
                            format "kdl-query"
                        }
                        prop "description" description="A description of this property's purpose." {
                            type "string"
                        }
                        children description="Property-specific validations." {
                            node "required" description="Whether this property is required if its parent is present." {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "boolean"
                                }
                            }
                        }
                        children id="validations" description="General value validations." {
                            node "tag" id="value-tag-node" description="The tags associated with this value" {
                                max 1
                                children ref=r#"[id="validations"]"#
                            }
                            node "type" description="The type for this prop's value." {
                                max 1
                                value {
                                    min 1
                                    type "string"
                                }
                            }
                            node "enum" description="An enumeration of possible values" {
                                max 1
                                value description="Enumeration choices" {
                                    min 1
                                }
                            }
                            node "pattern" description="PCRE (Regex) pattern or patterns to test prop values against." {
                                value {
                                    min 1
                                    type "string"
                                }
                            }
                            node "min-length" description="Minimum length of prop value, if it's a string." {
                                max 1
                                value {
                                    min 1
                                    type "number"
                                }
                            }
                            node "max-length" description="Maximum length of prop value, if it's a string." {
                                max 1
                                value {
                                    min 1
                                    type "number"
                                }
                            }
                            node "format" description="Intended data format." {
                                max 1
                                value {
                                    min 1
                                    type "string"
                                    // https://json-schema.org/understanding-json-schema/reference/string.html#format
                                    enum "date-time" "date" "time" "duration" "decimal" "currency" "country-2" "country-3" "country-subdivision" "email" "idn-email" "hostname" "idn-hostname" "ipv4" "ipv6" "url" "url-reference" "irl" "irl-reference" "url-template" "regex" "uuid" "kdl-query" "i8" "i16" "i32" "i64" "u8" "u16" "u32" "u64" "isize" "usize" "f32" "f64" "decimal64" "decimal128"
                                }
                            }
                            node "%" description="Only used for numeric values. Constrains them to be multiples of the given number(s)" {
                                max 1
                                value {
                                    min 1
                                    type "number"
                                }
                            }
                            node ">" description="Only used for numeric values. Constrains them to be greater than the given number(s)" {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                            node ">=" description="Only used for numeric values. Constrains them to be greater than or equal to the given number(s)" {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                            node "<" description="Only used for numeric values. Constrains them to be less than the given number(s)" {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                            node "<=" description="Only used for numeric values. Constrains them to be less than or equal to the given number(s)" {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                        }
                    }
                    node "value" id="value-node" description="one or more direct node values" {
                        prop "id" description="A globally-unique ID of this value." {
                            type "string"
                        }
                        prop "ref" description="A globally unique reference to another value node." {
                            type "string"
                            format "kdl-query"
                        }
                        prop "description" description="A description of this property's purpose." {
                            type "string"
                        }
                        children ref=r#"[id="validations"]"#
                        children description="Node value-specific validations" {
                            node "min" description="minimum number of values for this node." {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                            node "max" description="maximum number of values for this node." {
                                max 1
                                value {
                                    min 1
                                    max 1
                                    type "number"
                                }
                            }
                        }
                    }
                    node "children" id="children-node" {
                        prop "id" description="A globally-unique ID of this children node." {
                            type "string"
                        }
                        prop "ref" description="A globally unique reference to another children node." {
                            type "string"
                            format "kdl-query"
                        }
                        prop "description" description="A description of this these children's purpose." {
                            type "string"
                        }
                        children ref=r#"[id="node-children"]"#
                    }
                }
            }
            node "definitions" description="Definitions to reference in parts of the top-level nodes" {
                children {
                    node ref=r#"[id="node-node"]"#
                    node ref=r#"[id="value-node"]"#
                    node ref=r#"[id="prop-node"]"#
                    node ref=r#"[id="children-node"]"#
                    node ref=r#"[id="tag-node"]"#
                }
            }
        }
    }
}
`

func TestTokenize(t *testing.T) {
	input := []byte(kdlSchema)
	scanner := NewSlice(input)
	tokens, err := scanner.ScanAll()
	if err != nil {
		t.Errorf("failed to tokenize: %+v", err)
	}
	fmt.Printf("Tokens: %+v\n", tokens)
}

type kdlTestCase struct {
	input     []byte
	expectErr bool
}

func loadTestCases() map[string]kdlTestCase {
	cwd, _ := os.Getwd()
	testcasePath := filepath.Join(filepath.Dir(cwd), "kdl-org", "tests", "test_cases")
	inputPath := filepath.Join(testcasePath, "input")
	expectedPath := filepath.Join(testcasePath, "expected_kdl")
	cases, err := filepath.Glob(filepath.Join(inputPath, "*.kdl"))
	if err != nil {
		panic(fmt.Sprintf("can't find test cases: %v", err))
	}
	if len(cases) == 0 {
		panic("can't find any test cases")
	}

	testCases := make(map[string]kdlTestCase)

	for _, testCase := range cases {
		input, err := os.ReadFile(testCase)
		if err != nil {
			panic(fmt.Sprintf("cannot open test case %q: %v", testCase, err))
		}

		expectedPath := filepath.Join(expectedPath, filepath.Base(testCase))
		// println("!!! expected: ", expectedPath)
		_, err = os.Stat(expectedPath)
		expectAnError := err != nil

		testCases[testCase] = kdlTestCase{input, expectAnError}
	}

	return testCases
}

func TestTokenizeTestCases(t *testing.T) {
	testCases := loadTestCases()

	for testCase, tc := range testCases {
		println("===== ", testCase)
		scanner := NewSlice(tc.input)
		scanner.Logger = nil
		_, err := scanner.ScanAll()
		if err != nil {
			println("Error: ", err.Error())
		}
		if err != nil && !tc.expectErr {
			t.Fatalf("test case %q failed: %v", testCase, err)
		} else if err == nil && tc.expectErr {
			t.Errorf("test case %q succeeded incorrectly", testCase)
		}
	}
}

func runTestCases(testCases map[string]kdlTestCase, alternate bool) error {
	for testCase, tc := range testCases {
		scanner := NewSlice(tc.input)
		scanner.Logger = nil
		scanner.Alt = alternate
		_, err := scanner.ScanAll()
		if err != nil && !tc.expectErr {
			return fmt.Errorf("test case %q failed: %v", testCase, err)
		}
	}

	return nil
}

func BenchmarkTokenizeTestCases(b *testing.B) {
	testCases := loadTestCases()

	b.Run("baseline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := runTestCases(testCases, false); err != nil {
				b.Fatalf("failed: %v", err)
			}
		}
	})

	b.Run("alternate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := runTestCases(testCases, true); err != nil {
				b.Fatalf("failed: %v", err)
			}
		}
	})
}

func TestTokenizerProfile(t *testing.T) {
	testCases := loadTestCases()
	println(len(testCases), "test cases")

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

	for i := 0; i < 1000; i++ {
		if err := runTestCases(testCases, true); err != nil {
			t.Fatalf("failed: %v", err)
		}
	}

}

func TestScanner_mark(t *testing.T) {
	s := NewSlice([]byte("node \"fungamunga\"; check 1 2 3; testing \"one\" \"two\";"))
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	s.pushMark()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	v := s.copyFromMark()
	s.popMark()

	if string(v) != "check 1 2 3;" {
		t.Fatalf("expected \"check 1 2 3;\", got %q", string(v))
	}

	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	_, _ = s.readNext()
	s.pushMark()
	_, _ = s.readNext()
	_, _ = s.readNext()
	v = s.copyFromMark()
	s.popMark()

	if string(v) != "\"two\";" {
		t.Fatalf("expected \"\"two\";\", got %q", string(v))
	}
}

func TestScanner_refill(t *testing.T) {
	kdl := kdlSchema

	// kdl = "node \"superdonkeylovingchickentickler is my favorite uncle chucker\""
	preload := NewSlice([]byte(kdl))
	stream := NewBuffer(bytes.NewReader([]byte(kdl)), make([]byte, 32))

	for i := 0; i < 15; i++ {
		pb := preload.Scan()
		ps := stream.Scan()
		if !reflect.DeepEqual(preload.err, stream.err) {
			t.Fatalf("token %d\ngot error %v from preload,\ngot error %v from stream", i, preload.err, stream.err)
		}
		if pb != ps {
			t.Fatalf("token %d\ngot result %v from preload,\ngot result %v from stream", i, pb, ps)
		}
		pt := preload.Token().String()
		st := stream.Token().String()
		if pt != st {
			t.Fatalf("token %d\ngot %q from preload\ngot %q from stream", i, pt, st)
		}
	}

}

func TestScanner_extractLineAtOffset(t *testing.T) {
	scanner := NewSlice([]byte(
		`document {

  test {
    testy woop woop;
  }
	tests {
		small-integer-unsigned 3
		large-integer-unsigned 314159
		small-integer-signed -3
		large-integer-signed -314159
		small-float-unsigned 3.14159
		large-float-unsigned 31415.9
		small-float-signed -3.14159
		large-float-signed -31415.9
		exp-unsigned-unsigned 3.14159E3
		exp-signed-unsigned -3.14159E4
		exp-unsigned-signed 3.14159E-10
		exp-signed-signed -3.14159E-20
		hex 0xdeadbeef
		octal 0o1755
		binary 0b11011011
		int64 (i64)1234
		float (f64)1234.56
		url (url)"http://www.google.ca"
		custom (custom)"this is my custom type"
		boolean true
		nullean null
		quoted-with-escape-seqs "this\tis a test\nnice, right?"
		multiline-comment-at-end 42 /* comment
with multiple
lines */
		-tricky-name 3
		interrupted /* this is a comment */ 42
		comment-at-end 42 // this is a comment
	}

	spec-examples {
		foo 1 key="val" 3 {
			bar
			(role)baz 1 2
		}
		my-node 1 2 \  // comments are ok after \
        		3 4    // This is the actual end of the Node.
		node a=1 a=2
		my-node 1 2 3 "a" "b" "c"
		
		parent {
			child1
			child2
		}
		
		parent { child1; child2; }

		node (u8)123
		node prop=(regex)".*"
		(published)date "1970-01-01"
		(contributor)person name="Foo McBar"

		just-escapes r"\n will be literal"
		quotes-and-escapes r#"hello\n\r\asd"world"#

		my-node true value=false
		my-node null key=null
		foo {
			bar
		}
		baz
	}
	semicolons {
		title "Yep";
		value 42;
	}
}
`))
	tests := []struct {
		offset int
		want   string
	}{
		{9, "document {\n         ^"},
		{31, "    testy woop woop;\n          ^"},
		{137, "  small-integer-signed -3\n                       ^"},
		{690, "  quoted-with-escape-seqs \"this\\tis a test\\nnice, right?\"\n                                                  ^"},
		{817, "  interrupted /* this is a comment */ 42\n                                      ^"},
		{1450, "  value 42;\n  ^"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("offset-%d", tt.offset), func(t *testing.T) {
			if got := scanner.extractLineAtOffset(tt.offset); got != tt.want {
				t.Errorf("extractLineAtOffset()\ngot :\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRelaxed(t *testing.T) {
	input := []byte(`
location / {
	# this is a comment
	root /var/www/html;
}
`)
	scanner := NewSlice(input)
	scanner.RelaxedNonCompliant = relaxed.NGINXSyntax | relaxed.YAMLTOMLAssignments
	tokens, err := scanner.ScanAll()
	if err != nil {
		t.Errorf("failed to tokenize: %+v", err)
	}
	fmt.Printf("Tokens: %+v\n", tokens)
}
