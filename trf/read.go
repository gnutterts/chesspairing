// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Read parses a TRF file from the reader (supports both TRF16 and TRF-2026).
func Read(r io.Reader) (*Document, error) {
	doc := &Document{}
	scanner := bufio.NewScanner(r)
	// Increase scanner buffer for long lines (801/802 records can be very long).
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimRight(line, "\r")

		if len(line) < 3 {
			continue
		}

		code := line[:3]
		data := ""
		if len(line) > 4 {
			data = line[4:]
		}

		switch code {
		// --- TRF16 header lines ---
		case "012":
			doc.Name = data
		case "022":
			doc.City = data
		case "032":
			doc.Federation = data
		case "042":
			doc.StartDate = data
		case "052":
			doc.EndDate = data
		case "062":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid player count: %q", data)}
			}
			doc.NumPlayers = n
		case "072":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid rated count: %q", data)}
			}
			doc.NumRated = n
		case "082":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid team count: %q", data)}
			}
			doc.NumTeams = n
		case "092":
			doc.TournamentType = data
		case "102":
			doc.ChiefArbiter = data
		case "112":
			// TRF-2026 allows multiple 112 lines.
			if doc.DeputyArbiter == "" {
				doc.DeputyArbiter = data
			}
			doc.DeputyArbiters = append(doc.DeputyArbiters, data)
		case "122":
			doc.TimeControl = data
		case "132":
			doc.RoundDates = append(doc.RoundDates, data)

		// --- TRF-2026 new header lines ---
		case "142":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid total rounds: %q", data)}
			}
			doc.TotalRounds26 = n
		case "152":
			doc.InitialColor26 = strings.TrimSpace(data)
		case "162":
			doc.ScoringSystem = data
		case "172":
			doc.StartingRankMethod = strings.TrimSpace(data)
		case "192":
			doc.CodedTournamentType = strings.TrimSpace(data)
		case "202":
			doc.TieBreakDef = strings.TrimSpace(data)
		case "222":
			doc.EncodedTimeControl = strings.TrimSpace(data)
		case "352":
			doc.TeamInitialColor = strings.TrimSpace(data)
		case "362":
			doc.TeamScoringSystem = data

		// --- TRF-2026 data records ---
		case "240":
			rec, err := parseAbsenceRecord(data, lineNum)
			if err != nil {
				return nil, err
			}
			doc.Absences = append(doc.Absences, rec)
		case "250":
			rec, err := parseAccelerationRecord(data, lineNum)
			if err != nil {
				return nil, err
			}
			doc.Accelerations26 = append(doc.Accelerations26, rec)
		case "260":
			rec, err := parseForbiddenPairRecord(data, lineNum)
			if err != nil {
				return nil, err
			}
			doc.ForbiddenPairs26 = append(doc.ForbiddenPairs26, rec)
		case "300":
			rec, err := parseTeamRoundEntry(data, lineNum)
			if err != nil {
				return nil, err
			}
			doc.TeamRoundData = append(doc.TeamRoundData, rec)
		case "310":
			rec, err := parseNewTeamLine(line, lineNum)
			if err != nil {
				return nil, err
			}
			doc.NewTeams = append(doc.NewTeams, rec)
		case "320":
			rec := parseTeamRoundScoreEntry(data)
			doc.TeamRoundScores = append(doc.TeamRoundScores, rec)
		case "330":
			rec, err := parseOldAbsentForfeit(data, lineNum)
			if err != nil {
				return nil, err
			}
			doc.OldAbsentForfeits = append(doc.OldAbsentForfeits, rec)
		case "801":
			rec := parseDetailedTeamResult(data)
			doc.DetailedTeamResults = append(doc.DetailedTeamResults, rec)
		case "802":
			rec := parseSimpleTeamResult(data)
			doc.SimpleTeamResults = append(doc.SimpleTeamResults, rec)

		// --- TRF16 legacy extension lines ---
		case "XXR":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid total rounds: %q", data)}
			}
			doc.TotalRounds = n
		case "XXC":
			doc.InitialColor = strings.TrimSpace(data)
		case "XXS":
			doc.Acceleration = append(doc.Acceleration, data)
		case "XXP":
			fp, err := parseForbiddenPair(data)
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: err.Error()}
			}
			doc.ForbiddenPairs = append(doc.ForbiddenPairs, fp)
		case "XXY":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid cycles: %q", data)}
			}
			doc.Cycles = n
		case "XXB":
			b, err := strconv.ParseBool(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid color balance: %q", data)}
			}
			doc.ColorBalance = &b
		case "XXM":
			b, err := strconv.ParseBool(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid maxi tournament: %q", data)}
			}
			doc.MaxiTournament = &b
		case "XXT":
			doc.ColorPreferenceType = strings.TrimSpace(data)
		case "XXG":
			doc.PrimaryScore = strings.TrimSpace(data)
		case "XXA":
			b, err := strconv.ParseBool(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid allow repeat pairings: %q", data)}
			}
			doc.AllowRepeatPairings = &b
		case "XXK":
			n, err := strconv.Atoi(strings.TrimSpace(data))
			if err != nil {
				return nil, &ParseError{Line: lineNum, Code: code, Message: fmt.Sprintf("invalid min rounds between repeats: %q", data)}
			}
			doc.MinRoundsBetweenRepeats = n

		// --- Data lines ---
		case "001":
			pl, err := parsePlayerLine(line, lineNum)
			if err != nil {
				return nil, err
			}
			doc.Players = append(doc.Players, pl)
		case "013":
			tl, err := parseTeamLine(line, lineNum)
			if err != nil {
				return nil, err
			}
			doc.Teams = append(doc.Teams, tl)
		case "###":
			doc.Comments = append(doc.Comments, data)
		default:
			// Check if this is an NRS record (3-letter alpha code with
			// player-line layout: at least 68 chars with a numeric start
			// number at bytes 4-7).
			if isNRSCode(code, line) {
				rec := parseNRSRecord(code, line)
				doc.NRSRecords = append(doc.NRSRecords, rec)
			} else {
				doc.Other = append(doc.Other, RawLine{Code: code, Data: data})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("trf: read error: %w", err)
	}

	return doc, nil
}

