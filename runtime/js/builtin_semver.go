// SPDX-License-Identifier: Apache-2.0

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
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.com/dop251/goja"
)

var BuiltinSemverGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("compare", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("compare requires exactly 2 arguments"))
		}

		aStr := call.Argument(0).String()
		bStr := call.Argument(1).String()

		a, err := semver.NewVersion(aStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		b, err := semver.NewVersion(bStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		result := a.Compare(b)
		return vm.ToValue(result)
	})

	_ = ex.Set("isValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isValid requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		_, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.ToValue(false)
		}

		return vm.ToValue(true)
	})

	_ = ex.Set("stripPrefix", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("stripPrefix requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		// Strip "v" or "V" prefix if present
		if len(versionStr) > 0 && (versionStr[0] == 'v' || versionStr[0] == 'V') {
			versionStr = versionStr[1:]
		}

		return vm.ToValue(versionStr)
	})

	_ = ex.Set("satisfies", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("satisfies requires exactly 2 arguments"))
		}

		versionStr := call.Argument(0).String()
		constraintStr := call.Argument(1).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		constraint, err := semver.NewConstraint(constraintStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(constraint.Check(version))
	})

	_ = ex.Set("major", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("major requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(int64(version.Major()))
	})

	_ = ex.Set("minor", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("minor requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(int64(version.Minor()))
	})

	_ = ex.Set("patch", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("patch requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(int64(version.Patch()))
	})

	_ = ex.Set("prerelease", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("prerelease requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		prerelease := version.Prerelease()
		if prerelease == "" {
			return goja.Null()
		}

		return vm.ToValue(prerelease)
	})

	_ = ex.Set("metadata", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("metadata requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		metadata := version.Metadata()
		if metadata == "" {
			return goja.Null()
		}

		return vm.ToValue(metadata)
	})

	return ex, nil
}
