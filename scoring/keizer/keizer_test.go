package keizer

import (
	"context"
	"math"
	"testing"

	chesspairing "github.com/gnutterts/chesspairing"
)

func TestNew(t *testing.T) {
	s := New(Options{})
	if s == nil {
		t.Fatal("New returned nil")
	}
}

func TestNewFromMap(t *testing.T) {
	s := NewFromMap(map[string]any{
		"winFraction":  0.8,
		"drawFraction": 0.4,
	})
	if s == nil {
		t.Fatal("NewFromMap returned nil")
	}
	if *s.opts.WinFraction != 0.8 {
		t.Errorf("WinFraction = %v, want 0.8", *s.opts.WinFraction)
	}
}

func TestScoreNoPlayers(t *testing.T) {
	s := New(Options{})
	scores, err := s.Score(context.Background(), &chesspairing.TournamentState{})
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("expected no scores, got %d", len(scores))
	}
}

func TestScoreNoRounds(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 3 {
		t.Fatalf("expected 3 scores, got %d", len(scores))
	}
	// With no rounds, all scores should be zero, ranked by rating.
	for _, ps := range scores {
		if ps.Score != 0 {
			t.Errorf("player %s score = %v, want 0", ps.PlayerID, ps.Score)
		}
	}
	if scores[0].PlayerID != "p1" {
		t.Errorf("rank 1 = %s, want p1 (highest rated)", scores[0].PlayerID)
	}
	if scores[1].PlayerID != "p2" {
		t.Errorf("rank 2 = %s, want p2", scores[1].PlayerID)
	}
	if scores[2].PlayerID != "p3" {
		t.Errorf("rank 3 = %s, want p3 (lowest rated)", scores[2].PlayerID)
	}
}

func TestScoreOneRoundAllPlay(t *testing.T) {
	// 4 players, 1 round: p1 beats p2, p3 beats p4.
	// Default: base=4, step=1. Win=1.0, Loss=0.0. SelfVictory=true.
	// Initial ranking by rating: p1(val4), p2(val3), p3(val2), p4(val1).
	//
	// p1 wins p2(val3): scoreX2(3,1.0) = 6
	// p3 wins p4(val1): scoreX2(1,1.0) = 2
	// Self-victory: p1+=4×2=8, p2+=3×2=6, p3+=2×2=4, p4+=1×2=2
	// TotalX2: p1=14, p2=6, p3=6, p4=2. Scores: p1=7, p2=3, p3=3, p4=1.
	// Ranking: p1(7) > p2(3)=p3(3) > p4(1). p2 vs p3: rating 1800>1600.
	// Converges immediately (ranking unchanged).
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 4 {
		t.Fatalf("expected 4 scores, got %d", len(scores))
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// With self-victory, even losers get their own value.
	assertScore(t, scoreMap, "p1", 7.0)
	assertScore(t, scoreMap, "p2", 3.0) // self-victory only (lost game = 0)
	assertScore(t, scoreMap, "p3", 3.0) // win(val1=1) + self(val2=2)
	assertScore(t, scoreMap, "p4", 1.0) // self-victory only
	assertRank(t, scoreMap, "p1", 1)
	assertRank(t, scoreMap, "p2", 2) // tiebreak by rating over p3
	assertRank(t, scoreMap, "p3", 3)
	assertRank(t, scoreMap, "p4", 4)
}

func TestScoreDraws(t *testing.T) {
	// 2 players draw: each gets 50% of opponent's value + self-victory.
	// N=2, base=2, step=1. SelfVictory=true.
	//
	// Iter 0: Initial ranking: p1(rank1,val2), p2(rank2,val1).
	//   p1 draws p2(val1): scoreX2(1,0.5)=1. p2 draws p1(val2): scoreX2(2,0.5)=2.
	//   Self: p1+=2×2=4, p2+=1×2=2.
	//   TotalX2: p1=5, p2=4. Scores: p1=2.5, p2=2.0.
	//   Re-rank: p1(5) > p2(4). Ranking: [p1,p2]. Same as initial → converged.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// With self-victory, higher-valued player benefits more.
	// p1 converges at rank 1 due to self-victory advantage.
	assertScore(t, scoreMap, "p1", 2.5)
	assertScore(t, scoreMap, "p2", 2.0)
	assertRank(t, scoreMap, "p1", 1)
	assertRank(t, scoreMap, "p2", 2)
}

func TestScoreAbsentPlayer(t *testing.T) {
	// 4 players, 2 rounds. p1 beats p3, p2 beats p4 in round 1.
	// Round 2: p1 beats p2, p3 and p4 are absent.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p3", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				// p3, p4 absent from round 2
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// With self-victory + absence penalty (0.35), absent players still get some points.
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("absent player p3 score = %v, want > 0", scoreMap["p3"].Score)
	}
	if scoreMap["p4"].Score <= 0 {
		t.Errorf("absent player p4 score = %v, want > 0", scoreMap["p4"].Score)
	}
	// p1 won both rounds — should be ranked first.
	assertRank(t, scoreMap, "p1", 1)
	// p2 won round 1 and lost round 2 — should outscore at least one absent player.
	if scoreMap["p2"].Rank >= scoreMap["p3"].Rank && scoreMap["p2"].Rank >= scoreMap["p4"].Rank {
		t.Errorf("p2 rank = %d, expected to outscore at least one absent player", scoreMap["p2"].Rank)
	}
}

