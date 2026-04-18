// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestColor_IsValid(t *testing.T) {
	valid := []Color{ColorNone, ColorWhite, ColorBlack}
	for _, c := range valid {
		if !c.IsValid() {
			t.Errorf("IsValid(%v) = false, want true", c)
		}
	}
	if Color(-1).IsValid() {
		t.Error("IsValid(-1) = true, want false")
	}
	if Color(3).IsValid() {
		t.Error("IsValid(3) = true, want false")
	}
}

func TestColor_String(t *testing.T) {
	tests := []struct {
		c    Color
		want string
	}{
		{ColorNone, "None"},
		{ColorWhite, "White"},
		{ColorBlack, "Black"},
		{Color(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.c.String(); got != tt.want {
			t.Errorf("String(%d) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestColor_Char(t *testing.T) {
	tests := []struct {
		c    Color
		want byte
	}{
		{ColorNone, '-'},
		{ColorWhite, 'w'},
		{ColorBlack, 'b'},
	}
	for _, tt := range tests {
		if got := tt.c.Char(); got != tt.want {
			t.Errorf("Char(%d) = %q, want %q", tt.c, got, tt.want)
		}
	}
}

func TestResultCode_String(t *testing.T) {
	tests := []struct {
		rc   ResultCode
		want string
	}{
		{ResultWin, "Win"},
		{ResultLoss, "Loss"},
		{ResultDraw, "Draw"},
		{ResultForfeitWin, "ForfeitWin"},
		{ResultForfeitLoss, "ForfeitLoss"},
		{ResultHalfBye, "HalfBye"},
		{ResultFullBye, "FullBye"},
		{ResultUnpaired, "Unpaired"},
		{ResultZeroBye, "ZeroBye"},
		{ResultNotPlayed, "NotPlayed"},
		{ResultWinByDefault, "WinByDefault"},
		{ResultDrawByDefault, "DrawByDefault"},
		{ResultLossByDefault, "LossByDefault"},
		{ResultCode(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.rc.String(); got != tt.want {
			t.Errorf("String(%d) = %q, want %q", tt.rc, got, tt.want)
		}
	}
}

func TestResultCode_Char_default(t *testing.T) {
	if got := ResultCode(99).Char(); got != '?' {
		t.Errorf("Char(99) = %q, want '?'", got)
	}
}

func TestResultCode_IsValid(t *testing.T) {
	for rc := ResultWin; rc <= ResultLossByDefault; rc++ {
		if !rc.IsValid() {
			t.Errorf("IsValid(%v) = false, want true", rc)
		}
	}
	if ResultCode(-1).IsValid() {
		t.Error("IsValid(-1) = true, want false")
	}
	if ResultCode(13).IsValid() {
		t.Error("IsValid(13) = true, want false")
	}
}

func TestResultCode_Char_roundtrip(t *testing.T) {
	// Every valid ResultCode should round-trip through Char() and parseResultChar().
	for rc := ResultWin; rc <= ResultLossByDefault; rc++ {
		ch := rc.Char()
		got, ok := parseResultChar(ch)
		if !ok {
			t.Errorf("parseResultChar(%q) failed for %v", ch, rc)
			continue
		}
		if got != rc {
			t.Errorf("parseResultChar(%q) = %v, want %v", ch, got, rc)
		}
	}
}

func TestColor_Char_roundtrip(t *testing.T) {
	for _, c := range []Color{ColorNone, ColorWhite, ColorBlack} {
		ch := c.Char()
		got, ok := parseColorChar(ch)
		if !ok {
			t.Errorf("parseColorChar(%q) failed for %v", ch, c)
			continue
		}
		if got != c {
			t.Errorf("parseColorChar(%q) = %v, want %v", ch, got, c)
		}
	}
}

func TestResultCode_isByeResult(t *testing.T) {
	byes := []ResultCode{ResultHalfBye, ResultFullBye, ResultUnpaired, ResultZeroBye}
	for _, rc := range byes {
		if !rc.isByeResult() {
			t.Errorf("isByeResult(%v) = false, want true", rc)
		}
	}
	nonByes := []ResultCode{
		ResultWin, ResultLoss, ResultDraw, ResultForfeitWin,
		ResultForfeitLoss, ResultNotPlayed, ResultWinByDefault, ResultDrawByDefault, ResultLossByDefault,
	}
	for _, rc := range nonByes {
		if rc.isByeResult() {
			t.Errorf("isByeResult(%v) = true, want false", rc)
		}
	}
}

func TestParseResultChar_unknown(t *testing.T) {
	_, ok := parseResultChar('X')
	if ok {
		t.Error("parseResultChar('X') should fail")
	}
}

func TestParseColorChar_unknown(t *testing.T) {
	_, ok := parseColorChar('x')
	if ok {
		t.Error("parseColorChar('x') should fail")
	}
}

func TestParseError_Error(t *testing.T) {
	e := &ParseError{Line: 42, Code: "001", Message: "invalid rating"}
	want := "trf: line 42 (001): invalid rating"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestRead_headers(t *testing.T) {
	input := `012 Test Tournament
022 Amsterdam
032 NED
042 2025/01/15
052 2025/01/20
062 8
072 6
082 0
092 Swiss Dutch
102 Jan de Vries
112 Maria Jansen
122 90min/40moves+30min+30sec
132 2025/01/15
132 2025/01/16
`
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if doc.Name != "Test Tournament" {
		t.Errorf("Name = %q, want %q", doc.Name, "Test Tournament")
	}
	if doc.City != "Amsterdam" {
		t.Errorf("City = %q, want %q", doc.City, "Amsterdam")
	}
	if doc.Federation != "NED" {
		t.Errorf("Federation = %q, want %q", doc.Federation, "NED")
	}
	if doc.StartDate != "2025/01/15" {
		t.Errorf("StartDate = %q, want %q", doc.StartDate, "2025/01/15")
	}
	if doc.EndDate != "2025/01/20" {
		t.Errorf("EndDate = %q, want %q", doc.EndDate, "2025/01/20")
	}
	if doc.NumPlayers != 8 {
		t.Errorf("NumPlayers = %d, want 8", doc.NumPlayers)
	}
	if doc.NumRated != 6 {
		t.Errorf("NumRated = %d, want 6", doc.NumRated)
	}
	if doc.NumTeams != 0 {
		t.Errorf("NumTeams = %d, want 0", doc.NumTeams)
	}
	if doc.TournamentType != "Swiss Dutch" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Swiss Dutch")
	}
	if doc.ChiefArbiter != "Jan de Vries" {
		t.Errorf("ChiefArbiter = %q, want %q", doc.ChiefArbiter, "Jan de Vries")
	}
	if doc.DeputyArbiter != "Maria Jansen" {
		t.Errorf("DeputyArbiter = %q, want %q", doc.DeputyArbiter, "Maria Jansen")
	}
	if doc.TimeControl != "90min/40moves+30min+30sec" {
		t.Errorf("TimeControl = %q, want %q", doc.TimeControl, "90min/40moves+30min+30sec")
	}
	if len(doc.RoundDates) != 2 || doc.RoundDates[0] != "2025/01/15" || doc.RoundDates[1] != "2025/01/16" {
		t.Errorf("RoundDates = %v, want [2025/01/15, 2025/01/16]", doc.RoundDates)
	}
}

func TestRead_playerLine(t *testing.T) {
	// Minimal TRF with one player who played 2 rounds
	input := "001    1 m GM Kasparov, Garry                   2812 RUS 4100018     1963/04/13  1.5    1  0002 w 1  0003 b =\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Players) != 1 {
		t.Fatalf("Players count = %d, want 1", len(doc.Players))
	}
	p := doc.Players[0]
	if p.StartNumber != 1 {
		t.Errorf("StartNumber = %d, want 1", p.StartNumber)
	}
	if p.Sex != "m" {
		t.Errorf("Sex = %q, want %q", p.Sex, "m")
	}
	if p.Title != "GM" {
		t.Errorf("Title = %q, want %q", p.Title, "GM")
	}
	if p.Name != "Kasparov, Garry" {
		t.Errorf("Name = %q, want %q", p.Name, "Kasparov, Garry")
	}
	if p.Rating != 2812 {
		t.Errorf("Rating = %d, want 2812", p.Rating)
	}
	if p.Federation != "RUS" {
		t.Errorf("Federation = %q, want %q", p.Federation, "RUS")
	}
	if p.FideID != "4100018" {
		t.Errorf("FideID = %q, want %q", p.FideID, "4100018")
	}
	if p.BirthDate != "1963/04/13" {
		t.Errorf("BirthDate = %q, want %q", p.BirthDate, "1963/04/13")
	}
	if p.Points != 1.5 {
		t.Errorf("Points = %v, want 1.5", p.Points)
	}
	if p.Rank != 1 {
		t.Errorf("Rank = %d, want 1", p.Rank)
	}
	if len(p.Rounds) != 2 {
		t.Fatalf("Rounds count = %d, want 2", len(p.Rounds))
	}
	r1 := p.Rounds[0]
	if r1.Opponent != 2 || r1.Color != ColorWhite || r1.Result != ResultWin {
		t.Errorf("Round 1 = %+v, want {Opponent:2 Color:White Result:Win}", r1)
	}
	r2 := p.Rounds[1]
	if r2.Opponent != 3 || r2.Color != ColorBlack || r2.Result != ResultDraw {
		t.Errorf("Round 2 = %+v, want {Opponent:3 Color:Black Result:Draw}", r2)
	}
}

func TestRead_XXlines(t *testing.T) {
	input := "XXR 7\nXXC white1\nXXS 1 2.0 2.0 1.5 1.5 1.0 1.0\nXXP 3 5\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if doc.TotalRounds != 7 {
		t.Errorf("TotalRounds = %d, want 7", doc.TotalRounds)
	}
	if doc.InitialColor != "white1" {
		t.Errorf("InitialColor = %q, want %q", doc.InitialColor, "white1")
	}
	if len(doc.Acceleration) != 1 || doc.Acceleration[0] != "1 2.0 2.0 1.5 1.5 1.0 1.0" {
		t.Errorf("Acceleration = %v, want [\"1 2.0 2.0 1.5 1.5 1.0 1.0\"]", doc.Acceleration)
	}
	if len(doc.ForbiddenPairs) != 1 || doc.ForbiddenPairs[0].Player1 != 3 || doc.ForbiddenPairs[0].Player2 != 5 {
		t.Errorf("ForbiddenPairs = %v, want [{3 5}]", doc.ForbiddenPairs)
	}
}

func TestRead_teamLine(t *testing.T) {
	input := "013    1 Chess Club Amsterdam               1  2  3  4\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Teams) != 1 {
		t.Fatalf("Teams count = %d, want 1", len(doc.Teams))
	}
	team := doc.Teams[0]
	if team.TeamNumber != 1 {
		t.Errorf("TeamNumber = %d, want 1", team.TeamNumber)
	}
	if team.TeamName != "Chess Club Amsterdam" {
		t.Errorf("TeamName = %q, want %q", team.TeamName, "Chess Club Amsterdam")
	}
	if len(team.Members) != 4 || team.Members[0] != 1 || team.Members[3] != 4 {
		t.Errorf("Members = %v, want [1 2 3 4]", team.Members)
	}
}

func TestRead_unknownLines(t *testing.T) {
	input := "012 Test\nXYZ custom data here\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Other) != 1 {
		t.Fatalf("Other count = %d, want 1", len(doc.Other))
	}
	if doc.Other[0].Code != "XYZ" || doc.Other[0].Data != "custom data here" {
		t.Errorf("Other[0] = %+v, want {Code:XYZ Data:custom data here}", doc.Other[0])
	}
}

func TestRead_byeRoundResult(t *testing.T) {
	// Player with a full-point bye in round 1
	input := "001    1      Player One                        2000 NED             2000/01/01  1.0    1  0000 - F\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Players) != 1 || len(doc.Players[0].Rounds) != 1 {
		t.Fatalf("unexpected player/round count")
	}
	r := doc.Players[0].Rounds[0]
	if r.Opponent != 0 || r.Color != ColorNone || r.Result != ResultFullBye {
		t.Errorf("bye round = %+v, want {Opponent:0 Color:None Result:FullBye}", r)
	}
}

func TestRead_parseError(t *testing.T) {
	// Malformed player line (too short)
	input := "001    1\n"
	_, err := Read(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for malformed player line")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Line != 1 || pe.Code != "001" {
		t.Errorf("ParseError = %+v, want Line=1 Code=001", pe)
	}
}

func TestRead_empty(t *testing.T) {
	doc, err := Read(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Players) != 0 {
		t.Errorf("Players = %v, want empty", doc.Players)
	}
}

func TestRead_crlfLineEndings(t *testing.T) {
	input := "012 Test\r\nXXR 5\r\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if doc.Name != "Test" {
		t.Errorf("Name = %q, want %q", doc.Name, "Test")
	}
	if doc.TotalRounds != 5 {
		t.Errorf("TotalRounds = %d, want 5", doc.TotalRounds)
	}
}

func TestWrite_headers(t *testing.T) {
	doc := &Document{
		Name:           "Test Tournament",
		City:           "Amsterdam",
		Federation:     "NED",
		TournamentType: "Swiss Dutch",
		TotalRounds:    7,
		InitialColor:   "white1",
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "012 Test Tournament\n") {
		t.Errorf("missing tournament name line in:\n%s", output)
	}
	if !strings.Contains(output, "022 Amsterdam\n") {
		t.Errorf("missing city line in:\n%s", output)
	}
	if !strings.Contains(output, "092 Swiss Dutch\n") {
		t.Errorf("missing tournament type line in:\n%s", output)
	}
	if !strings.Contains(output, "XXR 7\n") {
		t.Errorf("missing XXR line in:\n%s", output)
	}
	if !strings.Contains(output, "XXC white1\n") {
		t.Errorf("missing XXC line in:\n%s", output)
	}
}

func TestWrite_playerLine(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{
				StartNumber: 1,
				Sex:         "m",
				Title:       "GM",
				Name:        "Kasparov, Garry",
				Rating:      2812,
				Federation:  "RUS",
				FideID:      "4100018",
				BirthDate:   "1963/04/13",
				Points:      1.5,
				Rank:        1,
				Rounds: []RoundResult{
					{Opponent: 2, Color: ColorWhite, Result: ResultWin},
					{Opponent: 3, Color: ColorBlack, Result: ResultDraw},
				},
			},
		},
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001    1") {
		t.Errorf("missing player start number in:\n%s", output)
	}
	if !strings.Contains(output, "0002 w 1") {
		t.Errorf("missing round 1 result in:\n%s", output)
	}
	if !strings.Contains(output, "0003 b =") {
		t.Errorf("missing round 2 result in:\n%s", output)
	}
}

