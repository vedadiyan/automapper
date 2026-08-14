package mapper

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type (
	Shared struct {
		Value int
		Test  []int
	}
	Shared2 struct {
		Value string
		Test  []int
	}
	Left struct {
		Name   string
		Skill  int
		Shared Shared
		Time   time.Time
	}

	Right struct {
		Name   string
		Skill  int
		Shared Shared2
		Time   **string
	}

	TI interface {
		Test() bool
	}

	T2 interface {
		A() bool
	}

	TII struct {
		xxx string
	}
	TIII struct {
		xxx string
	}
)

func (TII) Test() bool {
	return false
}

func (TIII) A() bool {
	return false
}

func TestConvert(t *testing.T) {

	var intA TI

	intA = &TII{}

	xxxx, err := FastConvertFor[TI, T2](&intA)

	zzzz := *xxxx

	yyyy := zzzz.A()
	_ = yyyy
	n := reflect.TypeOf(nil)

	_ = n
	left := &Left{
		Name:   "Pouya",
		Skill:  10,
		Shared: Shared{1, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		Time:   time.Now(),
	}

	strType := reflect.TypeFor[string]()
	converters[reflect.TypeFor[time.Time]()] = map[reflect.Type]Converter{
		strType: func(value reflect.Value, typ reflect.Type) (reflect.Value, error) {
			return reflect.ValueOf(fmt.Sprintf("%s", value.Interface())), nil
		},
	}
	converters[reflect.TypeFor[int]()] = map[reflect.Type]Converter{
		strType: func(value reflect.Value, typ reflect.Type) (reflect.Value, error) {
			return reflect.ValueOf(fmt.Sprintf("%d", value.Interface())), nil
		},
	}

	val, err := ConvertFor[Left, Right](left)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(val.Skill)
	_ = val
}
