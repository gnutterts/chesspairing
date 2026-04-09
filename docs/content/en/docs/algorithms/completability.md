---
title: "Completability"
linkTitle: "Completability"
weight: 3
description: "Stage 0.5 pre-matching — determining which player receives the bye in odd-count tournaments."
---

## The Bye Problem

When a round has an odd number of active players, exactly one player must
receive a pairing-allocated bye (PAB). The question is: _which one?_

A naive approach -- assigning the bye to the lowest-ranked eligible player --
can lead to situations where the remaining players cannot all be paired. For
example, if removing the lowest-ranked player leaves two players who have
already played each other and have no other compatible opponents, the pairing
fails.

Stage 0.5 solves this by running a simplified [Blossom matching](../blossom/)
over all bye-eligible candidates _before_ the real pairing begins. The
candidate whose removal leaves the best completable matching is chosen as the
bye recipient.

The implementation lives in `pairing/swisslib/global_matching.go`, within the
`PairBracketsGlobal` function.

---

## When Stage 0.5 Runs

Stage 0.5 is triggered only when all three conditions hold:

1. The number of active players is **odd**.
2. At least one player is **eligible** for a bye (has not already received a
   PAB in a prior round, or other system-specific restrictions are met).
3. The pairing system uses the global Blossom architecture (Dutch, Burstein).

For even player counts, Stage 0.5 is skipped entirely and the algorithm
proceeds directly to the bracket loop.

---

## Simplified Edge Weights

The real Blossom matching (Stages 1--2) uses a complex
[edge weight encoding](../edge-weights/) with 20 fields encoding criteria
$C_5$--$C_{21}$. Stage 0.5 uses a much simpler 3-field weight that asks a
single question: _if these two players are paired, which bye candidate is left
unmatched?_

The three fields, from most-significant to least-significant:

| Field                | Width    | Purpose                                                                                                                                                 |
| -------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Top-score protection | 1 bit    | Prefers leaving a lower-scoring player as the bye recipient over a higher-scoring one. Set to 1 when neither endpoint is the top-scoring bye candidate. |
| Score sum            | $s$ bits | Sum of the two players' scores. Maximizing this pushes higher-scoring players into pairings, leaving lower-scoring players for the bye.                 |
| Bye eligibility      | 1 bit    | Set to 1 when neither player is a bye candidate. Prefers matching non-bye-eligible players together, leaving bye candidates free to receive the bye.    |

Here $s$ is the number of bits needed to represent the maximum possible score
sum (twice the maximum individual score).

The total weight width is $s + 2$ bits -- small enough for standard `int64`
arithmetic. No `*big.Int` is needed for Stage 0.5.

---

## Algorithm

The Stage 0.5 procedure:

1. **Build the edge set.** For every pair of players $(i, j)$ that satisfies
   the absolute criteria (C1: no rematches, C3: no absolute color conflicts,
   plus forbidden pairs), create an edge with the simplified 3-field weight.

2. **Run Blossom with maximum cardinality.** Call `MaxWeightMatching` with
   `maxCardinality = true`. This finds a matching that pairs as many players
   as possible, breaking ties by total weight.

3. **Identify the unmatched player.** In a maximum-cardinality matching of an
   odd number of players, exactly one player is left unmatched. This player
   becomes the bye recipient.

4. **Store the result.** The bye recipient's identity is recorded and passed
   to the bracket loop (Stages 1--2), which excludes that player from the
   main matching and assigns them a PAB.

### Why maximum cardinality?

The `maxCardinality = true` flag is essential. Without it, Blossom would
optimize pure weight and might leave multiple players unmatched if doing so
increased total weight. We need exactly one unmatched player -- the bye
recipient -- and all others must be paired.

Among all maximum-cardinality matchings (which all leave exactly one player
unmatched), Blossom selects the one with the highest total weight. The
3-field weight encoding ensures this is the matching that:

1. Protects top-scoring players from receiving the bye (top-score protection
   bit).
2. Among equal protection levels, leaves the lowest-scoring bye candidate
   unmatched (score sum field).
3. Among equal scores, prefers matching non-bye-eligible players together
   (bye eligibility bit).

---

## Correctness Sketch

**Claim.** The player left unmatched by Stage 0.5 is the one whose removal
leaves a completable matching for the remaining players.

**Argument.** The Stage 0.5 Blossom matching considers the same absolute
criteria (C1, C3, forbidden pairs) as the real matching. If player $p$ is
left unmatched, it means a valid matching exists for all other players. The
real matching (Stages 1--2) works with the same player set minus $p$ and adds
optimization criteria (C5--C21) that refine but never invalidate a matching
that satisfies the absolute criteria.

The only way Stage 0.5 could choose a "wrong" bye recipient is if the
simplified weight encoding produced a matching where the unmatched player's
removal left the remaining set unpair-able under the full criteria. This
cannot happen because the full criteria (C5--C21) are optimization targets
encoded in edge weights -- they affect _which_ matching is chosen, not
_whether_ a matching exists. Existence depends only on the absolute criteria,
which Stage 0.5 fully enforces.

---

## Comparison with Other Systems

Not all pairing systems use Stage 0.5:

| System                | Bye selection method                                                                |
| --------------------- | ----------------------------------------------------------------------------------- |
| Dutch (C.04.3)        | Stage 0.5 completability matching                                                   |
| Burstein (C.04.4.2)   | Stage 0.5 completability matching                                                   |
| Dubov (C.04.4.1)      | Dedicated `DubovByeSelector` (Art. 2.3): lowest score group, highest pairing number |
| Lim (C.04.4.3)        | `LimByeSelector` (Art. 1.1): lowest rank in lowest score group                      |
| Double-Swiss (C.04.5) | `AssignPAB` from lexswiss: lowest score, highest TPN                                |
| Team Swiss (C.04.6)   | `AssignPAB` from lexswiss: lowest score, highest TPN                                |
| Keizer                | Lowest Keizer score                                                                 |
| Round-Robin           | Dummy player (no real bye needed)                                                   |

The completability approach (Dutch, Burstein) is the most computationally
expensive but also the most robust: it guarantees by construction that the
remaining players can be paired. The simpler selectors used by other systems
rely on heuristics that work well in practice but do not carry the same
structural guarantee.

---

## Complexity

Stage 0.5 runs one Blossom matching on $n$ players with $O(n^2)$ edges
(all compatible pairs). The Blossom algorithm is $O(n^3)$. Since Stage 0.5
uses `int64` weights (not `*big.Int`), the constant factor is small.

For a 200-player tournament, Stage 0.5 adds roughly 5--10% to the total
pairing time. For typical club tournaments (20--60 players), it is
negligible.

---

## Related Pages

- [Blossom Matching](../blossom/) -- the matching algorithm used by Stage 0.5.
- [Edge Weight Encoding](../edge-weights/) -- the full 20-field encoding used
  in Stages 1--2 (contrast with Stage 0.5's simplified 3-field encoding).
- [Dutch Criteria](../dutch-criteria/) -- the absolute criteria (C1, C3) that
  determine edge eligibility in Stage 0.5.