func TestReadWrite_roundtrip(t *testing.T) {
	input := "012 Test Tournament\n022 Amsterdam\n092 Swiss Dutch\nXXR 5\nXXC white1\n"
	input += "001    1 m GM Kasparov, Garry                   2812 RUS 4100018     1963/04/13  1.5    1  0002 w 1  0003 b =\n"
	input += "001    2   IM Kramnik, Vladimir                 2750 RUS 4101588     1975/06/25  0.5    2  0001 b 0  0003 w =\n"
	input += "001    3      Player Three                      2000 NED                         1.0    3  0000 - F  0002 b =\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Re-read and compare
	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Re-read failed: %v", err)
	}

	// Compare key fields
	if doc.Name != doc2.Name {
		t.Errorf("Name mismatch: %q vs %q", doc.Name, doc2.Name)
	}
	if len(doc.Players) != len(doc2.Players) {
		t.Fatalf("Player count mismatch: %d vs %d", len(doc.Players), len(doc2.Players))
	}
	for i, p1 := range doc.Players {
		p2 := doc2.Players[i]
		if p1.StartNumber != p2.StartNumber {
			t.Errorf("Player %d StartNumber: %d vs %d", i, p1.StartNumber, p2.StartNumber)
		}
		if p1.Sex != p2.Sex {
			t.Errorf("Player %d Sex: %q vs %q", i, p1.Sex, p2.Sex)
		}
		if p1.Title != p2.Title {
			t.Errorf("Player %d Title: %q vs %q", i, p1.Title, p2.Title)
		}
		if p1.Name != p2.Name {
			t.Errorf("Player %d Name: %q vs %q", i, p1.Name, p2.Name)
		}
		if p1.Rating != p2.Rating {
			t.Errorf("Player %d Rating: %d vs %d", i, p1.Rating, p2.Rating)
		}
		if p1.Federation != p2.Federation {
			t.Errorf("Player %d Federation: %q vs %q", i, p1.Federation, p2.Federation)
		}
		if p1.FideID != p2.FideID {
			t.Errorf("Player %d FideID: %q vs %q", i, p1.FideID, p2.FideID)
		}
		if p1.BirthDate != p2.BirthDate {
			t.Errorf("Player %d BirthDate: %q vs %q", i, p1.BirthDate, p2.BirthDate)
		}
		if p1.Points != p2.Points {
			t.Errorf("Player %d Points: %v vs %v", i, p1.Points, p2.Points)
		}
		if p1.Rank != p2.Rank {
			t.Errorf("Player %d Rank: %d vs %d", i, p1.Rank, p2.Rank)
		}
		if len(p1.Rounds) != len(p2.Rounds) {
			t.Errorf("Player %d Rounds count: %d vs %d", i, len(p1.Rounds), len(p2.Rounds))
			continue
		}
		for j, r1 := range p1.Rounds {
			r2 := p2.Rounds[j]
			if r1.Opponent != r2.Opponent || r1.Color != r2.Color || r1.Result != r2.Result {
				t.Errorf("Player %d Round %d: %+v vs %+v", i, j, r1, r2)
			}
		}
	}
}

