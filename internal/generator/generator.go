package generator

import (
	"io"

	"github.com/sblinch/kdl-go/document"
)

type Options struct {
	// Indent specifies the character(s) to use for indenting child nodes
	Indent string
	// IgnoreFlags causes certain formatting (such as hex/octal/binary styling for numbers, and raw/quoted/bare for
	// strings) from an input document to be ignored on output (if true, decimal is used for numbers, quoted for strings
	// unless bare is required)
	IgnoreFlags bool
	// AddSemicolons causes lines to be terminated with semicolons
	AddSemicolons bool
	// AddEquals causes '=' symbols to be inserted between nodes and their values, which is noncompliant with the KDL spec
	AddEquals bool
	// AddColon causes ':' symbols to be inserted between nodes and their values, which is noncompliant with the KDL spec
	AddColons bool
}

// Generator generates a KDL document from a parsed Document
type Generator struct {
	w       io.Writer
	depth   int
	options Options
}

// DefaultOptions sets the default options for a new Generator
var DefaultOptions = Options{
	Indent: "\t",
}

// NewOptions creates a new Generator with the provided Options, that writes to w
func NewOptions(w io.Writer, opts Options) *Generator {
	return &Generator{
		w:       w,
		options: opts,
	}
}

// New creates a new Generator with the default options, that writes to w
func New(w io.Writer) *Generator {
	return NewOptions(w, DefaultOptions)
}

// generateNode generates the KDL for a single Node (and its children by recursively calling itself) and returns a non-
// nil error on failure
func (g *Generator) generateNode(n *document.Node, leadingTrailingSpace, nameAndType bool) error {
	opts := document.NodeWriteOptions{
		LeadingTrailingSpace: leadingTrailingSpace,
		NameAndType:          nameAndType,
		Depth:                g.depth,
		Indent:               []byte(g.options.Indent),
		IgnoreFlags:          g.options.IgnoreFlags,
		AddSemicolons:        g.options.AddSemicolons,
		AddEquals:            g.options.AddEquals,
		AddColons:            g.options.AddColons,
	}
	_, err := n.WriteToOptions(g.w, opts)
	return err
}

// generateNodes generates the KDL for a slice of Nodes and returns a non-nil error on failure
func (g *Generator) generateNodes(nodes []*document.Node) error {
	opts := document.NodeWriteOptions{
		LeadingTrailingSpace: true,
		NameAndType:          true,
		Depth:                g.depth,
		Indent:               []byte(g.options.Indent),
		IgnoreFlags:          g.options.IgnoreFlags,
		AddSemicolons:        g.options.AddSemicolons,
		AddEquals:            g.options.AddEquals,
		AddColons:            g.options.AddColons,
	}

	for _, node := range nodes {
		if _, err := node.WriteToOptions(g.w, opts); err != nil {
			return err
		}
	}
	return nil
}

// Generate generates the KDL for a Document, and returns a non-nil error on failure
func (g *Generator) Generate(d *document.Document) error {
	return g.generateNodes(d.Nodes)
}
