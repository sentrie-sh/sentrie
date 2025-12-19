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

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

type RuleExportStatement struct {
	*baseNode
	Of          string              // Name of the exported variable or decision
	Attachments []*AttachmentClause // Optional attachments for the export
}

type AttachmentClause struct {
	*baseNode
	What string     // Name of the attachment
	As   Expression // Value of the attachment
}

func NewAttachmentClause(what string, as Expression, ssp tokens.Range) *AttachmentClause {
	return &AttachmentClause{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "attachment_clause",
		},
		What: what,
		As:   as,
	}
}

func NewRuleExportStatement(of string, attachments []*AttachmentClause, ssp tokens.Range) *RuleExportStatement {
	return &RuleExportStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "rule_export",
		},
		Of:          of,
		Attachments: attachments,
	}
}

func (v RuleExportStatement) statementNode() {}

func (v RuleExportStatement) String() string {
	return v.Of
}

var _ Statement = &RuleExportStatement{}
var _ Node = &RuleExportStatement{}

func (a AttachmentClause) String() string {
	return fmt.Sprintf("attach %s as %s", a.What, a.As)
}
func (a *AttachmentClause) expressionNode() {}

var _ Expression = &AttachmentClause{}
var _ Node = &AttachmentClause{}
