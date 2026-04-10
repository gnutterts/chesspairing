// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestToTournamentState_basic(t *testing.T) {
	input := "012 Test Tournament\n022 Amsterdam\n092 Swiss Dutch\nXXR 5\nXXC white1\n"
	input += "001    1 m GM Kasparov, Garry                   2812 RUS 4100018     1963/04/13  1.5    1  0002 w 1  0003 b =\n"
	input += "001    2   IM Kramnik, Vladimir                 2750 RUS 4101588     1975/06/25  0.5    2  0001 b 0  0003 w =\n"
	input += "001    3      Player Three                      2000 NED                         1.0    3  0000 - F  0002 b =\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	// Check players
	if len(state.Players) != 3 {
		t.Fatalf("Players count = %d, want 3", len(state.Players))
	}

	p1 := state.Players[0]
	if p1.ID != "1" {
		t.Errorf("Player 1 ID = %q, want %q", p1.ID, "1")
	}
	if p1.DisplayName != "Kasparov, Garry" {
		t.Errorf("Player 1 Name = %q, want %q", p1.DisplayName, "Kasparov, Garry")
	}
	if p1.Rating != 2812 {
		t.Errorf("Player 1 Rating = %d, want 2812", p1.Rating)
	}
	if p1.Federation != "RUS" {
		t.Errorf("Player 1 Federation = %q, want %q", p1.Federation, "RUS")
	}
	if p1.FideID != "4100018" {
		t.Errorf("Player 1 FideID = %q, want %q", p1.FideID, "4100018")
	}
	if p1.Title != "GM" {
		t.Errorf("Player 1 Title = %q, want %q", p1.Title, "GM")
	}
	if p1.Sex != "m" {
		t.Errorf("Player 1 Sex = %q, want %q", p1.Sex, "m")
	}

	// Check rounds
	if len(state.Rounds) != 2 {
		t.Fatalf("Rounds count = %d, want 2", len(state.Rounds))
	}

	// Round 1: game 1v2 (white wins) + bye for player 3
	r1 := state.Rounds[0]
	if r1.Number != 1 {
		t.Errorf("Round 1 Number = %d, want 1", r1.Number)
	}
	if len(r1.Games) != 1 {
		t.Fatalf("Round 1 Games = %d, want 1", len(r1.Games))
	}
	g1 := r1.Games[0]
	if g1.WhiteID != "1" || g1.BlackID != "2" {
		t.Errorf("Round 1 Game 1: White=%q Black=%q, want White=1 Black=2", g1.WhiteID, g1.BlackID)
	}
	if g1.Result != chesspairing.ResultWhiteWins {
		t.Errorf("Round 1 Game 1 Result = %q, want %q", g1.Result, chesspairing.ResultWhiteWins)
	}
	if len(r1.Byes) != 1 || r1.Byes[0].PlayerID != "3" || r1.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("Round 1 Byes = %+v, want [{PlayerID:3 Type:ByePAB}]", r1.Byes)
	}

	// Round 2: two draw games (1v3, 2v3)
	r2 := state.Rounds[1]
	if len(r2.Games) != 2 {
		t.Fatalf("Round 2 Games = %d, want 2", len(r2.Games))
	}

	// Check tournament info
	if state.Info.Name != "Test Tournament" {
		t.Errorf("Info.Name = %q, want %q", state.Info.Name, "Test Tournament")
	}
	if state.Info.City != "Amsterdam" {
		t.Errorf("Info.City = %q, want %q", state.Info.City, "Amsterdam")
	}

	// Check pairing config
	if state.PairingConfig.System != chesspairing.PairingDutch {
		t.Errorf("PairingConfig.System = %q, want %q", state.PairingConfig.System, chesspairing.PairingDutch)
	}
	if state.CurrentRound != 2 {
		t.Errorf("CurrentRound = %d, want 2", state.CurrentRound)
	}
}

