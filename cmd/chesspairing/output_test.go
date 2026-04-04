// cmd/chesspairing/output_test.go
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	cp "github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/trf"
)

func TestFormatPairList(t *testing.T) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "1", BlackID: "4"},
			{Board: 2, WhiteID: "2", BlackID: "3"},
		},
		Byes: []cp.ByeEntry{
			{PlayerID: "5", Type: cp.ByePAB},
		},
	}
	// Build a minimal player ID→start number map
	playerNumbers := map[string]int{"1": 1, "2": 2, "3": 3, "4": 4, "5": 5}

	var buf bytes.Buffer
	formatPairList(&buf, result, playerNumbers)
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	// First line: number of pairings (not including byes)
	if lines[0] != "2" {
		t.Errorf("first line: got %q, want %q", lines[0], "2")
	}
	// Pairing lines: "white black"
	if lines[1] != "1 4" {
		t.Errorf("pairing 1: got %q, want %q", lines[1], "1 4")
	}
	if lines[2] != "2 3" {
		t.Errorf("pairing 2: got %q, want %q", lines[2], "2 3")
	}
	// Bye line: "player 0"
	if lines[3] != "5 0" {
		t.Errorf("bye: got %q, want %q", lines[3], "5 0")
	}
}

func TestFormatStandingsText(t *testing.T) {
	standings := []cp.Standing{
		{Rank: 1, PlayerID: "1", DisplayName: "Fischer, Robert", Score: 6.5,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 32.0}},
			GamesPlayed: 9, Wins: 6, Draws: 1, Losses: 2},
		{Rank: 2, PlayerID: "2", DisplayName: "Karpov, Anatoly", Score: 6.0,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 30.0}},
			GamesPlayed: 9, Wins: 5, Draws: 2, Losses: 2},
	}
	var buf bytes.Buffer
	formatStandingsText(&buf, standings)
	out := buf.String()
	if !strings.Contains(out, "Fischer") {
		t.Errorf("should contain Fischer, got: %s", out)
	}
	if !strings.Contains(out, "6.5") {
		t.Errorf("should contain 6.5, got: %s", out)
	}
}

func TestFormatStandingsJSON(t *testing.T) {
	standings := []cp.Standing{
		{Rank: 1, PlayerID: "1", DisplayName: "Fischer, Robert", Score: 6.5,
			TieBreakers: []cp.NamedValue{{ID: "buchholz", Name: "Buchholz", Value: 32.0}},
			GamesPlayed: 9, Wins: 6, Draws: 1, Losses: 2},
	}
	var buf bytes.Buffer
	if err := formatStandingsJSON(&buf, standings, "standard", []string{"buchholz"}); err != nil {
		t.Fatalf("formatStandingsJSON: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	stArr, ok := result["standings"].([]any)
	if !ok || len(stArr) != 1 {
		t.Fatalf("expected 1 standing, got %v", result["standings"])
	}
}

func TestFormatValidationText(t *testing.T) {
	issues := []trf.ValidationIssue{
		{Field: "XXR", Severity: trf.SeverityError, Message: "missing total rounds"},
		{Field: "player.2.rating", Severity: trf.SeverityWarning, Message: "no rating"},
	}
	var buf bytes.Buffer
	formatValidationText(&buf, "test.trf", issues)
	out := buf.String()
	if !strings.Contains(out, "1 error") {
		t.Errorf("should report 1 error, got: %s", out)
	}
	if !strings.Contains(out, "1 warning") {
		t.Errorf("should report 1 warning, got: %s", out)
	}
}

func TestFormatValidationJSON(t *testing.T) {
	issues := []trf.ValidationIssue{
		{Field: "XXR", Severity: trf.SeverityError, Message: "missing total rounds"},
	}
	var buf bytes.Buffer
	if err := formatValidationJSON(&buf, issues, "standard", "trfx"); err != nil {
		t.Fatalf("formatValidationJSON: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false")
	}
}

// testPairingData returns shared test fixtures for format tests.
func testPairingData() (*cp.PairingResult, map[string]int, *cp.TournamentState) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "5", BlackID: "1"},
			{Board: 2, WhiteID: "4", BlackID: "2"},
		},
		Byes: []cp.ByeEntry{
			{PlayerID: "3", Type: cp.ByePAB},
		},
	}
	playerNumbers := map[string]int{"1": 1, "2": 2, "3": 3, "4": 4, "5": 5}
	state := &cp.TournamentState{
		CurrentRound: 2,
		Players: []cp.PlayerEntry{
			{ID: "1", DisplayName: "Kasparov, Garry", Rating: 2812, Title: "GM", Active: true},
			{ID: "2", DisplayName: "Kramnik, Vladimir", Rating: 2750, Title: "IM", Active: true},
			{ID: "3", DisplayName: "Player Three", Rating: 2000, Active: true},
			{ID: "4", DisplayName: "Polgar, Judit", Rating: 2735, Title: "WGM", Active: true},
			{ID: "5", DisplayName: "Player Five", Rating: 1800, Active: true},
		},
	}
	return result, playerNumbers, state
}

