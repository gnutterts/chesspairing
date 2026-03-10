package trf

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Read parses a TRF16 file from the reader.
func Read(r io.Reader) (*Document, error) {
	doc := &Document{}
	scanner := bufio.NewScanner(r)
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
			doc.DeputyArbiter = data
		case "122":
			doc.TimeControl = data
		case "132":
			doc.RoundDates = append(doc.RoundDates, data)
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
		default:
			doc.Other = append(doc.Other, RawLine{Code: code, Data: data})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("trf: read error: %w", err)
	}

	return doc, nil
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
