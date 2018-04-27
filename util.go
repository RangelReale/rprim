package rprim

import "reflect"

func indirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}
