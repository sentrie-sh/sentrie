// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	sentrt "github.com/sentrie-sh/sentrie/runtime"
)

// defaultBenchtime matches `go test -benchtime`.
const defaultBenchtime = "1s"

// maxBenchIterations bounds the iteration count of one timed run, as
// `go test` does.
const maxBenchIterations = 1_000_000_000

func addBenchCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("bench", benchCmd).
			WithDescription("Evaluate a rule or policy repeatedly and print Go benchmark lines (usable with benchstat)").
			WithArgument(cling.NewStringCmdInput("target").
				WithDescription("Rule or policy to benchmark, as for exec").
				Required().
				AsArgument(),
			).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault(".").
				WithDescription("Pack directory to load").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("fact-file").
				WithDefault("").
				WithDescription("File to load facts from").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("facts").
				WithDefault("{}").
				WithDescription("Facts to evaluate the target with").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("benchtime").
				WithDefault(defaultBenchtime).
				WithDescription("Measure for this long (a Go duration such as 1s) or for exactly this many evaluations (Nx, such as 100x), as for go test").
				AsFlag(),
			).
			WithFlag(cling.
				NewIntCmdInput("count").
				WithDefault(1).
				WithDescription("Repeat the whole measurement this many times, one output line per run").
				AsFlag(),
			).
			WithFlag(evalTimeoutFlag()).
			WithFlag(capabilitiesFlag()).
			WithFlag(skipEngineCheckFlag()),
	)
}

type benchCmdArgs struct {
	PackLocation    string `cling-name:"pack-location"`
	Target          string `cling-name:"target"`
	Facts           string `cling-name:"facts"`
	FactFile        string `cling-name:"fact-file"`
	Benchtime       string `cling-name:"benchtime"`
	Count           int    `cling-name:"count"`
	EvalTimeout     string `cling-name:"eval-timeout"`
	Capabilities    string `cling-name:"capabilities"`
	SkipEngineCheck bool   `cling-name:"skip-engine-check"`
}

func benchCmd(ctx context.Context, args []string) error {
	input := benchCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	spec, err := parseBenchtime(input.Benchtime)
	if err != nil {
		return err
	}
	if input.Count < 1 {
		return fmt.Errorf("invalid --count %d: must be at least 1", input.Count)
	}
	evalTimeout, err := parseEvalTimeout(input.EvalTimeout)
	if err != nil {
		return err
	}
	facts, err := loadBenchFacts(input.FactFile, input.Facts)
	if err != nil {
		return err
	}

	// Load, index and validate once; none of it is inside the timed loop.
	idx, err := loadBenchIndex(ctx, input.PackLocation, input.SkipEngineCheck)
	if err != nil {
		return err
	}
	exec, err := newGrantedExecutor(ctx, idx, input.Capabilities, sentrt.WithEvalTimeout(evalTimeout))
	if err != nil {
		return err
	}
	defer func() { _ = exec.Close() }()

	namespace, policy, rule, err := exec.Index().ResolveSegments(input.Target)
	if err != nil {
		return err
	}

	op := func() error {
		if rule == "" {
			_, err := exec.ExecPolicy(ctx, namespace, policy, facts)
			return err
		}
		_, err := exec.ExecRule(ctx, namespace, policy, rule, facts)
		return err
	}

	// One untimed evaluation warms caches and module loads, and surfaces an
	// unknown target or a failing evaluation before anything is printed.
	if err := op(); err != nil {
		return fmt.Errorf("benchmark target %q failed to evaluate: %w", input.Target, err)
	}

	name := benchName(input.Target)
	out := os.Stdout
	_, _ = fmt.Fprintf(out, "goos: %s\ngoarch: %s\n", runtime.GOOS, runtime.GOARCH)
	for range input.Count {
		res, err := runBench(spec, op)
		if err != nil {
			return fmt.Errorf("benchmark target %q failed to evaluate: %w", input.Target, err)
		}
		_, _ = fmt.Fprintln(out, res.line(name))
	}
	return nil
}

// loadBenchFacts merges the fact file and the --facts flag, the flag winning,
// exactly as exec does.
func loadBenchFacts(factFile, factsJSON string) (map[string]any, error) {
	facts := make(map[string]any)
	if factFile != "" {
		content, err := os.ReadFile(factFile)
		if err != nil {
			return nil, err
		}
		var fromFile map[string]any
		if err := json.NewDecoder(bytes.NewReader(content)).Decode(&fromFile); err != nil {
			return nil, err
		}
		facts = mergeFacts(facts, fromFile)
	}
	var fromFlag map[string]any
	if err := json.NewDecoder(strings.NewReader(factsJSON)).Decode(&fromFlag); err != nil {
		return nil, fmt.Errorf("invalid --facts: %w", err)
	}
	// Merged as exec merges them: objects deeply, lists replaced (#147).
	return mergeFacts(facts, fromFlag), nil
}

