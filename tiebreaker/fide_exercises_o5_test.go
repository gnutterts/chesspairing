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
// A team is modelled as a single participant (PlayerEntry) and a team match
// as a single GameData; that is the same convention pairing/team uses. The
// primary score MP can therefore be expressed exactly (PointWin=2,
// PointDraw=1, PointLoss=0, PointBye=1, PointForfeitWin=2,
// PointForfeitLoss=0).
//
// The secondary score GP is NOT expressible: GameData.Result only knows
// 1-0, 0-1, ½-½ and the forfeit variants, whereas a team match has a partial
// score such as 2½-1½ or 3½-½. A TournamentState carries a single score per
// participant per game; a tournament with two parallel scores (MP and GP)
// does not fit in it. All exercises that use GP as the primary score or as a
// factor (34b, 37, 39, 40, 41, 49) therefore log NOT_EXPRESSIBLE.
//
// Board order (needed for BC/TBR/BBE, exercises 46-48) is present in
// GamePairing.Board but not in GameData: a played game carries no board
// number. Exercises 46-48 can therefore only be rebuilt at board level as
// loose games without board identification.

import (
	"context"
	"math"
	"sort"
	"strconv"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

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
	win := chesspairing.ResultWhiteWins
	loss := chesspairing.ResultBlackWins
	draw := chesspairing.ResultDraw

	// match creates a team match with the given result.
	match := func(white, black string, result chesspairing.GameResult) chesspairing.GameData {
		return chesspairing.GameData{WhiteID: white, BlackID: black, Result: result}
	}

	state := &chesspairing.TournamentState{
		Players: make([]chesspairing.PlayerEntry, 0, len(o5Teams)),
		Rounds: []chesspairing.RoundData{
			{Number: 1, Games: []chesspairing.GameData{
				match("T1", "T8", win),   // 8w2½ / 1b1½
				match("T9", "T2", loss),  // 9b3 / 2w1
				match("T3", "T10", win),  // 10w2½ / 3b1½
				match("T11", "T4", loss), // 11b3 / 4w1
				match("T5", "T12", win),  // 12w4 / 5b0
				match("T6", "T13", draw), // 13b2 / 6w2
				match("T7", "T14", draw), // 14w2 / 7b2
			}},
			{Number: 2, Games: []chesspairing.GameData{
				match("T4", "T1", win),    // 4b1½ / 1w2½
				match("T2", "T3", win),    // 3w4 / 2b0
				match("T6", "T5", loss),   // 6b3 / 5w1
				match("T8", "T7", draw),   // 8b2 / 7w2
				match("T12", "T9", loss),  // 12b4 / 9w0
				match("T10", "T11", loss), // 11w1½ / 10b2½
				match("T14", "T13", loss), // 14b2½ / 13w1½
			}},
			{Number: 3, Games: []chesspairing.GameData{
				match("T1", "T7", win),    // 7w3 / 1b1
				match("T5", "T2", win),    // 5b0 / 2w4
				match("T3", "T9", win),    // 9w3 / 3b1
				match("T13", "T4", loss),  // 13b3 / 4w1
				match("T8", "T6", loss),   // 8b3 / 6w1
				match("T12", "T10", loss), // 12b2½ / 10w1½
				match("T11", "T14", draw), // 14w2 / 11b2
			}},
			{Number: 4, Games: []chesspairing.GameData{
				match("T2", "T1", loss),  // 2b2½ / 1w1½
				match("T6", "T3", loss),  // 6b2½ / 3w1½
				match("T4", "T5", win),   // 5w3½ / 4b½
				match("T7", "T10", win),  // 10w2½ / 7b1½
				match("T12", "T8", loss), // 12b3½ / 8w½
				match("T9", "T14", win),  // 14w3 / 9b1
				match("T13", "T11", win), // 13b1½ / 11w2½
			}},
			{Number: 5, Games: []chesspairing.GameData{
				match("T1", "T5", loss),  // 5w1½ / 1b2½
				match("T2", "T13", win),  // 13w3 / 2b1
				match("T3", "T4", win),   // 4w3 / 3b1
				match("T11", "T6", loss), // 11b2½ / 6w1½
				match("T9", "T7", draw),  // 9b2 / 7w2
				match("T10", "T8", draw), // 10b2 / 8w2
			}, Byes: []chesspairing.ByeEntry{
				{PlayerID: "T12", Type: chesspairing.ByePAB},  // PAB: 1 MP, 2 GP
				{PlayerID: "T14", Type: chesspairing.ByeHalf}, // HPB: 1 MP, 2 GP
			}},
			{Number: 6, Games: []chesspairing.GameData{
				// -F / +F: team 6 loses, team 1 wins by forfeit.
				{WhiteID: "T6", BlackID: "T1", Result: chesspairing.ResultForfeitBlackWins, IsForfeit: true},
				match("T4", "T2", loss),   // 4b2½ / 2w1½
				match("T5", "T3", loss),   // 5b3½ / 3w½
				match("T13", "T7", draw),  // 7b2 / 13w2
				match("T8", "T9", draw),   // 9w2 / 8b2
				match("T14", "T10", loss), // 14b3 / 10w1
				match("T11", "T12", draw), // 12w2 / 11b2
			}},
			{Number: 7, Games: []chesspairing.GameData{
				match("T3", "T1", loss),  // 3b2½ / 1w1½
				match("T10", "T2", loss), // 10b3 / 2w1
				match("T9", "T4", loss),  // 9b2½ / 4w1½
				match("T13", "T5", loss), // 13b3½ / 5w½
				match("T6", "T12", win),  // 12w2½ / 6b1½
				match("T8", "T11", win),  // 11w3 / 8b1
			}, Byes: []chesspairing.ByeEntry{
				{PlayerID: "T7", Type: chesspairing.ByeZero}, // ZPB: 0 MP, 0 GP
				{PlayerID: "T14", Type: chesspairing.ByePAB}, // PAB: 1 MP, 2 GP
			}},
		},
	}

	// Teams receive no rating in the exercise; Rating stays 0.
	for _, tm := range o5Teams {
		state.Players = append(state.Players, chesspairing.PlayerEntry{
			ID:          tm.id,
			DisplayName: tm.name,
		})
	}
	return state
}

