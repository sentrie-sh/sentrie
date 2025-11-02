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
 * Regex module provides regular expression pattern matching and manipulation utilities.
 * All patterns are compiled and cached for performance.
 */
declare module "@sentrie/regex" {
  /**
   * Tests if a string matches a regular expression pattern.
   * @param pattern - The regular expression pattern to match
   * @param str - The string to test
   * @returns true if the string matches the pattern, false otherwise
   * @throws Error if the pattern is invalid
   */
  export function match(pattern: string, str: string): boolean;

  /**
   * Finds the first match of a pattern in a string.
   * @param pattern - The regular expression pattern to search for
   * @param str - The string to search in
   * @returns The first match found, or null if no match
   * @throws Error if the pattern is invalid
   */
  export function find(pattern: string, str: string): string | null;

  /**
   * Finds all matches of a pattern in a string.
   * @param pattern - The regular expression pattern to search for
   * @param str - The string to search in
   * @returns Array of all matches found (empty array if none)
   * @throws Error if the pattern is invalid
   */
  export function findAll(pattern: string, str: string): string[];

  /**
   * Replaces the first occurrence of a pattern in a string.
   * @param pattern - The regular expression pattern to match
   * @param str - The string to perform replacement on
   * @param replacement - The replacement string
   * @returns The string with the first match replaced
   * @throws Error if the pattern is invalid
   */
  export function replace(pattern: string, str: string, replacement: string): string;

  /**
   * Replaces all occurrences of a pattern in a string.
   * @param pattern - The regular expression pattern to match
   * @param str - The string to perform replacement on
   * @param replacement - The replacement string
   * @returns The string with all matches replaced
   * @throws Error if the pattern is invalid
   */
  export function replaceAll(pattern: string, str: string, replacement: string): string;

  /**
   * Splits a string by a regular expression pattern.
   * @param pattern - The regular expression pattern to split on
   * @param str - The string to split
   * @returns Array of substrings split by the pattern
   * @throws Error if the pattern is invalid
   */
  export function split(pattern: string, str: string): string[];
}