func TestToTournamentState_forfeits(t *testing.T) {
	// Player 1 wins by forfeit against player 2
	input := "001    1      Player One                        2000 NED                         1.0    1  0002 w +\n"
	input += "001    2      Player Two                        1800 NED                         0.0    2  0001 b -\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 1 || len(state.Rounds[0].Games) != 1 {
		t.Fatalf("unexpected round/game count")
	}
	g := state.Rounds[0].Games[0]
	if g.Result != chesspairing.ResultForfeitWhiteWins {
		t.Errorf("Result = %q, want %q", g.Result, chesspairing.ResultForfeitWhiteWins)
	}
	if !g.IsForfeit {
		t.Error("IsForfeit = false, want true")
	}
}

func TestToTournamentState_byeTypes(t *testing.T) {
	// 4 players, each with a different bye type
	input := "001    1      Player One                        2000 NED                         1.0    1  0000 - F\n"
	input += "001    2      Player Two                        1800 NED                         0.5    2  0000 - H\n"
	input += "001    3      Player Three                      1600 NED                         0.0    3  0000 - Z\n"
	input += "001    4      Player Four                       1400 NED                         0.0    4  0000 - U\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 1 {
		t.Fatalf("Rounds = %d, want 1", len(state.Rounds))
	}

	byes := state.Rounds[0].Byes
	if len(byes) != 4 {
		t.Fatalf("Byes = %d, want 4", len(byes))
	}

	wantTypes := map[string]chesspairing.ByeType{
		"1": chesspairing.ByePAB,
		"2": chesspairing.ByeHalf,
		"3": chesspairing.ByeZero,
		"4": chesspairing.ByeAbsent,
	}
	for _, bye := range byes {
		want, ok := wantTypes[bye.PlayerID]
		if !ok {
			t.Errorf("unexpected bye player %q", bye.PlayerID)
			continue
		}
		if bye.Type != want {
			t.Errorf("player %s bye type = %v, want %v", bye.PlayerID, bye.Type, want)
		}
	}
}

func TestFromTournamentState_basic(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2200, Active: true, Federation: "NED"},
			{ID: "b", DisplayName: "Bob", Rating: 2000, Active: true, Federation: "BEL"},
			{ID: "c", DisplayName: "Carol", Rating: 1800, Active: true, Federation: "NED"},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "c", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 1,
		Info: chesspairing.TournamentInfo{
			Name: "Test",
			City: "Antwerp",
		},
	}

	doc, playerMap := FromTournamentState(state)

	// Players sorted by rating desc: Alice(2200)=1, Bob(2000)=2, Carol(1800)=3
	if playerMap["a"] != 1 || playerMap["b"] != 2 || playerMap["c"] != 3 {
		t.Errorf("playerMap = %v, want a=1 b=2 c=3", playerMap)
	}

	if doc.Name != "Test" {
		t.Errorf("Name = %q, want %q", doc.Name, "Test")
	}
	if len(doc.Players) != 3 {
		t.Fatalf("Players = %d, want 3", len(doc.Players))
	}

	// Check Alice's round result
	alice := doc.Players[0]
	if alice.StartNumber != 1 {
		t.Errorf("Alice StartNumber = %d, want 1", alice.StartNumber)
	}
	if len(alice.Rounds) != 1 {
		t.Fatalf("Alice Rounds = %d, want 1", len(alice.Rounds))
	}
	if alice.Rounds[0].Opponent != 2 || alice.Rounds[0].Color != ColorWhite || alice.Rounds[0].Result != ResultWin {
		t.Errorf("Alice Round 1 = %+v, want {Opponent:2 Color:White Result:Win}", alice.Rounds[0])
	}

	// Check Carol's bye
	carol := doc.Players[2]
	if len(carol.Rounds) != 1 {
		t.Fatalf("Carol Rounds = %d, want 1", len(carol.Rounds))
	}
	if carol.Rounds[0].Opponent != 0 || carol.Rounds[0].Result != ResultFullBye {
		t.Errorf("Carol Round 1 = %+v, want {Opponent:0 Result:FullBye}", carol.Rounds[0])
	}
}

