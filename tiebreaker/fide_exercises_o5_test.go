// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

// FIDE exercise set accompanying C.07 (2023): "Exercises in tie-breaking",
// IA Mario Held, Rev. 2403220900 / C.07-2023, V01 2024-03-22. Block o5 =
// PART TWO – TEAM TOURNAMENTS, exercises 34 through 49 (pages 52 through 70).
//
// # Reconstructed tournament
//
// The Swiss team tournament from section 2.3 (pages 6 through 9): 14 teams of
// 4 players, 7 rounds, team score 2-1-0 (MP) and game score 1-½-0 (GP)
// [11.1]. Point values for unplayed matches (page 8):
//
//	PAB: 1 MP, 2 GP, 0 player points   (#12 round 5, #14 round 7)
//	HPB: 1 MP, 2 GP, 0 player points   (#14 round 5)
//	ZPB: 0 MP, 0 GP, 0 player points   (#7 round 7)
//	-F:  0 MP, 0 GP, 0 player points   (#6 round 6)
//	+F:  2 MP, 4 GP, 1 point per player (#1 round 6)
//
// # Rebuild used and the limits of the data model
//
// A team is modelled as a single participant (PlayerEntry whose ID equals the
// team ID) and a team match as a MatchData. Matches with only reported GP
// totals use MatchData.Result; the round-3 match #11-#14 carries its four
// board results in MatchData.Boards (for BC/TBR/BBE), and the round-6 forfeit
// match #6-#1 carries four forfeit boards so Article 16's dummy handling is
// expressible.
//
// What remains inexpressible: the team direct encounter (EDE, Article 13.3)
// because the registry has no team variant of direct-encounter; the GP-based
// Buchholz (exercise 37) because only the Buchholz-on-MP variants exist; and
// SSSC (Article 13.4) because the registry has no SSSC variant. The inverted
// {GP, MP} standings of exercise 34b cannot be produced because standings
// carry a single primary score.

import (
	"context"
	"math"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
)

var FideAOrderForExternalTest map[string]int

func ExportO5TeamState() *chesspairing.TournamentState { return o5TeamState() }
func ExportO5MPScores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	return o5MPScores(t, state)
}

// o5Team is one team from the cross table of section 2.3 (page 6).
// mp and gp are the final scores as printed by FIDE.
type o5Team struct {
	id   string
	name string
	mp   float64
	gp   float64
}

// o5Teams in the order of the team number (#1 through #14), with the MP and
// GP final scores from the cross table (page 6, repeated in every exercise
// table).
var o5Teams = []o5Team{
	{id: "T1", name: "Antelopes", mp: 10, gp: 17.5},
	{id: "T2", name: "Bonobos", mp: 10, gp: 17},
	{id: "T3", name: "Cougars", mp: 10, gp: 16},
	{id: "T4", name: "Deer", mp: 10, gp: 17},
	{id: "T5", name: "Elephants", mp: 10, gp: 18},
	{id: "T6", name: "Falcons", mp: 7, gp: 12.5},
	{id: "T7", name: "Giraffes", mp: 6, gp: 11.5},
	{id: "T8", name: "Hippopotami", mp: 7, gp: 15},
	{id: "T9", name: "Iguanas", mp: 6, gp: 14.5},
	{id: "T10", name: "Jackals", mp: 5, gp: 13},
	{id: "T11", name: "Koalas", mp: 4, gp: 11.5},
	{id: "T12", name: "Lynxes", mp: 2, gp: 7.5},
	{id: "T13", name: "Moose", mp: 6, gp: 11.5},
	{id: "T14", name: "Narwhals", mp: 4, gp: 11.5},
}

