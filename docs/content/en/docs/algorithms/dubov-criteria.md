---
title: "Dubov Criteria"
linkTitle: "Dubov Criteria"
weight: 11
description: "FIDE C.04.4.1 criteria C1-C10 for Dubov pairings: definitions, measurement, comparison order and examples."
---

Dubov uses the criteria of C.04.4.1 (effective from 1 February 2026), article 2.
There are ten of them: three absolute criteria (C1--C3), one completion
criterion (C4) and six quality criteria (C5--C10). This page gives for each the
text of the article, the function that measures it in `pairing/dubov`, and
worked examples. The procedure that uses them is on the
[Dubov system](../../pairing-systems/dubov/) page.

## Where the criteria are used

| Step | Criteria | Article |
| --- | --- | --- |
| Choosing the pairing-allocated bye (PAB) | C2, and C4 through the full bracket procedure | 3.1 |
| Choosing the number of upfloaters | C1, C3, C4, C5 | 3.2.1 |
| Choosing the set of upfloaters | score multiset first (4.2.3), then C6--C10 | 3.2.2 |
| Choosing a recomposition | first legal recomposition in 3.2.4 and 4.3 order | 3.2.4 |
| Choosing the transposition of G2 | C1, C3 only; the first legal one wins | 3.2.6 |

Forbidden pairs (option `forbiddenPairs`) are not part of C.04.4.1. The code
treats them as one more absolute criterion: `SatisfiesAbsoluteCriteria` tests
C1, C3 and the forbidden pairs together.

## Absolute criteria

### C1 -- no rematch (2.1.1)

*Text:* two participants shall not play against each other more than once.

*Measured by* `C1NoRematches` in `criteria.go`, which is true when
`swisslib.HasPlayed` finds no earlier game between the two. Only games played
over the board count: a forfeited game is not in the opponent list, so those
two players may meet again.

*Example:* p2 and p5 met in round 1. In round 2 the transposition that puts p2
against p5 is rejected and the next transposition in article 4.4 order is
tried.

### C2 -- no second pairing-allocated bye (2.1.2)

*Text:* a participant who has already received a pairing-allocated bye, or has
already scored in one single round, without playing, as many points as rewarded
for a win, shall not receive the pairing-allocated bye.

*Measured by* `filterNoByeReceived` in `bye.go`, which drops every player whose
`PABIneligible` flag is set. The flag is set by an earlier PAB, by an earlier
full-point bye and by an earlier forfeit win. `orderedByeCandidates` in
`dubov.go` uses it when `Pair` picks the bye.

*Example:* a player who received the PAB in round 3 is not a candidate in
round 6. A player with a half-point bye is still a candidate; a player who won
by forfeit is not.

### C3 -- no two absolute colour preferences for the same colour (2.1.3)

*Text:* two players with the same absolute colour preference shall not meet.
Article 1.6.1: a preference is absolute when the colour difference is greater
than +1 or less than −1, or when the player had the same colour in the two
latest rounds they played.

*Measured by* `C3NoAbsoluteColorConflict`, which reads both preferences through
the local `preference` function (article 1.6.4 included) and rejects the pair
only when both are absolute and for the same colour. There is no exception for
top scorers. A player who has yet to play has a mild preference for Black and
is never absolute.

*Example:* a player whose games were Black, Black has an absolute preference
for White; so has a player with Black, Black, White, Black (three Blacks, one
White: difference above +1 for Black). Those two cannot meet. A player with
one Black game has only a strong preference for White and can be paired
against either of them (C7 then counts the preference that is not met).

## Completion criterion

### C4 -- a legal pairing must exist (2.2.1)

*Text:* a pairing complying with all the absolute criteria shall always exist
for all players not yet paired.

*Measured by* `completable` in `dubov.go` (not in `criteria.go`). It builds a
graph in which two players are connected when `SatisfiesAbsoluteCriteria`
holds and asks the Blossom algorithm for a maximum-cardinality matching;
the answer is yes only when every player is matched. It is used twice:

- When choosing an upfloater set (3.2.1/3.2.2), the players that stay behind in
  the lower scoregroups must still pass the test; otherwise the set is skipped.
- For the PAB (3.1.2) the code removes each candidate in turn and runs the whole
  bracket procedure on the rest (which contains the test above for every
  bracket); the first candidate for whom that succeeds receives the bye.

*Example:* a bracket of three residents needs one upfloater from the players
below. If choosing p4 would leave p5 and p6 behind and those two have already
met, p4 is skipped, however well p4 would fit the bracket.

## Quality criteria

Quality is stored in `candidateQuality` and computed by `qualityFor(pairs,
upfloaterIDs, lastRound, maxT)` for the fully paired bracket, where
`upfloaterIDs` are the upfloaters of the set being judged.

