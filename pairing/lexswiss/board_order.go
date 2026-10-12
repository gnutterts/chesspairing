// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package lexswiss

import "github.com/gnutterts/chesspairing/pairing/swisslib"

// BoardPlayer returns the primary pairing score and current TPN of a participant.
func BoardPlayer(p *ParticipantState) swisslib.BoardParticipant {
	return swisslib.BoardParticipant{Score: p.Score, TPN: p.TPN}
}
