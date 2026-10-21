// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * String module provides Unicode-aware string helpers backed by Go.
 *
 * Import it from policy code with an explicit alias, because `string` is a
 * keyword and cannot be the default alias:
 *
 *     use { lower, startsWith } from @sentrie/string as str
 *
 * Conventions for every function:
 * - Argument counts and types are checked exactly. A missing, extra or
 *   wrong-typed argument (including `null` and `undefined`) throws; nothing
 *   is coerced to a string.
 * - Errors are thrown as exceptions. No function returns an `Error` object.
 * - Strings are processed as Unicode code points (UTF-8 in Go). A lone UTF-16
 *   surrogate is replaced with U+FFFD.
 */
declare module "@sentrie/string" {
  /**
   * Converts a string to lowercase using simple (one-to-one) Unicode case
   * mapping, so `"ẞ"` becomes `"ß"` and every `"Σ"` becomes `"σ"`.
   * @throws Error if `str` is not a string
   */
  export function lower(str: string): string;

  /**
   * Converts a string to uppercase using simple (one-to-one) Unicode case
   * mapping, so `"ß"` stays `"ß"` rather than becoming `"SS"`.
   * @throws Error if `str` is not a string
   */
  export function upper(str: string): string;

  /**
   * Removes leading and trailing Unicode white space (spaces, tabs, line
   * breaks, U+00A0, U+3000 and the rest of the White_Space property).
   * @throws Error if `str` is not a string
   */
  export function trim(str: string): string;

  /**
   * Removes leading Unicode white space.
   * @throws Error if `str` is not a string
   */
  export function trimStart(str: string): string;

  /**
   * Removes trailing Unicode white space.
   * @throws Error if `str` is not a string
   */
  export function trimEnd(str: string): string;

  /**
   * Reports whether `str` begins with `prefix`. Case-sensitive.
   * @throws Error if either argument is not a string
   */
  export function startsWith(str: string, prefix: string): boolean;

  /**
   * Reports whether `str` ends with `suffix`. Case-sensitive.
   * @throws Error if either argument is not a string
   */
  export function endsWith(str: string, suffix: string): boolean;

  /**
   * Reports whether `substr` occurs within `str`. An empty `substr` is always
   * found. Case-sensitive.
   * @throws Error if either argument is not a string
   */
  export function includes(str: string, substr: string): boolean;

  /**
   * Splits `str` around every occurrence of `sep`. An empty `sep` splits into
   * Unicode code points, so an emoji stays one element. Splitting a string
   * that does not contain `sep` returns a one-element array.
   * @throws Error if either argument is not a string
   */
  export function split(str: string, sep: string): string[];

  /**
   * Joins the elements of `parts` with `sep` between them.
   * @throws Error if `parts` is not an array, any element is not a string, or
   *   `sep` is not a string
   */
  export function join(parts: string[], sep: string): string;

  /**
   * Replaces the first occurrence of `oldStr` with `newStr`. An empty `oldStr`
   * inserts `newStr` at the start.
   * @throws Error if any argument is not a string
   */
  export function replace(str: string, oldStr: string, newStr: string): string;

  /**
   * Replaces every occurrence of `oldStr` with `newStr`. An empty `oldStr`
   * inserts `newStr` before every code point and at the end.
   * @throws Error if any argument is not a string
   */
  export function replaceAll(str: string, oldStr: string, newStr: string): string;

  /**
   * Reports whether two strings are equal under simple Unicode case folding,
   * for case-insensitive comparisons such as `equalFold("Admin", "ADMIN")`.
   * @throws Error if either argument is not a string
   */
  export function equalFold(a: string, b: string): boolean;

  /**
   * Formats a string from a template and arguments. The supported verbs are a
   * closed set; anything else throws:
   *
   * - `%s` - a string argument, as is
   * - `%q` - a string argument, double-quoted with escapes (`"a\"b"`)
   * - `%d` - an integral, finite number, without a fraction or exponent
   * - `%f` - a finite number with 6 decimal places
   * - `%.Nf` - a finite number with N decimal places, 0 to 20
   * - `%v` - a string, number, boolean or `null`, rendered as JavaScript's
   *   `String()` would render it
   * - `%%` - a literal percent sign
   *
   * `%f` and `%.Nf` round the exact binary value to nearest, ties to even
   * (`format("%.0f", 2.5)` is `"2"`). There are no flags or widths.
   *
   * @example format("user %s has %d roles", "alice", 3) // "user alice has 3 roles"
   * @throws Error if the template is not a string, a verb is unsupported, an
   *   argument has the wrong type for its verb, or the number of arguments
   *   differs from the number of verbs that consume one
   */
  export function format(template: string, ...args: (string | number | boolean | null)[]): string;
}