func TestScoreByePlayer(t *testing.T) {
	// 3 players, 1 round. p1 plays p2, p3 gets a bye.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByePAB}},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// p3 gets bye: should have positive score (fraction of own value + self-victory).
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("bye player p3 score = %v, want > 0", scoreMap["p3"].Score)
	}
}

func TestScoreInactivePlayersExcluded(t *testing.T) {
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: false}, // withdrawn
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("expected 1 score (inactive excluded), got %d", len(scores))
	}
	if scores[0].PlayerID != "p1" {
		t.Errorf("expected p1, got %s", scores[0].PlayerID)
	}
}

func TestScoreCustomOptions(t *testing.T) {
	// Custom: draws worth 40%, absence penalty 0, self-victory OFF.
	draw := 0.4
	absent := 0.0
	selfVictory := false
	s := New(Options{
		DrawFraction:          &draw,
		AbsentPenaltyFraction: &absent,
		SelfVictory:           &selfVictory,
	})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
				},
				// p3 absent, but penalty is 0.
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, ps := range scores {
		scoreMap[ps.PlayerID] = ps
	}

	// p3 absent with 0 penalty and no self-victory → score should be 0.
	assertScore(t, scoreMap, "p3", 0)
}

func TestScoreMultipleRounds(t *testing.T) {
	// 4 players, 2 rounds. Tests that scoring accumulates across rounds
	// and rankings can shift.
	s := New(Options{})
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultWhiteWins}, // p2 bounces back
					{WhiteID: "p4", BlackID: "p1", Result: chesspairing.ResultBlackWins}, // p1 wins again
				},
			},
		},
	}
	scores, err := s.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score error: %v", err)
	}
	if len(scores) != 4 {
		t.Fatalf("expected 4 scores, got %d", len(scores))
	}

	// p1 won both games — should be rank 1.
	if scores[0].PlayerID != "p1" {
		t.Errorf("rank 1 = %s, want p1 (won both games)", scores[0].PlayerID)
	}
	// All ranks should be 1-4.
	for i, ps := range scores {
		if ps.Rank != i+1 {
			t.Errorf("scores[%d].Rank = %d, want %d", i, ps.Rank, i+1)
		}
	}
}

func TestPointsForResultWin(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 5,
	}
	pts := s.PointsForResult(chesspairing.ResultWhiteWins, rctx)
	// scoreX2(5, 1.0) = 10, /2 = 5.0
	if pts != 5.0 {
		t.Errorf("PointsForResult(win) = %v, want 5.0", pts)
	}
}

func TestPointsForResultDraw(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 4,
	}
	pts := s.PointsForResult(chesspairing.ResultDraw, rctx)
	// scoreX2(4, 0.5) = 4, /2 = 2.0
	if pts != 2.0 {
		t.Errorf("PointsForResult(draw) = %v, want 2.0", pts)
	}
}

func TestPointsForResultAbsent(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		PlayerValueNumber: 6,
		IsAbsent:          true,
	}
	pts := s.PointsForResult(chesspairing.ResultPending, rctx)
	// scoreX2(6, 0.35) = round(6 * 0.35 * 2) = round(4.2) = 4, /2 = 2.0
	if pts != 2.0 {
		t.Errorf("PointsForResult(absent) = %v, want 2.0", pts)
	}
}

func TestPointsForResultBye(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		PlayerValueNumber: 4,
		IsBye:             true,
	}
	pts := s.PointsForResult(chesspairing.ResultPending, rctx)
	// scoreX2(4, 0.50) = round(4 * 0.50 * 2) = 4, /2 = 2.0
	if pts != 2.0 {
		t.Errorf("PointsForResult(bye) = %v, want 2.0 (0.50 of 4)", pts)
	}
}

func TestPointsForResultForfeitWin(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 5,
	}
	pts := s.PointsForResult(chesspairing.ResultForfeitWhiteWins, rctx)
	// scoreX2(5, 1.0) = 10, /2 = 5.0
	if pts != 5.0 {
		t.Errorf("PointsForResult(forfeit win) = %v, want 5.0", pts)
	}
}

func TestPointsForResultDoubleForfeit(t *testing.T) {
	s := New(Options{})
	rctx := chesspairing.ResultContext{
		OpponentValueNumber: 5,
	}
	pts := s.PointsForResult(chesspairing.ResultDoubleForfeit, rctx)
	// scoreX2(5, 0.0) = 0, /2 = 0.0
	if pts != 0.0 {
		t.Errorf("PointsForResult(double forfeit) = %v, want 0.0", pts)
	}
}