// o5TeamState rebuilds the team tournament from the cross table on page 6.
// Each round lists the cross-table cell that determines the result behind
// every match (white team first).
//
//	R1: 1w2½-8, 9w1-2b3, 3w2½-10, 11w1-4b3, 5w4-12, 6w2-13, 7w2-14
//	R2: 4w2½-1, 2w4-3, 6w1-5b3, 8w2-7, 12w0-9b4, 10w1½-11b2½, 14w1½-13b2½
//	R3: 1w3-7, 5w4-2, 3w3-9, 13w1-4b3, 8w1-6b3, 12w1½-10b2½, 11w2-14
//	R4: 2w1½-1b2½, 6w1½-3b2½, 4w3½-5, 7w2½-10, 12w½-8b3½, 9w3-14, 13w2½-11
//	R5: 1w1½-5b2½, 2w3-13, 3w3-4, 11w1½-6b2½, 9w2-7, 10w2-8; PAB #12, HPB #14
//	R6: 6w -F/+F 1b, 4w1½-2b2½, 5w½-3b3½, 13w2-7, 8w2-9, 14w1-10b3, 11w2-12
//	R7: 3w1½-1b2½, 10w1-2b3, 9w1½-4b2½, 13w½-5b3½, 6w2½-12, 8w3-11; ZPB #7, PAB #14
//
// The colour in round 6 of the forfeit match follows from the individual
// cross table (page 7): player 1 (team 1) had "18b+", so team 1 was black.
func o5TeamState() *chesspairing.TournamentState {
	// match creates a team match with the given result.
	match := func(white, black string, homeGP, awayGP float64) chesspairing.MatchData {
		return chesspairing.MatchData{
			HomeID: white,
			AwayID: black,
			Result: &chesspairing.TeamMatchResult{HomeGame: homeGP, AwayGame: awayGP},
		}
	}

	state := &chesspairing.TournamentState{
		Players: make([]chesspairing.PlayerEntry, 0, len(o5Teams)),
		Rounds: []chesspairing.RoundData{
			{Number: 1, Matches: []chesspairing.MatchData{
				match("T1", "T8", 2.5, 1.5),  // 8w2½ / 1b1½
				match("T9", "T2", 1, 3),      // 9b3 / 2w1
				match("T3", "T10", 2.5, 1.5), // 10w2½ / 3b1½
				match("T11", "T4", 1, 3),     // 11b3 / 4w1
				match("T5", "T12", 4, 0),     // 12w4 / 5b0
				match("T6", "T13", 2, 2),     // 13b2 / 6w2
				match("T7", "T14", 2, 2),     // 14w2 / 7b2
			}},
			{Number: 2, Matches: []chesspairing.MatchData{
				match("T4", "T1", 2.5, 1.5),   // 4b1½ / 1w2½
				match("T2", "T3", 4, 0),       // 3w4 / 2b0
				match("T6", "T5", 1, 3),       // 6b3 / 5w1
				match("T8", "T7", 2, 2),       // 8b2 / 7w2
				match("T12", "T9", 0, 4),      // 12b4 / 9w0
				match("T10", "T11", 1.5, 2.5), // 11w1½ / 10b2½
				match("T14", "T13", 1.5, 2.5), // 14b2½ / 13w1½
			}},
			{Number: 3, Matches: []chesspairing.MatchData{
				match("T1", "T7", 3, 1),       // 7w3 / 1b1
				match("T5", "T2", 4, 0),       // 5b4 / 2w0
				match("T3", "T9", 3, 1),       // 9w3 / 3b1
				match("T13", "T4", 1, 3),      // 13b3 / 4w1
				match("T8", "T6", 1, 3),       // 8b3 / 6w1
				match("T12", "T10", 1.5, 2.5), // 12b2½ / 10w1½
				{
					HomeID: "T11", AwayID: "T14",
					Boards: []chesspairing.GameData{
						{WhiteID: "T11", BlackID: "T14", Result: chesspairing.ResultWhiteWins},
						{WhiteID: "T14", BlackID: "T11", Result: chesspairing.ResultDraw},
						{WhiteID: "T11", BlackID: "T14", Result: chesspairing.ResultDraw},
						{WhiteID: "T14", BlackID: "T11", Result: chesspairing.ResultWhiteWins},
					},
				},
			}},
			{Number: 4, Matches: []chesspairing.MatchData{
				match("T2", "T1", 1.5, 2.5),   // 2b2½ / 1w1½
				match("T6", "T3", 1.5, 2.5),   // 6b2½ / 3w1½
				match("T4", "T5", 3.5, 0.5),   // 5w3½ / 4b½
				match("T7", "T10", 2.5, 1.5),  // 10w2½ / 7b1½
				match("T12", "T8", 0.5, 3.5),  // 12b3½ / 8w½
				match("T9", "T14", 3, 1),      // 14w3 / 9b1
				match("T13", "T11", 2.5, 1.5), // 13b1½ / 11w2½
			}},
			{Number: 5, Matches: []chesspairing.MatchData{
				match("T1", "T5", 1.5, 2.5),  // 5w1½ / 1b2½
				match("T2", "T13", 3, 1),     // 13w3 / 2b1
				match("T3", "T4", 3, 1),      // 4w3 / 3b1
				match("T11", "T6", 1.5, 2.5), // 11b2½ / 6w1½
				match("T9", "T7", 2, 2),      // 9b2 / 7w2
				match("T10", "T8", 2, 2),     // 10b2 / 8w2
			}, TeamByes: []chesspairing.ByeEntry{
				{PlayerID: "T12", Type: chesspairing.ByePAB},  // PAB: 1 MP, 2 GP
				{PlayerID: "T14", Type: chesspairing.ByeHalf}, // HPB: 1 MP, 2 GP
			}},
			{Number: 6, Matches: []chesspairing.MatchData{
				// -F / +F: team 6 loses, team 1 wins by forfeit.
				{
					HomeID: "T6", AwayID: "T1",
					Boards: []chesspairing.GameData{
						{WhiteID: "T6", BlackID: "T1", Result: chesspairing.ResultForfeitBlackWins},
						{WhiteID: "T1", BlackID: "T6", Result: chesspairing.ResultForfeitWhiteWins},
						{WhiteID: "T6", BlackID: "T1", Result: chesspairing.ResultForfeitBlackWins},
						{WhiteID: "T1", BlackID: "T6", Result: chesspairing.ResultForfeitWhiteWins},
					},
				},
				match("T4", "T2", 1.5, 2.5), // 4b2½ / 2w1½
				match("T5", "T3", 0.5, 3.5), // 5b3½ / 3w½
				match("T13", "T7", 2, 2),    // 7b2 / 13w2
				match("T8", "T9", 2, 2),     // 9w2 / 8b2
				match("T14", "T10", 1, 3),   // 14b3 / 10w1
				match("T11", "T12", 2, 2),   // 12w2 / 11b2
			}},
			{Number: 7, Matches: []chesspairing.MatchData{
				match("T3", "T1", 1.5, 2.5),  // 3b2½ / 1w1½
				match("T10", "T2", 1, 3),     // 10b3 / 2w1
				match("T9", "T4", 1.5, 2.5),  // 9b2½ / 4w1½
				match("T13", "T5", 0.5, 3.5), // 13b3½ / 5w½
				match("T6", "T12", 2.5, 1.5), // 12w2½ / 6b1½
				match("T8", "T11", 3, 1),     // 11w3 / 8b1
			}, TeamByes: []chesspairing.ByeEntry{
				{PlayerID: "T7", Type: chesspairing.ByeZero}, // ZPB: 0 MP, 0 GP
				{PlayerID: "T14", Type: chesspairing.ByePAB}, // PAB: 1 MP, 2 GP
			}},
		},
	}

	// Teams receive no rating in the exercise; Rating stays 0. TeamID is set
	// to the team's own ID so board games stored under team IDs are attributed
	// to the right side regardless of board colour.
	for _, tm := range o5Teams {
		state.Players = append(state.Players, chesspairing.PlayerEntry{
			ID:          tm.id,
			TeamID:      tm.id,
			DisplayName: tm.name,
		})
	}
	return state
}

