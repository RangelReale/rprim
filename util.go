package rprim

import "reflect"

func IndirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}