// isNRSCode returns true if the 3-character code looks like a National Rating
// System record. NRS records use a 3-letter uppercase federation code and have
// the same fixed-width player-line layout as 001 records (start number at
// bytes 4-7). We verify both the code format and a minimum line length with a
// numeric start number to avoid false positives on arbitrary unknown lines.
func isNRSCode(code, line string) bool {
	if len(code) != 3 {
		return false
	}
	for _, r := range code {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	// Exclude known XX-prefix codes.
	if strings.HasPrefix(code, "XX") {
		return false
	}
	// NRS lines follow the 001 column layout: at least ~68 chars with a
	// numeric start number at bytes 4-7 (cols 5-8).
	if len(line) < 68 {
		return false
	}
	_, err := strconv.Atoi(strings.TrimSpace(line[4:8]))
	return err == nil
}

// parsePlayerLine parses a single 001 player line using fixed-width columns.
// TRF16 column layout (1-indexed columns, 0-indexed bytes):
//
//	Col  1-3   (bytes  0-2):   "001"
//	Col  5-8   (bytes  4-7):   Starting rank (4 chars)
//	Col 10     (byte   9):     Sex (1 char)
//	Col 11-13  (bytes 10-12):  Title (3 chars)
//	Col 15-47  (bytes 14-46):  Name (33 chars)
//	Col 49-52  (bytes 48-51):  Rating (4 chars)
//	Col 54-56  (bytes 53-55):  Federation (3 chars)
//	Col 58-68  (bytes 57-67):  FIDE number (11 chars)
//	Col 70-79  (bytes 69-78):  Birth date (10 chars)
//	Col 81-84  (bytes 80-83):  Points (4 chars)
//	Col 86-89  (bytes 85-88):  Rank (4 chars)
//	Col 90+    (bytes 89+):    Round results (10 chars each)
func parsePlayerLine(line string, lineNum int) (PlayerLine, error) {
	if len(line) < 84 {
		return PlayerLine{}, &ParseError{
			Line:    lineNum,
			Code:    "001",
			Message: fmt.Sprintf("line too short (%d chars, need at least 84)", len(line)),
		}
	}

	var pl PlayerLine

	// Col 5-8: start number (bytes 4-7)
	sn, err := strconv.Atoi(strings.TrimSpace(line[4:8]))
	if err != nil {
		return PlayerLine{}, &ParseError{
			Line:    lineNum,
			Code:    "001",
			Message: fmt.Sprintf("invalid start number: %q", line[4:8]),
		}
	}
	pl.StartNumber = sn

	// Col 10: sex (byte 9)
	pl.Sex = strings.TrimSpace(string(line[9]))

	// Col 11-13: title (bytes 10-12)
	if len(line) > 12 {
		pl.Title = strings.TrimSpace(line[10:13])
	}

	// Col 15-47: name (bytes 14-46)
	if len(line) > 46 {
		pl.Name = strings.TrimSpace(line[14:47])
	} else if len(line) > 14 {
		pl.Name = strings.TrimSpace(line[14:])
	}

	// Col 49-52: rating (bytes 48-51)
	if len(line) > 51 {
		ratingStr := strings.TrimSpace(line[48:52])
		if ratingStr != "" {
			r, err := strconv.Atoi(ratingStr)
			if err != nil {
				return PlayerLine{}, &ParseError{
					Line:    lineNum,
					Code:    "001",
					Message: fmt.Sprintf("invalid rating: %q", ratingStr),
				}
			}
			pl.Rating = r
		}
	}

	// Col 54-56: federation (bytes 53-55)
	if len(line) > 55 {
		pl.Federation = strings.TrimSpace(line[53:56])
	}

	// Col 58-68: FIDE ID (bytes 57-67)
	if len(line) > 67 {
		pl.FideID = strings.TrimSpace(line[57:68])
	}

	// Col 70-79: birth date (bytes 69-78)
	if len(line) > 78 {
		pl.BirthDate = strings.TrimSpace(line[69:79])
	}

	// Col 81-84: points (bytes 80-83)
	pointsStr := strings.TrimSpace(line[80:84])
	if pointsStr != "" {
		pts, err := strconv.ParseFloat(pointsStr, 64)
		if err != nil {
			return PlayerLine{}, &ParseError{
				Line:    lineNum,
				Code:    "001",
				Message: fmt.Sprintf("invalid points: %q", pointsStr),
			}
		}
		pl.Points = pts
	}

	// Col 86-89: rank (bytes 85-88)
	if len(line) > 88 {
		rankStr := strings.TrimSpace(line[85:89])
		if rankStr != "" {
			rank, err := strconv.Atoi(rankStr)
			if err != nil {
				return PlayerLine{}, &ParseError{
					Line:    lineNum,
					Code:    "001",
					Message: fmt.Sprintf("invalid rank: %q", rankStr),
				}
			}
			pl.Rank = rank
		}
	}

	// Col 90+: round results (bytes 89+, 10 chars each)
	// Format per round: 2 spaces + 4-digit opponent + space + color + space + result
	if len(line) > 89 {
		roundData := line[89:]
		for i := 0; i+10 <= len(roundData); i += 10 {
			chunk := roundData[i : i+10]
			rr, err := parseRoundResult(chunk)
			if err != nil {
				return PlayerLine{}, &ParseError{
					Line:    lineNum,
					Code:    "001",
					Message: fmt.Sprintf("round %d: %v", len(pl.Rounds)+1, err),
				}
			}
			pl.Rounds = append(pl.Rounds, rr)
		}
	}

	return pl, nil
}

// parseRoundResult parses a 10-character round result chunk.
// Format: "  OOOO C R" where OOOO=opponent(4), C=color(1), R=result(1)
func parseRoundResult(chunk string) (RoundResult, error) {
	if len(chunk) < 10 {
		return RoundResult{}, fmt.Errorf("chunk too short: %q", chunk)
	}

	// Bytes 2-5: opponent start number
	oppStr := strings.TrimSpace(chunk[2:6])
	opp := 0
	if oppStr != "" {
		var err error
		opp, err = strconv.Atoi(oppStr)
		if err != nil {
			return RoundResult{}, fmt.Errorf("invalid opponent: %q", oppStr)
		}
	}

	// Byte 7: color
	color, ok := parseColorChar(chunk[7])
	if !ok {
		return RoundResult{}, fmt.Errorf("invalid color: %q", string(chunk[7]))
	}

	// Byte 9: result
	result, ok := parseResultChar(chunk[9])
	if !ok {
		return RoundResult{}, fmt.Errorf("invalid result: %q", string(chunk[9]))
	}

	return RoundResult{
		Opponent: opp,
		Color:    color,
		Result:   result,
	}, nil
}

// parseTeamLine parses a 013 team line.
// Format: "013" + 4-char team number + 32-char team name + member start numbers (4 chars each)
func parseTeamLine(line string, lineNum int) (TeamLine, error) {
	if len(line) < 40 {
		return TeamLine{}, &ParseError{
			Line:    lineNum,
			Code:    "013",
			Message: fmt.Sprintf("line too short (%d chars, need at least 40)", len(line)),
		}
	}

	var tl TeamLine

	// Team number: bytes 4-7
	tn, err := strconv.Atoi(strings.TrimSpace(line[4:8]))
	if err != nil {
		return TeamLine{}, &ParseError{
			Line:    lineNum,
			Code:    "013",
			Message: fmt.Sprintf("invalid team number: %q", line[4:8]),
		}
	}
	tl.TeamNumber = tn

	// Team name: bytes 8-40 (32 chars)
	if len(line) > 40 {
		tl.TeamName = strings.TrimSpace(line[8:40])
	} else {
		tl.TeamName = strings.TrimSpace(line[8:])
	}

	// Members: bytes 40+ (whitespace-separated start numbers)
	if len(line) > 40 {
		for _, s := range strings.Fields(line[40:]) {
			m, err := strconv.Atoi(s)
			if err != nil {
				return TeamLine{}, &ParseError{
					Line:    lineNum,
					Code:    "013",
					Message: fmt.Sprintf("invalid team member number: %q", s),
				}
			}
			tl.Members = append(tl.Members, m)
		}
	}

	return tl, nil
}

// parseForbiddenPair parses an XXP value "P1 P2".
func parseForbiddenPair(data string) (ForbiddenPair, error) {
	fields := strings.Fields(data)
	if len(fields) != 2 {
		return ForbiddenPair{}, fmt.Errorf("expected 2 player numbers, got %d", len(fields))
	}
	p1, err := strconv.Atoi(fields[0])
	if err != nil {
		return ForbiddenPair{}, fmt.Errorf("invalid player 1: %q", fields[0])
	}
	p2, err := strconv.Atoi(fields[1])
	if err != nil {
		return ForbiddenPair{}, fmt.Errorf("invalid player 2: %q", fields[1])
	}
	return ForbiddenPair{Player1: p1, Player2: p2}, nil
}

// --- TRF-2026 record parsers ---

// parseAbsenceRecord parses a 240 data string.
// Format: "T RRR TOI1 TOI2 ..."
func parseAbsenceRecord(data string, lineNum int) (AbsenceRecord, error) {
	fields := strings.Fields(data)
	if len(fields) < 2 {
		return AbsenceRecord{}, &ParseError{
			Line: lineNum, Code: "240",
			Message: fmt.Sprintf("too few fields: %q", data),
		}
	}
	absType := fields[0]
	if absType != "F" && absType != "H" {
		return AbsenceRecord{}, &ParseError{
			Line: lineNum, Code: "240",
			Message: fmt.Sprintf("invalid absence type %q (expected F or H)", absType),
		}
	}
	round, err := strconv.Atoi(fields[1])
	if err != nil {
		return AbsenceRecord{}, &ParseError{
			Line: lineNum, Code: "240",
			Message: fmt.Sprintf("invalid round: %q", fields[1]),
		}
	}
	var players []int
	for _, f := range fields[2:] {
		p, err := strconv.Atoi(f)
		if err != nil {
			return AbsenceRecord{}, &ParseError{
				Line: lineNum, Code: "240",
				Message: fmt.Sprintf("invalid player number: %q", f),
			}
		}
		players = append(players, p)
	}
	return AbsenceRecord{Type: absType, Round: round, Players: players}, nil
}

// parseAccelerationRecord parses a 250 data string.
// Format: "MMMM GGGG RRF RRL PPPF PPPL" (fields may be empty/spaces)
func parseAccelerationRecord(data string, lineNum int) (AccelerationRecord, error) {
	fields := strings.Fields(data)
	if len(fields) < 4 {
		return AccelerationRecord{}, &ParseError{
			Line: lineNum, Code: "250",
			Message: fmt.Sprintf("too few fields: %q", data),
		}
	}
	rec := AccelerationRecord{Raw: data}

	// Parse match points (MMMM).
	mp, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return AccelerationRecord{}, &ParseError{
			Line: lineNum, Code: "250",
			Message: fmt.Sprintf("invalid match points: %q", fields[0]),
		}
	}
	rec.MatchPoints = mp

	// Parse game points (GGGG) — may be empty.
	if len(fields) > 1 && fields[1] != "" {
		gp, err := strconv.ParseFloat(fields[1], 64)
		if err == nil {
			rec.GamePoints = gp
		}
	}

	// Parse first/last round.
	idx := 1
	if rec.GamePoints != 0 || (len(fields) > 1 && fields[1] != fields[0]) {
		idx = 2
	}
	// Flexible parsing: the remaining fields are RRF RRL PPPF PPPL.
	remaining := fields[idx:]
	if len(remaining) >= 1 {
		if v, err := strconv.Atoi(remaining[0]); err == nil {
			rec.FirstRound = v
		}
	}
	if len(remaining) >= 2 {
		if v, err := strconv.Atoi(remaining[1]); err == nil {
			rec.LastRound = v
		}
	}
	if len(remaining) >= 3 {
		if v, err := strconv.Atoi(remaining[2]); err == nil {
			rec.FirstPlayer = v
		}
	}
	if len(remaining) >= 4 {
		if v, err := strconv.Atoi(remaining[3]); err == nil {
			rec.LastPlayer = v
		}
	}

	return rec, nil
}