// TestOptionsWithDefaults verifies that defaults are applied correctly.
func TestOptionsWithDefaults(t *testing.T) {
	o := Options{}.WithDefaults(10)
	if *o.ValueNumberBase != 10 {
		t.Errorf("ValueNumberBase = %d, want 10", *o.ValueNumberBase)
	}
	if *o.ValueNumberStep != 1 {
		t.Errorf("ValueNumberStep = %d, want 1", *o.ValueNumberStep)
	}
	if *o.WinFraction != 1.0 {
		t.Errorf("WinFraction = %v, want 1.0", *o.WinFraction)
	}
	if *o.DrawFraction != 0.5 {
		t.Errorf("DrawFraction = %v, want 0.5", *o.DrawFraction)
	}
	if *o.LossFraction != 0.0 {
		t.Errorf("LossFraction = %v, want 0.0", *o.LossFraction)
	}
	if *o.ForfeitWinFraction != 1.0 {
		t.Errorf("ForfeitWinFraction = %v, want 1.0", *o.ForfeitWinFraction)
	}
	if *o.ForfeitLossFraction != 0.0 {
		t.Errorf("ForfeitLossFraction = %v, want 0.0", *o.ForfeitLossFraction)
	}
	if *o.DoubleForfeitFraction != 0.0 {
		t.Errorf("DoubleForfeitFraction = %v, want 0.0", *o.DoubleForfeitFraction)
	}
	if *o.AbsentPenaltyFraction != 0.35 {
		t.Errorf("AbsentPenaltyFraction = %v, want 0.35", *o.AbsentPenaltyFraction)
	}
	if *o.ByeValueFraction != 0.50 {
		t.Errorf("ByeValueFraction = %v, want 0.50", *o.ByeValueFraction)
	}
	if *o.HalfByeFraction != 0.50 {
		t.Errorf("HalfByeFraction = %v, want 0.50", *o.HalfByeFraction)
	}
	if *o.ZeroByeFraction != 0.0 {
		t.Errorf("ZeroByeFraction = %v, want 0.0", *o.ZeroByeFraction)
	}
	if *o.ExcusedAbsentFraction != 0.35 {
		t.Errorf("ExcusedAbsentFraction = %v, want 0.35", *o.ExcusedAbsentFraction)
	}
	if *o.ClubCommitmentFraction != 0.70 {
		t.Errorf("ClubCommitmentFraction = %v, want 0.70", *o.ClubCommitmentFraction)
	}
	if *o.SelfVictory != true {
		t.Errorf("SelfVictory = %v, want true", *o.SelfVictory)
	}
	if *o.AbsenceLimit != 5 {
		t.Errorf("AbsenceLimit = %d, want 5", *o.AbsenceLimit)
	}
	if *o.AbsenceDecay != false {
		t.Errorf("AbsenceDecay = %v, want false", *o.AbsenceDecay)
	}

	// Fixed-value overrides should remain nil (not defaulted).
	if o.ByeFixedValue != nil {
		t.Errorf("ByeFixedValue = %v, want nil", o.ByeFixedValue)
	}
	if o.AbsentFixedValue != nil {
		t.Errorf("AbsentFixedValue = %v, want nil", o.AbsentFixedValue)
	}
}

// TestOptionsWithDefaultsPreservesExplicit verifies that explicit values are kept.
func TestOptionsWithDefaultsPreservesExplicit(t *testing.T) {
	win := 0.75
	base := 20
	selfVictory := false
	o := Options{
		WinFraction:     &win,
		ValueNumberBase: &base,
		SelfVictory:     &selfVictory,
	}.WithDefaults(10)
	if *o.WinFraction != 0.75 {
		t.Errorf("WinFraction = %v, want 0.75 (explicit)", *o.WinFraction)
	}
	if *o.ValueNumberBase != 20 {
		t.Errorf("ValueNumberBase = %d, want 20 (explicit, not default 10)", *o.ValueNumberBase)
	}
	if *o.SelfVictory != false {
		t.Errorf("SelfVictory = %v, want false (explicit)", *o.SelfVictory)
	}
}

func TestValueNumber(t *testing.T) {
	base := 10
	step := 1
	o := Options{ValueNumberBase: &base, ValueNumberStep: &step}
	tests := []struct {
		rank int
		want int
	}{
		{1, 10},
		{2, 9},
		{5, 6},
		{10, 1},
	}
	for _, tt := range tests {
		got := o.ValueNumber(tt.rank)
		if got != tt.want {
			t.Errorf("ValueNumber(%d) = %d, want %d", tt.rank, got, tt.want)
		}
	}
}

func TestParseOptions(t *testing.T) {
	m := map[string]any{
		"valueNumberBase":        12,
		"absentPenaltyFraction":  0.3,
		"winFraction":            0.9,
		"selfVictory":            false,
		"absenceLimit":           3,
		"absenceDecay":           true,
		"clubCommitmentFraction": 0.65,
		"unknownField":           "ignored",
	}
	o := ParseOptions(m)
	if o.ValueNumberBase == nil || *o.ValueNumberBase != 12 {
		t.Errorf("ValueNumberBase = %v, want 12", o.ValueNumberBase)
	}
	if o.AbsentPenaltyFraction == nil || *o.AbsentPenaltyFraction != 0.3 {
		t.Errorf("AbsentPenaltyFraction = %v, want 0.3", o.AbsentPenaltyFraction)
	}
	if o.WinFraction == nil || *o.WinFraction != 0.9 {
		t.Errorf("WinFraction = %v, want 0.9", o.WinFraction)
	}
	if o.SelfVictory == nil || *o.SelfVictory != false {
		t.Errorf("SelfVictory = %v, want false", o.SelfVictory)
	}
	if o.AbsenceLimit == nil || *o.AbsenceLimit != 3 {
		t.Errorf("AbsenceLimit = %v, want 3", o.AbsenceLimit)
	}
	if o.AbsenceDecay == nil || *o.AbsenceDecay != true {
		t.Errorf("AbsenceDecay = %v, want true", o.AbsenceDecay)
	}
	if o.ClubCommitmentFraction == nil || *o.ClubCommitmentFraction != 0.65 {
		t.Errorf("ClubCommitmentFraction = %v, want 0.65", o.ClubCommitmentFraction)
	}
	// Fields not in the map should remain nil.
	if o.DrawFraction != nil {
		t.Errorf("DrawFraction = %v, want nil (not in map)", o.DrawFraction)
	}
}

