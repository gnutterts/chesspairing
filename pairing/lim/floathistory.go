// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// previousRoundFloaters reconstructs Lim floaters from the last completed
// round. swisslib.FloatHistory follows Dutch conventions, which mark both
// sides of unequal-score games and point-scoring byes; Lim Art. 3.1 transfers
// only the player that moved, and byes and absences are not transfers. Two
// floaters from the same side of the median who meet each other, or a
// cracked-pair player paired with a down-floater from a higher score, cannot
// be distinguished from the games; only the player farther from the median is
// marked.
func previousRoundFloaters(state *chesspairing.TournamentState) map[string]bool {
	historyEnd := state.CurrentRound - 1
	if historyEnd <= 0 || len(state.Rounds) == 0 {
		return nil
	}
	if historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	previousRound := state.Rounds[historyEnd-1]
	scores := make(map[string]float64)
	points := swisslib.PairingPoints(state.ScoringConfig)
	for ri, round := range state.Rounds[:historyEnd-1] {
		addRoundScores(scores, round, state, ri+1)
	}
	median := medianScoreFor(points.Win(), historyEnd-1)
	floaters := make(map[string]bool)
	for _, game := range previousRound.Games {
		if game.WhiteID == "" || game.BlackID == "" {
			continue
		}
		whiteScore, blackScore := swisslib.ScoreGroupKey(scores[game.WhiteID]), swisslib.ScoreGroupKey(scores[game.BlackID])
		if whiteScore == blackScore {
			continue
		}
		whiteSide := whiteScore - median
		blackSide := blackScore - median
		if whiteSide*blackSide < 0 {
			floaters[game.WhiteID] = true
			floaters[game.BlackID] = true
			continue
		}
		if swisslib.ScoreGroupKey(absFloat(whiteSide)) > swisslib.ScoreGroupKey(absFloat(blackSide)) {
			floaters[game.WhiteID] = true
		} else {
			floaters[game.BlackID] = true
		}
	}
	return floaters
}

func addRoundScores(scores map[string]float64, round chesspairing.RoundData, state *chesspairing.TournamentState, roundNumber int) {
	points := swisslib.PairingPoints(state.ScoringConfig)
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
		if !recorded[player.ID] && state.IsActiveInRound(player.ID, roundNumber) {
			scores[player.ID] += points.Bye(chesspairing.ByeAbsent)
		}
	}
	for id, score := range scores {
		scores[id] = swisslib.ScoreGroupKey(score)
	}
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

// medianScoreFor returns half of the maximum possible score after the given
// number of completed rounds, normalised like the scores it is compared with.
func medianScoreFor(win float64, rounds int) float64 {
	return swisslib.ScoreGroupKey(win * float64(rounds) / 2)
}
