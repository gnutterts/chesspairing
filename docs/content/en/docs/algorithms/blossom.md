---
title: "Blossom Matching"
linkTitle: "Blossom"
weight: 1
description: "Edmonds' maximum weight matching algorithm — O(n^3) for general graphs with blossom contraction."
---

## Problem Statement

Given an undirected weighted graph $G = (V, E)$ with edge weights
$w : E \to \mathbb{Z}$, find a **matching** $M \subseteq E$ -- a set of edges
no two of which share an endpoint -- that maximizes the total weight:

$$\max_{M} \sum_{e \in M} w(e)$$

In the **maximum cardinality** variant the primary objective is to maximize
$|M|$; among all matchings of maximum cardinality, we then maximize total
weight.

The implementation lives in `algorithm/blossom/`.

---

## Why Blossom?

Swiss pairing is _not_ bipartite. In a bipartite matching problem every vertex
belongs to one of two fixed groups and edges only connect vertices from
different groups. In Swiss pairing, any two players who have not already met
can potentially be paired -- the graph is a **general** (non-bipartite) graph.

Classical algorithms such as the Hungarian method or Hopcroft-Karp are
restricted to bipartite graphs. The maximum weight matching problem on general
graphs requires Edmonds' **Blossom algorithm** (1965), which handles odd
cycles through a contraction technique that no bipartite algorithm provides.

---

## LP Relaxation

The matching problem has a clean integer programming formulation. Assign a
binary variable $x_e \in \{0, 1\}$ to each edge $e \in E$:

$$\max \sum_{e \in E} w(e) \, x_e$$

subject to:

$$\sum_{e \ni v} x_e \leq 1 \quad \text{for every } v \in V$$

For bipartite graphs this LP relaxation has integral optima (by total
unimodularity). For general graphs it does not -- fractional half-integral
solutions can appear on odd cycles. The fix is the **odd set constraints**
(Edmonds, 1965):

$$\sum_{e \subseteq B} x_e \leq \frac{|B| - 1}{2} \quad \text{for every odd subset } B \subseteq V, \; |B| \geq 3$$

where $e \subseteq B$ means both endpoints of $e$ lie in $B$. Adding these
constraints (one for each odd subset) restores integrality. The Blossom
algorithm implicitly enforces them through dual variables on contracted
blossoms.

---

## Dual Variables

The LP dual associates two kinds of variables with the primal problem:

- **Vertex duals** $u_v$ for each $v \in V$.
- **Blossom duals** $z_B \geq 0$ for each non-trivial blossom $B$ (odd
  subset with $|B| \geq 3$).

The **complementary slackness** condition for an edge $(i, j)$ is:

$$\pi(i, j) = u_i + u_j + \sum_{\substack{B \ni i \\ B \ni j}} z_B - w(i, j) \geq 0$$

An edge is **tight** when $\pi(i, j) = 0$. A primal-dual pair $(M, u, z)$ is
optimal when:

1. Every matched edge is tight.
2. Every blossom $B$ with $z_B > 0$ is "full" (matched on $\frac{|B|-1}{2}$
   edges).

### Storage convention

The implementation stores $2u_v$ in `dualvar[v]` to avoid fractions (all
arithmetic stays in integers). The slack of edge $k$ connecting vertices $i$
and $j$ is therefore:

$$\text{slack}(k) = \text{dualvar}[i] + \text{dualvar}[j] - 2 \, w(k)$$

Initial vertex duals are set to the maximum edge weight:

$$\text{dualvar}[v] = w_{\max} \quad \text{for all } v \in V$$

Initial blossom duals are zero: $z_B = 0$.

---

## Algorithm Structure

The algorithm proceeds in **stages**. Each stage attempts to find one
**augmenting path** -- a path between two unmatched (exposed) vertices that
alternates between unmatched and matched edges. Augmenting along such a path
increases $|M|$ by one. After at most $\lfloor n/2 \rfloor$ stages the
matching is maximum.

Within each stage the algorithm maintains a forest of alternating trees rooted
at exposed vertices. Vertices are labeled:

