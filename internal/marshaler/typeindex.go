package marshaler

// typeIndexer creates an index of frequently-used information about each Golang type into which
// configurations will be unmarshaled

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/sblinch/kdl-go/document"
)

type structFieldAttrs []string

func (s structFieldAttrs) Has(attr string) bool {
	for _, v := range s {
		if v == attr {
			return true
		}
	}
	return false
}

type structFieldDetails struct {
	FieldIndex int   // index of the field in the struct
	EmbedIndex []int // if non-nil, the index(es) of the embedded struct to which FieldIndex refers
	Format     string
	Attrs      structFieldAttrs
}

func (f *structFieldDetails) GetValueFrom(structVal reflect.Value) reflect.Value {
	for _, embedIndex := range f.EmbedIndex {
		structVal = structVal.Field(embedIndex)
	}
	return structVal.Field(f.FieldIndex)
}

func (f *structFieldDetails) IsMultiple() bool {
	return f.Attrs.Has("multiple")
}

func (f *structFieldDetails) IsCapture() bool {
	for _, v := range f.Attrs {
		switch v {
		case "arg", "args", "props", "children":
			return true
		}
	}
	return false
}

type typeDetails struct {
	StructFields              map[string]*structFieldDetails   // if this is a struct type, this is an index of the field names and their indexes
	StructAttrs               map[string][]*structFieldDetails // if this is a struct type, this is map of attribute names to a list of fields that have this attribute
	StructFieldNameList       []string                         // if this is a struct type, this is a list of field names in order
	StructureStructField      *structFieldDetails              // if this is a struct type that includes a "kdl:,structure" field, this identifies that field
	TextUnmarshalerMethod     int                              // index of the UnmarshalText method, if this type satisfies the encoding.TextUnmarshaler interface
	KDLUnmarshalerMethod      int                              // index of the UnmarshalKDL method, if this type satisfies the kdl.Unmarshaler interface
	KDLValueUnmarshalerMethod int                              // index of the UnmarshalKDLValue method, if this type satisfies the kdl.ValueUnmarshaler interface
	TextMarshalerMethod       int                              // index of the MarshalText method, if this type satisfies the encoding.TextMarshaler interface
	KDLMarshalerMethod        int                              // index of the MarshalKDL method, if this type satisfies the kdl.Marshaler interface
	KDLValueMarshalerMethod   int                              // index of the MarshalKDLValue method, if this type satisfies the kdl.ValueMarshaler interface
}

func (t *typeDetails) CanUnmarshalText() bool {
	return t.TextUnmarshalerMethod != -1
}
func (t *typeDetails) CanUnmarshalKDL() bool {
	return t.KDLUnmarshalerMethod != -1
}
func (t *typeDetails) CanUnmarshalKDLValue() bool {
	return t.KDLValueUnmarshalerMethod != -1
}
func (t *typeDetails) CanMarshalText() bool {
	return t.TextMarshalerMethod != -1
}
func (t *typeDetails) CanMarshalKDL() bool {
	return t.KDLMarshalerMethod != -1
}
func (t *typeDetails) CanMarshalKDLValue() bool {
	return t.KDLValueMarshalerMethod != -1
}

func newTypeDetails() *typeDetails {
	return &typeDetails{
		TextUnmarshalerMethod:     -1,
		KDLUnmarshalerMethod:      -1,
		KDLValueUnmarshalerMethod: -1,
		TextMarshalerMethod:       -1,
		KDLMarshalerMethod:        -1,
		KDLValueMarshalerMethod:   -1,
	}
}

func (d *typeDetails) Dump() {
	for k, v := range d.StructFields {
		fmt.Printf("  [%s](%d)=%#v\n", k, len(k), v)
	}

}

type structStructure struct {
	Comments map[string]*document.Comment
}

