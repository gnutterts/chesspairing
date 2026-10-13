// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/team"
)

func TestToTournamentStateTeamMatches(t *testing.T) {
	win, draw, loss := 1.0, 0.5, 0.0
	doc := &Document{
		TournamentType: "Team Swiss",
		Players: []PlayerLine{
			{StartNumber: 1, Name: "A1"},
			{StartNumber: 2, Name: "A2"},
			{StartNumber: 3, Name: "B1"},
			{StartNumber: 4, Name: "B2"},
		},
		Teams: []TeamLine{
			{TeamNumber: 1, TeamName: "Alpha", Members: []int{1, 2}},
			{TeamNumber: 2, TeamName: "Beta", Members: []int{3, 4}},
		},
		DetailedTeamResults: []DetailedTeamResult{{
			TeamNumber: 1,
			Rounds:     []DetailedTeamRound{{Opponent: 2, Color: "w", Results: "1="}},
		}},
		ScoringSystem: &ScoringPoints{W: &win, D: &draw, L: &loss},
		PrimaryScore:  "game",
	}

	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if state.ScoringConfig.System != "team" || state.ScoringConfig.Options["primaryScore"] != "game" {
		t.Errorf("ScoringConfig = %+v, want team scoring with game primary score", state.ScoringConfig)
	}
	if state.Players[0].TeamID != "1" || state.Players[3].TeamID != "2" {
		t.Errorf("team links = %q, %q", state.Players[0].TeamID, state.Players[3].TeamID)
	}
	if len(state.Rounds) != 1 || len(state.Rounds[0].Matches) != 1 {
		t.Fatalf("Matches = %+v, want one match", state.Rounds)
	}
	match := state.Rounds[0].Matches[0]
	if match.HomeID != "1" || match.AwayID != "2" || len(match.Boards) != 2 {
		t.Errorf("Match = %+v, want 1-2 with two boards", match)
	}
	if match.Result != nil {
		t.Errorf("Match.Result = %+v, want nil", match.Result)
	}

	written, _ := FromTournamentState(state)
	if len(written.Teams) != 2 || len(written.DetailedTeamResults) != 2 || len(written.SimpleTeamResults) != 2 {
		t.Fatalf("written team records = teams:%d 801:%d 802:%d, want 2 each", len(written.Teams), len(written.DetailedTeamResults), len(written.SimpleTeamResults))
	}
	roundTrip, err := written.ToTournamentState()
	if err != nil {
		t.Fatalf("round-trip ToTournamentState: %v", err)
	}
	got := roundTrip.Rounds[0].Matches
	if len(got) != 1 || got[0].Result != nil {
		t.Errorf("round-trip Matches = %+v, want one match with nil Result", got)
	}
}

func TestTeamMatchesStateWriteReadRoundTrip(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "1", TeamID: "1"},
			{ID: "2", TeamID: "1"},
			{ID: "3", TeamID: "2"},
			{ID: "4", TeamID: "2"},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Matches: []chesspairing.MatchData{
					{
						HomeID: "1",
						AwayID: "2",
						Boards: []chesspairing.GameData{
							{WhiteID: "1", BlackID: "3", Result: chesspairing.ResultWhiteWins},
							{WhiteID: "4", BlackID: "2", Result: chesspairing.ResultBlackWins},
						},
					},
				},
			},
			{
				Number: 2,
				Matches: []chesspairing.MatchData{
					{
						HomeID: "2",
						AwayID: "1",
						Boards: []chesspairing.GameData{
							{WhiteID: "3", BlackID: "1", Result: chesspairing.ResultDraw},
							{WhiteID: "2", BlackID: "4", Result: chesspairing.ResultBlackWins},
						},
					},
				},
			},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingTeam},
	}
	doc, _, err := FromTournamentStateWithError(state)
	if err != nil {
		t.Fatalf("FromTournamentStateWithError: %v", err)
	}
	var data bytes.Buffer
	if err := Write(&data, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	read, err := Read(&data)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	roundTrip, err := read.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	if len(roundTrip.Rounds) != 2 || len(roundTrip.Rounds[0].Matches) != 1 || len(roundTrip.Rounds[1].Matches) != 1 {
		t.Fatalf("round-trip matches = %+v, want one match in both rounds", roundTrip.Rounds)
	}
	if got := roundTrip.Rounds[0].Matches[0].Result; got != nil {
		t.Errorf("round 1 result = %+v, want nil", got)
	}
	if got := roundTrip.Rounds[1].Matches[0].Result; got != nil {
		t.Errorf("round 2 result = %+v, want nil", got)
	}
}

