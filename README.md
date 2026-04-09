# chesspairing

Chess tournament pairing, scoring, and tiebreaking algorithms in pure Go.

Eight pairing systems, three scoring engines, twenty-five tiebreakers — zero
external dependencies.

## What it does

Chesspairing generates pairings, computes standings, and calculates tiebreakers
for chess tournaments. It implements all six FIDE-approved Swiss pairing systems,
plus Keizer and round-robin. Any pairing system can be combined with any scoring
system — Swiss pairing with Keizer scoring, round-robin with football scoring,
whatever the tournament needs.

Everything operates on in-memory data structures. No I/O, no database, no
network calls. Build a `TournamentState`, pass it to an engine, get results
back.

### Pairing systems

| System       | FIDE Regulation                                      | Description                                            |
| ------------ | ---------------------------------------------------- | ------------------------------------------------------ |
| Dutch        | [C.04.3](https://handbook.fide.com/chapter/C0403)    | Global Blossom matching with 21 quality criteria       |
| Burstein     | [C.04.4.2](https://handbook.fide.com/chapter/C04042) | Seeding rounds with opposition-index re-ranking        |
| Dubov        | [C.04.4.1](https://handbook.fide.com/chapter/C04041) | ARO-equalization with transposition matching           |
| Lim          | [C.04.4.3](https://handbook.fide.com/chapter/C04043) | Median-first processing with exchange matching         |
| Double-Swiss | [C.04.5](https://handbook.fide.com/chapter/C0405)    | Lexicographic bracket pairing                          |
| Team Swiss   | [C.04.6](https://handbook.fide.com/chapter/C0406)    | Team-level pairing with configurable colour preference |
| Keizer       | —                                                    | Top-down by Keizer score with repeat avoidance         |
| Round-Robin  | [C.05](https://handbook.fide.com/chapter/C05)        | FIDE Berger tables with multi-cycle support            |

### Scoring systems

**Standard** (1–½–0), **Keizer** (iterative convergence), and **Football**
(3–1–0), all with configurable point values for wins, draws, byes, forfeits,
and absences.

### Tiebreakers

Twenty-five implementations covering all FIDE-recognized methods: Buchholz
(five variants), Sonneborn-Berger, Direct Encounter, Performance Rating,
Progressive Score, Koya, ARO, and more. Self-registering registry — look up
any tiebreaker by string ID.

## Use as a Go library

```
go get github.com/gnutterts/chesspairing
```

```go
pairer := dutch.New(dutch.Options{})
result, err := pairer.Pair(ctx, state)

scorer := standard.New(standard.Options{})
scores, err := scorer.Score(ctx, state)

tb, _ := tiebreaker.Get("buchholz-cut1")
values, err := tb.Compute(ctx, state, scores)
```

Three interfaces (`Pairer`, `Scorer`, `TieBreaker`) with a shared
`TournamentState` input. Safe for concurrent use when each goroutine
supplies its own state.

## Use as a CLI tool

The `chesspairing` command reads FIDE Tournament Report Files (TRF16 and
TRF-2026) and produces pairings, standings, and validation reports:

```
chesspairing pair tournament.trf
chesspairing standings tournament.trf
chesspairing validate tournament.trf
```

Output in five formats: plain list, wide tabular, board view, XML, and JSON. A
legacy mode provides drop-in compatibility with bbpPairings and JaVaFo
command-line conventions.

## Documentation

Full documentation is available at
**[gnutterts.github.io/chesspairing](https://gnutterts.github.io/chesspairing/)** —
including getting started guides, API reference, algorithm deep-dives with
mathematical notation, and FIDE regulation mappings.

## Testing

```
go test -race -count=1 ./...
```

1325 tests across 19 packages, including golden file comparisons against
bbpPairings and JaVaFo reference output, plus fuzz testing for the TRF parser.

## Acknowledgements

This project builds on the work of two chess pairing engines that came before
it:

- **[bbpPairings](https://github.com/BieremaBoyzProgramming/bbpPairings)** by
  Bierema Boyz Programming — a C++ implementation of the FIDE Dutch system
  using Blossom matching. The Dutch pairer in chesspairing follows the same
  architectural approach (global Blossom matching with completability
  pre-matching for bye determination), and bbpPairings' own test cases are
  included in the test suite for cross-validation.

- **[JaVaFo](http://www.rrweb.org/javafo)** by Roberto Ricca — a Java
  implementation that served as the FIDE reference pairer. Seven golden test
  scenarios from JaVaFo 2.2 are used to verify pairing correctness, and the
  CLI's legacy mode accepts JaVaFo command-line conventions.

The FIDE Handbook ([handbook.fide.com](https://handbook.fide.com)) is the
primary specification reference for all pairing and tiebreaking algorithms.

## License

Copyright © 2026 Gert Nutterts. Free for personal use and for chess clubs
affiliated with a FIDE member federation, including club activities that
generate income (entry fees, sponsorship, ticket sales). No commercial use
outside of club activities without permission. No distribution of modified
versions. See [LICENSE.md](LICENSE.md) for the full terms (governed by Dutch
law; the [Dutch text](LICENSE-NL.md) is legally binding).

## Disclaimer

This software is provided as-is, without warranty of any kind. Parts of this
software were developed with the assistance of AI tools. While the code is
tested against reference implementations and verified by over 1300 tests,
errors may exist that testing has not uncovered. In a chess tournament context
this means:

- **Pairings** may violate FIDE regulations, requiring correction mid-tournament.
- **Scores** may be calculated incorrectly, affecting standings and prizes.
- **Tiebreakers** may produce wrong values, changing final rankings among
  equal-scored players.

If you use this software for rated or official events, verify the output
independently. The author accepts no liability for errors in pairings, scores,
or rankings.
