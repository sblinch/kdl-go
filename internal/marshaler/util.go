package marshaler

import (
	"reflect"
	"strings"
)

func normalizeKey(name string, caseSensitive bool) string {
	name = strings.TrimSpace(name)
	if !caseSensitive {
		name = strings.ToLower(name)
	}
	return name
}

func parseTagName(tag string) string {
	tagspl := strings.SplitN(tag, ",", 2)
	if len(tagspl) > 0 {
		return strings.ToLower(tagspl[0])
	}
	return ""
}
func fieldTagOrName(tag reflect.StructTag, name string, caseSensitive bool) string {
	if tagdata := tag.Get("kdl"); tagdata != "" {
		if fieldname := parseTagName(tagdata); fieldname != "" {
			return fieldname
		}
	}

	return normalizeKey(name, caseSensitive)
}

func fieldAttrs(tag reflect.StructTag) []string {
	if tagdata := tag.Get("kdl"); tagdata != "" {
		attrs := strings.Split(tagdata, ",")
		return attrs[1:]
	} else {
		return nil
	}
}
