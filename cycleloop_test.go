package automapper

import (
	"fmt"
	"testing"
)

type Node struct {
	Value int
	Next  struct {
		Next *Node
	}
}

type NodeDTO struct {
	Value int
	Next  *NodeDTO
}

func TestCreateCodecFor_SelfReferentialStruct(t *testing.T) {
	codec := CreateCodecFor[Node, NodeDTO]()
	if codec == nil {
		t.Fatal("expected a non-nil codec for a self-referential but acyclic type")
	}

	src := &Node{Value: 1, Next: struct{ Next *Node }{Next: &Node{}}}
	_, err := codec(src)
	if err == nil {
		t.Fatal(fmt.Errorf("expected error"))
	}
}
