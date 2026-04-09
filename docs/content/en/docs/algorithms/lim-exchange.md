---
title: "Lim Exchange Matching"
linkTitle: "Lim Exchange"
weight: 12
description: "Exchange matching with scrutiny order and four floater types in the Lim system."
---

## Overview

The Lim system (FIDE C.04.4.3) pairs each score group using an
**exchange-based algorithm** that processes players in a specific order and
systematically tries alternative partners when the natural pairing fails.
The system also defines four **floater types** (A--D) with distinct
selection rules for choosing which players move between score groups.

The implementation lives in `pairing/lim/`.

---

## Compatibility (Art. 2.1)

Before any pairing or exchange, two players must be **compatible**. The Lim
system defines compatibility more strictly than most other systems:

Two players $a$ and $b$ are compatible if and only if:

1. They have **not already played** each other (excluding forfeits).
2. They are **not a forbidden pair**.
3. At least one valid **color assignment** exists for the pair. Specifically,
   there must exist colors $(c_a, c_b)$ with $c_a \neq c_b$ such that both
   `CanReceiveColor(a, c_a)` and `CanReceiveColor(b, c_b)` return true.

A player **can receive** color $c$ if:

- They have not played 2 consecutive games with color $c$ already (no 3
  consecutive same-color games).
- Receiving $c$ would not create a color imbalance of 3 or more.

$$\text{CanReceiveColor}(p, c) = \begin{cases} \text{false} & \text{if last 2 games were color } c \\ \text{false} & \text{if } |\text{imbalance after } c| \geq 3 \\ \text{true} & \text{otherwise} \end{cases}$$

Implementation: `pairing/lim/compatibility.go`.

---

## Score Group Pairing: S1/S2 Split

Each score group is split into two halves:

- **S1** (upper half): the higher-ranked players (by TPN).
- **S2** (lower half): the lower-ranked players.

If the group has an odd number of players, S1 gets the extra player.

The natural pairing is S1[1] vs S2[1], S1[2] vs S2[2], etc. When this
natural pairing fails due to compatibility constraints, the exchange
algorithm tries alternative partner assignments.

---

## The Exchange Algorithm (Art. 4)

The exchange algorithm processes S1 players in **scrutiny order** (ascending
TPN within S1). For each S1 player, it generates a sequence of candidate
partners and tries them in order.

### Entry Point

```text
function ExchangeMatch(players, pairingDownward, forbidden):
    split players into S1, S2
    result ← tryExchangePairing(S1, S2, forbidden, pairingDownward)
    if result == nil:
        result ← greedyPair(players, forbidden)
    return result
```

### Exchange Pairing

```text
function tryExchangePairing(S1, S2, forbidden, pairingDownward):
    unified ← [S1 | S2]        // S1 first, then S2
    pairs ← []

    for each player p in S1 (scrutiny order):
        if p is already paired:
            continue

        candidates ← generateExchangeOrder(p, unified, pairingDownward)

        for each candidate c in candidates:
            if c is already paired:
                continue
            if not IsCompatible(p, c, forbidden):
                continue
            pairs ← pairs + (p, c)
            break

    // Pair remaining unpaired S2 players among themselves
    pairRemainingS2(pairs)

    return pairs if valid, nil otherwise
```

### Candidate Generation (Art. 4.2)

For an S1 player at position $i$, the `generateExchangeOrder` function
produces candidates in this priority:

1. **Proposed S2 partner**: S2[$i$] (the "natural" partner).
2. **Remaining S2 players**: other S2 players in exchange order (by distance
   from the proposed position).
3. **Cross-half S1 partners**: if no S2 partner is available, try pairing
   with another S1 player. This only happens when S1 is larger than S2 or
   when all S2 partners are incompatible.

The exchange order within S2 follows the FIDE-specified sequence: try the
nearest S2 players first, then progressively more distant ones. The
`pairingDownward` flag affects whether the exchange searches downward (normal)
or upward (when the bracket is processing upfloaters).

---

## Floater Types (Art. 3.9)

When a score group cannot be fully paired, some players must float to an
adjacent group. The Lim system classifies potential floaters into four types
based on their history and compatibility:

| Type | Already floated? | Compatible with adjacent group? | Priority            |
| ---- | ---------------- | ------------------------------- | ------------------- |
| D    | No               | Yes                             | Best (chosen first) |
| C    | No               | No                              |                     |
| B    | Yes              | Yes                             |                     |
| A    | Yes              | No                              | Worst (chosen last) |

The classification considers whether the player has already floated in a
previous round and whether compatible opponents exist in the adjacent score
group.

