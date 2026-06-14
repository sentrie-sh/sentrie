// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
)

// kindCheckCtx carries lexical scope and policy context for gradual kind resolution.
type kindCheckCtx struct {
	idx    *Index
	policy *Policy
	derive *Derive
	scope  map[string]bindingInfo
}

type bindingInfo struct {
	kindKnown bool
	kind      box.ValueKind
	typeRef   ast.TypeRef
	lambda    *ast.LambdaExpression
	derive    *Derive
	letInit   ast.Expression
}

func cloneBindingScope(scope map[string]bindingInfo) map[string]bindingInfo {
	out := make(map[string]bindingInfo, len(scope)+1)
	for k, v := range scope {
		out[k] = v
	}
	return out
}

func newRuleKindCheckCtx(idx *Index, policy *Policy) *kindCheckCtx {
	scope := make(map[string]bindingInfo)
	for alias, fact := range policy.Facts {
		scope[alias] = bindingFromFact(idx, policy, fact)
	}
	for name, let := range policy.Lets {
		scope[name] = bindingFromLet(idx, policy, let)
	}
	for name, d := range policy.Derives {
		scope[name] = bindingInfo{kindKnown: true, kind: box.ValueCallable, derive: d}
	}
	if policy.Namespace != nil {
		for name, d := range policy.Namespace.Derives {
			if _, exists := scope[name]; !exists {
				scope[name] = bindingInfo{kindKnown: true, kind: box.ValueCallable, derive: d}
			}
		}
	}
	return &kindCheckCtx{idx: idx, policy: policy, scope: scope}
}

func newDeriveKindCheckCtx(idx *Index, d *Derive) *kindCheckCtx {
	scope := make(map[string]bindingInfo)
	if d.Lambda != nil {
		for i, p := range d.Lambda.Params {
			info := bindingInfo{}
			if d.Lambda.ParamTypes != nil && i < len(d.Lambda.ParamTypes) && d.Lambda.ParamTypes[i] != nil {
				info.typeRef = d.Lambda.ParamTypes[i]
				if k, ok := typeRefKind(idx, d.Policy, d.Lambda.ParamTypes[i]); ok {
					info.kindKnown = true
					info.kind = k
				}
			}
			scope[p] = info
		}
	}
	if d.DefineShort != nil {
		for name, derive := range d.DefineShort {
			scope[name] = bindingInfo{kindKnown: true, kind: box.ValueCallable, derive: derive}
		}
	}
	var policy *Policy
	if d.Policy != nil {
		policy = d.Policy
	}
	return &kindCheckCtx{idx: idx, policy: policy, derive: d, scope: scope}
}

func bindingFromFact(idx *Index, policy *Policy, fact *ast.FactStatement) bindingInfo {
	info := bindingInfo{typeRef: fact.Type}
	if fact.Type != nil {
		if k, ok := typeRefKind(idx, policy, fact.Type); ok {
			info.kindKnown = true
			info.kind = k
		}
	}
	return info
}

func bindingFromLet(idx *Index, policy *Policy, let *ast.VarDeclaration) bindingInfo {
	info := bindingInfo{letInit: let.Value, typeRef: let.Type}
	if let.Type != nil {
		if k, ok := typeRefKind(idx, policy, let.Type); ok {
			info.kindKnown = true
			info.kind = k
		}
	}
	if lam, ok := let.Value.(*ast.LambdaExpression); ok {
		info.lambda = lam
		info.kindKnown = true
		info.kind = box.ValueCallable
	}
	return info
}

func (k *kindCheckCtx) bindLet(vd *ast.VarDeclaration) {
	info := bindingFromLet(k.idx, k.policy, vd)
	if !info.kindKnown && vd.Value != nil {
		if kind, ok := k.resolveKind(vd.Value); ok {
			info.kindKnown = true
			info.kind = kind
		}
	}
	k.scope[vd.Name] = info
}

func (k *kindCheckCtx) pushLambdaScope(lam *ast.LambdaExpression) *kindCheckCtx {
	scope := cloneBindingScope(k.scope)
	for i, p := range lam.Params {
		info := bindingInfo{}
		if lam.ParamTypes != nil && i < len(lam.ParamTypes) && lam.ParamTypes[i] != nil {
			info.typeRef = lam.ParamTypes[i]
			if kind, ok := typeRefKind(k.idx, k.policy, lam.ParamTypes[i]); ok {
				info.kindKnown = true
				info.kind = kind
			}
		}
		scope[p] = info
	}
	child := *k
	child.scope = scope
	return &child
}

func (k *kindCheckCtx) lookupDerive(name string) *Derive {
	if k.derive != nil {
		if d, ok := k.derive.DefineShort[name]; ok {
			return d
		}
		return nil
	}
	if k.policy == nil {
		return nil
	}
	if d, ok := k.policy.Derives[name]; ok {
		return d
	}
	if k.policy.Namespace != nil {
		if d, ok := k.policy.Namespace.Derives[name]; ok {
			return d
		}
	}
	return nil
}

