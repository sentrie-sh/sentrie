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
 * Net module provides network and IP address utilities for network-based policies.
 * Supports both IPv4 and IPv6 addresses and CIDR notation.
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
   * Expands a CIDR block to a list of all host IP addresses within that block.
   * @param cidr - The CIDR block to expand (e.g., "192.168.1.0/28")
   * @returns Array of all IP addresses in the CIDR block
   * @throws Error if the CIDR notation is invalid
   */
  export function cidrExpand(cidr: string): string[];

  /**
   * Merges a list of IP addresses and subnets into the smallest possible list of CIDR blocks.
   * @param addrs - Array of IP addresses and/or CIDR blocks (e.g., ["192.168.1.1", "192.168.1.2", "10.0.0.0/24"])
   * @returns Array of merged CIDR blocks
   * @throws Error if any address is invalid
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
   * @returns true if the IP is IPv4, false otherwise
   */
  export function isIPv4(ip: string): boolean;

  /**
   * Checks if an IP address is IPv6.
   * @param ip - The IP address to check
   * @returns true if the IP is IPv6, false otherwise
   */
  export function isIPv6(ip: string): boolean;

  /**
   * Checks if an IP address is in a private address range.
   * Private ranges include: 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, and IPv6 equivalents.
   * @param ip - The IP address to check
   * @returns true if the IP is private, false otherwise
   */
  export function isPrivate(ip: string): boolean;

  /**
   * Checks if an IP address is a public (globally routable) address.
   * Public addresses are globally routable unicast addresses that are not private or loopback.
   * @param ip - The IP address to check
   * @returns true if the IP is public, false otherwise
   */
  export function isPublic(ip: string): boolean;

  /**
   * Checks if an IP address is a loopback address.
   * Loopback addresses include 127.0.0.0/8 for IPv4 and ::1 for IPv6.
   * @param ip - The IP address to check
   * @returns true if the IP is a loopback address, false otherwise
   */
  export function isLoopback(ip: string): boolean;

  /**
   * Checks if an IP address is a multicast address.
   * @param ip - The IP address to check
   * @returns true if the IP is multicast, false otherwise
   */
  export function isMulticast(ip: string): boolean;
}
