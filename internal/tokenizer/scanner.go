package tokenizer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/sblinch/kdl-go/relaxed"
)

var (
	ErrInvalidRune = errors.New("invalid UTF8 input")
	ErrEndOfToken  = errors.New("unexpected end of token")
)

// markedReaderPosition represents a copy of the Scanner.input slice at a given position in the stream
type markedReaderPosition int

// peeked represents a character peeked from the input buffer and the number of bytes it occupied in the buffer
type peeked struct {
	c    rune
	size int
}

// Scanner implements a scanner for tokenizing a KDL input stream
type Scanner struct {
	Logger              func(string, ...interface{})
	raw                 []byte
	input               []byte
	len                 int
	peeked              []peeked
	line                int
	column              int
	token               Token
	err                 error
	marks               []int
	Alt                 bool
	RelaxedNonCompliant relaxed.Flags
	r                   io.Reader
}

// log records a log message if a logger has been configured
func (s *Scanner) log(msg string, v ...interface{}) {
	if s.Logger != nil {
		s.Logger(msg, v...)
	}
}

// peek returns the next character from the input bufer without consuming it; returns an non-nil error on failure
func (s *Scanner) peek() (rune, error) {
	c, _, err := s.peekSize()
	return c, err
}

// peekSize returns the next character and its size from the input buffer without consuming it; returns a non-nil error
// on failure
func (s *Scanner) peekSize() (rune, int, error) {
	if len(s.peeked) == 0 {
		c, size := utf8.DecodeRune(s.input)
		if size == 0 {
			return 0, 0, io.EOF
		}
		if c == utf8.RuneError {
			return 0, 0, ErrInvalidRune
		}
		s.peeked = s.peeked[0:1]
		s.peeked[0].c = c
		s.peeked[0].size = size
		return c, size, nil
	}

	p := s.peeked[0]
	return p.c, p.size, nil
}

// peekTwo is analogous to peek(), but returns the next two characters from the input buffer
func (s *Scanner) peekTwo() (rune, rune, error) {
	c1, _, c2, _, err := s.peekTwoSize()
	return c1, c2, err
}

// peekTwoSize is analogous to peekSize(), but returns the next two characters from the input buffer and their sizes
func (s *Scanner) peekTwoSize() (rune, int, rune, int, error) {
	// r.log("peekTwo: peek buffer length is", "len", len(r.peek))
	peekedSize := 0
	if len(s.peeked) == 1 {
		peekedSize = s.peeked[0].size
	}
	for len(s.peeked) < 2 {
		c, size := utf8.DecodeRune(s.input[peekedSize:])
		if size == 0 {
			return 0, 0, 0, 0, io.EOF
		}
		if c == utf8.RuneError {
			return 0, 0, 0, 0, ErrInvalidRune
		}
		peekedSize += size
		// r.log("peekTwo: peeked", "c", string(c))
		n := len(s.peeked)
		s.peeked = s.peeked[0 : n+1]
		s.peeked[n].c = c
		s.peeked[n].size = size
	}

	// r.log("peekTwo: returning", "c1", string(r.peek[0].c), "c2", string(r.peek[1].c))
	return s.peeked[0].c, s.peeked[0].size, s.peeked[1].c, s.peeked[1].size, nil
}

