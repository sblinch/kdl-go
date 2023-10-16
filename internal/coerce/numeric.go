package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func stringToNumeric(s string) (i int64, f float64, isint bool) {
	if strings.IndexByte(s, '.') != -1 {
		f, _ := strconv.ParseFloat(s, 64)
		return 0, f, false
	} else {
		i, _ := strconv.ParseInt(s, 10, 64)
		return i, 0, true
	}
}

func ToNumeric(v interface{}) (i int64, f float64, isint bool) {
	switch x := v.(type) {
	case int:
		return int64(x), 0, true
	case int8:
		return int64(x), 0, true
	case int16:
		return int64(x), 0, true
	case int32:
		return int64(x), 0, true
	case int64:
		return x, 0, true
	case uint:
		return int64(x), 0, true
	case uint8:
		return int64(x), 0, true
	case uint16:
		return int64(x), 0, true
	case uint32:
		return int64(x), 0, true
	case uint64:
		return int64(x), 0, true
	case float32:
		return 0, float64(x), false
	case float64:
		return 0, x, false

	case time.Time:
		return x.Unix(), 0, true

	case []byte:
		return stringToNumeric(string(x))
	case []rune:
		return stringToNumeric(string(x))
	case string:
		return stringToNumeric(x)
	case fmt.Stringer:
		return stringToNumeric(x.String())
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		return stringToNumeric(string(b))
	case error:
		return stringToNumeric(x.Error())

	case complex64:
		return 0, float64(real(x)), false
	case complex128:
		return 0, real(x), false

	case *big.Int:
		return x.Int64(), 0, true
	case *big.Float:
		f, _ := x.Float64()
		return 0, f, false
	default:
		if ToBool(v) {
			return 1, 0, true
		} else {
			return 0, 0, true
		}
	}
}
