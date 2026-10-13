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
	ID          string
	DisplayName string
	// PairingNumber is the fixed, 1-based FIDE Tournament Pairing Number.
	PairingNumber int
	// InitialRank is the starting rank and is filled with PairingNumber.
	InitialRank int
	// TPN is the live rank within the current pairing; NOT the FIDE TPN, see PairingNumber.
	TPN           int
	Score         float64  // tournament points before acceleration
	PairingScore  float64  // pairing score = Score + virtual points (0 if no acceleration)
	ColorHistory  []Color  // color per round (index 0 = round 1)
	FloatHistory  []Float  // float per round (index 0 = round 1, Dutch only)
	Opponents     []string // IDs of opponents faced (forfeits excluded)
	PABIneligible PABIneligibility
	Active        bool
	Rating        int
}

// PABIneligibility records reasons a player cannot receive a pairing-allocated bye.
type PABIneligibility struct {
	PriorPAB          bool
	FullPointUnplayed bool
}

// Any reports whether the player is ineligible for a pairing-allocated bye.
func (p PABIneligibility) Any() bool {
	return p.PriorPAB || p.FullPointUnplayed
}

// IsForfeit reports whether a game ended by forfeit (C.04.2 Articles 3.4 and 3.5).
//
// trf/convert.go encodes TRF unrated short games (W/D/L, lasted less than one move)
// as a plain result with IsForfeit: true, and TRF forfeit results (+/-) as
// ResultForfeitWhiteWins, ResultForfeitBlackWins, or ResultDoubleForfeit.
// Callers may set either game.IsForfeit or game.Result.IsForfeit(), so both
// signals are honoured via their union.
func IsForfeit(game chesspairing.GameData) bool {
	return game.IsForfeit || game.Result.IsForfeit()
}

// ComputePABIneligibility computes C2 ineligibility for all players based on completed rounds.
func ComputePABIneligibility(state *chesspairing.TournamentState) map[string]PABIneligibility {
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	ineligible := make(map[string]PABIneligibility)
	points := pointsFor(state.ScoringConfig)
	for ri := 0; ri < historyEnd; ri++ {
		round := state.Rounds[ri]
		participated := make(map[string]bool)
		for _, game := range round.Games {
			participated[game.WhiteID] = true
			participated[game.BlackID] = true
			if IsForfeit(game) {
				white, black := points.unplayedGame(game.Result)
				if game.Result == chesspairing.ResultDoubleForfeit {
					white, black = points.forfeitLoss, points.forfeitLoss
				}
				if white >= points.win {
					entry := ineligible[game.WhiteID]
					entry.FullPointUnplayed = true
					ineligible[game.WhiteID] = entry
				}
				if black >= points.win {
					entry := ineligible[game.BlackID]
					entry.FullPointUnplayed = true
					ineligible[game.BlackID] = entry
				}
			}
		}
		for _, bye := range round.Byes {
			participated[bye.PlayerID] = true
			entry := ineligible[bye.PlayerID]
			if bye.Type == chesspairing.ByePAB {
				entry.PriorPAB = true
			}
			if points.bye(bye.Type) >= points.win {
				entry.FullPointUnplayed = true
			}
			ineligible[bye.PlayerID] = entry
		}
		if points.absent >= points.win {
			for _, player := range state.Players {
				if !participated[player.ID] && state.IsActiveInRound(player.ID, ri+1) {
					entry := ineligible[player.ID]
					entry.FullPointUnplayed = true
					ineligible[player.ID] = entry
				}
			}
		}
	}
	return ineligible
}

// ByeReceived reports whether the player is ineligible for a pairing-allocated bye.
func (p PlayerState) ByeReceived() bool {
	return p.PABIneligible.Any()
}

// EffectivePairingNumber returns the fixed tournament pairing number. The
// fallbacks keep manually constructed PlayerState values compatible; states
// produced by BuildPlayerStates always have PairingNumber set.
func EffectivePairingNumber(p *PlayerState) int {
	if p.PairingNumber != 0 {
		return p.PairingNumber
	}
	if p.InitialRank != 0 {
		return p.InitialRank
	}
	return p.TPN
}