// get consumes and returns the next character from the input buffer; returns a non-nil error on failure
func (s *Scanner) get() (rune, error) {
	if s.len <= utf8.UTFMax*2 {
		s.refill()
	}

	var (
		c    rune
		size int
	)
	if len(s.peeked) > 0 {
		size = s.peeked[0].size
		c = s.peeked[0].c
		if len(s.peeked) == 1 {
			s.peeked = s.peeked[:0]
		} else {
			s.peeked[0] = s.peeked[1]
			s.peeked = s.peeked[0:1]
		}
	} else {
		c, size = utf8.DecodeRune(s.input)
		if size == 0 {
			return 0, io.EOF
		}
		if c == utf8.RuneError {
			return 0, ErrInvalidRune
		}
	}

	if isNewline(c) {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	s.input = s.input[size:]
	s.len -= size

	return c, nil
}

// copyInput returns size bytes from the beginning of s.input.
//
// If input is being streamed from an io.Reader and more data needs to be read from the reader, copyInput returns a copy
// of the bytes.
// If input is not being streamed from a reader, or no further data needs to be read from the reader, copyInput returns
// a subslice of the input buffer to avoid unnecessary memory allocations.
func (s *Scanner) copyInput(size int) []byte {
	if s.r == nil {
		return s.input[0:size]
	} else {
		return append(make([]byte, 0, size), s.input[0:size]...)
	}
}

// refill refills the input buffer (when streaming from an io.Reader) when the input buffer is nearly empty
func (s *Scanner) refill() {
	// if no reader is assigned, we have nothing to do
	if s.r == nil {
		return
	}

	raw := s.raw[0:cap(s.raw)]

	inputOffset := 0
	unreadLen := len(s.input)
	retainLen := unreadLen
	if len(s.marks) > 0 {
		// if any marks have been made, we need to preserve the buffer contents starting from the oldest mark
		oldestMark := s.marks[0]
		copy(raw, s.raw[oldestMark:])
		retainLen = len(s.raw) - oldestMark
		inputOffset = retainLen - unreadLen

		// if it appears that a marked sequence (typically a single token) may end up being larger than our buffer, we
		// have no choice but to enlarge the buffer
		if retainLen > cap(s.raw)*3/4 {
			b := make([]byte, cap(s.raw)*2)
			copy(b, raw)
			raw = b
		}

		for k, m := range s.marks {
			s.marks[k] = m - oldestMark
		}
	} else {
		// otherwise we only need to preserve the remaining bytes in the input buffer
		copy(raw, s.input)
	}

	// fill the remainder of the input buffer from the reader
	remain := raw[retainLen:]
	nr, err := io.ReadFull(s.r, remain)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			s.err = err
			s.r = nil
			// fmt.Fprintf(os.Stderr, "REFILL: failed with error %v\n", err)
			return
		}
		s.r = nil
	}

	s.len = unreadLen + nr
	s.raw = raw[0 : retainLen+nr]
	s.input = s.raw[inputOffset:]
	// fmt.Fprintf(os.Stderr, "REFILL: after=%q\n", string(s.input))
}

// skip consumes the next character from the input buffer
func (s *Scanner) skip() {
	_, _ = s.get()
}

// pushMark pushes the current position in the input buffer onto the mark stack, for use with copyFromMark
func (s *Scanner) pushMark() {
	pos := len(s.raw) - len(s.input)
	s.marks = append(s.marks, pos)
	return
}

// popMark pops the last marked position from the mark stack
func (s *Scanner) popMark() {
	if len(s.marks) > 0 {
		s.marks = s.marks[0 : len(s.marks)-1]
	}
}

// copyFromMark returns a slice of bytes starting from the most recently marked position and ending at the current
// input buffer position
func (s *Scanner) copyFromMark() []byte {
	p := s.marks[len(s.marks)-1]
	newPos := len(s.raw) - len(s.input)
	r := s.raw[p:newPos]
	return r
}

// readWhile reads a rune from the input buffer and passes it to validRune, repeating for as long as validRune returns
// true; returns a slice of bytes from the input buffer on success, or a non-nil error on failure or if fewer than
// minLength runes are read from the buffer
func (s *Scanner) readWhile(validRune func(c rune) bool, minLength int) ([]byte, error) {
	s.pushMark()
	defer s.popMark()

	valid := false
	for {
		c, err := s.peek()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err != io.EOF && validRune(c) {
			s.skip()
			valid = true
		} else {
			if !valid {
				return nil, fmt.Errorf("unexpected character %c", c)
			}
			literal := s.copyFromMark()
			if len(literal) < minLength {
				return nil, ErrEndOfToken
			}
			return literal, nil
		}
	}
}

