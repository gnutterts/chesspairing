---
title: "TRF-2026 Extensions"
linkTitle: "Extensions"
weight: 2
description: "System-specific XX fields and TRF-2026 record types for pairing engine configuration."
---

The `trf` package supports both TRF16 legacy extensions (XX-prefixed fields) and the newer TRF-2026 record types. This page documents all extension fields and data records.

## System-specific XX fields (TRF16 legacy)

These fields carry per-engine configuration in the TRF16 format. Each is a single line with the XX code followed by a space and the value.

| Code  | Field                   | Type   | Used by     | Description                                              |
| ----- | ----------------------- | ------ | ----------- | -------------------------------------------------------- |
| `XXY` | Cycles                  | int    | Round-Robin | Number of cycles: `1` = single, `2` = double round-robin |
| `XXB` | ColorBalance            | bool   | Round-Robin | Enable color balancing (`true` or `false`)               |
| `XXM` | MaxiTournament          | bool   | Lim         | Enable maxi-tournament mode (`true` or `false`)          |
| `XXT` | ColorPreferenceType     | string | Team        | Color preference algorithm: `A`, `B`, or `none`          |
| `XXG` | PrimaryScore            | string | Team        | Primary scoring metric: `match` or `game`                |
| `XXA` | AllowRepeatPairings     | bool   | Keizer      | Allow repeat pairings (`true` or `false`)                |
| `XXK` | MinRoundsBetweenRepeats | int    | Keizer      | Minimum rounds between rematches                         |

Examples:

```text
XXY 2
XXB true
XXM false
XXT A
XXG match
XXA true
XXK 3
```

These fields are mapped to pairing option keys during `ToTournamentState()` conversion:

| XX code | Option key                |
| ------- | ------------------------- |
| `XXY`   | `cycles`                  |
| `XXB`   | `colorBalance`            |
| `XXM`   | `maxiTournament`          |
| `XXT`   | `colorPreferenceType`     |
| `XXG`   | `primaryScore`            |
| `XXA`   | `allowRepeatPairings`     |
| `XXK`   | `minRoundsBetweenRepeats` |

## Common XX fields

In addition to the system-specific fields, TRF16 defines general-purpose extension codes:

| Code  | Field          | Type     | Description                                                   |
| ----- | -------------- | -------- | ------------------------------------------------------------- |
| `XXR` | TotalRounds    | int      | Total number of rounds planned                                |
| `XXC` | InitialColor   | string   | Initial color assignment (e.g. `white1`)                      |
| `XXS` | Acceleration   | string   | Baku acceleration data (one line per entry)                   |
| `XXP` | ForbiddenPairs | int pair | Two start numbers that must not be paired (one pair per line) |

`XXR` maps to `totalRounds`, `XXC` maps to `topSeedColor`, and `XXS` sets `acceleration` to `"baku"` in the pairing options. Multiple `XXS` and `XXP` lines are accumulated.

## TRF-2026 header fields

TRF-2026 introduces new header codes that replace or extend TRF16 fields:

| Code  | Field               | Type   | Description                                                          |
| ----- | ------------------- | ------ | -------------------------------------------------------------------- |
| `142` | TotalRounds26       | int    | Total number of rounds (replaces `XXR`)                              |
| `152` | InitialColor26      | string | Initial color assignment: `B` or `W` (replaces `XXC`)                |
| `162` | ScoringSystem       | string | Scoring algorithm (e.g. `W 1.0    D 0.5    L 0.0`)                   |
| `172` | StartingRankMethod  | string | How start numbers are assigned (e.g. `IND FIDE`)                     |
| `192` | CodedTournamentType | string | Machine-readable tournament type (e.g. `FIDE_TEAM_BAKU`)             |
| `202` | TieBreakDef         | string | Tiebreaker configuration (e.g. `EDET/P,EMGSB/C1/P,BH:MP/C1/P`)       |
| `222` | EncodedTimeControl  | string | Machine-readable time control (e.g. `40/6000+30:20/3000+30:1500+30`) |
| `352` | TeamInitialColor    | string | Team color assignment pattern (e.g. `WBWB`)                          |
| `362` | TeamScoringSystem   | string | Team scoring algorithm (e.g. `TW 2     TD 1     TL 0`)               |

