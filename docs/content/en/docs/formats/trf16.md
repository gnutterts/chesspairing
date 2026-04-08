---
title: "TRF16 Format"
linkTitle: "TRF16"
weight: 1
description: "The FIDE Tournament Report File format — line types, player records, and round results."
---

## Overview

TRF16 is a fixed-width text format used by FIDE for tournament data exchange. Each line begins with a 3-character code identifying its type, followed by a space, then the data. Lines are separated by newlines. The format is designed for portability across pairing engines.

The `trf` package (`trf/read.go`, `trf/write.go`) provides a complete reader and writer that faithfully preserves the document structure, including unknown line codes for round-trip fidelity.

## Header lines

Header lines carry tournament metadata. Each code maps to a single field:

| Code  | Field                   | Example                         |
| ----- | ----------------------- | ------------------------------- |
| `012` | Tournament name         | `012 World Championship 2024`   |
| `022` | City                    | `022 Budapest`                  |
| `032` | Federation              | `032 HUN`                       |
| `042` | Start date              | `042 2024/11/01`                |
| `052` | End date                | `052 2024/11/30`                |
| `062` | Number of players       | `062 14`                        |
| `072` | Number of rated players | `072 14`                        |
| `082` | Number of teams         | `082 0`                         |
| `092` | Tournament type         | `092 Swiss Dutch`               |
| `102` | Chief arbiter           | `102 IA FirstName LastName`     |
| `112` | Deputy arbiter          | `112 FA FirstName LastName`     |
| `122` | Time control            | `122 90/40+30+30`               |
| `132` | Round dates             | `132 2024/11/01 2024/11/02 ...` |

Multiple `112` lines are supported (TRF-2026 allows multiple deputy arbiters). The first `112` line populates the `DeputyArbiter` field; all `112` lines are collected into `DeputyArbiters`.

Multiple `132` lines are appended to the `RoundDates` slice.

The `092` tournament type is used by `ToTournamentState()` to infer the pairing system. Recognized values:

| Tournament type      | Pairing system |
| -------------------- | -------------- |
| `Swiss Dutch`        | `dutch`        |
| `Swiss Burstein`     | `burstein`     |
| `Swiss Dubov`        | `dubov`        |
| `Swiss Lim`          | `lim`          |
| `Double Swiss`       | `doubleswiss`  |
| `Team Swiss`         | `team`         |
| `Round Robin`        | `roundrobin`   |
| `Double Round Robin` | `roundrobin`   |
| `Keizer`             | `keizer`       |

Unrecognized types default to `dutch`.

## Player lines (001)

Player lines use a fixed-width column layout. Byte positions are 0-indexed:

```text
001 SSSS SEX TTT NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN RRRR FED FFFFFFFFFFF BBBBBBBBBB PPPP RRRR  round-results...
```

| Byte range | Field         | Width        | Alignment | Description                   |
| ---------- | ------------- | ------------ | --------- | ----------------------------- |
| 0-2        | Line code     | 3            |           | Always `001`                  |
| 4-7        | Start number  | 4            | Right     | Tournament pairing number     |
| 9          | Sex           | 1            |           | `m` or `w`                    |
| 10-12      | Title         | 3            | Right     | `GM`, `IM`, `FM`, `WGM`, etc. |
| 14-46      | Name          | 33           | Left      | `LastName, FirstName`         |
| 48-51      | Rating        | 4            | Right     | FIDE rating                   |
| 53-55      | Federation    | 3            | Left      | 3-letter code                 |
| 57-67      | FIDE ID       | 11           | Left      | FIDE player number            |
| 69-78      | Birth date    | 10           | Left      | `YYYY/MM/DD`                  |
| 80-83      | Points        | 4            | Right     | Total points (e.g. ` 5.5`)    |
| 85-88      | Rank          | 4            | Right     | Current rank                  |
| 89+        | Round results | 10 per round |           | See below                     |

The minimum line length is 84 characters (through the points field). The parser requires at least 84 bytes.

## Round result format

Each round occupies exactly 10 characters within the player line, starting at byte 89:

```text
  OOOO C R
```

