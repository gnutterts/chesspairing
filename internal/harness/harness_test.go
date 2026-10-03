// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/factory"
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

type difference struct {
	Seed        int      `json:"seed"`
	Round       int      `json:"round"`
	Class       string   `json:"class"`
	IsLastRound bool     `json:"isLaatsteRonde"`
	Ours        []string `json:"onzeUitvoer"`
	BBP         []string `json:"bbpUitvoer"`
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
	// Build the pairer from the options the TRF conversion produced, so that the
	// total number of rounds (XXR) reaches the last-round rules as it does in
	// the command line tool.
	pairer, err := factory.NewPairer(string(chesspairing.PairingDutch), state.PairingConfig.Options)
	if err != nil {
		return nil, err
	}
	result, err := pairer.Pair(context.Background(), state)
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

// The corpus is a set of complete tournaments generated once with
// bbpPairings' random tournament generator (see testdata/README.md), together
// with bbpPairings' own pairing for every round (testdata/oracle.json). The
// generator is not portable across standard libraries, the pairing is, so the
// stored inputs and answers make the comparison identical on every platform.

type oracleFile struct {
	Oracle string                `json:"oracle"`
	Rounds map[string][][]string `json:"rounds"`
}

type corpusBaseline struct {
	Tournaments    int               `json:"tournaments"`
	Rounds         int               `json:"rounds"`
	MinEqual       int               `json:"minEqual"`
	MaxDifferences map[string]int    `json:"maxDifferences"`
	Reasons        map[string]string `json:"reasons"`
}

func corpusFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "corpus", "*.trf"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus files: %v", err)
	}
	sort.Strings(files)
	return files
}

func corpusSeed(file string) string {
	return strings.TrimSuffix(filepath.Base(file), ".trf")
}

func loadOracle(t *testing.T) oracleFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "oracle.json"))
	if err != nil {
		t.Fatal(err)
	}
	var o oracleFile
	if err := json.Unmarshal(data, &o); err != nil {
		t.Fatal(err)
	}
	return o
}

// roundInput returns the TRF for pairing the given round of a corpus
// tournament: the history before it, plus the requested byes of that round.
func roundInput(t *testing.T, doc *trf.Document, rounds, round int) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := trf.Write(&b, truncate(doc, round-1, rounds)); err != nil {
		t.Fatalf("round %d: write TRF: %v", round, err)
	}
	return b.Bytes()
}

func tournamentRounds(doc *trf.Document) int {
	rounds := 0
	for _, p := range doc.Players {
		if len(p.Rounds) > rounds {
			rounds = len(p.Rounds)
		}
	}
	return rounds
}

// TestHarnessCorpus compares our Dutch pairing with the stored bbpPairings
// answers for every round of every corpus tournament. It needs no external
// program and gives the same result on every platform.
func TestHarnessCorpus(t *testing.T) {
	oracle := loadOracle(t)
	data, err := os.ReadFile(filepath.Join("testdata", "baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	var want corpusBaseline
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatal(err)
	}
	s := harnessSummary{Classes: make(map[string]int)}
	for _, file := range corpusFiles(t) {
		seed := corpusSeed(file)
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := readTRF(raw)
		if err != nil {
			t.Fatalf("seed %s: %v", seed, err)
		}
		rounds := tournamentRounds(doc)
		stored := oracle.Rounds[seed]
		if len(stored) != rounds {
			t.Fatalf("seed %s: oracle has %d rounds, tournament has %d", seed, len(stored), rounds)
		}
		s.Tournaments++
		for round := 1; round <= rounds; round++ {
			s.Rounds++
			input := roundInput(t, doc, rounds, round)
			reread, err := readTRF(input)
			if err != nil {
				t.Fatalf("seed %s round %d: reread TRF: %v", seed, round, err)
			}
			ours, ourErr := pairOurs(input, round)
			class := classify(ours, ourErr, stored[round-1], nil, activeFromDocument(reread, round))
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
			t.Logf("seed %s round %d: %s", seed, round, class)
		}
	}
	t.Logf("HARNESS corpus: toernooien=%d rondes=%d gelijk=%d verschillen=%v laatste=%d eerder=%d", s.Tournaments, s.Rounds, s.Equal, orderedCounts(s.Classes), s.Last, s.Earlier)
	if s.Tournaments != want.Tournaments || s.Rounds != want.Rounds {
		t.Errorf("corpus has %d tournaments and %d rounds, baseline says %d and %d", s.Tournaments, s.Rounds, want.Tournaments, want.Rounds)
	}
	if s.Equal < want.MinEqual {
		t.Errorf("equal rounds = %d, want at least %d", s.Equal, want.MinEqual)
	}
	for class, got := range s.Classes {
		maximum, ok := want.MaxDifferences[class]
		if !ok {
			t.Errorf("new difference class %q (%d)", class, got)
			continue
		}
		if got > maximum {
			t.Errorf("difference class %q = %d, want at most %d", class, got, maximum)
		}
	}
	if s.Equal > want.MinEqual {
		t.Log("HARNESS results improved; the baseline can be tightened")
	}
}