func TestRead_goldenBasic(t *testing.T) {
	data, err := os.ReadFile("testdata/basic.trf")
	if err != nil {
		t.Fatalf("read testdata/basic.trf: %v", err)
	}

	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// ── Header fields ──────────────────────────────────────────────

	if doc.Name != "FIDE Grand Prix 2025" {
		t.Errorf("Name = %q, want %q", doc.Name, "FIDE Grand Prix 2025")
	}
	if doc.City != "Amsterdam" {
		t.Errorf("City = %q, want %q", doc.City, "Amsterdam")
	}
	if doc.Federation != "NED" {
		t.Errorf("Federation = %q, want %q", doc.Federation, "NED")
	}
	if doc.StartDate != "2025/01/15" {
		t.Errorf("StartDate = %q, want %q", doc.StartDate, "2025/01/15")
	}
	if doc.EndDate != "2025/01/20" {
		t.Errorf("EndDate = %q, want %q", doc.EndDate, "2025/01/20")
	}
	if doc.NumPlayers != 5 {
		t.Errorf("NumPlayers = %d, want 5", doc.NumPlayers)
	}
	if doc.NumRated != 3 {
		t.Errorf("NumRated = %d, want 3", doc.NumRated)
	}
	if doc.TournamentType != "Swiss Dutch" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Swiss Dutch")
	}
	if doc.ChiefArbiter != "Jan de Vries" {
		t.Errorf("ChiefArbiter = %q, want %q", doc.ChiefArbiter, "Jan de Vries")
	}
	if doc.DeputyArbiter != "Maria Jansen" {
		t.Errorf("DeputyArbiter = %q, want %q", doc.DeputyArbiter, "Maria Jansen")
	}
	if doc.TimeControl != "90min/40moves+30min+30sec" {
		t.Errorf("TimeControl = %q, want %q", doc.TimeControl, "90min/40moves+30min+30sec")
	}
	if doc.TotalRounds != 3 {
		t.Errorf("TotalRounds = %d, want 3", doc.TotalRounds)
	}
	if doc.InitialColor != "white1" {
		t.Errorf("InitialColor = %q, want %q", doc.InitialColor, "white1")
	}
	if len(doc.RoundDates) != 2 {
		t.Errorf("RoundDates count = %d, want 2", len(doc.RoundDates))
	}

	// ── Players ────────────────────────────────────────────────────

	if len(doc.Players) != 5 {
		t.Fatalf("Players = %d, want 5", len(doc.Players))
	}

	// Verify per-player identity, rating, and points.
	type wantPlayer struct {
		name   string
		rating int
		points float64
	}
	wantPlayers := []wantPlayer{
		{"Kasparov, Garry", 2812, 1.5},
		{"Kramnik, Vladimir", 2750, 0.5},
		{"Player Three", 2000, 0.0},
		{"Polgar, Judit", 2735, 1.5},
		{"Player Five", 1800, 2.0},
	}
	for i, wp := range wantPlayers {
		p := doc.Players[i]
		if p.Name != wp.name {
			t.Errorf("Player %d Name = %q, want %q", i+1, p.Name, wp.name)
		}
		if p.Rating != wp.rating {
			t.Errorf("Player %d Rating = %d, want %d", i+1, p.Rating, wp.rating)
		}
		if p.Points != wp.points {
			t.Errorf("Player %d Points = %v, want %v", i+1, p.Points, wp.points)
		}
	}

	// Verify round results for every player (cross-reference consistency).
	type wantRound struct {
		opp    int
		color  Color
		result ResultCode
	}
	wantRounds := [][]wantRound{
		// P1: R1 vs2 w 1, R2 vs4 b =
		{{2, ColorWhite, ResultWin}, {4, ColorBlack, ResultDraw}},
		// P2: R1 vs1 b 0, R2 bye H
		{{1, ColorBlack, ResultLoss}, {0, ColorNone, ResultHalfBye}},
		// P3: R1 vs4 b 0, R2 vs5 b 0
		{{4, ColorBlack, ResultLoss}, {5, ColorBlack, ResultLoss}},
		// P4: R1 vs3 w 1, R2 vs1 w =
		{{3, ColorWhite, ResultWin}, {1, ColorWhite, ResultDraw}},
		// P5: R1 bye F, R2 vs3 w 1
		{{0, ColorNone, ResultFullBye}, {3, ColorWhite, ResultWin}},
	}
	for i, wr := range wantRounds {
		p := doc.Players[i]
		if len(p.Rounds) != len(wr) {
			t.Errorf("Player %d rounds = %d, want %d", i+1, len(p.Rounds), len(wr))
			continue
		}
		for j, want := range wr {
			got := p.Rounds[j]
			if got.Opponent != want.opp || got.Color != want.color || got.Result != want.result {
				t.Errorf("Player %d Round %d = {Opp:%d Color:%v Result:%v}, want {Opp:%d Color:%v Result:%v}",
					i+1, j+1, got.Opponent, got.Color, got.Result, want.opp, want.color, want.result)
			}
		}
	}

	// ── Bye coverage ───────────────────────────────────────────────

	// Player 5 has a full-point bye in round 1.
	if doc.Players[4].Rounds[0].Result != ResultFullBye {
		t.Errorf("Player 5 Round 1 = %v, want FullBye", doc.Players[4].Rounds[0].Result)
	}
	// Player 2 has a half-point bye in round 2.
	if doc.Players[1].Rounds[1].Result != ResultHalfBye {
		t.Errorf("Player 2 Round 2 = %v, want HalfBye", doc.Players[1].Rounds[1].Result)
	}

	// ── Round-trip: Write → re-Read ────────────────────────────────

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Re-read failed: %v", err)
	}

	// Header fields survive round-trip.
	if doc2.Name != doc.Name {
		t.Errorf("Round-trip Name: %q vs %q", doc.Name, doc2.Name)
	}
	if doc2.City != doc.City {
		t.Errorf("Round-trip City: %q vs %q", doc.City, doc2.City)
	}
	if doc2.TotalRounds != doc.TotalRounds {
		t.Errorf("Round-trip TotalRounds: %d vs %d", doc.TotalRounds, doc2.TotalRounds)
	}
	if doc2.InitialColor != doc.InitialColor {
		t.Errorf("Round-trip InitialColor: %q vs %q", doc.InitialColor, doc2.InitialColor)
	}
	if doc2.TournamentType != doc.TournamentType {
		t.Errorf("Round-trip TournamentType: %q vs %q", doc.TournamentType, doc2.TournamentType)
	}
	if doc2.ChiefArbiter != doc.ChiefArbiter {
		t.Errorf("Round-trip ChiefArbiter: %q vs %q", doc.ChiefArbiter, doc2.ChiefArbiter)
	}

	// Player data survives round-trip.
	if len(doc2.Players) != len(doc.Players) {
		t.Fatalf("Round-trip player count: %d vs %d", len(doc2.Players), len(doc.Players))
	}
	for i := range doc.Players {
		p1, p2 := doc.Players[i], doc2.Players[i]
		if p1.Name != p2.Name {
			t.Errorf("Round-trip player %d Name: %q vs %q", i+1, p1.Name, p2.Name)
		}
		if p1.Rating != p2.Rating {
			t.Errorf("Round-trip player %d Rating: %d vs %d", i+1, p1.Rating, p2.Rating)
		}
		if p1.Points != p2.Points {
			t.Errorf("Round-trip player %d Points: %v vs %v", i+1, p1.Points, p2.Points)
		}
		if len(p1.Rounds) != len(p2.Rounds) {
			t.Errorf("Round-trip player %d rounds: %d vs %d", i+1, len(p1.Rounds), len(p2.Rounds))
			continue
		}
		for j, r1 := range p1.Rounds {
			r2 := p2.Rounds[j]
			if r1.Opponent != r2.Opponent || r1.Color != r2.Color || r1.Result != r2.Result {
				t.Errorf("Round-trip player %d round %d: %+v vs %+v", i+1, j+1, r1, r2)
			}
		}
	}
}

