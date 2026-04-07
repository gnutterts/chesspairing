---
title: "The Swiss System"
linkTitle: "Swiss System"
weight: 1
description: "How the Swiss pairing system works: matching players of similar score while avoiding repeat pairings."
---

## The problem Swiss solves

A round-robin tournament with 40 players needs 39 rounds. Most events
can afford 7 or 9. The Swiss system was invented to produce a reliable
ranking from far fewer rounds than a full round-robin, by pairing
players of similar strength in every round rather than scheduling every
possible matchup.

The idea dates back to 1895 in Zurich, and FIDE has been refining the
rules ever since. Today the Swiss system is the default format for most
rated chess events worldwide.

## Core principles

Every Swiss system, regardless of variant, follows the same three
principles:

1. **Similar scores play each other.** Players with equal (or close)
   scores are paired together. This concentrates the decisive games near
   the top of the standings as the tournament progresses.

2. **No repeat pairings.** Two players may not meet more than once in
   the same tournament. This forces the pairer to find fresh opponents
   every round.

3. **Color balance.** Each player should alternate between White and
   Black from round to round, and the total number of games as each
   color should stay as close to equal as possible.

These three principles create a constraint-satisfaction problem. The
pairer must find pairings that satisfy all three simultaneously, relaxing
the weaker constraints only when the stronger ones leave no other
option.

## Score groups and brackets

After each round, players are grouped by their current score. A
**score group** is the set of all players on the same point total --
for example, everyone on 3/4 after four rounds.

When a score group has an odd number of players, or when the repeat and
color constraints make it impossible to pair everyone within the group,
one or more players must be paired against someone from an adjacent
group. This creates **brackets** -- working units that may span two
neighbouring score groups. The player who moves between groups is called
a **floater**: an upfloater joins a higher group, a downfloater joins a
lower one.

## How a Swiss round is paired

At a high level, every Swiss pairer follows the same flow:

1. **Rank all active players** by score, then by initial ranking
   (pairing number) within the same score.
2. **Form score groups** from the current standings.
3. **Assign the bye.** If the player count is odd, one player receives a
   pairing-allocated bye (PAB). The bye typically goes to the
   lowest-ranked player in the lowest score group who has not already
   had a bye.
4. **Pair each score group** from the highest score down, subject to the
   absolute criteria: no repeat opponents, respect color constraints
   that cannot be violated, and ensure remaining players can still be
   paired (completability).
5. **Optimise.** Within the space of legal pairings, apply quality
   criteria to choose the best pairing -- for example, minimising
   score differences between opponents, minimising floaters, or
   maximising color preference satisfaction.
6. **Allocate colors.** Once the pairing is determined, decide who
   plays White and who plays Black based on color history.
7. **Order the boards.** Higher-scoring pairings go on the top boards.

The details of steps 3-5 are where the variants differ. Each system
defines its own criteria hierarchy, its own matching strategy, and its
own tiebreaking rules for edge cases.

## Six Swiss variants

chesspairing implements all six FIDE-endorsed Swiss pairing systems,
plus two additional systems:

| System                                              | Description                                                                                                                       |
| --------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| [Dutch](/docs/pairing-systems/dutch/)               | The standard FIDE system (C.04.3). Uses global Blossom matching to optimise across all score groups simultaneously.               |
| [Burstein](/docs/pairing-systems/burstein/)         | FIDE C.04.4.2. Separates the tournament into seeding rounds and post-seeding rounds, using opposition indices to re-rank players. |
| [Dubov](/docs/pairing-systems/dubov/)               | FIDE C.04.4.1. Uses Average Rating of Opponents (ARO) for bracket ordering and has its own set of ten pairing criteria.           |
| [Lim](/docs/pairing-systems/lim/)                   | FIDE C.04.4.3. Processes score groups from the median outward and uses exchange-based matching within each group.                 |
| [Double-Swiss](/docs/pairing-systems/double-swiss/) | FIDE C.04.5. Each round consists of a two-game match. Uses lexicographic bracket pairing.                                         |
| [Team Swiss](/docs/pairing-systems/team/)           | FIDE C.04.6. Swiss pairing for team events, with a 9-step color allocation and first-team concept.                                |

All six share the same `Pairer` interface, accept the same
`TournamentState` input, and return the same `PairingResult` output.
You can swap one for another without changing the rest of your code.

## Matching algorithms

The Dutch and Burstein pairers use Edmonds' maximum weight matching
(Blossom algorithm) to find optimal pairings across all brackets at
once. The Dubov and Lim systems use transposition and exchange-based
matching within individual score groups. The Double-Swiss and Team Swiss
systems use lexicographic bracket pairing.

For more on these algorithms, see the [Algorithms](/docs/algorithms/)
section.
