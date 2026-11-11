// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	stdErr "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

// RuleError is one rule's failure, carried in that rule's ExecutorOutput as
// its "error" field (#143). It reads like an RFC 9457 problem document so a
// client can treat a failed rule the way it treats a failed request: Status is
// the HTTP status the failure maps to (see ErrorStatus) and Detail is the
// error message. The underlying error is kept for errors.Is and errors.As.
type RuleError struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
	err    error
}

// ruleErrorTitle is the title of every rule error, matching the problem
// document serve returns for a failed evaluation.
const ruleErrorTitle = "Evaluation Failed"

func newRuleError(err error) *RuleError {
	return &RuleError{Title: ruleErrorTitle, Status: ErrorStatus(err), Detail: err.Error(), err: err}
}

func (e *RuleError) Error() string { return e.Detail }

func (e *RuleError) Unwrap() error { return e.err }

// RuleErrors is the error ExecPolicy returns when at least one rule failed and
// the policy itself resolved. Every rule still has an output, and each failing
// one carries its RuleError, so a caller that wants partial results checks for
// this type with errors.As rather than discarding the outputs. Its message is
// the rule messages, one per line, in no particular order.
type RuleErrors []*RuleError

func (e RuleErrors) Error() string {
	msgs := make([]string, len(e))
	for i, re := range e {
		msgs[i] = re.Error()
	}
	return strings.Join(msgs, "\n")
}

func (e RuleErrors) Unwrap() []error {
	errs := make([]error, len(e))
	for i, re := range e {
		errs[i] = re
	}
	return errs
}

// ErrorStatus maps an evaluation error to the HTTP status serve reports for
// it: 400 for an invalid invocation (a missing or ill-typed fact), 503 when the
// evaluation deadline passed, and 500 for anything else.
func ErrorStatus(err error) int {
	var invErr xerr.InvalidInvocationError
	switch {
	case stdErr.As(err, &invErr):
		return http.StatusBadRequest
	case stdErr.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// withRuleError makes sure a failed rule has an output carrying its error.
// A rule that failed before it was evaluated (an unexported rule, a missing
// fact) has no output yet, so it gets one reporting unknown.
func withRuleError(output *ExecutorOutput, namespace, policy, rule string, err error) *ExecutorOutput {
	if err == nil {
		return output
	}
	if output == nil {
		output = &ExecutorOutput{
			PolicyName: policy,
			Namespace:  namespace,
			RuleName:   rule,
			Decision:   DecisionOf(box.Trinary(trinary.Unknown)).reported(),
		}
	}
	output.Error = newRuleError(err)
	return output
}

// factError is a fact that failed its declared type, found by execRule for
// the execution context ec before the rule evaluates. name is the fact's
// declared name, the key the caller supplied it under, and isDefault says the
// failing value is the policy's own default rather than a supplied one.
type factError struct {
	ec        *ExecutionContext
	name      string
	isDefault bool
	err       error
}

func (e *factError) Error() string {
	if e.isDefault {
		return fmt.Sprintf("default value of fact '%s': %s", e.name, e.err.Error())
	}
	return fmt.Sprintf("fact '%s': %s", e.name, e.err.Error())
}

func (e *factError) Unwrap() error { return e.err }

// requestError marks the error of a top-level call whose facts did not bind or
// validate. It is the request's error rather than one rule's (#143), so
// ExecPolicy returns it alone. It adds nothing to the message.
type requestError struct{ err error }

func (e *requestError) Error() string { return e.err.Error() }

func (e *requestError) Unwrap() error { return e.err }

// asRequestError marks err as a request error when it arose while binding the
// facts of ec (bindingFacts) or validating them (a factError for ec, not for
// an imported policy's context). A context error stays a rule failure: a
// deadline is evaluation time, whatever it interrupted.
//
// A supplied fact that fails its type is the caller's mistake, so it is an
// invalid invocation (400 on serve, #147). A default that fails its type is the
// policy's, so it stays a 500; the index rejects the statically wrong ones.
func asRequestError(ec *ExecutionContext, bindingFacts bool, err error) error {
	if err == nil || stdErr.Is(err, context.DeadlineExceeded) || stdErr.Is(err, context.Canceled) {
		return err
	}
	// An imported policy's wrongly typed fact, already marked by
	// asImportedFactError, stays the request's error.
	if reqErr := (*requestError)(nil); stdErr.As(err, &reqErr) {
		return err
	}
	var fe *factError
	if stdErr.As(err, &fe) && fe.ec == ec {
		if !fe.isDefault {
			err = fmt.Errorf("%w: %w", err, xerr.ErrInvalidInvocation(""))
		}
		return &requestError{err: err}
	}
	if bindingFacts {
		return &requestError{err: err}
	}
	return err
}

// asImportedFactError marks err as a request error when it is a fact error of
// the imported policy's context ec caused by the importing policy's `with`:
// a supplied value failing its declared type, or a binding error
// (withFactError) of a required fact with no `with` or a null for a fact that
// does not accept null. Like a caller's fact error it is an invalid
// invocation, so the whole request fails with a 400 (#147, user decision 11;
// the user confirmed the same for the missing fact). The imported policy's
// own default failing its type and every other error stay the importing
// rule's failure.
func asImportedFactError(ec *ExecutionContext, withFactError bool, err error) error {
	if err == nil || stdErr.Is(err, context.DeadlineExceeded) || stdErr.Is(err, context.Canceled) {
		return err
	}
	// The message names the imported policy: the fact is its, not the caller's.
	if withFactError {
		return &requestError{err: fmt.Errorf("imported policy %s: %w", ec.policy.FQN, err)}
	}
	var fe *factError
	if stdErr.As(err, &fe) && fe.ec == ec && !fe.isDefault {
		return &requestError{err: fmt.Errorf("imported policy %s: %w: %w", ec.policy.FQN, err, xerr.ErrInvalidInvocation(""))}
	}
	return err
}
