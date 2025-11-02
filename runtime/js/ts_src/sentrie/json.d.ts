/**
 * JSON module provides JSON marshaling (encoding) and unmarshaling (decoding) utilities.
 * Functions take exactly one argument as specified.
 */
declare module "@sentrie/json" {
  /**
   * Marshals (encodes) a JavaScript value to a JSON string.
   * @param value - The value to marshal (any JavaScript type: object, array, string, number, boolean, null)
   * @returns The JSON string representation of the value
   * @throws Error if the value cannot be marshaled (e.g., circular references)
   */
  export function marshal(value: any): string;

  /**
   * Unmarshals (decodes) a JSON string to a JavaScript value.
   * @param str - The JSON string to unmarshal
   * @returns The decoded value (object, array, string, number, boolean, or null)
   * @throws Error if the JSON string is invalid or cannot be parsed
   */
  export function unmarshal(str: string): any;

  /**
   * Validates whether a string is valid JSON.
   * @param str - The JSON string to validate
   * @returns true if the string is valid JSON, false otherwise
   */
  export function isValid(str: string): boolean;
}
