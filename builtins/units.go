// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/sentrie-sh/sentrie/box"
)

const (
	maxUnitStringLen    = 128
	maxQuantityExponent = 400
)

var (
	declParseQuantity = &Decl{
		Name: "parse_quantity",
		Description: "Parses a Kubernetes-style quantity string (\"10Gi\", \"500m\", \"1.5k\", \"2e3\") to a number. " +
			"Suffixes: binary Ki Mi Gi Ti Pi Ei; decimal n u m k M G T P E; or a decimal exponent (e3, E-2). " +
			"A plain number is returned as is. \"500m\" is 0.5 (milli), not minutes.",
		DeriveSafe: true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "value", Kinds: kindString, KindError: "parse_quantity: argument must be a string"}},
			TooFewError:  "parse_quantity requires 1 argument",
			TooManyError: "parse_quantity requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindNumber},
		},
		Impl: implParseQuantity,
	}

	declParseDuration = &Decl{
		Name: "parse_duration",
		Description: "Parses a Go-style duration string (\"2h\", \"1h30m\", \"1.5s\", \"250ms\") to a number of seconds. " +
			"Units: ns, us (or µs), ms, s, m (minutes), h; a leading sign is allowed and components may be combined. " +
			"Days and weeks are not units. The result is in seconds, like now().",
		DeriveSafe: true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "value", Kinds: kindString, KindError: "parse_duration: argument must be a string"}},
			TooFewError:  "parse_duration requires 1 argument",
			TooManyError: "parse_duration requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindNumber},
		},
		Impl: implParseDuration,
	}
)

// quantitySuffixes is the Kubernetes resource.Quantity suffix table.
var quantitySuffixes = map[string]*big.Rat{
	"n":  big.NewRat(1, 1_000_000_000),
	"u":  big.NewRat(1, 1_000_000),
	"m":  big.NewRat(1, 1000),
	"":   big.NewRat(1, 1),
	"k":  big.NewRat(1000, 1),
	"M":  big.NewRat(1_000_000, 1),
	"G":  big.NewRat(1_000_000_000, 1),
	"T":  big.NewRat(1_000_000_000_000, 1),
	"P":  big.NewRat(1_000_000_000_000_000, 1),
	"E":  big.NewRat(1_000_000_000_000_000_000, 1),
	"Ki": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 10)),
	"Mi": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 20)),
	"Gi": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 30)),
	"Ti": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 40)),
	"Pi": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 50)),
	"Ei": new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 60)),
}

func implParseQuantity(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	if args[0].IsUndefined() {
		return box.Undefined(), nil
	}
	s, ok := args[0].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("parse_quantity: argument must be a string")
	}
	n, err := parseQuantity(s)
	if err != nil {
		return box.Undefined(), fmt.Errorf("parse_quantity: %w", err)
	}
	return box.Number(n), nil
}

func parseQuantity(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty quantity")
	}
	if len(s) > maxUnitStringLen {
		return 0, fmt.Errorf("quantity is longer than %d characters", maxUnitStringLen)
	}
	i := 0
	if s[i] == '+' || s[i] == '-' {
		i++
	}
	digits, dot := 0, false
	for ; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			digits++
		} else if c == '.' && !dot {
			dot = true
		} else {
			break
		}
	}
	if digits == 0 {
		return 0, fmt.Errorf("invalid quantity %q: no number", s)
	}
	mantissa, rest := s[:i], s[i:]

	r, ok := new(big.Rat).SetString(strings.TrimPrefix(mantissa, "+"))
	if !ok {
		return 0, fmt.Errorf("invalid quantity %q", s)
	}

	factor, known := quantitySuffixes[rest]
	if !known {
		exp, isExp := parseQuantityExponent(rest)
		if !isExp {
			return 0, fmt.Errorf("invalid quantity %q: unknown suffix %q", s, rest)
		}
		if exp > maxQuantityExponent || exp < -maxQuantityExponent {
			return 0, fmt.Errorf("quantity %q is out of range", s)
		}
		pow := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(max(exp, -exp))), nil)
		factor = new(big.Rat).SetInt(pow)
		if exp < 0 {
			factor = new(big.Rat).Inv(factor)
		}
	}

	f, _ := r.Mul(r, factor).Float64()
	if math.IsInf(f, 0) || (f == 0 && r.Sign() != 0) {
		return 0, fmt.Errorf("quantity %q is out of range", s)
	}
	return f, nil
}

// parseQuantityExponent reads a decimal exponent such as "e3" or "E-2". A bare
// "E" is the exa suffix and is handled by the suffix table instead.
func parseQuantityExponent(rest string) (int, bool) {
	if len(rest) < 2 || (rest[0] != 'e' && rest[0] != 'E') {
		return 0, false
	}
	body := rest[1:]
	digits := strings.TrimLeft(body, "+-")
	if len(body)-len(digits) > 1 || digits == "" || strings.Trim(digits, "0123456789") != "" {
		return 0, false
	}
	n, err := strconv.Atoi(body)
	if err != nil {
		return 0, false
	}
	return n, true
}

func implParseDuration(_ context.Context, _ Env, args ...box.Value) (box.Value, error) {
	if args[0].IsUndefined() {
		return box.Undefined(), nil
	}
	s, ok := args[0].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("parse_duration: argument must be a string")
	}
	if s == "" {
		return box.Undefined(), fmt.Errorf("parse_duration: empty duration")
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return box.Undefined(), fmt.Errorf("parse_duration: %w", err)
	}
	return box.Number(d.Seconds()), nil
}
