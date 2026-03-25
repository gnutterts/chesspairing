// Package keizer implements Keizer-style pairing for chess tournaments.
//
// Keizer pairing works by ranking players by their current Keizer points
// (or rating if no rounds have been played), then pairing from the outside
// in: the top-ranked player plays the bottom-ranked, second plays
// second-from-bottom, and so on. The middle player gets a bye if there's
// an odd number of players.
//
// Repeat avoidance: by default, players must wait at least 3 rounds before
// being paired against the same opponent again. When a conflict occurs,
// the lower-ranked player swaps with the nearest available player.
//
// Color assignment: the higher-ranked player gets white, unless they had
// white in their most recent game (then colors are swapped for balance).
package keizer

import (
	"context"
	"sort"

	chesspairing "github.com/gnutterts/chesspairing"
)

// Pairer implements the chesspairing.Pairer interface for Keizer pairing.
type Pairer struct {
	opts Options
}

// New creates a new Keizer pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts}
}

// NewFromMap creates a new Keizer pairer from a map[string]any config.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Pair generates pairings for the next round using the Keizer method.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	opts := p.opts.WithDefaults()

	// Get active players.
	active := activePlayerIDs(state.Players)
	if len(active) < 2 {
		// Not enough players to pair.
		result := &chesspairing.PairingResult{}
		if len(active) == 1 {
			result.Byes = []chesspairing.ByeEntry{{PlayerID: active[0], Type: chesspairing.ByePAB}}
			result.Notes = []string{active[0] + " receives a bye (only player)"}
		}
		return result, nil
	}

	// Build player entries lookup.
	entries := make(map[string]chesspairing.PlayerEntry, len(state.Players))
	for _, pl := range state.Players {
		entries[pl.ID] = pl
	}

	// Rank players: by Keizer score (if rounds exist) or by rating.
	ranked := rankPlayers(active, state, entries)

	// Build pairing history for repeat avoidance.
	history := buildHistory(state.Rounds)

	// Build last-color map for color balance.
	lastColor := buildLastColor(state.Rounds)

	// Pair from outside in.
	return pairRanked(ranked, opts, history, lastColor, state.CurrentRound), nil
}

