package trf

import (
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
	nonByes := []ResultCode{ResultWin, ResultLoss, ResultDraw, ResultForfeitWin,
		ResultForfeitLoss, ResultNotPlayed, ResultWinByDefault, ResultDrawByDefault, ResultLossByDefault}
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