func (k *kindCheckCtx) isShadowed(name string) bool {
	_, ok := k.scope[name]
	return ok
}

func (k *kindCheckCtx) isBuiltinCall(c *ast.CallExpression) (*builtins.Decl, bool) {
	id, ok := c.Callee.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	name := id.Value
	// Precedence: local binding > derive > builtin (matches runtime getTarget).
	if k.isShadowed(name) {
		return nil, false
	}
	if k.lookupDerive(name) != nil {
		return nil, false
	}
	decl, ok := builtins.Table[name]
	return decl, ok
}

// typeRefKind maps a static TypeRef to a runtime ValueKind when certain.
// ShapeTypeRef delegates to lookupShapeKind (alias vs complex). NullableTypeRef is
// unknown in v1 because runtime Precheck passes null through.
// RecordTypeRef maps to ValueList (runtime requires ListValue).
func typeRefKind(idx *Index, policy *Policy, ref ast.TypeRef) (box.ValueKind, bool) {
	if ref == nil {
		return 0, false
	}
	switch tr := ref.(type) {
	case *ast.StringTypeRef:
		return box.ValueString, true
	case *ast.NumberTypeRef:
		return box.ValueNumber, true
	case *ast.TrinaryTypeRef:
		return box.ValueTrinary, true
	case *ast.ListTypeRef:
		return box.ValueList, true
	case *ast.DictTypeRef:
		return box.ValueDict, true
	case *ast.DocumentTypeRef:
		return box.ValueDocument, true
	case *ast.NullableTypeRef:
		return 0, false
	case *ast.RecordTypeRef:
		return box.ValueList, true
	case *ast.ShapeTypeRef:
		return lookupShapeKind(idx, policy, tr)
	default:
		return 0, false
	}
}

// lookupShapeKind resolves a shape FQN to a ValueKind. Alias shapes recurse through
// AliasOf; complex shapes (Model with fields) are ValueDict.
func lookupShapeKind(idx *Index, policy *Policy, tr *ast.ShapeTypeRef) (box.ValueKind, bool) {
	if tr == nil || tr.Ref == nil {
		return 0, false
	}
	shape := resolveShapeReadOnly(idx, policy, tr.Ref)
	if shape == nil {
		return 0, false
	}
	if shape.AliasOf != nil {
		return typeRefKind(idx, policy, shape.AliasOf)
	}
	if shape.Model != nil && (len(shape.Model.Fields) > 0 || shape.Model.WithFQN != nil) {
		return box.ValueDict, true
	}
	return 0, false
}

func resolveShapeReadOnly(idx *Index, policy *Policy, ref *ast.FQN) *Shape {
	if ref == nil || ref.IsEmpty() {
		return nil
	}
	name := ref.LastSegment()

	if policy != nil {
		if s, ok := policy.Shapes[name]; ok {
			return s
		}
		if policy.Namespace != nil {
			if s, ok := policy.Namespace.Shapes[name]; ok {
				return s
			}
		}
	}

	if len(ref.Parts) > 2 {
		nsFQN := ref.Parent()
		if namespace, err := idx.ResolveNamespace(nsFQN.String()); err == nil && namespace != nil {
			if s, ok := namespace.Shapes[name]; ok {
				return s
			}
		}
		namespace, err := idx.ResolveNamespace(nsFQN.String())
		if err != nil {
			return nil
		}
		if err := namespace.VerifyShapeExported(name); err != nil {
			return nil
		}
		s, err := idx.ResolveShape(nsFQN.String(), name)
		if err != nil {
			return nil
		}
		return s
	}
	return nil
}

func lookupShapeFieldTypeReadOnly(idx *Index, policy *Policy, shape *Shape, fieldName string) (ast.TypeRef, bool) {
	if shape == nil {
		return nil, false
	}
	if shape.Model != nil {
		if f, ok := shape.Model.Fields[fieldName]; ok {
			return f.TypeRef, true
		}
		if shape.Model.WithFQN != nil {
			parent := resolveShapeReadOnly(idx, policy, shape.Model.WithFQN)
			if parent != nil {
				return lookupShapeFieldTypeReadOnly(idx, policy, parent, fieldName)
			}
		}
	}
	return nil, false
}

func fieldTypeRefHop(idx *Index, policy *Policy, ref ast.TypeRef, field string) (ast.TypeRef, bool) {
	if ref == nil {
		return nil, false
	}
	switch tr := ref.(type) {
	case *ast.DocumentTypeRef:
		return nil, false
	case *ast.ShapeTypeRef:
		shape := resolveShapeReadOnly(idx, policy, tr.Ref)
		if shape == nil {
			return nil, false
		}
		return lookupShapeFieldTypeReadOnly(idx, policy, shape, field)
	default:
		return nil, false
	}
}

func (k *kindCheckCtx) resolveKind(expr ast.Expression) (box.ValueKind, bool) {
	return k.resolveKindGuarded(expr, nil)
}

