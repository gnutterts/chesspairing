---
title: "Algorithms"
linkTitle: "Algorithms"
weight: 90
description: "Mathematical deep-dives into the algorithms behind tournament pairing — with formulas, proof sketches, and pseudocode."
---

This section explores the mathematics behind chesspairing's pairing, scoring,
and tiebreaking engines. The pages that follow present formulas, proof sketches,
complexity analyses, and pseudocode -- together with links to the Go source
where each algorithm is implemented.

**Target audience.** Researchers, mathematicians, and developers who want to
understand the _why_ behind the code -- not just the API surface. If you are
looking for usage examples or configuration options, see the
[Getting Started](../getting-started/) and [Formats](../formats/)
sections instead.

**Scope and rigour.** These are proof sketches and worked intuitions, not formal
proofs. Where a result is well-known (e.g. the LP relaxation of maximum weight
matching), we state the theorem and reference a textbook proof. Where the
reasoning is specific to this codebase (e.g. the edge weight bit layout), we
give a complete derivation.

**FIDE regulations.** Regulation text throughout this section is paraphrased,
not quoted verbatim. For the authoritative source, consult the
[FIDE Handbook -- Chess Regulations](https://handbook.fide.com/chapter/C0403).

---

## Pages by category

### Core Algorithms

| Page                                  | Summary                                                                                                                            |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| [Blossom Matching](blossom/)          | Edmonds' algorithm for maximum weight matching in general graphs -- $O(n^3)$ with blossom contraction.                             |
| [Edge Weight Encoding](edge-weights/) | How 16+ pairing criteria are packed into a single `*big.Int` edge weight so that Blossom maximization respects criterion priority. |
| [Completability](completability/)     | Stage 0.5 pre-matching that determines the bye recipient before the real matching begins.                                          |

### Pairing System Specifics

| Page                                    | Summary                                                                                                                                       |
| --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| [Berger Tables](berger-tables/)         | FIDE round-robin rotation (C.05 Annex 1) -- constructing the $n-1$ round schedule from Berger's 1895 table.                                   |
| [Varma Tables](varma-tables/)           | Federation-aware pairing number assignment for round-robin tournaments (C.05 Annex 2).                                                        |
| [Baku Acceleration](baku-acceleration/) | Virtual points in early rounds (C.04.7) -- reducing draws among top seeds by inflating initial score groups.                                  |
| [Dutch Criteria](dutch-criteria/)       | The 21 optimization criteria of the Dutch system (C.04.3) -- from absolute constraints $C_1$--$C_4$ through quality criteria $C_8$--$C_{21}$. |
| [Dubov Criteria](dubov-criteria/)       | The 10 criteria of the Dubov system (C.04.4.1) with MaxT tracking and ascending-ARO sorting.                                                  |
| [Lim Exchange Matching](lim-exchange/)  | Exchange-based matching (C.04.4.3) with four floater types (A--D) and median tiebreaking.                                                     |
| [Lexicographic Pairing](lexicographic/) | DFS backtracking over criteria functions, shared by Double-Swiss (C.04.5) and Team Swiss (C.04.6).                                            |

### Scoring and Tiebreaking

| Page                                      | Summary                                                                                                             |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| [Keizer Convergence](keizer-convergence/) | Iterative scoring with oscillation detection -- convergence proof sketch for the fixed-point iteration.             |
| [Elo Model](elo-model/)                   | The expected score function $E = \frac{1}{1 + 10^{-d/400}}$ and its use in performance rating tiebreakers.          |
| [FIDE B.02 Table](fide-b02/)              | The rating difference $\leftrightarrow$ expected score lookup table, including interpolation and boundary handling. |
| [Color Allocation](color-allocation/)     | Six different color allocation algorithms compared: Dutch, Burstein, Dubov, Lim, Double-Swiss, and Team Swiss.      |
