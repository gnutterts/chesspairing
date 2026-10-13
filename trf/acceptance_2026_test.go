// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"errors"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestAcceleration250BlankGamePoints(t *testing.T) {
	doc, err := Read(strings.NewReader("250 01.0      004 006 0001 0090\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Accelerations26) != 1 {
		t.Fatalf("accelerations = %v", doc.Accelerations26)
	}
	got := doc.Accelerations26[0]
	if got.MatchPoints != 1 || got.GamePoints != 0 || got.FirstRound != 4 || got.LastRound != 6 || got.FirstPlayer != 1 || got.LastPlayer != 90 {
		t.Errorf("250 fields = %+v", got)
	}
}

func TestAcceleration250NoData(t *testing.T) {
	for _, line := range []string{"250\n", "250   \n"} {
		_, err := Read(strings.NewReader(line))
		var pe *ParseError
		if !errors.As(err, &pe) || pe.Code != "250" {
			t.Errorf("%q: error = %v, want 250 ParseError", line, err)
		}
	}
}

func TestTeamRound300BlankBoards(t *testing.T) {
	doc, err := Read(strings.NewReader("300 011 002 006 0002           0027\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.TeamRoundData) != 1 {
		t.Fatalf("team rounds = %v", doc.TeamRoundData)
	}
	got := doc.TeamRoundData[0]
	if got.Round != 11 || got.Team1 != 2 || got.Team2 != 6 || len(got.Boards) != 4 || got.Boards[0] != 2 || got.Boards[1] != 0 || got.Boards[2] != 0 || got.Boards[3] != 27 {
		t.Errorf("300 fields = %+v, want boards [2 0 0 27]", got)
	}
}

func TestTeamRound300TrailingPadding(t *testing.T) {
	doc, err := Read(strings.NewReader("300 011 002 006 0002 0001 0003 0004  \n"))
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.TeamRoundData[0].Boards; len(got) != 4 || got[0] != 2 || got[1] != 1 || got[2] != 3 || got[3] != 4 {
		t.Errorf("boards = %v, want [2 1 3 4]", got)
	}
}

func TestScoring162Columns(t *testing.T) {
	w, d, l := 1.0, 0.5, 0.0
	var out strings.Builder
	if err := Write(&out, &Document{ScoringSystem: &ScoringPoints{W: &w, D: &d, L: &l}}); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "162  W 1.0    D 0.5    L 0.0\r\n"; got != want {
		t.Errorf("162 bytes = %q, want %q", got, want)
	}
}

func TestScoring162FullWidthRoundTrip(t *testing.T) {
	w, d, l := 10.0, 0.5, 0.0
	var out strings.Builder
	if err := Write(&out, &Document{ScoringSystem: &ScoringPoints{W: &w, D: &d, L: &l}}); err != nil {
		t.Fatal(err)
	}
	doc, err := Read(strings.NewReader(out.String()))
	if err != nil {
		t.Fatal(err)
	}
	if doc.ScoringSystem.W == nil || *doc.ScoringSystem.W != w || doc.ScoringSystem.D == nil || *doc.ScoringSystem.D != d || doc.ScoringSystem.L == nil || *doc.ScoringSystem.L != l {
		t.Errorf("scoring after round trip = %+v", doc.ScoringSystem)
	}
}

func TestScoring162SparseFullWidth(t *testing.T) {
	value := 11.0
	var out strings.Builder
	if err := Write(&out, &Document{ScoringSystem: &ScoringPoints{D: &value}}); err != nil {
		t.Fatal(err)
	}
	doc, err := Read(strings.NewReader(out.String()))
	if err != nil {
		t.Fatal(err)
	}
	if doc.ScoringSystem.D == nil || *doc.ScoringSystem.D != value {
		t.Errorf("sparse scoring after round trip = %+v", doc.ScoringSystem)
	}
}

func TestTeamScoring362FullWidthRoundTrip(t *testing.T) {
	w, d, l := 10.0, 1.0, 0.0
	var out strings.Builder
	if err := Write(&out, &Document{TeamScoringPoints: &TeamScoringPoints{TW: &w, TD: &d, TL: &l}}); err != nil {
		t.Fatal(err)
	}
	doc, err := Read(strings.NewReader(out.String()))
	if err != nil {
		t.Fatal(err)
	}
	if doc.TeamScoringPoints.TW == nil || *doc.TeamScoringPoints.TW != w || doc.TeamScoringPoints.TD == nil || *doc.TeamScoringPoints.TD != d || doc.TeamScoringPoints.TL == nil || *doc.TeamScoringPoints.TL != l {
		t.Errorf("team scoring after round trip = %+v", doc.TeamScoringPoints)
	}
}

func TestScoringColumnsAcceptSpacedInput(t *testing.T) {
	doc, err := Read(strings.NewReader("162 W 10.0 D 0.5 L 0.0\n362 TW 10.0 TD 1.0 TL 0.0\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.ScoringSystem.W == nil || *doc.ScoringSystem.W != 10 || doc.ScoringSystem.D == nil || *doc.ScoringSystem.D != 0.5 || doc.TeamScoringPoints.TW == nil || *doc.TeamScoringPoints.TW != 10 || doc.TeamScoringPoints.TD == nil || *doc.TeamScoringPoints.TD != 1 {
		t.Errorf("space-separated scoring: 162=%+v 362=%+v", doc.ScoringSystem, doc.TeamScoringPoints)
	}
}

func TestStartingRank172Ratings(t *testing.T) {
	for _, tc := range []struct {
		method string
		want   [3]int
	}{
		{"FIDE", [3]int{2100, 0, 2300}},
		{"NRO", [3]int{2000, 2200, 0}},
		{"FIDON", [3]int{2100, 2200, 2300}},
		{"NIDOF", [3]int{2000, 2200, 2300}},
		{"HBFN", [3]int{2100, 2200, 2300}},
		{"LBFN", [3]int{2000, 2200, 2300}},
		{"OTHER", [3]int{2100, 0, 2300}},
		{"", [3]int{2100, 0, 2300}},
	} {
		t.Run(tc.method, func(t *testing.T) {
			doc := &Document{
				StartingRankMethod: "NED " + tc.method,
				Players: []PlayerLine{
					{StartNumber: 9, Rating: 2100}, {StartNumber: 3}, {StartNumber: 7, Rating: 2300},
				},
				NRSRecords: []NRSRecord{
					{Federation: "NED", StartNumber: 9, NationalRating: 2000},
					{Federation: "NED", StartNumber: 3, NationalRating: 2200},
					{Federation: "BEL", StartNumber: 7, NationalRating: 1900},
				},
			}
			if tc.method == "" {
				doc.StartingRankMethod = ""
			}
			state, err := doc.ToTournamentState()
			if err != nil {
				t.Fatal(err)
			}
			for i, player := range state.Players {
				if player.Rating != tc.want[i] || player.PairingNumber != doc.Players[i].StartNumber {
					t.Errorf("player %d: rating=%d pairingNumber=%d, want %d/%d", i, player.Rating, player.PairingNumber, tc.want[i], doc.Players[i].StartNumber)
				}
			}
		})
	}
}

func TestStartingRank172MissingNationalRatingsFallback(t *testing.T) {
	for _, method := range []string{"FIDON", "NIDOF", "HBFN", "LBFN"} {
		t.Run(method, func(t *testing.T) {
			doc := &Document{
				StartingRankMethod: "NED " + method,
				Players:            []PlayerLine{{StartNumber: 1, Rating: 2100}},
			}
			state, err := doc.ToTournamentState()
			if err != nil {
				t.Fatal(err)
			}
			if got := state.Players[0].Rating; got != 2100 {
				t.Errorf("rating = %d, want 2100", got)
			}
		})
	}
}

func TestScoring162DanglingSymbol(t *testing.T) {
	_, err := Read(strings.NewReader("162  W 1.0 D\n"))
	var pe *ParseError
	if !errors.As(err, &pe) || pe.Code != "162" {
		t.Errorf("error = %v, want 162 ParseError", err)
	}
}

func TestScoringPointsOverflowError(t *testing.T) {
	value := 100.0
	for _, tc := range []struct {
		name string
		doc  *Document
		want string
	}{
		{"162", &Document{ScoringSystem: &ScoringPoints{W: &value}}, "162 W points 100.0"},
		{"362", &Document{TeamScoringPoints: &TeamScoringPoints{TW: &value}}, "362 TW points 100.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			err := Write(&out, tc.doc)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("write error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestStartingRank172FederationMatching(t *testing.T) {
	doc := &Document{
		StartingRankMethod: "ned NRO",
		Players:            []PlayerLine{{StartNumber: 1, Rating: 2100}},
		NRSRecords:         []NRSRecord{{Federation: "NED", StartNumber: 1, NationalRating: 1950}},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	if state.Players[0].Rating != 1950 {
		t.Errorf("rating = %d, want 1950", state.Players[0].Rating)
	}
	doc.StartingRankMethod = "BEL NRO"
	_, err = doc.ToTournamentState()
	if err == nil || !strings.Contains(err.Error(), "BEL") {
		t.Errorf("missing NRS error = %v, want federation diagnostic", err)
	}
}

func TestPairingController182(t *testing.T) {
	doc, err := Read(strings.NewReader("182 Tournament controller 1.2\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.PairingController != "Tournament controller 1.2" || len(doc.Other) != 0 {
		t.Errorf("182 parsed = %+v", doc)
	}
	var out strings.Builder
	if err := Write(&out, doc); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "182 Tournament controller 1.2\r\n" {
		t.Errorf("182 output = %q", got)
	}
	generated, _ := FromTournamentState(&chesspairing.TournamentState{})
	if !strings.HasPrefix(generated.PairingController, "chesspairing") {
		t.Errorf("default controller = %q", generated.PairingController)
	}
}
