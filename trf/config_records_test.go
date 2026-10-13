// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"os"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestCodedTournamentTypeConfig(t *testing.T) {
	doc := &Document{CodedTournamentType: "FIDE_TEAM_TYPEB_GP_MP"}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if state.PairingConfig.System != chesspairing.PairingTeam {
		t.Errorf("System = %q, want team", state.PairingConfig.System)
	}
	if state.PairingConfig.Options["colorPreferenceType"] != "B" || state.PairingConfig.Options["primaryScore"] != "game" {
		t.Errorf("team options = %#v, want Type B/game", state.PairingConfig.Options)
	}
	doc.ColorPreferenceType, doc.PrimaryScore = "A", "match"
	state, err = doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState with XX options: %v", err)
	}
	if state.PairingConfig.Options["colorPreferenceType"] != "A" || state.PairingConfig.Options["primaryScore"] != "match" {
		t.Errorf("explicit team options = %#v, want Type A/match", state.PairingConfig.Options)
	}
	baku, err := (&Document{CodedTournamentType: "FIDE_TEAM_BAKU"}).ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState Baku: %v", err)
	}
	if baku.PairingConfig.Options["acceleration"] != "baku" {
		t.Errorf("Baku acceleration = %v", baku.PairingConfig.Options["acceleration"])
	}
}

func TestTieBreak212AndConfigurationRoundTrip(t *testing.T) {
	doc, err := Read(strings.NewReader("212 PTS,EDET/P,BH:MP/C1/P\n222 5400+30\n162 W 3.0    D 1.0    L 0.0\n362 TW 3     TD 1     TL 0\n"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if got := strings.Join(state.ScoringConfig.Tiebreakers, ","); got != "direct-encounter,top-board-results,buchholz-mp-cut1" {
		t.Errorf("Tiebreakers = %q", got)
	}
	if state.Info.EncodedTimeControl != "5400+30" || state.ScoringConfig.Options["pointMatchWin"] != 3.0 {
		t.Errorf("configuration was not converted: %+v %#v", state.Info, state.ScoringConfig.Options)
	}
	written, _ := FromTournamentState(state)
	if written.TieBreakRecord != "212" || written.TieBreakDef != "EDET/P,BH:MP/C1/P" {
		t.Errorf("tie-break round trip = %q %q", written.TieBreakRecord, written.TieBreakDef)
	}
	if written.ScoringSystem == nil || written.TeamScoringPoints == nil {
		t.Errorf("scoring records were not written: %+v %+v", written.ScoringSystem, written.TeamScoringPoints)
	}
}

func TestRecord212FileRoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team-212.trf")
	if err != nil {
		t.Fatalf("read testdata/trf2026-team-212.trf: %v", err)
	}
	doc, err := Read(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if doc.TieBreakRecord != "212" || doc.TieBreakDef != "BH,SB" {
		t.Fatalf("212 record = %q/%q, want 212/BH,SB", doc.TieBreakRecord, doc.TieBreakDef)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if got := strings.Join(state.ScoringConfig.Tiebreakers, ","); got != "buchholz,sonneborn-berger" {
		t.Errorf("Tiebreakers = %q, want buchholz,sonneborn-berger", got)
	}
	written, _ := FromTournamentState(state)
	var out strings.Builder
	if err := Write(&out, written); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(out.String(), "212 PTS,BH") {
		t.Errorf("written TRF lacks 212 PTS,BH record")
	}
	doc2, err := Read(strings.NewReader(out.String()))
	if err != nil {
		t.Fatalf("re-Read: %v", err)
	}
	if doc2.TieBreakRecord != "212" {
		t.Errorf("TieBreakRecord = %q, want 212", doc2.TieBreakRecord)
	}
}

func TestTeamESTBCut1TieBreakCodes(t *testing.T) {
	ids, err := parseTiebreakers("EMGSB/C1/P")
	if err != nil {
		t.Fatalf("parseTiebreakers(EMGSB/C1/P): %v", err)
	}
	if strings.Join(ids, ",") != "emgsb-cut1" {
		t.Errorf("parseTiebreakers(EMGSB/C1/P) = %v, want [emgsb-cut1]", ids)
	}
	doc := &Document{CodedTournamentType: "FIDE_TEAM_TYPEA_MP_GP", TieBreakDef: "EMGSB/C1/P,BH:MP/C1/P,MPvGP"}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if got := strings.Join(state.ScoringConfig.Tiebreakers, ","); got != "emgsb-cut1,buchholz-mp-cut1,mpvgp" {
		t.Errorf("Tiebreakers = %q, want emgsb-cut1,buchholz-mp-cut1,mpvgp", got)
	}
}

func TestTeamTypeCodeRoundTrip(t *testing.T) {
	tests := []struct {
		code             string
		wantPrimary      string
		wantColorPref    string
		wantAcceleration bool
	}{
		{"FIDE_TEAM_GP_MP", "game", "", false},
		{"FIDE_TEAM_TYPEA_MP", "match", "A", false},
		{"FIDE_TEAM_TYPEB_GP_MP_BAKU", "game", "B", true},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			file := strings.Join([]string{
				"012 T", "022 C", "032 NED", "042 2026/01/01", "052 2026/01/02",
				"062 0", "072 0", "142 1", "152 W", "192 " + tt.code,
			}, "\n") + "\n"
			doc, err := Read(strings.NewReader(file))
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			state, err := doc.ToTournamentState()
			if err != nil {
				t.Fatalf("ToTournamentState: %v", err)
			}
			colorPref, _ := state.PairingConfig.Options["colorPreferenceType"].(string)
			if state.PairingConfig.Options["primaryScore"] != tt.wantPrimary ||
				colorPref != tt.wantColorPref ||
				(state.PairingConfig.Options["acceleration"] == "baku") != tt.wantAcceleration {
				t.Fatalf("state options = %#v", state.PairingConfig.Options)
			}
			written, _, err := FromTournamentStateWithError(state)
			if err != nil {
				t.Fatalf("FromTournamentStateWithError: %v", err)
			}
			if written.CodedTournamentType != tt.code {
				t.Errorf("CodedTournamentType round trip = %q, want %q", written.CodedTournamentType, tt.code)
			}
			var out strings.Builder
			if err := Write(&out, written); err != nil {
				t.Fatalf("Write: %v", err)
			}
			if !strings.Contains(out.String(), "192 "+tt.code) {
				t.Errorf("written TRF lacks 192 %s", tt.code)
			}
		})
	}
}

func TestTeamRoundRobinCodeUsesFallback(t *testing.T) {
	for _, code := range []string{"FIDE_TEAM_ROUNDROBIN", "FIDE_TEAM_DOUBLEROUNDROBIN", "BERGER_TEAM_ROUNDROBIN_G2", "FIDE_TEAM_SCHEVENINGEN"} {
		doc := &Document{TournamentType: "Round Robin", CodedTournamentType: code}
		state, err := doc.ToTournamentState()
		if err != nil || state.PairingConfig.System != chesspairing.PairingRoundRobin {
			t.Fatalf("%s: state=%+v, err=%v", code, state, err)
		}
		out, _ := FromTournamentState(state)
		if out.CodedTournamentType != code {
			t.Errorf("%s became %s", code, out.CodedTournamentType)
		}
	}
}

func TestGeneratedTeamAndTieBreakCodes(t *testing.T) {
	state := &chesspairing.TournamentState{
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingTeam, Options: map[string]any{"primaryScore": "game"}},
		ScoringConfig: chesspairing.ScoringConfig{Tiebreakers: []string{"buchholz", "direct-encounter"}},
	}
	out, _ := FromTournamentState(state)
	if out.CodedTournamentType != "FIDE_TEAM_GP_MP" || out.TieBreakDef != "BH,DE" {
		t.Fatalf("generated codes: %s, %s", out.CodedTournamentType, out.TieBreakDef)
	}
}

