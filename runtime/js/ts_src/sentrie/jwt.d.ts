/**
 * JWT module provides JSON Web Token decoding and verification utilities.
 * This module only decodes and verifies tokens; it does NOT create/generate tokens.
 */
declare module "@sentrie/jwt" {
  /**
   * Decodes a JWT token and optionally verifies its signature.
   * @param token - The JWT token string to decode
   * @param secret - Optional secret key for signature verification. If provided, verifies the token signature.
   * @returns The decoded payload as an object
   * @throws Error if the token format is invalid, signature verification fails, or the algorithm is unsupported
   * @remarks Supported algorithms: HS256, HS384, HS512
   */
  export function decode(token: string, secret?: string): Record<string, any>;

  /**
   * Verifies a JWT token's signature.
   * @param token - The JWT token string to verify
   * @param secret - The secret key used for signature verification
   * @param algorithm - Optional algorithm name (default: "HS256"). Supported: HS256, HS384, HS512
   * @returns true if the signature is valid, false otherwise
   */
  export function verify(token: string, secret: string, algorithm?: string): boolean;

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
