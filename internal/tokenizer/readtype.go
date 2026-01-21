package tokenizer

import (
	"fmt"
	"io"

	"github.com/sblinch/kdl-go/relaxed"
)

// readWhitespace reads all whitespace starting from the current position. It does not return an error as in practice it
// is only called after r.peek() has already been invoked and returned a whitespace character, and thus at least one
// whitespace character will always be available.
func (s *Scanner) readWhitespace() []byte {
	ws, _ := s.readWhile(isWhiteSpace, 1)
	return ws
}

// skipWhitespace skips zero or more whitespace characters from the current position, and returns a non-nil error on
// failure
func (s *Scanner) skipWhitespace() error {
	_, err := s.readWhile(isWhiteSpace, 0)
	return err
}

// readMultiLineComment reads and returns a multiline comment from the current position, supporting nested /* and */
// sequences. It returns a non-nil error on failure.
func (s *Scanner) readMultiLineComment() ([]byte, error) {
	s.pushMark()
	defer s.popMark()

	depth := 0
	for {
		c, err := s.get()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}

		switch c {
		case '*':
			if next, err := s.peek(); err == nil && next == '/' {
				depth--
				s.skip()

				if depth == 0 {
					return s.copyFromMark(), nil
				}
			}

		case '/':
			if next, err := s.peek(); err == nil && next == '*' {
				depth++
				s.skip()
			}
		}
	}
}

// skipUntilNewline skips all characters from the current position until the next newline. It returns a non-nil error on
// failure.
func (s *Scanner) skipUntilNewline() error {
	escaped := false
	for {
		c, err := s.get()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		switch c {
		case '\\':
			escaped = true
			if err := s.skipWhitespace(); err != nil {
				return err
			}
		case '\r':
			// swallow error on peek, as it's still a valid newline if \r is not followed by \n
			if c, err := s.peek(); err == nil && c == '\n' {
				s.skip()
			}
			if escaped {
				escaped = false
			} else {
				return nil
			}

		case '\n', '\u0085', '\u000c', '\u2028', '\u2029':
			if escaped {
				escaped = false
			} else {
				return nil
			}
		default:
			escaped = false
		}
	}
}

// readSingleLineComment reads and returns a single-line comment from the current position, or a non-nil error on
// failure.
func (s *Scanner) readSingleLineComment() ([]byte, error) {
	literal, err := s.readUntil(isNewline, false)
	if err == io.ErrUnexpectedEOF {
		err = nil
	}
	return literal, err
}

// readRawString reads and returns a raw string from the input, or returns a non-nil error on failure
func (s *Scanner) readRawString() ([]byte, error) {
	s.pushMark()
	defer s.popMark()

	var (
		c   rune
		err error
	)

	startHashes := 0

	if c, err = s.get(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if c != 'r' {
		return nil, fmt.Errorf("unexpected character %c", c)
	}

hashLoop:
	for {
		if c, err = s.get(); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		switch c {
		case '"':
			break hashLoop
		case '#':
			startHashes++
		default:
			return nil, fmt.Errorf("unexpected character %c", c)
		}
	}

	foundQuote := false
	endHashes := 0
	for {
		if c, err = s.get(); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return nil, err
		}
		if foundQuote {
			if c == '#' {
				endHashes++
				if endHashes == startHashes {
					return s.copyFromMark(), nil
				}
			} else if c == '"' {
				endHashes = 0
			} else {
				foundQuote = false
				endHashes = 0
			}
		} else if c == '"' {
			foundQuote = true
			if startHashes == 0 {
				return s.copyFromMark(), nil
			}
		}
	}
}

func (s *Scanner) readQuotedString() ([]byte, error) {
	return s.readQuotedStringQ('"')
}

func (s *Scanner) readSingleQuotedString() ([]byte, error) {
	return s.readQuotedStringQ('\'')
}

