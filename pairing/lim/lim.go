// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lim

import (
	"context"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// Pair implements chesspairing.Pairer for the Lim Swiss system.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := swisslib.ValidatePreAssignedPAB(state); err != nil {
		return nil, fmt.Errorf("lim: %w", err)
	}
	originalState := state

	// Honour pre-assigned byes for the upcoming round.
	state, preAssignedByes := swisslib.FilterPreAssignedByes(state)

	result := &chesspairing.PairingResult{}

	// Build player states.
	players, err := swisslib.BuildPlayerStates(state)
	if err != nil {
		return nil, err
	}
	if len(players) <= 1 {
		// 0 or 1 player: just assign bye if needed.
		if len(players) == 1 {
			if players[0].PABIneligible.Any() {
				return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "lim", Err: swisslib.ErrNoPABCandidate}
			}
			result.Byes = append(result.Byes, chesspairing.ByeEntry{
				PlayerID: players[0].ID,
				Type:     chesspairing.ByePAB,
			})
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		if err := chesspairing.ValidatePairing(originalState, result); err != nil {
			if pairingErr, ok := err.(*chesspairing.PairingError); ok {
				pairingErr.System = "lim"
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

	// Assign PAB if odd number of players.
	playerPtrs := make([]*swisslib.PlayerState, len(players))
	for i := range players {
		playerPtrs[i] = &players[i]
	}

	if swisslib.NeedsBye(len(playerPtrs)) {
		byeSelector := LimByeSelector{}
		byePlayer := byeSelector.SelectBye(playerPtrs)
		if byePlayer == nil {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingNoPABCandidate, System: "lim", Err: swisslib.ErrNoPABCandidate}
		}
		result.Byes = append(result.Byes, chesspairing.ByeEntry{
			PlayerID: byePlayer.ID,
			Type:     chesspairing.ByePAB,
		})
		// Remove bye player from pairing pool.
		playerPtrs = removePtrs(playerPtrs, byePlayer)
	}

	// Build score groups.
	scoreGroups := swisslib.BuildScoreGroups(derefPlayers(playerPtrs))

	// Article 2.2 uses completed rounds, not a possibly staged upcoming round.
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	medianScore := float64(historyEnd) / 2.0
	lastRound := p.opts.TotalRounds > 0 && state.CurrentRound == p.opts.TotalRounds
	compatible := func(a, b *swisslib.PlayerState) bool {
		return IsCompatibleLastRound(a, b, forbidden, lastRound)
	}

	// Determine processing order (Art. 2.2).
	aboveMedian, belowMedian, medianGroup := splitByMedian(scoreGroups, medianScore)

	floatedIDs := make(map[string]bool)
	previousFloaters := previousRoundFloaters(state)
	isMaxi := *p.opts.MaxiTournament

	// Process higher scoregroups
	higherPairs := make([][2]*swisslib.PlayerState, 0)
	var pendingDownFloaters []FloaterEntry

	for _, sg := range aboveMedian {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("lim: %w", err)
		}
		groupPlayers := make([]*swisslib.PlayerState, len(sg.Players))
		copy(groupPlayers, sg.Players)

		var earlierDownFloaters []FloaterEntry
		if len(pendingDownFloaters) > 0 {
			floaterPairs, remaining, targets := PairFloaters(pendingDownFloaters, groupPlayers, true, isMaxi, forbidden)
			higherPairs = append(higherPairs, floaterPairs...)
			groupPlayers = targets
			// A floater that cannot be paired directly is now part of this
			// scoregroup (Art. 3.2.1); let Art. 2.4 exchanges pair it here.
			for _, floater := range remaining {
				groupPlayers = append(groupPlayers, floater.Player)
			}
			earlierDownFloaters = remaining
			pendingDownFloaters = nil
		}

		pairs, unpaired := exchangeMatchWithFloaterSelector(groupPlayers, true, compatible, func(pairable []*swisslib.PlayerState) *swisslib.PlayerState {
			return selectFloaterWithEarlier(pairable, adjacentScoreGroup(scoreGroups, sg.Score, FloatDown), forbidden, floatedIDs, isMaxi, true, previousFloaters, true, earlierDownFloaters, compatible)
		})
		pairs = ColourExchange(pairs, forbidden, isMaxi)
		higherPairs = append(higherPairs, pairs...)

		for _, u := range unpaired {
			floatedIDs[u.ID] = true
			pendingDownFloaters = append(pendingDownFloaters, FloaterEntry{Player: u, Direction: FloatDown, SourceScore: sg.Score})
		}
	}

	lowerPairs := make([][2]*swisslib.PlayerState, 0)

	// Process lower scoregroups
	var pendingUpFloaters []FloaterEntry

	for i := len(belowMedian) - 1; i >= 0; i-- {
		sg := belowMedian[i]
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("lim: %w", err)
		}
		groupPlayers := make([]*swisslib.PlayerState, len(sg.Players))
		copy(groupPlayers, sg.Players)

		var earlierUpFloaters []FloaterEntry
		if len(pendingUpFloaters) > 0 {
			floaterPairs, remaining, targets := PairFloaters(pendingUpFloaters, groupPlayers, false, isMaxi, forbidden)
			lowerPairs = append(lowerPairs, floaterPairs...)
			groupPlayers = targets
			// A floater that cannot be paired directly is now part of this
			// scoregroup (Art. 3.2.1); let Art. 2.4 exchanges pair it here.
			for _, floater := range remaining {
				groupPlayers = append(groupPlayers, floater.Player)
			}
			earlierUpFloaters = remaining
			pendingUpFloaters = nil
		}

		pairs, unpaired := exchangeMatchWithFloaterSelector(groupPlayers, false, compatible, func(pairable []*swisslib.PlayerState) *swisslib.PlayerState {
			return selectFloaterWithEarlier(pairable, adjacentScoreGroup(scoreGroups, sg.Score, FloatUp), forbidden, floatedIDs, isMaxi, false, previousFloaters, true, earlierUpFloaters, compatible)
		})
		pairs = ColourExchange(pairs, forbidden, isMaxi)
		lowerPairs = append(lowerPairs, pairs...)

		for _, u := range unpaired {
			floatedIDs[u.ID] = true
			pendingUpFloaters = append(pendingUpFloaters, FloaterEntry{Player: u, Direction: FloatUp, SourceScore: sg.Score})
		}
	}

	// Median group + Cracking loop
	var medianPairs [][2]*swisslib.PlayerState
	medianPlayers := make([]*swisslib.PlayerState, 0)
	// Interpretation of Arts. 2.2 and 3.6.3: if no player has the score
	// "equal to half the number of rounds", its median position contains the
	// arriving floaters. It is paired last, downward: "first pair the DF,
	// then the UF and finally the remaining players".
	if medianGroup != nil {
		medianPlayers = append(medianPlayers, medianGroup.Players...)
	}

	for {
		currentPendingDown := append([]FloaterEntry(nil), pendingDownFloaters...)
		currentPendingUp := append([]FloaterEntry(nil), pendingUpFloaters...)
		groupPlayers := append([]*swisslib.PlayerState(nil), medianPlayers...)
		medianPlayerCount := len(groupPlayers) + len(currentPendingDown) + len(currentPendingUp)

		var currentMedianPairs [][2]*swisslib.PlayerState

		// Pair down-floaters first
		if len(currentPendingDown) > 0 {
			floaterPairs, remaining, targets := PairFloaters(currentPendingDown, groupPlayers, true, isMaxi, forbidden)
			currentMedianPairs = append(currentMedianPairs, floaterPairs...)
			currentPendingDown = remaining
			groupPlayers = targets
		}

		// Then up-floaters
		if len(currentPendingUp) > 0 {
			floaterPairs, remaining, targets := PairFloaters(currentPendingUp, groupPlayers, false, isMaxi, forbidden)
			currentMedianPairs = append(currentMedianPairs, floaterPairs...)
			currentPendingUp = remaining
			groupPlayers = targets
		}

		// If both up and down floaters are still remaining, they cannot be paired.
		// We'll leave them to ExchangeMatch to fail.
		// Wait, they are floaters. The remaining players must be paired.
		// Put remaining floaters back into groupPlayers for ExchangeMatch.
		for _, f := range currentPendingDown {
			groupPlayers = append(groupPlayers, f.Player)
		}
		for _, f := range currentPendingUp {
			groupPlayers = append(groupPlayers, f.Player)
		}

		// Art. 2.3.4 applies here too: a median-group down-floater uses the
		// same Arts. 3.2-3.5, 3.9 and 3.10 selector as every other group.
		pairs, unpaired := exchangeMatchWithFloaterSelector(groupPlayers, true, compatible, func(pairable []*swisslib.PlayerState) *swisslib.PlayerState {
			return selectFloaterWithEarlier(pairable, adjacentScoreGroup(scoreGroups, medianScore, FloatDown), forbidden, floatedIDs, isMaxi, true, previousFloaters, true, nil, compatible)
		})
		pairs = ColourExchange(pairs, forbidden, isMaxi)
		currentMedianPairs = append(currentMedianPairs, pairs...)

		if (medianGroup != nil && (len(unpaired) == 0 || len(pairs) > 0)) ||
			(medianGroup == nil && len(unpaired) == 0 && len(currentMedianPairs)*2 == medianPlayerCount) {
			medianPairs = currentMedianPairs
			break
		}

		// Cracking needed (Art 2.6)
		// 2.6.1 if down-floaters > up-floaters -> crack lower. Else crack higher.
		if len(pendingDownFloaters) > len(pendingUpFloaters) && len(lowerPairs) > 0 {
			pair := lowerPairs[len(lowerPairs)-1]
			lowerPairs = lowerPairs[:len(lowerPairs)-1]
			pendingUpFloaters = append(pendingUpFloaters, FloaterEntry{Player: pair[0], Direction: FloatUp, SourceScore: pair[0].Score})
			pendingUpFloaters = append(pendingUpFloaters, FloaterEntry{Player: pair[1], Direction: FloatUp, SourceScore: pair[1].Score})
		} else if len(higherPairs) > 0 {
			pair := higherPairs[len(higherPairs)-1]
			higherPairs = higherPairs[:len(higherPairs)-1]
			pendingDownFloaters = append(pendingDownFloaters, FloaterEntry{Player: pair[0], Direction: FloatDown, SourceScore: pair[0].Score})
			pendingDownFloaters = append(pendingDownFloaters, FloaterEntry{Player: pair[1], Direction: FloatDown, SourceScore: pair[1].Score})
		} else if len(lowerPairs) > 0 {
			pair := lowerPairs[len(lowerPairs)-1]
			lowerPairs = lowerPairs[:len(lowerPairs)-1]
			pendingUpFloaters = append(pendingUpFloaters, FloaterEntry{Player: pair[0], Direction: FloatUp, SourceScore: pair[0].Score})
			pendingUpFloaters = append(pendingUpFloaters, FloaterEntry{Player: pair[1], Direction: FloatUp, SourceScore: pair[1].Score})
		} else {
			return nil, &chesspairing.PairingError{Kind: chesspairing.PairingImpossible, System: "lim", Err: fmt.Errorf("no pairing satisfies Lim scoregroup rules (cracking failed)")}
		}
	}

	var allPairs [][2]*swisslib.PlayerState
	allPairs = append(allPairs, higherPairs...)
	allPairs = append(allPairs, lowerPairs...)
	allPairs = append(allPairs, medianPairs...)

	// Assign colours and build final result.
	for boardNum, pair := range allPairs {
		isAboveMedian := pair[0].PairingScore > medianScore ||
			pair[0].PairingScore == medianScore
		wID, bID := AllocateColor(pair[0], pair[1], state.CurrentRound, isAboveMedian, topSeedColorPtr(p.opts.TopSeedColor))
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   boardNum + 1,
			WhiteID: wID,
			BlackID: bID,
		})
	}

	// Re-number boards: sort by max score desc, then min TPN asc.
	sortBoards(result.Pairings, playerPtrs)

	if len(preAssignedByes) > 0 {
		result.Byes = append(preAssignedByes, result.Byes...)
	}

	if err := chesspairing.ValidatePairing(originalState, result); err != nil {
		if pairingErr, ok := err.(*chesspairing.PairingError); ok {
			pairingErr.System = "lim"
			if pairingErr.Kind == chesspairing.PairingIncomplete {
				pairingErr.Partial = result
			}
			if (pairingErr.Kind == chesspairing.PairingIncomplete && len(result.Pairings) == 0) || pairingAllocatedByes(result) > 1 {
				pairingErr.Kind = chesspairing.PairingImpossible
				pairingErr.Missing = nil
				pairingErr.Partial = nil
			}
		}
		return nil, err
	}
	return result, nil
}

