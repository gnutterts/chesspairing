---
title: "Head-to-Head Tiebreakers"
linkTitle: "Head-to-Head"
weight: 4
description: "Direct Encounter and Sonneborn-Berger — tiebreakers that consider results against specific opponents."
---

Head-to-head tiebreakers resolve ties by examining results between specific opponents rather than aggregating across the full field. Direct Encounter looks only at games among tied players, while Sonneborn-Berger weights each result by the opponent's final score.

Both tiebreakers belong to **FIDE Category A** (based on results of opponents).

## Forfeit handling

Both tiebreakers use `buildOpponentData()`, which excludes all forfeits from game entries. Only OTB results (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) produce game entries. Forfeit wins, forfeit losses, double forfeits, and pending games generate no game entries and do not contribute to either tiebreaker.

## Tiebreakers

### direct-encounter

**ID:** `direct-encounter`
**Name:** Direct Encounter
**FIDE Category:** A

The direct encounter tiebreaker considers only games played between members of the same tied group. Players who are not tied with anyone receive a value of 0.

**Algorithm:**

1. Group all players by their primary score. Each group of players sharing the same score forms a "tied group."
2. For groups with only one player (no tie), the value is 0.
3. For each player in a tied group, iterate their OTB game entries:
   - If the opponent is also in the same tied group:
     - Win: +1.0
     - Draw: +0.5
     - Loss: +0.0
   - If the opponent is NOT in the tied group: skip.
4. Sum the contributions.

**Special considerations:**

- Games against players outside the tied group are completely ignored.
- If two tied players never played each other OTB, their direct encounter value reflects only games against other members of the tied group.
- In Swiss tournaments with many players on the same score, this tiebreaker can be decisive when specific head-to-head matchups occurred.

**Formula:** `SUM(standard results from OTB games against tied-group opponents)`

### sonneborn-berger

**ID:** `sonneborn-berger`
**Name:** Sonneborn-Berger
**FIDE Category:** A

Sonneborn-Berger (SB) weights each game result by the opponent's final tournament score. Wins against strong opponents contribute more than wins against weak opponents. This is one of the most widely used tiebreakers in round-robin tournaments.

**Algorithm:**

For each OTB game in the player's game list:

1. Look up the opponent's final score from the score map.
2. Apply the result weight:
   - Win: add the opponent's full score.
   - Draw: add half the opponent's score.
   - Loss: add 0.
3. Sum all contributions.

**Example:** A player who beat an opponent with 5.0 points and drew with an opponent who has 4.0 points:

- Win contribution: 5.0
- Draw contribution: 4.0 / 2 = 2.0
- SB = 5.0 + 2.0 = **7.0**

**Formula:** `SUM(win: opponent score, draw: opponent score / 2, loss: 0)`

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Direct Encounter
tb, err := tiebreaker.Get("direct-encounter")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Sonneborn-Berger
tb, err = tiebreaker.Get("sonneborn-berger")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player.
