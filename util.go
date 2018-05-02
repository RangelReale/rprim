package rprim

import (
	"reflect"
)

func UnderliningValueKind(v reflect.Value) reflect.Kind {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	// if there is nil at any point, must walk using the type instead
	vt := v.Type()
	for vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}
	return vt.Kind()
}

func UnderliningValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	return v
}

func UnderliningTypeKind(v reflect.Type) reflect.Kind {
	if v == nil {
		return reflect.Interface
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Kind()
}

func NewUnderliningValue(v reflect.Type) (root reflect.Value, last reflect.Value) {
	root = reflect.Value{}
	last = reflect.Value{}
	for {
		newi := reflect.New(v)
		if !root.IsValid() {
			root = newi.Elem()
		} else {
			last.Set(newi)
		}
		last = newi.Elem()
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		} else {
			break
		}
	}

	return
}

func KindIsSimpleValue(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	default:
		return false
	}
}

func IndirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

/*
func IndirectTypeLast(v reflect.Type) reflect.Type {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func IndirectPtrInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		return v
	}
	return v.Elem()
}

func IndirectPtrInterfaceLast(v reflect.Value) reflect.Value {
	for (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) && !v.IsNil() {
		v = v.Elem()
	}
	return v
}
*/
