// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

//go:build harness

package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/dutch"
	"github.com/gnutterts/chesspairing/trf"
)

type harnessSummary struct {
	Tournaments int
	Rounds      int
	Equal       int
	Sanity      int
	Classes     map[string]int
	Last        int
	Earlier     int
}

type baseline struct {
	Tournaments    int               `json:"tournaments"`
	Seed           int               `json:"seed"`
	Rounds         int               `json:"rounds"`
	MinEqual       int               `json:"minEqual"`
	MaxDifferences map[string]int    `json:"maxDifferences"`
	Reasons        map[string]string `json:"reasons"`
}

type difference struct {
	Seed        int      `json:"seed"`
	Round       int      `json:"round"`
	Class       string   `json:"class"`
	IsLastRound bool     `json:"isLaatsteRonde"`
	Ours        []string `json:"onzeUitvoer"`
	BBP         []string `json:"bbpUitvoer"`
}

func TestHarnessDutch(t *testing.T) {
	runHarness(t, envInt("HARNESS_N", 10), envInt("HARNESS_SEED", 1))
}
func TestHarnessSmoke(t *testing.T) { runHarness(t, 3, 1) }

func runHarness(t *testing.T, n, firstSeed int) {
	t.Helper()
	if _, err := exec.LookPath("bbpPairings"); err != nil {
		t.Skip("bbpPairings is not on PATH; Dutch differential harness cannot run")
	}
	if n < 1 {
		t.Fatalf("HARNESS_N must be positive, got %d", n)
	}
	// Differences are written outside the repository: HARNESS_OUT when set,
	// otherwise a temporary directory, so a run never changes the work tree.
	root := os.Getenv("HARNESS_OUT")
	if root == "" {
		root = t.TempDir()
	} else if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	index, err := os.Create(filepath.Join(root, "index.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer index.Close()

	s := harnessSummary{Tournaments: n, Classes: make(map[string]int)}
	for i := 0; i < n; i++ {
		seed := firstSeed + i
		doc, rounds := generate(t, seed, i%4)
		for round := 1; round <= rounds; round++ {
			s.Rounds++
			cut := truncate(doc, round-1, rounds)
			var trfData bytes.Buffer
			if err := trf.Write(&trfData, cut); err != nil {
				t.Fatalf("seed %d round %d: write TRF: %v", seed, round, err)
			}
			// Pair both engines from the serialized form.  In particular, this
			// checks that the requested byes in the current-round column survive
			// the TRF writer and reader.
			input, err := readTRF(trfData.Bytes())
			if err != nil {
				t.Fatalf("seed %d round %d: reread TRF: %v", seed, round, err)
			}
			if err := checkPreassignedByes(doc, input, round); err != nil {
				t.Fatalf("seed %d round %d: %v", seed, round, err)
			}
			bbp, bbpErr := pairBBP(t, trfData.Bytes())
			want := pairsFromDocument(doc, round)
			if bbpErr != nil || !sameStrings(uncolored(bbp), uncolored(want)) {
				s.Sanity++
			}
			ours, ourErr := pairOurs(trfData.Bytes(), round)
			class := classify(ours, ourErr, bbp, bbpErr, activeFromDocument(input, round))
			if class == "gelijk" {
				s.Equal++
				continue
			}
			s.Classes[class]++
			if round == rounds {
				s.Last++
			} else {
				s.Earlier++
			}
			name := fmt.Sprintf("%d-r%d.trf", seed, round)
			if err := os.WriteFile(filepath.Join(root, name), trfData.Bytes(), 0o644); err != nil {
				t.Fatal(err)
			}
			rec := difference{seed, round, class, round == rounds, printable(ours, ourErr), printable(bbp, bbpErr)}
			line, err := json.Marshal(rec)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := index.Write(append(line, '\n')); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Logf("HARNESS Dutch: toernooien=%d rondes=%d gelijk=%d sanity-afwijkingen=%d verschillen=%v laatste=%d eerder=%d", s.Tournaments, s.Rounds, s.Equal, s.Sanity, orderedCounts(s.Classes), s.Last, s.Earlier)
	checkBaseline(t, s, firstSeed)
}

func checkBaseline(t *testing.T, summary harnessSummary, seed int) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want baseline
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	if summary.Tournaments != want.Tournaments || seed != want.Seed {
		t.Logf("HARNESS baseline not checked for tournaments=%d seed=%d", summary.Tournaments, seed)
		return
	}
	if summary.Sanity != 0 {
		t.Errorf("sanity deviations = %d, want 0", summary.Sanity)
	}
	if summary.Equal < want.MinEqual {
		t.Errorf("equal rounds = %d, want at least %d", summary.Equal, want.MinEqual)
	}
	for class, got := range summary.Classes {
		maximum, ok := want.MaxDifferences[class]
		if !ok {
			t.Errorf("new difference class %q (%d)", class, got)
			continue
		}
		if got > maximum {
			t.Errorf("difference class %q = %d, want at most %d", class, got, maximum)
		}
	}
	if summary.Rounds > want.Rounds || summary.Equal > want.MinEqual {
		t.Log("HARNESS results improved; the baseline can be tightened")
	}
}

func generate(t *testing.T, seed, variant int) (*trf.Document, int) {
	t.Helper()
	configs := [][]string{
		{"PlayersNumber=9", "RoundsNumber=3", "DrawPercentage=10", "ForfeitRate=100", "RetiredRate=100", "HalfPointByeRate=100"},
		{"PlayersNumber=16", "RoundsNumber=7", "DrawPercentage=60", "ForfeitRate=8", "RetiredRate=100", "HalfPointByeRate=8"},
		{"PlayersNumber=31", "RoundsNumber=5", "DrawPercentage=35", "ForfeitRate=5", "RetiredRate=20", "HalfPointByeRate=5"},
		{"PlayersNumber=48", "RoundsNumber=9", "DrawPercentage=50", "ForfeitRate=12", "RetiredRate=12", "HalfPointByeRate=12"},
	}
	lines := append([]string{}, configs[variant]...)
	lines = append(lines, "HighestRating=2600", "LowestRating=1400", "PointsForWin=1.0", "PointsForDraw=0.5", "PointsForLoss=0.0", "PointsForPAB=1.0", "PointsForZPB=0.0", "PointsForForfeitLoss=0.0")
	dir := t.TempDir()
	cfg, out := filepath.Join(dir, "rtg.cfg"), filepath.Join(dir, "out.trf")
	if err := os.WriteFile(cfg, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bbpPairings", "--dutch", "-g", cfg, "-o", out, "-s", strconv.Itoa(seed))
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("seed %d: bbp generator: %v: %s", seed, err, data)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := readTRF(data)
	if err != nil {
		t.Fatalf("seed %d: parse generated TRF: %v", seed, err)
	}
	rounds := 0
	for _, p := range doc.Players {
		if len(p.Rounds) > rounds {
			rounds = len(p.Rounds)
		}
	}
	return doc, rounds
}

func truncate(src *trf.Document, played, total int) *trf.Document {
	out := *src
	out.Players = append([]trf.PlayerLine(nil), src.Players...)
	for i := range out.Players {
		p := &out.Players[i]
		p.Rounds = append([]trf.RoundResult(nil), p.Rounds[:min(played, len(p.Rounds))]...)
		// bbpPairings reads requested H/Z/absence byes from the column of
		// the round it is about to pair. Keep that column, but deliberately
		// omit U: that is bbp's pairing-allocated bye (PAB), which the
		// engine must choose rather than inherit from the generated event.
		if played < total && played < len(src.Players[i].Rounds) {
			rr := src.Players[i].Rounds[played]
			if isPreassignedBye(rr) {
				p.Rounds = append(p.Rounds, rr)
			}
		}
		p.Points = 0
		for _, rr := range p.Rounds[:min(played, len(p.Rounds))] {
			p.Points += trfPoints(rr.Result)
		}
	}
	out.TotalRounds, out.TotalRounds26 = total, 0
	if out.InitialColor == "" {
		out.InitialColor = "white1"
	}
	rankPlayers(out.Players)
	return &out
}

func pairBBP(t *testing.T, input []byte) ([]string, error) {
	t.Helper()
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.trf"), filepath.Join(dir, "out.txt")
	if err := os.WriteFile(in, input, 0o644); err != nil {
		return nil, err
	}
	cmd := exec.Command("bbpPairings", "--dutch", in, "-p", out)
	if data, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(string(data)))
	}
	data, err := os.ReadFile(out)
	if err != nil {
		return nil, err
	}
	return parseList(string(data))
}

func readTRF(data []byte) (*trf.Document, error) {
	return trf.Read(bytes.NewReader(bytes.ReplaceAll(data, []byte{'\r'}, []byte{'\n'})))
}

func pairOurs(input []byte, round int) ([]string, error) {
	doc, err := readTRF(input)
	if err != nil {
		return nil, err
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		return nil, err
	}
	// The round-r records are inputs to the next pairing, not completed
	// results. Convert them to the state field understood by Dutch and keep
	// only rounds before r in the played history.
	if len(state.Rounds) >= round {
		state.Rounds = state.Rounds[:round-1]
	}
	state.CurrentRound = round
	state.PreAssignedByes = preAssignedByes(doc, round)
	result, err := dutch.New(dutch.Options{}).Pair(context.Background(), state)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(result.Pairings)+len(result.Byes))
	for _, p := range result.Pairings {
		out = append(out, p.WhiteID+" "+p.BlackID)
	}
	preassigned := make(map[string]bool, len(state.PreAssignedByes))
	for _, b := range state.PreAssignedByes {
		preassigned[b.PlayerID] = true
	}
	for _, b := range result.Byes {
		if !preassigned[b.PlayerID] {
			out = append(out, b.PlayerID+" 0")
		}
	}
	sort.Strings(out)
	return out, nil
}

func parseList(s string) ([]string, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty bbp pairing output")
	}
	n, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, err
	}
	if len(fields) != 1+2*n {
		return nil, fmt.Errorf("bbp output says %d pairs, has %d fields", n, len(fields)-1)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, fields[1+2*i]+" "+fields[2+2*i])
	}
	sort.Strings(out)
	return out, nil
}

