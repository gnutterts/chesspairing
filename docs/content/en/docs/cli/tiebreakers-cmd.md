---
title: "tiebreakers"
linkTitle: "tiebreakers"
weight: 7
description: "List all available tiebreaker implementations."
---

The `tiebreakers` subcommand prints every registered tiebreaker identifier along with its display name.

## Synopsis

```text
chesspairing tiebreakers [options]
```

No arguments or pairing system flag required.

## Options

| Flag     | Type | Default | Description           |
| -------- | ---- | ------- | --------------------- |
| `--json` | bool | `false` | Output as JSON object |
| `--help` | --   | --      | Show usage help       |

## Text output

The default output is a tab-aligned two-column table sorted alphabetically by ID:

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

## JSON output

With `--json`, the output is a JSON object wrapping an array of tiebreaker entries, with 2-space indentation:

```json
{
  "tiebreakers": [
    {
      "id": "aro",
      "name": "Average Rating of Opponents"
    },
    {
      "id": "buchholz",
      "name": "Buchholz"
    }
  ]
}
```

| Field                | Type   | Description                 |
| -------------------- | ------ | --------------------------- |
| `tiebreakers`        | array  | All registered tiebreakers  |
| `tiebreakers[].id`   | string | Tiebreaker registration ID  |
| `tiebreakers[].name` | string | Human-readable display name |

## Exit codes

| Code | Meaning             |
| ---- | ------------------- |
| 0    | Success             |
| 2    | JSON encoding error |
| 3    | Flag parse error    |

## See also

- [Tiebreakers](/docs/tiebreakers/) -- detailed descriptions of each tiebreaker algorithm
- [standings](../standings/) -- compute standings with tiebreakers applied