// o5MPScorer is the scorer with the point values of the team tournament:
// 2-1-0 for match points, PAB/HPB = 1 MP, ZPB = 0 MP, +F = 2 MP,
// -F = 0 MP (page 5 and page 8).
func o5MPScorer() *team.Scorer {
	return team.New(team.Options{
		PointMatchWin:  chesspairing.Float64Ptr(2),
		PointMatchDraw: chesspairing.Float64Ptr(1),
		PointMatchLoss: chesspairing.Float64Ptr(0),
	})
}

// o5MPScores computes the match points (MP) of all teams.
func o5MPScores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	t.Helper()
	scores, err := o5MPScorer().Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/team (MP): %v", err)
	}
	return scores
}

// o5AssertMP is the hard assertion of the transcription: the match points
// computed with scoring/team must equal the MP column of the cross table
// (page 6).
func o5AssertMP(t *testing.T, nr int, scores []chesspairing.PlayerScore) {
	t.Helper()
	got := make(map[string]float64, len(scores))
	for _, ps := range scores {
		got[ps.PlayerID] = ps.Score
	}
	for _, tm := range o5Teams {
		v, ok := got[tm.id]
		if !ok {
			t.Errorf("EXERCISE %d: transcription error: team %s (%s) missing from the score list", nr, tm.id, tm.name)
			continue
		}
		if math.Abs(v-tm.mp) > 0.001 {
			t.Errorf("EXERCISE %d: transcription error: %s %s MP = %v, FIDE cross table (p.6) = %v",
				nr, tm.id, tm.name, v, tm.mp)
		}
	}
}

