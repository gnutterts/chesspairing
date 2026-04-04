// cmd/chesspairing/convert.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnutterts/chesspairing/trf"
)

const convertUsage = `Usage: chesspairing convert input-file -o output-file [options]

Convert between TRF file formats.

Arguments:
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  -o FILE          Output file (required)
  --trf-format FMT Output format: trf, trfbx, trf2026 (default: trf2026)
  --help           Show this help

Exit codes:
  0  Success
  3  Invalid input
  5  File access error

Examples:
  chesspairing convert tournament.trf -o output.trf
  chesspairing convert - -o output.trf < tournament.trf
`

func runConvert(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, convertUsage)
			return ExitSuccess
		}
	}

	flags, positional := separateFlags(args, map[string]bool{"-o": true, "--trf-format": true})

	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputFile := fs.String("o", "", "output file (required)")
	trfFormat := fs.String("trf-format", "trf2026", "output format: trf, trfbx, trf2026")
	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		fmt.Fprintf(stderr, "\nRun 'chesspairing convert --help' for usage.\n")
		return ExitInvalidInput
	}

	if *outputFile == "" {
		fmt.Fprintln(stderr, "error: -o output file required")
		fmt.Fprintf(stderr, "\nRun 'chesspairing convert --help' for usage.\n")
		return ExitInvalidInput
	}

	switch *trfFormat {
	case "trf", "trfbx", "trf2026":
		// valid
	default:
		fmt.Fprintf(stderr, "error: unknown TRF format %q (use trf, trfbx, or trf2026)\n", *trfFormat)
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

	if *trfFormat != "trf2026" {
		fmt.Fprintf(stderr, "warning: --trf-format %s not yet supported by library, writing default format\n", *trfFormat)
	}

	out, err := os.Create(*outputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot create %s: %v\n", *outputFile, err)
		return ExitFileAccess
	}
	defer out.Close()

	if err := trf.Write(out, doc); err != nil {
		fmt.Fprintf(stderr, "error: cannot write TRF: %v\n", err)
		return ExitUnexpected
	}

	return ExitSuccess
}
