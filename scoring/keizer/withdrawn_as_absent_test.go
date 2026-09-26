// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// A withdrawal after round 2 must not remove X's already played games from
// the standings. In round 3 X is absent, while A and B receive zero-point
// byes so their wins over X remain in their scores.
func TestRepro_KZS05_WithdrawalRetainsHistoryAndScoresAbsence(t *testing.T) {
	no := false
	frozen := true
	absentFixed := 1
	withdrawnAsAbsent := true
	opts := Options{
		SelfVictory:       &no,
		Frozen:            &frozen,
		AbsentFixedValue:  &absentFixed,
		WithdrawnAsAbsent: &withdrawnAsAbsent,
	}
	players := []chesspairing.PlayerEntry{
		{ID: "A", Rating: 3},
		{ID: "B", Rating: 2},
		{ID: "X", Rating: 1},
	}
	rounds12 := []chesspairing.RoundData{
		{Number: 1, Games: []chesspairing.GameData{{WhiteID: "X", BlackID: "A", Result: chesspairing.ResultBlackWins}}, Byes: []chesspairing.ByeEntry{{PlayerID: "B", Type: chesspairing.ByeZero}}},
		{Number: 2, Games: []chesspairing.GameData{{WhiteID: "X", BlackID: "B", Result: chesspairing.ResultBlackWins}}, Byes: []chesspairing.ByeEntry{{PlayerID: "A", Type: chesspairing.ByeZero}}},
	}

	before, err := New(opts).Score(context.Background(), &chesspairing.TournamentState{Players: players, Rounds: rounds12, CurrentRound: 2})
	if err != nil {
		t.Fatal(err)
	}
	withdrawnAfter := 2
	players[2].WithdrawnAfterRound = &withdrawnAfter
	after, err := New(opts).Score(context.Background(), &chesspairing.TournamentState{
		Players: players,
		Rounds: append(rounds12, chesspairing.RoundData{Number: 3, Byes: []chesspairing.ByeEntry{
			{PlayerID: "A", Type: chesspairing.ByeZero},
			{PlayerID: "B", Type: chesspairing.ByeZero},
		}}),
		CurrentRound: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("before withdrawal=%+v; after round 3 withdrawal=%+v", before, after)

	beforeMap := scoreByID(before)
	afterMap := scoreByID(after)
	beforeA, beforeB, beforeX := beforeMap["A"].Score, beforeMap["B"].Score, beforeMap["X"].Score
	afterA, afterB, afterX := afterMap["A"].Score, afterMap["B"].Score, afterMap["X"].Score

	if _, ok := afterMap["X"]; !ok {
		t.Errorf("X is absent from the standings after withdrawing after round 2: %+v", after)
	}
	if afterX != beforeX+float64(absentFixed) {
		t.Errorf("X score after round 3 = %v, want prior score %v plus absent points %d", afterX, beforeX, absentFixed)
	}
	if afterA != beforeA || afterB != beforeB {
		t.Errorf("A and B lost or changed points from rounds 1-2: before A=%v B=%v, after A=%v B=%v", beforeA, beforeB, afterA, afterB)
	}
}

func scoreByID(scores []chesspairing.PlayerScore) map[string]chesspairing.PlayerScore {
	m := make(map[string]chesspairing.PlayerScore, len(scores))
	for _, s := range scores {
		m[s.PlayerID] = s
	}
	return m
}
