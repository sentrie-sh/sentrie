// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/format"
)

func addFmtCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("fmt", fmtCmd).
			WithFlag(cling.NewBoolCmdInput("check").
				WithDefault(false).
				WithDescription("Write nothing; list the files that would change and fail if there are any").
				AsFlag(),
			).
			WithFlag(cling.NewBoolCmdInput("diff").
				WithDefault(false).
				WithDescription("Write nothing; print a unified diff for each file that would change").
				AsFlag(),
			),
	)
}

type fmtOptions struct {
	check bool
	diff  bool
	paths []string
}

// parseFmtArgs reads the arguments of `sentrie fmt` itself instead of using
// cling.Hydrate. cling takes the word after a bare `--check` as that flag's
// value and passes it on as `--check=<word>`, and its arguments are fixed
// positions, so `fmt --check dir` would lose the path. A flag value is therefore
// taken as a path when it names an existing path or is not a boolean, and such
// paths are processed after the positional ones, in sorted order.
func parseFmtArgs(args []string) (fmtOptions, error) {
	var opts fmtOptions
	var flagPaths []string
	for _, arg := range args {
		name, value, hasValue := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
		if !strings.HasPrefix(arg, "--") || (name != "check" && name != "diff") {
			// An unknown flag only reaches here when parseFmtArgs is called
			// directly: cling rejects undefined flags before the command runs.
			if strings.HasPrefix(arg, "--") {
				return opts, fmt.Errorf("unknown flag %q for fmt", arg)
			}
			opts.paths = append(opts.paths, arg)
			continue
		}
		on := true
		if hasValue && value != "" {
			// A value naming an existing path is a path, even when it also
			// reads as a boolean (`--check f` with a directory `f`).
			if _, err := os.Stat(value); err == nil {
				flagPaths = append(flagPaths, value)
			} else if b, err := strconv.ParseBool(value); err == nil {
				on = b
			} else {
				flagPaths = append(flagPaths, value)
			}
		}
		if name == "check" {
			opts.check = opts.check || on
		} else {
			opts.diff = opts.diff || on
		}
	}
	slices.Sort(flagPaths)
	opts.paths = append(opts.paths, flagPaths...)
	if len(opts.paths) == 0 {
		opts.paths = []string{"."}
	}
	return opts, nil
}

// fmtCmd formats .sentrie files in place. With --check or --diff it writes
// nothing. --check lists the files that would change and fails when there are
// any. --diff prints a unified diff of each and, like gofmt -d, succeeds on
// its own; combined with --check it fails too. A file that cannot be read,
// parsed or formatted is reported on stderr and fails the command, and the
// remaining files are still processed.
func fmtCmd(ctx context.Context, args []string) error {
	opts, err := parseFmtArgs(args)
	if err != nil {
		return err
	}
	return runFmt(ctx, os.Stdout, os.Stderr, opts)
}

func runFmt(ctx context.Context, stdout, stderr io.Writer, opts fmtOptions) error {
	files, walkErrs := collectFmtFiles(ctx, opts.paths)
	for _, e := range walkErrs {
		_, _ = fmt.Fprintf(stderr, "%v\n", e)
	}
	failed := len(walkErrs)
	changed := 0

	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		did, err := fmtFile(ctx, stdout, path, opts)
		if err != nil {
			// every error here already names the file (parse and format errors
			// carry the filename; os errors carry the path)
			_, _ = fmt.Fprintf(stderr, "%v\n", err)
			failed++
			continue
		}
		if did {
			changed++
		}
	}

	switch {
	case failed > 0:
		return fmt.Errorf("%d file(s) could not be formatted", failed)
	case opts.check && changed > 0:
		return fmt.Errorf("%d file(s) are not formatted", changed)
	}
	return nil
}

// fmtFile formats one file and reports whether its content differs from the
// formatted text.
func fmtFile(ctx context.Context, stdout io.Writer, path string, opts fmtOptions) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := format.Source(ctx, path, src)
	if err != nil {
		return false, err
	}
	if bytes.Equal(src, out) {
		return false, nil
	}

	switch {
	case opts.diff:
		_, err = io.WriteString(stdout, unifiedDiff(filepath.ToSlash(path), string(src), string(out)))
		if err == nil && opts.check {
			_, err = fmt.Fprintln(stdout, path)
		}
	case opts.check:
		_, err = fmt.Fprintln(stdout, path)
	default:
		err = writeFileInPlace(path, src, out, info.Mode().Perm())
	}
	return true, err
}

