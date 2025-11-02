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
 * Semver module provides semantic version comparison and validation utilities.
 * Supports the "v" prefix (e.g., "v1.2.3" is equivalent to "1.2.3").
 */
declare module "@sentrie/semver" {
  /**
   * Compares two semantic version strings.
   * Supports the "v" prefix - "v1.2.3" and "1.2.3" are treated as equivalent.
   * @param a - The first version string (e.g., "1.2.3" or "v1.2.3")
   * @param b - The second version string (e.g., "1.2.4" or "v1.2.4")
   * @returns -1 if a < b, 1 if a > b, 0 if a == b
   * @throws Error if either version string is invalid
   */
  export function compare(a: string, b: string): number;

  /**
   * Validates whether a string is a valid semantic version.
   * Supports the "v" prefix - "v1.2.3" is considered valid.
   * @param a - The version string to validate (e.g., "1.2.3", "v1.2.3", "1.0.0-alpha")
   * @returns true if the string is a valid semantic version, false otherwise
   */
  export function isValid(a: string): boolean;

  /**
   * Strips the "v" or "V" prefix from a version string if present.
   * @param a - The version string (e.g., "v1.2.3" or "V1.2.3")
   * @returns The version string without the prefix (e.g., "1.2.3")
   * @example
   * stripPrefix("v1.2.3") // returns "1.2.3"
   * stripPrefix("1.2.3")  // returns "1.2.3"
   */
  export function stripPrefix(a: string): string;

  /**
   * Checks if a version satisfies a constraint.
   * Supports constraint ranges like ">=1.0.0 <2.0.0", "^1.2.0", "~1.2.0", etc.
   * @param version - The version string to check (e.g., "1.2.3" or "v1.2.3")
   * @param constraint - The constraint string (e.g., ">=1.0.0 <2.0.0", "^1.2.0")
   * @returns true if the version satisfies the constraint, false otherwise
   * @throws Error if either version or constraint string is invalid
   * @example
   * satisfies("1.2.3", ">=1.0.0 <2.0.0") // returns true
   * satisfies("2.0.0", "^1.2.0") // returns false
   */
  export function satisfies(version: string, constraint: string): boolean;

  /**
   * Gets the major version number from a version string.
   * @param version - The version string (e.g., "1.2.3" or "v1.2.3")
   * @returns The major version number (e.g., 1 for "1.2.3")
   * @throws Error if the version string is invalid
   * @example
   * major("1.2.3") // returns 1
   * major("2.0.0-alpha") // returns 2
   */
  export function major(version: string): number;

  /**
   * Gets the minor version number from a version string.
   * @param version - The version string (e.g., "1.2.3" or "v1.2.3")
   * @returns The minor version number (e.g., 2 for "1.2.3")
   * @throws Error if the version string is invalid
   * @example
   * minor("1.2.3") // returns 2
   * minor("1.0.0") // returns 0
   */
  export function minor(version: string): number;

  /**
   * Gets the patch version number from a version string.
   * @param version - The version string (e.g., "1.2.3" or "v1.2.3")
   * @returns The patch version number (e.g., 3 for "1.2.3")
   * @throws Error if the version string is invalid
   * @example
   * patch("1.2.3") // returns 3
   * patch("1.2.0") // returns 0
   */
  export function patch(version: string): number;

  /**
   * Gets the prerelease identifier from a version string.
   * @param version - The version string (e.g., "1.2.3-alpha.1" or "1.0.0-beta")
   * @returns The prerelease identifier (e.g., "alpha.1") or null if not present
   * @throws Error if the version string is invalid
   * @example
   * prerelease("1.2.3-alpha.1") // returns "alpha.1"
   * prerelease("1.2.3") // returns null
   */
  export function prerelease(version: string): string | null;

  /**
   * Gets the build metadata from a version string.
   * @param version - The version string (e.g., "1.2.3+001" or "1.0.0+exp.sha.5114f85")
   * @returns The build metadata (e.g., "001") or null if not present
   * @throws Error if the version string is invalid
   * @example
   * metadata("1.2.3+001") // returns "001"
   * metadata("1.2.3") // returns null
   */
  export function metadata(version: string): string | null;
}

