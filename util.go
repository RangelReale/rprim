package rprim

import (
	"errors"
	"reflect"
)

// Returns the underlining kind of the value, after all pointer and interface dereferences.
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

// Returns the underlining value, after all pointer and interface dereferences.
func UnderliningValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	return v
}

// Returns if the underlining value is nil, even in any amount of pointer indirection
func UnderliningValueIsNil(v reflect.Value) bool {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	return false
}

// Returns the underlining kind of the type, after all pointer and interface dereferences.
func UnderliningTypeKind(v reflect.Type) reflect.Kind {
	if v == nil {
		return reflect.Interface
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Kind()
}

// Returns the underlining type, after all pointer and interface dereferences.
func UnderliningType(v reflect.Type) reflect.Type {
	if v == nil {
		return nil
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

// Creates a new instance of the type, returning the root item, and the last if the type contains any
// pointer or interface, else returns the same as root.
// The last item is always de zero-value of the type (nil if value is nillable).
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

// Ensure that all the passed value pointer indirections are not nil, and returns the last
// non-pointer value.
func EnsureUnderliningValue(v reflect.Value) (last reflect.Value, err error) {
	cur := v
	for cur.Kind() == reflect.Ptr {
		if cur.IsNil() {
			if cur.CanSet() {
				cur.Set(reflect.New(cur.Type().Elem()))
			} else {
				return reflect.Value{}, errors.New("Pointer value is not settable")
			}
		}
		cur = cur.Elem()
	}
	return cur, nil
}

// Checks if the kind is a simple value (no array, slice, interface, map or chan).
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

// Equivalent to reflect.Indirect, but for reflect.Type.
func IndirectType(v reflect.Type) reflect.Type {
	if v.Kind() == reflect.Ptr {
		return v.Elem()
	}
	return v
}
