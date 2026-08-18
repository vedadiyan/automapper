package automapper

import (
	"reflect"
)

type (
	RType struct {
		reflect.Type
		ct       reflect.Type
		ptrCount int
	}
)

func typeOf(t reflect.Type) RType {
	n, ct := DeReferenceType(t)
	out := RType{
		Type:     t,
		ct:       ct,
		ptrCount: n,
	}

	return out
}

func TypeOf(i any) RType {
	return typeOf(reflect.TypeOf(i))
}

func TypeFor[T any]() RType {
	return typeOf(reflect.TypeFor[T]())
}

func (rt RType) GoType() reflect.Type {
	return rt.Type
}

func (rt RType) ConcreteType() RType {
	return typeOf(rt.ct)
}

func (rt RType) PointerCount() int {
	return rt.ptrCount
}

func DeReferenceType(v reflect.Type) (int, reflect.Type) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}
