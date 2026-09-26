// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"math"
	"sort"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// Cases pending: Exercise 4 BH #4=15; Exercise 6 BH-C1 #9=7.5; Exercise 7
// BH-C1 #4=11.5; Exercise 8 BH-C1 #14=9 — see FIDE C.07:2023 §§3.1-3.2;
// activated by C6.
//
// This file transcribes the FIDE C.07 (2023) Exercises in Tie-Breaking,
// Swiss individual tournament (crosstable §2.1), into regression tests for
// Buchholz exercises 1-8 (§§3.1-3.2).

var swissPlayers = []chesspairing.PlayerEntry{
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
	{ID: "12", DisplayName: "Nick", Rating: 1650},
	{ID: "13", DisplayName: "Opal", Rating: 1600},
	{ID: "14", DisplayName: "Paul", Rating: 1550},
	{ID: "15", DisplayName: "Reine", Rating: 1500},
	{ID: "16", DisplayName: "Stephan", Rating: 1450},
}

// swissState builds the FIDE C.07 (2023) Swiss individual crosstable (§2.1).
//
// Nick (#12) withdrew after round 3. The library's scoring/standard only
// scores players that are active in CurrentRound and, when a player is marked
// WithdrawnAfterRound, skips every game involving that player (which would
// drop David's round-1 win and Paul's round-3 forfeit). To keep the score
// oracle intact we therefore model the withdrawal as absences in rounds 4-5:
// Nick simply has no game and no bye in those rounds, which is exactly what
// the crosstable shows ("--").
func swissState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players:      swissPlayers,
		CurrentRound: 5,
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
	}
}

// swissExpectedScores returns the published final scores (crosstable §2.1).
func swissExpectedScores() map[string]float64 {
	return map[string]float64{
		"1": 3.5, "2": 4.0, "3": 3.5, "4": 3.5,
		"5": 2.5, "6": 3.0, "7": 1.5, "8": 2.5,
		"9": 1.5, "10": 1.0, "11": 2.5, "12": 2.0,
		"13": 1.5, "14": 2.0, "15": 2.0, "16": 3.5,
	}
}

// verifySwissScores is the single hard assertion: scoring/standard must
// reproduce the published final scores. It returns the computed scores for
// reuse by the tiebreaker comparison.
func verifySwissScores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	t.Helper()
	scorer := standard.New(standard.Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard: %v", err)
	}
	want := swissExpectedScores()
	got := make(map[string]float64, len(scores))
	for _, ps := range scores {
		got[ps.PlayerID] = ps.Score
	}
	for id, w := range want {
		g, ok := got[id]
		if !ok {
			t.Errorf("scoring/standard: no score for player #%s", id)
			continue
		}
		if g != w {
			t.Errorf("scoring/standard: player #%s score = %v, want %v", id, g, w)
		}
	}
	for id := range got {
		if _, ok := want[id]; !ok {
			t.Errorf("scoring/standard: unexpected score for player #%s", id)
		}
	}
	return scores
}

func swissLabel(id string) string {
	for _, p := range swissPlayers {
		if p.ID == id {
			return "#" + id + " " + p.DisplayName
		}
	}
	return "#" + id
}

func getC072023(id string) (chesspairing.TieBreaker, error) {
	switch id {
	case "buchholz":
		return &Buchholz{variant: buchholzFull, legacy: true}, nil
	case "buchholz-cut1":
		return &Buchholz{variant: buchholzCut1, legacy: true}, nil
	case "buchholz-cut2":
		return &Buchholz{variant: buchholzCut2, legacy: true}, nil
	case "buchholz-median":
		return &Buchholz{variant: buchholzMedian, legacy: true}, nil
	case "buchholz-median2":
		return &Buchholz{variant: buchholzMedian2, legacy: true}, nil
	case "avg-opponent-buchholz":
		return &AvgOpponentBuchholz{legacy: true}, nil
	case "fore-buchholz":
		return &ForeBuchholz{legacy: true}, nil
	case "sonneborn-berger":
		return &SonnebornBerger{legacy: true}, nil
	case "sonneborn-berger-cut1":
		return &SonnebornBerger{cut1: true, legacy: true}, nil
	default:
		return Get(id)
	}
}

