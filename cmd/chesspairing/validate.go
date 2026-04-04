// cmd/chesspairing/validate.go
package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/gnutterts/chesspairing/trf"
)

var profileMap = map[string]trf.ValidationProfile{
	"minimal":  trf.ValidateGeneral,
	"standard": trf.ValidatePairingEngine,
	"strict":   trf.ValidateFIDE,
}

const validateUsage = `Usage: chesspairing validate input-file [options]

Validate a TRF16 tournament file against a validation profile.

Arguments:
  input-file   TRF16 tournament file, or "-" for stdin

Options:
  --profile PROFILE  Validation profile (default: standard)
                     minimal  — basic structural checks
                     standard — checks required for pairing engines
                     strict   — full FIDE compliance checks
  --json             Output as JSON
  --help             Show this help

Exit codes:
  0  Valid (may have warnings)
  3  Validation errors found
  5  File access error

Examples:
  chesspairing validate tournament.trf
  chesspairing validate tournament.trf --profile strict
  chesspairing validate tournament.trf --json
  chesspairing validate - < tournament.trf
`

func runValidate(args []string, stdout, stderr io.Writer) int {
	// Check for --help before any parsing
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, validateUsage)
			return ExitSuccess
		}
	}

	flags, positional := separateFlags(args, map[string]bool{"--profile": true})

	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	profile := fs.String("profile", "standard", "validation profile: minimal, standard, strict")
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(flags); err != nil {
		return ExitInvalidInput
	}

	if len(positional) < 1 {
		fmt.Fprintln(stderr, "error: input file required")
		fmt.Fprintf(stderr, "\nRun 'chesspairing validate --help' for usage.\n")
		return ExitInvalidInput
	}

	inputName := positional[0]

	vp, ok := profileMap[*profile]
	if !ok {
		fmt.Fprintf(stderr, "error: unknown profile %q (use minimal, standard, or strict)\n", *profile)
		return ExitInvalidInput
	}

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

	issues := doc.Validate(vp)

	if *jsonOut {
		if err := formatValidationJSON(stdout, issues, *profile, "auto"); err != nil {
			fmt.Fprintf(stderr, "error: encoding JSON: %v\n", err)
			return ExitUnexpected
		}
	} else {
		formatValidationText(stdout, inputName, issues)
	}

	for _, issue := range issues {
		if issue.Severity == trf.SeverityError {
			return ExitInvalidInput
		}
	}
	return ExitSuccess
}
