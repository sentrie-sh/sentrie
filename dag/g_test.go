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
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestNode implements fmt.Stringer for testing
type TestNode struct {
	ID string
}

func (n TestNode) String() string {
	return n.ID
}

// TestNodeWithData implements fmt.Stringer for testing with additional data
type TestNodeWithData struct {
	ID   string
	Data int
}

func (n TestNodeWithData) String() string {
	return n.ID
}

// StringNode implements fmt.Stringer for testing
type StringNode string

func (s StringNode) String() string {
	return string(s)
}

type DagTestSuite struct {
	suite.Suite
}

func (suite *DagTestSuite) SetupSuite() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(suite.T().Output(), nil)))
}

func (suite *DagTestSuite) BeforeTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "BeforeTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "BeforeTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *DagTestSuite) AfterTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "AfterTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "AfterTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *DagTestSuite) TearDownSuite() {
	slog.InfoContext(suite.T().Context(), "TearDownSuite")
	defer slog.InfoContext(suite.T().Context(), "TearDownSuite end")
}

// TestNew tests the New() function
func (s *DagTestSuite) TestNew() {
	// Test with TestNode type
	graph1 := New[TestNode]()
	s.NotNil(graph1)
	s.Implements((*G[TestNode])(nil), graph1)

	// Test with TestNodeWithData type
	graph2 := New[TestNodeWithData]()
	s.NotNil(graph2)
	s.Implements((*G[TestNodeWithData])(nil), graph2)

	// Test with custom string type that implements fmt.Stringer
	graph3 := New[StringNode]()
	s.NotNil(graph3)
	s.Implements((*G[StringNode])(nil), graph3)
}

// TestAddNode tests the AddNode() method
func (s *DagTestSuite) TestAddNode() {
	graph := New[TestNode]()

	// Test adding single node
	node1 := TestNode{ID: "node1"}
	graph.AddNode(node1)

	// Test adding multiple nodes
	node2 := TestNode{ID: "node2"}
	node3 := TestNode{ID: "node3"}
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Test adding duplicate node (should not error, but should overwrite)
	duplicateNode := TestNode{ID: "node1"}
	graph.AddNode(duplicateNode)

	// Test adding node with empty string ID
	emptyNode := TestNode{ID: ""}
	graph.AddNode(emptyNode)

	// Test adding node with special characters
	specialNode := TestNode{ID: "node-with-special-chars!@#$%"}
	graph.AddNode(specialNode)
}

// TestAddNodeWithData tests AddNode with nodes containing additional data
func (s *DagTestSuite) TestAddNodeWithData() {
	graph := New[TestNodeWithData]()

	// Test adding nodes with different data
	node1 := TestNodeWithData{ID: "node1", Data: 100}
	node2 := TestNodeWithData{ID: "node2", Data: 200}
	node3 := TestNodeWithData{ID: "node3", Data: 300}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Test that nodes with same ID but different data overwrite
	node1Updated := TestNodeWithData{ID: "node1", Data: 999}
	graph.AddNode(node1Updated)
}

// TestAddEdge tests the AddEdge() method
func (s *DagTestSuite) TestAddEdge() {
	graph := New[TestNode]()

	// Add nodes first
	node1 := TestNode{ID: "node1"}
	node2 := TestNode{ID: "node2"}
	node3 := TestNode{ID: "node3"}
	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	// Test adding valid edges
	err := graph.AddEdge(node1, node2)
	s.NoError(err)

	err = graph.AddEdge(node2, node3)
	s.NoError(err)

	err = graph.AddEdge(node1, node3)
	s.NoError(err)

	// Test adding duplicate edge (should not error)
	err = graph.AddEdge(node1, node2)
	s.NoError(err)

	// Test adding edge with non-existent source node
	nonExistentNode := TestNode{ID: "non-existent"}
	err = graph.AddEdge(nonExistentNode, node1)
	s.NoError(err) // AddEdge doesn't check for node existence

	// Test adding edge with non-existent destination node
	err = graph.AddEdge(node1, nonExistentNode)
	s.NoError(err) // AddEdge doesn't check for node existence
}

