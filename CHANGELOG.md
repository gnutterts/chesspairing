# Changelog

All notable changes to this project will be documented in this file. The
format is loosely based on [Keep a Changelog](https://keepachangelog.com/),
and the project follows [Semantic Versioning](https://semver.org/) once it
reaches a tagged release.

## [Unreleased]

### Changed

- Koya accepts a `koyaLimit` percentage of the maximum possible score, defaulting to 50 percent.
- Lim colour-history and Team Swiss Type A colour-preference interpretations are documented in both languages.
- Tie-break documentation now notes the round-robin prohibition on Buchholz and the half-point standard result for a double forfeit.
- Tiebreaker docs now state that the ratings-based tie-breaks of Article 10 count the games of a Double Swiss match separately (forfeits stay unplayed for them) and that the `unratedRating` scoring option is set through the Go API; the command-line tool and the TRF reader do not set it.
- Pairing and standings documentation now clarifies the Team Swiss secondary score, round-robin withdrawals, and Swiss colour and bye interpretations.

### Added

- Team standings now support Extended Direct Encounter and Scores and Schedule Strength Combination, including team TRF codes.
- Stored Dutch differential coverage against bbpPairings 6.0.0, with nightly exploration and TRF fuzz regression coverage.
- TRF: record 192 with `FIDE_DUTCH_2025_BAKU`, `FIDE_DUTCH_BAKU` or `FIDE_BURSTEIN_BAKU` now selects Baku acceleration, as bbpPairings reads it; before only `XXS` lines did.
- Burstein: `ErrBracketTooLarge` is returned when a bracket's floater-set enumeration exceeds 200000 candidates, rather than silently skipping criteria C5 and C7.

### Fixed

- Progressive scores now use configured tournament points.
- Koya qualification now uses half the maximum possible configured score.
- Wins now count only rounds awarding the configured win points.
- Perfect tournament performance now selects the lowest qualifying integer rating using symmetric table probabilities.
- Player rating now requires an unrated-player rule when an unrated player is present.
- Tie-break documentation and the rating comment now use the correct FIDE type labels.
- Performance documentation now describes nearest-entry table lookup rather than interpolation.
- Direct Encounter now resolves each subgroup still tied after earlier tie-breaks and reapplies within its remaining tied subsets.
- Repeated direct encounters against one opponent now contribute their average result rather than their sum.
- Direct Encounter and standings documentation now describes subgroup resolution, reapplication and repeated-encounter averaging correctly.
- Team Swiss pre-assigned team byes now remove all members of the team before pairing.
- Team Swiss reports when a pre-assigned pairing-allocated bye leaves an odd number of teams.
- Standard and football standings now award a win for an unset PAB value (win plus draw in Double Swiss), including with custom win points.
- Team match history now rejects a board-1 White player from the away team, and documents the home-team colour convention.
- TRF output now uses CRLF line endings as required by TRF-2026 Remark 1; the reader still accepts CR, LF, and CRLF.
- TRF team records 310/300/320/330/299/352/362 now carry rosters, assignments, byes, forfeits and scoring across conversion instead of losing team data.
- TRF tournament configuration records 192, 202/212, 222, and 162 now round-trip through tournament state and apply their FIDE settings.
- Round-robin standings now retain withdrawal dates for counted players' tie-breaks and preserve the TRF-inferred system when the CLI receives tie-breakers without a system flag.
- Team Swiss colour allocation now uses the secondary score by default under C.04.6 Article 4.2.2, with an option to disable it.
- The CLI now recognises a trailing all-blank TRF round as the round to pair instead of a completed round.
- Round-robin final standings now exclude results of withdrawn players who completed fewer than half their scheduled games under C.05 Article 6.6(2).
- Opponent records and standard points now use configured tournament points for games and byes, including C.07 Article 16 adjusted scores and dummy limits; previously football scoring used fixed 1-½-0 values.
- Board Count now rejects a tied set of teams with unequal game points as required by C.07 Article 12.1.
- Double-Swiss tie-breaks now merge the two games of a match into one per-round record, with both games' points added and the round played unless both games were forfeits (C.07 7.1, 7.7, 9.1 and 16; C.04.5 Preface); before, the second game replaced the first for Buchholz, Sonneborn-Berger, direct encounter, standard points and the other record-based tie-breaks, while the per-game tie-breaks of Articles 7.2-7.4 keep counting each game.
- The performance rating and performance points of Article 10 now count each game of a Double Swiss match separately when computing the fractional score (Article 10.2) and the expected scores (Article 10.3), while the average opponent rating keeps one entry per round; before, a split match (1-0, 0-1) was read as one played game worth the match total, inflating the performance rating by dp +800.
- Double-Swiss record merging is now order-independent: a match with one forfeited game and one pending game keeps the forfeit category of the decided game; before, the order of the round's two games decided whether the round kept its forfeit category or dropped out of Buchholz, Sonneborn-Berger and games elected entirely.
- Fore Buchholz now treats every game of a Double Swiss final-round match as drawn (C.07 8.3), replacing the match total with the per-match draw value; before, a 2-0 last round lowered the player's opponent-adjusted score by 1.5 instead of the match draw value.
- Double-Swiss matches now end by forfeit only when one player forfeited both games, decided from both games of the match before the per-player records are split (C.04.5 Preface); before, a double forfeit combined with an opponent's forfeit win marked the match regularly played.
- Games won (C.07 7.2) now ignores rounds behind a player's withdrawal, like the record model of the other tie-breaks; before, a recorded win after the withdrawal still counted.
- Ratings-based tie-breaks (ARO, ARO Cut-1, TPR, PTP, APRO and APPO) now return an error naming the unrated player who played a game over the board, instead of averaging a missing rating as 0, unless the scoring option `unratedRating` supplies the detailed rules on unrated players required by C.07 Article 10.
- Lim, Burstein, Double-Swiss and Team Swiss pairing scores now follow configured tournament points under C.04.2 Article 3.1 and C.04.4/C.04.5; before, their histories, opposition indices and game scores used fixed 1-½-0 points.
- Standard and football scoring now give both players the draw points for a draw in a game of less than one move (TRF `D`, a draw result with `IsForfeit` set); before, both scored nothing.
- Swiss pairing scores now follow tournament points under C.04.3 Articles 1.2–1.3 and C.04.2 Article 3.1, including byes and proportional Baku virtual points; before, games and byes always used 1-½-0.
- Late entries now receive ranked TPNs under C.04.2 Articles 2.4–2.5, shifting lower-ranked participants; before, they were appended after existing numbers.
- FIDE Swiss boards now follow C.04.2 Article 3.6 by score, score sum and higher-ranked participant's TPN; before, bracket or pairing order could determine boards.
- Team Swiss now follows C.04.6 Articles 2.3.5-2.3.7 and 3.6.4 by minimising quality criteria C8, C9, and C10 lexicographically in bracket pairing; before, colour preferences and C10 filtered pairings with an absolute-only retry.
- Swiss pairers now recognize forfeits from either the result or the IsForfeit flag (C.04.2 3.4-3.5); before, only the IsForfeit flag was checked, treating games with forfeit results as played.
- Adjourned and pending games now count as draws (0.5 points) for Swiss pairing scores under C.04.2 Article 3.1; before, pending games contributed 0 points to the pairing score.
- Team Swiss pairing now treats an adjourned match as a drawn match under C.04.2 Article 3.1: both teams receive the match points of a draw, and each adjourned board counts as a draw in game points, for the pairing score only; before, the match scored as its decided boards.
- Pre-assigned pairing-allocated byes now validate C2 eligibility (C.04.3 2.1.2 / Basic Rules Art. 4) and reject ineligible participants; before, pre-assigned PABs bypassed C2 checks.
- Double-Swiss colour history now follows C.04.5 Article 1.6 with one first-game colour per played match; before, both games and forfeited matches added colours.
- Double-Swiss forfeits now follow the C.04.5 Preface at match level for C1, C2, matches played, and colours, recognising forfeits from either the result or the IsForfeit flag; before, one forfeited game made a match unplayed and barred its winner from a pairing-allocated bye.
- Team Swiss colour alternation now follows C.04.6 Article 4.3.6 and C.04.2 Article 3.4 by comparing played histories from the most recent match; before, histories were compared from their oldest aligned round.
- Double-Swiss and Team Swiss now assign pairing-allocated byes in C.04.5/C.04.6 Article 3.4.1-3.4.4 order, checking a complete legal pairing, lowest score, most played matches, and largest fixed TPN; before, they did not check completion and Double-Swiss omitted matches played.
- Double-Swiss and Team Swiss now keep the fixed TPN required by C.04.5/C.04.6 Articles 1.1-1.2 for pairing, identifiers, byes, and colours; before, the TPN was renumbered by each round's score order.
- Double-Swiss now follows C.04.5 Articles 2.3.5, 3.6 and 4 by minimising C8 only while choosing a bracket pairing and allocating colours afterwards; before, a colour preference filtered pairings.
- Double-Swiss and Team Swiss now apply C.04.5/C.04.6 Article 2.3.3 [C6] by preferring the upfloater set that minimises the following bracket's required upfloaters; before, C6 never distinguished working sets.
- Double-Swiss and Team Swiss now search C.04.5/C.04.6 Article 3.5 upfloater sets profile by profile and charge the search limit per evaluated set; before, listing every set of a size exhausted the limit in large rounds.
- Double-Swiss and Team Swiss now build brackets top-down with Article 3.5 upfloater sets and C3 completion under C.04.5/C.04.6; before, bottom-up parity floats could require pooling all remaining participants.
- Double-Swiss and Team Swiss bracket pairing now follows C.04.5/C.04.6 Article 3.6 identifier order; before, it paired each lowest TPN with the next available TPN.
- Dubov pairing now follows FIDE C.04.4.1: the pairing-allocated bye and each bracket's upfloater set are chosen only when the remaining players stay completable (C4), upfloaters are grouped by score multiset, Article 3.2.4 first shifts players out of the smaller subgroup for unavoidable same-subgroup pairs and then equalises, and the first legal recomposition and depth-first 4.4 transposition are used. A shared per-bracket search budget returns `ErrBracketTooLarge` rather than skipping prescribed order; ARO, MaxT, C5--C10, last-round relaxations, and colour allocation also follow the 2026 procedure. Before, the bye ignored completability, residents were downfloated, a failed bracket recomposed the whole round, and the criteria and colours followed an older, capped procedure.
- Lim numbering direction now follows FIDE C.04.4.3 Article 4.4.1 (lowest numbered means largest TPN); before, it used the reverse direction.
- Lim exchange matching now follows FIDE C.04.4.3 Articles 4.1-4.4, including Article 4.4.1; before, it used greedy fallback matching.
- Lim scoregroup transfers now follow FIDE C.04.4.3 Article 2.3 by applying 2.3.1-2.3.3 before 2.3.4; before, parity selection could happen first.
- Lim floater selection now follows FIDE C.04.4.3 Articles 3.2-3.10, including FIDE Article 3.9 priorities and Article 3.10's previous-floater limitation; before, those rules were incomplete.
- Lim floater pairing now follows FIDE C.04.4.3 Articles 3.6-3.8 and the Maxi rule; before, its priority and exchanged-opponent rating rules differed.
- Lim median processing now follows FIDE C.04.4.3 Articles 2.2 and 2.6 cracking; before, floaters were routed across phases and cracking was absent.
- Lim last-round colours now follow FIDE C.04.4.3 Article 6 when `TotalRounds` is set; before, the last-round exception was ignored.
- Lim errors now report impossible pairings instead of best-effort results where FIDE C.04.4.3 permits none; before, partial results could be returned.
- Lim colour exchanges now follow FIDE C.04.4.3 Article 5.2 over all pairs of a scoregroup; before, a colour conflict was only exchanged with later pairs.
- Round-robin rest rounds now follow TRF-2026 code Z and receive zero points; before, the dummy pairing gave a full-point pairing-allocated bye.
- Round-robin Berger slots now follow FIDE C.05 Annex 1 pairing numbers; without explicit numbers they are assigned by rating, title and name (C.04.2 2.2), and `varma.Assign` sets them. Late entries are rejected because the table is fixed. Before, the player slice order determined the table.
- Round-robin withdrawals now follow FIDE C.05 6.6 with a fixed Berger table and forfeit notes for the caller to record; before, removing inactive players changed the table, repeated pairings, or ended the schedule early.
- Team tie-breaks now apply FIDE C.07:2026 Articles 12 and 16: board tie-breaks include standard board wins for team byes, compare board by board in board order (half points included), and cap match-point and game-point dummy opponents; previously they omitted team byes, stopped at one board, and left dummies uncapped.
- Round-robin tie-breaks now apply FIDE C.07:2026 Article 15.2 to forfeits for games won, black-game measures, and Koya, and `wins` is now named "Games Won" instead of "Games Won (OTB)"; previously these results were excluded.
- BPG and REP now follow FIDE C.07:2026 Articles 7.3, 7.6, and 16.1.1: pending games do not count as games with Black, and rounds-played subtracts rounds after a withdrawal and rounds before a late join (matching Games Elected); previously those rounds were counted.
- Team pairing now uses team-specific library default tie-breaks; previously it used individual tie-breaks that returned zero for teams.
- TRF records 013 and 310 now follow the TRF-2026 team-section columns; before, 013 wrote a non-existent team number and 310 shifted fields.
- TRF record 132 now follows the TRF-2026 round-date columns (YY/MM/DD) on one line, and a legacy YYYY/MM/DD date is written in that form; before, each date was written on a separate short line.
- TRF records 250 and 300 now follow their TRF-2026 fixed-width fields; before, generated records shifted later fields.
- TRF record 320 now follows the TRF-2026 pairing-allocated-bye layout and converts its teams to `RoundData.TeamByes`; before, it was treated as round scores.
- TRF player round results now follow the TRF-2026 equivalence of four blank opponent columns to `0000`; before, the four-blank spelling of the opponent was rejected.
- Burstein pairing now follows FIDE C.04.4.2: Dutch seeding rounds, opposition-index ranking, bye selection, bracket ordering, and colour allocation follow the published rules. Article 1.7.2 is interpreted so that a series of zero-point byes benefits only actual opponents, while unplayed rounds count as games the player plays against themselves. The output no longer matches bbpPairings' Burstein, which deviates from the published rules. Fixed two bracket-search faults: a bracket could be judged unpairable, and the Article 4.3 order did not break ties between floater sets.
- Dutch pairing: the colour of two players without a colour preference (C.04.3 5.2.5) is decided by the higher ranked player, ranked by score first and then by TPN (article 1.2), and by the parity of the number among the players who have entered the tournament, in every round. A player whose only entries so far are requested byes (a late entry, C.04.2 2.4) does not count until they take part. Before, the order used the TPN alone and the numbering was done in round 1 only, which gave the opposite colours in some rounds.
- Dutch pairing: the repeated-float criteria are weighed in the order of article 2.4: C16 and C17 (floats two rounds ago) now outrank C18 and C19 (the scores involved in the previous round's floats), and the C11 comparison of colour imbalances uses the absolute difference. Before, some rounds, mostly the last, were paired with a repeated float or a third equal colour that the rules allow avoiding.
- Dutch and Burstein pairing with Baku acceleration (C.04.7): group A is counted over participants before the first round, and its last player stays the same, so late entries numbered above that player join group A; the scoring premise of article 1.1 is checked and Burstein requires total rounds. Virtual points are those of a win on the 1-½-0 scale of the pairing score, so 2-1-0 scoring no longer doubles them, and they count in the scores that decide the scoregroups, the criteria, the bye and who floated in earlier rounds (pairing score, 1.5). Before, late entries counted in the size of group A, so group A grew and its last player changed, any scoring was accepted, and Burstein used the current round as the tournament length when the total was not given.
- Pairing: equal edge weights no longer give a different matching from run to run (the edges are sorted before matching).
- CLI `pair`: when the last round column of the TRF holds only byes and announced absences of the round that is about to be paired, as written by bbpPairings, JaVaFo and tournament programs, those become the pre-assigned byes of the next round instead of a played round. Before, the tournament was paired one round too late (and, with a known number of rounds, failed with "current round exceeds total rounds") and those players were credited with the points.
- TRF writer: blank 001, 013 and 310 records are padded to the shortest width the reader accepts, team member numbers above 9999 are rejected instead of running into the next field, and a 4-digit member number after another member is separated by a space; before, such files could not be read back.

## [0.6.0] — 2026-10-03

### Added

- Team scoring now records match and game points, including TRF-2026 team match data.
- New `team` scoring system (`ScoringTeam`, `chesspairing standings --scoring team`) with match points and game points.
- Team standings now show match and game points and support FIDE C.07 team tie-break variants.

### Fixed

- Team scoring, tie-breaks, pairing-allocated bye eligibility, and TRF team round trips now handle incomplete, forfeited, and totals-only team matches correctly.
- Double Swiss and Team Swiss choose the upfloater as C.04.5 and C.04.6 articles 3.5.3-3.5.5 prescribe: by score and then lowest pairing number, and only when the rest of the bracket can still be paired. Before, the lowest-ranked player always floated up, which could give a pairing the rules do not allow (for example E-C and A-B instead of E-B and A-C).
- Team Swiss now pairs on its configured primary score, keeps team byes separate from player byes, and avoids previous-round floaters when selecting upfloaters; Double Swiss pairing-allocated byes now score 1.5 points by default.

## [0.5.0] — 2026-10-03

### Fixed

- Dutch pairing: in the first round the colour rule for players without a colour preference (C.04.3 5.2.5) counts the parity of the player's number among the players taking part in the round, so a player with a requested bye no longer shifts the colours of everyone below them; this is how JaVaFo and bbpPairings number them. Later rounds keep the fixed pairing number.
- Double Swiss colour allocation now applies the five rules of Article 4.3 in priority order; who plays whom does not change.
- Tournament performance rating (TPR) and the tie-breaks built on it (APRO): the fractional score is rounded to two decimals (0.5 up) before the rating-difference table lookup instead of being interpolated between table entries, as C.07 Article 10.2 and the official exercises prescribe. Values change by a few points where the fractional score is not a whole hundredth.

## [0.4.0] — 2026-10-02

### Added

- Added a deterministic ESG club-season regression fixture for Keizer scoring and pairing.
- Keizer scoring: added the `Method` option (`iterative`, `frozen`, `keizer1956`), `withdrawnAsAbsent`, `forfeitCountsAsMet`, and `ESGOptions()`; defaults and existing behaviour are unchanged.
- Keizer pairing: added `byePolicy`, `periodLength`, `noRepeatWithinPeriod`, `forfeitCountsAsMet`, and `ESGOptions()`; the default output is unchanged where greedy already produces a legal matching.

### Changed

- Faster maximum weight matching: fewer big-integer allocations in the Blossom implementation (output unchanged).
- Fixed FIDE tournament pairing numbers are now preserved and used for seeding, byes, board order, acceleration, and Dutch colour parity.
- Opponent-based tie-breaks now follow FIDE C.07:2026 Articles 15-16 for unplayed rounds, including adjusted opponent scores, capped dummy opponents, and the Cut-1 exception.

### Fixed

- Dutch pairing: a topscorer in the last round is a player with over 50% of the points possible after the rounds played so far (4.5 or more after eight rounds), not over 50% of the whole tournament. Before, a player with 2.5 points before round 5 of 5 was not a topscorer, so the color exception was refused and the round could be paired differently or fail.
- Dutch and Burstein pairing: with an odd number of players, a round no longer fails with "no player is eligible for the pairing-allocated bye" when a complete pairing exists. The search for the bye player now only considers pairs that can be played, so it respects the absolute color rules.
- Dutch pairing: C9 now counts forfeits as unplayed rounds and applies only in the bracket supplying the pairing-allocated bye.
- TRF writer: the points column now counts a pairing-allocated bye (`U`) as one point; before, a player with such a bye was written with no points for it.
- TRF writer: a player without a game and without a bye in a round (also a withdrawn player after the last round) is written as a zero-point bye (`Z`) instead of `U`, so reading the file back no longer gives that player a pairing-allocated bye.
- Keizer pairing with `byePolicy="lowest-without-bye"`: absences and external games no longer count as byes, and once everyone has had a bye the next round of byes again starts from the lowest-ranked player with the fewest byes.
- Pairers now report incomplete or impossible rounds as typed errors instead of returning partial pairings.
- Double Swiss and Team Swiss retry the whole field in TPN order when bracket pairing fails, which can pair across score groups; the affected CLI golden cases now exit 1.
- TRF bye codes now distinguish pairing-allocated byes (`U`) from requested full-point byes (`F`), and write absences as zero-point byes (`Z`).
- Swiss pairers now reject a pairing-allocated bye when every active player is ineligible under C2.
- Dubov: the average rating of opponents now includes opponents who withdrew later (C.04.4.1 art. 1.7.1); before, they counted as zero.
- Keizer pairing: when top-down greedy hit a forbidden repeat it emitted the note "Could not avoid repeat pairing"; the pairer now falls back to a full matching search (blossom, minimizing the sum of rank distances) and returns an error when no legal pairing exists.

## [0.3.0] — 2026-09-26

### Added

- Regression tests against the official FIDE examples (Dutch 2026, Dubov, Lim, Baku) and the C.07 tie-break exercises; cases that still deviate are listed as pending in the test files.
- A hard C.07:2026 Article 16.2.5 regression test for Buchholz, Buchholz
  Cut-1, and Sonneborn-Berger.

### Fixed

- All pairing systems now honour context cancellation and deadlines.
- Baku acceleration: virtual points are now derived from the points awarded for a win instead of a fixed 1.0/0.5.
- TRF: the initial colour from record 152 (`B`) or `XXC black1` was ignored, so the top seed always received White in round 1.
- Dutch: the last-round rules (topscorers, Baku virtual points) applied to every round when the planned number of rounds was not known; the pairer now reads `totalRounds` and only treats the final round as such.
- Tiebreaks: ARO is rounded to the nearest whole number (0.5 up) and TPR uses only points scored in games played over the board (C.07 art. 10.1-10.2).
- CLI: pair on a TRF with played rounds paired the last played round again instead of the next round (regression in v0.2.0).
- TRF: a double forfeit (both players marked `-`) is read as `0-0f` instead of a forfeit win for Black, and mutually inconsistent round results are rejected instead of silently crediting one side.
- TRF: files with CR-only line endings (as the TRF-2026 specification prescribes) were read as a single line; lower-case result and colour codes and a blank result (equivalent to `Z`) are now accepted. A CRLF pair split across two read boundaries no longer yields an empty line, and a blank or dash colour is now only accepted in a bye round.
- Keizer pairing now reads nested `scoringOptions`, with nested values taking precedence over compatible flat options.
- Keizer `PointsForResult` is consistent with `Score` for all bye types.
- Keizer rounds are scored in round-number order.
- TRF: 260 forbidden-pair records only apply when the round to be paired falls within their FirstRound..LastRound window (0 means unbounded); legacy XXP pairs stay permanent.
- TRF: 162 scoring-points records now drive the standard scorer's point options instead of always using the default scoring.

### Changed

- Keizer self-victory now awards value number × `WinFraction` when a tournament has no rounds.
- Raised minimum Go version to 1.26. Go 1.25 reached end-of-life with the
  Go 1.27 release; the Go team supports only the two most recent releases.
- CI now tests on both supported Go releases (`oldstable` and `stable`) on
  Linux, macOS and Windows, pins govulncheck and golangci-lint to explicit
  versions, and rejects committed AI-assistant configuration files. A single
  `CI OK` job aggregates the required checks.
- Dependency updates are managed by Renovate, with a seven-day minimum release
  age and GitHub Actions pinned to commit digests.
- The `main` branch only accepts changes through pull requests with passing CI.

### Removed

- The GitHub Pages workflow. The documentation moves into the repository and
  the project wiki.

## [0.2.2] — 2026-04-21

CI maintenance release. No library or CLI changes.

### Changed

- Raised minimum Go version to 1.25. Go 1.24 reached end-of-life with
  the Go 1.26 release on 2026-02-10 and no longer receives security
  patches. The April 2026 security batch (Go 1.25.9 / 1.26.2) patched
  six stdlib CVEs in archive/tar, crypto/tls, crypto/x509, html/template,
  and os; bumping the floor lets the CI vulnerability scan run against
  a maintained stdlib. The module's import surface does not reach any
  of the affected packages, so this is a defence-in-depth change rather
  than a fix for a known reachable vulnerability.
- Dropped the temporary `govulncheck` v1.1.4 pin introduced in v0.2.1.
  With the module floor at Go 1.25, `golang.org/x/vuln@latest` (v1.2.0
  and beyond) builds and runs again.

## [0.2.1] — 2026-04-20

CI maintenance release. No library or CLI changes.

### Fixed

- Pinned `govulncheck` to v1.1.4. The `@latest` resolver jumped to
  v1.2.0 which requires Go 1.25, breaking the vuln scan against the
  module's 1.24 floor under `GOTOOLCHAIN: local`. v1.1.4 is the last
  release that builds on 1.24 and still reports zero vulnerabilities
  affecting the module.

### Changed

- Bumped CI actions off Node.js 20 ahead of the 2026-06-02 deprecation:
  `golangci/golangci-lint-action` v8 → v9 and `actions/upload-artifact`
  v5 → v7. Both are drop-in for our usage.

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

[Unreleased]: https://github.com/gnutterts/chesspairing/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/gnutterts/chesspairing/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/gnutterts/chesspairing/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/gnutterts/chesspairing/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/gnutterts/chesspairing/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/gnutterts/chesspairing/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/gnutterts/chesspairing/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/gnutterts/chesspairing/compare/v0.0.0...v0.2.0
