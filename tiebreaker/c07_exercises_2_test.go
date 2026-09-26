// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

// FIDE C.07 exercise set (Rev. 2403220900 / C.07-2023), block o2.
//
// Exercises in this file:
//   9  AOB    (p. 20-22)
//   10 FB     (p. 22-23)
//   11 SB     players on 3.5 points (p. 25-26)
//   12 SB     all players (p. 26-27)
//   13 SB-C1  all players (p. 27-29)
//   14 SB     round-robin (p. 29-30)
//   15 SB-C1  round-robin (p. 30)
//   16 KS     round-robin (p. 30-31)
//
// Every executable FIDE value is asserted against the implementation.
//
// Nick (#12) withdrew after round 3. The standard scorer excludes players
// with WithdrawnAfterRound from the score list, while the FIDE solution still
// lists Nick's score (2.0). Nick is therefore modelled as active and absent
// in rounds 4-5. This reproduces his score and own Buchholz/FB exactly; his
// opponents use Nick's raw score (2.0), rather than FIDE's adjusted 3.0.

import (
	"context"
	"math"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

var (
	fideO2SwissOrder = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16"}
	fideO2RROrder    = []string{"1", "2", "3", "4", "5", "6"}
	fideO2SB35Order  = []string{"1", "3", "4", "16"}
)

// --- FIDE values transcribed from the solution ---

// Swiss final scores (p. 5-6 / §2.1).
var fideO2SwissScores = map[string]float64{
	"1": 3.5, "2": 4.0, "3": 3.5, "4": 3.5,
	"5": 2.5, "6": 3.0, "7": 1.5, "8": 2.5,
	"9": 1.5, "10": 1.0, "11": 2.5, "12": 2.0,
	"13": 1.5, "14": 2.0, "15": 2.0, "16": 3.5,
}

// Swiss adjusted scores for Fore Buchholz (p. 22-23, exercise 10).
var fideO2SwissForeScores = map[string]float64{
	"1": 3.5, "2": 4.0, "3": 3.5, "4": 3.5,
	"5": 2.0, "6": 2.5, "7": 2.0, "8": 3.0,
	"9": 1.5, "10": 1.5, "11": 2.0, "12": 2.0,
	"13": 2.0, "14": 1.5, "15": 2.5, "16": 3.0,
}

// Round-robin final scores (p. 6 / §2.2, "Swiss" format).
var fideO2RRScores = map[string]float64{
	"1": 3.5, "2": 3.5, "3": 3.5, "4": 1.5, "5": 1.5, "6": 1.5,
}

// Exercise 9: Buchholz (p. 21).
var fideO2BH = map[string]float64{
	"1": 12.5, "2": 13.0, "3": 15.5, "4": 15.0,
	"5": 8.5, "6": 12.0, "7": 14.5, "8": 13.5,
	"9": 9.0, "10": 13.0, "11": 13.5, "12": 11.5,
	"13": 14.0, "14": 11.0, "15": 12.0, "16": 12.5,
}

// Exercise 9: Average Buchholz of Opponents (p. 22).
var fideO2AOB = map[string]float64{
	"1": 12.60, "2": 13.60, "3": 13.40, "4": 13.38,
	"5": 13.40, "6": 13.25, "7": 11.90, "8": 13.00,
	"9": 12.75, "10": 10.90, "11": 12.75, "12": 15.00,
	"13": 12.10, "14": 13.17, "15": 12.20, "16": 13.30,
}

// Exercise 10: Fore Buchholz (p. 23).
var fideO2FB = map[string]float64{
	"1": 13.5, "2": 13.5, "3": 15.0, "4": 15.5,
	"5": 10.0, "6": 12.0, "7": 13.5, "8": 12.5,
	"9": 9.5, "10": 12.5, "11": 12.5, "12": 11.5,
	"13": 13.5, "14": 10.5, "15": 12.0, "16": 13.5,
}

// Exercise 11: Sonneborn-Berger for players on 3.5 points (p. 25-26).
var fideO2SB35 = map[string]float64{
	"1": 8.00, "3": 10.50, "4": 9.75, "16": 7.25,
}

// Exercise 12: Sonneborn-Berger, all players (p. 26-27).
var fideO2SB = map[string]float64{
	"1": 8.00, "2": 9.50, "3": 10.50, "4": 9.75,
	"5": 4.25, "6": 6.50, "7": 3.25, "8": 5.25,
	"9": 2.25, "10": 1.50, "11": 5.75, "12": 4.00,
	"13": 4.25, "14": 4.50, "15": 3.50, "16": 7.25,
}

var fideO2SBC1 = map[string]float64{
	"1": 7.25, "2": 8.50, "3": 9.25, "4": 8.00,
	"5": 3.25, "6": 5.50, "7": 1.25, "8": 3.75,
	"9": 2.25, "10": 0.00, "11": 4.25, "12": 4.00,
	"13": 4.25, "14": 3.00, "15": 2.50, "16": 5.75,
}

// Exercise 14: Sonneborn-Berger, round-robin (p. 29-30).
var fideO2RRSB = map[string]float64{
	"1": 9.25, "2": 6.25, "3": 6.25, "4": 4.25, "5": 3.25, "6": 1.50,
}

var fideO2RRSBC1 = map[string]float64{
	"1": 9.25, "2": 4.75, "3": 4.75, "4": 4.25, "5": 3.25, "6": 0.75,
}

// Exercise 16: Koya System, round-robin (p. 31).
var fideO2RRKS = map[string]float64{
	"1": 2.0, "2": 0.5, "3": 0.5, "4": 1.0, "5": 0.5, "6": 0.0,
}

// --- TournamentState builders ---

func fideO2SwissPlayers() []chesspairing.PlayerEntry {
	return []chesspairing.PlayerEntry{
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
		{ID: "12", DisplayName: "Nick (W)", Rating: 1650},
		{ID: "13", DisplayName: "Opal", Rating: 1600},
		{ID: "14", DisplayName: "Paul", Rating: 1550},
		{ID: "15", DisplayName: "Reine", Rating: 1500},
		{ID: "16", DisplayName: "Stephan", Rating: 1450},
	}
}

func fideO2SwissRounds14() []chesspairing.RoundData {
	return []chesspairing.RoundData{
		{
			Number: 1,
			Games: []chesspairing.GameData{
				{WhiteID: "1", BlackID: "9", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "10", BlackID: "2", Result: chesspairing.ResultBlackWins},
				{WhiteID: "3", BlackID: "11", Result: chesspairing.ResultDraw},
				{WhiteID: "12", BlackID: "4", Result: chesspairing.ResultBlackWins},
				{WhiteID: "5", BlackID: "13", Result: chesspairing.ResultBlackWins},
				{WhiteID: "14", BlackID: "6", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "7", BlackID: "15", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "16", BlackID: "8", Result: chesspairing.ResultDraw},
			},
		},
		{
			Number: 2,
			Games: []chesspairing.GameData{
				{WhiteID: "13", BlackID: "1", Result: chesspairing.ResultDraw},
				{WhiteID: "2", BlackID: "7", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "6", BlackID: "3", Result: chesspairing.ResultBlackWins},
				{WhiteID: "15", BlackID: "5", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "8", BlackID: "14", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "9", BlackID: "10", Result: chesspairing.ResultBlackWins},
				{WhiteID: "11", BlackID: "16", Result: chesspairing.ResultBlackWins},
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
				{WhiteID: "5", BlackID: "11", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "16", BlackID: "7", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "10", BlackID: "15", Result: chesspairing.ResultBlackWins},
				{WhiteID: "12", BlackID: "14", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "6", Type: chesspairing.ByePAB},
				{PlayerID: "9", Type: chesspairing.ByeHalf},
			},
		},
		{
			Number: 4,
			Games: []chesspairing.GameData{
				{WhiteID: "15", BlackID: "1", Result: chesspairing.ResultBlackWins},
				{WhiteID: "2", BlackID: "16", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "4", BlackID: "3", Result: chesspairing.ResultDraw},
				{WhiteID: "7", BlackID: "5", Result: chesspairing.ResultDraw},
				{WhiteID: "6", BlackID: "10", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "8", BlackID: "13", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "11", BlackID: "9", Result: chesspairing.ResultForfeitWhiteWins, IsForfeit: true},
			},
			Byes: []chesspairing.ByeEntry{
				{PlayerID: "14", Type: chesspairing.ByeZero},
			},
		},
	}
}

func fideO2SwissState() *chesspairing.TournamentState {
	rounds := fideO2SwissRounds14()
	rounds = append(rounds, chesspairing.RoundData{
		Number: 5,
		Games: []chesspairing.GameData{
			{WhiteID: "1", BlackID: "4", Result: chesspairing.ResultDraw},
			{WhiteID: "3", BlackID: "2", Result: chesspairing.ResultDraw},
			{WhiteID: "5", BlackID: "10", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "8", BlackID: "6", Result: chesspairing.ResultBlackWins},
			{WhiteID: "11", BlackID: "7", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "13", BlackID: "14", Result: chesspairing.ResultBlackWins},
			{WhiteID: "16", BlackID: "15", Result: chesspairing.ResultWhiteWins},
		},
		Byes: []chesspairing.ByeEntry{
			{PlayerID: "9", Type: chesspairing.ByePAB},
		},
	})
	return &chesspairing.TournamentState{
		Players:      fideO2SwissPlayers(),
		Rounds:       rounds,
		CurrentRound: 5,
	}
}

func fideO2SwissForeState() *chesspairing.TournamentState {
	rounds := fideO2SwissRounds14()
	rounds = append(rounds, chesspairing.RoundData{
		Number: 5,
		Games: []chesspairing.GameData{
			{WhiteID: "1", BlackID: "4", Result: chesspairing.ResultDraw},
			{WhiteID: "3", BlackID: "2", Result: chesspairing.ResultDraw},
			{WhiteID: "5", BlackID: "10", Result: chesspairing.ResultDraw},
			{WhiteID: "8", BlackID: "6", Result: chesspairing.ResultDraw},
			{WhiteID: "11", BlackID: "7", Result: chesspairing.ResultDraw},
			{WhiteID: "13", BlackID: "14", Result: chesspairing.ResultDraw},
			{WhiteID: "16", BlackID: "15", Result: chesspairing.ResultDraw},
		},
		Byes: []chesspairing.ByeEntry{
			{PlayerID: "9", Type: chesspairing.ByePAB},
		},
	})
	return &chesspairing.TournamentState{
		Players:      fideO2SwissPlayers(),
		Rounds:       rounds,
		CurrentRound: 5,
	}
}

func fideO2RoundRobinState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "1", DisplayName: "Alyx", Rating: 2200},
			{ID: "2", DisplayName: "Bruno", Rating: 2150},
			{ID: "3", DisplayName: "Charline", Rating: 2100},
			{ID: "4", DisplayName: "David", Rating: 2050},
			{ID: "5", DisplayName: "Franck", Rating: 1950},
			{ID: "6", DisplayName: "Helene", Rating: 2000},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "1", BlackID: "5", Result: chesspairing.ResultDraw},
					{WhiteID: "2", BlackID: "6", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "3", BlackID: "4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "1", BlackID: "2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "6", BlackID: "3", Result: chesspairing.ResultBlackWins},
					{WhiteID: "5", BlackID: "4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 3,
				Games: []chesspairing.GameData{
					{WhiteID: "3", BlackID: "1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "2", BlackID: "5", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "4", BlackID: "6", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 4,
				Games: []chesspairing.GameData{
					{WhiteID: "1", BlackID: "4", Result: chesspairing.ResultBlackWins},
					{WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
					{WhiteID: "5", BlackID: "6", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true},
				},
			},
			{
				Number: 5,
				Games: []chesspairing.GameData{
					{WhiteID: "6", BlackID: "1", Result: chesspairing.ResultBlackWins},
					{WhiteID: "4", BlackID: "2", Result: chesspairing.ResultBlackWins},
					{WhiteID: "3", BlackID: "5", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
		CurrentRound:  5,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingRoundRobin},
	}
}

// --- Helpers ---

func fideO2ComputeScores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	t.Helper()
	scorer := standard.New(standard.Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard: %v", err)
	}
	return scores
}

func fideO2AssertScores(t *testing.T, exercise int, scores []chesspairing.PlayerScore, want map[string]float64) {
	t.Helper()
	got := make(map[string]float64, len(scores))
	for _, ps := range scores {
		got[ps.PlayerID] = ps.Score
	}
	for id, w := range want {
		g, ok := got[id]
		if !ok {
			t.Errorf("exercise %d: score for %s is missing from scoring/standard output", exercise, id)
			continue
		}
		if math.Abs(g-w) > 0.001 {
			t.Errorf("exercise %d: score for %s = %v, want FIDE %v", exercise, id, g, w)
		}
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("exercise %d: unexpected player %s in scoring/standard output", exercise, id)
		}
	}
}

func fideO2AssertTiebreak(t *testing.T, exercise int, abbreviation, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore, want map[string]float64, order []string) {
	t.Helper()

	names := make(map[string]string, len(state.Players))
	for _, p := range state.Players {
		names[p.ID] = p.DisplayName
	}

	tb, err := getC072023(id)
	if err != nil {
		t.Fatalf("FIDE exercise %d: tiebreaker %s: %v", exercise, id, err)
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("exercise %d: tiebreak %s: %v", exercise, abbreviation, err)
	}
	vm := valueMap(values)

	for _, p := range order {
		w, ok := want[p]
		if !ok {
			continue
		}
		code, ok := vm[p]
		if !ok {
			t.Errorf("FIDE exercise %d %s: player %s is missing, want FIDE %v", exercise, abbreviation, names[p], w)
			continue
		}
		if math.Abs(code-w) > 0.001 {
			t.Errorf("FIDE exercise %d %s: player %s = %v, want FIDE %v", exercise, abbreviation, names[p], code, w)
		}
	}
}

// --- Exercises ---

func TestFIDEExercise_09_AOB(t *testing.T) {
	state := fideO2SwissState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 9, scores, fideO2SwissScores)

	fideO2AssertTiebreak(t, 9, "BH", "buchholz", state, scores, fideO2BH, fideO2SwissOrder)
	fideO2AssertTiebreak(t, 9, "AOB", "avg-opponent-buchholz", state, scores, fideO2AOB, fideO2SwissOrder)
}

