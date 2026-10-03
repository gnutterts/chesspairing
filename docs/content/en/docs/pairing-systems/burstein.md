---
title: "Burstein System"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 Burstein Swiss pairing system."
---

Burstein follows FIDE C.04.4.2 (effective 1 February 2026).

The first `min(floor(totalRounds / 2), 4)` rounds are Dutch seeding rounds
(Article 1.6), paired by the Dutch pairer itself, including its colour
allocation. Afterwards players are grouped by pairing score and brackets are
processed from high to low (Article 1.9). Within a bracket they are ranked by
Buchholz, Sonneborn-Berger, and fixed TPN (Articles 1.7 and 1.8).

A pairing-allocated bye is assigned before brackets: eligible candidates are
considered by lowest score, most over-the-board games, then lowest ranking, and
the remaining players must be completely pairable (Article 3.1). `ForbiddenPairs`
is a library option, not part of C1; it also applies when testing whether a
candidate leaves a complete pairing. Brackets avoid rematches and matching
absolute colour preferences (C1 and C3). They choose floater sets under C5--C8,
then pairs in the Article 4.3 order (Articles 3.2 and 4); this is bracket
pairing, not a global Blossom matching.

For C7, the quality of the next bracket is the number of pairs it can still
make and, then, the scores of the floaters it would send down, with C5 and C6
satisfied in both comparisons. Brackets of at most ten players enumerate the
Article 4.3 order. Larger brackets enumerate floater sets and fix the Article
4.3 partners for each floater set; ties between floater sets are broken by that
same order. A bracket whose enumeration of floater sets exceeds 200000
candidates fails with `ErrBracketTooLarge` instead of silently skipping C5/C7.

The Article 1.7.2 treatment is an interpretation. An unplayed round (a bye, a
forfeit, or a round without a record) counts as played against the player
himself with the registered points. The benefit of a series of consecutive
zero-point byes that ends at the last completed round goes only to the player's
actual over-the-board opponents. In those opponents' Buchholz and
Sonneborn-Berger, the player's score is 0.5 higher for each bye in the series;
rounds without any record count as zero-point byes in the series. In the
player's own index, including unplayed rounds treated as games against himself,
the registered score is used without the extra half-points. Virtual acceleration
points are excluded from the index, although they determine pairing
scoregroups.

Colours follow Articles 5.2.1--5.2.5, rather than the Dutch colour cascade.
For two players without played games, 5.2.1 gives the higher-ranked player the
initial colour; `TopSeedColor` sets that colour after the seeding rounds as
well as in them. The parity is the TPN parity among all players who have entered
the tournament, including the player who receives the bye in this round. The
remaining rules consider compatible preferences, preference strength and colour
difference, the most recent opposite colours, and ranking. The Article 5.2.4
"most recent time" is read over played games only, aligned from the most recent
game backwards, as in the Dutch pairer and General Handling Rules 3.4; this
reading is not independently verified.

## Configuration

### CLI

```bash
chesspairing pair --burstein tournament.trf
```

### Go API

```go
p := burstein.New(burstein.Options{TotalRounds: chesspairing.IntPtr(9)})
result, err := p.Pair(ctx, &state)
```

### Options

| Option | Description |
| --- | --- |
| `TotalRounds` | Planned rounds; determines the Article 1.6 seeding phase. |
| `Acceleration` | Optional `"baku"` acceleration. |
| `TopSeedColor` | Initial top-seed colour in seeding rounds and for Article 5.2.1 afterwards. |
| `ForbiddenPairs` | Library-configured player-ID pairs that must not be paired; not a C1 rule. |

### Errors

| Error | Condition |
| --- | --- |
| `ErrTooFewPlayers` | The post-seeding field is empty (unless it consists only of pre-assigned byes). A single remaining player receives the bye. |
| `PairingNoPABCandidate` | Returned in a `chesspairing.PairingError` when no player may receive the pairing-allocated bye. |
| `ErrNoPairingPossible` | Returned in a `chesspairing.PairingError` of kind `PairingImpossible` when no complete pairing exists. |
| `ErrBracketTooLarge` | Returned in a `chesspairing.PairingError` of kind `PairingImpossible` when a bracket's floater-set enumeration exceeds 200000 candidates. |

bbpPairings' Burstein mode is not used as a reference: it has no seeding
rounds, uses a different index order, treats unplayed games as draws, omits the
most-games step in bye choice, and merges outgoing floaters; it therefore
deviates from the published text. There is no FIDE-endorsed program or public
set of verified Burstein pairings to compare with.
