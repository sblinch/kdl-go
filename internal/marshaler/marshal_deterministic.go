//go:build kdldeterministic

package marshaler

import (
	"reflect"
	"sort"

	"github.com/sblinch/kdl-go/internal/coerce"
)

func sortMapKeys(v []reflect.Value) []reflect.Value {
	ss := make(map[reflect.Value]string)
	for _, rv := range v {
		ss[rv] = coerce.ToString(rv.Interface())
	}

	sort.SliceStable(v, func(i, j int) bool {
		vi := v[i]
		vj := v[j]
		return ss[vi] < ss[vj]
	})

	return v
}