func (k *kindCheckCtx) resolveKindGuarded(expr ast.Expression, seen map[string]struct{}) (box.ValueKind, bool) {
	if expr == nil {
		return 0, false
	}

	switch n := expr.(type) {
	case *ast.StringLiteral:
		return box.ValueString, true
	case *ast.IntegerLiteral, *ast.FloatLiteral:
		return box.ValueNumber, true
	case *ast.ListLiteral:
		return box.ValueList, true
	case *ast.MapLiteral:
		return box.ValueDict, true
	case *ast.TrinaryLiteral:
		return box.ValueTrinary, true
	case *ast.NullLiteral:
		return 0, false
	case *ast.LambdaExpression:
		return box.ValueCallable, true
	case *ast.CastExpression:
		return typeRefKind(k.idx, k.policy, n.TargetType)
	case *ast.Identifier:
		return k.resolveIdentKindGuarded(n, seen)
	case *ast.FieldAccessExpression:
		return k.resolveFieldAccessChain(n)
	case *ast.CallExpression:
		if decl, ok := k.isBuiltinCall(n); ok {
			kinds := decl.Sig.Result.Kinds
			if len(kinds) == 1 {
				return kinds[0], true
			}
		}
		return 0, false
	default:
		return 0, false
	}
}

func (k *kindCheckCtx) resolveIdentKind(id *ast.Identifier) (box.ValueKind, bool) {
	return k.resolveIdentKindGuarded(id, nil)
}

func (k *kindCheckCtx) resolveIdentKindGuarded(id *ast.Identifier, seen map[string]struct{}) (box.ValueKind, bool) {
	if seen != nil {
		if _, loop := seen[id.Value]; loop {
			return 0, false
		}
	}
	if info, ok := k.scope[id.Value]; ok {
		if info.kindKnown {
			return info.kind, true
		}
		if info.derive != nil {
			return box.ValueCallable, true
		}
		if info.lambda != nil {
			return box.ValueCallable, true
		}
		if info.typeRef != nil {
			return typeRefKind(k.idx, k.policy, info.typeRef)
		}
		if info.letInit != nil {
			branch := make(map[string]struct{}, len(seen)+1)
			for name := range seen {
				branch[name] = struct{}{}
			}
			branch[id.Value] = struct{}{}
			return k.resolveKindGuarded(info.letInit, branch)
		}
	}
	if d := k.lookupDerive(id.Value); d != nil {
		return box.ValueCallable, true
	}
	return 0, false
}

func (k *kindCheckCtx) resolveFieldAccessChain(e *ast.FieldAccessExpression) (box.ValueKind, bool) {
	fields := make([]string, 0, 4)
	cur := ast.Expression(e)
	for {
		fa, ok := cur.(*ast.FieldAccessExpression)
		if !ok {
			break
		}
		fields = append([]string{fa.Field}, fields...)
		cur = fa.Left
	}

	rootTypeRef, ok := k.resolveFieldChainRootTypeRef(cur)
	if !ok {
		return 0, false
	}

	currentRef := rootTypeRef
	for _, field := range fields {
		nextRef, ok := fieldTypeRefHop(k.idx, k.policy, currentRef, field)
		if !ok {
			return 0, false
		}
		currentRef = nextRef
	}
	return typeRefKind(k.idx, k.policy, currentRef)
}

func (k *kindCheckCtx) resolveFieldChainRootTypeRef(root ast.Expression) (ast.TypeRef, bool) {
	id, ok := root.(*ast.Identifier)
	if !ok {
		return nil, false
	}
	info, ok := k.scope[id.Value]
	if !ok {
		return nil, false
	}
	// v1: only annotated roots (fact/let/param TypeRef). Unannotated lets whose
	// initializer resolves to a shape are intentionally unknown here — safe, no false positive.
	if info.typeRef != nil {
		return info.typeRef, true
	}
	return nil, false
}

func (k *kindCheckCtx) resolveCallableArity(expr ast.Expression) (int, bool) {
	if expr == nil {
		return 0, false
	}
	switch n := expr.(type) {
	case *ast.LambdaExpression:
		return ast.RequiredLambdaArity(n), true
	case *ast.Identifier:
		if info, ok := k.scope[n.Value]; ok {
			if info.lambda != nil {
				return ast.RequiredLambdaArity(info.lambda), true
			}
			if info.derive != nil && info.derive.Lambda != nil {
				return ast.RequiredLambdaArity(info.derive.Lambda), true
			}
		}
		if d := k.lookupDerive(n.Value); d != nil && d.Lambda != nil {
			return ast.RequiredLambdaArity(d.Lambda), true
		}
	}
	return 0, false
}

func kindAllowed(kinds []box.ValueKind, k box.ValueKind) bool {
	return slices.Contains(kinds, k)
}

func builtinParamSigAt(sig builtins.Sig, i int) (builtins.ParamSig, bool) {
	if i < len(sig.Params) {
		return sig.Params[i], true
	}
	if sig.Variadic != nil {
		return *sig.Variadic, true
	}
	return builtins.ParamSig{}, false
}
