package rprim

// Most of this logic comes from the reflect package (value.go), but the result is different from it.

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const (
	// Whether nil pointer source values can be assigned to non-pointer destination as its zero-value
	COP_ALLOW_NIL_TO_ZERO_VALUE = 1
	// Whether to allow string to slice conversion ([]uint8 or []int32 only)
	COP_ALLOW_STRING_TO_SLICE = 2
	// Whether to allow slice to string conversion ([]uint8 or []int32 only)
	COP_ALLOW_SLICE_TO_SRING = 4
)

// ConvertOp returns the function to convert a primitive value of type src
// to a value of type dst. If the conversion is illegal, ConvertOp returns nil.
// String conversion are supported using the strconv package
// Conversing between pointers are allowed, including of different types.
func ConvertOp(src, dst reflect.Value, flags uint) ConvertOpFunc {
	return NewConfig().SetFlags(flags).ConvertOp(src, dst)
}

func ConvertOpType(src reflect.Value, dstType reflect.Type, flags uint) ConvertOpFunc {
	return NewConfig().SetFlags(flags).ConvertOpType(src, dstType)
}

// Struct to set optional parameters
type Config struct {
	Flags         uint
	FloatFormat   string
	ComplexFormat string
}

func NewConfig() *Config {
	return &Config{
		FloatFormat:   "%f",
		ComplexFormat: "%g",
	}
}

func (c *Config) SetFlags(flags uint) *Config {
	c.Flags = flags
	return c
}

func (c *Config) AddFlags(flags uint) *Config {
	c.Flags |= flags
	return c
}

func (c Config) Dup() *Config {
	return &Config{
		Flags:         c.Flags,
		FloatFormat:   c.FloatFormat,
		ComplexFormat: c.ComplexFormat,
	}
}

// Convert function
type ConvertOpFunc func(reflect.Value, reflect.Type) (reflect.Value, error)

func (c Config) ConvertOp(src, dst reflect.Value) ConvertOpFunc {
	return c.ConvertOpType(src, dst.Type())
}

func (c Config) ConvertOpType(src reflect.Value, dstType reflect.Type) ConvertOpFunc {
	uk_src := UnderliningValueKind(src)
	uk_dst := UnderliningTypeKind(dstType)

	may_be_direct_assignable := (dstType == nil || UnderliningType(src.Type()).AssignableTo(UnderliningType(dstType))) ||
		src.Kind() == reflect.Interface || dstType.Kind() == reflect.Interface

	// these funcions are used to only allow setting nil after all the type compatibility checks are done
	proc_ret := func(f ConvertOpFunc) ConvertOpFunc {
		return f
	}

	proc_ret_nil := func(f ConvertOpFunc) ConvertOpFunc {
		return cvtNil
	}

	// if src is nil, check if dst is nullable
	if (src.Kind() == reflect.Ptr || src.Kind() == reflect.Interface) && src.IsNil() {
		if dstType == nil || dstType.Kind() == reflect.Ptr || dstType.Kind() == reflect.Interface {
			//return cvtNil
			proc_ret = proc_ret_nil
		} else if !((c.Flags & COP_ALLOW_NIL_TO_ZERO_VALUE) == COP_ALLOW_NIL_TO_ZERO_VALUE) {
			return cvtError(errors.New("Copying nil to zero value not allowed"))
		} else {
			//return cvtNil
			proc_ret = proc_ret_nil
		}
	}

	// target type is interface
	if uk_dst == reflect.Interface {
		return proc_ret(cvtAnyInterface)
	}

	// dst and src have same underlying type.
	if may_be_direct_assignable && uk_src == uk_dst && KindIsSimpleValue(uk_src) && KindIsSimpleValue(uk_dst) {
		if dstType == nil || src.Kind() == reflect.Ptr || dstType.Kind() == reflect.Ptr || src.Kind() == reflect.Interface || dstType.Kind() == reflect.Interface {
			return proc_ret(cvtDirectPointer)
		} else {
			return proc_ret(cvtDirect)
		}
	}

	switch uk_src {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch uk_dst {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return proc_ret(cvtInt)
		case reflect.Float32, reflect.Float64:
			return proc_ret(cvtIntFloat)
		case reflect.String:
			return proc_ret(cvtIntString)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch uk_dst {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return proc_ret(cvtUint)
		case reflect.Float32, reflect.Float64:
			return proc_ret(cvtUintFloat)
		case reflect.String:
			return proc_ret(cvtUintString)
		}

	case reflect.Float32, reflect.Float64:
		switch uk_dst {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return proc_ret(cvtFloatInt)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return proc_ret(cvtFloatUint)
		case reflect.Float32, reflect.Float64:
			return proc_ret(cvtFloat)
		case reflect.String:
			return proc_ret(cvtFloatString(c.FloatFormat))
		}

	case reflect.Complex64, reflect.Complex128:
		switch uk_dst {
		case reflect.Complex64, reflect.Complex128:
			return proc_ret(cvtComplex)
		case reflect.String:
			return proc_ret(cvtComplexString(c.ComplexFormat))
		}

	case reflect.String:
		switch uk_dst {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return proc_ret(cvtStringInt)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return proc_ret(cvtStringUint)
		case reflect.Float32, reflect.Float64:
			return proc_ret(cvtStringFloat(c.FloatFormat))
		case reflect.Complex64, reflect.Complex128:
			return proc_ret(cvtStringComplex(c.ComplexFormat))
		case reflect.String:
			return proc_ret(cvtDirectPointer)
		case reflect.Slice:
			if (c.Flags & COP_ALLOW_STRING_TO_SLICE) == COP_ALLOW_STRING_TO_SLICE {
				switch dstType.Elem().Kind() {
				case reflect.Uint8:
					return proc_ret(cvtStringBytes)
				case reflect.Int32:
					return proc_ret(cvtStringRunes)
				}
			}
		}

	case reflect.Slice:
		if uk_dst == reflect.String {
			if (c.Flags & COP_ALLOW_SLICE_TO_SRING) == COP_ALLOW_SLICE_TO_SRING {
				switch src.Elem().Kind() {
				case reflect.Uint8:
					return proc_ret(cvtBytesString)
				case reflect.Int32:
					return proc_ret(cvtRunesString)
				}
			}
		}
	}

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
	root, last := NewUnderliningValue(t)
	var newvalue reflect.Value
	switch last.Kind() {
	case reflect.Uint:
		newvalue = reflect.ValueOf(uint(bits))
	case reflect.Uint8:
		newvalue = reflect.ValueOf(uint8(bits))
	case reflect.Uint16:
		newvalue = reflect.ValueOf(uint16(bits))
	case reflect.Uint32:
		newvalue = reflect.ValueOf(uint32(bits))
	case reflect.Uint64:
		newvalue = reflect.ValueOf(uint64(bits))
	case reflect.Int:
		newvalue = reflect.ValueOf(int(bits))
	case reflect.Int8:
		newvalue = reflect.ValueOf(int8(bits))
	case reflect.Int16:
		newvalue = reflect.ValueOf(int16(bits))
	case reflect.Int32:
		newvalue = reflect.ValueOf(int32(bits))
	case reflect.Int64:
		newvalue = reflect.ValueOf(int64(bits))
	default:
		panic(fmt.Sprintf("Invalid value for makeInt: %s", last.Kind().String()))
	}

	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}

	last.Set(newvalue)
	return root
}

