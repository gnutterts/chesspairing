// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

// Pair implements chesspairing.Pairer for the Double-Swiss system.
func (p *Pairer) Pair(_ context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := lexswiss.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// Build participant states.
	participants := lexswiss.BuildParticipantStates(state)
	if len(participants) <= 1 {
		if len(participants) == 1 {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		return result, nil
	}

	// Build forbidden pairs map.
	forbidden := buildForbiddenMap(p.opts.ForbiddenPairs)

	// Build participant pointer slice.
	ptrs := make([]*lexswiss.ParticipantState, len(participants))
	for i := range participants {
		ptrs[i] = &participants[i]
	}

	// Assign PAB if odd number.
	if lexswiss.NeedsBye(len(ptrs)) {
		byePlayer := lexswiss.AssignPAB(ptrs)
		if byePlayer != nil {
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: byePlayer.ID,
				Type:     chesspairing.ByePAB,
			})
			ptrs = removeParticipant(ptrs, byePlayer)
		}
	}

	// Build score groups.
	participantValues := make([]lexswiss.ParticipantState, len(ptrs))
	for i, ptr := range ptrs {
		participantValues[i] = *ptr
	}
	scoreGroups := lexswiss.BuildScoreGroups(participantValues)

	// Determine if this is the last round (for criteria relaxation).
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds

	// Build criteria function for C8 (colour preferences).
	// In the last round, C8 is relaxed (no colour criteria checked).
	var criteriaFn lexswiss.CriteriaFunc
	if !isLastRound {
		criteriaFn = func(a, b *lexswiss.ParticipantState) bool {
			return checkC8ColorPreference(a, b)
		}
	}

	// Pair brackets from top to bottom with upfloater handling.
	allPairs := pairAllBrackets(scoreGroups, forbidden, criteriaFn)

	// Build participant map for lookups.
	participantMap := make(map[string]*lexswiss.ParticipantState, len(ptrs))
	for _, ptr := range ptrs {
		participantMap[ptr.ID] = ptr
	}

	// Allocate colours and build final pairings.
	for boardNum, pair := range allPairs {
		wID, bID := AllocateColor(pair[0], pair[1], state.CurrentRound, boardNum+1, p.opts.TopSeedColor)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Sort boards: max score desc, then min TPN asc.
	sortBoards(result.Pairings, participantMap)

	if len(preAssignedByes) > 0 {
		result.Byes = append(preAssignedByes, result.Byes...)
	}

	return result, nil
}

// pairAllBrackets pairs all scoregroups from top to bottom, handling
// upfloaters when a bracket has an odd number of participants.
func pairAllBrackets(scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, criteriaFn lexswiss.CriteriaFunc) [][2]*lexswiss.ParticipantState {
	if len(scoreGroups) == 0 {
		return nil
	}

	// Work with mutable bracket copies.
	type bracket struct {
		participants []*lexswiss.ParticipantState
		score        float64
	}
	brackets := make([]bracket, len(scoreGroups))
	for i, sg := range scoreGroups {
		participants := make([]*lexswiss.ParticipantState, len(sg.Participants))
		copy(participants, sg.Participants)
		brackets[i] = bracket{
			participants: participants,
			score:        sg.Score,
		}
	}

	// Handle upfloaters: if a bracket has odd participants, float the
	// lowest-ranked up to the bracket above.
	for i := len(brackets) - 1; i > 0; i-- {
		if len(brackets[i].participants)%2 == 1 {
			floater := lexswiss.SelectUpfloater(brackets[i].participants, brackets[i-1].participants, forbidden)
			if floater != nil {
				// Remove from current bracket.
				brackets[i].participants = removeParticipant(brackets[i].participants, floater)
				// Add to bracket above.
				brackets[i-1].participants = append(brackets[i-1].participants, floater)
			}
		}
	}

	// Pair each bracket.
	var allPairs [][2]*lexswiss.ParticipantState
	for _, b := range brackets {
		if len(b.participants) < 2 {
			continue
		}
		pairs := lexswiss.PairBracket(b.participants, forbidden, criteriaFn)
		allPairs = append(allPairs, pairs...)
	}

	return allPairs
}

// checkC8ColorPreference checks the Double-Swiss colour criterion (C8):
// both participants should not have had the same Game 1 colour in all
// previous rounds if there's an alternative.
//
// This is a soft criterion — if no pairing satisfies it, PairBracket will
// fall back to the first pairing without it (via the lexicographic
// enumeration trying all candidates).
//
// For the initial implementation, C8 checks that pairing two participants
// with identical colour histories is acceptable. A stricter implementation
// would check if the resulting colour assignment violates the 3-consecutive
// constraint. The 3-consecutive constraint is already enforced by
// AllocateColor, so the pairing itself is always valid — C8 just prefers
// pairings that don't require overriding colour preferences.
func checkC8ColorPreference(a, b *lexswiss.ParticipantState) bool {
	histA := filterPlayed(a.ColorHistory)
	histB := filterPlayed(b.ColorHistory)

	if len(histA) == 0 || len(histB) == 0 {
		return true // no history → no colour constraint
	}

	// Both have 2 consecutive same colours.
	// They MUST get opposite colours next round.
	// If they both need the SAME colour, this pairing violates C8.
	if len(histA) >= 2 && histA[len(histA)-1] == histA[len(histA)-2] &&
		len(histB) >= 2 && histB[len(histB)-1] == histB[len(histB)-2] {
		mustA := histA[len(histA)-1].Opposite()
		mustB := histB[len(histB)-1].Opposite()
		if mustA == mustB {
			return false // both need same colour → can't satisfy both
		}
	}

	return true
}

// buildForbiddenMap builds a lookup map from forbidden pair slices.
func buildForbiddenMap(pairs [][]string) map[[2]string]bool {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[[2]string]bool, len(pairs)*2)
	for _, pair := range pairs {
		if len(pair) == 2 {
			m[[2]string{pair[0], pair[1]}] = true
			m[[2]string{pair[1], pair[0]}] = true
		}
	}
	return m
}

// removeParticipant removes a specific participant from the pointer slice.
func removeParticipant(participants []*lexswiss.ParticipantState, remove *lexswiss.ParticipantState) []*lexswiss.ParticipantState {
	result := make([]*lexswiss.ParticipantState, 0, len(participants)-1)
	for _, p := range participants {
		if p.ID != remove.ID {
			result = append(result, p)
		}
	}
	return result
}

// sortBoards sorts pairings for board ordering:
// max score of pair (desc), then min TPN of pair (asc).
func sortBoards(pairings []chesspairing.GamePairing, participants map[string]*lexswiss.ParticipantState) {
	sort.SliceStable(pairings, func(i, j int) bool {
		pi1 := participants[pairings[i].WhiteID]
		pi2 := participants[pairings[i].BlackID]
		pj1 := participants[pairings[j].WhiteID]
		pj2 := participants[pairings[j].BlackID]

		maxI := pi1.Score
		if pi2 != nil && pi2.Score > maxI {
			maxI = pi2.Score
		}
		maxJ := pj1.Score
		if pj2 != nil && pj2.Score > maxJ {
			maxJ = pj2.Score
		}
		if maxI != maxJ {
			return maxI > maxJ
		}

		minI := pi1.TPN
		if pi2 != nil && pi2.TPN < minI {
			minI = pi2.TPN
		}
		minJ := pj1.TPN
		if pj2 != nil && pj2.TPN < minJ {
			minJ = pj2.TPN
		}
		return minI < minJ
	})

	for i := range pairings {
		pairings[i].Board = i + 1
	}
}
