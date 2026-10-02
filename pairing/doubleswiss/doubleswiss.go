// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"context"
	"errors"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Double-Swiss system.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	originalState := state
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := lexswiss.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// Build participant states.
	participants, err := lexswiss.BuildParticipantStates(state)
	if err != nil {
		return nil, err
	}
	if len(participants) <= 1 {
		if len(participants) == 1 {
			if participants[0].PABIneligible.Any() {
				return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "doubleswiss", Err: swisslib.ErrNoPABCandidate}
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		if err := chesspairing.ValidatePairing(originalState, result); err != nil {
			if pairingErr, ok := err.(*chesspairing.PairingError); ok {
				pairingErr.System = "doubleswiss"
				if pairingErr.Kind == chesspairing.PairingIncomplete {
					pairingErr.Partial = result
				}
			}
			return nil, err
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
		if byePlayer == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "doubleswiss", Err: swisslib.ErrNoPABCandidate}
		}
		result.Byes = append(result.Byes, chesspairing.ByeEntry{
			PlayerID: byePlayer.ID,
			Type:     chesspairing.ByePAB,
		})
		ptrs = removeParticipant(ptrs, byePlayer)
	}

	// Build score groups.
	participantValues := make([]lexswiss.ParticipantState, len(ptrs))
	for i, ptr := range ptrs {
		participantValues[i] = *ptr
	}
	scoreGroups := lexswiss.BuildScoreGroups(participantValues)

	// Determine if this is the last round (for criteria relaxation).
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds

	// Build the colour pairing criterion. In the last round it is relaxed.
	var criteriaFn lexswiss.CriteriaFunc
	if !isLastRound {
		criteriaFn = colorPairingPreference
	}

	// Pair brackets from top to bottom with upfloater handling.
	allPairs, err := pairAllBrackets(ctx, scoreGroups, forbidden, criteriaFn)
	if err != nil {
		if errors.Is(err, lexswiss.ErrNoCompletePairing) {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "doubleswiss", Err: err}
		}
		return nil, err
	}

	// Build participant map for lookups.
	participantMap := make(map[string]*lexswiss.ParticipantState, len(ptrs))
	for _, ptr := range ptrs {
		participantMap[ptr.ID] = ptr
	}

	parityRank := participantRanks(participants)

	// Allocate colours and build final pairings.
	for boardNum, pair := range allPairs {
		first, second := *pair[0], *pair[1]
		first.PairingNumber, second.PairingNumber = parityRank[first.ID], parityRank[second.ID]
		wID, bID := AllocateColor(&first, &second, p.opts.TopSeedColor)
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

	if err := chesspairing.ValidatePairing(originalState, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "doubleswiss"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
		}
		return nil, err
	}
	return result, nil
}

// pairAllBrackets pairs all scoregroups from top to bottom, handling
// upfloaters when a bracket has an odd number of participants.
func pairAllBrackets(ctx context.Context, scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, criteriaFn lexswiss.CriteriaFunc) ([][2]*lexswiss.ParticipantState, error) {
	if len(scoreGroups) == 0 {
		return nil, nil
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

	// Pair each bracket. If independent brackets cannot be completed, retry
	// all participants together: a legal cross-bracket pairing is preferable
	// to silently omitting participants.
	var allPairs [][2]*lexswiss.ParticipantState
	complete := true
	for _, b := range brackets {
		if len(b.participants)%2 == 1 {
			complete = false
			break
		}
		if len(b.participants) == 0 {
			continue
		}
		pairs, err := lexswiss.PairBracket(ctx, b.participants, forbidden, criteriaFn)
		if err != nil {
			if !errors.Is(err, lexswiss.ErrNoCompletePairing) {
				return nil, err
			}
			complete = false
			break
		}
		allPairs = append(allPairs, pairs...)
	}
	if complete {
		return allPairs, nil
	}
	var participants []*lexswiss.ParticipantState
	for _, b := range brackets {
		participants = append(participants, b.participants...)
	}
	return lexswiss.PairBracket(ctx, participants, forbidden, criteriaFn)
}

// colorPairingPreference avoids pairing participants that both require the
// same next colour when an alternative pairing is available.
func colorPairingPreference(a, b *lexswiss.ParticipantState) bool {
	histA := filterPlayed(a.ColorHistory)
	histB := filterPlayed(b.ColorHistory)

	if len(histA) < 2 || len(histB) < 2 {
		return true
	}
	if histA[len(histA)-1] != histA[len(histA)-2] || histB[len(histB)-1] != histB[len(histB)-2] {
		return true
	}
	return histA[len(histA)-1] != histB[len(histB)-1]
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

// participantRanks numbers the participants of the round 1, 2, 3, ... in order
// of their fixed pairing number. Rules 4.2 and 4.3.1 look at the TPN of a
// player; players who do not take part in the round (a requested bye, an
// absence) do not count, the way the FIDE reference implementation numbers the
// players of the Dutch system. The player who receives the pairing-allocated
// bye takes part in the round and does count.
func participantRanks(participants []lexswiss.ParticipantState) map[string]int {
	sorted := make([]lexswiss.ParticipantState, len(participants))
	copy(sorted, participants)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].PairingNumber < sorted[j].PairingNumber })
	ranks := make(map[string]int, len(sorted))
	for i := range sorted {
		ranks[sorted[i].ID] = i + 1
	}
	return ranks
}
