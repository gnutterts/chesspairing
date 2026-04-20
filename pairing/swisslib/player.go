// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package swisslib

import (
	"sort"

	"github.com/gnutterts/chesspairing"
)

// Color represents a player's color assignment in a round.
type Color int

const (
	ColorNone Color = iota // bye, absent, or no game
	ColorWhite
	ColorBlack
)

// String returns the color name for debugging.
func (c Color) String() string {
	switch c {
	case ColorWhite:
		return "White"
	case ColorBlack:
		return "Black"
	default:
		return "None"
	}
}

// Opposite returns the opposite color (White↔Black). None returns None.
func (c Color) Opposite() Color {
	switch c {
	case ColorWhite:
		return ColorBlack
	case ColorBlack:
		return ColorWhite
	default:
		return ColorNone
	}
}

// Float represents a player's float direction in a round.
type Float int

const (
	FloatNone Float = iota
	FloatUp
	FloatDown
)

// PlayerState holds the computed state of a single player for the pairing
// algorithm. Built once per Pair() call from the engine's TournamentState.
type PlayerState struct {
	ID           string
	DisplayName  string
	InitialRank  int      // starting rank (by rating desc, then name asc), 1-based
	TPN          int      // Tournament Pairing Number (re-ranked each round), 1-based
	Score        float64  // actual score (standard 1-½-0, not tournament scoring)
	PairingScore float64  // pairing score = Score + virtual points (0 if no acceleration)
	ColorHistory []Color  // color per round (index 0 = round 1)
	FloatHistory []Float  // float per round (index 0 = round 1, Dutch only)
	Opponents    []string // IDs of opponents faced (forfeits excluded)
	ByeReceived  bool     // already received a PAB
	Active       bool
	Rating       int
}

