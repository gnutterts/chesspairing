---
title: "Burstein System"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 Burstein Swiss pairing system."
---

Burstein follows FIDE C.04.4.2 (effective 1 February 2026).

The first `min(floor(totalRounds / 2), 4)` rounds are Dutch seeding rounds
(Article 1.6), with exactly the Dutch pairing and colour allocation. Afterwards
players are grouped by pairing score and brackets are processed from high to low
(Article 1.9). Within each bracket they are ranked by Buchholz,
Sonneborn-Berger, and fixed TPN (Articles 1.7 and 1.8).

A pairing-allocated bye is assigned before brackets: eligible candidates are
considered by lowest score, most over-the-board games, then lowest ranking, and
the remaining players must be completely pairable (Article 3.1). Brackets use
no rematches and no matching absolute colour preferences (C1 and C3), maximise
pairs, and choose floaters and pairs in the Article 4.3 order using C5--C8
(Articles 3.2 and 4).

Unplayed rounds, including forfeits, are evaluated as games against the player
themself for the index (Article 1.7.2). Virtual acceleration points determine
pairing scoregroups but do not enter that index. Colours follow Article 5.2:
the entered-player TPN parity settles a pair of players with no played games;
then preferences, the most recent opposite colours, and ranking are used.

`totalRounds`, `acceleration`, `topSeedColor`, and `forbiddenPairs` are
available through `burstein.Options`.

bbpPairings' Burstein mode is not used as a reference: it describes itself as
a flawed implementation of an earlier version and differs from this regulation.
