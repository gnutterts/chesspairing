---
title: "Performance-Based Tiebreakers"
linkTitle: "Performance"
weight: 2
description: "TPR, PTP, APRO, APPO, and ARO — tiebreakers derived from ratings and expected scores."
---

Performance-based tiebreakers use player ratings and the FIDE B.02 conversion table to derive tiebreak values. They measure how well a player performed relative to the rating strength of their opponents.

All five tiebreakers in this group belong to **FIDE Category D** (based on ratings).

## FIDE B.02 conversion table

Two lookup functions underpin the rating-based tiebreakers:

- **`dpFromP(p)`** -- Given a fractional score `p` (0.0 to 1.0), returns the rating difference `dp`. The table maps 101 entries from p=0.00 (dp=-800) through p=0.50 (dp=0) to p=1.00 (dp=+800). Values between table entries are linearly interpolated.
- **`expectedScore(dp)`** -- The inverse lookup. Given a rating difference `dp` (-800 to +800), returns the expected fractional score. Also linearly interpolated between table entries.

## Forfeit handling

All performance-based tiebreakers use `buildOpponentData()`, which excludes all forfeits from game entries. Only OTB results (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) produce game entries. Forfeit wins, forfeit losses, double forfeits, and pending games are skipped. Players with no OTB games receive a value of 0.

## Tiebreakers

### aro

**ID:** `aro`
**Name:** Avg Rating of Opponents
**FIDE Category:** D

The arithmetic mean of the ratings of all OTB opponents.

**Algorithm:**

1. For each OTB game in the player's game list, look up the opponent's rating.
2. Sum all opponent ratings.
3. Divide by the number of OTB games.

If the player has no OTB games, the value is 0.

**Formula:** `SUM(opponent ratings) / number of OTB games`

### performance-rating

**ID:** `performance-rating`
**Name:** Performance Rating (TPR)
**FIDE Category:** D

The Tournament Performance Rating combines the average opponent rating with a rating-difference adjustment derived from the player's fractional score.

**Algorithm:**

1. Compute ARO (average rating of opponents from OTB games).
2. Compute the fractional score: `p = player score / number of OTB games`, clamped to [0.0, 1.0].
3. Look up `dp = dpFromP(p)` from the FIDE B.02 table (with linear interpolation).
4. `TPR = round(ARO + dp)`.

If the player has no OTB games, the value is 0. The result is rounded to the nearest integer (0.5 rounds up).

**Formula:** `round(ARO + dpFromP(score / games))`

### performance-points

**ID:** `performance-points`
**Name:** Performance Points (PTP)
**FIDE Category:** D

PTP finds the lowest hypothetical rating R such that the sum of expected scores against all opponents would reach or exceed the player's actual score. It uses binary search over the FIDE expected-score function.

**Algorithm:**

1. Collect all OTB opponent ratings.
2. **Zero score:** value = `round(lowest opponent rating - 800)`.
3. **Perfect score:** value = `round(highest opponent rating + 800)`.
4. **Otherwise:** binary search for the lowest R in range `[min opponent rating - 800, max opponent rating + 800]` where `SUM(expectedScore(R - oppRating_i))` >= actual score. Search precision is 0.5 rating points.
5. Result is rounded to the nearest integer.

If the player has no OTB games, the value is 0.

**Formula:** `min R such that SUM(expectedScore(R - oppRating_i)) >= score`

### avg-opponent-tpr

**ID:** `avg-opponent-tpr`
**Name:** Avg Opponent TPR (APRO)
**FIDE Category:** D

The average of all OTB opponents' Tournament Performance Ratings.

**Algorithm:**

1. Compute TPR for every player in the tournament (using the `performance-rating` tiebreaker).
2. For each of the player's OTB opponents, collect their TPR value.
3. Average those values.
4. Round to the nearest integer.

If the player has no OTB games, the value is 0.

**Formula:** `round(SUM(opponent TPR values) / number of OTB games)`

### avg-opponent-ptp

**ID:** `avg-opponent-ptp`
**Name:** Avg Opponent PTP (APPO)
**FIDE Category:** D

The average of all OTB opponents' Performance Points values.

**Algorithm:**

1. Compute PTP for every player in the tournament (using the `performance-points` tiebreaker).
2. For each of the player's OTB opponents, collect their PTP value.
3. Average those values.
4. Round to the nearest integer.

If the player has no OTB games, the value is 0.

**Formula:** `round(SUM(opponent PTP values) / number of OTB games)`

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Average Rating of Opponents
tb, err := tiebreaker.Get("aro")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Performance Rating (TPR)
tb, err = tiebreaker.Get("performance-rating")
values, err = tb.Compute(ctx, state, scores)

// Performance Points (PTP)
tb, err = tiebreaker.Get("performance-points")
values, err = tb.Compute(ctx, state, scores)

// Average Opponent TPR (APRO)
tb, err = tiebreaker.Get("avg-opponent-tpr")
values, err = tb.Compute(ctx, state, scores)

// Average Opponent PTP (APPO)
tb, err = tiebreaker.Get("avg-opponent-ptp")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player. For `aro`, the value is a float64 (not rounded). For all others, the value is rounded to the nearest integer.
