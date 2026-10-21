// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * JWT module provides JSON Web Token decoding and verification utilities.
 * This module only decodes and verifies tokens; it does NOT create/generate tokens.
 *
 * Keys: HS* algorithms take a shared secret string. RS* and ES* take a public key as a PEM block
 * ("PUBLIC KEY", "RSA PUBLIC KEY" or "CERTIFICATE") or a JWK (object or JSON string). The key's
 * form must match the algorithm's family, which blocks algorithm-confusion forgeries.
 */
declare module "@sentrie/jwt" {
  /** A public JWK (RFC 7517): RSA with n and e, or EC with crv, x and y. */
  export type JWK = Record<string, any>;

  /**
   * Time-claim validation settings for decode() and verify().
   */
  export interface VerifyOptions {
    /** Clock skew tolerated on exp and nbf, in seconds (default: 0). */
    leeway?: number;
    /**
     * Validate exp and nbf (default: true). Set false only to inspect a token
     * whose validity window does not matter.
     */
    checkTime?: boolean;
  }

  /**
   * Decodes a JWT token, verifying its signature when a secret is given.
   * The exp and nbf claims are validated against the evaluation's execution start time
   * (the same Unix-seconds clock as `@sentrie/time.now()`) unless `options.checkTime` is false.
   * @param token - The JWT token string to decode
   * @param key - Optional secret or public key for signature verification; the header's alg picks the algorithm
   * @param options - Optional time-claim settings
   * @returns The decoded payload as an object
   * @throws Error if the token is malformed, the signature is invalid, the algorithm is "none" or unsupported,
   * the token is expired or not yet valid, or the key or options are invalid
   * @remarks Supported algorithms: HS256, HS384, HS512, RS256, RS384, RS512, ES256, ES384, ES512
   */
  export function decode(token: string, key?: string | JWK | null, options?: VerifyOptions): Record<string, any>;

  /**
   * Verifies a JWT token's signature and its exp and nbf claims.
   * The token's alg header must equal `algorithm`. Time claims are checked against the
   * execution start time unless `options.checkTime` is false.
   * @param token - The JWT token string to verify
   * @param key - A non-empty shared secret for HS*, or a PEM or JWK public key for RS* and ES*
   * @param algorithm - Optional algorithm name (default: "HS256"). Supported: HS256, HS384, HS512, RS256, RS384,
   * RS512, ES256, ES384, ES512
   * @param options - Optional time-claim settings
   * @returns true if the token is correctly signed and currently valid; false for a malformed, mismatched,
   * badly signed, expired or not-yet-valid token
   * @throws Error if algorithm is "none" or unsupported, the key does not suit the algorithm (RSA keys need at least
   * 2048 bits), the options are invalid, or on a wrong argument count
   */
  export function verify(token: string, key: string | JWK, algorithm?: string, options?: VerifyOptions): boolean;

  /**
   * Extracts the payload from a JWT token without verification.
   * This does NOT verify the signature - use decode() with a secret for verification.
   * @param token - The JWT token string
   * @returns The decoded payload as an object
   * @throws Error if the token format is invalid
   */
  export function getPayload(token: string): Record<string, any>;

  /**
   * Extracts the header from a JWT token without verification.
   * @param token - The JWT token string
   * @returns The decoded header as an object
   * @throws Error if the token format is invalid
   */
  export function getHeader(token: string): Record<string, any>;
}