func TestTeamScoringSymbolsAnyOrder(t *testing.T) {
	doc, err := Read(strings.NewReader("362 TD 1     TW 3     TL 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.TeamScoringPoints.TW == nil || *doc.TeamScoringPoints.TW != 3 {
		t.Fatalf("reordered 362: %+v", doc.TeamScoringPoints)
	}
	if _, err := Read(strings.NewReader("362 TD 1     TD 2     TL 0\n")); err == nil {
		t.Fatal("duplicate symbol accepted")
	}
}

func TestUnknownFideTieBreakFails(t *testing.T) {
	for _, code := range []string{"UNKNOWN", "BH:MP/C2", "EDEUNKNOWN"} {
		_, err := (&Document{TieBreakDef: code}).ToTournamentState()
		if err == nil || !strings.Contains(err.Error(), code) {
			t.Errorf("error for %q = %v, want unsupported code error", code, err)
		}
	}
}

func TestCompositeEDETiebreakRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		code string
		ids  string
	}{
		{"EDET", "direct-encounter,top-board-results"},
		{"EDEB", "direct-encounter,bottom-board-elimination"},
		{"EDEBT", "direct-encounter,board-count,top-board-results"},
		{"EDEBB", "direct-encounter,board-count,bottom-board-elimination"},
		{"MPvGP", "mpvgp"},
	} {
		t.Run(tc.code, func(t *testing.T) {
			state, err := (&Document{TieBreakDef: tc.code}).ToTournamentState()
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(state.ScoringConfig.Tiebreakers, ","); got != tc.ids {
				t.Fatalf("tie-break IDs = %q, want %q", got, tc.ids)
			}
			out, _ := FromTournamentState(state)
			if out.TieBreakDef != tc.code {
				t.Errorf("tie-break code = %q, want %q", out.TieBreakDef, tc.code)
			}
		})
	}
}

func TestTeamColorSequenceConfiguresBoards(t *testing.T) {
	doc := &Document{
		CodedTournamentType: "FIDE_TEAM",
		TeamInitialColor:    "WWBB",
		Players: []PlayerLine{
			{StartNumber: 1},
			{StartNumber: 2},
			{StartNumber: 3},
			{StartNumber: 4},
			{StartNumber: 5},
			{StartNumber: 6},
			{StartNumber: 7},
			{StartNumber: 8},
		},
		Teams:               []TeamLine{{TeamNumber: 1, Members: []int{1, 2, 3, 4}}, {TeamNumber: 2, Members: []int{5, 6, 7, 8}}},
		DetailedTeamResults: []DetailedTeamResult{{TeamNumber: 1, Rounds: []DetailedTeamRound{{Opponent: 2, Color: "w", Results: "===="}}}},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	match := state.Rounds[0].Matches[0]
	if match.Boards[0].WhiteID != "1" || match.Boards[1].WhiteID != "2" || match.Boards[2].WhiteID != "7" {
		t.Errorf("board colours = %+v, want WWBB", match.Boards)
	}
	if state.ScoringConfig.Options["boardCount"] != 4 {
		t.Errorf("scoring boardCount = %v, want 4", state.ScoringConfig.Options["boardCount"])
	}
}
