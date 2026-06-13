// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/sentrie-sh/sentrie/box"
)

// Env is the complete set of capabilities a builtin may use from the
// evaluation engine. It is deliberately narrow: widening this interface is
// the visible review event for "does this builtin remain DeriveSafe?".
//
// DeriveSafe means: deterministic output within a single policy execution.
// It does NOT mean deterministic across executions — now is DeriveSafe
// because ExecutionStart is pinned per execution.
type Env interface {
	Call(ctx context.Context, fn box.Value, args []box.Value) (box.Value, error)
	CallableArity(fn box.Value) (int, error)
	ExecutionStart() time.Time
}

// Fn is a builtin implementation. It may rely on the dispatcher having
// already enforced Decl.Sig arity and kind rules (see Precheck), with the
// exception of undefined/null arguments, which are always passed through.
type Fn func(ctx context.Context, env Env, args ...box.Value) (box.Value, error)

// MismatchPolicy controls what Precheck does when a defined argument's kind
// is outside ParamSig.Kinds.
type MismatchPolicy uint8

const (
	MismatchError MismatchPolicy = iota
	MismatchUndefined
)

type ParamSig struct {
	Name              string
	Kinds             []box.ValueKind
	Optional          bool
	CallableArities   []int
	OnMismatch        MismatchPolicy
	KindError         string
	CallableArityError string
}

type Sig struct {
	Params       []ParamSig
	Variadic     *ParamSig
	Result       ParamSig
	TooFewError  string
	TooManyError string
}

type Decl struct {
	Name        string
	Description string
	Sig         Sig
	DeriveSafe  bool
	Impl        Fn
}

var (
	kindList      = []box.ValueKind{box.ValueList}
	kindDict      = []box.ValueKind{box.ValueDict}
	kindNumber    = []box.ValueKind{box.ValueNumber}
	kindBool      = []box.ValueKind{box.ValueBool}
	kindCallable  = []box.ValueKind{box.ValueCallable}
	kindCountable = []box.ValueKind{box.ValueList, box.ValueDict, box.ValueString}
)

// Table is the registry. Map key MUST equal Decl.Name (TestTableWellFormed).
var Table = map[string]*Decl{
	"all":            declAll,
	"any":            declAny,
	"as_list":        declAsList,
	"collect":        declCollect,
	"count":          declCount,
	"distinct":       declDistinct,
	"error":          declError,
	"filter":         declFilter,
	"first":          declFirst,
	"flatten":        declFlatten,
	"flatten_deep":   declFlattenDeep,
	"merge":          declMerge,
	"normalise_list": declNormaliseList,
	"now":            declNow,
	"reduce":         declReduce,
}

// IsDeriveSafe reports whether name is a builtin allowed inside derive bodies.
func IsDeriveSafe(name string) bool {
	d, ok := Table[name]
	return ok && d.DeriveSafe
}

// DeriveSafeNames returns a sorted copy of derive-safe builtin names.
func DeriveSafeNames() []string {
	out := make([]string, 0, len(Table))
	for name, d := range Table {
		if d.DeriveSafe {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}

// Precheck validates args against d.Sig.
func (d *Decl) Precheck(env Env, args []box.Value) (handled bool, val box.Value, err error) {
	sig := d.Sig

	min := 0
	for _, p := range sig.Params {
		if !p.Optional {
			min++
		}
	}
	max := len(sig.Params)

	if sig.Variadic == nil {
		if len(args) > max {
			msg := sig.TooManyError
			if msg == "" {
				msg = sig.TooFewError
			}
			return false, box.Undefined(), fmt.Errorf("%s", msg)
		}
	}

	if len(args) < min {
		return false, box.Undefined(), fmt.Errorf("%s", sig.TooFewError)
	}

	for i, arg := range args {
		ps, ok := paramSigAt(sig, i)
		if !ok {
			break
		}

		if arg.Kind() == box.ValueUndefined || arg.Kind() == box.ValueNull {
			continue
		}

		if len(ps.Kinds) > 0 && !kindAllowed(ps.Kinds, arg.Kind()) {
			if ps.OnMismatch == MismatchUndefined {
				return true, box.Undefined(), nil
			}
			if ps.KindError != "" {
				return false, box.Undefined(), fmt.Errorf("%s", ps.KindError)
			}
			if slices.Contains(ps.Kinds, box.ValueCallable) {
				_, err := env.CallableArity(arg)
				return false, box.Undefined(), err
			}
		}

		if len(ps.CallableArities) > 0 {
			n, err := env.CallableArity(arg)
			if err != nil {
				return false, box.Undefined(), err
			}
			if !slices.Contains(ps.CallableArities, n) {
				return false, box.Undefined(), fmt.Errorf("%s", ps.CallableArityError)
			}
		}
	}

	return false, box.Undefined(), nil
}

func paramSigAt(sig Sig, i int) (ParamSig, bool) {
	if i < len(sig.Params) {
		return sig.Params[i], true
	}
	if sig.Variadic != nil {
		return *sig.Variadic, true
	}
	return ParamSig{}, false
}

func kindAllowed(kinds []box.ValueKind, k box.ValueKind) bool {
	return slices.Contains(kinds, k)
}

func hofSig(name, callableArityErr string) Sig {
	return Sig{
		Params: []ParamSig{
			{Name: "collection", Kinds: kindList, KindError: name + ": first argument must be a list"},
			{
				Name:               "predicate",
				Kinds:              kindCallable,
				CallableArities:    []int{1, 2},
				CallableArityError: callableArityErr,
			},
		},
		TooFewError:  name + " requires 2 arguments",
		TooManyError: name + " requires 2 arguments",
		Result:       ParamSig{Name: "result", Kinds: kindBool},
	}
}
