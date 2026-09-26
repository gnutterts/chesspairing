// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// auditPlayer is a literal 001 layout with one round entry at bytes 89-98.
func auditPlayer(result byte) string {
	line := bytes.Repeat([]byte(" "), 99)
	copy(line[0:3], "001")
	copy(line[4:8], "   1")
	copy(line[14:47], "Audit, Player")
	copy(line[80:84], " 0.0")
	copy(line[85:89], "   1")
	copy(line[91:95], "0000")
	line[96] = '-'
	line[98] = result
	return string(line)
}

func TestResultCodes2026(t *testing.T) {
	// TRF-2026, Player Section, result column: the quoted symbols are "1",
	// "0", "=", "+", "-", "W", "D", "L", "H", "F", "U", "Z"; "(blank)
	// equivalent to Z" and "Letter codes are case-insensitive". F/U are
	// excluded here: their meaning changes in a later PR (D22), so these
	// test cases are deliberately withheld for now.
	uppercase := map[byte]ResultCode{
		'1': ResultWin,
		'0': ResultLoss,
		'=': ResultDraw,
		'+': ResultForfeitWin,
		'-': ResultForfeitLoss,
		'W': ResultWinByDefault,
		'D': ResultDrawByDefault,
		'L': ResultLossByDefault,
		'H': ResultHalfBye,
		'Z': ResultZeroBye,
	}
	for code, want := range uppercase {
		t.Run("upper-"+string(code), func(t *testing.T) {
			doc, err := Read(strings.NewReader(auditPlayer(code) + "\n"))
			if err != nil {
				t.Fatalf("Read(%q): %v", code, err)
			}
			if got := doc.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("Read(%q).Result = %v, want %v", code, got, want)
			}

			var out strings.Builder
			if err := Write(&out, doc); err != nil {
				t.Fatalf("Write(%q): %v", code, err)
			}
			doc2, err := Read(strings.NewReader(out.String()))
			if err != nil {
				t.Fatalf("re-Read(%q): %v", code, err)
			}
			if got := doc2.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("re-Read(%q).Result = %v, want %v", code, got, want)
			}
		})
	}

	lowercase := map[byte]ResultCode{
		'w': ResultWinByDefault,
		'd': ResultDrawByDefault,
		'l': ResultLossByDefault,
		'h': ResultHalfBye,
		'z': ResultZeroBye,
	}
	for code, want := range lowercase {
		t.Run("lower-"+string(code), func(t *testing.T) {
			doc, err := Read(strings.NewReader(auditPlayer(code) + "\n"))
			if err != nil {
				t.Fatalf("Read(%q): %v", code, err)
			}
			if got := doc.Players[0].Rounds[0].Result; got != want {
				t.Fatalf("Read(%q).Result = %v, want %v", code, got, want)
			}
		})
	}

	// A blank result with opponent 0000 is equivalent to Z.
	t.Run("blank", func(t *testing.T) {
		doc, err := Read(strings.NewReader(auditPlayer(' ') + "\n"))
		if err != nil {
			t.Fatalf("Read(blank): %v", err)
		}
		r := doc.Players[0].Rounds[0]
		if r.Opponent != 0 || r.Color != ColorNone || r.Result != ResultZeroBye {
			t.Fatalf("Read(blank) = %+v, want {0 None ZeroBye}", r)
		}
	})
}

