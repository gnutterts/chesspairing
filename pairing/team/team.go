// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

// Limitations:
// - It does not derive the team set from Matches; teams must be defined via PlayerEntry.TeamID.
// - It does not pass a secondary score to the colour allocation step (AllocateColor).

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Team Swiss system.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePreAssignedPAB(state, *p.opts.PrimaryScore); err != nil {
		return nil, fmt.Errorf("team: %w", err)
	}
	originalState := state
	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := lexswiss.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// Build participant states (each PlayerEntry represents a team). C.04.6
	// Article 1.2 makes the configured primary component the pairing score.
	participants, err := buildParticipantStates(state, *p.opts.PrimaryScore)
	if err != nil {
		return nil, err
	}
	if len(participants) <= 1 {
		if len(participants) == 1 {
			if participants[0].PABIneligible.Any() {
				return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "team", Err: swisslib.ErrNoPABCandidate}
			}
			result.TeamByes = append(result.TeamByes, chesspairing.ByeEntry{
				PlayerID: participants[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.TeamByes = append(preAssignedByes, result.TeamByes...)
		}
		if err := chesspairing.ValidatePairing(originalState, result); err != nil {
			if pairingErr, ok := err.(*chesspairing.PairingError); ok {
				pairingErr.System = "team"
				if pairingErr.Kind == chesspairing.PairingIncomplete {
					pairingErr.Partial = result
				}
			}
			return nil, err
		}
		return result, nil
	}

	// Resolve options.
	prefType := resolveColorPrefType(*p.opts.ColorPreferenceType)

	// Build forbidden pairs map.
	forbidden := buildForbiddenMap(p.opts.ForbiddenPairs)

	// Build participant pointer slice.
	ptrs := make([]*lexswiss.ParticipantState, len(participants))
	for i := range participants {
		ptrs[i] = &participants[i]
	}

	// Assign PAB if odd number (Art. 3.4).
	if lexswiss.NeedsBye(len(ptrs)) {
		byeTeam, err := lexswiss.AssignPAB(ctx, ptrs, forbidden)
		if err != nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "team", Err: err}
		}
		if byeTeam == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "team", Err: swisslib.ErrNoPABCandidate}
		}
		result.TeamByes = append(result.TeamByes, chesspairing.ByeEntry{
			PlayerID: byeTeam.ID,
			Type:     chesspairing.ByePAB,
		})
		ptrs = removeParticipant(ptrs, byeTeam)
	}

	// Build score groups.
	participantValues := make([]lexswiss.ParticipantState, len(ptrs))
	for i, ptr := range ptrs {
		participantValues[i] = *ptr
	}
	scoreGroups := lexswiss.BuildScoreGroups(participantValues)

	// Determine round context.
	isLastRound := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds
	isLastTwoRounds := p.opts.TotalRounds != nil && state.CurrentRound >= *p.opts.TotalRounds-1

	// Pair brackets from top to bottom with upfloater handling.
	allPairs, err := pairAllBrackets(ctx, scoreGroups, forbidden, prefType, isLastTwoRounds, isLastRound, len(ptrs))
	if err != nil {
		if errors.Is(err, lexswiss.ErrNoCompletePairing) || errors.Is(err, lexswiss.ErrPairingSearchLimit) {
			partial := partialResult(result, allPairs, prefType, isLastRound, p.opts.TopSeedColor)
			if len(preAssignedByes) > 0 {
				partial.TeamByes = append(preAssignedByes, partial.TeamByes...)
			}
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "team", Err: err, Partial: partial}
		}
		// Remaining failures (e.g. overflow from the C8-C10 cost weights) are
		// wrapped like every other Team failure path.
		return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "team", Err: err}
	}

	// Build participant map for lookups.
	participantMap := make(map[string]*lexswiss.ParticipantState, len(ptrs))
	for _, ptr := range ptrs {
		participantMap[ptr.ID] = ptr
	}

	// Allocate colours and build final pairings.
	for boardNum, pair := range allPairs {
		wID, bID := AllocateColor(pair[0], pair[1], prefType, isLastRound, p.opts.TopSeedColor, nil)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Sort boards: max score desc, then min TPN asc.
	sortBoards(result.Pairings, participantMap)

	if len(preAssignedByes) > 0 {
		result.TeamByes = append(preAssignedByes, result.TeamByes...)
	}

	if err := chesspairing.ValidatePairing(originalState, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "team"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
		}
		return nil, err
	}
	return result, nil
}

