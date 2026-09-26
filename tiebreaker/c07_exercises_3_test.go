// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"math"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// fideO3Entry is one expected FIDE tiebreak value for one player.
type fideO3Entry struct {
	playerID string
	value    float64
}

// fideO3SwissState builds the C.07-2023 Swiss individual crosstable
// (exercise set pp. 5-6) used by exercises 17-22.
//
// Nick (#12, "W") withdrew before round 4, hence WithdrawnAfterRound=3.
// CurrentRound is anchored at 3 so scoring/standard still emits Nick's
// final score together with the other 15 players.
func fideO3SwissState() *chesspairing.TournamentState {
	wd := 3
	players := []chesspairing.PlayerEntry{
		{ID: "1", DisplayName: "Alyx", Rating: 2200},
		{ID: "2", DisplayName: "Bruno", Rating: 2150},
		{ID: "3", DisplayName: "Charline", Rating: 2100},
		{ID: "4", DisplayName: "David", Rating: 2050},
		{ID: "5", DisplayName: "Helene", Rating: 2000},
		{ID: "6", DisplayName: "Franck", Rating: 1950},
		{ID: "7", DisplayName: "Genevieve", Rating: 1900},
		{ID: "8", DisplayName: "Irina", Rating: 1850},
		{ID: "9", DisplayName: "Jessica", Rating: 1800},
		{ID: "10", DisplayName: "Lais", Rating: 1750},
		{ID: "11", DisplayName: "Maria", Rating: 1700},
		{ID: "12", DisplayName: "Nick", Rating: 1650, WithdrawnAfterRound: &wd},
		{ID: "13", DisplayName: "Opal", Rating: 1600},
		{ID: "14", DisplayName: "Paul", Rating: 1550},
		{ID: "15", DisplayName: "Reine", Rating: 1500},
		{ID: "16", DisplayName: "Stephan", Rating: 1450},
	}

	rounds := []chesspairing.RoundData{
		{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "10", BlackID: "2", Result: chesspairing.ResultBlackWins},
				{WhiteID: "1", BlackID: "9", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "11", BlackID: "3", Result: chesspairing.ResultDraw},
				{WhiteID: "12", BlackID: "4", Result: chesspairing.ResultBlackWins},
				{WhiteID: "16", BlackID: "8", Result: chesspairing.ResultDraw},
				{WhiteID: "14", BlackID: "6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "5", BlackID: "13", Result: chesspairing.ResultBlackWins},
				{WhiteID: "7", BlackID: "15", Result: chesspairing.ResultWhiteWins},
			},
		},
		{
			Number: 2,
			Games: []chesspairing.GameData{
				{WhiteID: "2", BlackID: "7", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "13", BlackID: "1", Result: chesspairing.ResultDraw},
				{WhiteID: "6", BlackID: "3", Result: chesspairing.ResultBlackWins},
				{WhiteID: "11", BlackID: "16", Result: chesspairing.ResultBlackWins},
				{WhiteID: "15", BlackID: "5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "8", BlackID: "14", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "9", BlackID: "10", Result: chesspairing.ResultBlackWins},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "4", Type: chesspairing.ByeHalf}, // HPB
				{PlayerID: "12", Type: chesspairing.ByePAB}, // +BYE
			},
		},
		{
			Number: 3,
			Games: []chesspairing.GameData{
				{WhiteID: "1", BlackID: "2", Result: chesspairing.ResultDraw},
				{WhiteID: "3", BlackID: "8", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "4", BlackID: "13", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "16", BlackID: "7", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "5", BlackID: "11", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "12", BlackID: "14", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
				{WhiteID: "10", BlackID: "15", Result: chesspairing.ResultBlackWins},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "6", Type: chesspairing.ByePAB},  // +BYE
				{PlayerID: "9", Type: chesspairing.ByeHalf}, // =BYE HPB
			},
		},
		{
			Number: 4,
			Games: []chesspairing.GameData{
				{WhiteID: "2", BlackID: "16", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "15", BlackID: "1", Result: chesspairing.ResultBlackWins},
				{WhiteID: "4", BlackID: "3", Result: chesspairing.ResultDraw},
				{WhiteID: "6", BlackID: "10", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "7", BlackID: "5", Result: chesspairing.ResultDraw},
				{WhiteID: "8", BlackID: "13", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "11", BlackID: "9", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "14", Type: chesspairing.ByeZero}, // ZPB
			},
		},
		{
			Number: 5,
			Games: []chesspairing.GameData{
				{WhiteID: "3", BlackID: "2", Result: chesspairing.ResultDraw},
				{WhiteID: "1", BlackID: "4", Result: chesspairing.ResultDraw},
				{WhiteID: "16", BlackID: "15", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "8", BlackID: "6", Result: chesspairing.ResultBlackWins},
				{WhiteID: "5", BlackID: "10", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "11", BlackID: "7", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "13", BlackID: "14", Result: chesspairing.ResultBlackWins},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "9", Type: chesspairing.ByePAB}, // +BYE
			},
		},
	}

	return &chesspairing.TournamentState{
		Players:      players,
		Rounds:       rounds,
		CurrentRound: 3,
	}
}

