package marshaler

import (
	"reflect"
	"strings"
)

func normalizeKey(name string) string {
	out := make([]byte, len(name))
	o := 0
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			out[o] = c + 32
			o++
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || (c == '_') {
			out[o] = c
			o++
		}
	}
	return string(out[0:o])
}

func parseTagName(tag string) string {
	tagspl := strings.SplitN(tag, ",", 2)
	if len(tagspl) > 0 {
		return strings.ToLower(tagspl[0])
	}
	return ""
}
func fieldTagOrName(tag reflect.StructTag, name string) string {
	if tagdata := tag.Get("kdl"); tagdata != "" {
		if fieldname := parseTagName(tagdata); fieldname != "" {
			return fieldname
		}
	}

	return normalizeKey(strings.ToLower(name))
}

func fieldAttrs(tag reflect.StructTag) []string {
	if tagdata := tag.Get("kdl"); tagdata != "" {
		attrs := strings.Split(tagdata, ",")
		return attrs[1:]
	} else {
		return nil
	}
}
