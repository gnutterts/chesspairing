---
title: "Completeerbaarheid"
linkTitle: "Completeerbaarheid"
weight: 3
description: "Stage 0.5 pre-matching — bepalen welke speler de bye krijgt bij oneven aantallen."
---

## Het bye-probleem

Wanneer een ronde een oneven aantal actieve spelers heeft, moet precies een
speler een indeling-toegewezen bye (PAB) ontvangen. De vraag is: _welke?_

Een naieve aanpak — de bye toekennen aan de laagst gerangschikte speler die
in aanmerking komt — kan leiden tot situaties waarin de overige spelers niet
allemaal ingedeeld kunnen worden. Als het verwijderen van de laagst gerangschikte
speler bijvoorbeeld twee spelers overlaat die al tegen elkaar gespeeld hebben
en geen andere compatibele tegenstanders hebben, mislukt de indeling.

Stage 0.5 lost dit op door een vereenvoudigde [Blossom matching](../blossom/)
uit te voeren over alle bye-kandidaten _voordat_ de echte indeling begint. De
kandidaat wiens verwijdering de best completeerbare matching oplevert, wordt
als bye-ontvanger gekozen.

De implementatie staat in `pairing/swisslib/global_matching.go`, in de
functie `PairBracketsGlobal`.

---

## Wanneer Stage 0.5 draait

Stage 0.5 wordt alleen geactiveerd wanneer aan alle drie de voorwaarden
wordt voldaan:

1. Het aantal actieve spelers is **oneven**.
2. Ten minste een speler **komt in aanmerking** voor een bye (heeft niet al
   een PAB ontvangen in een eerdere ronde, of aan andere systeemspecifieke
   beperkingen is voldaan).
3. Het indelingssysteem gebruikt de globale Blossom-architectuur (Nederlands,
   Burstein).

Bij een even aantal spelers wordt Stage 0.5 volledig overgeslagen en gaat
het algoritme direct door naar de groepsloop.

---

## Vereenvoudigde kantgewichten

De echte Blossom matching (Stages 1--2) gebruikt een complexe
[kantgewicht-codering](../edge-weights/) met 20 velden die criteria
$C_5$--$C_{21}$ coderen. Stage 0.5 gebruikt een veel eenvoudiger
3-veldsgewicht dat een enkele vraag stelt: _als deze twee spelers worden
ingedeeld, welke bye-kandidaat blijft dan ongematcht?_

De drie velden, van meest-significant naar minst-significant:

| Veld                 | Breedte  | Doel                                                                                                                                                                    |
| -------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Topscorerbescherming | 1 bit    | Geeft voorkeur aan een lager scorende speler als bye-ontvanger boven een hoger scorende. Gezet op 1 wanneer geen van beide eindpunten de topscorende bye-kandidaat is.  |
| Scoresom             | $s$ bits | Som van de scores van de twee spelers. Maximalisatie hiervan duwt hoger scorende spelers naar indelingen, waardoor lager scorende spelers overblijven voor de bye.        |
| Bye-geschiktheid     | 1 bit    | Gezet op 1 wanneer geen van beide spelers een bye-kandidaat is. Geeft voorkeur aan het matchen van niet-bye-geschikte spelers samen, zodat bye-kandidaten vrij blijven. |

Hier is $s$ het aantal bits dat nodig is om de maximaal mogelijke scoresom
weer te geven (tweemaal de maximale individuele score).

De totale gewichtsbreedte is $s + 2$ bits — klein genoeg voor standaard
`int64`-rekenkunde. Geen `*big.Int` nodig voor Stage 0.5.

---

## Algoritme

De Stage 0.5-procedure:

1. **Bouw de kantenset.** Maak voor elk paar spelers $(i, j)$ dat voldoet
   aan de absolute criteria (C1: geen herparingen, C3: geen absolute
   kleurconflicten, plus verboden paren) een kant met het vereenvoudigde
   3-veldsgewicht.

2. **Voer Blossom uit met maximale cardinaliteit.** Roep `MaxWeightMatching`
   aan met `maxCardinality = true`. Dit vindt een matching die zoveel
   mogelijk spelers paart, met totaal gewicht als tiebreaker.

3. **Identificeer de ongematchte speler.** In een maximale-cardinaliteit-
   matching van een oneven aantal spelers blijft exact een speler ongematcht.
   Deze speler wordt de bye-ontvanger.

4. **Sla het resultaat op.** De identiteit van de bye-ontvanger wordt
   vastgelegd en doorgegeven aan de groepsloop (Stages 1--2), die deze
   speler uitsluit van de hoofdmatching en een PAB toewijst.

### Waarom maximale cardinaliteit?

