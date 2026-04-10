---
title: "tiebreakers"
linkTitle: "tiebreakers"
weight: 7
description: "Toon alle beschikbare tiebreaker-implementaties."
---

Het `tiebreakers`-subcommando toont elke geregistreerde tiebreaker-identifier samen met de weergavenaam.

## Synopsis

```text
chesspairing tiebreakers [options]
```

Geen argumenten of indelingssysteemvlag vereist.

## Opties

| Vlag     | Type | Standaard | Beschrijving           |
| -------- | ---- | --------- | ---------------------- |
| `--json` | bool | `false`   | Uitvoer als JSON-array |
| `--help` | --   | --        | Hulp tonen             |

## Tekstuitvoer

De standaarduitvoer is een tab-uitgelijnde tabel met twee kolommen, alfabetisch gesorteerd op ID:

```text
aro                      Average Rating of Opponents
avg-opponent-buchholz    Average Opponent Buchholz
avg-opponent-ptp         Average Opponent PTP
avg-opponent-tpr         Average Opponent TPR
black-games              Games with Black
black-wins               Black Wins
buchholz                 Buchholz
buchholz-cut1            Buchholz Cut 1
buchholz-cut2            Buchholz Cut 2
buchholz-median          Buchholz Median
buchholz-median2         Buchholz Median 2
direct-encounter         Direct Encounter
fore-buchholz            Fore Buchholz
games-played             Games Played
koya                     Koya System
pairing-number           Pairing Number
performance-points       Performance Points
performance-rating       Performance Rating
player-rating            Player Rating
progressive              Progressive Score
rounds-played            Rounds Played
sonneborn-berger         Sonneborn-Berger
standard-points          Standard Points
win                      Rounds Won
wins                     Games Won
```

## JSON-uitvoer

Met `--json` is de uitvoer een JSON-array van objecten met 2-spatie-inspringing:

```json
[
  {
    "id": "aro",
    "name": "Average Rating of Opponents"
  },
  {
    "id": "buchholz",
    "name": "Buchholz"
  }
]
```

Elk object heeft twee velden:

| Veld   | Type   | Beschrijving              |
| ------ | ------ | ------------------------- |
| `id`   | string | Tiebreaker-registratie-ID |
| `name` | string | Leesbare weergavenaam     |

## Exitcodes

| Code | Betekenis         |
| ---- | ----------------- |
| 0    | Gelukt            |
| 2    | JSON-encodingfout |
| 3    | Vlagparseerfout   |

## Zie ook

- [Tiebreakers](/docs/tiebreakers/) -- gedetailleerde beschrijvingen van elk tiebreaker-algoritme
- [standings](../standings/) -- bereken de stand met tiebreakers