## TRF-2026 data records

### 240 -- Absence records

Declares absent players for a round.

Format: `240 T RRR TOI1 TOI2 ...`

| Field | Description                                              |
| ----- | -------------------------------------------------------- |
| `T`   | Absence type: `F` (full forfeit) or `H` (half-point bye) |
| `RRR` | Round number                                             |
| `TOI` | Start numbers of absent players/teams                    |

Example: `240 F 3 5 12 18`

Section 240 only encodes the two FIDE-defined absence letters. Richer
bye types (`zero`, `absent`, `excused`, `clubcommitment`) and player
withdrawals travel via chesspairing comment directives -- see below.

### chesspairing comment directives

Lines beginning with `### chesspairing:` carry data the FIDE TRF
formats cannot express directly. They live in the comment block, so
parsers that do not recognise them simply preserve the line verbatim.
Two verbs are currently defined.

`### chesspairing:bye round=N player=SN type=TYPE` declares a
pre-assigned bye for the upcoming round. The valid `type` values are
the lowercased `ByeType.String()` spellings:

| Value            | ByeType             |
| ---------------- | ------------------- |
| `pab`            | `ByePAB`            |
| `half`           | `ByeHalf`           |
| `zero`           | `ByeZero`           |
| `absent`         | `ByeAbsent`         |
| `excused`        | `ByeExcused`        |
| `clubcommitment` | `ByeClubCommitment` |

`### chesspairing:withdrawn player=SN after-round=N` records a
permanent withdrawal: the player is excluded from pairing for every
round strictly greater than `N`. `N` must be a positive integer.

On read, both verbs are bridged into the `TournamentState`:
chesspairing:bye entries become `PreAssignedByes` for the current
round, and chesspairing:withdrawn entries set
`PlayerEntry.WithdrawnAfterRound`. When a Section 240 record and a
chesspairing:bye directive name the same `(round, player)`, the
directive wins; this lets richer types override the FIDE-encoded
default. Unknown player IDs in either verb produce a validation
error rather than being silently dropped, as do non-positive
`after-round` values.

On write, `PreAssignedByes` whose type is not expressible in
Section 240 are emitted as chesspairing:bye directives, and any
player with a non-nil `WithdrawnAfterRound` produces a
chesspairing:withdrawn directive. Unknown verbs encountered on read
are preserved verbatim so files written by a future version of the
library are not silently rewritten by an older one.

Example:

```text
### chesspairing:bye round=4 player=12 type=excused
### chesspairing:withdrawn player=18 after-round=3
```

Stored on `Document.ChesspairingDirectives` as a slice of
`Directive{Verb, Params}`.

### 250 -- Acceleration records

Baku acceleration parameters (replaces `XXS`).

Format: `250 MMMM GGGG RRF RRL PPPF PPPL`

| Field  | Description                                |
| ------ | ------------------------------------------ |
| `MMMM` | Match points to add (for team tournaments) |
| `GGGG` | Game points to add                         |
| `RRF`  | First round of acceleration                |
| `RRL`  | Last round of acceleration                 |
| `PPPF` | First player/team number in range          |
| `PPPL` | Last player/team number in range           |

Raw line data is preserved for round-trip fidelity.

### 260 -- Forbidden pair records

Round-scoped forbidden pairing restrictions (replaces `XXP`).

Format: `260 RR1 RRL TOI1 TOI2 ...`

| Field | Description                               |
| ----- | ----------------------------------------- |
| `RR1` | First round of restriction                |
| `RRL` | Last round of restriction                 |
| `TOI` | Start numbers that are mutually forbidden |

Unlike `XXP` which only specifies a single pair, a `260` record lists multiple players who are all mutually forbidden. The `ToTournamentState()` conversion generates all pairwise combinations.