func TestFromTournamentState_doubleForfeit(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "b", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true},
				},
			},
		},
	}

	doc, _ := FromTournamentState(state)
	if len(doc.Players) != 2 {
		t.Fatalf("Players = %d, want 2", len(doc.Players))
	}

	// Both players should have ForfeitLoss ("-")
	for _, p := range doc.Players {
		if len(p.Rounds) != 1 {
			t.Fatalf("Player %d Rounds = %d, want 1", p.StartNumber, len(p.Rounds))
		}
		if p.Rounds[0].Result != ResultForfeitLoss {
			t.Errorf("Player %d Result = %v, want ForfeitLoss", p.StartNumber, p.Rounds[0].Result)
		}
	}
}

func TestConversion_roundtrip(t *testing.T) {
	// Build a state, convert to TRF Document, write, read back, convert back to state.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Player 2400", Rating: 2400, Active: true, Federation: "NED"},
			{ID: "p2", DisplayName: "Player 2300", Rating: 2300, Active: true, Federation: "BEL"},
			{ID: "p3", DisplayName: "Player 2200", Rating: 2200, Active: true, Federation: "NED"},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 1,
		Info: chesspairing.TournamentInfo{
			Name: "Round-trip Test",
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard},
	}

	doc, playerMap := FromTournamentState(state)

	var buf strings.Builder
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	doc2, err := Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state2, err := doc2.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	// Verify players (IDs will be start numbers now, not original IDs)
	if len(state2.Players) != len(state.Players) {
		t.Fatalf("Player count: %d vs %d", len(state2.Players), len(state.Players))
	}

	// Verify round data is preserved
	if len(state2.Rounds) != len(state.Rounds) {
		t.Fatalf("Round count: %d vs %d", len(state2.Rounds), len(state.Rounds))
	}

	r1 := state2.Rounds[0]
	if len(r1.Games) != 1 {
		t.Fatalf("Round 1 games: %d, want 1", len(r1.Games))
	}

	// Check the game was reconstructed (white=1, black=2 in start numbers)
	g := r1.Games[0]
	sn1 := playerMap["p1"]
	sn2 := playerMap["p2"]
	if g.WhiteID != fmt.Sprintf("%d", sn1) || g.BlackID != fmt.Sprintf("%d", sn2) {
		t.Errorf("Game: White=%q Black=%q, want White=%d Black=%d", g.WhiteID, g.BlackID, sn1, sn2)
	}

	// Check bye preserved
	if len(r1.Byes) != 1 {
		t.Fatalf("Round 1 byes: %d, want 1", len(r1.Byes))
	}
	sn3 := playerMap["p3"]
	if r1.Byes[0].PlayerID != fmt.Sprintf("%d", sn3) || r1.Byes[0].Type != chesspairing.ByePAB {
		t.Errorf("Bye = %+v, want {PlayerID:%d Type:ByePAB}", r1.Byes[0], sn3)
	}
}

func TestFromTournamentState_numRated(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Rated", Rating: 2200, Active: true},
			{ID: "b", DisplayName: "Unrated", Rating: 0, Active: true},
			{ID: "c", DisplayName: "Also Rated", Rating: 1800, Active: true},
		},
	}

	doc, _ := FromTournamentState(state)

	if doc.NumPlayers != 3 {
		t.Errorf("NumPlayers = %d, want 3", doc.NumPlayers)
	}
	if doc.NumRated != 2 {
		t.Errorf("NumRated = %d, want 2", doc.NumRated)
	}
}

func TestFromTournamentState_totalRoundsFromOptions(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "b", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
			}},
		},
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingDutch,
			Options: map[string]any{"totalRounds": 7},
		},
	}

	doc, _ := FromTournamentState(state)

	// Explicit option takes precedence over len(Rounds).
	if doc.TotalRounds != 7 {
		t.Errorf("TotalRounds = %d, want 7", doc.TotalRounds)
	}
}