// readUntil reads a rune from the input buffer and passes it to stop, repeating for as long as stop returns false;
// returns a slice of bytes from the input buffer (including the final rune which caused stop to return alse if
// includeStopChar is true) on success, or a non-nil error on failure;
func (s *Scanner) readUntil(stop func(c rune) bool, includeStopChar bool) ([]byte, error) {
	s.pushMark()
	defer s.popMark()
	for {
		c, err := s.peek()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return s.copyFromMark(), err
		}

		if !stop(c) {
			s.skip()
		} else {
			if includeStopChar {
				s.skip()
			}
			return s.copyFromMark(), nil
		}
	}
}

// Offset returns the current byte offset in s.raw
func (s *Scanner) Offset() int {
	return len(s.raw) - len(s.input)
}

type staticScanner struct {
	s  Scanner
	mu sync.Mutex
}

var staticScan = staticScanner{
	s: Scanner{
		peeked: make([]peeked, 0, 2),
		marks:  make([]int, 0, 8),
	},
}

func ScanOne(b []byte) (Token, error) {
	staticScan.mu.Lock()
	defer staticScan.mu.Unlock()

	staticScan.s.raw = b
	staticScan.s.input = b
	staticScan.s.len = len(b)
	staticScan.s.peeked = staticScan.s.peeked[:0]
	staticScan.s.marks = staticScan.s.marks[:0]
	return staticScan.s.readNext()
}

