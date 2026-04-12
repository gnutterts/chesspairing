---
title: "Pairer-interface"
linkTitle: "Pairer"
weight: 3
description: "De Pairer-interface en het gebruik van elk van de acht indelingsimplementaties."
---

De `Pairer`-interface genereert indelingen voor een toernooironde. Acht implementaties dekken alle gangbare schaak-indelingssystemen.

## Interface-definitie

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}
```

`Pair` accepteert een `context.Context` en een pointer naar `TournamentState`. De state wordt als alleen-lezen behandeld en wordt nooit door de engine gewijzigd. Het geretourneerde `PairingResult` bevat bordtoewijzingen (`Pairings`) en eventuele byes (`Byes`).

Alle engines accepteren `context.Context` voor toekomstige compatibiliteit. Aangezien alle berekeningen CPU-gebonden en in-memory zijn, wordt de context momenteel niet op annulering gecontroleerd.

## Implementaties

Elke indelingsengine staat in een eigen pakket onder `pairing/` en biedt twee constructors:

- `New(opts Options) *Pairer` -- getypeerde options-struct.
- `NewFromMap(m map[string]any) *Pairer` -- generieke map (van JSON-configuratie of TRF).

Beide passen standaardwaarden toe voor niet-ingestelde (nil) velden. Zie [Optiepatroon](../options/) voor de nil-betekent-standaard-conventie.

### Een pairer aanmaken

```go
import "github.com/gnutterts/chesspairing/pairing/dutch"

// Vanuit een generieke optiemap (bijv. geparsed uit JSON-configuratie):
p := dutch.NewFromMap(nil) // alle standaardwaarden

