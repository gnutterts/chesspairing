package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func d2State() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", DisplayName: "P1", Rating: 2600}, {ID: "p2", DisplayName: "P2", Rating: 2500}, {ID: "p3", DisplayName: "P3", Rating: 2400}, {ID: "p4", DisplayName: "P4", Rating: 2300}, {ID: "p5", DisplayName: "P5", Rating: 2200}, {ID: "p6", DisplayName: "P6", Rating: 2100}},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p5", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p6", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDraw}}},
			{Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p6", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p5", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultDoubleForfeit, IsForfeit: true}}},
		},
		CurrentRound: 3,
	}
}

func paired(result *chesspairing.PairingResult, first, second string) bool {
	for _, pairing := range result.Pairings {
		if (pairing.WhiteID == first && pairing.BlackID == second) ||
			(pairing.WhiteID == second && pairing.BlackID == first) {
			return true
		}
	}
	return false
}

// FIDE C.04.3 art. 1.8 applies the topscorer exception only in the final round.
func TestD2_PlannedRoundsDelayTopscorers(t *testing.T) {
	totalRounds := 9
	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), d2State())
	if err != nil {
		t.Fatal(err)
	}
	if paired(result, "p1", "p2") {
		t.Fatalf("p1 and p2 must not meet before the final round: %v", result.Pairings)
	}
	if !paired(result, "p1", "p3") || !paired(result, "p2", "p4") {
		t.Fatalf("expected legal p1-p3 and p2-p4 pairings, got %v", result.Pairings)
	}
}

func TestD2_UnknownTotalRoundsDoesNotApplyTopscorers(t *testing.T) {
	result, err := New(Options{}).Pair(context.Background(), d2State())
	if err != nil {
		t.Fatal(err)
	}
	if paired(result, "p1", "p2") {
		t.Fatalf("p1 and p2 must not meet when total rounds are unknown: %v", result.Pairings)
	}
}

func TestD2_FinalRoundAppliesTopscorers(t *testing.T) {
	totalRounds := 3
	result, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), d2State())
	if err != nil {
		t.Fatal(err)
	}
	if !paired(result, "p1", "p2") {
		t.Fatalf("expected final-round topscorer pairing p1-p2, got %v", result.Pairings)
	}
}

// FIDE C.04.7 art. 1.4.2 assigns GA players one virtual point in round 3 of a
// nine-round tournament.
func TestD2_BakuUsesPlannedTournamentLength(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "p1", DisplayName: "P1", Rating: 2600}, {ID: "p2", DisplayName: "P2", Rating: 2500}, {ID: "p3", DisplayName: "P3", Rating: 2400}, {ID: "p4", DisplayName: "P4", Rating: 2300}}, CurrentRound: 3,
		Rounds: []chesspairing.RoundData{{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins}}}, {Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}}}},
	}
	result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Notes, " "), "VP=1.0") {
		t.Fatalf("expected nine-round VP=1.0, got notes=%v", result.Notes)
	}
}

func TestD2_BakuRequiresTotalRounds(t *testing.T) {
	acceleration := "baku"
	_, err := New(Options{Acceleration: &acceleration}).Pair(context.Background(), d2State())
	if err == nil || err.Error() != "dutch: Baku acceleration requires total rounds" {
		t.Fatalf("expected missing-total-rounds Baku error, got %v", err)
	}
}

func TestD2_ValidatesTotalRounds(t *testing.T) {
	for _, totalRounds := range []int{0, -1} {
		_, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), d2State())
		if err == nil {
			t.Errorf("totalRounds %d: expected error", totalRounds)
		}
	}

	totalRounds := 2
	_, err := New(Options{TotalRounds: &totalRounds}).Pair(context.Background(), d2State())
	want := "dutch: current round 3 exceeds total rounds 2"
	if err == nil || err.Error() != want {
		t.Fatalf("expected %q, got %v", want, err)
	}
}

func TestD2_RejectsNonIntegralTotalRounds(t *testing.T) {
	_, err := NewFromMap(map[string]any{"totalRounds": 9.9}).Pair(context.Background(), d2State())
	if err == nil || err.Error() != "dutch: total rounds must be an integer" {
		t.Fatalf("expected non-integral total-rounds error, got %v", err)
	}
}
