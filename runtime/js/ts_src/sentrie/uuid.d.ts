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
 * UUID module provides functions for generating UUIDs (Universally Unique Identifiers).
 */
declare module "@sentrie/uuid" {
  /**
   * Generates a version 4 UUID (random UUID).
   * Version 4 UUIDs are randomly generated and provide strong uniqueness guarantees.
   * @returns A UUID string in standard format (e.g., "550e8400-e29b-41d4-a716-446655440000")
   */
  export function v4(): string;

  /**
   * Generates a version 6 UUID (time-ordered UUID).
   * Version 6 UUIDs are time-ordered and provide better database indexing performance.
   * @returns A UUID string in standard format (e.g., "1b21dd213814000-8000-6000-0000-000000000000")
   * @throws Error if UUID generation fails
   */
  export function v6(): string;

  /**
   * Generates a version 7 UUID (time-ordered UUID with Unix timestamp).
   * Version 7 UUIDs are time-ordered and include a Unix timestamp for better sorting.
   * @returns A UUID string in standard format (e.g., "017f22e2-79b0-7cc3-8000-383fb6ef7b1a")
   * @throws Error if UUID generation fails
   */
  export function v7(): string;
}
