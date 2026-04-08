---
title: "Ordering Tiebreakers"
linkTitle: "Ordering"
weight: 6
description: "Pairing Number and Player Rating — deterministic ordering tiebreakers."
---

Ordering tiebreakers provide deterministic tie resolution when all other tiebreakers produce equal values. They are typically placed last in the tiebreaker chain to guarantee a total ordering of the standings.

## Tiebreakers

### pairing-number

**ID:** `pairing-number`
**Name:** Pairing Number (TPN)
**FIDE Category:** B

Uses the tournament pairing number (the 1-based index of the player in the `state.Players` slice) as the tiebreak value. The value is **negated** so that a lower pairing number (higher seeding) produces a higher tiebreak value, consistent with the convention that higher tiebreak values rank better.

**Algorithm:**

1. Assign each player their 1-based index in `state.Players`.
2. Negate the index: `value = -float64(index)`.

Player 1 gets value -1, player 2 gets value -2, and so on. Since -1 > -2, player 1 ranks above player 2.

**Formula:** `-(1-based player index)`

**Special handling:** This tiebreaker does not use `buildOpponentData()` and does not examine game results. It is purely positional.

### player-rating

**ID:** `player-rating`
**Name:** Player Rating (RTNG)
**FIDE Category:** D

Uses the player's registered rating as the tiebreak value. Higher rating ranks higher.

**Algorithm:**

1. Look up the player's `Rating` field from `state.Players`.
2. Return it as a float64.

**Formula:** `float64(player.Rating)`

**Special handling:** This tiebreaker does not use `buildOpponentData()` and does not examine game results. It uses only the static rating registered at the start of the tournament.

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Pairing Number (lower TPN = higher value)
tb, err := tiebreaker.Get("pairing-number")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Player Rating
tb, err = tiebreaker.Get("player-rating")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player.
