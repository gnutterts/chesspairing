// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

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
// Color assignment uses a priority cascade: absolute preferences (imbalance
// or consecutive same color) take priority, followed by strong preferences,
// color history differences, rank tiebreak, and board alternation.
package keizer

import (
	"context"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/algorithm/blossom"
	"github.com/gnutterts/chesspairing/pairing/swisslib"
	keizerscoring "github.com/gnutterts/chesspairing/scoring/keizer"
)

// Pairer implements the chesspairing.Pairer interface for Keizer pairing.
type Pairer struct {
	opts Options
}

// New creates a new Keizer pairer with the given options.
func New(opts Options) *Pairer {
	return &Pairer{opts: opts.WithDefaults()}
}

// NewFromMap creates a new Keizer pairer from a map[string]any config.
func NewFromMap(m map[string]any) *Pairer {
	return New(ParseOptions(m))
}

// Pair generates pairings for the next round using the Keizer method.
func (p *Pairer) Pair(ctx context.Context, state *chesspairing.TournamentState) (*chesspairing.PairingResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opts := p.opts

	// Honour pre-assigned byes for the upcoming round: those players are
	// excluded from the matching pool and echoed back in result.Byes.
	preAssigned := make(map[string]bool, len(state.PreAssignedByes))
	for _, b := range state.PreAssignedByes {
		preAssigned[b.PlayerID] = true
	}
	preAssignedByes := append([]chesspairing.ByeEntry(nil), state.PreAssignedByes...)

	// Get active players.
	allActive := state.ActivePlayerIDs(state.CurrentRound)
	active := make([]string, 0, len(allActive))
	for _, id := range allActive {
		if !preAssigned[id] {
			active = append(active, id)
		}
	}
	if len(active) < 2 {
		// Not enough players to pair.
		result := &chesspairing.PairingResult{}
		if len(active) == 1 {
			result.Byes = []chesspairing.ByeEntry{{PlayerID: active[0], Type: chesspairing.ByePAB}}
			result.Notes = []string{active[0] + " receives a bye (only player)"}
		}
		if len(preAssignedByes) > 0 {
			result.Byes = append(preAssignedByes, result.Byes...)
		}
		return result, nil
	}

	// Assign pairing numbers to players (used for tiebreaking)
	playersWithNum, err := chesspairing.AssignPairingNumbers(state.Players)
	if err != nil {
		return nil, err
	}

	// Build player entries lookup.
	entries := make(map[string]chesspairing.PlayerEntry, len(playersWithNum))
	entryOrder := make(map[string]int, len(playersWithNum))
	pairingNumbers := make(map[string]int, len(playersWithNum))
	for i, pl := range playersWithNum {
		entries[pl.ID] = pl
		entryOrder[pl.ID] = i
		pairingNumbers[pl.ID] = pl.PairingNumber
	}

	// Rank players: by Keizer score (if rounds exist) or by rating.
	ranked := rankPlayers(ctx, p, active, state, entries, entryOrder, opts.ScoringOptions)

	// Build pairing history for repeat avoidance.
	history := buildHistory(state.Rounds, *opts.ForfeitCountsAsMet)

	// Build color histories for color allocation.
	colorHistories := buildColorHistories(state.Rounds)

	// Pair top-down, falling back to a full search when greedy cannot
	// produce a legal complete matching.
	priorByes := buildByeHistory(state.Rounds)
	result, err := pairRankedWithNumbers(ranked, opts, history, colorHistories, pairingNumbers, state.CurrentRound, priorByes)
	if err != nil {
		return nil, err
	}
	if len(preAssignedByes) > 0 {
		result.Byes = append(preAssignedByes, result.Byes...)
	}
	return result, nil
}

