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

func UnderliningTypeKind(v reflect.Type) reflect.Kind {
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
		}
		if last.IsValid() {
			last.Set(newi)
		} else {
			last = newi.Elem()
		}
		if v.Kind() == reflect.Ptr {
			last = last.Elem()
			v = v.Elem()
		} else {
			break
		}
	}

	return
}

func IndirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}

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
