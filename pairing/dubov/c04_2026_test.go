// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dubov

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// TestFIDEArticle433ShifterOrder verifies the published middle-first example.
func TestFIDEArticle433ShifterOrder(t *testing.T) {
	players := []*swisslib.PlayerState{
		{ID: "A", TPN: 1},
		{ID: "B", TPN: 2},
		{ID: "C", TPN: 3},
		{ID: "D", TPN: 4},
		{ID: "E", TPN: 5},
		{ID: "F", TPN: 6},
		{ID: "G", TPN: 7},
	}
	order := shifterOrder(players, nil, false)
	want := []string{"D", "C", "E", "B", "F", "A", "G"}
	for i, player := range order {
		if player.ID != want[i] {
			t.Fatalf("article 4.3.3 number %d = %s, want %s", i+1, player.ID, want[i])
		}
	}
}

func TestPairRequiresTotalRounds(t *testing.T) {
	_, err := New(Options{}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		CurrentRound: 1,
	})
	if err == nil || err.Error() != "dubov: total rounds required" {
		t.Fatalf("error = %v, want total-rounds error", err)
	}
}