func TestFormatPairWide(t *testing.T) {
	result, playerNumbers, state := testPairingData()

	var buf bytes.Buffer
	formatPairWide(&buf, result, playerNumbers, state)
	out := buf.String()

	// Should contain header
	if !strings.Contains(out, "Board") {
		t.Errorf("should contain Board header, got:\n%s", out)
	}
	// Should contain player names with titles
	if !strings.Contains(out, "GM Kasparov, Garry") {
		t.Errorf("should contain titled player name, got:\n%s", out)
	}
	if !strings.Contains(out, "Player Five") {
		t.Errorf("should contain untitled player name, got:\n%s", out)
	}
	// Should contain ratings
	if !strings.Contains(out, "2812") {
		t.Errorf("should contain rating 2812, got:\n%s", out)
	}
	// Should contain bye line
	if !strings.Contains(out, "Bye (PAB)") {
		t.Errorf("should contain bye type, got:\n%s", out)
	}
	if !strings.Contains(out, "Player Three") {
		t.Errorf("should contain bye player name, got:\n%s", out)
	}
}

func TestFormatPairWide_NoByes(t *testing.T) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "1", BlackID: "2"},
		},
	}
	playerNumbers := map[string]int{"1": 1, "2": 2}
	state := &cp.TournamentState{
		CurrentRound: 0,
		Players: []cp.PlayerEntry{
			{ID: "1", DisplayName: "White Player", Rating: 2000, Active: true},
			{ID: "2", DisplayName: "Black Player", Rating: 1900, Active: true},
		},
	}

	var buf bytes.Buffer
	formatPairWide(&buf, result, playerNumbers, state)
	out := buf.String()

	if strings.Contains(out, "Bye") {
		t.Errorf("should not contain bye line when no byes, got:\n%s", out)
	}
}

func TestFormatPairBoard(t *testing.T) {
	result, playerNumbers, _ := testPairingData()

	var buf bytes.Buffer
	formatPairBoard(&buf, result, playerNumbers)
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %s", len(lines), out)
	}
	// Board lines
	if !strings.Contains(lines[0], "Board") {
		t.Errorf("line 1 should start with Board, got: %s", lines[0])
	}
	if !strings.Contains(lines[0], "5 - 1") {
		t.Errorf("line 1 should contain '5 - 1', got: %s", lines[0])
	}
	if !strings.Contains(lines[1], "4 - 2") {
		t.Errorf("line 2 should contain '4 - 2', got: %s", lines[1])
	}
	// Bye line
	if !strings.HasPrefix(lines[2], "Bye:") {
		t.Errorf("line 3 should start with 'Bye:', got: %s", lines[2])
	}
	if !strings.Contains(lines[2], "3") {
		t.Errorf("bye line should contain player 3, got: %s", lines[2])
	}
}

func TestFormatPairBoard_NoByes(t *testing.T) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "1", BlackID: "2"},
		},
	}
	playerNumbers := map[string]int{"1": 1, "2": 2}

	var buf bytes.Buffer
	formatPairBoard(&buf, result, playerNumbers)
	out := buf.String()

	if strings.Contains(out, "Bye") {
		t.Errorf("should not contain bye when no byes, got:\n%s", out)
	}
}

