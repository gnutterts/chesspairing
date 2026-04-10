---
title: "Buchholz-familie"
linkTitle: "Buchholz"
weight: 1
description: "Vijf Buchholz-varianten — Volledig, Cut-1, Cut-2, Mediaan en Mediaan-2 — gebaseerd op scores van tegenstanders."
---

De Buchholz-tiebreaker meet de sterkte van iemands tegenstand door de eindscores van alle tegenstanders op te tellen. Een hogere Buchholz-waarde geeft aan dat de speler sterkere tegenstand heeft gehad. Er zijn vijf varianten geregistreerd, die alleen verschillen in hoeveel extreme tegenstanderscores worden afgekapt voor het optellen.

Alle vijf varianten vallen onder **FIDE-categorie A** (gebaseerd op resultaten van tegenstanders).

## Gedeeld algoritme

Elke Buchholz-variant volgt dezelfde kernstappen:

1. **Verzamel tegenstanderscores.** Voor elke gespeelde partij (geen forfait, niet hangend), zoek de eindscore van de tegenstander op.
2. **Voeg virtuele tegenstanders toe.** Voor elke ronde waarin de speler bye had of afwezig was (geen echte tegenstander), voeg een virtuele tegenstanderscore toe gelijk aan de eigen eindscore van de speler.
3. **Sorteer oplopend.** De verzamelde scores worden van laag naar hoog gesorteerd.
4. **Kap af.** Afhankelijk van de variant worden scores van de onderkant, de bovenkant, of beide verwijderd.
5. **Tel op.** De resterende scores worden opgeteld tot de tiebreakwaarde.

### Forfait- en bye-afhandeling

De gedeelde functie `buildOpponentData()` sluit alle forfaits uit van de partijenlijst. Alleen resultaten aan het bord (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) leveren partij-items op met een echte tegenstander. Forfaitwinsten, forfaitverliezen, dubbele forfaits en hangende partijen worden volledig overgeslagen.

Voor ronden waarin een speler geen partij aan het bord speelde:

- **Byes** (PAB, halve punt, nul punten) verhogen de bye-teller van de speler.
- **Afwezigheden** (actieve speler die niet voorkomt in een partij of bye van een ronde) verhogen de afwezigheidsteller.

Elke bye en afwezigheid draagt een virtuele tegenstanderscore bij gelijk aan de eigen eindscore van de speler.

## Varianten

| ID                 | Naam              | Afkapregel                              |
| ------------------ | ----------------- | --------------------------------------- |
| `buchholz`         | Buchholz          | Geen -- som van alle tegenstanderscores |
| `buchholz-cut1`    | Buchholz Cut-1    | Laagste tegenstanderscore weglaten      |
| `buchholz-cut2`    | Buchholz Cut-2    | 2 laagste tegenstanderscores weglaten   |
| `buchholz-median`  | Buchholz Median   | Hoogste EN laagste weglaten             |
| `buchholz-median2` | Buchholz Median-2 | 2 hoogste EN 2 laagste weglaten         |

### buchholz

**ID:** `buchholz`
**Naam:** Buchholz
**FIDE-categorie:** A

De volledige Buchholz. Geen afkapping -- elke tegenstanderscore (echt of virtueel) wordt opgeteld.

**Formule:** `SUM(all opponent scores)`

### buchholz-cut1

**ID:** `buchholz-cut1`
**Naam:** Buchholz Cut-1
**FIDE-categorie:** A

Na oplopend sorteren van tegenstanderscores wordt de laagste score weggelaten voor het optellen.

**Formule:** `SUM(opponent scores[1..n])` (index 0 overgeslagen)

### buchholz-cut2

**ID:** `buchholz-cut2`
**Naam:** Buchholz Cut-2
**FIDE-categorie:** A

Na sorteren worden de twee laagste tegenstanderscores weggelaten. Als er minder dan drie tegenstanders zijn, blijft er niets over om op te tellen (waarde is 0).

**Formule:** `SUM(opponent scores[2..n])` (indices 0 en 1 overgeslagen)

### buchholz-median

**ID:** `buchholz-median`
**Naam:** Buchholz Median
**FIDE-categorie:** A

Zowel de hoogste als de laagste tegenstanderscore wordt weggelaten. Dit verwijdert de beste en de slechtste tegenstander, waardoor de invloed van extreme indelingen vermindert.

**Formule:** `SUM(opponent scores[1..n-1])` (eerste en laatste overgeslagen)

### buchholz-median2

**ID:** `buchholz-median2`
**Naam:** Buchholz Median-2
**FIDE-categorie:** A

De twee hoogste en twee laagste tegenstanderscores worden weggelaten. Er zijn minstens vijf tegenstanders nodig om nog resterende scores te hebben.

**Formule:** `SUM(opponent scores[2..n-2])` (eerste twee en laatste twee overgeslagen)

## Voorbeeld

Een speler met tegenstanders die [2.0, 3.0, 3.5, 4.0, 5.0] scoorden:

| Variant          | Weggelaten         | Resterend                 | Waarde |
| ---------------- | ------------------ | ------------------------- | ------ |
| buchholz         | geen               | [2.0, 3.0, 3.5, 4.0, 5.0] | 17.5   |
| buchholz-cut1    | 2.0                | [3.0, 3.5, 4.0, 5.0]      | 15.5   |
| buchholz-cut2    | 2.0, 3.0           | [3.5, 4.0, 5.0]           | 12.5   |
| buchholz-median  | 2.0, 5.0           | [3.0, 3.5, 4.0]           | 10.5   |
| buchholz-median2 | 2.0, 3.0, 4.0, 5.0 | [3.5]                     | 3.5    |

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Full Buchholz
tb, err := tiebreaker.Get("buchholz")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Buchholz Cut-1
tbCut1, err := tiebreaker.Get("buchholz-cut1")
values, err = tbCut1.Compute(ctx, state, scores)

// Buchholz Cut-2
tbCut2, err := tiebreaker.Get("buchholz-cut2")
values, err = tbCut2.Compute(ctx, state, scores)

// Buchholz Median
tbMedian, err := tiebreaker.Get("buchholz-median")
values, err = tbMedian.Compute(ctx, state, scores)

// Buchholz Median-2
tbMedian2, err := tiebreaker.Get("buchholz-median2")
values, err = tbMedian2.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item, waarbij `Value` de berekende Buchholz-score is.
