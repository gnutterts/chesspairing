// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/lexswiss"
)

func TestAllocateColor_Article43Rules(t *testing.T) {
	white := lexswiss.ColorWhite
	black := lexswiss.ColorBlack

	tests := []struct {
		name          string
		first, second lexswiss.ParticipantState
		initialColor  *string
		whiteID       string
		blackID       string
	}{
		{
			name:         "4.3.1 initial colour for an even TPN HRP",
			first:        lexswiss.ParticipantState{ID: "hrp", PairingNumber: 4, Score: 2},
			second:       lexswiss.ParticipantState{ID: "opponent", PairingNumber: 1, Score: 1},
			initialColor: ptr("black"),
			whiteID:      "hrp",
			blackID:      "opponent",
		},
		{
			name: "4.3.2 fewer Whites",
			first: lexswiss.ParticipantState{
				ID:            "hrp",
				PairingNumber: 1,
				ColorHistory:  []lexswiss.Color{white, white},
			},
			second: lexswiss.ParticipantState{
				ID:            "opponent",
				PairingNumber: 2,
				ColorHistory:  []lexswiss.Color{black},
			},
			whiteID: "opponent",
			blackID: "hrp",
		},
		{
			name: "4.3.3 most recent opposite colours",
			first: lexswiss.ParticipantState{
				ID:            "hrp",
				PairingNumber: 1,
				ColorHistory:  []lexswiss.Color{white, black},
			},
			second: lexswiss.ParticipantState{
				ID:            "opponent",
				PairingNumber: 2,
				ColorHistory:  []lexswiss.Color{black, white},
			},
			whiteID: "hrp",
			blackID: "opponent",
		},
		{
			name: "4.3.4 HRP last colour",
			first: lexswiss.ParticipantState{
				ID:            "hrp",
				PairingNumber: 1,
				ColorHistory:  []lexswiss.Color{white, white},
			},
			second: lexswiss.ParticipantState{
				ID:            "opponent",
				PairingNumber: 2,
				ColorHistory:  []lexswiss.Color{white, white},
			},
			whiteID: "opponent",
			blackID: "hrp",
		},
		{
			name:    "4.3.5 opponent last colour",
			first:   lexswiss.ParticipantState{ID: "hrp", PairingNumber: 1, Score: 2},
			second:  lexswiss.ParticipantState{ID: "opponent", PairingNumber: 2, Score: 1, ColorHistory: []lexswiss.Color{black, black}},
			whiteID: "opponent",
			blackID: "hrp",
		},
		{
			name: "4.3.3 unplayed rounds are left out of the histories",
			// The late joiner has played [W, W] after two unplayed rounds. Left out,
			// both histories end in W, W, so 4.3.3 finds no opposite colours and
			// 4.3.4 gives the HRP, whose last colour is W, Black.
			first: lexswiss.ParticipantState{
				ID:            "opponent", // lower TPN -> HRP
				PairingNumber: 1,
				ColorHistory:  []lexswiss.Color{black, black, white, white},
			},
			second: lexswiss.ParticipantState{
				ID:            "joiner",
				PairingNumber: 2,
				ColorHistory:  []lexswiss.Color{lexswiss.ColorNone, lexswiss.ColorNone, white, white},
			},
			whiteID: "joiner",
			blackID: "opponent",
		},
		{
			name: "4.3.3 compares played matches, not rounds",
			// C.04.2 3.4: BWBuW counts as uBWBW. HRP [W, B, -] and opponent [B, -, W]
			// left out give [W, B] and [B, W]: the latest matches had HRP Black and
			// the opponent White, so the HRP now gets White.
			first: lexswiss.ParticipantState{
				ID:            "hrp",
				PairingNumber: 1,
				Score:         2,
				ColorHistory:  []lexswiss.Color{white, black, lexswiss.ColorNone},
			},
			second: lexswiss.ParticipantState{
				ID:            "opponent",
				PairingNumber: 2,
				Score:         1,
				ColorHistory:  []lexswiss.Color{black, lexswiss.ColorNone, white},
			},
			whiteID: "hrp",
			blackID: "opponent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whiteID, blackID := AllocateColor(&tt.first, &tt.second, tt.initialColor)
			if whiteID != tt.whiteID || blackID != tt.blackID {
				t.Errorf("AllocateColor() = White %s, Black %s; want White %s, Black %s", whiteID, blackID, tt.whiteID, tt.blackID)
			}
		})
	}
}

func ptr(value string) *string {
	return &value
}