// readQuotedString reads and returns a quoted string from the current position, or returns a non-nil error on failure.
func (s *Scanner) readQuotedStringQ(q rune) ([]byte, error) {
	var (
		c   rune
		err error
	)
	if c, err = s.peek(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if c != q {
		return nil, fmt.Errorf("unexpected character %c", c)
	}

	escaped := false
	done := false
	first := true
	return s.readWhile(func(c rune) bool {
		if first {
			// skip "
			first = false
			return true
		}
		if done {
			return false
		}
		switch c {
		case '\\':
			escaped = !escaped
		case q:
			if escaped {
				escaped = false
			} else {
				done = true
			}
		default:
			if escaped {
				escaped = false
			}
		}
		return true
	}, 2)

}

// readBareIdentifier reads a bare identifier from the current position and returns a TokenID representing its type
// (either BareIdentifier, Boolean, or Null), the byte sequence for the identifier, and a non-nil error on failure
func (s *Scanner) readBareIdentifier() (TokenID, []byte, error) {
	var (
		c   rune
		err error
	)

	if c, err = s.peek(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}

		return Unknown, nil, err
	}

	switch c {
	case '+', '-':
		if _, c, err = s.peekTwo(); err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return Unknown, nil, err
		}
		if isDigit(c) {
			return Unknown, nil, fmt.Errorf("unexpected character %c", c)
		}
	default:
		if !isBareIdentifierStartChar(c, s.RelaxedNonCompliant) {
			return Unknown, nil, fmt.Errorf("unexpected character %c", c)
		}
	}

	var literal []byte

	isBareIdentifierCharClosure := func(c rune) bool {
		return isBareIdentifierChar(c, s.RelaxedNonCompliant)
	}

	if literal, err = s.readWhile(isBareIdentifierCharClosure, 1); err != nil {
		return Unknown, nil, err
	}
	tokenType := BareIdentifier

	if string(literal) == "true" || string(literal) == "false" {
		tokenType = Boolean
	} else if string(literal) == "null" {
		tokenType = Null
	}

	return tokenType, literal, nil
}

// readIdentifier reads an identifier from the current position and returns a TokenID representing the identifier's
// type, a byte sequence representeing the identifier, and a non-nil error on failure
func (s *Scanner) readIdentifier() (TokenID, []byte, error) {
	c, err := s.peek()
	if err != nil {
		return Unknown, nil, err
	}

	if c <= 0x20 || c > 0x10FFFF {
		return Unknown, nil, fmt.Errorf("unexpected character %c", c)
	}

	// r.log("reading an identifier", "start-with", string(c), "second-char", string(c2))
	switch c {
	case 'r':
		// r.log("maybe a raw string", "second-char", string(c2))
		_, c2, err := s.peekTwo()
		if err == nil && (c2 == '#' || c2 == '"') {
			// r.log("fo sho a raw string, reading")
			literal, err := s.readRawString()
			return RawString, literal, err
		} else {
			// r.log("must be a bare identifier")
			// possible bare identifier starting with 'r'
			tokenType, literal, err := s.readBareIdentifier()
			return tokenType, literal, err
		}

	case '"':
		s.log("quoted string, reading")
		literal, err := s.readQuotedString()
		return QuotedString, literal, err

	case '{', '}', '<', '>', ';', '[', ']', '=', ',':
		return Unknown, nil, fmt.Errorf("unexpected character %c", c)

	case '\\', '(', ')', '.', '_', '?', '/':
		if !s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			// some of these arent actually forbidden by the spec, but the test cases indicate that they should be disallowed
			return Unknown, nil, fmt.Errorf("unexpected character %c", c)
		}

	case '\'':
		if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			s.log("single quoted string, reading")
			literal, err := s.readSingleQuotedString()
			return QuotedString, literal, err
		}
	}

	_, c2, err := s.peekTwo()
	if err == nil && !isBareIdentifierStartChar(c, s.RelaxedNonCompliant) && !(c == '-' && !isDigit(c2)) {
		s.log("not a valid bare identifier")
		return Unknown, nil, fmt.Errorf("unexpected character %c", c)
	}

	s.log("bare identifier, reading")
	tokenType, literal, err := s.readBareIdentifier()
	return tokenType, literal, err
}

// readInteger reads and returns an integer from the current position, or a non-nil error on failure
func (s *Scanner) readInteger() (TokenID, []byte, error) {
	tokenID := Decimal

	first := true
	validRune := func(c rune) bool {
		if first {
			first = false
			return isDigit(c) // cannot start with _
		}
		return isDigit(c) || c == '_'
	}

	hasMultiplier := false
	if s.RelaxedNonCompliant.Permit(relaxed.MultiplierSuffixes) {
		multiplierOK := true
		validRune = func(c rune) bool {
			if first {
				first = false
				if c == '+' || c == '-' {
					multiplierOK = false
					return true
				}
			}

			if multiplierOK {
				switch c {
				case 'h', 'm', 's', 'u', 'µ', 'k', 'K', 'M', 'g', 'G', 't', 'T', 'b':
					hasMultiplier = true
					return true
				}
			}

			return isDigit(c) || c == '_'
		}
	}
	data, err := s.readWhile(validRune, 1)

	if hasMultiplier {
		tokenID = SuffixedDecimal
	}

	return tokenID, data, err
}

