package document

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

type SuffixedDecimal struct {
	Number []byte
	Suffix []byte
}

func (s SuffixedDecimal) AsDuration() (time.Duration, error) {
	if len(s.Number) > 0 {
		return time.ParseDuration(s.String())
	} else {
		return time.ParseDuration(string(s.Suffix))
	}
}

func (s SuffixedDecimal) String() string {
	b := make([]byte, 0, len(s.Number)+len(s.Suffix))
	b = append(b, s.Number...)
	b = append(b, s.Suffix...)
	return string(b)
}

func (s SuffixedDecimal) AsNumber() (interface{}, error) {
	n, err := parseNumber(s.Number, 10)
	if err != nil {
		return 0, fmt.Errorf("suffixed decimal: %w", err)
	}
	unit := float64(1000)

	switch len(s.Suffix) {
	case 0:
		return n, nil
	case 1:
	case 2:
		if s.Suffix[1] == 'b' || s.Suffix[1] == 'B' {
			unit = 1024
		} else {
			return 0, fmt.Errorf("invalid suffix: %s", string(s.Suffix))
		}
	default:
		return 0, fmt.Errorf("invalid suffix: %s", string(s.Suffix))
	}

	multiplier := float64(1)
	switch s.Suffix[0] {
	case 'k', 'K':
		multiplier = unit
	case 'm', 'M':
		multiplier = unit * unit
	case 'g', 'G':
		multiplier = unit * unit * unit
	case 't', 'T':
		multiplier = unit * unit * unit * unit
	default:
		return 0, fmt.Errorf("invalid suffix: %s", string(s.Suffix))
	}

	switch v := n.(type) {
	case int64:
		return float64(v) * multiplier, nil
	case float64:
		return v * multiplier, nil
	case *big.Int:
		bf := big.NewFloat(0)
		bf.SetInt(v)
		m := big.NewFloat(multiplier)
		z := big.NewFloat(0)
		return z.Mul(bf, m), nil
	case *big.Float:
		m := big.NewFloat(multiplier)
		z := big.NewFloat(0)
		return z.Mul(v, m), nil
	default:
		return 0, nil
	}
}

func ParseSuffixedDecimal(b []byte) (SuffixedDecimal, error) {
	duration := false
	suffixIndex := -1

	for i, c := range b {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		case 'h', 's', 'u', 'µ':
			if suffixIndex == -1 {
				suffixIndex = i
			}
			duration = true
			break
		case 'k', 'K', 'm', 'M', 'g', 'G', 't', 'T', 'b', 'B':
			if suffixIndex == -1 {
				suffixIndex = i
			} else if c != 'b' {
				duration = true
				break
			}
		default:
			return SuffixedDecimal{}, fmt.Errorf("unexpected character %c in suffixed decimal value", c)
		}
	}

	if suffixIndex == 0 {
		return SuffixedDecimal{}, errors.New("suffixed decimal starts with non-digit")

	} else if duration { // mixed digits and letters
		return SuffixedDecimal{Number: nil, Suffix: b}, nil

	} else if suffixIndex == -1 { // all digits
		return SuffixedDecimal{Number: b, Suffix: []byte{}}, nil

	} else {
		return SuffixedDecimal{Number: b[0:suffixIndex], Suffix: b[suffixIndex:]}, nil
	}
}