func pairsFromDocument(doc *trf.Document, round int) []string {
	var out []string
	for _, p := range doc.Players {
		if round <= len(p.Rounds) {
			rr := p.Rounds[round-1]
			if rr.Opponent > 0 && rr.Color == trf.ColorWhite {
				out = append(out, fmt.Sprintf("%d %d", p.StartNumber, rr.Opponent))
			} else if rr.Opponent == 0 && rr.Result == trf.ResultUnpaired {
				// U is bbp's pairing-allocated bye. Requested byes were already
				// supplied in the input and are not emitted by `bbpPairings -p`.
				out = append(out, fmt.Sprintf("%d 0", p.StartNumber))
			}
		}
	}
	sort.Strings(out)
	return out
}

func activeFromDocument(doc *trf.Document, round int) map[string]bool {
	out := make(map[string]bool, len(doc.Players))
	for _, p := range doc.Players {
		if round <= len(p.Rounds) && isPreassignedBye(p.Rounds[round-1]) {
			continue
		}
		out[strconv.Itoa(p.StartNumber)] = true
	}
	return out
}

func isPreassignedBye(rr trf.RoundResult) bool {
	return rr.Opponent == 0 && rr.Result != trf.ResultUnpaired
}

func preAssignedByes(doc *trf.Document, round int) []chesspairing.ByeEntry {
	var out []chesspairing.ByeEntry
	for _, p := range doc.Players {
		if round > len(p.Rounds) {
			continue
		}
		rr := p.Rounds[round-1]
		if !isPreassignedBye(rr) {
			continue
		}
		bt := chesspairing.ByeAbsent
		switch rr.Result {
		case trf.ResultHalfBye:
			bt = chesspairing.ByeHalf
		case trf.ResultZeroBye:
			bt = chesspairing.ByeZero
		case trf.ResultFullBye:
			bt = chesspairing.ByePAB
		}
		out = append(out, chesspairing.ByeEntry{PlayerID: strconv.Itoa(p.StartNumber), Type: bt})
	}
	return out
}

