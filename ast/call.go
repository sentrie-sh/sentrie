// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"fmt"
	"strings"
	"time"

	"github.com/sentrie-sh/sentrie/tokens"
)

type CallExpression struct {
	*baseNode
	Callee     Expression
	Arguments  []Expression
	Memoized   bool
	MemoizeTTL *time.Duration

	// Piped and PipeArgs record the source form of a call written as a
	// pipeline stage, `Piped |> Callee(PipeArgs...)`. The parser desugars the
	// stage into Arguments, which is all evaluation reads; these two fields
	// only let a printer reproduce the pipeline. Piped is nil for a plain call.
	Piped    Expression
	PipeArgs []Expression
}

func NewCallExpression(callee Expression, arguments []Expression, memoized bool, memoizeTTL *time.Duration, ssp tokens.Range) *CallExpression {
	return &CallExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "call",
		},
		Callee:     callee,
		Arguments:  arguments,
		Memoized:   memoized,
		MemoizeTTL: memoizeTTL,
	}
}

func (c *CallExpression) String() string {
	args := make([]string, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", c.Callee.String(), strings.Join(args, ", "))
}

func (c *CallExpression) expressionNode() {}

var _ Expression = &CallExpression{}
var _ Node = &CallExpression{}
