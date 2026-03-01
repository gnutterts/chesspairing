package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("direct-encounter", func() chesspairing.TieBreaker { return &DirectEncounter{} })
}

// DirectEncounter computes the direct encounter (head-to-head) tiebreaker.
//
// For each group of tied players (same primary score), the direct encounter
// value is the score from games played ONLY against other members of the
// tied group. Games against non-tied players are ignored.
//
// If a player is not tied with anyone, their direct encounter value is 0
// (it doesn't matter since there's nothing to break).
type DirectEncounter struct{}

func (de *DirectEncounter) ID() string   { return "direct-encounter" }
func (de *DirectEncounter) Name() string { return "Direct Encounter" }

func (de *DirectEncounter) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	data := buildOpponentData(state, scores)

	// Group players by their primary score to identify tied groups.
	groups := make(map[float64]map[string]bool)
	for _, ps := range scores {
		if groups[ps.Score] == nil {
			groups[ps.Score] = make(map[string]bool)
		}
		groups[ps.Score][ps.PlayerID] = true
	}

	// For each player, compute their score from games against tied opponents.
	deScores := make(map[string]float64, len(scores))
	for _, ps := range scores {
		tiedGroup := groups[ps.Score]
		if len(tiedGroup) <= 1 {
			continue // not tied with anyone
		}

		var deScore float64
		for _, g := range data.playerGames[ps.PlayerID] {
			if !tiedGroup[g.opponentID] {
				continue // skip games against non-tied players
			}
			switch g.result {
			case resultWin:
				deScore += 1.0
			case resultDraw:
				deScore += 0.5
			case resultLoss:
				// 0
			}
		}
		deScores[ps.PlayerID] = deScore
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    deScores[ps.PlayerID],
		}
	}
	return result, nil
}
