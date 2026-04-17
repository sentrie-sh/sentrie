// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import "github.com/sentrie-sh/sentrie/tokens"

// PipelineHoleExpression is a placeholder used by pipeline lowering.
type PipelineHoleExpression struct {
	*baseNode
}

func NewPipelineHoleExpression(ssp tokens.Range) *PipelineHoleExpression {
	return &PipelineHoleExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "pipeline_hole",
		},
	}
}

func (p *PipelineHoleExpression) String() string {
	return "#"
}

func (p *PipelineHoleExpression) expressionNode() {}

var _ Expression = &PipelineHoleExpression{}
var _ Node = &PipelineHoleExpression{}
