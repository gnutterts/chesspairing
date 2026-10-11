// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import (
	"context"
	"errors"
	"sort"
)

// ErrPairingSearchLimit is returned when top-down bracket selection exceeds
// its bounded search budget.
var ErrPairingSearchLimit = errors.New("pairing search limit reached")

const (
	topDownSearchLimit  = 100000
	topDownProfileLimit = 100000
	topDownSetLimit     = 100000
)

// TopDownOptions configures the common C.04.5/C.04.6 Article 3.3.2 bracket
// construction. Criteria applies the system's bracket-pairing quality rules.
// CriteriaForBracket may add criteria which depend on the chosen upfloaters.
type TopDownOptions struct {
	Criteria           CriteriaFunc
	CriteriaForBracket func(CriteriaFunc, []*ParticipantState) CriteriaFunc
	CostForBracket     func([]*ParticipantState) PairCostFunc
	RelaxCriteriaForC3 bool
	ApplyC7            bool
}

// PairTopDown pairs scoregroups from the highest remaining score downwards.
// For every top-scoregroup it considers upfloater sets in C4, C5, and 3.5.4
// order, and only accepts a set when its bracket and the remaining players
// comply with C1/C3. It deliberately does not pool failed brackets.
func PairTopDown(ctx context.Context, scoreGroups []ScoreGroup, forbidden map[[2]string]bool, options TopDownOptions) ([][2]*ParticipantState, error) {
	remaining := make([]*ParticipantState, 0)
	for _, group := range scoreGroups {
		remaining = append(remaining, group.Participants...)
	}
	var pairs [][2]*ParticipantState
	budget := topDownSearchLimit
	for len(remaining) > 0 {
		residents := topScoreGroup(remaining)
		set, bracketPairs, err := selectUpfloaterSet(ctx, residents, remaining[len(residents):], forbidden, options, &budget)
		if err != nil {
			return pairs, err
		}
		remaining = without(remaining, append(append([]*ParticipantState{}, residents...), set...))
		pairs = append(pairs, bracketPairs...)
	}
	return pairs, nil
}

func selectUpfloaterSet(ctx context.Context, residents, potential []*ParticipantState, forbidden map[[2]string]bool, options TopDownOptions, budget *int) ([]*ParticipantState, [][2]*ParticipantState, error) {
	for k := 0; k <= len(potential); k++ {
		if (len(residents)+k)%2 != 0 {
			continue
		}
		pairSet := func(set []*ParticipantState, cost PairCostFunc) ([][2]*ParticipantState, error) {
			bracket := append(append([]*ParticipantState{}, residents...), set...)
			criteria := options.Criteria
			if options.CriteriaForBracket != nil {
				criteria = options.CriteriaForBracket(criteria, set)
			}
			if err := spend(ctx, budget); err != nil {
				return nil, err
			}
			pairs, err := pairBracket(ctx, bracket, forbidden, criteria, cost, budget)
			if errors.Is(err, ErrNoCompletePairing) && options.RelaxCriteriaForC3 {
				// C3 takes priority over bracket quality criteria: only C1 and C2
				// are absolute (C.04.5/C.04.6 Article 2.1).
				if options.CriteriaForBracket != nil {
					if err := spend(ctx, budget); err != nil {
						return nil, err
					}
					pairs, err = pairBracket(ctx, bracket, forbidden, options.Criteria, cost, budget)
				}
				if errors.Is(err, ErrNoCompletePairing) && options.Criteria != nil {
					if err := spend(ctx, budget); err != nil {
						return nil, err
					}
					pairs, err = pairBracket(ctx, bracket, forbidden, nil, cost, budget)
				}
			}
			return pairs, err
		}
		var working []workingSet
		err := upfloaterSetProfiles(ctx, potential, k, func(run func(func([]*ParticipantState) error) error) error {
			err := run(func(set []*ParticipantState) error {
				// C8 does not take part in the set choice (3.5.5), so working
				// sets are paired without it; only the chosen set is minimised.
				pairs, err := pairSet(set, nil)
				if err != nil {
					if errors.Is(err, ErrNoCompletePairing) {
						return nil
					}
					return err
				}
				rest := without(append(append([]*ParticipantState{}, residents...), potential...), append(append([]*ParticipantState{}, residents...), set...))
				complete, err := completable(ctx, rest, forbidden, budget)
				if err != nil {
					return err
				}
				if complete {
					working = append(working, workingSet{set: set, pairs: pairs, c7: floaterCount(set, options.ApplyC7)})
				}
				return nil
			})
			if err != nil {
				return err
			}
			if len(working) > 0 {
				return errWorkingProfile
			}
			return nil
		})
		if errors.Is(err, errWorkingProfile) {
			err = nil
		}
		if err != nil {
			return nil, nil, err
		}
		if len(working) == 0 {
			continue
		}

		// C.04.5/C.04.6 2.3.3 says to choose the set so C1, C3 and C4 are
		// complied with in the following bracket. Like Dutch 2.4.3 [C8],
		// compare the working sets by how well they comply in that bracket.
		bestC6 := -1
		for i := range working {
			count, err := c6UpfloaterCount(ctx, residents, potential, working[i].set, forbidden, budget)
			if errors.Is(err, ErrNoCompletePairing) {
				// A working set leaves a completable rest, so this cannot
				// happen; rank it last rather than abort the round.
				count, err = len(potential)+1, nil
			}
			if err != nil {
				return nil, nil, err
			}
			working[i].c6Count = count
			if count >= 0 && (bestC6 < 0 || count < bestC6) {
				bestC6 = count
			}
		}
		for i := range working {
			working[i].c6 = working[i].c6Count < 0 || working[i].c6Count == bestC6
		}
		sort.SliceStable(working, func(i, j int) bool {
			if working[i].c6 != working[j].c6 {
				return working[i].c6
			}
			if working[i].c7 != working[j].c7 {
				return working[i].c7 < working[j].c7
			}
			return compareSetOrder(working[i].set, working[j].set) < 0
		})
		winner := working[0]
		if cost := pairCostForBracket(options, winner.set); cost != nil {
			pairs, err := pairSet(winner.set, cost)
			if err != nil {
				return nil, nil, err
			}
			winner.pairs = pairs
		}
		return winner.set, winner.pairs, nil
	}
	return nil, nil, ErrNoCompletePairing
}