| Label | Name      | Meaning                                                                                |
| ----- | --------- | -------------------------------------------------------------------------------------- |
| S     | outer     | Exposed vertex, or reached by an even-length alternating path from an exposed vertex.  |
| T     | inner     | Reached by an odd-length alternating path (partner of an S-vertex via a matched edge). |
| free  | unlabeled | Not yet reached by any alternating tree.                                               |

The stage repeatedly scans edges incident to S-vertices. Three events can
occur when processing a tight edge $(v, w)$ with $v$ an S-vertex:

1. **Grow**: $w$ is free -- label $w$ as T, label its mate as S. The
   alternating tree grows by two vertices.
2. **Blossom**: $w$ is an S-vertex in the _same_ tree -- an odd cycle is
   found. Contract it into a super-vertex (see below).
3. **Augment**: $w$ is an S-vertex in a _different_ tree -- an augmenting
   path exists. Flip matched/unmatched edges along it and end the stage.

When no tight S-edges remain, the algorithm updates dual variables to create
new tight edges.

---

## The Four Delta Types

At each dual update step, the algorithm computes four candidate step sizes and
takes the minimum:

| Type       | Formula                                                                    | Resulting action                                                                 |
| ---------- | -------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| $\delta_1$ | $\min_{v \in S} u_v$                                                       | Terminate (no augmenting path exists -- only used when `maxCardinality = false`) |
| $\delta_2$ | $\min_{\substack{v \text{ free} \\ (v,w) \in E,\, w \in S}} \pi(v, w)$     | Grow: a free vertex gains a tight edge to an S-vertex                            |
| $\delta_3$ | $\min_{\substack{B_1, B_2 \in S \\ B_1 \neq B_2}} \frac{\pi(B_1, B_2)}{2}$ | Augment or discover blossom: an S-S edge becomes tight                           |
| $\delta_4$ | $\min_{\substack{B \in T \\ B \text{ non-trivial}}} z_B$                   | Expand: a T-blossom's dual reaches zero                                          |

Set $\delta = \min(\delta_1, \delta_2, \delta_3, \delta_4)$. Then update duals:

- S-vertex: $\text{dualvar}[v] \mathrel{-}= \delta$
- T-vertex: $\text{dualvar}[v] \mathrel{+}= \delta$
- S-blossom (non-trivial): $z_B \mathrel{+}= \delta$
- T-blossom (non-trivial): $z_B \mathrel{-}= \delta$

This preserves the slack of edges within S-T or T-T pairs (their dual
adjustments cancel), while strictly decreasing the slack of S-free edges
($\delta_2$) and S-S edges ($\delta_3$, with factor 2 because both endpoints
decrease). Blossom duals $z_B$ stay non-negative because $\delta \leq \delta_4$.

---

## Blossom Contraction

When two S-vertices $v$ and $w$ in _different_ top-level blossoms share a
tight edge, the algorithm traces alternating paths from both back toward tree
roots. Two outcomes are possible:

**Same root (odd cycle found).** The paths meet at a common ancestor $b$. The
vertices on the cycle $v \to b \to w$ (through the alternating tree) plus the
edge $(v, w)$ form an odd cycle of length $2k + 1$. The algorithm contracts
this cycle into a single super-vertex -- a **blossom** -- with $b$ as its
**base**:

1. All vertices in the cycle are merged into blossom $B$.
2. $B$ inherits the S-label and all edges incident to its members.
3. $B$ gets a dual variable $z_B = 0$ (it will grow in subsequent dual
   updates while $B$ remains an S-blossom).
4. Matched edges inside $B$ are preserved; the matching on $B$'s boundary is
   determined by the base vertex.

**Different roots (augmenting path found).** The path from $v$'s root through
the S-T tree to $v$, across edge $(v, w)$, and from $w$ through its S-T tree
to $w$'s root forms an augmenting path. Flip matched/unmatched edges along it.

### Blossom expansion

When a non-trivial T-blossom's dual $z_B$ reaches zero ($\delta_4$), the
blossom is expanded back into its constituent sub-blossoms. Sub-blossoms are
relabeled (some become S, some T) so the alternating tree structure is
maintained. At end-of-stage, S-blossoms with $z_B = 0$ are also expanded.