// parseForbiddenPairRecord parses a 260 data string.
// Format: "RR1 RRL TOI1 TOI2 ..."
func parseForbiddenPairRecord(data string, lineNum int) (ForbiddenPairRecord, error) {
	fields := strings.Fields(data)
	if len(fields) < 3 {
		return ForbiddenPairRecord{}, &ParseError{
			Line: lineNum, Code: "260",
			Message: fmt.Sprintf("too few fields: %q", data),
		}
	}
	rec := ForbiddenPairRecord{Raw: data}

	rr1, err := strconv.Atoi(fields[0])
	if err != nil {
		return ForbiddenPairRecord{}, &ParseError{
			Line: lineNum, Code: "260",
			Message: fmt.Sprintf("invalid first round: %q", fields[0]),
		}
	}
	rec.FirstRound = rr1

	rrl, err := strconv.Atoi(fields[1])
	if err != nil {
		return ForbiddenPairRecord{}, &ParseError{
			Line: lineNum, Code: "260",
			Message: fmt.Sprintf("invalid last round: %q", fields[1]),
		}
	}
	rec.LastRound = rrl

	for _, f := range fields[2:] {
		p, err := strconv.Atoi(f)
		if err != nil {
			return ForbiddenPairRecord{}, &ParseError{
				Line: lineNum, Code: "260",
				Message: fmt.Sprintf("invalid player number: %q", f),
			}
		}
		rec.Players = append(rec.Players, p)
	}
	return rec, nil
}

