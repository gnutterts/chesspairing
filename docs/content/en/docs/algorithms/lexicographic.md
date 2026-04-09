---
title: "Lexicographic Pairing"
linkTitle: "Lexicographic"
weight: 13
description: "DFS backtracking for the lexicographically smallest valid pairing — used by Double-Swiss and Team Swiss."
---

## Overview

The Double-Swiss (FIDE C.04.5) and Team Swiss (FIDE C.04.6) pairing systems
use a shared algorithm for pairing within score groups: find the
**lexicographically smallest** valid pairing by depth-first search with
backtracking.

"Lexicographically smallest" means: among all valid pairings, choose the one
where the first pair (by pairing number order) is the smallest possible, then
the second pair is the smallest possible given the first, and so on. This
provides a deterministic, reproducible pairing that favors matching
lower-numbered players (higher-seeded) first.

The implementation lives in `pairing/lexswiss/bracket.go`.

---

## Definitions

Given $n$ participants in a score group, sorted by tournament pairing number
(TPN) in ascending order: $p_1, p_2, \ldots, p_n$.

A **valid pairing** is a set of pairs $\{(p_{a_1}, p_{b_1}), (p_{a_2},
p_{b_2}), \ldots\}$ where:

1. Each participant appears in at most one pair.
2. No pair violates the absolute criteria:
   - **C1**: The two players have not already played each other (excluding
     forfeits).
   - **Forbidden pairs**: The pair is not in the forbidden list.
   - **System criteria**: The pair passes the system-specific criteria
     function.
3. If $n$ is odd, exactly one participant is left unpaired (they will float
   to the next bracket).

A **lexicographic ordering** on pairings: pairing $A$ is lexicographically
smaller than pairing $B$ if, at the first position where they differ, $A$'s
participant has a smaller TPN.

---

## The Criteria Function

The algorithm accepts a **criteria function** that encodes system-specific
quality requirements beyond the basic C1/forbidden checks:

```go
type CriteriaFunc func(pairs []Pair, remaining []Participant) bool
```

It receives the pairs formed so far and the remaining unmatched participants,
returning `true` if the current partial pairing is acceptable.

| System                | Criteria checked                                                                                        |
| --------------------- | ------------------------------------------------------------------------------------------------------- |
| Double-Swiss (C.04.5) | C8: Minimize the number of upfloaters                                                                   |
| Team Swiss (C.04.6)   | C8: Minimize upfloaters, C9: Minimize score-difference of paired teams, C10: Minimize upfloaters' score |

The criteria function is called at each node of the DFS tree, enabling
early pruning of branches that cannot satisfy the quality requirements.

---

## Algorithm: pairRecursive

The core is a recursive DFS that builds pairings one pair at a time:

```text
function pairRecursive(participants, forbidden, criteriaFn, pairs):
    if no unpaired participants remain:
        return pairs                  // Complete valid pairing found

    first ← smallest-TPN unpaired participant

    for each candidate in remaining participants (ascending TPN):
        if first == candidate:
            continue
        if alreadyPlayed(first, candidate):
            continue                  // C1 violation
        if isForbidden(first, candidate):
            continue

        newPairs ← pairs + (first, candidate)
        remaining ← participants - {first, candidate}

        if criteriaFn(newPairs, remaining) == false:
            continue                  // System criteria violated

        result ← pairRecursive(remaining, forbidden, criteriaFn, newPairs)
        if result != nil:
            return result             // Success — propagate up

    // If n is odd and first is the last unpaired, allow leaving them unpaired
    if only one participant remains:
        return pairs                  // first floats

    return nil                        // Backtrack — no valid partner for first
```

### Key Properties

1. **First unused, smallest TPN.** At each recursion level, the algorithm
   picks the unpaired participant with the smallest TPN. This ensures the
   lexicographic property: the first pair is determined by the lowest-TPN
   player's best available partner.

2. **Partner search in TPN order.** Candidates are tried in ascending TPN
   order. The first valid partner found produces the lexicographically
   smallest pairing for this position.