func TestFromTournamentState_totalRoundsFallback(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "b", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultDraw},
			}},
		},
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingDutch,
		},
	}

	doc, _ := FromTournamentState(state)

	// No explicit totalRounds option — falls back to len(state.Rounds).
	if doc.TotalRounds != 2 {
		t.Errorf("TotalRounds = %d, want 2", doc.TotalRounds)
	}
}

func TestConversion_bursteinRoundTrip(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "b", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingBurstein,
		},
	}

	doc, _ := FromTournamentState(state)
	if doc.TournamentType != "Swiss Burstein" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Swiss Burstein")
	}

	state2, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	if state2.PairingConfig.System != chesspairing.PairingBurstein {
		t.Errorf("round-trip System = %q, want %q", state2.PairingConfig.System, chesspairing.PairingBurstein)
	}
}

func TestInferPairingSystem_allSystems(t *testing.T) {
	tests := []struct {
		tournamentType string
		want           chesspairing.PairingSystem
	}{
		{"Swiss Dutch", chesspairing.PairingDutch},
		{"Swiss Burstein", chesspairing.PairingBurstein},
		{"Swiss Dubov", chesspairing.PairingDubov},
		{"Swiss Lim", chesspairing.PairingLim},
		{"Double Swiss", chesspairing.PairingDoubleSwiss},
		{"Team Swiss", chesspairing.PairingTeam},
		{"Round Robin", chesspairing.PairingRoundRobin},
		{"Double Round Robin", chesspairing.PairingRoundRobin},
		{"Keizer", chesspairing.PairingKeizer},
	}

	for _, tt := range tests {
		t.Run(tt.tournamentType, func(t *testing.T) {
			got := inferPairingSystem(tt.tournamentType)
			if got != tt.want {
				t.Errorf("inferPairingSystem(%q) = %q, want %q", tt.tournamentType, got, tt.want)
			}
		})
	}
}

func TestToTournamentState_accelerationFromXXS(t *testing.T) {
	input := "012 Test\n092 Swiss Dutch\nXXS 1 2.0 3.0\nXXS 2 1.0 2.0\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if state.PairingConfig.Options["acceleration"] != "baku" {
		t.Errorf("acceleration = %v, want %q", state.PairingConfig.Options["acceleration"], "baku")
	}
}

func TestToTournamentState_roundRobinOptions(t *testing.T) {
	input := "012 Test\n092 Double Round Robin\nXXY 2\nXXB true\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if state.PairingConfig.System != chesspairing.PairingRoundRobin {
		t.Errorf("System = %q, want %q", state.PairingConfig.System, chesspairing.PairingRoundRobin)
	}
	if state.PairingConfig.Options["cycles"] != 2 {
		t.Errorf("cycles = %v, want 2", state.PairingConfig.Options["cycles"])
	}
	if state.PairingConfig.Options["colorBalance"] != true {
		t.Errorf("colorBalance = %v, want true", state.PairingConfig.Options["colorBalance"])
	}
}

func TestToTournamentState_limOptions(t *testing.T) {
	input := "012 Test\n092 Swiss Lim\nXXM true\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if state.PairingConfig.Options["maxiTournament"] != true {
		t.Errorf("maxiTournament = %v, want true", state.PairingConfig.Options["maxiTournament"])
	}
}

func TestToTournamentState_teamOptions(t *testing.T) {
	input := "012 Test\n092 Team Swiss\nXXT B\nXXG game\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if state.PairingConfig.Options["colorPreferenceType"] != "B" {
		t.Errorf("colorPreferenceType = %v, want %q", state.PairingConfig.Options["colorPreferenceType"], "B")
	}
	if state.PairingConfig.Options["primaryScore"] != "game" {
		t.Errorf("primaryScore = %v, want %q", state.PairingConfig.Options["primaryScore"], "game")
	}
}

func TestToTournamentState_keizerOptions(t *testing.T) {
	input := "012 Test\n092 Keizer\nXXA false\nXXK 5\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if state.PairingConfig.Options["allowRepeatPairings"] != false {
		t.Errorf("allowRepeatPairings = %v, want false", state.PairingConfig.Options["allowRepeatPairings"])
	}
	if state.PairingConfig.Options["minRoundsBetweenRepeats"] != 5 {
		t.Errorf("minRoundsBetweenRepeats = %v, want 5", state.PairingConfig.Options["minRoundsBetweenRepeats"])
	}
}

