// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
			return throwErr(vm, errors.New("compare requires exactly 2 arguments"))
		}

		aStr := call.Argument(0).String()
		bStr := call.Argument(1).String()

		a, err := semver.NewVersion(aStr)
		if err != nil {
			return throwErr(vm, err)
		}

		b, err := semver.NewVersion(bStr)
		if err != nil {
			return throwErr(vm, err)
		}

		result := a.Compare(b)
		return vm.ToValue(result)
	})

	_ = ex.Set("isValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("isValid requires exactly 1 argument"))
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
			return throwErr(vm, errors.New("stripPrefix requires exactly 1 argument"))
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
			return throwErr(vm, errors.New("satisfies requires exactly 2 arguments"))
		}

		versionStr := call.Argument(0).String()
		constraintStr := call.Argument(1).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		constraint, err := semver.NewConstraint(constraintStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(constraint.Check(version))
	})

	_ = ex.Set("major", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("major requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(int64(version.Major()))
	})

	_ = ex.Set("minor", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("minor requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(int64(version.Minor()))
	})

	_ = ex.Set("patch", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("patch requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(int64(version.Patch()))
	})

	_ = ex.Set("prerelease", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("prerelease requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		prerelease := version.Prerelease()
		if prerelease == "" {
			return goja.Null()
		}

		return vm.ToValue(prerelease)
	})

	_ = ex.Set("metadata", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("metadata requires exactly 1 argument"))
		}

		versionStr := call.Argument(0).String()

		version, err := semver.NewVersion(versionStr)
		if err != nil {
			return throwErr(vm, err)
		}

		metadata := version.Metadata()
		if metadata == "" {
			return goja.Null()
		}

		return vm.ToValue(metadata)
	})

	return ex, nil
}
