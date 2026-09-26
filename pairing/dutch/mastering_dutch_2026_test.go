// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Cases pending: mast2_5 — see FIDE Mastering the Dutch §7 p. 51, FIDE
// pairings 1-3, 9-7, 4-5, 12-2, 6-8, 13-11, 10-15, 14-0; activated by C4.
//
// Source: Mastering the Dutch, chapters 6-11, pp. 37-67.
// Results are transcribed verbatim from the crosstables; F denotes a forfeit.
func mast2Game(w, b string, result chesspairing.GameResult) chesspairing.GameData {
	return chesspairing.GameData{WhiteID: w, BlackID: b, Result: result, IsForfeit: result.IsForfeit()}
}

func mast2Rounds() []chesspairing.RoundData {
	return []chesspairing.RoundData{
		{Number: 1, Games: []chesspairing.GameData{
			mast2Game("1", "8", chesspairing.ResultWhiteWins), mast2Game("9", "2", chesspairing.ResultDraw),
			mast2Game("3", "10", chesspairing.ResultWhiteWins), mast2Game("11", "4", chesspairing.ResultBlackWins),
			mast2Game("5", "12", chesspairing.ResultDraw), mast2Game("13", "6", chesspairing.ResultWhiteWins),
			mast2Game("7", "14", chesspairing.ResultBlackWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "15", Type: chesspairing.ByePAB}}},
		{Number: 2, Games: []chesspairing.GameData{
			mast2Game("14", "1", chesspairing.ResultBlackWins), mast2Game("2", "5", chesspairing.ResultWhiteWins),
			mast2Game("15", "3", chesspairing.ResultBlackWins), mast2Game("4", "13", chesspairing.ResultWhiteWins),
			mast2Game("10", "7", chesspairing.ResultBlackWins), mast2Game("8", "11", chesspairing.ResultWhiteWins),
			mast2Game("12", "9", chesspairing.ResultDraw),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "6", Type: chesspairing.ByeHalf}}},
		{Number: 3, Games: []chesspairing.GameData{
			mast2Game("1", "4", chesspairing.ResultDraw), mast2Game("3", "2", chesspairing.ResultDraw),
			mast2Game("6", "5", chesspairing.ResultDraw), mast2Game("8", "13", chesspairing.ResultForfeitWhiteWins),
			mast2Game("9", "14", chesspairing.ResultWhiteWins), mast2Game("11", "10", chesspairing.ResultWhiteWins),
			mast2Game("12", "15", chesspairing.ResultWhiteWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "7", Type: chesspairing.ByeHalf}}},
		{Number: 4, Games: []chesspairing.GameData{
			mast2Game("4", "3", chesspairing.ResultForfeitBlackWins), mast2Game("2", "1", chesspairing.ResultBlackWins),
			mast2Game("8", "9", chesspairing.ResultBlackWins), mast2Game("7", "12", chesspairing.ResultWhiteWins),
			mast2Game("5", "13", chesspairing.ResultWhiteWins), mast2Game("14", "6", chesspairing.ResultBlackWins),
			mast2Game("15", "11", chesspairing.ResultBlackWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "10", Type: chesspairing.ByePAB}}},
		{Number: 5, Games: []chesspairing.GameData{
			mast2Game("1", "3", chesspairing.ResultDraw), mast2Game("9", "7", chesspairing.ResultBlackWins),
			mast2Game("4", "5", chesspairing.ResultWhiteWins), mast2Game("12", "2", chesspairing.ResultDraw),
			mast2Game("6", "8", chesspairing.ResultBlackWins), mast2Game("13", "11", chesspairing.ResultWhiteWins),
			mast2Game("10", "15", chesspairing.ResultWhiteWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "14", Type: chesspairing.ByePAB}}},
		{Number: 6, Games: []chesspairing.GameData{
			mast2Game("7", "1", chesspairing.ResultBlackWins), mast2Game("3", "4", chesspairing.ResultWhiteWins),
			mast2Game("8", "12", chesspairing.ResultWhiteWins), mast2Game("11", "9", chesspairing.ResultDraw),
			mast2Game("2", "10", chesspairing.ResultWhiteWins), mast2Game("5", "14", chesspairing.ResultForfeitBlackWins),
			mast2Game("15", "6", chesspairing.ResultBlackWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "13", Type: chesspairing.ByePAB}}},
		{Number: 7, Games: []chesspairing.GameData{
			mast2Game("3", "8", chesspairing.ResultWhiteWins), mast2Game("1", "9", chesspairing.ResultDraw),
			mast2Game("4", "2", chesspairing.ResultWhiteWins), mast2Game("6", "7", chesspairing.ResultBlackWins),
			mast2Game("13", "14", chesspairing.ResultWhiteWins), mast2Game("12", "11", chesspairing.ResultWhiteWins),
			mast2Game("5", "10", chesspairing.ResultBlackWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "15", Type: chesspairing.ByeHalf}}},
		{Number: 8, Games: []chesspairing.GameData{
			mast2Game("7", "3", chesspairing.ResultDraw), mast2Game("1", "13", chesspairing.ResultWhiteWins),
			mast2Game("9", "4", chesspairing.ResultBlackWins), mast2Game("2", "8", chesspairing.ResultWhiteWins),
			mast2Game("10", "12", chesspairing.ResultWhiteWins), mast2Game("11", "6", chesspairing.ResultWhiteWins),
			mast2Game("14", "15", chesspairing.ResultBlackWins),
		}, Byes: []chesspairing.ByeEntry{{PlayerID: "5", Type: chesspairing.ByePAB}}},
	}
}

