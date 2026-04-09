---
title: "Edge Weight Encoding"
linkTitle: "Edge Weights"
weight: 2
description: "How 16+ pairing criteria are encoded into a single big.Int edge weight for Blossom matching."
---

## The Problem

Swiss pairing systems define a strict priority order over many criteria. The
Dutch system (FIDE C.04.3), for example, has 21 optimization criteria numbered
$C_1$ through $C_{21}$. Criterion $C_i$ takes absolute priority over $C_j$
whenever $i < j$ -- no amount of improvement in lower-priority criteria can
compensate for a single violation of a higher-priority criterion.

The [Blossom algorithm](../blossom/) maximizes a single numeric weight per
edge. We need a way to encode all criteria into one number such that Blossom's
weight maximization automatically satisfies criteria in the correct priority
order.

---

## The Solution: Bit-Field Encoding

Each criterion occupies a contiguous range of bits in a `*big.Int` value.
Higher-priority criteria occupy more-significant bits. Because the value of a
single bit at position $p$ exceeds the combined value of all bits below it:

$$2^p > \sum_{k=0}^{p-1} 2^k = 2^p - 1$$

a single violation in a high-priority criterion (clearing its bit) reduces the
weight by more than the maximum possible contribution of all lower-priority
criteria combined. Blossom, maximizing total weight, will therefore always
resolve higher-priority criteria first.

**Positive logic convention:** a set bit (1) means "no violation." A higher
weight means a better pairing. This matches the implementation in
`pairing/swisslib/criteria_pairs.go`.

---

## Bit Width Parameters

Three derived values control the layout:

**$\text{sgBits}$** -- Score group size bits.

$$\text{sgBits} = \lceil \log_2(\max_i |\text{SG}_i|) \rceil$$

where $|\text{SG}_i|$ is the number of players in score group $i$. This is the
width of each boolean-style field (the field can hold a count up to the
largest score group size). Computed by `bitsToRepresent(maxScoreGroupSize)`.

**$\text{sgsShift}$** -- Score-groups shift (cumulative bit width).

$$\text{sgsShift} = \sum_{i} \text{bitsToRepresent}(|\text{SG}_i|)$$

Each score group gets a sub-field whose width depends on its own size. The
total width of a score-indexed field is $\text{sgsShift}$. Within such a field,
score group $i$'s sub-field starts at offset:

$$\text{sgShifts}[\text{score}_i] = \sum_{j < i} \text{bitsToRepresent}(|\text{SG}_j|)$$

where groups are ordered lowest score first (matching bbpPairings' low-to-high
iteration). This is stored in `EdgeWeightParams.ScoreGroupShifts`.

**$\text{reserveBits}$** -- Reserve for the per-pair addend.

$$\text{reserveBits} = 3 \cdot \text{sgBits} + 1$$

---

## Bit Layout (High to Low)

The complete layout from most-significant to least-significant bits. Fields
are listed top-down in order of decreasing significance. "Width" is in bits.

