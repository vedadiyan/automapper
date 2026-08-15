package mapper

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

func TargetFieldName(field StructField) string {
	if val, ok := field.Tag.Lookup("mapto"); ok {
		return val
	}
	return field.Name
}

func Zero[T any]() T {
	var out T
	return out
}
