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
	// Article 1.7.2 is interpreted here as making each zero-point bye in a
	// current series a draw for the player's actual over-the-board opponents.
	// The player's registered points for the bye remain unchanged.
	indexScores := make(map[string]float64, len(state.Players))
	for _, entry := range state.Players {
		indexScores[entry.ID] = scores[entry.ID] + .5*float64(precedingZeroByeRun(entry.ID, state))
	}
	own := indexScores[player.ID]
	var buchholz, sb float64
	for _, round := range completedRounds(state) {
		recorded := false
		for _, bye := range round.Byes {
			if bye.PlayerID != player.ID {
				continue
			}
			recorded = true
			points := byeIndexPoints(bye.Type)
			buchholz += own
			sb += points * own
		}
		for _, game := range round.Games {
			if game.WhiteID != player.ID && game.BlackID != player.ID {
				continue
			}
			recorded = true
			if game.IsForfeit {
				buchholz += own
				sb += gamePoints(player.ID, game) * own
				continue
			}
			opponent := game.WhiteID
			if opponent == player.ID {
				opponent = game.BlackID
			}
			points := gamePoints(player.ID, game)
			buchholz += indexScores[opponent]
			sb += points * indexScores[opponent]
		}
		if !recorded {
			// Article 1.7.2 treats an entirely absent round as self-play.
			buchholz += own
		}
	}
	return OppositionIndex{Buchholz: buchholz, SonnebornBerger: sb, TPN: swisslib.EffectivePairingNumber(player)}
}

func completedRounds(state *chesspairing.TournamentState) []chesspairing.RoundData {
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	return state.Rounds[:historyEnd]
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
	rounds := completedRounds(state)
	run := 0
	for i := len(rounds) - 1; i >= 0; i-- {
		recorded := false
		zeroBye := false
		for _, bye := range rounds[i].Byes {
			if bye.PlayerID == id {
				recorded = true
				zeroBye = byeIndexPoints(bye.Type) == 0
				break
			}
		}
		for _, game := range rounds[i].Games {
			if game.WhiteID == id || game.BlackID == id {
				recorded = true
				zeroBye = false
				break
			}
		}
		if recorded && !zeroBye {
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
	for _, round := range completedRounds(state) {
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
