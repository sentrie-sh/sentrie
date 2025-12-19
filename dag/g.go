// SPDX-License-Identifier: Apache-2.0

// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	TopoSort() ([]T, error)
	DetectFirstCycle() []T
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
		}()

		visited[node] = struct{}{}
		for neighbor := range d.edges[node] {
			if err := dfs(neighbor); err != nil {
				return err
			}
		}
		// Add node to stack after all its dependencies have been processed
		stack = append(stack, node)
		return nil
	}

	for node := range d.nodes {
		if err := dfs(node); err != nil {
			return nil, err
		}
	}

	// Reverse the stack to get the correct topological order
	slices.Reverse(stack)

	nodes := make([]T, 0, len(stack))
	for _, node := range stack {
		nodes = append(nodes, d.nodes[node])
	}
	return nodes, nil
}

func (d *gImpl[T]) DetectFirstCycle() []T {
	d.lock.RLock()
	defer d.lock.RUnlock()

	visited := make(map[string]struct{})
	stack := make([]string, 0, len(d.nodes))
	visiting := make([]string, 0, len(d.nodes))

	var dfs func(node string) []string
	dfs = func(node string) []string {
		if slices.Contains(visiting, node) {
			// if we are already visiting this node, we have a cycle
			idx := slices.Index(visiting, node)
			path := append(visiting[idx:], node)
			return path
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
		if cycle := dfs(node); len(cycle) > 0 {
			// Convert cycle path to []T
			result := make([]T, len(cycle))
			for i, nodeStr := range cycle {
				result[i] = d.nodes[nodeStr]
			}
			return result
		}
	}

	return []T{}
}
