# Golang primitive value copying

[![GoDoc](https://godoc.org/github.com/RangelReale/rprim?status.svg)](https://godoc.org/github.com/RangelReale/rprim)

This library contains functions to convert between any two Go primitive values (int, float, string, complex).

The values can have any number of pointer indirections or interface{} containment in either side of the conversion, 
even if in different number.

For string targets, it is possible to convert a printf/scanf format.

### Install

```
go get github.com/RangelReale/rprim
```

### Examples

String helpers:
```go
// Int
str, err := rprim.ConvertToString(reflect.ValueOf(910))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %s\n", str)

// Float
str, err = rprim.ConvertToString(reflect.ValueOf(910.15))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %s\n", str)

// Int pointer
var x *int
x = new(int)
*x = 765
str, err = rprim.ConvertToString(reflect.ValueOf(x))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %s\n", str)
```
Output:
```
Value: 910
Value: 910.150000
Value: 765
```

Int to string:
```go
i := 199
conv, err := rprim.Convert(reflect.ValueOf(i), reflect.TypeOf(""))
if err != nil {
    log.Fatal(err)
}

fmt.Printf("%s\n", conv.Interface().(string))
```
Output:
```
199
```

Int to float:
```go
i := 199
conv, err := rprim.Convert(reflect.ValueOf(i), reflect.TypeOf(float32(0)))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%f\n", conv.Interface().(float32))
```
Output:
```
199.000000
```

With pointer target:
```go
i := 199
var pi ****int
conv, err := rprim.Convert(reflect.ValueOf(i), reflect.TypeOf(pi))
if err != nil {
    log.Fatal(err)
}
pi = conv.Interface().(****int)
fmt.Printf("%d\n", ****pi)
```
Output:
```
199
```

Nil to zero conversion:
```go
// nil pointer to non-pointer target
var x *int
conv, err := rprim.Convert(reflect.ValueOf(x), reflect.TypeOf(0))
if err != nil {
    fmt.Printf("ERROR: %v\n", err)
} else {
    fmt.Printf("%d\n", conv.Int())
}

// allow nil-to-zero conversion
conv, err = rprim.NewConfig().AddFlags(rprim.COP_ALLOW_NIL_TO_ZERO_VALUE).
    Convert(reflect.ValueOf(x), reflect.TypeOf(0))
if err != nil {
    fmt.Printf("ERROR: %v\n", err)
} else {
    fmt.Printf("%d\n", conv.Int())
}
```
Output:
```
ERROR: Copying nil to zero value not allowed
0
```

### Author

Rangel Reale (rangelspam@gmail.com) 
