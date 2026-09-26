---
title: "Keizer Scoring"
linkTitle: "Keizer"
weight: 2
description: "Iterative ranking-based scoring where points depend on opponent strength — converges through up to 20 iterations."
---

Keizer scoring is a ranking-based system popular in club tournaments in Belgium and the Netherlands. The central idea: beating a strong opponent earns more points than beating a weak one. Each player is assigned a value number based on their current rank, and game points are computed as fractions of the opponent's value number. Because value numbers depend on rankings, and rankings depend on scores, the system must iterate until it converges.

The engine uses x2 integer arithmetic internally to eliminate floating-point drift while preserving half-point granularity. Convergence typically happens within 3-5 iterations, with a hard cap at 20 and oscillation detection to guarantee termination.

## When to use

- **Club internal competitions.** Keizer is designed for long-running club leagues where players may miss rounds regularly. The absence-handling system (fractions, limits, decay, club commitments) is far richer than standard scoring.
- **Tournaments where opponent strength should matter.** In standard scoring, a win is always worth 1 point regardless of the opponent. In Keizer, beating the top-ranked player earns significantly more than beating the bottom-ranked player.
- **Combined pairing and scoring.** The Keizer pairer uses the Keizer scorer internally to rank players before pairing, creating a tightly integrated system. However, the scorer can also be used independently with any pairing system.

Keizer scoring is not suitable for FIDE-rated events, which require standard 1-half-0 scoring.

## Configuration

### CLI

Pass Keizer scoring options through the TRF `XXY` field or via `--config`:

