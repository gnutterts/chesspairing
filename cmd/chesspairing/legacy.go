// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/legacy.go
package main

import (
	"fmt"
	"io"

	cp "github.com/gnutterts/chesspairing"
)

// parsedLegacyArgs holds the parsed state from legacy CLI arguments.
type parsedLegacyArgs struct {
	system      cp.PairingSystem
	inputFile   string
	mode        string // "pair", "check", "generate"
	outputFile  string // for -p file or -g -o file
	seed        string // for -g -s
	configFile  string // for -g config
	showVersion bool   // -r
	wide        bool   // -w: wide output format
}

// parseLegacyArgs parses bbpPairings-compatible positional arguments.
// The args slice should NOT include the program name (os.Args[0]).
func parseLegacyArgs(args []string) (*parsedLegacyArgs, error) {
	p := &parsedLegacyArgs{}
	i := 0

	for i < len(args) {
		arg := args[i]
		switch arg {
		case "-r":
			p.showVersion = true
			i++

		case "-w":
			p.wide = true
			i++

		case "-q":
			// JaVaFo compat: accepted, ignored. May have optional numeric argument.
			i++
			if i < len(args) && len(args[i]) > 0 && args[i][0] >= '0' && args[i][0] <= '9' {
				i++ // skip the numeric value
			}

		case "-p":
			p.mode = "pair"
			i++
			// Optional output file follows
			if i < len(args) && !isFlag(args[i]) {
				p.outputFile = args[i]
				i++
			}

		case "-c":
			p.mode = "check"
			i++

		case "-g":
			p.mode = "generate"
			i++
			// Optional config file follows
			if i < len(args) && !isFlag(args[i]) {
				p.configFile = args[i]
				i++
			}

		case "-o":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("-o requires a filename argument")
			}
			p.outputFile = args[i]
			i++

		case "-s":
			i++
			if i >= len(args) {
				return nil, fmt.Errorf("-s requires a seed argument")
			}
			p.seed = args[i]
			i++

		default:
			// Try as system flag
			if sys, ok := parseSystemFlag(arg); ok {
				p.system = sys
				i++
				continue
			}
			// Try as input file (must not start with -)
			if p.inputFile == "" && !isFlag(arg) {
				p.inputFile = arg
				i++
				continue
			}
			return nil, fmt.Errorf("unexpected argument: %s", arg)
		}
	}

	return p, nil
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func runLegacy(args []string, stdout, stderr io.Writer) int {
	parsed, err := parseLegacyArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		printUsage(stderr)
		return ExitInvalidInput
	}

	// -r with no mode: just show version
	if parsed.showVersion && parsed.mode == "" {
		return runVersion(nil, stdout, stderr)
	}

	// Print version first if -r is combined with a mode flag
	if parsed.showVersion {
		runVersion(nil, stdout, stderr)
		fmt.Fprintln(stdout) // blank line separator
	}

	if parsed.system == "" {
		fmt.Fprintln(stderr, "error: system flag required (e.g. --dutch, --burstein)")
		printUsage(stderr)
		return ExitInvalidInput
	}

	switch parsed.mode {
	case "pair":
		return runPair(buildPairArgs(parsed), stdout, stderr)
	case "check":
		return runCheck(buildCheckArgs(parsed), stdout, stderr)
	case "generate":
		return runGenerate(buildGenerateArgs(parsed), stdout, stderr)
	default:
		fmt.Fprintln(stderr, "error: mode flag required (-p, -c, or -g)")
		printUsage(stderr)
		return ExitInvalidInput
	}
}

// buildPairArgs reconstructs args from parsed legacy state for runPair.
func buildPairArgs(p *parsedLegacyArgs) []string {
	var args []string
	if p.system != "" {
		for flag, sys := range systemFlags {
			if sys == p.system {
				args = append(args, flag)
				break
			}
		}
	}
	if p.inputFile != "" {
		args = append(args, p.inputFile)
	}
	if p.outputFile != "" {
		args = append(args, "-o", p.outputFile)
	}
	if p.wide {
		args = append(args, "-w")
	}
	return args
}

// buildCheckArgs reconstructs args from parsed legacy state for runCheck.
func buildCheckArgs(p *parsedLegacyArgs) []string {
	var args []string
	if p.system != "" {
		for flag, sys := range systemFlags {
			if sys == p.system {
				args = append(args, flag)
				break
			}
		}
	}
	if p.inputFile != "" {
		args = append(args, p.inputFile)
	}
	return args
}

// pairingsMatch checks if a PairingResult matches an existing round's games.
// Comparison is by player IDs per board, ignoring board numbering differences.
func pairingsMatch(result *cp.PairingResult, round *cp.RoundData) bool {
	if len(result.Pairings) != len(round.Games) {
		return false
	}

	// Build set of pairs from result (order-independent)
	type pair struct{ w, b string }
	resultPairs := make(map[pair]bool, len(result.Pairings))
	for _, p := range result.Pairings {
		resultPairs[pair{p.WhiteID, p.BlackID}] = true
	}

	// Check all existing games are in result
	for _, g := range round.Games {
		if !resultPairs[pair{g.WhiteID, g.BlackID}] {
			return false
		}
	}

	// Check byes match
	if len(result.Byes) != len(round.Byes) {
		return false
	}
	resultByes := make(map[string]bool, len(result.Byes))
	for _, b := range result.Byes {
		resultByes[b.PlayerID] = true
	}
	for _, b := range round.Byes {
		if !resultByes[b.PlayerID] {
			return false
		}
	}

	return true
}

// buildGenerateArgs reconstructs args from parsed legacy state for runGenerate.
func buildGenerateArgs(p *parsedLegacyArgs) []string {
	var args []string
	if p.system != "" {
		// Find the flag string for this system.
		for flag, sys := range systemFlags {
			if sys == p.system {
				args = append(args, flag)
				break
			}
		}
	}
	if p.configFile != "" {
		args = append(args, "--config", p.configFile)
	}
	if p.outputFile != "" {
		args = append(args, "-o", p.outputFile)
	}
	if p.seed != "" {
		args = append(args, "-s", p.seed)
	}
	return args
}
