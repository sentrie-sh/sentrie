// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	fmtMessy = "namespace   com/ex\n\npolicy p {\n  fact x: number\n  rule r = default false { yield x>1 }\n}\n"
	fmtBad   = "namespace com/ex\npolicy {{{\n"
)

// fmtClean returns the formatted form of fmtMessy by formatting it once.
func (s *CmdTestSuite) fmtClean() string {
	dir := s.T().TempDir()
	p := filepath.Join(dir, "a.sentrie")
	s.Require().NoError(os.WriteFile(p, []byte(fmtMessy), 0o600))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{p}}))
	b, err := os.ReadFile(p)
	s.Require().NoError(err)
	s.Require().NotEqual(fmtMessy, string(b))
	return string(b)
}

func (s *CmdTestSuite) write(dir, name, content string) string {
	p := filepath.Join(dir, name)
	s.Require().NoError(os.MkdirAll(filepath.Dir(p), 0o755))
	s.Require().NoError(os.WriteFile(p, []byte(content), 0o600))
	return p
}

func (s *CmdTestSuite) TestFmtRewritesInPlaceKeepingMode() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(os.Chmod(p, 0o640))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{p}}))
	got, _ := os.ReadFile(p)
	s.Equal(s.fmtClean(), string(got))
	st, err := os.Stat(p)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o640), st.Mode().Perm())
}

func (s *CmdTestSuite) TestFmtLeavesFormattedFileUntouched() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", s.fmtClean())
	old := time.Now().Add(-time.Hour).Truncate(time.Second)
	s.Require().NoError(os.Chtimes(p, old, old))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{p}, check: true}))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{p}}))
	st, _ := os.Stat(p)
	s.True(st.ModTime().Equal(old))
}

func (s *CmdTestSuite) TestFmtCheckListsAndFailsWithoutWriting() {
	dir := s.T().TempDir()
	bad := s.write(dir, "bad.sentrie", fmtMessy)
	ok := s.write(dir, "ok.sentrie", s.fmtClean())
	var out strings.Builder
	err := runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{dir}, check: true})
	s.Require().Error(err)
	s.Equal(bad+"\n", out.String())
	got, _ := os.ReadFile(bad)
	s.Equal(fmtMessy, string(got))
	_ = ok
}

func (s *CmdTestSuite) TestFmtDiffPrintsUnifiedDiffAndSucceeds() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	var out strings.Builder
	s.Require().NoError(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{p}, diff: true}))
	d := out.String()
	s.Contains(d, "--- "+p+"\n+++ "+p+"\n@@ ")
	s.Contains(d, "-namespace   com/ex\n")
	s.Contains(d, "+namespace com/ex\n")
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))

	out.Reset()
	s.Require().Error(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{p}, diff: true, check: true}))
	s.Contains(out.String(), "@@ ")
}

func (s *CmdTestSuite) TestFmtRecursesAndSkipsHiddenAndOtherFiles() {
	dir := s.T().TempDir()
	a := s.write(dir, "a.sentrie", fmtMessy)
	b := s.write(dir, "sub/deep/b.sentrie", fmtMessy)
	hidden := s.write(dir, ".git/h.sentrie", fmtMessy)
	other := s.write(dir, "notes.txt", fmtMessy)
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{dir}}))
	for _, p := range []string{a, b} {
		got, _ := os.ReadFile(p)
		s.NotEqual(fmtMessy, string(got), p)
	}
	for _, p := range []string{hidden, other} {
		got, _ := os.ReadFile(p)
		s.Equal(fmtMessy, string(got), p)
	}
}

