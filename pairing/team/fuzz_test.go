// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func FuzzPair(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(1001))
	f.Fuzz(func(t *testing.T, seed uint64) {
		state := fuzzState(seed)
		result, err := New(Options{}).Pair(context.Background(), state)
		if err != nil {
			var pairingErr *chesspairing.PairingError
			if !errors.As(err, &pairingErr) {
				t.Fatalf("Pair() error = %v, want *PairingError", err)
			}
			return
		}
		if err := chesspairing.ValidatePairing(state, result); err != nil {
			t.Fatalf("ValidatePairing() error = %v", err)
		}
	})
}

func fuzzState(seed uint64) *chesspairing.TournamentState {
	players := make([]chesspairing.PlayerEntry, int(seed%4)*2+4)
	for i := range players {
		seed = seed*6364136223846793005 + 1
		players[i] = chesspairing.PlayerEntry{ID: fmt.Sprintf("t%d", i+1), Rating: 1200 + int(seed%1600), PairingNumber: i + 1}
	}
	roundCount := int(seed%uint64(len(players)-3)) + 2
	rounds := make([]chesspairing.RoundData, 0, roundCount)
	results := []chesspairing.GameResult{chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins, chesspairing.ResultDraw}
	for number := 1; number <= roundCount; number++ {
		games := make([]chesspairing.GameData, 0, len(players)/2)
		for i := range len(players) / 2 {
			white := (number - 1 + i) % (len(players) - 1)
			black := len(players) - 1
			if i != 0 {
				black = (number - 1 + len(players) - 1 - i) % (len(players) - 1)
			}
			games = append(games, chesspairing.GameData{WhiteID: players[white].ID, BlackID: players[black].ID, Result: results[(seed+uint64(number+i))%uint64(len(results))]})
		}
		rounds = append(rounds, chesspairing.RoundData{Number: number, Games: games})
	}
	return &chesspairing.TournamentState{
		Players:       players,
		CurrentRound:  roundCount + 1,
		Rounds:        rounds,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingTeam},
	}
}
