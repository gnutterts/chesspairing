// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestScore_Round1PAB(t *testing.T) {
	// 3 teams, team 'c' gets a PAB in round 1, no matches played yet.
	// Board order/membership gives boardCount = 4
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", TeamID: "a"},
			{ID: "p2", TeamID: "a"},
			{ID: "p3", TeamID: "a"},
			{ID: "p4", TeamID: "a"},
			{ID: "p5", TeamID: "b"},
			{ID: "p6", TeamID: "b"},
			{ID: "p7", TeamID: "b"},
			{ID: "p8", TeamID: "b"},
			{ID: "p9", TeamID: "c"},
			{ID: "p10", TeamID: "c"},
			{ID: "p11", TeamID: "c"},
			{ID: "p12", TeamID: "c"},
		},
		Rounds: []chesspairing.RoundData{{
			Number:   1,
			TeamByes: []chesspairing.ByeEntry{{PlayerID: "c", Type: chesspairing.ByePAB}},
		}},
	}
	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	for _, sc := range scores {
		if sc.PlayerID == "c" {
			if sc.Team.Match != 1 || sc.Team.Game != 2.0 {
				t.Errorf("expected MP 1, GP 2.0, got MP %v GP %v", sc.Team.Match, sc.Team.Game)
			}
		}
	}
}

func TestScore_802Only(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a"}, {ID: "b"}, {ID: "c"},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{
				{HomeID: "a", AwayID: "b", Result: &chesspairing.TeamMatchResult{HomeGame: 3, AwayGame: 1}},
			},
			TeamByes: []chesspairing.ByeEntry{{PlayerID: "c", Type: chesspairing.ByePAB}},
		}},
	}
	bc := 4
	scorer := New(Options{BoardCount: &bc})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	for _, sc := range scores {
		if sc.PlayerID == "c" {
			if sc.Team.Match != 1 || sc.Team.Game != 2.0 {
				t.Errorf("expected MP 1, GP 2.0, got MP %v GP %v", sc.Team.Match, sc.Team.Game)
			}
		}
	}
}

func TestScoreMatchAndGamePoints(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{
				{
					HomeID: "a",
					AwayID: "b",
					Boards: []chesspairing.GameData{
						{Result: chesspairing.ResultWhiteWins},
						{Result: chesspairing.ResultDraw},
					},
				},
				{
					HomeID: "c",
					AwayID: "d",
					Boards: []chesspairing.GameData{
						{Result: chesspairing.ResultForfeitWhiteWins},
						{Result: chesspairing.ResultBlackWins},
					},
				},
			},
			TeamByes: []chesspairing.ByeEntry{{PlayerID: "d", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
	}

	scores, err := New(Options{}).Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	got := make(map[string]chesspairing.PlayerScore)
	for _, score := range scores {
		got[score.PlayerID] = score
	}
	for id, want := range map[string]chesspairing.TeamPoints{
		"a": {Match: 2, Game: 1.5},
		"b": {Match: 0, Game: 0.5},
		"c": {Match: 1, Game: 1},
		"d": {Match: 2, Game: 2},
	} {
		if got[id].Team == nil || *got[id].Team != want {
			t.Errorf("%s Team = %+v, want %+v", id, got[id].Team, want)
		}
	}
	if got["a"].Score != 2 {
		t.Errorf("a Score = %v, want primary match points 2", got["a"].Score)
	}
}

func TestTeamByeUsesPointsForEveryBoard(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		Rounds: []chesspairing.RoundData{{
			Matches: []chesspairing.MatchData{{
				HomeID: "a", AwayID: "b",
				Boards: []chesspairing.GameData{{Result: chesspairing.ResultDraw}, {Result: chesspairing.ResultDraw}, {Result: chesspairing.ResultDraw}, {Result: chesspairing.ResultDraw}},
			}},
			TeamByes: []chesspairing.ByeEntry{{PlayerID: "c", Type: chesspairing.ByePAB}},
		}},
		CurrentRound: 2,
	}
	scores, err := New(Options{}).Score(context.Background(), state)
	if err != nil {
		t.Fatal(err)
	}
	for _, score := range scores {
		if score.PlayerID != "c" {
			continue
		}
		if *score.Team != (chesspairing.TeamPoints{Match: 1, Game: 2}) {
			t.Fatalf("PAB points = %+v, want 1 MP and 2 GP", score.Team)
		}
		return
	}
	t.Fatal("missing team c score")
}

func TestMatchGamePointsPreferCompleteBoards(t *testing.T) {
	match := chesspairing.MatchData{
		HomeID: "home",
		AwayID: "away",
		Boards: []chesspairing.GameData{
			{WhiteID: "home-player", BlackID: "away-player", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "away-player", BlackID: "home-player", Result: chesspairing.ResultBlackWins},
		},
		Result: &chesspairing.TeamMatchResult{HomeGame: 0, AwayGame: 2},
	}
	home, away := MatchGamePoints(match, Options{}.WithDefaults().Options, map[string]string{
		"home-player": "home",
		"away-player": "away",
	})
	if home != 2 || away != 0 {
		t.Fatalf("MatchGamePoints = %v-%v, want board totals 2-0", home, away)
	}
}