func (s *CmdTestSuite) TestFmtParseErrorDoesNotStopOtherFiles() {
	dir := s.T().TempDir()
	bad := s.write(dir, "a_bad.sentrie", fmtBad)
	good := s.write(dir, "b_good.sentrie", fmtMessy)
	var errOut strings.Builder
	err := runFmt(s.T().Context(), &strings.Builder{}, &errOut, fmtOptions{paths: []string{dir}})
	s.Require().Error(err)
	s.Contains(errOut.String(), bad)
	got, _ := os.ReadFile(good)
	s.NotEqual(fmtMessy, string(got))
	got, _ = os.ReadFile(bad)
	s.Equal(fmtBad, string(got))
}

func (s *CmdTestSuite) TestFmtMissingPathFailsAndFormatsRest() {
	dir := s.T().TempDir()
	good := s.write(dir, "g.sentrie", fmtMessy)
	var errOut strings.Builder
	err := runFmt(s.T().Context(), &strings.Builder{}, &errOut, fmtOptions{paths: []string{filepath.Join(dir, "nope"), good}})
	s.Require().Error(err)
	s.Contains(errOut.String(), "nope")
	got, _ := os.ReadFile(good)
	s.NotEqual(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtDefaultsToCurrentDirectory() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.T().Chdir(dir)
	s.Require().NoError(Execute(s.T().Context(), Setup(s.T().Context(), "test"), []string{"sentrie", "fmt"}))
	got, _ := os.ReadFile(p)
	s.NotEqual(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtCLIFlagBeforePath() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	err := Execute(s.T().Context(), Setup(s.T().Context(), "test"), []string{"sentrie", "fmt", "--check", dir})
	s.Require().Error(err)
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	s.Require().NoError(Execute(s.T().Context(), Setup(s.T().Context(), "test"), []string{"sentrie", "fmt", dir, "--check=false"}))
	got, _ = os.ReadFile(p)
	s.NotEqual(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestUnifiedDiffHunks() {
	s.Equal("", unifiedDiff("f", "a\n", "a\n"))
	var old, nw []string
	for i := range 30 {
		old = append(old, string(rune('a'+i%26))+strings.Repeat("x", i))
	}
	nw = append(nw, old...)
	nw[2] = "changed"
	nw[25] = "changed2"
	d := unifiedDiff("f", strings.Join(old, "\n")+"\n", strings.Join(nw, "\n")+"\n")
	s.Equal(2, strings.Count(d, "@@ -"))
	s.Contains(d, "-"+old[2]+"\n+changed\n")

	d = unifiedDiff("f", "a\nb", "a\nb\n")
	s.Contains(d, "-b\n\\ No newline at end of file\n+b\n")
}

func (s *CmdTestSuite) TestFmtCheckWithDiffPrintsDiffAndPath() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	var out strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{p}, diff: true, check: true}))
	s.Contains(out.String(), "--- "+p+"\n")
	s.True(strings.HasSuffix(out.String(), p+"\n"))
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtDirectFileOfAnyExtensionIsFormatted() {
	dir := s.T().TempDir()
	p := s.write(dir, "policy.txt", fmtMessy)
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{p}}))
	got, _ := os.ReadFile(p)
	s.NotEqual(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtSkipsSymlinksAndHiddenDirsOnWalk() {
	dir := s.T().TempDir()
	real := s.write(s.T().TempDir(), "real.sentrie", fmtMessy)
	s.Require().NoError(os.Symlink(real, filepath.Join(dir, "link.sentrie")))
	hidden := s.write(dir, ".hid/deep/h.sentrie", fmtMessy)
	var out strings.Builder
	s.Require().NoError(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{dir}, check: true}))
	s.Empty(out.String())
	for _, p := range []string{real, hidden} {
		got, _ := os.ReadFile(p)
		s.Equal(fmtMessy, string(got), p)
	}
}

func (s *CmdTestSuite) TestFmtHiddenRootDirectoryIsWalkedWhenNamed() {
	dir := filepath.Join(s.T().TempDir(), ".root")
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{dir}}))
	got, _ := os.ReadFile(p)
	s.NotEqual(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtFileNamedTwiceIsReportedOnce() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	var out strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{p, dir, filepath.Join(dir, ".", "a.sentrie")}, check: true}))
	s.Equal(p+"\n", out.String())
}

