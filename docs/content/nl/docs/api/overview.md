---
title: "Pakketorganisatie"
linkTitle: "Overzicht"
weight: 1
description: "Hoe de chesspairing-module in pakketten is georganiseerd en hoe ze zich tot elkaar verhouden."
---

## Rootpakket

Het `chesspairing`-pakket definieert drie interfaces (`Pairer`, `Scorer`, `TieBreaker`) en alle gedeelde typen (`TournamentState`, `PlayerEntry`, `RoundData`, `GameData`, `PairingResult`, `PlayerScore`, `TieBreakValue`, etc.). Het bevat geen implementatiecode -- alleen contracten en datastructuren.

Configuratie-enums (`PairingSystem`, `ScoringSystem`, `GameResult`, `ByeType`) en helperfuncties (`DefaultTiebreakers()`) staan hier ook. Elke enum heeft een `Parse*`-constructor (`ParsePairingSystem`, `ParseScoringSystem`, `ParseGameResult`, `ParseByeType`) om configuratie vanuit strings te kunnen ronddraaien. `PlayedPairs(state, HistoryOptions{})` retourneert de verzameling reeds gespeelde ongeordende spelersparen, geschikt als verboden-paar-beperking bij het bouwen van de volgende ronde; `HistoryOptions.IncludeForfeits` bepaalt of enkele forfait-partijen als gespeeld tellen.

## Engine-pakketten

Elke pairer en scorer staat in een eigen pakket. Elk engine-pakket volgt dezelfde structuur:

- Een publieke `Pairer`- of `Scorer`-struct.
- Een `Options`-struct met pointervelden. Een nil-veld betekent "gebruik de standaardwaarde."
- `WithDefaults()`-methode -- vult nil-velden met verstandige standaardwaarden.
- `ParseOptions(map[string]any)` -- parst opties uit een generieke map (doorgaans vanuit JSON-configuratie).
- `NewFromMap(map[string]any)` -- constructor voor factory-instantiatie. Retourneert de engine met geparsede opties en toegepaste standaardwaarden.
- Compile-time interface-controle:

```go
var _ chesspairing.Pairer = (*Pairer)(nil)
```

### Pairer-pakketten

| Pakket                | Systeem             | Specificatie  |
| --------------------- | ------------------- | ------------- |
| `pairing/dutch`       | Nederlands Zwitsers | FIDE C.04.3   |
| `pairing/burstein`    | Burstein Zwitsers   | FIDE C.04.4.2 |
| `pairing/dubov`       | Dubov Zwitsers      | FIDE C.04.4.1 |
| `pairing/lim`         | Lim Zwitsers        | FIDE C.04.4.3 |
| `pairing/doubleswiss` | Dubbel-Zwitsers     | FIDE C.04.5   |
| `pairing/team`        | Team-Zwitsers       | FIDE C.04.6   |
| `pairing/keizer`      | Keizer              | --            |
| `pairing/roundrobin`  | Round-robin         | FIDE C.05     |

### Scorer-pakketten

| Pakket             | Systeem                                                 |
| ------------------ | ------------------------------------------------------- |
| `scoring/standard` | Standaard (1-0.5-0, configureerbare puntwaarden)        |
| `scoring/keizer`   | Keizer (iteratieve convergentie, variant-ondersteuning) |
| `scoring/football` | Football (3-1-0, dunne wrapper rond standaard)          |

## Gedeelde bibliotheken

Twee interne bibliotheken bieden gedeelde logica voor Zwitserse pairers. Aanroepers importeren deze nooit rechtstreeks -- ze werken uitsluitend via de publieke pairer-interfaces.

### `pairing/swisslib`

Gebruikt door de Dutch- en Burstein-pairers. Biedt:

- `PlayerState`-constructie vanuit toernooigeschiedenis
- Scoregroepen en brackets
- Bye-selectie (op basis van voltooibaarheid)
- Kleurvoorkeur en -toewijzing
- Absolute criteria (C1--C4) en optimalisatiecriteria (C8--C21)
- Randgewichtberekening voor Blossom-matching (`*big.Int`)
- `PairBracketsGlobal()` -- globale Blossom-matching met Stage 0.5 voltooibaarheids-pre-matching
- Structurele validatie

