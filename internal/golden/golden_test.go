// Package golden preserves the library's current externally observable output.
// Set GOLDEN_UPDATE=1 and run the Golden tests to review and accept changes.
package golden_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/factory"
	"github.com/gnutterts/chesspairing/standings"
	"github.com/gnutterts/chesspairing/trf"
)

var allResults = []cp.GameResult{
	cp.ResultWhiteWins, cp.ResultBlackWins, cp.ResultDraw, cp.ResultPending,
	cp.ResultForfeitWhiteWins, cp.ResultForfeitBlackWins, cp.ResultDoubleForfeit,
}

var allByes = []cp.ByeType{
	cp.ByePAB, cp.ByeHalf, cp.ByeZero, cp.ByeAbsent, cp.ByeExcused, cp.ByeClubCommitment,
}

type tournamentCase struct {
	name  string
	seed1 uint64
	seed2 uint64
	count int
	equal bool
}

var tournamentCases = []tournamentCase{
	{name: "even", seed1: 101, seed2: 1001, count: 16},
	{name: "odd", seed1: 202, seed2: 2002, count: 15},
	{name: "features", seed1: 303, seed2: 3003, count: 18, equal: true},
}

func TestGoldenPairings(t *testing.T) {
	for _, system := range factory.PairerNames() {
		for _, tc := range tournamentCases {
			opts := pairingOptions(system, tc.name)
			name := system + "_" + tc.name
			t.Run(name, func(t *testing.T) {
				golden(t, filepath.Join("testdata", "pairing", name+".txt"), simulatePairing(system, opts, tc))
			})
		}
	}
	variants := []struct {
		name string
		opts map[string]any
	}{
		{"no-repeats", map[string]any{"allowRepeatPairings": false}},
		{"short-gap", map[string]any{"allowRepeatPairings": true, "minRoundsBetweenRepeats": 1}},
		{"custom-ranking", map[string]any{"valueNumberBase": 40, "valueNumberStep": 2, "selfVictory": false}},
	}
	for _, v := range variants {
		for i, tc := range tournamentCases {
			v, tc := v, tc
			name := "keizer_" + v.name
			if i > 0 {
				name += "_" + tc.name
			}
			t.Run(name, func(t *testing.T) {
				golden(t, filepath.Join("testdata", "pairing", name+".txt"), simulatePairing("keizer", v.opts, tc))
			})
		}
	}
}

func TestGoldenScoring(t *testing.T) {
	variants := map[string][]struct {
		name string
		opts map[string]any
	}{
		"standard": {
			{"default", nil},
			{"custom-byes", map[string]any{"pointBye": .75, "pointAbsent": -.25, "pointExcused": .4, "pointClubCommitment": .6}},
		},
		"football": {{"default", nil}, {"custom", map[string]any{"pointWin": 4.0, "pointDraw": 2.0, "pointLoss": 1.0}}},
		"keizer": {
			{"default", nil},
			{"frozen", map[string]any{"frozen": true}},
			{"fixed-absence-limit", map[string]any{"absentFixedValue": 7, "excusedAbsentFixedValue": 9, "clubCommitmentFixedValue": 11, "absenceLimit": 1}},
			{"fixed-byes", map[string]any{"byeFixedValue": 8, "halfByeFixedValue": 6, "zeroByeFixedValue": 2}},
			{"loss-fraction", map[string]any{"lossFraction": 1.0 / 6.0}},
			{"absence-decay", map[string]any{"absenceDecay": true, "absenceLimit": 0}},
			{"late-join-handicap", map[string]any{"lateJoinHandicap": 4.5}},
			{"forfeit-fractions", map[string]any{"forfeitWinFraction": .8, "forfeitLossFraction": .2, "doubleForfeitFraction": .1}},
			{"no-self-victory", map[string]any{"selfVictory": false}},
			{"base-step", map[string]any{"valueNumberBase": 50, "valueNumberStep": 3}},
		},
	}
	for _, scorerName := range factory.ScorerNames() {
		scorerVariants, ok := variants[scorerName]
		if !ok || len(scorerVariants) == 0 {
			t.Fatalf("scorer %q has no golden variants", scorerName)
		}
		for _, variant := range scorerVariants {
			for _, tc := range tournamentCases {
				name := scorerName + "_" + variant.name + "_" + tc.name
				t.Run(name, func(t *testing.T) {
					state := completedTournament(tc)
					scorer, err := factory.NewScorer(scorerName, variant.opts)
					if err != nil {
						t.Fatal(err)
					}
					golden(t, filepath.Join("testdata", "scoring", name+".txt"), scoringText(scorer, state))
				})
			}
		}
	}
}

