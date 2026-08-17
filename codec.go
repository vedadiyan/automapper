package mapper

import "reflect"

type (
	CodecFn func(sourceField Type, targetField Type) func(sourceValue RValue, targetValue RValue) error
)

var (
	codecs []CodecFn
)

func init() {
	codecs = []CodecFn{
		AssignCodec,
		ConvertCodec,
		StructCodec,
	}
}

func AssignCodec(sourceField Type, targetField Type) func(sourceValue RValue, targetValue RValue) error {
	if !sourceField.AssignableTo(targetField) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(sourceValue, targetField.PointerCount())
		return nil
	}
}

func ConvertCodec(sourceField Type, targetField Type) func(sourceValue RValue, targetValue RValue) error {
	if !sourceField.ConvertibleTo(targetField) {
		return nil
	}

	return func(sourceValue RValue, targetValue RValue) error {
		targetValue.SetAt(valueOf(sourceValue.Convert(targetField.GoType())), targetField.PointerCount())
		return nil
	}
}

func StructCodec(sourceField Type, targetField Type) func(sourceValue RValue, targetValue RValue) error {
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
		codec := Codec(f.Type, target.Type)

		if codec == nil {
			continue
		}
		out = append(out, func(sourceValue RValue, targetValue RValue) error {
			err := codec(valueOf(sourceValue.Field(i)), valueOf(targetValue.FieldByName(f.Name)))
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

func SlowConvertCodec(src Type, target Type) func(sourceValue RValue, targetValue RValue) error {
	sct := src.ConcreteType()
	tct := target.ConcreteType()

	for _, p := range codecs {
		if fn := p(sct, tct); fn != nil {
			return fn
		}

	}
	return nil
}
func Codec(src Type, target Type) func(sourceValue RValue, targetValue RValue) error {
	if src.MemoryLayout().IdenticalTo(target.MemoryLayout()) {
		return nil
	}
	return SlowConvertCodec(src, target)
}

func CodecFor[T any, R any]() func(in *T) (*R, error) {
	left := TypeFor[T]()
	right := TypeFor[R]()

	fn := Codec(left, right)
	if fn == nil {
		return nil
	}
	return func(in *T) (*R, error) {
		targetValue := valueOf(New(right).Elem())
		if err := fn(ValueOf(in).ConcreteValue(), targetValue); err != nil {
			return nil, err
		}
		out := targetValue.Interface().(R)
		return &out, nil
	}
}
