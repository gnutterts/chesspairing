// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBuildParticipantStates_BasicFourPlayers(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200},
			{ID: "p4", DisplayName: "Diana", Rating: 2100},
		},
		CurrentRound: 1,
	}

	participants := mustBuildParticipantStates(t, state)
	if len(participants) != 4 {
		t.Fatalf("expected 4 participants, got %d", len(participants))
	}

	// Should be sorted by rating desc → TPN 1=Alice, 2=Bob, 3=Charlie, 4=Diana.
	if participants[0].ID != "p1" || participants[0].TPN != 1 {
		t.Errorf("expected p1 with TPN 1, got %s with TPN %d", participants[0].ID, participants[0].TPN)
	}
	if participants[3].ID != "p4" || participants[3].TPN != 4 {
		t.Errorf("expected p4 with TPN 4, got %s with TPN %d", participants[3].ID, participants[3].TPN)
	}

	// Round 1: all scores should be 0.
	for _, p := range participants {
		if p.Score != 0 {
			t.Errorf("round 1: expected score 0 for %s, got %f", p.ID, p.Score)
		}
	}
}

func TestSortParticipantsKeepsFixedTPN(t *testing.T) {
	participants := []ParticipantState{
		{ID: "p2", TPN: 2, Score: 1},
		{ID: "p10", TPN: 10, Score: 2},
		{ID: "p3", TPN: 3, Score: 2},
	}

	got := SortParticipants(participants)
	if got[0].ID != "p3" || got[1].ID != "p10" || got[2].ID != "p2" {
		t.Fatalf("order = %s, %s, %s; want score then fixed TPN", got[0].ID, got[1].ID, got[2].ID)
	}
	if got[0].TPN != 3 || got[1].TPN != 10 || got[2].TPN != 2 {
		t.Fatalf("SortParticipants renumbered TPNs: %+v", got)
	}
}

func TestBuildParticipantStates_WithHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
			{ID: "p3", DisplayName: "Charlie", Rating: 2200},
			{ID: "p4", DisplayName: "Diana", Rating: 2100},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultBlackWins},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)

	// After round 1: p1 scored 1.0, p2 scored 1.0, p3 scored 0.0, p4 scored 0.0
	// TPN ordering: score desc, then initial rank asc.
	// p1 (1.0, rank 1) → TPN 1, p2 (1.0, rank 2) → TPN 2
	// p3 (0.0, rank 3) → TPN 3, p4 (0.0, rank 4) → TPN 4
	scoreMap := make(map[string]float64)
	for _, p := range participants {
		scoreMap[p.ID] = p.Score
	}
	if scoreMap["p1"] != 1.0 {
		t.Errorf("p1 score: expected 1.0, got %f", scoreMap["p1"])
	}
	if scoreMap["p3"] != 0.0 {
		t.Errorf("p3 score: expected 0.0, got %f", scoreMap["p3"])
	}
}

func TestBuildParticipantStates_OpponentHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if len(p1.Opponents) != 1 || p1.Opponents[0] != "p2" {
		t.Errorf("p1 opponents: expected [p2], got %v", p1.Opponents)
	}
}

func TestBuildParticipantStates_ForfeitExcludedFromOpponents(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	// Forfeits excluded from opponent history — can be re-paired.
	if len(p1.Opponents) != 0 {
		t.Errorf("forfeit should not add opponent: got %v", p1.Opponents)
	}
	// But forfeit winner still gets points.
	if p1.Score != 1.0 {
		t.Errorf("forfeit winner score: expected 1.0, got %f", p1.Score)
	}
}