func (s *structStructure) getKey(name string, node *document.Node) string {
	if node == nil || len(node.Children) == 0 || (len(node.Arguments) == 0 && len(node.Properties) == 0) {
		return name
	}
	b := make([]byte, 0, len(name)+1+len(node.Arguments)*32+1+len(node.Properties)*64)
	b = append(b, name...)
	b = append(b, ':')
	for i, arg := range node.Arguments {
		if i > 0 {
			b = append(b, ',')
		}
		b = arg.AppendTo(b)
	}
	b = append(b, ':')
	b = node.Properties.AppendTo(b)
	return string(b)
}

func (s *structStructure) Set(name string, node *document.Node) {
	if node.Comment == nil {
		return
	}

	key := s.getKey(name, node)
	s.Comments[key] = node.Comment
}

func (s *structStructure) Get(name string, node *document.Node) *document.Comment {
	key := s.getKey(name, node)
	return s.Comments[key]
}

func (t *typeDetails) getStructureStructField(structValue reflect.Value, create bool) *structStructure {
	if t.StructureStructField == nil {
		return nil
	}
	destFieldValue := t.StructureStructField.GetValueFrom(structValue)
	existingIntf := destFieldValue.Interface()

	var ss *structStructure
	if existingIntf != nil {
		ss, _ = existingIntf.(*structStructure)
	}
	if ss == nil && create {
		ss = &structStructure{Comments: make(map[string]*document.Comment)}
		destFieldValue.Set(reflect.ValueOf(ss))
	}
	return ss
}

func (t *typeDetails) SetStructure(destStruct reflect.Value, name string, node *document.Node) {
	ss := t.getStructureStructField(destStruct, true)
	if ss != nil {
		ss.Set(name, node)
	}
}

func (t *typeDetails) GetStructure(structValue reflect.Value) *structStructure {
	return t.getStructureStructField(structValue, false)
}

type typeIndexer struct {
	index         map[string]*typeDetails
	caseSensitive bool
}

func newTypeIndexer(caseSensitive bool) *typeIndexer {
	return &typeIndexer{
		index:         make(map[string]*typeDetails),
		caseSensitive: caseSensitive,
	}
}

func (i *typeIndexer) Dump() {
	for k, v := range i.index {
		fmt.Printf("[%s](%d)=%#v\n", k, len(k), v)
	}
}

func Debug(s string, v ...interface{}) {
	// fmt.Printf(s, v...)
	// fmt.Println()
}

// Indexes a type and (in the case of structs, maps, pointers, etc.) any of the types it contains
func (i *typeIndexer) indexType(typ reflect.Type) error {
	for ; typ != nil && typ.Kind() == reflect.Ptr; typ = typ.Elem() {
	}
	Debug("  indexType: %s", typ)

	if typ == nil {
		return nil
	}

	typName := typ.String()
	if _, exists := i.index[typName]; exists {
		Debug("    already indexed, skipping")
		return nil
	}

	typeDetails := newTypeDetails()
	i.index[typName] = typeDetails

	ptrTyp := reflect.PtrTo(typ)
	if ptrTyp.NumMethod() > 0 {
		Debug("    have methods on type %s", typName)
		v := reflect.New(typ)
		intf := v.Interface()

		for i := 0; i < ptrTyp.NumMethod(); i++ {
			switch ptrTyp.Method(i).Name {
			case "UnmarshalText":
				if _, ok := intf.(encoding.TextUnmarshaler); ok {
					typeDetails.TextUnmarshalerMethod = i
				}
			case "UnmarshalKDL":
				if _, ok := intf.(unmarshaler); ok {
					typeDetails.KDLUnmarshalerMethod = i
				}
			case "UnmarshalKDLValue":
				if _, ok := intf.(valueUnmarshaler); ok {
					typeDetails.KDLValueUnmarshalerMethod = i
				}
			case "MarshalText":
				if _, ok := intf.(encoding.TextMarshaler); ok {
					typeDetails.TextMarshalerMethod = i
				}
			case "MarshalKDL":
				if _, ok := intf.(marshaler); ok {
					typeDetails.KDLMarshalerMethod = i
				}
			case "MarshalKDLValue":
				if _, ok := intf.(valueMarshaler); ok {
					typeDetails.KDLValueMarshalerMethod = i
				}
			}
		}

	} else {
		Debug("    have no methods on type %s", typ.String())
	}

	switch typ.Kind() {
	case reflect.Map:
		Debug("    this is a map: Map's key type is: %s, value type is: %s\n", typ.Key().String(), typ.Elem().String())
		if err := i.indexType(typ.Key()); err != nil {
			return err
		}
		if err := i.indexType(typ.Elem()); err != nil {
			return err
		}

	case reflect.Struct:
		Debug("    this is a struct: %s", typ.String())

		typeDetails.StructFields = make(map[string]*structFieldDetails)
		typeDetails.StructAttrs = make(map[string][]*structFieldDetails)
		typeDetails.StructFieldNameList = make([]string, 0, typ.NumField())

		return i.indexStructFields(typ, typeDetails, nil)

	case reflect.Slice, reflect.Array:
		Debug("    this is a slice: slice's element type is: %s\n", typ.Elem().String())
		if err := i.indexType(typ.Elem()); err != nil {
			return err
		}
	}

	return nil
}

