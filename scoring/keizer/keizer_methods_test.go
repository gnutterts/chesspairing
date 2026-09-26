// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"reflect"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestOptionsMethodDefault(t *testing.T) {
	o := Options{}.WithDefaults(10)
	if o.Method == nil || *o.Method != methodIterative {
		t.Fatalf("Method = %v, want %q", o.Method, methodIterative)
	}
}

func TestOptionsWithdrawnAsAbsentDefault(t *testing.T) {
	o := Options{}.WithDefaults(10)
	if o.WithdrawnAsAbsent == nil || *o.WithdrawnAsAbsent {
		t.Fatalf("WithdrawnAsAbsent default = %v, want false", o.WithdrawnAsAbsent)
	}
}

func TestOptionsForfeitCountsAsMetDefault(t *testing.T) {
	o := Options{}.WithDefaults(10)
	if o.ForfeitCountsAsMet == nil || !*o.ForfeitCountsAsMet {
		t.Fatalf("ForfeitCountsAsMet default = %v, want true", o.ForfeitCountsAsMet)
	}
}

func TestParseOptionsMethod(t *testing.T) {
	o := ParseOptions(map[string]any{"method": methodKeizer1956})
	if o.Method == nil || *o.Method != methodKeizer1956 {
		t.Fatalf("Method = %v, want %q", o.Method, methodKeizer1956)
	}
}

func TestParseOptionsWithdrawnAsAbsent(t *testing.T) {
	o := ParseOptions(map[string]any{"withdrawnAsAbsent": true})
	if o.WithdrawnAsAbsent == nil || !*o.WithdrawnAsAbsent {
		t.Fatalf("WithdrawnAsAbsent = %v, want true", o.WithdrawnAsAbsent)
	}
}

func TestParseOptionsForfeitCountsAsMet(t *testing.T) {
	o := ParseOptions(map[string]any{"forfeitCountsAsMet": false})
	if o.ForfeitCountsAsMet == nil || *o.ForfeitCountsAsMet {
		t.Fatalf("ForfeitCountsAsMet = %v, want false", o.ForfeitCountsAsMet)
	}
}

// TestMethodFrozenEquivalentToFrozenFlag verifies that Method="frozen" and the
// deprecated Frozen=true select the same scoring path.
func TestMethodFrozenEquivalentToFrozenFlag(t *testing.T) {
	no := false
	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000},
		{ID: "p2", DisplayName: "Bob", Rating: 1800},
		{ID: "p3", DisplayName: "Carol", Rating: 1600},
	}
	rounds := []chesspairing.RoundData{
		{
			Number: 1,
			Games:  []chesspairing.GameData{{WhiteID: "p3", BlackID: "p1", Result: chesspairing.ResultWhiteWins}},
			Byes:   []chesspairing.ByeEntry{{PlayerID: "p2", Type: chesspairing.ByePAB}},
		},
		{
			Number: 2,
			Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins}},
		},
	}
	state := &chesspairing.TournamentState{Players: players, Rounds: rounds}

	frozen := true
	viaFlag, err := New(Options{SelfVictory: &no, Frozen: &frozen}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}

	method := methodFrozen
	viaMethod, err := New(Options{SelfVictory: &no, Method: &method}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(viaFlag, viaMethod) {
		t.Fatalf("Method=%q and Frozen=true differ: flag=%+v method=%+v", methodFrozen, viaFlag, viaMethod)
	}
}

// TestScoreWithdrawnAsAbsent verifies that a withdrawn player stays in the
// standings and scores absence points from the round after withdrawal, while
// earlier game points of all players are preserved.
func TestScoreWithdrawnAsAbsent(t *testing.T) {
	no := false
	frozen := true
	absentFixed := 5
	withdrawnAsAbsent := true

	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "Alice", Rating: 2000},
		{ID: "B", DisplayName: "Bob", Rating: 1800},
		{ID: "X", DisplayName: "Xavier", Rating: 1600},
	}
	withdrawnAfter := 1
	players[2].WithdrawnAfterRound = &withdrawnAfter

	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games:  []chesspairing.GameData{{WhiteID: "A", BlackID: "X", Result: chesspairing.ResultWhiteWins}},
				Byes:   []chesspairing.ByeEntry{{PlayerID: "B", Type: chesspairing.ByeZero}},
			},
			{
				Number: 2,
				Games:  []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins}},
			},
		},
		CurrentRound: 2,
	}

	opts := Options{
		SelfVictory:       &no,
		Frozen:            &frozen,
		AbsentFixedValue:  &absentFixed,
		WithdrawnAsAbsent: &withdrawnAsAbsent,
	}
	scores, err := New(opts).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores (withdrawn X stays in the standings), got %d", len(scores))
	}
	sm := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		sm[s.PlayerID] = s
	}
	// A: round 1 win over X (1) + round 2 win over B (2) = 3.
	assertScore(t, sm, "A", 3.0)
	assertScore(t, sm, "B", 0.0)
	// X: round 1 loss (0) + round 2 absence (5) = 5.
	assertScore(t, sm, "X", 5.0)
}

