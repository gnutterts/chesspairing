// cmd/chesspairing/output.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

// formatPairList writes bbpPairings-compatible pair output.
// Line 1: number of pairings. Lines 2+: "white black" (start numbers).
// Byes: "player 0".
func formatPairList(w io.Writer, result *cp.PairingResult, playerNumbers map[string]int) {
	fmt.Fprintf(w, "%d\n", len(result.Pairings))
	for _, p := range result.Pairings {
		fmt.Fprintf(w, "%d %d\n", playerNumbers[p.WhiteID], playerNumbers[p.BlackID])
	}
	for _, b := range result.Byes {
		fmt.Fprintf(w, "%d 0\n", playerNumbers[b.PlayerID])
	}
}

// formatStandingsText writes a human-readable standings table.
func formatStandingsText(w io.Writer, standings []cp.Standing) {
	if len(standings) == 0 {
		fmt.Fprintln(w, "(no standings)")
		return
	}

	// Determine tiebreaker columns from first entry
	var tbNames []string
	if len(standings) > 0 {
		for _, tb := range standings[0].TieBreakers {
			tbNames = append(tbNames, tb.Name)
		}
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Header
	header := "Rank\tID\tName\tScore"
	for _, name := range tbNames {
		header += "\t" + name
	}
	fmt.Fprintln(tw, header)

	// Separator
	sep := "----\t--\t----\t-----"
	for range tbNames {
		sep += "\t" + strings.Repeat("-", 8)
	}
	fmt.Fprintln(tw, sep)

	// Rows
	for _, s := range standings {
		line := fmt.Sprintf("%d\t%s\t%s\t%s", s.Rank, s.PlayerID, s.DisplayName, formatScore(s.Score))
		for _, tb := range s.TieBreakers {
			line += "\t" + formatScore(tb.Value)
		}
		fmt.Fprintln(tw, line)
	}
	tw.Flush()
}

// formatStandingsJSON writes standings as JSON. Returns any encoding error.
func formatStandingsJSON(w io.Writer, standings []cp.Standing, scoring string, tbIDs []string) error {
	output := map[string]any{
		"standings":   standings,
		"scoring":     scoring,
		"tiebreakers": tbIDs,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// formatValidationText writes validation issues in human-readable form.
func formatValidationText(w io.Writer, filename string, issues []trf.ValidationIssue) {
	var errors, warnings int
	for _, issue := range issues {
		if issue.Severity == trf.SeverityError {
			errors++
		} else {
			warnings++
		}
	}

	fmt.Fprintf(w, "%s: %d error%s, %d warning%s\n", filename, errors, plural(errors), warnings, plural(warnings))

	if errors > 0 {
		fmt.Fprintln(w, "\nErrors:")
		for _, issue := range issues {
			if issue.Severity == trf.SeverityError {
				fmt.Fprintf(w, "  %s: %s\n", issue.Field, issue.Message)
			}
		}
	}
	if warnings > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, issue := range issues {
			if issue.Severity == trf.SeverityWarning {
				fmt.Fprintf(w, "  %s: %s\n", issue.Field, issue.Message)
			}
		}
	}
}

// formatValidationJSON writes validation issues as JSON. Returns any encoding error.
func formatValidationJSON(w io.Writer, issues []trf.ValidationIssue, profile, format string) error {
	var errors, warnings []map[string]string
	for _, issue := range issues {
		entry := map[string]string{
			"field":    issue.Field,
			"severity": severityString(issue.Severity),
			"message":  issue.Message,
		}
		if issue.Severity == trf.SeverityError {
			errors = append(errors, entry)
		} else {
			warnings = append(warnings, entry)
		}
	}
	output := map[string]any{
		"valid":    len(errors) == 0,
		"errors":   errors,
		"warnings": warnings,
		"profile":  profile,
		"format":   format,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func formatScore(v float64) string {
	if v == float64(int(v)) {
		return fmt.Sprintf("%d", int(v))
	}
	return fmt.Sprintf("%.1f", v)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func severityString(s trf.Severity) string {
	if s == trf.SeverityError {
		return "error"
	}
	return "warning"
}
