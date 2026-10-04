---
title: "Fore Buchholz and Average Opponent Buchholz"
linkTitle: "Opponent Buchholz"
weight: 7
description: "Buchholz variants that handle incomplete rounds or average opponent contributions."
---

These two tiebreakers extend the [Buchholz](../buchholz/) concept for specific situations: Fore Buchholz handles tournaments with an incomplete final round, and Average Opponent Buchholz normalizes Buchholz values across opponents.

Both belong to **FIDE Category C** (based on opponents' results, advanced variants).

## Forfeit handling

Both tiebreakers use the same per-round records as the [Buchholz family](../buchholz/). Swiss unplayed rounds follow the Article 16 adjusted-score and capped-dummy model; Article 15.2 uses forfeits as regular encounters in predetermined pairings. Pending games are not completed encounters.

## Tiebreakers

### fore-buchholz

**ID:** `fore-buchholz`
**Name:** Fore Buchholz
**FIDE Category:** C

Fore Buchholz computes the full Buchholz score as if all pending final-round games ended in draws. This allows standings to be computed before the last round is complete. If all games in the final round are already finished, Fore Buchholz equals regular Buchholz.

**Algorithm:**

1. Start with actual scores for all players.
2. Identify pending games in the last round (`ResultPending` in `state.Rounds[last]`).
3. For each pending game, add +0.5 to both the White and Black player's virtual score.
4. Build the normal per-round records, then treat pending final-round games as draws for this calculation.
5. Use the virtual scores from step 3 and apply the Article 16 adjusted-score and capped-dummy rules when calculating Buchholz.

**Formula:** `Buchholz(modified state where pending last-round games = draws)`

**Special handling:**

- Only the final round's pending games are treated as draws. Pending games in earlier rounds (if any) remain excluded.
- The virtual score adjustment (+0.5 per pending game) propagates through opponent score lookups, affecting all players whose opponents have pending games.
- If no rounds exist, all values are 0.

### avg-opponent-buchholz

**ID:** `avg-opponent-buchholz`
**Name:** Avg Opponent Buchholz (AOB)
**FIDE Category:** C

Average Opponent Buchholz first computes the full Buchholz for every player, then for each player averages the Buchholz values of their OTB opponents. This normalizes the tiebreaker across players who may have played different numbers of games.

**Algorithm:**

1. Compute full Buchholz for every scored player from the per-round records, applying Articles 15 and 16.
2. For each player, iterate their OTB game entries:
   - Sum the Buchholz values of each opponent.
   - Divide by the number of OTB games.

If the player has no OTB games, the value is 0.

**Formula:** `SUM(opponent Buchholz values) / number of OTB games`

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Fore Buchholz (handles pending final-round games)
tb, err := tiebreaker.Get("fore-buchholz")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Average Opponent Buchholz
tb, err = tiebreaker.Get("avg-opponent-buchholz")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player.