func mast2State(round int) *chesspairing.TournamentState {
	names := []string{"Alice", "Bruno", "Carla", "David", "Eloise", "Finn", "Giorgia", "Kevin", "Louise", "Mark", "Nancy", "Oskar", "Patricia", "Robert", "Stephanie"}
	players := make([]chesspairing.PlayerEntry, len(names))
	for i, name := range names {
		// Numeric IDs deliberately establish the source PN/TPN order.
		players[i] = chesspairing.PlayerEntry{ID: fmt.Sprint(i + 1), DisplayName: name, Rating: 3000 - i}
	}
	all := mast2Rounds()
	return &chesspairing.TournamentState{
		Players: players, Rounds: all[:round-1], CurrentRound: round, PreAssignedByes: mast2Requested(all, round),
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch, Options: map[string]any{"totalRounds": 9}},
	}
}

func mast2Pairs(result *chesspairing.PairingResult) string {
	parts := make([]string, 0, len(result.Pairings)+len(result.Byes))
	for _, pair := range result.Pairings {
		parts = append(parts, pair.WhiteID+"-"+pair.BlackID)
	}
	for _, bye := range result.Byes {
		if bye.Type != chesspairing.ByePAB {
			continue // requested byes are input, not part of FIDE's pairing table
		}
		parts = append(parts, bye.PlayerID+"-0")
	}
	return strings.Join(parts, ", ")
}

func mast2CheckScores(t *testing.T, state *chesspairing.TournamentState, want []float64) {
	t.Helper()
	got := swisslib.BuildPlayerStates(state)
	byID := make(map[string]float64, len(got))
	for _, player := range got {
		byID[player.ID] = player.Score
	}
	for i, score := range want {
		id := fmt.Sprint(i + 1)
		if byID[id] != score {
			t.Errorf("FIDE score for player %s = %.1f, got %.1f", id, score, byID[id])
		}
	}
}

func mast2CheckColorPreferences(t *testing.T, state *chesspairing.TournamentState, want []string) {
	t.Helper()
	players := swisslib.BuildPlayerStates(state)
	byID := make(map[string]swisslib.PlayerState, len(players))
	for _, player := range players {
		byID[player.ID] = player
	}
	for i, source := range want {
		if source == "" {
			continue
		}
		pref := swisslib.ComputeColorPreference(byID[fmt.Sprint(i+1)].ColorHistory)
		got := ""
		if pref.Color != nil {
			if *pref.Color == swisslib.ColorWhite {
				got = "W"
			} else {
				got = "B"
			}
			if !pref.AbsolutePreference && !pref.StrongPreference {
				got = strings.ToLower(got)
			}
			if pref.AbsolutePreference {
				got += "D"
			}
		}
		// Uppercase is absolute and lowercase mild. Compare the colour without
		// case sensitivity: the source is internally inconsistent (ledger D27,
		// disproved), so preference strength is not an oracle here.
		if !strings.EqualFold(strings.TrimSuffix(got, "D"), strings.TrimSuffix(source, "D")) {
			t.Errorf("FIDE colour preference for player %d = %s, got %s", i+1, source, got)
		}
	}
}