func TestRecords2026(t *testing.T) {
	// Literal TRF-2026 records: Read must parse every record type and a
	// Write+Read round-trip must preserve the parsed values.
	input := strings.Join([]string{
		"012 Audit", "022 City", "032 NED", "042 2026/01/01", "052 2026/01/02", "062 1", "072 1", "092 Swiss Dutch", "102 Chief", "112 Deputy", "122 90min", "132 26/01/01",
		"142 7", "152 B", "162 W 1.0    D 0.5    L 0.0", "172 NED FIDE", "182 controller", "192 FIDE_DUTCH_2025", "202 BH", "212 PTS,BH", "222 5400+30", "352 WBWB", "362 TW 2.0   TD 1.0   TL 0.0",
		auditPlayer('Z'),
		"NED    1      Audit, Player                   2000 NED             2000/01/01",
		"013    1 Audit Team                      1",
		"240 H 003 026 047", "250 00.0 02.0 001 003 0001 0090", "260 001 002 125 180 184 216", "299 +    2.0      2.5", "300 008 021 047 0058 0203 0105 0162",
		"310   1 India                            IND     2486   15.0   28.0 11                            1    5   15   28   44",
		"320 01.0 02.0 000 000 050 049 000 046 048 045 000 036 043", "330 +- 004 023 047", "330 -- 004 023 047",
		"801 03 GEO   19 32.5       FFFF       16 w 11=0 1254", "802   3 GEO     19.0   32.5          FPB    4.0         16 w 2.5",
		"XXR 7", "XXC B", "XXS acceleration", "XXP 1 2", "XXY 2", "XXB true", "XXM false", "XXT A", "XXG match", "XXA true", "XXK 1",
	}, "\n") + "\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read literal records: %v", err)
	}

	// TRF16 headers.
	if doc.Name != "Audit" || doc.City != "City" || doc.Federation != "NED" {
		t.Errorf("012/022/032 = %q/%q/%q, want Audit/City/NED", doc.Name, doc.City, doc.Federation)
	}
	if doc.NumPlayers != 1 || doc.NumRated != 1 {
		t.Errorf("062/072 = %d/%d, want 1/1", doc.NumPlayers, doc.NumRated)
	}
	if doc.TournamentType != "Swiss Dutch" || doc.ChiefArbiter != "Chief" || doc.DeputyArbiter != "Deputy" {
		t.Errorf("092/102/112 = %q/%q/%q", doc.TournamentType, doc.ChiefArbiter, doc.DeputyArbiter)
	}
	if doc.StartDate != "2026/01/01" || doc.EndDate != "2026/01/02" || doc.TimeControl != "90min" {
		t.Errorf("042/052/122 = %q/%q/%q", doc.StartDate, doc.EndDate, doc.TimeControl)
	}
	if len(doc.RoundDates) != 1 || doc.RoundDates[0] != "26/01/01" {
		t.Errorf("132 = %v, want [26/01/01]", doc.RoundDates)
	}

	// TRF-2026 header records.
	if doc.TotalRounds26 != 7 || doc.InitialColor26 != "B" {
		t.Errorf("142/152 = %d/%q, want 7/B", doc.TotalRounds26, doc.InitialColor26)
	}
	if doc.ScoringSystem == nil || doc.ScoringSystem.W == nil || *doc.ScoringSystem.W != 1.0 ||
		doc.ScoringSystem.D == nil || *doc.ScoringSystem.D != 0.5 ||
		doc.ScoringSystem.L == nil || *doc.ScoringSystem.L != 0.0 {
		t.Errorf("162 = %+v, want W=1.0 D=0.5 L=0.0", doc.ScoringSystem)
	}
	if doc.StartingRankMethod != "NED FIDE" || doc.CodedTournamentType != "FIDE_DUTCH_2025" {
		t.Errorf("172/192 = %q/%q, want NED FIDE/FIDE_DUTCH_2025", doc.StartingRankMethod, doc.CodedTournamentType)
	}
	if doc.TieBreakDef != "BH" || doc.EncodedTimeControl != "5400+30" {
		t.Errorf("202/222 = %q/%q, want BH/5400+30", doc.TieBreakDef, doc.EncodedTimeControl)
	}
	if doc.TeamInitialColor != "WBWB" || doc.TeamScoringSystem != "TW 2.0   TD 1.0   TL 0.0" {
		t.Errorf("352/362 = %q/%q", doc.TeamInitialColor, doc.TeamScoringSystem)
	}

	// TRF-2026 data records.
	if len(doc.ForbiddenPairs26) != 1 {
		t.Fatalf("ForbiddenPairs26 count = %d, want 1", len(doc.ForbiddenPairs26))
	}
	fp := doc.ForbiddenPairs26[0]
	if fp.FirstRound != 1 || fp.LastRound != 2 || fmt.Sprint(fp.Players) != fmt.Sprint([]int{125, 180, 184, 216}) {
		t.Errorf("260 = %+v, want FirstRound=1 LastRound=2 Players=[125 180 184 216]", fp)
	}
	if fp.Raw != "001 002 125 180 184 216" {
		t.Errorf("260.Raw = %q, want %q", fp.Raw, "001 002 125 180 184 216")
	}
	if len(doc.Absences) != 1 || doc.Absences[0].Type != "H" || doc.Absences[0].Round != 3 ||
		fmt.Sprint(doc.Absences[0].Players) != fmt.Sprint([]int{26, 47}) {
		t.Errorf("240 = %+v, want H round 3 players [26 47]", doc.Absences)
	}
	if len(doc.Accelerations26) != 1 || doc.Accelerations26[0].FirstRound != 1 || doc.Accelerations26[0].LastRound != 3 {
		t.Errorf("250 = %+v, want FirstRound=1 LastRound=3", doc.Accelerations26)
	}
	if len(doc.TeamRoundData) != 1 || doc.TeamRoundData[0].Round != 8 || len(doc.TeamRoundData[0].Boards) != 4 {
		t.Errorf("300 = %+v, want round 8 with 4 boards", doc.TeamRoundData)
	}
	if len(doc.NewTeams) != 1 || doc.NewTeams[0].TeamNumber != 1 || doc.NewTeams[0].TeamName != "India" {
		t.Errorf("310 = %+v, want team 1 India", doc.NewTeams)
	}
	if len(doc.TeamRoundScores) != 1 || doc.TeamRoundScores[0].Raw != "01.0 02.0 000 000 050 049 000 046 048 045 000 036 043" {
		t.Errorf("320 = %+v", doc.TeamRoundScores)
	}
	if len(doc.OldAbsentForfeits) != 2 {
		t.Errorf("330 count = %d, want 2", len(doc.OldAbsentForfeits))
	}
	if len(doc.DetailedTeamResults) != 1 || len(doc.SimpleTeamResults) != 1 {
		t.Errorf("801/802 counts = %d/%d, want 1/1", len(doc.DetailedTeamResults), len(doc.SimpleTeamResults))
	}
	if len(doc.Other) != 3 {
		t.Errorf("Other count = %d, want 3 (182, 212, 299)", len(doc.Other))
	}

	// Legacy extension records.
	if doc.TotalRounds != 7 || doc.InitialColor != "B" {
		t.Errorf("XXR/XXC = %d/%q, want 7/B", doc.TotalRounds, doc.InitialColor)
	}
	if len(doc.ForbiddenPairs) != 1 || doc.ForbiddenPairs[0] != (ForbiddenPair{Player1: 1, Player2: 2}) {
		t.Errorf("XXP = %+v, want [{1 2}]", doc.ForbiddenPairs)
	}
	if len(doc.Acceleration) != 1 || doc.Acceleration[0] != "acceleration" {
		t.Errorf("XXS = %v, want [acceleration]", doc.Acceleration)
	}
	if doc.Cycles != 2 || doc.ColorBalance == nil || !*doc.ColorBalance {
		t.Errorf("XXY/XXB = %d/%v, want 2/true", doc.Cycles, doc.ColorBalance)
	}
	if doc.MaxiTournament == nil || *doc.MaxiTournament {
		t.Errorf("XXM = %v, want false", doc.MaxiTournament)
	}
	if doc.ColorPreferenceType != "A" || doc.PrimaryScore != "match" {
		t.Errorf("XXT/XXG = %q/%q, want A/match", doc.ColorPreferenceType, doc.PrimaryScore)
	}
	if doc.AllowRepeatPairings == nil || !*doc.AllowRepeatPairings || doc.MinRoundsBetweenRepeats != 1 {
		t.Errorf("XXA/XXK = %v/%d, want true/1", doc.AllowRepeatPairings, doc.MinRoundsBetweenRepeats)
	}

	// Write+Read must preserve the same parsed values.
	var out strings.Builder
	if err := Write(&out, doc); err != nil {
		t.Fatalf("Write literal records: %v", err)
	}
	doc2, err := Read(strings.NewReader(out.String()))
	if err != nil {
		t.Fatalf("re-Read literal records: %v", err)
	}
	if doc2.ScoringSystem == nil || !scoringPointsEqual(doc.ScoringSystem, doc2.ScoringSystem) {
		t.Errorf("162 round-trip = %+v, want %+v", doc2.ScoringSystem, doc.ScoringSystem)
	}
	if len(doc2.ForbiddenPairs26) != 1 {
		t.Fatalf("260 round-trip count = %d, want 1", len(doc2.ForbiddenPairs26))
	}
	fp2 := doc2.ForbiddenPairs26[0]
	if fp2.FirstRound != fp.FirstRound || fp2.LastRound != fp.LastRound ||
		fmt.Sprint(fp2.Players) != fmt.Sprint(fp.Players) || fp2.Raw != fp.Raw {
		t.Errorf("260 round-trip = %+v, want %+v", fp2, fp)
	}
}

