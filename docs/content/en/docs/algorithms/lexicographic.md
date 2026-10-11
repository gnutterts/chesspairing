---
title: "Lexicographic Pairing"
linkTitle: "Lexicographic"
weight: 13
description: "Article 3.6 identifier-order bracket pairing for Double-Swiss and Team Swiss."
---

## Overview

Double-Swiss (FIDE C.04.5) and Team Swiss (FIDE C.04.6) share the bracket
pairer in `pairing/lexswiss/bracket.go`. For even-sized brackets it selects the
first valid pairing in the identifier order required by Articles 3.6.1--3.6.3.
Odd-sized brackets retain the existing sequence-order handling.

## Identifier order

For every pair, the participant with the smaller fixed tournament pairing number
(TPN) is its top member and the other is its bottom member. The TPN is assigned
once under Article 1.1 and is not the participant's current score rank. A pairing identifier
is the ascending sequence of top-member TPNs followed by the corresponding
bottom-member TPNs. Pairings are compared lexicographically by that identifier.

The implementation generates this order lazily:

1. Select top-member sets in ascending lexicographic order.
2. For each set, assign bottom members in ascending lexicographic order.
3. Reject a candidate pair that violates C1, a forbidden-pair restriction, or a
   system-specific criterion. Infeasible top-member prefixes are pruned with a
   matching check.

For six unconstrained participants, the first identifier is `1 2 3 4 5 6`,
which represents `1-4, 2-5, 3-6`.

## Criteria function

```go
type CriteriaFunc func(a, b *ParticipantState) bool
```

The function is called for each candidate pair after the C1 and forbidden-pair
checks. It returns whether that pair meets the system-specific criterion.
Double-Swiss uses it for its colour preference criterion; Team Swiss uses it
for its colour and floater criteria.

## Example

Consider TPNs 3, 7, 12, 15, 22 and 28. Participant 3 has played 7, and 12
has played 15. The first possible top-member sequence is `3 7 12`. Its first
legal bottom assignment is `15 22 28`, so the pairer returns:

```
(3, 15), (7, 22), (12, 28)
```

Its identifier is `3 7 12 15 22 28`.

## Top-down brackets

Double-Swiss and Team Swiss start with the highest scoregroup still unpaired.
They combine all of its residents with the first upfloater set that works: first
the smallest number of upfloaters (C4), then the best score profile (C5), then
C6, C7, and finally the Article 3.5.4 TPN order. A set works only when its
Article 3.6 bracket is legal and the remaining players have a C1/C2 completion.
C6 compares the working sets: it prefers the sets after which the following
scoregroup needs the fewest upfloaters, as Article 2.4.3 [C8] of the Dutch
system (C.04.3) does for downfloaters. A colour preference never makes a legal
pairing impossible: only C1 and C2 are absolute. Brackets are never pooled as a
fallback; an impossible round is reported for the Chief Arbiter under Article
3.3.3.

## Complexity

The number of possible identifiers is exponential in the worst case. The
pairer stops at the first valid complete identifier and prunes top-member
prefixes for which the selected tops cannot have distinct legal bottom members,
and skips a bottom member when the remaining tops could no longer all receive
one. This keeps ordinary brackets fast while preserving the required order.

## Related Pages

- [Double-Swiss Pairing](/docs/pairing-systems/double-swiss/)
- [Team Swiss Pairing](/docs/pairing-systems/team/)
- [Dutch Criteria](../dutch-criteria/)
