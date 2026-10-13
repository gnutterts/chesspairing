---
title: "Team Swiss System"
linkTitle: "Team Swiss"
weight: 6
description: "FIDE C.04.6 — Swiss pairing for team competitions with board-level color allocation."
---

The Team Swiss system pairs teams rather than individual players, using the same lexicographic bracket pairing infrastructure as the Double-Swiss system. Players link to their team through `PlayerEntry.TeamID`; legacy states may use one entry per team. The system features a nine-step colour allocation procedure and two configurable colour preference types that control how aggressively the engine tries to satisfy colour balance at the team level. Approved by FIDE in October 2025 and effective from February 2026.

## When to Use

Team Swiss is appropriate when:

- The competition is between teams (e.g. club championships, Olympiad-style events).
- You need a Swiss format that handles team-level pairing while respecting board-level colour assignments.
- The event uses match points (team wins a match) or game points (sum of individual board results) as the primary ranking criterion.

It is not suitable for individual tournaments. For individual two-game-match Swiss tournaments, use the [Double-Swiss](../double-swiss/) system.

## Configuration

### CLI

```bash
chesspairing pair --team tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/team"

// With typed options
p := team.New(team.Options{
    TopSeedColor:        chesspairing.StringPtr("white"),
    TotalRounds:         chesspairing.IntPtr(11),
    ColorPreferenceType: chesspairing.StringPtr("A"),
    PrimaryScore:        chesspairing.StringPtr("match"),
    ForbiddenPairs:      [][]string{{"team-1", "team-2"}},
})

// From a generic map (JSON config)
p := team.NewFromMap(map[string]any{
    "topSeedColor":        "white",
    "totalRounds":         11,
    "colorPreferenceType": "B",
    "primaryScore":        "game",
})
```

### Options Reference

| Option                | Type         | Default   | Description                                                                                                                                                                  |
| --------------------- | ------------ | --------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `topSeedColor`        | `string`     | `"white"` | Initial colour from drawing of lots before round 1 (Art. 4.1). Values: `"white"`, `"black"`. Note: unlike other systems, this defaults to `"white"`, not `"auto"`.           |
| `totalRounds`         | `int`        | nil       | Total number of rounds. Used to determine the last two rounds (C7/C10 relaxation) and last round (Type B mild preference calculation).                                       |
| `colorPreferenceType` | `string`     | `"A"`     | Colour preference rules (Art. 1.7). Values: `"A"` (simple), `"B"` (strong + mild), `"none"` (disabled).                                                                      |
| `primaryScore`        | `string`     | `"match"` | Score used for pairing (Art. 1.2): `"match"` or `"game"`. The other score is used for colour allocation (Art. 4.2.2). |
| `secondaryScoreForColour` | `bool` | `true` | Set to `false` if competition rules omit the secondary score from colour allocation (Art. 4.2.2). |
| `forbiddenPairs`      | `[][]string` | nil       | Pairs of team IDs that must never be paired together.                                                                                                                        |

## How It Works

The pairing algorithm proceeds in five stages, sharing most of its infrastructure with the Double-Swiss system through the `pairing/lexswiss` package.

### 1. Build Participant States

The team pairer constructs one pairing state per active team. Pairing scores use the configured primary score: match points by default, or game points when `primaryScore` is `"game"`. The underlying points follow the tournament's scoring configuration: game points use the configured board points, forfeit boards score the configured forfeit values, and a pending or adjourned match scores both teams the match-draw points for the pairing score only. Zero-point byes and absences add the configured value of their type to the game points, so they leave match points untouched. Each team receives a distinct fixed TPN from 1 through the number of teams (Art. 1.1). When competition rules do not prescribe an order, the library uses its initial team order before round one; teams are then sorted for pairing by score descending and TPN ascending, without round-by-round renumbering.

### 2. Assign the Pairing-Allocated Bye

