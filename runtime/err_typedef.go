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

package runtime

import (
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

var (
	ErrTypeRef           = errors.New("typeref error")
	errConstraintFailed  = errors.Wrapf(ErrTypeRef, "constraint failed")
	errUnknownConstraint = errors.Wrapf(ErrTypeRef, "unknown constraint")
)

func ErrUnknownConstraint(c *ast.TypeRefConstraint) error {
	return errors.Wrapf(errUnknownConstraint, "unknown constraint: '%s' at %s", c.Name, c.Span())
}

func ErrConstraintFailed(pos tokens.Range, c *ast.TypeRefConstraint, err error) error {
	return errors.Wrapf(errConstraintFailed, "constraint failed: '%s' at %s", c.Name, c.Span())
}

func IsUnknownConstraint(err error) bool {
	return errors.Is(err, errUnknownConstraint)
}

func IsConstraintFailed(err error) bool {
	return errors.Is(err, errConstraintFailed)
}