### C5 -- number of upfloaters (2.3.1)

*Text:* minimise the number of upfloaters.

*Measured by* `candidateQuality.upfloaters`, the size of the set. Because
`selectBracket` starts with the smallest number that fits the parity of the
bracket and only goes up when no admissible set exists, the sets that are
compared always have the same size; C5 is satisfied by construction.

*Example:* in a scoregroup of three residents at least one upfloater is needed
(three players cannot be paired). The code first tries every set of one. Only
if none gives a legal pairing does it try sets of three.

### C6 -- score differences in pairs with upfloaters (2.3.2)

*Text:* minimise the score differences (taken in descending order) in the pairs
involving upfloaters, i.e. maximise the scores (taken in ascending order) of the
upfloaters.

*Measured by* `candidateQuality.differences`: for every pair that contains an
upfloater, the absolute score difference of the two players; the list is sorted
in descending order. Lists are compared element by element, the smaller value
wins at the first difference. When one list is the beginning of the other the
shorter one wins (fewer pairs with upfloaters, for example two upfloaters paired
together). This is the vector comparison of the text, not a sum.

*Example:* the residents have 4 points and two upfloaters are needed from
players with 3, 2, 2 and 1 point.

| Upfloaters | Differences, descending | Rank |
| --- | --- | --- |
| 3 and 2 | (2, 1) | best |
| 2 and 2 | (2, 2) | |
| 3 and 1 | (3, 1) | |
| 2 and 1 | (3, 2) | worst |

`{2,2}` beats `{3,1}` although both sums are 4, because the largest difference
is smaller.

### C7 -- players who miss their colour preference (2.3.3)

*Text:* minimise the number of players who do not get their colour preference.

*Measured by* `candidateQuality.colorViolations` (`CriterionC7` is the exported
form). The G1 player of each pair counts as White and the G2 player as Black;
every player who prefers the other colour adds 1. The preference is the one of
`preference`, so a player without a game prefers Black. The final colours of
article 5.2 are not used here; they are fixed after pairing.

Article 3.2.4 takes the first legal recomposition; it does not compare
recompositions on C7. C7 separates admissible upfloater sets within a score-
multiset group, after C6.

*Example:* six players, round 1 pairs p1--p4, p5--p2 and p3--p6; p1, p2 and p3 win.
In round 2 the residents p1, p2, p3 need one upfloater from p4, p5, p6 (4.2
order: p4, p5, p6 at equal score and ascending TPN). p1 and p3 prefer Black, p2
prefers White (it played Black).

| Upfloater | G1 (White) / G2 (Black) | Violations |
| --- | --- | --- |
| p4 (played Black, prefers White) | G1 {p2, p4}, G2 {p1, p3} | 0 |
| p5 (played White, prefers Black) | G1 {p2, p3}, G2 {p1, p5} after shifting p3 | 1 |
| p6 (played Black, prefers White) | G1 {p2, p6}, G2 {p1, p3} | 0 |

C5 and C6 are equal for the three sets (one upfloater, difference 1). p5 loses
on C7. p4 and p6 tie on C7 and on C8--C10, so the earlier set, p4, is used. The
pairing is p2--p1 and p4--p3, and p6--p5 in the next bracket.

### C8 -- maximum upfloaters (2.3.4)

*Text:* unless it is the last round, minimise the number of upfloaters who are
maximum upfloaters (article 1.8: a player is a maximum upfloater when they have
already been upfloated `MaxT` times).

*Measured by* `candidateQuality.maximumUpfloaters`: the number of upfloaters of
the set whose `UpfloatCount` is at least `MaxT`. `UpfloatCount` counts the past
rounds in which the player's score at the start of the round was lower than the
opponent's. `CriterionC8(players, maxT)` is the exported form; it counts over
the players it is given.

*Example:* `totalRounds` is 9, so `MaxT = 2 + floor(9/5) = 3`. X has upfloated
three times, Y twice. A set with X has C8 = 1, a set with Y has C8 = 0; if C5--C7
tie, the set with Y is used.

### C9 -- times a maximum upfloater is upfloated (2.3.5)

*Text:* unless it is the last round, minimise the number of times a maximum
upfloater is upfloated.

*Measured by* `candidateQuality.maximumTimes`: for the maximum upfloaters of the
set, the sum of their `UpfloatCount`. A player who has already been upfloated
more often weighs more.

*Example:* with `MaxT` 3, a set with X (3 earlier upfloats) has C9 = 3 and a set
with Z (4 earlier upfloats) has C9 = 4. With equal C5--C8 the set with X is used.

### C10 -- upfloaters from the previous round (2.3.6)

*Text:* unless it is the last round, minimise the number of upfloaters who
upfloated in the previous round.

