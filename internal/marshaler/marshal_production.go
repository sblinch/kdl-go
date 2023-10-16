//go:build !kdldeterministic

package marshaler

import (
	"reflect"
)

func sortMapKeys(v []reflect.Value) []reflect.Value {
	return v
}