// TestAddEdgeErrors tests error conditions for AddEdge
func (s *DagTestSuite) TestAddEdgeErrors() {
	graph := New[TestNode]()

	node1 := TestNode{ID: "node1"}
	graph.AddNode(node1)

	// Test self-loop error
	err := graph.AddEdge(node1, node1)
	s.Error(err)
	s.Equal(ErrSelfLoop, err)

	// Test self-loop with different node objects but same ID
	node1Copy := TestNode{ID: "node1"}
	err = graph.AddEdge(node1, node1Copy)
	s.Error(err)
	s.Equal(ErrSelfLoop, err)
}

// TestTopoSort tests the TopoSort() method
func (s *DagTestSuite) TestTopoSort() {
	graph := New[TestNode]()

	// Test empty graph
	nodes, err := s.topoSort(graph)
	s.NoError(err)
	s.Empty(nodes)

	// Test single node
	node1 := TestNode{ID: "node1"}
	graph.AddNode(node1)
	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 1)
	s.Equal(node1, nodes[0])

	// Test linear graph: A -> B -> C
	nodeA := TestNode{ID: "A"}
	nodeB := TestNode{ID: "B"}
	nodeC := TestNode{ID: "C"}
	graph = New[TestNode]()
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeC)

	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 3)
	// Topological order should be A, B, C (or any valid order)
	s.Contains(nodes, nodeA)
	s.Contains(nodes, nodeB)
	s.Contains(nodes, nodeC)

	// Test diamond graph: A -> B -> D, A -> C -> D
	graph = New[TestNode]()
	nodeA = TestNode{ID: "A"}
	nodeB = TestNode{ID: "B"}
	nodeC = TestNode{ID: "C"}
	nodeD := TestNode{ID: "D"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeA, nodeC)
	graph.AddEdge(nodeB, nodeD)
	graph.AddEdge(nodeC, nodeD)

	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 4)
	s.Contains(nodes, nodeA)
	s.Contains(nodes, nodeB)
	s.Contains(nodes, nodeC)
	s.Contains(nodes, nodeD)

	// Verify A comes before B and C, and B and C come before D
	aIndex := s.findNodeIndex(nodes, nodeA)
	bIndex := s.findNodeIndex(nodes, nodeB)
	cIndex := s.findNodeIndex(nodes, nodeC)
	dIndex := s.findNodeIndex(nodes, nodeD)

	s.True(aIndex < bIndex, "A should come before B")
	s.True(aIndex < cIndex, "A should come before C")
	s.True(bIndex < dIndex, "B should come before D")
	s.True(cIndex < dIndex, "C should come before D")
}

// TestTopoSortWithCycle tests TopoSort with cyclic graphs
func (s *DagTestSuite) TestTopoSortWithCycle() {
	// Test simple cycle: A -> B -> A
	graph := New[TestNode]()
	nodeA := TestNode{ID: "A"}
	nodeB := TestNode{ID: "B"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeA)

	nodes, err := s.topoSort(graph)
	s.Error(err)
	s.Nil(nodes)

	// Check if it's a cycle error
	cycleErr, ok := err.(ErrCycle)
	s.True(ok, "Error should be of type ErrCycle")
	s.NotEmpty(cycleErr.Path)
	s.Contains(cycleErr.Path, "A")
	s.Contains(cycleErr.Path, "B")

	// Test complex cycle: A -> B -> C -> A
	graph = New[TestNode]()
	nodeA = TestNode{ID: "A"}
	nodeB = TestNode{ID: "B"}
	nodeC := TestNode{ID: "C"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeC)
	graph.AddEdge(nodeC, nodeA)

	nodes, err = s.topoSort(graph)
	s.Error(err)
	s.Nil(nodes)

	cycleErr, ok = err.(ErrCycle)
	s.True(ok, "Error should be of type ErrCycle")
	s.NotEmpty(cycleErr.Path)
	s.Contains(cycleErr.Path, "A")
	s.Contains(cycleErr.Path, "B")
	s.Contains(cycleErr.Path, "C")
}

