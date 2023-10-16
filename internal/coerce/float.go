package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func ToFloat64(v interface{}) float64 {
	switch x := v.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case time.Time:
		return float64(x.Unix())
	case time.Duration:
		return x.Seconds()

	case []byte:
		f, _ := strconv.ParseFloat(string(x), 64)
		return f
	case []rune:
		f, _ := strconv.ParseFloat(string(x), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	case fmt.Stringer:
		s := x.String()
		f, _ := strconv.ParseFloat(s, 64)
		return f
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		f, _ := strconv.ParseFloat(string(b), 64)
		return f
	case error:
		f, _ := strconv.ParseFloat(x.Error(), 64)
		return f

	case complex64:
		return float64(real(x))
	case complex128:
		return real(x)

	case *big.Int:
		// this seems inefficient, but not sure if there's a more efficient, equally accurate solution
		bf := big.NewFloat(0)
		_, _, _ = bf.Parse(x.String(), 10)
		f, _ := bf.Float64()
		return f
	case *big.Float:
		f, _ := x.Float64()
		return f

	default:
		if ToBool(v) {
			return 1
		} else {
			return 0
		}
	}
}

func parseSuffixedFloat(s string) (float64, error) {
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
		return strconv.ParseFloat(s, 64)
	} else if n, err := strconv.ParseFloat(s[0:suffixIndex], 64); err != nil {
		return 0, err
	} else if multiplier, err := suffixToMultiplier(s[suffixIndex:]); err != nil {
		return 0, err
	} else {
		return n * float64(multiplier), nil
	}
}

func ToFloat64Suffix(v interface{}) float64 {
	switch x := v.(type) {
	case float32:
		return float64(x)
	case float64:
		return x
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case time.Time:
		return float64(x.Unix())
	case time.Duration:
		return x.Seconds()

	case []byte:
		f, _ := parseSuffixedFloat(string(x))
		return f
	case []rune:
		f, _ := parseSuffixedFloat(string(x))
		return f
	case string:
		f, _ := parseSuffixedFloat(x)
		return f
	case fmt.Stringer:
		s := x.String()
		f, _ := parseSuffixedFloat(s)
		return f
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		f, _ := parseSuffixedFloat(string(b))
		return f
	case error:
		f, _ := parseSuffixedFloat(x.Error())
		return f

	case complex64:
		return float64(real(x))
	case complex128:
		return real(x)

	case *big.Int:
		// this seems inefficient, but not sure if there's a more efficient, equally accurate solution
		bf := big.NewFloat(0)
		_, _, _ = bf.Parse(x.String(), 10)
		f, _ := bf.Float64()
		return f
	case *big.Float:
		f, _ := x.Float64()
		return f

	default:
		if ToBool(v) {
			return 1
		} else {
			return 0
		}
	}
}