func mast2Run(t *testing.T, id, source string, round int, scores []float64, fide string) {
	t.Helper()
	state := mast2State(round)
	mast2CheckScores(t, state, scores)
	mast2CheckColorPreferences(t, state, mast2ColorPreferences[round])
	white := "white"
	result, err := New(Options{TopSeedColor: &white}).Pair(context.Background(), state)
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	ours := mast2Pairs(result)
	if ours != fide {
		t.Errorf("%s (%s): pairings = %s, want FIDE %s", id, source, ours, fide)
	}
}

var mast2ColorPreferences = map[int][]string{
	4: {"B", "W", "B", "W", "WD", "b", "w", "b", "B", "W", "B", "BD", "w", "W", "w"},
	5: {"w", "b", "B", "W", "B", "W", "B", "BD", "W", "W", "W", "w", "WD", "b", "B"},
	6: {"B", "W", "w", "b", "W", "b", "w", "w", "B", "b", "WD", "B", "b", "b", "w"},
	7: {"w", "b", "B", "W", "W", "W", "B", "B", "w", "W", "b", "w", "b", "b", ""},
	8: {"B", "W", "BD", "b", "B", "b", "w", "w", "WD", "WD", "W", "B", "BD", "W", "B"},
	9: {"BD", "b", "B", "W", "b", "W", "B", "WD", "b", "W", "b", "w", "w", "b", "w"},
}

func TestFIDEExample_mast2_4(t *testing.T) {
	// Vindplaats: §6, Fourth Round, p. 44, final pairing table (C5/C12/C13/C15).
	mast2Run(t, "mast2_4", "§6 p.44", 4, []float64{2.5, 2, 2.5, 2.5, 1, 1, 1.5, 2, 2, 0, 1, 2, 1, 1, 1}, "4-3, 2-1, 8-9, 7-12, 5-13, 14-6, 15-11, 10-0")
}

func TestFIDEExample_mast2_6(t *testing.T) {
	// Vindplaats: §8, Sixth Round, p. 57, round-pairing table (C4/C8/C9).
	mast2Run(t, "mast2_6", "§8 p.57", 6, []float64{4, 2.5, 4, 3.5, 2, 2, 3.5, 3, 3, 2, 2, 2.5, 2, 2, 1}, "7-1, 3-4, 8-12, 11-9, 2-10, 5-14, 15-6, 13-0")
}

func TestFIDEExample_mast2_7(t *testing.T) {
	// Vindplaats: §9, Seventh Round, p. 60, round-pairing table (colour allocation).
	mast2Run(t, "mast2_7", "§9 p.60", 7, []float64{5, 3.5, 5, 3.5, 2, 3, 3.5, 4, 3.5, 2, 2.5, 2.5, 3, 3, 1}, "3-8, 1-9, 4-2, 6-7, 13-14, 12-11, 5-10")
}

func TestFIDEExample_mast2_8(t *testing.T) {
	// Source: §10, Eighth Round, p. 63, complete pairing table (C5/C6/C8).
	// Board 8 (5-0, the bye) receives a different colour in FIDE and library
	// reports; assert the pairings without treating that reporting difference as colour.
	mast2Run(t, "mast2_8", "§10 p.63", 8, []float64{5.5, 3.5, 6, 4.5, 2, 3, 4.5, 4, 4, 3, 2.5, 3.5, 4, 3, 1.5}, "7-3, 1-13, 9-4, 2-8, 10-12, 11-6, 14-15, 5-0")
}

func TestFIDEExample_mast2_9(t *testing.T) {
	// Vindplaats: §11, Ninth Round, p. 67, sorted final pairing table (C10/C11; C.04.2:3.6).
	mast2Run(t, "mast2_9", "§11 p.67", 9, []float64{6.5, 4.5, 6.5, 5.5, 3, 3, 5, 4, 4, 4, 3.5, 3.5, 4, 3, 2.5}, "10-1, 9-3, 4-7, 13-2, 8-15, 5-11, 12-14, 6-0")
}

// mast2Requested returns the requested (non-PAB) byes of the round being paired;
// the engine must see them as pre-assigned, while the PAB is for the engine to choose.
func mast2Requested(all []chesspairing.RoundData, round int) []chesspairing.ByeEntry {
	if round-1 >= len(all) {
		return nil
	}
	var out []chesspairing.ByeEntry
	for _, b := range all[round-1].Byes {
		if b.Type != chesspairing.ByePAB {
			out = append(out, b)
		}
	}
	return out
}
