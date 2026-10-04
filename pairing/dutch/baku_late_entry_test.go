// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBakuLateEntryGroupA(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:      bakuLateEntryPlayers(2500),
		CurrentRound: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Notes, "\n"), "GA=5 players") {
		t.Errorf("notes = %v, want GA=5", result.Notes)
	}
	groupA := map[string]bool{"p1": true, "p2": true, "p3": true, "p4": true}
	for _, pairing := range result.Pairings {
		if groupA[pairing.WhiteID] != groupA[pairing.BlackID] {
			t.Errorf("GA-GB pairing: %+v", pairing)
		}
	}
}

func TestBakuLateEntryBelowGroupA(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:      bakuLateEntryPlayers(1500),
		CurrentRound: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Notes, "\n"), "GA=4 players") {
		t.Errorf("notes = %v, want GA=4", result.Notes)
	}
}

func TestBakuLateEntryAboveLastGroupAPlayer(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	players := bakuLateEntryPlayers(2500)
	for i := range players[:8] {
		players[i].PairingNumber = i + 2
	}
	players[8].PairingNumber = 1
	result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
		Players:      players,
		CurrentRound: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(result.Notes, "\n"), "GA=5 players") {
		t.Errorf("notes = %v, want GA=5", result.Notes)
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

func bakuLateEntryPlayers(lateRating int) []chesspairing.PlayerEntry {
	players := make([]chesspairing.PlayerEntry, 8)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{ID: "p" + string(rune('1'+i)), Rating: 2400 - i*100}
	}
	return append(players, chesspairing.PlayerEntry{ID: "late", Rating: lateRating, JoinedRound: 3})
}

// C.04.7 1.4.2: a group A player gets the points of a win. The pairing scores
// are on the 1-½-0 scale, so 2-1-0 scoring must pair exactly like 1-½-0.
func TestBakuVirtualPointsFollowPairingScale(t *testing.T) {
	acceleration := "baku"
	totalRounds := 9
	players := make([]chesspairing.PlayerEntry, 8)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{ID: "p" + string(rune('1'+i)), Rating: 2400 - i*100}
	}
	history := []chesspairing.RoundData{{
		Number: 1,
		Games: []chesspairing.GameData{
			{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultBlackWins},
			{WhiteID: "p4", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "p5", BlackID: "p7", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "p8", BlackID: "p6", Result: chesspairing.ResultWhiteWins},
		},
	}}
	pair := func(options map[string]any) []chesspairing.GamePairing {
		result, err := New(Options{Acceleration: &acceleration, TotalRounds: &totalRounds}).Pair(context.Background(), &chesspairing.TournamentState{
			Players:       players,
			Rounds:        history,
			CurrentRound:  2,
			ScoringConfig: chesspairing.ScoringConfig{System: chesspairing.ScoringStandard, Options: options},
		})
		if err != nil {
			t.Fatal(err)
		}
		return result.Pairings
	}
	standardPairs := pair(nil)
	doubledPairs := pair(map[string]any{"pointWin": 2.0, "pointDraw": 1.0})
	if fmt.Sprint(standardPairs) != fmt.Sprint(doubledPairs) {
		t.Errorf("2-1-0 pairs %v, 1-½-0 pairs %v", doubledPairs, standardPairs)
	}
}