func TestParseOptionsFixedValues(t *testing.T) {
	m := map[string]any{
		"byeFixedValue":            15,
		"absentFixedValue":         10,
		"excusedAbsentFixedValue":  10,
		"clubCommitmentFixedValue": 25,
	}
	o := ParseOptions(m)
	if o.ByeFixedValue == nil || *o.ByeFixedValue != 15 {
		t.Errorf("ByeFixedValue = %v, want 15", o.ByeFixedValue)
	}
	if o.AbsentFixedValue == nil || *o.AbsentFixedValue != 10 {
		t.Errorf("AbsentFixedValue = %v, want 10", o.AbsentFixedValue)
	}
	if o.ExcusedAbsentFixedValue == nil || *o.ExcusedAbsentFixedValue != 10 {
		t.Errorf("ExcusedAbsentFixedValue = %v, want 10", o.ExcusedAbsentFixedValue)
	}
	if o.ClubCommitmentFixedValue == nil || *o.ClubCommitmentFixedValue != 25 {
		t.Errorf("ClubCommitmentFixedValue = %v, want 25", o.ClubCommitmentFixedValue)
	}
}

func TestScoreExactConvergence(t *testing.T) {
	// 4 players, 2 rounds. Self-victory ON (default).
	// Ratings: p1=2000, p2=1800, p3=1600, p4=1400
	// Round 1: p1 beats p4 (1-0), p2 draws p3 (½-½)
	// Round 2: p1 draws p2 (½-½), p3 beats p4 (1-0)
	//
	// N=4, base=4, step=1. Win=1.0, Draw=0.5, Loss=0.0.
	//
	// ITERATION 0:
	//   Ranking: p1(rank1,val4), p2(rank2,val3), p3(rank3,val2), p4(rank4,val1)
	//   R1: p1 wins p4(1)→2, p4 loss→0, p2 draws p3(2)→2, p3 draws p2(3)→3
	//   (scoreX2: p1 wins: round(1*1.0*2)=2, p2 draws: round(2*0.5*2)=2, p3 draws: round(3*0.5*2)=3)
	//   R2: p1 draws p2(3)→3, p2 draws p1(4)→4, p3 wins p4(1)→2, p4 loss→0
	//   Game totals X2: p1=2+3=5, p2=2+4=6, p3=3+2=5, p4=0
	//   Self X2: p1+=4*2=8, p2+=3*2=6, p3+=2*2=4, p4+=1*2=2
	//   Total X2: p1=13, p2=12, p3=9, p4=2
	//   Re-rank: p1(13)>p2(12)>p3(9)>p4(2). Ranking: [p1,p2,p3,p4].
	//   Same as initial → converged!
	//
	// Final scores: p1=6.5, p2=6, p3=4.5, p4=1

	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}

	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 6.5)
	assertScore(t, scoreMap, "p2", 6.0)
	assertScore(t, scoreMap, "p3", 4.5)
	assertScore(t, scoreMap, "p4", 1.0)

	assertRank(t, scoreMap, "p1", 1)
	assertRank(t, scoreMap, "p2", 2)
	assertRank(t, scoreMap, "p3", 3)
	assertRank(t, scoreMap, "p4", 4)
}

func TestScoreExactConvergenceNoSelfVictory(t *testing.T) {
	// Same scenario as TestScoreExactConvergence but with SelfVictory=false.
	// This should match the old behavior (pre-self-victory).
	//
	// ITERATION 0: Ranking: p1(val4), p2(val3), p3(val2), p4(val1)
	//   R1: p1 wins p4(1)→X2=2, p2 draws p3(2)→X2=2, p3 draws p2(3)→X2=3
	//   R2: p1 draws p2(3)→X2=3, p2 draws p1(4)→X2=4, p3 wins p4(1)→X2=2
	//   Total X2: p1=5, p2=6, p3=5, p4=0
	//   Scores: p1=2.5, p2=3, p3=2.5, p4=0
	//   Re-rank: p2(6)>p1(5)=p3(5)>p4(0). p1 vs p3: rating 2000>1600.
	//   New ranking: [p2,p1,p3,p4]. Different from [p1,p2,p3,p4].
	//
	// ITERATION 1: Ranking: p2(val4), p1(val3), p3(val2), p4(val1)
	//   R1: p1 wins p4(1)→X2=2, p2 draws p3(2)→X2=2, p3 draws p2(4)→X2=4
	//   R2: p1 draws p2(4)→X2=4, p2 draws p1(3)→X2=3, p3 wins p4(1)→X2=2
	//   Total X2: p1=6, p2=5, p3=6, p4=0
	//   Scores: p1=3, p2=2.5, p3=3, p4=0
	//   Re-rank: p1(6)=p3(6)>p2(5)>p4(0). p1 vs p3: rating. [p1,p3,p2,p4].
	//
	// ITERATION 2: Ranking: p1(val4), p3(val3), p2(val2), p4(val1)
	//   R1: p1 wins p4(1)→X2=2, p2 draws p3(3)→X2=3, p3 draws p2(2)→X2=2
	//   R2: p1 draws p2(2)→X2=2, p2 draws p1(4)→X2=4, p3 wins p4(1)→X2=2
	//   Total X2: p1=4, p2=7, p3=4, p4=0
	//   Scores: p1=2, p2=3.5, p3=2, p4=0
	//   Re-rank: p2(7)>p1(4)=p3(4)>p4(0). [p2,p1,p3,p4].
	//   twoAgoRanking (iter 0 output) = [p2,p1,p3,p4] == current → oscillation!
	//
	// Average iter1 and iter2 X2: p1=(6+4)/2=5, p2=(5+7)/2=6, p3=(6+4)/2=5, p4=0
	// Scores: p1=2.5, p2=3, p3=2.5, p4=0.
	// Final: [p2,p1,p3,p4]

	selfVictory := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
					{WhiteID: "p2", BlackID: "p3", Result: chesspairing.ResultDraw},
				},
			},
			{
				Number: 2,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
					{WhiteID: "p3", BlackID: "p4", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}

	scorer := New(Options{SelfVictory: &selfVictory})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 2.5)
	assertScore(t, scoreMap, "p2", 3.0)
	assertScore(t, scoreMap, "p3", 2.5)
	assertScore(t, scoreMap, "p4", 0.0)

	assertRank(t, scoreMap, "p2", 1)
	assertRank(t, scoreMap, "p1", 2)
	assertRank(t, scoreMap, "p3", 3)
	assertRank(t, scoreMap, "p4", 4)
}

