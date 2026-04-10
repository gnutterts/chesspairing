---
title: "Prestatietiebreakers"
linkTitle: "Prestatie"
weight: 2
description: "TPR, PTP, APRO, APPO en ARO — tiebreakers afgeleid van ratings en verwachte scores."
---

Prestatietiebreakers gebruiken spelerratings en de FIDE B.02-conversietabel om tiebreakwaarden af te leiden. Ze meten hoe goed een speler heeft gepresteerd ten opzichte van de ratingsterkte van de tegenstand.

Alle vijf tiebreakers in deze groep vallen onder **FIDE-categorie D** (gebaseerd op ratings).

## FIDE B.02-conversietabel

Twee opzoekfuncties vormen de basis van de ratinggebaseerde tiebreakers:

- **`dpFromP(p)`** -- Gegeven een fractionele score `p` (0.0 tot 1.0), retourneert het ratingverschil `dp`. De tabel bevat 101 items van p=0.00 (dp=-800) via p=0.50 (dp=0) tot p=1.00 (dp=+800). Waarden tussen tabelitems worden lineair geinterpoleerd.
- **`expectedScore(dp)`** -- De omgekeerde opzoeking. Gegeven een ratingverschil `dp` (-800 tot +800), retourneert de verwachte fractionele score. Eveneens lineair geinterpoleerd tussen tabelitems.

## Forfait-afhandeling

Alle prestatietiebreakers gebruiken `buildOpponentData()`, die alle forfaits uitsluit van partij-items. Alleen resultaten aan het bord (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) leveren partij-items op. Forfaitwinsten, forfaitverliezen, dubbele forfaits en hangende partijen worden overgeslagen. Spelers zonder partijen aan het bord krijgen een waarde van 0.

## Tiebreakers

### aro

**ID:** `aro`
**Naam:** Gem. rating van tegenstanders
**FIDE-categorie:** D

Het rekenkundig gemiddelde van de ratings van alle tegenstanders aan het bord.

**Algoritme:**

1. Voor elke partij aan het bord in de partijenlijst van de speler, zoek de rating van de tegenstander op.
2. Tel alle tegenstanderratings op.
3. Deel door het aantal partijen aan het bord.

Als de speler geen partijen aan het bord heeft, is de waarde 0.

**Formule:** `SUM(opponent ratings) / number of OTB games`

### performance-rating

**ID:** `performance-rating`
**Naam:** Prestatierating (TPR)
**FIDE-categorie:** D

De Tournament Performance Rating combineert de gemiddelde tegenstanderrating met een ratingverschil-correctie afgeleid van de fractionele score van de speler.

**Algoritme:**

1. Bereken ARO (gemiddelde rating van tegenstanders uit partijen aan het bord).
2. Bereken de fractionele score: `p = player score / number of OTB games`, begrensd op [0.0, 1.0].
3. Zoek `dp = dpFromP(p)` op uit de FIDE B.02-tabel (met lineaire interpolatie).
4. `TPR = round(ARO + dp)`.

Als de speler geen partijen aan het bord heeft, is de waarde 0. Het resultaat wordt afgerond op het dichtstbijzijnde gehele getal (0.5 wordt naar boven afgerond).

**Formule:** `round(ARO + dpFromP(score / games))`

### performance-points

**ID:** `performance-points`
**Naam:** Prestatiepunten (PTP)
**FIDE-categorie:** D

PTP zoekt de laagste hypothetische rating R waarvoor de som van verwachte scores tegen alle tegenstanders de werkelijke score van de speler zou bereiken of overschrijden. Hiervoor wordt binair gezocht over de FIDE verwachte-score-functie.

**Algoritme:**

1. Verzamel alle ratings van tegenstanders aan het bord.
2. **Nulscore:** waarde = `round(lowest opponent rating - 800)`.
3. **Perfecte score:** waarde = `round(highest opponent rating + 800)`.
4. **Anders:** binair zoeken naar de laagste R in het bereik `[min opponent rating - 800, max opponent rating + 800]` waarvoor `SUM(expectedScore(R - oppRating_i))` >= werkelijke score. Zoekprecisie is 0.5 ratingpunten.
5. Het resultaat wordt afgerond op het dichtstbijzijnde gehele getal.

Als de speler geen partijen aan het bord heeft, is de waarde 0.

**Formule:** `min R such that SUM(expectedScore(R - oppRating_i)) >= score`

### avg-opponent-tpr

**ID:** `avg-opponent-tpr`
**Naam:** Gem. tegenstander-TPR (APRO)
**FIDE-categorie:** D

Het gemiddelde van de Tournament Performance Ratings van alle tegenstanders aan het bord.

**Algoritme:**

1. Bereken TPR voor elke speler in het toernooi (via de `performance-rating`-tiebreaker).
2. Verzamel voor elk van de tegenstanders aan het bord van de speler hun TPR-waarde.
3. Bereken het gemiddelde.
4. Rond af op het dichtstbijzijnde gehele getal.

Als de speler geen partijen aan het bord heeft, is de waarde 0.

**Formule:** `round(SUM(opponent TPR values) / number of OTB games)`

### avg-opponent-ptp

**ID:** `avg-opponent-ptp`
**Naam:** Gem. tegenstander-PTP (APPO)
**FIDE-categorie:** D

Het gemiddelde van de Performance Points-waarden van alle tegenstanders aan het bord.

**Algoritme:**

1. Bereken PTP voor elke speler in het toernooi (via de `performance-points`-tiebreaker).
2. Verzamel voor elk van de tegenstanders aan het bord van de speler hun PTP-waarde.
3. Bereken het gemiddelde.
4. Rond af op het dichtstbijzijnde gehele getal.

Als de speler geen partijen aan het bord heeft, is de waarde 0.

**Formule:** `round(SUM(opponent PTP values) / number of OTB games)`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Average Rating of Opponents
tb, err := tiebreaker.Get("aro")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Performance Rating (TPR)
tb, err = tiebreaker.Get("performance-rating")
values, err = tb.Compute(ctx, state, scores)

// Performance Points (PTP)
tb, err = tiebreaker.Get("performance-points")
values, err = tb.Compute(ctx, state, scores)

// Average Opponent TPR (APRO)
tb, err = tiebreaker.Get("avg-opponent-tpr")
values, err = tb.Compute(ctx, state, scores)

// Average Opponent PTP (APPO)
tb, err = tiebreaker.Get("avg-opponent-ptp")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item. Voor `aro` is de waarde een float64 (niet afgerond). Voor alle andere wordt de waarde afgerond op het dichtstbijzijnde gehele getal.
