package dag

import (
	"testing"
)

type String string

func (s String) String() string {
	return string(s)
}

func TestG(t *testing.T) {
	g := New[String]()
	g.AddNode(String("A"))
	g.AddNode(String("B"))
	g.AddNode(String("C"))
	if err := g.AddEdge(String("A"), String("B")); err != nil {
		t.Errorf("expected no error, got %v", err)
		t.Fail()
	}
	if err := g.AddEdge(String("B"), String("C")); err != nil {
		t.Errorf("expected no error, got %v", err)
		t.Fail()
	}

	found := g.DetectAllCycles()
	if len(found) > 0 {
		t.Errorf("did not expect a cycle, got one")
		t.Fail()
	}

	// add an edge that creates a cycle
	if err := g.AddEdge(String("C"), String("A")); err != nil {
		t.Errorf("expected an error, got none")
		t.Fail()
	}

	cl := g.DetectAllCycles()
	if len(cl) == 0 {
		t.Errorf("expected a cycle, got none")
		t.Fail()
	}
}
