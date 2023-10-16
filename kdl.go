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

// Parse parses a KDL document from r and returns the parsed Document, or a non-nil error on failure
func Parse(r io.Reader) (*document.Document, error) {
	s := tokenizer.New(r)
	return parse(s)
}

// Generate writes to w a well-formatted KDL document generated from doc, or a non-nil error on failure
func Generate(doc *document.Document, w io.Writer) error {
	g := generator.New(w)
	return g.Generate(doc)
}