| Byte offset | Field    | Description                                              |
| ----------- | -------- | -------------------------------------------------------- |
| 0-1         | Padding  | Two spaces                                               |
| 2-5         | Opponent | 4-digit start number, zero-padded (`0000` = no opponent) |
| 6           | Space    | Separator                                                |
| 7           | Color    | `w` (White), `b` (Black), or `-` (no color / bye)        |
| 8           | Space    | Separator                                                |
| 9           | Result   | Result code character                                    |

Example: `  0012 w 1` means opponent 12, playing White, win.

## Result codes

| Code | Constant              | Meaning                     |
| ---- | --------------------- | --------------------------- |
| `1`  | `ResultWin`           | Win (played)                |
| `0`  | `ResultLoss`          | Loss (played)               |
| `=`  | `ResultDraw`          | Draw                        |
| `+`  | `ResultForfeitWin`    | Win by forfeit              |
| `-`  | `ResultForfeitLoss`   | Loss by forfeit             |
| `H`  | `ResultHalfBye`       | Half-point bye              |
| `F`  | `ResultFullBye`       | Full-point bye (PAB)        |
| `U`  | `ResultUnpaired`      | Unpaired (absent, 0 points) |
| `Z`  | `ResultZeroBye`       | Zero-point bye              |
| `*`  | `ResultNotPlayed`     | Not yet played              |
| `W`  | `ResultWinByDefault`  | Win, opponent absent        |
| `D`  | `ResultDrawByDefault` | Draw by default             |
| `L`  | `ResultLossByDefault` | Loss by default             |

Bye results (`H`, `F`, `U`, `Z`) have opponent `0000` and color `-`.

When converting to `TournamentState`, bye results create `ByeEntry` records:

| TRF code | ByeType     |
| -------- | ----------- |
| `F`      | `ByePAB`    |
| `H`      | `ByeHalf`   |
| `Z`      | `ByeZero`   |
| `U`      | `ByeAbsent` |

## Extended data lines (XX fields)

TRF16 defines several extension codes for data not covered by the base specification:

| Code  | Field                    | Type     | Example      |
| ----- | ------------------------ | -------- | ------------ |
| `XXR` | Total rounds             | int      | `XXR 9`      |
| `XXC` | Initial color assignment | string   | `XXC white1` |
| `XXS` | Acceleration data        | string   | `XXS baku`   |
| `XXP` | Forbidden pairs          | int pair | `XXP 5 12`   |

`XXR` and `XXC` have TRF-2026 replacements (`142` and `152` respectively). The `trf` package provides `EffectiveTotalRounds()` and `EffectiveInitialColor()` methods that check TRF-2026 fields first, falling back to TRF16 XX fields.

Multiple `XXS` and `XXP` lines are supported. Each `XXP` line specifies a pair of start numbers that must not be paired against each other.

System-specific XX fields (`XXY`, `XXB`, `XXM`, `XXT`, `XXG`, `XXA`, `XXK`) are documented in [TRF-2026 Extensions](../trf-extensions/).

## Team lines (013)

Team lines define team composition:

```text
013 SSSS NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN MMMM MMMM ...
```

| Byte range | Field       | Width  | Description                        |
| ---------- | ----------- | ------ | ---------------------------------- |
| 0-2        | Line code   | 3      | Always `013`                       |
| 4-7        | Team number | 4      | Right-aligned                      |
| 8-39       | Team name   | 32     | Left-aligned                       |
| 40+        | Members     | 4 each | Whitespace-separated start numbers |

The minimum line length is 40 characters.

## Comment lines

Lines starting with `###` are comments and are ignored by the parser:

```text
### This is a comment
```

Comments are stored in the `Comments` slice and written back during serialization.

## Unknown lines

Lines with unrecognized 3-character codes are preserved as `RawLine` entries in the `Other` slice. Each `RawLine` stores the code and data separately. During write-back, these lines are reproduced as `CODE DATA`, ensuring no data is lost during read/write cycles.

## National Rating System records (NRS)

Lines starting with a 3-letter uppercase code that is not an `XX` prefix and that follow the `001` column layout (at least 68 characters with a numeric start number at bytes 4-7) are recognized as National Rating System records. These carry national ratings and sub-federation data alongside the standard player fields. NRS records are stored in `NRSRecords` and written back using their raw line data.

## Parse errors

Parse errors include the 1-based line number and line code for diagnostic context:

```text
trf: line 42 (001): invalid rating: "XXXX"
```

The parser returns on the first error encountered.
