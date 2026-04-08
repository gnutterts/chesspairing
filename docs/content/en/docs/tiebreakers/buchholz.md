---
title: "Buchholz Family"
linkTitle: "Buchholz"
weight: 1
description: "Five Buchholz variants — Full, Cut-1, Cut-2, Median, and Median-2 — based on opponents' scores."
---

The Buchholz tiebreaker measures the strength of a player's opposition by summing the final scores of all their opponents. A higher Buchholz value indicates the player faced stronger competition. Five variants are registered, each differing only in how many extreme opponent scores are trimmed before summing.

All five variants belong to **FIDE Category A** (based on results of opponents).

## Shared algorithm

Every Buchholz variant follows the same core steps:

1. **Collect opponent scores.** For each OTB game (non-forfeit, non-pending), look up the opponent's final tournament score.
2. **Add virtual opponents.** For each round where the player had a bye or was absent (no real opponent), add a virtual opponent score equal to the player's own final score.
3. **Sort ascending.** The collected scores are sorted from lowest to highest.
4. **Trim.** Depending on the variant, remove scores from the low end, the high end, or both.
5. **Sum.** Add the remaining scores to produce the tiebreak value.

### Forfeit and bye handling

The shared `buildOpponentData()` function excludes all forfeits from the game entry list. Only OTB results (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) produce game entries with a real opponent. Forfeit wins, forfeit losses, double forfeits, and pending games are skipped entirely.

For rounds where a player did not play an OTB game:

- **Byes** (PAB, half-point, zero-point) increment the player's bye count.
- **Absences** (active player not appearing in any game or bye for a round) increment the absence count.

Each bye and absence contributes a virtual opponent score equal to the player's own final score.

## Variants

| ID                 | Name              | Trimming rule                     |
| ------------------ | ----------------- | --------------------------------- |
| `buchholz`         | Buchholz          | None -- sum all opponent scores   |
| `buchholz-cut1`    | Buchholz Cut-1    | Drop the 1 lowest opponent score  |
| `buchholz-cut2`    | Buchholz Cut-2    | Drop the 2 lowest opponent scores |
| `buchholz-median`  | Buchholz Median   | Drop the 1 highest AND 1 lowest   |
| `buchholz-median2` | Buchholz Median-2 | Drop the 2 highest AND 2 lowest   |

### buchholz

**ID:** `buchholz`
**Name:** Buchholz
**FIDE Category:** A

The full Buchholz. No trimming -- every opponent score (real or virtual) is summed.

**Formula:** `SUM(all opponent scores)`

### buchholz-cut1

**ID:** `buchholz-cut1`
**Name:** Buchholz Cut-1
**FIDE Category:** A

After sorting opponent scores ascending, the single lowest score is dropped before summing.

**Formula:** `SUM(opponent scores[1..n])` (skip index 0)

### buchholz-cut2

**ID:** `buchholz-cut2`
**Name:** Buchholz Cut-2
**FIDE Category:** A

After sorting, the two lowest opponent scores are dropped. If fewer than three opponents exist, nothing remains to sum (value is 0).

**Formula:** `SUM(opponent scores[2..n])` (skip indices 0 and 1)

### buchholz-median

**ID:** `buchholz-median`
**Name:** Buchholz Median
**FIDE Category:** A

Both the highest and lowest opponent scores are dropped. This removes the single best and single worst opponent, reducing the impact of extreme pairings.

**Formula:** `SUM(opponent scores[1..n-1])` (skip first and last)

### buchholz-median2

**ID:** `buchholz-median2`
**Name:** Buchholz Median-2
**FIDE Category:** A

The two highest and two lowest opponent scores are dropped. Requires at least five opponents to have any remaining scores.

**Formula:** `SUM(opponent scores[2..n-2])` (skip first two and last two)

## Example

A player with opponents who scored [2.0, 3.0, 3.5, 4.0, 5.0]:

| Variant          | Dropped            | Remaining                 | Value |
| ---------------- | ------------------ | ------------------------- | ----- |
| buchholz         | none               | [2.0, 3.0, 3.5, 4.0, 5.0] | 17.5  |
| buchholz-cut1    | 2.0                | [3.0, 3.5, 4.0, 5.0]      | 15.5  |
| buchholz-cut2    | 2.0, 3.0           | [3.5, 4.0, 5.0]           | 12.5  |
| buchholz-median  | 2.0, 5.0           | [3.0, 3.5, 4.0]           | 10.5  |
| buchholz-median2 | 2.0, 3.0, 4.0, 5.0 | [3.5]                     | 3.5   |

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Full Buchholz
tb, err := tiebreaker.Get("buchholz")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Buchholz Cut-1
tbCut1, err := tiebreaker.Get("buchholz-cut1")
values, err = tbCut1.Compute(ctx, state, scores)

// Buchholz Cut-2
tbCut2, err := tiebreaker.Get("buchholz-cut2")
values, err = tbCut2.Compute(ctx, state, scores)

// Buchholz Median
tbMedian, err := tiebreaker.Get("buchholz-median")
values, err = tbMedian.Compute(ctx, state, scores)

// Buchholz Median-2
tbMedian2, err := tiebreaker.Get("buchholz-median2")
values, err = tbMedian2.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player, where `Value` is the computed Buchholz score.
