package mapper

import (
	"fmt"
	"reflect"
)

func FloatToInt(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetInt(int64(src.Float()))
	return nil
}

func FloatToUint(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetUint(uint64(src.Float()))
	return nil
}

func FloatToComplex(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetComplex(complex(src.Float(), 0))
	return nil
}

func FloatToString(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetString(fmt.Sprintf("%f", src.Float()))
	return nil
}
