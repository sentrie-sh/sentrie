package dag

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
)

// G is a directed acyclic graph.
type G[T fmt.Stringer] interface {
	AddNode(T)
	AddEdge(T, T) error
	DetectAllCycles() [][]T
}

type gImpl[T fmt.Stringer] struct {
	lock  *sync.RWMutex
	nodes map[string]T
	edges map[string]map[string]struct{}
}

func New[T fmt.Stringer]() G[T] {
	return &gImpl[T]{
		lock:  &sync.RWMutex{},
		nodes: make(map[string]T),
		edges: make(map[string]map[string]struct{}),
	}
}

func (g *gImpl[T]) AddNode(node T) {
	g.nodes[node.String()] = node
	// add an empty edge map for this node
	g.edges[node.String()] = make(map[string]struct{})
}

var (
	ErrNodeMissing = errors.New("node not found")
	ErrSelfLoop    = errors.New("self-loop not allowed")
	ErrNotADAG     = errors.New("graph contains a cycle (not a DAG)")
)

type ErrCycle struct {
	Path []string
}

func (e ErrCycle) Error() string {
	return fmt.Sprintf("cycle detected: %v", strings.Join(e.Path, " -> "))
}

// AddEdge adds a directed edge from source to destination.
// This function does not check for cycles. It errors only if the source or destination node is missing or self-looping.
func (d *gImpl[T]) AddEdge(sourceID, destID T) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if sourceID.String() == destID.String() {
		return ErrSelfLoop
	}

	if _, ok := d.edges[sourceID.String()]; !ok {
		d.edges[sourceID.String()] = make(map[string]struct{})
	}

	// avoid duplicate edges
	if _, ok := d.edges[sourceID.String()][destID.String()]; !ok {
		d.edges[sourceID.String()][destID.String()] = struct{}{}
	}

	return nil
}

// Strategy: DFS
// Returns an error if the graph contains a cycle.
func (d *gImpl[T]) TopoSort() ([]T, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	visited := make(map[string]struct{})
	stack := make([]string, 0, len(d.nodes))
	visiting := make([]string, 0, len(d.nodes))

	var dfs func(node string) error
	dfs = func(node string) error {
		if slices.Contains(visiting, node) {
			// if we are already visiting this node, we have a cycle
			idx := slices.Index(visiting, node)
			path := append(visiting[idx:], node)
			return ErrCycle{Path: path}
		}
		if _, ok := visited[node]; ok {
			return nil
		}
		visiting = append(visiting, node) // push
		defer func() {
			visiting = visiting[:len(visiting)-1] // pop
			stack = append(stack, node)
		}()

		visited[node] = struct{}{}
		for neighbor := range d.edges[node] {
			if err := dfs(neighbor); err != nil {
				return err
			}
		}
		return nil
	}

	for node := range d.nodes {
		if err := dfs(node); err != nil {
			return nil, err
		}
	}

	slices.Reverse(stack)

	nodes := make([]T, 0, len(stack))
	for _, node := range stack {
		nodes = append(nodes, d.nodes[node])
	}
	return nodes, nil
}

func (d *gImpl[T]) DetectAllCycles() [][]T {
	d.lock.RLock()
	defer d.lock.RUnlock()

	detectedCycles := make([][]string, 0, len(d.edges)) // start pessimistically, we cannot have more cycles than edges

	visited := make(map[string]struct{})
	stack := make([]string, 0, len(d.nodes))
	visiting := make([]string, 0, len(d.nodes))

	var dfs func(node string) error
	dfs = func(node string) error {
		if slices.Contains(visiting, node) {
			// if we are already visiting this node, we have a cycle
			idx := slices.Index(visiting, node)
			path := append(visiting[idx:], node)
			return ErrCycle{Path: path}
		}
		if _, ok := visited[node]; ok {
			return nil
		}
		visiting = append(visiting, node) // push
		defer func() {
			visiting = visiting[:len(visiting)-1] // pop
			stack = append(stack, node)
		}()

		visited[node] = struct{}{}
		for neighbor := range d.edges[node] {
			if slices.Contains(visiting, neighbor) {
				idx := slices.Index(visiting, neighbor)
				path := append(visiting[idx:], neighbor)
				detectedCycles = append(detectedCycles, path)
				// ignore this edge and move on
				continue
			}

			if err := dfs(neighbor); err != nil {
				return err
			}
		}
		return nil
	}

	for node := range d.nodes {
		if err := dfs(node); err != nil {
			return nil
		}
	}

	slices.Reverse(stack)

	cycles := make([][]T, 0, len(detectedCycles))
	for _, cycle := range detectedCycles {
		thisCycle := make([]T, 0, len(cycle))
		for _, n := range cycle {
			thisCycle = append(thisCycle, d.nodes[n])
		}
		cycles = append(cycles, thisCycle)
	}
	return cycles
}