func (s *CmdTestSuite) TestFmtCheckKeepsModeAndParseErrorWithCheck() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(os.Chmod(p, 0o640))
	bad := s.write(dir, "b.sentrie", fmtBad)
	var out, errOut strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &out, &errOut, fmtOptions{paths: []string{dir}, check: true}))
	s.Contains(out.String(), p)
	s.NotContains(out.String(), bad)
	s.Contains(errOut.String(), bad)
	st, _ := os.Stat(p)
	s.Equal(os.FileMode(0o640), st.Mode().Perm())
}

func (s *CmdTestSuite) TestParseFmtArgs() {
	o, err := parseFmtArgs(nil)
	s.Require().NoError(err)
	s.Equal([]string{"."}, o.paths)
	s.False(o.check || o.diff)

	o, err = parseFmtArgs([]string{"--check", "--diff=true", "a", "b"})
	s.Require().NoError(err)
	s.True(o.check && o.diff)
	s.Equal([]string{"a", "b"}, o.paths)

	o, err = parseFmtArgs([]string{"--check=dir", "x"})
	s.Require().NoError(err)
	s.True(o.check)
	s.Equal([]string{"x", "dir"}, o.paths)

	o, err = parseFmtArgs([]string{"--check=false", "x"})
	s.Require().NoError(err)
	s.False(o.check)

	_, err = parseFmtArgs([]string{"--nope"})
	s.Require().Error(err)
}

func (s *CmdTestSuite) TestUnifiedDiffEdges() {
	s.Equal("--- f\n+++ f\n@@ -0,0 +1 @@\n+new\n", unifiedDiff("f", "", "new\n"))
	s.Equal("--- f\n+++ f\n@@ -1 +0,0 @@\n-old\n", unifiedDiff("f", "old\n", ""))
	s.Equal("--- f\n+++ f\n@@ -1,2 +1,3 @@\n+top\n a\n b\n", unifiedDiff("f", "a\nb\n", "top\na\nb\n"))
	s.Equal("--- f\n+++ f\n@@ -1,3 +1,2 @@\n a\n b\n-c\n", unifiedDiff("f", "a\nb\nc\n", "a\nb\n"))
	s.Equal("--- f\n+++ f\n@@ -1,2 +1,3 @@\n a\n b\n+c\n", unifiedDiff("f", "a\nb\n", "a\nb\nc\n"))
	s.Equal("--- f\n+++ f\n@@ -1,2 +1 @@\n-a\n b\n", unifiedDiff("f", "a\nb\n", "b\n"))
}

func (s *CmdTestSuite) TestFmtFlagValueNamingExistingPathIsAPath() {
	for _, name := range []string{"f", "1"} {
		dir := s.T().TempDir()
		p := s.write(dir, name+"/a.sentrie", fmtMessy)
		other := s.write(dir, "o.sentrie", fmtMessy)
		s.T().Chdir(dir)
		opts, err := parseFmtArgs([]string{"--check=" + name})
		s.Require().NoError(err)
		s.True(opts.check)
		s.Equal([]string{name}, opts.paths)
		s.Require().Error(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, opts))
		for _, f := range []string{p, other} {
			got, _ := os.ReadFile(f)
			s.Equal(fmtMessy, string(got))
		}
	}
}

func (s *CmdTestSuite) TestFmtFlagPathsAreSorted() {
	dir := s.T().TempDir()
	s.write(dir, "b/x.sentrie", "")
	s.write(dir, "a/x.sentrie", "")
	s.T().Chdir(dir)
	opts, err := parseFmtArgs([]string{"--diff=b", "pos", "--check=a"})
	s.Require().NoError(err)
	s.Equal([]string{"pos", "a", "b"}, opts.paths)
}

