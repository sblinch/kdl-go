package document

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	// noEscapeTable maps each ASCII value to a boolean value indicating whether it does NOT require escapement
	noEscapeTable = [256]bool{}
	// hexTable maps each hexadecimal digit (0-9, a-f, and A-F) to its decimal value
	hexTable = [256]rune{}
)

func init() {
	// initialize the maps
	for i := 0; i <= 0x7e; i++ {
		noEscapeTable[i] = i >= 0x20 && i != '\\' && i != '"'
	}

	for r := '0'; r <= '9'; r++ {
		hexTable[r] = r - '0'
	}
	for r := 'a'; r <= 'f'; r++ {
		hexTable[r] = r - 'a' + 10
	}
	for r := 'A'; r <= 'F'; r++ {
		hexTable[r] = r - 'A' + 10
	}
}

// QuoteString returns s quoted for use as a KDL FormattedString
func QuoteString(s string) string {
	b := make([]byte, 0, len(s)*5/4)
	return string(AppendQuotedString(b, s, '"'))
}

// AppendQuotedString appends s, quoted for use as a KDL FormattedString, to b, and returns the expanded buffer.
//
// AppendQuotedString is based on the JSON string quoting function from the MIT-Licensed ZeroLog, Copyright (c) 2017
// Olivier Poitrey, but has been heavily modified to improve performance and use KDL string escapes instead of JSON.
func AppendQuotedString(b []byte, s string, quote byte) []byte {
	b = append(b, quote)

	// use uints for bounds-check elimination
	lenS := uint(len(s))
	// Loop through each character in the string.
	for i := uint(0); i < lenS; i++ {
		// Check if the character needs encoding. Control characters, slashes,
		// and the double quote need json encoding. Bytes above the ascii
		// boundary needs utf8 encoding.
		if !noEscapeTable[s[i]] {
			// We encountered a character that needs to be encoded. Switch
			// to complex version of the algorithm.

			start := uint(0)
			for i < lenS {
				c := s[i]
				if noEscapeTable[c] {
					i++
					continue
				}

				if c >= utf8.RuneSelf {
					r, size := utf8.DecodeRuneInString(s[i:])
					if r == utf8.RuneError && size == 1 {
						// In case of error, first append previous simple characters to
						// the byte slice if any and append a replacement character code
						// in place of the invalid sequence.
						if start < i {
							b = append(b, s[start:i]...)
						}
						b = append(b, `\ufffd`...)
						i += uint(size)
						start = i
						continue
					}
					i += uint(size)
					continue
				}

				// We encountered a character that needs to be encoded.
				// Let's append the previous simple characters to the byte slice
				// and switch our operation to read and encode the remainder
				// characters byte-by-byte.
				if start < i {
					b = append(b, s[start:i]...)
				}

				switch c {
				case quote, '\\', '/':
					b = append(b, '\\', c)
				case '\n':
					b = append(b, '\\', 'n')
				case '\r':
					b = append(b, '\\', 'r')
				case '\t':
					b = append(b, '\\', 't')
				case '\b':
					b = append(b, '\\', 'b')
				case '\f':
					b = append(b, '\\', 'f')
				default:
					b = append(b, '\\', 'u')
					b = strconv.AppendUint(b, uint64(c), 16)
				}
				i++
				start = i
			}
			if start < lenS {
				b = append(b, s[start:]...)
			}

			b = append(b, quote)
			return b
		}
	}
	// The string has no need for encoding an therefore is directly
	// appended to the byte slice.
	b = append(b, s...)
	b = append(b, quote)

	return b
}

const empty = ""

// UnquoteString returns s unquoted from KDL FormattedString notation
func UnquoteString(s string) (string, error) {
	if len(s) == 0 {
		return empty, nil
	}
	q := s[0]
	switch q {
	case '"', '\'':
	default:
		return "", ErrInvalid
	}

	b := make([]byte, 0, len(s))
	b, err := AppendUnquotedString(b, s, q)
	return string(b), err
}

var ErrInvalid = errors.New("invalid quoted string")

