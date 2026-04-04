// cmd/chesspairing/main.go
package main

import (
	"fmt"
	"io"
	"os"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args, os.Stdout, os.Stderr))
}

// subcommands lists the recognized extended subcommands.
var subcommands = map[string]func([]string, io.Writer, io.Writer) int{
	"pair":        runPair,
	"check":       runCheck,
	"generate":    runGenerate,
	"version":     runVersion,
	"tiebreakers": runTiebreakers,
	"validate":    runValidate,
	"standings":   runStandings,
	"convert":     runConvert,
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		printUsage(stderr)
		return ExitInvalidInput
	}

	first := args[1]

	// Help flags
	if first == "--help" || first == "-h" || first == "help" {
		printUsage(stdout)
		return ExitSuccess
	}

	// -r flag alone: print version
	if first == "-r" && len(args) == 2 {
		return runVersion(nil, stdout, stderr)
	}

	// Check for known subcommand
	if fn, ok := subcommands[first]; ok {
		return fn(args[2:], stdout, stderr)
	}

	// Legacy mode: first arg should be a system flag (--dutch etc.) or -r combined with system
	return runLegacy(args[1:], stdout, stderr)
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `Usage: chesspairing <command> [options]

Pairing commands:
  pair         Generate pairings for the next round
  check        Verify last round's pairings by re-pairing

Tournament tools:
  standings    Compute and display tournament standings
  validate     Validate a TRF16 tournament file
  convert      Convert between TRF formats
  generate     Generate a random tournament (RTG)

Info:
  version      Show version and supported systems
  tiebreakers  List available tiebreaker algorithms
  help         Show this help

Pairing systems:
  --dutch, --burstein, --dubov, --lim,
  --double-swiss, --team, --keizer, --roundrobin

Legacy mode (bbpPairings-compatible):
  chesspairing SYSTEM input-file -p [output-file]
  chesspairing SYSTEM input-file -c
  chesspairing SYSTEM -g [config-file] -o output-file [-s seed]

Use "chesspairing <command> --help" for detailed help on any command.
`)
}