var errUnexportedStructure = errors.New("fields tagged kdl:\",structure\" must be exported")

func (i *typeIndexer) indexStructFields(typ reflect.Type, typeDetails *typeDetails, embedIndexes []int) error {
	numFields := typ.NumField()
	for n := 0; n < numFields; n++ {
		field := typ.Field(n)
		ft := field.Type

		if field.Type.Kind() == reflect.Struct && field.Anonymous {
			ei := append(append([]int(nil), embedIndexes...), n)
			if err := i.indexStructFields(field.Type, typeDetails, ei); err != nil {
				return err
			}
			continue
		}

		attrs := fieldAttrs(field.Tag)
		structure := slices.Contains(attrs, "structure")
		if !field.IsExported() {
			if structure {
				return errUnexportedStructure
			}
			continue
		}

		normalized := fieldTagOrName(field.Tag, field.Name, i.caseSensitive)
		Debug("  field %s (normalized %s, type %s) is at index %d", field.Name, normalized, ft.String(), n)

		fld := &structFieldDetails{
			FieldIndex: n,
			EmbedIndex: embedIndexes,
			Format:     "",
			Attrs:      nil,
		}

		fld.Attrs = attrs
		for _, name := range fld.Attrs {
			if strings.HasPrefix(name, "format:") {
				fld.Format = strings.TrimPrefix(name, "format:")
			}
			typeDetails.StructAttrs[name] = append(typeDetails.StructAttrs[name], fld)
		}

		if structure {
			typeDetails.StructureStructField = fld
		} else {
			typeDetails.StructFields[normalized] = fld
			typeDetails.StructFieldNameList = append(typeDetails.StructFieldNameList, normalized)

			if err := i.indexType(ft); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i *typeIndexer) IndexIntf(dest interface{}) error {
	Debug("=== INDEXING ===")
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	var t reflect.Type
	if v.Kind() == reflect.Ptr {
		t = v.Type().Elem()
	} else if !v.IsValid() {
		return nil
	} else {
		t = v.Type()
	}

	return i.indexType(t)
}

func (i *typeIndexer) Get(name string) *typeDetails {
	if len(name) == 0 {
		return nil
	}
	if name[0] == '*' {
		name = name[1:]
	}
	if v, ok := i.index[name]; ok {
		Debug("typeIndexer \"%s\" exists", name)
		return v
	} else {
		Debug("typeIndexer \"%s\" does not exist", name)
		return nil
	}
}

func (i *typeIndexer) GetEmpty() *typeDetails {
	d := newTypeDetails()
	d.StructFields = make(map[string]*structFieldDetails)
	d.StructAttrs = make(map[string][]*structFieldDetails)
	d.StructFieldNameList = make([]string, 0, 8)
	return d
}