// TestScoreForfeitCountsAsMetDefault documents the existing default: a double
// forfeit counts as an encounter, so both players receive DoubleForfeitFraction
// (0 by default) and are not scored as absent.
func TestScoreForfeitCountsAsMetDefault(t *testing.T) {
	no := false
	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "Alice", Rating: 2000},
		{ID: "B", DisplayName: "Bob", Rating: 1800},
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDoubleForfeit}}},
		},
	}
	scores, err := New(Options{SelfVictory: &no}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	sm := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		sm[s.PlayerID] = s
	}
	assertScore(t, sm, "A", 0.0)
	assertScore(t, sm, "B", 0.0)
}

// TestScoreForfeitCountsAsMetFalse verifies that a double forfeit does not
// count as an encounter, so both players receive absence points instead.
func TestScoreForfeitCountsAsMetFalse(t *testing.T) {
	no := false
	absentFixed := 7
	forfeitCountsAsMet := false
	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "Alice", Rating: 2000},
		{ID: "B", DisplayName: "Bob", Rating: 1800},
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDoubleForfeit}}},
		},
	}
	scores, err := New(Options{
		SelfVictory:        &no,
		AbsentFixedValue:   &absentFixed,
		ForfeitCountsAsMet: &forfeitCountsAsMet,
	}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	sm := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		sm[s.PlayerID] = s
	}
	assertScore(t, sm, "A", 7.0)
	assertScore(t, sm, "B", 7.0)
}

func TestESGOptions(t *testing.T) {
	o := ESGOptions()
	checks := []struct {
		name string
		ok   bool
	}{
		{"Method=frozen", o.Method != nil && *o.Method == methodFrozen},
		{"ValueNumberBase=60", o.ValueNumberBase != nil && *o.ValueNumberBase == 60},
		{"ValueNumberStep=1", o.ValueNumberStep != nil && *o.ValueNumberStep == 1},
		{"WinFraction=1.0", o.WinFraction != nil && *o.WinFraction == 1.0},
		{"DrawFraction=0.5", o.DrawFraction != nil && *o.DrawFraction == 0.5},
		{"LossFraction=0.0", o.LossFraction != nil && *o.LossFraction == 0.0},
		{"SelfVictory=false", o.SelfVictory != nil && !*o.SelfVictory},
		{"AbsentFixedValue=20", o.AbsentFixedValue != nil && *o.AbsentFixedValue == 20},
		{"AbsenceLimit=5", o.AbsenceLimit != nil && *o.AbsenceLimit == 5},
		{"AbsenceDecay=false", o.AbsenceDecay != nil && !*o.AbsenceDecay},
		{"ClubCommitmentFixedValue=40", o.ClubCommitmentFixedValue != nil && *o.ClubCommitmentFixedValue == 40},
		{"ByeFixedValue=40", o.ByeFixedValue != nil && *o.ByeFixedValue == 40},
		{"LateJoinHandicap=15", o.LateJoinHandicap != nil && *o.LateJoinHandicap == 15.0},
		{"WithdrawnAsAbsent=true", o.WithdrawnAsAbsent != nil && *o.WithdrawnAsAbsent},
		{"ForfeitCountsAsMet=false", o.ForfeitCountsAsMet != nil && !*o.ForfeitCountsAsMet},
	}
	for _, check := range checks {
		if !check.ok {
			t.Errorf("ESGOptions %s not satisfied", check.name)
		}
	}
}

// TestESGDoubleForfeitScoresAbsenceForBoth verifies that under ESGOptions a
// double forfeit is not an encounter: both players score absence points and
// the round counts toward AbsenceLimit.
func TestESGDoubleForfeitScoresAbsenceForBoth(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "Alice", Rating: 2000},
		{ID: "B", DisplayName: "Bob", Rating: 1800},
	}
	state := &chesspairing.TournamentState{
		Players: players,
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDoubleForfeit}}},
		},
	}
	scores, err := New(ESGOptions()).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	sm := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		sm[s.PlayerID] = s
	}
	// AbsentFixedValue=20 and SelfVictory=false under ESG: both score 20.
	assertScore(t, sm, "A", 20.0)
	assertScore(t, sm, "B", 20.0)
}

// TestESGDoubleForfeitCountsTowardAbsenceLimit verifies that under ESGOptions
// a double-forfeit round counts toward AbsenceLimit: with AbsenceLimit=5 only
// the first five such rounds score absence points.
func TestESGDoubleForfeitCountsTowardAbsenceLimit(t *testing.T) {
	players := []chesspairing.PlayerEntry{
		{ID: "A", DisplayName: "Alice", Rating: 2000},
		{ID: "B", DisplayName: "Bob", Rating: 1800},
	}
	rounds := make([]chesspairing.RoundData, 6)
	for i := range rounds {
		rounds[i] = chesspairing.RoundData{
			Number: i + 1,
			Games:  []chesspairing.GameData{{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDoubleForfeit}},
		}
	}
	state := &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: len(rounds)}
	scores, err := New(ESGOptions()).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	sm := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		sm[s.PlayerID] = s
	}
	// 5 scored absences × 20 = 100; the sixth round is beyond AbsenceLimit.
	assertScore(t, sm, "A", 100.0)
	assertScore(t, sm, "B", 100.0)
}
