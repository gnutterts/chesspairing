// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/version.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/tiebreaker"
)

const versionUsage = `Usage: chesspairing version [options]

Show version information and supported pairing/scoring systems.

Options:
  --json   Output as JSON (includes full tiebreaker list)
  --help   Show this help
`

func runVersion(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, versionUsage)
			return ExitSuccess
		}
	}

	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		return ExitInvalidInput
	}

	pairingSystems := []string{
		string(cp.PairingDutch),
		string(cp.PairingBurstein),
		string(cp.PairingDubov),
		string(cp.PairingLim),
		string(cp.PairingDoubleSwiss),
		string(cp.PairingTeam),
		string(cp.PairingKeizer),
		string(cp.PairingRoundRobin),
	}
	scoringSystems := []string{
		string(cp.ScoringStandard),
		string(cp.ScoringKeizer),
		string(cp.ScoringFootball),
	}
	tbs := tiebreaker.All()
	sort.Strings(tbs)

	if *jsonOut {
		v := map[string]any{
			"version":        version,
			"pairingSystems": pairingSystems,
			"scoringSystems": scoringSystems,
			"tiebreakers":    tbs,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(v); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return ExitUnexpected
		}
		return ExitSuccess
	}

	fmt.Fprintf(stdout, "chesspairing %s\n\n", version)
	fmt.Fprintf(stdout, "Pairing systems:  %s\n", joinComma(pairingSystems))
	fmt.Fprintf(stdout, "Scoring systems:  %s\n", joinComma(scoringSystems))
	fmt.Fprintf(stdout, "Tiebreakers:      %d available\n", len(tbs))
	return ExitSuccess
}

func joinComma(ss []string) string {
	if len(ss) == 0 {
		return "(none)"
	}
	result := ss[0]
	for _, s := range ss[1:] {
		result += ", " + s
	}
	return result
}