func (s *CmdTestSuite) TestFmtSymlinkRootIsWalkedAndReportedUnderGivenName() {
	dir := s.T().TempDir()
	real := filepath.Join(dir, "real")
	s.write(dir, "real/a.sentrie", fmtMessy)
	lnk := filepath.Join(dir, "lnk")
	s.Require().NoError(os.Symlink(real, lnk))
	var out strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{lnk}, check: true}))
	s.Equal(filepath.Join(lnk, "a.sentrie")+"\n", out.String())
}

func (s *CmdTestSuite) TestFmtParseErrorNamesFileOnce() {
	dir := s.T().TempDir()
	bad := s.write(dir, "bad.sentrie", fmtBad)
	var errOut strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &strings.Builder{}, &errOut, fmtOptions{paths: []string{bad}}))
	s.Contains(errOut.String(), bad)
	s.NotContains(errOut.String(), bad+": parsing error")
}

func (s *CmdTestSuite) TestUnifiedDiffFallsBackAboveCellCap() {
	old := maxLCSCells
	maxLCSCells = 4
	defer func() { maxLCSCells = old }()
	d := unifiedDiff("f", "k\na\nb\nc\nz\n", "k\nx\ny\nw\nz\n")
	s.Contains(d, "-a\n-b\n-c\n+x\n+y\n+w\n")
}

func (s *CmdTestSuite) TestFmtLeavesNoBackupAfterSuccess() {
	dir := s.T().TempDir()
	s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{dir}}))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

func (s *CmdTestSuite) TestFmtSymlinkedFileNamedDirectlyKeepsLink() {
	target := s.write(s.T().TempDir(), "real.sentrie", fmtMessy)
	link := filepath.Join(s.T().TempDir(), "link.sentrie")
	s.Require().NoError(os.Symlink(target, link))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{link}}))
	st, err := os.Lstat(link)
	s.Require().NoError(err)
	s.NotZero(st.Mode() & os.ModeSymlink)
	tgt, _ := os.ReadFile(target)
	s.NotEqual(fmtMessy, string(tgt))
}

func (s *CmdTestSuite) TestFmtHardLinkIsKept() {
	dir := s.T().TempDir()
	a := s.write(dir, "a.sentrie", fmtMessy)
	b := filepath.Join(dir, "b.sentrie")
	s.Require().NoError(os.Link(a, b))
	s.Require().NoError(runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, fmtOptions{paths: []string{a}}))
	ga, _ := os.ReadFile(a)
	gb, _ := os.ReadFile(b)
	s.NotEqual(fmtMessy, string(ga))
	s.Equal(string(ga), string(gb))
	sa, _ := os.Stat(a)
	sb, _ := os.Stat(b)
	s.True(os.SameFile(sa, sb))
}

func (s *CmdTestSuite) TestFmtBackupFailureLeavesFileAndNamesIt() {
	if os.Geteuid() == 0 {
		s.T().Skip("permissions are not enforced for root")
	}
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(os.Chmod(dir, 0o500))
	defer func() { _ = os.Chmod(dir, 0o755) }()
	var errOut strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &strings.Builder{}, &errOut, fmtOptions{paths: []string{p}}))
	s.Contains(errOut.String(), p)
	s.Contains(errOut.String(), "permission denied")
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	s.Require().NoError(os.Chmod(dir, 0o755))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

func (s *CmdTestSuite) TestFmtFailedRestoreKeepsBackupAndNamesIt() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	old := openForRewrite
	defer func() { openForRewrite = old }()
	calls := 0
	openForRewrite = func(path string) (rewriteTarget, error) {
		calls++
		if calls > 1 {
			return nil, errors.New("cannot reopen")
		}
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		return &failingTarget{File: f, failWrite: true}, err
	}
	err := writeFileInPlace(p, []byte(fmtMessy), []byte("replacement text that is longer than nothing"), 0o600)
	s.Require().Error(err)
	s.Contains(err.Error(), p)
	entries, _ := os.ReadDir(dir)
	s.Require().Len(entries, 2)
	for _, e := range entries {
		if e.Name() != "a.sentrie" {
			s.Contains(err.Error(), e.Name())
			bak, _ := os.ReadFile(filepath.Join(dir, e.Name()))
			s.Equal(fmtMessy, string(bak))
		}
	}
}

