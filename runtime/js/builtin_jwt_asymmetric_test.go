// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var jwtTestHashes = map[string]crypto.Hash{
	"256": crypto.SHA256, "384": crypto.SHA384, "512": crypto.SHA512,
}

func jwtDigest(alg, input string) (crypto.Hash, []byte) {
	h := jwtTestHashes[alg[2:]]
	d := h.New()
	d.Write([]byte(input))
	return h, d.Sum(nil)
}

// jwtRS signs claims with RSASSA-PKCS1-v1_5.
func (s *JSTestSuite) jwtRS(alg string, key *rsa.PrivateKey, claims map[string]any) string {
	input := s.jwtUnsigned(map[string]any{"alg": alg, "typ": "JWT"}, claims)
	h, digest := jwtDigest(alg, input)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, h, digest)
	s.Require().NoError(err)
	return input + "." + b64url(sig)
}

// jwtES signs claims with ECDSA, encoding the signature as fixed-width R || S.
func (s *JSTestSuite) jwtES(alg string, key *ecdsa.PrivateKey, claims map[string]any) string {
	input := s.jwtUnsigned(map[string]any{"alg": alg, "typ": "JWT"}, claims)
	_, digest := jwtDigest(alg, input)
	r, sv, err := ecdsa.Sign(rand.Reader, key, digest)
	s.Require().NoError(err)
	size := (key.Curve.Params().BitSize + 7) / 8
	sig := make([]byte, 2*size)
	r.FillBytes(sig[:size])
	sv.FillBytes(sig[size:])
	return input + "." + b64url(sig)
}