// parseTeamRoundEntry parses a 300 data string.
// Format: "RRR TT1 TT2 PPP1 PPP2 PPP3 PPP4"
func parseTeamRoundEntry(data string, lineNum int) (TeamRoundEntry, error) {
	fields := strings.Fields(data)
	if len(fields) < 3 {
		return TeamRoundEntry{}, &ParseError{
			Line: lineNum, Code: "300",
			Message: fmt.Sprintf("too few fields: %q", data),
		}
	}
	round, err := strconv.Atoi(fields[0])
	if err != nil {
		return TeamRoundEntry{}, &ParseError{
			Line: lineNum, Code: "300",
			Message: fmt.Sprintf("invalid round: %q", fields[0]),
		}
	}
	t1, err := strconv.Atoi(fields[1])
	if err != nil {
		return TeamRoundEntry{}, &ParseError{
			Line: lineNum, Code: "300",
			Message: fmt.Sprintf("invalid team1: %q", fields[1]),
		}
	}
	t2, err := strconv.Atoi(fields[2])
	if err != nil {
		return TeamRoundEntry{}, &ParseError{
			Line: lineNum, Code: "300",
			Message: fmt.Sprintf("invalid team2: %q", fields[2]),
		}
	}
	var boards []int
	for _, f := range fields[3:] {
		b, err := strconv.Atoi(f)
		if err != nil {
			return TeamRoundEntry{}, &ParseError{
				Line: lineNum, Code: "300",
				Message: fmt.Sprintf("invalid board player: %q", f),
			}
		}
		boards = append(boards, b)
	}
	return TeamRoundEntry{Round: round, Team1: t1, Team2: t2, Boards: boards}, nil
}

