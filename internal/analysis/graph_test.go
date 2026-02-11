package analysis

import (
	"testing"
)

func TestGraphCycleDetection(t *testing.T) {
	// Case 1: No Cycle
	g1 := NewGraph()
	g1.AddEdge("A", "B")
	g1.AddEdge("B", "C")

	if hasCycle, path := g1.HasCycle(); hasCycle {
		t.Errorf("g1 should not have a cycle, but got one: %v", path)
	}

	// Case 2: Simple Cycle A -> B -> A
	g2 := NewGraph()
	g2.AddEdge("A", "B")
	g2.AddEdge("B", "A") // Cycle!

	if hasCycle, _ := g2.HasCycle(); !hasCycle {
		t.Error("g2 should have a cycle A->B->A, but none detected")
	}

	// Case 3: Indirect Cycle A -> B -> C -> A
	g3 := NewGraph()
	g3.AddEdge("A", "B")
	g3.AddEdge("B", "C")
	g3.AddEdge("C", "A") // Cycle!

	if hasCycle, _ := g3.HasCycle(); !hasCycle {
		t.Error("g3 should have a cycle A->B->C->A, but none detected")
	}

	// Case 4: Disconnected components with cycle
	g4 := NewGraph()
	g4.AddEdge("A", "B")
	g4.AddEdge("X", "Y")
	g4.AddEdge("Y", "X") // Cycle in component 2

	if hasCycle, _ := g4.HasCycle(); !hasCycle {
		t.Error("g4 should have a cycle in X-Y component")
	}
}
