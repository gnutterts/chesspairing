// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import "github.com/gnutterts/chesspairing"

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
	for _, round := range state.Rounds[:historyEnd-1] {
		addRoundScores(scores, round)
	}
	median := float64(historyEnd-1) / 2
	floaters := make(map[string]bool)
	for _, game := range previousRound.Games {
		if game.WhiteID == "" || game.BlackID == "" {
			continue
		}
		whiteScore, blackScore := scores[game.WhiteID], scores[game.BlackID]
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
		if absFloat(whiteSide) > absFloat(blackSide) {
			floaters[game.WhiteID] = true
		} else {
			floaters[game.BlackID] = true
		}
	}
	return floaters
}

func addRoundScores(scores map[string]float64, round chesspairing.RoundData) {
	for _, game := range round.Games {
		switch game.Result {
		case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
			scores[game.WhiteID]++
		case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
			scores[game.BlackID]++
		case chesspairing.ResultDraw, chesspairing.ResultPending:
			scores[game.WhiteID] += 0.5
			scores[game.BlackID] += 0.5
		}
	}
	for _, bye := range round.Byes {
		switch bye.Type {
		case chesspairing.ByePAB, chesspairing.ByeFullPoint:
			scores[bye.PlayerID]++
		case chesspairing.ByeHalf:
			scores[bye.PlayerID] += 0.5
		}
	}
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
