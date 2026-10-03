// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package burstein

import (
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// OppositionIndex is the ranking index specified by C.04.4.2 Article 1.8.
type OppositionIndex struct {
	Buchholz, SonnebornBerger float64
	TPN                       int
}

// ComputeOppositionIndex implements C.04.4.2 Article 1.7.
func ComputeOppositionIndex(player *swisslib.PlayerState, state *chesspairing.TournamentState) OppositionIndex {
	scores := computePairingScores(state)
	own := scores[player.ID]
	var buchholz, sb float64
	zeroRun := precedingZeroByeRun(player.ID, state)
	stagedZeroBye := false
	for _, bye := range state.PreAssignedByes {
		if bye.PlayerID == player.ID && byeIndexPoints(bye.Type) == 0 {
			stagedZeroBye = true
			break
		}
	}
	for ri, round := range state.Rounds {
		for _, bye := range round.Byes {
			if bye.PlayerID == player.ID {
				points := byeIndexPoints(bye.Type)
				// Article 1.7.2's treatment of this series is open to
				// interpretation. Read literally, only earlier byes in a
				// zero-point series that reaches the present count as draws.
				if points == 0 && zeroRun > 0 && ri >= len(state.Rounds)-zeroRun && (ri < len(state.Rounds)-1 || stagedZeroBye) {
					points = .5
				}
				buchholz += own
				sb += points * own
			}
		}
		for _, game := range round.Games {
			if game.WhiteID != player.ID && game.BlackID != player.ID {
				continue
			}
			if game.IsForfeit {
				buchholz += own
				sb += gamePoints(player.ID, game) * own
				continue
			}
			var opponent string
			var points float64
			if game.WhiteID == player.ID {
				opponent = game.BlackID
				points = gamePoints(player.ID, game)
			} else {
				opponent = game.WhiteID
				points = gamePoints(player.ID, game)
			}
			buchholz += scores[opponent]
			sb += points * scores[opponent]
		}
	}
	return OppositionIndex{Buchholz: buchholz, SonnebornBerger: sb, TPN: swisslib.EffectivePairingNumber(player)}
}

func gamePoints(id string, game chesspairing.GameData) float64 {
	switch game.Result {
	case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
		if game.WhiteID == id {
			return 1
		}
	case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
		if game.BlackID == id {
			return 1
		}
	case chesspairing.ResultDraw:
		return .5
	}
	return 0
}

func byeIndexPoints(bye chesspairing.ByeType) float64 {
	switch bye {
	case chesspairing.ByePAB, chesspairing.ByeFullPoint:
		return 1
	case chesspairing.ByeHalf:
		return .5
	}
	return 0
}

func precedingZeroByeRun(id string, state *chesspairing.TournamentState) int {
	run := 0
	for i := len(state.Rounds) - 1; i >= 0; i-- {
		found := false
		for _, bye := range state.Rounds[i].Byes {
			if bye.PlayerID == id && byeIndexPoints(bye.Type) == 0 {
				found = true
				break
			}
		}
		if !found {
			break
		}
		run++
	}
	return run
}

// RankByOppositionIndex ranks players by C.04.4.2 Article 1.8 without changing
// their fixed tournament pairing numbers.
func RankByOppositionIndex(players []swisslib.PlayerState, state *chesspairing.TournamentState) []swisslib.PlayerState {
	sorted := append([]swisslib.PlayerState{}, players...)
	indices := make(map[string]OppositionIndex, len(sorted))
	for i := range sorted {
		indices[sorted[i].ID] = ComputeOppositionIndex(&sorted[i], state)
	}
	sort.SliceStable(sorted, func(i, j int) bool { return rankingCompare(indices[sorted[i].ID], indices[sorted[j].ID]) < 0 })
	return sorted
}

func computePairingScores(state *chesspairing.TournamentState) map[string]float64 {
	scores := make(map[string]float64)
	for _, round := range state.Rounds {
		for _, game := range round.Games {
			scores[game.WhiteID] += gamePoints(game.WhiteID, game)
			scores[game.BlackID] += gamePoints(game.BlackID, game)
		}
		for _, bye := range round.Byes {
			scores[bye.PlayerID] += byeIndexPoints(bye.Type)
		}
	}
	return scores
}