// o5MPScorer is the scorer with the point values of the team tournament:
// 2-1-0 for match points, PAB/HPB = 1 MP, ZPB = 0 MP, +F = 2 MP,
// -F = 0 MP (page 5 and page 8).
func o5MPScorer() *standard.Scorer {
	return standard.New(standard.Options{
		PointWin:         chesspairing.Float64Ptr(2),
		PointDraw:        chesspairing.Float64Ptr(1),
		PointLoss:        chesspairing.Float64Ptr(0),
		PointBye:         chesspairing.Float64Ptr(1),
		PointForfeitWin:  chesspairing.Float64Ptr(2),
		PointForfeitLoss: chesspairing.Float64Ptr(0),
	})
}

// o5MPScores computes the match points (MP) of all teams.
func o5MPScores(t *testing.T, state *chesspairing.TournamentState) []chesspairing.PlayerScore {
	t.Helper()
	scores, err := o5MPScorer().Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard (MP): %v", err)
	}
	return scores
}

// o5AssertMP is the hard assertion of the transcription: the match points
// computed with scoring/standard must equal the MP column of the cross
// table (page 6). GP cannot be verified: the partial score per match is not
// expressible (see the file comment).
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

// o5LogCmp logs one comparison line per team; FIDE and code are compared
// with tolerance 0.001. A difference is a DISCREPANCY (data, not a test
// failure): the test only fails when the transcription of the tournament
// itself is wrong.
func o5LogCmp(t *testing.T, nr int, tb, teamID string, fide, code float64) {
	t.Helper()
	status := "OK"
	if math.Abs(fide-code) > 0.001 {
		status = "DISCREPANCY"
	}
	t.Logf("EXERCISE %d | %s | %s | FIDE=%v | CODE=%v | %s", nr, tb, o5Label(teamID), fide, code, status)
}

