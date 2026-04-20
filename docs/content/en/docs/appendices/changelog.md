---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Version history and notable changes."
---

The canonical changelog lives in [`CHANGELOG.md`](https://github.com/gnutterts/chesspairing/blob/main/CHANGELOG.md) at the repository root. It follows a loose [Keep a Changelog](https://keepachangelog.com/) format and Semantic Versioning.

## Latest release: v0.2.0 (2026-04-20)

This release reworks the unplayed-round vocabulary across every subsystem alongside a batch of public-API additions that had been sitting unreleased. The six `ByeType` values now flow consistently through scoring, tiebreaking, pairing, TRF I/O, and the standings table. Player withdrawals move from a per-round `Active` boolean to a one-shot `WithdrawnAfterRound` pointer. Pre-assigned byes for the upcoming round get a first-class field on `TournamentState`. `ResultContext` exposes the bye type directly instead of collapsing it to a single "is a bye" flag.

The new `Parse*` helpers, `PlayedPairs`, the `factory` sub-package, and the `standings` sub-package round out the public API for downstream tools that previously re-implemented these pieces themselves. Pre-1.0 breaking changes are listed in the root `CHANGELOG.md` under Removed; there are no shims.

## Build-time version string

During development, the version string defaults to `"dev"`. Release versions are set at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=v0.2.0" ./cmd/chesspairing
```
