package mapper

import (
	"reflect"
)

type (
	CodecFn func(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error
)

var (
	codecs []CodecFn
)

func init() {
	codecs = []CodecFn{
		AssignCodec,
		ConvertCodec,
		StructCodec,
		ArrayCodec,
		MapCodec,
	}
}

func AssignCodec(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error {
	if !sourceField.AssignableTo(targetField) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(sourceValue, targetField.PointerCount())
		return nil
	}
}

func ConvertCodec(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error {
	if !sourceField.ConvertibleTo(targetField) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(valueOf(sourceValue.Convert(targetField.GoType())), targetField.PointerCount())
		return nil
	}
}

func StructCodec(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error {
	if sourceField.Kind() != reflect.Struct || targetField.Kind() != reflect.Struct {
		return nil
	}

	out := make([]func(sourceValue RValue, targetValue RValue) error, 0)

	for i := range sourceField.NumField() {
		f := sourceField.Field(i)
		target, ok := targetField.FieldByName(f.Name)
		if !ok {
			continue
		}
		codec := Codec(typeOf(f.Type).ConcreteType(), typeOf(target.Type).ConcreteType())

		if codec == nil {
			continue
		}
		sourceIndex := i
		targetIndex := target.Index
		out = append(out, func(sourceValue RValue, targetValue RValue) error {
			err := codec(valueOf(sourceValue.Field(sourceIndex)), valueOf(targetValue.FieldByIndex(targetIndex)))
			if err != nil {
				return err
			}
			return nil
		})

	}

	return func(sourceValue, targetValue RValue) error {
		for _, o := range out {
			if err := o(sourceValue, targetValue); err != nil {
				return err
			}
		}
		return nil
	}
}

func ArrayCodec(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error {
	if (sourceField.Kind() != reflect.Slice && sourceField.Kind() != reflect.Array) || (targetField.Kind() != reflect.Slice && targetField.Kind() != reflect.Array) {
		return nil
	}

	targetElem := typeOf(targetField.Elem())
	targetType := targetElem.ConcreteType()
	codec := Codec(typeOf(sourceField.Elem()).ConcreteType(), targetType)
	if codec == nil {
		return nil
	}
	n := targetElem.PointerCount()

	return func(sourceValue, targetValue RValue) error {
		if targetField.Kind() == reflect.Slice {
			if !sourceValue.IsZero() {
				targetValue.SetAt(valueOf(reflect.MakeSlice(targetField.GoType(), sourceValue.Len(), sourceValue.Len())), targetValue.PointerCount())
			}
		}
		target := valueOf(reflect.New(targetType.GoType()).Elem())
		targetValue = targetValue.Refresh().ConcreteValue()
		for i := range sourceValue.Len() {
			src := valueOf(sourceValue.Index(i))
			err := codec(src, target)
			if err != nil {
				return err
			}
			targetValue.Index(i).Set(target.Reference(n).Value)
		}
		return nil
	}
}

func MapCodec(sourceField RType, targetField RType) func(sourceValue RValue, targetValue RValue) error {
	if sourceField.Kind() != reflect.Map || targetField.Kind() != reflect.Map {
		return nil
	}

	targetKeyRawType := typeOf(targetField.Key())
	targetValueRawType := typeOf(targetField.Elem())

	sourceKeyRawType := typeOf(sourceField.Key())
	sourceValueRawType := typeOf(sourceField.Elem())

	keyType := targetKeyRawType.ConcreteType()
	valueType := targetValueRawType.ConcreteType()

	keyN := targetKeyRawType.PointerCount()
	valueN := targetValueRawType.PointerCount()

	keyCodec := Codec(sourceKeyRawType.ConcreteType(), keyType.ConcreteType())
	if keyCodec == nil {
		return nil
	}
	valueCodec := Codec(sourceValueRawType.ConcreteType(), valueType.ConcreteType())
	if valueCodec == nil {
		return nil
	}

	return func(sourceValue, targetValue RValue) error {
		mapRange := sourceValue.MapRange()
		targetValue.Set(reflect.MakeMapWithSize(targetField.GoType(), sourceValue.Len()))
		key := valueOf(New(keyType).Elem())
		value := valueOf(New(valueType).Elem())
		for mapRange.Next() {
			if err := keyCodec(valueOf(mapRange.Key()).ConcreteValue(), key); err != nil {
				return err
			}
			if err := valueCodec(valueOf(mapRange.Value()).ConcreteValue(), value); err != nil {

				return err
			}
			targetValue.SetMapIndex(key.Reference(keyN).Value, value.Reference(valueN).Value)
		}
		return nil
	}
}

func Codec(src RType, target RType) func(sourceValue RValue, targetValue RValue) error {
	for _, p := range codecs {
		if fn := p(src, target); fn != nil {
			return fn
		}
	}
	return nil
}

func CodecFor[T any, R any]() func(in *T) (*R, error) {
	src := TypeFor[T]()
	target := TypeFor[R]()

	fn := Codec(src.ConcreteType(), target.ConcreteType())
	if fn == nil {
		return nil
	}
	return func(in *T) (*R, error) {
		targetValue := New(target.ConcreteType())
		if err := fn(ValueOf(in).ConcreteValue(), targetValue.ConcreteValue()); err != nil {
			return nil, err
		}
		return targetValue.Reference(target.PointerCount()).Interface().(*R), nil
	}
}
