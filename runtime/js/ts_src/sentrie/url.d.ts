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
 * URL module provides URL parsing and manipulation utilities.
 * Note: URL encoding/decoding is provided by the encoding module.
 */
declare module "@sentrie/url" {
  /**
   * Represents a parsed URL with its components.
   */
  export interface ParsedURL {
    /** URL scheme (e.g., "http", "https", "ftp") */
    scheme: string;
    /** URL host (e.g., "example.com:8080") */
    host: string;
    /** URL path (e.g., "/path/to/resource") */
    path: string;
    /** URL query string without the leading "?" (e.g., "key=value&foo=bar") */
    query: string;
    /** URL fragment without the leading "#" (e.g., "section1") */
    fragment: string;
    /** URL user info (e.g., "user:password") */
    user: string;
  }

  /**
   * Parses a URL string into its components.
   * @param urlStr - The URL string to parse
   * @returns A ParsedURL object containing the URL components
   * @throws Error if the URL string is invalid
   */
  export function parse(urlStr: string): ParsedURL;

  /**
   * Joins multiple URL parts into a single URL.
   * Resolves relative paths against the base URL.
   * @param parts - Variable number of URL part strings to join
   * @returns The joined URL string
   * @throws Error if any part is invalid
   */
  export function join(...parts: string[]): string;

  /**
   * Extracts the host component from a URL.
   * @param url - The URL string (can be full URL or just host)
   * @returns The host component of the URL
   * @throws Error if the URL is invalid
   */
  export function getHost(url: string): string;

  /**
   * Extracts the path component from a URL.
   * @param url - The URL string
   * @returns The path component of the URL (e.g., "/path/to/resource")
   * @throws Error if the URL is invalid
   */
  export function getPath(url: string): string;

  /**
   * Extracts the query string component from a URL.
   * @param url - The URL string
   * @returns The query string without the leading "?" (e.g., "key=value")
   * @throws Error if the URL is invalid
   */
  export function getQuery(url: string): string;

  /**
   * Validates whether a string is a valid URL.
   * Basic validation checks for scheme and host presence (or leading "/" for relative URLs).
   * @param urlStr - The URL string to validate
   * @returns true if the URL appears valid, false otherwise
   */
  export function isValid(urlStr: string): boolean;
}