// TestDetectAllCycles tests the DetectAllCycles() method
func (s *DagTestSuite) TestDetectAllCycles() {
	// Test empty graph
	graph := New[TestNode]()
	cycles := graph.DetectFirstCycle()
	s.Empty(cycles)

	// Test acyclic graph
	graph = New[TestNode]()
	nodeA := TestNode{ID: "A"}
	nodeB := TestNode{ID: "B"}
	nodeC := TestNode{ID: "C"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeC)

	cycles = graph.DetectFirstCycle()
	s.Empty(cycles)

	// Test single cycle: A -> B -> A
	graph = New[TestNode]()
	nodeA = TestNode{ID: "A"}
	nodeB = TestNode{ID: "B"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeA)

	cycles = graph.DetectFirstCycle()
	s.Len(cycles, 3)
	s.Contains(cycles, nodeA)
	s.Contains(cycles, nodeB)

	// Test multiple cycles in a connected graph
	graph = New[TestNode]()
	nodeA = TestNode{ID: "A"}
	nodeB = TestNode{ID: "B"}
	nodeC = TestNode{ID: "C"}
	nodeD := TestNode{ID: "D"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)
	graph.AddEdge(nodeA, nodeB)
	graph.AddEdge(nodeB, nodeA) // Cycle 1: A -> B -> A
	graph.AddEdge(nodeB, nodeC) // Connect the components
	graph.AddEdge(nodeC, nodeD)
	graph.AddEdge(nodeD, nodeC) // Cycle 2: C -> D -> C

	cycles = graph.DetectFirstCycle()
	s.NotEmpty(cycles, "Should detect at least one cycle")

	s.True(len(cycles) > 0, "At least one cycle should be detected")
}

// TestComplexGraphStructures tests complex graph structures
func (s *DagTestSuite) TestComplexGraphStructures() {
	// Test disconnected components
	graph := New[TestNode]()
	nodeA := TestNode{ID: "A"}
	nodeB := TestNode{ID: "B"}
	nodeC := TestNode{ID: "C"}
	nodeD := TestNode{ID: "D"}
	graph.AddNode(nodeA)
	graph.AddNode(nodeB)
	graph.AddNode(nodeC)
	graph.AddNode(nodeD)
	graph.AddEdge(nodeA, nodeB) // Component 1: A -> B
	graph.AddEdge(nodeC, nodeD) // Component 2: C -> D

	nodes, err := s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 4)
	s.Contains(nodes, nodeA)
	s.Contains(nodes, nodeB)
	s.Contains(nodes, nodeC)
	s.Contains(nodes, nodeD)

	// Test tree structure
	graph = New[TestNode]()
	root := TestNode{ID: "root"}
	left := TestNode{ID: "left"}
	right := TestNode{ID: "right"}
	leftLeft := TestNode{ID: "leftLeft"}
	leftRight := TestNode{ID: "leftRight"}
	rightLeft := TestNode{ID: "rightLeft"}
	rightRight := TestNode{ID: "rightRight"}

	graph.AddNode(root)
	graph.AddNode(left)
	graph.AddNode(right)
	graph.AddNode(leftLeft)
	graph.AddNode(leftRight)
	graph.AddNode(rightLeft)
	graph.AddNode(rightRight)

	graph.AddEdge(root, left)
	graph.AddEdge(root, right)
	graph.AddEdge(left, leftLeft)
	graph.AddEdge(left, leftRight)
	graph.AddEdge(right, rightLeft)
	graph.AddEdge(right, rightRight)

	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 7)

	// Verify root comes before all children
	rootIndex := s.findNodeIndex(nodes, root)
	leftIndex := s.findNodeIndex(nodes, left)
	rightIndex := s.findNodeIndex(nodes, right)
	s.True(rootIndex < leftIndex, "Root should come before left child")
	s.True(rootIndex < rightIndex, "Root should come before right child")
}

