// cmd/chesspairing/tiebreakers.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/gnutterts/chesspairing/tiebreaker"
)

const tiebreakersUsage = `Usage: chesspairing tiebreakers [options]

List all available tiebreaker algorithms with their IDs and display names.

Options:
  --json   Output as JSON array
  --help   Show this help
`

func runTiebreakers(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			fmt.Fprint(stdout, tiebreakersUsage)
			return ExitSuccess
		}
	}

	fs := flag.NewFlagSet("tiebreakers", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "output as JSON")
	if err := fs.Parse(args); err != nil {
		return ExitInvalidInput
	}

	ids := tiebreaker.All()
	sort.Strings(ids)

	if *jsonOut {
		type tbEntry struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		entries := make([]tbEntry, 0, len(ids))
		for _, id := range ids {
			tb, err := tiebreaker.Get(id)
			if err != nil {
				continue
			}
			entries = append(entries, tbEntry{ID: id, Name: tb.Name()})
		}
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(entries); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return ExitUnexpected
		}
		return ExitSuccess
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	for _, id := range ids {
		tb, err := tiebreaker.Get(id)
		if err != nil {
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", id, tb.Name())
	}
	_ = tw.Flush()
	return ExitSuccess
}
