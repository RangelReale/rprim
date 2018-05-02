package rprim

import (
	"fmt"
	"reflect"
	"testing"
)

type testValueType int

const (
	VT_INT testValueType = iota
	VT_UINT
	VT_FLOAT
	VT_COMPLEX
	VT_STRING
)

type testValue struct {
	valueType testValueType
	name      string
	value     interface{}
}

var (
	test_int_value     = 109
	test_uint_value    = uint(109)
	test_float_value   = float64(109)
	test_string_value  = "109"
	test_complex_value = complex128(109)
)

func getTestList() []*testValue {
	return []*testValue{
		{VT_INT, "int", int(test_int_value)},
		{VT_INT, "int8", int8(test_int_value)},
		{VT_INT, "int16", int16(test_int_value)},
		{VT_INT, "int32", int32(test_int_value)},
		{VT_INT, "int64", int64(test_int_value)},
		{VT_UINT, "uint", uint(test_uint_value)},
		{VT_UINT, "uint8", uint8(test_uint_value)},
		{VT_UINT, "uint16", uint16(test_uint_value)},
		{VT_UINT, "uint32", uint32(test_uint_value)},
		{VT_UINT, "uint64", uint64(test_uint_value)},
		{VT_FLOAT, "float32", float32(test_float_value)},
		{VT_FLOAT, "float64", float64(test_float_value)},
		{VT_COMPLEX, "complex64", complex64(test_complex_value)},
		{VT_COMPLEX, "complex128", complex128(test_complex_value)},
		{VT_STRING, "string", test_string_value},
	}
}

func getTestListByType(valueType testValueType) []*testValue {
	var ret []*testValue
	for _, test := range getTestList() {
		if test.valueType == valueType {
			ret = append(ret, test)
		}
	}
	return ret
}

func TestPrimitives(t *testing.T) {
	tests := getTestList()
	for _, target_test := range tests {
		for _, source_test := range tests {
			if source_test.valueType != VT_COMPLEX && target_test.valueType != VT_COMPLEX {
				//fmt.Printf("Converting %s to %s\n", source_test.name, target_test.name)

				sourceV := reflect.ValueOf(source_test.value)
				targetT := reflect.TypeOf(target_test.value)

				cop := NewConfig().ConvertOpType(sourceV, targetT)
				if cop == nil {
					t.Fatalf("Converter not found for %s to %s", source_test.name, target_test.name)
				}

				copValue, err := cop(sourceV, targetT)
				if err != nil {
					t.Fatal(err)
				}

				expected := target_test.value
				if source_test.valueType == VT_FLOAT && target_test.valueType == VT_STRING {
					expected = fmt.Sprintf("%f", source_test.value)
				}

				if copValue.Interface() != expected {
					t.Fatalf("Values are different converting %s to %s: got %v expected %v",
						source_test.name, target_test.name, copValue.Interface(), target_test.value)
				}
			}
		}
	}
}

func TestPointerTarget(t *testing.T) {
	tests := getTestList()
	for _, target_test := range tests {
		for _, source_test := range tests {
			if source_test.valueType != VT_COMPLEX && target_test.valueType != VT_COMPLEX {
				//fmt.Printf("Converting %s to *%s\n", source_test.name, target_test.name)

				sourceV := reflect.ValueOf(source_test.value)
				targetT := reflect.PtrTo(reflect.TypeOf(target_test.value))

				cop := NewConfig().ConvertOpType(sourceV, targetT)
				if cop == nil {
					t.Fatalf("Converter not found for %s to *%s", source_test.name, target_test.name)
				}

				copValue, err := cop(sourceV, targetT)
				if err != nil {
					t.Fatal(err)
				}

				if copValue.Kind() != reflect.Ptr {
					t.Fatalf("Expected a pointer, got %s", copValue.Kind().String())
				}

				expected := target_test.value
				if source_test.valueType == VT_FLOAT && target_test.valueType == VT_STRING {
					expected = fmt.Sprintf("%f", source_test.value)
				}

				if copValue.Elem().Interface() != expected {
					t.Fatalf("Values are different converting %s to *%s: got %v expected %v",
						source_test.name, target_test.name, copValue.Elem().Interface(), target_test.value)
				}
			}
		}
	}
}

func TestPointerSource(t *testing.T) {
	tests := getTestList()
	for _, target_test := range tests {
		for _, source_test := range tests {
			if source_test.valueType != VT_COMPLEX && target_test.valueType != VT_COMPLEX {
				//fmt.Printf("Converting *%s to %s\n", source_test.name, target_test.name)

				var sourceI interface{}
				// create pointer of value
				sourceI = &source_test.value

				sourceV := reflect.ValueOf(sourceI)
				targetT := reflect.TypeOf(target_test.value)

				cop := NewConfig().ConvertOpType(sourceV, targetT)
				if cop == nil {
					t.Fatalf("Converter not found for *%s to %s", source_test.name, target_test.name)
				}

				copValue, err := cop(sourceV, targetT)
				if err != nil {
					t.Fatal(err)
				}

				if copValue.Kind() == reflect.Ptr {
					t.Fatal("Expected a value, got a pointer")
				}

				expected := target_test.value
				if source_test.valueType == VT_FLOAT && target_test.valueType == VT_STRING {
					expected = fmt.Sprintf("%f", source_test.value)
				}

				if copValue.Interface() != expected {
					t.Fatalf("Values are different converting *%s to %s: got %v expected %v",
						source_test.name, target_test.name, copValue.Interface(), target_test.value)
				}
			}
		}
	}
}

func TestInterfaceTarget(t *testing.T) {
	src := test_int_value
	var dst interface{}

	sourceV := reflect.ValueOf(src)
	targetT := reflect.TypeOf(dst)

	cop := NewConfig().ConvertOpType(sourceV, targetT)
	if cop == nil {
		t.Fatal("Converter not found for int to interface{}")
	}

	copValue, err := cop(sourceV, targetT)
	if err != nil {
		t.Fatal(err)
	}

	if copValue.Kind() != reflect.Interface && copValue.Kind() != reflect.Int {
		t.Fatal("Expected an interface with int, got %s", copValue.Kind().String())
	}

	if copValue.Int() != int64(test_int_value) {
		t.Fatal("Expected value %d, got %d", int64(test_int_value), copValue.Int())
	}
}