3. **Backtracking.** If no valid partner exists for the current player, the
   algorithm backtracks to the previous level and tries the next candidate
   there. This handles situations where a locally valid choice creates a
   dead end deeper in the tree.

4. **Early termination.** The criteria function enables pruning. If a partial
   pairing already violates quality criteria, the branch is abandoned without
   exploring its children.

---

## Greedy Fallback

If the DFS finds no complete valid pairing (all branches are pruned by the
criteria function), the algorithm falls back to a greedy partial pairing:

```text
function greedyPartialPair(participants, forbidden):
    pairs ← []
    for each unpaired participant p (ascending TPN):
        for each unpaired candidate c (ascending TPN, c ≠ p):
            if not alreadyPlayed(p, c) and not isForbidden(p, c):
                pairs ← pairs + (p, c)
                mark p and c as paired
                break
    return pairs
```

The greedy fallback does not check the criteria function -- it only enforces
C1 and forbidden pairs. It may leave some participants unpaired (as
floaters). This fallback ensures the algorithm always produces _some_ pairing
even when the criteria function is too restrictive.

---

## Complexity

### Worst Case

The DFS explores a search tree of depth $n/2$ (one level per pair) with
branching factor up to $n - 1$ at the first level, $n - 3$ at the second,
and so on:

$$\text{nodes} \leq \prod_{k=0}^{n/2 - 1} (n - 2k - 1) = (n-1)!! \quad \text{(double factorial)}$$

For $n = 20$, this is $19!! = 654{,}729{,}075$ -- too large for brute force.
However, the criteria function prunes aggressively, and the early-termination
property means the first valid pairing is found without exploring the full
tree.

### Practical Performance

In practice, the DFS terminates quickly because:

- **Most pairs are compatible.** In a typical score group, few players have
  already played each other, so the first candidate tried is usually valid.
- **Criteria pruning.** The criteria function eliminates invalid branches
  early.
- **Small score groups.** Score groups in Swiss tournaments rarely exceed
  20--30 players (and are often much smaller), keeping the search space
  manageable.

For typical tournament sizes, the DFS completes in microseconds.

---

## Example

Score group with 6 participants: TPN 3, 7, 12, 15, 22, 28.

Previous games: 3 has played 7; 12 has played 15.

**DFS execution:**

1. Pick TPN 3 (smallest). Try TPN 7 -- already played (C1). Try TPN 12 -- valid.
   Pair (3, 12).
2. Pick TPN 7 (next smallest). Try TPN 15 -- valid. Pair (7, 15).
3. Pick TPN 22 (next smallest). Try TPN 28 -- valid. Pair (22, 28).
4. No unpaired participants remain. Return {(3, 12), (7, 15), (22, 28)}.

This is the lexicographically smallest valid pairing. If TPN 12 had also
been incompatible with TPN 3, the DFS would try TPN 15 as 3's partner
instead, producing a different pairing.

---

## Contrast with Blossom-Based Systems

| Property          | Lexicographic DFS                        | Blossom matching         |
| ----------------- | ---------------------------------------- | ------------------------ |
| Used by           | Double-Swiss, Team Swiss                 | Dutch, Burstein          |
| Optimality        | Lexicographically first                  | Maximum weight           |
| Criteria          | Checked per-pair, with backtracking      | Encoded in edge weights  |
| Complexity        | Exponential worst case, fast in practice | $O(n^3)$ guaranteed      |
| Score group scope | One group at a time                      | Global across all groups |

The Blossom approach is theoretically more powerful: it finds the
globally optimal matching across all criteria simultaneously. The
lexicographic approach is simpler, deterministic, and well-suited to the
Double-Swiss and Team Swiss regulations where the criteria are fewer and
the "first valid pairing" definition is explicit in the rules.

---

## Related Pages

- [Double-Swiss Pairing](/docs/pairing-systems/double-swiss/) -- uses
  lexicographic pairing with C8 criteria.
- [Team Swiss Pairing](/docs/pairing-systems/team/) -- uses
  lexicographic pairing with C8--C10 criteria.
- [Dutch Criteria](../dutch-criteria/) -- the Blossom-based alternative
  with 21 optimization criteria.