func TestOracleRoundTrip(t *testing.T) {
	lf := strings.Join([]string{
		"012 Audit",
		"022 City",
		"092 Swiss Dutch",
		"142 3",
		"152 W",
		"XXR 3",
		"XXC B",
		auditPlayer('1'),
		auditPlayer('0'),
	}, "\n") + "\n"

	tests := []struct {
		name  string
		input string
	}{
		{"lf", lf},
		{"cr", strings.ReplaceAll(lf, "\n", "\r")},
		{"crlf", strings.ReplaceAll(lf, "\n", "\r\n")},
		{"noeol", strings.TrimSuffix(lf, "\n")},
	}

	var want string
	for i, tt := range tests {
		doc, err := Read(strings.NewReader(tt.input))
		if err != nil {
			t.Fatalf("Read(%s): %v", tt.name, err)
		}
		var out strings.Builder
		if err := Write(&out, doc); err != nil {
			t.Fatalf("Write(%s): %v", tt.name, err)
		}
		if i == 0 {
			want = out.String()
			continue
		}
		if out.String() != want {
			t.Errorf("%s round-trip output differs from LF-normalized output", tt.name)
		}
	}
}

func TestToTournamentState_forbiddenPairRecordRoundRange(t *testing.T) {
	players := func(rounds int) []PlayerLine {
		out := make([]PlayerLine, 5)
		for i := range out {
			out[i] = PlayerLine{StartNumber: i + 1, Name: "Audit", Rating: 2000 - i}
			for r := 0; r < rounds; r++ {
				out[i].Rounds = append(out[i].Rounds, RoundResult{Opponent: 0, Color: ColorNone, Result: ResultFullBye})
			}
		}
		return out
	}

	doc := &Document{
		ForbiddenPairs:   []ForbiddenPair{{Player1: 4, Player2: 5}},
		ForbiddenPairs26: []ForbiddenPairRecord{{FirstRound: 1, LastRound: 2, Players: []int{1, 2, 3}}},
		Players:          players(1),
	}

	// The round to be paired is round 2, inside 1..2: the 260 record
	// contributes its pairs and the legacy XXP pair stays permanent.
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState (round 2): %v", err)
	}
	if state.CurrentRound != 2 {
		t.Fatalf("CurrentRound = %d, want 2", state.CurrentRound)
	}
	assertForbiddenPairs(t, state, [][2]int{{4, 5}, {1, 2}, {1, 3}, {2, 3}})

	// The round to be paired is round 3, outside 1..2: only the permanent
	// legacy XXP pair remains.
	doc.Players = players(2)
	state, err = doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState (round 3): %v", err)
	}
	if state.CurrentRound != 3 {
		t.Fatalf("CurrentRound = %d, want 3", state.CurrentRound)
	}
	assertForbiddenPairs(t, state, [][2]int{{4, 5}})
}

