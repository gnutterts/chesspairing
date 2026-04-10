---
title: "Onderlinge tiebreakers"
linkTitle: "Onderling resultaat"
weight: 4
description: "Direct Encounter en Sonneborn-Berger — tiebreakers op basis van resultaten tegen specifieke tegenstanders."
---

Onderlinge tiebreakers beslechten gelijke standen door resultaten tussen specifieke tegenstanders te bekijken, in plaats van over het hele deelnemersveld te aggregeren. Direct Encounter kijkt alleen naar partijen tussen gelijk geëindigde spelers, terwijl Sonneborn-Berger elk resultaat weegt naar de eindscore van de tegenstander.

Beide tiebreakers vallen onder **FIDE-categorie A** (gebaseerd op resultaten van tegenstanders).

## Forfait-afhandeling

Beide tiebreakers gebruiken `buildOpponentData()`, die alle forfaits uitsluit van partij-items. Alleen resultaten aan het bord (`ResultWhiteWins`, `ResultBlackWins`, `ResultDraw`) leveren partij-items op. Forfaitwinsten, forfaitverliezen, dubbele forfaits en hangende partijen genereren geen partij-items en dragen niet bij aan deze tiebreakers.

## Tiebreakers

### direct-encounter

**ID:** `direct-encounter`
**Naam:** Onderling resultaat
**FIDE-categorie:** A

De direct-encounter-tiebreaker beschouwt alleen partijen gespeeld tussen leden van dezelfde gelijke groep. Spelers die met niemand gelijk staan krijgen een waarde van 0.

**Algoritme:**

1. Groepeer alle spelers op hun primaire score. Elke groep spelers met dezelfde score vormt een "gelijke groep".
2. Voor groepen met slechts een speler (geen gelijke stand) is de waarde 0.
3. Voor elke speler in een gelijke groep, doorloop hun partij-items aan het bord:
   - Als de tegenstander ook in dezelfde gelijke groep zit:
     - Winst: +1.0
     - Remise: +0.5
     - Verlies: +0.0
   - Als de tegenstander NIET in de gelijke groep zit: overslaan.
4. Tel de bijdragen op.

**Bijzonderheden:**

- Partijen tegen spelers buiten de gelijke groep worden volledig genegeerd.
- Als twee gelijk staande spelers nooit tegen elkaar aan het bord hebben gespeeld, weerspiegelt hun direct-encounter-waarde alleen partijen tegen andere leden van de gelijke groep.
- In Zwitserse toernooien met veel spelers op dezelfde score kan deze tiebreaker doorslaggevend zijn wanneer specifieke onderlinge partijen hebben plaatsgevonden.

**Formule:** `SUM(standard results from OTB games against tied-group opponents)`

### sonneborn-berger

**ID:** `sonneborn-berger`
**Naam:** Sonneborn-Berger
**FIDE-categorie:** A

Sonneborn-Berger (SB) weegt elk partijresultaat naar de eindscore van de tegenstander. Winsten tegen sterke tegenstanders dragen meer bij dan winsten tegen zwakke tegenstanders. Dit is een van de meest gebruikte tiebreakers bij round-robin-toernooien.

**Algoritme:**

Voor elke partij aan het bord in de partijenlijst van de speler:

1. Zoek de eindscore van de tegenstander op uit de scorelijst.
2. Pas het resultaatgewicht toe:
   - Winst: tel de volledige score van de tegenstander op.
   - Remise: tel de helft van de score van de tegenstander op.
   - Verlies: tel 0 op.
3. Tel alle bijdragen op.

**Voorbeeld:** Een speler die een tegenstander met 5.0 punten versloeg en remise speelde tegen een tegenstander met 4.0 punten:

- Winstbijdrage: 5.0
- Remisebijdrage: 4.0 / 2 = 2.0
- SB = 5.0 + 2.0 = **7.0**

**Formule:** `SUM(win: opponent score, draw: opponent score / 2, loss: 0)`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Direct Encounter
tb, err := tiebreaker.Get("direct-encounter")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Sonneborn-Berger
tb, err = tiebreaker.Get("sonneborn-berger")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item.
