// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/tokens"
)

// MaxDataFileSize bounds each `[data]` file. Data is held in memory for the
// executor's lifetime and boxed once, so an unbounded file would be an
// unbounded allocation at pack load.
const MaxDataFileSize int64 = 16 << 20 // 16 MiB

// loadPackData reads, decodes and boxes every `[data]` entry of p, once, into
// p.DataValues. Each path must name a regular file inside the pack root:
// absolute paths, `..` escapes and symlinks that resolve outside the root are
// rejected. Malformed JSON is an error naming the entry and its file.
func loadPackData(ctx context.Context, p *pack.PackFile) error {
	if len(p.Data) == 0 {
		return nil
	}

	// os.Root confines every open to the pack directory, including through
	// symlinks, so a link inside the pack cannot reach a file outside it.
	root, err := os.OpenRoot(p.Location)
	if err != nil {
		return fmt.Errorf("open pack root for data: %w", err)
	}
	defer func() { _ = root.Close() }()

	values := make(map[string]box.Value, len(p.Data))
	// Sorted so that the first error reported is the same on every run.
	for _, name := range slices.Sorted(maps.Keys(p.Data)) {
		if err := ctx.Err(); err != nil {
			return err
		}
		file := p.Data[name]
		if _, isKeyword := tokens.IsKeyword(name); isKeyword {
			return fmt.Errorf("data %q: name is a keyword and cannot be declared with 'data %s'", name, name)
		}
		v, err := readDataFile(root, file)
		if err != nil {
			return fmt.Errorf("data %q (%s): %w", name, file, err)
		}
		values[name] = v
	}
	p.DataValues = values
	return nil
}

// readDataFile reads one data file under root and decodes it as JSON.
func readDataFile(root *os.Root, file string) (box.Value, error) {
	if filepath.IsAbs(file) {
		return box.Undefined(), errors.New("path must be relative to the pack root")
	}
	// IsLocal rejects empty paths, `..` escapes and (on Windows) reserved names.
	rel := filepath.FromSlash(file)
	if !filepath.IsLocal(rel) {
		return box.Undefined(), errors.New("path escapes the pack root")
	}

	f, err := root.Open(rel)
	if err != nil {
		return box.Undefined(), err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return box.Undefined(), err
	}
	if !info.Mode().IsRegular() {
		return box.Undefined(), errors.New("not a regular file")
	}
	if info.Size() > MaxDataFileSize {
		return box.Undefined(), fmt.Errorf("file is %d bytes, larger than the %d byte limit", info.Size(), MaxDataFileSize)
	}
	// The limit is enforced on the bytes read too, in case the file grows
	// between Stat and Read.
	content, err := io.ReadAll(io.LimitReader(f, MaxDataFileSize+1))
	if err != nil {
		return box.Undefined(), err
	}
	if int64(len(content)) > MaxDataFileSize {
		return box.Undefined(), fmt.Errorf("file is larger than the %d byte limit", MaxDataFileSize)
	}

	return decodeDataJSON(content)
}

// decodeDataJSON decodes one JSON document the way request facts are decoded
// (encoding/json into Go values, then box.FromBoundaryAny), so data and facts
// with the same JSON produce the same values. Number handling therefore follows
// any change to fact decoding (#166 tracks exact large integers; when fact
// decoding moves to json.Decoder.UseNumber this decoder must move with it).
// Unlike the fact decoder, trailing content after the document is an error: a
// data file holds exactly one value.
func decodeDataJSON(content []byte) (box.Value, error) {
	dec := json.NewDecoder(bytes.NewReader(content))
	var raw any
	if err := dec.Decode(&raw); err != nil {
		return box.Undefined(), fmt.Errorf("invalid JSON: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return box.Undefined(), errors.New("invalid JSON: unexpected content after the top-level value")
	}
	return box.FromBoundaryAny(raw), nil
}
