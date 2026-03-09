package dutch

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ErrTooFewPlayers is returned when there aren't enough active players.
var ErrTooFewPlayers = errors.New("swiss pairing requires at least 2 active players")

// ErrNoPairingPossible is returned when no valid pairing can be found.
var ErrNoPairingPossible = errors.New("no valid pairing exists for the remaining players")

// Pair generates pairings for the next round using the FIDE Dutch system (C.04.3).
//
// Algorithm:
//  1. Build PlayerState for all active players
//  2. Assign PAB (bye) if odd count
//  3. Build score groups and initial brackets
//  4. Process brackets top-down with backtracking
//  5. Allocate colors for all paired games
//  6. Order boards per FIDE A.6
//  7. Validate and return PairingResult
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	// Build player states.
	players := swisslib.BuildPlayerStates(state)

	if len(players) == 0 {
		return nil, ErrTooFewPlayers
	}

	var notes []string

	// Handle single player.
	if len(players) == 1 {
		return &chesspairing.PairingResult{
			Byes:  []chesspairing.ByeEntry{{PlayerID: players[0].ID, Type: chesspairing.ByePAB}},
			Notes: []string{players[0].ID + " receives a bye (only active player)"},
		}, nil
	}

	// Assign PAB if odd number of players.
	var byePlayer *swisslib.PlayerState
	activePlayers := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		activePlayers[i] = &players[i]
	}

	if swisslib.NeedsBye(len(activePlayers)) {
		byeSelector := swisslib.DutchByeSelector{}
		byePlayer = byeSelector.SelectBye(activePlayers)
		if byePlayer != nil {
			notes = append(notes, fmt.Sprintf("%s receives PAB (bye)", byePlayer.ID))
			// Remove bye player from pairing pool.
			var remaining []*swisslib.PlayerState
			for _, ap := range activePlayers {
				if ap.ID != byePlayer.ID {
					remaining = append(remaining, ap)
				}
			}
			activePlayers = remaining
		}
	}

	// Build player states slice (for BuildScoreGroups which takes []PlayerState).
	playerStates := make([]swisslib.PlayerState, len(activePlayers))
	for i, ap := range activePlayers {
		playerStates[i] = *ap
	}

	totalRounds := state.CurrentRound // approximate
	if totalRounds < len(state.Rounds)+1 {
		totalRounds = len(state.Rounds) + 1
	}

	// Apply Baku acceleration if configured.
	if p.opts.Acceleration != nil && *p.opts.Acceleration == "baku" {
		gaSize := swisslib.BakuGASize(len(state.Players))
		swisslib.ApplyBakuAcceleration(playerStates, state.CurrentRound, totalRounds, gaSize)
		// Also update the pointer-based activePlayers to reflect PairingScore.
		for i := range activePlayers {
			activePlayers[i].PairingScore = playerStates[i].PairingScore
		}
		notes = append(notes, fmt.Sprintf("Baku acceleration: GA=%d players, VP=%.1f",
			gaSize, swisslib.BakuVirtualPoints(totalRounds, state.CurrentRound, true)))
	}

	// Build score groups.
	scoreGroups := swisslib.BuildScoreGroups(playerStates)

	// Build criteria context.
	playerMap := make(map[string]*swisslib.PlayerState, len(activePlayers))
	for _, ap := range activePlayers {
		playerMap[ap.ID] = ap
	}

	critCtx := &swisslib.CriteriaContext{
		Players:        playerMap,
		TotalRounds:    totalRounds,
		CurrentRound:   state.CurrentRound,
		IsLastRound:    state.CurrentRound == totalRounds,
		TopScorers:     computeTopScorers(activePlayers, totalRounds),
		ForbiddenPairs: buildForbiddenPairSet(p.opts.ForbiddenPairs),
	}

	// Set up C8 look-ahead: wraps MatchBracketFeasible from matching.go.
	critCtx.LookAhead = MatchBracketFeasible

	// Global Blossom matching — mirrors bbpPairings architecture.
	// Processes score groups top-down with a single global matching graph.
	allPairs, pairNotes := pairBracketsGlobal(scoreGroups, critCtx, playerMap)
	notes = append(notes, pairNotes...)

	// Order boards: pairs with higher-scoring players come first. When two
	// pairs share the same max-player score, the pair from the higher bracket
	// (homogeneous pairing) comes before a pair from a lower bracket (floater
	// pairing). Finally, ties are broken by the stronger player's TPN ascending.
	sort.SliceStable(allPairs, func(i, j int) bool {
		pi, pj := allPairs[i], allPairs[j]

		// Primary: maximum player pairing score in each pair (descending).
		maxScoreI := pi.White.PairingScore
		if pi.Black.PairingScore > maxScoreI {
			maxScoreI = pi.Black.PairingScore
		}
		maxScoreJ := pj.White.PairingScore
		if pj.Black.PairingScore > maxScoreJ {
			maxScoreJ = pj.Black.PairingScore
		}
		if maxScoreI != maxScoreJ {
			return maxScoreI > maxScoreJ
		}

		// Secondary: originating bracket score (descending).
		// Distinguishes homogeneous pairs from floater pairs at the same
		// max player score.
		if pi.BracketScore != pj.BracketScore {
			return pi.BracketScore > pj.BracketScore
		}

		// Tertiary: stronger player = lower TPN (ascending).
		minTPNi := pi.White.TPN
		if pi.Black.TPN < minTPNi {
			minTPNi = pi.Black.TPN
		}
		minTPNj := pj.White.TPN
		if pj.Black.TPN < minTPNj {
			minTPNj = pj.Black.TPN
		}
		return minTPNi < minTPNj
	})

	// Allocate colors and build final pairings.
	topSeedColor := parseTopSeedColor(p.opts.TopSeedColor)
	pairings := make([]chesspairing.GamePairing, len(allPairs))
	for i, pair := range allPairs {
		whiteID, blackID := swisslib.AllocateColor(pair.White, pair.Black, critCtx.IsLastRound, i+1, topSeedColor)
		pairings[i] = chesspairing.GamePairing{
			Board:   i + 1,
			WhiteID: whiteID,
			BlackID: blackID,
		}
	}

	// Build result.
	result := &chesspairing.PairingResult{
		Pairings: pairings,
		Notes:    notes,
	}

	if byePlayer != nil {
		result.Byes = []chesspairing.ByeEntry{{PlayerID: byePlayer.ID, Type: chesspairing.ByePAB}}
	}

	result.Notes = append(result.Notes, "Pairings generated by Dutch Swiss system (FIDE C.04.3)")

	return result, nil
}

