// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// doubleswissMatchState is a one-round Double Swiss between Alice (A) and
// Bob (B) in which the round holds both games of the match. rounds carries
// the two GameData entries of the match.
func doubleswissMatchState(games []chesspairing.GameData) *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", DisplayName: "Alice", Rating: 2000},
			{ID: "B", DisplayName: "Bob", Rating: 1900},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: games},
		},
	}
}

// doubleswissAssertValues computes one tie-break and asserts one value per
// player.
func doubleswissAssertValues(t *testing.T, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore, want map[string]float64) {
	t.Helper()
	tb, err := Get(id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("%s.Compute: %v", id, err)
	}
	got := make(map[string]float64, len(values))
	for _, value := range values {
		got[value.PlayerID] = value.Value
	}
	for playerID, expected := range want {
		if actual := got[playerID]; actual != expected {
			t.Errorf("%s: player %s = %v, want %v", id, playerID, actual, expected)
		}
	}
}

var doubleswissABScores = []chesspairing.PlayerScore{
	{PlayerID: "A", Score: 1, Rank: 1},
	{PlayerID: "B", Score: 1, Rank: 1},
}

// TestRoundWithDifferentOpponentsKeepsFirstPair checks that malformed round
// data cannot choose a different opponent record according to map iteration.
func TestRoundWithDifferentOpponentsKeepsFirstPair(t *testing.T) {
	for _, games := range [][]chesspairing.GameData{
		{
			{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "A", BlackID: "C", Result: chesspairing.ResultWhiteWins},
		},
		{
			{WhiteID: "A", BlackID: "C", Result: chesspairing.ResultWhiteWins},
			{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		},
	} {
		state := &chesspairing.TournamentState{
			Players: []chesspairing.PlayerEntry{{ID: "A"}, {ID: "B"}, {ID: "C"}},
			Rounds:  []chesspairing.RoundData{{Number: 1, Games: games}},
		}
		for i := 0; i < 100; i++ {
			records := buildOpponentRecords(state, nil).records
			if got := records["A"][0].OpponentID; got != games[0].BlackID {
				t.Fatalf("first opponent = %s, got %s", games[0].BlackID, got)
			}
			if got := records[games[0].BlackID][0].OpponentID; got != "A" {
				t.Fatalf("first pair lost reciprocal record: %s", got)
			}
			if got := records[games[1].BlackID][0].OpponentID; got != "" {
				t.Fatalf("conflicting pair was stored: %s", got)
			}
		}
	}
}

// TestDoubleswissSplitRoundStandardPoints asserts C.07 Article 7.7: the two
// games of one Double Swiss round are added before the comparison with the
// scheduled opponent, so a split match (1-0 followed by 0-1) counts as half
// a standard point for both players. Before, the second game replaced the
// first and STD was 0 for Alice and 1 for Bob.
func TestDoubleswissSplitRoundStandardPoints(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultWhiteWins},
	})
	doubleswissAssertValues(t, "standard-points", state, doubleswissABScores, map[string]float64{
		"A": 0.5,
		"B": 0.5,
	})
}

// TestDoubleswissSplitRoundMatchTotal asserts that the per-round opponent
// records hold the match total: Buchholz (Article 8.1) sums the opponents'
// scores, Sonneborn-Berger (Article 9.1) multiplies the final score by the
// points scored against them over the whole match, and the direct
// encounter counts the match once. Before, the second game replaced the
// first, and Sonneborn-Berger and the direct encounter only saw the second
// game's 0 points for Alice.
func TestDoubleswissSplitRoundMatchTotal(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultWhiteWins},
	})
	doubleswissAssertValues(t, "buchholz", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
	doubleswissAssertValues(t, "sonneborn-berger", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
	doubleswissAssertValues(t, "direct-encounter", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
}

// TestDoubleswissPerGameTiebreaksCountBothGames asserts the tie-breaks that
// C.07:2026 defines per game: games won over the board (Article 7.2) and
// games played with Black (Article 7.3) still count each game of the match
// separately. Before, the record table held only the second game and
// counted Alice's win of the first game as a loss.
func TestDoubleswissPerGameTiebreaksCountBothGames(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultWhiteWins},
	})
	doubleswissAssertValues(t, "wins", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
	doubleswissAssertValues(t, "black-games", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
}

// TestDoubleswissForfeitedSecondGameMergedAsPlayed asserts the C.04.5
// Preface: a match with a single forfeited game (Bob forfeits the second
// game) is regularly played, and any forfeited single game is treated as
// played. Bob's round is therefore not a forfeit loss (REP 1, not 0) and
// Bob's Buchholz contribution is the scheduled opponent's adjusted score
// (2), not a dummy score; Alice's merged round record adds both games'
// points, so her Buchholz contribution sees Bob's adjusted score of 0.
func TestDoubleswissForfeitedSecondGameMergedAsPlayed(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins},
	})
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 2, Rank: 1},
		{PlayerID: "B", Score: 0, Rank: 2},
	}
	doubleswissAssertValues(t, "ge", state, scores, map[string]float64{
		"A": 1,
		"B": 1,
	})
	doubleswissAssertValues(t, "buchholz", state, scores, map[string]float64{
		"A": 0,
		"B": 2,
	})
}

