// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"errors"
	"math"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

// C8Cost returns the number of teams in a proposed pair whose colour preference,
// if any, is not fulfilled (Art. 2.3.5 [C8]).
//
// In a pair where both teams have a colour preference for the same colour,
// exactly one team cannot get its preferred colour because one team must play
// White and the other Black.
//
// A pair can never leave two teams unfulfilled: one of the two always receives
// its preferred colour (via Art. 4.3 colour allocation). When only one team has
// a preference, or the two teams have opposite preferences, both preferences
// are fulfilled (0 unfulfilled).
//
// Therefore, the number of unfulfilled teams in the pair is 1 if both have a
// preference for the same colour, and 0 otherwise.
func C8Cost(a, b *lexswiss.ParticipantState, prefType ColorPrefType, isLastRound bool) int {
	if prefType == ColorPrefTypeNone {
		return 0
	}
	prefA := ComputeColorPreference(a, prefType, isLastRound)
	prefB := ComputeColorPreference(b, prefType, isLastRound)
	if (prefA.IsWhite() && prefB.IsWhite()) || (prefA.IsBlack() && prefB.IsBlack()) {
		return 1
	}
	return 0
}

// C9Cost returns the number of teams in a proposed pair whose strong colour
// preference, if any, is not fulfilled (Art. 2.3.6 [C9], Type B only).
//
// If prefType is not ColorPrefTypeB, C9 does not apply and returns 0.
// Under Type B, if both teams have a strong preference for the same colour,
// exactly one of them cannot be fulfilled (1 unfulfilled). If only one team has
// a strong preference, Art. 4.3.4 grants it (0 unfulfilled). If both have strong
// preferences for opposite colours, Art. 4.3.3 grants both (0 unfulfilled).
func C9Cost(a, b *lexswiss.ParticipantState, prefType ColorPrefType, isLastRound bool) int {
	if prefType != ColorPrefTypeB {
		return 0
	}
	prefA := ComputeColorPreference(a, prefType, isLastRound)
	prefB := ComputeColorPreference(b, prefType, isLastRound)
	if prefA.IsStrong() && prefB.IsStrong() && ((prefA.IsWhite() && prefB.IsWhite()) || (prefA.IsBlack() && prefB.IsBlack())) {
		return 1
	}
	return 0
}

// C10Cost counts the number of upfloaters in the pair whose opponent was a floater
// in the previous round (Art. 2.3.7 [C10]).
func C10Cost(a, b *lexswiss.ParticipantState, upfloaterIDs map[string]bool) int {
	cost := 0
	if upfloaterIDs[a.ID] && b.WasFloater {
		cost++
	}
	if upfloaterIDs[b.ID] && a.WasFloater {
		cost++
	}
	return cost
}

// BracketPairCost builds the combined quality cost function for a bracket of size
// bracketSize, minimising criteria C8, C9, and C10 in descending priority
// (Art. 2.3, 3.6.4).
//
// It combines the criteria into a single integer cost:
//
//	cost = c8 * W^2 + c9 * W + c10
//
// with weight W larger than the maximum possible sum of the lower terms in the
// bracket, ensuring that C8 strictly dominates C9 and C9 strictly dominates C10.
func BracketPairCost(bracketSize int, upfloaters []*lexswiss.ParticipantState, prefType ColorPrefType, isLastTwoRounds, isLastRound bool) (lexswiss.PairCostFunc, error) {
	W := bracketSize + 1
	if W < 2 {
		W = 2
	}

	// Check for overflow: ensure W*W and maximum total bracket cost do not overflow int.
	if W > math.MaxInt/W {
		return nil, errors.New("cost weight overflows integer")
	}
	W2 := W * W
	if W2 > math.MaxInt-W-2 {
		return nil, errors.New("pair cost calculation overflows integer")
	}
	maxPairs := bracketSize / 2
	if maxPairs > 0 && (W2+W+2) > math.MaxInt/maxPairs {
		return nil, errors.New("total bracket cost overflows integer")
	}

	upfloaterIDs := make(map[string]bool, len(upfloaters))
	for _, u := range upfloaters {
		upfloaterIDs[u.ID] = true
	}

	return func(a, b *lexswiss.ParticipantState) int {
		c8 := C8Cost(a, b, prefType, isLastRound)
		c9 := C9Cost(a, b, prefType, isLastRound)
		c10 := 0
		if !isLastTwoRounds {
			c10 = C10Cost(a, b, upfloaterIDs)
		}
		return c8*W2 + c9*W + c10
	}, nil
}

// bracketCostFunc returns a bracket cost generator for Article 3.6.4 minimisation
// in top-down pairing.
func bracketCostFunc(numTeams int, prefType ColorPrefType, isLastTwoRounds, isLastRound bool) (func([]*lexswiss.ParticipantState) lexswiss.PairCostFunc, error) {
	if prefType == ColorPrefTypeNone && isLastTwoRounds {
		return nil, nil
	}
	// Validate overflow upfront for tournament size.
	if _, err := BracketPairCost(numTeams, nil, prefType, isLastTwoRounds, isLastRound); err != nil {
		return nil, err
	}
	return func(upfloaters []*lexswiss.ParticipantState) lexswiss.PairCostFunc {
		cost, _ := BracketPairCost(numTeams, upfloaters, prefType, isLastTwoRounds, isLastRound)
		return cost
	}, nil
}
