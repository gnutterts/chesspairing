---
title: "Scorer-interface"
linkTitle: "Scorer"
weight: 4
description: "De Scorer-interface en de drie scoringsimplementaties."
---

De `Scorer`-interface berekent de stand op basis van partijresultaten. Drie implementaties dekken de belangrijkste scoringssystemen: Standaard (1-0.5-0), Keizer (iteratief rangafhankelijk) en Football (3-1-0).

## Interface-definitie

```go
type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}
```

Twee methoden:

- **`Score`** -- Berekent scores voor alle actieve spelers in het toernooi. Retourneert een `PlayerScore` per actieve speler, elk met `PlayerID`, `Score` en `Rank`. Spelers worden gerangschikt op aflopende score, met rating als secundaire tiebreak.
- **`PointsForResult`** -- Retourneert de puntwaarde voor een enkel partijresultaat gegeven de context. Pairers roepen dit intern aan wanneer ze scoringsinformatie nodig hebben (bijv. de Keizer-pairer heeft Keizer-scores nodig voor rangschikking). De `ResultContext` biedt informatie over rang van tegenstander/speler die nodig is voor Keizer-scoring; Standaard en Football negeren dit.

Alle engines accepteren `context.Context` voor toekomstige compatibiliteit. Aangezien alle berekeningen CPU-gebonden en in-memory zijn, wordt de context momenteel niet op annulering gecontroleerd.

### ResultContext

```go
type ResultContext struct {
    OpponentRank        int
    OpponentValueNumber int
    PlayerRank          int
    PlayerValueNumber   int
    IsBye               bool
    IsAbsent            bool
    IsForfeit           bool
}
```

De `ResultContext`-vlaggen (`IsBye`, `IsAbsent`, `IsForfeit`) bepalen welke puntwaarde wordt toegepast. De rang- en waardenummervelden worden uitsluitend door Keizer-scoring gebruikt.

## Implementaties

Elke scoring-engine biedt twee constructors:

- `New(opts Options) *Scorer` -- getypeerde options-struct.
- `NewFromMap(m map[string]any) *Scorer` -- generieke map (van JSON-configuratie of TRF).

Beide passen standaardwaarden toe voor niet-ingestelde (nil) velden. Zie [Optiepatroon](../options/) voor de nil-betekent-standaard-conventie.

### Standaard

**Pakket:** `github.com/gnutterts/chesspairing/scoring/standard`

Standaard FIDE-scoring: vaste punten per resultaat, onafhankelijk van de sterkte van de tegenstander. Enkele doorgang, deterministisch.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/standard"

// Vanuit een generieke optiemap:
scorer := standard.NewFromMap(nil) // alle standaardwaarden

// Vanuit een getypeerde Options-struct:
scorer := standard.New(standard.Options{
    PointWin:  chesspairing.Float64Ptr(1.0),
    PointDraw: chesspairing.Float64Ptr(0.5),
})
```

**Opties:**

| Veld               | Type       | Standaard | Beschrijving                                |
| ------------------ | ---------- | --------- | ------------------------------------------- |
| `PointWin`         | `*float64` | `1.0`     | Punten voor een overwinning                 |
| `PointDraw`        | `*float64` | `0.5`     | Punten voor remise                          |
| `PointLoss`        | `*float64` | `0.0`     | Punten voor verlies                         |
| `PointBye`         | `*float64` | `1.0`     | Punten voor een indelings-toegewezen bye      |
| `PointForfeitWin`  | `*float64` | `1.0`     | Punten voor een forfait-overwinning         |
| `PointForfeitLoss` | `*float64` | `0.0`     | Punten voor een forfait-verlies             |
| `PointAbsent`      | `*float64` | `0.0`     | Punten bij afwezigheid (geen partij of bye) |

Half-punt byes gebruiken `PointDraw`, nul-punt byes gebruiken `PointLoss`, en afwezigheidsbyes gebruiken `PointAbsent`. Dubbele forfaits kennen nul toe aan beide spelers.

### Keizer

**Pakket:** `github.com/gnutterts/chesspairing/scoring/keizer`

Bij Keizer-scoring krijgt elke speler een waardenummer gebaseerd op zijn huidige rang. Een sterke tegenstander verslaan (hoog waardenummer) levert meer punten op dan een zwakkere verslaan. Afwezigheden ontvangen een fractie van het eigen waardenummer van de speler.

Het algoritme is iteratief: scores bepalen rangschikkingen, die waardenummers bepalen, die scores veranderen. Het convergeert binnen 20 iteraties met x2-integer-rekenkunde (verdubbelde scores) om drijvende-komma-drift te elimineren. Een 2-cyclus oscillatiedetector middelt afwisselende rangschikkingen wanneer convergentie stagneert.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/keizer"

// Vanuit een generieke optiemap:
scorer := keizer.NewFromMap(nil) // KeizerForClubs standaardwaarden

// Vanuit een getypeerde Options-struct:
scorer := keizer.New(keizer.Options{
    WinFraction:  chesspairing.Float64Ptr(1.0),
    DrawFraction: chesspairing.Float64Ptr(0.5),
    SelfVictory:  chesspairing.BoolPtr(true),
})
```

