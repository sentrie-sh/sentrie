// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/dop251/goja"
)

// cidrExpandMaxHosts is the largest range cidrExpand enumerates: a /16 in
// IPv4, a /112 in IPv6. CIDRs can come from facts, so an unbounded expansion
// lets a request exhaust memory and CPU. Use cidrContains and friends to test
// membership in larger ranges without enumerating them.
const (
	cidrExpandMaxHostBits = 16
	cidrExpandMaxHosts    = 1 << cidrExpandMaxHostBits // 65,536
)

// classifyIP builds a classification predicate (isPrivate, isPublic,
// isLoopback, isMulticast). Unlike the validity predicates isIPv4 and isIPv6,
// which answer false for junk, these throw on input that is not an IP address:
// policies negate them (`!isPrivate(ip)`), and a false for junk would make the
// negation true and let a malformed address through (#213).
func classifyIP(vm *goja.Runtime, name string, classify func(net.IP) bool) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, fmt.Errorf("%s requires exactly 1 argument", name))
		}
		ipStr := call.Argument(0).String()
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return throwErr(vm, fmt.Errorf("%s: invalid IP address: %q", name, ipStr))
		}
		return vm.ToValue(classify(ip))
	}
}

var BuiltinNetGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("cidrContains", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("cidrContains requires exactly 2 arguments"))
		}
		cidrStr := call.Argument(0).String()
		cidrOrIpStr := call.Argument(1).String()

		_, cidrNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return throwErr(vm, err)
		}

		var ip net.IP
		// Check if second argument is a CIDR or IP
		if _, parsedCIDR, err := net.ParseCIDR(cidrOrIpStr); err == nil {
			ip = parsedCIDR.IP
		} else {
			ip = net.ParseIP(cidrOrIpStr)
			if ip == nil {
				return throwErr(vm, errors.New("invalid IP or CIDR: "+cidrOrIpStr))
			}
		}

		return vm.ToValue(cidrNet.Contains(ip))
	})

	_ = ex.Set("cidrIntersects", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("cidrIntersects requires exactly 2 arguments"))
		}
		cidr1Str := call.Argument(0).String()
		cidr2Str := call.Argument(1).String()

		_, net1, err := net.ParseCIDR(cidr1Str)
		if err != nil {
			return throwErr(vm, err)
		}

		_, net2, err := net.ParseCIDR(cidr2Str)
		if err != nil {
			return throwErr(vm, err)
		}

		// Two CIDRs intersect if either contains the other's network address
		intersects := net1.Contains(net2.IP) || net2.Contains(net1.IP)
		return vm.ToValue(intersects)
	})

	_ = ex.Set("cidrIsValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("cidrIsValid requires exactly 1 argument"))
		}
		cidrStr := call.Argument(0).String()

		_, _, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(true)
	})

	_ = ex.Set("cidrExpand", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("cidrExpand requires exactly 1 argument"))
		}
		cidrStr := call.Argument(0).String()

		ip, ipNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return throwErr(vm, err)
		}

		// Refuse from the prefix length alone, before allocating anything:
		// 10.0.0.0/8 would build 16 million strings and ::/0 never ends (#211).
		ones, bits := ipNet.Mask.Size()
		hostBits := bits - ones
		if hostBits > cidrExpandMaxHostBits {
			return throwErr(vm, fmt.Errorf("cidrExpand: %s has 2^%d addresses, more than the limit of %d", cidrStr, hostBits, cidrExpandMaxHosts))
		}

		hosts := make([]string, 0, 1<<hostBits)
		for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
			hosts = append(hosts, ip.String())
		}

		return vm.ToValue(hosts)
	})

	_ = ex.Set("cidrMerge", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("cidrMerge requires exactly 1 argument"))
		}

		addrsArg := call.Argument(0)
		var addrs []string

		// Handle array or set
		addrsExported := addrsArg.Export()
		switch v := addrsExported.(type) {
		case []interface{}:
			for _, item := range v {
				str, ok := item.(string)
				if !ok {
					return throwErr(vm, errors.New("cidrMerge requires an array of IPs/CIDRs"))
				}
				addrs = append(addrs, str)
			}
		case []string:
			addrs = v
		default:
			return throwErr(vm, errors.New("cidrMerge requires an array of IPs/CIDRs"))
		}

		// Parse all addresses
		var parsedIPs []netip.Prefix
		for _, addr := range addrs {
			prefix, err := netip.ParsePrefix(addr)
			if err != nil {
				// Try as IP (add /32 or /128)
				ip, err := netip.ParseAddr(addr)
				if err != nil {
					return throwErr(vm, errors.New("invalid IP or CIDR: "+addr))
				}
				if ip.Is4() {
					prefix = netip.PrefixFrom(ip, 32)
				} else {
					prefix = netip.PrefixFrom(ip, 128)
				}
			}
			parsedIPs = append(parsedIPs, prefix)
		}

		// Merge using netip prefix operations
		// This is a simplified merge - for full merge logic, we'd need more complex algorithm
		// For now, return unique prefixes
		merged := netipPrefixMerge(parsedIPs)

		result := make([]string, len(merged))
		for i, p := range merged {
			result[i] = p.String()
		}

		return vm.ToValue(result)
	})

	_ = ex.Set("parseIP", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("parseIP requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return goja.Null()
		}
		return vm.ToValue(ip.String())
	})

	_ = ex.Set("isIPv4", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("isIPv4 requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.To4() != nil)
	})

	_ = ex.Set("isIPv6", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("isIPv6 requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.To4() == nil && ip.To16() != nil)
	})

	_ = ex.Set("isPrivate", classifyIP(vm, "isPrivate", func(ip net.IP) bool {
		return ip.IsPrivate()
	}))

	_ = ex.Set("isPublic", classifyIP(vm, "isPublic", func(ip net.IP) bool {
		return ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback()
	}))

	_ = ex.Set("isLoopback", classifyIP(vm, "isLoopback", func(ip net.IP) bool {
		return ip.IsLoopback()
	}))

	_ = ex.Set("isMulticast", classifyIP(vm, "isMulticast", func(ip net.IP) bool {
		return ip.IsMulticast()
	}))

	return ex, nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// netipPrefixMerge merges a slice of netip.Prefix into the smallest possible list
// This is a simplified implementation - a full implementation would merge adjacent/contained prefixes
func netipPrefixMerge(prefixes []netip.Prefix) []netip.Prefix {
	if len(prefixes) == 0 {
		return nil
	}

	// Remove duplicates and sort by prefix length
	seen := make(map[netip.Prefix]bool)
	var unique []netip.Prefix
	for _, p := range prefixes {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}

	// For now, return unique prefixes without full merge logic
	// A full merge would:
	// 1. Remove prefixes that are contained in others
	// 2. Merge adjacent prefixes where possible
	// This is complex and can be enhanced later
	return unique
}
