---
title: "Dutch Criteria"
linkTitle: "Dutch Criteria"
weight: 10
description: "The 21 criteria (C1-C21) governing Dutch Swiss pairings — absolute constraints and optimization targets."
---

## Overview

The Dutch Swiss system (FIDE C.04.3) defines 21 criteria numbered $C_1$
through $C_{21}$ that govern how players are paired. These criteria form a
strict priority hierarchy: $C_i$ takes absolute precedence over $C_j$ when
$i < j$.

The criteria divide into two categories:

- **Absolute criteria** ($C_1$--$C_4$): constraints that _must_ be satisfied.
  A pairing that violates any absolute criterion is rejected entirely.
- **Optimization criteria** ($C_5$--$C_{21}$): targets that are _maximized_
  through [edge weight encoding](../edge-weights/). Violations are penalized
  in the Blossom matching weight, with higher-priority criteria occupying
  more-significant bits.

The absolute criteria are implemented in `pairing/swisslib/criteria.go`. The
optimization criteria are encoded in `pairing/swisslib/criteria_pairs.go`
(see [Edge Weight Encoding](../edge-weights/)).

---

## Absolute Criteria

### C1: No Rematches

Two players who have already played each other shall not be paired again.

$$\text{C1}(i, j) = \begin{cases} \text{pass} & \text{if } (i, j) \notin H \\ \text{fail} & \text{otherwise} \end{cases}$$

where $H$ is the set of played pairings. Forfeits are excluded from the
history: a game that was forfeited does not count as "played" for C1
purposes, meaning the players _can_ be re-paired. Double forfeits are also
excluded.

Implementation: `C1NoRematches` in `pairing/swisslib/criteria.go`.

### C2: No Second PAB

A player who has already received a pairing-allocated bye (PAB) shall not
receive a second one.

This criterion is enforced during bye selection, not during pair
construction. The [completability](../completability/) matching and the
bye selectors ensure that only PAB-eligible players can receive the bye.

Implementation: `C2NoSecondPAB` in `pairing/swisslib/criteria.go`.

### C3: No Absolute Color Conflict

Two non-top-scorers who both have an absolute color preference for the same
color shall not be paired.

A player has an **absolute color preference** when:

- Their color imbalance exceeds 1 (e.g., 3 Whites vs 1 Black), or
- They have played 2 or more consecutive games with the same color.

If both players have an absolute preference for (say) White, pairing them
would force one to play Black despite the absolute preference -- violating
the color rules.

$$\text{C3}(i, j) = \begin{cases} \text{fail} & \text{if both have absolute preference for the same color} \\ & \text{and neither is a top scorer} \\ \text{pass} & \text{otherwise} \end{cases}$$

**Top-scorer exception.** In the final round, when both players are top
scorers (in the highest non-empty score group), C3 is relaxed to allow the
pairing. This prevents situations where the tournament leaders cannot be
paired due to color constraints.

Implementation: `C3AbsoluteColorConflict` in `pairing/swisslib/criteria.go`.

### C4: Bracket Completeness

Not a per-pair criterion but a structural requirement: after processing a
score group, all players must either be paired or have floated to an
adjacent group. No player may be left "stranded" without a pair or a float
destination.

Implementation: validated during the bracket loop in
`pairing/swisslib/global_matching.go`.

---

## Optimization Criteria

The optimization criteria are encoded as bit fields in the Blossom edge
weight. See [Edge Weight Encoding](../edge-weights/) for the complete bit
layout. Here we describe the semantic meaning of each criterion.

### C5: Maximize Pairs in Current Bracket

Within each score group, maximize the number of players paired with opponents
from the _same_ score group (as opposed to floaters from adjacent groups).

**Edge weight field.** 1 bit (field 2, width $\text{sgBits}$): set when both
players belong to the current score group.

### C6: Maximize Score Sum in Current Bracket

Among pairs within the current score group, maximize the sum of scores. This
prefers pairing higher-scoring players within the bracket over lower-scoring
ones.

**Edge weight field.** Score-indexed sub-fields (field 3, width
$\text{sgsShift}$).

### C7: Maximize Pairs in Next Bracket

When players must float down to the next score group, maximize the number of
such downfloat pairings. This ensures the bracket extends smoothly into the
next group.

**Edge weight field.** 1 bit (field 4, width $\text{sgBits}$): set when the
lower player is in the next score group.

### C8: Maximize Score Sum in Next Bracket

Analogous to C6 but for the next-bracket extension.

**Edge weight field.** Score-indexed sub-fields (field 5, width
$\text{sgsShift}$).

### C9: Minimize Bye Recipient's Unplayed Games

When a bye must be assigned, prefer giving it to the player with the fewest
unplayed games (among bye-eligible candidates). This is encoded as two
sub-fields for the lower and higher player in each pair.

**Edge weight fields.** Fields 6--7, each $\text{sgBits}$ wide.

### C10--C13: Color Criteria

Four criteria govern color preference compatibility, listed in decreasing
priority:

| Criterion | Meaning                                                                                                                  |
| --------- | ------------------------------------------------------------------------------------------------------------------------ |
| C10       | No absolute imbalance conflict: avoid pairing two players who both have color imbalance $> 1$ and prefer the same color. |
| C11       | No absolute preference conflict: a more nuanced check considering imbalance magnitude and consecutive-color history.     |
| C12       | Color preferences compatible: the two players prefer different colors, or at least one has no preference.                |
| C13       | No strong preference conflict: avoid pairing two players with strong (but not absolute) same-color preferences.          |

**Edge weight fields.** Fields 8--11, each $\text{sgBits}$ wide.

See [Color Allocation](../color-allocation/) for how these preferences are
computed and resolved.

### C14--C15: Float Repeat Avoidance (Round $R-1$)

