package trf

import (
	"fmt"
	"io"
	"strings"
)

// Write serializes a Document to TRF16 format.
func Write(w io.Writer, doc *Document) error {
	if err := writeHeaders(w, doc); err != nil {
		return err
	}
	if err := writeXXLines(w, doc); err != nil {
		return err
	}
	for _, p := range doc.Players {
		if err := writePlayerLine(w, p); err != nil {
			return err
		}
	}
	for _, t := range doc.Teams {
		if err := writeTeamLine(w, t); err != nil {
			return err
		}
	}
	for _, rl := range doc.Other {
		if _, err := fmt.Fprintf(w, "%s %s\n", rl.Code, rl.Data); err != nil {
			return err
		}
	}
	return nil
}

func writeHeaders(w io.Writer, doc *Document) error {
	headers := []struct {
		code string
		val  string
	}{
		{"012", doc.Name},
		{"022", doc.City},
		{"032", doc.Federation},
		{"042", doc.StartDate},
		{"052", doc.EndDate},
		{"092", doc.TournamentType},
		{"102", doc.ChiefArbiter},
		{"112", doc.DeputyArbiter},
		{"122", doc.TimeControl},
	}
	for _, h := range headers {
		if h.val == "" {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s %s\n", h.code, h.val); err != nil {
			return err
		}
	}
	if doc.NumPlayers > 0 {
		if _, err := fmt.Fprintf(w, "062 %d\n", doc.NumPlayers); err != nil {
			return err
		}
	}
	if doc.NumRated > 0 {
		if _, err := fmt.Fprintf(w, "072 %d\n", doc.NumRated); err != nil {
			return err
		}
	}
	if doc.NumTeams > 0 {
		if _, err := fmt.Fprintf(w, "082 %d\n", doc.NumTeams); err != nil {
			return err
		}
	}
	for _, rd := range doc.RoundDates {
		if _, err := fmt.Fprintf(w, "132 %s\n", rd); err != nil {
			return err
		}
	}
	return nil
}

func writeXXLines(w io.Writer, doc *Document) error {
	if doc.TotalRounds > 0 {
		if _, err := fmt.Fprintf(w, "XXR %d\n", doc.TotalRounds); err != nil {
			return err
		}
	}
	if doc.InitialColor != "" {
		if _, err := fmt.Fprintf(w, "XXC %s\n", doc.InitialColor); err != nil {
			return err
		}
	}
	for _, acc := range doc.Acceleration {
		if _, err := fmt.Fprintf(w, "XXS %s\n", acc); err != nil {
			return err
		}
	}
	for _, fp := range doc.ForbiddenPairs {
		if _, err := fmt.Fprintf(w, "XXP %d %d\n", fp.Player1, fp.Player2); err != nil {
			return err
		}
	}
	if doc.Cycles > 0 {
		if _, err := fmt.Fprintf(w, "XXY %d\n", doc.Cycles); err != nil {
			return err
		}
	}
	if doc.ColorBalance != nil {
		if _, err := fmt.Fprintf(w, "XXB %t\n", *doc.ColorBalance); err != nil {
			return err
		}
	}
	if doc.MaxiTournament != nil {
		if _, err := fmt.Fprintf(w, "XXM %t\n", *doc.MaxiTournament); err != nil {
			return err
		}
	}
	if doc.ColorPreferenceType != "" {
		if _, err := fmt.Fprintf(w, "XXT %s\n", doc.ColorPreferenceType); err != nil {
			return err
		}
	}
	if doc.PrimaryScore != "" {
		if _, err := fmt.Fprintf(w, "XXG %s\n", doc.PrimaryScore); err != nil {
			return err
		}
	}
	if doc.AllowRepeatPairings != nil {
		if _, err := fmt.Fprintf(w, "XXA %t\n", *doc.AllowRepeatPairings); err != nil {
			return err
		}
	}
	if doc.MinRoundsBetweenRepeats > 0 {
		if _, err := fmt.Fprintf(w, "XXK %d\n", doc.MinRoundsBetweenRepeats); err != nil {
			return err
		}
	}
	return nil
}

