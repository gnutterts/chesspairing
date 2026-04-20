// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnutterts/chesspairing"
)

// TestParseChesspairingDirective_byeRoundTrip checks that a typed bye
// directive parses into a Directive with verb and params, then writes back
// to byte-equivalent output.
func TestParseChesspairingDirective_byeRoundTrip(t *testing.T) {
	input := "### chesspairing:bye round=3 player=5 type=excused\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(doc.Comments) != 0 {
		t.Errorf("Comments = %v, want none (directive should not become a comment)", doc.Comments)
	}
	if len(doc.ChesspairingDirectives) != 1 {
		t.Fatalf("ChesspairingDirectives = %d, want 1", len(doc.ChesspairingDirectives))
	}
	d := doc.ChesspairingDirectives[0]
	if d.Verb != "bye" {
		t.Errorf("Verb = %q, want %q", d.Verb, "bye")
	}
	wantParams := map[string]string{"round": "3", "player": "5", "type": "excused"}
	if len(d.Params) != len(wantParams) {
		t.Errorf("Params length = %d, want %d (params=%v)", len(d.Params), len(wantParams), d.Params)
	}
	for k, v := range wantParams {
		if d.Params[k] != v {
			t.Errorf("Params[%q] = %q, want %q", k, d.Params[k], v)
		}
	}

	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if got := buf.String(); got != input {
		t.Errorf("round-trip mismatch:\n got: %q\nwant: %q", got, input)
	}
}

// TestParseChesspairingDirective_withdrawnRoundTrip checks that a withdrawn
// directive survives Read/Write at the directive level. Bridging into
// PlayerEntry.WithdrawnAfterRound is exercised separately in
// TestBridgeWithdrawnDirectives_roundTrip.
func TestParseChesspairingDirective_withdrawnRoundTrip(t *testing.T) {
	input := "### chesspairing:withdrawn player=3 after-round=4\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.ChesspairingDirectives) != 1 {
		t.Fatalf("ChesspairingDirectives = %d, want 1", len(doc.ChesspairingDirectives))
	}
	d := doc.ChesspairingDirectives[0]
	if d.Verb != "withdrawn" {
		t.Errorf("Verb = %q, want %q", d.Verb, "withdrawn")
	}
	if d.Params["player"] != "3" || d.Params["after-round"] != "4" {
		t.Errorf("Params = %v", d.Params)
	}

	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if got := buf.String(); got != input {
		t.Errorf("round-trip mismatch:\n got: %q\nwant: %q", got, input)
	}
}

// TestParseChesspairingDirective_unknownVerbPreserved keeps an unfamiliar
// verb so newer parsers can still understand a TRF that an older library
// version wrote without dropping data.
func TestParseChesspairingDirective_unknownVerbPreserved(t *testing.T) {
	input := "### chesspairing:future-thing key=value\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.ChesspairingDirectives) != 1 || doc.ChesspairingDirectives[0].Verb != "future-thing" {
		t.Fatalf("directive not preserved: %+v", doc.ChesspairingDirectives)
	}

	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if got := buf.String(); got != input {
		t.Errorf("round-trip mismatch:\n got: %q\nwant: %q", got, input)
	}
}

// TestParseChesspairingDirective_malformedFallsBack ensures a typo'd
// directive is preserved as a plain comment rather than silently rewritten.
func TestParseChesspairingDirective_malformedFallsBack(t *testing.T) {
	input := "### chesspairing:bye round=3 brokenparam type=excused\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(doc.ChesspairingDirectives) != 0 {
		t.Errorf("ChesspairingDirectives = %v, want none", doc.ChesspairingDirectives)
	}
	if len(doc.Comments) != 1 {
		t.Fatalf("Comments = %v, want one", doc.Comments)
	}
	if !strings.Contains(doc.Comments[0], "brokenparam") {
		t.Errorf("Comment text = %q", doc.Comments[0])
	}
}

