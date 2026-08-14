package mapper

import (
	"fmt"
	"reflect"
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
	}
}

func FastConvert(in Value, o Type) (out Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	if in.Kind() != reflect.Pointer {
		in = in.Addr()
	}
	return NewAt(o, in.UnsafePointer()), nil
}

func TryAssign(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if !sourceField.AssignableTo(targetField) {
		return false, nil
	}
	targetValue.Set(Reference(targetField.PointerCount(), sourceValue))
	return true, nil
}

func TryConvert(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if !sourceField.ConvertibleTo(targetField) {
		return false, nil
	}
	targetValue.Set(Reference(targetField.PointerCount(), sourceValue.Convert(targetField)))
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
		_, realValue := DeReference(sourceValue.FieldByIndex(i.Index))
		val, err := Convert(realValue, target.Type.ConcreteType())
		if err != nil {
			return false, err
		}
		targetValue.FieldByIndex(target.Index).Set(Reference(target.Type.PointerCount(), val.Elem()))
	}

	return true, nil
}

func TryChangeArrayType(sourceField Type, targetField Type, sourceValue Value, targetValue Value) (bool, error) {
	if (sourceField.Kind() != reflect.Slice && sourceField.Kind() != reflect.Array) || (targetField.Kind() != reflect.Slice && targetField.Kind() != reflect.Array) {
		return false, nil
	}

	for i := range sourceValue.Len() {
		_, realValue := DeReference(sourceValue.Index(i))
		val, err := Convert(realValue, targetField.Elem().ConcreteType())
		if err != nil {
			return false, err
		}
		targetValue.Set(Reference(targetField.PointerCount(), valueOf(reflect.Append(targetValue.GoValue(), Reference(targetField.Elem().PointerCount(), val.Elem()).GoValue()))))
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
	targetValue.Set(Reference(targetField.PointerCount(), val))
	return true, nil
}

func SlowConvert(sourceType Type, targetType Type, source Value) (Value, error) {
	_, leftValue := DeReference(source)
	targetValue := New(targetType.ConcreteType()).Elem()

	for _, p := range pipeline {
		ok, err := p(sourceType.ConcreteType(), targetType.ConcreteType(), leftValue, targetValue)
		if err != nil {
			return Zero[Value](), err
		}
		if ok {
			break
		}
	}
	ref := Reference(targetType.PointerCount(), targetValue)

	return ref.Addr(), nil
}
func Convert(source Value, target Type) (Value, error) {

	leftType := source.Type()

	if leftType.MemoryLayout().IdenticalTo(target.MemoryLayout()) {
		return FastConvert(source, target)
	}
	return SlowConvert(leftType, target, source)
}
