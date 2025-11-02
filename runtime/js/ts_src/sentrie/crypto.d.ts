/**
 * Copyright 2025 Binaek Sarkar
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