// o5LogMissing logs a line for a tiebreak the code does not know
// (NOT_IMPLEMENTED) or for a value the data model cannot express
// (NOT_EXPRESSIBLE, with the reason in marker).
func o5LogMissing(t *testing.T, nr int, tb, teamID string, fide float64, marker string) {
	t.Helper()
	t.Logf("EXERCISE %d | %s | %s | FIDE=%v | CODE=missing | %s", nr, tb, o5Label(teamID), fide, marker)
}

// o5GPMarker is the standard reason why GP-based values cannot be computed.
const o5GPMarker = "NOT_EXPRESSIBLE: GP partial score per match (e.g. 2.5-1.5) does not fit in GameData.Result"

// o5PlacesByScore ranks teams by score (descending) and assigns league
// places ("1224"); equal scores share a place. IDs are sorted
// alphanumerically by team number when scores are equal.
func o5PlacesByScore(scores []chesspairing.PlayerScore) map[string]int {
	type row struct {
		id    string
		num   int
		score float64
	}
	rows := make([]row, 0, len(scores))
	for _, ps := range scores {
		n := 0
		for i, tm := range o5Teams {
			if tm.id == ps.PlayerID {
				n = i + 1
			}
		}
		rows = append(rows, row{id: ps.PlayerID, num: n, score: ps.Score})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].num < rows[j].num
	})
	places := make(map[string]int, len(rows))
	for i, r := range rows {
		if i > 0 && math.Abs(r.score-rows[i-1].score) < 0.001 {
			places[r.id] = places[rows[i-1].id]
			continue
		}
		places[r.id] = i + 1
	}
	return places
}

// ---------------------------------------------------------------------------
// Exercise 34 – Match points versus game points (MPVGP), pages 52-53.
//
// FIDE values: the two final standings on page 53. First table (a): {MP, GP}
// with the order 5, 1, 2, 4, 3, 8, 6, 9, 7, 13, 10, 11, 14, 12.
// Second table (b): {GP, MP} with the order 5, 1, 2, 4, 3, 8, 9, 10, 6, 7,
// 13, 11, 14, 12.
//
// The MPVGP tiebreak (secondary score as the first criterion after the
// primary score) is missing from the registry, and the secondary score GP is
// moreover not expressible. Part (a) is partially verifiable: the primary
// score MP can be rebuilt exactly, so the place is determined as far as MP
// alone suffices. Part (b) and the GP final scores cannot be computed at
// all.
// ---------------------------------------------------------------------------

func TestFIDEExercise_34_MPVGPStand(t *testing.T) {
	const nr = 34
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// Places from the {MP, GP} table (page 53, column "#").
	fideA := map[string]int{
		"T5": 1, "T1": 2, "T2": 3, "T4": 4, "T3": 5, "T8": 6, "T6": 7,
		"T9": 8, "T7": 9, "T13": 10, "T10": 11, "T11": 12, "T14": 13, "T12": 14,
	}
	// Places from the {GP, MP} table (page 53, column "#").
	fideB := map[string]int{
		"T5": 1, "T1": 2, "T2": 3, "T4": 4, "T3": 5, "T8": 6, "T9": 7,
		"T10": 8, "T6": 9, "T7": 10, "T13": 11, "T11": 12, "T14": 13, "T12": 14,
	}

	// GP final scores (cross table p.6, reprinted in both tables on p.53):
	// the secondary score cannot be computed with scoring/standard.
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "GP final score", tm.id, tm.gp, o5GPMarker)
	}

	places := o5PlacesByScore(scores)
	for _, tm := range o5Teams {
		fide := float64(fideA[tm.id])
		code := float64(places[tm.id])
		if math.Abs(fide-code) < 0.001 {
			o5LogCmp(t, nr, "MPVGP(a) place {MP,GP}", tm.id, fide, code)
			continue
		}
		// Places within an MP group derive from GP: not expressible.
		o5LogMissing(t, nr, "MPVGP(a) place {MP,GP}", tm.id, fide, o5GPMarker+
			"; CODE place on MP alone = "+strconv.Itoa(places[tm.id]))
	}
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "MPVGP(b) place {GP,MP}", tm.id, float64(fideB[tm.id]), o5GPMarker)
	}
}

