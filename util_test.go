package rprim

import (
	"reflect"
	"testing"
)

func TestUnderliningValueKind(t *testing.T) {
	var k reflect.Kind

	// string
	k = UnderliningValueKind(reflect.ValueOf(""))
	if k != reflect.String {
		t.Fatalf("Kind should be string, is %s", k.String())
	}

	// 3-level pointer to int
	var i ***int
	k = UnderliningValueKind(reflect.ValueOf(i))
	if k != reflect.Int {
		t.Fatalf("Kind should be int, is %s", k.String())
	}

	// 3-level pointer to nil interface
	var x ***interface{}
	k = UnderliningValueKind(reflect.ValueOf(x))
	if k != reflect.Interface {
		t.Fatalf("Kind should be interface, is %s", k.String())
	}

	// 3-level pointer to int interface
	xiv := 10
	xi := new(***interface{})
	*xi = new(**interface{})
	**xi = new(*interface{})
	***xi = new(interface{})
	****xi = xiv

	k = UnderliningValueKind(reflect.ValueOf(xi))
	if k != reflect.Int {
		t.Fatalf("Kind should be int, is %s", k.String())
	}

}

func TestUnderliningTypeKind(t *testing.T) {
	var k reflect.Kind

	// string
	k = UnderliningTypeKind(reflect.TypeOf(""))
	if k != reflect.String {
		t.Fatalf("Kind should be string, is %s", k.String())
	}

	// 3-level pointer to int
	var i ***int
	k = UnderliningTypeKind(reflect.TypeOf(i))
	if k != reflect.Int {
		t.Fatalf("Kind should be int, is %s", k.String())
	}

	// 3-level pointer to nil interface
	var x ***interface{}
	k = UnderliningTypeKind(reflect.TypeOf(x))
	if k != reflect.Interface {
		t.Fatalf("Kind should be interface, is %s", k.String())
	}
}

func TestNewUnderliningValue(t *testing.T) {
	var root, last reflect.Value

	// creates a new string
	root, last = NewUnderliningValue(reflect.TypeOf(""))
	if root.Kind() != reflect.String || last.Kind() != reflect.String {
		t.Fatalf("Both values should be string, but are %s and %s", root.Kind().String(), last.Kind().String())
	}

	// creates a new 3-level pointer to int
	var i ***int
	itype := reflect.TypeOf(i)
	root, last = NewUnderliningValue(itype)
	_, isi := root.Interface().(***int)
	if !isi {
		t.Fatal("Returned type is different from the source type")
	}
	if root.Kind() != reflect.Ptr || last.Kind() != reflect.Int {
		t.Fatalf("Values should be pointer and int, but are %s and %s", root.Kind().String(), last.Kind().String())
	}
}
