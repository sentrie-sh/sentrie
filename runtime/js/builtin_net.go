// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package js

import (
	"errors"
	"net"
	"net/netip"

	"github.com/dop251/goja"
)

var BuiltinNetGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("cidrContains", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("cidrContains requires exactly 2 arguments"))
		}
		cidrStr := call.Argument(0).String()
		cidrOrIpStr := call.Argument(1).String()

		_, cidrNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		var ip net.IP
		// Check if second argument is a CIDR or IP
		if _, parsedCIDR, err := net.ParseCIDR(cidrOrIpStr); err == nil {
			ip = parsedCIDR.IP
		} else {
			ip = net.ParseIP(cidrOrIpStr)
			if ip == nil {
				return vm.NewGoError(errors.New("invalid IP or CIDR: " + cidrOrIpStr))
			}
		}

		return vm.ToValue(cidrNet.Contains(ip))
	})

	_ = ex.Set("cidrIntersects", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("cidrIntersects requires exactly 2 arguments"))
		}
		cidr1Str := call.Argument(0).String()
		cidr2Str := call.Argument(1).String()

		_, net1, err := net.ParseCIDR(cidr1Str)
		if err != nil {
			return vm.NewGoError(err)
		}

		_, net2, err := net.ParseCIDR(cidr2Str)
		if err != nil {
			return vm.NewGoError(err)
		}

		// Two CIDRs intersect if either contains the other's network address
		intersects := net1.Contains(net2.IP) || net2.Contains(net1.IP)
		return vm.ToValue(intersects)
	})

	_ = ex.Set("cidrIsValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("cidrIsValid requires exactly 1 argument"))
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
			return vm.NewGoError(errors.New("cidrExpand requires exactly 1 argument"))
		}
		cidrStr := call.Argument(0).String()

		ip, ipNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		var hosts []string
		for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
			hosts = append(hosts, ip.String())
		}

		return vm.ToValue(hosts)
	})

	_ = ex.Set("cidrMerge", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("cidrMerge requires exactly 1 argument"))
		}

		addrsArg := call.Argument(0)
		var addrs []string

		// Handle array or set
		addrsExported := addrsArg.Export()
		switch v := addrsExported.(type) {
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					addrs = append(addrs, str)
				}
			}
		case []string:
			addrs = v
		default:
			return vm.NewGoError(errors.New("cidrMerge requires an array of IPs/CIDRs"))
		}

		// Parse all addresses
		var parsedIPs []netip.Prefix
		for _, addr := range addrs {
			prefix, err := netip.ParsePrefix(addr)
			if err != nil {
				// Try as IP (add /32 or /128)
				ip, err := netip.ParseAddr(addr)
				if err != nil {
					return vm.NewGoError(errors.New("invalid IP or CIDR: " + addr))
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
			return vm.NewGoError(errors.New("parseIP requires exactly 1 argument"))
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
			return vm.NewGoError(errors.New("isIPv4 requires exactly 1 argument"))
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
			return vm.NewGoError(errors.New("isIPv6 requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.To4() == nil && ip.To16() != nil)
	})

	_ = ex.Set("isPrivate", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isPrivate requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.IsPrivate())
	})

	_ = ex.Set("isPublic", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isPublic requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback())
	})

	_ = ex.Set("isLoopback", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isLoopback requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.IsLoopback())
	})

	_ = ex.Set("isMulticast", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isMulticast requires exactly 1 argument"))
		}
		ipStr := call.Argument(0).String()

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(ip.IsMulticast())
	})

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

