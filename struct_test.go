package mapper

import (
	"fmt"
	"testing"
)

type (
	Left struct {
		Name  string
		Skill int
	}

	Right struct {
		Name  string
		Skill int
	}
)

func TestConvert(t *testing.T) {
	left := &Left{
		Name:  "Pouya",
		Skill: 10,
	}

	val, err := Convert[Left, Right](left)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(val.Skill)
	_ = val
}
