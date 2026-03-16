package dubov

import (
	"testing"

	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

func TestComputeARO(t *testing.T) {
	ratings := map[string]int{
		"a": 2000,
		"b": 1800,
		"c": 2200,
		"d": 1600,
	}

	tests := []struct {
		name      string
		opponents []string
		want      float64
	}{
		{
			name:      "no opponents",
			opponents: nil,
			want:      0,
		},
		{
			name:      "one opponent",
			opponents: []string{"b"},
			want:      1800,
		},
		{
			name:      "two opponents",
			opponents: []string{"b", "c"},
			want:      2000, // (1800 + 2200) / 2
		},
		{
			name:      "three opponents",
			opponents: []string{"b", "c", "d"},
			want:      1866.6666666666667, // (1800 + 2200 + 1600) / 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := &swisslib.PlayerState{
				ID:        "a",
				Opponents: tt.opponents,
			}
			got := ComputeARO(player, ratings)
			if got != tt.want {
				t.Errorf("ComputeARO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildRatingMap(t *testing.T) {
	players := []swisslib.PlayerState{
		{ID: "a", Rating: 2000},
		{ID: "b", Rating: 1800},
	}
	m := BuildRatingMap(players)
	if m["a"] != 2000 {
		t.Errorf("expected 2000, got %d", m["a"])
	}
	if m["b"] != 1800 {
		t.Errorf("expected 1800, got %d", m["b"])
	}
}

func TestSortByAROAscending(t *testing.T) {
	ratings := map[string]int{
		"a": 2000,
		"b": 1800,
		"c": 2200,
	}

	players := []*swisslib.PlayerState{
		{ID: "a", Opponents: []string{"c"}, TPN: 1},      // ARO = 2200
		{ID: "b", Opponents: []string{"a"}, TPN: 2},      // ARO = 2000
		{ID: "c", Opponents: []string{"a", "b"}, TPN: 3}, // ARO = 1900
	}

	SortByAROAscending(players, ratings)

	// Expected order: c (1900), b (2000), a (2200)
	if players[0].ID != "c" || players[1].ID != "b" || players[2].ID != "a" {
		t.Errorf("unexpected order: %s, %s, %s", players[0].ID, players[1].ID, players[2].ID)
	}
}

func TestSortByAROAscendingTiebreakByTPN(t *testing.T) {
	ratings := map[string]int{
		"a": 2000,
		"b": 2000,
	}

	players := []*swisslib.PlayerState{
		{ID: "p1", Opponents: []string{"b"}, TPN: 3}, // ARO = 2000, TPN 3
		{ID: "p2", Opponents: []string{"a"}, TPN: 1}, // ARO = 2000, TPN 1
	}

	SortByAROAscending(players, ratings)

	// Same ARO -> ascending TPN
	if players[0].ID != "p2" || players[1].ID != "p1" {
		t.Errorf("unexpected order: %s, %s", players[0].ID, players[1].ID)
	}
}
