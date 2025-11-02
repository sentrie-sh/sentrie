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
