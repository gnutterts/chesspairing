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
	pointsPolicy := swisslib.PairingPoints(state.ScoringConfig)
	scores := computePairingScores(state)
	// Article 1.7.2 is interpreted here as making each zero-point bye in a
	// current series a draw only for the player's actual over-the-board
	// opponents. The player's own self-play uses the registered score.
	ids := make(map[string]bool, len(scores)+len(state.Players))
	for id := range scores {
		ids[id] = true
	}
	for _, entry := range state.Players {
		ids[entry.ID] = true
	}
	for _, round := range completedRounds(state) {
		for _, game := range round.Games {
			ids[game.WhiteID], ids[game.BlackID] = true, true
		}
		for _, bye := range round.Byes {
			ids[bye.PlayerID] = true
		}
	}
	indexScores := make(map[string]float64, len(ids))
	for id := range ids {
		indexScores[id] = scores[id] + pointsPolicy.Draw()*float64(precedingZeroByeRun(id, state, pointsPolicy.Bye))
	}
	own := scores[player.ID]
	var buchholz, sb float64
	for _, round := range completedRounds(state) {
		recorded := false
		for _, bye := range round.Byes {
			if bye.PlayerID != player.ID {
				continue
			}
			recorded = true
			points := pointsPolicy.Bye(bye.Type)
			buchholz += own
			sb += points * own
		}
		for _, game := range round.Games {
			if game.WhiteID != player.ID && game.BlackID != player.ID {
				continue
			}
			recorded = true
			if swisslib.IsForfeit(game) {
				buchholz += own
				white, black := pointsPolicy.Game(game)
				sb += gamePoints(player.ID, game, white, black) * own
				continue
			}
			opponent := game.WhiteID
			if opponent == player.ID {
				opponent = game.BlackID
			}
			white, black := pointsPolicy.Game(game)
			points := gamePoints(player.ID, game, white, black)
			buchholz += indexScores[opponent]
			sb += points * indexScores[opponent]
		}
		if !recorded {
			// Article 1.7.2 treats an entirely absent round as self-play.
			buchholz += own
		}
	}
	return OppositionIndex{Buchholz: swisslib.ScoreGroupKey(buchholz), SonnebornBerger: swisslib.ScoreGroupKey(sb), TPN: swisslib.EffectivePairingNumber(player)}
}

func completedRounds(state *chesspairing.TournamentState) []chesspairing.RoundData {
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	return state.Rounds[:historyEnd]
}

func gamePoints(id string, game chesspairing.GameData, white, black float64) float64 {
	if id == game.WhiteID {
		return white
	}
	return black
}

func precedingZeroByeRun(id string, state *chesspairing.TournamentState, byePoints func(chesspairing.ByeType) float64) int {
	rounds := completedRounds(state)
	run := 0
	for i := len(rounds) - 1; i >= 0; i-- {
		recorded := false
		zeroBye := false
		for _, bye := range rounds[i].Byes {
			if bye.PlayerID == id {
				recorded = true
				zeroBye = byePoints(bye.Type) == 0
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
	points := swisslib.PairingPoints(state.ScoringConfig)
	for ri, round := range completedRounds(state) {
		recorded := make(map[string]bool)
		for _, game := range round.Games {
			white, black := points.Game(game)
			scores[game.WhiteID] += white
			scores[game.BlackID] += black
			recorded[game.WhiteID], recorded[game.BlackID] = true, true
		}
		for _, bye := range round.Byes {
			scores[bye.PlayerID] += points.Bye(bye.Type)
			recorded[bye.PlayerID] = true
		}
		for _, player := range state.Players {
			if !recorded[player.ID] && state.IsActiveInRound(player.ID, ri+1) {
				scores[player.ID] += points.Bye(chesspairing.ByeAbsent)
			}
		}
	}
	for id, score := range scores {
		scores[id] = swisslib.ScoreGroupKey(score)
	}
	return scores
}