// BuildPlayerStates converts a TournamentState into a sorted slice of
// PlayerState values ready for the pairing algorithm.
//
// Active players only. Sorted by score (desc), then initial rank (asc).
// TPN assigned sequentially after sorting.
//
// Pairing scores use standard 1-½-0 regardless of tournament scoring system.
// Forfeit games are excluded from opponent history (players can be paired again).
func BuildPlayerStates(state *chesspairing.TournamentState) []PlayerState {
	// Step 1: Sort all players by rating desc, name asc to assign initial ranks.
	allPlayers := make([]chesspairing.PlayerEntry, len(state.Players))
	copy(allPlayers, state.Players)
	sort.SliceStable(allPlayers, func(i, j int) bool {
		if allPlayers[i].Rating != allPlayers[j].Rating {
			return allPlayers[i].Rating > allPlayers[j].Rating
		}
		return allPlayers[i].DisplayName < allPlayers[j].DisplayName
	})

	initialRanks := make(map[string]int, len(allPlayers))
	for i, p := range allPlayers {
		initialRanks[p.ID] = i + 1
	}

	// Step 2: Filter to active players.
	activeSet := make(map[string]bool)
	var activePlayers []chesspairing.PlayerEntry
	for _, p := range state.Players {
		if p.Active {
			activePlayers = append(activePlayers, p)
			activeSet[p.ID] = true
		}
	}

	// Step 3: Compute pairing scores, color history, opponents, bye status,
	// and float history.
	//
	// Float history mirrors bbpPairings' getFloat() (dutch.cpp lines 110-133):
	// for each round, compare the player's score at the START of that round
	// with their opponent's score at the START of that round.
	//   playerScore > opponentScore → FLOAT_DOWN
	//   playerScore < opponentScore → FLOAT_UP
	//   equal → FLOAT_NONE
	//   bye with points > loss (i.e. PAB = 1 point) → FLOAT_DOWN
	//
	// We track scores for ALL players (including withdrawn) since they may
	// have been opponents in earlier rounds.
	scores := make(map[string]float64)
	colorHistories := make(map[string][]Color)
	floatHistories := make(map[string][]Float)
	opponents := make(map[string][]string)
	byeReceived := make(map[string]bool)

	// Walk only completed rounds. By contract state.Rounds holds rounds
	// 1..CurrentRound-1; in case a caller staged the upcoming round in
	// state.Rounds[CurrentRound-1] (e.g. to record pre-assigned byes
	// before pairing) we must not consume it here. Pre-assigned byes
	// for the upcoming round live in state.PreAssignedByes and are
	// applied by the pairer after BuildPlayerStates returns.
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}

	for ri := 0; ri < historyEnd; ri++ {
		round := state.Rounds[ri]
		// Capture scores at the START of this round (before processing results).
		scoresBeforeRound := make(map[string]float64, len(scores))
		for id, s := range scores {
			scoresBeforeRound[id] = s
		}

		// Track which players participated this round and their float status.
		type roundParticipant struct {
			playerID   string
			opponentID string // empty for byes
			isBye      bool
			byePoints  float64 // points awarded for the bye
		}
		var participants []roundParticipant

		for _, game := range round.Games {
			// Score: standard 1-½-0 for all players.
			switch game.Result {
			case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
				scores[game.WhiteID] += 1.0
			case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
				scores[game.BlackID] += 1.0
			case chesspairing.ResultDraw:
				scores[game.WhiteID] += 0.5
				scores[game.BlackID] += 0.5
				// DoubleForfeit and NoResult: 0 points for both
			}

			// Color history: exclude forfeits (game not actually played → no color assigned).
			// This matches FIDE C.04.3 and bbpPairings: only played games count for
			// color preference, color difference, and consecutive-same-color tracking.
			if !game.IsForfeit {
				if activeSet[game.WhiteID] {
					colorHistories[game.WhiteID] = append(colorHistories[game.WhiteID], ColorWhite)
				}
				if activeSet[game.BlackID] {
					colorHistories[game.BlackID] = append(colorHistories[game.BlackID], ColorBlack)
				}
			}

			// Opponent history: exclude forfeits (forfeit = can be paired again).
			if !game.IsForfeit {
				opponents[game.WhiteID] = append(opponents[game.WhiteID], game.BlackID)
				opponents[game.BlackID] = append(opponents[game.BlackID], game.WhiteID)
			}

			// Track participation for float computation.
			// bbpPairings computes float for ALL matches, including forfeits
			// where gameWasPlayed is false. For forfeits, it uses the bye logic
			// (points > loss → FLOAT_DOWN). For played games, it compares scores.
			if game.IsForfeit {
				// Forfeit: treat like byes for float purposes.
				// Winner gets 1 point (> 0 = loss), loser gets 0.
				switch game.Result {
				case chesspairing.ResultForfeitWhiteWins:
					participants = append(participants, roundParticipant{
						playerID: game.WhiteID, isBye: true, byePoints: 1.0,
					})
					participants = append(participants, roundParticipant{
						playerID: game.BlackID, isBye: true, byePoints: 0.0,
					})
				case chesspairing.ResultForfeitBlackWins:
					participants = append(participants, roundParticipant{
						playerID: game.WhiteID, isBye: true, byePoints: 0.0,
					})
					participants = append(participants, roundParticipant{
						playerID: game.BlackID, isBye: true, byePoints: 1.0,
					})
				default: // double forfeit
					participants = append(participants, roundParticipant{
						playerID: game.WhiteID, isBye: true, byePoints: 0.0,
					})
					participants = append(participants, roundParticipant{
						playerID: game.BlackID, isBye: true, byePoints: 0.0,
					})
				}
			} else {
				participants = append(participants, roundParticipant{
					playerID: game.WhiteID, opponentID: game.BlackID,
				})
				participants = append(participants, roundParticipant{
					playerID: game.BlackID, opponentID: game.WhiteID,
				})
			}
		}

		// Byes
		for _, bye := range round.Byes {
			// Only PAB consumes the one-time pairing-allocated-bye
			// allowance. Other bye types (half, zero, absent, excused,
			// club commitment) leave the player eligible for a future
			// PAB.
			if bye.Type == chesspairing.ByePAB {
				byeReceived[bye.PlayerID] = true
			}
			switch bye.Type {
			case chesspairing.ByePAB:
				scores[bye.PlayerID] += 1.0
			case chesspairing.ByeHalf:
				scores[bye.PlayerID] += 0.5
			case chesspairing.ByeZero, chesspairing.ByeAbsent,
				chesspairing.ByeExcused, chesspairing.ByeClubCommitment:
				// 0 points for pairing-score purposes.
			}
			if activeSet[bye.PlayerID] {
				colorHistories[bye.PlayerID] = append(colorHistories[bye.PlayerID], ColorNone)
			}
			participants = append(participants, roundParticipant{
				playerID: bye.PlayerID, isBye: true, byePoints: byePoints(bye.Type),
			})
		}

		// Compute float for each participant this round.
		for _, p := range participants {
			if !activeSet[p.playerID] {
				continue // only track floats for active players
			}
			var f Float
			if p.isBye {
				// bbpPairings: bye with points > loss → FLOAT_DOWN, else FLOAT_NONE.
				// Loss points = 0 in standard scoring.
				if p.byePoints > 0 {
					f = FloatDown
				} else {
					f = FloatNone
				}
			} else {
				// Compare scores at start of this round.
				playerScore := scoresBeforeRound[p.playerID]
				opponentScore := scoresBeforeRound[p.opponentID]
				switch {
				case playerScore > opponentScore+0.001:
					f = FloatDown
				case playerScore+0.001 < opponentScore:
					f = FloatUp
				default:
					f = FloatNone
				}
			}
			floatHistories[p.playerID] = append(floatHistories[p.playerID], f)
		}
	}

	// Step 4: Build PlayerState slice.
	players := make([]PlayerState, 0, len(activePlayers))
	for _, p := range activePlayers {
		ps := PlayerState{
			ID:           p.ID,
			DisplayName:  p.DisplayName,
			InitialRank:  initialRanks[p.ID],
			TPN:          0, // assigned after sorting
			Score:        scores[p.ID],
			PairingScore: scores[p.ID],
			ColorHistory: colorHistories[p.ID],
			FloatHistory: floatHistories[p.ID], // computed from historical game data above
			Opponents:    opponents[p.ID],
			ByeReceived:  byeReceived[p.ID],
			Active:       true,
			Rating:       p.Rating,
		}
		players = append(players, ps)
	}

	// Step 5: Sort by score desc, then initial rank asc. Assign TPN.
	sort.SliceStable(players, func(i, j int) bool {
		if players[i].Score != players[j].Score {
			return players[i].Score > players[j].Score
		}
		return players[i].InitialRank < players[j].InitialRank
	})

	for i := range players {
		players[i].TPN = i + 1
	}

	return players
}

// HasPlayed returns true if player a has played against player b
// (based on opponent history, which excludes forfeits).
func HasPlayed(a, b *PlayerState) bool {
	for _, opp := range a.Opponents {
		if opp == b.ID {
			return true
		}
	}
	return false
}

// byePoints returns the pairing score points for a given bye type.
// These values must agree with the score loop in BuildPlayerStates so
// that float computation (which compares byePoints against the loss
// threshold) sees the same value the player's score reflects.
func byePoints(bt chesspairing.ByeType) float64 {
	switch bt {
	case chesspairing.ByePAB:
		return 1.0
	case chesspairing.ByeHalf:
		return 0.5
	case chesspairing.ByeZero, chesspairing.ByeAbsent,
		chesspairing.ByeExcused, chesspairing.ByeClubCommitment:
		return 0.0
	default:
		// Unknown bye type: treat as zero so float computation does not
		// spuriously flag FloatDown. Adding a new ByeType requires an
		// explicit case here; the byetypes_test sentinel will catch
		// omissions.
		return 0.0
	}
}
