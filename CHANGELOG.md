# Changelog

All notable changes to this project will be documented in this file. The
format is loosely based on [Keep a Changelog](https://keepachangelog.com/),
and the project follows [Semantic Versioning](https://semver.org/) once it
reaches a tagged release.

## [Unreleased]

### Added

- `SECURITY.md` describing the (small) attack surface and how to report
  vulnerabilities.
- `govulncheck` step in CI.
- Cross-platform CI matrix: tests now run on Ubuntu, macOS, and Windows.
- Coverage profile uploaded as a CI artifact.
- Package documentation for `pairing/swisslib` explaining its testing
  strategy (low per-package coverage is intentional; integration coverage
  via dependents is ~94%).
- Unit tests for `loadRTGConfig` in the CLI: full key parsing, missing
  files, malformed values, and unknown keys. Coverage of that function
  rose from ~35% to ~84%; CLI overall from ~79% to ~83%.

### Changed

- Minimum Go version reduced from 1.26.1 to 1.24. The actual feature
  floor was Go 1.24 (`testing.B.Loop`); the previous pin was overly
  restrictive.
- Removed unused root-level Node tooling (`package.json`,
  `node_modules/`). Documentation site keeps its own `docs/package.json`
  for PostCSS, which is the only real use.

### Removed

- Dead C8 look-ahead infrastructure (~720 LOC across two packages).
  Investigation against FIDE C.04.3 (effective 1 Feb 2026) confirmed
  that C8 ("choose downfloaters so the next bracket complies with
  C1–C7") is structurally subsumed by the global Blossom matching: the
  edge weight encoding in `swisslib.ComputeBaseEdgeWeight` already
  maximizes pairs and scores in the next bracket, which is what C8
  demands. The bracket-by-bracket scaffolding was a remnant of the
  pre-global-matching architecture and was never wired into the active
  pairing path. Removed: `pairing/dutch/matching.go`,
  `pairing/dutch/matching_test.go`, `pairing/swisslib/candidate.go`
  (Candidate, CandidateScore, IdxC8..IdxC21, NumViolations),
  `LookAheadFunc`/`LookAhead`/`RemainingBrackets` from
  `swisslib.CriteriaContext`, `SatisfiesAbsolute`, and the unused
  `recordFloats` helper. Dubov's local `IdxC8 = 4` (its own consecutive-
  upfloaters criterion per FIDE C.04.4.1) is unrelated and unchanged.

## [0.0.0] — Pre-history

The `git log` is the changelog for everything before the first tag.
Highlights:

- Eight FIDE-aligned pairing engines (Dutch, Burstein, Dubov, Lim,
  Double-Swiss, Team Swiss, Keizer, Round-Robin)
- Three scoring engines (Standard, Keizer, Football)
- Twenty-five tiebreakers via a self-registering registry
- TRF16 / TRF-2026 reader, writer, validator, and JSON converter
- CLI with eight subcommands and a legacy compatibility mode
- Bilingual (EN/NL) documentation site at https://chesspairing.nl
- Apache-2.0 licensing with SPDX headers throughout

[Unreleased]: https://github.com/gnutterts/chesspairing/compare/main...HEAD
