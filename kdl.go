package kdl

import (
	"io"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/internal/generator"
	"github.com/sblinch/kdl-go/internal/parser"
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

func parse(s *tokenizer.Scanner) (*document.Document, error) {
	defer s.Close()

	p := parser.New()
	c := p.NewContextOptions(parser.ParseContextOptions{RelaxedNonCompliant: s.RelaxedNonCompliant})
	for s.Scan() {
		if err := p.Parse(c, s.Token()); err != nil {
			return nil, err
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	return c.Document(), nil
}

type ParseOptions = parser.ParseContextOptions

var DefaultParseOptions = parser.ParseContextOptions{}

// Parse parses a KDL document from r and returns the parsed Document, or a non-nil error on failure
func Parse(r io.Reader) (*document.Document, error) {
	return ParseWithOptions(r, DefaultParseOptions)
}

func ParseWithOptions(r io.Reader, opts ParseOptions) (*document.Document, error) {
	s := tokenizer.New(r)
	s.RelaxedNonCompliant = opts.RelaxedNonCompliant
	return parse(s)
}

type GenerateOptions = generator.Options

var DefaultGenerateOptions = generator.DefaultOptions

// Generate writes to w a well-formatted KDL document generated from doc, or a non-nil error on failure
func Generate(doc *document.Document, w io.Writer) error {
	return GenerateWithOptions(doc, w, DefaultGenerateOptions)
}

// Generate writes to w a well-formatted KDL document generated from doc, or a non-nil error on failure
func GenerateWithOptions(doc *document.Document, w io.Writer, opts GenerateOptions) error {
	g := generator.NewOptions(w, opts)
	return g.Generate(doc)
}