// TestBridgePreAssignedByes_section240 exercises Section 240 → ByePAB / ByeHalf.
func TestBridgePreAssignedByes_section240(t *testing.T) {
	input := "012 Test\n092 Swiss Dutch\nXXR 5\nXXC white1\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"
	input += "001    2      Player Two                        1900 NED                         0.0    2\n"
	input += "001    3      Player Three                      1800 NED                         0.0    3\n"
	input += "240 F   1  0002\n"
	input += "240 H   1  0003\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	state.CurrentRound = 1 // ToTournamentState defaults to maxRounds; pin it.

	state.PreAssignedByes = nil
	if err := bridgePreAssignedByes(doc, state); err != nil {
		t.Fatalf("bridgePreAssignedByes failed: %v", err)
	}

	if len(state.PreAssignedByes) != 2 {
		t.Fatalf("PreAssignedByes = %+v, want 2 entries", state.PreAssignedByes)
	}
	got := map[string]chesspairing.ByeType{}
	for _, b := range state.PreAssignedByes {
		got[b.PlayerID] = b.Type
	}
	if got["2"] != chesspairing.ByePAB {
		t.Errorf("player 2 type = %v, want ByePAB", got["2"])
	}
	if got["3"] != chesspairing.ByeHalf {
		t.Errorf("player 3 type = %v, want ByeHalf", got["3"])
	}
}

// TestBridgePreAssignedByes_directiveOverrides ensures a chesspairing:bye
// directive wins over a Section 240 record for the same (round, player).
func TestBridgePreAssignedByes_directiveOverrides(t *testing.T) {
	input := "012 Test\n092 Swiss Dutch\nXXR 5\nXXC white1\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"
	input += "001    2      Player Two                        1900 NED                         0.0    2\n"
	input += "240 F   1  0002\n"
	input += "### chesspairing:bye round=1 player=2 type=excused\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	state.CurrentRound = 1
	state.PreAssignedByes = nil
	if err := bridgePreAssignedByes(doc, state); err != nil {
		t.Fatalf("bridgePreAssignedByes failed: %v", err)
	}

	if len(state.PreAssignedByes) != 1 {
		t.Fatalf("PreAssignedByes = %+v, want 1 entry", state.PreAssignedByes)
	}
	if state.PreAssignedByes[0].Type != chesspairing.ByeExcused {
		t.Errorf("type = %v, want ByeExcused (directive should override 240)", state.PreAssignedByes[0].Type)
	}
}

// TestBridgePreAssignedByes_unknownPlayer rejects Section 240 records
// that refer to a start number with no matching player line.
func TestBridgePreAssignedByes_unknownPlayer(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Player One", Rating: 2000},
		},
		Absences: []AbsenceRecord{
			{Type: "F", Round: 1, Players: []int{99}},
		},
	}
	state := &chesspairing.TournamentState{
		Players:      []chesspairing.PlayerEntry{{ID: "1", DisplayName: "Player One", Rating: 2000}},
		CurrentRound: 1,
	}
	err := bridgePreAssignedByes(doc, state)
	if err == nil {
		t.Fatal("bridgePreAssignedByes succeeded, want error for unknown player ID")
	}
	if !strings.Contains(err.Error(), "unknown player") {
		t.Errorf("error = %v, want one mentioning 'unknown player'", err)
	}
}