func TestFromTournamentState_allTournamentTypes(t *testing.T) {
	tests := []struct {
		system chesspairing.PairingSystem
		want   string
	}{
		{chesspairing.PairingDutch, "Swiss Dutch"},
		{chesspairing.PairingBurstein, "Swiss Burstein"},
		{chesspairing.PairingDubov, "Swiss Dubov"},
		{chesspairing.PairingLim, "Swiss Lim"},
		{chesspairing.PairingDoubleSwiss, "Double Swiss"},
		{chesspairing.PairingTeam, "Team Swiss"},
		{chesspairing.PairingRoundRobin, "Round Robin"},
		{chesspairing.PairingKeizer, "Keizer"},
	}

	for _, tt := range tests {
		t.Run(string(tt.system), func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
				},
				PairingConfig: chesspairing.PairingConfig{System: tt.system},
			}
			doc, _ := FromTournamentState(state)
			if doc.TournamentType != tt.want {
				t.Errorf("TournamentType = %q, want %q", doc.TournamentType, tt.want)
			}
		})
	}
}

func TestFromTournamentState_doubleRoundRobin(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingRoundRobin,
			Options: map[string]any{"cycles": 2},
		},
	}

	doc, _ := FromTournamentState(state)
	if doc.TournamentType != "Double Round Robin" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Double Round Robin")
	}
	if doc.Cycles != 2 {
		t.Errorf("Cycles = %d, want 2", doc.Cycles)
	}
}

func TestFromTournamentState_singleRoundRobin(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingRoundRobin,
			Options: map[string]any{"cycles": 1},
		},
	}

	doc, _ := FromTournamentState(state)
	if doc.TournamentType != "Round Robin" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Round Robin")
	}
}

func TestFromTournamentState_roundRobinNoOptions(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System: chesspairing.PairingRoundRobin,
		},
	}

	doc, _ := FromTournamentState(state)
	if doc.TournamentType != "Round Robin" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Round Robin")
	}
}

func TestFromTournamentState_accelerationToXXS(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
		},
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingDutch,
			Options: map[string]any{"acceleration": "baku"},
		},
	}

	doc, _ := FromTournamentState(state)
	if doc.TournamentType != "Swiss Dutch" {
		t.Errorf("TournamentType = %q, want %q", doc.TournamentType, "Swiss Dutch")
	}
	if len(doc.Acceleration) == 0 {
		t.Error("Acceleration should be non-empty when acceleration=baku")
	}
}

func TestFromTournamentState_systemSpecificOptions(t *testing.T) {
	tests := []struct {
		name    string
		system  chesspairing.PairingSystem
		options map[string]any
		check   func(t *testing.T, doc *Document)
	}{
		{
			name:    "round-robin colorBalance",
			system:  chesspairing.PairingRoundRobin,
			options: map[string]any{"colorBalance": false},
			check: func(t *testing.T, doc *Document) {
				t.Helper()
				if doc.ColorBalance == nil || *doc.ColorBalance {
					t.Errorf("ColorBalance = %v, want false", doc.ColorBalance)
				}
			},
		},
		{
			name:    "lim maxiTournament",
			system:  chesspairing.PairingLim,
			options: map[string]any{"maxiTournament": true},
			check: func(t *testing.T, doc *Document) {
				t.Helper()
				if doc.MaxiTournament == nil || !*doc.MaxiTournament {
					t.Errorf("MaxiTournament = %v, want true", doc.MaxiTournament)
				}
			},
		},
		{
			name:    "team colorPreferenceType and primaryScore",
			system:  chesspairing.PairingTeam,
			options: map[string]any{"colorPreferenceType": "B", "primaryScore": "game"},
			check: func(t *testing.T, doc *Document) {
				t.Helper()
				if doc.ColorPreferenceType != "B" {
					t.Errorf("ColorPreferenceType = %q, want %q", doc.ColorPreferenceType, "B")
				}
				if doc.PrimaryScore != "game" {
					t.Errorf("PrimaryScore = %q, want %q", doc.PrimaryScore, "game")
				}
			},
		},
		{
			name:    "keizer options",
			system:  chesspairing.PairingKeizer,
			options: map[string]any{"allowRepeatPairings": false, "minRoundsBetweenRepeats": 5},
			check: func(t *testing.T, doc *Document) {
				t.Helper()
				if doc.AllowRepeatPairings == nil || *doc.AllowRepeatPairings {
					t.Errorf("AllowRepeatPairings = %v, want false", doc.AllowRepeatPairings)
				}
				if doc.MinRoundsBetweenRepeats != 5 {
					t.Errorf("MinRoundsBetweenRepeats = %d, want 5", doc.MinRoundsBetweenRepeats)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true},
				},
				PairingConfig: chesspairing.PairingConfig{
					System:  tt.system,
					Options: tt.options,
				},
			}
			doc, _ := FromTournamentState(state)
			tt.check(t, doc)
		})
	}
}