// readNext reads and returns the next token from the input buffer, or a non-nil error on failure
func (s *Scanner) readNext() (Token, error) {
	token := Token{
		Line:   s.line,
		Column: s.column,
	}

	c, size, err := s.peekSize()
	if err != nil {
		return token, err
	}

	ignore := true

	switch c {
	case
		// unicode-space
		'\t', ' ',
		'\u00A0',
		'\u1680',
		'\u2000',
		'\u2001',
		'\u2002',
		'\u2003',
		'\u2004',
		'\u2005',
		'\u2006',
		'\u2007',
		'\u2008',
		'\u2009',
		'\u200A',
		'\u202F',
		'\u205F',
		'\u3000',
		// BOM
		'\uFEFF':
		s.log("reading whitespace")
		token.ID = Whitespace
		token.Data = s.readWhitespace()
		s.log("read whitespace")

	case '\r':
		s.log("reading carriage return")
		token.ID = Newline

		if _, _, nextC, nextSize, err := s.peekTwoSize(); err == nil && nextC == '\n' {
			token.Data = s.copyInput(size + nextSize)
			s.skip()
			s.skip()
		} else {
			token.Data = s.copyInput(size)
			s.skip()
		}

	case '\n', '\u0085', '\u000c', '\u2028', '\u2029':
		s.log("reading newline")
		token.ID = Newline
		token.Data = s.copyInput(size)
		s.skip()

	case '/':
		_, c, err := s.peekTwo()
		if err != nil {
			return token, err
		}
		// s.log("reading potential comment", "c2", string(c))

		switch c {
		case '*':
			s.log("reading multiline comment")
			token.ID = MultiLineComment
			token.Data, err = s.readMultiLineComment()
			if err != nil {
				return token, err
			}
		case '/':
			s.log("reading single line comment")
			token.ID = SingleLineComment
			token.Data, err = s.readSingleLineComment()
			if err != nil {
				return token, err
			}
		case '-':
			token.ID = TokenComment
			token.Data = s.copyInput(2)
			s.skip()
			s.skip()

		default:
			if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
				s.log("reading identifier")
				if token.ID, token.Data, err = s.readIdentifier(); err != nil {
					return token, err
				}
			} else {
				return token, fmt.Errorf("unexpected character %c", c)
			}
		}

	case '(':
		if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			ignore = false
		} else {
			s.log("reading open paren")
			token.ID = ParensOpen
			token.Data = s.copyInput(1)
			s.skip()
		}

	case ')':
		if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			ignore = false
		} else {
			s.log("reading close paren")
			token.ID = ParensClose
			token.Data = s.copyInput(1)
			s.skip()
		}

	case '{':
		s.log("reading open brace")
		token.ID = BraceOpen
		token.Data = s.copyInput(1)
		s.skip()
	case '}':
		s.log("reading close brace")
		token.ID = BraceClose
		token.Data = s.copyInput(1)
		s.skip()
	case '=':
		s.log("reading equals")
		token.ID = Equals
		token.Data = s.copyInput(1)
		s.skip()
	case ';':
		s.log("reading semicolon")
		token.ID = Semicolon
		token.Data = s.copyInput(1)
		s.skip()
	case '\\':
		if s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			ignore = false
		} else {
			s.log("reading continuation")
			token.ID = Continuation
			token.Data = s.copyInput(1)
			s.skip()
		}

	case '+', '-': // sign
		s.log("reading signed value")
		_, c, err := s.peekTwo()
		if err != nil {
			s.log("oh noes")
			return token, err
		}
		if isDigit(c) {
			if token.ID, token.Data, err = s.readDecimal(); err != nil {
				s.log("decimal oh noes")
				return token, err
			}
		} else {
			if token.ID, token.Data, err = s.readIdentifier(); err != nil {
				return token, err
			}
		}

	case '0':
		s.log("reading value starting with 0")
		if _, c, err := s.peekTwo(); err != nil {
			return token, err
		} else {
			switch c {
			case 'x':
				token.ID = Hexadecimal
				if token.Data, err = s.readHexadecimal(); err != nil {
					return token, err
				}

			case 'o':
				token.ID = Octal
				if token.Data, err = s.readOctal(); err != nil {
					return token, err
				}

			case 'b':
				token.ID = Binary
				if token.Data, err = s.readBinary(); err != nil {
					return token, err
				}

			default:
				if token.ID, token.Data, err = s.readDecimal(); err != nil {
					return token, err
				}
			}
		}

	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.log("reading decimal value")
		if token.ID, token.Data, err = s.readDecimal(); err != nil {
			return token, err
		}

	default:
		if c == '#' && s.RelaxedNonCompliant.Permit(relaxed.NGINXSyntax) {
			s.log("reading single line comment")
			token.ID = SingleLineComment
			token.Data, err = s.readSingleLineComment()
			if err != nil {
				return token, err
			}
		} else if c == ':' && s.RelaxedNonCompliant.Permit(relaxed.YAMLTOMLAssignments) {
			s.log("reading colon")
			token.ID = Whitespace
			token.Data = s.copyInput(1)
			s.skip()
		} else {
			ignore = false
		}

	}

	if !ignore {
		s.log("reading identifier")
		if token.ID, token.Data, err = s.readIdentifier(); err != nil {
			return token, err
		}
	}

	if s.Logger != nil {
		s.log("got token", "token", token)
	}
	return token, nil
}

// Pos returns the current line and column number from the input buffer; the current implementation is best-effort only
// and may be approximate but is usually fairly accurate
func (s *Scanner) Pos() (int, int) {
	return s.line + 1, s.column + 1
}

// extractLineAtOffset returns a string containing the line at the specified offset in the input buffer, a newline, and
// a caret positioned to indicate the current position in the input buffer
func (s *Scanner) extractLineAtOffset(offset int) string {
	start := offset
	for start > 0 {
		start--
		if isNewline(rune(s.raw[start])) {
			start++
			break
		}
	}

	end := offset
	for end < len(s.raw)-1 {
		end++
		if isNewline(rune(s.raw[end])) {
			break
		}
	}
	caretOffset := offset - start + 1

	elided := caretOffset > 64
	if elided {
		start += caretOffset - 64
		caretOffset = 64
	}

	line := make([]byte, 0, end-start+1+1+caretOffset)
	line = append(line, s.raw[start:end]...)
	for i, c := range line {
		if c == '\t' {
			line[i] = ' '
		}
	}

	if elided {
		line[0], line[1], line[2] = '.', '.', '.'
	}

	line = append(line, '\n')
	for i := 0; i < caretOffset-1; i++ {
		line = append(line, ' ')
	}
	line = append(line, '^')

	return string(line)
}