// activePlayerIDs returns IDs of active players.
func activePlayerIDs(players []chesspairing.PlayerEntry) []string {
	ids := make([]string, 0, len(players))
	for _, p := range players {
		if p.Active {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// rankPlayers returns player IDs sorted by score (descending) if rounds exist,
// otherwise by rating (descending).
func rankPlayers(ids []string, state *chesspairing.TournamentState, entries map[string]chesspairing.PlayerEntry) []string {
	ranked := make([]string, len(ids))
	copy(ranked, ids)

	if len(state.Rounds) == 0 {
		// No rounds: sort by rating descending.
		sort.Slice(ranked, func(i, j int) bool {
			ri := entries[ranked[i]].Rating
			rj := entries[ranked[j]].Rating
			if ri != rj {
				return ri > rj
			}
			return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
		})
		return ranked
	}

	// Compute simple game points for ranking (wins=1, draws=0.5).
	// This is used for pairing order, not for standings.
	gamePoints := computeGamePoints(ids, state.Rounds)

	sort.Slice(ranked, func(i, j int) bool {
		si := gamePoints[ranked[i]]
		sj := gamePoints[ranked[j]]
		if si != sj {
			return si > sj
		}
		// Tiebreak by rating.
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// computeGamePoints calculates simple game points (1-½-0) for pairing ranking.
func computeGamePoints(ids []string, rounds []chesspairing.RoundData) map[string]float64 {
	points := make(map[string]float64, len(ids))
	for _, round := range rounds {
		for _, game := range round.Games {
			// Double forfeits are excluded entirely.
			if game.Result.IsDoubleForfeit() {
				continue
			}
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				points[game.WhiteID] += 1.0
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				points[game.BlackID] += 1.0
			case chesspairing.ResultDraw:
				points[game.WhiteID] += 0.5
				points[game.BlackID] += 0.5
			}
		}
	}
	return points
}

// pairingHistory tracks which round each pair of players last played.
type pairingHistory map[string]map[string]int // playerA → playerB → round number

// buildHistory builds the pairing history from completed rounds.
func buildHistory(rounds []chesspairing.RoundData) pairingHistory {
	h := make(pairingHistory)
	for _, round := range rounds {
		for _, game := range round.Games {
			// Skip forfeits — per project convention, forfeits are
			// excluded from pairing history (players can re-pair).
			if game.IsForfeit {
				continue
			}
			if h[game.WhiteID] == nil {
				h[game.WhiteID] = make(map[string]int)
			}
			if h[game.BlackID] == nil {
				h[game.BlackID] = make(map[string]int)
			}
			h[game.WhiteID][game.BlackID] = round.Number
			h[game.BlackID][game.WhiteID] = round.Number
		}
	}
	return h
}

// canPair checks if two players can be paired given the repeat rules.
func canPair(a, b string, opts Options, history pairingHistory, currentRound int) bool {
	if !*opts.AllowRepeatPairings {
		// Check if they've ever played.
		if _, played := history[a][b]; played {
			return false
		}
		return true
	}

	// Allow repeats but with minimum rounds between.
	lastRound, played := history[a][b]
	if !played {
		return true
	}
	return currentRound-lastRound >= *opts.MinRoundsBetweenRepeats
}

// colorForPlayer represents which color a player had.
type colorForPlayer int

const (
	colorUnknown colorForPlayer = iota
	colorWhite
	colorBlack
)

// buildLastColor returns the last color each player had.
func buildLastColor(rounds []chesspairing.RoundData) map[string]colorForPlayer {
	lastColor := make(map[string]colorForPlayer)
	for _, round := range rounds {
		for _, game := range round.Games {
			// Skip forfeits — forfeit games don't contribute to color history.
			if game.IsForfeit {
				continue
			}
			lastColor[game.WhiteID] = colorWhite
			lastColor[game.BlackID] = colorBlack
		}
	}
	return lastColor
}

// pairRanked creates pairings from a ranked list of players.
// It pairs from outside in: rank 1 vs rank N, rank 2 vs rank N-1, etc.
// If odd number of players, the middle player gets a bye.
func pairRanked(ranked []string, opts Options, history pairingHistory, lastColor map[string]colorForPlayer, currentRound int) *chesspairing.PairingResult {
	n := len(ranked)
	result := &chesspairing.PairingResult{}

	paired := make(map[string]bool, n)

	// If odd, identify the bye candidate (lowest-ranked unpaired player
	// who hasn't had a bye recently, or simply the middle player).
	var byePlayer string
	if n%2 == 1 {
		// Give bye to the lowest-ranked player who hasn't had a bye
		// in the most recent round. For simplicity, use the middle player.
		byePlayer = ranked[n/2]
		paired[byePlayer] = true
		result.Byes = []chesspairing.ByeEntry{{PlayerID: byePlayer, Type: chesspairing.ByePAB}}
		result.Notes = append(result.Notes, byePlayer+" receives a bye (odd number of players)")
	}

	// Pair from outside in.
	board := 1
	lo, hi := 0, n-1
	for lo < hi {
		// Skip already paired (bye) players.
		for lo < hi && paired[ranked[lo]] {
			lo++
		}
		for lo < hi && paired[ranked[hi]] {
			hi--
		}
		if lo >= hi {
			break
		}

		topPlayer := ranked[lo]
		bottomPlayer := ranked[hi]

		// Check repeat avoidance.
		if !canPair(topPlayer, bottomPlayer, opts, history, currentRound) {
			// Try swapping the bottom player with the next one up.
			swapped := false
			for alt := hi - 1; alt > lo; alt-- {
				if paired[ranked[alt]] {
					continue
				}
				if canPair(topPlayer, ranked[alt], opts, history, currentRound) {
					// Swap bottom and alt in our iteration.
					ranked[hi], ranked[alt] = ranked[alt], ranked[hi]
					bottomPlayer = ranked[hi]
					swapped = true
					result.Notes = append(result.Notes,
						"Swapped "+ranked[alt]+" and "+bottomPlayer+" to avoid repeat pairing with "+topPlayer)
					break
				}
			}
			if !swapped {
				// Can't avoid repeat — pair them anyway.
				result.Notes = append(result.Notes,
					"Could not avoid repeat pairing: "+topPlayer+" vs "+bottomPlayer)
			}
		}

		// Assign colors.
		whiteID, blackID := assignColors(topPlayer, bottomPlayer, lastColor)

		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board,
			WhiteID: whiteID,
			BlackID: blackID,
		})

		paired[topPlayer] = true
		paired[bottomPlayer] = true
		board++
		lo++
		hi--
	}

	return result
}

// assignColors assigns white/black based on color balance.
// The higher-ranked player gets white unless they had white last time.
func assignColors(topPlayer, bottomPlayer string, lastColor map[string]colorForPlayer) (whiteID, blackID string) {
	topLast := lastColor[topPlayer]

	switch topLast {
	case colorWhite:
		// Top had white last → give them black this time.
		return bottomPlayer, topPlayer
	default:
		// Top had black or unknown → give them white.
		return topPlayer, bottomPlayer
	}
}

// Ensure Pairer implements chesspairing.Pairer.
var _ chesspairing.Pairer = (*Pairer)(nil)