**Belangrijke opties (24 totaal):**

| Veld                     | Type       | Standaard      | Beschrijving                                                                           |
| ------------------------ | ---------- | -------------- | -------------------------------------------------------------------------------------- |
| `ValueNumberBase`        | `*int`     | aantal spelers | Waardenummer van de hoogst gerangschikte speler                                        |
| `ValueNumberStep`        | `*int`     | `1`            | Afname per rangpositie                                                                 |
| `WinFraction`            | `*float64` | `1.0`          | Fractie van tegenstanders waardenummer bij winst                                       |
| `DrawFraction`           | `*float64` | `0.5`          | Fractie van tegenstanders waardenummer bij remise                                      |
| `LossFraction`           | `*float64` | `0.0`          | Fractie van tegenstanders waardenummer bij verlies                                     |
| `ForfeitWinFraction`     | `*float64` | `1.0`          | Fractie van tegenstanders waardenummer bij forfait-winst                               |
| `ForfeitLossFraction`    | `*float64` | `0.0`          | Fractie van tegenstanders waardenummer bij forfait-verlies                             |
| `DoubleForfeitFraction`  | `*float64` | `0.0`          | Fractie van tegenstanders waardenummer bij dubbel forfait                              |
| `ByeValueFraction`       | `*float64` | `0.50`         | Fractie van eigen waardenummer bij een PAB                                             |
| `HalfByeFraction`        | `*float64` | `0.50`         | Fractie van eigen waardenummer bij een half-punt bye                                   |
| `ZeroByeFraction`        | `*float64` | `0.0`          | Fractie van eigen waardenummer bij een nul-punt bye                                    |
| `AbsentPenaltyFraction`  | `*float64` | `0.35`         | Fractie van eigen waardenummer bij ongeoorloofde afwezigheid                           |
| `ExcusedAbsentFraction`  | `*float64` | `0.35`         | Fractie van eigen waardenummer bij verontschuldigde afwezigheid                        |
| `ClubCommitmentFraction` | `*float64` | `0.70`         | Fractie van eigen waardenummer bij afwezigheid door intercompetitie                    |
| `SelfVictory`            | `*bool`    | `true`         | Eigen waardenummer optellen bij totaal (eenmalig, niet per ronde)                      |
| `AbsenceLimit`           | `*int`     | `5`            | Max. afwezigheden die punten opleveren (0 = onbeperkt). Clubverplichtingen vrijgesteld |
| `AbsenceDecay`           | `*bool`    | `false`        | Halveer afwezigheidsbonus voor elke volgende afwezigheid                               |
| `LateJoinHandicap`       | `*float64` | `0`            | Gereserveerd voor toekomstig gebruik (gedeclareerd maar nog niet geïmplementeerd)      |

Zes vaste-waarde-overschrijvingsvelden (`ByeFixedValue`, `HalfByeFixedValue`, `ZeroByeFixedValue`, `AbsentFixedValue`, `ExcusedAbsentFixedValue`, `ClubCommitmentFixedValue`) vervangen de corresponderende fractieberekening door een vast geheel getal wanneer ze niet nil zijn.

**Variant-presets:**

