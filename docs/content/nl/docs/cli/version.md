---
title: "version"
linkTitle: "version"
weight: 9
description: "Toon de chesspairing-versie en ondersteunde engines."
---

Het `version`-subcommando toont de buildversie, ondersteunde indelings- en scoringssystemen en het aantal beschikbare tiebreakers. `chesspairing --version` is equivalent aan `chesspairing version`.

## Synopsis

```text
chesspairing version [options]
```

Geen argumenten vereist.

## Opties

| Vlag     | Type | Standaard | Beschrijving                                           |
| -------- | ---- | --------- | ------------------------------------------------------ |
| `--json` | bool | `false`   | Uitvoer als JSON (inclusief volledige tiebreakerlijst) |
| `--help` | --   | --        | Hulp tonen                                             |

## Tekstuitvoer

```text
chesspairing dev

Pairing systems:  dutch, burstein, dubov, lim, doubleswiss, team, keizer, roundrobin
Scoring systems:  standard, keizer, football
Tiebreakers:      25 available
```

De versiestring wordt bij het bouwen ingesteld via linker-vlaggen. Bij ontwikkelingsbuilds is de standaardwaarde `dev`.

## JSON-uitvoer

Met `--json` bevat de uitvoer de volledige lijst met tiebreaker-ID's:

```json
{
  "version": "dev",
  "pairingSystems": [
    "dutch",
    "burstein",
    "dubov",
    "lim",
    "doubleswiss",
    "team",
    "keizer",
    "roundrobin"
  ],
  "scoringSystems": ["standard", "keizer", "football"],
  "tiebreakers": [
    "aro",
    "avg-opponent-buchholz",
    "avg-opponent-ptp",
    "avg-opponent-tpr",
    "black-games",
    "black-wins",
    "buchholz",
    "buchholz-cut1",
    "buchholz-cut2",
    "buchholz-median",
    "buchholz-median2",
    "direct-encounter",
    "fore-buchholz",
    "games-played",
    "koya",
    "pairing-number",
    "performance-points",
    "performance-rating",
    "player-rating",
    "progressive",
    "rounds-played",
    "sonneborn-berger",
    "standard-points",
    "win",
    "wins"
  ]
}
```

## Exitcodes

| Code | Betekenis         |
| ---- | ----------------- |
| 0    | Gelukt            |
| 2    | JSON-encodingfout |
| 3    | Vlagparseerfout   |

## Zie ook

- [tiebreakers](../tiebreakers-cmd/) -- toon alle tiebreaker-ID's met weergavenamen
