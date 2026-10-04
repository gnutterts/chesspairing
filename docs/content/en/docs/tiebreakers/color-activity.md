---
title: "Color and Activity Tiebreakers"
linkTitle: "Color & Activity"
weight: 5
description: "Black Games, Black Wins, Rounds Played, and Games Played — activity-based tiebreakers."
---

Color and activity tiebreakers measure participation and color distribution rather than opponent quality. They count games played with the Black pieces, wins achieved with Black, total rounds effectively played, and total non-forfeit games. These tiebreakers reward active participation and help differentiate players who overcame the disadvantage of playing Black more often.

Three of the four belong to **FIDE Category B**. Games Played has no FIDE category assignment and is primarily used in Keizer tournaments.

## Tiebreakers

### black-games

**ID:** `black-games`
**Name:** Games with Black
**FIDE Category:** B

Counts the number of games played over the board with the Black pieces. A higher value indicates the player overcame the first-move disadvantage more frequently.

**Algorithm:**

1. Iterate all games in all rounds.
2. For each completed OTB game (`ResultWhiteWins`, `ResultBlackWins`, or `ResultDraw`):
   - Increment the Black player's count.
3. In a round robin, a forfeit win with Black (`ResultForfeitBlackWins`) also counts under FIDE C.07:2026 Article 15.2.
4. Forfeit losses, double forfeits, and pending games are excluded.

This tiebreaker works directly on round data, not through the opponent records.

**Formula:** `COUNT(completed OTB games as Black, plus round-robin forfeit wins with Black)`

### black-wins

**ID:** `black-wins`
**Name:** Black Wins
**FIDE Category:** B

Counts wins achieved with the Black pieces. In Swiss tournaments only over-the-board wins (`ResultBlackWins`) count; in a round robin a forfeit win with Black (`ResultForfeitBlackWins`) also counts under FIDE C.07:2026 Article 15.2.

**Algorithm:**

1. Iterate all games in all rounds.
2. For each game with result `ResultBlackWins`, or `ResultForfeitBlackWins` in a round robin:
   - Increment the Black player's count.
3. Draws and all other results are excluded.

**Formula:** `COUNT(games won with Black)`

### rounds-played

**ID:** `rounds-played`
**Name:** Rounds Played
**FIDE Category:** B

Computes the number of rounds effectively played by subtracting unplayed rounds from the total round count.

**Algorithm:**

Start with `totalRounds`. For each round, determine which rounds count as "unplayed":

**Unplayed (subtracted from total):**

- Forfeit loss (the losing side of `ResultForfeitWhiteWins` or `ResultForfeitBlackWins`)
- Double forfeit (both players in a `ResultDoubleForfeit` game)
- Half-point bye (`ByeHalf`)
- Zero-point bye (`ByeZero`)
- Absent bye (`ByeAbsent`)
- Missing from round entirely (active player not in any game or bye)
- Rounds after a withdrawal (a zero-point bye under FIDE C.07:2026 Article 16.1.1)
- Rounds before a late join

**Played (not subtracted):**

- OTB games (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`)
- Forfeit win (the winning side)
- PAB (`ByePAB`)
- A pending game is not yet an unplayed round, so its round counts.

**Formula:** `totalRounds - COUNT(unplayed rounds)`

### games-played

**ID:** `games-played`
**Name:** Games Played
**FIDE Category:** --

Counts the number of non-forfeit games played. Both players in each qualifying game are counted. This tiebreaker is primarily useful for Keizer tournaments, where players who attended more club evenings and played more games should rank higher than those with the same score but fewer actual games.

**Algorithm:**

1. Iterate all games in all rounds.
2. For each game where `IsForfeit` is false on the `GameData` struct:
   - Increment both White's and Black's count.
3. Byes, absences, and forfeited games are excluded.

Note: this tiebreaker uses the `IsForfeit` field on the `GameData` struct directly, rather than checking the result type.

**Formula:** `COUNT(non-forfeit games the player participated in)`

## Usage

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Games with Black
tb, err := tiebreaker.Get("black-games")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Black Wins
tb, err = tiebreaker.Get("black-wins")
values, err = tb.Compute(ctx, state, scores)

// Rounds Played
tb, err = tiebreaker.Get("rounds-played")
values, err = tb.Compute(ctx, state, scores)

// Games Played
tb, err = tiebreaker.Get("games-played")
values, err = tb.Compute(ctx, state, scores)
```

Each call returns a `[]chesspairing.TieBreakValue` with one entry per player.
