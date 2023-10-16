package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func ToInt64(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return int64(x)
	case float64:
		return int64(x)
	case time.Time:
		return x.Unix()
	case time.Duration:
		return int64(x.Seconds())

	case []byte:
		i, _ := strconv.ParseInt(string(x), 10, 64)
		return i
	case []rune:
		i, _ := strconv.ParseInt(string(x), 10, 64)
		return i
	case string:
		i, _ := strconv.ParseInt(x, 10, 64)
		return i
	case fmt.Stringer:
		s := x.String()
		i, _ := strconv.ParseInt(s, 10, 64)
		return i
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		i, _ := strconv.ParseInt(string(b), 10, 64)
		return i
	case error:
		i, _ := strconv.ParseInt(x.Error(), 10, 64)
		return i

	case complex64:
		return int64(real(x))
	case complex128:
		return int64(real(x))

	case *big.Int:
		return x.Int64()
	case *big.Float:
		i, _ := x.Int64()
		return i

	default:
		if ToBool(v) {
			return 1
		} else {
			return 0
		}
	}
}

func ToByte(v interface{}) byte {
	return byte(ToInt64(v))
}

var numericUnits = map[bool]map[uint8]int64{
	true:  {'k': 1e3, 'm': 1e6, 'g': 1e9, 't': 1e12, 'p': 1e15, 'e': 1e18},                      // si
	false: {'k': 1 << 10, 'm': 1 << 20, 'g': 1 << 30, 't': 1 << 40, 'p': 1 << 50, 'e': 1 << 60}, // iec
}

func suffixToMultiplier(suffix string) (int64, error) {
	si := true

	switch len(suffix) {
	case 0:
		return 1, nil
	case 1:
	case 2:
		if suffix[1] == 'b' || suffix[1] == 'B' {
			si = false
		} else {
			return 0, fmt.Errorf("invalid suffix: %s", suffix)
		}
	default:
		return 0, fmt.Errorf("invalid suffix: %s", suffix)
	}

	u := suffix[0]
	if u >= 'A' && u <= 'Z' {
		u += 32
	}

	return numericUnits[si][u], nil
}

func parseSuffixedInt(s string) (int64, error) {
	invalid := false
	suffixIndex := -1
	for i, c := range s {
		if (c >= '0' && c <= '9') || c == '.' {
			if suffixIndex != -1 {
				invalid = true
				break
			}
		} else if suffixIndex == -1 {
			suffixIndex = i
		}
	}
	if suffixIndex <= 0 {
		invalid = true
	}
	if invalid {
		return strconv.ParseInt(s, 10, 64)
	} else if n, err := strconv.ParseFloat(s[0:suffixIndex], 64); err != nil {
		return 0, err
	} else if multiplier, err := suffixToMultiplier(s[suffixIndex:]); err != nil {
		return 0, err
	} else {
		return int64(n * float64(multiplier)), nil
	}
}

func ToInt64Suffix(v interface{}) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return int64(x)
	case float64:
		return int64(x)
	case time.Time:
		return x.Unix()
	case time.Duration:
		return int64(x.Seconds())

	case []byte:
		i, _ := parseSuffixedInt(string(x))
		return i
	case []rune:
		i, _ := parseSuffixedInt(string(x))
		return i
	case string:
		i, _ := parseSuffixedInt(x)
		return i
	case fmt.Stringer:
		s := x.String()
		i, _ := parseSuffixedInt(s)
		return i
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		i, _ := parseSuffixedInt(string(b))
		return i
	case error:
		i, _ := parseSuffixedInt(x.Error())
		return i

	case complex64:
		return int64(real(x))
	case complex128:
		return int64(real(x))

	case *big.Int:
		return x.Int64()
	case *big.Float:
		i, _ := x.Int64()
		return i

	default:
		if ToBool(v) {
			return 1
		} else {
			return 0
		}
	}
}