// writePlayerLine writes a single 001 player line in fixed-width TRF16 format.
//
// Column layout (0-indexed bytes):
//
//	[0:3]   "001"
//	[4:8]   start number (4 chars, right-aligned)
//	[9]     sex (1 char)
//	[10:13] title (3 chars, right-aligned)
//	[14:47] name (33 chars, left-aligned)
//	[48:52] rating (4 chars, right-aligned)
//	[53:56] federation (3 chars, left-aligned)
//	[57:68] FIDE ID (11 chars, left-aligned)
//	[69:79] birth date (10 chars, left-aligned)
//	[80:84] points (4 chars, right-aligned, e.g. " 1.5")
//	[85:89] rank (4 chars, right-aligned)
//	[89:]   round results (10 chars each: "  OOOO C R")
func writePlayerLine(w io.Writer, p PlayerLine) error {
	// Allocate 89 bytes for the fixed-width header, filled with spaces.
	header := make([]byte, 89)
	for i := range header {
		header[i] = ' '
	}

	// [0:3] line code
	copy(header[0:3], "001")

	// [4:8] start number, right-aligned in 4 chars
	if err := putRight(header[4:8], fmt.Sprintf("%d", p.StartNumber)); err != nil {
		return fmt.Errorf("start number: %w", err)
	}

	// [9] sex
	if p.Sex != "" {
		header[9] = p.Sex[0]
	}

	// [10:13] title, right-aligned in 3 chars
	if p.Title != "" {
		if err := putRight(header[10:13], p.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
	}

	// [14:47] name, left-aligned in 33 chars
	if p.Name != "" {
		putLeft(header[14:47], p.Name)
	}

	// [48:52] rating, right-aligned in 4 chars
	if p.Rating > 0 {
		if err := putRight(header[48:52], fmt.Sprintf("%d", p.Rating)); err != nil {
			return fmt.Errorf("rating: %w", err)
		}
	}

	// [53:56] federation, left-aligned in 3 chars
	if p.Federation != "" {
		putLeft(header[53:56], p.Federation)
	}

	// [57:68] FIDE ID, left-aligned in 11 chars
	if p.FideID != "" {
		putLeft(header[57:68], p.FideID)
	}

	// [69:79] birth date, left-aligned in 10 chars
	if p.BirthDate != "" {
		putLeft(header[69:79], p.BirthDate)
	}

	// [80:84] points, right-aligned in 4 chars
	if err := putRight(header[80:84], formatPoints(p.Points)); err != nil {
		return fmt.Errorf("points: %w", err)
	}

	// [85:89] rank, right-aligned in 4 chars
	if p.Rank > 0 {
		if err := putRight(header[85:89], fmt.Sprintf("%d", p.Rank)); err != nil {
			return fmt.Errorf("rank: %w", err)
		}
	}

	// Build round results (10 chars each).
	var rounds strings.Builder
	for _, rr := range p.Rounds {
		_, _ = fmt.Fprintf(&rounds, "  %04d %c %c", rr.Opponent, rr.Color.Char(), rr.Result.Char())
	}

	line := strings.TrimRight(string(header)+rounds.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

// writeTeamLine writes a single 013 team line.
func writeTeamLine(w io.Writer, t TeamLine) error {
	header := make([]byte, 40)
	for i := range header {
		header[i] = ' '
	}
	copy(header[0:3], "013")
	if err := putRight(header[4:8], fmt.Sprintf("%d", t.TeamNumber)); err != nil {
		return fmt.Errorf("team number: %w", err)
	}

	if t.TeamName != "" {
		putLeft(header[8:40], t.TeamName)
	}

	var members strings.Builder
	for _, m := range t.Members {
		_, _ = fmt.Fprintf(&members, "%4d", m)
	}

	line := strings.TrimRight(string(header)+members.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

// formatPoints formats a float64 as "X.X" (e.g. "1.5", "10.0").
func formatPoints(pts float64) string {
	return fmt.Sprintf("%.1f", pts)
}

// putRight places s right-aligned within dst.
// It returns an error if s is wider than the destination field.
func putRight(dst []byte, s string) error {
	w := len(dst)
	if len(s) > w {
		return fmt.Errorf("value %q overflows %d-char field", s, w)
	}
	// Left-pad with spaces already present in dst; copy into the rightmost positions.
	copy(dst[w-len(s):], s)
	return nil
}

// putLeft places s left-aligned within dst, truncating if too long.
func putLeft(dst []byte, s string) {
	if len(s) > len(dst) {
		s = s[:len(dst)]
	}
	copy(dst, s)
}