// compareFIDETiebreak computes and asserts one FIDE tiebreaker value.
func compareFIDETiebreak(t *testing.T, exercise int, abbreviation, tbID string, want map[string]float64) {
	t.Helper()
	state := swissState()
	scores := verifySwissScores(t, state)

	tb, err := getC072023(tbID)
	if err != nil {
		t.Fatalf("FIDE exercise %d: tiebreaker %s: %v", exercise, tbID, err)
	}

	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("tiebreaker %s: %v", tbID, err)
	}
	got := make(map[string]float64, len(values))
	for _, v := range values {
		got[v.PlayerID] = v.Value
	}

	for _, id := range sortedIDs(want) {
		fide := want[id]
		code, ok := got[id]
		if !ok {
			t.Errorf("FIDE exercise %d %s: player %s is missing, want FIDE %v", exercise, abbreviation, swissLabel(id), fide)
			continue
		}
		if math.Abs(code-fide) > 0.001 {
			t.Errorf("FIDE exercise %d %s: player %s = %v, want FIDE %v", exercise, abbreviation, swissLabel(id), code, fide)
		}
	}
}

func sortedIDs(m map[string]float64) []string {
	ids := make([]string, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func TestFIDEExercise_1_Buchholz_Bruno(t *testing.T) {
	// §3.1 Exercise 1, p. 10: BH(#2) = 1.0+1.5+3.5+3.5+3.5 = 13.0.
	compareFIDETiebreak(t, 1, "BH", "buchholz", map[string]float64{"2": 13.0})
}

func TestFIDEExercise_2_Buchholz_Alyx_Charline(t *testing.T) {
	// §3.1 Exercise 2: BH(#1)=12.5 (p. 11), BH(#3)=15.5 (p. 12).
	compareFIDETiebreak(t, 2, "BH", "buchholz", map[string]float64{"1": 12.5, "3": 15.5})
}

func TestFIDEExercise_3_Buchholz_Helene_Irina_Maria(t *testing.T) {
	// §3.1 Exercise 3: BH(#5)=8.5 (p. 12), BH(#8)=13.5 and BH(#11)=13.5 (p. 13).
	compareFIDETiebreak(t, 3, "BH", "buchholz", map[string]float64{"5": 8.5, "8": 13.5, "11": 13.5})
}

func TestFIDEExercise_4_Buchholz_Score35(t *testing.T) {
	// §3.1 Exercise 4: BH(#1)=12.5, BH(#3)=15.5 (p. 14); BH(#16)=12.5 (p. 15).
	compareFIDETiebreak(t, 4, "BH", "buchholz", map[string]float64{"1": 12.5, "3": 15.5, "16": 12.5})
}

func TestFIDEExercise_5_BuchholzCut1_Score25(t *testing.T) {
	// §3.2 Exercise 5, p. 16: BH-C1(#5)=7.5, BH-C1(#8)=12.0, BH-C1(#11)=12.0.
	compareFIDETiebreak(t, 5, "BH-C1", "buchholz-cut1", map[string]float64{"5": 7.5, "8": 12.0, "11": 12.0})
}

func TestFIDEExercise_6_BuchholzCut1_Score15(t *testing.T) {
	// §3.2 Exercise 6: BH-C1(#7)=12.5 (p. 17); BH-C1(#13)=12.0 (p. 18).
	compareFIDETiebreak(t, 6, "BH-C1", "buchholz-cut1", map[string]float64{"7": 12.5, "13": 12.0})
}

func TestFIDEExercise_7_BuchholzCut1_Score35(t *testing.T) {
	// §3.2 Exercise 7: BH-C1(#1)=11.0, BH-C1(#3)=13.0 (p. 18); BH-C1(#16)=11.0 (p. 19).
	compareFIDETiebreak(t, 7, "BH-C1", "buchholz-cut1", map[string]float64{"1": 11.0, "3": 13.0, "16": 11.0})
}

func TestFIDEExercise_8_BuchholzCut1_Score20(t *testing.T) {
	// §3.2 Exercise 8: BH-C1(#12)=9.5 (p. 19); BH-C1(#15)=11.0 (p. 20).
	compareFIDETiebreak(t, 8, "BH-C1", "buchholz-cut1", map[string]float64{"12": 9.5, "15": 11.0})
}
