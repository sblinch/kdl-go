package marshaler

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/sblinch/kdl-go/document"
	"github.com/sblinch/kdl-go/internal/coerce"
	"github.com/sblinch/kdl-go/internal/tokenizer"
)

type marshaler interface {
	MarshalKDL(node *document.Node) error
}

type valueMarshaler interface {
	MarshalKDLValue(value *document.Value) error
}

type MarshalOptions struct {
}

type marshalContext struct {
	indexer *typeIndexer
	opts    MarshalOptions
}

func Marshal(v interface{}, doc *document.Document) error {
	opts := MarshalOptions{}
	return MarshalWithOptions(v, doc, opts)
}

func MarshalWithOptions(v interface{}, doc *document.Document, opts MarshalOptions) error {
	c := &marshalContext{}
	c.indexer = newTypeIndexer()
	if err := c.indexer.IndexIntf(v); err != nil {
		return err
	}

	nodes, err := marshalValueToNodes(c, reflect.ValueOf(v))
	if err != nil {
		return err
	}

	doc.Nodes = append(doc.Nodes, nodes...)
	return nil
}

func marshalMapToNodes(c *marshalContext, srcMap reflect.Value, nodes []*document.Node) ([]*document.Node, error) {
	// mapKeyType := srcMap.Type().Key()
	// mapValType := srcMap.Type().Elem()

	mapKeys := sortMapKeys(srcMap.MapKeys())
	if nodes == nil {
		nodes = make([]*document.Node, 0, len(mapKeys))
	}

	for _, keyVal := range mapKeys {
		node := document.NewNode()
		node.SetName(coerce.ToString(keyVal.Interface()))
		valVal := srcMap.MapIndex(keyVal)

		if node, err := marshalValueToNode(c, coerce.ToString(keyVal.Interface()), valVal, nil); err != nil {
			return nil, err
		} else if node != nil {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

func marshalByteSliceToNode(c *marshalContext, name string, slice reflect.Value, fldDetails *structFieldDetails) (*document.Node, error) {
	node := document.NewNode()
	node.SetName(name)

	bs, ok := slice.Interface().([]byte)
	if !ok {
		return nil, errors.New("byte slice was not a []byte?")
	}

	format := fldDetails.Format

	if format == "" {
		format = "base64"
	}

	if format == "array" {
		for _, c := range bs {
			node.AddArgument(c, "")
		}
		return node, nil
	}

	switch format {
	case "base64":
		node.AddArgument(base64.StdEncoding.EncodeToString(bs), "")
	case "base64url":
		node.AddArgument(base64.URLEncoding.EncodeToString(bs), "")
	case "base32":
		node.AddArgument(base32.StdEncoding.EncodeToString(bs), "")
	case "base32hex":
		node.AddArgument(base32.HexEncoding.EncodeToString(bs), "")
	case "base16", "hex":
		node.AddArgument(hex.EncodeToString(bs), "")
	case "string":
		node.AddArgument(string(bs), "")
	default:
		return nil, fmt.Errorf("invalid []byte encoding format: %s", format)
	}

	return node, nil
}

func marshalSliceToNode(c *marshalContext, name string, slice reflect.Value, fldDetails *structFieldDetails) (*document.Node, error) {
	elemKind := slice.Type().Elem().Kind()
	if elemKind == reflect.Uint8 {
		return marshalByteSliceToNode(c, name, slice, fldDetails)
	}

	var format string
	if fldDetails != nil {
		format = fldDetails.Format
	}

	isStringSlice := elemKind == reflect.String

	node := document.NewNode()
	node.SetName(name)

	n := slice.Len()
	for i := 0; i < n; i++ {
		el := reflect.Indirect(slice.Index(i))
		if isStringSlice {
			// if this is a []string and this element contains a `=`, marshal it as a property
			if k, v, ok := strings.Cut(el.String(), "="); ok {
				node.AddProperty(k, coerce.FromString(v), "")
				continue
			}
		}

		if el.Kind() == reflect.Interface && el.Elem().IsValid() {
			el = el.Elem()
		}

		// if this is a []interface{} and the element is a []interface{}{"key", value}, marshal it as a property
		if el.Kind() == reflect.Slice {
			v := el.Interface()
			if kv, ok := v.([]interface{}); ok && len(kv) == 2 {
				node.AddProperty(coerce.ToString(kv[0]), kv[1], "")
				continue
			}
		}

		arg := node.AddArgument(nil, "")
		if err := reflectValueToDocumentValue(c, el, arg, format); err != nil {
			return nil, err
		}
	}

	return node, nil
}

func marshalMapToNode(c *marshalContext, name string, m reflect.Value, fldDetails *structFieldDetails) (*document.Node, error) {
	node := document.NewNode()
	node.SetName(name)

	keys := sortMapKeys(m.MapKeys())

	var args map[int64]interface{}

	var format string
	if fldDetails != nil {
		format = fldDetails.Format
	}

	// we assume that all map values are properties unless they are maps, structs, or slices, in which case they have
	// to be represented as children; this may result in an input document having a node with values represented as
	// children and the output document having those same values represented as properties, but this is an ambiguous
	// scenario and there's no other solution, so we may as well choose the most compact approach (props, not children)
	for _, key := range keys {
		val := reflect.Indirect(m.MapIndex(key))
		keyIntf := key.Interface()

		if child, skip, err := tryMarshalValueAsChild(c, keyIntf, val, nil); err != nil {
			return nil, err
		} else if child != nil {
			node.AddNode(child)
		} else if !skip {
			dv := &document.Value{}
			if err := reflectValueToDocumentValue(c, val, dv, format); err != nil {
				return nil, err
			} else if coerce.IsInteger(keyIntf) {
				if args == nil {
					args = make(map[int64]interface{})
				}
				args[coerce.ToInt64(keyIntf)] = dv.Value
			} else {
				node.AddPropertyValue(coerce.ToString(keyIntf), dv, "")
			}
		}

	}

	lenArgs := int64(len(args))
	for i := int64(0); i < lenArgs; i++ {
		node.AddArgument(args[i], "")
	}

	return node, nil
}

// tryMarshalValueAsChild checks whether val must be unmarshaled as a child node (not a property); if so, it returns
// the child node to be added, otherwise it returns (nil, nil) to indicate that val can be added as a property
func tryMarshalValueAsChild(c *marshalContext, nameIntf interface{}, val reflect.Value, fldDetails *structFieldDetails) (n *document.Node, skip bool, e error) {

	if val.Kind() == reflect.Interface && val.Elem().IsValid() {
		val = val.Elem()
	}

	if fldDetails != nil && fldDetails.Attrs.Has("omitempty") && val.IsZero() {
		return nil, true, nil
	}

	typeDetails := c.indexer.Get(val.Type().String())
	// if it implements a marshaler interface, it definitely doesn't marshal into child nodes
	if typeDetails != nil && typeDetails.CanMarshalKDL() {
		return nil, false, nil
	}

	marshalAsChild := fldDetails != nil && fldDetails.Attrs.Has("child")
	hasMarshaler := typeDetails != nil && (typeDetails.CanMarshalKDLValue() || typeDetails.CanMarshalText())

	if !hasMarshaler {
		switch val.Kind() {
		case reflect.Map:
			n, err := marshalMapToNode(c, coerce.ToString(nameIntf), val, &structFieldDetails{})
			return n, false, err
		case reflect.Slice:
			n, err := marshalSliceToNode(c, coerce.ToString(nameIntf), val, &structFieldDetails{})
			return n, false, err
		case reflect.Struct:
			n, err := marshalStructToNode(c, coerce.ToString(nameIntf), val, &structFieldDetails{})
			return n, false, err
		}
	}

	if marshalAsChild {
		n, err := marshalValueToNode(c, coerce.ToString(nameIntf), val, fldDetails)
		return n, false, err
	}
	return nil, false, nil
}

const msgMarshalTextErr = "parsing value returned from MarshalText(): %w"

func marshalKDLNode(name string, srcStruct reflect.Value, typeDetails *typeDetails) (*document.Node, error) {
	node := document.NewNode()
	node.SetName(name)
	if _, err := callStructMethod(srcStruct, typeDetails.KDLMarshalerMethod, reflect.ValueOf(node)); err != nil {
		return nil, err
	} else {
		return node, nil
	}
}

func marshalTimeValue(srcTime reflect.Value, format string) (interface{}, error) {
	t, ok := srcTime.Interface().(time.Time)
	if !ok {
		return nil, errors.New("not a time.Time")
	}
	switch format {
	case "":
		// if no format is specified, use RFC3339
		return t.Format(time.RFC3339), nil
	case "unix":
		return t.Unix(), nil
	case "unixmilli":
		return t.UnixMilli(), nil
	case "unixmicro":
		return t.UnixMicro(), nil
	case "unixnano":
		return t.UnixNano(), nil
	default:
		if fmtStr := timeFormat(format); fmtStr != "" {
			return t.Format(fmtStr), nil
		} else {
			return nil, fmt.Errorf("invalid format string: %s", format)
		}
	}
}

func formatDuration(d time.Duration) string {
	ms := d.Nanoseconds()
	s := ms / 1000000000
	ms -= s * 1000000000
	m := s / 60
	s -= m * 60
	h := m / 60
	m -= h * 60
	return fmt.Sprintf("%02d:%02d:%02d.%d", h, m, s, ms)
}

func intOrFloat(f float64) interface{} {
	if f-math.Floor(f) == 0 {
		return int(f)
	} else {
		return f
	}
}

func marshalDurationValue(srcDuration reflect.Value, format string) (interface{}, error) {
	d, ok := srcDuration.Interface().(time.Duration)
	if !ok {
		return nil, errors.New("not a time.Duration")
	}
	switch format {
	case "":
		// if no format is specified, use NhNmNs
		return d.String(), nil
	case "sec":
		return intOrFloat(d.Seconds()), nil
	case "milli":
		return d.Milliseconds(), nil
	case "micro":
		return d.Microseconds(), nil
	case "nano":
		return d.Nanoseconds(), nil
	case "base60":
		return formatDuration(d), nil
	default:
		return nil, fmt.Errorf("invalid format string: %s", format)
	}
}

func marshalTextValue(srcStruct reflect.Value, typeDetails *typeDetails, format string) (interface{}, error) {
	if srcStruct.Type().String() == "time.Time" {
		return marshalTimeValue(srcStruct, format)
	} else if values, err := callStructMethod(srcStruct, typeDetails.TextMarshalerMethod); err != nil {
		return nil, err
	} else {
		if b, ok := values[0].Interface().([]byte); ok {
			if len(b) == 0 {
				return "", nil
			} else {
				token, err := tokenizer.ScanOne(b)
				if err != nil {
					// if it can't be scanned, it's a just a bare string
					return string(b), nil
				}
				switch token.ID {
				case tokenizer.BareIdentifier,
					tokenizer.RawString,
					tokenizer.QuotedString:
					// return all string types as-is, as encoding.TextMarshaler is a generic interface that does not
					// know about KDL and so if a quote is in the string, it should be included and escaped in the
					// generated output
					return string(token.Data), nil

				case tokenizer.Decimal,
					tokenizer.Hexadecimal,
					tokenizer.Octal,
					tokenizer.Binary,
					tokenizer.Boolean,
					tokenizer.Null:
					// perhaps a bad idea, but parse out numeric, boolean, and null values to a their actual typed values
					if v, err := document.ValueFromToken(token); err != nil {
						return nil, fmt.Errorf(msgMarshalTextErr, err)
					} else {
						return v.ResolvedValue(), nil
					}
				default:
					return nil, fmt.Errorf("MarshalText returned a %s, but a Value type is required", token.ID.String())
				}
			}
		} else {
			return nil, fmt.Errorf("invalid UnmarshalText on %s", srcStruct.Type().String())
		}
	}

}

func marshalKDLValue(srcStruct reflect.Value, typeDetails *typeDetails, format string, v *document.Value) error {
	_, err := callStructMethod(srcStruct, typeDetails.KDLValueMarshalerMethod, reflect.ValueOf(v))
	return err
}

func reflectValueToDocumentValue(c *marshalContext, rv reflect.Value, dv *document.Value, format string) (err error) {
	typeStr := rv.Type().String()
	typeDetails := c.indexer.Get(typeStr)
	if typeDetails != nil && typeDetails.CanMarshalKDLValue() {
		err = marshalKDLValue(rv, typeDetails, format, dv)
	} else if typeDetails != nil && typeDetails.CanMarshalText() {
		dv.Value, err = marshalTextValue(rv, typeDetails, format)
	} else if typeStr == "time.Duration" {
		dv.Value, err = marshalDurationValue(rv, format)
	} else {
		dv.Value = rv.Interface()
	}

	return
}

func marshalStructToNode(c *marshalContext, name string, s reflect.Value, fldDetails *structFieldDetails) (*document.Node, error) {
	typeDetails := c.indexer.Get(s.Type().String())

	node := document.NewNode()
	node.SetName(name)

	argFieldInfo := typeDetails.StructAttrs["arg"]
	argsFieldInfo := typeDetails.StructAttrs["args"]
	propsFieldInfo := typeDetails.StructAttrs["props"]
	childrenFieldInfo := typeDetails.StructAttrs["children"]

	// pull arguments from fields tagged `,arg`
	node.ExpectArguments(len(argFieldInfo))
	for _, argField := range argFieldInfo {
		v := reflect.Indirect(argField.GetValueFrom(s))
		dv := node.AddArgument(nil, "")
		if err := reflectValueToDocumentValue(c, v, dv, argField.Format); err != nil {
			return nil, err
		}
	}

	// pull arguments from field tagged `,args`
	for _, argsField := range argsFieldInfo {
		slice := reflect.Indirect(argsField.GetValueFrom(s))
		if slice.Kind() != reflect.Slice {
			return nil, fmt.Errorf("non-slice type %s tagged with ',args'", slice.Kind().String())
		}

		n := slice.Len()
		node.ExpectArguments(n)
		for i := 0; i < n; i++ {
			el := reflect.Indirect(slice.Index(i))
			dv := node.AddArgument(nil, "")
			if err := reflectValueToDocumentValue(c, el, dv, argsField.Format); err != nil {
				return nil, err
			}
		}
		break
	}

	// pull properties from field tagged `,props`
	for _, propsField := range propsFieldInfo {
		m := reflect.Indirect(propsField.GetValueFrom(s))
		if m.Kind() != reflect.Map {
			return nil, fmt.Errorf("non-map type %s tagged with ',props'", m.Kind().String())
		}

		keys := sortMapKeys(m.MapKeys())

		for _, key := range keys {
			val := m.MapIndex(key)
			dv := node.AddProperty(coerce.ToString(key.Interface()), nil, "")
			if err := reflectValueToDocumentValue(c, val, dv, propsField.Format); err != nil {
				return nil, err
			}
		}

		break
	}

	// pull properties from fields tagged with kdl property names
	for _, fldName := range typeDetails.StructFieldNameList {
		fldDetails := typeDetails.StructFields[fldName]
		if fldName != "-" && !fldDetails.IsCapture() {
			val := reflect.Indirect(fldDetails.GetValueFrom(s))

			if child, skip, err := tryMarshalValueAsChild(c, fldName, val, fldDetails); err != nil {
				return nil, err
			} else if child != nil {
				node.AddNode(child)
			} else if !skip {
				dv := node.AddProperty(fldName, nil, "")
				if err := reflectValueToDocumentValue(c, val, dv, fldDetails.Format); err != nil {
					return nil, err
				}
			}
		}
	}

	// pull children from fields tagged with `,children`
	node.ExpectChildren(len(childrenFieldInfo))
	for _, childrenField := range childrenFieldInfo {
		if children, err := marshalValueToNodes(c, childrenField.GetValueFrom(s)); err != nil {
			return nil, err
		} else if node.Children == nil {
			node.Children = children
		} else {
			node.Children = append(node.Children, children...)
		}
		break
	}

	return node, nil
}

func marshalValueToNode(c *marshalContext, name string, value reflect.Value, fldDetails *structFieldDetails) (*document.Node, error) {
	v := reflect.Indirect(value)

	if fldDetails != nil && fldDetails.Attrs.Has("omitempty") && v.IsZero() {
		return nil, nil
	}

	typeDetails := c.indexer.Get(value.Type().String())
	if typeDetails != nil {
		if typeDetails.CanMarshalKDL() {
			return marshalKDLNode(name, value, typeDetails)
		} else if typeDetails.CanMarshalKDLValue() {
			node := document.NewNode()
			node.SetName(name)
			dv := node.AddArgument(nil, "")
			if err := marshalKDLValue(v, typeDetails, fldDetails.Format, dv); err != nil {
				return nil, err
			}
			return node, nil
		} else if typeDetails.CanMarshalText() {
			node := document.NewNode()
			node.SetName(name)
			if arg, err := marshalTextValue(v, typeDetails, fldDetails.Format); err != nil {
				return nil, err
			} else {
				node.AddArgument(arg, "")
			}
			return node, nil
		}
	}

	var format string
	if fldDetails != nil {
		format = fldDetails.Format
	}

	switch v.Kind() {
	case reflect.Struct:
		return marshalStructToNode(c, name, v, fldDetails)
	case reflect.Map:
		return marshalMapToNode(c, name, v, fldDetails)
	case reflect.Slice:
		// this will need to handle byte slices too
		return marshalSliceToNode(c, name, v, fldDetails)
	case reflect.Interface:
		el := v.Elem()
		if el.IsValid() {
			return marshalValueToNode(c, name, v.Elem(), fldDetails)
		} else {
			node := document.NewNode()
			node.SetName(name)
			node.AddArgument(nil, "")
			return node, nil
		}
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Complex64, reflect.Complex128:

		node := document.NewNode()
		node.SetName(name)

		dv := node.AddArgument(nil, "")
		if err := reflectValueToDocumentValue(c, v, dv, format); err != nil {
			return nil, err
		}
		return node, nil
	case reflect.Float32, reflect.Float64:
		node := document.NewNode()
		node.SetName(name)

		dv := node.AddArgument(nil, "")
		if err := reflectValueToDocumentValue(c, v, dv, format); err != nil {
			return nil, err
		} else {
			f := coerce.ToFloat64(dv.Value)
			if math.IsInf(f, 0) {
				if format == "nonfinite" {
					inf := "+Inf"
					if math.IsInf(f, -1) {
						inf = "-Inf"
					}
					dv.Value = inf
				} else {
					dv.Value = 0.0
				}

			} else if math.IsNaN(f) {
				if format == "nonfinite" {
					dv.Value = "NaN"
				} else {
					dv.Value = 0.0
				}
			}
		}
		return node, nil
	case reflect.String:
		node := document.NewNode()
		node.SetName(name)

		dv := node.AddArgument(nil, "")
		if err := reflectValueToDocumentValue(c, v, dv, format); err != nil {
			return nil, err
		}

		dv.Flag |= document.FlagQuoted
		return node, nil
	default:
		// cain't do nuffin
		return nil, nil
	}

}

func prependArguments(node *document.Node, argStr ...string) {
	args := node.Arguments
	node.Arguments = make([]*document.Value, 0, len(args)+len(argStr))
	for _, arg := range argStr {
		node.AddArgument(arg, "")
	}
	node.Arguments = append(node.Arguments, args...)
}

func marshalMultiSliceToNodes(c *marshalContext, name string, value reflect.Value, fldDetails *structFieldDetails) ([]*document.Node, error) {
	nodes := make([]*document.Node, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		el := reflect.Indirect(value.Index(i))
		if child, err := marshalValueToNode(c, name, el, fldDetails); err != nil {
			return nil, err
		} else if child != nil {
			nodes = append(nodes, child)
		}
	}
	return nodes, nil
}

func marshalMultiMapToNodes(c *marshalContext, names []string, value reflect.Value, fldDetails *structFieldDetails) ([]*document.Node, error) {
	nestedFurther := value.Type().Elem().Kind() == reflect.Map

	nodes := make([]*document.Node, 0, value.Len())
	mapKeys := sortMapKeys(value.MapKeys())

	for _, key := range mapKeys {
		moreNames := make([]string, 0, len(names)+1)
		moreNames = append(moreNames, names...)
		moreNames = append(moreNames, coerce.ToString(key.Interface()))

		el := reflect.Indirect(value.MapIndex(key))
		if nestedFurther {
			if more, err := marshalMultiMapToNodes(c, moreNames, el, fldDetails); err != nil {
				return nil, err
			} else {
				nodes = append(nodes, more...)
			}
		} else {
			if child, err := marshalValueToNode(c, moreNames[0], el, fldDetails); err != nil {
				return nil, err
			} else if child != nil {
				// prepend the map key as the first argument
				prependArguments(child, moreNames[1:]...)
				nodes = append(nodes, child)
			}
		}
	}
	return nodes, nil
}

func marshalValueToNodeOrNodes(c *marshalContext, name string, value reflect.Value, fldDetails *structFieldDetails) ([]*document.Node, error) {
	isMultiple := fldDetails.IsMultiple()
	if isMultiple {
		value = reflect.Indirect(value)
		switch value.Kind() {
		case reflect.Slice:
			return marshalMultiSliceToNodes(c, name, value, fldDetails)

		case reflect.Map:
			return marshalMultiMapToNodes(c, []string{name}, value, fldDetails)

		default:
			return nil, fmt.Errorf("tag `,multiple` used on %s; must be slice or map", value.Type().String())
		}
	} else {
		if node, err := marshalValueToNode(c, name, value, fldDetails); err != nil {
			return nil, err
		} else if node == nil {
			return nil, nil
		} else {
			return []*document.Node{node}, nil
		}
	}
}

func marshalStructToNodes(c *marshalContext, value reflect.Value, nodes []*document.Node) ([]*document.Node, error) {
	typeDetails := c.indexer.Get(value.Type().String())

	if nodes == nil {
		nodes = make([]*document.Node, 0, value.NumField())
	}

	for _, nodeName := range typeDetails.StructFieldNameList {
		if nodeName == "-" {
			continue
		}
		fldDetails := typeDetails.StructFields[nodeName]

		childNodes, err := marshalValueToNodeOrNodes(c, nodeName, fldDetails.GetValueFrom(value), fldDetails)
		if err != nil {
			return nil, err
		} else if childNodes == nil {
			// unhandled
			continue
		} else {
			nodes = append(nodes, childNodes...)
		}
	}

	return nodes, nil
}

func marshalValueToNodes(c *marshalContext, value reflect.Value) ([]*document.Node, error) {
	v := reflect.Indirect(value)

	switch v.Kind() {
	case reflect.Map:
		return marshalMapToNodes(c, value, nil)
	case reflect.Struct:
		return marshalStructToNodes(c, v, nil)
	case reflect.Interface:
		return marshalValueToNodes(c, v.Elem())
	default:
		return nil, fmt.Errorf("cannot marshal Nodes from type %s", v.Type().String())
	}
}
