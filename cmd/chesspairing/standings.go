// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// cmd/chesspairing/standings.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/standings"
	"github.com/gnutterts/chesspairing/trf"
)

const standingsUsage = `Usage: chesspairing standings [SYSTEM] input-file [options]

Compute and display tournament standings.

Arguments:
  SYSTEM       Pairing system flag (required unless --tiebreakers is given):
               --dutch, --burstein, --dubov, --lim,
               --double-swiss, --team, --keizer, --roundrobin
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  -o FILE            Write output to FILE instead of stdout
  --scoring SYSTEM   Scoring system: standard, keizer, football (default: standard)
  --tiebreakers IDS  Comma-separated tiebreaker IDs (default: system-specific)
  --win N            Points for a win (overrides default)
  --draw N           Points for a draw
  --loss N           Points for a loss
  --forfeit-win N    Points for a forfeit win
  --bye N            Points for a bye
  --forfeit-loss N   Points for a forfeit loss
  --json             Output as JSON
  --help             Show this help

Exit codes:
  0  Success
  3  Invalid input
  5  File access error

Examples:
  chesspairing standings --dutch tournament.trf
  chesspairing standings --dutch tournament.trf --tiebreakers buchholz,wins
  chesspairing standings tournament.trf --tiebreakers buchholz,wins
  chesspairing standings --dutch tournament.trf --json
  chesspairing standings --dutch tournament.trf -o standings.txt
  chesspairing standings --dutch - < tournament.trf
`

func runStandings(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, standingsUsage)
			return ExitSuccess
		}
	}

	// First pass: extract system flag (before flag parsing, since --dutch etc. aren't flag-package flags)
	var system cp.PairingSystem
	var remaining []string
	for _, arg := range args {
		if sys, ok := parseSystemFlag(arg); ok {
			if system != "" {
				fmt.Fprintf(stderr, "warning: multiple system flags, using %s\n", arg)
			}
			system = sys
		} else {
			remaining = append(remaining, arg)
		}
	}

	flags, positional := separateFlags(remaining, map[string]bool{
		"-o": true, "--scoring": true, "--tiebreakers": true,
		"--win": true, "--draw": true, "--loss": true,
		"--forfeit-win": true, "--bye": true, "--forfeit-loss": true,
	})

	fs := flag.NewFlagSet("standings", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputFile := fs.String("o", "", "output file")
	scoring := fs.String("scoring", "standard", "scoring system: standard, keizer, football")
	tbFlag := fs.String("tiebreakers", "", "comma-separated tiebreaker IDs (default: system-specific)")
	jsonOut := fs.Bool("json", false, "output as JSON")
	win := fs.Float64("win", -1, "points for a win")
	draw := fs.Float64("draw", -1, "points for a draw")
	loss := fs.Float64("loss", -1, "points for a loss")
	forfeitWin := fs.Float64("forfeit-win", -1, "points for a forfeit win")
	bye := fs.Float64("bye", -1, "points for a bye")
	forfeitLoss := fs.Float64("forfeit-loss", -1, "points for a forfeit loss")

	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		return ExitInvalidInput
	}

	// System flag is required unless --tiebreakers is explicitly given
	if system == "" && *tbFlag == "" {
		fmt.Fprintln(stderr, "error: system flag required when --tiebreakers is not specified")
		fmt.Fprintf(stderr, "\nRun 'chesspairing standings --help' for usage.\n")
		return ExitInvalidInput
	}

	inputFile := positional[0]

	rc, err := openInput(inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if inputFile == "" {
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
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	// Build scoring options from CLI flags (override TRF / defaults)
	scoringOpts := state.ScoringConfig.Options
	if scoringOpts == nil {
		scoringOpts = map[string]any{}
	}
	if *win >= 0 {
		scoringOpts["pointWin"] = *win
	}
	if *draw >= 0 {
		scoringOpts["pointDraw"] = *draw
	}
	if *loss >= 0 {
		scoringOpts["pointLoss"] = *loss
	}
	if *forfeitWin >= 0 {
		scoringOpts["pointForfeitWin"] = *forfeitWin
	}
	if *bye >= 0 {
		scoringOpts["pointBye"] = *bye
	}
	if *forfeitLoss >= 0 {
		scoringOpts["pointForfeitLoss"] = *forfeitLoss
	}

	scoringSystem := cp.ScoringSystem(*scoring)
	scorer, err := newScorer(scoringSystem, scoringOpts)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitInvalidInput
	}

	ctx := rootContext()

	// Determine tiebreakers
	var tbIDs []string
	if *tbFlag != "" {
		tbIDs = strings.Split(*tbFlag, ",")
	} else if system != "" {
		tbIDs = cp.DefaultTiebreakers(system)
	} else {
		fmt.Fprintln(stderr, "error: system flag required when --tiebreakers is not specified")
		fmt.Fprintf(stderr, "\nRun 'chesspairing standings --help' for usage.\n")
		return ExitInvalidInput
	}

	standingsRows, err := standings.BuildByID(ctx, state, scorer, tbIDs)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return ExitUnexpected
	}

	// Determine output destination
	out := io.Writer(stdout)
	var outF *os.File
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(stderr, "error: cannot create %s: %v\n", *outputFile, err)
			return ExitFileAccess
		}
		outF = f
		out = f
	}

	var writeErr error
	if *jsonOut {
		writeErr = formatStandingsJSON(out, standingsRows, *scoring, tbIDs)
	} else {
		formatStandingsText(out, standingsRows)
	}

	if writeErr != nil {
		if outF != nil {
			_ = outF.Close()
		}
		fmt.Fprintf(stderr, "error: encoding output: %v\n", writeErr)
		return ExitUnexpected
	}

	if outF != nil {
		if err := outF.Close(); err != nil {
			fmt.Fprintf(stderr, "error: closing %s: %v\n", *outputFile, err)
			return ExitUnexpected
		}
	}

	return ExitSuccess
}