func checkPreassignedByes(src, got *trf.Document, round int) error {
	for i, p := range src.Players {
		want := round <= len(p.Rounds) && isPreassignedBye(p.Rounds[round-1])
		found := round <= len(got.Players[i].Rounds) && isPreassignedBye(got.Players[i].Rounds[round-1])
		if want != found {
			return fmt.Errorf("preassigned bye for player %d was not preserved", p.StartNumber)
		}
	}
	return nil
}

func classify(ours []string, ourErr error, bbp []string, bbpErr error, active map[string]bool) string {
	if ourErr != nil || bbpErr != nil {
		return "fout"
	}
	if sameExact(ours, bbp) {
		return "gelijk"
	}
	if !complete(ours, active) || !complete(bbp, active) {
		return "onvolledig"
	}
	ob, bb := byes(ours), byes(bbp)
	if !sameStrings(ob, bb) {
		return "andere-bye"
	}
	if sameStrings(uncolored(ours), uncolored(bbp)) {
		return "zelfde-paren-andere-kleur"
	}
	return "andere-paren"
}

func complete(pairs []string, active map[string]bool) bool {
	seen := map[string]bool{}
	for _, p := range pairs {
		f := strings.Fields(p)
		if len(f) != 2 || seen[f[0]] || (f[1] != "0" && seen[f[1]]) {
			return false
		}
		seen[f[0]] = true
		if f[1] != "0" {
			seen[f[1]] = true
		}
	}
	return sameSet(seen, active)
}

