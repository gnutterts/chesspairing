// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package main

import cp "github.com/gnutterts/chesspairing"

// stageUnpairedRound handles a TRF whose last round column is the round that is
// about to be paired: bbpPairings, JaVaFo and the tournament programs that
// write TRF16 put the requested byes and announced absences of that round in
// its column, next to the finished rounds. Read as played rounds they would
// make the tournament one round longer and give those players their points;
// here they become the pre-assigned byes of the round to pair.
//
// The last round counts as the unpaired one when nobody has a game in it, it
// holds byes, none of which is a pairing-allocated bye, and the file does not
// state pre-assigned byes of its own (Section 240 records).
func stageUnpairedRound(state *cp.TournamentState) {
	n := len(state.Rounds)
	if n == 0 || state.CurrentRound != n+1 || len(state.PreAssignedByes) > 0 {
		return
	}
	last := state.Rounds[n-1]
	if len(last.Games) > 0 || len(last.Byes) == 0 {
		return
	}
	for _, bye := range last.Byes {
		if bye.Type == cp.ByePAB {
			return
		}
	}
	state.PreAssignedByes = append(state.PreAssignedByes, last.Byes...)
	state.Rounds = state.Rounds[:n-1]
	state.CurrentRound = n
}
