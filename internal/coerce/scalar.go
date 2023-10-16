package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"regexp"
	"time"
)

func IsScalar(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, time.Time, []byte, []rune, string, fmt.Stringer, encoding.TextMarshaler, error, *big.Int, *big.Float:
		return true
	default:
		return false
	}
}

func IsScalarStrict(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, []byte, []rune, string:
		return true
	default:
		return false
	}
}

var (
	reNumeric = regexp.MustCompile("^[0-9]+(\\.[0-9]+)?([eE][0-9]+(\\.[0-9]+)?)?$")
	reInteger = regexp.MustCompile("^[0-9]+$")
)

func IsNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128, *big.Int, *big.Float:
		return true
	case []byte, []rune, string, fmt.Stringer, encoding.TextMarshaler:
		return reNumeric.MatchString(ToString(v))
	default:
		return false
	}
}

func IsInteger(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, *big.Int:
		return true
	case []byte, []rune, string, fmt.Stringer, encoding.TextMarshaler:
		return reInteger.MatchString(ToString(v))
	default:
		return false
	}
}