// annotatedError annotates err with the input line/column and positionSummary from the input buffer
func (s *Scanner) annotatedError(err error) error {
	line, column := s.Pos()
	return fmt.Errorf("scan failed: %w at line %d, column %d\n%s", err, line, column, s.extractLineAtOffset(len(s.raw)-len(s.input)))
}

// SimpleLogger provides a simple logger that writes to stderr; this can be assigned to Scanner.Logger for debugging
func SimpleLogger(s string, v ...interface{}) {
	b := strings.Builder{}
	b.WriteString(s)
	if len(v) > 0 {
		b.WriteByte('\t')
		key := true
		for _, x := range v {
			if fs, ok := x.(fmt.Stringer); ok {
				b.WriteString(fs.String())
			} else {
				fmt.Fprintf(&b, "%v", x)
			}
			if key {
				b.WriteByte('=')
			} else {
				b.WriteByte(' ')
			}
			key = !key
		}
	}
	fmt.Fprintln(os.Stderr, b.String())
}

func newScanner() *Scanner {
	return &Scanner{
		Logger: nil,
		peeked: make([]peeked, 0, 2),
		marks:  make([]int, 0, 8),
	}
}

// NewSlice creates a new Scanner that reads from input
func NewSlice(input []byte) *Scanner {
	s := newScanner()
	s.input = input
	s.raw = input
	s.len = len(input)

	return s
}

// NewBuffer creates a new scanner that reads from r, using a preallocated buffer b
func NewBuffer(r io.Reader, b []byte) *Scanner {
	s := newScanner()

	nr, err := io.ReadFull(r, b)
	if err != nil {
		if err != io.ErrUnexpectedEOF {
			s.err = err
			return s
		}
		// nothing more to read; don't retain the reader
		r = nil
	}

	s.r = r
	s.input = b[0:nr]
	s.raw = s.input
	s.len = nr

	return s
}

var DefaultBufferSize = 64 * 1024

// New creates a new Scanner that reads from r
func New(r io.Reader) *Scanner {
	b := make([]byte, DefaultBufferSize)
	return NewBuffer(r, b)
}

// ScanAll is a convenience function that scans and returns all tokens; a non-nil error is returned on failure
func (s *Scanner) ScanAll() ([]Token, error) {
	if s.err != nil {
		return nil, s.err
	}

	tokens := make([]Token, 0, len(s.input)/2)
	for {
		if token, err := s.readNext(); err == nil {
			tokens = append(tokens, token)
		} else if err == io.EOF {
			break
		} else {
			s.log("failed", "error", err)
			return nil, s.annotatedError(err)
		}
	}

	return tokens, nil

}

var eofToken = Token{
	ID:   EOF,
	Data: []byte{},
}

// Scan reads the next token from the input stream and returns true if a token was read, otherwise false.
//
// If Scan returns false, Err will return an error indicating the nature of the failure. On EOF, Scan will return
// false and Err will return nil.
//
// If Scan returns true, a token was read and will be returned by Token.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		if s.err == io.EOF {
			s.err = nil
		}
		return false
	}

	if s.token, s.err = s.readNext(); s.err == nil {
		return true
	} else if s.err == io.EOF {
		s.token = eofToken
		return true
	} else {
		s.err = s.annotatedError(s.err)
		return false
	}
}

// Token returns the token read by Scan
func (s *Scanner) Token() Token {
	return s.token
}

// Err returns the error encountered by Scan
func (s *Scanner) Err() error {
	return s.err
}

var ErrClosed = errors.New("use of closed Scanner")

// Close closes the scanner and releases its resources
func (s *Scanner) Close() error {
	s.input = nil
	s.raw = nil
	s.r = nil
	s.marks = nil
	s.peeked = nil
	s.err = ErrClosed
	return nil
}
