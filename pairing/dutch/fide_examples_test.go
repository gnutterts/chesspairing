// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

const fideMast1Source = "Mastering the Dutch, §3/round 1 (p. 11-13), §4/round 2 (p. 31-32), §5/round 3 (p. 33-37)"

func fideMast1Players() []chesspairing.PlayerEntry {
	return []chesspairing.PlayerEntry{
		{ID: "1", DisplayName: "Alice", Rating: 2509},
		{ID: "2", DisplayName: "Bruno", Rating: 2445},
		{ID: "3", DisplayName: "Carla", Rating: 2421},
		{ID: "4", DisplayName: "David", Rating: 2402},
		{ID: "5", DisplayName: "Eloise", Rating: 2309},
		{ID: "6", DisplayName: "Finn", Rating: 2307},
		{ID: "7", DisplayName: "Giorgia", Rating: 2286},
		{ID: "8", DisplayName: "Kevin", Rating: 2158},
		{ID: "9", DisplayName: "Louise", Rating: 2132},
		{ID: "10", DisplayName: "Marco", Rating: 2123},
		{ID: "11", DisplayName: "Nancy", Rating: 2116},
		{ID: "12", DisplayName: "Oskar", Rating: 2113},
		{ID: "13", DisplayName: "Patricia", Rating: 2105},
		{ID: "14", DisplayName: "Robert", Rating: 1936},
		{ID: "15", DisplayName: "Stephanie", Rating: 1900},
	}
}

func fideMast1Round1() chesspairing.RoundData {
	return chesspairing.RoundData{Number: 1, Games: []chesspairing.GameData{
		{WhiteID: "1", BlackID: "8", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "9", BlackID: "2", Result: chesspairing.ResultDraw},
		{WhiteID: "3", BlackID: "10", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "11", BlackID: "4", Result: chesspairing.ResultBlackWins},
		{WhiteID: "5", BlackID: "12", Result: chesspairing.ResultDraw},
		{WhiteID: "13", BlackID: "6", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "7", BlackID: "14", Result: chesspairing.ResultBlackWins},
	}, Byes: []chesspairing.ByeEntry{{PlayerID: "15", Type: chesspairing.ByePAB}}}
}

func fideMast1Round2() chesspairing.RoundData {
	return chesspairing.RoundData{Number: 2, Games: []chesspairing.GameData{
		{WhiteID: "14", BlackID: "1", Result: chesspairing.ResultBlackWins},
		{WhiteID: "15", BlackID: "3", Result: chesspairing.ResultBlackWins},
		{WhiteID: "4", BlackID: "13", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "2", BlackID: "5", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "12", BlackID: "9", Result: chesspairing.ResultDraw},
		{WhiteID: "10", BlackID: "7", Result: chesspairing.ResultBlackWins},
		{WhiteID: "8", BlackID: "11", Result: chesspairing.ResultWhiteWins},
	}, Byes: []chesspairing.ByeEntry{{PlayerID: "6", Type: chesspairing.ByeHalf}}}
}

func fideMast1State(round int, rounds []chesspairing.RoundData, byes []chesspairing.ByeEntry) *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players:         fideMast1Players(),
		Rounds:          rounds,
		CurrentRound:    round,
		PreAssignedByes: byes,
		PairingConfig: chesspairing.PairingConfig{
			System:  chesspairing.PairingDutch,
			Options: map[string]any{"totalRounds": 9},
		},
	}
}

func fideMast1Preference(history []swisslib.Color) string {
	var played []swisslib.Color
	for _, color := range history {
		if color != swisslib.ColorNone {
			played = append(played, color)
		}
	}
	if len(played) == 0 {
		return "A"
	}
	white, black := 0, 0
	for _, color := range played {
		if color == swisslib.ColorWhite {
			white++
		} else {
			black++
		}
	}
	wanted := "W"
	if white > black {
		wanted = "B"
	} else if white == black && played[len(played)-1] == swisslib.ColorWhite {
		wanted = "B"
	}
	if len(played) >= 2 && played[len(played)-1] == played[len(played)-2] {
		return wanted + "S"
	}
	if white-black > 1 || black-white > 1 {
		return wanted + "D"
	}
	if white != black {
		return wanted
	}
	return strings.ToLower(wanted)
}

