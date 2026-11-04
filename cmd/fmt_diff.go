// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"strings"
)

// diffContext is the number of unchanged lines shown around each change.
const diffContext = 3

// maxLCSCells caps the LCS table (4 bytes a cell); beyond it diffLines falls
// back to deleting and re-inserting the changed middle. A variable so tests
// can lower it.
var maxLCSCells = 1 << 22

type diffOp byte

const (
	opKeep   diffOp = ' '
	opDelete diffOp = '-'
	opInsert diffOp = '+'
)

type diffEdit struct {
	op   diffOp
	line string // includes its trailing newline, if it has one
}

// unifiedDiff renders the difference between oldText and newText as a unified
// diff for the file name. It returns "" when the texts are equal.
func unifiedDiff(name, oldText, newText string) string {
	if oldText == newText {
		return ""
	}
	edits := diffLines(splitKeepNewline(oldText), splitKeepNewline(newText))

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- %s\n+++ %s\n", name, name)

	// Walk the edit script and cut it into hunks: a hunk covers a run of
	// changes plus diffContext lines either side, and runs closer together
	// than 2*diffContext unchanged lines share one hunk.
	n := len(edits)
	i := 0
	for i < n {
		for i < n && edits[i].op == opKeep {
			i++
		}
		if i >= n {
			break
		}
		start := max(i-diffContext, 0)
		end := i
		// extend end past the changes, merging runs that are close together
		for {
			for end < n && edits[end].op != opKeep {
				end++
			}
			gap := 0
			for end+gap < n && edits[end+gap].op == opKeep {
				gap++
			}
			if end+gap >= n || gap > 2*diffContext {
				end = min(end+diffContext, n)
				break
			}
			end += gap
		}
		writeHunk(&sb, edits, start, end)
		i = end
	}
	return sb.String()
}

func writeHunk(sb *strings.Builder, edits []diffEdit, start, end int) {
	// line numbers (1-based) of the first line of the hunk in each file
	oldLine, newLine := 1, 1
	for _, e := range edits[:start] {
		if e.op != opInsert {
			oldLine++
		}
		if e.op != opDelete {
			newLine++
		}
	}
	oldCount, newCount := 0, 0
	for _, e := range edits[start:end] {
		if e.op != opInsert {
			oldCount++
		}
		if e.op != opDelete {
			newCount++
		}
	}
	// unified diff convention: an empty range reports the line before it
	if oldCount == 0 {
		oldLine--
	}
	if newCount == 0 {
		newLine--
	}
	fmt.Fprintf(sb, "@@ -%s +%s @@\n", hunkRange(oldLine, oldCount), hunkRange(newLine, newCount))
	for _, e := range edits[start:end] {
		sb.WriteByte(byte(e.op))
		sb.WriteString(e.line)
		if !strings.HasSuffix(e.line, "\n") {
			sb.WriteString("\n\\ No newline at end of file\n")
		}
	}
}

func hunkRange(start, count int) string {
	if count == 1 {
		return fmt.Sprint(start)
	}
	return fmt.Sprintf("%d,%d", start, count)
}

// splitKeepNewline splits s into lines, each keeping its trailing "\n". A last
// line without a newline is kept as is, so "a" and "a\n" differ.
func splitKeepNewline(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.SplitAfter(s, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffLines returns an edit script turning a into b, built from a longest
// common subsequence over the lines left after trimming the common prefix and
// suffix.
func diffLines(a, b []string) []diffEdit {
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix &&
		a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	midA := a[prefix : len(a)-suffix]
	midB := b[prefix : len(b)-suffix]

	edits := make([]diffEdit, 0, len(a)+len(b))
	for _, l := range a[:prefix] {
		edits = append(edits, diffEdit{opKeep, l})
	}

	if len(midA)*len(midB) > maxLCSCells {
		// too large for the table: report the middle as replaced wholesale
		for _, l := range midA {
			edits = append(edits, diffEdit{opDelete, l})
		}
		for _, l := range midB {
			edits = append(edits, diffEdit{opInsert, l})
		}
		for _, l := range a[len(a)-suffix:] {
			edits = append(edits, diffEdit{opKeep, l})
		}
		return edits
	}

	// lcs[i][j] is the LCS length of midA[i:] and midB[j:]
	rows, cols := len(midA)+1, len(midB)+1
	lcs := make([]int32, rows*cols)
	for i := len(midA) - 1; i >= 0; i-- {
		for j := len(midB) - 1; j >= 0; j-- {
			if midA[i] == midB[j] {
				lcs[i*cols+j] = lcs[(i+1)*cols+j+1] + 1
			} else {
				lcs[i*cols+j] = max(lcs[(i+1)*cols+j], lcs[i*cols+j+1])
			}
		}
	}
	i, j := 0, 0
	for i < len(midA) && j < len(midB) {
		switch {
		case midA[i] == midB[j]:
			edits = append(edits, diffEdit{opKeep, midA[i]})
			i++
			j++
		case lcs[(i+1)*cols+j] >= lcs[i*cols+j+1]:
			edits = append(edits, diffEdit{opDelete, midA[i]})
			i++
		default:
			edits = append(edits, diffEdit{opInsert, midB[j]})
			j++
		}
	}
	for ; i < len(midA); i++ {
		edits = append(edits, diffEdit{opDelete, midA[i]})
	}
	for ; j < len(midB); j++ {
		edits = append(edits, diffEdit{opInsert, midB[j]})
	}

	for _, l := range a[len(a)-suffix:] {
		edits = append(edits, diffEdit{opKeep, l})
	}
	return edits
}