// rewriteTarget is the part of *os.File that writeFileInPlace uses; a test
// replaces openForRewrite to inject write failures.
type rewriteTarget interface {
	Write(p []byte) (int, error)
	WriteAt(p []byte, off int64) (int, error)
	Truncate(size int64) error
	Close() error
}

var openForRewrite = func(path string) (rewriteTarget, error) {
	return os.OpenFile(path, os.O_WRONLY, 0)
}

// writeFileInPlace rewrites path with data following gofmt -w. It saves the
// original bytes in a backup beside the file, opens the file write-only
// without truncating it (so a file that cannot be opened is never touched),
// writes data over the start and truncates to its length, then removes the
// backup. If the write fails after bytes were written, the original is written
// back and the backup removed; the backup is kept, and named in the error, only
// when that restore fails too. If the backup cannot be made the file is left
// untouched. Writing in place keeps symlinks, hard links, owner and xattrs.
func writeFileInPlace(path string, orig, data []byte, mode fs.FileMode) error {
	bak, err := backupFile(path, orig, mode)
	if err != nil {
		return fmt.Errorf("%s: cannot back up before writing: %w", path, err)
	}
	f, err := openForRewrite(path)
	if err != nil {
		_ = os.Remove(bak)
		return err
	}
	n, err := f.Write(data)
	writeFailed := err != nil
	if err == nil {
		err = f.Truncate(int64(n))
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		_ = os.Remove(bak)
		return nil
	}
	if writeFailed && n == 0 {
		// nothing reached the file
		_ = os.Remove(bak)
		return fmt.Errorf("%s: %w", path, err)
	}
	if rerr := restoreFile(path, orig); rerr != nil {
		return fmt.Errorf("%s: %w (restoring it failed: %v; original kept in %s)", path, err, rerr, bak)
	}
	_ = os.Remove(bak)
	return fmt.Errorf("%s: %w (original restored)", path, err)
}

// restoreFile writes orig back over path.
func restoreFile(path string, orig []byte) error {
	f, err := openForRewrite(path)
	if err != nil {
		return err
	}
	_, err = f.WriteAt(orig, 0)
	if err == nil {
		err = f.Truncate(int64(len(orig)))
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// backupFile writes orig to a new file in the directory of path and returns
// its name.
func backupFile(path string, orig []byte, mode fs.FileMode) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".bak-*")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, err = f.Write(orig)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(name, mode)
	}
	if err != nil {
		_ = os.Remove(name)
		return "", err
	}
	return name, nil
}

// collectFmtFiles expands the paths into the files to format, in the order
// given and each file once. A directory contributes its *.sentrie files,
// recursively, skipping hidden directories and symbolic links. A path named
// directly is formatted whatever its extension.
func collectFmtFiles(ctx context.Context, paths []string) ([]string, []error) {
	var files []string
	var errs []error
	seen := map[string]bool{}
	add := func(p string) {
		key := filepath.Clean(p)
		if r, err := filepath.EvalSymlinks(p); err == nil {
			key = r
		}
		if !seen[key] {
			seen[key] = true
			files = append(files, p)
		}
	}

	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !info.IsDir() {
			add(root)
			continue
		}
		// WalkDir does not descend into a root that is a symlink, so walk the
		// resolved directory and report paths under the name the user gave.
		walkRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = filepath.WalkDir(walkRoot, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				errs = append(errs, err)
				if d != nil && d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if d.IsDir() {
				if p != walkRoot && strings.HasPrefix(d.Name(), ".") {
					return fs.SkipDir
				}
				return nil
			}
			if d.Type().IsRegular() && filepath.Ext(d.Name()) == "."+constants.PolicyFileExtension {
				rel, relErr := filepath.Rel(walkRoot, p)
				if relErr != nil {
					rel = p
				}
				add(filepath.Join(root, rel))
			}
			return nil
		})
		if err != nil && !errors.Is(err, fs.SkipDir) {
			errs = append(errs, err)
		}
	}
	return files, errs
}
