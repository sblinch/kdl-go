package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"time"
)

func boolString(s string) bool {
	ls := len(s)
	switch ls {
	case 1: // 1, y(es), or t(rue)
		return (s[0] > '0' && s[0] <= '9') || s[0] == 'y' || s[0] == 'Y' || s[0] == 't' || s[0] == 'T'
	case 2: // no, or any other 2-letter string
		return false
	case 3: // yes
		return s[0] == 'y' || s[0] == 'Y'
	case 4: // true
		return s[0] == 't' || s[0] == 'T'
	case 5: // false, or any other 5-letter string
		return false
	default:
		return false
	}
}

func ToBool(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int8:
		return x != 0
	case int16:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case uint:
		return x != 0
	case uint8:
		return x != 0
	case uint16:
		return x != 0
	case uint32:
		return x != 0
	case uint64:
		return x != 0
	case float32:
		return x != 0
	case float64:
		return x != 0
	case error:
		return x != nil
	case []byte:
		return boolString(string(x))
	case []rune:
		return boolString(string(x))
	case string:
		return boolString(x)
	case fmt.Stringer:
		return boolString(x.String())
	case nil:
		return false

	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		return boolString(string(b))
	case time.Time:
		return !x.IsZero()

	case complex64:
		return x != 0
	case complex128:
		return x != 0

	case *big.Int:
		return x.Int64() != 0
	case *big.Float:
		f, _ := x.Float64()
		return f != 0
	default:
		return false
	}
}
