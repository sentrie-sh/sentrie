/**
 * Crypto module provides basic cryptographic utilities.
 * For additional hash algorithms, see the hash module.
 */
declare module "@sentrie/crypto" {
  /**
   * Computes the SHA-256 hash of a string.
   * This function uses a streaming hash implementation.
   * @param str - The string to hash
   * @returns SHA-256 hash as a hexadecimal string
   * @remarks For consistent hashing, consider using the hash module's sha256 function instead.
   */
  export function sha256(str: string): string;
}
