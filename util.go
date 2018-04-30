package rprim

import "reflect"

func IndirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

func IndirectPtrInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		return v
	}
	return v.Elem()
}
