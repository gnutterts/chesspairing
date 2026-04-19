// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// allByeTypes returns every defined ByeType value. Adding a new ByeType
// to the enum without updating tiebreakers will surface here as a test
// gap rather than a silent default.
func allByeTypes() []chesspairing.ByeType {
	out := make([]chesspairing.ByeType, 0, 8)
	for b := chesspairing.ByePAB; b.IsValid(); b++ {
		out = append(out, b)
	}
	return out
}

// stateWithSingleBye builds a 2-round, 2-player tournament where p1
// gets a bye of the given type in round 1 and p2 plays nobody (absent).
// Round 2 has no games or byes. Used to isolate per-bye-type behaviour.
func stateWithSingleBye(byeType chesspairing.ByeType) *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p1", Type: byeType},
				},
			},
			{Number: 2},
		},
	}
}

func scoresFor(ids ...string) []chesspairing.PlayerScore {
	out := make([]chesspairing.PlayerScore, len(ids))
	for i, id := range ids {
		out[i] = chesspairing.PlayerScore{PlayerID: id}
	}
	return out
}

// TestRoundsPlayedAllByeTypes verifies that every defined ByeType is
// handled. PAB counts as played; everything else counts as unplayed.
// Total rounds = 2; p1 has a bye in round 1 and is absent in round 2.
func TestRoundsPlayedAllByeTypes(t *testing.T) {
	rp := &RoundsPlayed{}
	ctx := context.Background()

	for _, bt := range allByeTypes() {
		t.Run(bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(bt)
			values, err := rp.Compute(ctx, state, scoresFor("p1", "p2"))
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}
			vm := valueMap(values)

			// p1: round 1 bye, round 2 absent -> 1 unplayed for absence,
			// plus 1 if the bye itself counts as unplayed.
			var want float64
			if bt == chesspairing.ByePAB {
				want = 1.0 // played round 1, absent round 2
			} else {
				want = 0.0 // both rounds unplayed
			}
			if got := vm["p1"]; got != want {
				t.Errorf("p1 with %v: rounds-played = %v, want %v", bt, got, want)
			}
			// p2 is absent both rounds.
			if got := vm["p2"]; got != 0.0 {
				t.Errorf("p2: rounds-played = %v, want 0", got)
			}
		})
	}
}

// TestProgressiveAllByeTypes exercises the round-score builder that
// progressive (and other tiebreakers) depend on. PAB=1, Half=0.5,
// everything else=0.
func TestProgressiveAllByeTypes(t *testing.T) {
	pr := &Progressive{}
	ctx := context.Background()

	for _, bt := range allByeTypes() {
		t.Run(bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(bt)
			values, err := pr.Compute(ctx, state, scoresFor("p1", "p2"))
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}
			vm := valueMap(values)

			// p1's round-1 score becomes the cumulative for both rounds.
			// progressive = round1_cum + round2_cum = 2 * round1_score.
			var roundScore float64
			switch bt {
			case chesspairing.ByePAB:
				roundScore = 1.0
			case chesspairing.ByeHalf:
				roundScore = 0.5
			default:
				roundScore = 0.0
			}
			want := 2 * roundScore
			if got := vm["p1"]; got != want {
				t.Errorf("p1 with %v: progressive = %v, want %v", bt, got, want)
			}
		})
	}
}

// TestStandardPointsAllByeTypes verifies the unplayed-round comparison
// against the draw value (0.5). PAB(1.0)>0.5 -> +1, Half(0.5)=0.5 -> +0.5,
// rest(0)<0.5 -> 0.
func TestStandardPointsAllByeTypes(t *testing.T) {
	sp := &StandardPoints{}
	ctx := context.Background()

	for _, bt := range allByeTypes() {
		t.Run(bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(bt)
			values, err := sp.Compute(ctx, state, scoresFor("p1", "p2"))
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}
			vm := valueMap(values)

			var want float64
			switch bt {
			case chesspairing.ByePAB:
				want = 1.0
			case chesspairing.ByeHalf:
				want = 0.5
			default:
				want = 0.0
			}
			if got := vm["p1"]; got != want {
				t.Errorf("p1 with %v: standard-points = %v, want %v", bt, got, want)
			}
		})
	}
}

// TestWinAllByeTypes locks in the PAB-only contract. Only PAB awards
// a win; every other bye type contributes zero, regardless of how the
// scorer values it.
func TestWinAllByeTypes(t *testing.T) {
	w := &Win{}
	ctx := context.Background()

	for _, bt := range allByeTypes() {
		t.Run(bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(bt)
			values, err := w.Compute(ctx, state, scoresFor("p1", "p2"))
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}
			vm := valueMap(values)

			var want float64
			if bt == chesspairing.ByePAB {
				want = 1.0
			}
			if got := vm["p1"]; got != want {
				t.Errorf("p1 with %v: win = %v, want %v", bt, got, want)
			}
		})
	}
}

// TestOpponentDataByeTypeBucketing verifies that buildOpponentData
// preserves the bye type (rather than collapsing all types into a
// single counter as v0.1 did). Each ByeType lands in its own bucket
// of playerByes[player][type].
func TestOpponentDataByeTypeBucketing(t *testing.T) {
	for _, bt := range allByeTypes() {
		t.Run(bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(bt)
			data := buildOpponentData(state, scoresFor("p1", "p2"))

			if data.playerByes["p1"] == nil {
				t.Fatal("playerByes[p1] is nil; expected one entry")
			}
			if got := data.playerByes["p1"][bt]; got != 1 {
				t.Errorf("playerByes[p1][%v] = %d, want 1", bt, got)
			}
			// No other bye type should be incremented.
			for _, other := range allByeTypes() {
				if other == bt {
					continue
				}
				if got := data.playerByes["p1"][other]; got != 0 {
					t.Errorf("playerByes[p1][%v] = %d, want 0", other, got)
				}
			}
		})
	}
}

// TestCountsAsPlayedByeTypes verifies that PAB, Half and Zero count
// as played while Absent, Excused and ClubCommitment do not, per the
// v0.2.0 bye-type semantics matrix.
func TestCountsAsPlayedByeTypes(t *testing.T) {
	cases := []struct {
		bt   chesspairing.ByeType
		want int
	}{
		{chesspairing.ByePAB, 1},
		{chesspairing.ByeHalf, 1},
		{chesspairing.ByeZero, 1},
		{chesspairing.ByeAbsent, 0},
		{chesspairing.ByeExcused, 0},
		{chesspairing.ByeClubCommitment, 0},
	}
	for _, c := range cases {
		t.Run(c.bt.String(), func(t *testing.T) {
			state := stateWithSingleBye(c.bt)
			data := buildOpponentData(state, scoresFor("p1", "p2"))
			if got := data.countsAsPlayed("p1"); got != c.want {
				t.Errorf("countsAsPlayed(p1) with %v = %d, want %d", c.bt, got, c.want)
			}
		})
	}
}
