---
title: "Dubov System"
linkTitle: "Dubov"
weight: 3
description: "FIDE C.04.4.1 Dubov pairing system."
---

## Overview

The Dubov system implements FIDE C.04.4.1 (approved by the Council on
28 October 2025, applied from 1 February 2026). Its aim (article 0) is that a
player with more points than another should also have a higher performance
rating. It tries to reach that by equalising the Average Rating of Opponents
(ARO, article 1.7) of all players in a scoregroup: players with a low ARO are
paired against opponents with a high rating.

The package is `pairing/dubov`. A round is paired in this order (article 1.9.2):

1. Players named in `state.PreAssignedByes` are taken out of the field and
   returned unchanged in `PairingResult.Byes`.
2. With an odd number of remaining players, the pairing-allocated bye (PAB) is
   assigned first (article 3.1).
3. The scoregroups are paired in descending order of score (article 3.2), each
   as a bracket of residents plus the upfloaters pulled up from lower
   scoregroups.
4. The pairs are put in board order and given colours (article 5.2).

If the round cannot be completed, no pairing is returned at all; see
[When no pairing exists](#when-no-pairing-exists).

## Configuration

### CLI

```bash
chesspairing pair --dubov tournament.trf
```

The TRF reader copies the total number of rounds (record `142` of TRF-2026,
otherwise `XXR`), the initial colour and the forbidden pairs into the
options below.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dubov"

// With typed options
p := dubov.New(dubov.Options{
    TotalRounds:    chesspairing.IntPtr(9),
    TopSeedColor:   chesspairing.StringPtr("auto"),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// From a generic map (e.g. parsed from JSON config)
p := dubov.NewFromMap(map[string]any{
    "totalRounds":  9,
    "topSeedColor": "black",
})
result, err := p.Pair(ctx, &state)
```

### Options

| Option (JSON key) | Go field | Type | Default | Meaning |
| --- | --- | --- | --- | --- |
| `totalRounds` | `TotalRounds` | `*int` / integer | none; **required** | Planned number of rounds `Rnds`. It fixes `MaxT = 2 + floor(Rnds/5)` (article 1.8.2) and decides which round is the last round (`CurrentRound == totalRounds`; C8--C10 are then not applied). |
| `topSeedColor` | `TopSeedColor` | `*string` / string | `"auto"` | The initial colour of article 5.1, used by article 5.2.1 for every pair in which both players have yet to play a game. `"auto"` and `"white"` both mean White; `"black"` means Black; any other value is treated as White. |
| `forbiddenPairs` | `ForbiddenPairs` | `[][]string` | none | Pairs of player IDs that must not meet. They count as an absolute criterion next to C1 and C3: they apply in every bracket, in the C4 completability test and in the bye selection. Entries that do not have exactly two IDs are ignored. |

Details:

- **`totalRounds` is required.** When it is missing, or smaller than 1,
  `Pair` returns the plain error `dubov: total rounds required` before it pairs
  anything. In a map it is read as an `int`, an `int64` or a whole-number
  `float64` (the type JSON decoding produces); a fractional number or another
  type counts as missing. The code does not derive it from the rounds played
  so far, because `MaxT` and the last-round relaxation are defined on the
  planned length of the tournament.
- **`forbiddenPairs` from a map** must be a JSON-style `[]any` whose elements
  are `[]any` pairs of two strings; a Go `[][]string` stored in the map is not
  recognised. With the typed `Options` it is a plain `[][]string`.
- `WithDefaults` only fills `topSeedColor`; the other options have no default.

## Terms the procedure uses

| Term | As implemented |
| --- | --- |
| Score | The pairing score with 1 / ½ / 0 for win / draw / loss, whatever scoring system the tournament uses for standings. A PAB or full-point bye counts 1, a half-point bye ½, a forfeit win 1. |
| TPN (1.2) | The fixed tournament pairing number: `PlayerEntry.PairingNumber` when it is set, otherwise assigned by rating, title and name (C.04.2 article 2). Every TPN sort below uses this fixed number, not the rank in the current round. The code does not recalculate TPNs (1.2.2). |
| Rating (1.1) | `PlayerEntry.Rating`, used as given. The code does not invent a provisional rating (1.1.2); a player without a rating has rating 0. |
| Games played | Rounds with a colour. A bye or a forfeit is not a game: no colour, no opponent, not counted in the ARO, and the two players of a forfeited game may meet again. |
| ARO (1.7) | Mean of the ratings of the opponents met over the board, rounded to the nearest integer with .5 rounded up (`ComputeARO`, integer arithmetic `(2·sum + n) / (2·n)`); 0 for a player who has yet to play (1.7.3). Opponents who withdrew later still count. Example: opponents rated 1000 and 1001 give 1000.5, so ARO 1001. |
| MaxT (1.8.2) | `MaxT(Rnds) = 2 + floor(Rnds/5)`: 2 for 1--4 rounds, 3 for 5--9, 4 for 10--14, 5 for 15--19. |
| Colour preference (1.6) | Classified by the shared `swisslib.ComputeColorPreference` over played games only: absolute (colour difference above +1 or below −1, or the same colour in the last two games), strong (difference ±1), mild (difference 0: alternate). Dubov adds article 1.6.4 itself: a player without a game has a mild preference for Black. |
| Upfloat | A past round counts as an upfloat when the player's score at the start of that round was lower than the opponent's. `UpfloatCount` counts those rounds; the last entry of the float history tells whether the player upfloated in the previous round. |

## Procedure

### 3.1 Pairing-allocated bye

With an even number of players there is no bye. With an odd number, the
candidates are the players that may still receive one (3.1.1, criterion C2):
players who have not yet received a PAB and have not scored a full point
without playing (a full-point bye or a forfeit win). They are sorted by

1. lowest score (3.1.3),
2. most games played (3.1.4),
3. largest TPN (3.1.5).

For each candidate, in that order, the candidate is removed and the complete
bracket procedure of 3.2 is run on the other players. The first candidate for
whom that procedure yields a complete pairing receives the bye (3.1.2, C4);
the pairing found for the others is used. The bye is recorded as a
`ByePAB` entry. When no candidate qualifies, the round fails with
`PairingNoPABCandidate`.

### 3.2 Bracket for a scoregroup

Scoregroups are handled from the highest score down. The players of the
highest remaining score are the **residents** (1.3.2). All players with a lower
score are the pool of possible upfloaters.

**3.2.1 Minimum number of upfloaters.** The code tries 0 or 1 upfloater (the
number that makes residents plus upfloaters even), then 2 or 3, and so on in
steps of 2. For each number it enumerates the possible sets (see 4.2 and 4.1.3
below). A set is *admissible* when

- the residents plus the set can be paired by the steps 3.2.3--3.2.6 below, so
  with C1, C3 and the forbidden pairs respected, and
- the players that stay behind in the lower scoregroups still admit a complete
  pairing under C1, C3 and the forbidden pairs (C4, function `completable`: a
  maximum-cardinality matching in which every player must be matched).

The first number that has an admissible set is the minimum (3.2.1, C5). A
resident is never moved down: when residents cannot be paired, more upfloaters
are pulled up, and if no number works the round fails.

**3.2.2 Choosing the set.** Sets of the minimum size are grouped by their score
multiset (4.2.3) and the groups are considered in C6 order. The first group with
an admissible set is used; within that group the code chooses the set that best
satisfies C6--C10, in that order (function `betterC6ToC10`; see
[Dubov Criteria](../../algorithms/dubov-criteria/)). A later set replaces the
current best only when it is strictly better, so with equal quality the first set
in article 4.1.3 order stays ("the first set of upfloaters that ... complies at
best"). C5 is already fixed by the minimum size.

After the choice, the residents and the selected upfloaters leave the field,
the next scoregroup is handled with its remaining residents, and so on until
everyone is paired.

**3.2.3 G1 and G2.** The bracket is split in two subgroups:

- When no player of the bracket (upfloaters included) has played a game, as in
  round 1, G1 holds the first half of the bracket by TPN and G2 the second half.
  This depends on the players, not on the round number.
- Otherwise G1 holds the players with a colour preference for White and G2 all
  other players. A player without a game prefers Black (1.6.4) and so belongs
  in G2.

**3.2.4 Recomposition.** The code first follows the first sentence of the
article. For `s = 0, 1, 2, ...`, it shifts `s` players out of the smaller
subgroup (G1 when the subgroups are equal), in 4.3 shifter order, to avoid
unavoidable same-subgroup pairs. It then equalises the subgroup sizes: players
who still seek the colour of the larger subgroup are shifted from it to the
smaller subgroup, in 4.3 order. A player shifted in the first stage keeps the
colour preference of the subgroup it left, so it is not shifted back. The
exception is a bracket where nobody has played yet (round 1): everyone prefers
Black there, the article does not say whether a first-stage shifter may return,
and the code allows it. G1 is paired against G2, so the resulting subgroups must
be equally large. For each recomposition the first legal transposition (3.2.6)
is sought; the **first legal recomposition** is used.

**3.2.5 S1.** The players of (the possibly new) G1 are sorted by ascending ARO
and, with equal ARO, by ascending TPN. The result is S1. The ARO is the rounded
value of 1.7.1.

**3.2.6 T2.** G2 is put in ascending TPN order and its transpositions are
tried in the order of 4.4. S1 and T2 are paired position by position: first
with first, second with second. The **first** transposition in which every pair
is legal (C1, C3, forbidden pairs) is used. The code does not compare
transpositions on quality, and there is no limit of 120 transpositions.

## Sorting criteria

### 4.1 Sets by sequence number

A pool of P players receives sequence numbers #1 to #P. The sets of K players
are ordered by the smallest differing sequence number, which is lexicographic
order on the sorted numbers: with K = 2, {#1,#2} comes before {#1,#3}, and
{#1,#P} before {#2,#3}. `forEachCombination` enumerates in exactly this order.

### 4.2 Upfloaters

The pool is every player with a score below that of the residents. The sequence
number follows descending score and, with equal scores, ascending TPN.

Example: the pool is A (1.5 points, TPN 7), B (1.0, TPN 4), C (1.0, TPN 9) and
D (0.5, TPN 2). The numbers are #1 A, #2 B, #3 C, #4 D. Sets of two come in the
order {A,B}, {A,C}, {A,D}, {B,C}, {B,D}, {C,D}.

### 4.3 Shifters

A player with a colour preference for White (Black) is a possible White (Black)
shifter (4.3.1). White seekers are sorted by ascending ARO and, with equal ARO,
ascending TPN; Black seekers by ascending TPN (4.3.2). Sequence numbers are
then assigned starting with the player in the middle of the remaining list;
when two players are in the middle, the one higher in the list goes first
(4.3.3). `shifterOrder` takes the element at index `(n−1)/2` of the remaining
list, removes it, and repeats.

Example of 4.3.3: the sorted list is A, B, C, D, E, F, G. #1 is D (middle of
seven), #2 is C (higher of C and E, the middle of the remaining six), #3 is E,
#4 is B, #5 is F, #6 is A, #7 is G. Sets of two shifters are therefore tried
as {D,C}, {D,E}, {D,B}, {D,F}, {D,A}, {D,G}, {C,E}, and so on.

### 4.4 Transpositions of G2

The players that are in G2 after the recomposition form the pool and receive
sequence numbers by ascending TPN (4.4.1, 4.4.2). The code finds the first legal
transposition with a depth-first search in the lexicographic order of those
permutations.

Example of 4.4.2: with A, B, C in G2 (ascending TPN) the order is {A,B,C},
{A,C,B}, {B,A,C}, {B,C,A}, {C,A,B}, {C,B,A}. If S1 is (x1, x2, x3) and x2 has
already played B, the first transposition fails on x2–B, the second gives
x1–A, x2–C, x3–B, which is used when it is legal.

To keep a pathological bracket from running forever, one shared budget of
1,000,000 search nodes covers its recompositions and depth-first transposition
search; a single enumeration of sets is also capped at 200,000 candidates.
Beyond either limit `Pair` returns `ErrBracketTooLarge` rather than skip part of
the prescribed order. A 200-player, 11-round tournament pairs in a few seconds.

## Criteria

The absolute criteria C1--C3 and the completion criterion C4 must always hold.
The quality criteria C5--C10 decide between candidates, in descending priority:

| Criterion | Article | Rule | Measured by |
| --- | --- | --- | --- |
| C1 | 2.1.1 | No rematch. | `C1NoRematches` |
| C2 | 2.1.2 | No second PAB, none after a full-point unplayed result. | `filterNoByeReceived` (in `orderedByeCandidates`) |
| C3 | 2.1.3 | No two players with the same absolute colour preference. | `C3NoAbsoluteColorConflict` |
| C4 | 2.2.1 | A legal pairing must exist for everyone not yet paired. | `completable`, and the bye loop of 3.1 |
| C5 | 2.3.1 | Minimise the number of upfloaters. | `selectBracket` (smallest number first), `candidateQuality.upfloaters` |
| C6 | 2.3.2 | Minimise the score differences, in descending order, in pairs with upfloaters. | `candidateQuality.differences` |
| C7 | 2.3.3 | Minimise the number of players who miss their colour preference. | `candidateQuality.colorViolations` |
| C8 | 2.3.4 | Unless it is the last round, minimise the number of maximum upfloaters. | `candidateQuality.maximumUpfloaters` |
| C9 | 2.3.5 | Unless it is the last round, minimise the number of times a maximum upfloater is upfloated. | `candidateQuality.maximumTimes` |
| C10 | 2.3.6 | Unless it is the last round, minimise the number of upfloaters who upfloated in the previous round. | `candidateQuality.previousUpfloaters` |

See [Dubov Criteria](../../algorithms/dubov-criteria/) for the definition,
measurement, order of comparison and worked examples of every criterion.

## Colours

The colour of each pair follows article 5.2, in this order (`AllocateColor` in
`pairing/dubov/color.go`). Dubov has its own procedure; it does not use the
Dutch cascade of `swisslib`, only its classification of preferences.

1. **5.2.1** When both players have yet to play a game, the higher-ranked player
   (more points, or with equal points the smaller TPN) gets the initial colour
   when their TPN is odd, and the opposite colour when it is even. The TPN is the
   fixed pairing number, not the board number. The initial colour comes from
   `topSeedColor` (White unless it is `"black"`).
2. **5.2.2** Otherwise, when the preferences are for different colours, both are
   granted. A player without a game prefers Black (1.6.4).
3. **5.2.3** When both want the same colour, the stronger preference wins: an
   absolute preference beats a non-absolute one (with two absolute preferences,
   the larger colour difference wins), then a strong preference beats a mild one.
4. **5.2.4** When that does not decide, the colours alternate relative to the
   most recent time in which one player had White and the other Black. The
   played games of both are compared from the most recent one backwards; at the
   first difference each player gets the colour the other had then.
5. **5.2.5** Finally the preference of the higher-ranked player (as in 5.2.1)
   is granted.

Example of 5.2.1: in round 1 with the initial colour White, the higher-ranked
player of each pair gets White when their TPN is odd and Black when it is even:
TPN 1 plays White, TPN 2 plays Black, TPN 3 plays White, and so on. With
`topSeedColor` set to `"black"` this is reversed.

The colours measured for C7 while pairing a bracket are not these final colours:
there the G1 player counts as White and the G2 player as Black. Colour
allocation only runs after the pairs are fixed and never changes who plays whom.

## Board ordering

Under C.04.2 article 3.6, pairs are sorted (stable sort) by the higher-ranked
player's pairing score (descending), then the sum of both scores (descending),
then that player's TPN (ascending). Within a pair, higher score ranks first;
smaller TPN breaks ties.

Boards are numbered from 1 in that order. A PAB and pre-assigned byes have no
board; they are listed in `Byes`, pre-assigned byes first.

## When no pairing exists

Article 1.9.3 leaves an impossible round to the Chief Arbiter. The code never
repairs it with a best-effort pairing: it does not pair players that already
met, does not give a second PAB, and does not recompose the field to make it
fit. It returns an error and no result:

| Situation | Error |
| --- | --- |
| Even field, a bracket cannot be paired | `*chesspairing.PairingError` of kind `PairingIncomplete`, `System` `"dubov"`. `Missing` lists the IDs that stay unpaired, sorted. `Partial` holds the pairs already fixed for the higher brackets (with boards and colours) plus the pre-assigned byes. Example message: `dubov: pairing is incomplete: missing p1, p2, p3, p4`. |
| Odd field, no player can take the PAB, or each one who can leaves a field that 3.2 cannot pair | `*chesspairing.PairingError` of kind `PairingNoPABCandidate`, wrapping `swisslib.ErrNoPABCandidate`; no `Partial`. |
| No active player and no pre-assigned bye | `ErrTooFewPlayers` |
| `totalRounds` missing | the plain error `dubov: total rounds required` |
| Shared bracket-search budget or a set-enumeration limit exceeded | `ErrBracketTooLarge` |
| Context cancelled | the context's error |

Use `errors.As` to reach the `*chesspairing.PairingError`. The exported
`ErrNoPairingPossible` is the internal signal for a bracket that fails; the
caller sees it as `PairingIncomplete`, not as a wrapped error.

## Worked example

Eight players p1--p8 rated 2400, 2300, ..., 1700 (TPN 1--8), `totalRounds` 9.

**Round 1.** Nobody has played, so G1 = {p1..p4} and G2 = {p5..p8} (3.2.3). The
first transposition is legal: p1–p5, p2–p6, p3–p7, p4–p8. By 5.2.1 p1 (odd TPN)
gets the initial colour White; p2 (even) gets Black, so p6 plays White; p3 plays
White; p4 gets Black, so p8 plays White. Boards: 1 p1–p5, 2 p6–p2, 3 p3–p7,
4 p8–p4. Suppose the White players p1, p6, p3 and p8 win.

**Round 2, scoregroup 1.0** = {p1, p3, p6, p8}. All four played White once, so
all prefer Black: G1 is empty, G2 has four players. Two players must be shifted
(3.2.4). Black seekers by TPN are p1, p3, p6, p8; middle first they are numbered
#1 p3, #2 p6, #3 p1, #4 p8. The first set is {p3, p6}, so G1 = {p3, p6} and
G2 = {p1, p8}. G1 by ARO: p3 (opponent p7, 1800) before p6 (opponent p2, 2300).
The first transposition gives p3–p1 and p6–p8, which is legal. In both pairs the
two players want Black with equal strength and equal history, so 5.2.5 gives
Black to the higher-ranked player (p1, p6); p3 and p8 play White.

**Round 2, scoregroup 0** = {p2, p4, p5, p7}. All played Black, all prefer White:
G1 has four players, G2 none. White seekers by ARO: p4 (1700), p2 (1900),
p7 (2200), p5 (2400). Numbered from the middle: #1 p2, #2 p7, #3 p4, #4 p5. The
first set {p2, p7} moves to G2; G1 = {p4, p5}, G2 = {p2, p7}. The first
transposition gives p4–p2 and p5–p7. By 5.2.5 p2 and p5 (the higher-ranked) play
White.

Boards: 1 p3–p1, 2 p8–p6, 3 p2–p4, 4 p5–p7.

## Interpretations

The text is not explicit on the points below. The code follows the words as
closely as it can and does the following; these are interpretations.

- **C7 while pairing.** When upfloater sets are compared, a player counts as
  missing their preference when the G1 player (White) or the G2 player (Black)
  prefers the other colour. The final colours of 5.2 can differ.
- **3.1.2.** "Allows a complete pairing of all the remaining players" is tested
  by running the 3.2 procedure on the remaining players. A candidate whose
  remaining players have a legal pairing in theory, but cannot be completed by
  that procedure, is skipped.
- **3.2.1 and 3.2.2.** The minimum number of upfloaters is the smallest number
  for which an admissible set exists (including the C4 test on the players left
  behind); among sets of that size the quality criteria decide.
- **C6.** The differences are taken per pair that contains an upfloater
  (|score of one player − score of the other|) and compared as descending
  vectors, element by element. When one vector is the beginning of the other,
  the shorter one wins.
- **C8--C10.** A maximum upfloater is a player with at least `MaxT` upfloats.
  C9 adds up the earlier upfloat counts of the maximum upfloaters in the set.
  C10 looks at the last entry of the float history.
- **Upfloat history.** A past round counts as an upfloat when the player had
  fewer points than the opponent at the start of that round, with 1 / ½ / 0
  scores.
- **5.2.4.** "Most recent time" is read over played games only, aligned from the
  most recent game backwards. The note on article 3.4 of the General Handling
  Rules is not modelled beyond that.
- **Board order.** C.04.4.1 has no separate rule; the order above follows C.04.2 3.6.
- **Enumeration limits.** The text has no limit; the code searches in the
  prescribed order and fails with `ErrBracketTooLarge` when the shared
  per-bracket search budget or a set-enumeration limit is exceeded.

## FIDE reference

FIDE Handbook C.04.4.1 Dubov System, effective from 1 February 2026
([handbook.fide.com/chapter/C04041](https://handbook.fide.com/chapter/C04041)).
Articles used: 0 (preface), 1.1--1.9 (definitions), 2.1--2.3 (criteria C1--C10),
3.1--3.2 (bye and bracket procedure), 4.1--4.4 (sorting criteria) and 5.1--5.2
(colours).
