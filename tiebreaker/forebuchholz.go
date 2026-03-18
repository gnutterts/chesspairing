package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("fore-buchholz", func() chesspairing.TieBreaker { return &ForeBuchholz{} })
}

// ForeBuchholz computes the Fore Buchholz tiebreaker (FIDE Art. 8.3, FB).
//
// This is Buchholz calculated as if all final-round games that have not
// yet been played ended in draws. If all games are complete, Fore Buchholz
// equals regular Buchholz.
//
// The "final round" is the last round in state.Rounds. Any game in that
// round with ResultPending is treated as a draw for scoring purposes.
//
// FIDE Category C tiebreaker.
type ForeBuchholz struct{}

func (fb *ForeBuchholz) ID() string   { return "fore-buchholz" }
func (fb *ForeBuchholz) Name() string { return "Fore Buchholz" }

func (fb *ForeBuchholz) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	if len(state.Rounds) == 0 {
		result := make([]chesspairing.TieBreakValue, len(scores))
		for i, ps := range scores {
			result[i] = chesspairing.TieBreakValue{PlayerID: ps.PlayerID, Value: 0}
		}
		return result, nil
	}

	// Build virtual scores: start from actual scores, then adjust for
	// pending final-round games assumed to be draws.
	virtualScores := make(map[string]float64, len(scores))
	for _, ps := range scores {
		virtualScores[ps.PlayerID] = ps.Score
	}

	lastRound := state.Rounds[len(state.Rounds)-1]
	for _, game := range lastRound.Games {
		if game.Result == chesspairing.ResultPending {
			// Assume draw: +0.5 for each player.
			virtualScores[game.WhiteID] += 0.5
			virtualScores[game.BlackID] += 0.5
		}
	}

	// Build opponent data using all rounds, but skip pending games
	// (buildOpponentData already does this). We need to manually
	// account for pending final-round games as opponent pairings.
	data := buildOpponentData(state, scores)

	// For pending final-round games, add them as virtual game entries
	// and remove the corresponding absence counts (buildOpponentData
	// marks players as absent when their games are pending).
	for _, game := range lastRound.Games {
		if game.Result == chesspairing.ResultPending {
			data.playerGames[game.WhiteID] = append(data.playerGames[game.WhiteID], gameEntry{
				opponentID: game.BlackID,
				result:     resultDraw,
			})
			data.playerGames[game.BlackID] = append(data.playerGames[game.BlackID], gameEntry{
				opponentID: game.WhiteID,
				result:     resultDraw,
			})
			if data.playerAbsences[game.WhiteID] > 0 {
				data.playerAbsences[game.WhiteID]--
			}
			if data.playerAbsences[game.BlackID] > 0 {
				data.playerAbsences[game.BlackID]--
			}
		}
	}

	// Override the score map with virtual scores.
	data.playerScoreMap = virtualScores

	totalRounds := len(state.Rounds)

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		oppScores := opponentScores(ps.PlayerID, data, totalRounds)
		var sum float64
		for _, s := range oppScores {
			sum += s
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    sum,
		}
	}
	return result, nil
}