func TestRead_systemSpecificXXLines(t *testing.T) {
	// Test Round-Robin XX lines
	rrInput := "012 Test\n092 Round Robin\nXXY 2\nXXB true\n"
	doc, err := Read(strings.NewReader(rrInput))
	if err != nil {
		t.Fatalf("Read Round-Robin XX lines failed: %v", err)
	}
	if doc.Cycles != 2 {
		t.Errorf("Cycles = %d, want 2", doc.Cycles)
	}
	if doc.ColorBalance == nil || !*doc.ColorBalance {
		t.Errorf("ColorBalance = %v, want true", doc.ColorBalance)
	}

	// Test Lim XX lines
	limInput := "012 Test\n092 Swiss Lim\nXXM true\n"
	doc, err = Read(strings.NewReader(limInput))
	if err != nil {
		t.Fatalf("Read Lim XX lines failed: %v", err)
	}
	if doc.MaxiTournament == nil || !*doc.MaxiTournament {
		t.Errorf("MaxiTournament = %v, want true", doc.MaxiTournament)
	}

	// Test Team XX lines
	teamInput := "012 Test\n092 Team Swiss\nXXT A\nXXG match\n"
	doc, err = Read(strings.NewReader(teamInput))
	if err != nil {
		t.Fatalf("Read Team XX lines failed: %v", err)
	}
	if doc.ColorPreferenceType != "A" {
		t.Errorf("ColorPreferenceType = %q, want %q", doc.ColorPreferenceType, "A")
	}
	if doc.PrimaryScore != "match" {
		t.Errorf("PrimaryScore = %q, want %q", doc.PrimaryScore, "match")
	}

	// Test Keizer XX lines
	keizerInput := "012 Test\n092 Keizer\nXXA true\nXXK 5\n"
	doc, err = Read(strings.NewReader(keizerInput))
	if err != nil {
		t.Fatalf("Read Keizer XX lines failed: %v", err)
	}
	if doc.AllowRepeatPairings == nil || !*doc.AllowRepeatPairings {
		t.Errorf("AllowRepeatPairings = %v, want true", doc.AllowRepeatPairings)
	}
	if doc.MinRoundsBetweenRepeats != 5 {
		t.Errorf("MinRoundsBetweenRepeats = %d, want 5", doc.MinRoundsBetweenRepeats)
	}

	// Test false booleans
	falseInput := "012 Test\nXXB false\nXXM false\nXXA false\n"
	doc, err = Read(strings.NewReader(falseInput))
	if err != nil {
		t.Fatalf("Read false booleans failed: %v", err)
	}
	if doc.ColorBalance == nil || *doc.ColorBalance {
		t.Errorf("ColorBalance = %v, want false", doc.ColorBalance)
	}
	if doc.MaxiTournament == nil || *doc.MaxiTournament {
		t.Errorf("MaxiTournament = %v, want false", doc.MaxiTournament)
	}
	if doc.AllowRepeatPairings == nil || *doc.AllowRepeatPairings {
		t.Errorf("AllowRepeatPairings = %v, want false", doc.AllowRepeatPairings)
	}
}

func TestWrite_systemSpecificXXLines(t *testing.T) {
	trueVal := true
	falseVal := false

	doc := &Document{
		Name:                    "Test",
		TournamentType:          "Round Robin",
		Cycles:                  2,
		ColorBalance:            &trueVal,
		MaxiTournament:          &falseVal,
		ColorPreferenceType:     "B",
		PrimaryScore:            "game",
		AllowRepeatPairings:     &trueVal,
		MinRoundsBetweenRepeats: 5,
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()

	wantLines := []string{
		"XXY 2",
		"XXB true",
		"XXM false",
		"XXT B",
		"XXG game",
		"XXA true",
		"XXK 5",
	}
	for _, want := range wantLines {
		if !strings.Contains(output, want+"\n") {
			t.Errorf("output missing line %q\nGot:\n%s", want, output)
		}
	}
}

func TestWrite_emptyDocument(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, &Document{}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %d bytes: %q", buf.Len(), buf.String())
	}
}

func TestWrite_longPlayerName(t *testing.T) {
	longName := "Abcdefghijklmnopqrstuvwxyz1234567890" // 36 chars
	doc := &Document{
		Players: []PlayerLine{
			{
				StartNumber: 1,
				Name:        longName,
				Rank:        1,
			},
		},
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	// Name field is 33 chars max — putLeft truncates silently.
	truncated := longName[:33]
	if !strings.Contains(output, truncated) {
		t.Errorf("output should contain first 33 chars %q\nGot:\n%s", truncated, output)
	}
	// The full 36-char name must NOT appear.
	if strings.Contains(output, longName) {
		t.Errorf("output should NOT contain full 36-char name %q\nGot:\n%s", longName, output)
	}
}

func TestWrite_startNumberOverflow(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{
				StartNumber: 10000, // 5 digits overflows 4-char field
				Name:        "Overflow Player",
				Rank:        1,
			},
		},
	}

	var buf strings.Builder
	err := Write(&buf, doc)
	if err == nil {
		t.Fatal("expected error for start number overflow, got nil")
	}
}

func TestWrite_teamLine(t *testing.T) {
	doc := &Document{
		Teams: []TeamLine{
			{
				TeamNumber: 1,
				TeamName:   "Test Team",
				Members:    []int{1, 2, 3, 4},
			},
		},
	}

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "013") {
		t.Errorf("output should contain team line code '013'\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Test Team") {
		t.Errorf("output should contain team name 'Test Team'\nGot:\n%s", output)
	}
}

func TestReadWrite_otherLinesRoundTrip(t *testing.T) {
	input := "XYZ Some custom data\nABC Another custom line\n"

	// First Read.
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.Other) != 2 {
		t.Fatalf("Other count = %d, want 2", len(doc.Other))
	}

	// Write.
	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Re-Read.
	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Re-read failed: %v", err)
	}
	if len(doc2.Other) != 2 {
		t.Fatalf("Re-read Other count = %d, want 2", len(doc2.Other))
	}

	// Verify Code and Data match.
	for i := range doc.Other {
		if doc.Other[i].Code != doc2.Other[i].Code {
			t.Errorf("Other[%d].Code: %q vs %q", i, doc.Other[i].Code, doc2.Other[i].Code)
		}
		if doc.Other[i].Data != doc2.Other[i].Data {
			t.Errorf("Other[%d].Data: %q vs %q", i, doc.Other[i].Data, doc2.Other[i].Data)
		}
	}
}

// ---------------------------------------------------------------------------
// Error path coverage tests
// ---------------------------------------------------------------------------

func TestRead_invalidRating(t *testing.T) {
	// Build a 001 line with non-numeric rating at bytes 48-51.
	// Pad to >=84 chars. Bytes 48-51 contain "ABCD" (invalid).
	line := "001    1      Player One                        ABCD                             0.0    1"
	if len(line) < 84 {
		t.Fatalf("test line too short: %d chars", len(line))
	}

	_, err := Read(strings.NewReader(line + "\n"))
	if err == nil {
		t.Fatal("expected error for non-numeric rating")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if pe.Code != "001" {
		t.Errorf("ParseError.Code = %q, want %q", pe.Code, "001")
	}
	if !strings.Contains(pe.Message, "rating") {
		t.Errorf("ParseError.Message = %q, want it to mention 'rating'", pe.Message)
	}
}

func TestRead_invalidXXP(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"three numbers", "XXP 1 2 3"},
		{"one number", "XXP 1"},
		{"non-numeric", "XXP abc def"},
		{"empty", "XXP "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(strings.NewReader(tt.input + "\n"))
			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *ParseError, got %T: %v", err, err)
			}
			if pe.Code != "XXP" {
				t.Errorf("ParseError.Code = %q, want %q", pe.Code, "XXP")
			}
		})
	}
}

func TestRead_invalidBooleans(t *testing.T) {
	for _, code := range []string{"XXB", "XXM", "XXA"} {
		t.Run(code, func(t *testing.T) {
			_, err := Read(strings.NewReader(code + " notabool\n"))
			if err == nil {
				t.Fatalf("expected error for %s with invalid boolean", code)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *ParseError, got %T: %v", err, err)
			}
			if pe.Code != code {
				t.Errorf("ParseError.Code = %q, want %q", pe.Code, code)
			}
		})
	}
}

func TestRead_invalidIntegers(t *testing.T) {
	for _, code := range []string{"XXR", "XXY", "XXK", "062", "072", "082"} {
		t.Run(code, func(t *testing.T) {
			_, err := Read(strings.NewReader(code + " notanint\n"))
			if err == nil {
				t.Fatalf("expected error for %s with invalid integer", code)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *ParseError, got %T: %v", err, err)
			}
			if pe.Code != code {
				t.Errorf("ParseError.Code = %q, want %q", pe.Code, code)
			}
		})
	}
}

