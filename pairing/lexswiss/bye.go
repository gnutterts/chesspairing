// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"sort"
)

// NeedsBye returns true if the participant count is odd.
func NeedsBye(count int) bool {
	return count%2 == 1
}

// AssignPAB selects the participant to receive the pairing-allocated bye per
// C.04.5/C.04.6 Article 3.4. It first retains C2-eligible participants whose
// removal leaves a complete C1 pairing, then applies lowest score, most
// matches played, and largest fixed TPN in that order. A nil participant and
// nil error means that no participant is C2-eligible.
func AssignPAB(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool) (*ParticipantState, error) {
	eligible := make([]*ParticipantState, 0, len(participants))
	hasC2Eligible := false
	budget := topDownSearchLimit
	for _, participant := range participants {
		if participant.PABIneligible.Any() {
			continue
		}
		hasC2Eligible = true
		complete, err := completable(ctx, without(participants, []*ParticipantState{participant}), forbidden, &budget)
		if err != nil {
			return nil, err
		}
		if complete {
			eligible = append(eligible, participant)
		}
	}
	if len(eligible) == 0 {
		if hasC2Eligible {
			return nil, ErrNoCompletePairing
		}
		return nil, nil
	}

	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		if eligible[i].MatchesPlayed != eligible[j].MatchesPlayed {
			return eligible[i].MatchesPlayed > eligible[j].MatchesPlayed
		}
		return eligible[i].TPN > eligible[j].TPN
	})
	return eligible[0], nil
}
