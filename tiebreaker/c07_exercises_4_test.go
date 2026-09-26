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

// FIDE C.07-2023 exercise set (Rev. 2403220900), block o4: exercises 26-33.
//
// Every executable FIDE value is transcribed from the exercise-set solution
// and asserted against the implementation.

const exerciseTol = 0.001

func exerciseAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= exerciseTol
}

// --- Shared players for the Swiss individual tournament (2.1) ---

var exerciseSwissOrder = []string{
	"1", "2", "3", "4", "5", "6", "7", "8",
	"9", "10", "11", "12", "13", "14", "15", "16",
}

var exerciseSwissNames = map[string]string{
	"1": "Alyx", "2": "Bruno", "3": "Charline", "4": "David",
	"5": "Helene", "6": "Franck", "7": "Genevieve", "8": "Irina",
	"9": "Jessica", "10": "Lais", "11": "Maria", "12": "Nick (W)",
	"13": "Opal", "14": "Paul", "15": "Reine", "16": "Stephan",
}

var exerciseSwissScores = map[string]float64{
	"1": 3.5, "2": 4.0, "3": 3.5, "4": 3.5,
	"5": 2.5, "6": 3.0, "7": 1.5, "8": 2.5,
	"9": 1.5, "10": 1.0, "11": 2.5, "12": 2.0,
	"13": 1.5, "14": 2.0, "15": 2.0, "16": 3.5,
}

// fideSwissState reproduces the Swiss tournament from section 2.1.
//
// CurrentRound is 3, rather than 5, because scoring/standard determines the
// active players from it. At CurrentRound=5, Nick (#12, withdrawn after round
// 3) would be excluded, as would his round-1 game against David. With
// CurrentRound=3, when Nick was last active, all sixteen players are scored.
// The B-type tie-breaks in this block do not consult CurrentRound.
func fideSwissState() *chesspairing.TournamentState {
	nickWithdrawnAfter := 3
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
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
			{ID: "12", DisplayName: "Nick (W)", Rating: 1650, WithdrawnAfterRound: &nickWithdrawnAfter},
			{ID: "13", DisplayName: "Opal", Rating: 1600},
			{ID: "14", DisplayName: "Paul", Rating: 1550},
			{ID: "15", DisplayName: "Reine", Rating: 1500},
			{ID: "16", DisplayName: "Stephan", Rating: 1450},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "10", BlackID: "2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "1", BlackID: "9", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "3", BlackID: "11", Result: chesspairing.ResultDraw},
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
					{PlayerID: "4", Type: chesspairing.ByeHalf},
					{PlayerID: "12", Type: chesspairing.ByePAB},
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
					{PlayerID: "6", Type: chesspairing.ByePAB},
					{PlayerID: "9", Type: chesspairing.ByeHalf},
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
					{PlayerID: "14", Type: chesspairing.ByeZero},
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
					{PlayerID: "9", Type: chesspairing.ByePAB},
				},
			},
		},
		CurrentRound: 3,
	}
}

// --- Shared players for the round-robin tournament (2.2) ---

var exerciseRRScores = map[string]float64{
	"1": 3.5, "2": 3.5, "3": 3.5,
	"4": 1.5, "5": 1.5, "6": 1.5,
}

// fideRoundRobinState reproduces the round-robin tournament from section 2.2.
// The games are derived from the conventional crosstable (first format), in
// which Franck is #6 and Helene is #5. Helene's round-4 forfeit win over
// Franck is recorded as ResultForfeitBlackWins.
func fideRoundRobinState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "1", DisplayName: "Alyx", Rating: 2200},
			{ID: "2", DisplayName: "Bruno", Rating: 2150},
			{ID: "3", DisplayName: "Charline", Rating: 2100},
			{ID: "4", DisplayName: "David", Rating: 2050},
			{ID: "5", DisplayName: "Helene", Rating: 2000},
			{ID: "6", DisplayName: "Franck", Rating: 1950},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "1", BlackID: "6", Result: chesspairing.ResultDraw},
					{WhiteID: "2", BlackID: "5", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "3", BlackID: "4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "1", BlackID: "2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "5", BlackID: "3", Result: chesspairing.ResultBlackWins},
					{WhiteID: "6", BlackID: "4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "3", BlackID: "1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "2", BlackID: "6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "4", BlackID: "5", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 4,
				Games: []chesspairing.GameData{
					{WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
					{WhiteID: "1", BlackID: "4", Result: chesspairing.ResultBlackWins},
					{WhiteID: "6", BlackID: "5", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true},
				},
			},
			{
				Number: 5,
				Games: []chesspairing.GameData{
					{WhiteID: "5", BlackID: "1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "4", BlackID: "2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "3", BlackID: "6", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound: 5,
	}
}

// checkExerciseScores compares scoring/standard with the FIDE final scores.
func checkExerciseScores(t *testing.T, state *chesspairing.TournamentState, want map[string]float64) []chesspairing.PlayerScore {
	t.Helper()

	scores, err := standard.New(standard.Options{}).Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard: %v", err)
	}

	got := make(map[string]float64, len(scores))
	for _, ps := range scores {
		got[ps.PlayerID] = ps.Score
	}

	for id, fide := range want {
		id, fide := id, fide
		t.Run("score/player-"+id, func(t *testing.T) {
			code, ok := got[id]
			if !ok {
				t.Errorf("score: player %s is missing from scoring/standard, want FIDE %v", id, fide)
				return
			}
			if !exerciseAlmostEqual(code, fide) {
				t.Errorf("score: player %s = %v, want FIDE %v", id, code, fide)
			}
		})
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("score transcription: unexpected player %s in scoring/standard", id)
		}
	}

	return scores
}

