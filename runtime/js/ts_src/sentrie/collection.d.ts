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
 * Collection module provides utilities for both list/array and map/object manipulation and operations.
 * Functions are prefixed with `list_` for list operations and `map_` for map operations.
 */
declare module "@sentrie/collection" {
  /**
   * Checks if an array includes a specific item.
   * Uses deep equality comparison.
   * @param arr - The array to search in
   * @param item - The item to search for
   * @returns true if the item is found in the array, false otherwise
   */
  export function list_includes(arr: any[], item: any): boolean;

  /**
   * Finds the first index of an item in an array.
   * Uses deep equality comparison.
   * @param arr - The array to search in
   * @param item - The item to search for
   * @returns The index of the first occurrence, or -1 if not found
   */
  export function list_indexOf(arr: any[], item: any): number;

  /**
   * Finds the last index of an item in an array.
   * Uses deep equality comparison.
   * @param arr - The array to search in
   * @param item - The item to search for
   * @returns The index of the last occurrence, or -1 if not found
   */
  export function list_lastIndexOf(arr: any[], item: any): number;

  /**
   * Sorts an array in ascending order.
   * Sorts numbers numerically and strings lexicographically.
   * @param arr - The array to sort
   * @returns A new sorted array (original array is not modified)
   * @throws Error if the input is not an array
   */
  export function list_sort(arr: any[]): any[];

  /**
   * Removes duplicate values from an array.
   * Uses equality comparison to detect duplicates.
   * @param arr - The array to deduplicate
   * @returns A new array with unique values (original array is not modified)
   * @throws Error if the input is not an array
   */
  export function list_unique(arr: any[]): any[];

  /**
   * Splits an array into chunks of a specified size.
   * @param arr - The array to chunk
   * @param size - The size of each chunk (must be positive)
   * @returns Array of chunks, where each chunk is an array of the specified size (except possibly the last)
   * @throws Error if the input is not an array or size is not positive
   */
  export function list_chunk(arr: any[], size: number): any[][];

  /**
   * Flattens a nested array structure by one level or recursively.
   * @param arr - The nested array to flatten
   * @returns A new flattened array (original array is not modified)
   * @throws Error if the input is not an array
   */
  export function list_flatten(arr: any[]): any[];

  /**
   * Gets all keys from a map/object.
   * @param map - The map/object to extract keys from
   * @returns Array of all keys in the map
   * @throws Error if the input is not a map
   */
  export function map_keys(map: Record<string, any>): any[];

  /**
   * Gets all values from a map/object.
   * @param map - The map/object to extract values from
   * @returns Array of all values in the map
   * @throws Error if the input is not a map
   */
  export function map_values(map: Record<string, any>): any[];

  /**
   * Gets all key-value pairs from a map/object as an array of [key, value] tuples.
   * @param map - The map/object to extract entries from
   * @returns Array of [key, value] pairs
   * @throws Error if the input is not a map
   */
  export function map_entries(map: Record<string, any>): [any, any][];

  /**
   * Checks if a map/object contains a specific key.
   * @param map - The map/object to check
   * @param key - The key to check for
   * @returns true if the key exists in the map, false otherwise
   */
  export function map_has(map: Record<string, any>, key: any): boolean;

  /**
   * Gets a value from a map/object by key, with an optional default value.
   * @param map - The map/object to get the value from
   * @param key - The key to look up
   * @param defaultValue - Optional default value to return if the key is not found
   * @returns The value associated with the key, or the default value if the key is not found (or undefined if no default provided)
   */
  export function map_get(map: Record<string, any>, key: any, defaultValue?: any): any;

  /**
   * Gets the number of key-value pairs in a map/object.
   * @param map - The map/object to get the size of
   * @returns The number of entries in the map
   * @throws Error if the input is not a map
   */
  export function map_size(map: Record<string, any>): number;

  /**
   * Checks if a map/object is empty (has no entries).
   * @param map - The map/object to check
   * @returns true if the map has no entries, false otherwise
   * @throws Error if the input is not a map
   */
  export function map_isEmpty(map: Record<string, any>): boolean;

  /**
   * Merges multiple maps/objects into a single map.
   * Later maps override earlier maps if they have the same keys.
   * @param map1 - The first map/object
   * @param map2 - The second map/object
   * @param ...maps - Additional maps/objects to merge
   * @returns A new merged map (original maps are not modified)
   * @throws Error if any argument is not a map
   */
  export function map_merge(map1: Record<string, any>, map2: Record<string, any>, ...maps: Record<string, any>[]): Record<string, any>;
}