### 300 -- Team round data

Board assignments for team matches.

Format: `300 RRR TT1 TT2 PPP1 PPP2 PPP3 PPP4`

| Field | Description                                             |
| ----- | ------------------------------------------------------- |
| `RRR` | Round number                                            |
| `TT1` | First team number                                       |
| `TT2` | Second team number                                      |
| `PPP` | Player start numbers for each board (`0` = empty board) |

### 310 -- Team definition

TRF-2026 team records (replaces `013`). Fixed-width column layout:

| Byte range | Field          | Width  | Description          |
| ---------- | -------------- | ------ | -------------------- |
| 0-2        | Code           | 3      | Always `310`         |
| 4-6        | Team number    | 3      | Right-aligned        |
| 8-40       | Team name      | 33     | Left-aligned         |
| 41-45      | Federation     | 5      | Left-aligned         |
| 46-52      | Average rating | 7      | Right-aligned        |
| 53-58      | Match points   | 6      | Right-aligned        |
| 59-66      | Game points    | 8      | Right-aligned        |
| 67-70      | Rank           | 4      | Right-aligned        |
| 72+        | Members        | 4 each | Member start numbers |

### 320 -- Team round scores

Per-round team scores.

Format: `320 TTT GGGG RRR1 RRR2 ...`

| Field  | Description             |
| ------ | ----------------------- |
| `TTT`  | Team number             |
| `GGGG` | Total game points       |
| `RRR`  | Per-round score strings |

Raw line data is preserved for round-trip fidelity.

### 330 -- Old absent forfeits

Legacy absent/forfeit records for team tournaments.

Format: `330 TT RRR WWW BBB`

| Field | Description                      |
| ----- | -------------------------------- |
| `TT`  | Result code: `+-`, `-+`, or `--` |
| `RRR` | Round number                     |
| `WWW` | White team number                |
| `BBB` | Black team number                |

### 801 -- Detailed team results

Detailed per-board team match results. Contains team number, team name, match points, game points, and per-round data with opponent, color, individual board results, and board order.

Bye rounds use markers: `FFFF`, `HHHH`, `ZZZZ`, `UUUU`.

Raw line data is preserved for round-trip fidelity due to complex variable-width formatting.

### 802 -- Simple team results

Simplified team round results. Contains team number, team name, match points, game points, and per-round entries with opponent, color, and game points.

Bye rounds use markers: `FPB`, `HPB`, `ZPB`, `PAB` followed by game points.

A trailing `f` on game points indicates forfeit involvement.

Raw line data is preserved for round-trip fidelity.

### NRS -- National rating records

Lines starting with a 3-letter uppercase federation code (not `XX`-prefixed) that follow the `001` column layout are parsed as National Rating System records. These carry national ratings, sub-federation codes, and national IDs alongside the standard player fields.

NRS records are stored with their raw line data and written back unchanged.

## Fallback methods

The `trf.Document` type provides fallback accessors that check TRF-2026 fields first, then fall back to TRF16 legacy fields:

- `EffectiveTotalRounds()` -- returns `TotalRounds26` (code `142`) if set, otherwise `TotalRounds` (code `XXR`). Returns `0` if neither is set.
- `EffectiveInitialColor()` -- returns `InitialColor26` (code `152`) if set, otherwise `InitialColor` (code `XXC`). Returns `""` if neither is set.

These methods are used by `ToTournamentState()` and should be preferred over accessing the fields directly when both TRF16 and TRF-2026 data may be present.

## Round-trip fidelity

The `trf` package preserves all data during read/write cycles:

- Unknown line codes are stored as `RawLine` entries in the `Other` slice and written back as `CODE DATA`.
- TRF-2026 records with complex formatting (`250`, `260`, `320`, `801`, `802`) store the raw line data and use it during serialization when available.
- NRS records are stored and written back using their original raw lines.
- Comment lines (`###`) are stored in the `Comments` slice and reproduced during write.

This ensures that a read-then-write cycle produces an identical document for any data the parser does not structurally modify.