func byes(p []string) []string {
	var o []string
	for _, x := range p {
		f := strings.Fields(x)
		if len(f) == 2 && f[1] == "0" {
			o = append(o, f[0])
		}
	}
	sort.Strings(o)
	return o
}

func uncolored(p []string) []string {
	o := make([]string, 0, len(p))
	for _, x := range p {
		f := strings.Fields(x)
		if len(f) != 2 {
			continue
		}
		if f[1] != "0" && f[1] < f[0] {
			f[0], f[1] = f[1], f[0]
		}
		o = append(o, f[0]+" "+f[1])
	}
	sort.Strings(o)
	return o
}
func sameExact(a, b []string) bool { return sameStrings(a, b) }
func sameStrings(a, b []string) bool {
	return len(a) == len(b) && func() bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}()
}

func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func printable(p []string, err error) []string {
	if err != nil {
		return []string{"ERROR: " + err.Error()}
	}
	return p
}

func trfPoints(r trf.ResultCode) float64 {
	switch r {
	case trf.ResultWin, trf.ResultForfeitWin, trf.ResultWinByDefault, trf.ResultFullBye, trf.ResultUnpaired:
		return 1
	case trf.ResultDraw, trf.ResultDrawByDefault, trf.ResultHalfBye:
		return .5
	}
	return 0
}

func rankPlayers(p []trf.PlayerLine) {
	order := make([]int, len(p))
	for i := range p {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		if p[order[i]].Points != p[order[j]].Points {
			return p[order[i]].Points > p[order[j]].Points
		}
		return p[order[i]].StartNumber < p[order[j]].StartNumber
	})
	for i, x := range order {
		p[x].Rank = i + 1
	}
}

func orderedCounts(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var x []string
	for _, k := range keys {
		x = append(x, k+"="+strconv.Itoa(m[k]))
	}
	return "{" + strings.Join(x, ", ") + "}"
}

func envInt(name string, fallback int) int {
	if s := os.Getenv(name); s != "" {
		if n, e := strconv.Atoi(s); e == nil {
			return n
		}
	}
	return fallback
}
