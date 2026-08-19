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

func (rt RType) DetectCycleLoop() bool {
	visited := make(map[reflect.Type]bool)

	var detect func(RType) bool
	detect = func(current RType) bool {
		current = current.ConcreteType()

		if visited[current.Type] {
			return true
		}

		visited[current.Type] = true
		defer delete(visited, current.Type)

		switch current.Kind() {
		case reflect.Struct:
			for i := 0; i < current.NumField(); i++ {
				f := current.Field(i)
				if _, ignore := parseTag(f.Name, f.Tag.Get("mapper")); ignore {
					continue
				}
				if detect(typeOf(f.Type)) {
					return true
				}
			}
		case reflect.Array, reflect.Slice, reflect.Map:
			return detect(typeOf(current.Elem()))
		}

		return false
	}

	return detect(rt)
}

func DeReferenceType(v reflect.Type) (int, reflect.Type) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}

func parseTag(fieldName, value string) (string, bool) {
	if value == "-" {
		return "", true
	}
	if value == "" {
		return fieldName, false
	}
	return value, false
}