func TestRead_playerLineNoRounds(t *testing.T) {
	// A valid 001 line with exactly 89 chars: header fields through rank, no round data.
	// Round data starts at byte 89, so len==89 means no rounds parsed.
	line := "001    1      Player One                        2000 NED             2000/01/01  0.0    1"
	if len(line) != 89 {
		t.Fatalf("test line length = %d, want 89", len(line))
	}

	doc, err := Read(strings.NewReader(line + "\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(doc.Players) != 1 {
		t.Fatalf("Players count = %d, want 1", len(doc.Players))
	}
	if len(doc.Players[0].Rounds) != 0 {
		t.Errorf("Rounds count = %d, want 0", len(doc.Players[0].Rounds))
	}
	if doc.Players[0].Name != "Player One" {
		t.Errorf("Name = %q, want %q", doc.Players[0].Name, "Player One")
	}
}

func TestRead_malformedTeamLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// "013 AB" = 6 chars, well under the 40-char minimum
		{"too short", "013 AB"},
		// Valid length (>=40) but team number field (bytes 4-7) is non-numeric
		{"invalid team number", "013 ABCD                                    1"},
		// Valid team number but non-numeric member after the 32-char name field
		{"invalid member number", "013    1 Team Name Here                  abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(strings.NewReader(tt.input + "\n"))
			if err == nil {
				t.Fatalf("expected error for input %q", tt.input)
			}
			var pe *ParseError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *ParseError, got %T: %v", err, err)
			}
			if pe.Code != "013" {
				t.Errorf("ParseError.Code = %q, want %q", pe.Code, "013")
			}
		})
	}
}

func TestRead_shortLines(t *testing.T) {
	// Lines < 3 chars should be silently skipped.
	input := "AB\n012 Test Tournament\nX\n"
	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Name != "Test Tournament" {
		t.Errorf("Name = %q, want %q", doc.Name, "Test Tournament")
	}
}

// ---------------------------------------------------------------------------
// TRF-2026 tests
// ---------------------------------------------------------------------------

func TestRead_TRF2026_headers(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Standard TRF16 headers.
	if doc.Name != "TRF-2026 Test Tournament" {
		t.Errorf("Name = %q, want %q", doc.Name, "TRF-2026 Test Tournament")
	}
	if doc.City != "Antwerp" {
		t.Errorf("City = %q, want %q", doc.City, "Antwerp")
	}
	if doc.Federation != "BEL" {
		t.Errorf("Federation = %q, want %q", doc.Federation, "BEL")
	}
	if doc.StartDate != "2026/03/01" {
		t.Errorf("StartDate = %q, want %q", doc.StartDate, "2026/03/01")
	}
	if doc.EndDate != "2026/03/07" {
		t.Errorf("EndDate = %q, want %q", doc.EndDate, "2026/03/07")
	}
	if doc.NumPlayers != 8 {
		t.Errorf("NumPlayers = %d, want 8", doc.NumPlayers)
	}
	if doc.NumRated != 8 {
		t.Errorf("NumRated = %d, want 8", doc.NumRated)
	}
	if doc.NumTeams != 2 {
		t.Errorf("NumTeams = %d, want 2", doc.NumTeams)
	}
	if doc.TournamentType != "Team Swiss" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Team Swiss")
	}
	if doc.ChiefArbiter != "Jan de Vries" {
		t.Errorf("ChiefArbiter = %q, want %q", doc.ChiefArbiter, "Jan de Vries")
	}
	if doc.TimeControl != "90min/40moves+30min+30sec" {
		t.Errorf("TimeControl = %q, want %q", doc.TimeControl, "90min/40moves+30min+30sec")
	}
	if len(doc.RoundDates) != 3 {
		t.Errorf("RoundDates count = %d, want 3", len(doc.RoundDates))
	}

	// Multiple deputy arbiters.
	if doc.DeputyArbiter != "Maria Jansen" {
		t.Errorf("DeputyArbiter = %q, want %q", doc.DeputyArbiter, "Maria Jansen")
	}
	if len(doc.DeputyArbiters) != 2 {
		t.Fatalf("DeputyArbiters count = %d, want 2", len(doc.DeputyArbiters))
	}
	if doc.DeputyArbiters[0] != "Maria Jansen" {
		t.Errorf("DeputyArbiters[0] = %q, want %q", doc.DeputyArbiters[0], "Maria Jansen")
	}
	if doc.DeputyArbiters[1] != "Pieter Bakker" {
		t.Errorf("DeputyArbiters[1] = %q, want %q", doc.DeputyArbiters[1], "Pieter Bakker")
	}

	// Legacy XX fields.
	if doc.TotalRounds != 3 {
		t.Errorf("TotalRounds (XXR) = %d, want 3", doc.TotalRounds)
	}
	if doc.InitialColor != "white1" {
		t.Errorf("InitialColor (XXC) = %q, want %q", doc.InitialColor, "white1")
	}

	// TRF-2026 header fields.
	if doc.TotalRounds26 != 3 {
		t.Errorf("TotalRounds26 (142) = %d, want 3", doc.TotalRounds26)
	}
	if doc.InitialColor26 != "W" {
		t.Errorf("InitialColor26 (152) = %q, want %q", doc.InitialColor26, "W")
	}
	if doc.ScoringSystem != " W 1.0    D 0.5    L 0.0" {
		t.Errorf("ScoringSystem (162) = %q, want %q", doc.ScoringSystem, " W 1.0    D 0.5    L 0.0")
	}
	if doc.StartingRankMethod != "IND FIDE" {
		t.Errorf("StartingRankMethod (172) = %q, want %q", doc.StartingRankMethod, "IND FIDE")
	}
	if doc.CodedTournamentType != "FIDE_TEAM_BAKU" {
		t.Errorf("CodedTournamentType (192) = %q, want %q", doc.CodedTournamentType, "FIDE_TEAM_BAKU")
	}
	if doc.TieBreakDef != "EDET/P,BH:MP/C1/P" {
		t.Errorf("TieBreakDef (202) = %q, want %q", doc.TieBreakDef, "EDET/P,BH:MP/C1/P")
	}
	if doc.EncodedTimeControl != "40/6000+30:20/3000+30:1500+30" {
		t.Errorf("EncodedTimeControl (222) = %q, want %q", doc.EncodedTimeControl, "40/6000+30:20/3000+30:1500+30")
	}
	if doc.TeamInitialColor != "WBWB" {
		t.Errorf("TeamInitialColor (352) = %q, want %q", doc.TeamInitialColor, "WBWB")
	}
	if doc.TeamScoringSystem != "TW 2     TD 1     TL 0" {
		t.Errorf("TeamScoringSystem (362) = %q, want %q", doc.TeamScoringSystem, "TW 2     TD 1     TL 0")
	}

	// Effective* helpers prefer TRF-2026 values.
	if doc.EffectiveTotalRounds() != 3 {
		t.Errorf("EffectiveTotalRounds() = %d, want 3", doc.EffectiveTotalRounds())
	}
	if doc.EffectiveInitialColor() != "W" {
		t.Errorf("EffectiveInitialColor() = %q, want %q", doc.EffectiveInitialColor(), "W")
	}
}

func TestRead_TRF2026_players(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.Players) != 8 {
		t.Fatalf("Players count = %d, want 8", len(doc.Players))
	}

	// Spot-check player 1.
	p1 := doc.Players[0]
	if p1.StartNumber != 1 {
		t.Errorf("P1 StartNumber = %d, want 1", p1.StartNumber)
	}
	if p1.Name != "Fischer, Bobby" {
		t.Errorf("P1 Name = %q, want %q", p1.Name, "Fischer, Bobby")
	}
	if p1.Title != "GM" {
		t.Errorf("P1 Title = %q, want %q", p1.Title, "GM")
	}
	if p1.Rating != 2785 {
		t.Errorf("P1 Rating = %d, want 2785", p1.Rating)
	}
	if p1.Federation != "USA" {
		t.Errorf("P1 Federation = %q, want %q", p1.Federation, "USA")
	}
	if p1.Points != 2.0 {
		t.Errorf("P1 Points = %v, want 2.0", p1.Points)
	}
	if p1.Rank != 1 {
		t.Errorf("P1 Rank = %d, want 1", p1.Rank)
	}
	if len(p1.Rounds) != 3 {
		t.Fatalf("P1 Rounds count = %d, want 3", len(p1.Rounds))
	}
	// R1: 0005 w 1, R2: 0006 b 1, R3: 0000 - *
	if r := p1.Rounds[0]; r.Opponent != 5 || r.Color != ColorWhite || r.Result != ResultWin {
		t.Errorf("P1 R1 = %+v, want {5 White Win}", r)
	}
	if r := p1.Rounds[1]; r.Opponent != 6 || r.Color != ColorBlack || r.Result != ResultWin {
		t.Errorf("P1 R2 = %+v, want {6 Black Win}", r)
	}
	if r := p1.Rounds[2]; r.Opponent != 0 || r.Color != ColorNone || r.Result != ResultNotPlayed {
		t.Errorf("P1 R3 = %+v, want {0 None NotPlayed}", r)
	}

	// Spot-check player 8 (last player).
	p8 := doc.Players[7]
	if p8.StartNumber != 8 {
		t.Errorf("P8 StartNumber = %d, want 8", p8.StartNumber)
	}
	if p8.Name != "Player Eight" {
		t.Errorf("P8 Name = %q, want %q", p8.Name, "Player Eight")
	}
	if p8.Rating != 2050 {
		t.Errorf("P8 Rating = %d, want 2050", p8.Rating)
	}
	if p8.Points != 0.5 {
		t.Errorf("P8 Points = %v, want 0.5", p8.Points)
	}
}

