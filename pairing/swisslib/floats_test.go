// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import "testing"

func TestLastFloat(t *testing.T) {
	tests := []struct {
		name    string
		history []Float
		want    Float
	}{
		{"empty", nil, FloatNone},
		{"none", []Float{FloatNone}, FloatNone},
		{"up", []Float{FloatNone, FloatUp}, FloatUp},
		{"down then none", []Float{FloatDown, FloatNone}, FloatNone},
		{"down", []Float{FloatDown}, FloatDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastFloat(tt.history)
			if got != tt.want {
				t.Errorf("LastFloat(%v) = %v, want %v", tt.history, got, tt.want)
			}
		})
	}
}

func TestConsecutiveSameFloat(t *testing.T) {
	tests := []struct {
		name    string
		history []Float
		dir     Float
		want    int
	}{
		{"empty", nil, FloatDown, 0},
		{"one down", []Float{FloatDown}, FloatDown, 1},
		{"two down", []Float{FloatDown, FloatDown}, FloatDown, 2},
		{"down then none", []Float{FloatDown, FloatNone}, FloatDown, 0},
		{"none then down", []Float{FloatNone, FloatDown}, FloatDown, 1},
		{"up down up", []Float{FloatUp, FloatDown, FloatUp}, FloatUp, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConsecutiveSameFloat(tt.history, tt.dir)
			if got != tt.want {
				t.Errorf("ConsecutiveSameFloat(%v, %v) = %d, want %d", tt.history, tt.dir, got, tt.want)
			}
		})
	}
}

func TestFloatedToSameScoreGroup(t *testing.T) {
	tests := []struct {
		name        string
		history     []Float
		targetScore float64
		scores      []float64 // score after each round
		want        bool
	}{
		{
			"no float history",
			nil, 1.0, nil, false,
		},
		{
			"floated down to 1.0 last round, targeting 1.0 again",
			[]Float{FloatDown},
			1.0,
			[]float64{1.0},
			true,
		},
		{
			"no float last round",
			[]Float{FloatNone},
			1.0,
			[]float64{1.0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FloatedToSameScoreGroup(tt.history, tt.targetScore, tt.scores)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