func (s *CmdTestSuite) TestFmtSymlinkAndTargetCountOnce() {
	dir := s.T().TempDir()
	real := s.write(dir, "real/a.sentrie", fmtMessy)
	lnk := filepath.Join(dir, "lnk")
	s.Require().NoError(os.Symlink(filepath.Dir(real), lnk))
	var out strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{lnk, filepath.Dir(real)}, check: true}))
	s.Equal(filepath.Join(lnk, "a.sentrie")+"\n", out.String())
}

func (s *CmdTestSuite) TestUnifiedDiffFallbackHunkCounts() {
	old := maxLCSCells
	maxLCSCells = 4
	defer func() { maxLCSCells = old }()
	d := unifiedDiff("f", "k\na\nb\nc\nz\n", "k\nx\ny\nw\nz\n")
	s.Equal("--- f\n+++ f\n@@ -1,5 +1,5 @@\n k\n-a\n-b\n-c\n+x\n+y\n+w\n z\n", d)
	d = unifiedDiff("f", "a\nb\nc\nd\n", "p\nq\n")
	s.Equal("--- f\n+++ f\n@@ -1,4 +1,2 @@\n-a\n-b\n-c\n-d\n+p\n+q\n", d)
}

func (s *CmdTestSuite) TestUnifiedDiffFallbackDoesNotKeepCommonMiddleLines() {
	old := maxLCSCells
	maxLCSCells = 4
	defer func() { maxLCSCells = old }()
	d := unifiedDiff("f", "a\nb\nc\n", "x\nb\ny\n")
	s.Equal("--- f\n+++ f\n@@ -1,3 +1,3 @@\n-a\n-b\n-c\n+x\n+b\n+y\n", d)
}

func (s *CmdTestSuite) TestFmtCheckAndDiffCreateNoBackup() {
	dir := s.T().TempDir()
	s.write(dir, "a.sentrie", fmtMessy)
	for _, o := range []fmtOptions{{paths: []string{dir}, check: true}, {paths: []string{dir}, diff: true}} {
		_ = runFmt(s.T().Context(), &strings.Builder{}, &strings.Builder{}, o)
		entries, _ := os.ReadDir(dir)
		s.Len(entries, 1)
	}
}