// rankPlayers returns player IDs sorted by Keizer score if rounds exist,
// otherwise by rating (descending). Uses the Keizer scorer internally
// because Keizer pairing rank = Keizer scoring rank.
func rankPlayers(ctx context.Context, p *Pairer, ids []string, state *chesspairing.TournamentState, entries map[string]chesspairing.PlayerEntry, entryOrder map[string]int, scoringOpts *keizerscoring.Options) []string {
	ranked := make([]string, len(ids))
	copy(ranked, ids)

	initialOrder := "rating-name"
	if p.opts.InitialOrder != nil {
		initialOrder = *p.opts.InitialOrder
	}

	if len(state.Rounds) == 0 {
		// No rounds: sort by rating descending.
		sortByRating(ranked, entries, entryOrder, initialOrder)
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
		sortByRating(ranked, entries, entryOrder, initialOrder)
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
		if initialOrder == "rating-entry" {
			if entryOrder[ranked[i]] != entryOrder[ranked[j]] {
				return entryOrder[ranked[i]] < entryOrder[ranked[j]]
			}
		} else {
			if entries[ranked[i]].DisplayName != entries[ranked[j]].DisplayName {
				return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
			}
		}
		return entries[ranked[i]].PairingNumber < entries[ranked[j]].PairingNumber
	})
	return ranked
}

// sortByRating sorts player IDs by rating descending, with display name
// as alphabetical tiebreak for deterministic ordering, and PairingNumber as final.
func sortByRating(ranked []string, entries map[string]chesspairing.PlayerEntry, entryOrder map[string]int, initialOrder string) {
	sort.Slice(ranked, func(i, j int) bool {
		ri := entries[ranked[i]].Rating
		rj := entries[ranked[j]].Rating
		if ri != rj {
			return ri > rj
		}
		if initialOrder == "rating-entry" {
			if entryOrder[ranked[i]] != entryOrder[ranked[j]] {
				return entryOrder[ranked[i]] < entryOrder[ranked[j]]
			}
		} else {
			if entries[ranked[i]].DisplayName != entries[ranked[j]].DisplayName {
				return entries[ranked[i]].DisplayName < entries[ranked[j]].DisplayName
			}
		}
		return entries[ranked[i]].PairingNumber < entries[ranked[j]].PairingNumber
	})
}

// pairingHistory tracks in which rounds each pair of players encountered each
// other. A player maps to their opponents, and each opponent maps to the set
// of round numbers in which the two met.
type pairingHistory map[string]map[string]map[int]struct{}

// addEncounter records one encounter between a and b in the given round.
func addEncounter(h pairingHistory, a, b string, round int) {
	if h[a] == nil {
		h[a] = make(map[string]map[int]struct{})
	}
	if h[a][b] == nil {
		h[a][b] = make(map[int]struct{})
	}
	if h[b] == nil {
		h[b] = make(map[string]map[int]struct{})
	}
	if h[b][a] == nil {
		h[b][a] = make(map[int]struct{})
	}
	h[a][b][round] = struct{}{}
	h[b][a][round] = struct{}{}
}

// buildHistory builds the pairing history from completed rounds. Single
// forfeits are never counted as encounters. A double forfeit is counted only
// when forfeitCountsAsMet is true.
func buildHistory(rounds []chesspairing.RoundData, forfeitCountsAsMet bool) pairingHistory {
	h := make(pairingHistory)
	for _, round := range rounds {
		for _, game := range round.Games {
			if game.Result.IsDoubleForfeit() {
				if !forfeitCountsAsMet {
					continue
				}
			} else if game.IsForfeit {
				continue
			}
			addEncounter(h, game.WhiteID, game.BlackID, round.Number)
		}
	}
	return h
}

// buildByeHistory returns the set of players who received a bye in any
// completed round.
func buildByeHistory(rounds []chesspairing.RoundData) map[string]bool {
	h := make(map[string]bool)
	for _, round := range rounds {
		for _, bye := range round.Byes {
			h[bye.PlayerID] = true
		}
	}
	return h
}

// samePeriod reports whether two round numbers fall in the same period of
// length periodLength (rounds 1-N, N+1-2N, ...).
func samePeriod(a, b, periodLength int) bool {
	if periodLength <= 0 {
		return false
	}
	return (a-1)/periodLength == (b-1)/periodLength
}

