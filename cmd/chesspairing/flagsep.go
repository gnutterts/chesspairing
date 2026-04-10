// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/flagsep.go
package main

// separateFlags splits args into flag arguments and positional arguments.
// valuedFlags lists flags that consume the next argument as their value
// (e.g., "--profile", "-o"). Flags and positional args can appear in any order.
// A bare "--" terminates flag processing: everything after it is positional.
func separateFlags(args []string, valuedFlags map[string]bool) (flags, positional []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(args[i]) > 0 && args[i][0] == '-' {
			flags = append(flags, args[i])
			if valuedFlags[args[i]] && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, args[i])
		}
	}
	return flags, positional
}
