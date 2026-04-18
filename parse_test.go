// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package chesspairing_test

import (
	"testing"

	cp "github.com/gnutterts/chesspairing"
)

func TestParseScoringSystem(t *testing.T) {
	cases := []struct {
		in      string
		want    cp.ScoringSystem
		wantErr bool
	}{
		{"standard", cp.ScoringStandard, false},
		{"Standard", cp.ScoringStandard, false},
		{"  STANDARD  ", cp.ScoringStandard, false},
		{"keizer", cp.ScoringKeizer, false},
		{"KEIZER", cp.ScoringKeizer, false},
		{"football", cp.ScoringFootball, false},
		{"", "", true},
		{"   ", "", true},
		{"swiss", "", true},
		{"standardx", "", true},
	}
	for _, c := range cases {
		got, err := cp.ParseScoringSystem(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseScoringSystem(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("ParseScoringSystem(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseScoringSystem_RoundTrip(t *testing.T) {
	for _, s := range []cp.ScoringSystem{cp.ScoringStandard, cp.ScoringKeizer, cp.ScoringFootball} {
		got, err := cp.ParseScoringSystem(string(s))
		if err != nil {
			t.Errorf("ParseScoringSystem(%q): %v", s, err)
		}
		if got != s {
			t.Errorf("round-trip %q = %q", s, got)
		}
	}
}

func TestParsePairingSystem(t *testing.T) {
	cases := []struct {
		in      string
		want    cp.PairingSystem
		wantErr bool
	}{
		{"dutch", cp.PairingDutch, false},
		{"Dutch", cp.PairingDutch, false},
		{"DUTCH", cp.PairingDutch, false},
		{"fide-dutch", cp.PairingDutch, false},
		{"FIDE-Dutch", cp.PairingDutch, false},
		{"  dutch  ", cp.PairingDutch, false},
		{"burstein", cp.PairingBurstein, false},
		{"fide-burstein", cp.PairingBurstein, false},
		{"dubov", cp.PairingDubov, false},
		{"fide-dubov", cp.PairingDubov, false},
		{"lim", cp.PairingLim, false},
		{"fide-lim", cp.PairingLim, false},
		{"doubleswiss", cp.PairingDoubleSwiss, false},
		{"double-swiss", cp.PairingDoubleSwiss, false},
		{"team", cp.PairingTeam, false},
		{"keizer", cp.PairingKeizer, false},
		{"roundrobin", cp.PairingRoundRobin, false},
		{"round-robin", cp.PairingRoundRobin, false},
		{"rr", cp.PairingRoundRobin, false},
		{"RR", cp.PairingRoundRobin, false},
		{"", "", true},
		{"   ", "", true},
		{"unknown", "", true},
		{"fide-swiss", "", true},
	}
	for _, c := range cases {
		got, err := cp.ParsePairingSystem(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParsePairingSystem(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("ParsePairingSystem(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParsePairingSystem_RoundTrip(t *testing.T) {
	all := []cp.PairingSystem{
		cp.PairingDutch, cp.PairingBurstein, cp.PairingDubov, cp.PairingLim,
		cp.PairingDoubleSwiss, cp.PairingTeam, cp.PairingKeizer, cp.PairingRoundRobin,
	}
	for _, s := range all {
		got, err := cp.ParsePairingSystem(string(s))
		if err != nil {
			t.Errorf("ParsePairingSystem(%q): %v", s, err)
		}
		if got != s {
			t.Errorf("round-trip %q = %q", s, got)
		}
	}
}

func TestParseGameResult(t *testing.T) {
	cases := []struct {
		in      string
		want    cp.GameResult
		wantErr bool
	}{
		{"1-0", cp.ResultWhiteWins, false},
		{"1 - 0", cp.ResultWhiteWins, false},
		{"  1-0  ", cp.ResultWhiteWins, false},
		{"0-1", cp.ResultBlackWins, false},
		{"0 - 1", cp.ResultBlackWins, false},
		{"0.5-0.5", cp.ResultDraw, false},
		{"0.5 - 0.5", cp.ResultDraw, false},
		{"1/2-1/2", cp.ResultDraw, false},
		{"1/2 - 1/2", cp.ResultDraw, false},
		{"½-½", cp.ResultDraw, false},
		{"*", cp.ResultPending, false},
		{"1-0f", cp.ResultForfeitWhiteWins, false},
		{"1-0F", cp.ResultForfeitWhiteWins, false},
		{"1 - 0 f", cp.ResultForfeitWhiteWins, false},
		{"0-1f", cp.ResultForfeitBlackWins, false},
		{"0-0f", cp.ResultDoubleForfeit, false},
		{"", "", true},
		{"   ", "", true},
		{"2-0", "", true},
		{"draw", "", true},
		{"win", "", true},
	}
	for _, c := range cases {
		got, err := cp.ParseGameResult(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseGameResult(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("ParseGameResult(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseGameResult_RoundTrip(t *testing.T) {
	all := []cp.GameResult{
		cp.ResultWhiteWins, cp.ResultBlackWins, cp.ResultDraw, cp.ResultPending,
		cp.ResultForfeitWhiteWins, cp.ResultForfeitBlackWins, cp.ResultDoubleForfeit,
	}
	for _, r := range all {
		got, err := cp.ParseGameResult(string(r))
		if err != nil {
			t.Errorf("ParseGameResult(%q): %v", r, err)
		}
		if got != r {
			t.Errorf("round-trip %q = %q", r, got)
		}
	}
}

func TestParseByeType(t *testing.T) {
	cases := []struct {
		in      string
		want    cp.ByeType
		wantErr bool
	}{
		{"PAB", cp.ByePAB, false},
		{"pab", cp.ByePAB, false},
		{"  PAB  ", cp.ByePAB, false},
		{"F", cp.ByePAB, false},
		{"f", cp.ByePAB, false},
		{"Half", cp.ByeHalf, false},
		{"half", cp.ByeHalf, false},
		{"H", cp.ByeHalf, false},
		{"Zero", cp.ByeZero, false},
		{"Z", cp.ByeZero, false},
		{"Absent", cp.ByeAbsent, false},
		{"U", cp.ByeAbsent, false},
		{"Excused", cp.ByeExcused, false},
		{"excused", cp.ByeExcused, false},
		{"ClubCommitment", cp.ByeClubCommitment, false},
		{"clubcommitment", cp.ByeClubCommitment, false},
		{"CLUBCOMMITMENT", cp.ByeClubCommitment, false},
		{"", 0, true},
		{"   ", 0, true},
		{"X", 0, true},
		{"unknown", 0, true},
	}
	for _, c := range cases {
		got, err := cp.ParseByeType(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("ParseByeType(%q) err=%v wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if got != c.want {
			t.Errorf("ParseByeType(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseByeType_RoundTripFromString(t *testing.T) {
	all := []cp.ByeType{
		cp.ByePAB, cp.ByeHalf, cp.ByeZero,
		cp.ByeAbsent, cp.ByeExcused, cp.ByeClubCommitment,
	}
	for _, b := range all {
		got, err := cp.ParseByeType(b.String())
		if err != nil {
			t.Errorf("ParseByeType(%q): %v", b.String(), err)
		}
		if got != b {
			t.Errorf("round-trip %v = %v", b, got)
		}
	}
}
