// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * JWT module provides JSON Web Token decoding and verification utilities.
 * This module only decodes and verifies tokens; it does NOT create/generate tokens.
 */
declare module "@sentrie/jwt" {
  /**
   * Time-claim validation settings for decode() and verify().
   */
  export interface VerifyOptions {
    /** Clock skew tolerated on exp and nbf, in seconds (default: 0, at most 86400). */
    leeway?: number;
    /**
     * Validate exp and nbf (default: true). Set false only to inspect a token
     * whose validity window does not matter.
     */
    checkTime?: boolean;
  }

  /**
   * Verifies a JWT token's signature and returns its payload. A key is required: use getPayload()
   * or getHeader() to read a token without verifying it.
   * The exp and nbf claims are validated against the evaluation's execution start time
   * (the same Unix-seconds clock as `@sentrie/time.now()`) unless `options.checkTime` is false.
   * @param token - The JWT token string to decode
   * @param secret - The secret for signature verification; the header's alg picks the algorithm
   * @param options - Optional time-claim settings
   * @returns The decoded payload as an object
   * @throws Error if the token is malformed, the signature is invalid, the algorithm is "none" or unsupported,
   * the token is expired or not yet valid, or the secret is missing (undefined or null) or invalid, or the options are invalid
   * @remarks Supported algorithms: HS256, HS384, HS512
   */
  export function decode(token: string, secret: string, options?: VerifyOptions): Record<string, any>;

  /**
   * Verifies a JWT token's signature and its exp and nbf claims.
   * The token's alg header must equal `algorithm`. Time claims are checked against the
   * execution start time unless `options.checkTime` is false.
   * @param token - The JWT token string to verify
   * @param secret - The secret used for signature verification; must be a non-empty string
   * @param algorithm - Optional algorithm name (default: "HS256"). Supported: HS256, HS384, HS512
   * @param options - Optional time-claim settings
   * @returns true if the token is correctly signed and currently valid; false for a malformed, mismatched,
   * badly signed, expired or not-yet-valid token
   * @throws Error if algorithm is "none" or unsupported, the secret or options are invalid, or on a wrong argument count
   */
  export function verify(token: string, secret: string, algorithm?: string, options?: VerifyOptions): boolean;

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