func TestRead_TRF2026_comments(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.Comments) != 2 {
		t.Fatalf("Comments count = %d, want 2", len(doc.Comments))
	}
	if doc.Comments[0] != "Column guide for 001 lines" {
		t.Errorf("Comments[0] = %q, want %q", doc.Comments[0], "Column guide for 001 lines")
	}
	if doc.Comments[1] != "SSSS sTTT NNNNN RRRR FFF" {
		t.Errorf("Comments[1] = %q, want %q", doc.Comments[1], "SSSS sTTT NNNNN RRRR FFF")
	}
}

func TestRead_TRF2026_NRSRecords(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.NRSRecords) != 2 {
		t.Fatalf("NRSRecords count = %d, want 2", len(doc.NRSRecords))
	}

	nrs1 := doc.NRSRecords[0]
	if nrs1.Federation != "IND" {
		t.Errorf("NRS[0] Federation = %q, want %q", nrs1.Federation, "IND")
	}
	if nrs1.StartNumber != 1 {
		t.Errorf("NRS[0] StartNumber = %d, want 1", nrs1.StartNumber)
	}
	if nrs1.Title != "GM" {
		t.Errorf("NRS[0] Title = %q, want %q", nrs1.Title, "GM")
	}
	if nrs1.Name != "Fischer, Bobby" {
		t.Errorf("NRS[0] Name = %q, want %q", nrs1.Name, "Fischer, Bobby")
	}
	if nrs1.NationalRating != 2785 {
		t.Errorf("NRS[0] NationalRating = %d, want 2785", nrs1.NationalRating)
	}

	nrs2 := doc.NRSRecords[1]
	if nrs2.StartNumber != 5 {
		t.Errorf("NRS[1] StartNumber = %d, want 5", nrs2.StartNumber)
	}
	if nrs2.Name != "Player Five" {
		t.Errorf("NRS[1] Name = %q, want %q", nrs2.Name, "Player Five")
	}
}

func TestRead_TRF2026_legacyTeams(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// 013 legacy teams.
	if len(doc.Teams) != 2 {
		t.Fatalf("Teams count = %d, want 2", len(doc.Teams))
	}
	if doc.Teams[0].TeamName != "Team Alpha" {
		t.Errorf("Teams[0].TeamName = %q, want %q", doc.Teams[0].TeamName, "Team Alpha")
	}
	if len(doc.Teams[0].Members) != 4 || doc.Teams[0].Members[0] != 1 || doc.Teams[0].Members[3] != 4 {
		t.Errorf("Teams[0].Members = %v, want [1 2 3 4]", doc.Teams[0].Members)
	}
	if doc.Teams[1].TeamName != "Team Beta" {
		t.Errorf("Teams[1].TeamName = %q, want %q", doc.Teams[1].TeamName, "Team Beta")
	}
}

func TestRead_TRF2026_newTeams310(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// 310 new teams.
	if len(doc.NewTeams) != 2 {
		t.Fatalf("NewTeams count = %d, want 2", len(doc.NewTeams))
	}

	nt1 := doc.NewTeams[0]
	if nt1.TeamNumber != 1 {
		t.Errorf("NewTeams[0].TeamNumber = %d, want 1", nt1.TeamNumber)
	}
	if nt1.TeamName != "Team Alpha" {
		t.Errorf("NewTeams[0].TeamName = %q, want %q", nt1.TeamName, "Team Alpha")
	}
	if nt1.Federation != "USA" {
		t.Errorf("NewTeams[0].Federation = %q, want %q", nt1.Federation, "USA")
	}
	if nt1.AvgRating != 2746 {
		t.Errorf("NewTeams[0].AvgRating = %v, want 2746", nt1.AvgRating)
	}
	if nt1.MatchPoints != 4 {
		t.Errorf("NewTeams[0].MatchPoints = %v, want 4", nt1.MatchPoints)
	}
	if nt1.GamePoints != 6.5 {
		t.Errorf("NewTeams[0].GamePoints = %v, want 6.5", nt1.GamePoints)
	}
	if nt1.Rank != 1 {
		t.Errorf("NewTeams[0].Rank = %d, want 1", nt1.Rank)
	}
	if len(nt1.Members) != 4 || nt1.Members[0] != 1 || nt1.Members[3] != 4 {
		t.Errorf("NewTeams[0].Members = %v, want [1 2 3 4]", nt1.Members)
	}

	nt2 := doc.NewTeams[1]
	if nt2.TeamNumber != 2 {
		t.Errorf("NewTeams[1].TeamNumber = %d, want 2", nt2.TeamNumber)
	}
	if nt2.TeamName != "Team Beta" {
		t.Errorf("NewTeams[1].TeamName = %q, want %q", nt2.TeamName, "Team Beta")
	}
	if nt2.Federation != "NED" {
		t.Errorf("NewTeams[1].Federation = %q, want %q", nt2.Federation, "NED")
	}
	if nt2.GamePoints != 3.0 {
		t.Errorf("NewTeams[1].GamePoints = %v, want 3.0", nt2.GamePoints)
	}
}

func TestRead_TRF2026_absences240(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.Absences) != 2 {
		t.Fatalf("Absences count = %d, want 2", len(doc.Absences))
	}

	a1 := doc.Absences[0]
	if a1.Type != "F" {
		t.Errorf("Absences[0].Type = %q, want %q", a1.Type, "F")
	}
	if a1.Round != 1 {
		t.Errorf("Absences[0].Round = %d, want 1", a1.Round)
	}
	if len(a1.Players) != 1 || a1.Players[0] != 2 {
		t.Errorf("Absences[0].Players = %v, want [2]", a1.Players)
	}

	a2 := doc.Absences[1]
	if a2.Type != "H" {
		t.Errorf("Absences[1].Type = %q, want %q", a2.Type, "H")
	}
	if a2.Round != 2 {
		t.Errorf("Absences[1].Round = %d, want 2", a2.Round)
	}
	if len(a2.Players) != 2 || a2.Players[0] != 7 || a2.Players[1] != 8 {
		t.Errorf("Absences[1].Players = %v, want [7 8]", a2.Players)
	}
}

func TestRead_TRF2026_accelerations250(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.Accelerations26) != 1 {
		t.Fatalf("Accelerations26 count = %d, want 1", len(doc.Accelerations26))
	}

	acc := doc.Accelerations26[0]
	if acc.Raw == "" {
		t.Error("Accelerations26[0].Raw is empty, expected raw data for round-trip")
	}
	// The fixture line: "250  2          1   1    1    2"
	// Fields: 2 (match pts) then remaining fields.
	if acc.MatchPoints != 2 {
		t.Errorf("Accelerations26[0].MatchPoints = %v, want 2", acc.MatchPoints)
	}
}

func TestRead_TRF2026_forbiddenPairs260(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.ForbiddenPairs26) != 2 {
		t.Fatalf("ForbiddenPairs26 count = %d, want 2", len(doc.ForbiddenPairs26))
	}

	fp1 := doc.ForbiddenPairs26[0]
	if fp1.FirstRound != 1 {
		t.Errorf("FP26[0].FirstRound = %d, want 1", fp1.FirstRound)
	}
	if fp1.LastRound != 3 {
		t.Errorf("FP26[0].LastRound = %d, want 3", fp1.LastRound)
	}
	if len(fp1.Players) != 5 {
		t.Fatalf("FP26[0].Players count = %d, want 5", len(fp1.Players))
	}
	wantPlayers := []int{1, 2, 3, 4, 5}
	for i, want := range wantPlayers {
		if fp1.Players[i] != want {
			t.Errorf("FP26[0].Players[%d] = %d, want %d", i, fp1.Players[i], want)
		}
	}

	fp2 := doc.ForbiddenPairs26[1]
	if fp2.FirstRound != 1 || fp2.LastRound != 3 {
		t.Errorf("FP26[1] rounds = %d-%d, want 1-3", fp2.FirstRound, fp2.LastRound)
	}
	if len(fp2.Players) != 3 || fp2.Players[0] != 6 || fp2.Players[1] != 7 || fp2.Players[2] != 8 {
		t.Errorf("FP26[1].Players = %v, want [6 7 8]", fp2.Players)
	}
}