### `pairing/lexswiss`

Gebruikt door de Dubbel-Zwitserse en Team-Zwitserse pairers. Biedt:

- `ParticipantState`-constructie
- Scoregroepen
- Bye-toewijzing en up-floater-selectie
- `PairBracket()` -- lexicografische bracket-indeling met inplugbare criteria

## Ondersteunende pakketten

### `tiebreaker`

Zelfregistrerend register. Elk tiebreaker-bestand roept `Register()` aan in zijn `init()`-functie. Het register biedt `Get(name)` om een tiebreaker op ID op te halen en `All()` om beschikbare tiebreakers op te sommen. Er zijn 25 tiebreakers geregistreerd.

### `trf`

Bidirectionele TRF-conversie:

- `Read(io.Reader)` -- parst een TRF16-document.
- `Write(io.Writer, *Document)` -- serialiseert een TRF16-document.
- `ToTournamentState()` -- converteert een TRF-`Document` naar een `chesspairing.TournamentState`.
- `FromTournamentState()` -- converteert een `TournamentState` terug naar een TRF-`Document`.
- `Document.Validate()` -- multi-profiel validatie (Algemeen, Indelingsengine, FIDE).

### `factory`

Maakt engines aan op naam vanuit een generieke configuratiemap. Drie ingangen:

- `NewPairer(name string, opts map[string]any)` -- retourneert een geconfigureerde `chesspairing.Pairer`.
- `NewScorer(name string, opts map[string]any)` -- retourneert een geconfigureerde `chesspairing.Scorer`.
- `NewTieBreaker(name string)` -- zoekt een geregistreerde tiebreaker op via diens ID.

Handig wanneer het indelings- of scoring-systeem tijdens runtime wordt gekozen vanuit JSON of CLI-vlaggen in plaats van bij compileren te worden ingebouwd.

### `standings`

Combineert een `Scorer` en een reeks `TieBreaker`s tot een presentatieklare tabel. `Build(ctx, state, scorer, tieBreakers)` draait de scoring, draait elke tiebreaker, sorteert op score en daarna op tiebreaker-kolommen in volgorde, en retourneert `[]Standing` waarin echte gelijkstanden dezelfde rang delen (standaard "1224"-rangschikking). `BuildByID(ctx, state, scorer, tbIDs)` lost tiebreaker-ID's op via het register als gemak. Winst, remise en verlies worden uit de partijresultaten afgeleid, omdat W/R/V losstaat van de scoring-regel.

### `algorithm/blossom`

Op zichzelf staande implementatie van Edmonds' maximum weight matching voor algemene grafen. O(n^3). Twee varianten:

- `MaxWeightMatching(edges, maxCardinality)` -- int64-gewichten.
- `MaxWeightMatchingBig(edges, maxCardinality)` -- `*big.Int`-gewichten (nodig wanneer randgewicht-bitindelingen 64 bits overschrijden).

Geporteerd van Joris van Rantwijks Python-referentie-implementatie.

### `algorithm/varma`

Op zichzelf staande opzoektabellen uit FIDE C.05 Annex 2. Gebruikt voor federatiebewuste toewijzing van rangnummers bij round-robin-toernooien.

- `Groups()` -- retourneert Varma-groeperingen voor een gegeven aantal spelers.
- `Assign()` -- wijst rangnummers toe met inachtneming van federatiebeperkingen.

## Afhankelijkheidsstroom

```text
Aanroepende code
  -> chesspairing (interfaces + typen)
  -> pairing/* (implementeert Pairer)
      -> pairing/swisslib of pairing/lexswiss (gedeelde logica)
      -> algorithm/blossom (matching)
  -> scoring/* (implementeert Scorer)
  -> tiebreaker (implementeert TieBreaker)
  -> trf (I/O-laag)
```

Alle pijlen wijzen naar binnen, naar het rootpakket. Engine-pakketten hangen af van het rootpakket en hun gedeelde bibliotheken, maar nooit van elkaar. Een pairer-pakket importeert nooit een scorer-pakket en andersom. Het `trf`-pakket hangt af van het rootpakket voor typedefinities, maar is verder onafhankelijk van de engine-pakketten.
