---
title: "version"
linkTitle: "version"
weight: 9
description: "Display the chesspairing version and supported engines."
---

The `version` subcommand prints the build version, supported pairing and scoring systems, and the number of available tiebreakers. Running `chesspairing --version` is equivalent to `chesspairing version`.

## Synopsis

```text
chesspairing version [options]
```

No arguments required.

## Options

| Flag     | Type | Default | Description                                    |
| -------- | ---- | ------- | ---------------------------------------------- |
| `--json` | bool | `false` | Output as JSON (includes full tiebreaker list) |
| `--help` | --   | --      | Show usage help                                |

## Text output

```text
chesspairing dev

Pairing systems:  dutch, burstein, dubov, lim, doubleswiss, team, keizer, roundrobin
Scoring systems:  standard, keizer, football
Tiebreakers:      25 available
```

The version string is set at build time via linker flags. During development builds, it defaults to `dev`.

## JSON output

With `--json`, the output includes the full list of tiebreaker IDs:

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

## Exit codes

| Code | Meaning             |
| ---- | ------------------- |
| 0    | Success             |
| 2    | JSON encoding error |
| 3    | Flag parse error    |

## See also

- [tiebreakers](../tiebreakers-cmd/) -- list all tiebreaker IDs with display names