func TestScoreWithForfeit(t *testing.T) {
	// 2 players, 1 round: forfeit white wins.
	// N=2, base=2, step=1. SelfVictory=true.
	// Initial: p1(rank1,val2), p2(rank2,val1).
	// p1 forfeit wins p2(val1) × ForfeitWinFraction(1.0): scoreX2(1,1.0)=2.
	// p2 forfeit loses = ForfeitLossFraction(0.0): scoreX2(2,0.0)=0.
	// Self: p1+=2*2=4, p2+=1*2=2.
	// Total X2: p1=6, p2=2. Scores: p1=3, p2=1. Converges immediately.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{
						WhiteID:   "p1",
						BlackID:   "p2",
						Result:    chesspairing.ResultForfeitWhiteWins,
						IsForfeit: true,
					},
				},
			},
		},
	}

	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 3.0)
	assertScore(t, scoreMap, "p2", 1.0)
}

func TestScoreWithForfeitCustomFractions(t *testing.T) {
	// Custom forfeit fractions: forfeit win=0.5, forfeit loss=0.1.
	// SelfVictory OFF to isolate the forfeit scoring.
	// 2 players, N=2. p1(val2), p2(val1).
	// p1 forfeit wins p2(val1): scoreX2(1, 0.5) = 1.
	// p2 forfeit loses p1(val2): scoreX2(2, 0.1) = round(2*0.1*2) = round(0.4) = 0.
	// Total X2: p1=1, p2=0. Scores: p1=0.5, p2=0.
	forfeitWin := 0.5
	forfeitLoss := 0.1
	selfVictory := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{
						WhiteID:   "p1",
						BlackID:   "p2",
						Result:    chesspairing.ResultForfeitWhiteWins,
						IsForfeit: true,
					},
				},
			},
		},
	}

	scorer := New(Options{
		ForfeitWinFraction:  &forfeitWin,
		ForfeitLossFraction: &forfeitLoss,
		SelfVictory:         &selfVictory,
	})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 0.5) // scoreX2(1,0.5)=1 → 0.5
	assertScore(t, scoreMap, "p2", 0.0)
}

func TestScoreWithDoubleForfeit(t *testing.T) {
	// 4 players, 1 round: p1 beats p2 normally, p3 vs p4 double forfeit.
	// N=4, base=4, step=1. SelfVictory=true. DoubleForfeitFraction=0.0 (default).
	// p1 wins p2(val3): scoreX2(3,1.0)=6.
	// p3 double forfeit p4(val1): scoreX2(1,0.0)=0. p4 ditto: scoreX2(2,0.0)=0.
	// Self: p1+=4*2=8, p2+=3*2=6, p3+=2*2=4, p4+=1*2=2.
	// Total X2: p1=14, p2=6, p3=4, p4=2.
	// Scores: p1=7, p2=3, p3=2, p4=1.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
					{
						WhiteID:   "p3",
						BlackID:   "p4",
						Result:    chesspairing.ResultDoubleForfeit,
						IsForfeit: true,
					},
				},
			},
		},
	}

	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// Double forfeit: 0 game points, but still have self-victory.
	assertScore(t, scoreMap, "p3", 2.0) // self only: val2 = 2
	assertScore(t, scoreMap, "p4", 1.0) // self only: val1 = 1
	assertScore(t, scoreMap, "p1", 7.0) // win(3) + self(4) = 7
}

func TestScoreExactBye(t *testing.T) {
	// 3 players, 1 round: p1 beats p2, p3 gets PAB bye.
	// N=3, base=3, step=1. Default bye fraction = 0.50. SelfVictory=true.
	//
	// Iter 0: Ranking: p1(rank1,val3), p2(rank2,val2), p3(rank3,val1).
	//   p1 wins p2(val2): scoreX2(2,1.0)=4. p3 bye: scoreX2(1,0.50)=1.
	//   Self: p1+=3*2=6, p2+=2*2=4, p3+=1*2=2.
	//   Total X2: p1=10, p2=4, p3=3. Scores: p1=5, p2=2, p3=1.5.
	//   Re-rank: [p1,p2,p3]. Same → converged.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByePAB},
				},
			},
		},
	}

	scorer := New(Options{})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 5.0)
	assertScore(t, scoreMap, "p2", 2.0)
	assertScore(t, scoreMap, "p3", 1.5) // bye(0.5) + self(1) = 1.5
	assertRank(t, scoreMap, "p1", 1)
	assertRank(t, scoreMap, "p2", 2)
	assertRank(t, scoreMap, "p3", 3)
}

