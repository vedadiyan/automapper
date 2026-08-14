package mapper

import (
	"reflect"
	"unsafe"
)

type (
	MapIter = reflect.MapIter
	Value   interface {
		Addr() reflect.Value
		Bool() bool
		Bytes() []byte
		Call(in []Value) []Value
		CallSlice(in []Value) []Value
		CanAddr() bool
		CanComplex() bool
		CanConvert(t Type) bool
		CanFloat() bool
		CanInt() bool
		CanInterface() bool
		CanSet() bool
		CanUint() bool
		Cap() int
		Clear()
		Close()
		Comparable() bool
		Complex() complex128
		Convert(t Type) Value
		Elem() Value
		Equal(u Value) bool
		Field(i int) Value
		FieldByIndex(index []int) Value
		FieldByIndexErr(index []int) (Value, error)
		FieldByName(name string) Value
		FieldByNameFunc(match func(string) bool) Value
		Float() float64
		Index(i int) Value
		Int() int64
		IsNil() bool
		IsValid() bool
		IsZero() bool
		Kind() Kind
		Len() int
		MapIndex(key Value) Value
		MapKeys() []Value
		MapRange() *MapIter
		Method(i int) Value
		MethodByName(name string) Value
		NumField() int
		NumMethod() int
		OverflowComplex(x complex128) bool
		OverflowFloat(x float64) bool
		OverflowInt(x int64) bool
		OverflowUint(x uint64) bool
		Pointer() uintptr
		Recv() (x Value, ok bool)
		Send(x Value)
		Set(x Value)
		SetBool(x bool)
		SetBytes(x []byte)
		SetCap(n int)
		SetComplex(x complex128)
		SetFloat(x float64)
		SetInt(x int64)
		SetIterKey(iter *MapIter)
		SetIterValue(iter *MapIter)
		SetLen(n int)
		SetMapIndex(key, elem Value)
		SetPointer(p unsafe.Pointer)
		SetString(s string)
		SetUint(x uint64)
		SetZero()
		Slice(i, j int) Value
		Slice3(i, j, k int) Value
		String() string
		Type() Type
		UnsafePointer() unsafe.Pointer
		Uint() uint64
		Interface() any
		GoValue() reflect.Value
	}

	rvalue struct {
		v reflect.Value
	}
)

func valueOf(r reflect.Value) Value {
	return &rvalue{r}
}

func (rv rvalue) Addr() reflect.Value {
	return rv.v.Addr()
}
func (rv rvalue) Bool() bool {
	return rv.Bool()
}
func (rv rvalue) Bytes() []byte {
	return rv.Bytes()
}

func (rv rvalue) do(in []Value, fn func([]reflect.Value) []reflect.Value) []Value {
	inLen := len(in)
	ins := make([]reflect.Value, inLen)
	for i, v := range in {
		ins[i] = v.GoValue()
	}
	out := fn(ins)
	outLen := len(out)
	outs := make([]Value, outLen)
	for i, v := range out {
		outs[i] = valueOf(v)
	}
	return outs
}

