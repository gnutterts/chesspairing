---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Version history and notable changes."
---

## Current Status

The chesspairing module is in active development. No stable releases have been published yet.

During development, the version string defaults to `"dev"`. Release versions are set at build time via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.0.0" ./cmd/chesspairing
```

Once stable releases begin, this page will track the full version history.

## Format

Future changelog entries will follow this format:

```text
## v1.0.0 (YYYY-MM-DD)

### Added
- ...

### Changed
- ...

### Fixed
- ...
```
