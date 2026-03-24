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
	// Separate flags from positional args so flags work in any position.
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if len(args[i]) > 0 && args[i][0] == '-' {
			flags = append(flags, args[i])
			// If this flag takes a value (--profile), consume next arg too.
			if args[i] == "--profile" && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, args[i])
		}
	}

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
		formatValidationJSON(stdout, issues, *profile, "auto")
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
