package trf

import (
	"fmt"
	"io"
	"strings"
)

// Write serializes a Document to TRF format (supports both TRF16 and TRF-2026).
func Write(w io.Writer, doc *Document) error {
	if err := writeHeaders(w, doc); err != nil {
		return err
	}
	if err := writeTRF2026Headers(w, doc); err != nil {
		return err
	}
	if err := writeXXLines(w, doc); err != nil {
		return err
	}
	for _, c := range doc.Comments {
		if _, err := fmt.Fprintf(w, "### %s\n", c); err != nil {
			return err
		}
	}
	for _, p := range doc.Players {
		if err := writePlayerLine(w, p); err != nil {
			return err
		}
	}
	for _, rec := range doc.NRSRecords {
		if _, err := fmt.Fprintf(w, "%s\n", rec.Raw); err != nil {
			return err
		}
	}
	for _, t := range doc.Teams {
		if err := writeTeamLine(w, t); err != nil {
			return err
		}
	}
	if err := writeTRF2026Data(w, doc); err != nil {
		return err
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

	// 112: write all deputy arbiters if present (TRF-2026 supports multiple),
	// otherwise write the single legacy field.
	if len(doc.DeputyArbiters) > 0 {
		for _, da := range doc.DeputyArbiters {
			if _, err := fmt.Fprintf(w, "112 %s\n", da); err != nil {
				return err
			}
		}
	} else if doc.DeputyArbiter != "" {
		if _, err := fmt.Fprintf(w, "112 %s\n", doc.DeputyArbiter); err != nil {
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

// writeTRF2026Headers writes TRF-2026 header lines (142-362).
func writeTRF2026Headers(w io.Writer, doc *Document) error {
	if doc.TotalRounds26 > 0 {
		if _, err := fmt.Fprintf(w, "142 %d\n", doc.TotalRounds26); err != nil {
			return err
		}
	}
	if doc.InitialColor26 != "" {
		if _, err := fmt.Fprintf(w, "152 %s\n", doc.InitialColor26); err != nil {
			return err
		}
	}
	if doc.ScoringSystem != "" {
		if _, err := fmt.Fprintf(w, "162 %s\n", doc.ScoringSystem); err != nil {
			return err
		}
	}
	if doc.StartingRankMethod != "" {
		if _, err := fmt.Fprintf(w, "172 %s\n", doc.StartingRankMethod); err != nil {
			return err
		}
	}
	if doc.CodedTournamentType != "" {
		if _, err := fmt.Fprintf(w, "192 %s\n", doc.CodedTournamentType); err != nil {
			return err
		}
	}
	if doc.TieBreakDef != "" {
		if _, err := fmt.Fprintf(w, "202 %s\n", doc.TieBreakDef); err != nil {
			return err
		}
	}
	if doc.EncodedTimeControl != "" {
		if _, err := fmt.Fprintf(w, "222 %s\n", doc.EncodedTimeControl); err != nil {
			return err
		}
	}
	if doc.TeamInitialColor != "" {
		if _, err := fmt.Fprintf(w, "352 %s\n", doc.TeamInitialColor); err != nil {
			return err
		}
	}
	if doc.TeamScoringSystem != "" {
		if _, err := fmt.Fprintf(w, "362 %s\n", doc.TeamScoringSystem); err != nil {
			return err
		}
	}
	return nil
}

// writeTRF2026Data writes TRF-2026 data records (240, 250, 260, 300, 310, 320, 330, 801, 802).
func writeTRF2026Data(w io.Writer, doc *Document) error {
	for _, t := range doc.NewTeams {
		if err := writeNewTeamLine(w, t); err != nil {
			return err
		}
	}
	for _, a := range doc.Absences {
		if err := writeAbsenceRecord(w, a); err != nil {
			return err
		}
	}
	for _, a := range doc.Accelerations26 {
		if err := writeAccelerationRecord(w, a); err != nil {
			return err
		}
	}
	for _, fp := range doc.ForbiddenPairs26 {
		if err := writeForbiddenPairRecord(w, fp); err != nil {
			return err
		}
	}
	for _, tr := range doc.TeamRoundData {
		if err := writeTeamRoundEntry(w, tr); err != nil {
			return err
		}
	}
	for _, ts := range doc.TeamRoundScores {
		if err := writeTeamRoundScoreEntry(w, ts); err != nil {
			return err
		}
	}
	for _, oaf := range doc.OldAbsentForfeits {
		if err := writeOldAbsentForfeit(w, oaf); err != nil {
			return err
		}
	}
	for _, dtr := range doc.DetailedTeamResults {
		if err := writeDetailedTeamResult(w, dtr); err != nil {
			return err
		}
	}
	for _, str := range doc.SimpleTeamResults {
		if err := writeSimpleTeamResult(w, str); err != nil {
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

// --- TRF-2026 data record writers ---

// writeAbsenceRecord writes a 240 line.
func writeAbsenceRecord(w io.Writer, a AbsenceRecord) error {
	var b strings.Builder
	fmt.Fprintf(&b, "240 %s %3d", a.Type, a.Round)
	for _, p := range a.Players {
		fmt.Fprintf(&b, " %4d", p)
	}
	_, err := fmt.Fprintf(w, "%s\n", b.String())
	return err
}

// writeAccelerationRecord writes a 250 line. Uses raw data for round-trip
// fidelity when available.
func writeAccelerationRecord(w io.Writer, a AccelerationRecord) error {
	if a.Raw != "" {
		_, err := fmt.Fprintf(w, "250 %s\n", a.Raw)
		return err
	}
	_, err := fmt.Fprintf(w, "250 %2.0f %9.0f %3d %3d %4d %4d\n",
		a.MatchPoints, a.GamePoints, a.FirstRound, a.LastRound, a.FirstPlayer, a.LastPlayer)
	return err
}

// writeForbiddenPairRecord writes a 260 line. Uses raw data for round-trip
// fidelity when available.
func writeForbiddenPairRecord(w io.Writer, fp ForbiddenPairRecord) error {
	if fp.Raw != "" {
		_, err := fmt.Fprintf(w, "260 %s\n", fp.Raw)
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "260 %3d %3d", fp.FirstRound, fp.LastRound)
	for _, p := range fp.Players {
		fmt.Fprintf(&b, " %4d", p)
	}
	_, err := fmt.Fprintf(w, "%s\n", b.String())
	return err
}

// writeTeamRoundEntry writes a 300 line.
func writeTeamRoundEntry(w io.Writer, tr TeamRoundEntry) error {
	var b strings.Builder
	fmt.Fprintf(&b, "300 %3d %2d %2d", tr.Round, tr.Team1, tr.Team2)
	for _, bd := range tr.Boards {
		fmt.Fprintf(&b, " %4d", bd)
	}
	line := strings.TrimRight(b.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

// writeNewTeamLine writes a 310 line using fixed-width columns matching the
// parse layout:
//
//	[0:3]   "310"
//	[4:7]   team number (3 chars, right-aligned)
//	[8:40]  team name (32 chars, left-aligned)
//	[41:46] federation (5 chars, left-aligned)
//	[47:53] avg rating (6 chars, right-aligned)
//	[54:60] match points (6 chars, right-aligned)
//	[61:67] game points (6 chars, right-aligned)
//	[68:71] rank (3 chars, right-aligned)
//	[73:]   members (4 chars each, right-aligned)
func writeNewTeamLine(w io.Writer, t NewTeamLine) error {
	header := make([]byte, 71)
	for i := range header {
		header[i] = ' '
	}
	copy(header[0:3], "310")

	if err := putRight(header[4:7], fmt.Sprintf("%d", t.TeamNumber)); err != nil {
		return fmt.Errorf("310 team number: %w", err)
	}
	if t.TeamName != "" {
		putLeft(header[8:41], t.TeamName)
	}
	if t.Federation != "" {
		putLeft(header[41:46], t.Federation)
	}
	if t.AvgRating != 0 {
		if err := putRight(header[46:53], fmt.Sprintf("%.0f", t.AvgRating)); err != nil {
			return fmt.Errorf("310 avg rating: %w", err)
		}
	}
	if t.MatchPoints != 0 {
		if err := putRight(header[53:59], fmt.Sprintf("%.0f", t.MatchPoints)); err != nil {
			return fmt.Errorf("310 match points: %w", err)
		}
	}
	if t.GamePoints != 0 {
		if err := putRight(header[59:67], fmt.Sprintf("%.1f", t.GamePoints)); err != nil {
			return fmt.Errorf("310 game points: %w", err)
		}
	}
	if t.Rank > 0 {
		if err := putRight(header[67:71], fmt.Sprintf("%d", t.Rank)); err != nil {
			return fmt.Errorf("310 rank: %w", err)
		}
	}

	var members strings.Builder
	for _, m := range t.Members {
		fmt.Fprintf(&members, " %4d", m)
	}

	line := strings.TrimRight(string(header)+members.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

// writeTeamRoundScoreEntry writes a 320 line. Uses raw data for round-trip
// fidelity when available.
func writeTeamRoundScoreEntry(w io.Writer, ts TeamRoundScoreEntry) error {
	if ts.Raw != "" {
		_, err := fmt.Fprintf(w, "320 %s\n", ts.Raw)
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "320 %2d %6.1f", ts.TeamNumber, ts.GamePoints)
	for _, s := range ts.Scores {
		fmt.Fprintf(&b, " %s", s)
	}
	_, err := fmt.Fprintf(w, "%s\n", b.String())
	return err
}

// writeOldAbsentForfeit writes a 330 line.
func writeOldAbsentForfeit(w io.Writer, oaf OldAbsentForfeit) error {
	_, err := fmt.Fprintf(w, "330 %s %3d %3d %3d\n", oaf.ResultType, oaf.Round, oaf.WhiteTeam, oaf.BlackTeam)
	return err
}

// writeDetailedTeamResult writes an 801 line. Uses raw data for round-trip
// fidelity since the format is complex with variable-width fields.
func writeDetailedTeamResult(w io.Writer, dtr DetailedTeamResult) error {
	if dtr.Raw != "" {
		_, err := fmt.Fprintf(w, "801 %s\n", dtr.Raw)
		return err
	}
	// Reconstruct from parsed fields.
	var b strings.Builder
	fmt.Fprintf(&b, "801 %2d %-5s %4.0f %6.1f",
		dtr.TeamNumber, dtr.TeamName, dtr.MatchPoints, dtr.GamePoints)
	for _, rd := range dtr.Rounds {
		if rd.ByeType != "" {
			fmt.Fprintf(&b, "  %s      ", rd.ByeType)
		} else {
			fmt.Fprintf(&b, "  %2d %s %s %s", rd.Opponent, rd.Color, rd.Results, rd.BoardOrder)
		}
	}
	line := strings.TrimRight(b.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}

// writeSimpleTeamResult writes an 802 line. Uses raw data for round-trip
// fidelity since the format is complex with variable-width fields.
func writeSimpleTeamResult(w io.Writer, str SimpleTeamResult) error {
	if str.Raw != "" {
		_, err := fmt.Fprintf(w, "802 %s\n", str.Raw)
		return err
	}
	// Reconstruct from parsed fields.
	var b strings.Builder
	fmt.Fprintf(&b, "802 %3d %-5s %6.0f %8.1f",
		str.TeamNumber, str.TeamName, str.MatchPoints, str.GamePoints)
	for _, rd := range str.Rounds {
		if rd.ByeType != "" {
			fmt.Fprintf(&b, " %s %4.1f", rd.ByeType, rd.GamePoints)
		} else {
			gpStr := fmt.Sprintf("%.1f", rd.GamePoints)
			if rd.Forfeit {
				gpStr += "f"
			}
			fmt.Fprintf(&b, " %2d %s %s", rd.Opponent, rd.Color, gpStr)
		}
	}
	line := strings.TrimRight(b.String(), " ")
	_, err := fmt.Fprintf(w, "%s\n", line)
	return err
}
