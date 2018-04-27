package rprim

// Most of this logic comes from the reflect package (value.go), but the result is different from it.

import (
	"fmt"
	"reflect"
	"strconv"
)

const (
	// Whether nil pointer source values can be assigned to non-pointer destination as its zero-value
	COP_ALLOW_NIL_TO_ZERO = 1
)

// Convert function
type ConvertOpFunc func(reflect.Value, reflect.Type) (reflect.Value, error)

// ConvertOp returns the function to convert a primitive value of type src
// to a value of type dst. If the conversion is illegal, ConvertOp returns nil.
// String conversion are supported using the strconv package
// Conversing between pointers are allowed, including of different types.
func ConvertOp(dst, src reflect.Value, flags uint) ConvertOpFunc {
	srckind := indirectType(src.Type()).Kind()
	dstkind := indirectType(dst.Type()).Kind()

	if src.Kind() == reflect.Ptr && src.IsNil() {
		if dst.Kind() != reflect.Ptr && !((flags & COP_ALLOW_NIL_TO_ZERO) == COP_ALLOW_NIL_TO_ZERO) {
			return nil
		}
		return cvtNil
	}

	// dst and src have same underlying type.
	if dstkind == srckind {
		if src.Kind() == reflect.Ptr || dst.Kind() == reflect.Ptr {
			return cvtDirectPointer
		} else {
			return cvtDirect
		}
	}

	switch srckind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch dstkind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtInt
		case reflect.Float32, reflect.Float64:
			return cvtIntFloat
		case reflect.String:
			return cvtIntString
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch dstkind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtUint
		case reflect.Float32, reflect.Float64:
			return cvtUintFloat
		case reflect.String:
			return cvtUintString
		}

	case reflect.Float32, reflect.Float64:
		switch dstkind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return cvtFloatInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtFloatUint
		case reflect.Float32, reflect.Float64:
			return cvtFloat
		}

	case reflect.Complex64, reflect.Complex128:
		switch dstkind {
		case reflect.Complex64, reflect.Complex128:
			return cvtComplex
		}

	case reflect.String:
		switch dstkind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return cvtStringInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return cvtStringUint
			/*
				case reflect.Slice:
					switch dst.Elem().Kind() {
					case reflect.Uint8:
						return cvtStringBytes
					case reflect.Int32:
						return cvtStringRunes
					}
			*/
		}
	case reflect.Slice:
		if dstkind == reflect.String /*&& src.Elem().PkgPath() == ""*/ {
			switch src.Elem().Kind() {
			case reflect.Uint8:
				return cvtBytesString
				//case reflect.Int32:
				//return cvtRunesString
			}
		}
	}

	// dst and src are unnamed pointer types with same underlying base type.
	/*
		if dstkind == reflect.Ptr && dst.Name() == "" &&
			srckind == reflect.Ptr && src.Name() == "" &&
			haveIdenticalUnderlyingType(dst.Elem().common(), src.Elem().common(), false) {
			return cvtDirect
		}
	*/

	/*
		if implements(dst, src) {
			if srckind == Interface {
				return cvtI2I
			}
			return cvtT2I
		}
	*/

	return nil
}

// makeInt returns a Value of type t equal to bits (possibly truncated),
// where t is a signed or unsigned int type.
func makeInt(bits uint64, t reflect.Type) reflect.Value {
	xt := indirectType(t)

	x := reflect.New(indirectType(t)).Elem()
	switch xt.Kind() {
	case reflect.Uint:
		x.Set(reflect.ValueOf(uint(bits)))
	case reflect.Uint8:
		x.Set(reflect.ValueOf(uint8(bits)))
	case reflect.Uint16:
		x.Set(reflect.ValueOf(uint16(bits)))
	case reflect.Uint32:
		x.Set(reflect.ValueOf(uint32(bits)))
	case reflect.Uint64:
		x.Set(reflect.ValueOf(uint64(bits)))
	case reflect.Int:
		x.Set(reflect.ValueOf(int(bits)))
	case reflect.Int8:
		x.Set(reflect.ValueOf(int8(bits)))
	case reflect.Int16:
		x.Set(reflect.ValueOf(int16(bits)))
	case reflect.Int32:
		x.Set(reflect.ValueOf(int32(bits)))
	case reflect.Int64:
		x.Set(reflect.ValueOf(int64(bits)))
	default:
		panic("Invalid value for makeInt")
	}
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

// makeFloat returns a Value of type t equal to v (possibly truncated to float32),
// where t is a float32 or float64 type.
func makeFloat(v float64, t reflect.Type) reflect.Value {
	xt := indirectType(t)

	x := reflect.New(xt).Elem()
	switch xt.Kind() {
	case reflect.Float32:
		x.Set(reflect.ValueOf(float32(v)))
	case reflect.Float64:
		x.Set(reflect.ValueOf(float64(v)))
	default:
		panic("Invalid value for makeFloat")
	}
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

// makeComplex returns a Value of type t equal to v (possibly truncated to complex64),
// where t is a complex64 or complex128 type.
func makeComplex(v complex128, t reflect.Type) reflect.Value {
	xt := indirectType(t)

	x := reflect.New(xt).Elem()
	switch xt.Kind() {
	case reflect.Complex64:
		x.Set(reflect.ValueOf(complex64(v)))
	case reflect.Complex128:
		x.Set(reflect.ValueOf(complex128(v)))
	default:
		panic("Invalid value for makeComplex")
	}
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

func makeString(v string, t reflect.Type) reflect.Value {
	xt := indirectType(t)
	x := reflect.New(xt).Elem()
	x.SetString(v)
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

func makeBytes(v []byte, t reflect.Type) reflect.Value {
	xt := indirectType(t)
	x := reflect.New(xt).Elem()
	x.SetBytes(v)
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

func makeRunes(v []rune, t reflect.Type) reflect.Value {
	xt := indirectType(t)
	x := reflect.New(xt).Elem()
	x.SetString(string(v))
	if t.Kind() == reflect.Ptr {
		return x.Addr()
	}
	return x
}

// These conversion functions are returned by ConvertOp
// for classes of conversions. For example, the first function, cvtInt,
// takes any value v of signed int type and returns the value converted
// to type t, where t is any signed or unsigned int type.

// ConvertOp: intXX -> [u]intXX
func cvtInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(reflect.Indirect(v).Int()), t), nil
}

// ConvertOp: uintXX -> [u]intXX
func cvtUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(reflect.Indirect(v).Uint(), t), nil
}

// ConvertOp: floatXX -> intXX
func cvtFloatInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(int64(reflect.Indirect(v).Float())), t), nil
}

