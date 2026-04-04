// cmd/chesspairing/output.go
package main

import (
	"encoding/json"
	"encoding/xml"
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

// formatPairJSON writes pairing results as JSON.
func formatPairJSON(w io.Writer, result *cp.PairingResult, playerNumbers map[string]int) error {
	type jsonPairing struct {
		Board int `json:"board"`
		White int `json:"white"`
		Black int `json:"black"`
	}
	type jsonBye struct {
		Player int    `json:"player"`
		Type   string `json:"type"`
	}
	type jsonOutput struct {
		Pairings []jsonPairing `json:"pairings"`
		Byes     []jsonBye     `json:"byes,omitempty"`
	}

	out := jsonOutput{
		Pairings: make([]jsonPairing, len(result.Pairings)),
	}
	for i, p := range result.Pairings {
		out.Pairings[i] = jsonPairing{
			Board: i + 1,
			White: playerNumbers[p.WhiteID],
			Black: playerNumbers[p.BlackID],
		}
	}
	for _, b := range result.Byes {
		out.Byes = append(out.Byes, jsonBye{
			Player: playerNumbers[b.PlayerID],
			Type:   b.Type.String(),
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// formatPairWide writes a human-readable wide table with player names, titles, and ratings.
func formatPairWide(w io.Writer, result *cp.PairingResult, playerNumbers map[string]int, state *cp.TournamentState) {
	// Build player ID → PlayerEntry lookup
	players := make(map[string]*cp.PlayerEntry, len(state.Players))
	for i := range state.Players {
		players[state.Players[i].ID] = &state.Players[i]
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "Board\tWhite\tRtg\t\tBlack\tRtg")
	fmt.Fprintln(tw, "-----\t-----\t---\t\t-----\t---")

	for i, p := range result.Pairings {
		wNum := playerNumbers[p.WhiteID]
		bNum := playerNumbers[p.BlackID]
		wName := playerDisplayWide(players[p.WhiteID], wNum)
		bName := playerDisplayWide(players[p.BlackID], bNum)
		wRtg := playerRating(players[p.WhiteID])
		bRtg := playerRating(players[p.BlackID])
		fmt.Fprintf(tw, "%d\t%s\t%s\t-\t%s\t%s\n", i+1, wName, wRtg, bName, bRtg)
	}
	for _, b := range result.Byes {
		num := playerNumbers[b.PlayerID]
		name := playerDisplayWide(players[b.PlayerID], num)
		rtg := playerRating(players[b.PlayerID])
		fmt.Fprintf(tw, "\t%s\t%s\t\tBye (%s)\t\n", name, rtg, b.Type.String())
	}
	tw.Flush()
}

// playerDisplayWide formats "TPN Title LastName, FirstName" or "TPN LastName, FirstName".
func playerDisplayWide(pe *cp.PlayerEntry, num int) string {
	if pe == nil {
		return fmt.Sprintf("%d", num)
	}
	if pe.Title != "" {
		return fmt.Sprintf("%d %s %s", num, pe.Title, pe.DisplayName)
	}
	return fmt.Sprintf("%d %s", num, pe.DisplayName)
}

// playerRating returns the rating as a string, or "" if the player is unknown.
func playerRating(pe *cp.PlayerEntry) string {
	if pe == nil || pe.Rating == 0 {
		return ""
	}
	return fmt.Sprintf("%d", pe.Rating)
}

// formatPairBoard writes numbered board pairings: "Board  1:  5 -  1".
func formatPairBoard(w io.Writer, result *cp.PairingResult, playerNumbers map[string]int) {
	// Determine field width for board and player numbers
	boardWidth := digitWidth(len(result.Pairings))
	playerWidth := maxPlayerWidth(result, playerNumbers)

	for i, p := range result.Pairings {
		fmt.Fprintf(w, "Board %*d: %*d - %*d\n",
			boardWidth, i+1,
			playerWidth, playerNumbers[p.WhiteID],
			playerWidth, playerNumbers[p.BlackID])
	}
	for _, b := range result.Byes {
		fmt.Fprintf(w, "Bye: %*d\n", playerWidth, playerNumbers[b.PlayerID])
	}
}

// digitWidth returns the number of digits needed to display n.
func digitWidth(n int) int {
	if n <= 0 {
		return 1
	}
	w := 0
	for n > 0 {
		w++
		n /= 10
	}
	return w
}

// maxPlayerWidth returns the digit width of the largest player number in the result.
func maxPlayerWidth(result *cp.PairingResult, playerNumbers map[string]int) int {
	maxNum := 0
	for _, p := range result.Pairings {
		if n := playerNumbers[p.WhiteID]; n > maxNum {
			maxNum = n
		}
		if n := playerNumbers[p.BlackID]; n > maxNum {
			maxNum = n
		}
	}
	for _, b := range result.Byes {
		if n := playerNumbers[b.PlayerID]; n > maxNum {
			maxNum = n
		}
	}
	return digitWidth(maxNum)
}

// formatPairXML writes pairings as XML.
func formatPairXML(w io.Writer, result *cp.PairingResult, playerNumbers map[string]int, state *cp.TournamentState) error {
	// Build player ID → PlayerEntry lookup
	players := make(map[string]*cp.PlayerEntry, len(state.Players))
	for i := range state.Players {
		players[state.Players[i].ID] = &state.Players[i]
	}

	type xmlPlayer struct {
		Number int    `xml:"number,attr"`
		Name   string `xml:"name,attr,omitempty"`
		Rating int    `xml:"rating,attr,omitempty"`
		Title  string `xml:"title,attr,omitempty"`
	}
	type xmlBoard struct {
		Number int       `xml:"number,attr"`
		White  xmlPlayer `xml:"white"`
		Black  xmlPlayer `xml:"black"`
	}
	type xmlBye struct {
		Number int    `xml:"number,attr"`
		Name   string `xml:"name,attr,omitempty"`
		Type   string `xml:"type,attr"`
	}
	type xmlPairings struct {
		XMLName xml.Name   `xml:"pairings"`
		Round   int        `xml:"round,attr"`
		Boards  int        `xml:"boards,attr"`
		Byes    int        `xml:"byes,attr"`
		Board   []xmlBoard `xml:"board"`
		Bye     []xmlBye   `xml:"bye"`
	}

	makePlayer := func(id string) xmlPlayer {
		p := xmlPlayer{
			Number: playerNumbers[id],
		}
		if pe := players[id]; pe != nil {
			p.Name = pe.DisplayName
			p.Rating = pe.Rating
			p.Title = pe.Title
		}
		return p
	}

	out := xmlPairings{
		Round:  state.CurrentRound + 1,
		Boards: len(result.Pairings),
		Byes:   len(result.Byes),
	}
	for i, p := range result.Pairings {
		out.Board = append(out.Board, xmlBoard{
			Number: i + 1,
			White:  makePlayer(p.WhiteID),
			Black:  makePlayer(p.BlackID),
		})
	}
	for _, b := range result.Byes {
		xb := xmlBye{
			Number: playerNumbers[b.PlayerID],
			Type:   b.Type.String(),
		}
		if pe := players[b.PlayerID]; pe != nil {
			xb.Name = pe.DisplayName
		}
		out.Bye = append(out.Bye, xb)
	}

	fmt.Fprint(w, xml.Header)
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(out); err != nil {
		return err
	}
	fmt.Fprintln(w) // trailing newline
	return nil
}
