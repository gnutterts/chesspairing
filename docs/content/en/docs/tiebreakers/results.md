---
title: "Result-Based Tiebreakers"
linkTitle: "Results"
weight: 3
description: "Wins, Rounds Won, Standard Points, Progressive Score, and Koya System."
---

Result-based tiebreakers derive their values directly from game outcomes and round-by-round scoring. Unlike [Buchholz](../buchholz/) or [performance-based](../performance/) tiebreakers, they do not consider opponent strength or ratings. All five belong to **FIDE Category B** (based on the player's own results).

## Forfeit handling

The shared `buildOpponentData()` function excludes all forfeits from game entries. Only OTB results (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) produce game entries. Some tiebreakers in this group work directly with round data rather than the opponent data structure, and their specific forfeit handling is documented per tiebreaker below.

## Tiebreakers

### wins

**ID:** `wins`
**Name:** Games Won (OTB)
**FIDE Category:** B

Counts the number of OTB wins. Only `resultWin` entries from the player's game list (built by `buildOpponentData()`) are counted. Since `buildOpponentData()` excludes all forfeits, this strictly counts over-the-board victories.

**Algorithm:**

1. Iterate the player's game entries from `buildOpponentData()`.
2. Count entries where `result == resultWin`.

Byes, forfeit wins, and draws do not contribute.

**Formula:** `COUNT(games where result = win)`

### win

**ID:** `win`
**Name:** Rounds Won
**FIDE Category:** B

Counts the number of rounds where the player received win-equivalent points. This is broader than `wins` -- it includes OTB wins, forfeit wins, and full-point byes (PAB).

**Algorithm:**

1. Iterate all games in all rounds. For each game:
   - `ResultWhiteWins` or `ResultForfeitWhiteWins`: increment White's count.
   - `ResultBlackWins` or `ResultForfeitBlackWins`: increment Black's count.
2. Iterate all byes. For each `ByePAB`: increment the player's count.
3. Half-point byes, zero-point byes, draws, pending games, and double forfeits do not count.

**Formula:** `COUNT(OTB wins + forfeit wins + PAB byes)`

### standard-points

**ID:** `standard-points`
**Name:** Standard Points
**FIDE Category:** B

Normalizes each round's result to a standard 1/0.5/0 scale regardless of the tournament's scoring system. This is useful in tournaments with non-standard point values (e.g., football scoring 3-1-0).

**Algorithm:**

For each round, determine the player's awarded points and compare:

1. **If opponent exists:** compare the player's points with the opponent's points for that game.
   - Player scored more than opponent: +1.0
   - Equal: +0.5
   - Player scored less: +0.0
2. **If no opponent** (bye or absent): compare the player's awarded points to 0.5.
   - PAB (1.0 points) > 0.5: +1.0
   - Half-bye (0.5 points) = 0.5: +0.5
   - Zero-bye or absent (0.0 points) < 0.5: +0.0
3. Sum across all rounds.

**Formula:** `SUM(per-round standard result)`

### progressive

**ID:** `progressive`
**Name:** Progressive Score
**FIDE Category:** B

The progressive (cumulative) score rewards players who win early. It builds per-round scores, computes cumulative totals after each round, then sums all cumulative values.

**Algorithm:**

1. Build per-round scores for each player:
   - Win or forfeit win: 1.0
   - Draw: 0.5
   - Loss, forfeit loss, double forfeit: 0.0
   - PAB: 1.0, half-bye: 0.5, zero-bye/absent: 0.0
2. Compute cumulative scores: after round 1, after round 2, etc.
3. Sum all cumulative values.

**Example:** A player scoring 1, 0, 1, 1 across four rounds:

- Per-round: [1.0, 0.0, 1.0, 1.0]
- Cumulative: [1.0, 1.0, 2.0, 3.0]
- Progressive = 1.0 + 1.0 + 2.0 + 3.0 = **7.0**

Compare with a player scoring 0, 1, 1, 1 (same total of 3.0):

- Per-round: [0.0, 1.0, 1.0, 1.0]
- Cumulative: [0.0, 1.0, 2.0, 3.0]
- Progressive = 0.0 + 1.0 + 2.0 + 3.0 = **6.0**

The first player ranks higher because they won earlier.

**Formula:** `SUM(cumulative scores after each round)`

### koya

**ID:** `koya`
**Name:** Koya System
**FIDE Category:** B

The Koya system counts points scored against opponents in the top half of the standings. It is particularly useful in round-robin tournaments.

**Algorithm:**

1. Compute the qualifying threshold: `totalRounds / 2`.
2. Identify "qualifying opponents": players whose final score >= threshold.
3. For each of the player's OTB games against a qualifying opponent:
   - Win: +1.0
   - Draw: +0.5
   - Loss: +0.0
4. Sum the contributions.

Only OTB games contribute (via `buildOpponentData()`). Forfeits, byes, and games against non-qualifying opponents are excluded.

**Formula:** `SUM(results against opponents with score >= totalRounds/2)`

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Games Won (OTB only)
tb, err := tiebreaker.Get("wins")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Rounds Won (OTB wins + forfeit wins + PAB)
tb, err = tiebreaker.Get("win")
values, err = tb.Compute(ctx, state, scores)

// Standard Points
tb, err = tiebreaker.Get("standard-points")
values, err = tb.Compute(ctx, state, scores)

// Progressive Score
tb, err = tiebreaker.Get("progressive")
values, err = tb.Compute(ctx, state, scores)

// Koya System
tb, err = tiebreaker.Get("koya")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player.