// ConvertOp: floatXX -> uintXX
func cvtFloatUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(reflect.Indirect(v).Float()), t), nil
}

// ConvertOp: intXX -> floatXX
func cvtIntFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(float64(reflect.Indirect(v).Int()), t), nil
}

// ConvertOp: uintXX -> floatXX
func cvtUintFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(float64(reflect.Indirect(v).Uint()), t), nil
}

// ConvertOp: floatXX -> floatXX
func cvtFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(reflect.Indirect(v).Float(), t), nil
}

// ConvertOp: complexXX -> complexXX
func cvtComplex(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeComplex(reflect.Indirect(v).Complex(), t), nil
}

// ConvertOp: intXX -> string
func cvtIntString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(string(reflect.Indirect(v).Int()), t), nil
}

// ConvertOp: uintXX -> string
func cvtUintString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(string(reflect.Indirect(v).Uint()), t), nil
}

// ConvertOp: []byte -> string
func cvtBytesString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(string(reflect.Indirect(v).Bytes()), t), nil
}

// ConvertOp: string -> int
func cvtStringInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	cv, err := strconv.ParseInt(reflect.Indirect(v).String(), 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to int: %v", err)
	}
	return makeInt(uint64(cv), t), nil
}

// ConvertOp: string -> uint
func cvtStringUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	cv, err := strconv.ParseUint(reflect.Indirect(v).String(), 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to uint: %v", err)
	}
	return makeInt(uint64(cv), t), nil
}

// ConvertOp: string -> []byte
func cvtStringBytes(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeBytes([]byte(reflect.Indirect(v).String()), t), nil
}

// ConvertOp: []rune -> string
/*
func cvtRunesString(v reflect.Value, t reflect.Type) reflect.Value {
	return makeString(string(v.runes()), t)
}
*/

// ConvertOp: string -> []rune
func cvtStringRunes(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeRunes([]rune(reflect.Indirect(v).String()), t), nil
}

// ConvertOp: direct copy
func cvtDirect(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	x := reflect.New(typ).Elem()
	x.Set(v)
	return x, nil
}

// ConvertOp: direct copy with pointers involved
func cvtDirectPointer(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	xt := indirectType(typ)
	x := reflect.New(xt).Elem()
	x.Set(reflect.Indirect(v))
	if typ.Kind() == reflect.Ptr {
		return x.Addr(), nil
	}
	return x, nil
}

// converOp: nil when source value is nil
func cvtNil(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	return reflect.Zero(typ), nil
}

// ConvertOp: concrete -> interface
/*
func cvtT2I(v Value, typ Type) Value {
	target := unsafe_New(typ.common())
	x := valueInterface(v, false)
	if typ.NumMethod() == 0 {
		*(*interface{})(target) = x
	} else {
		ifaceE2I(typ.(*rtype), x, target)
	}
	return Value{typ.common(), target, v.flag.ro() | flagIndir | flag(Interface)}
}

// ConvertOp: interface -> interface
func cvtI2I(v Value, typ Type) Value {
	if v.IsNil() {
		ret := Zero(typ)
		ret.flag |= v.flag.ro()
		return ret
	}
	return cvtT2I(v.Elem(), typ)
}
*/
