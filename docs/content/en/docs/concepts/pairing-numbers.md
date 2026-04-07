---
title: "Pairing Numbers and Seeding"
linkTitle: "Pairing Numbers"
weight: 9
description: "Tournament pairing numbers, initial ranking, and how seeding order affects pairings."
---

Every player in a chess tournament is assigned a **pairing number**, formally called the Tournament Pairing Number (TPN). This number serves as the player's identity within the pairing engine and establishes their seeding position relative to all other participants.

## How pairing numbers are assigned

Before the first round, all players are sorted by rating (highest first). Players with the same rating are ordered alphabetically by name. The sorted position becomes the player's **initial rank**: position 1 is the highest-rated player, position 2 the second-highest, and so on.

At the start of each round, active players are re-ranked by current score (descending), with ties broken by initial rank (ascending). This re-ranking produces the TPN for that round. A player who started as initial rank 5 but leads the tournament after three rounds could have TPN 1 in round 4.

The key distinction:

- **Initial rank** is fixed for the entire tournament. It reflects the pre-tournament rating order.
- **TPN** is recalculated every round. It reflects the current standing order.

## Where pairing numbers matter

The TPN influences nearly every aspect of the pairing process:

### Score group splitting

In the [Dutch system](/docs/pairing-systems/dutch/), each score group is divided into two halves -- S1 (the top-seeded half) and S2 (the bottom half). The split is determined by TPN order: the top half of the group by TPN forms S1, and the rest form S2. S1 players are then paired against S2 players. This ensures that the highest-ranked players within a score group face opponents from the lower half, creating balanced matchups.

### Board ordering

After pairings are generated, games are assigned to boards in a specific order. The primary sort is by the higher scorer in each pairing (top boards feature the highest-scoring players). Within the same score level, the pairing with the lower minimum TPN is placed on the higher board. This means the game featuring the tournament leader appears on board 1.

### Bye assignment

When selecting which player receives the [pairing-allocated bye](/docs/concepts/byes/) (PAB), most systems prefer the player with the highest TPN (lowest ranking) in the lowest score group. The highest TPN typically belongs to the lowest-rated player, making them the natural bye candidate.

### Color allocation

When two players have no color history (typically in round 1), colors are assigned by board number: on odd-numbered boards the higher-seeded player (lower TPN) gets White, and on even-numbered boards they get Black. This alternating pattern ensures a balanced color distribution across the first round.

When both players have color preferences of equal strength, the higher-ranked player (lower TPN) gets their preferred color.

### Floater selection

The TPN affects which player [floats](/docs/concepts/floaters/) when a score group cannot be paired internally. In the Lim system, downfloaters are selected starting from the lowest TPN (strongest player), while upfloaters are selected starting from the highest TPN (weakest player). This design principle keeps the strongest players in their natural score group when possible.

### Tiebreaking

The "pairing number" tiebreaker uses the TPN directly as a tiebreak value. Since lower TPNs correspond to higher-rated players, this tiebreaker favors the higher-rated player when all other tiebreakers are equal.

## Round-robin: Varma tables

In [round-robin tournaments](/docs/concepts/round-robin/), pairing numbers take on special significance because they directly determine the game schedule through Berger tables. The order in which players are numbered controls who plays whom in each round.

When players come from multiple federations, it is desirable to avoid same-federation matchups in the early rounds. The [Varma tables](/docs/algorithms/varma-tables/) (defined in FIDE C.05 Annex 2) provide a federation-aware number assignment scheme. Players are distributed across four groups (A through D), with the largest federations spread across groups first, ensuring that players from the same country are placed at positions in the Berger table where they meet in later rounds rather than early ones.

The Varma assignment algorithm:

1. Groups players by federation, sorted by federation size (largest first).
2. Assigns each federation's players to Varma groups, picking the group with the most available slots.
3. If a federation is too large for any single group, spills players across multiple groups.

This supports tournaments with up to 24 players and works with the standard Berger table rotation.

## The PlayerEntry.ID field

In the chesspairing data model, each player has an `ID` field that serves as their unique identifier throughout the tournament. This ID is used in pairing results, game records, and bye entries. The TPN is computed from the ID-indexed player data each round -- it is not stored as a permanent attribute but derived from the current score and initial rank.

## See also

- [Varma tables algorithm](/docs/algorithms/varma-tables/) -- federation-aware number assignment for round-robin
- [Pairing systems](/docs/pairing-systems/) -- how different systems use seeding order