// ---------------------------------------------------------------------------
// Exercise 35 – Buchholz (total) on match points, pages 53-55.
//
// FIDE values: the "BH" column of the table on page 54 (teams 1, 3, 2, 4, 5,
// 6, 8) and page 55 (teams 13, 9, 7, 10, 11, 14, 12).
//
// Relevant articles from the solution: [16.4] (forfeit and bye count as a
// match against a dummy with the same score and the same result as the team
// itself) and [16.2.5]/[16.3.2] (the ZPB of #7 in the last round counts as a
// draw for the opponents: "Adjusted MP" of #7 is 7 instead of 6, see page
// 54).
// ---------------------------------------------------------------------------

func TestFIDEExercise_35_BuchholzMP(t *testing.T) {
	const nr = 35
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// BH (MP) per team, table pages 54-55.
	fide := map[string]float64{
		"T1": 64, "T2": 57, "T3": 58, "T4": 56, "T5": 55, "T6": 46, "T7": 44,
		"T8": 41, "T9": 50, "T10": 44, "T11": 41, "T12": 41, "T13": 52, "T14": 36,
	}

	code := o5Values(t, "buchholz", state, scores)
	for _, tm := range o5Teams {
		o5LogCmp(t, nr, "BH(MP)", tm.id, fide[tm.id], code[tm.id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 36 – Buchholz Cut-1 on match points, page 55.
//
// FIDE values: the "BH-C1" column of the table on page 55.
//
// Article from the solution: [16.5] – the contribution of a voluntary
// absence (forfeit loss #6 round 6, HPB #14 round 5, ZPB #7 round 7) is
// struck off first, even when that contribution is larger than the lowest
// value.
// ---------------------------------------------------------------------------

func TestFIDEExercise_36_BuchholzCut1MP(t *testing.T) {
	const nr = 36
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// BH-C1 (MP) per team, table page 55.
	fide := map[string]float64{
		"T1": 57, "T2": 52, "T3": 53, "T4": 52, "T5": 53, "T6": 39, "T7": 38,
		"T8": 39, "T9": 48, "T10": 42, "T11": 39, "T12": 39, "T13": 48, "T14": 32,
	}

	code := o5Values(t, "buchholz-cut1", state, scores)
	for _, tm := range o5Teams {
		o5LogCmp(t, nr, "BH-C1(MP)", tm.id, fide[tm.id], code[tm.id])
	}
}

// ---------------------------------------------------------------------------
// Exercise 37 – Buchholz Cut-1 on game points (GP as primary score), page 56.
//
// FIDE values: the "BH GP" and "BH-C1" columns of the table on page 56.
//
// GP is the primary score of this exercise; it is not expressible in the
// data model (see the file comment). The MP column of the same table is
// verified in o5AssertMP.
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
		o5LogMissing(t, nr, "BH(GP)", tm.id, fideBH[tm.id], o5GPMarker)
	}
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "BH-C1(GP)", tm.id, fideC1[tm.id], o5GPMarker)
	}
}

// ---------------------------------------------------------------------------
// Exercise 38 – EMMSB and EMMSB Cut-1 for the teams on 10 MP, pages 57-58.
//
// FIDE values: the "EMMSB" column and the "EMMSB Cut-1" column of the tables
// on page 57 (team #1) and page 58 (teams #2 through #5).
//
// EMMSB = Σ (opponent MP × own MP), so with 2-1-0 weighting [16]. The code
// only knows sonneborn-berger (individual, weighting 1-½-0) and has no
// Cut-1 variant of it; moreover the code does not count forfeit matches and
// byes as contributions, while [16.4] counts them at face value and
// [16.3.2] prescribes the adjusted score of #7 (7 MP).
// ---------------------------------------------------------------------------

func TestFIDEExercise_38_EMMSBCut1(t *testing.T) {
	const nr = 38
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// Teams on 10 MP; EMMSB / EMMSB-C1 from pages 57-58.
	fideSB := map[string]float64{"T1": 88, "T2": 74, "T3": 76, "T4": 72, "T5": 70}
	fideC1 := map[string]float64{"T1": 74, "T2": 64, "T3": 66, "T4": 64, "T5": 66}

	sb := o5Values(t, "sonneborn-berger", state, scores)
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogCmp(t, nr, "EMMSB", id, fideSB[id], sb[id])
	}
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogMissing(t, nr, "EMMSB-C1", id, fideC1[id],
			"NOT_IMPLEMENTED: no Sonneborn-Berger Cut-1 in the registry")
	}
}

