---
title: "JSON Schemas"
linkTitle: "JSON"
weight: 3
description: "JSON output schemas for pairings, standings, validation, version, and tiebreaker listing."
---

The CLI produces JSON output for several subcommands when invoked with `--json` or `--format json`. All JSON output uses 2-space indentation.

## Pairings

Command: `chesspairing pair SYSTEM input-file --format json`

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 12 }
  ],
  "byes": [{ "player": 7, "type": "PAB" }]
}
```

### Fields

| Field              | Type   | Description                        |
| ------------------ | ------ | ---------------------------------- |
| `pairings`         | array  | Board assignments for the round    |
| `pairings[].board` | int    | Board number (1-indexed)           |
| `pairings[].white` | int    | White player's start number        |
| `pairings[].black` | int    | Black player's start number        |
| `byes`             | array  | Bye assignments (omitted if empty) |
| `byes[].player`    | int    | Player's start number              |
| `byes[].type`      | string | Bye type string                    |

Bye type values correspond to `ByeType.String()`:

| Value            | Meaning                            |
| ---------------- | ---------------------------------- |
| `PAB`            | Pairing-Allocated Bye (full point) |
| `Half`           | Half-point bye                     |
| `Zero`           | Zero-point bye                     |
| `Absent`         | Absent/unpaired                    |
| `Excused`        | Excused absence                    |
| `ClubCommitment` | Club commitment absence            |

## Standings

Command: `chesspairing standings SYSTEM input-file --json`

```json
{
  "standings": [
    {
      "rank": 1,
      "playerId": "5",
      "displayName": "GM Smith, John",
      "score": 7,
      "tieBreakers": [
        { "id": "buchholz-cut1", "name": "Buchholz Cut 1", "value": 28.5 },
        { "id": "buchholz", "name": "Buchholz", "value": 35 },
        { "id": "sonneborn-berger", "name": "Sonneborn-Berger", "value": 24.5 }
      ],
      "gamesPlayed": 9,
      "wins": 6,
      "draws": 2,
      "losses": 1
    }
  ],
  "scoring": "standard",
  "tiebreakers": ["buchholz-cut1", "buchholz", "sonneborn-berger"]
}
```

### Fields

| Field                             | Type   | Description                                            |
| --------------------------------- | ------ | ------------------------------------------------------ |
| `standings`                       | array  | Ranked player entries                                  |
| `standings[].rank`                | int    | Rank (shared ranks for tied players)                   |
| `standings[].playerId`            | string | Player ID (start number as string)                     |
| `standings[].displayName`         | string | Player display name                                    |
| `standings[].score`               | float  | Total score from the scoring engine                    |
| `standings[].tieBreakers`         | array  | Tiebreaker values in configured order                  |
| `standings[].tieBreakers[].id`    | string | Tiebreaker registry ID                                 |
| `standings[].tieBreakers[].name`  | string | Tiebreaker display name                                |
| `standings[].tieBreakers[].value` | float  | Computed tiebreaker value                              |
| `standings[].gamesPlayed`         | int    | Total games played                                     |
| `standings[].wins`                | int    | Number of wins                                         |
| `standings[].draws`               | int    | Number of draws                                        |
| `standings[].losses`              | int    | Number of losses                                       |
| `scoring`                         | string | Scoring system used (`standard`, `keizer`, `football`) |
| `tiebreakers`                     | array  | Ordered list of tiebreaker IDs applied                 |

Standings are sorted by score descending, then by tiebreaker values in order. Players with identical scores and tiebreaker values share the same rank.

## Validation

Command: `chesspairing validate input-file --json`

```json
{
  "valid": false,
  "errors": [
    {
      "field": "player.3.rating",
      "severity": "error",
      "message": "rating must be a positive integer"
    }
  ],
  "warnings": [
    {
      "field": "player.5.title",
      "severity": "warning",
      "message": "title field is empty"
    }
  ],
  "profile": "standard"
}
```

### Fields

| Field                 | Type   | Description                                   |
| --------------------- | ------ | --------------------------------------------- |
| `valid`               | bool   | `true` if no errors (warnings are acceptable) |
| `errors`              | array  | Error-level issues (null if none)             |
| `warnings`            | array  | Warning-level issues (null if none)           |
| `errors[].field`      | string | Dot-separated field path                      |
| `errors[].severity`   | string | Always `"error"`                              |
| `errors[].message`    | string | Human-readable description                    |
| `warnings[].field`    | string | Dot-separated field path                      |
| `warnings[].severity` | string | Always `"warning"`                            |
| `warnings[].message`  | string | Human-readable description                    |
| `profile`             | string | Validation profile used                       |

## Check

Command: `chesspairing check SYSTEM input-file --json`

```json
{
  "match": true,
  "system": "dutch",
  "round": 5
}
```

### Fields

| Field    | Type   | Description                                                  |
| -------- | ------ | ------------------------------------------------------------ |
| `match`  | bool   | `true` if regenerated pairings match the existing last round |
| `system` | string | Pairing system used for re-pairing                           |
| `round`  | int    | Round number that was checked                                |

The `check` subcommand strips the last round, regenerates pairings, and compares them. Exit code `0` indicates a match; exit code `1` indicates a mismatch.

## Version

Command: `chesspairing version --json`

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

### Fields

| Field            | Type   | Description                                           |
| ---------------- | ------ | ----------------------------------------------------- |
| `version`        | string | Build version string                                  |
| `pairingSystems` | array  | All supported pairing system identifiers              |
| `scoringSystems` | array  | All supported scoring system identifiers              |
| `tiebreakers`    | array  | All registered tiebreaker IDs (sorted alphabetically) |

## Tiebreakers

Command: `chesspairing tiebreakers --json`

```json
{
  "tiebreakers": [
    { "id": "aro", "name": "Average Rating of Opponents" },
    { "id": "buchholz", "name": "Buchholz" },
    { "id": "buchholz-cut1", "name": "Buchholz Cut 1" }
  ]
}
```

### Fields

The output is a JSON object with a single `tiebreakers` array.

| Field                | Type   | Description                                           |
| -------------------- | ------ | ----------------------------------------------------- |
| `tiebreakers`        | array  | All registered tiebreakers                            |
| `tiebreakers[].id`   | string | Tiebreaker registry ID (used in `--tiebreakers` flag) |
| `tiebreakers[].name` | string | Human-readable display name                           |

Entries are sorted alphabetically by ID.