func (s *CmdTestSuite) TestFmtWalkIgnoresBackupFiles() {
	dir := s.T().TempDir()
	bak := s.write(dir, "a.sentrie.bak-123", fmtMessy)
	var out strings.Builder
	s.Require().NoError(runFmt(s.T().Context(), &out, &strings.Builder{}, fmtOptions{paths: []string{dir}, check: true}))
	s.Empty(out.String())
	got, _ := os.ReadFile(bak)
	s.Equal(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtReadOnlyFileFailsAndLeavesNoBackup() {
	if os.Geteuid() == 0 {
		s.T().Skip("permissions are not enforced for root")
	}
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	s.Require().NoError(os.Chmod(p, 0o400))
	var errOut strings.Builder
	s.Require().Error(runFmt(s.T().Context(), &strings.Builder{}, &errOut, fmtOptions{paths: []string{p}}))
	s.Contains(errOut.String(), p)
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
	s.NotContains(errOut.String(), "restor")
	s.NotContains(errOut.String(), ".bak-")
}

// failingTarget writes part of the data, then fails; the real file behind it
// takes the writes so the restore can be observed.
type failingTarget struct {
	*os.File
	failWrite bool
}

func (f *failingTarget) Write(p []byte) (int, error) {
	if !f.failWrite {
		return f.File.Write(p)
	}
	n, _ := f.File.Write(p[:len(p)/2])
	return n, errors.New("disk full")
}

func (s *CmdTestSuite) TestFmtPartialWriteIsRestoredAndBackupRemoved() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	old := openForRewrite
	defer func() { openForRewrite = old }()
	calls := 0
	openForRewrite = func(path string) (rewriteTarget, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		calls++
		return &failingTarget{File: f, failWrite: calls == 1}, err
	}
	err := writeFileInPlace(p, []byte(fmtMessy), []byte("replacement text that is longer than nothing"), 0o600)
	s.Require().Error(err)
	s.Contains(err.Error(), "original restored")
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

func (s *CmdTestSuite) TestWriteFileInPlaceShorterLongerAndEmpty() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", "0123456789abcdef\n")
	s.Require().NoError(writeFileInPlace(p, []byte("0123456789abcdef\n"), []byte("xyz\n"), 0o600))
	got, _ := os.ReadFile(p)
	s.Equal("xyz\n", string(got))
	s.Require().NoError(writeFileInPlace(p, []byte("xyz\n"), []byte("a much longer replacement\n"), 0o600))
	got, _ = os.ReadFile(p)
	s.Equal("a much longer replacement\n", string(got))
	s.Require().NoError(writeFileInPlace(p, []byte("a much longer replacement\n"), nil, 0o600))
	got, _ = os.ReadFile(p)
	s.Empty(got)
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

func (s *CmdTestSuite) TestFmtWriteFailingWithNothingWrittenRemovesBackup() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	old := openForRewrite
	defer func() { openForRewrite = old }()
	openForRewrite = func(path string) (rewriteTarget, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		return &zeroWriteTarget{File: f}, err
	}
	err := writeFileInPlace(p, []byte(fmtMessy), []byte("new text"), 0o600)
	s.Require().Error(err)
	s.Contains(err.Error(), p)
	s.NotContains(err.Error(), "restor")
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

type zeroWriteTarget struct{ *os.File }

func (z *zeroWriteTarget) Write([]byte) (int, error) { return 0, errors.New("disk full") }

func (s *CmdTestSuite) TestOpenForRewriteDoesNotTruncate() {
	p := s.write(s.T().TempDir(), "a.sentrie", fmtMessy)
	f, err := openForRewrite(p)
	s.Require().NoError(err)
	s.Require().NoError(f.Close())
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
}

func (s *CmdTestSuite) TestFmtPartialWriteLongerThanOriginalIsTruncatedOnRestore() {
	dir := s.T().TempDir()
	orig := "a b\n"
	p := s.write(dir, "a.sentrie", orig)
	old := openForRewrite
	defer func() { openForRewrite = old }()
	calls := 0
	openForRewrite = func(path string) (rewriteTarget, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		calls++
		return &failingTarget{File: f, failWrite: calls == 1}, err
	}
	s.Require().Error(writeFileInPlace(p, []byte(orig), []byte(strings.Repeat("x", 40)), 0o600))
	got, _ := os.ReadFile(p)
	s.Equal(orig, string(got))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}

type closeFailTarget struct{ *os.File }

func (c *closeFailTarget) Close() error {
	_ = c.File.Close()
	return errors.New("close failed")
}

func (s *CmdTestSuite) TestFmtCloseFailureAfterEmptyWriteRestoresOriginal() {
	dir := s.T().TempDir()
	p := s.write(dir, "a.sentrie", fmtMessy)
	old := openForRewrite
	defer func() { openForRewrite = old }()
	calls := 0
	openForRewrite = func(path string) (rewriteTarget, error) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		calls++
		if calls == 1 {
			return &closeFailTarget{f}, err
		}
		return f, err
	}
	err := writeFileInPlace(p, []byte(fmtMessy), nil, 0o600)
	s.Require().Error(err)
	got, _ := os.ReadFile(p)
	s.Equal(fmtMessy, string(got))
	entries, _ := os.ReadDir(dir)
	s.Len(entries, 1)
}
