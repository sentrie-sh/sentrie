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

import "github.com/binaek/sentra/tokens"

type RuleExportStatement struct {
	Pos         tokens.Position     // Position in the source code
	Of          string              // Name of the exported variable or decision
	Attachments []*AttachmentClause // Optional attachments for the export
}

func (v RuleExportStatement) String() string {
	return v.Of
}

func (v RuleExportStatement) Position() tokens.Position {
	return v.Pos
}

func (v RuleExportStatement) statementNode() {}

var _ Statement = &RuleExportStatement{}

type AttachmentClause struct {
	Pos  tokens.Position // Position in the source code
	What string          // Name of the attachment
	As   string          // Value of the attachment
}
