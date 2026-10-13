// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	scoringTeam "github.com/gnutterts/chesspairing/scoring/team"
)

func TestTeamRecords2026Conversion(t *testing.T) {
	doc := &Document{
		CodedTournamentType: "FIDE_TEAM",
		TeamInitialColor:    "WWBB",
		NewTeams: []NewTeamLine{
			{TeamNumber: 1, TeamName: "Alpha", Members: []int{1, 2}},
			{TeamNumber: 2, TeamName: "Beta", Members: []int{3, 4}},
		},
		Players:             []PlayerLine{{StartNumber: 1}, {StartNumber: 2}, {StartNumber: 3}, {StartNumber: 4}},
		DetailedTeamResults: []DetailedTeamResult{{TeamNumber: 1, Rounds: []DetailedTeamRound{{Opponent: 2, Color: "w", Results: "11"}}}},
		TeamRoundData:       []TeamRoundEntry{{Round: 1, Team1: 1, Team2: 2, Boards: []int{2, 1}}},
		OldAbsentForfeits:   []OldAbsentForfeit{{ResultType: "+-", Round: 2, WhiteTeam: 1, BlackTeam: 2}},
		TeamPABs:            []TeamPABRecord{{MatchPoints: 1, GamePoints: 1, RoundTeams: []int{0, 0, 1}}},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].TeamID != "1" || state.Players[3].TeamID != "2" {
		t.Fatalf("310 team membership: %+v", state.Players)
	}
	if len(state.Rounds) != 3 || len(state.Rounds[0].Matches) != 1 || len(state.Rounds[1].Matches) != 1 || len(state.Rounds[2].TeamByes) != 1 {
		t.Fatalf("team rounds: %+v", state.Rounds)
	}
	boards := state.Rounds[0].Matches[0].Boards
	if boards[0].WhiteID != "2" || boards[1].WhiteID != "1" {
		t.Fatalf("300 order and 352 colours: %+v", boards)
	}
	if state.Rounds[1].Matches[0].Result.HomeGame != 4 {
		t.Fatalf("330 forfeit: %+v", state.Rounds[1].Matches[0])
	}
	out, _ := FromTournamentState(state)
	if !reflect.DeepEqual(out.OldAbsentForfeits, doc.OldAbsentForfeits) {
		t.Fatalf("330 not preserved: %+v", out.OldAbsentForfeits)
	}
	if len(out.NewTeams) != 2 || len(out.Teams) != 2 || len(out.TeamRoundData) == 0 || len(out.TeamPABs) != 1 {
		t.Fatalf("emitted team records: %+v", out)
	}
	var buf bytes.Buffer
	if err := Write(&buf, out); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(&buf); err != nil {
		t.Fatal(err)
	}
}