---

## Augmenting Path

When an augmenting path is found, the `augmentMatching` function flips
matched/unmatched edges along it:

1. Trace from both endpoints of the discovering edge back through the
   alternating trees to their respective roots (exposed vertices).
2. Along each trace, swap the matched/unmatched status of every edge.
3. If the path passes through a non-trivial blossom, `augmentBlossom`
   recursively rotates the blossom's internal child list so the entry
   vertex becomes the new base.

After augmentation, $|M|$ increases by one and the stage ends.

---

## Two Implementation Variants

The `algorithm/blossom/` package provides two functions:

| Function                                                            | Weight type | Use case                                                                  |
| ------------------------------------------------------------------- | ----------- | ------------------------------------------------------------------------- |
| `MaxWeightMatching(edges []BlossomEdge, maxCardinality bool) []int` | `int64`     | Small problems or when 63 usable bits suffice                             |
| `MaxWeightMatchingBig(edges []BigEdge, maxCardinality bool) []int`  | `*big.Int`  | Swiss pairing edge weights (see [Edge Weight Encoding](../edge-weights/)) |

Both return a slice `m` where `m[i]` is the vertex matched to `i`, or `-1` if
unmatched.

The `*big.Int` variant exists because Swiss pairing edge weights encode 16+
criteria fields into a single integer. For a 100-player, 9-round tournament
the total bit width can reach approximately 294 bits -- far beyond `int64`'s
63 usable bits. The algorithm structure is identical in both variants; only the
arithmetic differs.

---

## Complexity

**Time:** $O(n^3)$ where $n = |V|$.

Each stage performs at most $O(n)$ dual updates (because each update either
grows the tree, discovers a blossom, or expands one). Each dual update scans
$O(n)$ vertices/blossoms to find the minimum delta. There are at most
$\lfloor n/2 \rfloor$ stages (one augmentation per stage). This gives:

$$O\!\left(\frac{n}{2}\right) \times O(n) \times O(n) = O(n^3)$$

**Space:** $O(n + m)$ where $m = |E|$. The dominant structures are the
neighbor lists ($O(m)$) and the per-vertex/blossom arrays ($O(n)$ each, with
up to $2n$ slots to accommodate blossoms).

---

## Implementation Notes

The Go implementation is a direct port of Joris van Rantwijk's Python
reference implementation (`mwmatching.py`, public domain). Key correspondences:

| Python             | Go (`blossom.go`)                                              |
| ------------------ | -------------------------------------------------------------- |
| `mate[v]`          | `mate[v]` -- remote endpoint index, or $-1$                    |
| `label[b]`         | `label[b]` -- `0` = unlabeled, `1` = S, `2` = T                |
| `inblossom[v]`     | `inblossom[v]` -- top-level blossom containing $v$             |
| `dualvar[v]`       | `dualvar[v]` -- stores $2u_v$ for vertices, $z_B$ for blossoms |
| `blossomchilds[b]` | `blossomchilds[b]` -- ordered child list of blossom $b$        |
| `blossombase[b]`   | `blossombase[b]` -- base vertex of blossom $b$                 |
| `bestedge[b]`      | `bestedge[b]` -- least-slack edge to a different S-blossom     |

Vertices are numbered $0, 1, \ldots, n-1$. Non-trivial blossoms are numbered
$n, n+1, \ldots, 2n-1$ and allocated from a free pool. Edges are addressed
by index $k$; their two endpoints are stored at positions $2k$ and $2k+1$ in
the `endpoint` array.

---

## References

- J. Edmonds, "Paths, trees, and flowers," _Canadian Journal of Mathematics_,
  vol. 17, pp. 449--467, 1965.
- J. van Rantwijk, `mwmatching.py` -- Python reference implementation
  (public domain).
- Z. Galil, "Efficient algorithms for finding maximum matching in graphs,"
  _ACM Computing Surveys_, vol. 18, no. 1, pp. 23--38, 1986.
