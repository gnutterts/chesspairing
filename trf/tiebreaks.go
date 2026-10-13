// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"fmt"
	"strings"
)

// fideTiebreaks maps the C.07:2026 acronym (and applicable modifier) to a
// registered library ID. IDs without an entry are intentionally not emitted
// by FromTournamentState, because TRF has no corresponding FIDE code.
var fideTiebreaks = []struct {
	code    string
	id      string
	article string
}{
	{"DE", "direct-encounter", "6"},
	{"EDE", "direct-encounter", "13.3"},
	{"BH", "buchholz", "8.1"},
	{"BH/C1", "buchholz-cut1", "8.1, 14.1"},
	{"BH/C2", "buchholz-cut2", "8.1, 14.2"},
	{"BH/M1", "buchholz-median", "8.1, 14.3"},
	{"BH/M2", "buchholz-median2", "8.1, 14.4"},
	{"BH:MP", "buchholz-mp", "8.1, 13"},
	{"BH:MP/C1", "buchholz-mp-cut1", "8.1, 13, 14.1"},
	{"FB", "fore-buchholz", "8.3"},
	{"AOB", "avg-opponent-buchholz", "8.2"},
	{"ARO", "aro", "10.1"},
	{"ARO/C1", "aro-cut1", "10.1, 14.1"},
	{"APRO", "avg-opponent-tpr", "10.4"},
	{"APPO", "avg-opponent-ptp", "10.5"},
	{"TPR", "performance-rating", "10.2"},
	{"PTP", "performance-points", "10.3"},
	{"RTNG", "player-rating", "10.6"},
	{"SB", "sonneborn-berger", "9.1"},
	{"SB/C1", "sonneborn-berger-cut1", "9.1, 14.1"},
	{"KS", "koya", "9.2"},
	{"BPG", "black-games", "7.3"},
	{"BWG", "black-wins", "7.4"},
	{"WIN", "win", "7.1"},
	{"WON", "wins", "7.2"},
	{"PS", "progressive", "7.5"},
	{"PS/C1", "progressive-cut1", "7.5, 14.1"},
	{"REP", "rounds-played", "7.6"},
	{"STD", "standard-points", "7.7"},
	{"TPN", "pairing-number", "7.8"},
	{"MPvGP", "mpvgp", "13.1"},
	{"EMMSB", "emmsb", "13.2.1"},
	{"EMMSB/C1", "emmsb-cut1", "13.2.1, 14.1"},
	{"EMGSB", "emgsb", "13.2.2"},
	{"EMGSB/C1", "emgsb-cut1", "13.2.2, 14.1"},
	{"EGMSB", "egmsb", "13.2.3"},
	{"EGMSB/C1", "egmsb-cut1", "13.2.3, 14.1"},
	{"EGGSB", "eggsb", "13.2.4"},
	{"EGGSB/C1", "eggsb-cut1", "13.2.4, 14.1"},
	{"BC", "board-count", "12.1"},
	{"TBR", "top-board-results", "12.2"},
	{"BBE", "bottom-board-elimination", "12.3"},
}

// C.07 13.3.2 combines extended direct encounter with board criteria.
var fideEDEComposites = []struct {
	code string
	ids  []string
}{
	{"EDEBT", []string{"direct-encounter", "board-count", "top-board-results"}},
	{"EDEBB", []string{"direct-encounter", "board-count", "bottom-board-elimination"}},
	{"EDET", []string{"direct-encounter", "top-board-results"}},
	{"EDEB", []string{"direct-encounter", "bottom-board-elimination"}},
}

func trfTiebreakID(code string) ([]string, bool) {
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, "/P", "")
	for _, entry := range fideEDEComposites {
		if strings.EqualFold(entry.code, code) {
			return entry.ids, true
		}
	}
	for _, entry := range fideTiebreaks {
		if strings.EqualFold(entry.code, code) {
			return []string{entry.id}, true
		}
	}
	return nil, false
}

func trfTiebreakCode(id string) (string, bool) {
	for _, entry := range fideTiebreaks {
		if entry.id == id {
			return entry.code, true
		}
	}
	return "", false
}

func parseTiebreakers(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	ids := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		expanded, ok := trfTiebreakID(part)
		if !ok {
			return nil, fmt.Errorf("trf: unsupported FIDE tie-break code %q", part)
		}
		ids = append(ids, expanded...)
	}
	return ids, nil
}

func formatTiebreakers(ids []string) (string, bool) {
	codes := make([]string, 0, len(ids))
	for i := 0; i < len(ids); i++ {
		matched := false
		for _, entry := range fideEDEComposites {
			if len(ids)-i >= len(entry.ids) && strings.Join(ids[i:i+len(entry.ids)], ",") == strings.Join(entry.ids, ",") {
				codes = append(codes, entry.code)
				i += len(entry.ids) - 1
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		code, ok := trfTiebreakCode(ids[i])
		if !ok {
			return "", false
		}
		codes = append(codes, code)
	}
	return strings.Join(codes, ","), true
}
