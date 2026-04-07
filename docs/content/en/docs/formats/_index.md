---
title: "Format Specifications"
linkTitle: "Formats"
weight: 80
description: "File format specifications — TRF16, TRF-2026 extensions, JSON schemas, and configuration."
---

This section documents the file formats used by chesspairing for tournament data exchange, engine configuration, and structured output.

Chesspairing uses TRF16 as its primary file format for tournament data exchange. TRF16 is a fixed-width text format defined by FIDE for tournament report files. The `trf` package provides a complete reader and writer with bidirectional conversion to and from the internal `TournamentState` representation.

TRF-2026 extends TRF16 with additional record types and system-specific fields. These extensions add header codes for total rounds, initial color, scoring systems, and tiebreaker definitions, as well as data records for absences, acceleration, forbidden pairs, and team data. The `trf` package transparently handles both TRF16 and TRF-2026 documents.

The CLI produces structured JSON output for pairings, standings, validation, version information, and tiebreaker listings. All JSON output uses 2-space indentation.

Engine options are configured through a `map[string]any` key-value format that can come from TRF fields, JSON configuration, or CLI flags. The `generate` subcommand also accepts an RTG configuration file with `key=value` syntax.

| Page                                   | Description                                                        |
| -------------------------------------- | ------------------------------------------------------------------ |
| [TRF16](trf16/)                        | The FIDE TRF16 format -- line types, player records, round results |
| [TRF-2026 Extensions](trf-extensions/) | System-specific XX fields for pairing engine configuration         |
| [JSON Schemas](json-schemas/)          | JSON output schemas for CLI commands                               |
| [Configuration](configuration/)        | Key-value configuration for engine factories and RTG               |
