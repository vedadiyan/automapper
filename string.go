package mapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func StringToInt(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	val, err := strconv.ParseInt(src.String(), 10, 64)
	if err != nil {
		return err
	}
	target.SetInt(val)
	return nil
}

func StringToFloat(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	val, err := strconv.ParseFloat(src.String(), 64)
	if err != nil {
		return err
	}
	target.SetFloat(val)
	return nil
}

func StringToComplex(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	val, err := strconv.ParseComplex(src.String(), 128)
	if err != nil {
		return err
	}
	target.SetComplex(val)
	return nil
}

func StringToUint(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	val, err := strconv.ParseUint(src.String(), 10, 64)
	if err != nil {
		return err
	}
	target.SetUint(val)
	return nil
}

func StringToBool(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	switch strings.ToLower(src.String()) {
	case "false":
		{
			target.SetBool(false)
			return nil
		}
	case "true":
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

func stringToTime(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	time, err := time.Parse(time.RFC3339, src.String())
	if err != nil {
		return err
	}
	target.Set(reflect.ValueOf(time))
	return nil
}

func timeToString(src reflect.Value, target reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	t, ok := src.Interface().(time.Time)
	if !ok {
		return fmt.Errorf("type is not time")
	}
	target.SetString(t.Format(time.RFC3339))
	return nil
}
