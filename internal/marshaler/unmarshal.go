package marshaler

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/internal/coerce"
	"github.com/sblinch/kdl-go/relaxed"
)

type unmarshaler interface {
	UnmarshalKDL(node *document.Node) error
}
type valueUnmarshaler interface {
	UnmarshalKDLValue(value *document.Value) error
}

type UnmarshalOptions struct {
	AllowUnhandledNodes    bool
	AllowUnhandledArgs     bool
	AllowUnhandledProps    bool
	AllowUnhandledChildren bool
	RelaxedNonCompliant    relaxed.Flags
}

type unmarshalContext struct {
	indexer *typeIndexer
	opts    UnmarshalOptions
}

// inStrSlice returns true if s is contained in ss
func inStrSlice(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// verifyArgsPropsChildren returns an error if the node does not contain the expected number of args or the expected properties
func verifyArgsPropsChildren(c *unmarshalContext, node *document.Node, expectedArgs int, expectedProps []string, allowChildren bool) error {
	argCount := len(node.Arguments)
	if argCount < expectedArgs || (!c.opts.AllowUnhandledArgs && argCount > expectedArgs) {
		return fmt.Errorf("%s expects %d argument(s), %d provided", node.Name.ValueString(), expectedArgs, argCount)
	}

	if len(expectedProps) > 0 {
		var missingProps []string
		if node.Properties.Len() == 0 {
			missingProps = append(missingProps, expectedProps...)
		} else {
			for _, expectedPropname := range expectedProps {
				if _, ok := node.Properties.Get(expectedPropname); !ok {
					missingProps = append(missingProps, expectedPropname)
				}
			}
		}
		if len(missingProps) > 0 {
			return fmt.Errorf("%s is missing properties %s", node.Name.ValueString(), strings.Join(missingProps, ", "))
		}

		if !c.opts.AllowUnhandledProps && node.Properties.Len() > 0 {
			var extraProps []string
			for name := range node.Properties.Unordered() {
				if !inStrSlice(expectedProps, name) {
					extraProps = append(extraProps, name)
				}
			}
			if len(extraProps) > 0 {
				return fmt.Errorf("%s has unexpected properties %s", node.Name.ValueString(), strings.Join(extraProps, ", "))
			}
		}

	} else if !c.opts.AllowUnhandledProps && node.Properties.Len() > 0 {
		var extraProps []string
		for name := range node.Properties.Unordered() {
			extraProps = append(extraProps, name)
		}

		return fmt.Errorf("%s has unexpected properties %s", node.Name.ValueString(), strings.Join(extraProps, ", "))
	}

	if !allowChildren && len(node.Children) > 0 {
		return fmt.Errorf("%s has unexpected children", node.Name.ValueString())
	}

	return nil
}

func callStructMethod(structValue reflect.Value, methodIndex int, args ...reflect.Value) ([]reflect.Value, error) {
	valuePtr := structValue

	createdPtr := false
	if valuePtr.Kind() != reflect.Pointer {
		valuePtr = reflect.New(structValue.Type())
		valuePtr.Elem().Set(structValue)
		createdPtr = true
	} else if valuePtr.IsNil() {
		valuePtr = reflect.New(structValue.Type().Elem())
		structValue.Set(valuePtr)
	}

	method := valuePtr.Method(methodIndex)

	ret := method.Call(args)

	if createdPtr && structValue.CanAddr() {
		structValue.Set(valuePtr.Elem())
	}

	var err error
	for i := len(ret) - 1; i >= 0; i-- {
		v := ret[i]
		if v.Type().Name() == "error" {
			intf := v.Interface()
			if e, ok := intf.(error); ok {
				err = e
			} else {
				err = nil
			}
			break
		}
	}

	return ret, err
}

func unmarshalIntf(c *unmarshalContext, dest reflect.Value, v interface{}, format string) (bool, error) {
	if typeDetails := c.indexer.Get(dest.Type().String()); typeDetails != nil {
		if typeDetails.CanUnmarshalKDLValue() {
			v := &document.Value{Value: v}
			_, err := callStructMethod(dest, typeDetails.KDLValueUnmarshalerMethod, reflect.ValueOf(v))
			return true, err

		} else if typeDetails.CanUnmarshalText() {
			_, err := callStructMethod(dest, typeDetails.TextUnmarshalerMethod, reflect.ValueOf([]byte(coerce.ToString(v))))
			return true, err
		}
	}
	return false, nil
}

func timeFormat(format string) string {
	fmtStr := ""
	switch format {
	case "ANSIC":
		fmtStr = time.ANSIC
	case "UnixDate":
		fmtStr = time.UnixDate
	case "RubyDate":
		fmtStr = time.RubyDate
	case "RFC822":
		fmtStr = time.RFC822
	case "RFC822Z":
		fmtStr = time.RFC822Z
	case "RFC850":
		fmtStr = time.RFC850
	case "RFC1123":
		fmtStr = time.RFC1123
	case "RFC1123Z":
		fmtStr = time.RFC1123Z
	case "RFC3339":
		fmtStr = time.RFC3339
	case "RFC3339Nano":
		fmtStr = time.RFC3339Nano
	case "Kitchen":
		fmtStr = time.Kitchen
	case "Stamp":
		fmtStr = time.Stamp
	case "StampMilli":
		fmtStr = time.StampMilli
	case "StampMicro":
		fmtStr = time.StampMicro
	case "StampNano":
		fmtStr = time.StampNano
	default:
		if len(format) >= 2 && format[0] == '\'' && format[len(format)-1] == '\'' {
			fmtStr = format[1 : len(format)-1]
		}
	}
	return fmtStr
}

func unmarshalValueTime(c *unmarshalContext, destTime reflect.Value, val interface{}, format string) error {
	switch format {
	case "unix":
		t := time.Unix(coerce.ToInt64(val), 0)
		destTime.Set(reflect.ValueOf(t))
		return nil
	case "unixmilli":
		t := time.UnixMilli(coerce.ToInt64(val))
		destTime.Set(reflect.ValueOf(t))
		return nil
	case "unixmicro":
		t := time.UnixMicro(coerce.ToInt64(val))
		destTime.Set(reflect.ValueOf(t))
		return nil
	case "unixnano":
		t := time.Unix(0, coerce.ToInt64(val))
		destTime.Set(reflect.ValueOf(t))
		return nil
	default:
		if fmtStr := timeFormat(format); fmtStr != "" {
			t, err := time.Parse(fmtStr, coerce.ToString(val))
			if err != nil {
				return err
			}
			destTime.Set(reflect.ValueOf(t))
			return nil
		} else {
			return fmt.Errorf("invalid format string: %s", format)
		}
	}
}

func unmarshalNode(c *unmarshalContext, dest reflect.Value, node *document.Node, format string) (bool, reflect.Value, error) {
	if typeDetails := c.indexer.Get(dest.Type().String()); typeDetails != nil {
		if typeDetails.CanUnmarshalKDL() {
			_, err := callStructMethod(dest, typeDetails.KDLUnmarshalerMethod, reflect.ValueOf(node))
			return true, dest, err
		} else if len(node.Arguments) == 1 && (typeDetails.CanUnmarshalText() || typeDetails.CanUnmarshalKDLValue()) {
			dest, err := setReflectValueFromIntf(c, dest, node.Arguments[0].ResolvedValue(), format)
			return true, dest, err
		}
	}

	return false, dest, nil
}

func handleFormatIntf(c *unmarshalContext, dest reflect.Value, iv interface{}, format string) (newDest reflect.Value, newVal interface{}, done bool, e error) {
	if format == "" {
		return dest, iv, false, nil
	}

	switch dest.Type().String() {
	case "time.Time":
		return dest, iv, true, unmarshalValueTime(c, dest, iv, format)

	case "time.Duration":
		return dest, iv, true, unmarshalValueDuration(c, dest, iv, format)

	case "float32", "float64":
		return dest, iv, false, nil
	}

	return dest, iv, true, fmt.Errorf("invalid format string: %s", format)
}

func parseHMSDuration(hms string) (time.Duration, error) {
	var (
		s, remain  string
		d          time.Duration
		ok, havens bool
	)

	s, remain, ok = strings.Cut(hms, ":")
	if ok {
		if n, err := strconv.ParseInt(s, 10, 32); err != nil {
			return 0, err
		} else {
			d += time.Hour * time.Duration(n)
		}
		s, remain, ok = strings.Cut(remain, ":")
	}
	if ok {
		if n, err := strconv.ParseInt(s, 10, 32); err != nil {
			return 0, err
		} else {
			d += time.Minute * time.Duration(n)
		}
		s, remain, havens = strings.Cut(remain, ".")

		ok := len(s) > 0
		if ok {
			if n, err := strconv.ParseInt(s, 10, 32); err != nil {
				return 0, err
			} else {
				d += time.Second * time.Duration(n)
			}
			if havens {
				if n, err := strconv.ParseInt(remain, 10, 32); err != nil {
					return 0, err
				} else {
					d += time.Nanosecond * time.Duration(n)
				}
			}
		}
	}
	if ok {
		return d, nil
	} else {
		return 0, fmt.Errorf("invalid H:MM:SS.SSSSSSSSS format: %s", s)
	}
}

var durationFormatFactor = map[string]time.Duration{
	"sec":   time.Second,
	"milli": time.Millisecond,
	"micro": time.Microsecond,
	"nano":  time.Nanosecond,
}

func unmarshalValueDuration(c *unmarshalContext, dest reflect.Value, iv interface{}, format string) error {
	var (
		d   time.Duration
		err error
	)

	if factor, ok := durationFormatFactor[format]; ok {
		i, f, isInt := coerce.ToNumeric(iv)

		if isInt {
			d = time.Duration(i) * factor
		} else {
			d = time.Duration(f * float64(factor))
		}
	} else {
		switch iv.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			switch format {
			case "base60":
				return errors.New("cannot unmarshal numeric value from base60 format")
			default:
				d = time.Duration(coerce.ToInt64(iv)) * time.Second
			}
		case float32, float64:
			switch format {
			case "base60":
				return errors.New("cannot unmarshal numeric value from base60 format")
			default:
				d = time.Duration(coerce.ToFloat64(iv) * float64(time.Second))
			}
		case document.SuffixedDecimal:
			if d, err = iv.(document.SuffixedDecimal).AsDuration(); err != nil {
				return err
			}
		default:
			switch format {
			case "base60":
				if d, err = parseHMSDuration(coerce.ToString(iv)); err != nil {
					return err
				}
			default:
				if d, err = time.ParseDuration(coerce.ToString(iv)); err != nil {
					return err
				}
			}
		}
	}

	dest.Set(reflect.ValueOf(d))
	return nil
}

