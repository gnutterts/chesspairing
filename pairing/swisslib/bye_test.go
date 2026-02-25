package swisslib

import "testing"

func TestDutchByeSelector_LowestRankedNoBye(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 1.0, ByeReceived: false},
		{ID: "p2", TPN: 2, Score: 1.0, ByeReceived: false},
		{ID: "p3", TPN: 3, Score: 0.0, ByeReceived: false},
		{ID: "p4", TPN: 4, Score: 0.0, ByeReceived: false},
		{ID: "p5", TPN: 5, Score: 0.0, ByeReceived: false},
	}
	sel := DutchByeSelector{}
	bye := sel.SelectBye(players)
	if bye == nil {
		t.Fatal("expected a bye player")
	}
	// Dutch: lowest score group → highest TPN → p5
	if bye.ID != "p5" {
		t.Errorf("expected p5 (lowest ranked in lowest score group), got %s", bye.ID)
	}
}

func TestDutchByeSelector_SkipsAlreadyReceivedBye(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 0.0, ByeReceived: false},
		{ID: "p2", TPN: 2, Score: 0.0, ByeReceived: false},
		{ID: "p3", TPN: 3, Score: 0.0, ByeReceived: true}, // already had bye
	}
	sel := DutchByeSelector{}
	bye := sel.SelectBye(players)
	if bye == nil {
		t.Fatal("expected a bye player")
	}
	if bye.ID != "p2" {
		t.Errorf("p3 had bye, next lowest is p2, got %s", bye.ID)
	}
}

func TestDutchByeSelector_AllHadBye(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 0.0, ByeReceived: true},
	}
	sel := DutchByeSelector{}
	bye := sel.SelectBye(players)
	// When all have had a bye, return nil (shouldn't normally happen).
	if bye != nil {
		t.Errorf("all had bye, expected nil, got %s", bye.ID)
	}
}

func TestDutchByeSelector_EvenPlayers(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 0.0, ByeReceived: false},
		{ID: "p2", TPN: 2, Score: 0.0, ByeReceived: false},
	}
	sel := DutchByeSelector{}
	bye := sel.SelectBye(players)
	// Even number — shouldn't be called, but if it is, still works.
	if bye == nil {
		t.Fatal("expected a bye player even with even count")
	}
}

func TestBursteinByeSelector_LowestScore_MostGames_LowestRank(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 2.0, ByeReceived: false},
		{ID: "p2", TPN: 2, Score: 1.0, ByeReceived: false},
		{ID: "p3", TPN: 3, Score: 0.0, ByeReceived: false},
		{ID: "p4", TPN: 4, Score: 0.0, ByeReceived: false},
	}
	// Give p3 more games played than p4 by adding color history.
	p3 := players[2]
	p3.ColorHistory = []Color{ColorWhite, ColorBlack} // 2 games
	p4 := players[3]
	p4.ColorHistory = []Color{ColorWhite} // 1 game

	sel := BursteinByeSelector{}
	bye := sel.SelectBye(players)
	if bye == nil {
		t.Fatal("expected a bye player")
	}
	// Burstein priority: 1. lowest score (p3,p4=0.0), 2. most games (p3=2),
	// 3. lowest ranking/highest TPN (p3=TPN3 vs p4 would only apply if games tied).
	// p3 has more games → p3 gets bye.
	if bye.ID != "p3" {
		t.Errorf("expected p3 (lowest score, most games), got %s", bye.ID)
	}
}

func TestBursteinByeSelector_TieGoesToHighestTPN(t *testing.T) {
	players := []*PlayerState{
		{ID: "p1", TPN: 1, Score: 0.0, ByeReceived: false, ColorHistory: []Color{ColorWhite}},
		{ID: "p2", TPN: 2, Score: 0.0, ByeReceived: false, ColorHistory: []Color{ColorBlack}},
		{ID: "p3", TPN: 3, Score: 0.0, ByeReceived: false, ColorHistory: []Color{ColorWhite}},
	}
	sel := BursteinByeSelector{}
	bye := sel.SelectBye(players)
	if bye == nil {
		t.Fatal("expected a bye player")
	}
	// All same score, same number of games → highest TPN (p3)
	if bye.ID != "p3" {
		t.Errorf("expected p3 (highest TPN for tiebreak), got %s", bye.ID)
	}
}

func TestNeedsBye(t *testing.T) {
	if NeedsBye(4) {
		t.Error("4 players should not need bye")
	}
	if !NeedsBye(5) {
		t.Error("5 players should need bye")
	}
	if NeedsBye(0) {
		t.Error("0 players should not need bye")
	}
	if !NeedsBye(1) {
		t.Error("1 player should need bye")
	}
}