func TestRoundTripTeamTRF(t *testing.T) {
	data, err := os.ReadFile("testdata/trf2026-team.trf")
	if err != nil {
		t.Fatal(err)
	}
	original, err := Read(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	state, err := original.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	converted, _ := FromTournamentState(state)
	var buf bytes.Buffer
	if err := Write(&buf, converted); err != nil {
		t.Fatal(err)
	}
	again, err := Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []struct {
		name      string
		want, got any
	}{
		{"162", original.ScoringSystem, again.ScoringSystem},
		{"192", original.CodedTournamentType, again.CodedTournamentType},
		{"202", original.TieBreakDef, again.TieBreakDef},
		{"222", original.EncodedTimeControl, again.EncodedTimeControl},
		{"299", original.AbnormalAssignments, again.AbnormalAssignments},
		{"300", original.TeamRoundData, again.TeamRoundData},
		{"310", original.NewTeams, again.NewTeams},
		{"320", original.TeamPABs, again.TeamPABs},
		{"330", original.OldAbsentForfeits, again.OldAbsentForfeits},
		{"352", original.TeamInitialColor, again.TeamInitialColor},
		{"362", original.TeamScoringPoints, again.TeamScoringPoints},
	} {
		if !reflect.DeepEqual(field.want, field.got) {
			t.Errorf("%s: want %+v, got %+v", field.name, field.want, field.got)
		}
	}
}

func TestForfeitWithoutColorSequenceAndBusyTeam(t *testing.T) {
	doc := &Document{
		CodedTournamentType: "FIDE_TEAM",
		Players:             []PlayerLine{{StartNumber: 1}, {StartNumber: 2}, {StartNumber: 3}, {StartNumber: 4}},
		NewTeams: []NewTeamLine{
			{TeamNumber: 1, Members: []int{1, 2}}, {TeamNumber: 2, Members: []int{3, 4}},
		},
		OldAbsentForfeits: []OldAbsentForfeit{
			{ResultType: "+-", Round: 1, WhiteTeam: 1, BlackTeam: 2},
			{ResultType: "--", Round: 2, WhiteTeam: 1, BlackTeam: 2},
		},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Rounds[0].Matches[0].Result.HomeGame != 2 || state.Rounds[1].Matches[0].Result != nil {
		t.Fatalf("forfeit points: %+v", state.Rounds)
	}
	if got := teamBoardResults(state.Rounds[1].Matches[0], teamPlayerTeams(doc.Teams)); got != "--" {
		t.Errorf("double-forfeit boards = %q, want --", got)
	}
	state.CurrentRound = 2
	scores, err := scoringTeam.NewFromMap(state.ScoringConfig.Options).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	points := make(map[string]float64)
	for _, score := range scores {
		points[score.PlayerID] = score.Team.Match
	}
	if points["1"] != 2 || points["2"] != 0 {
		t.Errorf("forfeit match points = %v, want 2-0", points)
	}
	out, _ := FromTournamentState(state)
	if !reflect.DeepEqual(out.OldAbsentForfeits, doc.OldAbsentForfeits) {
		t.Fatalf("forfeits: %+v", out.OldAbsentForfeits)
	}
	doc.DetailedTeamResults = []DetailedTeamResult{{TeamNumber: 1, Rounds: []DetailedTeamRound{{ByeType: "FFFF"}}}}
	state, err = doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Rounds[0].Matches) != 0 {
		t.Fatalf("forfeit duplicated a bye: %+v", state.Rounds[0])
	}
}

func TestForfeitUsesConfiguredBoardPoints(t *testing.T) {
	doc := &Document{OldAbsentForfeits: []OldAbsentForfeit{{ResultType: "+-", Round: 1, WhiteTeam: 1, BlackTeam: 2}}}
	teams := []TeamLine{{TeamNumber: 1, Members: []int{1, 2}}, {TeamNumber: 2, Members: []int{3, 4}}}
	rounds := make([]chesspairing.RoundData, 1)
	buildTeamMatches(doc, teams, rounds, map[string]any{"pointForfeitWin": 1.5})
	if got := rounds[0].Matches[0].Result.HomeGame; got != 3 {
		t.Errorf("forfeit game points = %v, want 3", got)
	}
}

func TestAbnormalByePreservedWithoutScoringOverride(t *testing.T) {
	raw := "299 F 1.0 1.0\n"
	doc, err := Read(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	doc.CodedTournamentType = "FIDE_TEAM"
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"pointWin", "pointMatchWin"} {
		if _, ok := state.ScoringConfig.Options[key]; ok {
			t.Errorf("299 F set %s: %+v", key, state.ScoringConfig.Options)
		}
	}
	out, _ := FromTournamentState(state)
	var buf bytes.Buffer
	if err := Write(&buf, out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), strings.ReplaceAll(raw, "\n", "\r\n")) {
		t.Errorf("299 F not preserved: %q", buf.String())
	}
}

func TestTeamRostersConflictAndAbnormalAssignments(t *testing.T) {
	doc := &Document{CodedTournamentType: "FIDE_TEAM", NewTeams: []NewTeamLine{{TeamNumber: 1, Members: []int{1}}}, Teams: []TeamLine{{TeamNumber: 1, Members: []int{2}}}}
	if _, err := doc.ToTournamentState(); err == nil || !strings.Contains(err.Error(), "team 1") {
		t.Fatalf("conflicting rosters: %v", err)
	}
	doc = &Document{CodedTournamentType: "FIDE_TEAM", TeamScoringPoints: &TeamScoringPoints{TW: chesspairing.Float64Ptr(2)}, AbnormalAssignments: []AbnormalAssignmentRecord{{Type: "W", MatchPoints: 3}}}
	if _, err := doc.ToTournamentState(); err == nil || !strings.Contains(err.Error(), "299 W") {
		t.Fatalf("conflicting points: %v", err)
	}
	raw := "299 +    2.0      2.5\n"
	parsed, err := Read(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Write(&buf, parsed); err != nil || buf.String() != strings.ReplaceAll(raw, "\n", "\r\n") {
		t.Fatalf("299 raw round trip: %q, %v", buf.String(), err)
	}
}