// canPair checks if two players can be paired given the repeat rules.
func canPair(a, b string, opts Options, history pairingHistory, currentRound int) bool {
	rounds := history[a][b]
	if len(rounds) == 0 {
		return true
	}

	if !*opts.AllowRepeatPairings {
		return false
	}

	lastRound := 0
	for round := range rounds {
		if round > lastRound {
			lastRound = round
		}
	}
	if currentRound-lastRound < *opts.MinRoundsBetweenRepeats {
		return false
	}

	if *opts.PeriodLength > 0 && *opts.NoRepeatWithinPeriod {
		for round := range rounds {
			if samePeriod(round, currentRound, *opts.PeriodLength) {
				return false
			}
		}
	}
	return true
}

// buildColorHistories returns the full color history for each player
// across all completed rounds. Forfeits are excluded (no color assigned).
// Byes produce ColorNone (filtered out by ComputeColorPreference).
func buildColorHistories(rounds []chesspairing.RoundData) map[string][]swisslib.Color {
	histories := make(map[string][]swisslib.Color)
	for _, round := range rounds {
		for _, game := range round.Games {
			if game.IsForfeit {
				continue
			}
			histories[game.WhiteID] = append(histories[game.WhiteID], swisslib.ColorWhite)
			histories[game.BlackID] = append(histories[game.BlackID], swisslib.ColorBlack)
		}
		for _, bye := range round.Byes {
			histories[bye.PlayerID] = append(histories[bye.PlayerID], swisslib.ColorNone)
		}
	}
	return histories
}

// pairRanked creates pairings from a ranked list of players.
// It pairs top-down: rank 1 vs rank 2, rank 3 vs rank 4, etc.
// If odd number of players, the lowest-ranked player gets a bye.
func pairRanked(ranked []string, opts Options, history pairingHistory, colorHistories map[string][]swisslib.Color, currentRound int) (*chesspairing.PairingResult, error) {
	pairingNumbers := make(map[string]int, len(ranked))
	for i, id := range ranked {
		pairingNumbers[id] = i + 1
	}
	return pairRankedWithNumbers(ranked, opts, history, colorHistories, pairingNumbers, currentRound, nil)
}

func pairRankedWithNumbers(ranked []string, opts Options, history pairingHistory, colorHistories map[string][]swisslib.Color, pairingNumbers map[string]int, currentRound int, priorByes map[string]bool) (*chesspairing.PairingResult, error) {
	opts = opts.WithDefaults()
	n := len(ranked)
	result := &chesspairing.PairingResult{}
	if n == 0 {
		return result, nil
	}
	if n == 1 {
		result.Byes = []chesspairing.ByeEntry{{PlayerID: ranked[0], Type: chesspairing.ByePAB}}
		result.Notes = []string{ranked[0] + " receives a bye (only player)"}
		return result, nil
	}

	ranks := make(map[string]int, n)
	for i, id := range ranked {
		ranks[id] = i + 1
	}

	matchList := ranked
	if n%2 == 1 {
		byePlayer := selectBye(ranked, opts, priorByes)
		result.Byes = []chesspairing.ByeEntry{{PlayerID: byePlayer, Type: chesspairing.ByePAB}}
		result.Notes = append(result.Notes, byeNote(byePlayer, opts))
		matchList = make([]string, 0, n-1)
		for _, id := range ranked {
			if id != byePlayer {
				matchList = append(matchList, id)
			}
		}
	}

	pairs, notes, legal := greedyPair(matchList, opts, history, currentRound)
	if !legal {
		pairs = fullSearch(matchList, ranks, opts, history, currentRound)
		if pairs == nil {
			return nil, fmt.Errorf("keizer: no pairing satisfies the repeat restrictions")
		}
		notes = nil
	}

	for board, pair := range pairs {
		a, b := pair[0], pair[1]
		if ranks[a] > ranks[b] {
			a, b = b, a
		}
		whiteID, blackID := allocateColor(a, b, colorHistories, pairingNumbers, ranks, board+1)
		result.Pairings = append(result.Pairings, chesspairing.GamePairing{
			Board:   board + 1,
			WhiteID: whiteID,
			BlackID: blackID,
		})
	}
	result.Notes = append(result.Notes, notes...)
	return result, nil
}

// selectBye returns the player who receives the pairing-allocated bye.
func selectBye(ranked []string, opts Options, priorByes map[string]bool) string {
	if *opts.ByePolicy == byePolicyLowestWithoutBye {
		for i := len(ranked) - 1; i >= 0; i-- {
			if !priorByes[ranked[i]] {
				return ranked[i]
			}
		}
	}
	return ranked[len(ranked)-1]
}

