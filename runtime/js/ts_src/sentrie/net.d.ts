// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Net module provides network and IP address utilities for network-based policies.
 * Supports both IPv4 and IPv6 addresses and CIDR notation.
 *
 * Functions that compute a result throw on invalid input. The predicates (`cidrIsValid` and the `is*` checks)
 * answer false for input they cannot parse, so prefer a positive check such as `isPublic(ip)` over a
 * negated one such as `!isPrivate(ip)`, which is true for garbage.
 */
declare module "@sentrie/net" {
  /**
   * Checks if a CIDR block or IP address is contained within another CIDR block.
   * Supports both IPv4 and IPv6.
   * @param cidr - The CIDR block to check against (e.g., "192.168.1.0/24")
   * @param cidrOrIp - Either a CIDR block or IP address to check (e.g., "192.168.1.5" or "192.168.1.0/28")
   * @returns true if cidrOrIp is contained within cidr, false otherwise
   * @throws Error if either argument is not a valid CIDR or IP address
   */
  export function cidrContains(cidr: string, cidrOrIp: string): boolean;

  /**
   * Checks if two CIDR blocks intersect or overlap.
   * Supports both IPv4 and IPv6.
   * @param cidr1 - First CIDR block (e.g., "192.168.1.0/24")
   * @param cidr2 - Second CIDR block (e.g., "192.168.1.0/28")
   * @returns true if the CIDR blocks intersect, false otherwise
   * @throws Error if either argument is not a valid CIDR block
   */
  export function cidrIntersects(cidr1: string, cidr2: string): boolean;

  /**
   * Validates whether a string is a valid CIDR notation.
   * @param cidr - The CIDR string to validate (e.g., "192.168.1.0/24")
   * @returns true if the CIDR notation is valid, false otherwise
   */
  export function cidrIsValid(cidr: string): boolean;

  /**
   * Expands a CIDR block to a list of every IP address within that block,
   * network and broadcast addresses included.
   * Ranges are limited to 65,536 addresses (a /16 in IPv4, a /112 in IPv6);
   * larger ones throw without being enumerated. Use cidrContains to test
   * membership in a larger range.
   * @param cidr - The CIDR block to expand (e.g., "192.168.1.0/28")
   * @returns Array of all IP addresses in the CIDR block
   * @throws Error if the CIDR notation is invalid or the range holds more than 65,536 addresses
   */
  export function cidrExpand(cidr: string): string[];

  /**
   * Merges a list of IP addresses and subnets into the smallest possible list of CIDR blocks.
   * @param addrs - Array of IP addresses and/or CIDR blocks (e.g., ["192.168.1.1", "192.168.1.2", "10.0.0.0/24"])
   * @returns Array of merged CIDR blocks
   * @throws Error if addrs is not an array of strings or any address is invalid
   */
  export function cidrMerge(addrs: string[] | any[]): string[];

  /**
   * Parses an IP address string (IPv4 or IPv6).
   * @param ipStr - The IP address string to parse
   * @returns The normalized IP address string, or null if invalid
   */
  export function parseIP(ipStr: string): string | null;

  /**
   * Checks if an IP address is IPv4.
   * @param ip - The IP address to check
   * @returns true if the IP is IPv4, false otherwise, including for input that is not an IP address
   */
  export function isIPv4(ip: string): boolean;

  /**
   * Checks if an IP address is IPv6.
   * @param ip - The IP address to check
   * @returns true if the IP is IPv6, false otherwise, including for input that is not an IP address
   */
  export function isIPv6(ip: string): boolean;

  /**
   * Checks if an IP address is in a private address range.
   * Private ranges include: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, and IPv6 equivalents.
   * @param ip - The IP address to check
   * @returns true if the IP is private, false otherwise
   * @throws Error if the input is not an IP address. Classification predicates throw
   * rather than answer false, so a negated check (`!isPrivate(ip)`) fails for junk
   * instead of passing. Check `isIPv4`/`isIPv6` first to handle untrusted input.
   */
  export function isPrivate(ip: string): boolean;

  /**
   * Checks if an IP address is a public (globally routable) address.
   * Public addresses are globally routable unicast addresses that are not private or loopback.
   * @param ip - The IP address to check
   * @returns true if the IP is public, false otherwise
   * @throws Error if the input is not an IP address. Classification predicates throw
   * rather than answer false, so a negated check (`!isPrivate(ip)`) fails for junk
   * instead of passing. Check `isIPv4`/`isIPv6` first to handle untrusted input.
   */
  export function isPublic(ip: string): boolean;

  /**
   * Checks if an IP address is a loopback address.
   * Loopback addresses include 127.0.0.0/8 for IPv4 and ::1 for IPv6.
   * @param ip - The IP address to check
   * @returns true if the IP is a loopback address, false otherwise
   * @throws Error if the input is not an IP address. Classification predicates throw
   * rather than answer false, so a negated check (`!isPrivate(ip)`) fails for junk
   * instead of passing. Check `isIPv4`/`isIPv6` first to handle untrusted input.
   */
  export function isLoopback(ip: string): boolean;

  /**
   * Checks if an IP address is a multicast address.
   * @param ip - The IP address to check
   * @returns true if the IP is multicast, false otherwise
   * @throws Error if the input is not an IP address. Classification predicates throw
   * rather than answer false, so a negated check (`!isPrivate(ip)`) fails for junk
   * instead of passing. Check `isIPv4`/`isIPv6` first to handle untrusted input.
   */
  export function isMulticast(ip: string): boolean;
}
