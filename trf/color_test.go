// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestToTournamentStateInitialColor(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		want  string
		valid bool
	}{
		{name: "white record", raw: "W", want: "white", valid: true},
		{name: "lowercase white record", raw: "w", want: "white", valid: true},
		{name: "white extension", raw: "white1", want: "white", valid: true},
		{name: "white name", raw: "white", want: "white", valid: true},
		{name: "black record", raw: "B", want: "black", valid: true},
		{name: "lowercase black record", raw: "b", want: "black", valid: true},
		{name: "black extension", raw: "black1", want: "black", valid: true},
		{name: "black name", raw: "black", want: "black", valid: true},
		{name: "invalid", raw: "green", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &Document{InitialColor: tt.raw}
			state, err := doc.ToTournamentState()
			if !tt.valid {
				if err == nil {
					t.Fatal("ToTournamentState returned nil error")
				}
				if !strings.Contains(err.Error(), tt.raw) {
					t.Errorf("error = %q, want raw value %q", err, tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ToTournamentState: %v", err)
			}
			if got := state.PairingConfig.Options["topSeedColor"]; got != tt.want {
				t.Errorf("topSeedColor = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFromTournamentStateInitialColor(t *testing.T) {
	tests := []struct {
		option string
		want   string
	}{
		{option: "white", want: "white1"},
		{option: "black", want: "black1"},
	}

	for _, tt := range tests {
		t.Run(tt.option, func(t *testing.T) {
			doc, _ := FromTournamentState(&chesspairing.TournamentState{PairingConfig: chesspairing.PairingConfig{Options: map[string]any{"topSeedColor": tt.option}}})
			if doc.InitialColor != tt.want {
				t.Errorf("InitialColor = %q, want %q", doc.InitialColor, tt.want)
			}
		})
	}
}
