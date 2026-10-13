---
title: "Head-to-Head Tiebreakers"
linkTitle: "Head-to-Head"
weight: 4
description: "Direct Encounter and Sonneborn-Berger — tiebreakers that consider results against specific opponents."
---

Head-to-head tiebreakers resolve ties by examining results between specific opponents rather than aggregating across the full field. Direct Encounter looks only at games among tied players, while Sonneborn-Berger weights each result by the opponent's final score.

Direct Encounter is **FIDE Type A** (tied participants' games); Sonneborn-Berger is **Type BC** (own and opponents' results).

## Forfeit handling

Both tiebreakers use per-round opponent records. Swiss unplayed rounds follow the Article 16 adjusted-score and dummy model; in predetermined pairings, Article 15.2 treats forfeits as regular encounters. Pending games are not completed encounters.

## Tiebreakers

### direct-encounter

**ID:** `direct-encounter`
**Name:** Direct Encounter
**FIDE Category:** A

Direct Encounter considers only encounters between members of the subgroup still tied after preceding tie-breaks. Players who are not tied with anyone receive a value of 0.

**Algorithm:**

1. Start with each score group, then apply each configured tie-break only to subgroups still tied.
2. For Direct Encounter, sum the awarded points against opponents within that subgroup. For repeated encounters with one opponent, use their average result as one addend (a Double Swiss match is already one encounter).
3. If Direct Encounter separates players but leaves a tied subset, repeat it on that subset until no further separation occurs. A singleton receives 0 if never evaluated.

In a Swiss tournament with missing encounters, only a player guaranteed to lead the separate standings regardless of those games is ranked first; repeat for the remaining players, leaving unresolved players tied.

**Special considerations:**

- Games against players outside the tied group are completely ignored.
- If two tied players never played each other, their value reflects only encounters against other members of the subgroup.
- In Swiss tournaments with many players on the same score, this tiebreaker can be decisive when specific head-to-head matchups occurred.

**Formula:** `SUM(average awarded points per opponent within the tied subgroup)`

### sonneborn-berger

**ID:** `sonneborn-berger`
**Name:** Sonneborn-Berger
**FIDE Category:** BC

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
