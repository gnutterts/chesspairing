package swisslib

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestValidatePairing_Valid(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
			{Board: 2, WhiteID: "p3", BlackID: "p4"},
		},
	}
	if err := ValidatePairing(players, result); err != nil {
		t.Errorf("valid pairing should not error: %v", err)
	}
}

func TestValidatePairing_DuplicatePlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
			{Board: 2, WhiteID: "p1", BlackID: "p3"}, // p1 paired twice!
		},
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("duplicate player should fail validation")
	}
}

func TestValidatePairing_UnknownPlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p99"}, // p99 not in player list
		},
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("unknown player should fail validation")
	}
}

func TestValidatePairing_OddWithBye(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
		},
		Byes: []string{"p3"},
	}
	if err := ValidatePairing(players, result); err != nil {
		t.Errorf("odd players with bye should be valid: %v", err)
	}
}

func TestValidatePairing_MissingPlayer(t *testing.T) {
	players := []PlayerState{
		{ID: "p1", Active: true},
		{ID: "p2", Active: true},
		{ID: "p3", Active: true},
		{ID: "p4", Active: true},
	}
	result := &chesspairing.PairingResult{
		Pairings: []chesspairing.GamePairing{
			{Board: 1, WhiteID: "p1", BlackID: "p2"},
		},
		// p3 and p4 are neither paired nor have bye
	}
	if err := ValidatePairing(players, result); err == nil {
		t.Error("missing players should fail validation")
	}
}
