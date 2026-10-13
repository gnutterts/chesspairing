---
title: "Resultaattiebreakers"
linkTitle: "Resultaten"
weight: 3
description: "Wins, Rounds Won, Standard Points, Progressive Score en Koya System."
---

Resultaattiebreakers leiden hun waarden rechtstreeks af van partijuitslagen en ronde-voor-ronde-scores. In tegenstelling tot [Buchholz](../buchholz/)- of [prestatietiebreakers](../performance/) houden ze geen rekening met de sterkte of rating van de tegenstand. Alle vijf vallen onder **FIDE-categorie B** (gebaseerd op de eigen resultaten van de speler).

## Niet-gespeelde ronden

Resultaattiebreakers bouwen één record op per speler en ronde. In Zwitserse toernooien deelt FIDE C.07:2026 artikel 16 niet-gespeelde ronden in: forfaitwinsten en -verliezen, aangevraagde byes en afwezigheden zijn ongespeeld, terwijl volle-punt-byes als gespeeld tellen. In round-robins en andere vooraf vastgelegde indelingen behandelt artikel 15.2 forfaits juist als gewone ontmoetingen, behalve bij ratingtiebreakers en Type-B-forfaitverliezen. De per tiebreaker specifieke afhandeling staat hieronder beschreven.

## Double-Swiss-ronden

In een Double Swiss (FIDE C.04.5) bevat één ronde beide partijen van een match tussen dezelfde tegenstanders, met gewisselde kleuren. Het recordmodel voegt de twee partijen samen tot één record per ronde: hun punten worden opgeteld en de ronde telt als gespeeld tenzij beide partijen forfaits waren. De preface van C.04.5 beschouwt een match met één forfaitpartij als regulier gespeeld en behandelt die partij als gespeeld, terwijl een match slechts door forfait eindigt als één speler beide partijen verliest door forfait.

Tiebreakers die C.07:2026 over ronden en tegenstanders telt, werken daarom op het totaal van de match: standard points (artikel 7.7, "het aantal ronden waarin een deelnemer meer punten schaalt dan zijn tegenstander"), Sonneborn-Berger (artikel 9.1, "per ronde ... de tegen hen geschaalde punten"), Buchholz en zijn varianten (artikel 8), het onderling resultaat en ronden one elected to play (artikel 7.6). De ratinggebaseerde tiebreakers van artikel 10 tellen in plaats daarvan partijen: de gebroken score van de TPR deelt "het aantal punten behaald in partijen aan het bord" door "het aantal partijen" (artikel 10.2), en de verwachte score van de PTP wordt per partij opgeteld over "elke tegenstander tegen wie de speler in het toernooi speelde" (artikel 10.3); de gemiddelde tegenstanderrating van ARO is hetzelfde over partijen en ronden, omdat beide partijen van een match dezelfde tegenstander hebben. Voor forfaitpartijen staat artikel 10 apart: de preface-regel die één forfaitpartij als gespeeld behandelt "geldt niet voor rating", dus forfaits blijven niet-gespeeld voor de ratinggebaseerde tiebreakers. Tiebreakers die partijen tellen, blijven elke partij afzonderlijk tellen: gewonnen partijen (artikel 7.2), partijen met zwart (artikel 7.3) en gewonnen partijen met zwart (artikel 7.4). Het Koya-systeem (artikel 9.2) somt de punten op die tegen een gekwalificeerde tegenstander zijn behaald, dus beide partijen dragen bij.

## Tiebreakers

### wins

**ID:** `wins`
**Naam:** Gewonnen partijen
**FIDE-categorie:** B

Telt het aantal gewonnen partijen. In Zwitserse toernooien tellen alleen winstpartijen aan het bord. In round-robins en andere vooraf vastgelegde indelingen behandelt FIDE C.07:2026 artikel 15.2 een forfaitwinst als een gewone partij, dus die telt daar ook mee.

**Algoritme:**

1. Doorloop alle partijen in alle ronden.
2. Tel een winst aan het bord (1-0 of 0-1); in evenementen met vooraf vastgelegde indelingen telt een forfaitwinst ook (FIDE C.07:2026 artikel 15.2).

In een Double Swiss tellen beide partijen van een match apart (artikel 7.2 telt partijen). Byes, remises en forfaitverliezen dragen niet bij.

**Formule:** `COUNT(games with a win result)`

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

1. **Als er een tegenstander is:** vergelijk de ingestelde punten van beide spelers voor de ronde.
   - Speler scoorde meer dan tegenstander: +1.0
   - Gelijk: +0.5
   - Speler scoorde minder: +0.0
   - Dubbel forfait: +0.5 (beide geplande tegenstanders scoren evenveel; C.07:2026 artikel 7.7).
2. **Als er geen tegenstander is** (bye of afwezig): vergelijk de punten met de ingestelde remisewaarde (standaard 0.5).
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

Alleen partijen aan het bord tellen mee in Zwitserse toernooien. In vooraf vastgelegde indelingen behandelt FIDE C.07:2026 artikel 15.2 alle forfaits als gewone ontmoetingen, dus forfaits tellen daar ook mee. Byes en partijen tegen niet-kwalificerende tegenstanders worden uitgesloten.

**Formule:** `SUM(results against opponents with score >= totalRounds/2)`

## Gebruik

```go
import (
    "context"

    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/tiebreaker"
)

// Gewonnen partijen
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
