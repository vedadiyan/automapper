package mapper

import "reflect"

func FindConverter(l Type, r Type) (Converter, bool) {
	lVal, ok := converters[l]
	if !ok {
		return nil, false
	}
	rVal, ok := lVal[r]
	if !ok {
		return nil, false
	}
	return rVal, true
}

func DeReference(v Value) (int, Value) {
	i := 0
	for ; v.Kind() == reflect.Pointer; i++ {
		v = v.Elem()
	}
	return i, v
}

func Reference(n int, v Value) Value {
	if n == 0 {
		return v
	}

	ref := New(v.Type())
	ref.Elem().Set(v)

	for range n - 1 {
		next := New(ref.Type())
		next.Elem().Set(ref)
		ref = next
	}

	return ref
}

func TargetFieldName(field reflect.StructField) string {
	if val, ok := field.Tag.Lookup("mapto"); ok {
		return val
	}
	return field.Name
}

func Zero[T any]() T {
	var out T
	return out
}
