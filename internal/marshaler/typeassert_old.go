//go:build !go1.25

package marshaler

import (
	"reflect"
)

func IsType[T any](rv reflect.Value) bool {
	_, ok := rv.Interface().(T)
	return ok
}

func TypeAssert[T any](rv reflect.Value) (T, bool) {
	t, ok := rv.Interface().(T)
	return t, ok
}