// parseNewTeamLine parses a 310 line using fixed-width columns.
// Format: "310 SSS NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN FFFFF EEEEEE MMMMMM GGGGGG RRR  PPP1 PPP2..."
// Columns (0-indexed bytes):
//
//	[0:3]   "310"
//	[4:7]   team number (3 chars)
//	[8:40]  team name (32 chars)
//	[41:46] federation (5 chars)
//	[47:53] avg rating (6 chars)
//	[54:60] match points (6 chars)
//	[61:67] game points (6 chars)
//	[68:71] rank (3 chars)
//	[73:]   members (4 chars each)
func parseNewTeamLine(line string, lineNum int) (NewTeamLine, error) {
	if len(line) < 8 {
		return NewTeamLine{}, &ParseError{
			Line: lineNum, Code: "310",
			Message: fmt.Sprintf("line too short (%d chars, need at least 8)", len(line)),
		}
	}
	var tl NewTeamLine

	// Team number: bytes 4-6 (3 chars, right-justified).
	tn, err := strconv.Atoi(strings.TrimSpace(line[4:7]))
	if err != nil {
		return NewTeamLine{}, &ParseError{
			Line: lineNum, Code: "310",
			Message: fmt.Sprintf("invalid team number: %q", line[4:7]),
		}
	}
	tl.TeamNumber = tn

	// Team name: bytes 8-40 (33 chars, left-justified).
	if len(line) <= 41 {
		tl.TeamName = strings.TrimSpace(line[8:])
		return tl, nil
	}
	tl.TeamName = strings.TrimSpace(line[8:41])

	// Federation: bytes 41-45 (5 chars, left-justified).
	end := min(len(line), 46)
	tl.Federation = strings.TrimSpace(line[41:end])

	// Remaining numeric fields: avg rating, match points, game points,
	// rank, members. Parsed as whitespace-separated tokens for robustness
	// against varying column widths across tournaments.
	if len(line) <= 46 {
		return tl, nil
	}
	fields := strings.Fields(line[46:])
	if len(fields) == 0 {
		return tl, nil
	}

	// Field 0: average rating.
	if r, err := strconv.ParseFloat(fields[0], 64); err == nil {
		tl.AvgRating = r
	}

	// Field 1: match points.
	if len(fields) > 1 {
		if mp, err := strconv.ParseFloat(fields[1], 64); err == nil {
			tl.MatchPoints = mp
		}
	}

	// Field 2: game points.
	if len(fields) > 2 {
		if gp, err := strconv.ParseFloat(fields[2], 64); err == nil {
			tl.GamePoints = gp
		}
	}

	// Field 3: rank.
	if len(fields) > 3 {
		if rank, err := strconv.Atoi(fields[3]); err == nil {
			tl.Rank = rank
		}
	}

	// Fields 4+: member start numbers.
	for i := 4; i < len(fields); i++ {
		if m, err := strconv.Atoi(fields[i]); err == nil {
			tl.Members = append(tl.Members, m)
		}
	}

	return tl, nil
}