// ---------- NEW FEATURE TESTS ----------

func TestScoreSelfVictoryOnVsOff(t *testing.T) {
	// 2 players, 1 round: p1 beats p2.
	// With self-victory: p1 gets win + own value, p2 gets own value.
	// Without self-victory: p1 gets win only, p2 gets 0.
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}

	// With self-victory (default).
	scorer := New(Options{})
	scores, _ := scorer.Score(context.Background(), state)
	scoreMapOn := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMapOn[s.PlayerID] = s
	}
	// p1: win(val1=1) + self(val2=2) = 3.
	// Wait — N=2, initial ranking p1(rank1,val2), p2(rank2,val1).
	// p1 wins p2(val1): scoreX2(1,1.0)=2. Self: p1+=2*2=4, p2+=1*2=2.
	// Total X2: p1=6, p2=2. Scores: p1=3, p2=1.
	assertScore(t, scoreMapOn, "p1", 3.0)
	assertScore(t, scoreMapOn, "p2", 1.0) // self-victory only

	// Without self-victory.
	selfVictory := false
	scorer2 := New(Options{SelfVictory: &selfVictory})
	scores2, _ := scorer2.Score(context.Background(), state)
	scoreMapOff := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores2 {
		scoreMapOff[s.PlayerID] = s
	}
	// Without self-victory, N=2. p1(rank1,val2), p2(rank2,val1).
	// p1 wins p2(val1): scoreX2(1,1.0)=2. /2=1.0. p2=0.
	// Converges immediately.
	assertScore(t, scoreMapOff, "p1", 1.0)
	assertScore(t, scoreMapOff, "p2", 0.0)
}

func TestScoreClubCommitmentByeType(t *testing.T) {
	// 3 players, 1 round: p1 beats p2, p3 has club commitment.
	// ClubCommitmentFraction default = 0.70.
	// SelfVictory OFF for clarity.
	selfVictory := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByeClubCommitment},
				},
			},
		},
	}

	// Iter 0: p1(val3), p2(val2), p3(val1).
	// p1 wins p2(val2): scoreX2(2,1.0)=4. p3 club: scoreX2(1,0.70)=round(1*0.70*2)=round(1.4)=1.
	// Total X2: p1=4, p2=0, p3=1.
	// Scores: p1=2, p2=0, p3=0.5.
	// Re-rank: [p1,p3,p2].
	//
	// Iter 1: p1(val3), p3(val2), p2(val1).
	// p1 wins p2(val1): scoreX2(1,1.0)=2. p3 club: scoreX2(2,0.70)=round(2*0.70*2)=round(2.8)=3.
	// Total X2: p1=2, p2=0, p3=3.
	// Scores: p1=1, p2=0, p3=1.5.
	// Re-rank: [p3,p1,p2].
	//
	// Iter 2: p3(val3), p1(val2), p2(val1).
	// p1 wins p2(val1): scoreX2(1,1.0)=2. p3 club: scoreX2(3,0.70)=round(3*0.70*2)=round(4.2)=4.
	// Total X2: p1=2, p2=0, p3=4.
	// Re-rank: [p3,p1,p2]. Same as iter1 → converged!
	// Scores: p1=1, p2=0, p3=2.

	scorer := New(Options{SelfVictory: &selfVictory})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p3", 2.0) // club commitment at 70% of val3=3
	assertScore(t, scoreMap, "p1", 1.0)
	assertScore(t, scoreMap, "p2", 0.0)
	assertRank(t, scoreMap, "p3", 1) // club commitment pushes p3 to top
}

func TestScoreExcusedAbsenceByeType(t *testing.T) {
	// 3 players, 1 round: p1 beats p2, p3 has excused absence.
	// ExcusedAbsentFraction default = 0.35. SelfVictory OFF.
	selfVictory := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByeExcused},
				},
			},
		},
	}

	scorer := New(Options{SelfVictory: &selfVictory})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// p3 should have a positive score from excused absence.
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("excused absent p3 score = %v, want > 0", scoreMap["p3"].Score)
	}
}

func TestScoreFixedValueOverride(t *testing.T) {
	// Fixed bye value overrides the fraction calculation.
	// SelfVictory OFF for clarity.
	selfVictory := false
	byeFixed := 10
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByePAB},
				},
			},
		},
	}

	scorer := New(Options{SelfVictory: &selfVictory, ByeFixedValue: &byeFixed})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// p3 gets fixed bye value = 10, regardless of own Keizer value.
	// fixedX2(10) = 20. /2 = 10.0.
	assertScore(t, scoreMap, "p3", 10.0)
}