// ---------------------------------------------------------------------------
// Exercise 39 – EGMSB for the teams on 10 MP, page 59.
//
// FIDE values: the "EGMSB" column of the tables on page 59 (158.0 / 144.0 /
// 150.0 / 146.0 / 132.0). EGMSB = Σ (opponent GP × own MP): the GP of the
// opponent is not expressible.
// ---------------------------------------------------------------------------

func TestFIDEExercise_39_EGMSB(t *testing.T) {
	const nr = 39
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T1": 158.0, "T2": 144.0, "T3": 150.0, "T4": 146.0, "T5": 132.0}
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogMissing(t, nr, "EGMSB", id, fide[id], o5GPMarker+"; moreover the EGMSB variant is missing")
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
	for _, id := range []string{"T7", "T9", "T13"} {
		o5LogMissing(t, nr, "EMGSB", id, fide[id], o5GPMarker+"; moreover the EMGSB variant is missing")
	}
}

// ---------------------------------------------------------------------------
// Exercise 41 – EGGSB (and EGGSB Cut-1) for the teams on 7 MP, pages 60-61.
//
// FIDE values: the "EGGSB" column of the tables on page 61 (#6 = 157.50;
// #8 = 181.50). Because the two values differ there is no tie among the
// remainder and the Cut-1 variant is not addressed. EGGSB = Σ (opponent GP ×
// own GP): both factors are not expressible.
// ---------------------------------------------------------------------------

