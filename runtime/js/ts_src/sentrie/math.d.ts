// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Math module provides number helpers backed by Go.
 *
 *     use { abs, clamp, max } from @sentrie/math
 *
 * Conventions for every function:
 * - Every argument must be a finite number. `NaN`, `Infinity`, `-Infinity`
 *   and any non-number (including numeric strings, `null` and `undefined`)
 *   throw; nothing is coerced.
 * - A result that would be `NaN` or infinite (for example `sqrt(-1)` or an
 *   overflowing `pow`) throws, so every value the module returns is finite
 *   and safe to compare or put in a decision.
 * - A negative-zero result is returned as `0`.
 * - Argument counts are checked exactly, and errors are thrown as
 *   exceptions. No function returns an `Error` object.
 *
 * For IEEE-754 semantics (`NaN` and `Infinity` propagation, `Math.round`
 * rounding) use the `Math` functions re-exported by `@sentrie/js`.
 */
declare module "@sentrie/math" {
  /**
   * Returns the absolute value of `x`.
   * @throws Error if `x` is not a finite number
   */
  export function abs(x: number): number;

  /**
   * Returns the largest integer less than or equal to `x`.
   * @throws Error if `x` is not a finite number
   */
  export function floor(x: number): number;

  /**
   * Returns the smallest integer greater than or equal to `x`.
   * @throws Error if `x` is not a finite number
   */
  export function ceil(x: number): number;

  /**
   * Rounds `x` to the nearest integer, with halves rounded away from zero:
   * `round(2.5)` is `3` and `round(-2.5)` is `-3`. This differs from
   * JavaScript's `Math.round`, which gives `-2` for `-2.5`.
   * @throws Error if `x` is not a finite number
   */
  export function round(x: number): number;

  /**
   * Returns the integer part of `x`, removing any fraction (rounds toward
   * zero).
   * @throws Error if `x` is not a finite number
   */
  export function trunc(x: number): number;

  /**
   * Returns the smallest of the given numbers. Accepts one or more numbers,
   * or a single non-empty array of numbers (such as a list from policy code).
   * @throws Error if there are no values, an array is empty or is not the only
   *   argument, or any value is not a finite number
   */
  export function min(...values: number[]): number;
  export function min(values: number[]): number;

  /**
   * Returns the largest of the given numbers. Accepts one or more numbers, or
   * a single non-empty array of numbers (such as a list from policy code).
   * @throws Error if there are no values, an array is empty or is not the only
   *   argument, or any value is not a finite number
   */
  export function max(...values: number[]): number;
  export function max(values: number[]): number;

  /**
   * Returns `base` raised to the power `exponent`.
   * @throws Error if either argument is not a finite number, or if the result
   *   is not finite (`pow(10, 400)`, `pow(-8, 1/3)`, `pow(0, -1)`)
   */
  export function pow(base: number, exponent: number): number;

  /**
   * Returns the square root of `x`.
   * @throws Error if `x` is negative or not a finite number
   */
  export function sqrt(x: number): number;

  /**
   * Restricts `x` to the inclusive range `[lo, hi]`.
   * @throws Error if any argument is not a finite number, or if `lo > hi`
   */
  export function clamp(x: number, lo: number, hi: number): number;
}
