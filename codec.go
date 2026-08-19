package automapper

import (
	"fmt"
	"reflect"
	"sync/atomic"
)

type (
	Codec        func(sourceValue RValue, targetValue RValue) error
	CodecFactory func(sourceType RType, targetType RType) Codec
)

var (
	codecs       []CodecFactory
	customCodecs atomic.Pointer[[]CodecFactory]
)

func init() {
	codecs = []CodecFactory{
		CustomCodec,
		AssignCodec,
		ConvertCodec,
		StructCodec,
		ArrayCodec,
		MapCodec,
	}
	customCodecs.Store(&[]CodecFactory{})
}

func CustomCodec(sourceType RType, targetType RType) Codec {
	codecs := customCodecs.Load()
	if codecs == nil {
		return nil
	}
	for _, codec := range *codecs {
		if fn := codec(sourceType, targetType); fn != nil {
			return fn
		}
	}
	return nil
}

func AssignCodec(sourceType RType, targetType RType) Codec {
	if !sourceType.AssignableTo(targetType) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(sourceValue, targetType.PointerCount())
		return nil
	}
}

func ConvertCodec(sourceType RType, targetType RType) Codec {
	if !sourceType.ConvertibleTo(targetType) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(valueOf(sourceValue.Convert(targetType.GoType())), targetType.PointerCount())
		return nil
	}
}

func StructCodec(sourceType RType, targetType RType) Codec {
	if sourceType.Kind() != reflect.Struct || targetType.Kind() != reflect.Struct {
		return nil
	}

	out := make([]Codec, 0)

	for i := range sourceType.NumField() {
		f := sourceType.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, ignore := parseTag(f.Name, f.Tag.Get("mapper"))
		if ignore {
			continue
		}
		target, ok := targetType.FieldByName(tag)
		if !ok {
			continue
		}
		if !target.IsExported() {
			continue
		}
		codec := CreateCodec(typeOf(f.Type).ConcreteType(), typeOf(target.Type).ConcreteType())

		if codec == nil {
			continue
		}
		sourceIndex := i
		targetIndex := target.Index
		out = append(out, func(sourceValue RValue, targetValue RValue) error {
			return codec(valueOf(sourceValue.Field(sourceIndex)), valueOf(targetValue.FieldByIndex(targetIndex)))
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

func ArrayCodec(sourceType RType, targetType RType) Codec {
	if (sourceType.Kind() != reflect.Slice && sourceType.Kind() != reflect.Array) || (targetType.Kind() != reflect.Slice && targetType.Kind() != reflect.Array) {
		return nil
	}

	targetElem := typeOf(targetType.Elem())
	targetElemType := targetElem.ConcreteType()
	codec := CreateCodec(typeOf(sourceType.Elem()).ConcreteType(), targetElemType)
	if codec == nil {
		return nil
	}
	n := targetElem.PointerCount()

	return func(sourceValue, targetValue RValue) error {
		if targetType.Kind() == reflect.Slice {
			if !sourceValue.IsZero() {
				targetValue.SetAt(valueOf(reflect.MakeSlice(targetType.GoType(), sourceValue.Len(), sourceValue.Len())), targetValue.PointerCount())
			}
		}
		target := valueOf(reflect.New(targetElemType.GoType()).Elem())
		targetValue = targetValue.Refresh().ConcreteValue()
		for i := range min(sourceValue.Len(), targetValue.Len()) {
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

func MapCodec(sourceType RType, targetType RType) Codec {
	if sourceType.Kind() != reflect.Map || targetType.Kind() != reflect.Map {
		return nil
	}

	targetKeyRawType := typeOf(targetType.Key())
	targetValueRawType := typeOf(targetType.Elem())

	sourceKeyRawType := typeOf(sourceType.Key())
	sourceValueRawType := typeOf(sourceType.Elem())

	keyType := targetKeyRawType.ConcreteType()
	valueType := targetValueRawType.ConcreteType()

	keyN := targetKeyRawType.PointerCount()
	valueN := targetValueRawType.PointerCount()

	keyCodec := CreateCodec(sourceKeyRawType.ConcreteType(), keyType.ConcreteType())
	if keyCodec == nil {
		return nil
	}
	valueCodec := CreateCodec(sourceValueRawType.ConcreteType(), valueType.ConcreteType())
	if valueCodec == nil {
		return nil
	}

	return func(sourceValue, targetValue RValue) error {
		mapRange := sourceValue.MapRange()
		targetValue.Set(reflect.MakeMapWithSize(targetType.GoType(), sourceValue.Len()))
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

func CreateCodec(src RType, target RType) Codec {
	for _, p := range codecs {
		if fn := p(src, target); fn != nil {
			return fn
		}
	}
	return nil
}

func CreateCodecFor[T any, R any]() func(in *T) (*R, error) {
	src := TypeFor[T]()
	target := TypeFor[R]()

	if target.DetectCycleLoop() {
		return func(in *T) (*R, error) {
			return nil, fmt.Errorf("cycle loop detected")
		}
	}

	fn := CreateCodec(src.ConcreteType(), target.ConcreteType())
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

func CreateCustomCodec[T any, R any](codec func(RValue) (RValue, error)) CodecFactory {
	src := TypeFor[T]()
	tgt := TypeFor[R]()
	if tgt.DetectCycleLoop() {
		return func(sourceType, targetType RType) Codec {
			return func(sourceValue, targetValue RValue) error {
				return fmt.Errorf("cycle loop detected")
			}
		}
	}

	return func(sourceType RType, targetType RType) Codec {
		if sourceType.ConcreteType() != src.ConcreteType() || targetType.ConcreteType() != tgt.ConcreteType() {
			return nil
		}
		return func(sourceValue RValue, targetValue RValue) error {
			r, err := codec(sourceValue)
			if err != nil {
				return err
			}
			targetValue.SetAt(r.ConcreteValue(), targetType.PointerCount())
			return nil
		}
	}
}

func SetCustomCodecs(codecs []CodecFactory) {
	copied := append([]CodecFactory(nil), codecs...)
	customCodecs.Store(&copied)
}
