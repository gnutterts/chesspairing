// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import (
	"github.com/gnutterts/chesspairing"
)

// Options holds configurable settings for Keizer point scoring.
// All fields are pointers to distinguish "not set" (nil = use default)
// from "explicitly set to zero."
//
// Defaults follow KeizerForClubs conventions (most widely used software).
//
// # Variant presets
//
// KeizerForClubs (default): all nil — use defaults.
//
// Classic KNSB sixths: WinFraction=1, DrawFraction=0.5, LossFraction=0,
// ByeValueFraction=4/6, AbsentPenaltyFraction=2/6, ClubCommitmentFraction=2/3,
// ExcusedAbsentFraction=2/6, AbsenceLimit=5.
//
// FreeKeizer: LossFraction=1/6, ByeValueFraction=4/6,
// AbsentPenaltyFraction=2/6, AbsenceLimit=5.
//
// No self-victory: SelfVictory=false.
//
// Fixed absences: AbsentFixedValue=15, ExcusedAbsentFixedValue=15,
// ClubCommitmentFixedValue=25.
//
// Decaying absences: AbsenceDecay=true, AbsenceLimit=0.
type Options struct {
	// --- Value number assignment ---

	// ValueNumberBase is the top-ranked player's value number.
	// Default: player count (N).
	ValueNumberBase *int `json:"valueNumberBase,omitempty"`

	// ValueNumberStep is the decrement per rank position.
	// Player at rank r gets: ValueNumberBase - (r-1) * ValueNumberStep.
	// Default: 1.
	ValueNumberStep *int `json:"valueNumberStep,omitempty"`

	// --- Game result fractions (fraction of OPPONENT's Keizer value) ---

	// WinFraction is the multiplier applied to the opponent's value number
	// for a win. Points for win = opponent_value × WinFraction.
	// Default: 1.0.
	WinFraction *float64 `json:"winFraction,omitempty"`

	// DrawFraction is the multiplier applied to the opponent's value number
	// for a draw. Points for draw = opponent_value × DrawFraction.
	// Default: 0.5.
	DrawFraction *float64 `json:"drawFraction,omitempty"`

	// LossFraction is the multiplier applied to the opponent's value number
	// for a loss. Points for loss = opponent_value × LossFraction.
	// Default: 0.0 (KeizerForClubs). Set to 1/6 for FreeKeizer toughness bonus.
	LossFraction *float64 `json:"lossFraction,omitempty"`

	// ForfeitWinFraction is the multiplier applied to the opponent's value number
	// when winning by forfeit. Default: 1.0.
	ForfeitWinFraction *float64 `json:"forfeitWinFraction,omitempty"`

	// ForfeitLossFraction is the multiplier applied to the opponent's value number
	// when losing by forfeit. Default: 0.0.
	ForfeitLossFraction *float64 `json:"forfeitLossFraction,omitempty"`

	// DoubleForfeitFraction is the multiplier applied to the opponent's value number
	// when both players forfeit. Applied to each player. Default: 0.0.
	DoubleForfeitFraction *float64 `json:"doubleForfeitFraction,omitempty"`

	// --- Non-game result fractions (fraction of OWN Keizer value) ---

	// ByeValueFraction is the fraction of own value number awarded for a
	// pairing-allocated bye (PAB). Default: 0.50 (KeizerForClubs).
	ByeValueFraction *float64 `json:"byeValueFraction,omitempty"`

	// HalfByeFraction is the fraction of own value number awarded for a
	// half-point bye. Default: 0.50.
	HalfByeFraction *float64 `json:"halfByeFraction,omitempty"`

	// ZeroByeFraction is the fraction of own value number awarded for a
	// zero-point bye. Default: 0.0.
	ZeroByeFraction *float64 `json:"zeroByeFraction,omitempty"`

	// AbsentPenaltyFraction is the fraction of own value number awarded
	// when a player is absent (unexcused). Default: 0.35 (KeizerForClubs).
	AbsentPenaltyFraction *float64 `json:"absentPenaltyFraction,omitempty"`

	// ExcusedAbsentFraction is the fraction of own value number awarded
	// when a player has an excused absence. Default: 0.35.
	ExcusedAbsentFraction *float64 `json:"excusedAbsentFraction,omitempty"`

	// ClubCommitmentFraction is the fraction of own value number awarded
	// when a player is absent for interclub team duty.
	// Default: 0.70 (KeizerForClubs). Club commitments are exempt from
	// absence limits and decay.
	ClubCommitmentFraction *float64 `json:"clubCommitmentFraction,omitempty"`

	// --- Fixed-value overrides (nil = use fraction, non-nil = fixed score) ---
	// When set, the fixed value is used instead of the fraction calculation.
	// Values are in real (non-doubled) units.

	// ByeFixedValue overrides ByeValueFraction with a fixed PAB bye score.
	ByeFixedValue *int `json:"byeFixedValue,omitempty"`

	// HalfByeFixedValue overrides HalfByeFraction with a fixed half-bye score.
	HalfByeFixedValue *int `json:"halfByeFixedValue,omitempty"`

	// ZeroByeFixedValue overrides ZeroByeFraction with a fixed zero-bye score.
	ZeroByeFixedValue *int `json:"zeroByeFixedValue,omitempty"`

	// AbsentFixedValue overrides AbsentPenaltyFraction with a fixed absence score.
	AbsentFixedValue *int `json:"absentFixedValue,omitempty"`

	// ExcusedAbsentFixedValue overrides ExcusedAbsentFraction with a fixed score.
	ExcusedAbsentFixedValue *int `json:"excusedAbsentFixedValue,omitempty"`

	// ClubCommitmentFixedValue overrides ClubCommitmentFraction with a fixed score.
	ClubCommitmentFixedValue *int `json:"clubCommitmentFixedValue,omitempty"`

	// --- Behavioral options ---

	// SelfVictory controls whether each player's own Keizer value is added
	// to their total (once, not per round). This is standard in every known
	// Keizer implementation. Default: true.
	SelfVictory *bool `json:"selfVictory,omitempty"`

	// AbsenceLimit is the maximum number of absences that score points.
	// Absences beyond this limit score 0. Club commitments are exempt.
	// 0 means unlimited. Default: 5.
	AbsenceLimit *int `json:"absenceLimit,omitempty"`

	// AbsenceDecay halves the absence bonus for each successive absence:
	// 1st absence = full fraction, 2nd = fraction/2, 3rd = fraction/4, etc.
	// Club commitments are exempt. Default: false.
	AbsenceDecay *bool `json:"absenceDecay,omitempty"`

	// Frozen disables the iterative convergence loop. Instead of rescoring
	// all rounds with the final ranking's value numbers, each round is scored
	// once using the ranking as it stood before that round. Points from
	// earlier rounds are never retroactively recalculated.
	// Default: false (standard iterative Keizer).
	Frozen *bool `json:"frozen,omitempty"`

	// --- Other ---

	// LateJoinHandicap is the fixed score awarded per round missed before
	// a player joined the tournament. Unlike absences (which use a fraction
	// of the player's own value or AbsentFixedValue), late-join rounds use
	// this value directly. Late-join rounds do not count toward AbsenceLimit
	// or AbsenceDecay.
	// Requires PlayerEntry.JoinedRound to be set (0 or 1 = original player).
	// Default: 0 (late-join rounds score nothing).
	LateJoinHandicap *float64 `json:"lateJoinHandicap,omitempty"`
}

