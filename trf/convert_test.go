package trf

import (
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
	if state.PairingConfig.System != chesspairing.PairingSwiss {
		t.Errorf("PairingConfig.System = %q, want %q", state.PairingConfig.System, chesspairing.PairingSwiss)
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
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingSwiss},
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