// parseTeamRoundScoreEntry parses a 320 data string.
// Format: "TTT GGGG RRR1 RRR2 ..." — store raw for round-trip.
func parseTeamRoundScoreEntry(data string) TeamRoundScoreEntry {
	rec := TeamRoundScoreEntry{Raw: data}
	fields := strings.Fields(data)
	if len(fields) >= 1 {
		if tn, err := strconv.Atoi(fields[0]); err == nil {
			rec.TeamNumber = tn
		}
	}
	if len(fields) >= 2 {
		if gp, err := strconv.ParseFloat(fields[1], 64); err == nil {
			rec.GamePoints = gp
		}
	}
	if len(fields) > 2 {
		rec.Scores = fields[2:]
	}
	return rec
}

// parseOldAbsentForfeit parses a 330 data string.
// Format: "TT RRR WWW BBB"
func parseOldAbsentForfeit(data string, lineNum int) (OldAbsentForfeit, error) {
	fields := strings.Fields(data)
	if len(fields) < 4 {
		return OldAbsentForfeit{}, &ParseError{
			Line: lineNum, Code: "330",
			Message: fmt.Sprintf("too few fields: %q", data),
		}
	}
	round, err := strconv.Atoi(fields[1])
	if err != nil {
		return OldAbsentForfeit{}, &ParseError{
			Line: lineNum, Code: "330",
			Message: fmt.Sprintf("invalid round: %q", fields[1]),
		}
	}
	wt, err := strconv.Atoi(fields[2])
	if err != nil {
		return OldAbsentForfeit{}, &ParseError{
			Line: lineNum, Code: "330",
			Message: fmt.Sprintf("invalid white team: %q", fields[2]),
		}
	}
	bt, err := strconv.Atoi(fields[3])
	if err != nil {
		return OldAbsentForfeit{}, &ParseError{
			Line: lineNum, Code: "330",
			Message: fmt.Sprintf("invalid black team: %q", fields[3]),
		}
	}
	return OldAbsentForfeit{
		ResultType: fields[0],
		Round:      round,
		WhiteTeam:  wt,
		BlackTeam:  bt,
	}, nil
}

