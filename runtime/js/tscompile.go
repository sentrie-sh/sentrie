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

package js

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

type TranspileResult struct {
	Code string
	Map  string
}

func isTS(module *ModuleSpec) bool {
	ext := strings.ToLower(filepath.Ext(module.Path))
	return ext == ".ts" || ext == ".tsx" || ext == ".mts" || ext == ".cts"
}

func TranspileTS(module *ModuleSpec, source string) (TranspileResult, error) {
	loader := api.LoaderJS
	if isTS(module) {
		loader = api.LoaderTS
	}

	res := api.Transform(source, api.TransformOptions{
		Loader:            loader,
		Target:            api.ES2019,
		Format:            api.FormatCommonJS, // keep CJS semantics for require/module/exports
		Platform:          api.PlatformDefault,
		Sourcemap:         api.SourceMapInline,
		LegalComments:     api.LegalCommentsNone,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
		KeepNames:         false,
		SourcesContent:    api.SourcesContentExclude,
		Charset:           api.CharsetUTF8,
	})

	if len(res.Errors) > 0 {
		return TranspileResult{}, fmt.Errorf("esbuild: %v", res.Errors[0].Text)
	}
	return TranspileResult{Code: string(res.Code), Map: string(res.Map)}, nil
}

// WrapAsIIFE wraps module JS to prevent global pollution AND produce a callable factory.
//
// We compile modules to a function form:
//
//	(function(require, module, exports) { /* transpiled code */ })
//
// The caller will invoke it with a VM-scoped require() and CJS module/exports.
func WrapAsIIFE(js string) string {
	return "(function(require, module, exports) {\n" + js + "\n})"
}