These criteria discourage repeating float patterns from the immediately
preceding round:

| Criterion | Meaning                                                                                                                |
| --------- | ---------------------------------------------------------------------------------------------------------------------- |
| C14       | Minimize the number of players who floated down in round $R - 1$ and are floating down again.                          |
| C15       | Avoid pairing an upfloater from round $R - 1$ against a higher-scoring opponent (which would make them upfloat again). |

**Edge weight fields.** Fields 12--13, each $\text{sgBits}$ wide.
Conditional: only present when at least 1 round has been played.

### C16--C17: Float Repeat Avoidance (Round $R-2$)

Same as C14--C15 but for two rounds ago:

| Criterion | Meaning                                                              |
| --------- | -------------------------------------------------------------------- |
| C16       | Minimize repeated downfloaters from round $R - 2$.                   |
| C17       | Avoid repeated upfloater-against-higher-opponent from round $R - 2$. |

**Edge weight fields.** Fields 16--17, each $\text{sgBits}$ wide.
Conditional: only present when at least 2 rounds have been played.

### C18--C19: Float Score Minimization (Round $R-1$)

| Criterion | Meaning                                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------------------------- |
| C18       | Minimize the score of downfloaters from round $R - 1$. A player who floated down with a high score should not float down again. |
| C19       | Minimize the opponent score of upfloaters from round $R - 1$. An upfloater should face the lowest-scoring available opponent.   |

**Edge weight fields.** Fields 14--15, each $\text{sgsShift}$ wide.
Score-indexed sub-fields provide granular optimization.

### C20--C21: Float Score Minimization (Round $R-2$)

Same as C18--C19 but for two rounds ago:

| Criterion | Meaning                                                       |
| --------- | ------------------------------------------------------------- |
| C20       | Minimize the score of downfloaters from round $R - 2$.        |
| C21       | Minimize the opponent score of upfloaters from round $R - 2$. |

**Edge weight fields.** Fields 18--19, each $\text{sgsShift}$ wide.
Conditional: only present when at least 2 rounds have been played.

---

## Float History

Several optimization criteria reference a player's **float history** --
whether they floated up or down in previous rounds. The float direction is
determined by comparing a player's score to the score group they were
paired in:

- **Float down**: player's score is higher than the bracket they were
  paired in (they "dropped" to find an opponent).
- **Float up**: player's score is lower than the bracket they were paired
  in (they were "pulled up" to fill a bracket).
- **No float**: player was paired within their own score group.

The helper `floatAtRound(p, roundIdx)` in
`pairing/swisslib/criteria_optimization.go` retrieves the float direction
for a specific past round.

---

## Criteria Interaction with Edge Weights

The 21 criteria are not applied sequentially. Instead, the optimization
criteria are encoded _simultaneously_ into each edge weight via the bit
layout described in [Edge Weight Encoding](../edge-weights/). The Blossom
algorithm then finds the maximum-weight matching, which automatically
resolves all criteria in the correct priority order.

This is the key insight of the bbpPairings architecture (which this
implementation follows): instead of processing criteria one at a time with
backtracking, encode them all into a single number and let the matching
algorithm handle the optimization.

The absolute criteria ($C_1$, $C_3$, forbidden pairs) are handled
differently: they determine which edges _exist_ in the graph. A pair that
violates an absolute criterion simply has no edge, so Blossom cannot select
it.

---

## Criteria Summary Table

| Criterion | Type         | Description                             | Edge weight field |
| --------- | ------------ | --------------------------------------- | ----------------- |
| C1        | Absolute     | No rematches                            | Edge existence    |
| C2        | Absolute     | No second PAB                           | Bye selection     |
| C3        | Absolute     | No absolute color conflict              | Edge existence    |
| C4        | Structural   | Bracket completeness                    | Bracket loop      |
| C5        | Optimization | Maximize within-bracket pairs           | Field 2           |
| C6        | Optimization | Maximize within-bracket score sum       | Field 3           |
| C7        | Optimization | Maximize next-bracket pairs             | Field 4           |
| C8        | Optimization | Maximize next-bracket score sum         | Field 5           |
| C9        | Optimization | Minimize bye recipient unplayed games   | Fields 6--7       |
| C10       | Optimization | No absolute imbalance conflict          | Field 8           |
| C11       | Optimization | No absolute preference conflict         | Field 9           |
| C12       | Optimization | Color preferences compatible            | Field 10          |
| C13       | Optimization | No strong preference conflict           | Field 11          |
| C14       | Optimization | No repeated downfloat ($R-1$)           | Field 12          |
| C15       | Optimization | No repeated upfloat ($R-1$)             | Field 13          |
| C16       | Optimization | No repeated downfloat ($R-2$)           | Field 16          |
| C17       | Optimization | No repeated upfloat ($R-2$)             | Field 17          |
| C18       | Optimization | Minimize downfloat score ($R-1$)        | Field 14          |
| C19       | Optimization | Minimize upfloat opponent score ($R-1$) | Field 15          |
| C20       | Optimization | Minimize downfloat score ($R-2$)        | Field 18          |
| C21       | Optimization | Minimize upfloat opponent score ($R-2$) | Field 19          |

---

## Related Pages

- [Edge Weight Encoding](../edge-weights/) -- how these criteria become
  Blossom edge weights.
- [Completability](../completability/) -- how the bye recipient (C2/C9) is
  determined.
- [Color Allocation](../color-allocation/) -- how color preferences (C10--C13)
  are resolved after pairing.
- [Baku Acceleration](../baku-acceleration/) -- how virtual points modify
  the score groups that C5--C8 operate on.
- [Dutch Pairing](/docs/pairing-systems/dutch/) -- the pairing system these
  criteria govern.
