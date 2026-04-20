# Changelog

All notable changes to this project will be documented in this file. The
format is loosely based on [Keep a Changelog](https://keepachangelog.com/),
and the project follows [Semantic Versioning](https://semver.org/) once it
reaches a tagged release.

## [Unreleased]

### Added

- `Parse*` helpers in the root package for the public enum types:
  `ParseScoringSystem`, `ParsePairingSystem`, `ParseGameResult`, and
  `ParseByeType`. Permissive (case-insensitive, whitespace-tolerant,
  accepting common aliases like `fide-dutch`, `rr`, and the TRF result
  letters `F`/`H`/`Z`/`U`).
- `PlayedPairs(state, HistoryOptions)` for deriving the set of unordered
  pairs that have already been played. The default semantics (single
  forfeits excluded, double forfeits always excluded) match FIDE's
  position that forfeited games may be replayed; setting
  `IncludeForfeits` is house-rule territory.
- `chesspairing/factory` sub-package with `NewPairer`, `NewScorer`, and
  `NewTieBreaker` constructors keyed by name, plus `PairerNames`,
  `ScorerNames`, and `TieBreakerIDs` for discovery. The CLI's internal
  factory now delegates to this public package.
- `chesspairing/standings` sub-package with `Build` and `BuildByID` for
  composing a Scorer with a list of TieBreakers into a presentation-ready
  table. Two opinionated choices, both documented on `Build`:
  double-forfeit games count as 0 across the board (no win, no draw, no
  loss, no game played); true ties on score and all tiebreaker values
  share a rank, with the next distinct row's rank skipping accordingly
  (standard "1224" competition ranking).
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
- Forfeit-handling matrix in the root package documentation, summarising
  how Scorer, TieBreaker, PlayedPairs, and `standings.Build` each treat
  single and double forfeits.
- `TournamentState.PreAssignedByes []ByeEntry` for declaring byes locked
  in for the upcoming round before pairing runs (e.g. a player notified
  the arbiter they will be absent). All Swiss-style pairers (Dutch,
  Burstein, Dubov, Lim, Team, Double-Swiss, Keizer) now drop those
  players from the matching pool and echo the entries back in
  `PairingResult.Byes` with the declared type intact. The roundrobin
  pairer rejects non-empty `PreAssignedByes` because the Berger schedule
  is fixed.
- TRF bridge for `PreAssignedByes`: `ToTournamentState` populates the
  field from Section 240 absence records (`F` → `ByePAB`, `H` →
  `ByeHalf`) and from typed `### chesspairing:bye round=N player=SN
  type=...` comment directives, and `FromTournamentState` writes them
  back the same way. Section 240 only carries the two FIDE-defined
  letters; richer types (`ByeZero`, `ByeAbsent`, `ByeExcused`,
  `ByeClubCommitment`) round-trip via the directive form. When both
  sources name the same player in the same round the directive wins.
  Unknown player IDs in either source are reported as a validation
  error rather than silently dropped.
- `### chesspairing:` typed comment directives parsed into a new
  `Document.ChesspairingDirectives []Directive` field. Unknown verbs
  are preserved verbatim through Read/Write so older parsers do not
  drop data a future library version understands. The
  `chesspairing:withdrawn player=N after-round=M` directive is
  currently parsed and round-tripped only; bridging it into a
  per-player withdrawal field is reserved for a follow-up commit.

### Changed

- The CLI's `standings` subcommand now delegates to `standings.BuildByID`
  rather than assembling rows itself. Visible behaviour change: a double
  forfeit no longer counts as one played game with zero W/D/L. It now
  contributes nothing to GamesPlayed or W/D/L, matching the documented
  forfeit semantics elsewhere.
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
