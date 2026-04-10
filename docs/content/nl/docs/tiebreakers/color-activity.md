---
title: "Kleur- en activiteitstiebreakers"
linkTitle: "Kleur & Activiteit"
weight: 5
description: "Black Games, Black Wins, Rounds Played en Games Played — tiebreakers op basis van activiteit."
---

Kleur- en activiteitstiebreakers meten deelname en kleurverdeling in plaats van de kwaliteit van de tegenstand. Ze tellen partijen gespeeld met zwart, winstpartijen met zwart, het totaal aantal effectief gespeelde ronden en het totaal aantal niet-forfaitpartijen. Deze tiebreakers belonen actieve deelname en helpen spelers te onderscheiden die vaker het nadeel van zwart moesten overwinnen.

Drie van de vier vallen onder **FIDE-categorie B**. Games Played heeft geen FIDE-categorietoewijzing en wordt vooral gebruikt bij Keizer-toernooien.

## Tiebreakers

### black-games

**ID:** `black-games`
**Naam:** Partijen met zwart
**FIDE-categorie:** B

Telt het aantal partijen gespeeld als zwart waarbij het resultaat geen forfait is. Een hogere waarde geeft aan dat de speler vaker het nadeel van de eerste zet heeft overwonnen.

**Algoritme:**

1. Doorloop alle partijen in alle ronden.
2. Voor elke partij waar `result.IsForfeit()` onwaar is (d.w.z. een partij aan het bord -- `ResultWhiteWins`, `ResultBlackWins`, `ResultDraw` of `ResultPending`):
   - Verhoog de teller van de zwartspeler.
3. Forfaitwinsten, forfaitverliezen en dubbele forfaits worden uitgesloten.

Deze tiebreaker werkt rechtstreeks op rondegegevens, niet via `buildOpponentData()`. Hij gebruikt de methode `IsForfeit()` op `GameResult`, die waar retourneert voor `ResultForfeitWhiteWins`, `ResultForfeitBlackWins` en `ResultDoubleForfeit`.

**Formule:** `COUNT(games as Black where IsForfeit() = false)`

### black-wins

**ID:** `black-wins`
**Naam:** Winstpartijen met zwart
**FIDE-categorie:** B

Telt winstpartijen aan het bord met de zwarte stukken. Alleen partijen met resultaat `ResultBlackWins` worden geteld -- forfaitwinsten met zwart zijn uitgesloten.

**Algoritme:**

1. Doorloop alle partijen in alle ronden.
2. Voor elke partij waar het resultaat exact `ResultBlackWins` is:
   - Verhoog de teller van de zwartspeler.
3. `ResultForfeitBlackWins`, remises en alle andere resultaten worden uitgesloten.

**Formule:** `COUNT(games where result = ResultBlackWins AND player is Black)`

### rounds-played

**ID:** `rounds-played`
**Naam:** Gespeelde ronden
**FIDE-categorie:** B

Berekent het aantal effectief gespeelde ronden door ongespeelde ronden af te trekken van het totale aantal ronden.

**Algoritme:**

Begin met `totalRounds`. Bepaal per ronde welke ronden als "ongespeeld" tellen:

**Ongespeeld (afgetrokken van totaal):**

- Forfaitverlies (de verliezende kant van `ResultForfeitWhiteWins` of `ResultForfeitBlackWins`)
- Dubbel forfait (beide spelers bij een `ResultDoubleForfeit`-partij)
- Halve-punt-bye (`ByeHalf`)
- Nul-punten-bye (`ByeZero`)
- Afwezigheidsbye (`ByeAbsent`)
- Niet aanwezig in de ronde (actieve speler niet in een partij of bye)

**Gespeeld (niet afgetrokken):**

- Partijen aan het bord (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`)
- Forfaitwinst (de winnende kant)
- PAB (`ByePAB`)

**Formule:** `totalRounds - COUNT(unplayed rounds)`

### games-played

**ID:** `games-played`
**Naam:** Gespeelde partijen
**FIDE-categorie:** --

Telt het aantal niet-forfaitpartijen. Beide spelers in elke kwalificerende partij worden geteld. Deze tiebreaker is vooral nuttig bij Keizer-toernooien, waar spelers die meer clubavonden bijwoonden en meer partijen speelden hoger moeten eindigen dan spelers met dezelfde score maar minder daadwerkelijke partijen.

**Algoritme:**

1. Doorloop alle partijen in alle ronden.
2. Voor elke partij waar `IsForfeit` onwaar is op de `GameData`-struct:
   - Verhoog de tellers van zowel wit als zwart.
3. Byes, afwezigheden en forfaitpartijen worden uitgesloten.

Let op: deze tiebreaker gebruikt het veld `IsForfeit` op de `GameData`-struct rechtstreeks, in plaats van het resultaattype te controleren.

**Formule:** `COUNT(non-forfeit games the player participated in)`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Games with Black
tb, err := tiebreaker.Get("black-games")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Black Wins
tb, err = tiebreaker.Get("black-wins")
values, err = tb.Compute(ctx, state, scores)

// Rounds Played
tb, err = tiebreaker.Get("rounds-played")
values, err = tb.Compute(ctx, state, scores)

// Games Played
tb, err = tiebreaker.Get("games-played")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item.