func TestScoreAbsenceLimit(t *testing.T) {
	// 4 players, 7 rounds. p3 is absent for all 7 rounds.
	// AbsenceLimit=5 (default). Only first 5 absences score.
	// SelfVictory OFF for clarity.
	selfVictory := false
	absenceLimit := 5

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
	}

	var rounds []chesspairing.RoundData
	for i := 1; i <= 7; i++ {
		rounds = append(rounds, chesspairing.RoundData{
			Number: i,
			Games: []chesspairing.GameData{
				{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw},
			},
			// p3 absent, p4 absent
		})
	}

	state := &chesspairing.TournamentState{
		Players: players,
		Rounds:  rounds,
	}

	scorer := New(Options{
		SelfVictory:  &selfVictory,
		AbsenceLimit: &absenceLimit,
	})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// p3 and p4 are absent for 7 rounds, but only 5 count.
	// Verify that 6th and 7th absences don't add more points than 5 absences would.
	// Compute expected: with limit=5, only 5 rounds of absence score.

	// Also run with limit=0 (unlimited) to compare.
	unlimitedLimit := 0
	scorer2 := New(Options{
		SelfVictory:  &selfVictory,
		AbsenceLimit: &unlimitedLimit,
	})
	scores2, _ := scorer2.Score(context.Background(), state)
	scoreMapUnlimited := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores2 {
		scoreMapUnlimited[s.PlayerID] = s
	}

	// With unlimited absences (7 rounds scoring), should be more than with limit 5.
	if scoreMapUnlimited["p3"].Score <= scoreMap["p3"].Score {
		t.Errorf("unlimited absence score (%v) should be > limited score (%v)",
			scoreMapUnlimited["p3"].Score, scoreMap["p3"].Score)
	}
}

func TestScoreAbsenceDecay(t *testing.T) {
	// 3 players, 3 rounds. p3 absent all 3 rounds.
	// AbsenceDecay=true, AbsenceLimit=0 (unlimited). SelfVictory OFF.
	// 1st absence: full, 2nd: /2, 3rd: /4.
	selfVictory := false
	decay := true
	noLimit := 0

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
	}

	rounds := []chesspairing.RoundData{
		{Number: 1, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}}},
		{Number: 2, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}}},
		{Number: 3, Games: []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}}},
	}

	state := &chesspairing.TournamentState{Players: players, Rounds: rounds}

	// With decay.
	scorer := New(Options{SelfVictory: &selfVictory, AbsenceDecay: &decay, AbsenceLimit: &noLimit})
	scores, _ := scorer.Score(context.Background(), state)
	var decayScore float64
	for _, s := range scores {
		if s.PlayerID == "p3" {
			decayScore = s.Score
		}
	}

	// Without decay.
	noDecay := false
	scorer2 := New(Options{SelfVictory: &selfVictory, AbsenceDecay: &noDecay, AbsenceLimit: &noLimit})
	scores2, _ := scorer2.Score(context.Background(), state)
	var noDecayScore float64
	for _, s := range scores2 {
		if s.PlayerID == "p3" {
			noDecayScore = s.Score
		}
	}

	// With decay, total should be less than without decay.
	if decayScore >= noDecayScore {
		t.Errorf("decay score (%v) should be < no-decay score (%v)", decayScore, noDecayScore)
	}
	// Decay score should still be > 0 (first absence is full).
	if decayScore <= 0 {
		t.Errorf("decay score = %v, want > 0", decayScore)
	}
}

func TestScoreClubCommitmentExemptFromLimitAndDecay(t *testing.T) {
	// Club commitments are NOT subject to absence limit or decay.
	// 3 players, 7 rounds. p3 has club commitment for all 7 rounds.
	// AbsenceLimit=5, AbsenceDecay=true. SelfVictory OFF.
	selfVictory := false
	limit := 5
	decay := true

	players := []chesspairing.PlayerEntry{
		{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
		{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
	}

	var rounds []chesspairing.RoundData
	for i := 1; i <= 7; i++ {
		rounds = append(rounds, chesspairing.RoundData{
			Number: i,
			Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}},
			Byes:   []chesspairing.ByeEntry{{PlayerID: "p3", Type: chesspairing.ByeClubCommitment}},
		})
	}

	state := &chesspairing.TournamentState{Players: players, Rounds: rounds}

	scorer := New(Options{
		SelfVictory:  &selfVictory,
		AbsenceLimit: &limit,
		AbsenceDecay: &decay,
	})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	var p3Score float64
	for _, s := range scores {
		if s.PlayerID == "p3" {
			p3Score = s.Score
		}
	}

	// All 7 club commitments should score — none skipped by limit, none decayed.
	// If limit+decay were applied, score would be much lower.
	// With 0.70 fraction and no limit/decay, 7 rounds × 0.70 × own_value.
	if p3Score <= 0 {
		t.Errorf("p3 club commitment score = %v, want > 0", p3Score)
	}

	// Compare: same scenario but as regular absences (should be much less).
	var roundsAbsent []chesspairing.RoundData
	for i := 1; i <= 7; i++ {
		roundsAbsent = append(roundsAbsent, chesspairing.RoundData{
			Number: i,
			Games:  []chesspairing.GameData{{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultDraw}},
			// p3 implicitly absent
		})
	}
	stateAbsent := &chesspairing.TournamentState{Players: players, Rounds: roundsAbsent}
	scoresAbsent, _ := scorer.Score(context.Background(), stateAbsent)
	var p3AbsentScore float64
	for _, s := range scoresAbsent {
		if s.PlayerID == "p3" {
			p3AbsentScore = s.Score
		}
	}

	if p3Score <= p3AbsentScore {
		t.Errorf("club commitment score (%v) should be > regular absent score (%v)",
			p3Score, p3AbsentScore)
	}
}

func TestScoreHalfAndZeroBye(t *testing.T) {
	// Test half-bye and zero-bye scoring.
	// SelfVictory OFF for clarity.
	selfVictory := false
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
			{ID: "p4", DisplayName: "Dave", Rating: 1400, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				Byes: []chesspairing.ByeEntry{
					{PlayerID: "p3", Type: chesspairing.ByeHalf},
					{PlayerID: "p4", Type: chesspairing.ByeZero},
				},
			},
		},
	}

	scorer := New(Options{SelfVictory: &selfVictory})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// p3 half-bye: HalfByeFraction=0.50.
	// p4 zero-bye: ZeroByeFraction=0.0.
	if scoreMap["p3"].Score <= 0 {
		t.Errorf("p3 (half-bye) score = %v, want > 0", scoreMap["p3"].Score)
	}
	assertScore(t, scoreMap, "p4", 0.0)
}