func TestTeamMatchesBlackFirstBoardRoundTrip(t *testing.T) {
	for _, detailed := range []bool{true, false} {
		t.Run(map[bool]string{true: "801_and_802", false: "802_only"}[detailed], func(t *testing.T) {
			doc := &Document{
				TournamentType:   "Team Swiss",
				TeamInitialColor: "BWBW",
				Players: []PlayerLine{
					{StartNumber: 1, Name: "A1"},
					{StartNumber: 2, Name: "A2"},
					{StartNumber: 3, Name: "A3"},
					{StartNumber: 4, Name: "A4"},
					{StartNumber: 5, Name: "B1"},
					{StartNumber: 6, Name: "B2"},
					{StartNumber: 7, Name: "B3"},
					{StartNumber: 8, Name: "B4"},
				},
				Teams: []TeamLine{
					{TeamNumber: 1, TeamName: "Alpha", Members: []int{1, 2, 3, 4}},
					{TeamNumber: 2, TeamName: "Beta", Members: []int{5, 6, 7, 8}},
				},
				SimpleTeamResults: []SimpleTeamResult{
					{TeamNumber: 1, Rounds: []SimpleTeamRound{{Opponent: 2, Color: "w", GamePoints: 3}}},
					{TeamNumber: 2, Rounds: []SimpleTeamRound{{Opponent: 1, Color: "b", GamePoints: 1}}},
				},
			}
			if detailed {
				doc.DetailedTeamResults = []DetailedTeamResult{
					{TeamNumber: 1, Rounds: []DetailedTeamRound{{Opponent: 2, Color: "w", Results: "1110", BoardOrder: "1234"}}},
					{TeamNumber: 2, Rounds: []DetailedTeamRound{{Opponent: 1, Color: "b", Results: "0001", BoardOrder: "1234"}}},
				}
			}
			for number := 9; number <= 16; number++ {
				doc.Players = append(doc.Players, PlayerLine{StartNumber: number, Name: "Reserve"})
			}
			doc.Teams = append(doc.Teams,
				TeamLine{TeamNumber: 3, TeamName: "Gamma", Members: []int{9, 10, 11, 12}},
				TeamLine{TeamNumber: 4, TeamName: "Delta", Members: []int{13, 14, 15, 16}},
			)
			var input bytes.Buffer
			if err := Write(&input, doc); err != nil {
				t.Fatalf("Write input: %v", err)
			}
			parsed, err := Read(&input)
			if err != nil {
				t.Fatalf("Read input: %v", err)
			}
			state, err := parsed.ToTournamentState()
			if err != nil {
				t.Fatalf("ToTournamentState: %v", err)
			}
			if len(state.Rounds) != 1 || len(state.Rounds[0].Matches) != 1 {
				t.Fatalf("matches = %+v", state.Rounds)
			}
			match := state.Rounds[0].Matches[0]
			if match.HomeID != "2" || match.AwayID != "1" {
				t.Errorf("match sides = %s/%s, want 2/1", match.HomeID, match.AwayID)
			}
			if detailed {
				if len(match.Boards) != 4 || match.Boards[0].WhiteID != "5" || match.Boards[0].BlackID != "1" {
					t.Errorf("board 1 = %+v, want White=5 Black=1", match.Boards)
				}
			} else if match.Result == nil || match.Result.HomeGame != 1 || match.Result.AwayGame != 3 {
				t.Errorf("game points = %+v, want home=1 away=3", match.Result)
			}
			if _, err := team.New(team.Options{}).Pair(context.Background(), state); err != nil {
				t.Errorf("team Pair: %v", err)
			}
			written, _ := FromTournamentState(state)
			var output bytes.Buffer
			if err := Write(&output, written); err != nil {
				t.Fatalf("Write state: %v", err)
			}
			again, err := Read(&output)
			if err != nil {
				t.Fatalf("Read state: %v", err)
			}
			if again.TeamInitialColor != parsed.TeamInitialColor {
				t.Errorf("352 = %q, want %q", again.TeamInitialColor, parsed.TeamInitialColor)
			}
			for i := range parsed.SimpleTeamResults {
				if !reflect.DeepEqual(again.SimpleTeamResults[i].Rounds, parsed.SimpleTeamResults[i].Rounds) {
					t.Errorf("802 team %d rounds = %+v, want %+v", i+1, again.SimpleTeamResults[i].Rounds, parsed.SimpleTeamResults[i].Rounds)
				}
			}
			if detailed {
				for i := range parsed.DetailedTeamResults {
					if !reflect.DeepEqual(again.DetailedTeamResults[i].Rounds, parsed.DetailedTeamResults[i].Rounds) {
						t.Errorf("801 team %d rounds = %+v, want %+v", i+1, again.DetailedTeamResults[i].Rounds, parsed.DetailedTeamResults[i].Rounds)
					}
				}
			}
		})
	}
}