// parseDetailedTeamResult parses an 801 data string. Since 801 lines have
// complex per-round data with variable formatting, we store the raw data and
// parse header fields only.
func parseDetailedTeamResult(data string) DetailedTeamResult {
	rec := DetailedTeamResult{Raw: data}
	fields := strings.Fields(data)
	if len(fields) >= 1 {
		if tn, err := strconv.Atoi(fields[0]); err == nil {
			rec.TeamNumber = tn
		}
	}
	if len(fields) >= 2 {
		rec.TeamName = fields[1]
	}
	if len(fields) >= 3 {
		if mp, err := strconv.ParseFloat(fields[2], 64); err == nil {
			rec.MatchPoints = mp
		}
	}
	if len(fields) >= 4 {
		if gp, err := strconv.ParseFloat(fields[3], 64); err == nil {
			rec.GamePoints = gp
		}
	}
	// Per-round data is complex (variable width with bye markers). Parse
	// structurally: each round entry is 4 fields (opponent color results boardorder)
	// or a bye marker like "FFFF", "HHHH", "ZZZZ", "UUUU".
	idx := 4
	for idx < len(fields) {
		var dr DetailedTeamRound
		token := fields[idx]
		// Check for bye markers.
		if token == "FFFF" || token == "HHHH" || token == "ZZZZ" || token == "UUUU" {
			dr.ByeType = token
			idx++
			rec.Rounds = append(rec.Rounds, dr)
			continue
		}
		// Normal round: opponent color results boardorder.
		if opp, err := strconv.Atoi(token); err == nil {
			dr.Opponent = opp
		}
		if idx+1 < len(fields) {
			dr.Color = fields[idx+1]
		}
		if idx+2 < len(fields) {
			dr.Results = fields[idx+2]
		}
		if idx+3 < len(fields) {
			dr.BoardOrder = fields[idx+3]
		}
		idx += 4
		rec.Rounds = append(rec.Rounds, dr)
	}
	return rec
}

