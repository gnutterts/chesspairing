---
title: "Go-bibliotheek snelstart"
linkTitle: "Go snelstart"
weight: 3
description: "Voeg chesspairing toe aan je Go-project en genereer je eerste indeling programmatisch."
---

Deze handleiding laat stap voor stap zien hoe je chesspairing toevoegt aan een
Go-project, een toernooistatus opbouwt, indelingen genereert, rondes scoort en
tiebreakers berekent. Aan het eind heb je een werkend programma dat een
Zwitsers toernooi met vier spelers indeelt.

## Vereisten

- Go 1.24 of nieuwer

chesspairing is een pure Go-module zonder externe afhankelijkheden.

## Installatie

Voeg de module toe aan je project:

```bash
go get github.com/gnutterts/chesspairing
```

Importeer vervolgens het root-package en de engine-packages die je nodig hebt:

```go
import (
    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
)
```

## Kerninterfaces

Alle engines implementeren een van drie interfaces die in het root-package
gedefinieerd zijn:

```go
// Pairer generates pairings for a round given tournament state.
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

// Scorer calculates standings from game results.
type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}

// TieBreaker computes a single tiebreak value for each player.
type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

Indeling en scoring zijn onafhankelijk van elkaar -- een toernooi kan elke
combinatie gebruiken (bijvoorbeeld Zwitserse indeling met Keizer-scoring).

## Een TournamentState opbouwen

Elke engine-methode ontvangt een `*TournamentState`. Dit is een read-only
snapshot die je uit je eigen databron samenstelt:

```go
state := &chesspairing.TournamentState{
    Players: []chesspairing.PlayerEntry{
        {ID: "1", DisplayName: "Alice",   Rating: 2100},
        {ID: "2", DisplayName: "Bob",     Rating: 1950},
        {ID: "3", DisplayName: "Charlie", Rating: 1800},
        {ID: "4", DisplayName: "Diana",   Rating: 1750},
    },
    Rounds:       nil, // nog geen rondes gespeeld
    CurrentRound: 0,
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System:      chesspairing.ScoringStandard,
        Tiebreakers: []string{"buchholz-cut1", "buchholz", "sonneborn-berger"},
    },
}
```

`PlayerEntry` heeft extra optionele velden zoals `Federation`, `FideID`,
`Title`, `Sex` en `BirthDate`. Vul alleen in wat je hebt. Om een speler
permanent af te melden na een ronde `N`, zet `WithdrawnAfterRound = &N`;
de speler wordt dan uitgesloten van indeling in elke ronde groter dan
`N`. Gebruik `state.IsActiveInRound(spelerID, ronde)` en
`state.ActivePlayerIDs(ronde)` om de actieve verzameling op te vragen.

`Rounds` bevat een `[]RoundData` met voltooide partijresultaten. Voor een
nieuw toernooi is dit nil of leeg.

## Indelingen genereren

Maak een pairer aan en roep `Pair` aan:

```go
pairer := dutch.New(dutch.Options{})
result, err := pairer.Pair(context.Background(), state)
if err != nil {
    log.Fatal(err)
}

for _, p := range result.Pairings {
    fmt.Printf("Board %d: %s (White) vs %s (Black)\n", p.Board, p.WhiteID, p.BlackID)
}
for _, bye := range result.Byes {
    fmt.Printf("Bye: %s (%s)\n", bye.PlayerID, bye.Type)
}
```

`PairingResult` bevat `Pairings` (een slice van `GamePairing` met `Board`,
`WhiteID`, `BlackID`), `Byes` (een slice van `ByeEntry`) en optioneel `Notes`.

### Beschikbare pairers

| Systeem       | Package               | FIDE-reglement |
| ------------- | --------------------- | -------------- |
| Dutch         | `pairing/dutch`       | C.04.3         |
| Burstein      | `pairing/burstein`    | C.04.4.2       |
| Dubov         | `pairing/dubov`       | C.04.4.1       |
| Lim           | `pairing/lim`         | C.04.4.3       |
| Double-Swiss  | `pairing/doubleswiss` | C.04.5         |
| Team-Zwitsers | `pairing/team`        | C.04.6         |
| Keizer        | `pairing/keizer`      | --             |
| Round-robin   | `pairing/roundrobin`  | C.05 Annex 1   |

Alle pairers volgen hetzelfde patroon: `New(Options{})` of `NewFromMap(map[string]any)`.

## Het toernooi scoren

Na het invoeren van de partijresultaten scoor je de ronde:

```go
scorer := standard.New(standard.Options{})

// Add round 1 results to the state.
state.Rounds = []chesspairing.RoundData{
    {
        Number: 1,
        Games: []chesspairing.GameData{
            {WhiteID: "1", BlackID: "4", Result: chesspairing.ResultWhiteWins},
            {WhiteID: "2", BlackID: "3", Result: chesspairing.ResultDraw},
        },
    },
}
state.CurrentRound = 1