func TestConversion_systemSpecificRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		system  chesspairing.PairingSystem
		options map[string]any
		check   func(t *testing.T, opts map[string]any)
	}{
		{
			name:    "Dutch with acceleration",
			system:  chesspairing.PairingDutch,
			options: map[string]any{"acceleration": "baku", "topSeedColor": "white"},
			check: func(t *testing.T, opts map[string]any) {
				t.Helper()
				if opts["acceleration"] != "baku" {
					t.Errorf("acceleration = %v, want %q", opts["acceleration"], "baku")
				}
				if opts["topSeedColor"] != "white" {
					t.Errorf("topSeedColor = %v, want %q", opts["topSeedColor"], "white")
				}
			},
		},
		{
			name:    "Double Round Robin",
			system:  chesspairing.PairingRoundRobin,
			options: map[string]any{"cycles": 2, "colorBalance": false},
			check: func(t *testing.T, opts map[string]any) {
				t.Helper()
				if opts["cycles"] != 2 {
					t.Errorf("cycles = %v, want 2", opts["cycles"])
				}
				if opts["colorBalance"] != false {
					t.Errorf("colorBalance = %v, want false", opts["colorBalance"])
				}
			},
		},
		{
			name:    "Lim with maxiTournament",
			system:  chesspairing.PairingLim,
			options: map[string]any{"maxiTournament": true},
			check: func(t *testing.T, opts map[string]any) {
				t.Helper()
				if opts["maxiTournament"] != true {
					t.Errorf("maxiTournament = %v, want true", opts["maxiTournament"])
				}
			},
		},
		{
			name:    "Team with all options",
			system:  chesspairing.PairingTeam,
			options: map[string]any{"colorPreferenceType": "B", "primaryScore": "game", "totalRounds": 9},
			check: func(t *testing.T, opts map[string]any) {
				t.Helper()
				if opts["colorPreferenceType"] != "B" {
					t.Errorf("colorPreferenceType = %v, want %q", opts["colorPreferenceType"], "B")
				}
				if opts["primaryScore"] != "game" {
					t.Errorf("primaryScore = %v, want %q", opts["primaryScore"], "game")
				}
				if opts["totalRounds"] != 9 {
					t.Errorf("totalRounds = %v, want 9", opts["totalRounds"])
				}
			},
		},
		{
			name:    "Keizer with repeat options",
			system:  chesspairing.PairingKeizer,
			options: map[string]any{"allowRepeatPairings": false, "minRoundsBetweenRepeats": 5},
			check: func(t *testing.T, opts map[string]any) {
				t.Helper()
				if opts["allowRepeatPairings"] != false {
					t.Errorf("allowRepeatPairings = %v, want false", opts["allowRepeatPairings"])
				}
				if opts["minRoundsBetweenRepeats"] != 5 {
					t.Errorf("minRoundsBetweenRepeats = %v, want 5", opts["minRoundsBetweenRepeats"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players: []chesspairing.PlayerEntry{
					{ID: "a", DisplayName: "Alice", Rating: 2000, Active: true, Federation: "NED"},
					{ID: "b", DisplayName: "Bob", Rating: 1800, Active: true, Federation: "BEL"},
				},
				PairingConfig: chesspairing.PairingConfig{
					System:  tt.system,
					Options: tt.options,
				},
			}

			// State -> Document -> TRF bytes -> Document -> State
			doc1, _ := FromTournamentState(state)

			var buf strings.Builder
			if err := Write(&buf, doc1); err != nil {
				t.Fatalf("Write failed: %v", err)
			}

			doc2, err := Read(strings.NewReader(buf.String()))
			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			state2, err := doc2.ToTournamentState()
			if err != nil {
				t.Fatalf("ToTournamentState failed: %v", err)
			}

			if state2.PairingConfig.System != tt.system {
				t.Errorf("System = %q, want %q", state2.PairingConfig.System, tt.system)
			}
			tt.check(t, state2.PairingConfig.Options)
		})
	}
}

func TestToTournamentState_winByDefault(t *testing.T) {
	// Player 1 has W (WinByDefault) vs player 2; player 2 has L (LossByDefault).
	input := "012 Test\n092 Swiss Dutch\nXXR 1\nXXC white1\n"
	input += "001    1      Alice                             2000 USA                         0.0    1  0002 w W\n"
	input += "001    2      Bob                               1800 USA                         0.0    2  0001 b L\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 1 {
		t.Fatalf("Rounds = %d, want 1", len(state.Rounds))
	}
	if len(state.Rounds[0].Games) != 1 {
		t.Fatalf("Games = %d, want 1", len(state.Rounds[0].Games))
	}

	g := state.Rounds[0].Games[0]
	if g.Result != chesspairing.ResultWhiteWins {
		t.Errorf("Result = %q, want %q", g.Result, chesspairing.ResultWhiteWins)
	}
	if !g.IsForfeit {
		t.Error("IsForfeit = false, want true")
	}
	if g.WhiteID != "1" || g.BlackID != "2" {
		t.Errorf("WhiteID=%q BlackID=%q, want WhiteID=1 BlackID=2", g.WhiteID, g.BlackID)
	}
}

