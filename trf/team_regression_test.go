package trf

import (
	"bytes"
	"strings"
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestMixedTeamMatchesRoundTrip(t *testing.T) {
	state := &cp.TournamentState{
		Players:       []cp.PlayerEntry{{ID: "1", TeamID: "1"}, {ID: "2", TeamID: "1"}, {ID: "3", TeamID: "2"}, {ID: "4", TeamID: "2"}},
		PairingConfig: cp.PairingConfig{System: cp.PairingTeam},
		Rounds: []cp.RoundData{
			{Number: 1, Matches: []cp.MatchData{{HomeID: "1", AwayID: "2", Boards: []cp.GameData{
				{WhiteID: "1", BlackID: "3", Result: cp.ResultWhiteWins}, {WhiteID: "4", BlackID: "2", Result: cp.ResultWhiteWins},
			}}}},
			{Number: 2, Matches: []cp.MatchData{{HomeID: "2", AwayID: "1", Result: &cp.TeamMatchResult{HomeGame: 1.5, AwayGame: 0.5}}}},
			{Number: 3, Matches: []cp.MatchData{{HomeID: "1", AwayID: "2", Boards: []cp.GameData{
				{WhiteID: "1", BlackID: "3", Result: cp.ResultDraw}, {WhiteID: "4", BlackID: "2", Result: cp.ResultBlackWins},
			}}}},
		},
	}
	doc, _, err := FromTournamentStateWithError(state)
	if err != nil {
		t.Fatal(err)
	}
	var data bytes.Buffer
	if err := Write(&data, doc); err != nil {
		t.Fatal(err)
	}
	t.Logf("\n%s", data.String())
	read, err := Read(bytes.NewReader(data.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	rt, err := read.ToTournamentState()
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range rt.Rounds {
		t.Logf("round %d matches=%+v teamByes=%+v", i+1, r.Matches, r.TeamByes)
		for _, mm := range r.Matches {
			if mm.Result != nil {
				t.Logf("  result %+v", *mm.Result)
			}
		}
	}
	if len(rt.Rounds[1].TeamByes) != 0 {
		t.Errorf("spurious team byes in round 2: %+v", rt.Rounds[1].TeamByes)
	}
	for i, round := range rt.Rounds {
		if len(round.Matches) != 1 {
			t.Errorf("round %d has %d matches, want 1", i+1, len(round.Matches))
		}
	}
	if got := rt.Rounds[2].Matches[0]; got.HomeID != "1" || got.AwayID != "2" || len(got.Boards) != 2 {
		t.Errorf("round 3 match = %+v, want its original board match", got)
	}
}

func TestWriteTeamLineWithoutNameKeepsHeaderWidth(t *testing.T) {
	var buf bytes.Buffer
	if err := writeTeamLine(&buf, TeamLine{TeamNumber: 3}); err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSuffix(buf.String(), "\n")
	if len(line) != 40 {
		t.Fatalf("line length = %d, want 40: %q", len(line), line)
	}
	if _, err := parseTeamLine(line, 1); err != nil {
		t.Fatalf("parseTeamLine() error = %v", err)
	}
}

func TestWriteBlankRecordsStayReadable(t *testing.T) {
	var buf bytes.Buffer
	if err := writeNewTeamLine(&buf, NewTeamLine{}); err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSuffix(buf.String(), "\n")
	if _, err := parseNewTeamLine(line, 1); err != nil {
		t.Fatalf("parseNewTeamLine(%q) error = %v", line, err)
	}
}

func TestWriteTeamMemberOutOfRange(t *testing.T) {
	var buf bytes.Buffer
	if err := writeTeamLine(&buf, TeamLine{TeamNumber: 1, Members: []int{10000}}); err == nil {
		t.Fatal("writeTeamLine() error = nil, want an overflow error")
	}
	if err := writeNewTeamLine(&buf, NewTeamLine{Members: []int{-1}}); err == nil {
		t.Fatal("writeNewTeamLine() error = nil, want an overflow error")
	}
}

func TestWriteTeamMembersDoNotRunTogether(t *testing.T) {
	var buf bytes.Buffer
	if err := writeTeamLine(&buf, TeamLine{TeamNumber: 1, Members: []int{1, 1000}}); err != nil {
		t.Fatal(err)
	}
	got, err := parseTeamLine(strings.TrimSuffix(buf.String(), "\n"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Members) != 2 || got.Members[0] != 1 || got.Members[1] != 1000 {
		t.Fatalf("members = %v, want [1 1000]", got.Members)
	}
}