func TestBuildParticipantStates_DoubleSwissMatches(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultDraw},
				{WhiteID: "c", BlackID: "d", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "d", BlackID: "c", Result: chesspairing.ResultBlackWins},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultDraw},
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultBlackWins},
				{WhiteID: "c", BlackID: "d", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "d", BlackID: "c", Result: chesspairing.ResultForfeitBlackWins},
			}},
		},
		CurrentRound: 3,
	}
	participants := mustBuildParticipantStates(t, state)
	a := findParticipant(participants, "a")
	b := findParticipant(participants, "b")
	c := findParticipant(participants, "c")
	if got, want := a.ColorHistory, []Color{ColorWhite, ColorBlack}; !equalColors(got, want) {
		t.Errorf("a colour history = %v, want %v", got, want)
	}
	if got, want := b.ColorHistory, []Color{ColorBlack, ColorWhite}; !equalColors(got, want) {
		t.Errorf("b colour history = %v, want %v", got, want)
	}
	if a.MatchesPlayed != 2 || len(a.Opponents) != 2 {
		t.Errorf("a played = %d, opponents = %v; want two regular matches", a.MatchesPlayed, a.Opponents)
	}
	if c.ColorHistory[0] != ColorWhite || c.ColorHistory[1] != ColorNone {
		t.Errorf("c colour history = %v, want [White None]", c.ColorHistory)
	}
	if c.MatchesPlayed != 1 || len(c.Opponents) != 1 {
		t.Errorf("c single-forfeit match = played %d, opponents %v; want regular match", c.MatchesPlayed, c.Opponents)
	}
	if findParticipant(participants, "d").PABIneligible.FullPointUnplayed {
		t.Error("d should remain eligible after forfeiting one game of a regular match")
	}
	if !c.PABIneligible.FullPointUnplayed {
		t.Error("c should be ineligible after d forfeited both games")
	}
}

func TestBuildParticipantStates_SingleForfeitGameDoesNotBarPAB(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		Rounds: []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{
			{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultForfeitWhiteWins},
			{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultDraw},
		}}},
		CurrentRound: 2,
	}
	a := findParticipant(mustBuildParticipantStates(t, state), "a")
	if a.PABIneligible.FullPointUnplayed {
		t.Error("a should remain eligible after winning only one forfeited game")
	}
}

func TestBuildParticipantStates_MatchesPlayedExcludesForfeits(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				// IsForfeit marks unplayed games even when the result is not encoded as a forfeit.
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultWhiteWins, IsForfeit: true},
				{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultBlackWins, IsForfeit: true},
			},
		}},
		CurrentRound: 2,
	}
	participants := mustBuildParticipantStates(t, state)
	for _, participant := range participants {
		if participant.ID == "a" || participant.ID == "b" {
			if participant.MatchesPlayed != 0 {
				t.Errorf("%s played %d matches, want 0 after a forfeited match", participant.ID, participant.MatchesPlayed)
			}
			if len(participant.Opponents) != 0 {
				t.Errorf("%s has opponents %v, want none after a forfeited match", participant.ID, participant.Opponents)
			}
			if got, want := participant.ColorHistory, []Color{ColorNone}; !equalColors(got, want) {
				t.Errorf("%s colour history = %v, want %v", participant.ID, got, want)
			}
		}
	}
}

func TestBuildParticipantStates_CrossForfeitAndForfeitPendingRegularMatch(t *testing.T) {
	// C.04.5 Preface: neither match ends by forfeit (no player forfeits both
	// games), so both are regularly played for C1 and matches played. Art. 1.6
	// and C.04.2 3.4: without an actually played game there is no colour.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Games: []chesspairing.GameData{
				// a vs b: each player forfeits one game (cross-forfeited, neither game played).
				{WhiteID: "a", BlackID: "b", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "b", BlackID: "a", Result: chesspairing.ResultForfeitWhiteWins},
				// c vs d: one forfeit, one pending (neither game played).
				{WhiteID: "c", BlackID: "d", Result: chesspairing.ResultForfeitWhiteWins},
				{WhiteID: "d", BlackID: "c", Result: chesspairing.ResultPending},
			},
		}},
		CurrentRound: 2,
	}
	partner := map[string]string{"a": "b", "b": "a", "c": "d", "d": "c"}
	participants := mustBuildParticipantStates(t, state)
	for _, id := range []string{"a", "b", "c", "d"} {
		p := findParticipant(participants, id)
		if got, want := p.ColorHistory, []Color{ColorNone}; !equalColors(got, want) {
			t.Errorf("%s colour history = %v, want %v", id, got, want)
		}
		if p.MatchesPlayed != 1 {
			t.Errorf("%s matches played = %d, want 1", id, p.MatchesPlayed)
		}
		if got, want := p.Opponents, []string{partner[id]}; len(got) != 1 || got[0] != want[0] {
			t.Errorf("%s opponents = %v, want %v", id, got, want)
		}
	}
}

