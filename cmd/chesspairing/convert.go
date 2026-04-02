// cmd/chesspairing/convert.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnutterts/chesspairing/trf"
)

func runConvert(args []string, stdout, stderr io.Writer) int {
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
		fmt.Fprintln(stderr, "usage: chesspairing convert input-file -o output-file [--trf-format FORMAT]")
		return ExitInvalidInput
	}

	if *outputFile == "" {
		fmt.Fprintln(stderr, "error: -o output file required")
		fmt.Fprintln(stderr, "usage: chesspairing convert input-file -o output-file [--trf-format FORMAT]")
		return ExitInvalidInput
	}

	// Validate format flag (even though we can only write one format currently)
	switch *trfFormat {
	case "trf", "trfbx", "trf2026":
		// valid
	default:
		fmt.Fprintf(stderr, "error: unknown TRF format %q (use trf, trfbx, or trf2026)\n", *trfFormat)
		return ExitInvalidInput
	}

	inputFile := positional[0]

	// Read
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot open %s: %v\n", inputFile, err)
		return ExitFileAccess
	}
	defer f.Close()

	doc, err := trf.Read(f)
	if err != nil {
		fmt.Fprintf(stderr, "error: cannot parse TRF: %v\n", err)
		return ExitInvalidInput
	}

	// Note: trf.Write() currently only supports one format (TRF16).
	// The --trf-format flag is accepted for future compatibility but
	// format selection is not yet available in the library.
	if *trfFormat != "trf2026" {
		fmt.Fprintf(stderr, "warning: --trf-format %s not yet supported by library, writing default format\n", *trfFormat)
	}

	// Write
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