// readSignedInteger reads and returns a signed integer from the current position, or a non-nil error on failure
func (s *Scanner) readSignedInteger() (TokenID, []byte, error) {
	s.pushMark()
	defer s.popMark()

	c, err := s.peek()
	if err != nil {
		return Unknown, nil, err
	}

	if c == '+' || c == '-' {
		s.skip()
	}

	tokenID, _, err := s.readInteger()
	return tokenID, s.copyFromMark(), err
}

// readDecimal reads and returns a a decimal value (either an integer or a floating point number) from the current
// position, or a non-nil error on failure
func (s *Scanner) readDecimal() (TokenID, []byte, error) {
	s.pushMark()
	defer s.popMark()

	tokenID, _, err := s.readSignedInteger()
	if err != nil {
		s.log("reading decimal: failed", "error", err)
		return tokenID, nil, err
	}

	// ignore any error at this point because we've already successfully read the initial signed integer
	// r.log("reading decimal: peeky")
	if c, err := s.peek(); err == nil {
		if c == '.' {
			s.skip()

			// r.log("reading decimal: unsigned integer")
			if tokenID, _, err = s.readInteger(); err != nil {
				s.log("reading decimal: failed", "error", err)
				return tokenID, nil, err
			}
		}

		// again, ignore any error
		if c, err := s.peek(); err == nil {
			if c == 'e' || c == 'E' {
				s.skip()
				// r.log("reading decimal: signed integer")
				if tokenID, _, err := s.readSignedInteger(); err != nil {
					s.log("reading decimal: failed", "error", err)
					return tokenID, nil, err
				}
			}
		}
	}

	if c, err := s.peek(); err == nil && !isSeparator(c) {
		if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) && isBareIdentifierChar(c, s.RelaxedNonCompliant) {
			// it's not actually a numeric identifier; parse as a bare string

			isBareIdentifierCharClosure := func(c rune) bool {
				return isBareIdentifierChar(c, s.RelaxedNonCompliant)
			}

			if _, err = s.readWhile(isBareIdentifierCharClosure, 1); err != nil {
				return Unknown, nil, err
			}

			tokenID = BareIdentifier
		} else {
			return tokenID, nil, fmt.Errorf("unexpected character %c", c)
		}
	}

	return tokenID, s.copyFromMark(), nil
}

// readNumericBase reads and returns a binary, octal, or hexadecimal number from the current position, ensuring that it
// is at least 3 characters in length (eg: 0xN), followed by whitespace or a newline, and that all characters are valid;
// returns a non-nil error on failure
func (s *Scanner) readNumericBase(valid func(c rune) bool) ([]byte, error) {
	lit, err := s.readWhile(valid, 3)
	if err == nil && lit[2] == '_' {
		// disallow 0x_
		return nil, fmt.Errorf("unexpected character _")
	}
	if err == nil {
		if c, err := s.peek(); err == nil && !isWhiteSpace(c) && !isNewline(c) {
			return nil, fmt.Errorf("unexpected character %c", c)
		}
	}
	return lit, err
}

// readHexadecimal reads and returns a hexadecimal number from the current position, or a non-nil error on failure
func (s *Scanner) readHexadecimal() ([]byte, error) {
	n := 0
	return s.readNumericBase(func(c rune) bool {
		if n < 2 {
			// skip 0x
			n++
			return true
		}
		return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == '_'
	})
}

// readOctal reads and returns an octal number from the current position, or a non-nil error on failure
func (s *Scanner) readOctal() ([]byte, error) {
	n := 0
	return s.readNumericBase(func(c rune) bool {
		if n < 2 {
			// skip 0o
			n++
			return true
		}
		return (c >= '0' && c <= '7') || c == '_'
	})
}

// readBinary reads and returns a binary number from the current position, or a non-nil error on failure
func (s *Scanner) readBinary() ([]byte, error) {
	n := 0
	return s.readNumericBase(func(c rune) bool {
		if n < 2 {
			// skip 0b
			n++
			return true
		}
		return c == '0' || c == '1' || c == '_'
	})
}
