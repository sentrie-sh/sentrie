// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/dop251/goja"
)

// jwtMinRSABits is the smallest RSA modulus accepted (RFC 7518 section 3.3).
const jwtMinRSABits = 2048

// parseJWTKey turns the key argument into what verifyJWTSignature needs for
// alg: a []byte secret for HMAC, an *rsa.PublicKey for RS*, or an
// *ecdsa.PublicKey on the algorithm's curve for ES*.
//
// The key's form must match the algorithm's family. That is what stops
// algorithm confusion: a token whose header says HS256 cannot be checked
// against a public key used as an HMAC secret, and a shared secret cannot
// stand in for a public key.
func parseJWTKey(alg string, key goja.Value) (any, error) {
	a := jwtAlgorithms[alg]
	if a.family == jwtHMAC {
		return hmacSecret(alg, key)
	}

	pub, err := jwtPublicKey(alg, key)
	if err != nil {
		return nil, err
	}
	switch a.family {
	case jwtRSA:
		rsaKey, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%s needs an RSA public key", alg)
		}
		if rsaKey.N.BitLen() < jwtMinRSABits {
			return nil, fmt.Errorf("%s needs an RSA key of at least %d bits", alg, jwtMinRSABits)
		}
		return rsaKey, nil
	case jwtECDSA:
		ecKey, ok := pub.(*ecdsa.PublicKey)
		if !ok || ecKey.Curve != a.curve {
			return nil, fmt.Errorf("%s needs an ECDSA %s public key", alg, a.curve.Params().Name)
		}
		return ecKey, nil
	}
	return nil, errors.New("unsupported algorithm: " + alg)
}

// hmacSecret validates the key argument for an HMAC algorithm.
func hmacSecret(alg string, key goja.Value) ([]byte, error) {
	secret, ok := key.Export().(string)
	if !ok {
		return nil, errors.New("secret must be a string")
	}
	if secret == "" {
		return nil, errors.New("secret must not be empty")
	}
	if trimmed := strings.TrimSpace(secret); strings.HasPrefix(trimmed, "-----BEGIN") || strings.HasPrefix(trimmed, "{") {
		return nil, fmt.Errorf("%s needs a shared secret, not a public key", alg)
	}
	return []byte(secret), nil
}

// jwtPublicKey parses a PEM block (PKIX "PUBLIC KEY", PKCS #1 "RSA PUBLIC
// KEY" or "CERTIFICATE") or a JWK, given as a JSON string or an object.
func jwtPublicKey(alg string, key goja.Value) (any, error) {
	switch v := key.Export().(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		switch {
		case strings.HasPrefix(trimmed, "-----BEGIN"):
			return pemPublicKey(trimmed)
		case strings.HasPrefix(trimmed, "{"):
			var jwk map[string]any
			if err := json.Unmarshal([]byte(trimmed), &jwk); err != nil {
				return nil, fmt.Errorf("invalid JWK: %w", err)
			}
			return jwkPublicKey(alg, jwk)
		}
	case map[string]any:
		return jwkPublicKey(alg, v)
	}
	return nil, fmt.Errorf("%s needs a public key as a PEM block or a JWK", alg)
}

func pemPublicKey(text string) (any, error) {
	block, _ := pem.Decode([]byte(text))
	if block == nil {
		return nil, errors.New("invalid PEM public key")
	}
	switch block.Type {
	case "PUBLIC KEY":
		return x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		return cert.PublicKey, nil
	}
	return nil, fmt.Errorf("unsupported PEM block %q; pass a public key", block.Type)
}

var jwkCurves = map[string]elliptic.Curve{
	"P-256": elliptic.P256(),
	"P-384": elliptic.P384(),
	"P-521": elliptic.P521(),
}

// jwkPublicKey parses an RSA or EC public JWK (RFC 7517, RFC 7518 section 6).
func jwkPublicKey(alg string, jwk map[string]any) (any, error) {
	if jwkAlg, present := jwk["alg"]; present && jwkAlg != alg {
		return nil, fmt.Errorf("JWK alg %v does not match %s", jwkAlg, alg)
	}
	if use, present := jwk["use"]; present && use != "sig" {
		return nil, fmt.Errorf("JWK use %v is not sig", use)
	}
	switch jwk["kty"] {
	case "RSA":
		n, err := jwkInt(jwk, "n")
		if err != nil {
			return nil, err
		}
		e, err := jwkInt(jwk, "e")
		if err != nil {
			return nil, err
		}
		if !e.IsInt64() || e.Int64() < 3 || e.Int64() > 1<<31-1 {
			return nil, errors.New("invalid JWK: unsupported RSA exponent")
		}
		return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
	case "EC":
		crv, _ := jwk["crv"].(string)
		curve, ok := jwkCurves[crv]
		if !ok {
			return nil, fmt.Errorf("invalid JWK: unsupported curve %q", crv)
		}
		size := (curve.Params().BitSize + 7) / 8
		x, err := jwkBytes(jwk, "x")
		if err != nil {
			return nil, err
		}
		y, err := jwkBytes(jwk, "y")
		if err != nil {
			return nil, err
		}
		if len(x) != size || len(y) != size {
			return nil, errors.New("invalid JWK: coordinate length does not match the curve")
		}
		point := append(append([]byte{4}, x...), y...)
		pub, err := ecdsa.ParseUncompressedPublicKey(curve, point)
		if err != nil {
			return nil, fmt.Errorf("invalid JWK: %w", err)
		}
		return pub, nil
	}
	return nil, fmt.Errorf("invalid JWK: unsupported kty %v", jwk["kty"])
}

func jwkBytes(jwk map[string]any, name string) ([]byte, error) {
	s, ok := jwk[name].(string)
	if !ok || s == "" {
		return nil, fmt.Errorf("invalid JWK: missing %q", name)
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid JWK %q: %w", name, err)
	}
	return b, nil
}

func jwkInt(jwk map[string]any, name string) (*big.Int, error) {
	b, err := jwkBytes(jwk, name)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(b), nil
}