// TestDoubleswissForfeitedMatchKeepsForfeitLoss asserts that when one
// player forfeits both games (C.04.5 Preface: the match ends by forfeit)
// the merged round record stays a forfeit loss for Alice and a forfeit win
// for Bob, so Alice's round remains a voluntary unplayed round (REP 0) and
// Bob still counts his round (REP 1).
func TestDoubleswissForfeitedMatchKeepsForfeitLoss(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitBlackWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultForfeitWhiteWins},
	})
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 0, Rank: 2},
		{PlayerID: "B", Score: 2, Rank: 1},
	}
	doubleswissAssertValues(t, "ge", state, scores, map[string]float64{
		"A": 0,
		"B": 1,
	})
}

// TestDoubleswissMixedForfeitGamesMergedAsPlayed asserts the C.04.5
// Preface for the case where each player forfeits a different game of the
// match: nobody forfeited both games, so the match is regularly played and
// both rows count (REP 1 for each), while no game was played over the
// board (black-games 0 for each).
func TestDoubleswissMixedForfeitGamesMergedAsPlayed(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultForfeitWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultForfeitWhiteWins},
	})
	doubleswissAssertValues(t, "ge", state, doubleswissABScores, map[string]float64{
		"A": 1,
		"B": 1,
	})
	doubleswissAssertValues(t, "black-games", state, doubleswissABScores, map[string]float64{
		"A": 0,
		"B": 0,
	})
}

// doubleswissValues computes one tie-break and maps the values by player.
func doubleswissValues(t *testing.T, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) map[string]float64 {
	t.Helper()
	tb, err := Get(id)
	if err != nil {
		t.Fatalf("Get(%q): %v", id, err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("%s.Compute: %v", id, err)
	}
	got := make(map[string]float64, len(values))
	for _, value := range values {
		got[value.PlayerID] = value.Value
	}
	return got
}

// TestDoubleswissPerformanceCountsGames asserts C.07:2026 Article 10.2 on a
// Double Swiss split match (1-0 followed by 0-1): the fractional score is
// "the number of points achieved in games played over the board divided by
// the number of games", so Alice's 1 point in 2 games gives dp 0 and both
// players' performance rating stays at the shared opponent rating of 1900,
// while the match total (Points 1 over one merged record) would have given
// dp +800 and a rating of 2700. PTP (Article 10.3) also faces "each
// opponent" per game, so the expected scores are summed over both games
// and PTP = 1900 as well. In the 2-0 match the perfect score gives dp +800
// and TPR = PTP = 2700.
// doubleswissRatedMatchState is a one-round Double Swiss match between
// Alice (A) and Bob (B) with both ratings set to rating.
func doubleswissRatedMatchState(games []chesspairing.GameData, aRating, bRating int) *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", DisplayName: "Alice", Rating: aRating},
			{ID: "B", DisplayName: "Bob", Rating: bRating},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss},
		Rounds:        []chesspairing.RoundData{{Number: 1, Games: games}},
	}
}

func TestDoubleswissPerformanceCountsGames(t *testing.T) {
	splitGames := []chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultWhiteWins},
	}
	splitState := doubleswissRatedMatchState(splitGames, 1900, 1900)
	splitScores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 1, Rank: 1},
		{PlayerID: "B", Score: 1, Rank: 1},
	}
	doubleswissAssertValues(t, "performance-rating", splitState, splitScores, map[string]float64{
		"A": 1900,
		"B": 1900,
	})
	doubleswissAssertValues(t, "performance-points", splitState, splitScores, map[string]float64{
		"A": 1897,
		"B": 1897,
	})

	sweepGames := []chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultBlackWins},
	}
	sweepState := doubleswissRatedMatchState(sweepGames, 1900, 1900)
	sweepScores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 2, Rank: 1},
		{PlayerID: "B", Score: 0, Rank: 2},
	}
	doubleswissAssertValues(t, "performance-rating", sweepState, sweepScores, map[string]float64{
		"A": 2700,
		"B": 1100,
	})
	doubleswissAssertValues(t, "performance-points", sweepState, sweepScores, map[string]float64{
		"A": 2700,
		"B": 1100,
	})
}

