// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"context"
	"errors"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// pairDutchSeeding implements C.04.4.2 Article 1.6: the seeding rounds are
// paired by the Dutch system (C.04.3), so the Dutch pairer is used as is.
func (p *Pairer) pairDutchSeeding(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	result, err := dutch.New(dutch.Options{
		Acceleration:   p.opts.Acceleration,
		TopSeedColor:   p.opts.TopSeedColor,
		TotalRounds:    p.opts.TotalRounds,
		ForbiddenPairs: p.opts.ForbiddenPairs,
	}).Pair(ctx, state)
	if err != nil {
		var pairingErr *chesspairing.PairingError
		switch {
		case errors.Is(err, dutch.ErrTooFewPlayers):
			return nil, ErrTooFewPlayers
		case errors.As(err, &pairingErr) && pairingErr.System == "dutch":
			relabelled := *pairingErr
			relabelled.System = "burstein"
			return nil, &relabelled
		}
		return nil, err
	}
	return result, nil
}

// colorParityRanks follows the Dutch entered-player TPN numbering used by
// C.04.4.2 Article 5.2.1.
func colorParityRanks(state *chesspairing.TournamentState, active []*swisslib.PlayerState) map[string]int {
	numbers := make(map[string]int, len(state.Players))
	if numbered, err := chesspairing.AssignPairingNumbers(state.Players); err == nil {
		for _, player := range numbered {
			numbers[player.ID] = player.PairingNumber
		}
	}
	entered := make(map[string]bool, len(active))
	for _, player := range active {
		entered[player.ID] = true
		if _, ok := numbers[player.ID]; !ok {
			numbers[player.ID] = swisslib.EffectivePairingNumber(player)
		}
	}
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	for _, round := range state.Rounds[:historyEnd] {
		for _, game := range round.Games {
			entered[game.WhiteID], entered[game.BlackID] = true, true
		}
		for _, bye := range round.Byes {
			if bye.Type == chesspairing.ByePAB {
				entered[bye.PlayerID] = true
			}
		}
	}
	ids := make([]string, 0, len(entered))
	for id := range entered {
		if _, ok := numbers[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		if numbers[ids[i]] != numbers[ids[j]] {
			return numbers[ids[i]] < numbers[ids[j]]
		}
		return ids[i] < ids[j]
	})
	ranks := make(map[string]int, len(ids))
	for i, id := range ids {
		ranks[id] = i + 1
	}
	return ranks
}
