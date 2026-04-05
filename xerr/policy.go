// SPDX-License-Identifier: Apache-2.0
//
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

package xerr

import pkgerrors "github.com/pkg/errors"

// Policy indexing: literal metadata and header ordering. Each value wraps ErrIndex;
// add location at the call site with errors.Wrapf(err, "at %s", span).
var (
	ErrPolicyMetadataContiguous = pkgerrors.Wrap(ErrIndex, "title, description, version, and tag may only appear in one contiguous block at the top of the policy, before all fact and use statements.")
	ErrPolicyFactAfterUse       = pkgerrors.Wrap(ErrIndex, "fact statements must appear before any use statements.")
	ErrPolicyInvalidVersion     = pkgerrors.Wrap(ErrIndex, `Invalid policy version: expected SemVer string (e.g., "1.2.3").`)
	ErrPolicyEmptyTitle         = pkgerrors.Wrap(ErrIndex, "policy title must not be empty or whitespace-only.")
	ErrPolicyEmptyTagKey        = pkgerrors.Wrap(ErrIndex, "tag key must not be empty or whitespace-only.")
)
