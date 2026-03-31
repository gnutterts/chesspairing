package chesspairing_test

import (
	"testing"

	"github.com/gnutterts/chesspairing"
)

func TestGameResult_IsValid(t *testing.T) {
	valid := []chesspairing.GameResult{
		chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins,
		chesspairing.ResultDraw, chesspairing.ResultPending,
		chesspairing.ResultForfeitWhiteWins, chesspairing.ResultForfeitBlackWins,
		chesspairing.ResultDoubleForfeit,
	}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("IsValid(%q) = false, want true", r)
		}
	}
	if chesspairing.GameResult("invalid").IsValid() {
		t.Error("IsValid(invalid) = true, want false")
	}
}

func TestGameResult_IsRecordable(t *testing.T) {
	recordable := []chesspairing.GameResult{
		chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins,
		chesspairing.ResultDraw,
		chesspairing.ResultForfeitWhiteWins, chesspairing.ResultForfeitBlackWins,
		chesspairing.ResultDoubleForfeit,
	}
	for _, r := range recordable {
		if !r.IsRecordable() {
			t.Errorf("IsRecordable(%q) = false, want true", r)
		}
	}
	if chesspairing.ResultPending.IsRecordable() {
		t.Error("IsRecordable(pending) = true, want false")
	}
}

func TestGameResult_IsForfeit(t *testing.T) {
	forfeits := []chesspairing.GameResult{
		chesspairing.ResultForfeitWhiteWins,
		chesspairing.ResultForfeitBlackWins,
		chesspairing.ResultDoubleForfeit,
	}
	for _, r := range forfeits {
		if !r.IsForfeit() {
			t.Errorf("IsForfeit(%q) = false, want true", r)
		}
	}
	nonForfeits := []chesspairing.GameResult{
		chesspairing.ResultWhiteWins, chesspairing.ResultBlackWins,
		chesspairing.ResultDraw, chesspairing.ResultPending,
	}
	for _, r := range nonForfeits {
		if r.IsForfeit() {
			t.Errorf("IsForfeit(%q) = true, want false", r)
		}
	}
}

func TestGameResult_IsDoubleForfeit(t *testing.T) {
	if !chesspairing.ResultDoubleForfeit.IsDoubleForfeit() {
		t.Error("IsDoubleForfeit(0-0f) = false, want true")
	}
	if chesspairing.ResultForfeitWhiteWins.IsDoubleForfeit() {
		t.Error("IsDoubleForfeit(1-0f) = true, want false")
	}
}

func TestPairingDubovIsValid(t *testing.T) {
	if !chesspairing.PairingDubov.IsValid() {
		t.Error("PairingDubov should be valid")
	}
}

func TestDefaultTiebreakersDubov(t *testing.T) {
	tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDubov)
	if len(tbs) == 0 {
		t.Error("Dubov should have default tiebreakers")
	}
}

func TestPairingLimIsValid(t *testing.T) {
	if !chesspairing.PairingLim.IsValid() {
		t.Error("PairingLim should be valid")
	}
}

func TestDefaultTiebreakersLim(t *testing.T) {
	tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingLim)
	if len(tbs) == 0 {
		t.Error("Lim should have default tiebreakers")
	}
	// Lim is a Swiss system — same tiebreakers as Dutch/Burstein/Dubov.
	expected := []string{"buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"}
	if len(tbs) != len(expected) {
		t.Errorf("expected %d tiebreakers, got %d", len(expected), len(tbs))
	}
	for i, tb := range tbs {
		if i < len(expected) && tb != expected[i] {
			t.Errorf("tiebreaker %d: expected %q, got %q", i, expected[i], tb)
		}
	}
}

func TestPairingDoubleSwissIsValid(t *testing.T) {
	if !chesspairing.PairingDoubleSwiss.IsValid() {
		t.Error("PairingDoubleSwiss should be valid")
	}
}

func TestDefaultTiebreakersDoubleSwiss(t *testing.T) {
	tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDoubleSwiss)
	if len(tbs) == 0 {
		t.Error("Double-Swiss should have default tiebreakers")
	}
	// Double-Swiss is a Swiss system — same tiebreakers as Dutch/Burstein/Dubov/Lim.
	expected := []string{"buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"}
	if len(tbs) != len(expected) {
		t.Errorf("expected %d tiebreakers, got %d", len(expected), len(tbs))
	}
	for i, tb := range tbs {
		if i < len(expected) && tb != expected[i] {
			t.Errorf("tiebreaker %d: expected %q, got %q", i, expected[i], tb)
		}
	}
}

func TestPairingTeamIsValid(t *testing.T) {
	if !chesspairing.PairingTeam.IsValid() {
		t.Error("PairingTeam should be valid")
	}
}

func TestDefaultTiebreakersTeam(t *testing.T) {
	tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingTeam)
	if len(tbs) == 0 {
		t.Error("Team Swiss should have default tiebreakers")
	}
	// Team Swiss uses the same tiebreakers as other Swiss systems.
	expected := []string{"buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"}
	if len(tbs) != len(expected) {
		t.Errorf("expected %d tiebreakers, got %d", len(expected), len(tbs))
	}
	for i, tb := range tbs {
		if i < len(expected) && tb != expected[i] {
			t.Errorf("tiebreaker %d: expected %q, got %q", i, expected[i], tb)
		}
	}
}

func TestByeType_IsValid(t *testing.T) {
	valid := []chesspairing.ByeType{
		chesspairing.ByePAB, chesspairing.ByeHalf,
		chesspairing.ByeZero, chesspairing.ByeAbsent,
		chesspairing.ByeExcused, chesspairing.ByeClubCommitment,
	}
	for _, bt := range valid {
		if !bt.IsValid() {
			t.Errorf("IsValid(%v) = false, want true", bt)
		}
	}
	if chesspairing.ByeType(-1).IsValid() {
		t.Error("IsValid(-1) = true, want false")
	}
	if chesspairing.ByeType(6).IsValid() {
		t.Error("IsValid(6) = true, want false")
	}
}

func TestByeType_String(t *testing.T) {
	tests := []struct {
		bt   chesspairing.ByeType
		want string
	}{
		{chesspairing.ByePAB, "PAB"},
		{chesspairing.ByeHalf, "Half"},
		{chesspairing.ByeZero, "Zero"},
		{chesspairing.ByeAbsent, "Absent"},
		{chesspairing.ByeExcused, "Excused"},
		{chesspairing.ByeClubCommitment, "ClubCommitment"},
		{chesspairing.ByeType(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.bt.String(); got != tt.want {
			t.Errorf("String(%d) = %q, want %q", tt.bt, got, tt.want)
		}
	}
}
