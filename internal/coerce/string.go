package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func ToString(v interface{}) string {
	switch x := v.(type) {
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)

	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)

	case float32:
		return strconv.FormatFloat(float64(x), 'G', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(x), 'G', -1, 64)
	case time.Time:
		return x.Format(time.RFC3339)

	case []byte:
		return string(x)
	case []rune:
		return string(x)
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		return string(b)
	case error:
		return x.Error()

	case complex64:
		return strconv.FormatComplex(complex128(x), 'G', -1, 64)
	case complex128:
		return strconv.FormatComplex(x, 'G', -1, 128)

	case *big.Int:
		return x.String()
	case *big.Float:
		return x.String()

	case bool:
		if x {
			return "true"
		} else {
			return "false"
		}

	case nil:
		return "<nil>"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FromString attempts to identify a supported type in s, and converts to it
func FromString(s string) interface{} {
	switch s {
	case "true", "false":
		return ToBool(s)
	case "null":
		return nil
	default:
		if reNumeric.MatchString(s) {
			i, f, isInt := ToNumeric(s)
			if isInt {
				return i
			} else {
				return f
			}
		}

		// assume string
		return s
	}
}
