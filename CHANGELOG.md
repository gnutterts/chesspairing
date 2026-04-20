# Changelog

All notable changes to this project will be documented in this file. The
format is loosely based on [Keep a Changelog](https://keepachangelog.com/),
and the project follows [Semantic Versioning](https://semver.org/) once it
reaches a tagged release.

## [Unreleased]

## [0.2.0] — 2026-04-20

This release reworks the unplayed-round vocabulary across every
subsystem alongside a batch of public-API additions that had been
sitting unreleased. The six `ByeType` values now flow consistently
through scoring, tiebreaking, pairing, TRF I/O, and the standings
table; player withdrawals move from a per-round `Active` boolean to a
one-shot `WithdrawnAfterRound` pointer; pre-assigned byes for the
upcoming round get a first-class field on `TournamentState`; and
`ResultContext` exposes the bye type directly instead of collapsing it
to a single "is a bye" flag. The new `Parse*` helpers, `PlayedPairs`,
the `factory` sub-package, and the `standings` sub-package round out
the public API for downstream tools that previously re-implemented
these pieces themselves. Pre-1.0 breaking changes are listed under
Removed below — there are no shims.

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
- Bye-type and absence matrix in the root package documentation,
  listing each `ByeType` with its default standard-scoring weight,
  whether it counts as a played round for tiebreakers, whether it is
  tracked under the PAB-uniqueness constraint, and its TRF round-column
  code (or directive form, for `ByeExcused` / `ByeClubCommitment`).
- `standard.Options.PointExcused` and `standard.Options.PointClubCommitment`
  (`*float64`, default `0.0`, JSON keys `pointExcused` and
  `pointClubCommitment`). The standard scorer's `Score` loop and
  `PointsForResult` both dispatch per `ByeType`, so excused and club-
  commitment byes can carry a configurable weight rather than being
  silently mapped to absent. The football scorer inherits the new
  options through composition.
- Per-bye-type bucketing in tiebreaker `opponentData`: `playerByes`
  is now `map[string]map[ByeType]int`. Every Category-B tiebreaker
  (Buchholz, Sonneborn-Berger, Foreheads, Average Opponent Buchholz,
  AOPR, Koya, Direct Encounter, Performance Points, Performance
  Rating) handles all six bye types explicitly. Virtual-opponent rules
  unchanged: every unplayed round contributes per FIDE simplified VOO.
- Documented Keizer scoring's `Frozen` option behaviour (single-pass,
  historical scoring path, no oscillation handling needed).
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
  `chesspairing:withdrawn player=N after-round=M` directive bridges
  to `PlayerEntry.WithdrawnAfterRound` in both directions.
- `PlayerEntry.WithdrawnAfterRound *int` for marking a player as
  withdrawn after a specific round. Nil means active. The companion
  `TournamentState.IsActiveInRound(playerID, round)` predicate and
  `ActivePlayerIDs(round)` helper give the contemporaneous view: a
  player who withdrew after round 5 still counts as having played
  rounds 1-5 and is excluded from round 6 onward. As a convenience for
  scoring callers without a specific round anchor, `round <= 0` means
  "no round filter" — any not-yet-withdrawn enrolled player is active.
  `Validate()` rejects zero, negative, or `WithdrawnAfterRound` values
  greater than `CurrentRound`.

### Changed

- The CLI's `standings` subcommand now delegates to `standings.BuildByID`
  rather than assembling rows itself. Visible behaviour change: a double
  forfeit no longer counts as one played game with zero W/D/L. It now
  contributes nothing to GamesPlayed or W/D/L, matching the documented
  forfeit semantics elsewhere.
- All Swiss pairers, scorers, and tiebreakers now use
  `IsActiveInRound`/`ActivePlayerIDs` instead of the removed
  `PlayerEntry.Active` flag. Tiebreakers use a per-round contemporaneous
  active set so a player's pre-withdrawal games still count for opponent
  tiebreakers like Buchholz.
- TRF: round-trip a withdrawn player's status via the
  `### chesspairing:withdrawn` directive on read and write.
- Minimum Go version reduced from 1.26.1 to 1.24. The actual feature
  floor was Go 1.24 (`testing.B.Loop`); the previous pin was overly
  restrictive.
- Removed unused root-level Node tooling (`package.json`,
  `node_modules/`). Documentation site keeps its own `docs/package.json`
  for PostCSS, which is the only real use.

### Fixed

- Removed a stale `FixedWinValue` paragraph from the Keizer convergence
  doc (EN+NL). No such option exists or has ever existed: fixed values
  in Keizer apply only to byes and absences (`ByeFixedValue`,
  `AbsentFixedValue`), and game results are necessarily fractional
  against opponent value — that is the defining property of Keizer.
- Documented that Keizer's `LateJoinHandicap` is intentionally fixed-
  only with no fraction companion. A late joiner has no value-number
  history to apply a fraction to, and the handicap exists so the
  arbiter can pick a single deterministic catch-up score.
- CLI generator (`cmd/chesspairing/generate.go`) now lists `ByeAbsent`,
  `ByeExcused`, and `ByeClubCommitment` explicitly in the bye-to-TRF
  switch instead of sweeping them into a single default branch. The
  three-way collapse was correct for `ByeAbsent` (TRF "U" = 0 pts) but
  silently lost the configurable point values of the other two.

### Removed

- `ResultContext.IsBye`, `ResultContext.IsAbsent`, and
  `ResultContext.IsForfeit`. Replaced by `ResultContext.ByeType *ByeType`
  (non-nil indicates a bye of that type) and `Result.IsForfeit()` for
  forfeit detection. Scorers' `PointsForResult` now dispatches on the
  bye type directly, which lets callers distinguish PAB / Half / Zero /
  Absent / Excused / ClubCommitment instead of collapsing them to a
  single "is a bye" flag. Pre-1.0 break, no shim.
- `PlayerEntry.Active bool`. Replaced by the nullable
  `WithdrawnAfterRound *int` and the `IsActiveInRound` /
  `ActivePlayerIDs` accessors. This is a clean break — no shim, no
  deprecation period, in line with the pre-1.0 API.
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

[Unreleased]: https://github.com/gnutterts/chesspairing/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/gnutterts/chesspairing/compare/v0.0.0...v0.2.0