// AppendUnquotedString appends s, unquoted from KDL FormattedString notation, to b and returns the expanded buffer.
//
// AppendUnquotedString was originally based on the JSON string quoting function from the MIT-Licensed ZeroLog,
// Copyright (c) 2017 Olivier Poitrey, but has been heavily modified to unquote KDL quoted strings.
func AppendUnquotedString(b []byte, s string, quote byte) ([]byte, error) {
	if len(s) < 2 || s[0] != quote || s[len(s)-1] != quote {
		return nil, ErrInvalid
	}
	// remove quotes
	s = s[1 : len(s)-1]

	// use uints for bounds-check elimination
	lenS := uint(len(s))
	// Loop through each character in the string.
	for i := uint(0); i < lenS; i++ {
		c := s[i]
		// Check if the character needs decoding.
		if c == '\\' || c >= utf8.RuneSelf {
			// We encountered a character that needs to be decoded. Switch
			// to complex version of the algorithm.

			start := uint(0)
			for i < lenS {
				c := s[i]
				if !(c == '\\' || c >= utf8.RuneSelf) {
					i++
					continue
				}

				if c >= utf8.RuneSelf {
					r, size := utf8.DecodeRuneInString(s[i:])
					if r == utf8.RuneError && size == 1 {
						// In case of error, first append previous simple characters to
						// the byte slice if any and append a replacement character code
						// in place of the invalid sequence.
						if start < i {
							b = append(b, s[start:i]...)
						}
						b = append(b, `\ufffd`...)
						i += uint(size)
						start = i
						continue
					}
					i += uint(size)
					continue
				}

				// We encountered a character that needs to be decoded.
				// Let's append the previous simple characters to the byte slice
				// and switch our operation to read and encode the remainder
				// characters byte-by-byte.
				if start < i {
					b = append(b, s[start:i]...)
				}

				i++
				if i == lenS {
					return b, ErrInvalid
				}
				c = s[i]

				switch c {
				case 'n':
					b = append(b, '\n')
				case 'r':
					b = append(b, '\r')
				case 't':
					b = append(b, '\t')
				case 'b':
					b = append(b, '\b')
				case 'f':
					b = append(b, '\f')
				case 'u':
					// make sure we have enough room for `{n}`
					if i+3 >= lenS || s[i+1] != '{' {
						return b, ErrInvalid
					}
					i += 2

					// find the closing `}`
					rstart := i
					for i < lenS && s[i] != '}' {
						i++
					}
					if i >= lenS {
						return b, ErrInvalid
					}
					if i-rstart > 6 {
						return b, ErrInvalid
					}

					// convert the hex digits, working backwards
					r := rune(0)
					factor := rune(1)
					for j := i - 1; j >= rstart; j-- {
						r += hexTable[s[j]] * factor
						factor *= 16
					}
					if r > 0x10FFFF {
						return b, ErrInvalid
					}
					b = utf8.AppendRune(b, r)
				default:
					b = append(b, c)
				}
				i++
				start = i
			}
			if start < lenS {
				b = append(b, s[start:]...)
			}

			return b, nil
		}
	}

	// The string has no need for decoding an therefore is directly
	// appended to the byte slice.
	b = append(b, s...)

	return b, nil
}

func rawString(s string) string {
	b := make([]byte, 0, 1+8*2+len(s))
	return string(AppendRawString(b, s))
}

// AppendRawString appends s, quoted for use as a KDL RawString, to b and returns the expanded buffer.
func AppendRawString(b []byte, s string) []byte {
	// inelegant brute force approach because generation is not something I really care about at this point
	marker := append(make([]byte, 0, 64), '"')
	ok := false
	for i := 0; i < cap(marker); i++ {
		if !strings.Contains(s, string(marker)) {
			ok = true
			break
		}
		marker = append(marker, '#')
	}
	if !ok {
		marker = append(marker, "r\"invalid\""...)
		return marker
	}

	minSpace := 1 + len(marker)*2 + len(s)
	if cap(b)-len(b) < minSpace {
		n := make([]byte, 0, len(b)+minSpace)
		n = append(n, b...)
		b = n
	}
	b = append(b, 'r')
	for i := 0; i < len(marker)-1; i++ {
		b = append(b, '#')
	}
	b = append(b, '"')
	b = append(b, s...)
	b = append(b, marker...)
	return b
}
