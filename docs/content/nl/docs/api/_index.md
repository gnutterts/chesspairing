---
title: "Go API-referentie"
linkTitle: "Go API"
weight: 70
description: "Handgeschreven API-documentatie voor de chesspairing Go-module — typen, interfaces en gebruikspatronen."
---

Dit is handgeschreven API-documentatie voor de `github.com/gnutterts/chesspairing` Go-module. De broncode is de enige bron van waarheid; deze pagina's beschrijven wat de code doet, niet andersom.

## Module-informatie

- **Modulepad**: `github.com/gnutterts/chesspairing`
- **Go-versie**: 1.24
- **Externe afhankelijkheden**: geen (alleen stdlib)

## Kern-interfaces

Het rootpakket definieert drie interfaces:

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}

type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

## Gegevensstroom

```text
*TournamentState
  -> Pairer.Pair()       -> *PairingResult
  -> Scorer.Score()      -> []PlayerScore
  -> TieBreaker.Compute() -> []TieBreakValue
```

Indeling en scoring zijn onafhankelijk. Elke pairer kan met elke scorer worden gecombineerd. Een toernooi kan Zwitserse indeling met Keizer-scoring gebruiken, of round-robin-indeling met Football-scoring.

## Context-parameter

Alle methoden accepteren `context.Context` als eerste parameter voor API-compatibiliteit. Op dit moment controleert geen enkele engine op annulering -- alle berekeningen zijn CPU-gebonden en in-memory.

## Gelijktijdigheid

Alle engines zijn veilig voor gelijktijdig gebruik wanneer elke goroutine zijn eigen `TournamentState` meestuurt. Er is geen gedeelde muteerbare staat.

## Pakketten

| Pakket                           | Doel                                           |
| -------------------------------- | ---------------------------------------------- |
| [`chesspairing`](overview/)      | Interfaces, gedeelde typen, configuratie-enums |
| [`pairing/dutch`](pairer/)       | Nederlandse Zwitserse pairer (C.04.3)          |
| [`pairing/burstein`](pairer/)    | Burstein Zwitserse pairer (C.04.4.2)           |
| [`pairing/dubov`](pairer/)       | Dubov Zwitserse pairer (C.04.4.1)              |
| [`pairing/lim`](pairer/)         | Lim Zwitserse pairer (C.04.4.3)                |
| [`pairing/doubleswiss`](pairer/) | Dubbel-Zwitserse pairer (C.04.5)               |
| [`pairing/team`](pairer/)        | Team-Zwitserse pairer (C.04.6)                 |
| [`pairing/keizer`](pairer/)      | Keizer-pairer                                  |
| [`pairing/roundrobin`](pairer/)  | Round-robin-pairer (C.05)                      |
| [`scoring/standard`](scorer/)    | Standaardscoring (1-0.5-0)                     |
| [`scoring/keizer`](scorer/)      | Keizer-scoring (iteratief)                     |
| [`scoring/football`](scorer/)    | Football-scoring (3-1-0)                       |
| [`tiebreaker`](tiebreaker/)      | 25 tiebreaker-implementaties + register        |
| [`trf`](trf/)                    | TRF16/TRF-2026 I/O en validatie                |
| [`factory`](overview/)           | Engines aanmaken op naam (`NewPairer`, `NewScorer`, `NewTieBreaker`) |
| [`standings`](overview/)         | Combineert Scorer en TieBreakers tot een presentatieklare tabel |
| [`algorithm/blossom`](overview/) | Edmonds' maximum weight matching               |
| [`algorithm/varma`](overview/)   | Varma-opzoektabellen (C.05 Annex 2)            |

Het rootpakket bevat ook `Parse*`-helpers voor de publieke enum-typen (`ParseScoringSystem`, `ParsePairingSystem`, `ParseGameResult`, `ParseByeType`) en `PlayedPairs(state, HistoryOptions{})` om de verzameling reeds gespeelde ongeordende paren af te leiden uit `TournamentState`.

## Subpagina's

- [Pakketorganisatie](overview/) -- hoe pakketten zich verhouden en de afhankelijkheidsstroom
- [Kerntypen](core-types/) -- `TournamentState`, `PlayerEntry`, `GameData`, resultaattypen
- [Pairer-interface](pairer/) -- het `Pairer`-contract, `PairingResult`, alle pairer-implementaties
- [Scorer-interface](scorer/) -- het `Scorer`-contract, `PlayerScore`, scoring-engines
- [TieBreaker-interface](tiebreaker/) -- het `TieBreaker`-contract, register, alle 25 tiebreakers
- [Optiepatroon](options/) -- hoe engine-opties werken (pointervelden, `WithDefaults`, `ParseOptions`)
- [Configuratie en enums](config/) -- `PairingSystem`, `ScoringSystem`, `DefaultTiebreakers()`
- [TRF-pakket](trf/) -- TRF16 lezen, schrijven, validatie en bidirectionele conversie