// makeFloat returns a Value of type t equal to v (possibly truncated to float32),
// where t is a float32 or float64 type.
func makeFloat(v float64, t reflect.Type) reflect.Value {
	root, last := NewUnderliningValue(t)
	var newvalue reflect.Value
	switch last.Kind() {
	case reflect.Float32:
		newvalue = reflect.ValueOf(float32(v))
	case reflect.Float64:
		newvalue = reflect.ValueOf(float64(v))
	default:
		panic("Invalid value for makeFloat")
	}

	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}
	last.Set(newvalue)

	return root
}

// makeComplex returns a Value of type t equal to v (possibly truncated to complex64),
// where t is a complex64 or complex128 type.
func makeComplex(v complex128, t reflect.Type) reflect.Value {
	root, last := NewUnderliningValue(t)
	var newvalue reflect.Value
	switch last.Kind() {
	case reflect.Complex64:
		newvalue = reflect.ValueOf(complex64(v))
	case reflect.Complex128:
		newvalue = reflect.ValueOf(complex128(v))
	default:
		panic("Invalid value for makeComplex")
	}

	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}
	last.Set(newvalue)

	return root
}

func makeString(v string, t reflect.Type) reflect.Value {
	root, last := NewUnderliningValue(t)
	newvalue := reflect.ValueOf(v)
	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}
	last.Set(newvalue)
	return root
}

func makeBytes(v []byte, t reflect.Type) reflect.Value {
	root, last := NewUnderliningValue(t)
	newvalue := reflect.ValueOf(v)
	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}
	last.Set(newvalue)
	return root
}

func makeRunes(v []rune, t reflect.Type) reflect.Value {
	root, last := NewUnderliningValue(t)
	newvalue := reflect.ValueOf(string(v))
	if !newvalue.Type().AssignableTo(last.Type()) {
		// named values are not directly assignable
		newvalue = newvalue.Convert(last.Type())
	}
	last.Set(newvalue)
	return root
}

func cvtError(err error) ConvertOpFunc {
	return func(v reflect.Value, t reflect.Type) (reflect.Value, error) {
		return reflect.Value{}, err
	}
}

// These conversion functions are returned by ConvertOp
// for classes of conversions. For example, the first function, cvtInt,
// takes any value v of signed int type and returns the value converted
// to type t, where t is any signed or unsigned int type.

// ConvertOp: intXX -> [u]intXX
func cvtInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(UnderliningValue(v).Int()), t), nil
}

// ConvertOp: uintXX -> [u]intXX
func cvtUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(UnderliningValue(v).Uint(), t), nil
}

