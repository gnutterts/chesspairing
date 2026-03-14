package trf

import (
	"encoding/json"
	"testing"
)

func TestColor_MarshalJSON(t *testing.T) {
	tests := []struct {
		color Color
		want  string
	}{
		{ColorNone, `"-"`},
		{ColorWhite, `"w"`},
		{ColorBlack, `"b"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.color)
		if err != nil {
			t.Fatalf("Marshal(%v) error: %v", tt.color, err)
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%v) = %s, want %s", tt.color, data, tt.want)
		}
	}
}

func TestColor_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Color
	}{
		{`"w"`, ColorWhite},
		{`"b"`, ColorBlack},
		{`"-"`, ColorNone},
	}
	for _, tt := range tests {
		var c Color
		if err := json.Unmarshal([]byte(tt.input), &c); err != nil {
			t.Fatalf("Unmarshal(%s) error: %v", tt.input, err)
		}
		if c != tt.want {
			t.Errorf("Unmarshal(%s) = %v, want %v", tt.input, c, tt.want)
		}
	}
}

func TestColor_UnmarshalJSON_invalid(t *testing.T) {
	invalid := []string{`"x"`, `"ww"`, `""`}
	for _, input := range invalid {
		var c Color
		if err := json.Unmarshal([]byte(input), &c); err == nil {
			t.Errorf("Unmarshal(%s) expected error, got %v", input, c)
		}
	}
}

func TestResultCode_MarshalJSON(t *testing.T) {
	tests := []struct {
		rc   ResultCode
		want string
	}{
		{ResultWin, `"1"`},
		{ResultLoss, `"0"`},
		{ResultDraw, `"="`},
		{ResultForfeitWin, `"+"`},
		{ResultForfeitLoss, `"-"`},
		{ResultHalfBye, `"H"`},
		{ResultFullBye, `"F"`},
		{ResultUnpaired, `"U"`},
		{ResultZeroBye, `"Z"`},
		{ResultNotPlayed, `"*"`},
		{ResultWinByDefault, `"W"`},
		{ResultDrawByDefault, `"D"`},
		{ResultLossByDefault, `"L"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.rc)
		if err != nil {
			t.Fatalf("Marshal(%v) error: %v", tt.rc, err)
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%v) = %s, want %s", tt.rc, data, tt.want)
		}
	}
}

func TestResultCode_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  ResultCode
	}{
		{`"1"`, ResultWin},
		{`"0"`, ResultLoss},
		{`"="`, ResultDraw},
		{`"+"`, ResultForfeitWin},
		{`"-"`, ResultForfeitLoss},
		{`"H"`, ResultHalfBye},
		{`"F"`, ResultFullBye},
		{`"U"`, ResultUnpaired},
		{`"Z"`, ResultZeroBye},
		{`"*"`, ResultNotPlayed},
		{`"W"`, ResultWinByDefault},
		{`"D"`, ResultDrawByDefault},
		{`"L"`, ResultLossByDefault},
	}
	for _, tt := range tests {
		var rc ResultCode
		if err := json.Unmarshal([]byte(tt.input), &rc); err != nil {
			t.Fatalf("Unmarshal(%s) error: %v", tt.input, err)
		}
		if rc != tt.want {
			t.Errorf("Unmarshal(%s) = %v, want %v", tt.input, rc, tt.want)
		}
	}
}

func TestResultCode_UnmarshalJSON_invalid(t *testing.T) {
	invalid := []string{`"x"`, `"11"`, `""`}
	for _, input := range invalid {
		var rc ResultCode
		if err := json.Unmarshal([]byte(input), &rc); err == nil {
			t.Errorf("Unmarshal(%s) expected error, got %v", input, rc)
		}
	}
}

func TestDocument_JSON_roundTrip(t *testing.T) {
	original := &Document{
		Name:           "Test Open",
		City:           "Amsterdam",
		TotalRounds:    7,
		TournamentType: "Swiss Dutch",
		InitialColor:   "white1",
		NumPlayers:     2,
		Players: []PlayerLine{
			{
				StartNumber: 1,
				Name:        "Alice",
				Rating:      2200,
				Federation:  "NED",
				FideID:      "12345",
				Points:      1.5,
				Rank:        1,
				Rounds: []RoundResult{
					{Opponent: 2, Color: ColorWhite, Result: ResultWin},
					{Opponent: 0, Color: ColorNone, Result: ResultHalfBye},
				},
			},
			{
				StartNumber: 2,
				Name:        "Bob",
				Rating:      1800,
				Points:      0.5,
				Rank:        2,
				Rounds: []RoundResult{
					{Opponent: 1, Color: ColorBlack, Result: ResultLoss},
					{Opponent: 0, Color: ColorNone, Result: ResultFullBye},
				},
			},
		},
		ForbiddenPairs: []ForbiddenPair{{Player1: 1, Player2: 2}},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var restored Document
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify key fields survived the round-trip.
	if restored.Name != original.Name {
		t.Errorf("Name = %q, want %q", restored.Name, original.Name)
	}
	if restored.TotalRounds != original.TotalRounds {
		t.Errorf("TotalRounds = %d, want %d", restored.TotalRounds, original.TotalRounds)
	}
	if len(restored.Players) != len(original.Players) {
		t.Fatalf("len(Players) = %d, want %d", len(restored.Players), len(original.Players))
	}

	p := restored.Players[0]
	if p.StartNumber != 1 || p.Name != "Alice" || p.Rating != 2200 {
		t.Errorf("Player 1 = {%d, %q, %d}, want {1, Alice, 2200}", p.StartNumber, p.Name, p.Rating)
	}
	if len(p.Rounds) != 2 {
		t.Fatalf("Player 1 rounds = %d, want 2", len(p.Rounds))
	}
	if p.Rounds[0].Color != ColorWhite || p.Rounds[0].Result != ResultWin || p.Rounds[0].Opponent != 2 {
		t.Errorf("Player 1 round 1 = %+v, want {2 White Win}", p.Rounds[0])
	}

	if len(restored.ForbiddenPairs) != 1 || restored.ForbiddenPairs[0].Player1 != 1 {
		t.Errorf("ForbiddenPairs = %+v, want [{1 2}]", restored.ForbiddenPairs)
	}
}

func TestDocument_JSON_omitsEmptyFields(t *testing.T) {
	doc := &Document{
		Players: []PlayerLine{
			{StartNumber: 1, Name: "Alice", Points: 0, Rank: 1},
		},
	}

	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Unmarshal to a map to check which fields are present.
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}

	// These optional fields should be absent.
	for _, field := range []string{"name", "city", "federation", "startDate", "endDate", "timeControl"} {
		if _, exists := m[field]; exists {
			t.Errorf("field %q should be omitted when empty", field)
		}
	}

	// Players should be present (non-empty slice).
	if _, exists := m["players"]; !exists {
		t.Error("players should be present")
	}
}
