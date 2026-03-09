package burstein

import (
	"context"
	"errors"
	"fmt"

	chesspairing "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// ErrTooFewPlayers is returned when there aren't enough active players.
var ErrTooFewPlayers = errors.New("burstein pairing requires at least 2 active players")

// ErrNoPairingPossible is returned when no valid pairing can be found.
var ErrNoPairingPossible = errors.New("no valid pairing exists for the remaining players")

// Pair generates pairings for the next round using the Burstein system (C.04.4.2).
//
// Algorithm:
//  1. Build PlayerState for all active players
//  2. Determine if this is a seeding round or post-seeding round
//  3. Seeding rounds: delegate to Dutch matching (same S1/S2 split + criteria)
//  4. Post-seeding rounds: re-rank by opposition index, then use Dutch matching
//  5. Use BursteinByeSelector for bye assignment
//  6. Use BursteinOptimizationCriteria (C10-C13 only, no float criteria)
//  7. AllocateColor with topScorerRules=false
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

	// Determine total rounds for seeding calculation.
	totalRounds := p.totalRounds(state)
	isSeeding := IsSeedingRound(state.CurrentRound, totalRounds)

	if isSeeding {
		notes = append(notes, fmt.Sprintf("Seeding round %d of %d", state.CurrentRound, SeedingRounds(totalRounds)))
	} else {
		notes = append(notes, fmt.Sprintf("Post-seeding round %d (opposition index ranking)", state.CurrentRound))
		// Re-rank players by opposition index.
		players = RankByOppositionIndex(players, state)
	}

	// Assign PAB if odd number of players.
	var byePlayer *swisslib.PlayerState
	activePlayers := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		activePlayers[i] = &players[i]
	}

	if swisslib.NeedsBye(len(activePlayers)) {
		byeSelector := swisslib.BursteinByeSelector{}
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

	// Build player states slice for BuildScoreGroups.
	playerStates := make([]swisslib.PlayerState, len(activePlayers))
	for i, ap := range activePlayers {
		playerStates[i] = *ap
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

	// Build score groups and brackets.
	scoreGroups := swisslib.BuildScoreGroups(playerStates)
	brackets := swisslib.BuildBrackets(scoreGroups)

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
		TopScorers:     map[string]bool{}, // Burstein: no topscorer rules
		ForbiddenPairs: buildForbiddenPairSet(p.opts.ForbiddenPairs),
	}

	// Burstein uses only color criteria C10-C13.
	// No look-ahead (C8) and no float criteria (C14-C21).
	criteria := bursteinOptimizationCriteria()

	// Process brackets top-down using Dutch matching algorithm.
	var allPairs []swisslib.ProposedPairing
	var pendingFloaters []*swisslib.PlayerState

	for i, bracket := range brackets {
		// Merge pending floaters into this bracket.
		if len(pendingFloaters) > 0 {
			bracket = swisslib.MergeIntoHeterogeneous(bracket, pendingFloaters)
			pendingFloaters = nil
		}

		result, err := dutch.MatchBracket(bracket, critCtx, criteria)
		if err != nil {
			// Bracket failed — collapse with next if possible.
			if i+1 < len(brackets) {
				brackets[i+1] = swisslib.CollapseBrackets(bracket, brackets[i+1])
				notes = append(notes, fmt.Sprintf("Collapsed brackets at score %.1f", bracket.OriginalScore))
				continue
			}
			return nil, fmt.Errorf("%w: failed at bracket score %.1f", ErrNoPairingPossible, bracket.OriginalScore)
		}

		allPairs = append(allPairs, result.Pairs...)

		// Floaters from this bracket float down to the next bracket.
		if len(result.Floaters) > 0 {
			if i+1 < len(brackets) {
				pendingFloaters = result.Floaters
			} else {
				// Last bracket — floaters have nowhere to go.
				if len(result.Pairs) == 0 {
					return nil, fmt.Errorf("%w: failed at bracket score %.1f", ErrNoPairingPossible, bracket.OriginalScore)
				}
				notes = append(notes, fmt.Sprintf("%d players could not be paired", len(result.Floaters)))
			}
		}
	}

	// Defensive: any remaining floaters after all brackets.
	if len(pendingFloaters) > 0 {
		notes = append(notes, fmt.Sprintf("%d players could not be paired", len(pendingFloaters)))
	}

	// Allocate colors and build final pairings.
	// Burstein: topScorerRules=false.
	topSeedColor := parseTopSeedColor(p.opts.TopSeedColor)
	pairings := make([]chesspairing.GamePairing, len(allPairs))
	for i, pair := range allPairs {
		whiteID, blackID := swisslib.AllocateColor(pair.White, pair.Black, false, i+1, topSeedColor)
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

	result.Notes = append(result.Notes, "Pairings generated by Burstein Swiss system (C.04.4.2)")

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

// bursteinOptimizationCriteria returns the Burstein-specific optimization
// criteria: only C10-C13 (color criteria). No C8 look-ahead and no float
// criteria (C14-C21), matching FIDE C.04.4.2 specification.
func bursteinOptimizationCriteria() []swisslib.OptimizationCriterion {
	criteria := make([]swisslib.OptimizationCriterion, swisslib.NumViolations)
	criteria[swisslib.IdxC10] = swisslib.CriterionC10
	criteria[swisslib.IdxC11] = swisslib.CriterionC11
	criteria[swisslib.IdxC12] = swisslib.CriterionC12
	criteria[swisslib.IdxC13] = swisslib.CriterionC13
	return criteria
}

// totalRounds returns the total number of rounds for seeding calculation.
// Uses options override if set, otherwise derives from state.
func (p *Pairer) totalRounds(state *chesspairing.TournamentState) int {
	if p.opts.TotalRounds != nil {
		return *p.opts.TotalRounds
	}

	// Derive from state: use CurrentRound as best estimate if larger
	// than completed rounds.
	total := state.CurrentRound
	if total < len(state.Rounds)+1 {
		total = len(state.Rounds) + 1
	}
	return total
}
