// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBakuRequiresTotalRounds(t *testing.T) {
	acceleration := "baku"
	_, err := New(Options{Acceleration: &acceleration}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		CurrentRound: 1,
	})
	if err == nil || err.Error() != "burstein: Baku acceleration requires total rounds" {
		t.Errorf("Pair() error = %v", err)
	}
}

func TestBakuLateEntryGroupA(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	tests := []struct {
		name       string
		lateRating int
		wantBye    string
	}{
		{name: "above Group A", lateRating: 2500, wantBye: "p7"},
		{name: "below Group A", lateRating: 1500, wantBye: "late"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
				Players:      bursteinBakuLateEntryPlayers(tt.lateRating),
				Rounds:       bursteinBakuHistory(),
				CurrentRound: 5,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Byes) != 1 || result.Byes[0].PlayerID != tt.wantBye {
				t.Errorf("byes = %+v, want PAB for %q", result.Byes, tt.wantBye)
			}
		})
	}
}

func TestBakuRejectsFootballScoring(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	_, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:       []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		CurrentRound:  1,
		ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringFootball},
	})
	if err == nil || !strings.Contains(err.Error(), "C.04.7 1.1") {
		t.Errorf("Pair() error = %v, want C.04.7 1.1", err)
	}
}

func bursteinBakuLateEntryPlayers(lateRating int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, 8)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{ID: "p" + string(rune('1'+i)), Rating: 2400 - i*100}
	}
	return append(players, chesspairing.PlayerEntry{ID: "late", Rating: lateRating, JoinedRound: 3})
}

func bursteinBakuHistory() []chesspairing.RoundData {
	draw := chesspairing.ResultDraw
	return []chesspairing.RoundData{
		{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: draw},
				{WhiteID: "p3", BlackID: "p4", Result: draw},
				{WhiteID: "p5", BlackID: "p6", Result: draw},
				{WhiteID: "p7", BlackID: "p8", Result: draw},
			},
		},
		{
			Number: 2,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p3", Result: draw},
				{WhiteID: "p2", BlackID: "p4", Result: draw},
				{WhiteID: "p5", BlackID: "p7", Result: draw},
				{WhiteID: "p6", BlackID: "p8", Result: draw},
			},
		},
		{
			Number: 3,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p4", Result: draw},
				{WhiteID: "p2", BlackID: "p3", Result: draw},
				{WhiteID: "p5", BlackID: "p8", Result: draw},
				{WhiteID: "late", BlackID: "p6", Result: draw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "p7", Type: chesspairing.ByeZero}},
		},
		{
			Number: 4,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p5", Result: draw},
				{WhiteID: "p2", BlackID: "p6", Result: draw},
				{WhiteID: "p3", BlackID: "p7", Result: draw},
				{WhiteID: "late", BlackID: "p8", Result: draw},
			},
			Byes: []chesspairing.ByeEntry{{PlayerID: "p4", Type: chesspairing.ByeZero}},
		},
	}
}

// Post-seeding rounds report group A like the Dutch seeding rounds do.
func TestBakuPostSeedingGroupANote(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	for lateRating, want := range map[int]string{2500: "GA=5 players", 1500: "GA=4 players"} {
		result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
			Players:      bursteinBakuLateEntryPlayers(lateRating),
			Rounds:       bursteinBakuHistory(),
			CurrentRound: 5,
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.Join(result.Notes, "\n"), want) {
			t.Errorf("late entry rated %d: notes = %v, want %s", lateRating, result.Notes, want)
		}
	}
}

// C.04.7 1.4.2/1.4.3: the virtual points are those of a win on the scale of
// the pairing score, so 2-1-0 scoring pairs exactly like 1-½-0.
func TestBakuVirtualPointsFollowPairingScale(t *testing.T) {
	acceleration := "baku"
	totalRounds := 3
	players := make([]chesspairing.PlayerEntry, 8)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{ID: "p" + string(rune('1'+i)), Rating: 2400 - i*100}
	}
	history := []chesspairing.RoundData{{
		Number: 1,
		Games: []chesspairing.GameData{
			{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultDraw},
			{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultDraw},
			{WhiteID: "p5", BlackID: "p7", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "p8", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
		},
	}}
	pair := func(options map[string]any) string {
		result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
			Players:       players,
			Rounds:        history,
			CurrentRound:  2,
			ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: options},
		})
		if err != nil {
			t.Fatal(err)
		}
		return fmt.Sprint(result.Pairings)
	}
	if standardPairs, doubledPairs := pair(nil), pair(map[string]any{"pointWin": 2.0, "pointDraw": 1.0}); standardPairs != doubledPairs {
		t.Errorf("2-1-0 pairs %s, 1-½-0 pairs %s", doubledPairs, standardPairs)
	}
}
