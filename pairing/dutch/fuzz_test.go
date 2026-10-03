// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package dutch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func FuzzPair(f *testing.F) {
	f.Add(uint64(1))
	f.Add(uint64(2))
	f.Add(uint64(1001))
	f.Fuzz(func(t *testing.T, seed uint64) {
		state := fuzzState(t, seed)
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

func fuzzState(t *testing.T, seed uint64) *chesspairing.TournamentState {
	t.Helper()
	files := []string{
		"testdata/golden/10-players-7-rounds/scenario.json",
		"testdata/golden/withdrawal/scenario.json",
	}
	scenarioFile := files[seed%uint64(len(files))]
	data, err := os.ReadFile(scenarioFile) //nolint:gosec // test fixture
	if err != nil {
		t.Fatal(err)
	}
	var scenario goldenScenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		t.Fatal(err)
	}
	roundToPair := int(seed%uint64(scenario.TotalRounds-1)) + 2
	players := append([]chesspairing.PlayerEntry(nil), scenario.Players...)
	for i := range players {
		seed = seed*6364136223846793005 + 1
		players[i].Rating += int(seed % 101)
	}
	if seed%2 == 0 {
		withdrawn := 1
		players[len(players)-1].WithdrawnAfterRound = &withdrawn
	}
	rounds := make([]chesspairing.RoundData, 0, roundToPair-1)
	for number := 1; number < roundToPair; number++ {
		data, err = os.ReadFile(filepath.Join(filepath.Dir(scenarioFile), fmt.Sprintf("round-%d.json", number))) //nolint:gosec // test fixture
		if err != nil {
			t.Fatal(err)
		}
		var pairing chesspairing.PairingResult
		if err := json.Unmarshal(data, &pairing); err != nil {
			t.Fatal(err)
		}
		rounds = append(rounds, chesspairing.RoundData{Number: number, Games: fuzzGames(pairing.Pairings, seed+uint64(number)), Byes: pairing.Byes})
	}
	return &chesspairing.TournamentState{
		Players:       players,
		CurrentRound:  roundToPair,
		Rounds:        rounds,
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
	}
}

func fuzzGames(pairings []chesspairing.GamePairing, seed uint64) []chesspairing.GameData {
	games := make([]chesspairing.GameData, len(pairings))
	results := []chesspairing.GameResult{chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins, chesspairing.ResultDraw}
	for i, pairing := range pairings {
		games[i] = chesspairing.GameData{WhiteID: pairing.WhiteID, BlackID: pairing.BlackID, Result: results[(seed+uint64(i))%uint64(len(results))]}
	}
	return games
}