// TestEmitPreAssignedByes_roundTrip is the inverse path: a state with mixed
// bye types becomes a Section 240 record (for PAB / Half) plus directives
// (for the richer types), and reading the emitted document back yields the
// original PreAssignedByes set.
func TestEmitPreAssignedByes_roundTrip(t *testing.T) {
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "a", DisplayName: "Alice", Rating: 2200},
			{ID: "b", DisplayName: "Bob", Rating: 2100},
			{ID: "c", DisplayName: "Carol", Rating: 2000},
			{ID: "d", DisplayName: "Dan", Rating: 1900},
		},
		CurrentRound: 4,
		PreAssignedByes: []chesspairing.ByeEntry{
			{PlayerID: "a", Type: chesspairing.ByePAB},
			{PlayerID: "b", Type: chesspairing.ByeHalf},
			{PlayerID: "c", Type: chesspairing.ByeExcused},
			{PlayerID: "d", Type: chesspairing.ByeClubCommitment},
		},
		PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
	}

	doc, _ := FromTournamentState(state)

	// Section 240 should hold Alice (F) and Bob (H), one record each.
	if len(doc.Absences) != 2 {
		t.Fatalf("Absences = %+v, want 2 records", doc.Absences)
	}
	for _, a := range doc.Absences {
		if a.Round != 4 {
			t.Errorf("Absence round = %d, want 4", a.Round)
		}
	}

	// Directives carry the two richer types.
	if len(doc.ChesspairingDirectives) != 2 {
		t.Fatalf("Directives = %+v, want 2", doc.ChesspairingDirectives)
	}
	for _, d := range doc.ChesspairingDirectives {
		if d.Verb != "bye" {
			t.Errorf("directive verb = %q, want %q", d.Verb, "bye")
		}
		if d.Params["round"] != "4" {
			t.Errorf("directive round = %q, want %q", d.Params["round"], "4")
		}
	}

	// Round-trip through Write/Read and back into a state.
	var buf bytes.Buffer
	if err := Write(&buf, doc); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	doc2, err := Read(&buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	state2, err := doc2.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	// CurrentRound after read defaults to len(rounds); pin to 4 and rerun the
	// bridge so the assertion compares like with like.
	state2.CurrentRound = 4
	state2.PreAssignedByes = nil
	if err := bridgePreAssignedByes(doc2, state2); err != nil {
		t.Fatalf("bridgePreAssignedByes failed: %v", err)
	}

	got := map[string]chesspairing.ByeType{}
	for _, b := range state2.PreAssignedByes {
		got[b.PlayerID] = b.Type
	}
	want := map[string]chesspairing.ByeType{
		"1": chesspairing.ByePAB,            // Alice, top seed
		"2": chesspairing.ByeHalf,           // Bob
		"3": chesspairing.ByeExcused,        // Carol
		"4": chesspairing.ByeClubCommitment, // Dan
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("player %s: got %v, want %v (full map: %v)", id, got[id], w, got)
		}
	}
}

// TestBridgeWithdrawnDirectives_roundTrip checks that a withdrawn directive
// flows into PlayerEntry.WithdrawnAfterRound on read and back into a
// directive on write.
func TestBridgeWithdrawnDirectives_roundTrip(t *testing.T) {
	input := "012 Test\n092 Swiss Dutch\nXXR 5\nXXC white1\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"
	input += "001    2      Player Two                        1900 NED                         0.0    2\n"
	input += "001    3      Player Three                      1800 NED                         0.0    3\n"
	input += "### chesspairing:withdrawn player=2 after-round=3\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	state, err := doc.ToTournamentState()
	if err != nil {
		t.Fatalf("ToTournamentState failed: %v", err)
	}
	if state.Players[1].WithdrawnAfterRound == nil || *state.Players[1].WithdrawnAfterRound != 3 {
		t.Fatalf("Players[1].WithdrawnAfterRound = %v, want *3", state.Players[1].WithdrawnAfterRound)
	}
	if state.Players[0].WithdrawnAfterRound != nil || state.Players[2].WithdrawnAfterRound != nil {
		t.Errorf("unexpected WithdrawnAfterRound on other players")
	}

	// Write back through FromTournamentState. The withdrawn directive
	// should reappear in doc2.ChesspairingDirectives.
	doc2, _ := FromTournamentState(state)
	var found bool
	for _, d := range doc2.ChesspairingDirectives {
		if d.Verb == "withdrawn" && d.Params["player"] == "2" && d.Params["after-round"] == "3" {
			found = true
		}
	}
	if !found {
		t.Errorf("withdrawn directive missing from FromTournamentState output: %+v", doc2.ChesspairingDirectives)
	}
}

// TestBridgeWithdrawnDirectives_unknownPlayer reports a validation error.
func TestBridgeWithdrawnDirectives_unknownPlayer(t *testing.T) {
	input := "012 Test\n092 Swiss Dutch\nXXR 5\n"
	input += "001    1      Player One                        2000 NED                         0.0    1\n"
	input += "### chesspairing:withdrawn player=9 after-round=2\n"

	doc, err := Read(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if _, err := doc.ToTournamentState(); err == nil {
		t.Error("expected error for unknown player")
	}
}