// validatePreAssignedPAB checks that no team with a pre-assigned ByePAB is
// ineligible under C2 (C.04.6 Article 2.1.2 / Basic Rules Art. 4). It applies
// the same participant states the pairer builds for its own PAB choice, so
// the option-derived match-point and game-point rules decide eligibility.
func validatePreAssignedPAB(state *chesspairing.TournamentState, primary string) error {
	preAssigned := make([]chesspairing.ByeEntry, 0, len(state.PreAssignedByes))
	for _, bye := range state.PreAssignedByes {
		if bye.Type == chesspairing.ByePAB {
			preAssigned = append(preAssigned, bye)
		}
	}
	if len(preAssigned) == 0 {
		return nil
	}
	participants, err := buildParticipantStates(state, primary)
	if err != nil {
		return err
	}
	ineligible := make(map[string]bool, len(participants))
	for _, participant := range participants {
		ineligible[participant.ID] = participant.PABIneligible.Any()
	}
	for _, bye := range preAssigned {
		if ineligible[bye.PlayerID] {
			return fmt.Errorf("team: participant %s is ineligible for pairing-allocated bye under C2 (C.04.6 2.1.2)", bye.PlayerID)
		}
	}
	return nil
}

// partialResult converts brackets completed before a failed top-down search to
// a valid partial pairing result.
func partialResult(result *chesspairing.PairingResult, pairs [][2]*lexswiss.ParticipantState, prefType ColorPrefType, isLastRound bool, topSeedColor *string) *chesspairing.PairingResult {
	for boardNum, pair := range pairs {
		whiteID, blackID := AllocateColor(pair[0], pair[1], prefType, isLastRound, topSeedColor, nil)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{Board: boardNum + 1, WhiteID: whiteID, BlackID: blackID})
	}
	return result
}

// pairAllBrackets applies the top-down bracket process in C.04.6 Article 3.3.2.
func pairAllBrackets(ctx context.Context, scoreGroups []lexswiss.ScoreGroup, forbidden map[[2]string]bool, prefType ColorPrefType, isLastTwoRounds, isLastRound bool, numTeams int) ([][2]*lexswiss.ParticipantState, error) {
	options := lexswiss.TopDownOptions{
		ApplyC7: !isLastTwoRounds,
	}
	costFn, err := bracketCostFunc(numTeams, prefType, isLastTwoRounds, isLastRound)
	if err != nil {
		return nil, err
	}
	options.CostForBracket = costFn
	return lexswiss.PairTopDown(ctx, scoreGroups, forbidden, options)
}

// selectTeamUpfloater applies C.04.6 Article 2.3.4 (C7): except in the
// last two rounds, an upfloater that was a floater in the previous round is
// avoided when C1 and C3 leave a non-floater candidate. C4-C6 are preserved
// by the bracket construction and compatible-candidate fallback.
func selectTeamUpfloater(bracket, target []*lexswiss.ParticipantState, forbidden map[[2]string]bool, isLastTwoRounds bool) *lexswiss.ParticipantState {
	return selectTeamUpfloaterWithContext(context.Background(), bracket, target, forbidden, isLastTwoRounds)
}

func selectTeamUpfloaterWithContext(ctx context.Context, bracket, target []*lexswiss.ParticipantState, forbidden map[[2]string]bool, isLastTwoRounds bool) *lexswiss.ParticipantState {
	if isLastTwoRounds {
		return lexswiss.SelectUpfloaterWithContext(ctx, bracket, target, forbidden)
	}
	return lexswiss.SelectUpfloaterWithPreferenceAndContext(ctx, bracket, target, forbidden, func(participant *lexswiss.ParticipantState) bool {
		return !participant.WasFloater
	})
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

// sortBoards sorts pairings under C.04.2 article 3.6.
func sortBoards(pairings []chesspairing.GamePairing, participants map[string]*lexswiss.ParticipantState) {
	sort.SliceStable(pairings, func(i, j int) bool {
		pi1 := participants[pairings[i].WhiteID]
		pi2 := participants[pairings[i].BlackID]
		pj1 := participants[pairings[j].WhiteID]
		pj2 := participants[pairings[j].BlackID]

		return swisslib.BoardBefore(lexswiss.BoardPlayer(pi1), lexswiss.BoardPlayer(pi2),
			lexswiss.BoardPlayer(pj1), lexswiss.BoardPlayer(pj2))
	})

	for i := range pairings {
		pairings[i].Board = i + 1
	}
}
