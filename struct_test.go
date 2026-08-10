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
	}
	Shared2 struct {
		Value int
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
)

func TestConvert(t *testing.T) {
	left := &Left{
		Name:   "Pouya",
		Skill:  10,
		Shared: Shared{1},
		Time:   time.Now(),
	}

	strType := reflect.TypeFor[string]()
	converters[reflect.TypeFor[time.Time]()] = map[reflect.Type]Converter{
		strType: func(value reflect.Value, typ reflect.Type) (reflect.Value, error) {
			return reflect.ValueOf(fmt.Sprintf("%s", value.Interface())), nil
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
