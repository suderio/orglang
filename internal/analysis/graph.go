package analysis

import (
	"fmt"
)

// Graph represents the flow of data between nodes (variables, expressions)
type Graph struct {
	nodes map[string]struct{}
	edges map[string][]string // Adjacency list: from -> [to]
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]struct{}),
		edges: make(map[string][]string),
	}
}

func (g *Graph) AddNode(name string) {
	if _, exists := g.nodes[name]; !exists {
		g.nodes[name] = struct{}{}
	}
}

func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.edges[from] = append(g.edges[from], to)
}

// HasCycle detects if there is a cycle in the graph using DFS
// Returns true if cycle found, and the path of the cycle
func (g *Graph) HasCycle() (bool, []string) {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for node := range g.nodes {
		if !visited[node] {
			if found, path := g.dfs(node, visited, recursionStack); found {
				return true, path
			}
		}
	}
	return false, nil
}

func (g *Graph) dfs(node string, visited, recursionStack map[string]bool) (bool, []string) {
	visited[node] = true
	recursionStack[node] = true

	if neighbors, ok := g.edges[node]; ok {
		for _, neighbor := range neighbors {
			if !visited[neighbor] {
				if found, path := g.dfs(neighbor, visited, recursionStack); found {
					return true, append([]string{node}, path...)
				}
			} else if recursionStack[neighbor] {
				return true, []string{node, neighbor}
			}
		}
	}

	recursionStack[node] = false
	return false, nil
}

func (g *Graph) String() string {
	out := "Graph:\n"
	for from, tos := range g.edges {
		out += fmt.Sprintf("  %s -> %v\n", from, tos)
	}
	return out
}
