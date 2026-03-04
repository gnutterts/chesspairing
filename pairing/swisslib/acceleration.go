package swisslib

import "math"

// BakuAccelerationRounds returns the number of accelerated rounds, full virtual
// point rounds, and half virtual point rounds for the Baku acceleration system
// (FIDE C.04.7).
//
//   - accelerated = ceil(totalRounds / 2)
//   - fullVP = ceil(accelerated / 2)
//   - halfVP = accelerated - fullVP
func BakuAccelerationRounds(totalRounds int) (accelerated, fullVP, halfVP int) {
	accelerated = int(math.Ceil(float64(totalRounds) / 2.0))
	fullVP = int(math.Ceil(float64(accelerated) / 2.0))
	halfVP = accelerated - fullVP
	return
}

// BakuGASize returns the size of Group A (top-ranked players) for Baku
// acceleration: 2 * ceil(totalPlayers / 4).
func BakuGASize(totalPlayers int) int {
	return 2 * int(math.Ceil(float64(totalPlayers)/4.0))
}

// BakuVirtualPoints returns the virtual points for a player in a given round
// under Baku acceleration.
//
//   - GA player in a full VP round: 1.0
//   - GA player in a half VP round: 0.5
//   - All other cases: 0.0
func BakuVirtualPoints(totalRounds, currentRound int, isGA bool) float64 {
	if !isGA {
		return 0.0
	}

	_, fullVP, _ := BakuAccelerationRounds(totalRounds)

	if currentRound <= fullVP {
		return 1.0
	}

	accelerated := int(math.Ceil(float64(totalRounds) / 2.0))
	if currentRound <= accelerated {
		return 0.5
	}

	return 0.0
}

// ApplyBakuAcceleration modifies PairingScore for each player by adding virtual
// points based on the Baku acceleration system.
//
// Players in Group A (InitialRank <= gaSize) receive virtual points. Players
// outside Group A are not modified.
func ApplyBakuAcceleration(players []PlayerState, currentRound, totalRounds, gaSize int) {
	for i := range players {
		isGA := players[i].InitialRank <= gaSize
		vp := BakuVirtualPoints(totalRounds, currentRound, isGA)
		players[i].PairingScore = players[i].Score + vp
	}
}
