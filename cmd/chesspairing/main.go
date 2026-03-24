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
	fmt.Fprintf(w, `Usage:
  chesspairing [-r]
  chesspairing SYSTEM input-file -p [output-file]
  chesspairing SYSTEM input-file -c
  chesspairing SYSTEM -g [config-file] -o output-file [-s seed]
  chesspairing standings SYSTEM input-file [options]
  chesspairing validate input-file [options]
  chesspairing convert input-file -o output-file [options]
  chesspairing version [--json]
  chesspairing tiebreakers [--json]

Systems: --dutch, --burstein, --dubov, --lim, --double-swiss, --team, --keizer, --roundrobin
`)
}
