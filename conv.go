package mapper

import (
	"fmt"
	"reflect"
	"unsafe"
)

type (
	Converter func(value Value, typ Type) (Value, error)
	Pipeline  func(sourceFiel, targetField Type, sourceValue, targetValue Value) (bool, error)
)

var (
	converters map[Type]map[Type]Converter
	pipeline   []Pipeline
)

func init() {
	converters = make(map[Type]map[Type]Converter)
	pipeline = []Pipeline{
		TryCustomConvert,
		TryAssign,
		TryConvert,
		TryChangeStructType,
		TryChangeArrayType,
		TryChangeMapType,
	}
}

func FastConvert(in Value, o Type) (out Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if in.Kind() != reflect.Pointer {
		in = in.Reference(1)
	}

	return NewAt(o, in.UnsafePointer()), nil
}

func TryAssign(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if !sourceField.AssignableTo(targetField) {
		return false, nil
	}

	targetValue.SetAt(sourceValue, targetField.PointerCount())

	return true, nil
}

func TryConvert(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if !sourceField.ConvertibleTo(targetField) {
		return false, nil
	}

	targetValue.SetAt(sourceValue.Convert(targetField), targetField.PointerCount())

	return true, nil
}

func TryChangeStructType(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if sourceField.Kind() != reflect.Struct || targetField.Kind() != reflect.Struct {
		return false, nil
	}

	for i := range sourceField.Fields() {
		target, ok := targetField.FieldByName(i.Name)
		if !ok {
			continue
		}
		val, err := Convert(sourceValue.FieldByIndex(i.Index).ConcreteValue(), target.Type.ConcreteType())
		if err != nil {
			return false, err
		}
		targetValue.FieldByIndex(target.Index).SetAt(val.Elem(), target.Type.PointerCount())
	}

	return true, nil
}

func TryChangeArrayType(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if (sourceField.Kind() != reflect.Slice && sourceField.Kind() != reflect.Array) || (targetField.Kind() != reflect.Slice && targetField.Kind() != reflect.Array) {
		return false, nil
	}

	if targetField.ConcreteType().Kind() == reflect.Slice && !sourceValue.ConcreteValue().IsZero() {
		targetValue.Set(valueOf(reflect.MakeSlice(targetField.GoType(), 0, 0)))
	}

	for i := range sourceValue.Len() {
		val, err := Convert(sourceValue.Index(i).ConcreteValue(), targetField.Elem().ConcreteType())
		if err != nil {
			return false, err
		}
		switch targetField.ConcreteType().Kind() {
		case reflect.Array:
			{
				targetValue.Index(i).Set(val.Elem().Reference(targetField.Elem().PointerCount()))
			}
		case reflect.Slice:
			{
				targetValue.SetAt(Append(targetValue, val.Elem().Reference(targetField.Elem().PointerCount())), targetField.PointerCount())
			}
		}

	}

	return true, nil
}

func TryChangeMapType(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if sourceField.Kind() != reflect.Map || targetField.Kind() != reflect.Map {
		return false, nil
	}

	mapRange := sourceValue.MapRange()

	init := false

	for mapRange.Next() {
		if !init {
			targetValue.Set(MakeMap(targetField))
			init = true
		}
		key, err := Convert(mapRange.Key().ConcreteValue(), targetField.Key().ConcreteType())
		if err != nil {
			return false, err
		}
		value, err := Convert(mapRange.Value().ConcreteValue(), targetField.Elem().ConcreteType())
		if err != nil {
			return false, err
		}
		targetValue.SetMapIndexAt(key.Elem(), targetField.Key().PointerCount(), value.Elem(), targetField.Elem().PointerCount())
	}

	return true, nil
}

func TryCustomConvert(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	conv, ok := FindConverter(sourceField, targetField)
	if !ok {
		return false, nil
	}
	val, err := conv(sourceValue, targetValue.Type())
	if err != nil {
		return false, err
	}
	targetValue.SetAt(val, targetField.PointerCount())

	return true, nil
}

func SlowConvert(sourceType Type, targetType Type, source Value) (Value, error) {
	targetValue := New(targetType.ConcreteType()).Elem()

	for _, p := range pipeline {
		ok, err := p(sourceType.ConcreteType(), targetType.ConcreteType(), source.ConcreteValue(), targetValue)
		if err != nil {
			return Zero[Value](), err
		}
		if ok {
			break
		}
	}

	ref := targetValue.Reference(targetType.PointerCount())

	return ref.Addr(), nil
}
func Convert(source Value, target Type) (Value, error) {

	leftType := source.Type()

	if leftType.MemoryLayout().IdenticalTo(target.MemoryLayout()) {
		return FastConvert(source, target)
	}

	return SlowConvert(leftType, target, source)
}

func FastConvertFor[T any, R any](in *T) (r *R, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return (*R)(unsafe.Pointer(in)), nil
}

func ConvertFor[T any, R any](in *T) (*R, error) {
	out, err := Convert(ValueOf(in), TypeFor[R]())
	if err != nil {
		return nil, err
	}
	return out.Interface().(*R), nil
}

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