func (s *JSTestSuite) pkixPEM(pub any) string {
	der, err := x509.MarshalPKIXPublicKey(pub)
	s.Require().NoError(err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func rsaJWK(pub *rsa.PublicKey) map[string]any {
	return map[string]any{"kty": "RSA", "n": b64url(pub.N.Bytes()), "e": b64url(big.NewInt(int64(pub.E)).Bytes())}
}

func (s *JSTestSuite) ecJWK(pub *ecdsa.PublicKey, crv string) map[string]any {
	point, err := pub.Bytes()
	s.Require().NoError(err)
	size := (len(point) - 1) / 2
	return map[string]any{"kty": "EC", "crv": crv, "x": b64url(point[1 : 1+size]), "y": b64url(point[1+size:])}
}

func jsString(v string) string { return strconv.Quote(v) }

func (s *JSTestSuite) jsJSON(v any) string {
	b, err := json.Marshal(v)
	s.Require().NoError(err)
	return string(b)
}

func (s *JSTestSuite) TestJwtAsymmetricAlgorithms() {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)
	otherRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)
	smallRSA, err := rsa.GenerateKey(rand.Reader, 1024)
	s.Require().NoError(err)
	ec := map[string]*ecdsa.PrivateKey{}
	for alg, curve := range map[string]elliptic.Curve{"ES256": elliptic.P256(), "ES384": elliptic.P384(), "ES512": elliptic.P521()} {
		ec[alg], err = ecdsa.GenerateKey(curve, rand.Reader)
		s.Require().NoError(err)
	}

	rsaPEM := s.pkixPEM(&rsaKey.PublicKey)
	pkcs1PEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&rsaKey.PublicKey)}))
	privPEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)}))
	p256PEM := s.pkixPEM(&ec["ES256"].PublicKey)
	p256JWK := s.ecJWK(&ec["ES256"].PublicKey, "P-256")
	claims := map[string]any{"sub": "alice"}
	now := testExecutionStartUnix

	rs256 := s.jwtRS("RS256", rsaKey, claims)
	es256 := s.jwtES("ES256", ec["ES256"], claims)
	tamper := func(tok string) string {
		parts := strings.Split(tok, ".")
		forged := strings.Split(s.jwtUnsigned(map[string]any{"alg": "x"}, map[string]any{"sub": "mallory"}), ".")
		return parts[0] + "." + forged[1] + "." + parts[2]
	}
	// HS256 signed with the RSA public key's PEM text as the HMAC secret: the
	// classic algorithm-confusion forgery.
	confusedInput := s.jwtUnsigned(map[string]any{"alg": "HS256"}, claims)
	mac := hmac.New(crypto.SHA256.New, []byte(rsaPEM))
	mac.Write([]byte(confusedInput))
	confused := confusedInput + "." + b64url(mac.Sum(nil))

	// The same forgery with the PEM text saved behind a byte-order mark.
	bomPEM := "\uFEFF" + rsaPEM
	bomMAC := hmac.New(crypto.SHA256.New, []byte(bomPEM))
	bomMAC.Write([]byte(confusedInput))
	bomConfused := confusedInput + "." + b64url(bomMAC.Sum(nil))

	// A self-signed certificate around the RSA key.
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), NotBefore: time.Unix(0, 0), NotAfter: time.Unix(4_000_000_000, 0)}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &rsaKey.PublicKey, rsaKey)
	s.Require().NoError(err)
	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}))

	// ES256 signed correctly but encoded as ASN.1 DER instead of R || S.
	esInput := s.jwtUnsigned(map[string]any{"alg": "ES256"}, claims)
	_, esDigest := jwtDigest("ES256", esInput)
	der, err := ecdsa.SignASN1(rand.Reader, ec["ES256"], esDigest)
	s.Require().NoError(err)
	derES256 := esInput + "." + b64url(der)

	// An RS256 token carrying a crit header.
	critInput := s.jwtUnsigned(map[string]any{"alg": "RS256", "crit": []string{"exp"}}, claims)
	critHash, critDigest := jwtDigest("RS256", critInput)
	critSig, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, critHash, critDigest)
	s.Require().NoError(err)
	critRS256 := critInput + "." + b64url(critSig)

	// A modulus over the 8192-bit ceiling; it never has to be a real key.
	hugeN := make([]byte, 1100)
	hugeN[0] = 0xC0
	hugeN[len(hugeN)-1] = 1

	cases := []struct {
		name string
		tok  string
		expr string
		want string
	}{
		{"RS256 PKIX PEM", rs256, `m.verify(token, ` + jsString(rsaPEM) + `, "RS256")`, "true"},
		{"RS384 PKIX PEM", s.jwtRS("RS384", rsaKey, claims), `m.verify(token, ` + jsString(rsaPEM) + `, "RS384")`, "true"},
		{"RS512 PKIX PEM", s.jwtRS("RS512", rsaKey, claims), `m.verify(token, ` + jsString(rsaPEM) + `, "RS512")`, "true"},
		{"RS256 PKCS1 PEM", rs256, `m.verify(token, ` + jsString(pkcs1PEM) + `, "RS256")`, "true"},
		{"RS256 JWK object", rs256, `m.verify(token, ` + s.jsJSON(rsaJWK(&rsaKey.PublicKey)) + `, "RS256")`, "true"},
		{"RS256 JWK string", rs256, `m.verify(token, ` + jsString(s.jsJSON(rsaJWK(&rsaKey.PublicKey))) + `, "RS256")`, "true"},
		{"RS256 decode", rs256, `m.decode(token, ` + jsString(rsaPEM) + `).sub`, `"alice"`},
		{"ES256 PEM", es256, `m.verify(token, ` + jsString(p256PEM) + `, "ES256")`, "true"},
		{"ES384 PEM", s.jwtES("ES384", ec["ES384"], claims), `m.verify(token, ` + jsString(s.pkixPEM(&ec["ES384"].PublicKey)) + `, "ES384")`, "true"},
		{"ES512 PEM", s.jwtES("ES512", ec["ES512"], claims), `m.verify(token, ` + jsString(s.pkixPEM(&ec["ES512"].PublicKey)) + `, "ES512")`, "true"},
		{"ES256 JWK", es256, `m.verify(token, ` + s.jsJSON(p256JWK) + `, "ES256")`, "true"},
		{"ES256 decode", es256, `m.decode(token, ` + s.jsJSON(p256JWK) + `).sub`, `"alice"`},

		{"RS256 tampered payload", tamper(rs256), `m.verify(token, ` + jsString(rsaPEM) + `, "RS256")`, "false"},
		{"ES256 tampered payload", tamper(es256), `m.verify(token, ` + jsString(p256PEM) + `, "ES256")`, "false"},
		{"ES256 tampered decode", tamper(es256), `m.decode(token, ` + jsString(p256PEM) + `)`, "threw:token signature is invalid"},
		{"RS256 wrong key", rs256, `m.verify(token, ` + jsString(s.pkixPEM(&otherRSA.PublicKey)) + `, "RS256")`, "false"},
		{"ES256 truncated signature", es256[:len(es256)-4], `m.verify(token, ` + jsString(p256PEM) + `, "ES256")`, "false"},
		{"RS256 token, ES256 expected", rs256, `m.verify(token, ` + jsString(p256PEM) + `, "ES256")`, "false"},
		{"RS256 expired", s.jwtRS("RS256", rsaKey, map[string]any{"exp": now - 1}), `m.verify(token, ` + jsString(rsaPEM) + `, "RS256")`, "false"},
		{"ES256 expired decode", s.jwtES("ES256", ec["ES256"], map[string]any{"exp": now - 1}), `m.decode(token, ` + jsString(p256PEM) + `)`, "threw:token is expired"},

		{"confusion decode", confused, `m.decode(token, ` + jsString(rsaPEM) + `)`, "threw:token algorithm HS256 is not allowed for this key (allowed: RS256, RS384, RS512)"},
		{"confusion decode with HS allowed", confused, `m.decode(token, ` + jsString(rsaPEM) + `, { algorithms: ["HS256"] })`, "threw:HS256 needs a shared secret, not a public key"},
		{"BOM PEM confusion decode", bomConfused, `m.decode(token, ` + jsString(bomPEM) + `)`, "threw:token algorithm HS256 is not allowed for this key (allowed: RS256, RS384, RS512)"},
		{"BOM PEM confusion verify", bomConfused, `m.verify(token, ` + jsString(bomPEM) + `, "HS256")`, "threw:HS256 needs a shared secret, not a public key"},
		{"BOM PEM genuine RS256", rs256, `m.verify(token, ` + jsString(bomPEM) + `, "RS256")`, "true"},
		{"PEM inside a secret", confused, `m.verify(token, "prefix " + ` + jsString(rsaPEM) + `, "HS256")`, "threw:HS256 needs a shared secret, not a public key"},
		{"confusion verify", confused, `m.verify(token, ` + jsString(rsaPEM) + `, "HS256")`, "threw:HS256 needs a shared secret, not a public key"},
		{"JWK as HMAC secret", confused, `m.verify(token, ` + s.jsJSON(p256JWK) + `, "HS256")`, "threw:secret must be a string"},
		{"secret as RSA key", rs256, `m.decode(token, "a-shared-secret")`, "threw:token algorithm RS256 is not allowed for this key (allowed: HS256, HS384, HS512)"},
		{"secret as RSA key with RS allowed", rs256, `m.decode(token, "a-shared-secret", { algorithms: ["RS256"] })`, "threw:RS256 needs a public key as a PEM block or a JWK"},
		{"decode algorithms pin", rs256, `m.decode(token, ` + jsString(rsaPEM) + `, { algorithms: ["RS384"] })`, "threw:token algorithm RS256 is not allowed for this key (allowed: RS384)"},
		{"decode algorithms allow", rs256, `m.decode(token, ` + jsString(rsaPEM) + `, { algorithms: ["RS512", "RS256"] }).sub`, `"alice"`},
		{"decode algorithms rejects none", rs256, `m.decode(token, ` + jsString(rsaPEM) + `, { algorithms: ["none"] })`, `threw:algorithm "none" is not allowed`},
		{"verify rejects algorithms option", rs256, `m.verify(token, ` + jsString(rsaPEM) + `, "RS256", { algorithms: ["RS256"] })`, "threw:options.algorithms applies to decode"},
		{"ES256 key default pins ES256", s.jwtES("ES384", ec["ES384"], claims), `m.decode(token, ` + jsString(p256PEM) + `)`, "threw:token algorithm ES384 is not allowed for this key (allowed: ES256)"},
		{"certificate PEM", rs256, `m.verify(token, ` + jsString(certPEM) + `, "RS256")`, "true"},
		{"ES256 ASN.1 DER signature", derES256, `m.verify(token, ` + jsString(p256PEM) + `, "ES256")`, "false"},
		{"crit header does not verify", critRS256, `m.verify(token, ` + jsString(rsaPEM) + `, "RS256")`, "false"},
		{"crit header decode throws", critRS256, `m.decode(token, ` + jsString(rsaPEM) + `)`, "threw:unsupported crit header"},
		{"private JWK", rs256, `m.verify(token, ` + s.jsJSON(map[string]any{"kty": "RSA", "n": rsaJWK(&rsaKey.PublicKey)["n"], "e": "AQAB", "d": "AQAB"}) + `, "RS256")`, "threw:invalid JWK: private key material (d) is not accepted"},
		{"huge RSA key", rs256, `m.verify(token, ` + s.jsJSON(map[string]any{"kty": "RSA", "n": b64url(hugeN), "e": "AQAB"}) + `, "RS256")`, "threw:RS256 needs an RSA key of 2048 to 8192 bits"},
		{"RSA key for ES256", es256, `m.verify(token, ` + jsString(rsaPEM) + `, "ES256")`, "threw:ES256 needs an ECDSA P-256 public key"},
		{"P-256 key for ES384", es256, `m.verify(token, ` + jsString(p256PEM) + `, "ES384")`, "threw:ES384 needs an ECDSA P-384 public key"},
		{"EC key for RS256", rs256, `m.verify(token, ` + jsString(p256PEM) + `, "RS256")`, "threw:RS256 needs an RSA public key"},
		{"private key PEM", rs256, `m.verify(token, ` + jsString(privPEM) + `, "RS256")`, `threw:unsupported PEM block "RSA PRIVATE KEY"; pass a public key`},
		{"small RSA key", rs256, `m.verify(token, ` + jsString(s.pkixPEM(&smallRSA.PublicKey)) + `, "RS256")`, "threw:RS256 needs an RSA key of 2048 to 8192 bits"},
		{"JWK alg mismatch", es256, `m.verify(token, ` + s.jsJSON(map[string]any{"kty": "EC", "crv": "P-256", "x": p256JWK["x"], "y": p256JWK["y"], "alg": "ES384"}) + `, "ES256")`, "threw:JWK alg ES384 does not match ES256"},
		{"JWK off-curve point", es256, `m.verify(token, ` + s.jsJSON(map[string]any{"kty": "EC", "crv": "P-256", "x": p256JWK["x"], "y": p256JWK["x"]}) + `, "ES256")`, "threw:invalid JWK"},
		{"JWK unknown kty", es256, `m.verify(token, {"kty": "oct", "k": "c2VjcmV0"}, "ES256")`, "threw:invalid JWK: unsupported kty oct"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			got := s.jwtEval(tc.tok, tc.expr)
			if strings.HasPrefix(tc.want, "threw:") {
				s.True(strings.HasPrefix(got, tc.want), "got %q, want prefix %q", got, tc.want)
				return
			}
			s.Equal(tc.want, got)
		})
	}
}
