---
title: "Forfeits and Absences"
linkTitle: "Forfeits"
weight: 8
description: "How forfeits and absences affect scoring, pairing history, and tiebreak calculations."
---

Not every game in a chess tournament ends with pieces being moved. Sometimes a player does not show up, or both players fail to appear. These situations produce **forfeit** results that behave very differently from regular game outcomes -- both for scoring and for future pairings.

## Forfeit game results

chesspairing recognizes three forfeit results:

| Result                 | Code   | Meaning                                       |
| ---------------------- | ------ | --------------------------------------------- |
| **Forfeit White wins** | `1-0f` | Black did not show; White is awarded the win. |
| **Forfeit Black wins** | `0-1f` | White did not show; Black is awarded the win. |
| **Double forfeit**     | `0-0f` | Neither player showed up.                     |

These are distinct from the four regular game results (`1-0`, `0-1`, `0.5-0.5`, and `*` for pending).

## The critical difference: pairing history

The most important thing to understand about forfeits is how they affect pairing history.

**Single forfeit (one player wins by forfeit):** The winner receives points (1.0 by default in standard scoring), but the game is **excluded from pairing history**. Because the players never actually sat across the board from each other, the pairing engine treats them as if they have not met. They can be paired again in a later round.

**Double forfeit:** The game is excluded from **both scoring and pairing history**. Neither player receives points, and the game is treated as though it never happened. The two players can be paired again.

This means that forfeit games do not count as "played" for the purpose of the rematch prohibition (the absolute criterion that prevents two players from meeting twice). A player who won by forfeit against an opponent in round 3 could face that same opponent again in round 5.

## Impact on color history

Forfeit games are also excluded from color history. Since no actual game was played, neither player receives a color assignment for that round. This affects:

- **Color preference calculations.** The round does not contribute to the player's white/black balance or consecutive-same-color tracking.
- **Color difference.** The player's color imbalance is computed only from rounds where they actually played.

In the player's color history, a forfeited round is recorded as "no color" (the same as a bye), so it does not influence future color allocation.

## Impact on tiebreakers

Tiebreaker calculations systematically exclude all forfeited games. The `buildOpponentData` function that feeds tiebreaker computations skips any game with a forfeit result (single or double). This means:

- **Buchholz** (all variants) does not count the forfeited opponent's score.
- **Sonneborn-Berger** does not include the forfeited game's result-times-opponent-score product.
- **ARO** (Average Rating of Opponents) only averages over opponents from actual games.
- **Direct Encounter** only considers over-the-board results between the tied players.
- **Performance Rating** and related tiebreakers (PTP, APRO, APPO) exclude forfeited games from their calculations.

Only actual over-the-board games -- where both players showed up and made moves -- contribute to opponent-based tiebreaker values.

## Absence types

Beyond forfeits that occur within a scheduled game, players can also miss a round entirely. chesspairing distinguishes three absence types, each recorded as a [bye](/docs/concepts/byes/):

| Type                | Description                                                                                | Default points |
| ------------------- | ------------------------------------------------------------------------------------------ | -------------- |
| **Absent**          | Unexcused absence. The player did not show up and did not notify the arbiter.              | 0.0            |
| **Excused**         | The player notified the arbiter in advance that they would miss the round.                 | 0.0            |
| **Club Commitment** | The player is absent because they are playing for a club team in an interclub competition. | 0.0            |

Each absence type can be assigned different point values through the scoring system's options. An organizer might choose to give excused absences a partial point while penalizing unexcused absences with zero, for example.

## Checking forfeit status in code

The `GameResult` type provides two methods for identifying forfeits:

- `IsForfeit()` returns `true` for all three forfeit results (`1-0f`, `0-1f`, `0-0f`).
- `IsDoubleForfeit()` returns `true` only for the double forfeit (`0-0f`).

This distinction matters because single forfeits still award points to the winner, while double forfeits award nothing. The `IsForfeit()` check is used throughout the codebase to exclude forfeited games from opponent lists, color histories, and tiebreaker data.

## See also

- [Scoring systems](/docs/scoring/) -- how forfeit wins and losses are scored
- [Tiebreakers](/docs/tiebreakers/) -- which tiebreakers are affected by forfeit exclusion