// o5Label is the player column in the log lines: "T1 Antelopes".
func o5Label(teamID string) string {
	for _, tm := range o5Teams {
		if tm.id == teamID {
			return tm.id + " " + tm.name
		}
	}
	return teamID
}

// o5Values computes a tiebreak from the registry and returns a map
// teamID → value.
func o5Values(t *testing.T, id string, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) map[string]float64 {
	t.Helper()
	tb, err := Get(id)
	if err != nil {
		t.Fatalf("tiebreaker.Get(%q): %v", id, err)
	}
	values, err := tb.Compute(context.Background(), state, scores)
	if err != nil {
		t.Fatalf("%s.Compute: %v", id, err)
	}
	out := make(map[string]float64, len(values))
	for _, v := range values {
		out[v.PlayerID] = v.Value
	}
	return out
}

// o5AssertCmp asserts one comparison per team with tolerance 0.001.
func o5AssertCmp(t *testing.T, nr int, tb, teamID string, fide, code float64) {
	t.Helper()
	if math.Abs(fide-code) > 0.001 {
		t.Errorf("EXERCISE %d | %s | %s | FIDE=%v | CODE=%v | DISCREPANCY", nr, tb, o5Label(teamID), fide, code)
	}
}

// o5LogMissing logs a line for a value the code cannot compute, with the
// reason in marker. The exercise report lists these as inexpressible.
func o5LogMissing(t *testing.T, nr int, tb, teamID string, fide float64, marker string) {
	t.Helper()
	t.Logf("EXERCISE %d | %s | %s | FIDE=%v | CODE=missing | %s", nr, tb, o5Label(teamID), fide, marker)
}

// ---------------------------------------------------------------------------
// Exercise 34 – Match points versus game points (MPVGP), pages 52-53.
//
// FIDE values: the two final standings on page 53. First table (a): {MP, GP}
// with the order 5, 1, 2, 4, 3, 8, 6, 9, 7, 13, 10, 11, 14, 12.
// Second table (b): {GP, MP} with the order 5, 1, 2, 4, 3, 8, 9, 10, 6, 7,
// 13, 11, 14, 12.
//
// The mpvgp tiebreak correctly resolves the standings. Part (a) asserts the
// places produced by a standings build. Part (b) evaluates {GP, MP} which
// is an inexpressible variant since mpvgp only orders by secondary score
// when primary scores are tied, but part (b) asks to sort by secondary score
// unconditionally.
// ---------------------------------------------------------------------------

func TestFIDEExercise_34_MPVGPStand(t *testing.T) {
	const nr = 34
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fideA := map[string]int{
		"T5": 1, "T1": 2, "T2": 3, "T4": 4, "T3": 5, "T8": 6, "T6": 7,
		"T9": 8, "T13": 9, "T7": 10, "T10": 11, "T11": 12, "T14": 13, "T12": 14,
	}
	fideB := map[string]int{
		"T5": 1, "T1": 2, "T2": 3, "T4": 4, "T3": 5, "T8": 6, "T9": 7,
		"T10": 8, "T6": 9, "T7": 10, "T13": 11, "T11": 12, "T14": 13, "T12": 14,
	}

	// GP final scores: can now be verified as MPVGP values.
	mpvgp := o5Values(t, "mpvgp", state, scores)
	for _, tm := range o5Teams {
		o5AssertCmp(t, nr, "MPVGP(a) GP final score", tm.id, tm.gp, mpvgp[tm.id])
	}

	// Assert {MP, GP} standings order using O5MPVGPOrder (in tiebreaker_test)
	FideAOrderForExternalTest = fideA

	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "MPVGP(b) place {GP,MP}", tm.id, float64(fideB[tm.id]), "NOT_EXPRESSIBLE: data model carries only one primary score; standings cannot be inverted to {GP, MP}")
	}
}