func TestGoldenTieBreakers(t *testing.T) {
	for _, id := range factory.TieBreakerIDs() {
		for _, tc := range tournamentCases {
			name := id + "_" + tc.name
			t.Run(name, func(t *testing.T) {
				state := completedTournament(tc)
				scorer, _ := factory.NewScorer("standard", nil)
				scores, err := scorer.Score(context.Background(), state)
				if err != nil {
					t.Fatal(err)
				}
				tb, err := factory.NewTieBreaker(id)
				if err != nil {
					t.Fatal(err)
				}
				values, err := tb.Compute(context.Background(), state, scores)
				var b strings.Builder
				fmt.Fprintf(&b, "id=%s\nname=%s\nerror=%v\n", tb.ID(), tb.Name(), err)
				for _, v := range values {
					fmt.Fprintf(&b, "%s %s\n", v.PlayerID, fl(v.Value))
				}
				golden(t, filepath.Join("testdata", "tiebreak", name+".txt"), []byte(b.String()))
			})
		}
	}
}

func TestGoldenStandings(t *testing.T) {
	for _, tc := range tournamentCases {
		t.Run(tc.name, func(t *testing.T) {
			state := completedTournament(tc)
			scorer, _ := factory.NewScorer("standard", nil)
			ids := factory.TieBreakerIDs()
			rows, err := standings.BuildByID(context.Background(), state, scorer, ids)
			var b strings.Builder
			fmt.Fprintf(&b, "error=%v\n", err)
			for _, row := range rows {
				fmt.Fprintf(&b, "%d %s %q score=%s games=%d w=%d d=%d l=%d", row.Rank, row.PlayerID, row.DisplayName, fl(row.Score), row.GamesPlayed, row.Wins, row.Draws, row.Losses)
				for _, v := range row.TieBreakers {
					fmt.Fprintf(&b, " %s=%s", v.ID, fl(v.Value))
				}
				b.WriteByte('\n')
			}
			golden(t, filepath.Join("testdata", "standings", tc.name+".txt"), []byte(b.String()))
		})
	}
}

func TestGoldenTRF(t *testing.T) {
	for _, path := range trfFixtures(t) {
		path := path
		name := strings.NewReplacer("/", "__", "\\", "__", ".", "_").Replace(path)
		t.Run(name, func(t *testing.T) {
			input, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			doc, readErr := trf.Read(bytes.NewReader(input))
			var out bytes.Buffer
			var writeErr error
			if doc != nil {
				writeErr = trf.Write(&out, doc)
			}
			var stateJSON, roundJSON []byte
			var stateErr, roundErr error
			if doc != nil {
				state, err := doc.ToTournamentState()
				stateErr = err
				if err == nil {
					stateJSON, err = json.MarshalIndent(state, "", "  ")
					if err != nil {
						stateJSON = []byte(fmt.Sprintf("JSON MARSHAL ERROR: %v", err))
					}
				}
				doc2, err := trf.Read(bytes.NewReader(out.Bytes()))
				roundErr = err
				if err == nil {
					state2, err := doc2.ToTournamentState()
					roundErr = err
					if err == nil {
						roundJSON, err = json.MarshalIndent(state2, "", "  ")
						if err != nil {
							roundJSON = []byte(fmt.Sprintf("JSON MARSHAL ERROR: %v", err))
						}
					}
				}
			}
			var b bytes.Buffer
			fmt.Fprintf(&b, "READ ERROR: %v\nSTATE ERROR: %v\nSTATE JSON:\n%s\nWRITE ERROR: %v\nWRITE BYTES:\n", readErr, stateErr, stateJSON, writeErr)
			b.Write(out.Bytes())
			fmt.Fprintf(&b, "ROUNDTRIP ERROR: %v\nROUNDTRIP STATE JSON:\n%s\n", roundErr, roundJSON)
			golden(t, filepath.Join("testdata", "trf", name+".txt"), b.Bytes())
		})
	}
}

