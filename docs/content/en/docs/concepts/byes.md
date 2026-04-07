---
title: "Byes"
linkTitle: "Byes"
weight: 7
description: "Pairing-allocated byes, half-point byes, and how odd player counts are handled."
---

A **bye** is a round in which a player does not have an opponent. Byes arise for different reasons -- an odd number of participants, a player requesting time off, or a player simply not showing up -- and each reason carries different point consequences.

## Bye types

chesspairing implements six bye types, each identified by a code used in TRF16 tournament files:

| Bye type                        | TRF code | Default points | Description                                                                       |
| ------------------------------- | -------- | -------------- | --------------------------------------------------------------------------------- |
| **PAB** (Pairing-Allocated Bye) | `F`      | 1.0            | Awarded automatically when there is an odd number of active players.              |
| **Half-point bye**              | `H`      | 0.5            | Requested by the player in advance. The player sits out a round for half a point. |
| **Zero-point bye**              | `Z`      | 0.0            | Requested by the player. No points awarded.                                       |
| **Absent**                      | `U`      | 0.0            | The player did not show up and did not notify the arbiter in advance.             |
| **Excused**                     | --       | 0.0            | The player notified the arbiter beforehand that they would miss the round.        |
| **Club Commitment**             | --       | 0.0            | The player is absent because of interclub team duty.                              |

The point values shown are defaults for [standard scoring](/docs/scoring/). Each scoring system can configure these values differently through its options.

## The Pairing-Allocated Bye (PAB)

The most significant bye type is the PAB. When a tournament has an odd number of active players, one player must sit out each round. The PAB is worth a full point by default, compensating the player for the game they could not play.

A fundamental rule across all pairing systems: **a player should not receive a PAB more than once** in a tournament. The engine filters out players who have already received one before selecting the next PAB recipient.

### How PAB assignment works

Each pairing system uses a different method to decide who receives the PAB:

**Dutch and Burstein** -- These systems use a completability-based approach. Before the main pairing begins, a pre-matching phase (called Stage 0.5) tests which player, when removed from the pool, still allows the remaining players to be completely paired. This ensures the bye goes to a player whose removal does not break the pairing. Among eligible candidates, the player with the lowest score, most games played, and lowest ranking (highest pairing number) is preferred. See [completability](/docs/algorithms/completability/) for details.

**Dubov** -- The bye goes to the lowest-ranked player (highest pairing number) in the lowest score group who has not already received a PAB. Among tied players, the one with the most games played is selected first.

**Lim** -- The bye is assigned to the lowest-ranked player in the lowest score group, provided they have not already received a PAB.

**Double-Swiss and Team Swiss** -- These lexicographic pairing systems assign the bye to the player with the lowest score, breaking ties by lowest ranking (highest pairing number).

**Keizer** -- The lowest-ranked player (by current Keizer score, or by rating before any rounds have been played) receives the bye.

**Round-Robin** -- Odd player counts are handled by adding a virtual "dummy" player to the rotation. Each round, the real player scheduled to face the dummy receives the bye. This rotates naturally through the Berger table, so every player gets exactly one bye across the cycle.

## Bye scoring

How many points a bye is worth depends on the [scoring system](/docs/scoring/) in use:

- **Standard scoring**: PAB = 1.0, Half-point bye = 0.5, all others = 0.0 by default. Each value is configurable through the `pointBye`, `pointAbsent`, and related options.
- **Football scoring**: follows the same defaults as standard scoring but with the football point scale (win = 3, draw = 1, loss = 0).
- **Keizer scoring**: byes are scored using configurable fractions of the player's own value number, with separate settings for PAB, half-point byes, and absences.

## Byes and tiebreakers

Bye rounds are not "real" games. Because no opponent was faced:

- Opponent-based tiebreakers (Buchholz, Sonneborn-Berger, ARO) have no opponent data for that round. The tiebreaker implementations handle this by only summing over actual game entries.
- The bye round itself does not count toward games-played or wins tallies used by certain tiebreakers.

The number of bye rounds a player has received is tracked separately and can affect how virtual opponents are computed in tiebreakers like Fore Buchholz.

## See also

- [Pairing systems overview](/docs/pairing-systems/) -- how each system selects the PAB recipient
- [Scoring systems](/docs/scoring/) -- configuring bye point values
- [Completability algorithm](/docs/algorithms/completability/) -- the Dutch/Burstein method for finding the optimal bye candidate