func TestTeamMatchByeRoundTrip(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "1", TeamID: "1"},
			{ID: "2", TeamID: "2"},
			{ID: "3", TeamID: "3"},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingTeam},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{{
				HomeID: "1",
				AwayID: "2",
				Result: &chesspairing.TeamMatchResult{HomeGame: 1, AwayGame: 0},
			}},
			TeamByes: []chesspairing.ByeEntry{{PlayerID: "3", Type: chesspairing.ByePAB}},
		}},
	}
	doc, _, err := FromTournamentStateWithError(state)
	if err != nil {
		t.Fatalf("FromTournamentStateWithError: %v", err)
	}
	var data bytes.Buffer
	if err := Write(&data, doc); err != nil {
		t.Fatalf("Write: %v", err)
	}
	read, err := Read(&data)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	roundTrip, err := read.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	found := false
	for _, bye := range roundTrip.Rounds[0].TeamByes {
		if bye == (chesspairing.ByeEntry{PlayerID: "3", Type: chesspairing.ByePAB}) {
			found = true
		}
	}
	if !found {
		t.Errorf("round-trip team byes = %+v, want PAB for team 3", roundTrip.Rounds[0].TeamByes)
	}
}

func TestFromTournamentStateWithErrorRejectsNonNumericTeamID(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "player", TeamID: "alpha"}},
	}
	if _, _, err := FromTournamentStateWithError(state); err == nil {
		t.Error("FromTournamentStateWithError succeeded for non-numeric team ID")
	}
}

func TestToTournamentStateSimpleTeamUsesRoundTotals(t *testing.T) {
	doc := &Document{
		TournamentType: "Team Swiss",
		SimpleTeamResults: []SimpleTeamResult{
			{TeamNumber: 1, GamePoints: 6.5, Rounds: []SimpleTeamRound{{Opponent: 2, Color: "w", GamePoints: 2.5}}},
			{TeamNumber: 2, GamePoints: 1.5, Rounds: []SimpleTeamRound{{Opponent: 1, Color: "b", GamePoints: 1.5}}},
		},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	match := state.Rounds[0].Matches[0]
	if match.Result == nil || match.Result.HomeGame != 2.5 || match.Result.AwayGame != 1.5 {
		t.Errorf("Match.Result = %+v, want round totals 2.5-1.5", match.Result)
	}
}

func TestToTournamentStateSimpleTeamTotals(t *testing.T) {
	doc := &Document{
		TournamentType: "Team Swiss",
		SimpleTeamResults: []SimpleTeamResult{
			{TeamNumber: 1, Rounds: []SimpleTeamRound{{Opponent: 2, Color: "w", GamePoints: 2.5}}},
			{TeamNumber: 2, Rounds: []SimpleTeamRound{{Opponent: 1, Color: "b", GamePoints: 1.5}}},
		},
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState: %v", err)
	}
	match := state.Rounds[0].Matches[0]
	if match.Result == nil || match.Result.HomeGame != 2.5 || match.Result.AwayGame != 1.5 {
		t.Errorf("Match.Result = %+v, want 2.5-1.5", match.Result)
	}
}