type workingSet struct {
	set     []*ParticipantState
	pairs   [][2]*ParticipantState
	c6      bool
	c6Count int
	c7      int
}

func pairBracket(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, criteria CriteriaFunc, cost PairCostFunc, budget *int) ([][2]*ParticipantState, error) {
	return PairBracketMinimizing(ctx, participants, forbidden, criteria, cost, budget)
}

func pairCostForBracket(options TopDownOptions, set []*ParticipantState) PairCostFunc {
	if options.CostForBracket == nil {
		return nil
	}
	return options.CostForBracket(set)
}

func topScoreGroup(participants []*ParticipantState) []*ParticipantState {
	score := participants[0].Score
	end := 0
	for end < len(participants) && participants[end].Score == score {
		end++
	}
	return participants[:end]
}

var errWorkingProfile = errors.New("working upfloater profile found")

// upfloaterSets visits sets profile-first: C5's best score profile first, and
// then Article 3.5.4 lexicographic TPN order within that profile.
func upfloaterSets(ctx context.Context, potential []*ParticipantState, k int, visit func([]*ParticipantState) error) error {
	return upfloaterSetProfiles(ctx, potential, k, func(run func(func([]*ParticipantState) error) error) error {
		return run(visit)
	})
}

// upfloaterSetProfiles generates score profiles lazily. This avoids building
// C(n,k) sets before the first C5 profile has been tested.
func upfloaterSetProfiles(ctx context.Context, potential []*ParticipantState, k int, visitProfile func(func(func([]*ParticipantState) error) error) error) error {
	ordered := append([]*ParticipantState(nil), potential...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Score != ordered[j].Score {
			return ordered[i].Score > ordered[j].Score
		}
		return ordered[i].TPN < ordered[j].TPN
	})
	var groups [][]*ParticipantState
	for start := 0; start < len(ordered); {
		end := start + 1
		for end < len(ordered) && ordered[end].Score == ordered[start].Score {
			end++
		}
		groups = append(groups, ordered[start:end])
		start = end
	}
	// capacity[g] is the number of players in groups 0..g-1.
	capacity := make([]int, len(groups)+1)
	for g, group := range groups {
		capacity[g+1] = capacity[g] + len(group)
	}
	profiles := 0
	visited := 0
	counts := make([]int, len(groups))
	visitCounts := func() error {
		if profiles >= topDownProfileLimit {
			return ErrPairingSearchLimit
		}
		profiles++
		return visitProfile(func(visit func([]*ParticipantState) error) error {
			var choose func(int, []*ParticipantState) error
			choose = func(group int, set []*ParticipantState) error {
				if group == len(groups) {
					if visited >= topDownSetLimit {
						return ErrPairingSearchLimit
					}
					visited++
					// Copy: set shares its backing array with sibling sets.
					return visit(append([]*ParticipantState(nil), set...))
				}
				return chooseFromGroup(groups[group], counts[group], func(chosen []*ParticipantState) error {
					return choose(group+1, append(set, chosen...))
				})
			}
			return choose(0, nil)
		})
	}
	// C5 maximises the upfloaters' scores taken in ascending order, so the
	// profile with fewer players from the lowest group comes first, then
	// fewer from the next lowest group, and so on: walk the groups from the
	// lowest score upwards and give each the smallest count first.
	var generate func(int, int) error
	generate = func(g, left int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if g < 0 {
			if left != 0 {
				return nil
			}
			return visitCounts()
		}
		for c := 0; c <= len(groups[g]) && c <= left; c++ {
			if left-c > capacity[g] {
				continue
			}
			counts[g] = c
			if err := generate(g-1, left-c); err != nil {
				return err
			}
		}
		counts[g] = 0
		return nil
	}
	return generate(len(groups)-1, k)
}

