---
title: "Volgordetiebreakers"
linkTitle: "Volgorde"
weight: 6
description: "Rangnummer en Spelerrating — deterministische volgordetiebreakers."
---

Volgordetiebreakers bieden een deterministische oplossing voor gelijke standen wanneer alle andere tiebreakers gelijke waarden opleveren. Ze worden doorgaans als laatste in de tiebreaker-keten geplaatst om een volledige ordening van de ranglijst te garanderen.

## Tiebreakers

### pairing-number

**ID:** `pairing-number`
**Naam:** Rangnummer (TPN)
**FIDE-categorie:** B

Gebruikt het rangnummer (de 1-gebaseerde index van de speler in de `state.Players`-slice) als tiebreakwaarde. De waarde wordt **genegeerd** zodat een lager rangnummer (hogere plaatsing) een hogere tiebreakwaarde oplevert, in overeenstemming met de conventie dat hogere tiebreakwaarden beter rangschikken.

**Algoritme:**

1. Ken elke speler hun 1-gebaseerde index toe in `state.Players`.
2. Negeer de index: `value = -float64(index)`.

Speler 1 krijgt waarde -1, speler 2 krijgt waarde -2, enzovoort. Omdat -1 > -2, staat speler 1 boven speler 2.

**Formule:** `-(1-based player index)`

**Bijzondere afhandeling:** Deze tiebreaker gebruikt `buildOpponentData()` niet en bekijkt geen partijresultaten. Het is puur positioneel.

### player-rating

**ID:** `player-rating`
**Naam:** Spelerrating (RTNG)
**FIDE-categorie:** D

Gebruikt de geregistreerde rating van de speler als tiebreakwaarde. Een hogere rating rangschikt hoger.

**Algoritme:**

1. Zoek het `Rating`-veld van de speler op uit `state.Players`.
2. Geef het terug als float64.

**Formule:** `float64(player.Rating)`

**Bijzondere afhandeling:** Deze tiebreaker gebruikt `buildOpponentData()` niet en bekijkt geen partijresultaten. Hij gebruikt alleen de statische rating die bij aanvang van het toernooi is geregistreerd.

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Pairing Number (lower TPN = higher value)
tb, err := tiebreaker.Get("pairing-number")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Player Rating
tb, err = tiebreaker.Get("player-rating")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item.