// WithDefaults returns a copy of Options with all nil fields filled
// in with system defaults. playerCount is the number of active players
// in the tournament.
func (o Options) WithDefaults(playerCount int) Options {
	// Value number assignment.
	if o.ValueNumberBase == nil {
		o.ValueNumberBase = chesspairing.IntPtr(playerCount)
	}
	if o.ValueNumberStep == nil {
		o.ValueNumberStep = chesspairing.IntPtr(1)
	}

	// Game result fractions (opponent's value).
	if o.WinFraction == nil {
		o.WinFraction = chesspairing.Float64Ptr(1.0)
	}
	if o.DrawFraction == nil {
		o.DrawFraction = chesspairing.Float64Ptr(0.5)
	}
	if o.LossFraction == nil {
		o.LossFraction = chesspairing.Float64Ptr(0.0)
	}
	if o.ForfeitWinFraction == nil {
		o.ForfeitWinFraction = chesspairing.Float64Ptr(1.0)
	}
	if o.ForfeitLossFraction == nil {
		o.ForfeitLossFraction = chesspairing.Float64Ptr(0.0)
	}
	if o.DoubleForfeitFraction == nil {
		o.DoubleForfeitFraction = chesspairing.Float64Ptr(0.0)
	}

	// Non-game result fractions (own value).
	if o.ByeValueFraction == nil {
		o.ByeValueFraction = chesspairing.Float64Ptr(0.50)
	}
	if o.HalfByeFraction == nil {
		o.HalfByeFraction = chesspairing.Float64Ptr(0.50)
	}
	if o.ZeroByeFraction == nil {
		o.ZeroByeFraction = chesspairing.Float64Ptr(0.0)
	}
	if o.AbsentPenaltyFraction == nil {
		o.AbsentPenaltyFraction = chesspairing.Float64Ptr(0.35)
	}
	if o.ExcusedAbsentFraction == nil {
		o.ExcusedAbsentFraction = chesspairing.Float64Ptr(0.35)
	}
	if o.ClubCommitmentFraction == nil {
		o.ClubCommitmentFraction = chesspairing.Float64Ptr(0.70)
	}

	// Fixed-value overrides: nil = not set (use fractions). No defaults needed.

	// Behavioral options.
	if o.SelfVictory == nil {
		o.SelfVictory = chesspairing.BoolPtr(true)
	}
	if o.AbsenceLimit == nil {
		o.AbsenceLimit = chesspairing.IntPtr(5)
	}
	if o.AbsenceDecay == nil {
		o.AbsenceDecay = chesspairing.BoolPtr(false)
	}
	if o.Frozen == nil {
		o.Frozen = chesspairing.BoolPtr(false)
	}

	// Other.
	if o.LateJoinHandicap == nil {
		o.LateJoinHandicap = chesspairing.Float64Ptr(0)
	}

	return o
}