func TestFormatPairXML(t *testing.T) {
	result, playerNumbers, state := testPairingData()

	var buf bytes.Buffer
	if err := formatPairXML(&buf, result, playerNumbers, state); err != nil {
		t.Fatalf("formatPairXML: %v", err)
	}
	out := buf.String()

	// Should be valid XML
	if !strings.Contains(out, "<?xml") {
		t.Errorf("should contain XML declaration, got:\n%s", out)
	}

	// Parse and verify structure
	type xmlPlayer struct {
		Number int    `xml:"number,attr"`
		Name   string `xml:"name,attr"`
		Rating int    `xml:"rating,attr"`
		Title  string `xml:"title,attr"`
	}
	type xmlBoard struct {
		Number int       `xml:"number,attr"`
		White  xmlPlayer `xml:"white"`
		Black  xmlPlayer `xml:"black"`
	}
	type xmlBye struct {
		Number int    `xml:"number,attr"`
		Name   string `xml:"name,attr"`
		Type   string `xml:"type,attr"`
	}
	type xmlPairings struct {
		XMLName xml.Name   `xml:"pairings"`
		Round   int        `xml:"round,attr"`
		Boards  int        `xml:"boards,attr"`
		Byes    int        `xml:"byes,attr"`
		Board   []xmlBoard `xml:"board"`
		Bye     []xmlBye   `xml:"bye"`
	}

	var parsed xmlPairings
	if err := xml.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("XML parse error: %v", err)
	}

	if parsed.Round != 3 {
		t.Errorf("round: got %d, want 3", parsed.Round)
	}
	if parsed.Boards != 2 {
		t.Errorf("boards: got %d, want 2", parsed.Boards)
	}
	if parsed.Byes != 1 {
		t.Errorf("byes: got %d, want 1", parsed.Byes)
	}

	// Check board 1
	if len(parsed.Board) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(parsed.Board))
	}
	b1 := parsed.Board[0]
	if b1.White.Number != 5 || b1.Black.Number != 1 {
		t.Errorf("board 1: got %d vs %d, want 5 vs 1", b1.White.Number, b1.Black.Number)
	}
	if b1.White.Name != "Player Five" {
		t.Errorf("board 1 white name: got %q, want %q", b1.White.Name, "Player Five")
	}
	if b1.Black.Title != "GM" {
		t.Errorf("board 1 black title: got %q, want %q", b1.Black.Title, "GM")
	}

	// Check bye
	if len(parsed.Bye) != 1 {
		t.Fatalf("expected 1 bye, got %d", len(parsed.Bye))
	}
	if parsed.Bye[0].Number != 3 {
		t.Errorf("bye player: got %d, want 3", parsed.Bye[0].Number)
	}
	if parsed.Bye[0].Type != "PAB" {
		t.Errorf("bye type: got %q, want %q", parsed.Bye[0].Type, "PAB")
	}
}

func TestFormatPairXML_NoByes(t *testing.T) {
	result := &cp.PairingResult{
		Pairings: []cp.GamePairing{
			{Board: 1, WhiteID: "1", BlackID: "2"},
		},
	}
	playerNumbers := map[string]int{"1": 1, "2": 2}
	state := &cp.TournamentState{
		CurrentRound: 0,
		Players: []cp.PlayerEntry{
			{ID: "1", DisplayName: "White Player", Rating: 2000, Active: true},
			{ID: "2", DisplayName: "Black Player", Rating: 1900, Active: true},
		},
	}

	var buf bytes.Buffer
	if err := formatPairXML(&buf, result, playerNumbers, state); err != nil {
		t.Fatalf("formatPairXML: %v", err)
	}
	if strings.Contains(buf.String(), "<bye") {
		t.Errorf("should not contain bye element when no byes")
	}
}

func TestDigitWidth(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{0, 1},
		{1, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
	}
	for _, tc := range tests {
		if got := digitWidth(tc.n); got != tc.want {
			t.Errorf("digitWidth(%d): got %d, want %d", tc.n, got, tc.want)
		}
	}
}
