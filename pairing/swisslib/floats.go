package swisslib

// LastFloat returns the most recent non-None float direction,
// or FloatNone if there is none.
func LastFloat(history []Float) Float {
	if len(history) == 0 {
		return FloatNone
	}
	return history[len(history)-1]
}

// ConsecutiveSameFloat counts how many times the player has floated
// in the given direction in consecutive recent rounds (from the end
// of history backwards). Stops at the first round that doesn't match.
func ConsecutiveSameFloat(history []Float, dir Float) int {
	count := 0
	for i := len(history) - 1; i >= 0; i-- {
		if history[i] == dir {
			count++
		} else {
			break
		}
	}
	return count
}

// FloatedToSameScoreGroup returns true if the player floated (down) to
// the same score group in the previous round. Used by Dutch C14 to prevent
// consecutive downfloats to the same score group.
//
// history: player's float history (one entry per round)
// targetScore: the score group the player would float to now
// scores: the score the player had after each round (parallel to history)
func FloatedToSameScoreGroup(history []Float, targetScore float64, scores []float64) bool {
	if len(history) == 0 || len(scores) == 0 {
		return false
	}
	lastIdx := len(history) - 1
	if history[lastIdx] != FloatDown {
		return false
	}
	// The score the player had after the last round indicates the group
	// they were in. If that matches the target, they'd be floating to
	// the same group twice.
	return lastIdx < len(scores) && scores[lastIdx] == targetScore
}
