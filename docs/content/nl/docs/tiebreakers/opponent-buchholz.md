---
title: "Fore Buchholz en Gemiddelde Tegenstander-Buchholz"
linkTitle: "Tegenstander-Buchholz"
weight: 7
description: "Buchholz-varianten voor onvolledige ronden of gemiddelde tegenstander-bijdragen."
---

Deze twee tiebreakers breiden het [Buchholz](../buchholz/)-concept uit voor specifieke situaties: Fore Buchholz is bedoeld voor toernooien met een onvolledige laatste ronde, en Average Opponent Buchholz normaliseert Buchholz-waarden over tegenstanders.

Beide vallen onder **FIDE-categorie C** (gebaseerd op resultaten van tegenstanders, geavanceerde varianten).

## Forfait-afhandeling

Beide tiebreakers gebruiken dezelfde records per ronde als de [Buchholz-familie](../buchholz/). Niet-gespeelde Zwitserse ronden volgen het model van artikel 16 met aangepaste scores en begrensde dummy's; artikel 15.2 gebruikt forfaits als gewone ontmoetingen bij vooraf vastgelegde indelingen. Hangende partijen zijn geen voltooide ontmoetingen.

## Tiebreakers

### fore-buchholz

**ID:** `fore-buchholz`
**Naam:** Fore Buchholz
**FIDE-categorie:** C

Fore Buchholz berekent de volledige Buchholz-score alsof alle hangende partijen in de laatste ronde in remise eindigen. Hierdoor kan de stand berekend worden voordat de laatste ronde is afgelopen. Als alle partijen in de laatste ronde al zijn afgelopen, is Fore Buchholz gelijk aan de gewone Buchholz.

**Algoritme:**

1. Begin met de werkelijke scores van alle spelers.
2. Identificeer hangende partijen in de laatste ronde (`ResultPending` in `state.Rounds[last]`).
3. Voor elke hangende partij, tel +0.5 op bij de virtuele score van zowel de wit- als de zwartspeler.
4. Bouw de gewone records per ronde op en behandel hangende partijen uit de laatste ronde voor deze berekening als remises.
5. Gebruik de virtuele scores uit stap 3 en pas bij Buchholz de regels van artikel 16 voor aangepaste scores en begrensde dummy's toe.

**Formule:** `Buchholz(modified state where pending last-round games = draws)`

**Bijzondere afhandeling:**

- Alleen hangende partijen in de laatste ronde worden als remise behandeld. Hangende partijen in eerdere ronden (indien aanwezig) blijven uitgesloten.
- De virtuele score-aanpassing (+0.5 per hangende partij) werkt door in de tegenstanderscore-opzoekingen en beïnvloedt alle spelers wier tegenstanders hangende partijen hebben.
- Als er geen ronden zijn, zijn alle waarden 0.

### avg-opponent-buchholz

**ID:** `avg-opponent-buchholz`
**Naam:** Gem. tegenstander-Buchholz (AOB)
**FIDE-categorie:** C

Average Opponent Buchholz berekent eerst de volledige Buchholz voor elke speler, en middelt vervolgens per speler de Buchholz-waarden van hun tegenstanders aan het bord. Dit normaliseert de tiebreaker over spelers die mogelijk een verschillend aantal partijen hebben gespeeld.

**Algoritme:**

1. Bereken de volledige Buchholz voor elke speler met een score uit de records per ronde, met toepassing van artikelen 15 en 16.
2. Voor elke speler, doorloop hun partij-items aan het bord:
   - Tel de Buchholz-waarden van elke tegenstander op.
   - Deel door het aantal partijen aan het bord.

Als de speler geen partijen aan het bord heeft, is de waarde 0.

**Formule:** `SUM(opponent Buchholz values) / number of OTB games`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Fore Buchholz (handles pending final-round games)
tb, err := tiebreaker.Get("fore-buchholz")
if err != nil {
    // handle error
}
values, err := tb.Compute(ctx, state, scores)

// Average Opponent Buchholz
tb, err = tiebreaker.Get("avg-opponent-buchholz")
values, err = tb.Compute(ctx, state, scores)
```

Elke aanroep geeft een `[]chesspairing.TieBreakValue` terug met per speler een item.
