package tokenizer

import (
	"fmt"
)

type TokenID int

const (
	Unknown TokenID = iota
	Newline
	Whitespace
	MultiLineComment
	SingleLineComment
	TokenComment
	TypeSpec
	Decimal
	Hexadecimal
	Octal
	Binary
	Boolean
	Null
	BareIdentifier
	SuffixedDecimal
	RawString
	QuotedString
	BraceOpen
	BraceClose
	ParensOpen
	ParensClose
	Equals
	Semicolon
	Continuation
	EOF

	ClassWhitespace
	ClassValue
	ClassIdentifier
	ClassNonStringValue
	ClassNumber
	ClassString
	ClassTerminator
	ClassEndOfLine
	ClassComment
)

var tokenClasses = map[TokenID][]TokenID{
	Newline:           {ClassTerminator, ClassWhitespace, ClassEndOfLine},
	Whitespace:        {ClassWhitespace},
	MultiLineComment:  {ClassComment},
	SingleLineComment: {ClassComment},
	TokenComment:      {},
	TypeSpec:          {},
	Decimal:           {ClassNumber, ClassValue, ClassNonStringValue},
	Hexadecimal:       {ClassNumber, ClassValue, ClassNonStringValue},
	Octal:             {ClassNumber, ClassValue, ClassNonStringValue},
	Binary:            {ClassNumber, ClassValue, ClassNonStringValue},
	SuffixedDecimal:   {ClassNumber, ClassValue, ClassNonStringValue},
	Boolean:           {ClassValue, ClassNonStringValue},
	Null:              {ClassValue, ClassNonStringValue},
	BareIdentifier:    {ClassValue, ClassIdentifier},
	RawString:         {ClassValue, ClassString, ClassIdentifier},
	QuotedString:      {ClassValue, ClassString, ClassIdentifier},
	BraceOpen:         {},
	BraceClose:        {},
	ParensOpen:        {},
	ParensClose:       {},
	Equals:            {},
	Semicolon:         {ClassTerminator},
	Continuation:      {},
	EOF:               {ClassTerminator, ClassEndOfLine},
}

func (t TokenID) Classes() []TokenID {
	return tokenClasses[t]
}

func (t TokenID) String() string {
	switch t {
	case Newline:
		return "Newline"
	case Whitespace:
		return "Whitespace"
	case MultiLineComment:
		return "MultiLineComment"
	case SingleLineComment:
		return "SingleLineComment"
	case TokenComment:
		return "TokenComment"
	case TypeSpec:
		return "TypeSpec"
	case Decimal:
		return "Decimal"
	case Hexadecimal:
		return "Hexadecimal"
	case Octal:
		return "Octal"
	case Binary:
		return "Binary"
	case Boolean:
		return "Boolean"
	case Null:
		return "Null"
	case BareIdentifier:
		return "BareIdentifier"
	case SuffixedDecimal:
		return "SuffixedDecimal"
	case RawString:
		return "RawString"
	case QuotedString:
		return "FormattedString"
	case BraceOpen:
		return "BraceOpen"
	case BraceClose:
		return "BraceClose"
	case ParensOpen:
		return "ParensOpen"
	case ParensClose:
		return "ParensClose"
	case Equals:
		return "Equals"
	case Semicolon:
		return "Semicolon"
	case Continuation:
		return "Continuation"
	case EOF:
		return "EOF"
	default:
		return "(invalid)"
	}
}

// Token contains a single token returned by a Scanner.
type Token struct {
	// ID indicates the token type
	ID TokenID
	// Data contains the literal data for the token; this may be a subslice of the input buffer (if the entire stream
	// could be read into a single buffer) or a copy of data from the input buffer, so it should not be modified.
	Data   []byte
	Line   int
	Column int
}

// String returns a string representation of the token for debugging
func (t Token) String() string {
	if len(t.Data) > 0 {
		return fmt.Sprintf("%s(%s)", t.ID.String(), string(t.Data))
	} else {
		return t.ID.String()
	}
}

// Valid returns true if this token has a valid ID
func (t Token) Valid() bool {
	return t.ID != Unknown
}

// Clear resets this token to its default (invalid) state
func (t *Token) Clear() {
	t.ID = Unknown
	t.Data = nil
	t.Line, t.Column = 0, 0
}
