// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

// BoardParticipant holds the pairing score and tournament pairing number used
// for board ordering under C.04.2 article 3.6.
type BoardParticipant struct {
	Score float64
	TPN   int
}

// BoardBefore reports whether pair (a, b) precedes pair (c, d). The higher
// ranked participant in each pair has the higher score, then the smaller TPN.
func BoardBefore(a, b, c, d BoardParticipant) bool {
	if b.Score > a.Score || (b.Score == a.Score && b.TPN < a.TPN) {
		a, b = b, a
	}
	if d.Score > c.Score || (d.Score == c.Score && d.TPN < c.TPN) {
		c, d = d, c
	}
	if a.Score != c.Score {
		return a.Score > c.Score
	}
	if a.Score+b.Score != c.Score+d.Score {
		return a.Score+b.Score > c.Score+d.Score
	}
	return a.TPN < c.TPN
}

// BoardPlayer returns the board-ordering values for an individual player.
func BoardPlayer(p *PlayerState) BoardParticipant {
	return BoardParticipant{Score: p.PairingScore, TPN: EffectivePairingNumber(p)}
}