func TestFIDEExercise_10_FB(t *testing.T) {
	state := fideO2SwissForeState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 10, scores, fideO2SwissForeScores)

	fideO2AssertTiebreak(t, 10, "FB", "fore-buchholz", state, scores, fideO2FB, fideO2SwissOrder)
}

func TestFIDEExercise_11_SB(t *testing.T) {
	state := fideO2SwissState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 11, scores, fideO2SwissScores)

	fideO2AssertTiebreak(t, 11, "SB", "sonneborn-berger", state, scores, fideO2SB35, fideO2SB35Order)
}

func TestFIDEExercise_12_SB(t *testing.T) {
	state := fideO2SwissState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 12, scores, fideO2SwissScores)

	fideO2AssertTiebreak(t, 12, "SB", "sonneborn-berger", state, scores, fideO2SB, fideO2SwissOrder)
}

func TestFIDEExercise_13_SBC1(t *testing.T) {
	state := fideO2SwissState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertTiebreak(t, 13, "SB-C1", "sonneborn-berger-cut1", state, scores, fideO2SBC1, fideO2SwissOrder)
}

func TestFIDEExercise_14_SB_RR(t *testing.T) {
	state := fideO2RoundRobinState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 14, scores, fideO2RRScores)

	fideO2AssertTiebreak(t, 14, "SB", "sonneborn-berger", state, scores, fideO2RRSB, fideO2RROrder)
}

func TestFIDEExercise_15_SBC1_RR(t *testing.T) {
	state := fideO2RoundRobinState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertTiebreak(t, 15, "SB-C1", "sonneborn-berger-cut1", state, scores, fideO2RRSBC1, fideO2RROrder)
}

func TestFIDEExercise_16_KS_RR(t *testing.T) {
	state := fideO2RoundRobinState()
	scores := fideO2ComputeScores(t, state)
	fideO2AssertScores(t, 16, scores, fideO2RRScores)

	fideO2AssertTiebreak(t, 16, "KS", "koya", state, scores, fideO2RRKS, fideO2RROrder)
}
