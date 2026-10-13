---
title: "Onderlinge tiebreakers"
linkTitle: "Onderling resultaat"
weight: 4
description: "Direct Encounter en Sonneborn-Berger — tiebreakers op basis van resultaten tegen specifieke tegenstanders."
---

Onderlinge tiebreakers beslechten gelijke standen door resultaten tussen specifieke tegenstanders te bekijken, in plaats van over het hele deelnemersveld te aggregeren. Direct Encounter kijkt alleen naar partijen tussen gelijk geëindigde spelers, terwijl Sonneborn-Berger elk resultaat weegt naar de eindscore van de tegenstander.

Onderling resultaat is **FIDE-type A** (partijen tussen gelijkstaande spelers); Sonneborn-Berger is **type BC** (eigen resultaten en die van tegenstanders).

## Forfait-afhandeling

Beide tiebreakers gebruiken tegenstanderrecords per ronde. Niet-gespeelde Zwitserse ronden volgen het model van artikel 16 met aangepaste scores en dummy's; bij vooraf vastgelegde indelingen behandelt artikel 15.2 forfaits als gewone ontmoetingen. Hangende partijen zijn geen voltooide ontmoetingen.

## Tiebreakers

### direct-encounter

**ID:** `direct-encounter`
**Naam:** Onderling resultaat
**FIDE-categorie:** A

Direct Encounter telt alleen ontmoetingen tussen spelers in de subgroep die na eerdere tiebreakers nog gelijk staat. Spelers die met niemand gelijk staan krijgen een waarde van 0.

**Algoritme:**

1. Begin met groepen op score en pas elke tiebreaker alleen toe op subgroepen die nog gelijk staan.
2. Tel voor Direct Encounter de toegekende punten tegen spelers in de subgroep. Bij herhaalde ontmoetingen met dezelfde tegenstander telt het gemiddelde resultaat als één bijdrage (een Double Swiss-match is al één ontmoeting).
3. Als Direct Encounter een groep splitst maar een subgroep gelijk blijft, pas het opnieuw toe op die subgroep totdat geen verdere splitsing optreedt. Een speler zonder gelijke stand krijgt 0 als hij niet is beoordeeld.

In een Zwitsers toernooi met ontbrekende onderlinge partijen krijgt alleen een speler die ongeacht de uitslagen daarvan alleen bovenaan de afzonderlijke stand blijft de eerste plaats; herhaal dit voor de overigen en laat onbesliste spelers gelijk staan.

**Bijzonderheden:**

- Partijen tegen spelers buiten de gelijke groep worden volledig genegeerd.
- Als twee gelijk staande spelers elkaar nooit hebben ontmoet, weerspiegelt hun waarde alleen ontmoetingen met andere leden van de subgroep.
- In Zwitserse toernooien met veel spelers op dezelfde score kan deze tiebreaker doorslaggevend zijn wanneer specifieke onderlinge partijen hebben plaatsgevonden.

**Formule:** `SOM(gemiddelde toegekende punten per tegenstander binnen de gelijke subgroep)`

### sonneborn-berger

**ID:** `sonneborn-berger`
**Naam:** Sonneborn-Berger
**FIDE-categorie:** BC

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
