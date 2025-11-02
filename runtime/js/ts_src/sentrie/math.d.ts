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
 * Math module provides mathematical constants and functions.
 * Similar to JavaScript's Math object, but with additional functions.
 */
declare module "@sentrie/math" {
  /** Euler's number (e ≈ 2.718281828459045) */
  export const E: number;

  /** Pi (π ≈ 3.141592653589793) */
  export const PI: number;

  /** Natural logarithm of 2 (ln(2) ≈ 0.6931471805599453) */
  export const LN2: number;

  /** Natural logarithm of 10 (ln(10) ≈ 2.302585092994046) */
  export const LN10: number;

  /** Base-2 logarithm of e (log₂(e) ≈ 1.4426950408889634) */
  export const LOG2E: number;

  /** Base-10 logarithm of e (log₁₀(e) ≈ 0.4342944819032518) */
  export const LOG10E: number;

  /** Square root of 2 (√2 ≈ 1.4142135623730951) */
  export const SQRT2: number;

  /** Square root of 0.5 (1/√2 ≈ 0.7071067811865476) */
  export const SQRT1_2: number;

  /** Maximum finite value representable as a float64 (1.7976931348623157e+308) */
  export const MAX_VALUE: number;

  /** Smallest positive non-zero value representable as a float64 (5e-324) */
  export const MIN_VALUE: number;

  /**
   * Returns the absolute value of a number.
   * @param x - The number
   * @returns The absolute value
   */
  export function abs(x: number): number;

  /**
   * Returns the smallest integer greater than or equal to a number (ceiling).
   * @param x - The number
   * @returns The ceiling of the number
   */
  export function ceil(x: number): number;

  /**
   * Returns the largest integer less than or equal to a number (floor).
   * @param x - The number
   * @returns The floor of the number
   */
  export function floor(x: number): number;

  /**
   * Returns the value of a number rounded to the nearest integer.
   * @param x - The number
   * @returns The rounded number
   */
  export function round(x: number): number;

  /**
   * Returns the maximum value from a list of numbers.
   * @param values - Variable number of numbers to compare
   * @returns The maximum value
   * @throws Error if no arguments are provided
   */
  export function max(...values: number[]): number;

  /**
   * Returns the minimum value from a list of numbers.
   * @param values - Variable number of numbers to compare
   * @returns The minimum value
   * @throws Error if no arguments are provided
   */
  export function min(...values: number[]): number;

  /**
   * Returns the square root of a number.
   * @param x - The number (must be non-negative)
   * @returns The square root
   * @throws Error if x is negative
   */
  export function sqrt(x: number): number;

  /**
   * Returns the value of base raised to the power of exponent.
   * @param base - The base number
   * @param exponent - The exponent
   * @returns base raised to the power of exponent
   */
  export function pow(base: number, exponent: number): number;

  /**
   * Returns e raised to the power of x (eˣ).
   * @param x - The exponent
   * @returns eˣ
   */
  export function exp(x: number): number;

  /**
   * Returns the natural logarithm (base e) of a number.
   * @param x - The number (must be positive)
   * @returns The natural logarithm
   * @throws Error if x is non-positive
   */
  export function log(x: number): number;

  /**
   * Returns the base-10 logarithm of a number.
   * @param x - The number (must be positive)
   * @returns The base-10 logarithm
   * @throws Error if x is non-positive
   */
  export function log10(x: number): number;

  /**
   * Returns the base-2 logarithm of a number.
   * @param x - The number (must be positive)
   * @returns The base-2 logarithm
   * @throws Error if x is non-positive
   */
  export function log2(x: number): number;

  /**
   * Returns the sine of an angle in radians.
   * @param x - The angle in radians
   * @returns The sine value (between -1 and 1)
   */
  export function sin(x: number): number;

  /**
   * Returns the cosine of an angle in radians.
   * @param x - The angle in radians
   * @returns The cosine value (between -1 and 1)
   */
  export function cos(x: number): number;

  /**
   * Returns the tangent of an angle in radians.
   * @param x - The angle in radians
   * @returns The tangent value
   */
  export function tan(x: number): number;

  /**
   * Returns the arcsine (inverse sine) of a number in radians.
   * @param x - The number (must be between -1 and 1)
   * @returns The angle in radians (between -π/2 and π/2)
   * @throws Error if x is not between -1 and 1
   */
  export function asin(x: number): number;

  /**
   * Returns the arccosine (inverse cosine) of a number in radians.
   * @param x - The number (must be between -1 and 1)
   * @returns The angle in radians (between 0 and π)
   * @throws Error if x is not between -1 and 1
   */
  export function acos(x: number): number;

  /**
   * Returns the arctangent (inverse tangent) of a number in radians.
   * @param x - The number
   * @returns The angle in radians (between -π/2 and π/2)
   */
  export function atan(x: number): number;

  /**
   * Returns the arctangent of the quotient of its arguments.
   * This is useful for converting rectangular coordinates to polar coordinates.
   * @param y - The y-coordinate
   * @param x - The x-coordinate
   * @returns The angle in radians (between -π and π)
   */
  export function atan2(y: number, x: number): number;

  /**
   * Returns the hyperbolic sine of a number.
   * @param x - The number
   * @returns The hyperbolic sine
   */
  export function sinh(x: number): number;

  /**
   * Returns the hyperbolic cosine of a number.
   * @param x - The number
   * @returns The hyperbolic cosine
   */
  export function cosh(x: number): number;

  /**
   * Returns the hyperbolic tangent of a number.
   * @param x - The number
   * @returns The hyperbolic tangent (between -1 and 1)
   */
  export function tanh(x: number): number;

  /**
   * Returns a pseudo-random number between 0 (inclusive) and 1 (exclusive).
   * Similar to JavaScript's Math.random().
   * @returns A random number in the range [0, 1)
   */
  export function random(): number;
}