If the team count is odd, the shared `lexswiss.AssignPAB` selects the bye recipient per Art. 3.4. Among C2-eligible teams (no earlier PAB, forfeit win, or deprecated full-point bye), it first retains only teams whose removal leaves a complete legal C1 pairing. It then selects the lowest score, the highest number of team matches actually played, and finally the largest fixed TPN, in that order. A team match that ended by forfeit is not played. A game is treated as a forfeit when either its result is a forfeit or its `IsForfeit` flag is set.

### 3. Build Score Groups

Teams are grouped by score into descending-order score groups. Within each group, teams are ordered by TPN ascending.

### 4. Lexicographic Bracket Pairing

Each bracket is paired by Article 3.6 identifier order. The search chooses the sorted set of top members first, then assigns their bottom members in ascending TPN order. An identifier consists of the top-member TPNs followed by the corresponding bottom-member TPNs. In a six-team first round without constraints this is `1-4, 2-5, 3-6`.

**Absolute criteria (always enforced):**

- C1: No two teams play each other more than once (Art. 2.1.1).
- Forbidden pairs are never paired.

**Quality criteria (C8, C9, C10):**

Under Art. 3.6.4, the algorithm selects the first identifier among pairings that minimise quality criteria C8, C9, and C10 in descending priority order (Art. 2.3):

- **C8 (colour preferences):** Minimises the number of teams whose colour preference, if any, is not fulfilled (Art. 2.3.5). In a pair where both teams prefer the same colour, exactly one team is left unfulfilled (cost 1); pairs with opposite or single preferences can both be fulfilled (cost 0).
- **C9 (strong colour preferences, Type B only):** Minimises the number of teams whose strong colour preference is not fulfilled (Art. 2.3.6). When both teams have a strong preference for the same colour, one is left unfulfilled (cost 1).
- **C10 (upfloater opponents):** With the exception of the last two rounds (when `TotalRounds` is set and `CurrentRound >= TotalRounds - 1`), minimises the number of upfloaters whose opponent was a floater in the previous round (Art. 2.3.7).

C8 strictly dominates C9, and C9 strictly dominates C10. These quality criteria do not filter candidate pairs; rather, they are minimised via weighted matching on the chosen bracket. Colour preferences are not an absolute restriction (only C1 and C2 are absolute, per Art. 2.1 and Preface).

Brackets are built from the highest unpaired scoregroup downwards under C.04.6 Art. 3.3.2, keeping every resident in its bracket. Upfloater sets are chosen by C4, C5, C6, C7, and Article 3.5.4 order; set choice does not use C8-C10 (minimisation applies only to the chosen set). Each set must leave a C1/C2 completion for the remaining teams. Except in the last two rounds, C7 prefers teams that were not floaters in the preceding round. Failed brackets are never pooled; Art. 3.3.3 reports an impossible round.

### 5. Colour Allocation (9-Step)

`AllocateColor` implements the nine-step colour allocation procedure of Art. 4:

1. **Determine the first-team** (Art. 4.2): Higher primary score, then higher secondary score (if available), then smaller TPN.
2. **No history** (Art. 4.3.1): If neither team has played a match, assign by TPN parity. Odd TPN gets the initial colour; even TPN gets the opposite.
3. **One preference** (Art. 4.3.2): If only one team has a colour preference, grant it.
4. **Opposite preferences** (Art. 4.3.3): If both teams have opposite preferences, grant both.
5. **Strong vs non-strong** (Art. 4.3.4, Type B only): If only one team has a strong preference, grant it.
6. **Colour difference** (Art. 4.3.5): The team with the lower colour difference (fewer whites minus blacks) gets White.
7. **Alternation** (Art. 4.3.6): Remove unplayed matches, align both histories at their most recent played match, then find the most recent opposing colours and alternate.
8. **First-team preference** (Art. 4.3.7): Grant the first-team's colour preference.
9. **Last colour alternation** (Art. 4.3.8-9): Alternate from the first-team's last played colour; if still tied, alternate from the other team's last colour.

Colour is determined by the first board assignment (Art. 1.6.1): whichever team gets White on board 1 is considered "White" for the match. `MatchData.HomeID` is the team that had White on board 1 (C.04.6 1.6.1).

