/**
 * Hash module provides various cryptographic hash functions.
 * All hash functions return hexadecimal-encoded strings.
 */
declare module "@sentrie/hash" {
  /**
   * Computes the MD5 hash of a string.
   * @param str - The string to hash
   * @returns MD5 hash as a hexadecimal string (32 characters)
   * @remarks MD5 is cryptographically broken and should not be used for security purposes.
   *          Use sha256 or sha512 for secure hashing.
   */
  export function md5(str: string): string;

  /**
   * Computes the SHA-1 hash of a string.
   * @param str - The string to hash
   * @returns SHA-1 hash as a hexadecimal string (40 characters)
   * @remarks SHA-1 is cryptographically broken and should not be used for security purposes.
   *          Use sha256 or sha512 for secure hashing.
   */
  export function sha1(str: string): string;

  /**
   * Computes the SHA-256 hash of a string.
   * @param str - The string to hash
   * @returns SHA-256 hash as a hexadecimal string (64 characters)
   */
  export function sha256(str: string): string;

  /**
   * Computes the SHA-512 hash of a string.
   * @param str - The string to hash
   * @returns SHA-512 hash as a hexadecimal string (128 characters)
   */
  export function sha512(str: string): string;

  /**
   * Computes HMAC (Hash-based Message Authentication Code) for data using a secret key.
   * @param algorithm - Hash algorithm to use: "md5", "sha1", "sha256", "sha384", or "sha512"
   * @param data - The data to authenticate
   * @param key - The secret key for HMAC computation
   * @returns HMAC as a hexadecimal string
   * @throws Error if the algorithm is unsupported
   * @remarks Recommended algorithms: "sha256" or "sha512" for security.
   */
  export function hmac(algorithm: string, data: string, key: string): string;
}