// byeNote returns the note describing why a player received the bye.
func byeNote(player string, opts Options) string {
	if *opts.ByePolicy == byePolicyLowestWithoutBye {
		return player + " receives a bye (lowest ranked without a bye)"
	}
	return player + " receives a bye (lowest ranked)"
}

// greedyPair runs the original top-down greedy pairing on an even-sized
// match list. It returns the pairs in board order and reports whether the
// result is legal (no repeat restriction was violated).
func greedyPair(matchList []string, opts Options, history pairingHistory, currentRound int) (pairs [][2]string, notes []string, legal bool) {
	players := append([]string(nil), matchList...)
	for i := 0; i < len(players); i += 2 {
		topPlayer := players[i]
		partner := players[i+1]

		if !canPair(topPlayer, partner, opts, history, currentRound) {
			swapped := false
			for alt := i + 2; alt < len(players); alt++ {
				if canPair(topPlayer, players[alt], opts, history, currentRound) {
					oldPartner := players[i+1]
					newPartner := players[alt]
					players[i+1], players[alt] = players[alt], players[i+1]
					partner = players[i+1]
					swapped = true
					notes = append(notes,
						"Swapped "+newPartner+" for "+oldPartner+" to avoid repeat pairing with "+topPlayer)
					break
				}
			}
			if !swapped {
				return nil, notes, false
			}
		}

		pairs = append(pairs, [2]string{topPlayer, partner})
	}
	return pairs, notes, true
}

// fullSearch finds a perfect matching on matchList that respects the repeat
// restrictions and minimizes the sum of rank distances. It uses Edmonds'
// blossom algorithm with weight -(rank distance). It returns nil when no such
// matching exists.
func fullSearch(matchList []string, ranks map[string]int, opts Options, history pairingHistory, currentRound int) [][2]string {
	if len(matchList) == 0 {
		return [][2]string{}
	}

	edges := make([]blossom.BlossomEdge, 0)
	for i := 0; i < len(matchList); i++ {
		for j := i + 1; j < len(matchList); j++ {
			if !canPair(matchList[i], matchList[j], opts, history, currentRound) {
				continue
			}
			distance := ranks[matchList[i]] - ranks[matchList[j]]
			if distance < 0 {
				distance = -distance
			}
			edges = append(edges, blossom.BlossomEdge{I: i, J: j, Weight: int64(-distance)})
		}
	}

	mate := blossom.MaxWeightMatching(edges, true)
	if mate == nil || len(mate) < len(matchList) {
		return nil
	}
	for _, partner := range mate {
		if partner == -1 {
			return nil
		}
	}

	pairs := make([][2]string, 0, len(matchList)/2)
	for i := 0; i < len(matchList); i++ {
		if i < mate[i] {
			pairs = append(pairs, [2]string{matchList[i], matchList[mate[i]]})
		}
	}
	sort.Slice(pairs, func(a, b int) bool {
		minA := ranks[pairs[a][0]]
		if ranks[pairs[a][1]] < minA {
			minA = ranks[pairs[a][1]]
		}
		minB := ranks[pairs[b][0]]
		if ranks[pairs[b][1]] < minB {
			minB = ranks[pairs[b][1]]
		}
		return minA < minB
	})
	return pairs
}

// allocateColor assigns white/black using the full swisslib color preference
// cascade: absolute > strong > color-history difference > rank > board alternation.
func allocateColor(a, b string, colorHistories map[string][]swisslib.Color, pairingNumbers, currentRanks map[string]int, board int) (string, string) {
	pa := &swisslib.PlayerState{
		ID:            a,
		PairingNumber: pairingNumbers[a],
		TPN:           currentRanks[a],
		ColorHistory:  colorHistories[a],
	}
	pb := &swisslib.PlayerState{
		ID:            b,
		PairingNumber: pairingNumbers[b],
		TPN:           currentRanks[b],
		ColorHistory:  colorHistories[b],
	}
	return swisslib.AllocateColor(pa, pb, false, board, nil, swisslib.AlternateByBoard)
}