// ValueNumber calculates the value number for a player at the given rank.
// Rank is 1-based (rank 1 = strongest player).
func (o Options) ValueNumber(rank int) int {
	return *o.ValueNumberBase - (rank-1)**o.ValueNumberStep
}

// ParseOptions converts a map[string]any (from Firestore/JSON) into
// typed Options. Unrecognized keys are ignored. Type mismatches use defaults.
func ParseOptions(m map[string]any) Options {
	var o Options

	// Value number assignment.
	if v, ok := chesspairing.GetInt(m, "valueNumberBase"); ok {
		o.ValueNumberBase = &v
	}
	if v, ok := chesspairing.GetInt(m, "valueNumberStep"); ok {
		o.ValueNumberStep = &v
	}

	// Game result fractions.
	if v, ok := chesspairing.GetFloat64(m, "winFraction"); ok {
		o.WinFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "drawFraction"); ok {
		o.DrawFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "lossFraction"); ok {
		o.LossFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "forfeitWinFraction"); ok {
		o.ForfeitWinFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "forfeitLossFraction"); ok {
		o.ForfeitLossFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "doubleForfeitFraction"); ok {
		o.DoubleForfeitFraction = &v
	}

	// Non-game result fractions.
	if v, ok := chesspairing.GetFloat64(m, "byeValueFraction"); ok {
		o.ByeValueFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "halfByeFraction"); ok {
		o.HalfByeFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "zeroByeFraction"); ok {
		o.ZeroByeFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "absentPenaltyFraction"); ok {
		o.AbsentPenaltyFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "excusedAbsentFraction"); ok {
		o.ExcusedAbsentFraction = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "clubCommitmentFraction"); ok {
		o.ClubCommitmentFraction = &v
	}

	// Fixed-value overrides.
	if v, ok := chesspairing.GetInt(m, "byeFixedValue"); ok {
		o.ByeFixedValue = &v
	}
	if v, ok := chesspairing.GetInt(m, "halfByeFixedValue"); ok {
		o.HalfByeFixedValue = &v
	}
	if v, ok := chesspairing.GetInt(m, "zeroByeFixedValue"); ok {
		o.ZeroByeFixedValue = &v
	}
	if v, ok := chesspairing.GetInt(m, "absentFixedValue"); ok {
		o.AbsentFixedValue = &v
	}
	if v, ok := chesspairing.GetInt(m, "excusedAbsentFixedValue"); ok {
		o.ExcusedAbsentFixedValue = &v
	}
	if v, ok := chesspairing.GetInt(m, "clubCommitmentFixedValue"); ok {
		o.ClubCommitmentFixedValue = &v
	}

	// Behavioral options.
	if v, ok := chesspairing.GetBool(m, "selfVictory"); ok {
		o.SelfVictory = &v
	}
	if v, ok := chesspairing.GetInt(m, "absenceLimit"); ok {
		o.AbsenceLimit = &v
	}
	if v, ok := chesspairing.GetBool(m, "absenceDecay"); ok {
		o.AbsenceDecay = &v
	}
	if v, ok := chesspairing.GetBool(m, "frozen"); ok {
		o.Frozen = &v
	}

	// Other.
	if v, ok := chesspairing.GetFloat64(m, "lateJoinHandicap"); ok {
		o.LateJoinHandicap = &v
	}

	return o
}
