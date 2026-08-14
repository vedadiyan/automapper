package mapper

import (
	"fmt"
	"reflect"
	"sync"
)

type (
	Converter func(value Value, typ Type) (Value, error)
	Pipeline  func(sourceFiel, targetField StructField, sourceValue, targetValue Value, n int) (bool, error)
)

var (
	converters map[Type]map[Type]Converter
	pipeline   []Pipeline
	mut        sync.RWMutex
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

	return NewAt(o, in.GoValue().UnsafePointer()), nil
}

func TryAssign(sourceField StructField, targetField StructField, sourceValue Value, targetValue Value, n int) (bool, error) {
	if !sourceField.Type.AssignableTo(targetField.Type) {
		return false, nil
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, sourceValue))
	return true, nil
}

func TryConvert(sourceField StructField, targetField StructField, sourceValue Value, targetValue Value, n int) (bool, error) {
	if !sourceField.Type.ConvertibleTo(targetField.Type) {
		return false, nil
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, sourceValue.Convert(targetField.Type)))
	return true, nil
}

func TryChangeStructType(sourceField StructField, targetField StructField, sourceValue Value, targetValue Value, n int) (bool, error) {
	if sourceField.Type.Kind() != reflect.Struct || targetField.Type.Kind() != reflect.Struct {
		return false, nil
	}
	val, err := Convert(sourceValue, targetField.Type)
	if err != nil {
		return false, nil
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, val.Elem()))
	return true, nil
}

func TryChangeArrayType(sourceField StructField, targetField StructField, sourceValue Value, targetValue Value, n int) (bool, error) {
	if (sourceField.Type.Kind() != reflect.Slice && sourceField.Type.Kind() != reflect.Array) || (targetField.Type.Kind() != reflect.Slice && targetField.Type.Kind() != reflect.Array) {
		return false, nil
	}

	return true, nil
}

func TryCustomConvert(sourceField StructField, targetField StructField, sourceValue Value, targetValue Value, n int) (bool, error) {
	conv, ok := FindConverter(sourceField.Type, targetField.Type)
	if !ok {
		return false, nil
	}
	val, err := conv(sourceValue, targetValue.Type())
	if err != nil {
		return false, err
	}
	targetValue.FieldByIndex(targetField.Index).Set(Reference(n, val))
	return true, nil
}

func SlowConvert(sourceType Type, targetType Type, source Value) (Value, error) {
	_, leftValue := DeReference(source)
	targetValue := New(targetType.ConcreteType()).Elem()

	for sourceField := range sourceType.Fields() {
		if !sourceField.IsExported() {
			continue
		}

		sourceValue := leftValue.FieldByIndex(sourceField.Index)

		targetField, ok := targetType.FieldByName(sourceField.Name)
		if !ok {
			continue
		}

		for _, p := range pipeline {
			ok, err := p(sourceField, targetField, sourceValue, targetValue, sourceField.Type.PointerCount())
			if err != nil {
				return Zero[Value](), err
			}
			if ok {
				break
			}
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
