package mapper

import (
	"reflect"
	"unsafe"
)

type (
	MapIter interface {
		Key() Value
		Value() Value
		Next() bool
		Reset(v Value)
		GoType() *reflect.MapIter
	}
	Value interface {
		ConcreteValue() Value
		PointerCount() int
		Reference(n int) Value
		SetAt(x Value, n int)
	}

	RValue struct {
		reflect.Value
		cv       reflect.Value
		ptrCount int
	}
)

func valueOf(r reflect.Value) RValue {
	n, cv := deReference(r)
	out := RValue{}
	out.Value = r
	out.ptrCount = n
	out.cv = cv
	return out
}

func (rv RValue) Reference(n int) RValue {
	return reference(n, rv)
}

func (rv RValue) SetAt(x RValue, n int) {
	rv.Set(reference(n, x).Value)
}

func (rv RValue) ConcreteValue() RValue {
	return valueOf(rv.cv)
}
func (rv RValue) Refresh() RValue {
	return valueOf(rv.Value)
}

func (rv RValue) PointerCount() int {
	return rv.ptrCount
}

func New(t RType) RValue {
	return valueOf(reflect.New(t.GoType()))
}

func ValueOf(i any) RValue {
	return valueOf(reflect.ValueOf(i))
}

func NewAt(t RType, p unsafe.Pointer) RValue {
	return valueOf(reflect.NewAt(t.GoType(), p))
}

func MakeMap(t RType) RValue {
	return valueOf(reflect.MakeMap(t.GoType()))
}

func Append(s RValue, x RValue) RValue {
	return valueOf(reflect.Append(s.Value, x.Value))
}
func AppendSlice(s RValue, x RValue) RValue {
	return valueOf(reflect.AppendSlice(s.Value, x.Value))
}

func reference(n int, v RValue) RValue {
	if n == 0 {
		return v
	}

	ref := New(typeOf(v.Type()))
	ref.Elem().Set(v.Value)

	for range n - 1 {
		next := New(typeOf(ref.Type()))
		next.Elem().Set(ref.Value)
		ref = next
	}

	return ref
}

func deReference(v reflect.Value) (int, reflect.Value) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}