// ---------------------------------------------------------------------------
// Exercise 35 – Buchholz (total) on match points, pages 53-55.
//
// The published 2023 answers are the "BH" column on pages 54-55 (team #1 is
// 64). The values below follow C.07:2026 instead: Article 16.4 caps a forfeit
// win's dummy opponent at the scheduled opponent's adjusted score (16.4.1),
// so #1's round-6 forfeit win over #6 contributes #6's adjusted 7 MP rather
// than #1's own 10 MP, and #1 totals 61. The 2023 answers are quoted only for
// reference and are deliberately not reproduced.
// ---------------------------------------------------------------------------

func TestFIDEExercise_35_BuchholzMP(t *testing.T) {
	const nr = 35
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// BH (MP) per team under C.07:2026 Article 16.4 (2023 table answer for #1: 64).
	fide := map[string]float64{
		"T1": 61, "T2": 57, "T3": 58, "T4": 56, "T5": 55, "T6": 46, "T7": 44,
		"T8": 41, "T9": 50, "T10": 44, "T11": 41, "T12": 41, "T13": 52, "T14": 36,
	}

	code := o5Values(t, "buchholz-mp", state, scores)
	for _, tm := range o5Teams {
		o5AssertCmp(t, nr, "BH(MP)", tm.id, fide[tm.id], code[tm.id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 36 – Buchholz Cut-1 on match points, page 55.
//
// The published 2023 answer for team #1 is 57 ("BH-C1" column, page 55).
// The values below follow C.07:2026 instead: the same Article 16.4 dummy cap
// as exercise 35 lowers #1's round-6 forfeit-win contribution, so #1 totals
// 54. The 2023 answers are quoted only for reference.
// ---------------------------------------------------------------------------

func TestFIDEExercise_36_BuchholzCut1MP(t *testing.T) {
	const nr = 36
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// BH-C1 (MP) per team under C.07:2026 (2023 table answer for #1: 57).
	fide := map[string]float64{
		"T1": 54, "T2": 52, "T3": 53, "T4": 52, "T5": 53, "T6": 39, "T7": 38,
		"T8": 39, "T9": 48, "T10": 42, "T11": 39, "T12": 39, "T13": 48, "T14": 32,
	}

	code := o5Values(t, "buchholz-mp-cut1", state, scores)
	for _, tm := range o5Teams {
		o5AssertCmp(t, nr, "BH-C1(MP)", tm.id, fide[tm.id], code[tm.id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 37 – Buchholz Cut-1 on game points (GP as primary score), page 56.
//
// FIDE values: the "BH GP" and "BH-C1" columns of the table on page 56.
//
// GP totals are expressible through TeamPoints, but the registry has no
// Buchholz-on-GP variant; only the Buchholz-on-MP variants (exercises 35-36)
// are implemented. The MP column of the same table is verified in o5AssertMP.
// ---------------------------------------------------------------------------

func TestFIDEExercise_37_BuchholzCut1GP(t *testing.T) {
	const nr = 37
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// Table page 56: BH(GP) and BH-C1(GP).
	fideBH := map[string]float64{
		"T1": 114.0, "T2": 107.5, "T3": 109.5, "T4": 106.0, "T5": 99.0,
		"T6": 92.0, "T7": 94.5, "T8": 90.0, "T9": 97.5, "T10": 92.0,
		"T11": 88.0, "T12": 92.0, "T13": 101.0, "T14": 87.0,
	}
	fideC1 := map[string]float64{
		"T1": 100.5, "T2": 96.0, "T3": 97.0, "T4": 94.5, "T5": 91.5,
		"T6": 79.5, "T7": 83.0, "T8": 82.5, "T9": 90.0, "T10": 84.5,
		"T11": 80.5, "T12": 84.5, "T13": 89.5, "T14": 75.5,
	}

	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "BH(GP)", tm.id, fideBH[tm.id], "NOT_IMPLEMENTED: no Buchholz-on-GP variant in the registry")
	}
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "BH-C1(GP)", tm.id, fideC1[tm.id], "NOT_IMPLEMENTED: no Buchholz-on-GP variant in the registry")
	}
}

// ---------------------------------------------------------------------------
// Exercise 38 – EMMSB and EMMSB Cut-1 for the teams on 10 MP, pages 57-58.
//
// EMMSB = Σ (opponent MP × own MP), with 2-1-0 weighting. The published 2023
// answers are the "EMMSB" and "EMMSB Cut-1" columns on pages 57-58 (team #1
// is 88 / 74). The values below follow C.07:2026: the Article 16.4 dummy cap
// lowers #1's round-6 forfeit-win dummy from 10 to 7 MP, so #1 totals 82 / 68.
// The 2023 answers are quoted only for reference.
// ---------------------------------------------------------------------------

func TestFIDEExercise_38_EMMSBCut1(t *testing.T) {
	const nr = 38
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// Teams on 10 MP; EMMSB / EMMSB-C1 under C.07:2026 (2023 answer for #1: 88 / 74).
	fideSB := map[string]float64{"T1": 82, "T2": 74, "T3": 76, "T4": 72, "T5": 70}
	fideC1 := map[string]float64{"T1": 68, "T2": 64, "T3": 66, "T4": 64, "T5": 66}

	sb := o5Values(t, "emmsb", state, scores)
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5AssertCmp(t, nr, "EMMSB", id, fideSB[id], sb[id])
	}
	sbC1 := o5Values(t, "emmsb-cut1", state, scores)
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5AssertCmp(t, nr, "EMMSB-C1", id, fideC1[id], sbC1[id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 39 – EGMSB for the teams on 10 MP, page 59.
//
// EGMSB = Σ (opponent GP × own MP). The published 2023 answers are the
// "EGMSB" column on page 59 (158.0 / 144.0 / 150.0 / 146.0 / 132.0). The
// values below follow C.07:2026: the Article 16.4 dummy cap lowers #1's
// round-6 forfeit-win game-point dummy, so #1 totals 148.0. The 2023 answers
// are quoted only for reference.
// ---------------------------------------------------------------------------

func TestFIDEExercise_39_EGMSB(t *testing.T) {
	const nr = 39
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T1": 148.0, "T2": 144.0, "T3": 150.0, "T4": 146.0, "T5": 132.0}
	egmsb := o5Values(t, "egmsb", state, scores)
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5AssertCmp(t, nr, "EGMSB", id, fide[id], egmsb[id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 40 – EMGSB for the teams on 6 MP, pages 59-60.
//
// FIDE values: the "EMGSB" column of the tables on page 60 (#7 = 68.5;
// #9 = 83.0; #13 = 73.0). EMGSB = Σ (opponent MP × own GP): the own GP per
// match is not expressible.
// ---------------------------------------------------------------------------

func TestFIDEExercise_40_EMGSB(t *testing.T) {
	const nr = 40
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T7": 68.5, "T9": 83.0, "T13": 73.0}
	emgsb := o5Values(t, "emgsb", state, scores)
	for _, id := range []string{"T7", "T9", "T13"} {
		o5AssertCmp(t, nr, "EMGSB", id, fide[id], emgsb[id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 41 – EGGSB (and EGGSB Cut-1) for the teams on 7 MP, pages 60-61.
//
// FIDE values: the "EGGSB" column of the tables on page 61 (#6 = 157.50;
// #8 = 181.50). Because the two values differ there is no tie among the
// remainder and the Cut-1 variant is not addressed. EGGSB = Σ (opponent GP ×
// own GP).
// ---------------------------------------------------------------------------

func TestFIDEExercise_41_EGGSB(t *testing.T) {
	const nr = 41
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T6": 157.50, "T8": 181.50}
	eggsb := o5Values(t, "eggsb", state, scores)
	for _, id := range []string{"T6", "T8"} {
		o5AssertCmp(t, nr, "EGGSB", id, fide[id], eggsb[id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 42 – EDE for the teams on 4 MP (#11 Koalas, #14 Narwhals),
// pages 62-63.
//
// FIDE values: the quoted cross-table cells on page 63, round 3: "14w2" for
// #11 and "11b2" for #14 → 2 GP each, so 1 MP each (2-1-0, page 52).
// Conclusion of the solution: equal, next tiebreak.
//
// [13.3]/[13.3.1]: the registry has no team direct-encounter (EDE) variant;
// the individual direct-encounter tiebreak reads Games, not Matches.
// ---------------------------------------------------------------------------

func TestFIDEExercise_42_EDE4MP(t *testing.T) {
	const nr = 42
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	for _, id := range []string{"T11", "T14"} {
		o5LogMissing(t, nr, "EDE(MP)", id, 1, "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
	for _, id := range []string{"T11", "T14"} {
		o5LogMissing(t, nr, "EDE(GP)", id, 2, "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
}

// ---------------------------------------------------------------------------
// Exercise 43 – EDE for the teams on 7 MP (#6 Falcons, #8 Hippopotami),
// page 63.
//
// FIDE values: the quoted cross-table cells on page 63, round 3: "8b3" for
// #6 and "6w1" for #8 → 3-1, so 2 MP for #6 and 0 MP for #8. Conclusion of
// the solution: #6 ranks above #8.
// ---------------------------------------------------------------------------

func TestFIDEExercise_43_EDE7MP(t *testing.T) {
	const nr = 43
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T6": 2, "T8": 0}
	for _, id := range []string{"T6", "T8"} {
		o5LogMissing(t, nr, "EDE(MP)", id, fide[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
}

// ---------------------------------------------------------------------------
// Exercise 44 – EDE for the teams on 6 MP (#7, #9, #13), page 63.
//
// FIDE values: page 63 – "The ranking, formulated based on the primary
// score (MP), sees #7 with two (MP) points, while #9 and #13 are tied at one
// point". Separate cross table (GP, page 63): 7-9 = B2/W2, 7-13 = W2/B2,
// 9-13 = not played; the GP totals per team are therefore 4 (#7), 2 (#9) and
// 2 (#13). Conclusion: EDE does not determine an order.
// ---------------------------------------------------------------------------

func TestFIDEExercise_44_EDE6MP(t *testing.T) {
	const nr = 44
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T7": 2, "T9": 1, "T13": 1}
	for _, id := range []string{"T7", "T9", "T13"} {
		o5LogMissing(t, nr, "EDE(MP)", id, fide[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
	// GP phase of [13.3.1]: the totals from the separate cross table (cells
	// B2/W2 on page 63).
	fideGP := map[string]float64{"T7": 4, "T9": 2, "T13": 2}
	for _, id := range []string{"T7", "T9", "T13"} {
		o5LogMissing(t, nr, "EDE(GP)", id, fideGP[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
}

// ---------------------------------------------------------------------------
// Exercise 45 – EDE for the teams on 10 MP, pages 63-64.
//
// FIDE values: the separate cross table on page 64 – the MP column (4 for
// each of the five teams), the GP column (8.0 / 8.0 / 8.0 / 8.5 / 7.5) and
// the second application with GP for the three remaining teams (GP column:
// #1 = 5.0; #2 = 5.5; #3 = 1.5). Final ranking: 4, 2, 1, 3, 5.
// ---------------------------------------------------------------------------

func TestFIDEExercise_45_EDE10MP(t *testing.T) {
	const nr = 45
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fideMP := map[string]float64{"T1": 4, "T2": 4, "T3": 4, "T4": 4, "T5": 4}
	fideGP := map[string]float64{"T1": 8.0, "T2": 8.0, "T3": 8.0, "T4": 8.5, "T5": 7.5}
	fideGP2 := map[string]float64{"T1": 5.0, "T2": 5.5, "T3": 1.5}

	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogMissing(t, nr, "EDE(MP)", id, fideMP[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogMissing(t, nr, "EDE(GP) phase 1", id, fideGP[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
	for _, id := range []string{"T1", "T2", "T3"} {
		o5LogMissing(t, nr, "EDE(GP) phase 2", id, fideGP2[id], "NOT_IMPLEMENTED: no team direct-encounter (EDE) variant")
	}
}

// ---------------------------------------------------------------------------
// Exercise 46 – Board count (BC) for #11 Koalas and #14 Narwhals, page 66.
//
// The published 2023 answer (page 66) is BC (#11) = 3.5 and BC (#14) = 6.5;
// #11 ranks above. The values below follow C.07:2026 instead: Article 12's
// preamble counts a team bye as a standard win on every board, so #14's
// round-7 pairing-allocated bye adds four board wins and lowers its count
// from −6.5 to −16.5 (BoardCount returns a negated value, where lower is
// better). The 2023 answers are quoted only for reference. The line-ups and
// board results are in the match table on page 66 (round 3, table 5): board
// 1 Kelpa 1-0 Neric, board 2 Kort ½-½ Negus, board 3 Koman ½-½ Neba, board 4
// Kontos 0-1 Negri. The round-3 match of o5TeamState carries these four
// boards in MatchData.Boards.
// ---------------------------------------------------------------------------

func TestFIDEExercise_46_BoardCount(t *testing.T) {
	const nr = 46
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	bc := o5Values(t, "board-count", state, scores)
	o5AssertCmp(t, nr, "BC", "T11", -3.5, bc["T11"])
	o5AssertCmp(t, nr, "BC", "T14", -16.5, bc["T14"])
}

// ---------------------------------------------------------------------------
// Exercise 47 – Top board results (TBR) for #11 and #14, pages 66-67.
//
// The published 2023 answer (page 67) is board-1 result 1 for #11 and 0 for
// #14, so #11 ranks above. The values below follow C.07:2026 instead: TBR
// encodes all board totals in board order (Article 12.2), and Article 12's
// preamble counts #14's round-7 pairing-allocated bye as a win on every
// board, so the encoded values (half points per digit) are 280 for #11 and
// 344 for #14 and rank #14
// above #11. The 2023 answers are quoted only for reference.
// ---------------------------------------------------------------------------

func TestFIDEExercise_47_TopBoardResults(t *testing.T) {
	const nr = 47
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	tbr := o5Values(t, "top-board-results", state, scores)
	o5AssertCmp(t, nr, "TBR", "T11", 280, tbr["T11"])
	o5AssertCmp(t, nr, "TBR", "T14", 344, tbr["T14"])
}

// ---------------------------------------------------------------------------
// Exercise 48 – Bottom board elimination (BBE) for #11 and #14, page 67.
//
// The published 2023 answer (page 67) is BBE (#11) = 2 and BBE (#14) = 1
// (lowest board struck off: 1 + ½ + ½ and 0 + ½ + ½ respectively); #11 ranks
// above. The values below follow C.07:2026 instead: BBE repeatedly excludes
// the bottom-most board (Article 12.3) and encodes each stage, and Article
// 12's preamble counts #14's round-7 pairing-allocated bye as a win on every
// board, so the encoded values (half points per digit) are 353 for #11 and
// 695 for #14 and rank
// #14 above #11. The 2023 answers are quoted only for reference.
// ---------------------------------------------------------------------------

func TestFIDEExercise_48_BottomBoardElimination(t *testing.T) {
	const nr = 48
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	bbe := o5Values(t, "bottom-board-elimination", state, scores)
	o5AssertCmp(t, nr, "BBE", "T11", 353, bbe["T11"])
	o5AssertCmp(t, nr, "BBE", "T14", 695, bbe["T14"])
}

// ---------------------------------------------------------------------------
// Exercise 49 – SSSC for all teams, pages 69-70.
//
// The "SSSC" column of the table on page 70 is the published 2023 answer; the
// "BH (MP)" column of the same table comes from exercise 35 (pages 54-55).
// Normalisation factor FN = 3 (page 69, [13.4.2.b]: 14 MP / 4 GP = 3.5 → 3).
// The BH(MP) term below follows C.07:2026 Article 16.4 as in exercise 35, so
// team #1 is 61 rather than the 2023 value 64.
//
// SSSC = GP + BH(MP)/FN: the registry has no SSSC variant (Article 13.4);
// only the BH(MP) term is verifiable.
// ---------------------------------------------------------------------------

func TestFIDEExercise_49_SSSC(t *testing.T) {
	const nr = 49
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// BH(MP) per team under C.07:2026 (2023 table answer for #1: 64); SSSC is the 2023 table column.
	fideBH := map[string]float64{
		"T1": 61, "T2": 57, "T3": 58, "T4": 56, "T5": 55, "T6": 46, "T7": 44,
		"T8": 41, "T9": 50, "T10": 44, "T11": 41, "T12": 41, "T13": 52, "T14": 36,
	}
	fideSSSC := map[string]float64{
		"T1": 38.8, "T2": 36.0, "T3": 35.3, "T4": 35.7, "T5": 36.3,
		"T6": 27.8, "T7": 26.2, "T8": 28.7, "T9": 31.2, "T10": 27.7,
		"T11": 25.2, "T12": 21.2, "T13": 28.8, "T14": 23.5,
	}

	bh := o5Values(t, "buchholz-mp", state, scores)
	for _, tm := range o5Teams {
		o5AssertCmp(t, nr, "BH(MP) for SSSC", tm.id, fideBH[tm.id], bh[tm.id])
	}
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "SSSC", tm.id, fideSSSC[tm.id],
			"NOT_IMPLEMENTED: no SSSC variant in the registry")
	}
}