// ConvertOp: floatXX -> intXX
func cvtFloatInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(int64(UnderliningValue(v).Float())), t), nil
}

// ConvertOp: floatXX -> uintXX
func cvtFloatUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeInt(uint64(UnderliningValue(v).Float()), t), nil
}

// ConvertOp: intXX -> floatXX
func cvtIntFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(float64(UnderliningValue(v).Int()), t), nil
}

// ConvertOp: uintXX -> floatXX
func cvtUintFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(float64(UnderliningValue(v).Uint()), t), nil
}

// ConvertOp: floatXX -> floatXX
func cvtFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeFloat(UnderliningValue(v).Float(), t), nil
}

// ConvertOp: complexXX -> complexXX
func cvtComplex(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeComplex(UnderliningValue(v).Complex(), t), nil
}

// ConvertOp: intXX -> string
func cvtIntString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(strconv.FormatInt(UnderliningValue(v).Int(), 10), t), nil
}

// ConvertOp: uintXX -> string
func cvtUintString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(strconv.FormatUint(UnderliningValue(v).Uint(), 10), t), nil
}

// ConvertOp: floatXX -> string
func cvtFloatString(format string) ConvertOpFunc {
	return func(v reflect.Value, t reflect.Type) (reflect.Value, error) {
		return makeString(fmt.Sprintf(format, UnderliningValue(v).Float()), t), nil
	}
}

/*
func cvtFloatString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(fmt.Sprintf("%f", IndirectPtrInterface(v).Float()), t), nil
}
*/

// ConvertOp: complexXX -> string
func cvtComplexString(format string) ConvertOpFunc {
	return func(v reflect.Value, t reflect.Type) (reflect.Value, error) {
		return makeString(fmt.Sprintf(format, UnderliningValue(v).Complex()), t), nil
	}
}

/*
func cvtComplexString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(fmt.Sprintf("%g", IndirectPtrInterface(v).Complex()), t), nil
}
*/

// ConvertOp: []byte -> string
func cvtBytesString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(string(UnderliningValue(v).Bytes()), t), nil
}

// ConvertOp: string -> intXX
func cvtStringInt(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	cv, err := strconv.ParseInt(UnderliningValue(v).String(), 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to int: %v", err)
	}
	return makeInt(uint64(cv), t), nil
}

// ConvertOp: string -> uintXX
func cvtStringUint(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	cv, err := strconv.ParseUint(UnderliningValue(v).String(), 10, 64)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to uint: %v", err)
	}
	return makeInt(uint64(cv), t), nil
}

// ConvertOp: string -> floatXX
func cvtStringFloat(format string) ConvertOpFunc {
	return func(v reflect.Value, t reflect.Type) (reflect.Value, error) {
		var cv float64
		_, err := fmt.Sscanf(UnderliningValue(v).String(), format, &cv)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Error converting string to float: %v", err)
		}
		return makeFloat(float64(cv), t), nil
	}
}

/*
func cvtStringFloat(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	var cv float64
	_, err := fmt.Sscanf(IndirectPtrInterface(v).String(), "%f", &cv)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to float: %v", err)
	}
	return makeFloat(float64(cv), t), nil
}
*/

// ConvertOp: string -> complexXX

func cvtStringComplex(format string) ConvertOpFunc {
	return func(v reflect.Value, t reflect.Type) (reflect.Value, error) {
		var cv complex128
		_, err := fmt.Sscanf(UnderliningValue(v).String(), format, &cv)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("Error converting string to complex: %v", err)
		}
		return makeComplex(complex128(cv), t), nil
	}
}

/*
func cvtStringComplex(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	var cv complex128
	_, err := fmt.Sscanf(IndirectPtrInterface(v).String(), "%g", &cv)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("Error converting string to complex: %v", err)
	}
	return makeComplex(complex128(cv), t), nil
}
*/

// ConvertOp: string -> []byte
func cvtStringBytes(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeBytes([]byte(UnderliningValue(v).String()), t), nil
}

// ConvertOp: []rune -> string
func cvtRunesString(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeString(string(UnderliningValue(v).Interface().([]int32)), t), nil
}

// ConvertOp: string -> []rune
func cvtStringRunes(v reflect.Value, t reflect.Type) (reflect.Value, error) {
	return makeRunes([]rune(UnderliningValue(v).String()), t), nil
}

// ConvertOp: direct copy
func cvtDirect(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	root, last := NewUnderliningValue(typ)
	last.Set(v)
	return root, nil
}

// ConvertOp: direct copy with pointers involved
func cvtDirectPointer(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	root, last := NewUnderliningValue(typ)
	last.Set(UnderliningValue(v))
	return root, nil
}

// converOp: nil when source value is nil
func cvtNil(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	return reflect.Zero(typ), nil
}

// converOp: any type to interface{}
func cvtAnyInterface(v reflect.Value, typ reflect.Type) (reflect.Value, error) {
	x := reflect.New(v.Type()).Elem()
	x.Set(v)
	return x, nil
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