// BuildPlayerStates converts a TournamentState into a sorted slice of
// PlayerState values ready for the pairing algorithm.
//
// Active players only. Sorted by score (desc), then initial rank (asc).
// TPN assigned sequentially after sorting.
//
// Pairing scores use configured tournament points (pending games count as draws).
// Forfeit games are excluded from opponent history (players can be paired again).
func BuildPlayerStates(state *chesspairing.TournamentState) ([]PlayerState, error) {
	return BuildPlayerStatesWithVirtualPoints(state, nil)
}

// BuildPlayerStatesWithVirtualPoints is BuildPlayerStates for an accelerated
// tournament. virtualPoints returns the virtual points a player was given when
// the given (1-based) round was paired; they count in the score a float is
// measured with (C.04.7 1.5: the pairing score defines the scoregroups, so who
// floated depends on it). A nil function means no acceleration.
func BuildPlayerStatesWithVirtualPoints(state *chesspairing.TournamentState, virtualPoints func(playerID string, round int) float64) ([]PlayerState, error) {
	points := pointsFor(state.ScoringConfig)
	// Step 1: Assign pairing numbers.
	playersWithNum, err := chesspairing.AssignPairingNumbers(state.Players)
	if err != nil {
		return nil, err
	}

	initialRanks := make(map[string]int, len(playersWithNum))
	for _, p := range playersWithNum {
		initialRanks[p.ID] = p.PairingNumber
	}

	// Step 2: Filter to active players.
	activeSet := make(map[string]bool)
	var activePlayers []chesspairing.PlayerEntry
	for _, p := range state.Players {
		if state.IsActiveInRound(p.ID, state.CurrentRound) {
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
	//   bye with points > played loss → FLOAT_DOWN
	//
	// We track scores for ALL players (including withdrawn) since they may
	// have been opponents in earlier rounds.
	scores := make(map[string]float64)
	colorHistories := make(map[string][]Color)
	floatHistories := make(map[string][]Float)
	opponents := make(map[string][]string)
	pabIneligibility := make(map[string]PABIneligibility)

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
		participated := make(map[string]bool)

		for _, game := range round.Games {
			participated[game.WhiteID] = true
			participated[game.BlackID] = true
			// C.04.2 3.1: pending games count as draws for pairing only.
			switch {
			case game.Result == chesspairing.ResultDoubleForfeit:
				// Neither player scores.
			case IsForfeit(game):
				switch game.Result {
				case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
					scores[game.WhiteID] += points.forfeitWin
					scores[game.BlackID] += points.forfeitLoss
				case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
					scores[game.BlackID] += points.forfeitWin
					scores[game.WhiteID] += points.forfeitLoss
				case chesspairing.ResultDraw:
					scores[game.WhiteID] += points.draw
					scores[game.BlackID] += points.draw
				}
			default:
				switch game.Result {
				case chesspairing.ResultWhiteWins:
					scores[game.WhiteID] += points.win
					scores[game.BlackID] += points.loss
				case chesspairing.ResultBlackWins:
					scores[game.BlackID] += points.win
					scores[game.WhiteID] += points.loss
				case chesspairing.ResultDraw, chesspairing.ResultPending:
					scores[game.WhiteID] += points.draw
					scores[game.BlackID] += points.draw
				}
			}

			if IsForfeit(game) {
				white, black := points.unplayedGame(game.Result)
				if game.Result == chesspairing.ResultDoubleForfeit {
					white, black = points.forfeitLoss, points.forfeitLoss
				}
				if white >= points.win {
					entry := pabIneligibility[game.WhiteID]
					entry.FullPointUnplayed = true
					pabIneligibility[game.WhiteID] = entry
				}
				if black >= points.win {
					entry := pabIneligibility[game.BlackID]
					entry.FullPointUnplayed = true
					pabIneligibility[game.BlackID] = entry
				}
			}

			// Color history: exclude forfeits (game not actually played → no color assigned).
			// This matches FIDE C.04.3 and bbpPairings: only played games count for
			// color preference, color difference, and consecutive-same-color tracking.
			if !IsForfeit(game) {
				if activeSet[game.WhiteID] {
					colorHistories[game.WhiteID] = append(colorHistories[game.WhiteID], ColorWhite)
				}
				if activeSet[game.BlackID] {
					colorHistories[game.BlackID] = append(colorHistories[game.BlackID], ColorBlack)
				}
			}

			// Opponent history: exclude forfeits (forfeit = can be paired again).
			if !IsForfeit(game) {
				opponents[game.WhiteID] = append(opponents[game.WhiteID], game.BlackID)
				opponents[game.BlackID] = append(opponents[game.BlackID], game.WhiteID)
			}

			// Track participation for float computation.
			// Unplayed games float down only when awarded more than a played loss.
			if IsForfeit(game) {
				whitePoints, blackPoints := points.forfeitLoss, points.forfeitLoss
				switch game.Result {
				case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
					whitePoints = points.forfeitWin
				case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
					blackPoints = points.forfeitWin
				case chesspairing.ResultDoubleForfeit:
					whitePoints, blackPoints = 0, 0
				case chesspairing.ResultDraw:
					whitePoints, blackPoints = points.draw, points.draw
				}
				participants = append(participants, roundParticipant{playerID: game.WhiteID, isBye: true, byePoints: whitePoints})
				participants = append(participants, roundParticipant{playerID: game.BlackID, isBye: true, byePoints: blackPoints})
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
			participated[bye.PlayerID] = true
			// Any unplayed round worth a win makes a player C2-ineligible.
			ineligible := pabIneligibility[bye.PlayerID]
			if bye.Type == chesspairing.ByePAB {
				ineligible.PriorPAB = true
			}
			if points.bye(bye.Type) >= points.win {
				ineligible.FullPointUnplayed = true
			}
			pabIneligibility[bye.PlayerID] = ineligible
			scores[bye.PlayerID] += points.bye(bye.Type)
			if activeSet[bye.PlayerID] {
				colorHistories[bye.PlayerID] = append(colorHistories[bye.PlayerID], ColorNone)
			}
			participants = append(participants, roundParticipant{
				playerID: bye.PlayerID, isBye: true, byePoints: points.bye(bye.Type),
			})
		}

		// C.04.2 3.3: use the configured absent value when no result is recorded.
		for _, player := range state.Players {
			if !participated[player.ID] && state.IsActiveInRound(player.ID, ri+1) {
				scores[player.ID] += points.absent
				if activeSet[player.ID] {
					f := FloatNone
					if points.absent > points.loss {
						f = FloatDown
					}
					floatHistories[player.ID] = append(floatHistories[player.ID], f)
				}
				if points.absent >= points.win {
					entry := pabIneligibility[player.ID]
					entry.FullPointUnplayed = true
					pabIneligibility[player.ID] = entry
				}
			}
		}

		// Compute float for each participant this round.
		for _, p := range participants {
			if !activeSet[p.playerID] {
				continue // only track floats for active players
			}
			var f Float
			if p.isBye {
				// bbpPairings: bye with points > loss → FLOAT_DOWN, else FLOAT_NONE.
				// Compare against the tournament's played-loss value.
				if p.byePoints > points.loss {
					f = FloatDown
				} else {
					f = FloatNone
				}
			} else {
				// Compare scores at start of this round.
				playerScore := scoresBeforeRound[p.playerID]
				opponentScore := scoresBeforeRound[p.opponentID]
				if virtualPoints != nil {
					playerScore += virtualPoints(p.playerID, ri+1)
					opponentScore += virtualPoints(p.opponentID, ri+1)
				}
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
			ID:            p.ID,
			DisplayName:   p.DisplayName,
			PairingNumber: initialRanks[p.ID],
			InitialRank:   initialRanks[p.ID],
			TPN:           0, // assigned after sorting
			Score:         ScoreGroupKey(scores[p.ID]),
			PairingScore:  ScoreGroupKey(scores[p.ID]),
			ColorHistory:  colorHistories[p.ID],
			FloatHistory:  floatHistories[p.ID], // computed from historical game data above
			Opponents:     opponents[p.ID],
			PABIneligible: pabIneligibility[p.ID],
			Active:        true,
			Rating:        p.Rating,
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

	return players, nil
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

// byePoints returns the default pairing points for a bye type.
func byePoints(bt chesspairing.ByeType) float64 {
	return pointsFor(chesspairing.ScoringConfig{}).bye(bt)
}
