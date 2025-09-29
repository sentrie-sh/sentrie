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

package trace

import (
	"fmt"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
)

// Node captures a single evaluation step in the decision tree.
type Node struct {
	// Kind is a high-level category: "literal", "identifier", "unary", "infix",
	// "block", "field", "index", "call", "import", "ternary", "quantifier",
	// "reduce", "transform", "rule", "policy"
	Kind string `json:"kind"`

	// Op is the operator or sub-kind (e.g., "not", "+", "any", "map", "filter"),
	// or rule/policy name for those node kinds.
	Op string `json:"op,omitempty"`

	// Duration is the time taken to evaluate the node.
	Duration time.Duration `json:"duration,omitempty"`

	// Node is the AST node that this node is associated with.
	Node ast.Node `json:"-"`

	// Meta holds node-specific metadata (e.g., field name, callee alias, etc.).
	Meta map[string]any `json:"meta,omitempty"`

	// Children are the nested steps under this node.
	Children []*Node `json:"children,omitempty"`

	// Result is the exported Go value resulting from this node's evaluation.
	Result any `json:"result,omitempty"`

	// Err (if set) is the error message produced during evaluation of this node.
	Err string `json:"err,omitempty"`
}

type DoneFn func()

// Helper to create a node with meta.
func New(kind, op string, n ast.Node, meta map[string]any) (*Node, DoneFn) {
	x := &Node{Kind: kind, Op: op, Node: n, Meta: meta}
	start := time.Now()
	return x, func() {
		x.Duration = time.Since(start)
	}
}

func IgnoredStmt(n ast.Node) *Node {
	return &Node{Kind: "stmt-ignored", Op: "", Node: n, Meta: map[string]any{"type": fmt.Sprintf("%T", n)}}
}

func UnsupportedExpression(n ast.Node) *Node {
	return &Node{Kind: "unsupported", Op: "", Node: n, Meta: map[string]any{"type": fmt.Sprintf("%T", n)}}
}

// Attach adds children and returns self for chaining.
func (n *Node) Attach(children ...*Node) *Node {
	if len(children) == 0 {
		return n
	}
	n.Children = append(n.Children, children...)
	return n
}

// SetResult sets the node’s result and returns self.
func (n *Node) SetResult(v any) *Node {
	n.Result = v
	return n
}

// SetErr annotates the node with an error string.
func (n *Node) SetErr(err error) *Node {
	if err != nil {
		n.Err = err.Error()
	}
	return n
}
