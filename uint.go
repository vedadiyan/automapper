package mapper

import (
	"fmt"
	"reflect"
	"strconv"
)

func UintToInt(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetInt(int64(src.Uint()))
	return nil
}

func UintToFloat(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetFloat(float64(src.Uint()))
	return nil
}

func UintToComplex(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetComplex(complex(float64(src.Uint()), 0))
	return nil
}

func UintToString(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetString(strconv.FormatUint(src.Uint(), 10))
	return nil
}

func UintToBool(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	switch src.Uint() {
	case 0:
		{
			target.SetBool(false)
			return nil
		}
	case 1:
		{
			target.SetBool(true)
			return nil
		}
	default:
		{
			return fmt.Errorf("")
		}
	}
}