func resolveSuffixedDecimal(rv *reflect.Value, val interface{}) (interface{}, error) {
	var err error

	sd, isSuffixed := val.(document.SuffixedDecimal)
	if !isSuffixed {
		// if val is not already a SuffixedDecimal, see if it is a string that contains a valid suffixed decimal
		if s, ok := val.(string); ok {
			if sd, err = document.ParseSuffixedDecimal([]byte(s)); err != nil {
				// if there's an error in parsing the suffixed decimal, it's just a regular identifier, so return it as-is
				return val, nil
			}
		} else {
			// if val is not a string, it can't be a suffixed decimal so return it as-is
			return val, nil
		}
	}

	switch rv.Kind() {
	case reflect.Bool, reflect.String, reflect.Slice, reflect.Array, reflect.Interface:
		return sd.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Type().String() == "time.Duration" {
			return sd.AsDuration()
		} else {
			return sd.AsNumber()
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128:
		return sd.AsNumber()
	}

	return val, nil
}

// setReflectValueFromIntf sets dest to the value of val, returning a non-nil error if dest is not of a compatible
// scalar type.
//
// val is coerced on a best-effort basis into the type of dest:
//
//   - if dest is a bool, any non-zero-value or any string of "y", "t", "yes", "true" (all case-insensitive) or a
//     nonzero numeric string is interpreted as true
//
//   - if dest is an integer, float, or complex type, any integer/float value or any numeric string is assigned
//     directly; time.Time is convered to a UNIX timestamp; time.Duration is converted to a duration in seconds;
//     otherwise the value is converted to a bool (per the rules above) and 1 is assigned if true, 0 if false
//
//   - if dest is a string, numeric values are stringified in decimal, booleans become "true"/"false", []rune and []byte
//     are stringified, .String() or .MarshalText() are invoked if available, and time.Time is rendered per RFC3339
//
//   - if dest is an interface, val is assigned to it directly
//
//   - if dest satisfies the encoding.UnmarshalText interface, val will be stringified per above and passed as a byte slice
//     to UnmarshalText.
func setReflectValueFromIntf(c *unmarshalContext, dest reflect.Value, val interface{}, format string) (reflect.Value, error) {
	return withCreatedAndIndirected(dest, func(rv *reflect.Value) error {
		dest, val, done, err := handleFormatIntf(c, dest, val, format)
		if err != nil {
			return err
		}
		if done {
			return nil
		}

		if unmarshaled, err := unmarshalIntf(c, dest, val, format); unmarshaled {
			return err
		}

		if c.opts.RelaxedNonCompliant.Permit(relaxed.MultiplierSuffixes) {
			if val, err = resolveSuffixedDecimal(rv, val); err != nil {
				return err
			}
		}

		switch rv.Kind() {
		case reflect.Bool:
			rv.SetBool(coerce.ToBool(val))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if rv.Type().String() == "time.Duration" {
				return unmarshalValueDuration(c, *rv, val, format)
			} else {
				rv.SetInt(coerce.ToInt64(val))
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			rv.SetUint(coerce.ToUint64(val))

		case reflect.Float32, reflect.Float64:
			if format != "nonfinite" {
				if s, ok := val.(string); ok {
					switch s {
					case "+Inf", "-Inf", "Inf", "NaN":
						val = 0.0
					}
				}
			}
			rv.SetFloat(coerce.ToFloat64(val))

		case reflect.Complex64, reflect.Complex128:
			rv.SetComplex(coerce.ToComplex128(val))

		case reflect.String:
			rv.SetString(coerce.ToString(val))

		case reflect.Interface:
			destVal := *rv
			sourceVal := reflect.ValueOf(val)
			v := destVal
			if v.IsValid() && v.Elem().IsValid() {
				v = v.Elem()
			}
			if sourceVal.IsValid() {
				v.Set(sourceVal)
			} else {
				v.Set(reflect.Zero(v.Type()))
			}

		default:
			return fmt.Errorf("cannot unmarshal value %q into %s", val, rv.Kind().String())
		}
		return nil

	})
}

func indirectKind(v reflect.Value) reflect.Kind {
	fk := v.Kind()
	if fk == reflect.Ptr {
		fk = v.Type().Elem().Kind()
	}
	return fk
}

func createMapIfNil(mapValue reflect.Value, size int) {
	if mapValue.IsNil() {
		mapValue.Set(reflect.MakeMapWithSize(mapValue.Type(), size))
	}
}

// unmarshalNodeToStruct unmarshals node into destStruct, which must represent a struct.
//
// Arguments will be assigned into destStruct, in order, as follows:
//   - into one or more fields tagged with `kdl:",arg"`; each field will be assigned one argument, in order
//   - into a slice field tagged with `kdl:",args"; all remaining arguments will be assigned as elements in the slice
//
// Properties will be assigned into destStruct as follows:
//   - into a field tagged with the property name, eg: `kdl:"myfield"` for a property named `myfield`
//   - into a single map field tagged with `kdl:",props"`
//
// Conversion rules for keys and values are per setReflectValueFromIntf.
func unmarshalNodeToStruct(c *unmarshalContext, node *document.Node, destStruct reflect.Value) (reflect.Value, error) {
	typeDetails := c.indexer.Get(destStruct.Type().String())

	argFieldInfo := typeDetails.StructAttrs["arg"]
	argsFieldInfo := typeDetails.StructAttrs["args"]
	propsFieldInfo := typeDetails.StructAttrs["props"]
	childrenFieldInfo := typeDetails.StructAttrs["children"]
	var err error

	if len(node.Arguments) > 0 {
		if len(node.Arguments) == 1 && (typeDetails.CanUnmarshalText() || typeDetails.CanUnmarshalKDLValue()) {
			return setReflectValueFromIntf(c, destStruct, node.Arguments[0].ResolvedValue(), "")
		}

		if !c.opts.AllowUnhandledArgs && len(argsFieldInfo) == 0 && len(argFieldInfo) < len(node.Arguments) {
			return reflect.Value{}, fmt.Errorf("%s has unexpected arguments", node.Name.ValueString())
		}

		if len(argsFieldInfo) > 1 {
			return reflect.Value{}, fmt.Errorf("%s must have no more than one field tagged ',args'", destStruct.Type().Name())
		}

		args := node.Arguments[:]

		// assign as many arguments as possible to fields tagged with ",arg"
		for _, fieldInfo := range argFieldInfo {
			field := fieldInfo.GetValueFrom(destStruct)
			field, err = withCreatedAndIndirected(field, func(field *reflect.Value) error {
				f, err := setReflectValueFromIntf(c, *field, args[0].ResolvedValue(), fieldInfo.Format)
				*field = f
				return err
			})
			if err != nil {
				return reflect.Value{}, err
			}
			args = args[1:]
		}

		// if remaining arguments exist, try to find a slice tagged with ",args" to which to assign them
		if len(argsFieldInfo) > 0 {
			fieldInfo := argsFieldInfo[0]
			field := fieldInfo.GetValueFrom(destStruct)
			field, err = withCreatedAndIndirected(field, func(slice *reflect.Value) error {
				sk := slice.Kind()
				if sk != reflect.Slice && sk != reflect.Array {
					return fmt.Errorf("cannot unmarshal arguments for %s into slice %s of non-slice type %s", node.Name.ValueString(), slice.Type().Name(), slice.Kind().String())
				}

				size := len(node.Arguments)
				_ = createSliceIfNil(slice, 0, size)

				var err error
				if err = addArgumentsToSlice(c, args, slice); err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	if node.Properties.Len() > 0 {

		// try to assign each property to a struct field tagged with the property's name
		handledProps := 0
		for propKey, propVal := range node.Properties.Unordered() {
			keyFieldInfo, exists := typeDetails.StructFields[propKey]
			if !exists {
				continue
			}
			field := keyFieldInfo.GetValueFrom(destStruct)
			if field, err = setReflectValueFromIntf(c, field, propVal.ResolvedValue(), keyFieldInfo.Format); err != nil {
				return reflect.Value{}, err
			}
			handledProps++
		}

		havePropsField := len(propsFieldInfo) > 0
		if !c.opts.AllowUnhandledProps && !havePropsField && handledProps < node.Properties.Len() {
			extraProps := make([]string, 0, node.Properties.Len())
			for name := range node.Properties.Unordered() {
				extraProps = append(extraProps, name)
			}
			return reflect.Value{}, fmt.Errorf("%s has unexpected properties %s", node.Name.ValueString(), strings.Join(extraProps, ", "))
		}

		// if we have a struct field tagged with ",props" and it's a map, add all of the properties to it
		if havePropsField {
			if len(propsFieldInfo) > 1 {
				return reflect.Value{}, fmt.Errorf("%s must have no more than one field tagged ',props'", destStruct.Type().Name())
			}

			fieldInfo := propsFieldInfo[0]
			field := fieldInfo.GetValueFrom(destStruct)

			fk := indirectKind(field)
			if fk != reflect.Map {
				return reflect.Value{}, fmt.Errorf("%s is tagged ',props' and must be a map, but is a %s", destStruct.Type().Name(), reflect.Indirect(field).Kind().String())
			}

			field, err := withCreatedAndIndirected(field, func(mapField *reflect.Value) error {
				createMapIfNil(*mapField, node.Properties.Len())

				mapKeyType := mapField.Type().Key()
				mapValType := mapField.Type().Elem()

				for propKey, propVal := range node.Properties.Unordered() {
					if err := setMapKeyValueFromIntf(c, *mapField, mapKeyType, mapValType, propKey, propVal.ResolvedValue()); err != nil {
						return err
					}
					// key := createTypeAndIndirect(mapKeyType)
					// if err := setReflectValueFromIntf(c, key, propKey, ""); err != nil {
					// 	return err
					// }
					//
					// val := createTypeAndIndirect(mapValType)
					// if err := setReflectValueFromIntf(c, val, propVal.ResolvedValue(), ""); err != nil {
					// 	return err
					// }
					//
					// mapField.SetMapIndex(key, val)
				}
				return nil
			})
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}

	if len(node.Children) > 0 {
		haveChildrenField := len(childrenFieldInfo) > 0
		if !c.opts.AllowUnhandledChildren && !haveChildrenField {
			// if we don't have a ",children" field in this struct to put the children into, try unmarshaling each child
			// directly into this struct to see if it has fields matching the node names
			return unmarshalNodesToStruct(c, node.Children, destStruct)
		}

		fieldInfo := childrenFieldInfo[0]
		field := fieldInfo.GetValueFrom(destStruct)

		fk := indirectKind(field)

		switch fk {
		case reflect.Map:
			if field, err = unmarshalNodesToMap(c, node.Children, field); err != nil {
				return reflect.Value{}, err
			}
		case reflect.Struct:
			if field, err = unmarshalNodesToStruct(c, node.Children, field); err != nil {
				return reflect.Value{}, err
			}
		case reflect.Interface:
			m := make(map[string]interface{})
			field.Set(reflect.ValueOf(m))
			if field, err = unmarshalNodesToMap(c, node.Children, field); err != nil {
				return reflect.Value{}, err
			}

		default:
			return reflect.Value{}, fmt.Errorf("%s is tagged ',children' and must be of type map or struct, but is %s", node.Name.ValueString(), fk.String())
		}

	}

	return destStruct, nil
}

// createSliceIfNil allocates a slice if it is nil, and ensures its capacity and length is at least size; it returns the
// index of the first assignable element
func createSliceIfNil(slice *reflect.Value, allocLen, allocCap int) int {
	if slice.IsNil() {
		slice.Set(reflect.MakeSlice(slice.Type(), allocLen, allocCap))
	}
	return slice.Len()
}

// newValueForSlice creates and returns new reflect.Value of the correct type for appending to slice
func newValueForSlice(slice reflect.Value) reflect.Value {
	sliceType := slice.Type().Elem()
	el := reflect.New(sliceType)
	if el.Type().Kind() == reflect.Ptr {
		el = el.Elem()
	}
	return el
}

// addArgumentsToSlice adds args to slice (which must represent a slice) starting at the specified startIdx; returns
// the index of the next assignable element and a non-nil error on failure
//
// Conversion rules for values are per setReflectValueFromIntf.
func addArgumentsToSlice(c *unmarshalContext, args []*document.Value, destSlice *reflect.Value) error {
	var slice = *destSlice
	for _, arg := range args {
		dst := newValueForSlice(slice)
		dst, err := setReflectValueFromIntf(c, dst, arg.ResolvedValue(), "")
		if err != nil {
			return err
		}
		slice = reflect.Append(slice, dst)
	}
	*destSlice = slice
	return nil
}

func unmarshalNodeToByteSlice(c *unmarshalContext, node *document.Node, destSlice *reflect.Value, format string) error {
	if len(node.Arguments) == 0 {
		return errors.New("cannot unmarshal node with no arguments []byte")
	}

	if node.Properties.Len() != 0 {
		return errors.New("cannot unmarshal node with properties into []byte")
	}
	if len(node.Children) != 0 {
		return errors.New("cannot unmarshal node with children into []byte")
	}

	if format == "" {
		if len(node.Arguments) > 1 || coerce.IsNumeric(node.Arguments[0].Value) {
			format = "array"
		} else {
			format = "base64"
		}
	}

	var bs []byte
	if format == "array" {
		bs = make([]byte, len(node.Arguments))
		for i, arg := range node.Arguments {
			bs[i] = coerce.ToByte(arg.ResolvedValue())
		}
	} else {

		if len(node.Arguments) != 1 {
			return fmt.Errorf("cannot unmarshal node with %d arguments into []byte", len(node.Arguments))
		}

		arg := node.Arguments[0]
		var err error

		switch format {
		case "base64":
			bs, err = base64.StdEncoding.DecodeString(arg.ValueString())
		case "base64url":
			bs, err = base64.URLEncoding.DecodeString(arg.ValueString())
		case "base32":
			bs, err = base32.StdEncoding.DecodeString(arg.ValueString())
		case "base32hex":
			bs, err = base32.HexEncoding.DecodeString(arg.ValueString())
		case "base16", "hex":
			bs, err = hex.DecodeString(arg.ValueString())
		case "string":
			bs = []byte(arg.ValueString())
		default:
			return fmt.Errorf("invalid []byte format: %s", format)
		}
		if err != nil {
			return err
		}
	}

	destSlice.Set(reflect.ValueOf(bs))
	return nil
}

// unmarshalNodeToSlice unmarshals node to destSlice, which must represent a slice.
//
// Arguments will be added to destSlice first, one element per argument.
//
// Properties will be added to destSlice following the arguments as follows:
//   - if destSlice is a []interface{}, each property will be added as an []interface{}{string(property-key), interface{}(property-value)}
//   - if destSlice is a []string, each property will be added as "property-key=property-value"
//   - if destSlice is of any other type and properties exist, an non-nil error will be returned
//
// Conversion rules for values are per setReflectValueFromIntf.
func unmarshalNodeToSlice(c *unmarshalContext, node *document.Node, destSlice *reflect.Value, format string) error {
	sliceElementType := destSlice.Type().Elem().Kind()
	if sliceElementType == reflect.Uint8 {
		return unmarshalNodeToByteSlice(c, node, destSlice, format)
	}

	size := len(node.Arguments) + node.Properties.Len()
	_ = createSliceIfNil(destSlice, 0, size)

	var err error
	if err = addArgumentsToSlice(c, node.Arguments, destSlice); err != nil {
		return err
	}

	if node.Properties.Len() > 0 {
		switch sliceElementType {
		case reflect.Interface, reflect.String:
		default:
			return fmt.Errorf("cannot unmarshal %s into slice of type []%s (must be []string or []interface{} because it has properties)", node.Name.ValueString(), sliceElementType.String())
		}

		b := strings.Builder{}
		for key, val := range node.Properties.Unordered() {
			dst := newValueForSlice(*destSlice)

			switch sliceElementType {
			case reflect.Interface:
				v := []interface{}{key, val.ResolvedValue()}
				dst.Set(reflect.ValueOf(v))
			case reflect.String:
				b.Reset()
				b.WriteString(key)
				b.WriteByte('=')
				b.WriteString(val.ValueString())
				if dst, err = setReflectValueFromIntf(c, dst, b.String(), ""); err != nil {
					return err
				}
			}
			*destSlice = reflect.Append(*destSlice, dst)
		}
	}

	return nil
}

func setMapKeyValueFromIntf(c *unmarshalContext, destMap reflect.Value, mapKeyType reflect.Type, mapValType reflect.Type, keyIntf interface{}, valueIntf interface{}) error {
	var (
		key           reflect.Value
		keyIndirect   reflect.Value
		keyIndirected bool
		val           reflect.Value
		valIndirect   reflect.Value
		valIndirected bool
	)

	key = reflect.New(mapKeyType).Elem()
	if key.Type().Kind() == reflect.Ptr {
		keyIndirect = key.Elem()
		keyIndirected = true
	} else {
		keyIndirect = key
	}

	var err error

	if keyIndirect, err = setReflectValueFromIntf(c, keyIndirect, keyIntf, ""); err != nil {
		return err
	}

	// set the value (creating if necessary)
	val = destMap.MapIndex(key)
	if !val.IsValid() {
		if mapValType.Kind() == reflect.Ptr {
			val = reflect.New(mapValType.Elem())
			valIndirect = val.Elem()
			valIndirected = true
		} else {
			val = reflect.New(mapValType).Elem()
			valIndirect = val
		}
	} else if val.Type().Kind() == reflect.Ptr {
		valIndirect = val.Elem()
		if !valIndirect.IsValid() {
			valIndirect = reflect.New(val.Type().Elem()).Elem()

		}
		valIndirected = true
	} else {
		valIndirect = val
	}

	if valIndirect, err = setReflectValueFromIntf(c, valIndirect, valueIntf, ""); err != nil {
		return err
	}

	if keyIndirected {
		key.Elem().Set(keyIndirect)
	} else {
		key.Set(keyIndirect)
	}

	if valIndirected {
		val.Elem().Set(valIndirect)
	} else {
		val.Set(valIndirect)
	}

	// add to the map
	destMap.SetMapIndex(key, val)
	return nil
}

func setMapKeyValueFromFunc(c *unmarshalContext, destMap reflect.Value, mapKeyType reflect.Type, mapValType reflect.Type, keyIntf interface{}, f func(val *reflect.Value) error) error {
	// set the key (creating if necessary)
	key := createTypeAndIndirect(mapKeyType)
	var err error

	if key, err = setReflectValueFromIntf(c, key, keyIntf, ""); err != nil {
		return err
	}

	// set the value (creating if necessary)
	val := destMap.MapIndex(key)
	if !val.IsValid() {
		val = reflect.New(mapValType)
		if mapValType.Kind() == reflect.Ptr {
			val = reflect.New(val.Type().Elem())
		}
		val = val.Elem()
	}

	if err := f(&val); err != nil {
		return err
	}

	// add to the map
	destMap.SetMapIndex(key, val)
	return nil
}

// unmarshalNodeToMap unmarshals node into destMap, which must represent a map.
//
// Arguments will be assigned into destMap keyed by the zero-based index of the argument (eg: the first argument is
// destMap[0] or destMap[0], etc., depending on the key type).
//
// Properties will be assigned into destMap by their keys.
//
// Children will be assigned into destMap by their names.
//
// Conversion rules for both keys and values are per setReflectValueFromIntf.
func unmarshalNodeToMap(c *unmarshalContext, node *document.Node, destMap reflect.Value) error {

	createMapIfNil(destMap, len(node.Arguments)+node.Properties.Len())
	mapKeyType := destMap.Type().Key()
	mapValType := destMap.Type().Elem()

	// unmarshal the node's arguments into the map with the argument number as the key, and the argument value as the value
	for i, arg := range node.Arguments {
		if err := setMapKeyValueFromIntf(c, destMap, mapKeyType, mapValType, i, arg.ResolvedValue()); err != nil {
			return err
		}
	}

	// unmarshal the node's properties into the map
	for propKey, propVal := range node.Properties.Unordered() {
		if err := setMapKeyValueFromIntf(c, destMap, mapKeyType, mapValType, propKey, propVal.ResolvedValue()); err != nil {
			return err
		}
	}

	// unmarshal the node's children into the map
	for _, childNode := range node.Children {
		err := setMapKeyValueFromFunc(c, destMap, mapKeyType, mapValType, childNode.Name.ResolvedValue(), func(val *reflect.Value) error {
			return unmarshalNodeToValue(c, childNode, val, "")
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// createTypeAndIndirect creates a new Value of type t and dereferences it, returning the Value
func createTypeAndIndirect(t reflect.Type) reflect.Value {
	var v reflect.Value
	if t.Kind() == reflect.Ptr {
		v = reflect.New(t.Elem())
	} else {
		v = reflect.New(t)
	}
	return v.Elem()
}

// createAndIndirect allocates a new value if v is a nil ptr and/or dereferences it if necessary, returning the
// resulting Value.
func createAndIndirect(v reflect.Value) (reflect.Value, bool) {
	var e reflect.Type
	if v.Kind() == reflect.Ptr {
		e = v.Type().Elem()
	} else {
		e = v.Type()
	}
	if v.Kind() == reflect.Ptr && v.IsNil() {
		v.Set(reflect.New(e))
	}

	indirected := false
	if v.Kind() == reflect.Ptr {
		indirected = true
		v = v.Elem()
	}

	return v, indirected
}

func withCreatedAndIndirected(v reflect.Value, f func(v *reflect.Value) error) (reflect.Value, error) {
	indirectCreatedV, indirected := createAndIndirect(v)
	err := f(&indirectCreatedV)

	if indirected {
		if e := v.Elem(); e.CanSet() {
			e.Set(indirectCreatedV)
		} else {
			ptr := reflect.New(v.Type())
			ptr.Elem().Set(indirectCreatedV)
			v = ptr
		}
	} else {
		if v.CanSet() {
			v.Set(indirectCreatedV)
		} else {
			v = indirectCreatedV
		}
	}
	return v, err
}

// unmarshalNodeToMultiDimensionalMap unmarshals node into a multi-dimensional map, using the first n nodes as map keys.
//
// For example, if destMap is of type map[string]map[string]int and the KDL is:
//
// person "contractor" "Foo Inc." 1234
// person "contractor" "Bar Inc." 3456
//
// ...then after unmarshaling, destMap will contain:
//
// {"contractor": {"Foo Inc.": 1234, "Bar Inc.": 3456}}
func unmarshalNodeToMultiDimensionalMap(c *unmarshalContext, node *document.Node, destMap reflect.Value) error {
	createMapIfNil(destMap, 4)

	mapKeyType := destMap.Type().Key()
	mapValType := destMap.Type().Elem()

	key := node.Arguments[0].ResolvedValue()
	node.Arguments = node.Arguments[1:]

	err := setMapKeyValueFromFunc(c, destMap, mapKeyType, mapValType, key, func(val *reflect.Value) error {
		if val.Kind() == reflect.Map && len(node.Arguments) > 0 {
			return unmarshalNodeToMultiDimensionalMap(c, node, *val)
		} else {
			return unmarshalNodeToValue(c, node, val, "")
		}
	})
	if err != nil {
		return err
	}

	return nil

}

// unmarshalNodeToMultiple unmarshals node into a map or slice supporting multiple node instances.
//
// Normally, unmarshaling into a map or slice has the semantics described in unmarshalNodeToMap or unmarshalNodeToSlice.
//
// Tagging map-type struct field with ",multiple" changes the semantics so that node's first n arguments are treated as
// map keys (where n represents the depth of nested maps) and the remaining arguments and properties are assigned into
// the value at the final map value.
//
// Tagging a slice-type struct field with ",multiple" changes the sematics so that each instance of the node is assigned
// to a new element appended to the slice.
//
// For example:
//
//	type foo struct {
//	  Items map[string]map[string]string `kdl:"items,multiple"`
//	}
//
// items "foo" "bar" "baz"
// items "foo" "abc" "def"
//
// yields:
// foo.Items = map[string]map[string]string{ "foo": {"bar": "baz", "abc": "def"}}
func unmarshalNodeToMultiple(c *unmarshalContext, node *document.Node, destValue *reflect.Value) error {
	newValue, err := withCreatedAndIndirected(*destValue, func(dest *reflect.Value) error {
		switch dest.Kind() {
		case reflect.Map:
			// make a copy of the node as we want to consume the first argument(s)
			node := node.ShallowCopy()

			return unmarshalNodeToMultiDimensionalMap(c, node, *dest)
		case reflect.Slice, reflect.Array:
			_ = createSliceIfNil(dest, 0, 2)

			el := newValueForSlice(*dest)
			if err := unmarshalNodeToValue(c, node, &el, ""); err != nil {
				return err
			}
			*dest = reflect.Append(*dest, el)
			return nil
		default:
			return fmt.Errorf("only maps and slices may have the 'multiple' tag")
		}
	})
	if err == nil {
		*destValue = newValue
	}
	return err
}

// unmarshalNodeToValue unmarshals node to dest, which can be of any supported type, and returns a non-nil error on
// failure.
func unmarshalNodeToValue(c *unmarshalContext, node *document.Node, destValue *reflect.Value, format string) error {
	var (
		unmarshaled bool
		err         error
	)
	if unmarshaled, *destValue, err = unmarshalNode(c, *destValue, node, format); unmarshaled {
		return err
	}

	rv, err := withCreatedAndIndirected(*destValue, func(dest *reflect.Value) error {
		switch dest.Kind() {
		case reflect.Struct:
			v, err := unmarshalNodeToStruct(c, node, *dest)
			if err == nil {
				*dest = v
			}
			return err
		case reflect.Map:
			return unmarshalNodeToMap(c, node, *dest)
		case reflect.Slice, reflect.Array:
			return unmarshalNodeToSlice(c, node, dest, format)
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64,
			reflect.Complex64, reflect.Complex128,
			reflect.String:
			if err := verifyArgsPropsChildren(c, node, 1, nil, false); err != nil {
				return err
			}

			v, err := setReflectValueFromIntf(c, *dest, node.Arguments[0].ResolvedValue(), format)
			if err == nil {
				*dest = v
			}
			return err

		case reflect.Interface:
			destVal := dest

			v := *destVal
			if v.IsValid() && v.Elem().IsValid() {
				v = v.Elem()
			}

			if len(node.Arguments) > 0 || node.Properties.Len() > 0 || len(node.Children) > 0 {
				if len(node.Arguments) == 1 && node.Properties.Len() == 0 && len(node.Children) == 0 {
					sourceVal := reflect.ValueOf(node.Arguments[0].ResolvedValue())
					if sourceVal.IsValid() {
						v.Set(sourceVal)
					} else {
						v.Set(reflect.Zero(v.Type()))
					}
				} else if len(node.Arguments) > 1 && node.Properties.Len() == 0 && len(node.Children) == 0 {
					destSlice := reflect.ValueOf(make([]interface{}, 0, len(node.Arguments)))
					if err := unmarshalNodeToSlice(c, node, &destSlice, format); err != nil {
						return err
					}
					v.Set(destSlice)
				} else {
					m := make(map[string]interface{}, len(node.Arguments)+node.Properties.Len()+len(node.Children))
					mv := reflect.ValueOf(m)
					if err := unmarshalNodeToMap(c, node, mv); err != nil {
						return err
					}
					if v.CanSet() {
						v.Set(mv)
					} else {
						*dest = mv
					}
				}
			}
		default:
			return fmt.Errorf("cannot unmarshal node %s into %s (originally %s)", node.Name.ValueString(), dest.Type().String(), destValue.Type().String())
		}

		return nil
	})
	if err == nil {
		*destValue = rv
	}
	return err
}

// unmarshalNodeToStructField unmarshals node into dest, which must represent a struct field.
//
// Internally, this will either call unmarshalNodeToMultiple (if this field is a map tagged ",multiple") or
// unmarshalNodeToValue.
func unmarshalNodeToStructField(c *unmarshalContext, node *document.Node, destStruct reflect.Value) error {
	name := node.Name.ValueString()
	typeDetails := c.indexer.Get(destStruct.Type().String())
	destFieldInfo, exists := typeDetails.StructFields[name]
	if !exists {
		if c.opts.AllowUnhandledNodes {
			return nil
		} else {
			// println(destStruct.Type().String())
			// for sn, sf := range typeDetails.StructFields {
			// 	println(sn, ": ", strings.Join(sf.Attrs, ","))
			// }
			return fmt.Errorf("no struct field into which to unmarshal node %q", name)
		}
	}

	destFieldValue := destFieldInfo.GetValueFrom(destStruct)

	if destFieldInfo.IsMultiple() {
		v := destFieldValue
		err := unmarshalNodeToMultiple(c, node, &v)
		destFieldValue.Set(v)
		return err
	}

	return unmarshalNodeToValue(c, node, &destFieldValue, destFieldInfo.Format)
}

// unmarshalNodesToStruct unmarshals each node in nodes into the destStruct, which must represent a struct value.
func unmarshalNodesToStruct(c *unmarshalContext, nodes []*document.Node, destStruct reflect.Value) (reflect.Value, error) {
	return withCreatedAndIndirected(destStruct, func(destStruct *reflect.Value) error {
		for _, node := range nodes {
			if err := unmarshalNodeToStructField(c, node, *destStruct); err != nil {
				return err
			}
		}

		return nil
	})

}

// unmarshalNodesToMap unmarshals each node in nodes into the destMap, which must represent a map value.
func unmarshalNodesToMap(c *unmarshalContext, nodes []*document.Node, destMap reflect.Value) (reflect.Value, error) {
	return withCreatedAndIndirected(destMap, func(destMap *reflect.Value) error {
		createMapIfNil(*destMap, len(nodes))

		mapKeyType := destMap.Type().Key()
		mapValType := destMap.Type().Elem()

		for _, node := range nodes {
			err := setMapKeyValueFromFunc(c, *destMap, mapKeyType, mapValType, node.Name.ResolvedValue(), func(val *reflect.Value) error {
				return unmarshalNodeToValue(c, node, val, "")
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

var ErrStructOrMap = errors.New("can only decode into struct or map")

// unmarshalNodes unmarshals each node in nodes into dest, which must represent a struct, map, slice, or interface{}
// type.
func unmarshalNodes(c *unmarshalContext, nodes []*document.Node, dest reflect.Value) (reflect.Value, error) {
	return withCreatedAndIndirected(dest, func(dest *reflect.Value) error {
		if !dest.IsValid() {
			v := make(map[string]interface{})
			dest.Set(reflect.ValueOf(v))
		}

		switch dest.Kind() {
		case reflect.Struct:
			v, err := unmarshalNodesToStruct(c, nodes, *dest)
			if err == nil {
				*dest = v
			}
			return err

		case reflect.Map:
			v, err := unmarshalNodesToMap(c, nodes, *dest)
			if err == nil {
				*dest = v
			}
			return err

		case reflect.Interface:
			v := make(map[string]interface{}, len(nodes))
			newMap := reflect.ValueOf(v)
			newMap, err := unmarshalNodesToMap(c, nodes, newMap)
			if err != nil {
				return err
			}
			dest.Set(newMap)
		default:
			return ErrStructOrMap
		}
		return nil
	})
}

var ErrNeedPointer = errors.New("must unmarshal into pointer type or map")

func UnmarshalWithOptions(doc *document.Document, v interface{}, opts UnmarshalOptions) error {
	c := &unmarshalContext{
		opts: opts,
	}
	c.indexer = newTypeIndexer()
	if err := c.indexer.IndexIntf(v); err != nil {
		return err
	}

	target := reflect.ValueOf(v)
	switch target.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Interface:
		_, err := unmarshalNodes(c, doc.Nodes, target)
		return err
	default:
		return ErrNeedPointer
	}
}

func Unmarshal(doc *document.Document, v interface{}) error {
	opts := UnmarshalOptions{}
	return UnmarshalWithOptions(doc, v, opts)
}
