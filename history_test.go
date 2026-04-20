// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing_test

import (
	"reflect"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func mkGame(t *testing.T, white, black, result string) cp.GameData {
	t.Helper()
	r, err := cp.ParseGameResult(result)
	if err != nil {
		t.Fatalf("parse %q: %v", result, err)
	}
	return cp.GameData{
		WhiteID:   white,
		BlackID:   black,
		Result:    r,
		IsForfeit: r == cp.ResultForfeitWhiteWins || r == cp.ResultForfeitBlackWins || r == cp.ResultDoubleForfeit,
	}
}

func TestPlayedPairs_Empty(t *testing.T) {
	got := cp.PlayedPairs(&cp.TournamentState{}, cp.HistoryOptions{})
	if got != nil {
		t.Errorf("empty state: got %v, want nil", got)
	}
}

func TestPlayedPairs_NilState(t *testing.T) {
	got := cp.PlayedPairs(nil, cp.HistoryOptions{})
	if got != nil {
		t.Errorf("nil state: got %v, want nil", got)
	}
}

func TestPlayedPairs_SingleDecisive(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "a", "b", "1-0")}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_LexicographicWithinPair(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "z", "a", "1-0")}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "z"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_OuterSorted(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				mkGame(t, "c", "d", "1-0"),
				mkGame(t, "a", "b", "0-1"),
			}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "b"}, {"c", "d"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_Dedup(t *testing.T) {
	// Same pair playing twice (could happen in round-robin double, or
	// in tests using malformed state). PlayedPairs returns the pair once.
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "a", "b", "1-0")}},
			{Number: 2, Games: []cp.GameData{mkGame(t, "b", "a", "0-1")}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_PendingSkipped(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				mkGame(t, "a", "b", "1-0"),
				mkGame(t, "c", "d", "*"),
			}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_DoubleForfeitNeverIncluded(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "a", "b", "0-0f")}},
		},
	}
	for _, include := range []bool{false, true} {
		got := cp.PlayedPairs(state, cp.HistoryOptions{IncludeForfeits: include})
		if got != nil {
			t.Errorf("IncludeForfeits=%v: double-forfeit should never be included, got %v", include, got)
		}
	}
}

func TestPlayedPairs_SingleForfeit_DefaultExcluded(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "a", "b", "1-0f")}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	if got != nil {
		t.Errorf("default IncludeForfeits=false: got %v, want nil", got)
	}
}

func TestPlayedPairs_SingleForfeit_OptIn(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{mkGame(t, "a", "b", "1-0f")}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{IncludeForfeits: true})
	want := [][]string{{"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("IncludeForfeits=true: got %v, want %v", got, want)
	}
}

func TestPlayedPairs_BothForfeitDirections(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				mkGame(t, "a", "b", "1-0f"),
				mkGame(t, "c", "d", "0-1f"),
			}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{IncludeForfeits: true})
	want := [][]string{{"a", "b"}, {"c", "d"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_ByesSkipped(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{
				Number: 1,
				Games:  []cp.GameData{mkGame(t, "a", "b", "1-0")},
				Byes:   []cp.ByeEntry{{PlayerID: "c", Type: cp.ByePAB}},
			},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"a", "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestPlayedPairs_AllByeTypesSkipped enumerates every ByeType through
// the enum sentinel and confirms that no bye type produces a played
// pair entry. The contract is "byes never create a pair", regardless
// of which type is recorded.
func TestPlayedPairs_AllByeTypesSkipped(t *testing.T) {
	for bt := cp.ByePAB; bt.IsValid(); bt++ {
		t.Run(bt.String(), func(t *testing.T) {
			state := &cp.TournamentState{
				Rounds: []cp.RoundData{{
					Number: 1,
					Byes:   []cp.ByeEntry{{PlayerID: "p1", Type: bt}},
				}},
			}
			got := cp.PlayedPairs(state, cp.HistoryOptions{})
			if len(got) != 0 {
				t.Errorf("bye type %s produced pairs %v, want none", bt, got)
			}
		})
	}
}

func TestPlayedPairs_EmptyPlayerIDsSkipped(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				{WhiteID: "", BlackID: "b", Result: cp.ResultWhiteWins},
				{WhiteID: "a", BlackID: "", Result: cp.ResultBlackWins},
				mkGame(t, "c", "d", "1-0"),
			}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{{"c", "d"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPlayedPairs_MultiRound(t *testing.T) {
	state := &cp.TournamentState{
		Rounds: []cp.RoundData{
			{Number: 1, Games: []cp.GameData{
				mkGame(t, "a", "b", "1-0"),
				mkGame(t, "c", "d", "0.5-0.5"),
			}},
			{Number: 2, Games: []cp.GameData{
				mkGame(t, "a", "c", "0-1"),
				mkGame(t, "b", "d", "1-0"),
			}},
		},
	}
	got := cp.PlayedPairs(state, cp.HistoryOptions{})
	want := [][]string{
		{"a", "b"},
		{"a", "c"},
		{"b", "d"},
		{"c", "d"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
