package automapper

import "testing"

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
	got, err := codec(src)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != 1 || got.Next.Value != 2 || got.Next.Next.Value != 3 || got.Next.Next.Next != nil {
		t.Fatalf("got %+v", got)
	}
}