| #   | Field                          | Width                       | Description                                                                                                                                                    |
| --- | ------------------------------ | --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Bye eligibility                | 2                           | Value $1 + [\text{not bye candidate}_i] + [\text{not bye candidate}_j]$. Prefers matching non-bye-eligible players together, leaving bye candidates unmatched. |
| 2   | Pairs in current bracket       | $\text{sgBits}$             | 1 if both players are in the current score group, 0 otherwise. Maximizes the number of within-bracket pairings (C5).                                           |
| 3   | Scores in current bracket      | $\text{sgsShift}$           | Sets a bit at the sub-field position corresponding to the higher player's score. Maximizes the sum of matched scores within the bracket (C6).                  |
| 4   | Pairs in next bracket          | $\text{sgBits}$             | 1 if the lower player is in the next score group. Extends the bracket downward (C7).                                                                           |
| 5   | Scores in next bracket         | $\text{sgsShift}$           | Analogous to field 3, but for the next-bracket extension.                                                                                                      |
| 6   | Bye assignee unplayed (lower)  | $\text{sgBits}$             | Unplayed game count of the lower player if they are a bye candidate. C9: minimize unplayed games of the eventual bye recipient.                                |
| 7   | Bye assignee unplayed (higher) | $\text{sgBits}$             | Same for the higher player.                                                                                                                                    |
| 8   | Color: absolute imbalance      | $\text{sgBits}$             | C10: 1 unless both players have absolute color imbalance ($> 1$) and prefer the same color.                                                                    |
| 9   | Color: absolute preference     | $\text{sgBits}$             | C11: complex check involving imbalance magnitude, repeated-color history, and preference direction.                                                            |
| 10  | Color: preference compatible   | $\text{sgBits}$             | C12: 1 if color preferences are compatible (different preferred colors, or at least one has no preference).                                                    |
| 11  | Color: strong preference       | $\text{sgBits}$             | C13: 1 unless both have strong preferences for the same color with no absolute override.                                                                       |
| 12  | C14: downfloat repeat R-1      | $\text{sgBits}$             | Count (0--2) of players who floated down in the previous round. Higher = fewer violations.                                                                     |
| 13  | C15: upfloat repeat R-1        | $\text{sgBits}$             | 1 unless the lower player was an upfloater last round and is paired against a higher-scoring opponent.                                                         |
| 14  | C18: downfloat score R-1       | $\text{sgsShift}$           | Score-indexed addend for each player who floated down last round. Minimizes the score of downfloaters.                                                         |
| 15  | C19: upfloat opp. score R-1    | $\text{sgsShift}$           | Score-indexed bit for the higher player's score, set when the lower player was not an upfloater last round. Minimizes the opponent score of upfloaters.        |
| 16  | C16: downfloat repeat R-2      | $\text{sgBits}$             | Same as C14 but for two rounds ago. Conditional: only present when played rounds $> 1$.                                                                        |
| 17  | C17: upfloat repeat R-2        | $\text{sgBits}$             | Same as C15 but for two rounds ago. Conditional.                                                                                                               |
| 18  | C20: downfloat score R-2       | $\text{sgsShift}$           | Same as C18 but for two rounds ago. Conditional.                                                                                                               |
| 19  | C21: upfloat opp. score R-2    | $\text{sgsShift}$           | Same as C19 but for two rounds ago. Conditional.                                                                                                               |
| 20  | Reserve                        | $3 \cdot \text{sgBits} + 1$ | Reserved for the per-pair addend filled during Phase 3 of bracket processing (S1/S2 split preference and BSN distance).                                        |

Fields 12--19 are conditional on the number of played rounds. Fields 12--15
require at least 1 played round ($R > 0$); fields 16--19 require at least 2
played rounds ($R > 1$). When absent, those bits are simply not allocated,
reducing total width.

---

## Total Bit Width Formula

Let $R$ denote the number of already-played rounds. The total width $W$ is:

$$W = 2 + 2\,\text{sgBits} + 2\,\text{sgsShift} + 2\,\text{sgBits} + 4\,\text{sgBits}$$

plus conditional fields:

$$+ \; [R > 0] \cdot (2\,\text{sgBits} + 2\,\text{sgsShift})$$

$$+ \; [R > 1] \cdot (2\,\text{sgBits} + 2\,\text{sgsShift})$$

plus the reserve:

$$+ \; 3\,\text{sgBits} + 1$$

Collecting terms for the common case $R > 1$:

$$W = 3 + 15\,\text{sgBits} + 6\,\text{sgsShift}$$

---

## Numerical Example

Consider a 100-player Swiss tournament with 9 rounds, currently pairing
round 5 ($R = 4$ played rounds). Suppose the largest score group has 30
players and there are 9 distinct score groups of varying sizes.