func TestFIDEExercise_41_EGGSB(t *testing.T) {
	const nr = 41
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	fide := map[string]float64{"T6": 157.50, "T8": 181.50}
	for _, id := range []string{"T6", "T8"} {
		o5LogMissing(t, nr, "EGGSB", id, fide[id], o5GPMarker+"; moreover the EGGSB variant is missing")
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
// [13.3]/[13.3.1]: the code only knows the individual direct encounter on a
// 1-½-0 basis and no EDE with MP/GP phases; incomparable scales.
// ---------------------------------------------------------------------------

func TestFIDEExercise_42_EDE4MP(t *testing.T) {
	const nr = 42
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	de := o5Values(t, "direct-encounter", state, scores)
	for _, id := range []string{"T11", "T14"} {
		o5LogCmp(t, nr, "EDE(MP)", id, 1, de[id])
	}
	for _, id := range []string{"T11", "T14"} {
		o5LogMissing(t, nr, "EDE(GP)", id, 2, o5GPMarker)
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
	de := o5Values(t, "direct-encounter", state, scores)
	for _, id := range []string{"T6", "T8"} {
		o5LogCmp(t, nr, "EDE(MP)", id, fide[id], de[id])
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
	de := o5Values(t, "direct-encounter", state, scores)
	for _, id := range []string{"T7", "T9", "T13"} {
		o5LogCmp(t, nr, "EDE(MP)", id, fide[id], de[id])
	}
	// GP phase of [13.3.1]: the totals from the separate cross table (cells
	// B2/W2 on page 63) cannot be recomputed.
	fideGP := map[string]float64{"T7": 4, "T9": 2, "T13": 2}
	for _, id := range []string{"T7", "T9", "T13"} {
		o5LogMissing(t, nr, "EDE(GP)", id, fideGP[id], o5GPMarker)
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

	de := o5Values(t, "direct-encounter", state, scores)
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogCmp(t, nr, "EDE(MP)", id, fideMP[id], de[id])
	}
	for _, id := range []string{"T1", "T2", "T3", "T4", "T5"} {
		o5LogMissing(t, nr, "EDE(GP) phase 1", id, fideGP[id], o5GPMarker)
	}
	for _, id := range []string{"T1", "T2", "T3"} {
		o5LogMissing(t, nr, "EDE(GP) phase 2", id, fideGP2[id], o5GPMarker)
	}
}

// ---------------------------------------------------------------------------
// Exercise 46 – Board count (BC) for #11 Koalas and #14 Narwhals, page 66.
//
// FIDE values: page 66 – BC (#11) = 3.5 and BC (#14) = 6.5; #11 ranks above.
// The line-ups and board results are in the match table on page 66 (round 3,
// table 5): board 1 Kelpa 1-0 Neric, board 2 Kort ½-½ Negus, board 3 Koman
// ½-½ Neba, board 4 Kontos 0-1 Negri.
//
// BC does not exist in the registry, and GameData has no board number.
// The transcription is verified on the individual player points and on the
// match GP (2-2) with scoring/standard (1-½-0).
// ---------------------------------------------------------------------------

// o5BoardState rebuilds the match Koalas (#11) - Narwhals (#14) from round 3
// at board level. The colours follow from the individual cross table
// (pages 7-8): 33w1 (Kelpa white), 43b= (Kort black), 52w= (Koman white),
// 62w1 (Negri white). Board numbers are not available in GameData.
func o5BoardState() *chesspairing.TournamentState {
	return &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "P33", DisplayName: "Kris Kelpa (Koalas board 1)", Rating: 1952},
			{ID: "P43", DisplayName: "Kelly Kort (Koalas board 2)", Rating: 1902},
			{ID: "P52", DisplayName: "Kirk Koman (Koalas board 3)", Rating: 1857},
			{ID: "P73", DisplayName: "Kurt Kontos (Koalas board 4)", Rating: 1752},
			{ID: "P34", DisplayName: "Nikola Neric (Narwhals board 1)", Rating: 1947},
			{ID: "P55", DisplayName: "Noah Negus (Narwhals board 2)", Rating: 1842},
			{ID: "P37", DisplayName: "Nicola Neba (Narwhals board 3)", Rating: 1932},
			{ID: "P62", DisplayName: "Nuccio Negri (Narwhals board 4)", Rating: 1807},
		},
		Rounds: []chesspairing.RoundData{
			{Number: 3, Games: []chesspairing.GameData{
				{WhiteID: "P33", BlackID: "P34", Result: chesspairing.ResultWhiteWins},
				{WhiteID: "P55", BlackID: "P43", Result: chesspairing.ResultDraw},
				{WhiteID: "P52", BlackID: "P37", Result: chesspairing.ResultDraw},
				{WhiteID: "P62", BlackID: "P73", Result: chesspairing.ResultWhiteWins},
			}},
		},
	}
}

// o5BoardScores computes the individual player points (1-½-0) of the board
// games and verifies them against the match GP 2-2 (page 66).
func o5BoardScores(t *testing.T, nr int, state *chesspairing.TournamentState) {
	t.Helper()
	scores, err := standard.New(standard.Options{}.WithDefaults()).Score(context.Background(), state)
	if err != nil {
		t.Fatalf("scoring/standard (board games): %v", err)
	}
	got := make(map[string]float64, len(scores))
	for _, ps := range scores {
		got[ps.PlayerID] = ps.Score
	}
	want := map[string]float64{
		"P33": 1, "P43": 0.5, "P52": 0.5, "P73": 0,
		"P34": 0, "P55": 0.5, "P37": 0.5, "P62": 1,
	}
	for id, w := range want {
		if math.Abs(got[id]-w) > 0.001 {
			t.Errorf("EXERCISE %d: transcription error: %s = %v points, FIDE board table = %v", nr, id, got[id], w)
		}
	}
	var koalas, narwhals float64
	for id, v := range got {
		switch id {
		case "P33", "P43", "P52", "P73":
			koalas += v
		case "P34", "P55", "P37", "P62":
			narwhals += v
		}
	}
	if math.Abs(koalas-2) > 0.001 || math.Abs(narwhals-2) > 0.001 {
		t.Errorf("EXERCISE %d: transcription error: match GP = %v-%v, FIDE (p.66) = 2-2", nr, koalas, narwhals)
	}
}

func TestFIDEExercise_46_BoardCount(t *testing.T) {
	const nr = 46
	state := o5BoardState()
	o5BoardScores(t, nr, state)

	o5LogMissing(t, nr, "BC", "T11", 3.5,
		"NOT_IMPLEMENTED: board count is missing; moreover GameData carries no board number")
	o5LogMissing(t, nr, "BC", "T14", 6.5,
		"NOT_IMPLEMENTED: board count is missing; moreover GameData carries no board number")
}

// ---------------------------------------------------------------------------
// Exercise 47 – Top board results (TBR) for #11 and #14, pages 66-67.
//
// FIDE value: page 67 – "On the first board, team #11 won, thus prevailing
// over the opponent"; the board result 1-0 on board 1 is in the table on
// page 67. As a number: board-1 result 1 for #11 and 0 for #14.
// ---------------------------------------------------------------------------

func TestFIDEExercise_47_TopBoardResults(t *testing.T) {
	const nr = 47
	state := o5BoardState()
	o5BoardScores(t, nr, state)

	o5LogMissing(t, nr, "TBR(board 1)", "T11", 1,
		"NOT_IMPLEMENTED: top board results is missing; moreover GameData carries no board number")
	o5LogMissing(t, nr, "TBR(board 1)", "T14", 0,
		"NOT_IMPLEMENTED: top board results is missing; moreover GameData carries no board number")
}

// ---------------------------------------------------------------------------
// Exercise 48 – Bottom board elimination (BBE) for #11 and #14, page 67.
//
// FIDE values: page 67 – BBE (#11) = 2 and BBE (#14) = 1 (lowest board
// struck off: 1 + ½ + ½ and 0 + ½ + ½ respectively); #11 ranks above.
// ---------------------------------------------------------------------------

func TestFIDEExercise_48_BottomBoardElimination(t *testing.T) {
	const nr = 48
	state := o5BoardState()
	o5BoardScores(t, nr, state)

	o5LogMissing(t, nr, "BBE", "T11", 2,
		"NOT_IMPLEMENTED: bottom board elimination is missing; moreover GameData carries no board number")
	o5LogMissing(t, nr, "BBE", "T14", 1,
		"NOT_IMPLEMENTED: bottom board elimination is missing; moreover GameData carries no board number")
}

// ---------------------------------------------------------------------------
// Exercise 49 – SSSC for all teams, pages 69-70.
//
// FIDE values: the "SSSC" column of the table on page 70; the "BH (MP)"
// column of the same table comes from exercise 35 (pages 54-55) and can be
// computed. Normalisation factor FN = 3 (page 69, [13.4.2.b]: 14 MP / 4 GP =
// 3.5 → 3).
//
// SSSC = GP + BH(MP)/FN: the GP term is not expressible and SSSC is missing
// from the registry.
// ---------------------------------------------------------------------------

func TestFIDEExercise_49_SSSC(t *testing.T) {
	const nr = 49
	state := o5TeamState()
	scores := o5MPScores(t, state)
	o5AssertMP(t, nr, scores)

	// Table page 70: BH(MP) and SSSC.
	fideBH := map[string]float64{
		"T1": 64, "T2": 57, "T3": 58, "T4": 56, "T5": 55, "T6": 46, "T7": 44,
		"T8": 41, "T9": 50, "T10": 44, "T11": 41, "T12": 41, "T13": 52, "T14": 36,
	}
	fideSSSC := map[string]float64{
		"T1": 38.8, "T2": 36.0, "T3": 35.3, "T4": 35.7, "T5": 36.3,
		"T6": 27.8, "T7": 26.2, "T8": 28.7, "T9": 31.2, "T10": 27.7,
		"T11": 25.2, "T12": 21.2, "T13": 28.8, "T14": 23.5,
	}

	bh := o5Values(t, "buchholz", state, scores)
	for _, tm := range o5Teams {
		o5LogCmp(t, nr, "BH(MP) for SSSC", tm.id, fideBH[tm.id], bh[tm.id])
	}
	for _, tm := range o5Teams {
		o5LogMissing(t, nr, "SSSC", tm.id, fideSSSC[tm.id],
			"NOT_IMPLEMENTED: SSSC is missing; "+o5GPMarker)
	}
}
