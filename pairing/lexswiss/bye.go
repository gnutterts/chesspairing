package lexswiss

import "sort"

// NeedsBye returns true if the participant count is odd.
func NeedsBye(count int) bool {
	return count%2 == 1
}

// AssignPAB selects the participant to receive the pairing-allocated bye
// per Art. 3.4 (shared by Double-Swiss and Team Swiss):
//
//  1. Lowest score
//  2. Among ties: highest TPN (lowest ranking)
//  3. Must not have already received a PAB
//
// Returns nil if all participants have already received a PAB.
func AssignPAB(participants []*ParticipantState) *ParticipantState {
	// Filter to participants who haven't received a bye.
	var eligible []*ParticipantState
	for _, p := range participants {
		if !p.ByeReceived {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) == 0 {
		return nil
	}

	// Sort by score ascending, then TPN descending (highest TPN = lowest ranking).
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].Score != eligible[j].Score {
			return eligible[i].Score < eligible[j].Score
		}
		return eligible[i].TPN > eligible[j].TPN
	})

	return eligible[0]
}
