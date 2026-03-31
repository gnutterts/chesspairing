// Package keizer implements Keizer-style pairing for chess tournaments.
//
// Keizer pairing works by ranking players by their current Keizer score
// (computed by the Keizer scorer, or by rating if no rounds have been played),
// then pairing top-down: rank 1 vs rank 2, rank 3 vs rank 4, and so on.
// The lowest-ranked player gets a bye if there's an odd number of players.
//
// Repeat avoidance: by default, players must wait at least 3 rounds before
// being paired against the same opponent again. When a conflict occurs,
// the partner is swapped with the nearest available lower-ranked player.
//
// Color assignment: the higher-ranked player gets white, unless they had
// white in their most recent game (then colors are swapped for balance).
package keizer

import (
	"context"
	"sort"

	chesspairing "github.com/gnutterts/chesspairing"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
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
	ranked := rankPlayers(ctx, active, state, entries, opts.ScoringOptions)

	// Build pairing history for repeat avoidance.
	history := buildHistory(state.Rounds)

	// Build last-color map for color balance.
	lastColor := buildLastColor(state.Rounds)

	// Pair top-down.
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

// rankPlayers returns player IDs sorted by Keizer score if rounds exist,
// otherwise by rating (descending). Uses the Keizer scorer internally
// because Keizer pairing rank = Keizer scoring rank.
func rankPlayers(ctx context.Context, ids []string, state *chesspairing.TournamentState, entries map[string]chesspairing.PlayerEntry, scoringOpts *keizerscoring.Options) []string {
	ranked := make([]string, len(ids))
	copy(ranked, ids)

	if len(state.Rounds) == 0 {
		// No rounds: sort by rating descending.
		sortByRating(ranked, entries)
		return ranked
	}

	// Use the Keizer scorer to compute scores for ranking.
	var opts keizerscoring.Options
	if scoringOpts != nil {
		opts = *scoringOpts
	}
	scorer := keizerscoring.New(opts)
	scores, err := scorer.Score(ctx, state)
	if err != nil {
		// Fall back to rating if scoring fails.
		sortByRating(ranked, entries)
		return ranked
	}

	// Build score lookup.
	scoreOf := make(map[string]float64, len(scores))
	for _, ps := range scores {
		scoreOf[ps.PlayerID] = ps.Score
	}

	sort.Slice(ranked, func(i, j int) bool {
		si := scoreOf[ranked[i]]
		sj := scoreOf[ranked[j]]
		if si != sj {
			return si > sj
		}
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
	return ranked
}

// sortByRating sorts player IDs by rating descending, with display name
// as alphabetical tiebreak for deterministic ordering.
func sortByRating(ranked []string, entries map[string]chesspairing.PlayerEntry) {
	sort.Slice(ranked, func(i, j int) bool {
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
	})
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
// It pairs top-down: rank 1 vs rank 2, rank 3 vs rank 4, etc.
// If odd number of players, the lowest-ranked player gets a bye.
func pairRanked(ranked []string, opts Options, history pairingHistory, lastColor map[string]colorForPlayer, currentRound int) *chesspairing.PairingResult {
	n := len(ranked)
	result := &chesspairing.PairingResult{}

	paired := make(map[string]bool, n)

	// If odd, the lowest-ranked player gets a bye.
	if n%2 == 1 {
		byePlayer := ranked[n-1]
		paired[byePlayer] = true
		result.Byes = []chesspairing.ByeEntry{{PlayerID: byePlayer, Type: chesspairing.ByePAB}}
		result.Notes = append(result.Notes, byePlayer+" receives a bye (lowest ranked)")
	}

	// Pair top-down: rank 1 vs rank 2, rank 3 vs rank 4, etc.
	board := 1
	for i := 0; i < n-1; i += 2 {
		if paired[ranked[i]] {
			// This player already has a bye — skip, adjust iteration.
			i--
			continue
		}

		topPlayer := ranked[i]
		partner := ranked[i+1]

		// Check repeat avoidance.
		if !canPair(topPlayer, partner, opts, history, currentRound) {
			// Try swapping the partner with the next available player.
			swapped := false
			for alt := i + 2; alt < n; alt++ {
				if paired[ranked[alt]] {
					continue
				}
				if canPair(topPlayer, ranked[alt], opts, history, currentRound) {
					oldPartner := ranked[i+1]
					newPartner := ranked[alt]
					ranked[i+1], ranked[alt] = ranked[alt], ranked[i+1]
					partner = ranked[i+1]
					swapped = true
					result.Notes = append(result.Notes,
						"Swapped "+newPartner+" for "+oldPartner+" to avoid repeat pairing with "+topPlayer)
					break
				}
			}
			if !swapped {
				result.Notes = append(result.Notes,
					"Could not avoid repeat pairing: "+topPlayer+" vs "+partner)
			}
		}

		whiteID, blackID := assignColors(topPlayer, partner, lastColor)

		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board,
			WhiteID: whiteID,
			BlackID: blackID,
		})

		paired[topPlayer] = true
		paired[partner] = true
		board++
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