// fideO3Scores verifies the transcribed crosstable against scoring/standard
// and returns the computed scores for use by the tiebreak comparisons.
func fideO3Scores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	t.Helper()

	scorer := standard.New(standard.Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard: %v", err)
	}

	want := map[string]float64{
		"1": 3.5, "2": 4.0, "3": 3.5, "4": 3.5,
		"5": 2.5, "6": 3.0, "7": 1.5, "8": 2.5,
		"9": 1.5, "10": 1.0, "11": 2.5, "12": 2.0,
		"13": 1.5, "14": 2.0, "15": 2.0, "16": 3.5,
	}

	got := make(map[string]float64, len(scores))
	for _, s := range scores {
		got[s.PlayerID] = s.Score
	}

	for id, w := range want {
		g, ok := got[id]
		if !ok {
			t.Errorf("transcription error: player %s is missing from scoring/standard; want score %v", id, w)
			continue
		}
		if math.Abs(g-w) > 0.001 {
			t.Errorf("transcription error: player %s score = %v, want %v", id, g, w)
		}
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("transcription error: unexpected player %s in scoring/standard", id)
		}
	}

	return scores
}

// fideO3AssertTiebreak computes and asserts one FIDE tiebreak value per player.
func fideO3AssertTiebreak(t *testing.T, exercise int, abbreviation, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore, entries []fideO3Entry) {
	t.Helper()

	tb, err := getC072023(id)
	if err != nil {
		t.Fatalf("FIDE exercise %d: tiebreaker %s: %v", exercise, id, err)
	}

	vals, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("tiebreaker %s: %v", id, err)
	}

	code := make(map[string]float64, len(vals))
	for _, v := range vals {
		code[v.PlayerID] = v.Value
	}

	for _, e := range entries {
		c, ok := code[e.playerID]
		if !ok {
			t.Errorf("FIDE exercise %d %s: player %s is missing, want FIDE %v", exercise, abbreviation, e.playerID, e.value)
			continue
		}
		if math.Abs(c-e.value) > 0.001 {
			t.Errorf("FIDE exercise %d %s: player %s = %v, want FIDE %v", exercise, abbreviation, e.playerID, c, e.value)
		}
	}
}

// FIDE C.07-2023 exercise set, Exercise 17 (ARO, Swiss) — p. 33.
func TestFIDEExercise_17_ARO(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 17, "ARO", "aro", state, scores, []fideO3Entry{
		{playerID: "2", value: 1880},
		{playerID: "3", value: 1940},
		{playerID: "4", value: 1888},
		{playerID: "1", value: 1820},
		{playerID: "16", value: 1820},
		{playerID: "6", value: 1813},
		{playerID: "11", value: 1863},
		{playerID: "8", value: 1730},
		{playerID: "5", value: 1690},
		{playerID: "12", value: 2050},
		{playerID: "15", value: 1860},
		{playerID: "14", value: 1800},
		{playerID: "9", value: 1975},
		{playerID: "13", value: 1930},
		{playerID: "7", value: 1760},
		{playerID: "10", value: 1880},
	})
}

// FIDE C.07-2023 exercise set, Exercise 18 (AROC, Swiss) — pp. 33-34.
// AROC is ARO Cut-1: the contribution of the lowest rated opponent
// (the least significant result on rating) is discarded before
// averaging. Player #12 played only one rated game, so after the cut
// nothing remains and FIDE lists 0 (undefined).
func TestFIDEExercise_18_AROC(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 18, "AROC", "aro-cut1", state, scores, []fideO3Entry{
		{playerID: "2", value: 1988},
		{playerID: "3", value: 2000},
		{playerID: "4", value: 1983},
		{playerID: "1", value: 1900},
		{playerID: "16", value: 1900},
		{playerID: "6", value: 1900},
		{playerID: "11", value: 2000},
		{playerID: "8", value: 1800},
		{playerID: "5", value: 1738},
		{playerID: "15", value: 1963},
		{playerID: "14", value: 1900},
		{playerID: "12", value: 0},
		{playerID: "9", value: 2200},
		{playerID: "13", value: 2025},
		{playerID: "7", value: 1838},
		{playerID: "10", value: 1975},
	})
}

// FIDE C.07-2023 exercise set, Exercise 19 (TPR, Swiss) — p. 37.
func TestFIDEExercise_19_TPR(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 19, "TPR", "performance-rating", state, scores, []fideO3Entry{
		{playerID: "2", value: 2120},
		{playerID: "6", value: 1813},
		{playerID: "12", value: 1250},
	})
}

// FIDE C.07-2023 exercise set, Exercise 20 (TPR, Swiss) — p. 38.
func TestFIDEExercise_20_TPR(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 20, "TPR", "performance-rating", state, scores, []fideO3Entry{
		{playerID: "2", value: 2120},
		{playerID: "3", value: 2089},
		{playerID: "4", value: 2081},
		{playerID: "1", value: 1969},
		{playerID: "16", value: 1969},
		{playerID: "6", value: 1813},

		{playerID: "8", value: 1730},
		{playerID: "5", value: 1690},

		{playerID: "15", value: 1788},
		{playerID: "12", value: 1250},
		{playerID: "13", value: 1781},
		{playerID: "7", value: 1611},
		{playerID: "9", value: 1175},
		{playerID: "10", value: 1640},
	})
}

// FIDE C.07-2023 exercise set, Exercise 21 (APRO, Swiss) — p. 38.
func TestFIDEExercise_21_APRO(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 21, "APRO", "avg-opponent-tpr", state, scores, []fideO3Entry{
		{playerID: "4", value: 1772},
		{playerID: "1", value: 1789},
	})
}

// FIDE C.07-2023 exercise set, Exercise 22 (PTP, Swiss) — pp. 39-40.
func TestFIDEExercise_22_PTP(t *testing.T) {
	state := fideO3SwissState()
	scores := fideO3Scores(t, state)

	fideO3AssertTiebreak(t, 22, "PTP", "performance-points", state, scores, []fideO3Entry{
		{playerID: "3", value: 2112},
	})
}