func TestToTournamentState_drawByDefault(t *testing.T) {
	// Both players have D (DrawByDefault).
	input := "012 Test\n092 Swiss Dutch\nXXR 1\nXXC white1\n"
	input += "001    1      Alice                             2000 USA                         0.0    1  0002 w D\n"
	input += "001    2      Bob                               1800 USA                         0.0    2  0001 b D\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 1 {
		t.Fatalf("Rounds = %d, want 1", len(state.Rounds))
	}
	if len(state.Rounds[0].Games) != 1 {
		t.Fatalf("Games = %d, want 1", len(state.Rounds[0].Games))
	}

	g := state.Rounds[0].Games[0]
	if g.Result != chesspairing.ResultDraw {
		t.Errorf("Result = %q, want %q", g.Result, chesspairing.ResultDraw)
	}
	if !g.IsForfeit {
		t.Error("IsForfeit = false, want true")
	}
}

func TestToTournamentState_pendingResult(t *testing.T) {
	// Three players: round 1 is played, round 2 has pending games (asterisk).
	input := "012 Test\n092 Swiss Dutch\nXXR 2\nXXC white1\n"
	input += "001    1      Alice                             2000 USA                         1.0    1  0002 w 1  0003 w *\n"
	input += "001    2      Bob                               1800 USA                         0.0    2  0001 b 0  0000 - U\n"
	input += "001    3      Carol                             1600 USA                         0.0    3  0000 - F  0001 b *\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	if len(state.Rounds) != 2 {
		t.Fatalf("Rounds = %d, want 2", len(state.Rounds))
	}

	// Round 1 should have 1 game and 1 bye.
	r1 := state.Rounds[0]
	if len(r1.Games) != 1 {
		t.Errorf("Round 1 Games = %d, want 1", len(r1.Games))
	}
	if len(r1.Byes) != 1 {
		t.Errorf("Round 1 Byes = %d, want 1", len(r1.Byes))
	}

	// Round 2: pending results (ResultNotPlayed / "*") are skipped by the converter.
	// Bob has a bye (U), so round 2 should have no games and 1 bye.
	r2 := state.Rounds[1]
	if len(r2.Games) != 0 {
		t.Errorf("Round 2 Games = %d, want 0 (pending results skipped)", len(r2.Games))
	}
	t.Log("pending results with '*' are skipped by the converter as expected")
	if len(r2.Byes) != 1 {
		t.Errorf("Round 2 Byes = %d, want 1", len(r2.Byes))
	}
}