func pairingAllocatedByes(result *chesspairing.PairingResult) int {
	count := 0
	for _, bye := range result.Byes {
		if bye.Type == chesspairing.ByePAB {
			count++
		}
	}
	return count
}

// splitByMedian divides score groups into above-median, below-median, and
// the median group itself.
func splitByMedian(groups []swisslib.ScoreGroup, medianScore float64) (above, below []swisslib.ScoreGroup, median *swisslib.ScoreGroup) {
	for i, sg := range groups {
		if floatEqual(sg.Score, medianScore) {
			// This is the median group.
			mg := groups[i]
			median = &mg
		} else if sg.Score > medianScore {
			above = append(above, sg)
		} else {
			below = append(below, sg)
		}
	}
	return
}

// floatEqual compares two float64 values for near-equality with a tolerance
// of 0.001 (sufficient for half-point chess scores).
func floatEqual(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.001
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

// adjacentScoreGroup returns the actual neighbouring scoregroup in the
// direction of a floater, as required by Art. 3.3-3.5.
func adjacentScoreGroup(groups []swisslib.ScoreGroup, score float64, direction FloatDirection) []*swisslib.PlayerState {
	for i, group := range groups {
		if !floatEqual(group.Score, score) {
			continue
		}
		if direction == FloatDown && i+1 < len(groups) {
			return groups[i+1].Players
		}
		if direction == FloatUp && i > 0 {
			return groups[i-1].Players
		}
		return nil
	}
	return nil
}

// removePtrs removes a specific player from the pointer slice.
func removePtrs(players []*swisslib.PlayerState, remove *swisslib.PlayerState) []*swisslib.PlayerState {
	result := make([]*swisslib.PlayerState, 0, len(players)-1)
	for _, p := range players {
		if p.ID != remove.ID {
			result = append(result, p)
		}
	}
	return result
}

// derefPlayers converts pointer slice to value slice for BuildScoreGroups.
func derefPlayers(ptrs []*swisslib.PlayerState) []swisslib.PlayerState {
	result := make([]swisslib.PlayerState, len(ptrs))
	for i, p := range ptrs {
		result[i] = *p
	}
	return result
}

// topSeedColorPtr converts the string option to a Color pointer.
func topSeedColorPtr(opt *string) *swisslib.Color {
	if opt == nil {
		return nil
	}
	switch *opt {
	case "white":
		c := swisslib.ColorWhite
		return &c
	case "black":
		c := swisslib.ColorBlack
		return &c
	default:
		return nil
	}
}

// sortBoards sorts pairings under C.04.2 article 3.6.
func sortBoards(pairings []chesspairing.GamePairing, players []*swisslib.PlayerState) {
	playerMap := make(map[string]*swisslib.PlayerState, len(players))
	for _, p := range players {
		playerMap[p.ID] = p
	}

	sort.SliceStable(pairings, func(i, j int) bool {
		pi1, pi2 := playerMap[pairings[i].WhiteID], playerMap[pairings[i].BlackID]
		pj1, pj2 := playerMap[pairings[j].WhiteID], playerMap[pairings[j].BlackID]

		return swisslib.BoardBefore(swisslib.BoardPlayer(pi1), swisslib.BoardPlayer(pi2),
			swisslib.BoardPlayer(pj1), swisslib.BoardPlayer(pj2))
	})

	// Renumber boards.
	for i := range pairings {
		pairings[i].Board = i + 1
	}
}
