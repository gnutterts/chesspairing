---
title: "Resultaattiebreakers"
linkTitle: "Resultaten"
weight: 3
description: "Wins, Rounds Won, Standard Points, Progressive Score en Koya System."
---

Resultaattiebreakers leiden hun waarden rechtstreeks af van partijuitslagen en ronde-voor-ronde-scores. In tegenstelling tot [Buchholz](../buchholz/)- of [prestatietiebreakers](../performance/) houden ze geen rekening met de sterkte of rating van de tegenstand. Alle vijf vallen onder **FIDE-categorie B** (gebaseerd op de eigen resultaten van de speler).

## Forfait-afhandeling

De gedeelde functie `buildOpponentData()` sluit alle forfaits uit van partij-items. Alleen resultaten aan het bord (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) leveren partij-items op. Sommige tiebreakers in deze groep werken rechtstreeks met rondegegevens in plaats van de tegenstanderdata-structuur; hun specifieke forfait-afhandeling staat hieronder per tiebreaker beschreven.

## Tiebreakers

### wins

**ID:** `wins`
**Naam:** Gewonnen partijen (aan het bord)
**FIDE-categorie:** B

Telt het aantal winstpartijen aan het bord. Alleen `resultWin`-items uit de partijenlijst van de speler (opgebouwd door `buildOpponentData()`) worden geteld. Omdat `buildOpponentData()` alle forfaits uitsluit, telt dit strikt alleen overwinningen aan het bord.

**Algoritme:**

1. Doorloop de partij-items van de speler uit `buildOpponentData()`.
2. Tel items waar `result == resultWin`.

Byes, forfaitwinsten en remises dragen niet bij.

**Formule:** `COUNT(games where result = win)`

### win

**ID:** `win`
**Naam:** Gewonnen ronden
**FIDE-categorie:** B

Telt het aantal ronden waarin de speler winstequivalente punten ontving. Dit is breder dan `wins` -- het omvat winstpartijen aan het bord, forfaitwinsten en volle-punt-byes (PAB).

**Algoritme:**

1. Doorloop alle partijen in alle ronden. Per partij:
   - `ResultWhiteWins` of `ResultForfeitWhiteWins`: verhoog de teller van wit.
   - `ResultBlackWins` of `ResultForfeitBlackWins`: verhoog de teller van zwart.
2. Doorloop alle byes. Per `ByePAB`: verhoog de teller van de speler.
3. Halve-punt-byes, nul-punten-byes, remises, hangende partijen en dubbele forfaits tellen niet mee.

**Formule:** `COUNT(OTB wins + forfeit wins + PAB byes)`

### standard-points

**ID:** `standard-points`
**Naam:** Standaardpunten
**FIDE-categorie:** B

Normaliseert het resultaat van elke ronde naar een standaard 1/0.5/0-schaal, ongeacht het puntensysteem van het toernooi. Dit is nuttig bij toernooien met niet-standaard puntwaarden (bijv. voetbalscoring 3-1-0).

**Algoritme:**

Bepaal per ronde de toegekende punten van de speler en vergelijk:

1. **Als er een tegenstander is:** vergelijk de punten van de speler met die van de tegenstander voor die partij.
   - Speler scoorde meer dan tegenstander: +1.0
   - Gelijk: +0.5
   - Speler scoorde minder: +0.0
2. **Als er geen tegenstander is** (bye of afwezig): vergelijk de toegekende punten van de speler met 0.5.
   - PAB (1.0 punten) > 0.5: +1.0
   - Halve-punt-bye (0.5 punten) = 0.5: +0.5
   - Nul-punt-bye of afwezig (0.0 punten) < 0.5: +0.0
3. Tel op over alle ronden.

**Formule:** `SUM(per-round standard result)`

### progressive

**ID:** `progressive`
**Naam:** Progressieve score
**FIDE-categorie:** B

De progressieve (cumulatieve) score beloont spelers die vroeg winnen. Het bouwt scores per ronde op, berekent cumulatieve totalen na elke ronde en telt vervolgens alle cumulatieve waarden op.

**Algoritme:**

1. Bouw scores per ronde op voor elke speler:
   - Winst of forfaitwinst: 1.0
   - Remise: 0.5
   - Verlies, forfaitverlies, dubbel forfait: 0.0
   - PAB: 1.0, halve-punt-bye: 0.5, nul-punt-bye/afwezig: 0.0
2. Bereken cumulatieve scores: na ronde 1, na ronde 2, enz.
3. Tel alle cumulatieve waarden op.

**Voorbeeld:** Een speler die 1, 0, 1, 1 scoort over vier ronden:

- Per ronde: [1.0, 0.0, 1.0, 1.0]
- Cumulatief: [1.0, 1.0, 2.0, 3.0]
- Progressief = 1.0 + 1.0 + 2.0 + 3.0 = **7.0**

Vergelijk met een speler die 0, 1, 1, 1 scoort (hetzelfde totaal van 3.0):

- Per ronde: [0.0, 1.0, 1.0, 1.0]
- Cumulatief: [0.0, 1.0, 2.0, 3.0]
- Progressief = 0.0 + 1.0 + 2.0 + 3.0 = **6.0**

De eerste speler staat hoger omdat die eerder won.

**Formule:** `SUM(cumulative scores after each round)`

### koya

**ID:** `koya`
**Naam:** Koya-systeem
**FIDE-categorie:** B

Het Koya-systeem telt de punten behaald tegen tegenstanders in de bovenste helft van de ranglijst. Het is vooral nuttig bij round-robin-toernooien.

**Algoritme:**

1. Bereken de kwalificatiedrempel: `totalRounds / 2`.
2. Identificeer "kwalificerende tegenstanders": spelers wier eindscore >= drempel is.
3. Voor elke partij aan het bord van de speler tegen een kwalificerende tegenstander:
   - Winst: +1.0
   - Remise: +0.5
   - Verlies: +0.0
4. Tel de bijdragen op.

Alleen partijen aan het bord tellen mee (via `buildOpponentData()`). Forfaits, byes en partijen tegen niet-kwalificerende tegenstanders worden uitgesloten.

**Formule:** `SUM(results against opponents with score >= totalRounds/2)`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Games Won (OTB only)
tb, err := tiebreaker.Get("wins")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Rounds Won (OTB wins + forfeit wins + PAB)
tb, err = tiebreaker.Get("win")
values, err = tb.Compute(ctx, state, scores)

// Standard Points
tb, err = tiebreaker.Get("standard-points")
values, err = tb.Compute(ctx, state, scores)

// Progressive Score
tb, err = tiebreaker.Get("progressive")
values, err = tb.Compute(ctx, state, scores)

// Koya System
tb, err = tiebreaker.Get("koya")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item.