**Selection preference:** Type D floaters are preferred because they have
not yet floated (minimizing repeat floats) and have compatible opponents in
the target group (ensuring they can actually be paired there). Type A
floaters are a last resort: they have already floated before and lack
compatible opponents.

Implementation: `ClassifyFloater` in `pairing/lim/floater.go`.

---

## Downfloater Selection (Art. 3.2--3.4)

When a player must float down from a score group, the selection algorithm:

1. **Classify** all unpaired players in the score group by floater type.
2. **Prefer type D**, then C, then B, then A.
3. **Within the same type**, apply tiebreakers:
   - **Color equalization**: prefer players whose color balance is closer
     to equal, or whose floating would help equalize colors in the target
     group.
   - **Lowest TPN**: among equal color balance, choose the player with the
     lowest tournament pairing number (highest rank).
4. **Compatibility check**: verify the selected player has at least one
   compatible opponent in the adjacent group. If not, try the next candidate.

### Maxi-Tournament Override

In maxi-tournaments (large open events), an additional constraint applies:
the downfloater's rating must be within 100 points of the highest-rated
player in the target group. This prevents extreme rating mismatches from
downfloating. When enabled via the `MaxiTournament` option, this 100-point
cap overrides the normal floater selection.

Implementation: `SelectDownFloater` in `pairing/lim/floater.go`.

---

## Upfloater Selection (Art. 3.2.4)

Upfloater selection mirrors downfloater selection with one key difference:
instead of preferring the lowest TPN, upfloaters prefer the **highest TPN**
(lowest rank). This ensures the weakest player in the lower group floats
up, preserving competitive balance.

The floater type preference and color equalization tiebreakers are the same.

Implementation: `SelectUpFloater` in `pairing/lim/floater.go`.

---

## Color Allocation (Art. 5)

After pairing, the Lim system allocates colors using a distinct algorithm
with **median-aware tiebreaking**. See [Color Allocation](../color-allocation/)
for the full comparison with other systems. Key features:

- **Round 1**: odd TPN gets the initial color (configurable, default White).
- **Art. 5.3**: a player with 2 consecutive same-color games _must_ get the
  opposite color.
- **Even rounds**: equalize color counts (Art. 5.2/5.6).
- **Odd rounds**: alternate from last played color (Art. 5.5).
- **History tiebreak** (Art. 5.4): walk backwards through game history to
  find the first divergence point. The player above the median gets priority.

The median tiebreak is unique to the Lim system. "Above the median" means
the player's rank is in the upper half of the current score group, reflecting
the philosophy that higher-ranked players should receive slight color
preference advantages.

---

## Greedy Fallback

If the exchange algorithm cannot find a valid pairing (all compatible
partners are exhausted), a greedy fallback pairs players in TPN order:

```text
function greedyPair(players, forbidden):
    for each unpaired player p (ascending TPN):
        for each unpaired candidate c (ascending TPN, c ≠ p):
            if IsCompatible(p, c, forbidden):
                pair (p, c)
                break
    return pairs
```

The greedy fallback may leave unpaired players (who become floaters). It
ensures the algorithm always terminates with _some_ pairing.

---

## Comparison with Other Matching Methods

| Property          | Lim Exchange                     | Dutch Blossom            | Dubov Transposition                | Lexicographic DFS                 |
| ----------------- | -------------------------------- | ------------------------ | ---------------------------------- | --------------------------------- |
| Scope             | One score group                  | Global                   | One score group                    | One score group                   |
| Search strategy   | Sequential + exchange            | Weight maximization      | Bounded transpositions             | Depth-first backtracking          |
| Floater selection | 4-type classification            | Implicit in edge weights | ARO-based                          | Leftover after DFS                |
| Color in pairing  | Part of compatibility            | Part of edge weight      | Part of absolute check             | Post-pairing                      |
| Guarantee         | All compatible paired or floated | Maximum weight matching  | Best within MaxT                   | Lexicographically first           |
| Complexity        | $O(n^2)$ per group               | $O(n^3)$ global          | $O(n \cdot \text{MaxT})$ per group | Exponential worst, fast practical |

---

## Related Pages

- [Lim Pairing](/docs/pairing-systems/lim/) -- the pairing system that uses
  exchange matching.
- [Color Allocation](../color-allocation/) -- the Lim color algorithm with
  median tiebreaking.
- [Dutch Criteria](../dutch-criteria/) -- the Blossom-based alternative.
- [Lexicographic Pairing](../lexicographic/) -- the DFS-based alternative
  used by Double-Swiss and Team Swiss.