// TestDoubleswissMergeOrderIndependent asserts that the per-round records
// do not depend on the order of the two GameData entries of a Double Swiss
// match: for every combination of two game results, Buchholz (Article 8.1),
// games elected to play (Article 7.6) and Sonneborn-Berger (Article 9.1)
// come out identical when the round lists the games in either order.
func TestDoubleswissMergeOrderIndependent(t *testing.T) {
	results := []chesspairing.GameResult{
		chesspairing.ResultWhiteWins,
		chesspairing.ResultBlackWins,
		chesspairing.ResultDraw,
		chesspairing.ResultForfeitWhiteWins,
		chesspairing.ResultForfeitBlackWins,
		chesspairing.ResultDoubleForfeit,
		chesspairing.ResultPending,
	}
	for _, first := range results {
		for _, second := range results {
			forward := doubleswissMatchState([]chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: first},
				{WhiteID: "B", BlackID: "A", Result: second},
			})
			reverse := doubleswissMatchState([]chesspairing.GameData{
				forward.Rounds[0].Games[1],
				forward.Rounds[0].Games[0],
			})
			for _, id := range []string{"buchholz", "ge", "sonneborn-berger"} {
				want := doubleswissValues(t, id, forward, doubleswissABScores)
				got := doubleswissValues(t, id, reverse, doubleswissABScores)
				for _, playerID := range []string{"A", "B"} {
					if got[playerID] != want[playerID] {
						t.Errorf("%s with results %v/%v: order affects player %s (%v, want %v)",
							id, first, second, playerID, got[playerID], want[playerID])
					}
				}
			}
		}
	}
}

// TestDoubleswissDoubleForfeitWithForfeitWin asserts the C.04.5 Preface at
// match level: Alice forfeits the first game (a double forfeit) and the
// second game (a forfeit win for Bob), so she forfeited both games and the
// match ends by forfeit. Bob's round stays a forfeit win he counts as
// played (REP 1); it is not a regularly played round, and Alice's round
// stays an unplayed forfeit loss (REP 0).
func TestDoubleswissDoubleForfeitWithForfeitWin(t *testing.T) {
	state := doubleswissMatchState([]chesspairing.GameData{
		{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultDoubleForfeit},
		{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultForfeitWhiteWins},
	})
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 0, Rank: 2},
		{PlayerID: "B", Score: 1, Rank: 1},
	}
	doubleswissAssertValues(t, "ge", state, scores, map[string]float64{
		"A": 0,
		"B": 1,
	})
}

// TestDoubleswissForeBuchholzDrawsWholeMatch asserts Fore Buchholz
// (Article 8.3) on a Double Swiss: every paired game of the last round is
// drawn, so a 2-0 final-round match contributes the draw value of both
// games (1.0) to the player's adjusted score instead of the match total of
// 2. Alice scores 4 in two 2-0 matches against Bob; with both final-round
// games drawn Bob's adjusted score is 1 (0 + 1.0), so her Fore Buchholz is
// 2 (two rounds against Bob), while Bob's sees Alice's adjusted 3.
func TestDoubleswissForeBuchholzDrawsWholeMatch(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "A", DisplayName: "Alice", Rating: 2000},
			{ID: "B", DisplayName: "Bob", Rating: 1900},
			{ID: "C", DisplayName: "Carol", Rating: 2000},
			{ID: "D", DisplayName: "Dave", Rating: 1900},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDoubleSwiss},
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultBlackWins},
				{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultDraw},
				{WhiteID: "D", BlackID: "C", Result: chesspairing.ResultDraw},
			}},
			{Number: 2, Games: []chesspairing.GameData{
				{WhiteID: "A", BlackID: "B", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "B", BlackID: "A", Result: chesspairing.ResultBlackWins},
				{WhiteID: "C", BlackID: "D", Result: chesspairing.ResultDraw},
				{WhiteID: "D", BlackID: "C", Result: chesspairing.ResultDraw},
			}},
		},
	}
	scores := []chesspairing.PlayerScore{
		{PlayerID: "A", Score: 4, Rank: 1},
		{PlayerID: "B", Score: 0, Rank: 4},
		{PlayerID: "C", Score: 1, Rank: 2},
		{PlayerID: "D", Score: 1, Rank: 2},
	}
	doubleswissAssertValues(t, "fore-buchholz", state, scores, map[string]float64{
		"A": 2,
		"B": 6,
		"C": 2,
		"D": 2,
	})
}
