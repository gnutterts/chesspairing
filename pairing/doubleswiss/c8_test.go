// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestPairAllBrackets_DoesNotFilterColours(t *testing.T) {
	participants := c8Participants()
	participants[0].ColorHistory = []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorWhite}
	participants[1].ColorHistory = []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}
	participants[2].ColorHistory = []lexswiss.Color{lexswiss.ColorWhite, lexswiss.ColorWhite}
	participants[3].ColorHistory = []lexswiss.Color{lexswiss.ColorBlack, lexswiss.ColorBlack}
	pairs, err := pairAllBrackets(context.Background(), []lexswiss.ScoreGroup{{Participants: participants}}, nil, false)
	if err != nil {
		t.Fatalf("pairAllBrackets() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p3"} {
		t.Errorf("first pair = %v, want [p1 p3] despite both players having White twice", got)
	}
}

func TestPairBracket_C8MinimizesBeforeIdentifier(t *testing.T) {
	participants := c8Participants()
	participants[0].WasFloater = true
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, map[[2]string]bool{{"p3", "p4"}: true}, nil, c8CostForBracket(participants[2:3]), nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p4"} {
		t.Errorf("first pair = %v, want [p1 p4] to avoid C8", got)
	}
}

func TestPairBracket_C8CountsEachUpfloaterOpponent(t *testing.T) {
	participants := c8Participants()
	for _, participant := range participants {
		participant.WasFloater = true
	}
	pairs, err := lexswiss.PairBracketMinimizing(context.Background(), participants, nil, nil, c8CostForBracket(participants[2:]), nil)
	if err != nil {
		t.Fatalf("PairBracketMinimizing() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p3"} {
		t.Errorf("first pair = %v, want [p1 p3], the first identifier at the minimum C8 count", got)
	}
}

func TestPairAllBrackets_C8IsOffInLastRound(t *testing.T) {
	participants := c8Participants()
	for i := range participants[:3] {
		participants[i].Score = 1
	}
	participants[1].WasFloater = true
	groups := []lexswiss.ScoreGroup{{Participants: participants[:3]}, {Participants: participants[3:]}}

	pairs, err := pairAllBrackets(context.Background(), groups, nil, true)
	if err != nil {
		t.Fatalf("last-round pairAllBrackets() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p3"} {
		t.Errorf("last-round first pair = %v, want [p1 p3], the first legal identifier", got)
	}

	pairs, err = pairAllBrackets(context.Background(), groups, nil, false)
	if err != nil {
		t.Fatalf("non-last-round pairAllBrackets() error: %v", err)
	}
	if got := [2]string{pairs[0][0].ID, pairs[0][1].ID}; got != [2]string{"p1", "p4"} {
		t.Errorf("non-last-round first pair = %v, want [p1 p4], the minimum C8 pairing", got)
	}
}

func c8Participants() []*lexswiss.ParticipantState {
	return []*lexswiss.ParticipantState{
		{ID: "p1", TPN: 1},
		{ID: "p2", TPN: 2},
		{ID: "p3", TPN: 3},
		{ID: "p4", TPN: 4},
	}
}
