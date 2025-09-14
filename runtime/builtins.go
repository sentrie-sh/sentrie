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

package runtime

import (
	"context"
	"fmt"
)

type Builtin func(ctx context.Context, args []any) (any, error)

func BuiltinCount(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("count requires 1 argument")
	}

	asList, ok := args[0].([]any)
	if ok {
		return len(asList), nil
	}

	asString, ok := args[0].(string)
	if ok {
		return len(asString), nil
	}

	asMap, ok := args[0].(map[string]any)
	if ok {
		return len(asMap), nil
	}

	return 0, nil
}

func BuiltinAdd(ctx context.Context, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("add requires 2 arguments")
	}
	return num(args[0]) + num(args[1]), nil
}

var Builtins = map[string]Builtin{
	"count": BuiltinCount,
	"add":   BuiltinAdd,
}