De `maxCardinality = true`-vlag is essentieel. Zonder deze vlag zou Blossom
puur op gewicht optimaliseren en mogelijk meerdere spelers ongematcht laten
als dat het totale gewicht zou verhogen. We hebben exact een ongematchte
speler nodig — de bye-ontvanger — en alle anderen moeten ingedeeld worden.

Onder alle maximale-cardinaliteitsmatchings (die allemaal exact een speler
ongematcht laten) selecteert Blossom die met het hoogste totale gewicht. De
3-veldsgewichtcodering zorgt ervoor dat dit de matching is die:

1. Topscorende spelers beschermt tegen het ontvangen van de bye
   (topscorerbeschermingsbit).
2. Bij gelijke beschermingsniveaus de laagst scorende bye-kandidaat
   ongematcht laat (scoresomveld).
3. Bij gelijke scores de voorkeur geeft aan het matchen van niet-bye-
   geschikte spelers samen (bye-geschiktheidsbit).

---

## Correctheidsschets

**Bewering.** De speler die door Stage 0.5 ongematcht wordt gelaten, is
degene wiens verwijdering een completeerbare matching voor de overige
spelers oplevert.

**Argument.** De Stage 0.5-Blossom matching beschouwt dezelfde absolute
criteria (C1, C3, verboden paren) als de echte matching. Als speler $p$
ongematcht blijft, betekent dit dat er een geldige matching bestaat voor
alle andere spelers. De echte matching (Stages 1--2) werkt met dezelfde
spelersset minus $p$ en voegt optimalisatiecriteria (C5--C21) toe die een
matching verfijnen maar nooit ongeldig maken als deze aan de absolute
criteria voldoet.

De enige manier waarop Stage 0.5 een "verkeerde" bye-ontvanger zou kunnen
kiezen is als de vereenvoudigde gewichtcodering een matching opleverde
waarbij de verwijdering van de ongematchte speler de resterende set
onlotbaar maakte onder de volledige criteria. Dit kan niet gebeuren omdat de
volledige criteria (C5--C21) optimalisatiedoelen zijn die in kantgewichten
zijn gecodeerd — ze beïnvloeden _welke_ matching wordt gekozen, niet _of_
er een matching bestaat. Bestaan hangt alleen af van de absolute criteria,
die Stage 0.5 volledig afdwingt.

---

## Vergelijking met andere systemen

Niet alle indelingssystemen gebruiken Stage 0.5:

| Systeem               | Bye-selectiemethode                                                               |
| --------------------- | --------------------------------------------------------------------------------- |
| Nederlands (C.04.3)   | Stage 0.5 completeerbaarheidsmatching                                             |
| Burstein (C.04.4.2)   | Stage 0.5 completeerbaarheidsmatching                                             |
| Dubov (C.04.4.1)      | Speciale `DubovByeSelector` (Art. 2.3): laagste scoregroep, hoogste rangnummer |
| Lim (C.04.4.3)        | `LimByeSelector` (Art. 1.1): laagste rang in laagste scoregroep                   |
| Double-Swiss (C.04.5) | `AssignPAB` uit lexswiss: laagste score, hoogste TPN                              |
| Team Swiss (C.04.6)   | `AssignPAB` uit lexswiss: laagste score, hoogste TPN                              |
| Keizer                | Laagste Keizer-score                                                              |
| Round-robin           | Dummyspeler (geen echte bye nodig)                                                |

De completeerbaarheidsaanpak (Nederlands, Burstein) is rekenkundig het
zwaarst maar ook het robuust: het garandeert door constructie dat de
overige spelers ingedeeld kunnen worden. De eenvoudigere selectoren van
andere systemen vertrouwen op heuristieken die in de praktijk goed werken
maar niet dezelfde structurele garantie bieden.

---

## Complexiteit

Stage 0.5 voert een Blossom matching uit op $n$ spelers met $O(n^2)$ kanten
(alle compatibele paren). Het Blossom-algoritme is $O(n^3)$. Aangezien
Stage 0.5 `int64`-gewichten gebruikt (geen `*big.Int`), is de constante
factor klein.

Voor een toernooi met 200 spelers voegt Stage 0.5 ruwweg 5--10% toe aan
de totale indelingstijd. Voor typische clubtoernooien (20--60 spelers) is
het verwaarloosbaar.

---

## Gerelateerde pagina's

- [Blossom Matching](../blossom/) — het matchingalgoritme dat Stage 0.5
  gebruikt.
- [Kantgewicht-codering](../edge-weights/) — de volledige 20-veldscodering
  van Stages 1--2 (in contrast met de vereenvoudigde 3-veldscodering van
  Stage 0.5).
- [Nederlandse criteria](../dutch-criteria/) — de absolute criteria (C1, C3)
  die kantgeschiktheid in Stage 0.5 bepalen.