// parseSimpleTeamResult parses an 802 data string. Since 802 lines have
// variable formatting with bye markers, we parse structurally.
func parseSimpleTeamResult(data string) SimpleTeamResult {
	rec := SimpleTeamResult{Raw: data}
	fields := strings.Fields(data)
	if len(fields) >= 1 {
		if tn, err := strconv.Atoi(fields[0]); err == nil {
			rec.TeamNumber = tn
		}
	}
	if len(fields) >= 2 {
		rec.TeamName = fields[1]
	}
	if len(fields) >= 3 {
		if mp, err := strconv.ParseFloat(fields[2], 64); err == nil {
			rec.MatchPoints = mp
		}
	}
	if len(fields) >= 4 {
		if gp, err := strconv.ParseFloat(fields[3], 64); err == nil {
			rec.GamePoints = gp
		}
	}
	// Per-round data: each entry is either:
	// - "TT C GGGG" or "TT C GGGGf" (opponent color gamepoints [forfeit])
	// - "FPB GGGG" / "HPB GGGG" / "ZPB GGGG" / "PAB GGGG" (bye with game points)
	idx := 4
	for idx < len(fields) {
		var sr SimpleTeamRound
		token := fields[idx]
		// Check for bye markers.
		if token == "FPB" || token == "HPB" || token == "ZPB" || token == "PAB" {
			sr.ByeType = token
			if idx+1 < len(fields) {
				gpStr := fields[idx+1]
				if gp, err := strconv.ParseFloat(gpStr, 64); err == nil {
					sr.GamePoints = gp
				}
			}
			idx += 2
			rec.Rounds = append(rec.Rounds, sr)
			continue
		}
		// Normal round: opponent color gamepoints[f].
		if opp, err := strconv.Atoi(token); err == nil {
			sr.Opponent = opp
		}
		if idx+1 < len(fields) {
			sr.Color = fields[idx+1]
		}
		if idx+2 < len(fields) {
			gpStr := fields[idx+2]
			if strings.HasSuffix(gpStr, "f") {
				sr.Forfeit = true
				gpStr = gpStr[:len(gpStr)-1]
			}
			if gp, err := strconv.ParseFloat(gpStr, 64); err == nil {
				sr.GamePoints = gp
			}
		}
		idx += 3
		rec.Rounds = append(rec.Rounds, sr)
	}
	return rec
}

// parseNRSRecord parses a National Rating System line. The format is similar
// to a 001 line but starts with a 3-letter federation code. We store the raw
// line and parse basic fields.
func parseNRSRecord(federation string, line string) NRSRecord {
	rec := NRSRecord{
		Federation: federation,
		Raw:        line,
	}
	// Parse using same column positions as 001 (start number at bytes 4-7).
	if len(line) >= 8 {
		if sn, err := strconv.Atoi(strings.TrimSpace(line[4:8])); err == nil {
			rec.StartNumber = sn
		}
	}
	if len(line) >= 13 {
		rec.Title = strings.TrimSpace(line[10:13])
	}
	if len(line) >= 47 {
		rec.Name = strings.TrimSpace(line[14:47])
	}
	if len(line) >= 52 {
		ratingStr := strings.TrimSpace(line[48:52])
		if ratingStr != "" {
			if r, err := strconv.Atoi(ratingStr); err == nil {
				rec.NationalRating = r
			}
		}
	}
	if len(line) >= 56 {
		rec.SubFederation = strings.TrimSpace(line[53:56])
	}
	if len(line) >= 68 {
		rec.NationalID = strings.TrimSpace(line[57:68])
	}
	if len(line) >= 79 {
		rec.BirthDate = strings.TrimSpace(line[69:79])
	}
	return rec
}