- **KeizerForClubs (standaard):** Alles nil -- gebruikt de hierboven vermelde standaardwaarden.
- **Klassiek KNSB zesden:** `WinFraction=1`, `DrawFraction=0.5`, `LossFraction=0`, `ByeValueFraction=4/6`, `AbsentPenaltyFraction=2/6`, `ClubCommitmentFraction=2/3`, `ExcusedAbsentFraction=2/6`, `AbsenceLimit=5`.
- **FreeKeizer:** `LossFraction=1/6`, `ByeValueFraction=4/6`, `AbsentPenaltyFraction=2/6`, `AbsenceLimit=5`.

### Football

**Pakket:** `github.com/gnutterts/chesspairing/scoring/football`

Dunne wrapper rond standaardscoring met football-standaardwaarden: 3 voor winst, 1 voor remise, 0 voor verlies. Beloont beslissende resultaten sterker dan standaardscoring.

Football gebruikt `standard.Options` rechtstreeks -- er is geen apart `football.Options`-type. Alle standaardopties zijn beschikbaar omdat Football intern volledig delegeert aan de standaardscorer.

**Constructors:**

```go
import "github.com/gnutterts/chesspairing/scoring/football"

// Vanuit een generieke optiemap:
scorer := football.NewFromMap(nil) // football standaardwaarden (3-1-0)

// Vanuit een getypeerde Options-struct (gebruikt standard.Options):
scorer := football.New(standard.Options{
    PointWin: chesspairing.Float64Ptr(3.0),
})
```

**Football standaardwaarden (vs. standaard):**

| Veld               | Football | Standaard |
| ------------------ | -------- | --------- |
| `PointWin`         | `3.0`    | `1.0`     |
| `PointDraw`        | `1.0`    | `0.5`     |
| `PointLoss`        | `0.0`    | `0.0`     |
| `PointBye`         | `3.0`    | `1.0`     |
| `PointForfeitWin`  | `3.0`    | `1.0`     |
| `PointForfeitLoss` | `0.0`    | `0.0`     |
| `PointAbsent`      | `0.0`    | `0.0`     |

## Compile-time interface-controle

Elk scoring-pakket bevat een compile-time assertie:

```go
var _ chesspairing.Scorer = (*Scorer)(nil)
```

Dit garandeert dat de implementatie de `Scorer`-interface op compileertijd voldoet.

## Gebruiksvoorbeeld

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/scoring/standard"
)

func main() {
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2400, Active: true},
            {ID: "2", DisplayName: "Bob",     Rating: 2350, Active: true},
            {ID: "3", DisplayName: "Charlie", Rating: 2300, Active: true},
            {ID: "4", DisplayName: "Diana",   Rating: 2250, Active: true},
        },
        Rounds: []chesspairing.RoundData{
            {
                Number: 1,
                Games: []chesspairing.GameData{
                    {WhiteID: "1", BlackID: "4", Result: chesspairing.ResultWhiteWins},
                    {WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
                },
            },
        },
        CurrentRound: 2,
    }

    scorer := standard.NewFromMap(nil)

    scores, err := scorer.Score(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    for _, ps := range scores {
        fmt.Printf("Rang %d: %s (%.1f ptn)\n", ps.Rank, ps.PlayerID, ps.Score)
    }
    // Uitvoer:
    // Rang 1: 1 (1.0 ptn)
    // Rang 2: 2 (0.5 ptn)
    // Rang 3: 3 (0.5 ptn)
    // Rang 4: 4 (0.0 ptn)
}
```

## Foutafhandeling

Scorers retourneren een fout bij ongeldige invoertoestanden. Als het toernooi geen spelers heeft, retourneert `Score` `nil, nil` (geen fout).

Scorers paniken nooit. Alle uitzonderlijke condities worden gemeld via de geretourneerde `error`-waarde.

## Gegevensstroom

```text
Aanroeper bouwt TournamentState
  -> Scorer.Score(ctx, state) retourneert []PlayerScore
     -> PlayerScore.PlayerID: spelersidentificatie
     -> PlayerScore.Score: totaal punten
     -> PlayerScore.Rank: 1-gebaseerde rangpositie

  -> Scorer.PointsForResult(result, rctx) retourneert float64
     -> Intern gebruikt door pairers voor scoregroepvorming
```

De `[]PlayerScore`-slice is geordend op rang (index 0 = rang 1). Geef deze slice door aan `TieBreaker.Compute()` om tiebreakwaarden te berekenen.
