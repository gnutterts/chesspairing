---
title: "Output Formats and Exit Codes"
linkTitle: "Output & Exits"
weight: 11
description: "The five output formats (list, wide, board, xml, json) and six exit codes."
---

This page documents the five output formats supported by the `pair` subcommand and the exit codes returned by the CLI.

## Output formats

The `pair` subcommand accepts `--format` with one of five values: `list`, `wide`, `board`, `xml`, `json`. The format can also be selected with the `-w` shorthand (wide) or `--json` shorthand.

### list (default)

Machine-readable format compatible with bbpPairings and JaVaFo.

- First line: the number of board pairings (excluding byes).
- Subsequent lines: `white_startnum black_startnum`, one pair per line.
- Byes appended after all pairings as `startnum 0`.

```text
5
5 1
3 12
8 2
6 9
4 11
7 0
```

This format contains start numbers only -- no names, ratings, or board numbers.

### wide

Human-readable tabular format rendered with Go's `tabwriter`.

Columns: `Board`, `White`, `Rtg`, `-`, `Black`, `Rtg`.

Player display format: `TPN Title Name` where `TPN` is the tournament pairing number (start number). The title is omitted when the player has no title. The rating column is empty when the player's rating is nil or 0.

Byes are listed after the last board pairing without a board number.

```text
Board  White              Rtg   -  Black              Rtg
-----  -----              ---      -----              ---
1      5 GM Smith          2600  -  1 IM Jones         2500
2      3 WGM Lee           2400  -  12 FM Petrov       2350
       7 Brown             1800     Bye (PAB)
```

### board

Compact board-style view with dynamic field widths.

Format: `Board N: W - B` where field widths for board and player numbers adjust based on the largest values present. Byes are shown as `Bye: N`.

```text
Board  1:  5 -  1
Board  2:  3 - 12
Board  3:  8 -  2
Bye:  7
```

Player number field width is determined by the largest start number across all pairings and byes. Board number width is determined by the total number of boards.

### json

Structured JSON output with 2-space indentation.

```json
{
  "pairings": [
    {
      "board": 1,
      "white": 5,
      "black": 1
    },
    {
      "board": 2,
      "white": 3,
      "black": 12
    }
  ],
  "byes": [
    {
      "player": 7,
      "type": "PAB"
    }
  ]
}
```

Structure:

| Field              | Type   | Description                                                            |
| ------------------ | ------ | ---------------------------------------------------------------------- |
| `pairings`         | array  | Board pairings                                                         |
| `pairings[].board` | int    | 1-indexed board number                                                 |
| `pairings[].white` | int    | White player start number                                              |
| `pairings[].black` | int    | Black player start number                                              |
| `byes`             | array  | Bye assignments (omitted when empty)                                   |
| `byes[].player`    | int    | Player start number                                                    |
| `byes[].type`      | string | Bye type: `PAB`, `Half`, `Zero`, `Absent`, `Excused`, `ClubCommitment` |

The `byes` key uses `omitempty` and is absent from the output when there are no byes.

### xml

Full XML document including `xml.Header` (`<?xml version="1.0" encoding="UTF-8"?>`).

```xml
<?xml version="1.0" encoding="UTF-8"?>
<pairings round="4" boards="3" byes="1">
  <board number="1">
    <white number="5" name="Smith" rating="2600" title="GM"></white>
    <black number="1" name="Jones" rating="2500" title="IM"></black>
  </board>
  <board number="2">
    <white number="3" name="Lee" rating="2400" title="WGM"></white>
    <black number="12" name="Petrov" rating="2350" title="FM"></black>
  </board>
  <bye number="7" name="Brown" type="PAB"></bye>
</pairings>
```

Root element attributes:

| Attribute | Type | Description                      |
| --------- | ---- | -------------------------------- |
| `round`   | int  | Round number (current round + 1) |
| `boards`  | int  | Number of board pairings         |
| `byes`    | int  | Number of byes                   |

`<board>` children (`<white>`, `<black>`) and `<bye>` elements share these attributes:

| Attribute | Type   | Description                            |
| --------- | ------ | -------------------------------------- |
| `number`  | int    | Player start number (always present)   |
| `name`    | string | Player display name (omitted if empty) |
| `rating`  | int    | Player rating (omitted if 0)           |
| `title`   | string | Player title (omitted if empty)        |

`<bye>` elements also have a `type` attribute with the bye type string.

## Exit codes

The CLI uses six exit codes, defined as constants in `exitcodes.go`:

| Code | Constant           | Meaning                                       | Used by                                             |
| ---- | ------------------ | --------------------------------------------- | --------------------------------------------------- |
| 0    | `ExitSuccess`      | Operation completed successfully              | All subcommands                                     |
| 1    | `ExitNoPairing`    | No valid pairing or pairings mismatch         | pair, check, generate                               |
| 2    | `ExitUnexpected`   | Unexpected error (JSON encoding, write error) | All subcommands                                     |
| 3    | `ExitInvalidInput` | Bad input, unknown flags, malformed TRF       | All subcommands                                     |
| 4    | `ExitSizeOverflow` | Tournament too large for implementation       | Defined but not currently used                      |
| 5    | `ExitFileAccess`   | File I/O error (open, read, or write)         | pair, check, generate, validate, standings, convert |

## Scripting guidance

Exit codes allow reliable error handling in shell scripts:

```bash
chesspairing pair --dutch tournament.trf -o pairings.txt
case $? in
  0) echo "Pairings generated" ;;
  1) echo "No valid pairing found" ;;
  3) echo "Invalid input file" ;;
  5) echo "File error" ;;
  *) echo "Unexpected error" ;;
esac
```

For JSON output, check the exit code before parsing:

```bash
if output=$(chesspairing pair --dutch tournament.trf --format json); then
  echo "$output" | jq '.pairings | length'
else
  echo "Pairing failed with exit code $?" >&2
fi
```

## See also

- [pair](../pair/) -- primary pairing subcommand
- [Legacy Mode](../legacy/) -- backward-compatible interface (supports list and wide formats only)