```bash
chesspairing pair --config '{"scoring": {"winFraction": 1.0, "drawFraction": 0.5, "selfVictory": true}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/keizer"

// With explicit options (nil fields use defaults).
scorer := keizer.New(keizer.Options{
    LossFraction: chesspairing.Float64Ptr(1.0 / 6.0),
    AbsenceLimit: chesspairing.IntPtr(0),
})

// From a generic map (e.g. parsed from JSON config).
scorer := keizer.NewFromMap(map[string]any{
    "lossFraction": 1.0 / 6.0,
    "absenceLimit": 0,
})

// Use the Scorer interface.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

The `Scorer` type satisfies the `chesspairing.Scorer` interface at compile time:

```go
var _ chesspairing.Scorer = (*keizer.Scorer)(nil)
```

### Options reference

Options are organized in five groups. All pointer fields use `nil` to mean "use the default."

#### Value numbers

These control how value numbers are assigned from rankings. The player at rank _r_ gets: `ValueNumberBase - (r-1) * ValueNumberStep`.

| Field             | Type   | JSON key          | Default | Description                                                           |
| ----------------- | ------ | ----------------- | ------- | --------------------------------------------------------------------- |
| `ValueNumberBase` | `*int` | `valueNumberBase` | N       | Value number for the top-ranked player. N = number of active players. |
| `ValueNumberStep` | `*int` | `valueNumberStep` | 1       | Decrement per rank position.                                          |

#### Game fractions

Fractions of the **opponent's** value number awarded for game results.

| Field                   | Type       | JSON key                | Default | Description                                  |
| ----------------------- | ---------- | ----------------------- | ------- | -------------------------------------------- |
| `WinFraction`           | `*float64` | `winFraction`           | 1.0     | Fraction for a win.                          |
| `DrawFraction`          | `*float64` | `drawFraction`          | 0.5     | Fraction for a draw.                         |
| `LossFraction`          | `*float64` | `lossFraction`          | 0.0     | Fraction for a loss.                         |
| `ForfeitWinFraction`    | `*float64` | `forfeitWinFraction`    | 1.0     | Fraction for winning by forfeit.             |
| `ForfeitLossFraction`   | `*float64` | `forfeitLossFraction`   | 0.0     | Fraction for losing by forfeit.              |
| `DoubleForfeitFraction` | `*float64` | `doubleForfeitFraction` | 0.0     | Fraction for a double forfeit (each player). |

For a double forfeit, `ResultDoubleForfeit` selects `DoubleForfeitFraction`
regardless of `GameData.IsForfeit`. For a single forfeit, `GameData.IsForfeit`
selects the forfeit win and loss fractions; the ordinary win/loss result still
identifies which player receives each fraction.

#### Non-game fractions

Fractions of the **player's own** value number awarded for non-game situations.

| Field                    | Type       | JSON key                 | Default | Description                                      |
| ------------------------ | ---------- | ------------------------ | ------- | ------------------------------------------------ |
| `ByeValueFraction`       | `*float64` | `byeValueFraction`       | 0.50    | Fraction for a pairing-allocated bye (PAB).      |
| `HalfByeFraction`        | `*float64` | `halfByeFraction`        | 0.50    | Fraction for a half-point bye.                   |
| `ZeroByeFraction`        | `*float64` | `zeroByeFraction`        | 0.0     | Fraction for a zero-point bye.                   |
| `AbsentPenaltyFraction`  | `*float64` | `absentPenaltyFraction`  | 0.35    | Fraction for an unexcused absence.               |
| `ExcusedAbsentFraction`  | `*float64` | `excusedAbsentFraction`  | 0.35    | Fraction for an excused absence.                 |
| `ClubCommitmentFraction` | `*float64` | `clubCommitmentFraction` | 0.70    | Fraction for absence due to interclub team duty. |

#### Fixed-value overrides

When set (non-nil), these replace the corresponding fraction calculation with a fixed score. Values are in real units (not x2). Leave `nil` to use the fraction-based calculation.

| Field                      | Type   | JSON key                   | Default | Description                        |
| -------------------------- | ------ | -------------------------- | ------- | ---------------------------------- |
| `ByeFixedValue`            | `*int` | `byeFixedValue`            | nil     | Fixed score for PAB.               |
| `HalfByeFixedValue`        | `*int` | `halfByeFixedValue`        | nil     | Fixed score for half-point bye.    |
| `ZeroByeFixedValue`        | `*int` | `zeroByeFixedValue`        | nil     | Fixed score for zero-point bye.    |
| `AbsentFixedValue`         | `*int` | `absentFixedValue`         | nil     | Fixed score for unexcused absence. |
| `ExcusedAbsentFixedValue`  | `*int` | `excusedAbsentFixedValue`  | nil     | Fixed score for excused absence.   |
| `ClubCommitmentFixedValue` | `*int` | `clubCommitmentFixedValue` | nil     | Fixed score for club commitment.   |

#### Behavioral options

| Field                | Type       | JSON key             | Default       | Description                                                                                                                                                   |
| -------------------- | ---------- | -------------------- | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Method`             | `*string`  | `method`             | `"iterative"` | Scoring method: `"iterative"` (default), `"frozen"`, or `"keizer1956"`. Takes precedence over the deprecated `Frozen` flag.                              |
| `Frozen`             | `*bool`    | `frozen`             | false         | Deprecated alias for `Method="frozen"`. When `Method` is unset and `Frozen` is true, the `"frozen"` method is used.                                         |
| `WithdrawnAsAbsent`  | `*bool`    | `withdrawnAsAbsent`  | false         | Keep a withdrawn player in the standings. From the round after `WithdrawnAfterRound` the player scores absence points up to `AbsenceLimit`.                    |
| `ForfeitCountsAsMet` | `*bool`    | `forfeitCountsAsMet` | true          | Whether a double forfeit counts as an encounter. When false, both players score absence points for that round instead.                                         |
| `SelfVictory`        | `*bool`    | `selfVictory`        | true          | Add each player's own value number to their total (once, not per round).                                                                                      |
| `AbsenceLimit`       | `*int`     | `absenceLimit`       | 5             | Maximum absences that score points. Beyond this limit, absences score 0. Club commitments are exempt. 0 = unlimited.                                          |
| `AbsenceDecay`       | `*bool`    | `absenceDecay`       | false         | Halve the absence score for each cumulative absence (1st = full, 2nd = half, 3rd = quarter, ...). Club commitments are exempt.                                |
| `LateJoinHandicap`   | `*float64` | `lateJoinHandicap`   | 0             | Fixed score awarded per round missed before the player joined. Requires `PlayerEntry.JoinedRound` to be set. Not subject to `AbsenceLimit` or `AbsenceDecay`. |

## How it works

### Score()

Keizer scoring is iterative because it has a circular dependency: scores depend on value numbers, value numbers depend on rankings, and rankings depend on scores. The algorithm resolves this through repeated recalculation:

1. **Initial ranking.** Rank all active players by rating descending, then by display name ascending.

