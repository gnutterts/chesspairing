---
title: "validate"
linkTitle: "validate"
weight: 5
description: "Validate a TRF file against one of three profiles."
---

The `validate` subcommand checks a TRF16 file for structural and semantic errors. Unlike most other subcommands, `validate` does not require a pairing system flag -- it operates purely on the TRF document structure.

Three validation profiles are available, each adding stricter checks on top of the previous level.

## Synopsis

```text
chesspairing validate input-file [options]
```

The input file can be a filesystem path or `-` for stdin.

## Flags

| Flag        | Type   | Default    | Description                                         |
| ----------- | ------ | ---------- | --------------------------------------------------- |
| `--profile` | string | `standard` | Validation profile: `minimal`, `standard`, `strict` |
| `--json`    | bool   | `false`    | Output as JSON                                      |
| `--help`    | --     | --         | Show usage help                                     |

## Profiles

**minimal** (`ValidateGeneral`): Basic structural validation. Checks that the TRF parses correctly and required fields are present.

**standard** (`ValidatePairingEngine`): Default. Adds checks needed by pairing engines -- player data completeness, round result consistency, cross-references between players and rounds.

**strict** (`ValidateFIDE`): Adds FIDE-specific requirements for official tournament reporting.

## Examples

```bash
# Default (standard profile, text output)
chesspairing validate tournament.trf

# Minimal validation
chesspairing validate tournament.trf --profile minimal

# Strict FIDE validation with JSON output
chesspairing validate tournament.trf --profile strict --json

# Read from stdin
cat tournament.trf | chesspairing validate -
```

## Output

### Text format (default)

When issues are found:

```text
tournament.trf: 2 errors, 1 warning

Errors:
  player.3.rating: rating must be a positive integer
  round.2.result: unknown result code

Warnings:
  player.5.title: title field is empty
```

When the file is clean:

```text
tournament.trf: 0 errors, 0 warnings
```

### JSON format

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
  "profile": "standard",
  "format": "auto"
}
```

The `valid` field is `true` only when there are zero errors. Warnings do not affect validity.

## Exit codes

| Code | Meaning                                            |
| ---- | -------------------------------------------------- |
| 0    | File is valid (no errors; warnings may be present) |
| 2    | JSON encoding error                                |
| 3    | Validation errors found, or invalid input          |
| 5    | File access error                                  |

## See also

- [convert](../convert/) -- re-serialize a TRF file
- [Output Formats and Exit Codes](../output-formats/) -- detailed format specifications and all exit codes