func pairingOptions(system, scenario string) map[string]any {
	opts := map[string]any{"topSeedColor": "black", "totalRounds": 7}
	if scenario == "features" && (system == "dutch" || system == "burstein") {
		opts["acceleration"] = "baku"
	}
	return opts
}

func simulatePairing(system string, opts map[string]any, tc tournamentCase) []byte {
	state := baseTournament(tc)
	rng := rand.New(rand.NewPCG(tc.seed1, tc.seed2))
	pairer, err := factory.NewPairer(system, opts)
	var b strings.Builder
	fmt.Fprintf(&b, "system=%s scenario=%s players=%d options=%s factory-error=%v\n", system, tc.name, tc.count, stableOptions(opts), err)
	if err != nil {
		return []byte(b.String())
	}
	for round := 1; round <= 5; round++ {
		state.CurrentRound = round
		state.PreAssignedByes = nil
		if tc.name == "features" && round == 2 {
			for i, bt := range allByes {
				state.PreAssignedByes = append(state.PreAssignedByes, cp.ByeEntry{PlayerID: state.Players[len(state.Players)-1-i].ID, Type: bt})
			}
		}
		if tc.name == "features" && round >= 4 {
			n := 3
			state.Players[1].WithdrawnAfterRound = &n
		}
		result, pairErr := pairer.Pair(context.Background(), state)
		fmt.Fprintf(&b, "round=%d error=%v\n", round, pairErr)
		if result != nil {
			for _, p := range result.Pairings {
				fmt.Fprintf(&b, "  board=%d white=%s black=%s\n", p.Board, p.WhiteID, p.BlackID)
			}
			for _, bye := range result.Byes {
				fmt.Fprintf(&b, "  bye=%s type=%s\n", bye.PlayerID, bye.Type)
			}
			for _, note := range result.Notes {
				fmt.Fprintf(&b, "  note=%s\n", note)
			}
		}
		if pairErr != nil || result == nil {
			break
		}
		rd := cp.RoundData{Number: round, Byes: append([]cp.ByeEntry(nil), result.Byes...)}
		for i, p := range result.Pairings {
			gr := allResults[(rng.IntN(len(allResults))+round+i)%len(allResults)]
			rd.Games = append(rd.Games, cp.GameData{WhiteID: p.WhiteID, BlackID: p.BlackID, Result: gr, IsForfeit: gr.IsForfeit()})
		}
		state.Rounds = append(state.Rounds, rd)
	}
	return []byte(b.String())
}

func baseTournament(tc tournamentCase) *cp.TournamentState {
	rng := rand.New(rand.NewPCG(tc.seed1, tc.seed2))
	players := make([]cp.PlayerEntry, tc.count)
	for i := range players {
		rating := 2450 - i*23 + rng.IntN(13)
		if tc.equal {
			rating = 1800
		}
		players[i] = cp.PlayerEntry{ID: fmt.Sprintf("P%02d", i+1), DisplayName: fmt.Sprintf("Player %02d", i+1), Rating: rating, Federation: []string{"NED", "BEL", "GER"}[i%3]}
	}
	return &cp.TournamentState{Players: players, CurrentRound: 1, ScoringConfig: cp.ScoringConfig{System: cp.ScoringStandard}}
}