// TestConcurrency tests thread safety
func (s *DagTestSuite) TestConcurrency() {
	graph := New[TestNode]()

	// First, add nodes sequentially to avoid race conditions
	// (AddNode is not thread-safe, but AddEdge and read operations are)
	for i := 0; i < 10; i++ {
		node := TestNode{ID: fmt.Sprintf("node%d", i)}
		graph.AddNode(node)
	}

	// Test concurrent AddEdge operations (these are thread-safe)
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			source := TestNode{ID: fmt.Sprintf("node%d", i)}
			dest := TestNode{ID: fmt.Sprintf("node%d", i+1)}
			graph.AddEdge(source, dest)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Test concurrent read operations (these are thread-safe)
	for i := 0; i < 5; i++ {
		go func() {
			s.topoSort(graph)
			graph.DetectFirstCycle()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestEdgeCases tests various edge cases
func (s *DagTestSuite) TestEdgeCases() {
	// Test graph with single node and no edges
	graph := New[TestNode]()
	node := TestNode{ID: "single"}
	graph.AddNode(node)

	nodes, err := s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 1)
	s.Equal(node, nodes[0])

	cycles := graph.DetectFirstCycle()
	s.Empty(cycles)

	// Test graph with nodes but no edges
	graph = New[TestNode]()
	node1 := TestNode{ID: "node1"}
	node2 := TestNode{ID: "node2"}
	node3 := TestNode{ID: "node3"}
	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)

	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 3)
	s.Contains(nodes, node1)
	s.Contains(nodes, node2)
	s.Contains(nodes, node3)

	// Test graph with special characters in node IDs
	graph = New[TestNode]()
	specialNode1 := TestNode{ID: "node-with-dashes"}
	specialNode2 := TestNode{ID: "node_with_underscores"}
	specialNode3 := TestNode{ID: "node.with.dots"}
	graph.AddNode(specialNode1)
	graph.AddNode(specialNode2)
	graph.AddNode(specialNode3)
	graph.AddEdge(specialNode1, specialNode2)
	graph.AddEdge(specialNode2, specialNode3)

	nodes, err = s.topoSort(graph)
	s.NoError(err)
	s.Len(nodes, 3)
}

// TestErrorTypes tests error types and messages
func (s *DagTestSuite) TestErrorTypes() {
	// Test ErrSelfLoop
	s.Equal("self-loop not allowed", ErrSelfLoop.Error())

	// Test ErrCycle
	cycleErr := ErrCycle{Path: []string{"A", "B", "C", "A"}}
	expectedMsg := "cycle detected: A -> B -> C -> A"
	s.Equal(expectedMsg, cycleErr.Error())

	// Test ErrNodeMissing
	s.Equal("node not found", ErrNodeMissing.Error())

	// Test ErrNotADAG
	s.Equal("graph contains a cycle (not a DAG)", ErrNotADAG.Error())
}

// Helper methods
func (s *DagTestSuite) findNodeIndex(nodes []TestNode, target TestNode) int {
	for i, node := range nodes {
		if node.ID == target.ID {
			return i
		}
	}
	return -1
}

func (s *DagTestSuite) containsNode(nodes []TestNode, target TestNode) bool {
	for _, node := range nodes {
		if node.ID == target.ID {
			return true
		}
	}
	return false
}

// topoSort is a helper method to call TopoSort on the concrete implementation
func (s *DagTestSuite) topoSort(graph G[TestNode]) ([]TestNode, error) {
	if impl, ok := graph.(*gImpl[TestNode]); ok {
		return impl.TopoSort()
	}
	return nil, fmt.Errorf("cannot cast to concrete implementation")
}

// TestTristateTestSuite runs the test suite
func TestDagTestSuite(t *testing.T) {
	suite.Run(t, new(DagTestSuite))
}
