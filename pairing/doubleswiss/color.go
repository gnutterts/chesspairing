// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package doubleswiss

import "github.com/gnutterts/chesspairing/pairing/lexswiss"

// AllocateColor decides which participant gets White in Game 1 of the match
// and which gets Black, implementing Article 4.3 of Double-Swiss.
//
// In Double-Swiss, "colour" means who gets White in Game 1 of the two-game
// match. The other participant gets White in Game 2.
//
// initialColor is the initial-colour drawn before the first round. A nil or
// "auto" value uses White.
//
// Returns (whiteID, blackID).
func AllocateColor(a, b *lexswiss.ParticipantState, initialColor *string) (string, string) {
	hrp, opponent := higherRankedPlayer(a, b)

	for _, rule := range []func(*lexswiss.ParticipantState, *lexswiss.ParticipantState, *string) (string, string, bool){
		colorRule43_1,
		colorRule43_2,
		colorRule43_3,
		colorRule43_4,
		colorRule43_5,
	} {
		if whiteID, blackID, ok := rule(hrp, opponent, initialColor); ok {
			return whiteID, blackID
		}
	}

	// Defensive default for the "no played colour and no rule matched" case
	// so a future rule that declines does not silently pick a side.
	return hrp.ID, opponent.ID
}

// higherRankedPlayer returns the HRP and its opponent under Article 4.2.
func higherRankedPlayer(a, b *lexswiss.ParticipantState) (*lexswiss.ParticipantState, *lexswiss.ParticipantState) {
	if a.Score > b.Score || a.Score == b.Score && a.PairingNumber < b.PairingNumber {
		return a, b
	}
	return b, a
}

// colorRule43_1 assigns the initial colour when both players are new.
func colorRule43_1(hrp, opponent *lexswiss.ParticipantState, initialColor *string) (string, string, bool) {
	if len(filterPlayed(hrp.ColorHistory)) != 0 || len(filterPlayed(opponent.ColorHistory)) != 0 {
		return "", "", false
	}

	colour := lexswiss.ColorWhite
	if initialColor != nil && *initialColor == "black" {
		colour = lexswiss.ColorBlack
	}
	if hrp.PairingNumber%2 == 0 {
		colour = colour.Opposite()
	}
	whiteID, blackID := assignColor(hrp, opponent, colour)
	return whiteID, blackID, true
}

// colorRule43_2 gives White to the player with fewer previous Whites.
func colorRule43_2(hrp, opponent *lexswiss.ParticipantState, _ *string) (string, string, bool) {
	whitesHRP := countColor(filterPlayed(hrp.ColorHistory), lexswiss.ColorWhite)
	whitesOpponent := countColor(filterPlayed(opponent.ColorHistory), lexswiss.ColorWhite)
	if whitesHRP < whitesOpponent {
		return hrp.ID, opponent.ID, true
	}
	if whitesOpponent < whitesHRP {
		return opponent.ID, hrp.ID, true
	}
	return "", "", false
}

// colorRule43_3 alternates the colours to the most recent time in which one
// player had White and the other Black. Rounds that were not played are left
// out of both histories first (C.04.2 Article 3.4), so the two histories are
// compared match by match from the latest one backwards.
func colorRule43_3(hrp, opponent *lexswiss.ParticipantState, _ *string) (string, string, bool) {
	hrpPlayed := filterPlayed(hrp.ColorHistory)
	opponentPlayed := filterPlayed(opponent.ColorHistory)
	for i, j := len(hrpPlayed)-1, len(opponentPlayed)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		hc, oc := hrpPlayed[i], opponentPlayed[j]
		if hc == lexswiss.ColorWhite && oc == lexswiss.ColorBlack {
			return opponent.ID, hrp.ID, true
		}
		if hc == lexswiss.ColorBlack && oc == lexswiss.ColorWhite {
			return hrp.ID, opponent.ID, true
		}
	}
	return "", "", false
}

// colorRule43_4 alternates the HRP's colour from its last played round.
func colorRule43_4(hrp, opponent *lexswiss.ParticipantState, _ *string) (string, string, bool) {
	colour := lastPlayedColor(hrp.ColorHistory)
	if colour == lexswiss.ColorNone {
		return "", "", false
	}
	whiteID, blackID := assignColor(hrp, opponent, colour.Opposite())
	return whiteID, blackID, true
}

// colorRule43_5 alternates the opponent's colour from its last played round.
func colorRule43_5(hrp, opponent *lexswiss.ParticipantState, _ *string) (string, string, bool) {
	colour := lastPlayedColor(opponent.ColorHistory)
	if colour == lexswiss.ColorNone {
		return "", "", false
	}
	whiteID, blackID := assignColor(opponent, hrp, colour.Opposite())
	return whiteID, blackID, true
}

func assignColor(player, opponent *lexswiss.ParticipantState, colour lexswiss.Color) (string, string) {
	if colour == lexswiss.ColorWhite {
		return player.ID, opponent.ID
	}
	return opponent.ID, player.ID
}

// filterPlayed returns only non-None colours from history.
func filterPlayed(history []lexswiss.Color) []lexswiss.Color {
	var played []lexswiss.Color
	for _, colour := range history {
		if colour != lexswiss.ColorNone {
			played = append(played, colour)
		}
	}
	return played
}

func lastPlayedColor(history []lexswiss.Color) lexswiss.Color {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] != lexswiss.ColorNone {
			return history[i]
		}
	}
	return lexswiss.ColorNone
}

func countColor(history []lexswiss.Color, target lexswiss.Color) int {
	count := 0
	for _, colour := range history {
		if colour == target {
			count++
		}
	}
	return count
}