// Vanuit een getypeerde Options-struct (string-opties vereisen een pointer;
// het pakket biedt geen StringPtr-helper, dus gebruik een lokale variabele):
accel := "baku"
color := "white"
p := dutch.New(dutch.Options{
    Acceleration: &accel,
    TopSeedColor: &color,
})
```

Alle acht engines volgen hetzelfde patroon. Vervang `dutch` door de naam van het gewenste pakket.

### Compile-time interface-controle

Elk engine-pakket bevat een compile-time assertie:

```go
var _ chesspairing.Pairer = (*Pairer)(nil)
```

Dit garandeert dat de implementatie de `Pairer`-interface op compileertijd voldoet.

## Opties per engine

| Engine          | Pakket                | Belangrijke opties                                                                                                                                                                              |
| --------------- | --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Dutch           | `pairing/dutch`       | `Acceleration` (`*string`: `"none"`, `"baku"`), `TopSeedColor` (`*string`: `"auto"`, `"white"`, `"black"`), `ForbiddenPairs` (`[][]string`)                                                     |
| Burstein        | `pairing/burstein`    | `Acceleration` (`*string`), `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                 |
| Dubov           | `pairing/dubov`       | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                                             |
| Lim             | `pairing/lim`         | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `MaxiTournament` (`*bool`)                                                                                                         |
| Dubbel-Zwitsers | `pairing/doubleswiss` | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`)                                                                                                             |
| Team-Zwitsers   | `pairing/team`        | `TopSeedColor` (`*string`), `ForbiddenPairs` (`[][]string`), `TotalRounds` (`*int`), `ColorPreferenceType` (`*string`: `"A"`, `"B"`, `"none"`), `PrimaryScore` (`*string`: `"match"`, `"game"`) |
| Keizer          | `pairing/keizer`      | `AllowRepeatPairings` (`*bool`, standaard `true`), `MinRoundsBetweenRepeats` (`*int`, standaard `3`), `ScoringOptions` (`*keizer.Options`)                                                      |
| Round-robin     | `pairing/roundrobin`  | `Cycles` (`*int`, standaard `1`), `ColorBalance` (`*bool`, standaard `true`), `SwapLastTwoRounds` (`*bool`, standaard `true`)                                                                   |

### Gemeenschappelijke opties

De meeste Zwitserse engines delen deze opties:

- **TopSeedColor** -- Forceert de kleur van de hoogst geplaatste speler in ronde 1. Waarden: `"auto"` (standaard, engine beslist), `"white"`, `"black"`.
- **ForbiddenPairs** -- Een lijst van speler-ID-paren `[id1, id2]` die niet tegen elkaar ingedeeld mogen worden. De engine behandelt deze als absolute beperkingen.

### Engine-specifieke opmerkingen

**Dutch (FIDE C.04.3):** Het meest gebruikte Zwitserse systeem. Gebruikt globale Blossom-matching met 21 kwaliteitscriteria. `Acceleration` schakelt Baku-acceleratie in (FIDE C.04.7), die in vroege ronden virtuele punten toekent om meer gevarieerde indelingen te produceren.

**Burstein (FIDE C.04.4.2):** Gebruikt seedingronden (gedelegeerd aan Dutch matching) gevolgd door oppositie-index-gebaseerde matching. Het aantal seedingronden is `min(floor(totalRounds/2), 4)`. Stel `TotalRounds` in om dit te regelen; bij nil wordt het afgeleid uit de state.

**Dubov (FIDE C.04.4.1):** Een ARO-egaliserende Zwitserse variant. Splitst scoregroepen op kleurvoorkeur, sorteert op oplopende ARO en gebruikt transpositie-gebaseerde matching met 10 criteria.

**Lim (FIDE C.04.4.3):** Verwerkt scoregroepen in mediaan-eerst-volgorde en gebruikt uitwisseling-gebaseerde matching. Heeft vier floater-typen (A-D). `MaxiTournament` schakelt de 100-punts ratingbeperking in voor uitwisselingen en floater-selectie.

**Dubbel-Zwitsers (FIDE C.04.5):** Elke ronde is een tweekamp. Gebruikt lexicografische bracket-indeling. `TotalRounds` wordt gebruikt om de laatste ronde te bepalen voor criteriaversoepeling.

**Team-Zwitsers (FIDE C.04.6):** Deelt teams in (elke `PlayerEntry` vertegenwoordigt een team). `ColorPreferenceType` selecteert Type A (eenvoudig) of Type B (sterk + mild) kleurvoorkeuren. `PrimaryScore` kiest tussen matchpunten en partijenreekspunten voor de indelingsrangschikking.

**Keizer:** Top-down indeling op basis van Keizer-score. `AllowRepeatPairings` regelt herindeling, met `MinRoundsBetweenRepeats` als tussenperiode. `ScoringOptions` configureert de interne Keizer-scorer die voor rangschikking wordt gebruikt; bij nil gelden de standaard-scoringswaarden. Kleurverdeling gebruikt de swisslib 6-stappencascade (dezelfde als Nederlands/Burstein).

**Round-robin (FIDE C.05 Annex 1):** Gebruikt FIDE Berger-tabellen. `Cycles` stelt het aantal volledige round-robins in (2 = dubbele round-robin met omgekeerde kleuren). `SwapLastTwoRounds` volgt de FIDE-aanbeveling om de laatste twee ronden van cyclus 1 om te wisselen bij een dubbele round-robin, om drie opeenvolgende partijen met dezelfde kleur op de cyclusgrens te voorkomen.

## Gebruiksvoorbeeld

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
)

func main() {
    state := &chesspairing.TournamentState{
        Players: []chesspairing.PlayerEntry{
            {ID: "1", DisplayName: "Alice",   Rating: 2400, Active: true},
            {ID: "2", DisplayName: "Bob",     Rating: 2350, Active: true},
            {ID: "3", DisplayName: "Charlie", Rating: 2300, Active: true},
            {ID: "4", DisplayName: "Diana",   Rating: 2250, Active: true},
            {ID: "5", DisplayName: "Eve",     Rating: 2200, Active: true},
        },
        CurrentRound: 1,
        PairingConfig: chesspairing.PairingConfig{
            System: chesspairing.PairingDutch,
        },
    }

    pairer := dutch.NewFromMap(nil)

    result, err := pairer.Pair(context.Background(), state)
    if err != nil {
        log.Fatal(err)
    }

    for _, g := range result.Pairings {
        fmt.Printf("Board %d: %s (W) vs %s (B)\n", g.Board, g.WhiteID, g.BlackID)
    }
    for _, b := range result.Byes {
        fmt.Printf("Bye: %s (%s)\n", b.PlayerID, b.Type)
    }
}
```

Bij een oneven aantal spelers ontvangt precies een speler een indelings-toegewezen bye (PAB). De bye-ontvanger wordt bepaald door het bye-selectiealgoritme van de engine.

## Foutafhandeling

Engines retourneren een fout bij ongeldige invoertoestanden. Veelvoorkomende foutcondities:

- Geen actieve spelers in het toernooi.
- `TournamentState.Validate()` faalt (lege ID's, duplicaten, ronde-telling komt niet overeen).
- Onvoldoende spelers om een indeling te vormen (bijv. een actieve speler zonder mogelijkheid tot bye).

Indelingsengines paniken nooit. Alle uitzonderlijke condities worden gemeld via de geretourneerde `error`-waarde.

## Gegevensstroom

```text
Aanroeper bouwt TournamentState
  -> Pairer.Pair(ctx, state) retourneert *PairingResult
     -> PairingResult.Pairings: bordtoewijzingen voor de ronde
     -> PairingResult.Byes: bye-toewijzingen (doorgaans 0 of 1)
     -> PairingResult.Notes: diagnostische berichten van de engine
```

De aanroeper is verantwoordelijk voor het vastleggen van het indelingsresultaat in `RoundData` voordat `Pair` opnieuw wordt aangeroepen voor de volgende ronde.