func (rv rvalue) Call(in []Value) []Value {
	return rv.do(in, rv.v.Call)
}
func (rv rvalue) CallSlice(in []Value) []Value {
	return rv.do(in, rv.v.CallSlice)
}
func (rv rvalue) CanAddr() bool {
	return rv.v.CanAddr()
}
func (rv rvalue) CanComplex() bool {
	return rv.v.CanComplex()
}
func (rv rvalue) CanConvert(t Type) bool {
	return rv.v.CanConvert(t.GoType())
}
func (rv rvalue) CanFloat() bool {
	return rv.v.CanFloat()
}
func (rv rvalue) CanInt() bool {
	return rv.v.CanInt()
}
func (rv rvalue) CanInterface() bool {
	return rv.CanInterface()
}
func (rv rvalue) CanSet() bool {
	return rv.v.CanSet()
}
func (rv rvalue) CanUint() bool {
	return rv.v.CanUint()
}
func (rv rvalue) Cap() int {
	return rv.v.Cap()
}
func (rv rvalue) Clear() {
	rv.v.Clear()
}
func (rv rvalue) Close() {
	rv.v.Close()
}
func (rv rvalue) Comparable() bool {
	return rv.v.Comparable()
}
func (rv rvalue) Complex() complex128 {
	return rv.v.Complex()
}
func (rv rvalue) Convert(t Type) Value {
	return valueOf(rv.v.Convert(t.GoType()))
}
func (rv rvalue) Elem() Value {
	return valueOf(rv.v.Elem())
}
func (rv rvalue) Equal(u Value) bool {
	return rv.v.Equal(u.GoValue())
}
func (rv rvalue) Field(i int) Value {
	return valueOf(rv.v.Field(i))
}
func (rv rvalue) FieldByIndex(index []int) Value {
	return valueOf(rv.v.FieldByIndex(index))
}
func (rv rvalue) FieldByIndexErr(index []int) (Value, error) {
	out, err := rv.v.FieldByIndexErr(index)
	if err != nil {
		return nil, err
	}
	return valueOf(out), nil
}
func (rv rvalue) FieldByName(name string) Value {
	return valueOf(rv.v.FieldByName(name))
}
func (rv rvalue) FieldByNameFunc(match func(string) bool) Value {
	return valueOf(rv.v.FieldByNameFunc(match))
}
func (rv rvalue) Float() float64 {
	return rv.v.Float()
}
func (rv rvalue) Index(i int) Value {
	return valueOf(rv.v.Index(i))
}
func (rv rvalue) Int() int64 {
	return rv.v.Int()
}
func (rv rvalue) IsNil() bool {
	return rv.v.IsNil()
}
func (rv rvalue) IsValid() bool {
	return rv.v.IsValid()
}
func (rv rvalue) IsZero() bool {
	return rv.v.IsZero()
}
func (rv rvalue) Kind() Kind {
	return rv.v.Kind()
}
func (rv rvalue) Len() int {
	return rv.v.Len()
}
func (rv rvalue) MapIndex(key Value) Value {
	return valueOf(rv.v.MapIndex(key.GoValue()))
}
func (rv rvalue) MapKeys() []Value {
	out := rv.v.MapKeys()
	outLen := len(out)
	outs := make([]Value, outLen)
	for i, v := range out {
		outs[i] = valueOf(v)
	}
	return outs
}
func (rv rvalue) MapRange() *MapIter {
	return rv.v.MapRange()
}
func (rv rvalue) Method(i int) Value {
	return valueOf(rv.v.Method(i))
}
func (rv rvalue) MethodByName(name string) Value {
	return valueOf(rv.v.MethodByName(name))
}
func (rv rvalue) NumField() int {
	return rv.v.NumField()
}
func (rv rvalue) NumMethod() int {
	return rv.v.NumMethod()
}
func (rv rvalue) OverflowComplex(x complex128) bool {
	return rv.v.OverflowComplex(x)
}
func (rv rvalue) OverflowFloat(x float64) bool {
	return rv.v.OverflowFloat(x)
}
func (rv rvalue) OverflowInt(x int64) bool {
	return rv.v.OverflowInt(x)
}
func (rv rvalue) OverflowUint(x uint64) bool {
	return rv.v.OverflowUint(x)
}
func (rv rvalue) Pointer() uintptr {
	return rv.v.Pointer()
}
func (rv rvalue) Recv() (x Value, ok bool) {
	out, ok := rv.v.Recv()
	if !ok {
		return nil, false
	}
	return valueOf(out), true
}
func (rv rvalue) Send(x Value) {
	rv.v.Send(x.GoValue())
}
func (rv rvalue) Set(x Value) {
	rv.v.Set(x.GoValue())
}
func (rv rvalue) SetBool(x bool) {
	rv.v.SetBool(x)
}
func (rv rvalue) SetBytes(x []byte) {
	rv.v.SetBytes(x)
}
func (rv rvalue) SetCap(n int) {
	rv.v.SetCap(n)
}
func (rv rvalue) SetComplex(x complex128) {
	rv.v.SetComplex(x)
}
func (rv rvalue) SetFloat(x float64) {
	rv.v.SetFloat(x)
}
func (rv rvalue) SetInt(x int64) {
	rv.v.SetInt(x)
}
func (rv rvalue) SetIterKey(iter *MapIter) {
	rv.v.SetIterKey(iter)
}
func (rv rvalue) SetIterValue(iter *MapIter) {
	rv.v.SetIterValue(iter)
}
func (rv rvalue) SetLen(n int) {
	rv.v.SetLen(n)
}
func (rv rvalue) SetMapIndex(key, elem Value) {
	rv.v.SetMapIndex(key.GoValue(), elem.GoValue())
}
func (rv rvalue) SetPointer(p unsafe.Pointer) {
	rv.v.SetPointer(p)
}
func (rv rvalue) SetString(s string) {
	rv.v.SetString(s)
}
func (rv rvalue) SetUint(x uint64) {
	rv.v.SetUint(x)
}
func (rv rvalue) SetZero() {
	rv.v.SetZero()
}
func (rv rvalue) Slice(i, j int) Value {
	return valueOf(rv.v.Slice(i, j))
}
func (rv rvalue) Slice3(i, j, k int) Value {
	return valueOf(rv.v.Slice3(i, j, k))
}
func (rv rvalue) String() string {
	return rv.v.String()
}
func (rv rvalue) Type() Type {
	return typeOf(rv.v.Type())
}
func (rv rvalue) UnsafePointer() unsafe.Pointer {
	return rv.v.UnsafePointer()
}
func (rv rvalue) Uint() uint64 {
	return rv.v.Uint()
}
func (rv rvalue) Interface() any {
	return rv.v.Interface()
}

func (rv rvalue) GoValue() reflect.Value {
	return rv.v
}
