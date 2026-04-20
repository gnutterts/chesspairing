// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"context"
	"fmt"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// buildKeizerState creates a tournament state with n players and r rounds.
// Each round pairs sequential players: 1v2, 3v4, etc. All results are white wins.
func buildKeizerState(n, r int) *chesspairing.TournamentState {
	players := make([]chesspairing.PlayerEntry, n)
	for i := range players {
		players[i] = chesspairing.PlayerEntry{
			ID:     fmt.Sprintf("%d", i+1),
			Rating: 2000 - i*10,
		}
	}

	rounds := make([]chesspairing.RoundData, r)
	for rd := range rounds {
		var games []chesspairing.GameData
		for i := 0; i+1 < n; i += 2 {
			games = append(games, chesspairing.GameData{
				WhiteID: fmt.Sprintf("%d", i+1),
				BlackID: fmt.Sprintf("%d", i+2),
				Result:  chesspairing.ResultWhiteWins,
			})
		}
		rounds[rd] = chesspairing.RoundData{Games: games}
	}

	return &chesspairing.TournamentState{
		Players:      players,
		Rounds:       rounds,
		CurrentRound: r,
	}
}

func BenchmarkKeizerScore_20Players_9Rounds(b *testing.B) {
	state := buildKeizerState(20, 9)
	scorer := New(Options{})
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, _ = scorer.Score(ctx, state)
	}
}

func BenchmarkKeizerScore_50Players_9Rounds(b *testing.B) {
	state := buildKeizerState(50, 9)
	scorer := New(Options{})
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, _ = scorer.Score(ctx, state)
	}
}

func BenchmarkKeizerScore_100Players_11Rounds(b *testing.B) {
	state := buildKeizerState(100, 11)
	scorer := New(Options{})
	ctx := context.Background()
	b.ResetTimer()
	for b.Loop() {
		_, _ = scorer.Score(ctx, state)
	}
}
