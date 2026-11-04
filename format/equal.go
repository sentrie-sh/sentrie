// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

var (
	rangeType = reflect.TypeFor[tokens.Range]()
	cmplxType = reflect.TypeFor[ast.Cmplx]()
)

// shapeOrder lists a shape body's fields and comments in source order.
// Fields live in a map and comments in a slice, so comparing the two alone
// would miss a comment that moved past a field.
func shapeOrder(c *ast.Cmplx) []string {
	type member struct {
		offset int
		label  string
	}
	members := make([]member, 0, len(c.Fields)+len(c.Comments))
	for name, f := range c.Fields {
		members = append(members, member{f.Range.From.Offset, "field " + name})
	}
	for _, cm := range c.Comments {
		members = append(members, member{cm.Span().From.Offset, "comment " + cm.Content})
	}
	slices.SortStableFunc(members, func(x, y member) int { return x.offset - y.offset })
	out := make([]string, len(members))
	for i, m := range members {
		out[i] = m.label
	}
	return out
}

// Equal reports, as an error naming the first difference, whether a and b
// are the same program ignoring source positions and the file name. Comments
// count, and a nil slice equals an empty one.
func Equal(a, b *ast.Program) error {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return nil
		}
		return fmt.Errorf("one program is empty")
	}
	return equal(reflect.ValueOf(a.Statements), reflect.ValueOf(b.Statements), "Statements")
}

func equal(a, b reflect.Value, path string) error {
	if a.Type() != b.Type() {
		return fmt.Errorf("%s: %s differs from %s", path, a.Type(), b.Type())
	}
	switch a.Kind() {
	case reflect.Pointer, reflect.Interface:
		if a.IsNil() || b.IsNil() {
			if a.IsNil() != b.IsNil() {
				return fmt.Errorf("%s: one side is nil", path)
			}
			return nil
		}
		if a.Kind() == reflect.Interface {
			path = fmt.Sprintf("%s.(%s)", path, a.Elem().Type())
		}
		return equal(a.Elem(), b.Elem(), path)
	case reflect.Struct:
		if a.Type() == rangeType {
			return nil
		}
		if a.Type() == cmplxType && a.CanInterface() {
			ca, cb := a.Interface().(ast.Cmplx), b.Interface().(ast.Cmplx)
			if oa, ob := shapeOrder(&ca), shapeOrder(&cb); !slices.Equal(oa, ob) {
				return fmt.Errorf("%s: shape members are in order %q, not %q", path, oa, ob)
			}
		}
		for i := range a.NumField() {
			if err := equal(a.Field(i), b.Field(i), path+"."+a.Type().Field(i).Name); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if a.Len() != b.Len() {
			return fmt.Errorf("%s: length %d differs from %d", path, a.Len(), b.Len())
		}
		for i := range a.Len() {
			if err := equal(a.Index(i), b.Index(i), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	case reflect.Map:
		if a.Len() != b.Len() {
			return fmt.Errorf("%s: %d entries differ from %d", path, a.Len(), b.Len())
		}
		for it := a.MapRange(); it.Next(); {
			bv := b.MapIndex(it.Key())
			if !bv.IsValid() {
				return fmt.Errorf("%s: key %v missing", path, it.Key())
			}
			if err := equal(it.Value(), bv, fmt.Sprintf("%s[%v]", path, it.Key())); err != nil {
				return err
			}
		}
	case reflect.String:
		if a.String() != b.String() {
			return fmt.Errorf("%s: %q differs from %q", path, a.String(), b.String())
		}
	case reflect.Bool:
		if a.Bool() != b.Bool() {
			return fmt.Errorf("%s: %t differs from %t", path, a.Bool(), b.Bool())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if a.Int() != b.Int() {
			return fmt.Errorf("%s: %d differs from %d", path, a.Int(), b.Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if a.Uint() != b.Uint() {
			return fmt.Errorf("%s: %d differs from %d", path, a.Uint(), b.Uint())
		}
	case reflect.Float32, reflect.Float64:
		if a.Float() != b.Float() {
			return fmt.Errorf("%s: %v differs from %v", path, a.Float(), b.Float())
		}
	}
	return nil
}
