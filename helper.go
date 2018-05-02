package rprim

import (
	"fmt"
	"reflect"
)

// Helper to convert between a value and a type.
func Convert(src reflect.Value, dstType reflect.Type) (reflect.Value, error) {
	return NewConfig().Convert(src, dstType)
}

// Helper to convert a value to string.
func ConvertToString(src reflect.Value) (string, error) {
	return NewConfig().ConvertToString(src)
}

// Helper to convert between a value and a type.
func (c *Config) Convert(src reflect.Value, dstType reflect.Type) (reflect.Value, error) {
	cop := c.ConvertOpType(src, dstType)
	if cop == nil {
		return reflect.Value{}, fmt.Errorf("Invalid conversion from %s to %s", src.Type().String(), dstType.String())
	}
	cv, err := cop(src, dstType)
	if err != nil {
		return reflect.Value{}, err
	}
	return cv, nil
}

// Helper to convert a value to string.
func (c *Config) ConvertToString(src reflect.Value) (string, error) {
	t_string := reflect.TypeOf("")

	cop := c.ConvertOpType(src, t_string)
	if cop == nil {
		return "", fmt.Errorf("Invalid conversion from %s to %s", src.Type().String(), "string")
	}
	cv, err := cop(src, t_string)
	if err != nil {
		return "", err
	}
	return cv.String(), nil
}
