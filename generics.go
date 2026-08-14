package mapper

import (
	"fmt"
	"reflect"
	"unsafe"
)

func FastConvertFor[T any, R any](in *T) (r *R, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return (*R)(unsafe.Pointer(in)), nil
}

func ConvertFor[T any, R any](in *T) (*R, error) {
	out, err := Convert(ValueOf(in), typeOf(reflect.TypeFor[R]()))
	if err != nil {
		return nil, err
	}
	return out.Interface().(*R), nil
}
