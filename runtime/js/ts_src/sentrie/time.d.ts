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
 * Time module provides date and time manipulation utilities.
 * All timestamps are Unix timestamps (seconds since epoch).
 */
declare module "@sentrie/time" {
  /** RFC3339 date format: "2006-01-02T15:04:05Z07:00" */
  export const RFC3339: string;

  /** RFC3339Nano date format: "2006-01-02T15:04:05.999999999Z07:00" */
  export const RFC3339Nano: string;

  /** RFC1123 date format: "Mon, 02 Jan 2006 15:04:05 MST" */
  export const RFC1123: string;

  /** RFC1123Z date format: "Mon, 02 Jan 2006 15:04:05 -0700" */
  export const RFC1123Z: string;

  /** RFC822 date format: "02 Jan 06 15:04 MST" */
  export const RFC822: string;

  /** RFC822Z date format: "02 Jan 06 15:04 -0700" */
  export const RFC822Z: string;

  /**
   * Returns the current timestamp as a Unix timestamp.
   * Within a single execution context, this returns the same value for consistency.
   * @returns Unix timestamp (seconds since epoch) as a number
   */
  export function now(): number;

  /**
   * Parses a date string and returns a Unix timestamp.
   * Supports RFC3339 and RFC3339Nano formats.
   * @param str - The date string to parse (e.g., "2006-01-02T15:04:05Z07:00")
   * @returns Unix timestamp (seconds since epoch) as a number
   * @throws Error if the date string cannot be parsed
   */
  export function parse(str: string): number;

  /**
   * Formats a Unix timestamp as a string using the specified format.
   * Format uses Go's time format reference time: Mon Jan 2 15:04:05 MST 2006.
   * @param timestamp - Unix timestamp (seconds since epoch)
   * @param formatStr - Format string (e.g., "2006-01-02 15:04:05")
   * @returns Formatted date string
   */
  export function format(timestamp: number, formatStr: string): string;

  /**
   * Checks if the first timestamp is before the second timestamp.
   * @param ts1 - First Unix timestamp
   * @param ts2 - Second Unix timestamp
   * @returns true if ts1 < ts2, false otherwise
   */
  export function isBefore(ts1: number, ts2: number): boolean;

  /**
   * Checks if the first timestamp is after the second timestamp.
   * @param ts1 - First Unix timestamp
   * @param ts2 - Second Unix timestamp
   * @returns true if ts1 > ts2, false otherwise
   */
  export function isAfter(ts1: number, ts2: number): boolean;

  /**
   * Checks if a timestamp is between two other timestamps (inclusive).
   * @param ts - The timestamp to check
   * @param start - Start timestamp (inclusive)
   * @param end - End timestamp (inclusive)
   * @returns true if start <= ts <= end, false otherwise
   */
  export function isBetween(ts: number, start: number, end: number): boolean;

  /**
   * Adds a duration to a timestamp.
   * Duration string format: "1h30m" (1 hour 30 minutes), "2d" (2 days), "5s" (5 seconds), etc.
   * @param timestamp - Unix timestamp (seconds since epoch)
   * @param durationStr - Duration string (e.g., "1h", "30m", "2h30m", "1d")
   * @returns New Unix timestamp after adding the duration
   * @throws Error if the duration string format is invalid
   */
  export function addDuration(timestamp: number, durationStr: string): number;

  /**
   * Subtracts a duration from a timestamp.
   * Duration string format: "1h30m" (1 hour 30 minutes), "2d" (2 days), "5s" (5 seconds), etc.
   * @param timestamp - Unix timestamp (seconds since epoch)
   * @param durationStr - Duration string (e.g., "1h", "30m", "2h30m", "1d")
   * @returns New Unix timestamp after subtracting the duration
   * @throws Error if the duration string format is invalid
   */
  export function subtractDuration(timestamp: number, durationStr: string): number;

  /**
   * Converts a Unix timestamp to a Unix timestamp (identity function for API consistency).
   * @param timestamp - Unix timestamp (seconds since epoch)
   * @returns The same Unix timestamp
   */
  export function unix(timestamp: number): number;
}