2. **Iterate** (up to 20 times):

   a. **Reset scores.** Clear all x2 scores to zero.

   b. **Compute value numbers.** From the current ranking, assign each player a value number: `ValueNumberBase - (rank-1) * ValueNumberStep`. The top-ranked player gets the highest value.

   c. **Score all rounds.** For each round, process games, byes, absences, and late-join rounds:
   - _Games:_ points = `round(opponent_value * fraction * 2)` using x2 integer arithmetic. For example, a win against a player with value 20 at WinFraction=1.0 yields `20 * 1.0 * 2 = 40` in x2 units.
   - _Byes:_ points = `round(own_value * fraction * 2)`, or `fixed_value * 2` when a fixed-value override is set. Club commitments are exempt from the absence limit and decay.
   - _Absences:_ same as byes using `AbsentPenaltyFraction` or `AbsentFixedValue`, subject to the absence limit and decay. Excused absences count toward the limit; club commitments do not.
   - _Late-join rounds:_ for players with `JoinedRound > 1`, rounds before the join round score `LateJoinHandicap` as a flat value instead of going through the absence calculation. These rounds do not count toward the absence limit and are not affected by decay.

   d. **Self-victory.** If enabled, add `own_value * 2` to each player's x2 total (once, not per round).

   e. **Re-rank.** Sort players by x2 score descending, rating descending, display name ascending.

   f. **Check convergence.** If the ranking is unchanged from the previous iteration, stop.

   g. **Check for oscillation.** If the ranking matches the one from two iterations ago (a 2-cycle), average the x2 scores from the last two iterations, re-rank, and stop. This handles cases where two players with very similar scores keep swapping positions.

3. **Convert to real scores.** Divide all x2 scores by 2 to produce the final values.

### Scoring methods

The `Method` option selects between three scoring methods:

| Method         | Behaviour                                                                                                     |
| -------------- | ------------------------------------------------------------------------------------------------------------- |
| `"iterative"`  | Default. Repeatedly rescore all rounds with the converged ranking (see the loop above).                       |
| `"frozen"`     | Single pass: each round is scored once with the ranking before that round; earlier rounds are never rescored. |
| `"keizer1956"` | Historical Keizer 1956 method: after each round all scores are recalculated once with the ranking before that round, and each player's own value is added. |

`Method="frozen"` is equivalent to the deprecated `Frozen=true`. When `Method` is set, it takes precedence over `Frozen`.

### Frozen mode

When `Frozen` is set to `true`, the iterative loop is replaced by a sequential pass through the rounds. Each round is scored once using the ranking as it stood before that round, and the ranking is updated afterward. Earlier rounds are never rescored when later results shift the standings.

The sequence:

1. Start with the initial rating-based ranking.
2. For each round in order, compute value numbers from the current ranking, score games/byes/absences for that round, and re-rank.
3. After all rounds, add self-victory (if enabled) using the final ranking.

This produces different results from the standard iterative mode. In standard mode, all rounds are rescored retroactively with the converged ranking, so a player's round-1 win is worth the opponent's final value number. In frozen mode, that same win is worth the opponent's value number at the time -- which may have been higher or lower before later rounds shifted things around.

Frozen mode is useful for clubs that want scores to reflect the standings as they developed over the season, rather than retroactively rewriting history from the endpoint.

### Keizer 1956 method

The `"keizer1956"` method implements the historical Keizer 1956 calculation.
After every round, all scores are recalculated once using the ranking that
stood before that round. Every game played up to and including the current
round uses the opponent's value number from that same ranking, and each
player's own value number is added to their total. The resulting ranking
becomes the input for the next round.

Unlike `"iterative"`, there is no convergence loop; unlike `"frozen"`, earlier
rounds are rescored on every round, so a change in the ranking can
retroactively revalue earlier results. Byes and absences follow the same
options as the other methods, scoring a fixed value or a fraction of the
player's own value number from the ranking before that round.

### Why x2 integer arithmetic

Each contribution is rounded in x2 units: `round(value * fraction * 2)` for a fraction and `fixed_value * 2` for a fixed value. Thus a half point is one x2 unit; all x2 totals are divided by 2 only for the final output.

### Absence rules

- **Absence limit.** After `AbsenceLimit` absences, all further absences score 0. This prevents players who barely participate from accumulating meaningful scores.
- **Absence decay.** When enabled, each cumulative absence earns half the previous one: 1st = full fraction, 2nd = fraction/2, 3rd = fraction/4, and so on (implemented as a right bit-shift on the x2 value).
- **Club commitments** are always exempt from both the limit and the decay. A player who misses rounds for interclub team duty is not penalized the way a regular absence would be.
- **Excused absences** receive their own fraction (`ExcusedAbsentFraction`) but do count toward the absence limit and decay.
- **Late joiners.** When a player has `JoinedRound > 1`, rounds before the join round are scored using `LateJoinHandicap` as a fixed value rather than the normal absence logic. These pre-join rounds bypass the absence limit and decay entirely, so a late joiner's actual absences (after joining) are counted from zero.

