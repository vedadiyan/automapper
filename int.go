package mapper

import (
	"fmt"
	"reflect"
	"strconv"
)

func IntToUint(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetUint(uint64(src.Int()))
	return nil
}

func IntToFloat(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetFloat(float64(src.Int()))
	return nil
}

func IntToComplex(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetComplex(complex(float64(src.Int()), 0))
	return nil
}

func IntToString(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	target.SetString(strconv.FormatInt(src.Int(), 10))
	return nil
}

func IntToBool(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	switch src.Int() {
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