func TestToTournamentState_raggedRounds(t *testing.T) {
	// Player 1 has 3 rounds, player 2 has only 2 rounds.
	input := "012 Test\n092 Swiss Dutch\nXXR 3\nXXC white1\n"
	input += "001    1      Alice                             2000 USA                         1.5    1  0002 w 1  0002 b =  0000 - F\n"
	input += "001    2      Bob                               1800 USA                         0.5    2  0001 b 0  0001 w =\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}

	// maxRounds is determined by the player with the most rounds (3).
	if len(state.Rounds) != 3 {
		t.Fatalf("Rounds = %d, want 3", len(state.Rounds))
	}

	// Round 1: one game (1 vs 2).
	if len(state.Rounds[0].Games) != 1 {
		t.Errorf("Round 1 Games = %d, want 1", len(state.Rounds[0].Games))
	}

	// Round 2: one game (1 vs 2, rematch).
	if len(state.Rounds[1].Games) != 1 {
		t.Errorf("Round 2 Games = %d, want 1", len(state.Rounds[1].Games))
	}

	// Round 3: player 1 has a bye (F), player 2 has no data for this round.
	r3 := state.Rounds[2]
	if len(r3.Byes) != 1 {
		t.Errorf("Round 3 Byes = %d, want 1", len(r3.Byes))
	}
	if len(r3.Byes) > 0 && r3.Byes[0].PlayerID != "1" {
		t.Errorf("Round 3 Bye PlayerID = %q, want %q", r3.Byes[0].PlayerID, "1")
	}
}

func TestConversion_teamDataRoundTrip(t *testing.T) {
	// Build a Document with 2 teams, Write it, Read it back, verify teams survive.
	doc := &Document{
		Name:           "Team Test",
		TournamentType: "Team Swiss",
		TotalRounds:    5,
		NumPlayers:     4,
		NumTeams:       2,
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice", Rating: 2200, Rank: 1, Points: 0},
			{StartNumber: 2, Name: "Bob", Rating: 2000, Rank: 2, Points: 0},
			{StartNumber: 3, Name: "Carol", Rating: 1800, Rank: 3, Points: 0},
			{StartNumber: 4, Name: "Dave", Rating: 1600, Rank: 4, Points: 0},
		},
		Teams: []TeamLine{
			{TeamNumber: 1, TeamName: "Alpha Team", Members: []int{1, 2}},
			{TeamNumber: 2, TeamName: "Beta Team", Members: []int{3, 4}},
		},
	}

	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	doc2, err := Read(&buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc2.Teams) != 2 {
		t.Fatalf("Teams = %d, want 2", len(doc2.Teams))
	}

	for i, want := range doc.Teams {
		got := doc2.Teams[i]
		if got.TeamNumber != want.TeamNumber {
			t.Errorf("Team %d: TeamNumber = %d, want %d", i, got.TeamNumber, want.TeamNumber)
		}
		if got.TeamName != want.TeamName {
			t.Errorf("Team %d: TeamName = %q, want %q", i, got.TeamName, want.TeamName)
		}
		if fmt.Sprint(got.Members) != fmt.Sprint(want.Members) {
			t.Errorf("Team %d: Members = %v, want %v", i, got.Members, want.Members)
		}
	}
}