// loadBenchIndex runs the load, engine check, index and validate steps that
// `exec` runs, for `bench`.
func loadBenchIndex(ctx context.Context, packLocation string, skipEngineCheck bool) (*index.Index, error) {
	pack, err := loader.LoadPack(ctx, packLocation)
	if err != nil {
		return nil, err
	}
	if err := checkEngine(pack, skipEngineCheck, os.Stderr); err != nil {
		return nil, err
	}
	if err := printUnenforcedPermissionWarnings(os.Stderr, pack); err != nil {
		return nil, err
	}
	idx := index.CreateIndex()
	if err := idx.SetPack(ctx, pack); err != nil {
		return nil, err
	}
	programs, err := loader.LoadPrograms(ctx, pack)
	if err != nil {
		return nil, err
	}
	for _, program := range programs {
		if err := idx.AddProgram(ctx, program); err != nil {
			return nil, err
		}
	}
	if err := idx.Validate(ctx); err != nil {
		return nil, err
	}
	return idx, nil
}

// benchSpec is a parsed --benchtime: exactly N evaluations when N > 0,
// otherwise as many as fit in Duration.
type benchSpec struct {
	N        int
	Duration time.Duration
}

// parseBenchtime parses a --benchtime value with the meaning it has in
// `go test`: a positive duration, or a positive count followed by "x".
func parseBenchtime(v string) (benchSpec, error) {
	if num, ok := strings.CutSuffix(v, "x"); ok {
		n, err := strconv.Atoi(num)
		if err != nil || n <= 0 {
			return benchSpec{}, fmt.Errorf("invalid --benchtime %q: the count before x must be a positive integer", v)
		}
		return benchSpec{N: n}, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return benchSpec{}, fmt.Errorf("invalid --benchtime %q: want a duration such as 1s or a count such as 100x: %w", v, err)
	}
	if d <= 0 {
		return benchSpec{}, fmt.Errorf("invalid --benchtime %q: must be positive", v)
	}
	return benchSpec{Duration: d}, nil
}

// benchName is the benchmark name printed for target. It carries the
// GOMAXPROCS suffix `go test` adds, so output lines compare with Go's own,
// and has no whitespace, which would end the name field.
func benchName(target string) string {
	clean := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		return r
	}, target)
	name := "BenchmarkExec/" + clean
	if procs := runtime.GOMAXPROCS(0); procs != 1 {
		name += "-" + strconv.Itoa(procs)
	}
	return name
}

// benchResult is one measured run.
type benchResult struct {
	N       int
	Elapsed time.Duration
	// Bytes and Mallocs are the process-wide heap allocation deltas over the run.
	Bytes   uint64
	Mallocs uint64
}

// line renders the result as a Go benchmark line.
func (r benchResult) line(name string) string {
	n := float64(r.N)
	nsPerOp := float64(r.Elapsed.Nanoseconds()) / n
	return fmt.Sprintf("%s\t%8d\t%12.2f ns/op\t%10d B/op\t%10d allocs/op",
		name, r.N, nsPerOp, r.Bytes/uint64(r.N), r.Mallocs/uint64(r.N))
}

// runN times n calls of op, with heap allocation deltas taken from
// runtime.MemStats outside the timed window.
func runN(n int, op func() error) (benchResult, error) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	for range n {
		if err := op(); err != nil {
			return benchResult{}, err
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	return benchResult{
		N:       n,
		Elapsed: elapsed,
		Bytes:   after.TotalAlloc - before.TotalAlloc,
		Mallocs: after.Mallocs - before.Mallocs,
	}, nil
}

// runBench measures op under spec. A fixed count runs once; a duration grows
// the iteration count until a run lasts at least the duration, as go test
// does, and reports that last run.
func runBench(spec benchSpec, op func() error) (benchResult, error) {
	if spec.N > 0 {
		return runN(spec.N, op)
	}
	n := 1
	for {
		res, err := runN(n, op)
		if err != nil {
			return benchResult{}, err
		}
		if res.Elapsed >= spec.Duration || n >= maxBenchIterations {
			return res, nil
		}
		perOp := max(float64(res.Elapsed.Nanoseconds())/float64(n), 1)
		next := int(min(float64(spec.Duration.Nanoseconds())*1.2/perOp, float64(maxBenchIterations)))
		next = min(next, 100*n)
		next = max(next, n+1)
		n = min(next, maxBenchIterations)
	}
}
