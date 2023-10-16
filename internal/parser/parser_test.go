package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"

	"github.com/sblinch/kdl-go/internal/generator"
	"github.com/sblinch/kdl-go/internal/tokenizer"
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
        exp-unsigned-unsigned 3.14159e+10
        exp-signed-unsigned -3.14159e+20
        exp-unsigned-signed 3.14159e-10
        exp-signed-signed -3.14159e-20
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
	readme-examples {
		title "Hello, World"
		bookmarks 12 15 188 1234
author "Alex Monad" email="alex@example.com" active=true
contents {
  section "First section" {
    paragraph "This is the first paragraph"
    paragraph "This is the second paragraph"
  }
}
node1; node2; node3;
node "this\nhas\tescapes"
other r"C:\Users\zkat\"
string "my
multiline
value"
other-raw r#"hello"world"#
num 1.234e-42
my-hex 0xdeadbeef
my-octal 0o755
my-binary 0b10101101
bignum 1_000_000
// C style

/*
C style multiline
*/

tag /*foo=true*/ bar=false

/*/*
hello
*/*/
// This entire node and its children are all commented out.
/-mynode "foo" key=1 {
  a
  b
  c
}

mynode /-"commented" "not commented" /-key="value" /-{
  a
  b
}
numbers (u8)10 (i32)20 myfloat=(f32)1.5 {
  strings (uuid)"123e4567-e89b-12d3-a456-426614174000" (date)"2021-02-03" filter=(regex)r"$\d+"
  (author)person name="Alex"
}
// Nodes can be separated into multiple lines
title \
  "Some title"


// Files must be utf8 encoded!
smile "😁"

// Instead of anonymous nodes, nodes and properties can be wrapped
// in "" for arbitrary node names.
"!@#$@$%Q#$%~@!40" "1.2.3" "!!!!!"=true

// The following is a legal bare identifier:
foo123~!@#$%^&*.:'|?+ "weeee"

// And you can also use unicode!
ノード　お名前="☜(ﾟヮﾟ☜)"

// kdl specifically allows properties and values to be
// interspersed with each other, much like CLI commands.
foo bar=true "baz" quux=false 1 2 3

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

func TestParser_ParseAll(t *testing.T) {

	testDoc := `

cat {
	name "Fernando"
	age 4
	eats fish chicken; color black white
}

dog {
	name "Maggie"
	age 14
	eats anything; color gray
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

`

	testDoc = kdlSchema

	s := tokenizer.NewSlice([]byte(testDoc))
	// s.Logger = tokenizer.SimpleLogger
	tokens, err := s.ScanAll()
	if err != nil {
		t.Fatalf("failed to tokenize: %v", err)
	}

	p := New()
	doc, err := p.ParseAll(tokens)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	fmt.Fprintf(os.Stderr, "%#v\n", doc)

	b := strings.Builder{}
	opts := generator.DefaultOptions
	opts.Indent = "    "
	g := generator.NewOptions(&b, opts)
	if err := g.Generate(doc); err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	// fmt.Fprintf(os.Stderr, "==== GENERATED ====\n%s\n", b.String())

	if testDoc != b.String() {
		t.Fatalf("want:\n%s\n got:\n%s\n", testDoc, b.String())
	}
	// d := diff.New([]byte(testDoc), []byte(b.String()))
	// fmt.Fprintf(os.Stderr, "==== DIFF ====\n%s\n", d.ANSIString())
}

var reSciNotFixup = regexp.MustCompile("([0-9.]+)[eE]([+-])")

func TestKDLOrgTestCases(t *testing.T) {
	testCases := loadTestCases()
	runTestCases(t, testCases, 0)
}

// we have to skip certain KDL test cases when NGINX syntax and YAML/TOML assignment modes are enabled because they
// will fail (or succeed) unexpectedly since they are not KDL
func skipTest(testCase string, relaxedFlag relaxed.Flags) bool {
	if relaxedFlag == 0 {
		return false
	}

	bn := filepath.Base(testCase)

	if relaxedFlag.Permit(relaxed.NGINXSyntax) {
		switch bn {
		case
			// invalid results because bare identifiers are allowed as arguments
			"bare_arg.kdl",
			"dash_dash.kdl",
			"false_prop_key.kdl",
			"null_prop_key.kdl",
			"true_prop_key.kdl",

			// invalid results because identifiers can contain special characters ?()./\ and (type) annotations
			// unsupported and treated as part of the identifier
			"arg_false_type.kdl",
			"arg_float_type.kdl",
			"arg_hex_type.kdl",
			"arg_null_type.kdl",
			"arg_raw_string_type.kdl",
			"arg_string_type.kdl",
			"arg_true_type.kdl",
			"arg_type.kdl",
			"arg_zero_type.kdl",
			"backslash_in_bare_id.kdl",
			"comment_in_arg_type.kdl",
			"comment_in_node_type.kdl",
			"comment_in_prop_type.kdl",
			"dot_zero.kdl",
			"just_space_in_prop_type.kdl",
			"just_type_no_node_id.kdl",
			"node_type.kdl",
			"parens_in_bare_id.kdl",
			"prop_false_type.kdl",
			"prop_string_type.kdl",
			"question_mark_at_start_of_int.kdl",
			"question_mark_before_number.kdl",
			"slash_in_bare_id.kdl",
			"quoted_prop_type.kdl",
			"space_in_node_type.kdl",
			"just_type_no_arg.kdl",
			"prop_true_type.kdl",
			"prop_zero_type.kdl",
			"prop_type.kdl",
			"multiple_dots_in_float.kdl",
			"just_type_no_prop.kdl",
			"multiple_es_in_float.kdl",
			"space_after_arg_type.kdl",
			"empty_prop_type.kdl",
			"prop_null_type.kdl",
			"empty_node_type.kdl",
			"comment_after_node_type.kdl",
			"prop_hex_type.kdl",
			"raw_arg_type.kdl",
			"quote_in_bare_id.kdl",
			"space_in_prop_type.kdl",
			"raw_prop_type.kdl",
			"space_after_node_type.kdl",
			"underscore_before_number.kdl",
			"prop_float_type.kdl",
			"underscore_at_start_of_int.kdl",
			"just_space_in_arg_type.kdl",
			"comment_after_arg_type.kdl",
			"empty_arg_type.kdl",
			"blank_prop_type.kdl",
			"space_after_prop_type.kdl",
			"just_space_in_node_type.kdl",
			"quoted_node_type.kdl",
			"raw_node_type.kdl",
			"blank_arg_type.kdl",
			"multiple_dots_in_float_before_exponent.kdl",
			"comment_after_prop_type.kdl",
			"quoted_arg_type.kdl",
			"space_in_arg_type.kdl",
			"type_before_prop_key.kdl",
			"blank_node_type.kdl",
			"dot_in_exponent.kdl",
			"prop_raw_string_type.kdl",

			// invalid results because continuations are unsupported
			"multiline_nodes.kdl",
			"escline.kdl",
			"escline_comment_node.kdl",
			"escline_line_comment.kdl",
			"slashdash_arg_before_newline_esc.kdl",
			"slashdash_arg_after_newline_esc.kdl":

			return true
		}

	}
	if relaxedFlag.Permit(relaxed.YAMLTOMLAssignments) {
		switch bn {
		case
			// invalid results because ':' is treated as whitespace
			"unusual_chars_in_bare_id.kdl":
			return true
		}
	}

	return false

}

func runTestCases(t *testing.T, testCases map[string]kdlTestCase, relaxedFlag relaxed.Flags) {
	out := strings.Builder{}
	opts := generator.DefaultOptions
	opts.Indent = "    "
	opts.IgnoreFlags = true // the expected_kdl documents expect basic formatting, whereas by default, kdl-go preserves the input formatting (hex input => hex output, etc)
	gen := generator.NewOptions(&out, opts)
	parser := New()

	for testCase, tc := range testCases {
		// if filepath.Base(testCase) != "dot_in_exponent.kdl" {
		// 	continue
		// }

		t.Run(filepath.Base(testCase), func(t *testing.T) {

			if skipTest(testCase, relaxedFlag) {
				return
			}
			out.Reset()

			wantErr := tc.expect == nil

			scanner := tokenizer.NewSlice(tc.input)
			scanner.RelaxedNonCompliant = relaxedFlag
			c := parser.NewContextOptions(ParseContextOptions{RelaxedNonCompliant: relaxedFlag})

			for scanner.Scan() {
				if err := parser.Parse(c, scanner.Token()); err != nil {
					if wantErr {
						// alles gute
						return
					}
					t.Fatalf("failed to parse: %v", err)
				}
			}
			if scanner.Err() != nil {
				// println("err = ", scanner.Err().Error())
				if wantErr {
					// alles gute
					return
				}
				t.Fatalf("failed to tokenize: %v", scanner.Err())
			}

			if err := gen.Generate(c.doc); err != nil {
				t.Fatalf("failed to generate output: %v", err)
			}

			if wantErr {
				t.Fatalf("successfully generated output that should have failed:\n%s", out.String())
			}

			output := out.String()

			bn := filepath.Base(testCase)
			switch bn {
			case "negative_exponent.kdl", "positive_exponent.kdl", "parse_all_arg_types.kdl", "underscore_in_exponent.kdl":
				// golang formats 1.0e-10 as 1E-10, which is valid per spec; we simply rewrite it as 1.0E-10 to match the expected result
				output = reSciNotFixup.ReplaceAllStringFunc(output, func(s string) string {
					matches := reSciNotFixup.FindStringSubmatch(s)
					if strings.IndexByte(matches[1], '.') == -1 {
						matches[1] += ".0"
					}
					return matches[1] + "E" + matches[2]
				})
			}

			if string(bytes.TrimSpace(tc.expect)) != strings.TrimSpace(output) {
				t.Fatalf("\nexpected: %s\ngot     : %s", string(tc.expect), output)
			}
		})
	}
}

func TestRelaxedNGINX(t *testing.T) {
	input := []byte(`
location / {
	# this is a comment
	root /var/www/html;
}
`)

	expect := []byte(`
location "/" {
    root "/var/www/html"
}
`)

	scanner := tokenizer.NewSlice(input)
	scanner.RelaxedNonCompliant = relaxed.NGINXSyntax | relaxed.MultiplierSuffixes

	t.Run("relaxedSyntax", func(t *testing.T) {
		tokens, err := scanner.ScanAll()
		if err != nil {
			t.Fatalf("failed to tokenize: %v", err)
		}

		p := New()
		c := p.NewContextOptions(ParseContextOptions{RelaxedNonCompliant: scanner.RelaxedNonCompliant})
		doc, err := p.ParseAllContext(c, tokens)
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		b := strings.Builder{}
		opts := generator.DefaultOptions
		opts.Indent = "    "
		g := generator.NewOptions(&b, opts)
		if err := g.Generate(doc); err != nil {
			t.Fatalf("failed to generate: %v", err)
		}

		// d := diff.New(bytes.TrimSpace(expect), bytes.TrimSpace([]byte(b.String())))
		// fmt.Fprintf(os.Stderr, "==== DIFF ====\n%s\n", d.ANSIString())

		want := string(bytes.TrimSpace(expect))
		got := strings.TrimSpace(b.String())
		if want != got {
			t.Fatalf("want:\n%s\n got:\n%s\n", want, got)
		}
	})

	// run the entire test suite in relaxed mode to make sure it doesn't interfere with standards-compliant documents
	testCases := loadTestCases()
	runTestCases(t, testCases, scanner.RelaxedNonCompliant)

}

func TestRelaxedYAMLTOML(t *testing.T) {
	input := []byte(`
yaml-like: 1234
toml-like=1234
toml-like-2 = 5678
`)

	expect := []byte(`
	yaml-like 1234
	toml-like 1234
	toml-like-2 5678
	`)

	scanner := tokenizer.NewSlice(input)
	scanner.RelaxedNonCompliant = relaxed.YAMLTOMLAssignments

	t.Run("relaxedSyntax", func(t *testing.T) {
		tokens, err := scanner.ScanAll()
		if err != nil {
			t.Fatalf("failed to tokenize: %v", err)
		}

		p := New()
		c := p.NewContextOptions(ParseContextOptions{RelaxedNonCompliant: scanner.RelaxedNonCompliant})
		doc, err := p.ParseAllContext(c, tokens)
		if err != nil {
			t.Fatalf("failed to parse: %v", err)
		}

		b := strings.Builder{}
		opts := generator.DefaultOptions
		opts.Indent = "    "
		g := generator.NewOptions(&b, opts)
		if err := g.Generate(doc); err != nil {
			t.Fatalf("failed to generate: %v", err)
		}

		// d := diff.New(bytes.TrimSpace(expect), bytes.TrimSpace([]byte(b.String())))
		// fmt.Fprintf(os.Stderr, "==== DIFF ====\n%s\n", d.ANSIString())

		want := string(bytes.TrimSpace(expect))
		got := strings.TrimSpace(b.String())
		if want != got {
			t.Fatalf("want:\n%s\n got:\n%s\n", want, got)
		}
	})

	// run the entire test suite in relaxed mode to make sure it doesn't interfere with standards-compliant documents
	testCases := loadTestCases()
	runTestCases(t, testCases, scanner.RelaxedNonCompliant)

}

type kdlTestCase struct {
	input  []byte
	expect []byte
}

func loadTestCases() map[string]kdlTestCase {
	cwd, _ := os.Getwd()
	testcasePath := filepath.Join(filepath.Dir(filepath.Dir(cwd)), "kdl-org", "kdl", "tests", "test_cases")
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
		expect, _ := os.ReadFile(expectedPath)

		testCases[testCase] = kdlTestCase{input, expect}
	}

	return testCases
}

func TestParserProfile(t *testing.T) {
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

	for i := 0; i < 2000; i++ {
		runTestCases(t, testCases, 0)
	}

}