func fideMast1VerifyTranscription(t *testing.T, state *chesspairing.TournamentState, scores map[string]float64, preferences map[string]string) {
	t.Helper()
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range players {
		if want, ok := scores[player.ID]; !ok || player.Score != want {
			t.Errorf("FIDE transcription: player %s score = %.1f, want %.1f (%s)", player.ID, player.Score, want, fideMast1Source)
		}
		if want, ok := preferences[player.ID]; ok {
			if got := fideMast1Preference(player.ColorHistory); got != want {
				t.Errorf("FIDE transcription: player %s colour preference = %s, want %s (%s)", player.ID, got, want, fideMast1Source)
			}
		}
	}
}

func fideMast1Pairs(result *chesspairing.PairingResult) string {
	parts := make([]string, 0, len(result.Pairings)+len(result.Byes))
	for _, pair := range result.Pairings {
		parts = append(parts, pair.WhiteID+"-"+pair.BlackID)
	}
	for _, bye := range result.Byes {
		parts = append(parts, bye.PlayerID+"-0")
	}
	return strings.Join(parts, ", ")
}

func fideMast1Pair(t *testing.T, id, fide string, state *chesspairing.TournamentState) {
	t.Helper()
	white := "white"
	result, err := New(Options{TopSeedColor: &white}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	got := fideMast1Pairs(result)
	if got != fide {
		t.Errorf("%s: pairings = %s, want FIDE %s", id, got, fide)
	}
}

func TestFIDEExample_mast1_1(t *testing.T) {
	state := fideMast1State(1, nil, nil)
	fideMast1VerifyTranscription(t, state,
		map[string]float64{"1": 0, "2": 0, "3": 0, "4": 0, "5": 0, "6": 0, "7": 0, "8": 0, "9": 0, "10": 0, "11": 0, "12": 0, "13": 0, "14": 0, "15": 0},
		map[string]string{"1": "A", "2": "A", "3": "A", "4": "A", "5": "A", "6": "A", "7": "A", "8": "A", "9": "A", "10": "A", "11": "A", "12": "A", "13": "A", "14": "A", "15": "A"})
	fideMast1Pair(t, "mast1_1", "1-8, 9-2, 3-10, 11-4, 5-12, 13-6, 7-14, 15-0", state)
}

func TestFIDEExample_mast1_2(t *testing.T) {
	state := fideMast1State(2, []chesspairing.RoundData{fideMast1Round1()}, []chesspairing.ByeEntry{{PlayerID: "6", Type: chesspairing.ByeHalf}})
	fideMast1VerifyTranscription(t, state,
		map[string]float64{"1": 1, "2": .5, "3": 1, "4": 1, "5": .5, "6": 0, "7": 0, "8": 0, "9": .5, "10": 0, "11": 0, "12": .5, "13": 1, "14": 1, "15": 1},
		map[string]string{"1": "B", "2": "W", "3": "B", "4": "W", "5": "B", "7": "B", "8": "W", "9": "B", "10": "W", "11": "B", "12": "W", "13": "B", "14": "W", "15": "A"})
	fideMast1Pair(t, "mast1_2", "14-1, 15-3, 4-13, 2-5, 12-9, 10-7, 8-11, 6-0", state)
}

func TestFIDEExample_mast1_3(t *testing.T) {
	state := fideMast1State(3, []chesspairing.RoundData{fideMast1Round1(), fideMast1Round2()}, []chesspairing.ByeEntry{{PlayerID: "7", Type: chesspairing.ByeHalf}})
	fideMast1VerifyTranscription(t, state,
		map[string]float64{"1": 2, "2": 1.5, "3": 2, "4": 2, "5": .5, "6": .5, "7": 1, "8": 1, "9": 1, "10": 0, "11": 0, "12": 1, "13": 1, "14": 1, "15": 1},
		map[string]string{"1": "w", "2": "b", "3": "w", "4": "b", "5": "w", "6": "W", "8": "b", "9": "w", "10": "b", "11": "w", "12": "b", "13": "w", "14": "b", "15": "B"})
	fideMast1Pair(t, "mast1_3", "1-4, 3-2, 13-8, 9-14, 12-15, 6-5, 11-10, 7-0", state)
}
