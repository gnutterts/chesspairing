package burstein

import (
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
)

// OppositionIndex holds the three components used to re-rank players
// after seeding rounds in the Burstein system.
//
// Per C.04.4.2: players are re-ranked by opposition index, which is
// computed as Buchholz → Sonneborn-Berger → TPN (as final tiebreak).
type OppositionIndex struct {
	Buchholz        float64 // sum of opponents' scores
	SonnebornBerger float64 // sum of (score vs opponent × opponent's score)
	TPN             int     // tournament pairing number (lower = higher ranked)
}

// ComputeOppositionIndex computes the opposition index for a single player.
//
// Buchholz = sum of opponents' pairing scores (standard 1-½-0).
// Sonneborn-Berger = sum of (result-against-opponent × opponent's score).
// TPN = current tournament pairing number (tiebreak of last resort).
func ComputeOppositionIndex(player *swisslib.PlayerState, state *chesspairing.TournamentState) OppositionIndex {
	// Build score map for all players (including inactive, for Buchholz).
	scores := computePairingScores(state)

	var buchholz float64
	for _, oppID := range player.Opponents {
		buchholz += scores[oppID]
	}

	// Build per-opponent result map for Sonneborn-Berger.
	var sb float64
	for _, round := range state.Rounds {
		for _, game := range round.Games {
			if game.IsForfeit {
				continue
			}
			var resultForPlayer float64
			var oppID string

			switch {
			case game.WhiteID == player.ID:
				oppID = game.BlackID
				switch game.Result {
				case chesspairing.ResultWhiteWins:
					resultForPlayer = 1.0
				case chesspairing.ResultDraw:
					resultForPlayer = 0.5
				default:
					resultForPlayer = 0.0
				}
			case game.BlackID == player.ID:
				oppID = game.WhiteID
				switch game.Result {
				case chesspairing.ResultBlackWins:
					resultForPlayer = 1.0
				case chesspairing.ResultDraw:
					resultForPlayer = 0.5
				default:
					resultForPlayer = 0.0
				}
			default:
				continue
			}

			sb += resultForPlayer * scores[oppID]
		}
	}

	return OppositionIndex{
		Buchholz:        buchholz,
		SonnebornBerger: sb,
		TPN:             player.TPN,
	}
}

// RankByOppositionIndex re-ranks players by opposition index and assigns
// new TPN values. Players are sorted by:
//  1. Score descending (primary, same as standard ranking)
//  2. Buchholz descending (higher opposition strength = better)
//  3. Sonneborn-Berger descending (better results against stronger opponents)
//  4. Original TPN ascending (tiebreak of last resort)
//
// After sorting, new TPN values are assigned sequentially (1, 2, 3, ...).
func RankByOppositionIndex(players []swisslib.PlayerState, state *chesspairing.TournamentState) []swisslib.PlayerState {
	// Compute opposition index for each player.
	indices := make(map[string]OppositionIndex, len(players))
	for i := range players {
		indices[players[i].ID] = ComputeOppositionIndex(&players[i], state)
	}

	// Sort by score desc, then opposition index components.
	sorted := make([]swisslib.PlayerState, len(players))
	copy(sorted, players)

	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]

		// 1. Score descending.
		if a.Score != b.Score {
			return a.Score > b.Score
		}

		idxA, idxB := indices[a.ID], indices[b.ID]

		// 2. Buchholz descending.
		if idxA.Buchholz != idxB.Buchholz {
			return idxA.Buchholz > idxB.Buchholz
		}

		// 3. Sonneborn-Berger descending.
		if idxA.SonnebornBerger != idxB.SonnebornBerger {
			return idxA.SonnebornBerger > idxB.SonnebornBerger
		}

		// 4. Original TPN ascending (lower = higher ranked).
		return idxA.TPN < idxB.TPN
	})

	// Assign new TPN values.
	for i := range sorted {
		sorted[i].TPN = i + 1
	}

	return sorted
}

// computePairingScores builds a map of player ID → pairing score (standard 1-½-0)
// for all players in the tournament state (including inactive, for Buchholz).
func computePairingScores(state *chesspairing.TournamentState) map[string]float64 {
	scores := make(map[string]float64)

	for _, round := range state.Rounds {
		for _, game := range round.Games {
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				scores[game.WhiteID] += 1.0
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				scores[game.BlackID] += 1.0
			case chesspairing.ResultDraw:
				scores[game.WhiteID] += 0.5
				scores[game.BlackID] += 0.5
			}
		}
		for _, bye := range round.Byes {
			scores[bye.PlayerID] += 1.0
		}
	}

	return scores
}
