package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
)

func init() {
	Register("pairing-number", func() chesspairing.TieBreaker { return &PairingNumber{} })
}

// PairingNumber computes the tournament pairing number tiebreaker
// (FIDE Art. 7.8, TPN).
//
// The value is the negated 1-based index of the player in state.Players.
// Negation ensures that lower TPN (= higher seeding) produces a higher
// tiebreak value, consistent with all other tiebreakers where higher = better.
//
// FIDE Category B tiebreaker.
type PairingNumber struct{}

func (pn *PairingNumber) ID() string   { return "pairing-number" }
func (pn *PairingNumber) Name() string { return "Pairing Number" }

func (pn *PairingNumber) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	// Build TPN lookup: 1-based index in Players slice.
	tpn := make(map[string]int, len(state.Players))
	for i, p := range state.Players {
		tpn[p.ID] = i + 1
	}

	result := make([]chesspairing.TieBreakValue, len(scores))
	for i, ps := range scores {
		result[i] = chesspairing.TieBreakValue{
			PlayerID: ps.PlayerID,
			Value:    -float64(tpn[ps.PlayerID]),
		}
	}
	return result, nil
}
