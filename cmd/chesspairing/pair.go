// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/pair.go
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

const pairUsage = `Usage: chesspairing pair SYSTEM input-file [options]

Generate pairings for the next round of a tournament.

Arguments:
  SYSTEM       Pairing system flag (required):
               --dutch, --burstein, --dubov, --lim,
               --double-swiss, --team, --keizer, --roundrobin
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  -o FILE          Write output to FILE instead of stdout
  --format FORMAT  Output format: list, wide, board, xml, json (default: list)
  -w               Shorthand for --format wide
  --json           Shorthand for --format json (backward compatible)
  --help           Show this help

Output formats:
  list   Compact pair list (default). First line: count. Then "white black".
         Byes: "player 0". Compatible with bbpPairings/JaVaFo.
  wide   Human-readable table with board numbers, player names, titles,
         and ratings.
  board  Numbered board list: "Board  1:  5 -  1".
  xml    XML document with player details.
  json   JSON with pairings array and byes array.

Exit codes:
  0  Success
  1  No valid pairing could be produced
  3  Invalid input
  5  File access error

Examples:
  chesspairing pair --dutch tournament.trf
  chesspairing pair --burstein tournament.trf -o pairings.txt
  chesspairing pair --dutch - < tournament.trf
  chesspairing pair --dutch tournament.trf --format wide
  chesspairing pair --dutch tournament.trf -w
  chesspairing pair --dutch tournament.trf --json
  chesspairing pair --dutch tournament.trf --format xml
`

func runPair(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, pairUsage)
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
		fmt.Fprintf(stderr, "\nRun 'chesspairing pair --help' for usage.\n")
		return ExitInvalidInput
	}

	flags, positional := separateFlags(remaining, map[string]bool{"-o": true, "--format": true})

	fs := flag.NewFlagSet("pair", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputFile := fs.String("o", "", "output file")
	formatFlag := fs.String("format", "", "output format: list, wide, board, xml, json")
	wideFlag := fs.Bool("w", false, "shorthand for --format wide")
	jsonFlag := fs.Bool("json", false, "shorthand for --format json")
	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	// Resolve output format: explicit --format wins over shorthands
	format := "list"
	if *formatFlag != "" {
		format = *formatFlag
	} else if *wideFlag {
		format = "wide"
	} else if *jsonFlag {
		format = "json"
	}
	switch format {
	case "list", "wide", "board", "xml", "json":
		// valid
	default:
		fmt.Fprintf(stderr, "error: unknown format %q (valid: list, wide, board, xml, json)\n", format)
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		fmt.Fprintf(stderr, "\nRun 'chesspairing pair --help' for usage.\n")
		return ExitInvalidInput
	}

	inputName := positional[0]

	// Open input
	rc, err := openInput(inputName)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if inputName == "" {
			return ExitInvalidInput
		}
		return ExitFileAccess
	}
	defer func() { _ = rc.Close() }()

	doc, err := trf.Read(rc)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot convert TRF to tournament state: %v\n", err)
		return ExitInvalidInput
	}

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

	// Build player ID → start number map
	playerNumbers := make(map[string]int, len(doc.Players))
	for _, pl := range doc.Players {
		playerNumbers[fmt.Sprintf("%d", pl.StartNumber)] = pl.StartNumber
	}

	// Determine output destination
	out := io.Writer(stdout)
	if *outputFile != "" {
		outF, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(stderr, "error: cannot create %s: %v\n", *outputFile, err)
			return ExitFileAccess
		}
		defer func() { _ = outF.Close() }()
		out = outF
	}

	switch format {
	case "json":
		if err := formatPairJSON(out, result, playerNumbers); err != nil {
			fmt.Fprintf(stderr, "error: encoding JSON: %v\n", err)
			return ExitUnexpected
		}
	case "wide":
		formatPairWide(out, result, playerNumbers, state)
	case "board":
		formatPairBoard(out, result, playerNumbers)
	case "xml":
		if err := formatPairXML(out, result, playerNumbers, state); err != nil {
			fmt.Fprintf(stderr, "error: encoding XML: %v\n", err)
			return ExitUnexpected
		}
	default:
		formatPairList(out, result, playerNumbers)
	}

	return ExitSuccess
}
