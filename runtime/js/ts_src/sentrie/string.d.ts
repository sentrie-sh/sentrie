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
 * String module provides comprehensive string manipulation utilities.
 */
declare module "@sentrie/string" {
  /**
   * Removes leading and trailing whitespace from a string.
   * @param str - The string to trim
   * @returns The trimmed string
   */
  export function trim(str: string): string;

  /**
   * Removes leading whitespace from a string.
   * @param str - The string to trim
   * @returns The string with leading whitespace removed
   */
  export function trimLeft(str: string): string;

  /**
   * Removes trailing whitespace from a string.
   * @param str - The string to trim
   * @returns The string with trailing whitespace removed
   */
  export function trimRight(str: string): string;

  /**
   * Converts a string to lowercase.
   * @param str - The string to convert
   * @returns The lowercase string
   */
  export function toLowerCase(str: string): string;

  /**
   * Converts a string to uppercase.
   * @param str - The string to convert
   * @returns The uppercase string
   */
  export function toUpperCase(str: string): string;

  /**
   * Replaces occurrences of a substring in a string.
   * @param str - The string to perform replacement on
   * @param oldStr - The substring to replace
   * @param newStr - The replacement substring
   * @param n - Optional number of replacements to make. If negative or omitted, replaces all occurrences.
   * @returns The string with replacements made
   */
  export function replace(str: string, oldStr: string, newStr: string, n?: number): string;

  /**
   * Replaces all occurrences of a substring in a string.
   * @param str - The string to perform replacement on
   * @param oldStr - The substring to replace
   * @param newStr - The replacement substring
   * @returns The string with all occurrences replaced
   */
  export function replaceAll(str: string, oldStr: string, newStr: string): string;

  /**
   * Splits a string into an array of substrings using a separator.
   * @param str - The string to split
   * @param sep - The separator string
   * @returns Array of substrings
   */
  export function split(str: string, sep: string): string[];

  /**
   * Extracts a substring from a string.
   * @param str - The string to extract from
   * @param start - The starting index (inclusive)
   * @param end - Optional ending index (exclusive). If omitted, extracts to the end of the string.
   * @returns The extracted substring
   */
  export function substring(str: string, start: number, end?: number): string;

  /**
   * Extracts a slice of a string.
   * Similar to substring, but supports negative indices.
   * @param str - The string to extract from
   * @param start - The starting index (inclusive)
   * @param end - Optional ending index (exclusive). If omitted, extracts to the end of the string.
   * @returns The extracted substring
   */
  export function slice(str: string, start: number, end?: number): string;

  /**
   * Checks if a string starts with a specific prefix.
   * @param str - The string to check
   * @param prefix - The prefix to check for
   * @returns true if the string starts with the prefix, false otherwise
   */
  export function startsWith(str: string, prefix: string): boolean;

  /**
   * Checks if a string ends with a specific suffix.
   * @param str - The string to check
   * @param suffix - The suffix to check for
   * @returns true if the string ends with the suffix, false otherwise
   */
  export function endsWith(str: string, suffix: string): boolean;

  /**
   * Finds the first index of a substring in a string.
   * @param str - The string to search in
   * @param substr - The substring to search for
   * @param fromIndex - Optional starting index for the search. If omitted, searches from the beginning.
   * @returns The index of the first occurrence, or -1 if not found
   */
  export function indexOf(str: string, substr: string, fromIndex?: number): number;

  /**
   * Finds the last index of a substring in a string.
   * @param str - The string to search in
   * @param substr - The substring to search for
   * @param fromIndex - Optional starting index for the search (searches backwards). If omitted, searches from the end.
   * @returns The index of the last occurrence, or -1 if not found
   */
  export function lastIndexOf(str: string, substr: string, fromIndex?: number): number;

  /**
   * Pads the start of a string to a specified length.
   * @param str - The string to pad
   * @param length - The target length for the padded string
   * @param padStr - Optional padding string (default: space " "). If omitted, uses a space.
   * @returns The padded string
   */
  export function padStart(str: string, length: number, padStr?: string): string;

  /**
   * Pads the end of a string to a specified length.
   * @param str - The string to pad
   * @param length - The target length for the padded string
   * @param padStr - Optional padding string (default: space " "). If omitted, uses a space.
   * @returns The padded string
   */
  export function padEnd(str: string, length: number, padStr?: string): string;

  /**
   * Repeats a string a specified number of times.
   * @param str - The string to repeat
   * @param count - The number of times to repeat (must be non-negative)
   * @returns The repeated string
   * @throws Error if count is negative
   */
  export function repeat(str: string, count: number): string;

  /**
   * Gets the character at a specific index in a string.
   * @param str - The string
   * @param index - The character index
   * @returns The character at the specified index, or empty string if index is out of bounds
   */
  export function charAt(str: string, index: number): string;

  /**
   * Checks if a string includes a specific substring.
   * @param str - The string to search in
   * @param substr - The substring to search for
   * @returns true if the substring is found, false otherwise
   */
  export function includes(str: string, substr: string): boolean;

  /**
   * Gets the length of a string.
   * @param str - The string
   * @returns The length of the string (number of characters)
   */
  export function length(str: string): number;
}