func TestScoreLossFraction(t *testing.T) {
	// FreeKeizer-style: loss gives 1/6 of opponent's value.
	// SelfVictory OFF. 2 players, 1 round.
	selfVictory := false
	lossFrac := 1.0 / 6.0
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
			},
		},
	}

	// N=2. p1(rank1,val2), p2(rank2,val1).
	// p1 wins p2(val1): scoreX2(1,1.0)=2. p2 loses p1(val2): scoreX2(2,1/6)=round(2*1/6*2)=round(0.667)=1.
	// Total X2: p1=2, p2=1. Scores: p1=1, p2=0.5.
	scorer := New(Options{SelfVictory: &selfVictory, LossFraction: &lossFrac})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	assertScore(t, scoreMap, "p1", 1.0)
	assertScore(t, scoreMap, "p2", 0.5) // toughness bonus!
}

func TestScoreX2Helper(t *testing.T) {
	tests := []struct {
		value    int
		fraction float64
		want     int
	}{
		{10, 1.0, 20},     // 10 * 1.0 * 2 = 20
		{10, 0.5, 10},     // 10 * 0.5 * 2 = 10
		{3, 0.35, 2},      // 3 * 0.35 * 2 = 2.1 → 2
		{4, 0.35, 3},      // 4 * 0.35 * 2 = 2.8 → 3
		{1, 0.70, 1},      // 1 * 0.70 * 2 = 1.4 → 1
		{3, 0.70, 4},      // 3 * 0.70 * 2 = 4.2 → 4
		{5, 1.0 / 6.0, 2}, // 5 * 1/6 * 2 = 1.667 → 2
		{0, 1.0, 0},       // zero value
		{10, 0.0, 0},      // zero fraction
	}
	for _, tt := range tests {
		got := scoreX2(tt.value, tt.fraction)
		if got != tt.want {
			t.Errorf("scoreX2(%d, %v) = %d, want %d", tt.value, tt.fraction, got, tt.want)
		}
	}
}

func TestFixedX2Helper(t *testing.T) {
	if fixedX2(10) != 20 {
		t.Errorf("fixedX2(10) = %d, want 20", fixedX2(10))
	}
	if fixedX2(0) != 0 {
		t.Errorf("fixedX2(0) = %d, want 0", fixedX2(0))
	}
}

func TestScoreAbsentFixedValueOverride(t *testing.T) {
	// Fixed absent value = 5, regardless of player's Keizer value.
	// SelfVictory OFF. 3 players, 1 round, p3 absent.
	selfVictory := false
	absentFixed := 5
	state := &chesspairing.TournamentState{
		Players: []chesspairing.PlayerEntry{
			{ID: "p1", DisplayName: "Alice", Rating: 2000, Active: true},
			{ID: "p2", DisplayName: "Bob", Rating: 1800, Active: true},
			{ID: "p3", DisplayName: "Carol", Rating: 1600, Active: true},
		},
		Rounds: []chesspairing.RoundData{
			{
				Number: 1,
				Games: []chesspairing.GameData{
					{WhiteID: "p1", BlackID: "p2", Result: chesspairing.ResultWhiteWins},
				},
				// p3 implicitly absent
			},
		},
	}

	scorer := New(Options{SelfVictory: &selfVictory, AbsentFixedValue: &absentFixed})
	scores, err := scorer.Score(context.Background(), state)
	if err != nil {
		t.Fatalf("Score: %v", err)
	}

	scoreMap := make(map[string]chesspairing.PlayerScore)
	for _, s := range scores {
		scoreMap[s.PlayerID] = s
	}

	// p3 gets fixed absent value = 5. fixedX2(5) = 10. /2 = 5.0.
	assertScore(t, scoreMap, "p3", 5.0)
}

func TestGetBool(t *testing.T) {
	m := map[string]any{
		"trueVal":  true,
		"falseVal": false,
		"notBool":  "hello",
	}
	if v, ok := getBool(m, "trueVal"); !ok || v != true {
		t.Errorf("getBool(trueVal) = %v, %v, want true, true", v, ok)
	}
	if v, ok := getBool(m, "falseVal"); !ok || v != false {
		t.Errorf("getBool(falseVal) = %v, %v, want false, true", v, ok)
	}
	if _, ok := getBool(m, "notBool"); ok {
		t.Error("getBool(notBool) should return false for non-bool")
	}
	if _, ok := getBool(m, "missing"); ok {
		t.Error("getBool(missing) should return false for missing key")
	}
}

// ---------- HELPERS ----------

func assertScore(t *testing.T, scoreMap map[string]chesspairing.PlayerScore, playerID string, want float64) {
	t.Helper()
	got := scoreMap[playerID].Score
	if math.Abs(got-want) > 0.001 {
		t.Errorf("%s score = %v, want %v", playerID, got, want)
	}
}

func assertRank(t *testing.T, scoreMap map[string]chesspairing.PlayerScore, playerID string, want int) {
	t.Helper()
	got := scoreMap[playerID].Rank
	if got != want {
		t.Errorf("%s rank = %d, want %d", playerID, got, want)
	}
}