func equalColors(got, want []Color) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestBuildParticipantStates_ByeTracking(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}},
			},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)
	if len(participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(participants))
	}
	if !participants[0].ByeReceived() {
		t.Error("expected ByeReceived=true after PAB")
	}
}

func TestBuildParticipantStates_InactivePlayers(t *testing.T) {
	withdrawnAfter := 1
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300, WithdrawnAfterRound: &withdrawnAfter},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)
	if len(participants) != 1 {
		t.Fatalf("expected 1 active participant, got %d", len(participants))
	}
	if participants[0].ID != "p1" {
		t.Errorf("expected p1, got %s", participants[0].ID)
	}
}

func TestBuildParticipantStates_ColorHistory(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2400},
			{ID: "p2", DisplayName: "Bob", Rating: 2300},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 2,
	}

	participants := mustBuildParticipantStates(t, state)
	p1 := findParticipant(participants, "p1")
	if p1 == nil {
		t.Fatal("p1 not found")
	}
	if len(p1.ColorHistory) != 1 || p1.ColorHistory[0] != ColorWhite {
		t.Errorf("p1 color history: expected [White], got %v", p1.ColorHistory)
	}
}

func TestBuildParticipantStates_PABIneligibility(t *testing.T) {
	tests := []struct {
		name       string
		round      chesspairing.RoundData
		playerIDs  []string
		ineligible PABIneligibility
	}{
		{
			name: "forfeit winner without flag",
			round: chesspairing.RoundData{
				Number: 1,
				Games: []chesspairing.GameData{{
					WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: false,
				}},
			},
			playerIDs:  []string{"p1"},
			ineligible: PABIneligibility{FullPointUnplayed: true},
		},
		{
			name: "double forfeit",
			round: chesspairing.RoundData{
				Number: 1,
				Games: []chesspairing.GameData{{
					WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDoubleForfeit,
				}},
			},
			playerIDs: []string{"p1", "p2"},
		},
		{
			name: "full-point bye",
			round: chesspairing.RoundData{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}},
			},
			playerIDs:  []string{"p1"},
			ineligible: PABIneligibility{FullPointUnplayed: true},
		},
		{
			name: "PAB",
			round: chesspairing.RoundData{
				Number: 1,
				Byes:   []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByePAB}},
			},
			playerIDs:  []string{"p1"},
			ineligible: PABIneligibility{PriorPAB: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &chesspairing.TournamentState{
				Players:      []chesspairing.PlayerEntry{{ID: "p1"}, {ID: "p2"}},
				Rounds:       []chesspairing.RoundData{tt.round},
				CurrentRound: 2,
			}
			participants := mustBuildParticipantStates(t, state)
			for _, playerID := range tt.playerIDs {
				participant := findParticipant(participants, playerID)
				if participant == nil {
					t.Fatalf("%s not found", playerID)
				}
				if participant.PABIneligible != tt.ineligible {
					t.Errorf("%s PABIneligible = %+v, want %+v", playerID, participant.PABIneligible, tt.ineligible)
				}
			}
		})
	}
}

// findParticipant is a test helper to locate a participant by ID.
func findParticipant(participants []ParticipantState, id string) *ParticipantState {
	for i := range participants {
		if participants[i].ID == id {
			return &participants[i]
		}
	}
	return nil
}

func mustBuildParticipantStates(t *testing.T, state *chesspairing.TournamentState) []ParticipantState {
	t.Helper()
	participants, err := BuildParticipantStates(state)
	if err != nil {
		t.Fatal(err)
	}
	return participants
}