scores, err := scorer.Score(context.Background(), state)
if err != nil {
    log.Fatal(err)
}
for _, s := range scores {
    fmt.Printf("Player %s: %.1f pts (rank %d)\n", s.PlayerID, s.Score, s.Rank)
}
```

`Score` retourneert `[]PlayerScore` gesorteerd op rang. Elk element bevat
`PlayerID`, `Score` en `Rank`.

### Beschikbare scorers

| Systeem  | Package            | Standaardpunten                   |
| -------- | ------------------ | --------------------------------- |
| Standard | `scoring/standard` | 1 - 0.5 - 0                       |
| Football | `scoring/football` | 3 - 1 - 0                         |
| Keizer   | `scoring/keizer`   | Iteratief rangschikkingsgebaseerd |

## Tiebreakers berekenen

Tiebreakers worden opgezocht in een globaal register aan de hand van hun ID:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
if err != nil {
    log.Fatal(err)
}

values, err := tb.Compute(context.Background(), state, scores)
if err != nil {
    log.Fatal(err)
}
for _, v := range values {
    fmt.Printf("Player %s: Buchholz Cut 1 = %.1f\n", v.PlayerID, v.Value)
}
```

Er zijn 25 geregistreerde tiebreakers. Veelgebruikte ID's zijn `buchholz`,
`buchholz-cut1`, `sonneborn-berger`, `direct-encounter`, `wins`, `aro`,
`performance-rating` en `koya`. Gebruik `tiebreaker.All()` om ze allemaal
op te vragen.

## Engine-opties

Elke engine heeft een `Options` struct met pointer-velden. Een nil-veld
betekent "gebruik de standaardwaarde." Geef een lege struct mee voor
standaardgedrag:

```go
// All defaults:
pairer := dutch.New(dutch.Options{})

// Override one setting:
accel := "baku"
pairer := dutch.New(dutch.Options{
    Acceleration: &accel,
})
```

Voor dynamische configuratie (JSON, configuratiebestanden) gebruik je
`NewFromMap`:

```go
pairer := dutch.NewFromMap(map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
})
```

## Volledig voorbeeld

Het volgende programma maakt een toernooi met vier spelers aan, genereert de
indeling voor ronde 1, voert resultaten in, scoort de ronde en berekent een
tiebreaker:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

func main() {
    ctx := context.Background()

    // 1. Define players and build tournament state.
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2100},
            {ID: "2", DisplayName: "Bob",     Rating: 1950},
            {ID: "3", DisplayName: "Charlie", Rating: 1800},
            {ID: "4", DisplayName: "Diana",   Rating: 1750},
        },
        CurrentRound: 0,
        PairingConfig: chesspairing.PairingConfig{
            System: chesspairing.PairingDutch,
        },
        ScoringConfig: chesspairing.ScoringConfig{
            System: chesspairing.ScoringStandard,
        },
    }

    // 2. Generate round 1 pairings.
    pairer := dutch.New(dutch.Options{})
    result, err := pairer.Pair(ctx, state)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Round 1 Pairings:")
    for _, p := range result.Pairings {
        fmt.Printf("  Board %d: %s (W) vs %s (B)\n", p.Board, p.WhiteID, p.BlackID)
    }

    // 3. Record results and add round to state.
    //    (In a real application, results come from user input.)
    state.Rounds = []chesspairing.RoundData{
        {
            Number: 1,
            Games:  toGameData(result.Pairings),
            Byes:   result.Byes,
        },
    }
    state.CurrentRound = 1

    // 4. Score the round.
    scorer := standard.New(standard.Options{})
    scores, err := scorer.Score(ctx, state)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\nStandings after Round 1:")
    for _, s := range scores {
        fmt.Printf("  %d. Player %s — %.1f pts\n", s.Rank, s.PlayerID, s.Score)
    }

    // 5. Compute a tiebreaker.
    tb, err := tiebreaker.Get("buchholz-cut1")
    if err != nil {
        log.Fatal(err)
    }
    values, err := tb.Compute(ctx, state, scores)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\nBuchholz Cut 1:")
    for _, v := range values {
        fmt.Printf("  Player %s: %.1f\n", v.PlayerID, v.Value)
    }
}

// toGameData converts pairings into game results for demonstration purposes.
// The first board is a white win, the rest are draws.
func toGameData(pairings []chesspairing.GamePairing) []chesspairing.GameData {
    games := make([]chesspairing.GameData, len(pairings))
    for i, p := range pairings {
        result := chesspairing.ResultDraw
        if i == 0 {
            result = chesspairing.ResultWhiteWins
        }
        games[i] = chesspairing.GameData{
            WhiteID: p.WhiteID,
            BlackID: p.BlackID,
            Result:  result,
        }
    }
    return games
}
```

## Volgende stappen

- [API-referentie](/docs/api/) -- volledige documentatie van alle types,
  interfaces en engine-opties
- [CLI-snelstart](../cli-quickstart/) -- gebruik chesspairing vanaf de
  opdrachtregel met TRF16-bestanden
- Bekijk de beschikbare [indelingssystemen](/docs/pairing-systems/) en
  [scoringssystemen](/docs/scoring/) om de juiste configuratie voor je
  toernooi te vinden