## Variant presets

Several well-known Keizer variants can be configured by setting specific options:

| Variant               | Key differences from defaults                                                                                                                                                             |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **KeizerForClubs**    | All defaults. The most widely used variant.                                                                                                                                               |
| **Classic KNSB**      | `ByeValueFraction`=4/6, `AbsentPenaltyFraction`=2/6, `ClubCommitmentFraction`=2/3, `ExcusedAbsentFraction`=2/6, `AbsenceLimit`=5. Based on the traditional Dutch chess federation system. |
| **FreeKeizer**        | `LossFraction`=1/6, `ByeValueFraction`=4/6, `AbsentPenaltyFraction`=2/6, `AbsenceLimit`=5. Adds a "toughness bonus" where losses against strong opponents earn a small reward.            |
| **No self-victory**   | `SelfVictory`=false. Removes the bonus for participation.                                                                                                                                 |
| **Fixed absences**    | `AbsentFixedValue`=15, `ExcusedAbsentFixedValue`=15, `ClubCommitmentFixedValue`=25. Uses absolute values instead of fractions.                                                            |
| **Decaying absences** | `AbsenceDecay`=true, `AbsenceLimit`=0. Each absence earns less than the previous one, with no hard cap.                                                                                   |
| **ESG Emmen**         | Use `keizer.ESGOptions()` (https://esgemmen.nl). Value numbers 60, 59, ..., win/draw/loss 1.0/0.5/0, first five absences 20 points, club commitment 40, bye 40, late-join handicap 15, no recalculation, no self-victory, withdrawn players stay in the standings, double forfeits do not count as encounters. |

## Examples

### Default KeizerForClubs

```go
scorer := keizer.New(keizer.Options{})
scores, _ := scorer.Score(ctx, &state)
```

With 10 players, the top-ranked player has value number 10. If that player beats the 3rd-ranked player (value 8): points = 8 _ 1.0 = 8.0. If they draw: 8 _ 0.5 = 4.0. Self-victory adds their own value (10) once.

### FreeKeizer with toughness bonus

```go
scorer := keizer.New(keizer.Options{
    LossFraction:          chesspairing.Float64Ptr(1.0 / 6.0),
    ByeValueFraction:      chesspairing.Float64Ptr(4.0 / 6.0),
    AbsentPenaltyFraction: chesspairing.Float64Ptr(2.0 / 6.0),
    AbsenceLimit:          chesspairing.IntPtr(5),
})
```

Losing to the top player (value 10) now earns 10 \* 1/6 = 1.5 points (rounded to x2 grid) instead of 0.

### Fixed absence values

```go
scorer := keizer.New(keizer.Options{
    AbsentFixedValue:         chesspairing.IntPtr(15),
    ExcusedAbsentFixedValue:  chesspairing.IntPtr(15),
    ClubCommitmentFixedValue: chesspairing.IntPtr(25),
})
```

All players receive the same fixed score for absences, regardless of their rank.

### Keizer 1956

```go
scorer := keizer.New(keizer.Options{
    Method:           chesspairing.StringPtr("keizer1956"),
    ValueNumberBase:  chesspairing.IntPtr(50),
    ValueNumberStep:  chesspairing.IntPtr(1),
})
```

This reproduces the historical Keizer 1956 two-round example: after each
round all scores are recalculated once using the ranking before that round.

### ESG Emmen profile

```go
scorer := keizer.New(keizer.ESGOptions())
```

`ESGOptions()` configures the scoring rules used by ESG Emmen
(https://esgemmen.nl): value numbers 60, 59, ..., a win worth the full
opponent value, the first five absences worth 20 points, a club commitment or
bye worth 40, a 15-point handicap per missed round for newcomers, no
recalculation and no self-victory. Withdrawn players stay in the standings
and a double forfeit does not count as an encounter.

## Related

- [Scoring concepts](/docs/concepts/scoring/) -- overview of all three scoring systems and how they interact with pairing
- [Keizer convergence algorithm](/docs/algorithms/keizer-convergence/) -- detailed analysis of the iterative convergence and oscillation detection
- [Keizer pairing system](/docs/pairing-systems/keizer/) -- the pairer that uses Keizer scoring internally for ranking
- [Standard scoring](/docs/scoring/standard/) -- the fixed-point alternative
- [Byes](/docs/concepts/byes/) -- bye types and their scoring across all systems
- [Scorer interface](/docs/api/scorer/) -- API reference for the `Scorer` interface