*Measured by* `candidateQuality.previousUpfloaters`: the number of upfloaters of
the set whose last entry in the float history is an upfloat (`CriterionC10` is
the exported form).

*Example:* U upfloated in round 4, V did not. Pairing round 5, a set with U has
C10 = 1, a set with V has C10 = 0.

## Order of comparison

`betterC6ToC10(left, right)` compares fully paired brackets within one
upfloater score-multiset group, lexicographically, and reports whether `left` is
better:

1. the C6 vectors as described above,
2. fewer colour violations (C7),
3. fewer maximum upfloaters (C8),
4. a smaller sum of earlier upfloats of maximum upfloaters (C9),
5. fewer upfloaters from the previous round (C10).

The minimum number of upfloaters (C5) is fixed before grouping, and groups are
considered in score-multiset order. A criterion is only looked at when all
criteria before it are equal. A candidate replaces the current best only when it
is strictly better, so with equal quality the first set in article 4.1.3 order
is kept.

*Example:* a set that gives C6 = (2, 1) and four colour violations beats a set
with C6 = (2, 2) and none: C6 has priority over C7.

## Last-round relaxation

In the last round, C8, C9 and C10 are not applied (articles 2.3.4--2.3.6).
`qualityFor` returns right after C5--C7 when `lastRound` is true, so those
fields stay 0 for every candidate. The last round is the round in which
`CurrentRound == totalRounds`. C5, C6 and C7 stay in force.

*Example:* in round 9 of 9, a set with a maximum upfloater and a set without one
tie when C5--C7 are equal, and the first set in 4.1.3 order is used.

## MaxT

`MaxT(totalRounds) = 2 + floor(totalRounds / 5)` (article 1.8.2), with
`totalRounds` the planned length of the tournament, not the number of rounds
played.

| Planned rounds | MaxT |
| --- | --- |
| 1--4 | 2 |
| 5--9 | 3 |
| 10--14 | 4 |
| 15--19 | 5 |

## Sorting orders used with the criteria

| Article | What is sorted | Order | Code |
| --- | --- | --- | --- |
| 4.1.3 | sets of K players from a numbered pool | smallest differing sequence number first | `forEachCombination` |
| 4.2 | possible upfloaters | descending score, then ascending TPN | `groupPlayers`, `lowerPlayers` |
| 4.3 | shifters | White seekers by ascending ARO then TPN, Black seekers by ascending TPN; numbers assigned middle first | `shifterOrder` |
| 3.2.5 | G1 (S1) | ascending ARO, then ascending TPN | `SortByAROAscending` |
| 4.4 | G2 (T2) | depth-first search in lexicographic order of ascending TPN; the first legal one is used | `firstLegalTransposition` |

The ARO is rounded to an integer, .5 up (`ComputeARO`, article 1.7.1), before it
is compared. TPN always means the fixed tournament pairing number.

*Example of 4.3.3:* the sorted list A, B, C, D, E, F, G is numbered D, C, E, B, F,
A, G. *Example of 4.4.2:* G2 with A, B, C gives {A,B,C}, {A,C,B}, {B,A,C},
{B,C,A}, {C,A,B}, {C,B,A}.

## Related code

`DubovCandidateScore` and `NumDubovViolations` are still exported for callers of
the former criterion inspector. The pairing procedure does not use them and
their seven violation slots do not follow the 2026 numbering. The 2026
criteria are evaluated through `candidateQuality` as described above.

## Interpretations

- **C6 prefix rule.** When one difference vector is the beginning of the other,
  the shorter one is better. The text does not say what happens then.
- **C9.** "The number of times a maximum upfloater is upfloated" is read as the
  sum of the earlier upfloat counts of the maximum upfloaters in the set.
- **C8--C10 on the set.** They are measured on the upfloaters of the set being
  judged, not on all players of the bracket.
- **Upfloat history.** A round counts as an upfloat when the player had fewer
  points than the opponent at the start of that round, with the tournament's configured points (and virtual points when accelerated).
  C10 uses the last entry in a player's float history.
- **C4 and the PAB.** Article 3.1.2 is tested by running the bracket procedure on
  the remaining players; a candidate for whom that procedure cannot complete is
  skipped.
- **Enumeration safety limit.** The text has no numeric limit. The code keeps
  the article 4.4 and 4.1.3 order and returns `ErrBracketTooLarge` when a shared
  per-bracket budget of 1,000,000 search nodes (recompositions and depth-first
  transposition search) or a 200,000-candidate set-enumeration limit is
  exceeded.

## Related pages

- [Dubov system](../../pairing-systems/dubov/) -- the complete procedure, options
  and colour rules.
- [Completability](../completability/) -- the matching test behind C4.
- [Color allocation](../color-allocation/) -- how the colours of article 5.2 are
  given after pairing.