// parseTopSeedColor converts the TopSeedColor string option to a *swisslib.Color.
// Returns nil for "auto" or "white" (default behavior), and &ColorBlack for "black".
func parseTopSeedColor(opt *string) *swisslib.Color {
	if opt == nil || *opt == "auto" || *opt == "white" {
		return nil
	}
	if *opt == "black" {
		c := swisslib.ColorBlack
		return &c
	}
	return nil
}

// recordFloats appends float direction to each player's FloatHistory for this round.
// floaters = players who floated down from a higher bracket.
// paired = players who were paired in their native bracket.
// playerMap provides mutable access to all players.
func recordFloats(floaters, paired []*swisslib.PlayerState, playerMap map[string]*swisslib.PlayerState) {
	for _, f := range floaters {
		if p, ok := playerMap[f.ID]; ok {
			p.FloatHistory = append(p.FloatHistory, swisslib.FloatDown)
		}
	}
	for _, p := range paired {
		if mp, ok := playerMap[p.ID]; ok {
			mp.FloatHistory = append(mp.FloatHistory, swisslib.FloatNone)
		}
	}
}

// buildForbiddenPairSet converts the options ForbiddenPairs slice into
// the canonicalized map format used by CriteriaContext.
func buildForbiddenPairSet(pairs [][]string) map[[2]string]bool {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[[2]string]bool, len(pairs))
	for _, pair := range pairs {
		if len(pair) == 2 {
			m[swisslib.CanonicalPairKey(pair[0], pair[1])] = true
		}
	}
	return m
}

// computeTopScorers identifies players with >50% of the maximum possible score.
// Only relevant in the final round.
func computeTopScorers(players []*swisslib.PlayerState, totalRounds int) map[string]bool {
	maxScore := float64(totalRounds)
	threshold := maxScore / 2.0

	topScorers := make(map[string]bool)
	for _, pl := range players {
		if pl.Score > threshold {
			topScorers[pl.ID] = true
		}
	}
	return topScorers
}
