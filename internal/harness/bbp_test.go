// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

//go:build harness

package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing/trf"
)

const oracleDescription = "bbpPairings v6.0.0 (tag commit 16a000f9811de322b0e835d5643198226165b5a9), --dutch -p, identical on macOS (libc++) and Linux (libstdc++) for all rounds"

func requireBBP(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bbpPairings"); err != nil {
		t.Skip("bbpPairings is not on PATH")
	}
}

// TestHarnessOracle runs bbpPairings again on every corpus round and requires
// the stored answers to be reproduced exactly. It guards the stored oracle
// against a different bbpPairings build or platform.
func TestHarnessOracle(t *testing.T) {
	requireBBP(t)
	oracle := loadOracle(t)
	checked := 0
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
		for round := 1; round <= rounds; round++ {
			got, err := pairBBP(t, roundInput(t, doc, rounds, round))
			if err != nil {
				t.Fatalf("seed %s round %d: %v", seed, round, err)
			}
			if !reflect.DeepEqual(got, oracle.Rounds[seed][round-1]) {
				t.Errorf("seed %s round %d: bbpPairings gives %v, stored %v", seed, round, got, oracle.Rounds[seed][round-1])
			}
			checked++
		}
	}
	t.Logf("HARNESS oracle: %d rounds reproduced", checked)
}

// TestHarnessWriteCorpus regenerates the corpus with bbpPairings' tournament
// generator when HARNESS_WRITE_CORPUS=1. Run it on a Linux machine with GCC's
// libstdc++, the platform bbpPairings is developed on.
func TestHarnessWriteCorpus(t *testing.T) {
	if os.Getenv("HARNESS_WRITE_CORPUS") != "1" {
		t.Skip("set HARNESS_WRITE_CORPUS=1 to regenerate testdata/corpus")
	}
	requireBBP(t)
	n, first := envInt("HARNESS_N", 150), envInt("HARNESS_SEED", 1001)
	if err := os.MkdirAll(filepath.Join("testdata", "corpus"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		seed := first + i
		data := generateTRF(t, seed, i%4)
		if err := os.WriteFile(filepath.Join("testdata", "corpus", fmt.Sprintf("%d.trf", seed)), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// TestHarnessWriteOracle rewrites testdata/oracle.json from bbpPairings when
// HARNESS_WRITE_ORACLE=1.
func TestHarnessWriteOracle(t *testing.T) {
	if os.Getenv("HARNESS_WRITE_ORACLE") != "1" {
		t.Skip("set HARNESS_WRITE_ORACLE=1 to rewrite testdata/oracle.json")
	}
	requireBBP(t)
	out := oracleFile{Oracle: oracleDescription, Rounds: map[string][][]string{}}
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
		for round := 1; round <= rounds; round++ {
			pairs, err := pairBBP(t, roundInput(t, doc, rounds, round))
			if err != nil {
				t.Fatalf("seed %s round %d: %v", seed, round, err)
			}
			out.Rounds[seed] = append(out.Rounds[seed], pairs)
		}
	}
	data, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("testdata", "oracle.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestHarnessExplore pairs freshly generated tournaments (HARNESS_N of them,
// from HARNESS_SEED) with both engines. Because the generator differs per
// platform it is a search for new differences, not a regression gate: it
// fails on any difference other than a colour choice.
func TestHarnessExplore(t *testing.T) {
	requireBBP(t)
	n, first := envInt("HARNESS_N", 10), envInt("HARNESS_SEED", 1)
	root := os.Getenv("HARNESS_OUT")
	if root == "" {
		root = t.TempDir()
	} else if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	s := harnessSummary{Tournaments: n, Classes: make(map[string]int)}
	for i := 0; i < n; i++ {
		seed := first + i
		variant := i % 4
		if os.Getenv("HARNESS_BAKU") == "1" {
			// bbpPairings makes group A the first ceil(N/2) players where C.04.7
			// 1.2 says 2*ceil(N/4) (161 participants: 82, not 81). The two agree
			// for the 16, 31 and 48 player configurations, not for the 9 player
			// one, so that one is left out of the Baku comparison.
			variant = 1 + i%3
		}
		doc, err := readTRF(generateTRF(t, seed, variant))
		if err != nil {
			t.Fatalf("seed %d: parse generated TRF: %v", seed, err)
		}
		rounds := tournamentRounds(doc)
		if os.Getenv("HARNESS_BAKU") == "1" {
			// Record 192 asks both engines for Baku acceleration.
			doc.CodedTournamentType = "FIDE_DUTCH_2025_BAKU"
		}
		for round := 1; round <= rounds; round++ {
			s.Rounds++
			input := roundInput(t, doc, rounds, round)
			reread, err := readTRF(input)
			if err != nil {
				t.Fatalf("seed %d round %d: reread TRF: %v", seed, round, err)
			}
			if err := checkPreassignedByes(doc, reread, round); err != nil {
				t.Fatalf("seed %d round %d: %v", seed, round, err)
			}
			bbp, bbpErr := pairBBP(t, input)
			ours, ourErr := pairOurs(input, round)
			class := classify(ours, ourErr, bbp, bbpErr, activeFromDocument(reread, round))
			if class == "gelijk" {
				s.Equal++
				continue
			}
			s.Classes[class]++
			name := fmt.Sprintf("%d-r%d", seed, round)
			if err := os.WriteFile(filepath.Join(root, name+".trf"), input, 0o644); err != nil {
				t.Fatal(err)
			}
			rec, _ := json.Marshal(difference{seed, round, class, round == rounds, printable(ours, ourErr), printable(bbp, bbpErr)})
			if err := os.WriteFile(filepath.Join(root, name+".json"), rec, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Logf("HARNESS explore: toernooien=%d rondes=%d gelijk=%d verschillen=%v", s.Tournaments, s.Rounds, s.Equal, orderedCounts(s.Classes))
	for class, got := range s.Classes {
		if class != "zelfde-paren-andere-kleur" {
			t.Errorf("difference class %q = %d", class, got)
		}
	}
}

func generateTRF(t *testing.T, seed, variant int) []byte {
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
	return data
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

type difference struct {
	Seed        int      `json:"seed"`
	Round       int      `json:"round"`
	Class       string   `json:"class"`
	IsLastRound bool     `json:"isLaatsteRonde"`
	Ours        []string `json:"onzeUitvoer"`
	BBP         []string `json:"bbpUitvoer"`
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

func printable(p []string, err error) []string {
	if err != nil {
		return []string{"ERROR: " + err.Error()}
	}
	return p
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
