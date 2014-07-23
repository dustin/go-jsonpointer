package jsonpointer

import (
	"reflect"
	"strconv"
	"strings"
)

// Reflect gets the value at the specified path from a struct.
func Reflect(o interface{}, path string) interface{} {
	if path == "" {
		return o
	}

	parts := strings.Split(path[1:], "/")
	var rv interface{} = o

OUTER:
	for _, p := range parts {
		val := reflect.ValueOf(rv)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			if strings.Contains(p, "~") {
				p = strings.Replace(p, "~1", "/", -1)
				p = strings.Replace(p, "~0", "~", -1)
			}

			// first look to see if path matches JSON tag name
			typ := val.Type()
			for i := 0; i < typ.NumField(); i++ {
				sf := typ.Field(i)
				tag := sf.Tag.Get("json")
				name := parseJSONTagName(tag)
				if name == p {
					rv = val.Field(i).Interface()
					continue OUTER
				}
			}

			// no JSON tag name matched, look for direct field match
			field := val.FieldByName(p)
			if field.IsValid() {
				rv = field.Interface()
			} else {
				return nil
			}
		} else if val.Kind() == reflect.Slice {
			i, err := strconv.Atoi(p)
			if err == nil && i < val.Len() {
				field := val.Index(i)
				rv = field.Interface()
			} else {
				return nil
			}
		} else {
			return nil
		}
	}

	return rv
}

// ReflectListPointers lists all possible pointers from the given struct.
func ReflectListPointers(o interface{}) ([]string, error) {
	return reflectListPointersRecursive(o, ""), nil
}

func reflectListPointersRecursive(o interface{}, prefix string) []string {
	rv := []string{prefix + ""}

	val := reflect.ValueOf(o)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {

		typ := val.Type()
		for i := 0; i < typ.NumField(); i++ {
			child := val.Field(i).Interface()
			sf := typ.Field(i)
			tag := sf.Tag.Get("json")
			name := parseJSONTagName(tag)
			if name != "" {
				// use the tag name
				childReults := reflectListPointersRecursive(child, prefix+encodePointer([]string{name}))
				rv = append(rv, childReults...)
			} else {
				// use the original field name
				childResults := reflectListPointersRecursive(child, prefix+encodePointer([]string{sf.Name}))
				rv = append(rv, childResults...)
			}
		}

	} else if val.Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			child := val.Index(i).Interface()
			childResults := reflectListPointersRecursive(child, prefix+encodePointer([]string{strconv.Itoa(i)}))
			rv = append(rv, childResults...)
		}
	}

	return rv
}

func parseJSONTagName(tag string) string {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}