func TestRead_TRF2026_teamRoundData300(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.TeamRoundData) != 2 {
		t.Fatalf("TeamRoundData count = %d, want 2", len(doc.TeamRoundData))
	}

	tr1 := doc.TeamRoundData[0]
	if tr1.Round != 1 {
		t.Errorf("TR[0].Round = %d, want 1", tr1.Round)
	}
	if tr1.Team1 != 1 || tr1.Team2 != 2 {
		t.Errorf("TR[0] teams = %d vs %d, want 1 vs 2", tr1.Team1, tr1.Team2)
	}
	if len(tr1.Boards) != 8 {
		t.Fatalf("TR[0].Boards count = %d, want 8", len(tr1.Boards))
	}
	// 300   1   1   2    1    5    2    6    3    7    4    8
	wantBoards := []int{1, 5, 2, 6, 3, 7, 4, 8}
	for i, want := range wantBoards {
		if tr1.Boards[i] != want {
			t.Errorf("TR[0].Boards[%d] = %d, want %d", i, tr1.Boards[i], want)
		}
	}

	tr2 := doc.TeamRoundData[1]
	if tr2.Round != 2 || tr2.Team1 != 2 || tr2.Team2 != 1 {
		t.Errorf("TR[1] = round %d team %d vs %d, want round 2 team 2 vs 1", tr2.Round, tr2.Team1, tr2.Team2)
	}
}

func TestRead_TRF2026_oldAbsentForfeits330(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.OldAbsentForfeits) != 2 {
		t.Fatalf("OldAbsentForfeits count = %d, want 2", len(doc.OldAbsentForfeits))
	}

	oaf1 := doc.OldAbsentForfeits[0]
	if oaf1.ResultType != "+-" {
		t.Errorf("OAF[0].ResultType = %q, want %q", oaf1.ResultType, "+-")
	}
	if oaf1.Round != 1 {
		t.Errorf("OAF[0].Round = %d, want 1", oaf1.Round)
	}
	if oaf1.WhiteTeam != 1 || oaf1.BlackTeam != 2 {
		t.Errorf("OAF[0] teams = %d vs %d, want 1 vs 2", oaf1.WhiteTeam, oaf1.BlackTeam)
	}

	oaf2 := doc.OldAbsentForfeits[1]
	if oaf2.ResultType != "-+" {
		t.Errorf("OAF[1].ResultType = %q, want %q", oaf2.ResultType, "-+")
	}
}

func TestRead_TRF2026_teamRoundScores320(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.TeamRoundScores) != 1 {
		t.Fatalf("TeamRoundScores count = %d, want 1", len(doc.TeamRoundScores))
	}

	ts := doc.TeamRoundScores[0]
	if ts.TeamNumber != 1 {
		t.Errorf("TRS[0].TeamNumber = %d, want 1", ts.TeamNumber)
	}
	if ts.GamePoints != 6.5 {
		t.Errorf("TRS[0].GamePoints = %v, want 6.5", ts.GamePoints)
	}
	if ts.Raw == "" {
		t.Error("TRS[0].Raw is empty, expected raw data")
	}
}

func TestRead_TRF2026_detailedTeamResults801(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.DetailedTeamResults) != 2 {
		t.Fatalf("DetailedTeamResults count = %d, want 2", len(doc.DetailedTeamResults))
	}

	dtr1 := doc.DetailedTeamResults[0]
	if dtr1.TeamNumber != 1 {
		t.Errorf("DTR[0].TeamNumber = %d, want 1", dtr1.TeamNumber)
	}
	if dtr1.TeamName != "Alpha" {
		t.Errorf("DTR[0].TeamName = %q, want %q", dtr1.TeamName, "Alpha")
	}
	if dtr1.MatchPoints != 4 {
		t.Errorf("DTR[0].MatchPoints = %v, want 4", dtr1.MatchPoints)
	}
	if dtr1.GamePoints != 6.5 {
		t.Errorf("DTR[0].GamePoints = %v, want 6.5", dtr1.GamePoints)
	}
	if dtr1.Raw == "" {
		t.Error("DTR[0].Raw is empty, expected raw data")
	}
}

func TestRead_TRF2026_simpleTeamResults802(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.SimpleTeamResults) != 2 {
		t.Fatalf("SimpleTeamResults count = %d, want 2", len(doc.SimpleTeamResults))
	}

	str1 := doc.SimpleTeamResults[0]
	if str1.TeamNumber != 1 {
		t.Errorf("STR[0].TeamNumber = %d, want 1", str1.TeamNumber)
	}
	if str1.TeamName != "Alpha" {
		t.Errorf("STR[0].TeamName = %q, want %q", str1.TeamName, "Alpha")
	}
	if str1.Raw == "" {
		t.Error("STR[0].Raw is empty, expected raw data")
	}
}

