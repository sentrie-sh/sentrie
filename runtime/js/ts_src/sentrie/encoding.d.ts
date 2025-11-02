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
 * Encoding module provides various encoding and decoding utilities.
 * Supports Base64, Hex, and URL encoding/decoding operations.
 */
declare module "@sentrie/encoding" {
  /**
   * Encodes a string using standard Base64 encoding.
   * @param str - The string to encode
   * @returns Base64-encoded string
   */
  export function base64Encode(str: string): string;

  /**
   * Decodes a Base64-encoded string.
   * @param str - The Base64-encoded string to decode
   * @returns Decoded string
   * @throws Error if the input is not valid Base64
   */
  export function base64Decode(str: string): string;

  /**
   * Encodes a string using URL-safe Base64 encoding.
   * Uses - and _ instead of + and /, and omits padding.
   * @param str - The string to encode
   * @returns URL-safe Base64-encoded string
   */
  export function base64UrlEncode(str: string): string;

  /**
   * Decodes a URL-safe Base64-encoded string.
   * @param str - The URL-safe Base64-encoded string to decode
   * @returns Decoded string
   * @throws Error if the input is not valid URL-safe Base64
   */
  export function base64UrlDecode(str: string): string;

  /**
   * Encodes a string to hexadecimal representation.
   * @param str - The string to encode
   * @returns Hexadecimal-encoded string (e.g., "48656c6c6f")
   */
  export function hexEncode(str: string): string;

  /**
   * Decodes a hexadecimal string.
   * @param str - The hexadecimal string to decode (e.g., "48656c6c6f")
   * @returns Decoded string
   * @throws Error if the input is not valid hexadecimal
   */
  export function hexDecode(str: string): string;

  /**
   * URL-encodes a string using percent encoding (query string encoding).
   * Encodes special characters as %XX hexadecimal sequences.
   * @param str - The string to encode
   * @returns URL-encoded string
   */
  export function urlEncode(str: string): string;

  /**
   * Decodes a URL-encoded string.
   * @param str - The URL-encoded string to decode
   * @returns Decoded string
   * @throws Error if the input contains invalid encoding sequences
   */
  export function urlDecode(str: string): string;
}