func TestScorePendingAndDoubleForfeitGiveNoMatchPoints(t *testing.T) {
	for _, result := range []chesspairing.GameResult{chesspairing.ResultPending, chesspairing.ResultDoubleForfeit} {
		state := &chesspairing.TournamentState{
			Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
			Rounds: []chesspairing.RoundData{{
				Number: 1,
				Matches: []chesspairing.MatchData{{
					HomeID: "a",
					AwayID: "b",
					Boards: []chesspairing.GameData{{Result: result}},
				}},
			}},
			CurrentRound: 2,
		}
		scores, err := New(Options{}).Score(context.Background(), state)
		if err != nil {
			t.Fatalf("Score: %v", err)
		}
		for _, score := range scores {
			if score.Team.Match != 0 {
				t.Errorf("%s match points = %v, want 0 for %v", score.PlayerID, score.Team.Match, result)
			}
		}
	}
}

func TestScoreExcludesWithdrawnAndUnknownTeams(t *testing.T) {
	withdrawn := 1
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a1", TeamID: "a"},
			{ID: "b1", TeamID: "b", WithdrawnAfterRound: &withdrawn},
		},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{
				{HomeID: "a", AwayID: "b", Result: &chesspairing.TeamMatchResult{HomeGame: 1, AwayGame: 0}},
				{HomeID: "a", AwayID: "unknown", Result: &chesspairing.TeamMatchResult{HomeGame: 1, AwayGame: 0}},
			},
		}},
		CurrentRound: 2,
	}
	scores, err := New(Options{}).Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if len(scores) != 1 || scores[0].PlayerID != "a" || scores[0].Team.Match != 0 {
		t.Errorf("scores = %+v, want only active team a without scored matches", scores)
	}
}

func TestScorePrimaryGameAndExplicitTotals(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{{ID: "a"}, {ID: "b"}},
		Rounds: []chesspairing.RoundData{{
			Number: 1,
			Matches: []chesspairing.MatchData{{
				HomeID: "a",
				AwayID: "b",
				Result: &chesspairing.TeamMatchResult{HomeGame: 2.5, AwayGame: 1.5},
			}},
		}},
		CurrentRound: 2,
	}
	scores, err := New(Options{PrimaryScore: "game"}).Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}
	if scores[0].PlayerID != "a" || scores[0].Score != 2.5 {
		t.Errorf("first score = %+v, want a with 2.5 game points", scores[0])
	}
}

func TestBoardCount_PAB_NoMatch(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", TeamID: "t1"},
			{ID: "p2", TeamID: "t1"},
			{ID: "p3", TeamID: "t2"},
			{ID: "p4", TeamID: "t2"},
			{ID: "p5", TeamID: "t3"},
			{ID: "p6", TeamID: "t3"},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				TeamByes: []chesspairing.ByeEntry{
					{PlayerID: "t1", Type: chesspairing.ByePAB},
				},
			},
		},
		CurrentRound: 2,
		ScoringConfig: chesspairing.ScoringConfig{
			System: chesspairing.ScoringTeam,
		},
	}
	// Should derive board count = 2 from team membership (each team has 2 players)
	// PAB should give MP 1, GP 1 * 2 = 2.
	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	var t1Score *chesspairing.TeamPoints
	for _, s := range scores {
		if s.PlayerID == "t1" {
			t1Score = s.Team
		}
	}
	if t1Score == nil || *t1Score != (chesspairing.TeamPoints{Match: 1, Game: 1}) {
		t.Errorf("t1Score = %#v, want MP: 1, GP: 1", t1Score)
	}

	// 802-only state
	state802 := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "t1"}, {ID: "t2"}, {ID: "t3"},
		}, // no team membership implies board count 0 (if not explicitly passed)
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Matches: []chesspairing.MatchData{
					{HomeID: "t2", AwayID: "t3", Result: &chesspairing.TeamMatchResult{HomeGame: 1, AwayGame: 1}},
				},
				TeamByes: []chesspairing.ByeEntry{
					{PlayerID: "t1", Type: chesspairing.ByePAB},
				},
			},
		},
		CurrentRound: 2,
		ScoringConfig: chesspairing.ScoringConfig{
			System: chesspairing.ScoringTeam,
		},
	}
	_, err = scorer.Score(context.Background(), state802)
	if err == nil {
		t.Errorf("expected error when board count cannot be inferred for team byes, got nil")
	} else if err.Error() != "cannot infer board count for team byes" {
		t.Errorf("expected 'cannot infer board count for team byes', got %v", err)
	}
}