func completedTournament(tc tournamentCase) *cp.TournamentState {
	state := baseTournament(tc)
	if tc.name == "features" {
		n := 3
		state.Players[1].WithdrawnAfterRound = &n
		state.Players[2].JoinedRound = 3
	}
	rng := rand.New(rand.NewPCG(tc.seed1+77, tc.seed2+99))
	for round := 1; round <= 6; round++ {
		rd := cp.RoundData{Number: round}
		shift := round % (tc.count - 1)
		used := make([]bool, tc.count)
		for i := 0; i < tc.count; i++ {
			if used[i] {
				continue
			}
			if !state.IsActiveInRound(state.Players[i].ID, round) {
				// Players outside their participation window have no round record.
				// In particular, this distinguishes late joining from absence.
				used[i] = true
				continue
			}
			j := (tc.count - 1 - i + shift) % tc.count
			if j == i || used[j] || !state.IsActiveInRound(state.Players[j].ID, round) {
				rd.Byes = append(rd.Byes, cp.ByeEntry{PlayerID: state.Players[i].ID, Type: allByes[(round+i)%len(allByes)]})
				used[i] = true
				continue
			}
			used[i], used[j] = true, true
			gr := allResults[(rng.IntN(len(allResults))+i+round)%len(allResults)]
			rd.Games = append(rd.Games, cp.GameData{WhiteID: state.Players[i].ID, BlackID: state.Players[j].ID, Result: gr, IsForfeit: gr.IsForfeit()})
		}
		state.Rounds = append(state.Rounds, rd)
	}
	state.CurrentRound = 7
	return state
}

func scoringText(scorer cp.Scorer, state *cp.TournamentState) []byte {
	scores, err := scorer.Score(context.Background(), state)
	var b strings.Builder
	fmt.Fprintf(&b, "error=%v\n", err)
	for _, s := range scores {
		fmt.Fprintf(&b, "%s score=%s rank=%d\n", s.PlayerID, fl(s.Score), s.Rank)
	}
	ctx := cp.ResultContext{OpponentRank: 3, OpponentValueNumber: 17, PlayerRank: 2, PlayerValueNumber: 19}
	for _, result := range allResults {
		fmt.Fprintf(&b, "result=%s points=%s\n", result, fl(scorer.PointsForResult(result, ctx)))
	}
	for _, bt := range allByes {
		bye := bt
		ctx.ByeType = &bye
		fmt.Fprintf(&b, "bye=%s points=%s\n", bt, fl(scorer.PointsForResult(cp.ResultPending, ctx)))
	}
	return []byte(b.String())
}

func trfFixtures(t *testing.T) []string {
	t.Helper()
	roots := []string{"../../trf/testdata", "../../pairing/dutch/testdata", "../../cmd/chesspairing/testdata"}
	var paths []string
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".trf" {
				// Always parse .trf fixtures so malformed inputs become golden errors.
				paths = append(paths, path)
				return nil
			}
			if ext != ".input" {
				// .input is included only when it is recognizable TRF test data.
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if len(data) >= 3 && (string(data[:3]) == "001" || string(data[:3]) == "012") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(paths)
	return paths
}

func stableOptions(opts map[string]any) string {
	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, opts[k]))
	}
	return strings.Join(parts, ",")
}

func fl(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

func golden(t *testing.T, path string, got []byte) {
	t.Helper()
	if os.Getenv("GOLDEN_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("GOLDEN_UPDATE=1: overwrote %s; output was not compared", path)
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s: %v (run with GOLDEN_UPDATE=1)", path, err)
	}
	if bytes.Equal(want, got) {
		return
	}
	t.Fatalf("golden mismatch: %s\n%s", path, compactDiff(want, got))
}

func compactDiff(want, got []byte) string {
	w, g := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	n := min(len(w), len(g))
	line := n
	for i := 0; i < n; i++ {
		if w[i] != g[i] {
			line = i
			break
		}
	}
	var b strings.Builder
	for i := max(0, line-1); i < min(max(len(w), len(g)), line+3); i++ {
		if i < len(w) && i < len(g) && w[i] == g[i] {
			fmt.Fprintf(&b, " %d %s\n", i+1, w[i])
			continue
		}
		if i < len(w) {
			fmt.Fprintf(&b, "-%d %s\n", i+1, w[i])
		}
		if i < len(g) {
			fmt.Fprintf(&b, "+%d %s\n", i+1, g[i])
		}
	}
	return b.String()
}