func TestToTournamentState_scoringPointsDriveStandardScorer(t *testing.T) {
	win, draw, loss := 3.0, 1.0, 0.0
	doc := &Document{
		ScoringSystem: &ScoringPoints{W: &win, D: &draw, L: &loss},
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Audit A", Rating: 2000},
			{StartNumber: 2, Name: "Audit B", Rating: 1800},
		},
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if got := state.ScoringConfig.Options["pointWin"]; got != 3.0 {
		t.Errorf("pointWin = %v, want 3.0", got)
	}
	if got := state.ScoringConfig.Options["pointDraw"]; got != 1.0 {
		t.Errorf("pointDraw = %v, want 1.0", got)
	}
	if got := state.ScoringConfig.Options["pointLoss"]; got != 0.0 {
		t.Errorf("pointLoss = %v, want 0.0", got)
	}

	scorer := standard.NewFromMap(state.ScoringConfig.Options)
	if got := scorer.PointsForResult(chesspairing.ResultWhiteWins, chesspairing.ResultContext{}); got != 3.0 {
		t.Errorf("PointsForResult(win) = %v, want 3.0", got)
	}
	if got := scorer.PointsForResult(chesspairing.ResultDraw, chesspairing.ResultContext{}); got != 1.0 {
		t.Errorf("PointsForResult(draw) = %v, want 1.0", got)
	}
}

func TestReadWrite_scoringPointsSpacing(t *testing.T) {
	input := "162 W 1.0    D 0.5    L 0.0\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	var out strings.Builder
	if err := Write(&out, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if got := out.String(); got != input {
		t.Errorf("162 round-trip = %q, want %q", got, input)
	}
}

func assertForbiddenPairs(t *testing.T, state *chesspairing.TournamentState, want [][2]int) {
	t.Helper()
	raw, ok := state.PairingConfig.Options["forbiddenPairs"]
	if !ok {
		if len(want) == 0 {
			return
		}
		t.Fatalf("forbiddenPairs option missing, want %v", want)
	}
	got, ok := raw.([][2]int)
	if !ok {
		t.Fatalf("forbiddenPairs = %T %v, want [][2]int", raw, raw)
	}
	if len(got) != len(want) {
		t.Fatalf("forbiddenPairs = %v, want %v", got, want)
	}
	set := make(map[[2]int]bool, len(got))
	for _, p := range got {
		set[p] = true
	}
	for _, p := range want {
		if !set[p] {
			t.Errorf("forbiddenPairs = %v, missing %v", got, p)
		}
	}
}
