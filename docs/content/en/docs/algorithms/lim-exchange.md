---
title: "Lim Exchange Matching"
linkTitle: "Lim Exchange"
weight: 12
description: "The FIDE C.04.4.3 Article 4 exchange procedure."
---

## Overview

This page describes the exchange engine used by the Lim pairer. It is the
Article 4 procedure applied after Article 2.3 has transferred players and made
the score group even. It is not greedy or global matching.

## Compatibility (Art. 2.1)

A pair must be new and must permit colours without three consecutive equal
colours or a difference of three in the colour counts. Forbidden pairs are also
not compatible. In the last round the Article 6 exception relaxes the colour
part so that an available same-score new pairing has priority.

## Proposed Pairings (Art. 2.4)

Players are ordered by pairing number and split into an upper and lower half.
The proposal is first upper against first lower, second against second lower,
and so on. The group has already had the mandatory 2.3.1--2.3.3 transfers; a
2.3.4 floater is selected only if it is still odd.

## Exchange Rules (Arts. 4.1--4.4)

The proposed pairings are scrutinised in turn. Pairing downward starts with
#1; pairing upward starts with #n. On a failed downward proposal the lower
numbered player is exchanged until a compatible pairing is found; upward uses
the corresponding higher numbered player. The implementation follows the
exchange sequences in the 4.2 and 4.3 tables, preserving an earlier compatible
pair while examining the next proposal.

Article 4.4 retains the earlier pair when the next player cannot be given a
compatible opponent. Under 4.4.1 the failing player is exchanged with the
floater that made an originally odd group even. Under 4.4.2, for an originally
even group, the failing player floats with the lowest-numbered remaining
player, keeping the group even.

No arbitrary fallback is used. Resulting floaters continue through Articles 3
and 2.6; an impossible round returns `PairingImpossible`, not partial results
with extra byes.

## Floater Rules Used by the Engine

Floater selection follows Arts. 3.2--3.5 and 3.9: the mandatory 3.3/3.5 rules
choose or exchange a compatible floater, and 3.9.2 orders alternatives D, B,
C, A. Article 3.10 protects a same-side previous-round transferred floater
(not a bye) from a 2.3.4 float if 3.10.1 and 3.10.2 allow it. Floater pairing
uses 3.6--3.8, including the Maxi comparison of the two exchanged opponents.

## Median Group and Cracking

Lim processes groups above the median downward, below it upward, and the
median last. A blocked median group is extended by Art. 2.6 cracking. If no
player has the median score, this implementation interprets the arriving
floaters as the median group at that position; 2.6 applies. In the median
group, it interprets 3.9 types against the already-paired group below. These
are interpretations where the FIDE text does not decide the case.

It also interprets 4.4.1 for the failing player whenever the group chose a
2.3.4 floater, including where 2.3.1--2.3.3 made an originally even group odd.

## Related Pages

- [Lim Pairing](/docs/pairing-systems/lim/) describes the complete system.
- [Colour Allocation](../color-allocation/) describes shared colour concepts.
