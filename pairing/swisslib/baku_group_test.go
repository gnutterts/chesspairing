// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestBakuGroupA(t *testing.T) {
	players := func(count int) []chesspairing.PlayerEntry {
		result := make([]chesspairing.PlayerEntry, count)
		for i := range result {
			result[i] = chesspairing.PlayerEntry{ID: string(rune('a' + i)), PairingNumber: i + 1}
		}
		return result
	}
	unnumberedStarters := func() []chesspairing.PlayerEntry {
		result := make([]chesspairing.PlayerEntry, 8)
		for i := range result {
			result[i] = chesspairing.PlayerEntry{ID: string(rune('a' + i)), Rating: 2400 - i*100}
		}
		return result
	}
	tests := []struct {
		name    string
		players []chesspairing.PlayerEntry
		want    []string
	}{
		{"161 starters", players(161), playersIDs(82)},
		{"nine starters", players(9), playersIDs(6)},
		{"eight starters", players(8), playersIDs(4)},
		{"one starter", players(1), playersIDs(1)},
		{"no starters", nil, nil},
		{"late after group A", append(players(8), chesspairing.PlayerEntry{ID: "late", JoinedRound: 3}), playersIDs(4)},
		{
			name:    "unnumbered late entry above last group A player",
			players: append(unnumberedStarters(), chesspairing.PlayerEntry{ID: "late", Rating: 2500, JoinedRound: 3}),
			want:    []string{"late", "a", "b", "c", "d"},
		},
		{
			name:    "unnumbered late entry below last group A player",
			players: append(unnumberedStarters(), chesspairing.PlayerEntry{ID: "late", Rating: 1500, JoinedRound: 3}),
			want:    []string{"a", "b", "c", "d"},
		},
		{
			name: "late above last group A player",
			players: []chesspairing.PlayerEntry{
				{ID: "a", PairingNumber: 1},
				{ID: "b", PairingNumber: 2},
				{ID: "late", PairingNumber: 3, JoinedRound: 3},
				{ID: "c", PairingNumber: 4},
				{ID: "d", PairingNumber: 5},
				{ID: "e", PairingNumber: 6},
				{ID: "f", PairingNumber: 7},
				{ID: "g", PairingNumber: 8},
				{ID: "h", PairingNumber: 9},
			},
			want: []string{"a", "b", "late", "c", "d"},
		},
		{
			name: "late below last group A player",
			players: []chesspairing.PlayerEntry{
				{ID: "a", PairingNumber: 1},
				{ID: "b", PairingNumber: 2},
				{ID: "c", PairingNumber: 3},
				{ID: "d", PairingNumber: 4},
				{ID: "e", PairingNumber: 5},
				{ID: "f", PairingNumber: 6},
				{ID: "g", PairingNumber: 7},
				{ID: "late", PairingNumber: 8, JoinedRound: 3},
				{ID: "h", PairingNumber: 9},
			},
			want: []string{"a", "b", "c", "d"},
		},
		{
			name: "several late entries above last group A player",
			players: []chesspairing.PlayerEntry{
				{ID: "a", PairingNumber: 1},
				{ID: "b", PairingNumber: 2},
				{ID: "late1", PairingNumber: 3, JoinedRound: 2},
				{ID: "late2", PairingNumber: 4, JoinedRound: 3},
				{ID: "c", PairingNumber: 5},
				{ID: "d", PairingNumber: 6},
				{ID: "e", PairingNumber: 7},
				{ID: "f", PairingNumber: 8},
				{ID: "g", PairingNumber: 9},
				{ID: "h", PairingNumber: 10},
			},
			want: []string{"a", "b", "late1", "late2", "c", "d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BakuGroupA(tt.players)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("Group A has %d players, want %d", len(got), len(tt.want))
			}
			for _, id := range tt.want {
				if !got[id] {
					t.Errorf("Group A does not contain %q", id)
				}
			}
		})
	}
	if _, err := BakuGroupA([]chesspairing.PlayerEntry{{ID: "a", PairingNumber: 1}, {ID: "b", PairingNumber: 1}}); err == nil {
		t.Error("BakuGroupA() error = nil for duplicate pairing number")
	}
}

func playersIDs(count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = string(rune('a' + i))
	}
	return ids
}

func TestBakuPremise(t *testing.T) {
	valid := []chesspairing.ScoringConfig{
		{System: chesspairing.ScoringStandard},
		{System: chesspairing.ScoringStandard, Options: map[string]any{"pointWin": 2.0, "pointDraw": 1.0, "pointLoss": 0.0}},
		{},
	}
	for _, cfg := range valid {
		if err := BakuPremise(cfg); err != nil {
			t.Errorf("BakuPremise(%+v) error = %v", cfg, err)
		}
	}
	invalid := []chesspairing.ScoringConfig{
		{System: chesspairing.ScoringStandard, Options: map[string]any{"pointDraw": 0.4}},
		{System: chesspairing.ScoringStandard, Options: map[string]any{"pointLoss": 0.5}},
		{System: chesspairing.ScoringFootball},
		{System: chesspairing.ScoringKeizer},
	}
	for _, cfg := range invalid {
		if err := BakuPremise(cfg); err == nil {
			t.Errorf("BakuPremise(%+v) error = nil", cfg)
		}
	}
}
