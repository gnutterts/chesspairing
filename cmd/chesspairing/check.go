// cmd/chesspairing/check.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

const checkUsage = `Usage: chesspairing check SYSTEM input-file [options]

Verify the last round's pairings by re-pairing and comparing.

Strips the last round from the tournament, generates fresh pairings using
the specified system, and compares them against the original last round.

Arguments:
  SYSTEM       Pairing system flag (required):
               --dutch, --burstein, --dubov, --lim,
               --double-swiss, --team, --keizer, --roundrobin
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  --json       Output result as JSON instead of text
  --help       Show this help

Exit codes:
  0  Pairings match
  1  Pairings do not match (mismatch)
  3  Invalid input (no rounds, bad file)
  5  File access error

Examples:
  chesspairing check --dutch tournament.trf
  chesspairing check --dutch - < tournament.trf
  chesspairing check --dutch tournament.trf --json
`

func runCheck(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, checkUsage)
			return ExitSuccess
		}
	}

	// Extract system flag
	var system cp.PairingSystem
	var remaining []string
	for _, arg := range args {
		if sys, ok := parseSystemFlag(arg); ok {
			system = sys
		} else {
			remaining = append(remaining, arg)
		}
	}

	if system == "" {
		fmt.Fprintln(stderr, "error: system flag required (e.g. --dutch)")
		fmt.Fprintf(stderr, "\nRun 'chesspairing check --help' for usage.\n")
		return ExitInvalidInput
	}

	flags, positional := separateFlags(remaining, nil)

	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		fmt.Fprintf(stderr, "\nRun 'chesspairing check --help' for usage.\n")
		return ExitInvalidInput
	}

	inputName := positional[0]

	rc, err := openInput(inputName)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if inputName == "" {
			return ExitInvalidInput
		}
		return ExitFileAccess
	}
	defer rc.Close()

	doc, err := trf.Read(rc)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	if len(state.Rounds) == 0 {
		fmt.Fprintln(stderr, "error: no rounds in tournament to check")
		return ExitInvalidInput
	}

	// Remove the last round and re-pair
	lastRound := state.Rounds[len(state.Rounds)-1]
	state.Rounds = state.Rounds[:len(state.Rounds)-1]
	state.CurrentRound = len(state.Rounds)

	state.PairingConfig.System = system

	pairer, err := newPairer(system, state.PairingConfig.Options)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	ctx := context.Background()
	result, err := pairer.Pair(ctx, state)
	if err != nil {
		fmt.Fprintf(stderr, "error: pairing failed: %v\n", err)
		return ExitNoPairing
	}

	match := pairingsMatch(result, &lastRound)

	if *jsonOut {
		out := map[string]any{
			"match":  match,
			"system": string(system),
			"round":  len(state.Rounds) + 1,
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(out); err != nil {
			fmt.Fprintf(stderr, "error: encoding JSON: %v\n", err)
			return ExitUnexpected
		}
	} else {
		if match {
			fmt.Fprintln(stdout, "OK: pairings match")
		} else {
			fmt.Fprintln(stdout, "MISMATCH: generated pairings differ from existing round")
		}
	}

	if match {
		return ExitSuccess
	}
	return ExitNoPairing
}
