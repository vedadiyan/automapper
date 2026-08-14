package mapper

import (
	"fmt"
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
		Test  []*int32
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

	left := &Left{
		Name:   "Pouya",
		Skill:  10,
		Shared: Shared{1, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		Time:   time.Now(),
	}

	strType := TypeFor[string]()
	converters[TypeFor[time.Time]()] = map[Type]Converter{
		strType: func(value Value, typ Type) (Value, error) {
			return ValueOf(fmt.Sprintf("%s", value.Interface())), nil
		},
	}
	converters[TypeFor[int]()] = map[Type]Converter{
		strType: func(value Value, typ Type) (Value, error) {
			return ValueOf(fmt.Sprintf("%d", value.Interface())), nil
		},
	}

	TypeFor[Left]().MemoryLayout()
	TypeFor[Right]().MemoryLayout()

	val, err := ConvertFor[Left, Right](left)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(val.Skill)
	_ = val
}
