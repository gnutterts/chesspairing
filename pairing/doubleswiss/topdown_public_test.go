// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/gnutterts/chesspairing"
)

func doubleState(n int, rounds []chesspairing.RoundData) *chesspairing.TournamentState {
	players := make([]chesspairing.PlayerEntry, n)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{ID: fmt.Sprintf("p%d", i+1), Rating: 3000 - i}
	}
	return &chesspairing.TournamentState{Players: players, Rounds: rounds, CurrentRound: len(rounds) + 1, PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss}}
}

func TestTopDownPublic_EvenGroupUsesUpfloaters(t *testing.T) {
	// C.04.5 3.3.2 makes a bracket from all residents plus the necessary
	// upfloaters.  The two .5 residents have met, so 3.5 must add two players
	// (one is left by the odd top bracket and one comes from the zero group).
	games := []chesspairing.GameData{
		{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p5", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p7", BlackID: "p15", Result: chesspairing.ResultDraw},
		{WhiteID: "p8", BlackID: "p9", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p10", BlackID: "p11", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p12", BlackID: "p13", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "p14", BlackID: "p16", Result: chesspairing.ResultWhiteWins},
	}
	result, err := New(Options{}).Pair(context.Background(), doubleState(16, []chesspairing.RoundData{{Number: 1, Games: games}}))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Pairings) != 8 {
		t.Fatalf("pairs = %d, want 8", len(result.Pairings))
	}
	// Articles 3.5.2-3.5.5: the odd 1-point group takes one upfloater from
	// the 0.5 group; p7 and p15 have the same profile and p7 comes first in
	// 3.5.4 order. p7's opponent inside that bracket follows from Article
	// 3.6 and is not asserted here. p15 is then the only 0.5 resident and
	// takes the first zero-point player it has not met: p2.
	if !hasTopDownPair(result.Pairings, "p15", "p2") {
		t.Fatalf("pairs = %v, want p15-p2", result.Pairings)
	}
	onePoint := map[string]bool{"p1": true, "p3": true, "p5": true, "p8": true, "p10": true, "p12": true, "p14": true}
	for _, p := range result.Pairings {
		if p.WhiteID == "p7" && !onePoint[p.BlackID] || p.BlackID == "p7" && !onePoint[p.WhiteID] {
			t.Fatalf("p7 paired outside the 1-point bracket: %v", result.Pairings)
		}
	}
}

func TestTopDownPublic_OddAndEvenTopGroups(t *testing.T) {
	odd := doubleState(6, []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeFullPoint}, {PlayerID: "p3", Type: chesspairing.ByeFullPoint}}}})
	result, err := New(Options{}).Pair(context.Background(), odd)
	if err != nil {
		t.Fatal(err)
	}
	cross := 0
	for _, p := range result.Pairings {
		if (p.WhiteID <= "p3") != (p.BlackID <= "p3") {
			cross++
		}
	}
	if cross != 1 {
		t.Fatalf("odd top group used %d upfloaters, want 1: %v", cross, result.Pairings)
	}
	even := doubleState(4, []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeFullPoint}}}})
	result, err = New(Options{}).Pair(context.Background(), even)
	if err != nil {
		t.Fatal(err)
	}
	if !hasTopDownPair(result.Pairings, "p1", "p2") {
		t.Fatalf("even pairable top group floated: %v", result.Pairings)
	}
}

func TestTopDownPublic_C6ChoosesBestFollowingBracket(t *testing.T) {
	state := doubleState(6, []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{
		{PlayerID: "p1", Type: chesspairing.ByeFullPoint},
		{PlayerID: "p2", Type: chesspairing.ByeHalf},
		{PlayerID: "p3", Type: chesspairing.ByeHalf},
		{PlayerID: "p4", Type: chesspairing.ByeHalf},
	}}})
	result, err := New(Options{ForbiddenPairs: [][]string{{"p3", "p4"}}}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	if !hasTopDownPair(result.Pairings, "p1", "p3") || !hasTopDownPair(result.Pairings, "p2", "p4") || !hasTopDownPair(result.Pairings, "p5", "p6") {
		t.Fatalf("C6 pairs = %v, want p1-p3, p2-p4, p5-p6", result.Pairings)
	}
}

func TestTopDownPublic_C3ChoosesLaterSet(t *testing.T) {
	state := doubleState(4, []chesspairing.RoundData{{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeHalf}, {PlayerID: "p3", Type: chesspairing.ByeHalf}, {PlayerID: "p4", Type: chesspairing.ByeHalf}}}})
	result, err := New(Options{ForbiddenPairs: [][]string{{"p3", "p4"}}}).Pair(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	// p2 is first under 3.5.4, but leaves forbidden p3-p4. C.04.5 2.2.1
	// requires a pairing for everybody not yet paired, so p3 is selected.
	if !hasTopDownPair(result.Pairings, "p1", "p3") {
		t.Fatalf("C3 did not select later upfloater: %v", result.Pairings)
	}
}

// Only C1 and C2 are absolute (Article 2.1): p3 and p4 both had White twice,
// but a colour preference never makes a legal pairing impossible.
func TestTopDownPublic_ColourNeverBlocksLegalPairing(t *testing.T) {
	state := doubleState(4, []chesspairing.RoundData{
		{Number: 1, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeFullPoint}}, Games: []chesspairing.GameData{{WhiteID: "p3", BlackID: "x1", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p4", BlackID: "x2", Result: chesspairing.ResultWhiteWins}}},
		{Number: 2, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeFullPoint}}, Games: []chesspairing.GameData{{WhiteID: "p3", BlackID: "x3", Result: chesspairing.ResultWhiteWins}, {WhiteID: "p4", BlackID: "x4", Result: chesspairing.ResultWhiteWins}}},
		{Number: 3, Byes: []chesspairing.ByeEntry{{PlayerID: "p1", Type: chesspairing.ByeFullPoint}, {PlayerID: "p2", Type: chesspairing.ByeFullPoint}}},
	})
	result, err := New(Options{}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair() error = %v, want a complete pairing", err)
	}
	if !hasTopDownPair(result.Pairings, "p1", "p2") || !hasTopDownPair(result.Pairings, "p3", "p4") {
		t.Fatalf("pairs = %v, want p1-p2 and p3-p4", result.Pairings)
	}
}

func TestTopDownPublic_100PlayersTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("slow test")
	}
	state := doubleState(100, nil)
	pairer := New(Options{})
	rng := rand.New(rand.NewPCG(1, 2))
	for round := 1; round <= 9; round++ {
		state.CurrentRound = round
		started := time.Now()
		result, err := pairer.Pair(context.Background(), state)
		if err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		if !raceEnabled && time.Since(started) >= 2*time.Second {
			t.Fatalf("round %d took %s", round, time.Since(started))
		}
		games := make([]chesspairing.GameData, 0, len(result.Pairings))
		for _, p := range result.Pairings {
			r := chesspairing.ResultWhiteWins
			if rng.IntN(2) == 0 {
				r = chesspairing.ResultBlackWins
			}
			games = append(games, chesspairing.GameData{WhiteID: p.WhiteID, BlackID: p.BlackID, Result: r})
		}
		state.Rounds = append(state.Rounds, chesspairing.RoundData{Number: round, Games: games})
	}
}

func hasTopDownPair(pairs []chesspairing.GamePairing, a, b string) bool {
	for _, p := range pairs {
		if (p.WhiteID == a && p.BlackID == b) || (p.WhiteID == b && p.BlackID == a) {
			return true
		}
	}
	return false
}
