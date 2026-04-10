---
title: "Configuratie en factory"
linkTitle: "Configuratie"
weight: 8
description: "PairingSystem- en ScoringSystem-enums, configuratiestructs en DefaultTiebreakers."
---

Het rootpakket definieert enumtypen en configuratiestructs die engine-selectie en -configuratie aansturen. Deze typen worden gebruikt door de CLI, het `trf`-pakket en elke aanroeper die een toernooi generiek wil configureren.

## PairingSystem

```go
type PairingSystem string
```

Geeft aan welk indelingsalgoritme gebruikt moet worden. Er zijn acht constanten gedefinieerd:

| Constante            | Waarde          | FIDE-reglement |
| -------------------- | --------------- | -------------- |
| `PairingDutch`       | `"dutch"`       | C.04.3         |
| `PairingBurstein`    | `"burstein"`    | C.04.4.2       |
| `PairingDubov`       | `"dubov"`       | C.04.4.1       |
| `PairingLim`         | `"lim"`         | C.04.4.3       |
| `PairingDoubleSwiss` | `"doubleswiss"` | C.04.5         |
| `PairingTeam`        | `"team"`        | C.04.6         |
| `PairingKeizer`      | `"keizer"`      | --             |
| `PairingRoundRobin`  | `"roundrobin"`  | C.05 Annex 1   |

### IsValid

```go
func (p PairingSystem) IsValid() bool
```

Retourneert `true` als `p` een van de acht erkende constanten is.

## ScoringSystem

```go
type ScoringSystem string
```

Geeft aan welk scoringsalgoritme gebruikt moet worden. Drie constanten:

| Constante         | Waarde       |
| ----------------- | ------------ |
| `ScoringStandard` | `"standard"` |
| `ScoringKeizer`   | `"keizer"`   |
| `ScoringFootball` | `"football"` |

### IsValid

```go
func (s ScoringSystem) IsValid() bool
```

Retourneert `true` als `s` een van de drie erkende constanten is.

## PairingConfig

```go
type PairingConfig struct {
    System  PairingSystem
    Options map[string]any
}
```

Bevat de selectie van het indelingssysteem en de engine-specifieke opties. De `Options`-map wordt direct doorgegeven aan de `NewFromMap()`-constructor van de engine. Welke sleutels en waarden geaccepteerd worden hangt af van de engine -- zie de pagina [Optiepatroon](../options/) voor details.

Voorbeeld:

```go
cfg := chesspairing.PairingConfig{
    System: chesspairing.PairingDutch,
    Options: map[string]any{
        "acceleration": "baku",
        "topSeedColor": "white",
    },
}
```

## ScoringConfig

```go
type ScoringConfig struct {
    System      ScoringSystem
    Tiebreakers []string
    Options     map[string]any
}
```

Bevat de selectie van het scoringssysteem, de geordende lijst van tiebreaker-ID's en engine-specifieke scoreropties.

- **System**: welk scoringsalgoritme gebruikt wordt.
- **Tiebreakers**: geordende lijst van tiebreaker-register-ID's (bijv. `"buchholz-cut1"`, `"sonneborn-berger"`). Worden in volgorde geëvalueerd om gelijke standen te breken.
- **Options**: wordt doorgegeven aan de `NewFromMap()` van de scorer.

Voorbeeld:

```go
cfg := chesspairing.ScoringConfig{
    System:      chesspairing.ScoringStandard,
    Tiebreakers: []string{"buchholz-cut1", "buchholz", "sonneborn-berger"},
    Options: map[string]any{
        "pointWin":  1.0,
        "pointDraw": 0.5,
    },
}
```

## DefaultTiebreakers

```go
func DefaultTiebreakers(system PairingSystem) []string
```

Retourneert de door de FIDE aanbevolen tiebreaker-volgorde voor het opgegeven indelingssysteem. Dit wordt als standaard gebruikt wanneer er geen tiebreakers expliciet zijn geconfigureerd.

| Indelingssysteem                                                | Standaard tiebreakers                                               |
| ------------------------------------------------------------- | ------------------------------------------------------------------- |
| Zwitsers (Dutch, Burstein, Dubov, Lim, Dubbel-Zwitsers, Team) | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Round-robin                                                   | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                                        | `games-played`, `direct-encounter`, `wins`                          |
| Overig/onbekend                                               | `direct-encounter`, `wins`                                          |

Gebruik:

```go
tbs := chesspairing.DefaultTiebreakers(chesspairing.PairingDutch)
// ["buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"]
```

## Alles samenvoegen

Een `TournamentState` bevat zowel een `PairingConfig` als een `ScoringConfig`. Samen beschrijven ze volledig hoe een toernooi ingedeeld en gescoord moet worden:

```go
state := chesspairing.TournamentState{
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
        Options: map[string]any{
            "totalRounds":  9,
            "acceleration": "baku",
        },
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System:      chesspairing.ScoringStandard,
        Tiebreakers: chesspairing.DefaultTiebreakers(chesspairing.PairingDutch),
    },
    // ... spelers, ronden, etc.
}
```

De CLI-factory en de functie `trf.ToTournamentState()` produceren beide `TournamentState`-waarden met deze configuraties ingevuld, zodat downstream-code engines generiek kan instantieren via `NewFromMap`.
