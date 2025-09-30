//go:build go1.25

package marshaler

import (
	"reflect"
)

func IsType[T any](rv reflect.Value) bool {
	_, ok := reflect.TypeAssert[T](rv)
	return ok
}

func TypeAssert[T any](rv reflect.Value) (T, bool) {
	return reflect.TypeAssert[T](rv)
}