- $\text{sgBits} = \lceil \log_2(30) \rceil = 5$
- $\text{sgsShift} = \sum_{i=1}^{9} \text{bitsToRepresent}(|\text{SG}_i|)$

  If score group sizes are roughly 2, 5, 10, 15, 30, 20, 10, 5, 3:

  $= 1 + 3 + 4 + 4 + 5 + 5 + 4 + 3 + 2 = 31$

- Total: $W = 3 + 15(5) + 6(31) = 3 + 75 + 186 = 264$ bits

For larger tournaments or tournaments with more granular score groups (e.g.
draws creating half-point gaps), $\text{sgsShift}$ grows further. Real-world
values can reach 294 bits or more.

This is why `int64` (63 usable bits) is insufficient and the `*big.Int`
variant of the Blossom algorithm is needed.

---

## Reserve Bits (Per-Pair Addend)

The lowest $3 \cdot \text{sgBits} + 1$ bits are reserved for the
**per-pair addend**, filled during Phase 3 of the bracket processing loop in
`PairBracketsGlobal`. This addend encodes within-bracket optimization:

- **S1/S2 split preference.** Within a score group, players are divided into
  an upper half (S1) and a lower half (S2). The addend rewards pairing S1
  players with S2 players over S1-S1 or S2-S2 pairings.

- **BSN distance minimization.** Among S1-S2 pairings, the addend prefers
  matching the first player in S1 with the first in S2, the second with the
  second, and so on. This minimizes the "board seeding number" distance.

The reserve width of $3 \cdot \text{sgBits} + 1$ provides enough room for
these values without overflowing into the C20/C21 fields above.

---

## Color Criteria Encoding

Four bit fields (fields 8--11) encode the color-related optimization criteria.
Each is $\text{sgBits}$ wide and represents a boolean condition on the pair's
color preference compatibility. From highest to lowest priority:

1. **Absolute imbalance** (C10). Set unless both players have color imbalance
   $> 1$ and prefer the same color. A violation means the pairing would force
   one player to increase an already-extreme imbalance.

2. **Absolute preference** (C11). A more nuanced check. When both players
   have absolute color preferences (from imbalance or consecutive same-color
   games) for the same color, the algorithm checks whether the conflict can
   be resolved by inspecting imbalance magnitudes and repeated-color history.

3. **Preference compatible** (C12). Set when the two players' preferred colors
   differ, or when at least one player has no color preference. This is the
   standard "can we assign colors to both players' satisfaction?" check.

4. **Strong preference** (C13). Set unless both players have strong (but not
   absolute) preferences for the same color and neither has an absolute
   preference that would override. This is the weakest color criterion.

The implementation computes these from each player's `ColorHistory` via the
`ComputeColorPreference` function in `pairing/swisslib/color.go`.

---

## Why big.Int?

A quick lower bound: even a modest tournament produces total bit widths that
exceed `int64`:

| Parameter         | Small (20 players) | Medium (60 players) | Large (200 players) |
| ----------------- | ------------------ | ------------------- | ------------------- |
| $\text{sgBits}$   | 3                  | 5                   | 7                   |
| $\text{sgsShift}$ | 12                 | 30                  | 55                  |
| $W$ ($R > 1$)     | 114                | 248                 | 424                 |

The `int64` type provides only 63 usable bits (the sign bit is unavailable
since edge weights must be non-negative). Any tournament with more than
approximately 10 players and 2+ rounds will likely produce edge weights
exceeding 63 bits.

The `MaxWeightMatchingBig` function in `algorithm/blossom/blossom_big.go`
uses `*big.Int` arithmetic throughout, ensuring no precision loss regardless
of tournament size. The performance overhead of `*big.Int` versus `int64` is
approximately 3--5x for typical tournament sizes, which remains negligible
compared to the $O(n^3)$ algorithm cost.

---

## Related Pages

- [Blossom Matching](../blossom/) -- the algorithm that consumes these edge
  weights.
- [Dutch Criteria](../dutch-criteria/) -- the 21 criteria that this encoding
  represents.
