package kdl

import (
	"bytes"
	"io"
	"reflect"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/internal/generator"
	"github.com/sblinch/kdl-go/internal/marshaler"
)

// Marshaler provides an interface for custom marshaling of a Go type into a Node
type Marshaler interface {
	MarshalKDL(node *document.Node) error
}

// ValueMarshaler provides an interface for custom marshaling of a Go type into a Value (such as a node argument or
// property)
type ValueMarshaler interface {
	MarshalKDLValue(value *document.Value) error
}

type MarshalerOptions = marshaler.MarshalOptions
type GeneratorOptions = generator.Options

type MarshalOptions struct {
	MarshalerOptions
	GeneratorOptions
}

// Encoder implements an encoder for KDL
type Encoder struct {
	w       io.Writer
	Options MarshalOptions
}

// Encode encodes v into KDL and writes it to the Encoder's writer, and returns a non-nil error on failure
func (e *Encoder) Encode(v interface{}) error {
	doc := document.New()
	if err := marshaler.MarshalWithOptions(v, doc, e.Options.MarshalerOptions); err != nil {
		return err
	}

	g := generator.NewOptions(e.w, e.Options.GeneratorOptions)
	if err := g.Generate(doc); err != nil {
		return err
	}

	return nil
}

// NewEncoder creates a new Encoder that writes to w
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
		Options: MarshalOptions{
			MarshalerOptions: marshaler.MarshalOptions{},
			GeneratorOptions: DefaultGenerateOptions,
		},
	}
}

// Marshal returns the KDL representation of v, or a non-nil error on failure
func Marshal(v interface{}) ([]byte, error) {
	doc := document.New()
	if err := marshaler.Marshal(v, doc); err != nil {
		return nil, err
	}

	b := bytes.Buffer{}
	g := generator.New(&b)
	if err := g.Generate(doc); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func MarshalDocument(v interface{}, doc *document.Document) error {
	return marshaler.Marshal(v, doc)
}

func MarshalDocumentWithOptions(v interface{}, doc *document.Document, opts marshaler.MarshalOptions) error {
	return marshaler.MarshalWithOptions(v, doc, opts)
}

func MarshalNode(v interface{}) (*document.Node, error) {
	return marshaler.MarshalNode(v)
}

func MarshalNodeWithOptions(v interface{}, opts marshaler.MarshalOptions) (*document.Node, error) {
	return marshaler.MarshalNodeWithOptions(v, opts)
}

func AddCustomMarshaler[T any](marshal func(v reflect.Value, node *document.Node) error) {
	marshaler.AddCustomMarshaler[T](marshal)
}

func AddCustomValueMarshaler[T any](marshal func(v reflect.Value, value *document.Value, format string) error) {
	marshaler.AddCustomValueMarshaler[T](marshal)
}
