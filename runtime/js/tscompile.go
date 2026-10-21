// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"path"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// The module factory wrapper. esbuild emits it as a banner and footer, so the
// inline source map it produces already accounts for the extra first line.
// The footer sits on its own line because goja only looks for a
// `//# sourceMappingURL=` comment past trailing lines that are exactly `})`.
const (
	factoryPrefix = "(function(require, module, exports) {"
	factorySuffix = "})"
)

type TranspileResult struct {
	// Code is the module factory expression
	// `(function(require, module, exports) { ... })`, ending in an inline
	// `//# sourceMappingURL=data:...` comment that maps back to the source.
	Code string
}

func isTS(module *ModuleSpec) bool {
	if module.Builtin {
		return true
	}
	ext := strings.ToLower(filepath.Ext(module.Path))
	return ext == ".ts" || ext == ".tsx" || ext == ".mts" || ext == ".cts"
}

// TranspileTS transpiles a module to an ES2019 CommonJS factory expression
// with an inline source map. Every esbuild error is returned, each as a
// *SourceError positioned in the original source.
func TranspileTS(module *ModuleSpec, source string) (TranspileResult, error) {
	loader := api.LoaderJS
	if isTS(module) {
		loader = api.LoaderTS
	}

	name := module.SourceName()
	res := api.Transform(source, api.TransformOptions{
		Loader:            loader,
		Target:            api.ES2019,
		Format:            api.FormatCommonJS, // keep CJS semantics for require/module/exports
		Platform:          api.PlatformDefault,
		Sourcemap:         api.SourceMapInline,
		Sourcefile:        path.Base(name), // goja resolves it against the program name
		Banner:            factoryPrefix,
		Footer:            factorySuffix,
		LegalComments:     api.LegalCommentsNone,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		KeepNames:         false,
		SourcesContent:    api.SourcesContentExclude,
		Charset:           api.CharsetUTF8,
	})

	if len(res.Errors) > 0 {
		errs := make([]error, 0, len(res.Errors))
		for _, msg := range res.Errors {
			errs = append(errs, esbuildError(name, msg))
		}
		return TranspileResult{}, errors.Join(errs...)
	}
	return TranspileResult{Code: string(res.Code)}, nil
}

func esbuildError(name string, msg api.Message) *SourceError {
	se := &SourceError{File: name, Message: msg.Text}
	if msg.Location != nil {
		se.Line = msg.Location.Line
		se.Column = msg.Location.Column + 1 // esbuild columns are 0-based
	}
	return se
}