func chooseFromGroup(group []*ParticipantState, count int, visit func([]*ParticipantState) error) error {
	var choose func(int, []*ParticipantState) error
	choose = func(start int, chosen []*ParticipantState) error {
		if len(chosen) == count {
			return visit(chosen)
		}
		for i := start; i <= len(group)-(count-len(chosen)); i++ {
			if err := choose(i+1, append(chosen, group[i])); err != nil {
				return err
			}
		}
		return nil
	}
	return choose(0, nil)
}

func compareSetOrder(first, second []*ParticipantState) int {
	for i := range first {
		if first[i].TPN < second[i].TPN {
			return -1
		}
		if first[i].TPN > second[i].TPN {
			return 1
		}
	}
	return 0
}

func sameProfile(first, second []*ParticipantState) bool {
	for i := range first {
		if first[i].Score != second[i].Score {
			return false
		}
	}
	return true
}

func floaterCount(set []*ParticipantState, apply bool) int {
	if !apply {
		return 0
	}
	count := 0
	for _, participant := range set {
		if participant.WasFloater {
			count++
		}
	}
	return count
}

// c6UpfloaterCount returns the smallest number of upfloaters which lets the
// following scoregroup satisfy C1 and C3. A negative result means that the
// selected set emptied that scoregroup, so C6 does not apply.
func c6UpfloaterCount(ctx context.Context, residents, potential, selected []*ParticipantState, forbidden map[[2]string]bool, budget *int) (int, error) {
	if len(potential) == 0 {
		return -1, nil
	}
	following := topScoreGroup(potential)
	selectedIDs := make(map[string]bool, len(selected))
	for _, participant := range selected {
		selectedIDs[participant.ID] = true
	}
	emptied := true
	for _, participant := range following {
		if !selectedIDs[participant.ID] {
			emptied = false
			break
		}
	}
	if emptied {
		return -1, nil
	}
	remaining := without(append(append([]*ParticipantState{}, residents...), potential...), append(append([]*ParticipantState{}, residents...), selected...))
	next := topScoreGroup(remaining)
	for k := 0; k <= len(remaining)-len(next); k++ {
		if (len(next)+k)%2 != 0 {
			continue
		}
		found := false
		err := upfloaterSets(ctx, remaining[len(next):], k, func(set []*ParticipantState) error {
			if err := spend(ctx, budget); err != nil {
				return err
			}
			if _, err := PairBracket(ctx, append(append([]*ParticipantState{}, next...), set...), forbidden, nil); err != nil {
				if errors.Is(err, ErrNoCompletePairing) {
					return nil
				}
				return err
			}
			rest := without(remaining, append(append([]*ParticipantState{}, next...), set...))
			complete, err := completable(ctx, rest, forbidden, budget)
			if err != nil {
				return err
			}
			if complete {
				found = true
				return errC6Found
			}
			return nil
		})
		if errors.Is(err, errC6Found) {
			return k, nil
		}
		if err != nil {
			return 0, err
		}
		if found {
			return k, nil
		}
	}
	return 0, ErrNoCompletePairing
}

var errC6Found = errors.New("C6 upfloater count found")

func completable(ctx context.Context, participants []*ParticipantState, forbidden map[[2]string]bool, budget *int) (bool, error) {
	if len(participants)%2 != 0 {
		return false, nil
	}
	if err := spend(ctx, budget); err != nil {
		return false, err
	}
	_, err := PairBracket(ctx, participants, forbidden, nil)
	if err == nil || errors.Is(err, ErrNoCompletePairing) {
		return err == nil, nil
	}
	return false, err
}

func without(participants, removed []*ParticipantState) []*ParticipantState {
	removedIDs := make(map[string]bool, len(removed))
	for _, participant := range removed {
		removedIDs[participant.ID] = true
	}
	result := make([]*ParticipantState, 0, len(participants)-len(removed))
	for _, participant := range participants {
		if !removedIDs[participant.ID] {
			result = append(result, participant)
		}
	}
	return result
}

func spend(ctx context.Context, budget *int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	(*budget)--
	if *budget < 0 {
		return ErrPairingSearchLimit
	}
	return nil
}
