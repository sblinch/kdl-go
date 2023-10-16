package tokenizer

import (
	"github.com/sblinch/kdl-go/relaxed"
)

// isWhiteSpace returns true if c is a whitespace character
func isWhiteSpace(c rune) bool {
	switch c {
	case // unicode-space
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
		return true
	default:
		return false
	}
}

// isNewline returns true if c is a newline character
func isNewline(c rune) bool {
	switch c {
	case '\r', '\n', '\u0085', '\u000c', '\u2028', '\u2029':
		return true
	default:
		return false
	}
}

// isLineSpace returns true if c is a whitespace or newline character
func isLineSpace(c rune) bool {
	return isWhiteSpace(c) || isNewline(c)
}

// isDigit returns true if c is a digit
func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

// isSign returns true if c is + or -
func isSign(c rune) bool {
	return c == '-' || c == '+'
}

// isSeparator returns true if c whitespace, a newline, or a semicolon
func isSeparator(c rune) bool {
	return isWhiteSpace(c) || isNewline(c) || c == ';'
}

// isBareIdentifierStartChar indicates whether c is a valid first character for a bare identifier. Note that this
// returns true if c is + or -, in which case the second character must not be a digit.
func isBareIdentifierStartChar(c rune, r relaxed.Flags) bool {
	if !isBareIdentifierChar(c, r) {
		return false
	}
	if isDigit(c) {
		return false
	}

	return true
}

// isBareIdentifierChar indicates whether c is a valid character for a bare identifier
func isBareIdentifierChar(c rune, r relaxed.Flags) bool {
	if isLineSpace(c) {
		return false
	}
	if c <= 0x20 || c > 0x10FFFF {
		return false
	}
	switch c {
	case '{', '}', '<', '>', ';', '[', ']', '=', ',':
		return false
	case '(', ')', '/', '\\', '"':
		return r.Permit(relaxed.NGINXSyntax)
	case ':':
		return !r.Permit(relaxed.YAMLTOMLAssignments)
	default:
		return true
	}
}

// IsBareIdentifier returns true if s contains a valid BareIdentifier (a string that requires no quoting in KDL)
func IsBareIdentifier(s string, rf relaxed.Flags) bool {
	if len(s) == 0 {
		return false
	}

	first := true
	for _, r := range s {
		if first {
			if !isBareIdentifierStartChar(r, rf) {
				return false
			}
			first = false
		} else {
			if !isBareIdentifierChar(r, rf) {
				return false
			}
		}
	}
	return true
}
