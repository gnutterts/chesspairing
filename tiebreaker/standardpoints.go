package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("standard-points", func() chesspairing.TieBreaker { return &StandardPoints{} })
}

// StandardPoints computes the standard points tiebreaker (FIDE Art. 7.7, STD).
//
// For each round, the player gets:
//   - 1 if they scored more points than their opponent
//   - 0.5 if they scored the same
//   - 0 if they scored fewer
//
// For unplayed rounds (byes/absences), the awarded points are compared to 0.5
// (the draw value): PAB(1.0)→1, half-bye(0.5)→0.5, zero-bye/absent(0)→0.
//
// FIDE Category B tiebreaker.
type StandardPoints struct{}

func (sp *StandardPoints) ID() string   { return "standard-points" }
func (sp *StandardPoints) Name() string { return "Standard Points" }

func (sp *StandardPoints) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	const drawValue = 0.5

	// Build per-round, per-player points awarded.
	type roundResult struct {
		points    float64
		hasOpp    bool
		oppPoints float64
	}
	playerRounds := make(map[string][]roundResult)

	// Initialize all active players.
	for _, p := range state.Players {
		if p.Active {
			playerRounds[p.ID] = make([]roundResult, len(state.Rounds))
		}
	}

	for roundIdx, round := range state.Rounds {
		played := make(map[string]bool)

		for _, game := range round.Games {
			var whitePoints, blackPoints float64

			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				whitePoints, blackPoints = 1.0, 0.0
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				whitePoints, blackPoints = 0.0, 1.0
			case chesspairing.ResultDraw:
				whitePoints, blackPoints = 0.5, 0.5
			case chesspairing.ResultDoubleForfeit:
				whitePoints, blackPoints = 0.0, 0.0
			default:
				continue
			}

			if _, ok := playerRounds[game.WhiteID]; ok {
				playerRounds[game.WhiteID][roundIdx] = roundResult{
					points: whitePoints, hasOpp: true, oppPoints: blackPoints,
				}
			}
			if _, ok := playerRounds[game.BlackID]; ok {
				playerRounds[game.BlackID][roundIdx] = roundResult{
					points: blackPoints, hasOpp: true, oppPoints: whitePoints,
				}
			}
			played[game.WhiteID] = true
			played[game.BlackID] = true
		}

		for _, bye := range round.Byes {
			if _, ok := playerRounds[bye.PlayerID]; !ok {
				continue
			}
			var pts float64
			switch bye.Type {
			case chesspairing.ByePAB:
				pts = 1.0
			case chesspairing.ByeHalf:
				pts = 0.5
			default:
				pts = 0.0
			}
			playerRounds[bye.PlayerID][roundIdx] = roundResult{
				points: pts, hasOpp: false,
			}
			played[bye.PlayerID] = true
		}

		// Absent players: 0 points, no opponent (already zero-initialized).
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		var std float64
		for _, rr := range playerRounds[ps.PlayerID] {
			if rr.hasOpp {
				// Compare with opponent.
				if rr.points > rr.oppPoints {
					std += 1.0
				} else if rr.points == rr.oppPoints {
					std += 0.5
				}
			} else {
				// Unplayed round: compare awarded points to draw value.
				if rr.points > drawValue {
					std += 1.0
				} else if rr.points == drawValue {
					std += 0.5
				}
			}
		}
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    std,
		}
	}
	return result, nil
}
