package mapper

import (
	"fmt"
	"reflect"
)

func ComplexToInt(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetInt(int64(real(src.Complex())))
	return nil
}

func ComplexToUint(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetUint(uint64(real(src.Complex())))
	return nil
}

func ComplexToFloat(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetFloat(real(src.Complex()))
	return nil
}

func ComplexToString(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetString(fmt.Sprintf("%f", src.Complex()))
	return nil
}
