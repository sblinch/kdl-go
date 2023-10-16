package coerce

import (
	"encoding"
	"fmt"
	"math/big"
	"strconv"
	"time"
)

func ToComplex128(v interface{}) complex128 {
	switch x := v.(type) {
	case float32:
		return complex128(complex(x, 0))
	case float64:
		return complex(x, 0)
	case int:
		return complex128(complex(float64(x), 0))
	case int8:
		return complex128(complex(float64(x), 0))
	case int16:
		return complex128(complex(float64(x), 0))
	case int32:
		return complex128(complex(float64(x), 0))
	case int64:
		return complex128(complex(float64(x), 0))
	case uint:
		return complex128(complex(float64(x), 0))
	case uint8:
		return complex128(complex(float64(x), 0))
	case uint16:
		return complex128(complex(float64(x), 0))
	case uint32:
		return complex128(complex(float64(x), 0))
	case uint64:
		return complex128(complex(float64(x), 0))
	case time.Time:
		return complex(float64(x.Unix()), 0)

	case []byte:
		f, _ := strconv.ParseComplex(string(x), 128)
		return f
	case []rune:
		f, _ := strconv.ParseComplex(string(x), 128)
		return f
	case string:
		f, _ := strconv.ParseComplex(x, 128)
		return f
	case fmt.Stringer:
		s := x.String()
		f, _ := strconv.ParseComplex(s, 128)
		return f
	case encoding.TextMarshaler:
		b, _ := x.MarshalText()
		f, _ := strconv.ParseComplex(string(b), 128)
		return f
	case error:
		f, _ := strconv.ParseComplex(x.Error(), 128)
		return f

	case complex64:
		return complex128(x)
	case complex128:
		return x

	case *big.Int:
		// this seems inefficient, but not sure if there's a more efficient, equally accurate solution
		bf := big.NewFloat(0)
		_, _, _ = bf.Parse(x.String(), 10)
		f, _ := bf.Float64()
		return complex(f, 0)
	case *big.Float:
		f, _ := x.Float64()
		return complex(f, 0)

	default:
		if ToBool(v) {
			return 1
		} else {
			return 0
		}
	}
}
