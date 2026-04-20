---
title: "Scoring Systems"
linkTitle: "Scoring"
weight: 3
description: "How game results become standings — from standard 1-½-0 to Keizer's iterative ranking."
---

## What scoring does

A scoring system converts raw game results into a numeric score per
player, then ranks players by that score to produce standings. Every
tournament needs a scoring system, and the choice of system affects how
the standings look, how draws are valued, and how absences are penalised.

chesspairing implements three scoring systems. All three implement the
same `Scorer` interface, which has two methods:

- **`Score()`** -- takes the full tournament state and returns a ranked
  list of player scores.
- **`PointsForResult()`** -- returns the points a specific result is
  worth in a given context (useful for displaying point values to
  players).

## Standard scoring (1-half-0)

The system used by FIDE for virtually all rated events. Each result
awards a fixed number of points:

| Result                      | Default points |
| --------------------------- | -------------- |
| Win                         | 1.0            |
| Draw                        | 0.5            |
| Loss                        | 0.0            |
| PAB (pairing-allocated bye) | 1.0            |
| Half-point bye              | 0.5            |
| Zero-point bye              | 0.0            |
| Forfeit win                 | 1.0            |
| Forfeit loss                | 0.0            |
| Absent (unexcused)          | 0.0            |
| Excused absence             | 0.0            |
| Club commitment             | 0.0            |

Every one of these values is configurable through the Options struct.
For example, some organisers award 0.5 for a PAB instead of 1.0, or
penalise unexcused absences with a negative score.

Standard scoring is simple and predictable: your points depend only on
your results, not on who you played. This also makes it the scoring
system used internally by Swiss pairers for forming score groups, even
when the tournament's public standings use a different system.

See [Standard scoring reference](/docs/scoring/standard/) for the full
options list.

## Keizer scoring

Keizer is an iterative scoring system popular in club tournaments,
particularly in Belgium and the Netherlands. The central idea: beating a
strong opponent is worth more than beating a weak one.

### How it works

1. **Value numbers.** Each player is assigned a value number based on
   their current ranking. The top-ranked player gets the highest value
   number (by default, equal to the player count); each subsequent rank
   gets one less.

2. **Game points.** When you beat an opponent, you receive their value
   number as points. A draw earns half their value number. A loss earns
   zero (by default, though a "toughness bonus" variant awards a
   fraction for losses too).

3. **Non-game points.** Byes, absences, and club commitments award a
   fraction of your own value number rather than an opponent's. For
   example, a PAB might give you 50% of your own value number.

4. **Self-victory.** By default, each player's own value number is
   added to their total once (not per round). This rewards participation
   and creates separation between active and inactive players.

5. **Iterative convergence.** Here is the key: since value numbers
   depend on rankings, and rankings depend on scores, and scores depend
   on value numbers, the system is circular. Keizer resolves this by
   iterating: compute scores, re-rank, recompute scores with the new
   value numbers, and repeat until the ranking stabilises. In practice,
   convergence happens within a few iterations.

Keizer scoring has many configurable knobs: absence limits, absence
decay, fixed-value overrides for byes, club commitment fractions, and
several variant presets (KeizerForClubs, Classic KNSB, FreeKeizer).

See [Keizer scoring reference](/docs/scoring/keizer/) for the full
options list and variant presets.

## Football scoring (3-1-0)

Football scoring uses the familiar football (soccer) point system:
3 points for a win, 1 for a draw, 0 for a loss. This rewards decisive
results more heavily than standard scoring -- a win is worth three
draws instead of two.

| Result          | Default points |
| --------------- | -------------- |
| Win             | 3.0            |
| Draw            | 1.0            |
| Loss            | 0.0            |
| PAB             | 3.0            |
| Forfeit win     | 3.0            |
| Forfeit loss    | 0.0            |
| Absent          | 0.0            |
| Excused absence | 0.0            |
| Club commitment | 0.0            |

Football scoring is implemented as a thin wrapper around standard
scoring with different defaults. All point values remain configurable.

See [Football scoring reference](/docs/scoring/football/) for details.

## Byes, forfeits, and absences

All three scoring systems handle special result types:

- **Pairing-allocated bye (PAB):** The system awards a bye when the
  player count is odd. Scored generously (a full point in standard, the
  player's own value fraction in Keizer).
- **Half-point bye / zero-point bye:** Requested byes that award
  reduced or no points.
- **Forfeit win:** The winner gets points, but the game is excluded
  from pairing history (the players can be re-paired in a later round).
- **Double forfeit:** Neither player gets points. The game is excluded
  from both scoring and pairing history -- it is treated as if it never
  happened.
- **Absent:** A player who neither played nor received a bye. Typically
  scored as zero, though Keizer awards a configurable fraction.

## Scoring is independent of pairing

A key design decision in chesspairing: **scoring and pairing are
completely independent**. Any scoring system works with any pairing
system. You can run a Swiss tournament with Keizer scoring, or a
round-robin with football scoring. The pairer and scorer never need to
know about each other -- they both operate on the same `TournamentState`
and produce independent outputs.

The one exception is the Keizer _pairer_, which uses the Keizer _scorer_
internally to rank players before pairing. But even there, the
tournament's public standings can use a different scoring system.

## Further reading

- [Standard scoring](/docs/scoring/standard/)
- [Keizer scoring](/docs/scoring/keizer/)
- [Football scoring](/docs/scoring/football/)