## Colour Preference Types

### Type A (Simple)

A team has a preference for White if its colour difference is below -1, or if CD is 0 or -1 and the last two played matches were both Black. The symmetric rule applies for Black. Otherwise, no preference. With fewer than two played matches only the colour-difference condition applies; Art. 4.3.5 then decides. This is a binary system: the team either has a preference or it does not.

### Type B (Strong + Mild)

Type B adds a second tier. Strong preferences use the same conditions as Type A. Mild preferences apply when:

- CD is -1: mild preference for White.
- CD is 0 and it is not the last round and the last played match was Black: mild preference for White.
- Symmetric rules apply for Black.

In the last round, mild preferences disappear (CD = 0 produces no preference). The strong/mild distinction affects step 5 of colour allocation: a strong preference takes priority over a non-strong one.

## Comparison

| Aspect                  | Team Swiss             | Double-Swiss       | Dutch                 |
| ----------------------- | ---------------------- | ------------------ | --------------------- |
| Participants            | Teams                  | Individual players | Individual players    |
| Matching algorithm      | Identifier order + weighted matching for C8-C10 | Identifier order + weighted matching for C8 | Global Blossom        |
| Colour allocation steps | 9                      | 5                  | FIDE C.04.3 rules     |
| Colour preference types | 3 (A, B, None)         | N/A                | N/A                   |
| Primary score options   | Match or game points   | Game points        | Standard points       |
| Criteria relaxation     | Last 2 rounds (C7/C10) | Last round (C8)    | No special relaxation |
| Default `topSeedColor`  | `"white"`              | `"auto"`           | `"auto"`              |
| Shared infrastructure   | `pairing/lexswiss`     | `pairing/lexswiss` | `pairing/swisslib`    |

## Mathematical Foundations

### Colour Difference

The colour difference (CD) for a team is defined as:

```text
CD = (number of White assignments) - (number of Black assignments)
```

Rounds where the team had no game (bye, absence) are excluded. CD drives the preference computation: negative CD pushes toward White, positive CD pushes toward Black.

### First-Team Ordering

The first-team determination in Art. 4.2 creates a strict total order for each pair:

```text
first(a, b) = a  if  score(a) > score(b)
            = a  if  score(a) == score(b) and secondary(a) > secondary(b)
            = a  if  scores equal and TPN(a) < TPN(b)
```

This ordering ensures deterministic colour allocation when all other tiebreakers are exhausted.

### Lexicographic Pairing

The Article 3.6 identifier enumeration is identical to Double-Swiss. See the [Double-Swiss](../double-swiss/#lexicographic-enumeration) mathematical foundations for the formal description.

## FIDE Reference

- **Regulation**: FIDE C.04.6 (Team Swiss System)
- **Adopted**: October 2025
- **Effective**: February 2026
- **Key articles**: Art. 1.2 (primary/secondary score), Art. 1.6.1 (colour by first board), Art. 1.7 (colour preference types), Art. 3.4.1-3.4.4 (PAB assignment), Art. 3.5 (upfloater selection), Art. 3.6 (lexicographic bracket pairing), Art. 4 (9-step colour allocation)

## Recording Team Byes

Copy `PairingResult.TeamByes` to `RoundData.TeamByes`. This is the preferred
team-bye channel. For compatibility, the team pairer also recognizes a
`RoundData.Byes` entry as a team bye when its ID identifies a team.

## Limitations

- The pairer does not derive the team set from `Matches`.
- The TRF converter does not derive the team set from `Matches`; teams appearing only in Matches get no 013/801/802 records.
- The TRF converter does not write real board-1 colours for 801/802 records. Totals-only matches cannot provide individual 001 opponents.
- The `primaryScore` option is not automatically merged between the pairer and scorer configurations.
- Tie-break PAB handling for Art. 12 TBR/BBE is incomplete.
- The scorer does not read `Games` for states without `Matches`.
