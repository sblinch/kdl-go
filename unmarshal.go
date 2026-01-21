package kdl

import (
	"io"
	"reflect"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/internal/marshaler"
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

type UnmarshalOptions = marshaler.UnmarshalOptions

// Unmarshaler provides an interface for custom unmarshaling of a node into a Go type
type Unmarshaler interface {
	UnmarshalKDL(node *document.Node) error
}

// ValueUnmarshaler provides an interface for custom unmarshaling of a Value (such as a node argument or property) into
// a Go type
type ValueUnmarshaler interface {
	UnmarshalKDLValue(value *document.Value) error
}

// Decoder implements a decoder for KDL
type Decoder struct {
	r       io.Reader
	Options marshaler.UnmarshalOptions
}

// Decode decodes KDL from the Decoder's reader into v; v must contain a pointer type. Returns a non-nil error on
// failure.
func (d *Decoder) Decode(v interface{}) error {
	s := tokenizer.New(d.r)
	s.RelaxedNonCompliant = d.Options.RelaxedNonCompliant
	s.ParseComments = d.Options.ParseComments
	if doc, err := parse(s); err != nil {
		return err
	} else {
		return marshaler.UnmarshalWithOptions(doc, v, d.Options)
	}
}

// NewDecoder returns a Decoder that reads from r
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Unmarshal unmarshals KDL from data into v; v must contain a pointer type. Returns a non-nil error on failure.
func Unmarshal(data []byte, v interface{}) error {
	s := tokenizer.NewSlice(data)
	if doc, err := parse(s); err != nil {
		return err
	} else {
		return marshaler.Unmarshal(doc, v)
	}
}

// UnmarshalWithOptions unmarshals KDL from data into v with the specified options; v must contain a pointer type.
// Returns a non-nil error on failure.
func UnmarshalWithOptions(data []byte, v interface{}, opts UnmarshalOptions) error {
	s := tokenizer.NewSlice(data)
	s.RelaxedNonCompliant = opts.RelaxedNonCompliant
	s.ParseComments = opts.ParseComments
	if doc, err := parse(s); err != nil {
		return err
	} else {
		return marshaler.Unmarshal(doc, v)
	}
}

func UnmarshalDocument(doc *document.Document, v interface{}) error {
	return marshaler.Unmarshal(doc, v)
}

func UnmarshalDocumentWithOptions(doc *document.Document, v interface{}, opts UnmarshalOptions) error {
	return marshaler.UnmarshalWithOptions(doc, v, opts)
}

func UnmarshalNode(node *document.Node, v interface{}) error {
	return marshaler.UnmarshalNode(node, v)
}

func UnmarshalNodeWithOptions(node *document.Node, v interface{}, opts UnmarshalOptions) error {
	return marshaler.UnmarshalNodeWithOptions(node, v, opts)
}

func AddCustomUnmarshaler[T any](unmarshal func(node *document.Node, v reflect.Value) error) {
	marshaler.AddCustomUnmarshaler[T](unmarshal)
}

func AddCustomValueUnmarshaler[T any](unmarshal func(value *document.Value, v reflect.Value, format string) error) {
	marshaler.AddCustomValueUnmarshaler[T](unmarshal)
}
