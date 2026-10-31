// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
)

func dataTestPack() *pack.PackFile {
	p := pack.NewPackFile("data_pack")
	p.Data = map[string]string{"roles": "roles.json", "allow": "allow.json"}
	p.DataValues = map[string]box.Value{
		"roles": box.FromAny(map[string]any{"admin": []any{"read"}}),
		"allow": box.FromAny([]any{"a"}),
	}
	return p
}

// indexData adds src to a fresh index carrying pk and returns the AddProgram
// and Validate errors.
func (suite *IndexTestSuite) indexData(pk *pack.PackFile, src string) (addErr, validateErr error) {
	idx := CreateIndex()
	if pk != nil {
		suite.Require().NoError(idx.SetPack(suite.ctx, pk))
	}
	prog, err := parser.NewParserFromString(src, "data.sentrie").ParseProgram(suite.ctx)
	suite.Require().NoError(err)
	if err := idx.AddProgram(suite.ctx, prog); err != nil {
		return err, nil
	}
	return nil, idx.Validate(suite.ctx)
}

func (suite *IndexTestSuite) TestPackDataDeclarationIndexed() {
	src := `namespace a
policy p {
  fact user: string
  data roles
  data allow as allowed
  use { now } from @sentrie/time as t
  rule r = count(allowed) > 0 and user in roles
  export decision of r
}
`
	addErr, valErr := suite.indexData(dataTestPack(), src)
	suite.Require().NoError(addErr)
	suite.Require().NoError(valErr)
}

func (suite *IndexTestSuite) TestPackDataDeclarationErrors() {
	cases := []struct {
		name    string
		pack    *pack.PackFile
		body    string
		atAdd   bool
		wantErr string
	}{
		{name: "unknown entry", pack: dataTestPack(), body: "data missing", wantErr: `no such entry in the pack's [data] table`},
		{name: "no pack", pack: nil, body: "data roles", wantErr: "declares no [data] table"},
		{name: "no data table", pack: pack.NewPackFile("x"), body: "data roles", wantErr: "declares no [data] table"},
		{name: "entry not loaded", pack: func() *pack.PackFile {
			p := dataTestPack()
			delete(p.DataValues, "roles")
			return p
		}(), body: "data roles", wantErr: "was not loaded"},
		{name: "clashes with fact", pack: dataTestPack(), body: "fact roles: string\n  data roles", atAdd: true, wantErr: "conflict"},
		{name: "clashes with let", pack: dataTestPack(), body: "data roles\n  let roles = 1", atAdd: true, wantErr: "conflict"},
		{name: "clashes with rule", pack: dataTestPack(), body: "data allow as r", atAdd: true, wantErr: "conflict"},
		{name: "duplicate data alias", pack: dataTestPack(), body: "data roles\n  data allow as roles", atAdd: true, wantErr: "conflict"},
		{name: "clashes with use alias after", pack: dataTestPack(), body: "data roles\n  use { now } from @sentrie/time as roles", atAdd: true, wantErr: "conflict"},
		{name: "clashes with use alias before", pack: dataTestPack(), body: "use { now } from @sentrie/time as roles\n  data roles", atAdd: true, wantErr: "conflict"},
		{name: "after body", pack: dataTestPack(), body: "let x = 1\n  data roles", atAdd: true, wantErr: "'data' must appear before"},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			src := "namespace a\npolicy p {\n  " + tc.body + "\n  rule r = true\n  export decision of r\n}\n"
			addErr, valErr := suite.indexData(tc.pack, src)
			if tc.atAdd {
				suite.Require().Error(addErr)
				suite.Contains(addErr.Error(), tc.wantErr)
				return
			}
			suite.Require().NoError(addErr)
			suite.Require().Error(valErr)
			suite.Contains(valErr.Error(), tc.wantErr)
		})
	}
}

// Pack data is policy-scoped: a derive body sees it only as an argument. The
// runtime guard in evalIdent gives the same message.
func (suite *IndexTestSuite) TestPackDataNotVisibleInsideDerive() {
	src := `namespace a
policy p {
  data roles
  let z = 1
  derive isAdmin = (u: string) => { yield u in roles }
  rule r = isAdmin("x")
  export decision of r
}
`
	addErr, valErr := suite.indexData(dataTestPack(), src)
	suite.Require().NoError(addErr)
	suite.Require().Error(valErr)
	suite.Contains(valErr.Error(), `pack data is not available inside a derive ("roles")`)
}

func (suite *IndexTestSuite) TestPackDataPassedToDeriveAsArgument() {
	src := `namespace a
policy p {
  data roles
  let z = 1
  derive has = (m: dict[list[string]], k: string) => { yield m[k] is defined }
  rule r = has(roles, "admin")
  export decision of r
}
`
	addErr, valErr := suite.indexData(dataTestPack(), src)
	suite.Require().NoError(addErr)
	suite.Require().NoError(valErr)
}

func (suite *IndexTestSuite) TestPackDataBindsLoadedKindStatically() {
	idx := CreateIndex()
	suite.Require().NoError(idx.SetPack(suite.ctx, dataTestPack()))
	prog, err := parser.NewParserFromString(`namespace a
policy p {
  data allow
  rule r = true
  export decision of r
}
`, "kind.sentrie").ParseProgram(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NoError(idx.AddProgram(suite.ctx, prog))
	pol := idx.Namespaces["a"].Policies["p"]
	k := newRuleKindCheckCtx(idx, pol)
	info, ok := k.scope["allow"]
	suite.Require().True(ok)
	suite.True(info.kindKnown)
	suite.Equal(box.ValueList, info.kind)
	suite.True(k.namesValue("allow"))
}
