// cmd/chesspairing/validate.go
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/gnutterts/chesspairing/trf"
)

var profileMap = map[string]trf.ValidationProfile{
	"minimal":  trf.ValidateGeneral,
	"standard": trf.ValidatePairingEngine,
	"strict":   trf.ValidateFIDE,
}

func runValidate(args []string, stdout, stderr io.Writer) int {
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
		fmt.Fprintln(stderr, "usage: chesspairing validate input-file [--profile PROFILE] [--json]")
		return ExitInvalidInput
	}

	inputFile := positional[0]

	vp, ok := profileMap[*profile]
	if !ok {
		fmt.Fprintf(stderr, "error: unknown profile %q (use minimal, standard, or strict)\n", *profile)
		return ExitInvalidInput
	}

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

	issues := doc.Validate(vp)

	if *jsonOut {
		if err := formatValidationJSON(stdout, issues, *profile, "auto"); err != nil {
			fmt.Fprintf(stderr, "error: encoding JSON: %v\n", err)
			return ExitUnexpected
		}
	} else {
		formatValidationText(stdout, inputFile, issues)
	}

	// Exit 3 if there are errors, 0 if only warnings
	for _, issue := range issues {
		if issue.Severity == trf.SeverityError {
			return ExitInvalidInput
		}
	}
	return ExitSuccess
}