// assertExerciseTiebreak computes and asserts a FIDE tie-break value per player.
func assertExerciseTiebreak(t *testing.T, exercise int, abbreviation, tbID string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore, order []string, names map[string]string, fide map[string]float64) {
	t.Helper()

	tb, err := Get(tbID)
	if err != nil {
		t.Fatalf("FIDE exercise %d: tiebreaker %s: %v", exercise, tbID, err)
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("tiebreaker %s: %v", tbID, err)
	}

	code := make(map[string]float64, len(values))
	for _, v := range values {
		code[v.PlayerID] = v.Value
	}

	for _, pid := range order {
		cv, ok := code[pid]
		if !ok {
			t.Errorf("FIDE exercise %d %s: player %s is missing, want FIDE %v", exercise, abbreviation, names[pid], fide[pid])
			continue
		}
		if !exerciseAlmostEqual(cv, fide[pid]) {
			t.Errorf("FIDE exercise %d %s: player %s = %v, want FIDE %v", exercise, abbreviation, names[pid], cv, fide[pid])
		}
	}
}

// Cases pending: Exercise 26 DE, exercise 31 GE, and exercise 33 PS-C1 —
// see FIDE C.07:2023; activated by C6.

// --- Exercise 26: Direct Encounter in a round-robin (p. 44) ---

func TestFIDEExercise_26_RoundRobinScoreTranscription(t *testing.T) {
	state := fideRoundRobinState()
	checkExerciseScores(t, state, exerciseRRScores)
}

// --- Exercise 27: WIN (pp. 44-45) ---

var exercise27WIN = map[string]float64{
	"1": 2, "2": 3, "3": 2, "4": 2,
	"5": 2, "6": 3, "7": 1, "8": 2,
	"9": 1, "10": 1, "11": 2, "12": 2,
	"13": 1, "14": 2, "15": 2, "16": 3,
}

func TestFIDEExercise_27_SwissWin(t *testing.T) {
	state := fideSwissState()
	scores := checkExerciseScores(t, state, exerciseSwissScores)
	assertExerciseTiebreak(t, 27, "WIN", "win", state, scores, exerciseSwissOrder, exerciseSwissNames, exercise27WIN)
}

// --- Exercise 28: WON (pp. 45-46) ---

var exercise28WON = map[string]float64{
	"1": 2, "2": 3, "3": 2, "4": 2,
	"5": 2, "6": 2, "7": 1, "8": 2,
	"9": 0, "10": 1, "11": 1, "12": 0,
	"13": 1, "14": 2, "15": 2, "16": 3,
}

func TestFIDEExercise_28_SwissWon(t *testing.T) {
	state := fideSwissState()
	scores := checkExerciseScores(t, state, exerciseSwissScores)
	assertExerciseTiebreak(t, 28, "WON", "wins", state, scores, exerciseSwissOrder, exerciseSwissNames, exercise28WON)
}

// --- Exercise 29: BPG (pp. 46-47) ---

var exercise29BPG = map[string]float64{
	"1": 2, "2": 3, "3": 2, "4": 2,
	"5": 2, "6": 2, "7": 3, "8": 2,
	"9": 1, "10": 3, "11": 2, "12": 0,
	"13": 3, "14": 2, "15": 3, "16": 2,
}

func TestFIDEExercise_29_SwissBlackGames(t *testing.T) {
	state := fideSwissState()
	scores := checkExerciseScores(t, state, exerciseSwissScores)
	assertExerciseTiebreak(t, 29, "BPG", "black-games", state, scores, exerciseSwissOrder, exerciseSwissNames, exercise29BPG)
}

// --- Exercise 30: BWG (p. 47) ---

var exercise30BWG = map[string]float64{
	"1": 1, "2": 1, "3": 1, "4": 1,
	"5": 0, "6": 1, "7": 0, "8": 0,
	"9": 0, "10": 1, "11": 0, "12": 0,
	"13": 1, "14": 1, "15": 1, "16": 1,
}

func TestFIDEExercise_30_SwissBlackWins(t *testing.T) {
	state := fideSwissState()
	scores := checkExerciseScores(t, state, exerciseSwissScores)
	assertExerciseTiebreak(t, 30, "BWG", "black-wins", state, scores, exerciseSwissOrder, exerciseSwissNames, exercise30BWG)
}

// --- Exercise 32: PS (pp. 49-50) ---

var exercise32PS = map[string]float64{
	"1": 11.0, "2": 13.0, "3": 11.0, "4": 11.5,
	"5": 5.0, "6": 6.0, "7": 6.0, "8": 8.5,
	"9": 2.5, "10": 4.0, "11": 5.5, "12": 7.0,
	"13": 7.0, "14": 6.0, "15": 7.0, "16": 10.5,
}

func TestFIDEExercise_32_SwissProgressive(t *testing.T) {
	state := fideSwissState()
	scores := checkExerciseScores(t, state, exerciseSwissScores)
	assertExerciseTiebreak(t, 32, "PS", "progressive", state, scores, exerciseSwissOrder, exerciseSwissNames, exercise32PS)
}
