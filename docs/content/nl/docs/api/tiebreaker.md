---
title: "TieBreaker-interface"
linkTitle: "TieBreaker"
weight: 5
description: "De TieBreaker-interface, het zelfregistrerende register en het berekenen van tiebreakwaarden."
---

De `TieBreaker`-interface berekent een enkele numerieke tiebreakwaarde per speler. Het `tiebreaker`-pakket biedt 25 implementaties en een zelfregistrerend register voor opzoeking op ID.

## Interface-definitie

```go
type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

Drie methoden:

- **`ID`** -- Kort machine-identificatie (bijv. `"buchholz-cut1"`). Gebruikt in configuratie en het register.
- **`Name`** -- Leesbare weergavenaam (bijv. `"Buchholz Cut-1"`).
- **`Compute`** -- Neemt de toernooi-state en huidige scores (van een `Scorer`) en retourneert een `TieBreakValue` per speler. De scores-slice is nodig omdat veel tiebreakers afhangen van de scores van tegenstanders.

Alle engines accepteren `context.Context` voor toekomstige compatibiliteit. Aangezien alle berekeningen CPU-gebonden en in-memory zijn, wordt de context momenteel niet op annulering gecontroleerd.

### TieBreakValue

```go
type TieBreakValue struct {
    PlayerID string
    Value    float64
}
```

De geretourneerde slice bevat een vermelding per speler in `scores`, in dezelfde volgorde.

## Register

Tiebreakers registreren zichzelf via `init()`-functies. Het `tiebreaker`-pakket stelt drie registerfuncties beschikbaar:

```go
import "github.com/gnutterts/chesspairing/tiebreaker"

// Haal een tiebreaker op via ID.
tb, err := tiebreaker.Get("buchholz-cut1")

// Lijst alle geregistreerde ID's op (ongesorteerd).
ids := tiebreaker.All() // retourneert []string

// Registreer een aangepaste tiebreaker (roep alleen aan in init()).
tiebreaker.Register("my-tb", func() chesspairing.TieBreaker {
    return &myTieBreaker{}
})
```

Alle schrijfacties naar het register vinden plaats tijdens `init()`. Na het voltooien van de initialisatie is het register alleen-lezen en veilig voor gelijktijdige toegang zonder synchronisatie.

## Alle 25 geregistreerde tiebreakers

| ID                      | Naam                        | Beschrijving                                                        |
| ----------------------- | --------------------------- | ------------------------------------------------------------------- |
| `buchholz`              | Buchholz                    | Som van alle scores van tegenstanders                               |
| `buchholz-cut1`         | Buchholz Cut-1              | Laagste tegenstander-score weggelaten                               |
| `buchholz-cut2`         | Buchholz Cut-2              | Twee laagste tegenstander-scores weggelaten                         |
| `buchholz-median`       | Buchholz Median             | Hoogste en laagste tegenstander-score weggelaten                    |
| `buchholz-median2`      | Buchholz Median-2           | Twee hoogste en twee laagste tegenstander-scores weggelaten         |
| `sonneborn-berger`      | Sonneborn-Berger            | Som van tegenstander-scores gewogen naar resultaat tegen elk        |
| `direct-encounter`      | Direct Encounter            | Onderlinge score tussen gelijk gerangschikte spelers                |
| `wins`                  | Gewonnen partijen          | Gewonnen partijen; forfaitwinsten tellen mee bij vooraf vastgelegde indelingen |
| `win`                   | Gewonnen ronden             | OTB-winsten + forfait-winsten + PAB                                 |
| `black-games`           | Partijen met zwart          | Aantal partijen gespeeld als zwart; forfaitwinsten tellen mee bij vooraf vastgelegde indelingen |
| `black-wins`            | Zwart-winsten               | Winsten met de zwarte stukken; forfaitwinsten tellen mee bij vooraf vastgelegde indelingen |
| `rounds-played`         | Gespeelde ronden            | Totaal ronden waarin de speler deelnam                              |
| `standard-points`       | Standaardpunten             | Score volgens 1-0.5-0 ongeacht het scoringssysteem van het toernooi |
| `pairing-number`        | Rangnummer               | Rangnummer (TPN, lager is beter)                                 |
| `koya`                  | Koya-systeem                | Score tegen tegenstanders met >= 50% score                          |
| `progressive`           | Progressieve score          | Cumulatieve ronde-voor-ronde score                                  |
| `aro`                   | Gem. rating tegenstanders   | Gemiddelde rating van tegenstanders                                 |
| `fore-buchholz`         | Fore Buchholz               | Buchholz met lopende partijen als remise behandeld                  |
| `avg-opponent-buchholz` | Gem. Buchholz tegenstanders | Gemiddelde van Buchholz-scores van tegenstanders                    |
| `performance-rating`    | Prestatierating             | Toernooi-prestatierating (TPR)                                      |
| `performance-points`    | Prestatiepunten             | Toernooi-prestatiepunten (PTP)                                      |
| `avg-opponent-tpr`      | Gem. TPR tegenstanders      | Gemiddelde van TPR van tegenstanders (APRO)                         |
| `avg-opponent-ptp`      | Gem. PTP tegenstanders      | Gemiddelde van PTP van tegenstanders (APPO)                         |
| `player-rating`         | Spelersrating               | Eigen rating van de speler (RTNG)                                   |
| `games-played`          | Gespeelde partijen          | Totaal gespeelde partijen (forfaits uitgesloten)                    |

## Niet-gespeelde ronden

Tiebreakers bouwen één record per speler en ronde op. In Zwitserse toernooien deelt FIDE C.07:2026 artikel 16 niet-gespeelde ronden in en gebruikt het begrensde dummy-tegenstanders; forfaitwinsten en -verliezen, aangevraagde byes en afwezigheden zijn ongespeeld, terwijl volle-punt-byes als gespeeld tellen. In round-robins en andere vooraf vastgelegde indelingen behandelt artikel 15.2 forfaits als gewone ontmoetingen, behalve bij ratingtiebreakers en Type-B-forfaitverliezen. Hangende partijen zijn geen voltooide ontmoetingen.

Dit betekent:

- Een forfaitwinst draagt bij aan Buchholz, Sonneborn-Berger en andere op tegenstander-score gebaseerde berekeningen via de artikel 16-dummy (Zwitsers) of de geplande tegenstander (vooraf vastgelegde indelingen).
- Hangende partijen zijn geen voltooide ontmoetingen.
- Ratingtiebreakers gebruiken alleen tegenstanders die aan het bord zijn gespeeld, ook bij vooraf vastgelegde indelingen.

## DefaultTiebreakers

Het rootpakket biedt standaard tiebreaker-volgordes van de bibliotheek per indelingssysteem. De hoofdorganisator kiest de feitelijke volgorde onder FIDE C.07:2026 artikel 4.1; de bibliotheekstandaarden worden gebruikt wanneer geen lijst is geconfigureerd. De teamstandaarden gaan uit van matchpunten als primaire score (artikel 13):

```go
import "github.com/gnutterts/chesspairing"

tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// Retourneert: ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

| Indelingssysteem                              | Standaard tiebreakers                                               |
| --------------------------------------------- | ------------------------------------------------------------------- |
| Dutch, Burstein, Dubov, Lim, Dubbel-Zwitsers  | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Team-Zwitsers                                 | `buchholz-mp-cut1`, `buchholz-mp`, `emmsb`, `mpvgp`                 |
| Round-robin                                   | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                        | `games-played`, `direct-encounter`, `wins`                          |

Zie [Tiebreakers](/docs/tiebreakers/) voor gedetailleerde uitleg van elk algoritme.

## Gebruiksvoorbeeld

Scoor het toernooi en bereken vervolgens tiebreakers in volgorde om de eindstand op te bouwen:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/scoring/standard"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

func main() {
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2400},
            {ID: "2", DisplayName: "Bob",     Rating: 2350},
            {ID: "3", DisplayName: "Charlie", Rating: 2300},
            {ID: "4", DisplayName: "Diana",   Rating: 2250},
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
        CurrentRound:  2,
        PairingConfig: chesspairing.PairingConfig{System: chesspairing.PairingDutch},
    }

    // Stap 1: Bereken scores.
    scorer := standard.NewFromMap(nil)
    scores, err := scorer.Score(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    // Stap 2: Bereken tiebreakers in de geconfigureerde bibliotheekvolgorde.
    tbIDs := chesspairing.DefaultTiebreakers(state.PairingConfig.System)
    tbResults := make(map[string][]chesspairing.TieBreakValue, len(tbIDs))

    for _, id := range tbIDs {
        tb, err := tiebreaker.Get(id)
        if err != nil {
            log.Fatal(err)
        }
        values, err := tb.Compute(context.Background(), state, scores)
        if err != nil {
            log.Fatal(err)
        }
        tbResults[id] = values
    }

    // Stap 3: Toon de stand met tiebreakwaarden.
    for _, ps := range scores {
        fmt.Printf("Rang %d: %s (%.1f ptn)", ps.Rank, ps.PlayerID, ps.Score)
        for _, id := range tbIDs {
            for _, tv := range tbResults[id] {
                if tv.PlayerID == ps.PlayerID {
                    fmt.Printf("  %s=%.2f", id, tv.Value)
                    break
                }
            }
        }
        fmt.Println()
    }
}
```

## Gegevensstroom

```text
Scorer.Score(ctx, state) retourneert []PlayerScore
  -> geef scores door aan elke TieBreaker.Compute(ctx, state, scores)
     -> retourneert []TieBreakValue (een per speler)
  -> combineer tot []Standing voor de uiteindelijke gerangschikte uitvoer
```

Het `Standing`-type combineert scores en tiebreakers in een enkele gerangschikte uitvoer:

```go
type Standing struct {
    Rank        int          `json:"rank"`
    PlayerID    string       `json:"playerId"`
    DisplayName string       `json:"displayName"`
    Score       float64      `json:"score"`
    TieBreakers []NamedValue `json:"tieBreakers"`
    GamesPlayed int          `json:"gamesPlayed"`
    Wins        int          `json:"wins"`
    Draws       int          `json:"draws"`
    Losses      int          `json:"losses"`
}
```