func TestReadWrite_TRF2026_roundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team.trf: %v", err)
	}

	// First read.
	doc1, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	// Write.
	var buf strings.Builder
	if err := Write(&buf, doc1); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Re-read.
	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Re-read failed: %v", err)
	}

	// Compare TRF16 headers.
	if doc1.Name != doc2.Name {
		t.Errorf("Name: %q vs %q", doc1.Name, doc2.Name)
	}
	if doc1.City != doc2.City {
		t.Errorf("City: %q vs %q", doc1.City, doc2.City)
	}
	if doc1.NumPlayers != doc2.NumPlayers {
		t.Errorf("NumPlayers: %d vs %d", doc1.NumPlayers, doc2.NumPlayers)
	}
	if doc1.TotalRounds != doc2.TotalRounds {
		t.Errorf("TotalRounds: %d vs %d", doc1.TotalRounds, doc2.TotalRounds)
	}
	if doc1.InitialColor != doc2.InitialColor {
		t.Errorf("InitialColor: %q vs %q", doc1.InitialColor, doc2.InitialColor)
	}

	// Compare TRF-2026 headers.
	if doc1.TotalRounds26 != doc2.TotalRounds26 {
		t.Errorf("TotalRounds26: %d vs %d", doc1.TotalRounds26, doc2.TotalRounds26)
	}
	if doc1.InitialColor26 != doc2.InitialColor26 {
		t.Errorf("InitialColor26: %q vs %q", doc1.InitialColor26, doc2.InitialColor26)
	}
	if doc1.ScoringSystem != doc2.ScoringSystem {
		t.Errorf("ScoringSystem: %q vs %q", doc1.ScoringSystem, doc2.ScoringSystem)
	}
	if doc1.StartingRankMethod != doc2.StartingRankMethod {
		t.Errorf("StartingRankMethod: %q vs %q", doc1.StartingRankMethod, doc2.StartingRankMethod)
	}
	if doc1.CodedTournamentType != doc2.CodedTournamentType {
		t.Errorf("CodedTournamentType: %q vs %q", doc1.CodedTournamentType, doc2.CodedTournamentType)
	}
	if doc1.TieBreakDef != doc2.TieBreakDef {
		t.Errorf("TieBreakDef: %q vs %q", doc1.TieBreakDef, doc2.TieBreakDef)
	}
	if doc1.EncodedTimeControl != doc2.EncodedTimeControl {
		t.Errorf("EncodedTimeControl: %q vs %q", doc1.EncodedTimeControl, doc2.EncodedTimeControl)
	}
	if doc1.TeamInitialColor != doc2.TeamInitialColor {
		t.Errorf("TeamInitialColor: %q vs %q", doc1.TeamInitialColor, doc2.TeamInitialColor)
	}
	if doc1.TeamScoringSystem != doc2.TeamScoringSystem {
		t.Errorf("TeamScoringSystem: %q vs %q", doc1.TeamScoringSystem, doc2.TeamScoringSystem)
	}

	// Compare deputy arbiters.
	if len(doc1.DeputyArbiters) != len(doc2.DeputyArbiters) {
		t.Errorf("DeputyArbiters count: %d vs %d", len(doc1.DeputyArbiters), len(doc2.DeputyArbiters))
	} else {
		for i := range doc1.DeputyArbiters {
			if doc1.DeputyArbiters[i] != doc2.DeputyArbiters[i] {
				t.Errorf("DeputyArbiters[%d]: %q vs %q", i, doc1.DeputyArbiters[i], doc2.DeputyArbiters[i])
			}
		}
	}

	// Compare comments.
	if len(doc1.Comments) != len(doc2.Comments) {
		t.Errorf("Comments count: %d vs %d", len(doc1.Comments), len(doc2.Comments))
	} else {
		for i := range doc1.Comments {
			if doc1.Comments[i] != doc2.Comments[i] {
				t.Errorf("Comments[%d]: %q vs %q", i, doc1.Comments[i], doc2.Comments[i])
			}
		}
	}

	// Compare players.
	if len(doc1.Players) != len(doc2.Players) {
		t.Fatalf("Players count: %d vs %d", len(doc1.Players), len(doc2.Players))
	}
	for i, p1 := range doc1.Players {
		p2 := doc2.Players[i]
		if p1.StartNumber != p2.StartNumber || p1.Name != p2.Name || p1.Rating != p2.Rating || p1.Points != p2.Points {
			t.Errorf("Player %d mismatch: {%d %q %d %.1f} vs {%d %q %d %.1f}",
				i+1, p1.StartNumber, p1.Name, p1.Rating, p1.Points,
				p2.StartNumber, p2.Name, p2.Rating, p2.Points)
		}
		if len(p1.Rounds) != len(p2.Rounds) {
			t.Errorf("Player %d rounds count: %d vs %d", i+1, len(p1.Rounds), len(p2.Rounds))
			continue
		}
		for j, r1 := range p1.Rounds {
			r2 := p2.Rounds[j]
			if r1.Opponent != r2.Opponent || r1.Color != r2.Color || r1.Result != r2.Result {
				t.Errorf("Player %d Round %d: %+v vs %+v", i+1, j+1, r1, r2)
			}
		}
	}

	// Compare NRS records.
	if len(doc1.NRSRecords) != len(doc2.NRSRecords) {
		t.Errorf("NRSRecords count: %d vs %d", len(doc1.NRSRecords), len(doc2.NRSRecords))
	} else {
		for i, n1 := range doc1.NRSRecords {
			n2 := doc2.NRSRecords[i]
			if n1.Federation != n2.Federation || n1.StartNumber != n2.StartNumber || n1.Name != n2.Name {
				t.Errorf("NRS[%d] mismatch: {%s %d %q} vs {%s %d %q}",
					i, n1.Federation, n1.StartNumber, n1.Name, n2.Federation, n2.StartNumber, n2.Name)
			}
		}
	}

	// Compare legacy teams (013).
	if len(doc1.Teams) != len(doc2.Teams) {
		t.Errorf("Teams count: %d vs %d", len(doc1.Teams), len(doc2.Teams))
	}

	// Compare new teams (310).
	if len(doc1.NewTeams) != len(doc2.NewTeams) {
		t.Errorf("NewTeams count: %d vs %d", len(doc1.NewTeams), len(doc2.NewTeams))
	} else {
		for i, t1 := range doc1.NewTeams {
			t2 := doc2.NewTeams[i]
			if t1.TeamNumber != t2.TeamNumber || t1.TeamName != t2.TeamName || t1.Federation != t2.Federation {
				t.Errorf("NewTeams[%d] mismatch: {%d %q %q} vs {%d %q %q}",
					i, t1.TeamNumber, t1.TeamName, t1.Federation,
					t2.TeamNumber, t2.TeamName, t2.Federation)
			}
			if t1.AvgRating != t2.AvgRating || t1.MatchPoints != t2.MatchPoints || t1.GamePoints != t2.GamePoints {
				t.Errorf("NewTeams[%d] stats: {%.0f %.0f %.1f} vs {%.0f %.0f %.1f}",
					i, t1.AvgRating, t1.MatchPoints, t1.GamePoints,
					t2.AvgRating, t2.MatchPoints, t2.GamePoints)
			}
		}
	}

	// Compare absences (240).
	if len(doc1.Absences) != len(doc2.Absences) {
		t.Errorf("Absences count: %d vs %d", len(doc1.Absences), len(doc2.Absences))
	} else {
		for i, a1 := range doc1.Absences {
			a2 := doc2.Absences[i]
			if a1.Type != a2.Type || a1.Round != a2.Round {
				t.Errorf("Absences[%d]: {%s %d} vs {%s %d}", i, a1.Type, a1.Round, a2.Type, a2.Round)
			}
			if len(a1.Players) != len(a2.Players) {
				t.Errorf("Absences[%d].Players count: %d vs %d", i, len(a1.Players), len(a2.Players))
			}
		}
	}

	// Compare forbidden pairs (260).
	if len(doc1.ForbiddenPairs26) != len(doc2.ForbiddenPairs26) {
		t.Errorf("ForbiddenPairs26 count: %d vs %d", len(doc1.ForbiddenPairs26), len(doc2.ForbiddenPairs26))
	} else {
		for i, fp1 := range doc1.ForbiddenPairs26 {
			fp2 := doc2.ForbiddenPairs26[i]
			if fp1.FirstRound != fp2.FirstRound || fp1.LastRound != fp2.LastRound {
				t.Errorf("FP26[%d] rounds: %d-%d vs %d-%d", i, fp1.FirstRound, fp1.LastRound, fp2.FirstRound, fp2.LastRound)
			}
			if len(fp1.Players) != len(fp2.Players) {
				t.Errorf("FP26[%d].Players count: %d vs %d", i, len(fp1.Players), len(fp2.Players))
			}
		}
	}

	// Compare team round data (300).
	if len(doc1.TeamRoundData) != len(doc2.TeamRoundData) {
		t.Errorf("TeamRoundData count: %d vs %d", len(doc1.TeamRoundData), len(doc2.TeamRoundData))
	} else {
		for i, tr1 := range doc1.TeamRoundData {
			tr2 := doc2.TeamRoundData[i]
			if tr1.Round != tr2.Round || tr1.Team1 != tr2.Team1 || tr1.Team2 != tr2.Team2 {
				t.Errorf("TeamRoundData[%d]: r%d t%d-t%d vs r%d t%d-t%d",
					i, tr1.Round, tr1.Team1, tr1.Team2, tr2.Round, tr2.Team1, tr2.Team2)
			}
			if len(tr1.Boards) != len(tr2.Boards) {
				t.Errorf("TeamRoundData[%d].Boards count: %d vs %d", i, len(tr1.Boards), len(tr2.Boards))
			}
		}
	}

	// Compare old absent forfeits (330).
	if len(doc1.OldAbsentForfeits) != len(doc2.OldAbsentForfeits) {
		t.Errorf("OldAbsentForfeits count: %d vs %d", len(doc1.OldAbsentForfeits), len(doc2.OldAbsentForfeits))
	} else {
		for i, o1 := range doc1.OldAbsentForfeits {
			o2 := doc2.OldAbsentForfeits[i]
			if o1.ResultType != o2.ResultType || o1.Round != o2.Round ||
				o1.WhiteTeam != o2.WhiteTeam || o1.BlackTeam != o2.BlackTeam {
				t.Errorf("OldAbsentForfeits[%d]: %+v vs %+v", i, o1, o2)
			}
		}
	}

	// Compare 801 and 802 counts (raw round-trip).
	if len(doc1.DetailedTeamResults) != len(doc2.DetailedTeamResults) {
		t.Errorf("DetailedTeamResults count: %d vs %d", len(doc1.DetailedTeamResults), len(doc2.DetailedTeamResults))
	}
	if len(doc1.SimpleTeamResults) != len(doc2.SimpleTeamResults) {
		t.Errorf("SimpleTeamResults count: %d vs %d", len(doc1.SimpleTeamResults), len(doc2.SimpleTeamResults))
	}

	// Compare team round scores (320).
	if len(doc1.TeamRoundScores) != len(doc2.TeamRoundScores) {
		t.Errorf("TeamRoundScores count: %d vs %d", len(doc1.TeamRoundScores), len(doc2.TeamRoundScores))
	}

	// Compare accelerations (250).
	if len(doc1.Accelerations26) != len(doc2.Accelerations26) {
		t.Errorf("Accelerations26 count: %d vs %d", len(doc1.Accelerations26), len(doc2.Accelerations26))
	}
}

func TestRead_bbpPairingsOutput(t *testing.T) {
	f, err := os.Open("testdata/bbp-8p5r-output.trf")
	if err != nil {
		t.Skipf("bbpPairings output file not found: %v (run bbpPairings to generate)", err)
	}
	defer func() { _ = f.Close() }()

	doc, err := Read(f)
	if err != nil {
		t.Fatalf("Read bbpPairings output: %v", err)
	}

	// Basic structural checks.
	if len(doc.Players) != 8 {
		t.Errorf("got %d players, want 8", len(doc.Players))
	}

	// Verify all players have round data.
	for _, p := range doc.Players {
		if len(p.Rounds) == 0 {
			t.Errorf("player %d has no round data", p.StartNumber)
		}
	}

	// Validate cross-references (if available — this test runs independently).
	issues := doc.Validate(ValidatePairingEngine)
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Errorf("validation error: %s: %s", issue.Field, issue.Message)
		}
	}

	// Round-trip: Write and re-Read.
	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Re-read: %v", err)
	}
	if len(doc2.Players) != len(doc.Players) {
		t.Errorf("re-read: %d players, want %d", len(doc2.Players), len(doc.Players))
	}

	// Verify ToTournamentState succeeds.
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if len(state.Players) != 8 {
		t.Errorf("state has %d players, want 8", len(state.Players))
	}
}
