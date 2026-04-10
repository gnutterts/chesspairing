---
title: "Nederlandse criteria"
linkTitle: "Nederlandse criteria"
weight: 10
description: "De 21 criteria (C1-C21) die Nederlandse Zwitserse indelingen besturen — absolute beperkingen en optimalisatiedoelen."
---

## Overzicht

Het Nederlandse Zwitserse systeem (FIDE C.04.3) definieert 21 criteria
genummerd $C_1$ tot en met $C_{21}$ die besturen hoe spelers worden ingedeeld.
Deze criteria vormen een strikte prioriteitshierarchie: $C_i$ heeft absolute
voorrang op $C_j$ wanneer $i < j$.

De criteria vallen uiteen in twee categorieën:

- **Absolute criteria** ($C_1$--$C_4$): beperkingen waaraan _moet_ worden
  voldaan. Een indeling die een absoluut criterium schendt, wordt volledig
  afgewezen.
- **Optimalisatiecriteria** ($C_5$--$C_{21}$): doelen die worden
  _gemaximaliseerd_ via [kantgewicht-codering](../edge-weights/). Schendingen
  worden beboet in het Blossom-matchinggewicht, waarbij hogere-prioriteits-
  criteria meer-significante bits bezetten.

De absolute criteria zijn geïmplementeerd in `pairing/swisslib/criteria.go`.
De optimalisatiecriteria worden gecodeerd in
`pairing/swisslib/criteria_pairs.go` (zie
[Kantgewicht-codering](../edge-weights/)).

---

## Absolute criteria

### C1: geen herparingen

Twee spelers die al tegen elkaar gespeeld hebben, worden niet opnieuw ingedeeld.

$$\text{C1}(i, j) = \begin{cases} \text{pass} & \text{if } (i, j) \notin H \\ \text{fail} & \text{otherwise} \end{cases}$$

waarbij $H$ de verzameling gespeelde indelingen is. Forfaits worden uitgesloten
van de historie: een partij die door forfait is verloren telt niet als
"gespeeld" voor C1-doeleinden, wat betekent dat de spelers _opnieuw_ ingedeeld
kunnen worden. Dubbele forfaits worden eveneens uitgesloten.

Implementatie: `C1NoRematches` in `pairing/swisslib/criteria.go`.

### C2: geen tweede PAB

Een speler die al een indeling-toegewezen bye (PAB) heeft ontvangen, mag geen
tweede ontvangen.

Dit criterium wordt afgedwongen tijdens de bye-selectie, niet tijdens de
indelingsopbouw. De [completeerbaarheids](../completability/)-matching en de
bye-selectoren zorgen ervoor dat alleen PAB-geschikte spelers de bye kunnen
ontvangen.

Implementatie: `C2NoSecondPAB` in `pairing/swisslib/criteria.go`.

### C3: geen absoluut kleurconflict

Twee niet-topscorers die beide een absolute kleurvoorkeur voor dezelfde
kleur hebben, worden niet ingedeeld.

Een speler heeft een **absolute kleurvoorkeur** wanneer:

- Zijn kleuronbalans groter is dan 1 (bijv. 3 wit tegen 1 zwart), of
- Hij 2 of meer opeenvolgende partijen met dezelfde kleur heeft gespeeld.

Als beide spelers een absolute voorkeur voor (zeg) wit hebben, zou het indelen
van hen een van beiden dwingen zwart te spelen ondanks de absolute voorkeur —
wat de kleurregels schendt.

$$\text{C3}(i, j) = \begin{cases} \text{fail} & \text{if both have absolute preference for the same color} \\ & \text{and neither is a top scorer} \\ \text{pass} & \text{otherwise} \end{cases}$$

**Topscorer-uitzondering.** In de laatste ronde, wanneer beide spelers
topscorers zijn (in de hoogste niet-lege scoregroep), wordt C3 versoepeld
om de indeling toe te staan. Dit voorkomt situaties waarin de toernooileiders
niet ingedeeld kunnen worden vanwege kleurbeperkingen.

Implementatie: `C3AbsoluteColorConflict` in `pairing/swisslib/criteria.go`.

### C4: groepsvolledigheid

Geen per-paarcriterium maar een structurele eis: na verwerking van een
scoregroep moeten alle spelers ofwel ingedeeld zijn ofwel naar een aangrenzende
groep zijn gedreven. Geen speler mag "gestrand" achterblijven zonder een
indeling of een floatbestemming.

Implementatie: gevalideerd tijdens de groepsloop in
`pairing/swisslib/global_matching.go`.

---

## Optimalisatiecriteria

De optimalisatiecriteria worden gecodeerd als bitvelden in het Blossom-
kantgewicht. Zie [Kantgewicht-codering](../edge-weights/) voor de volledige
bitindeling. Hier beschrijven we de semantische betekenis van elk criterium.

### C5: maximaliseer paren in huidige groep

Maximaliseer binnen elke scoregroep het aantal spelers dat ingedeeld wordt
tegen tegenstanders uit _dezelfde_ scoregroep (in tegenstelling tot
floaters uit aangrenzende groepen).

**Kantgewichtveld.** 1 bit (veld 2, breedte $\text{sgBits}$): gezet wanneer
beide spelers tot de huidige scoregroep behoren.

### C6: maximaliseer scoresom in huidige groep

Maximaliseer onder paren binnen de huidige scoregroep de som van de scores.
Dit geeft de voorkeur aan het indelen van hoger scorende spelers binnen de
groep boven lager scorende.

**Kantgewichtveld.** Score-geindexeerde subvelden (veld 3, breedte
$\text{sgsShift}$).

### C7: maximaliseer paren in volgende groep

Wanneer spelers moeten afdrijven naar de volgende scoregroep, maximaliseer
het aantal zulke afdrijfindelingen. Dit zorgt ervoor dat de groep soepel
doorloopt naar de volgende.

**Kantgewichtveld.** 1 bit (veld 4, breedte $\text{sgBits}$): gezet wanneer
de lagere speler in de volgende scoregroep zit.

### C8: maximaliseer scoresom in volgende groep

Analoog aan C6 maar voor de uitbreiding naar de volgende groep.

**Kantgewichtveld.** Score-geindexeerde subvelden (veld 5, breedte
$\text{sgsShift}$).

### C9: minimaliseer ongespeelde partijen bye-ontvanger

Wanneer een bye moet worden toegewezen, geef de voorkeur aan de speler met
de minste ongespeelde partijen (onder bye-geschikte kandidaten). Dit wordt
gecodeerd als twee subvelden voor de lagere en hogere speler in elk paar.

**Kantgewichtvelden.** Velden 6--7, elk $\text{sgBits}$ breed.

### C10--C13: kleurcriteria

Vier criteria besturen de compatibiliteit van kleurvoorkeuren, in aflopende
prioriteit:

| Criterium | Betekenis                                                                                                                                          |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| C10       | Geen absoluut onbalansconflict: vermijd het indelen van twee spelers die beide kleuronbalans $> 1$ hebben en dezelfde kleur prefereren.              |
| C11       | Geen absoluut voorkeurconflict: een meer genuanceerde controle rekening houdend met de grootte van de onbalans en opeenvolgende-kleurgeschiedenis. |
| C12       | Kleurvoorkeuren compatibel: de twee spelers prefereren verschillende kleuren, of ten minste een heeft geen voorkeur.                               |
| C13       | Geen sterk voorkeurconflict: vermijd het indelen van twee spelers met sterke (maar niet absolute) voorkeuren voor dezelfde kleur.                    |

**Kantgewichtvelden.** Velden 8--11, elk $\text{sgBits}$ breed.

Zie [Kleurverdeling](../color-allocation/) voor hoe deze voorkeuren
worden berekend en opgelost.

### C14--C15: floatherhaling vermijden (ronde $R-1$)

Deze criteria ontmoedigen het herhalen van floatpatronen uit de direct
voorgaande ronde:

| Criterium | Betekenis                                                                                                                    |
| --------- | ---------------------------------------------------------------------------------------------------------------------------- |
| C14       | Minimaliseer het aantal spelers dat in ronde $R - 1$ afdreef en opnieuw afdrijft.                                            |
| C15       | Vermijd het indelen van een upfloater uit ronde $R - 1$ tegen een hoger scorende tegenstander (waardoor hij opnieuw opdrijft). |

**Kantgewichtvelden.** Velden 12--13, elk $\text{sgBits}$ breed.
Voorwaardelijk: alleen aanwezig wanneer ten minste 1 ronde is gespeeld.

### C16--C17: floatherhaling vermijden (ronde $R-2$)

Hetzelfde als C14--C15 maar dan voor twee ronden geleden:

| Criterium | Betekenis                                                                |
| --------- | ------------------------------------------------------------------------ |
| C16       | Minimaliseer herhaalde afdrijvers uit ronde $R - 2$.                     |
| C17       | Vermijd herhaalde upfloater-tegen-hogere-tegenstander uit ronde $R - 2$. |

**Kantgewichtvelden.** Velden 16--17, elk $\text{sgBits}$ breed.
Voorwaardelijk: alleen aanwezig wanneer ten minste 2 ronden zijn gespeeld.

### C18--C19: floatscore minimaliseren (ronde $R-1$)

| Criterium | Betekenis                                                                                                                                          |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| C18       | Minimaliseer de score van afdrijvers uit ronde $R - 1$. Een speler die met een hoge score afdreef, zou niet opnieuw moeten afdrijven.              |
| C19       | Minimaliseer de tegenstander-score van upfloaters uit ronde $R - 1$. Een upfloater zou de laagst scorende beschikbare tegenstander moeten treffen. |

**Kantgewichtvelden.** Velden 14--15, elk $\text{sgsShift}$ breed.
Score-geindexeerde subvelden bieden granulaire optimalisatie.

### C20--C21: floatscore minimaliseren (ronde $R-2$)

Hetzelfde als C18--C19 maar dan voor twee ronden geleden:

| Criterium | Betekenis                                                            |
| --------- | -------------------------------------------------------------------- |
| C20       | Minimaliseer de score van afdrijvers uit ronde $R - 2$.              |
| C21       | Minimaliseer de tegenstander-score van upfloaters uit ronde $R - 2$. |

**Kantgewichtvelden.** Velden 18--19, elk $\text{sgsShift}$ breed.
Voorwaardelijk: alleen aanwezig wanneer ten minste 2 ronden zijn gespeeld.

---

## Floatgeschiedenis

Verschillende optimalisatiecriteria verwijzen naar de **floatgeschiedenis**
van een speler — of hij in voorgaande ronden op- of afdreef. De
floatrichting wordt bepaald door de score van een speler te vergelijken met
de scoregroep waarin hij werd ingedeeld:

- **Afdrijven**: de score van de speler is hoger dan de groep waarin hij
  werd ingedeeld (hij "daalde" om een tegenstander te vinden).
- **Opdrijven**: de score van de speler is lager dan de groep waarin hij
  werd ingedeeld (hij werd "omhooggetrokken" om een groep te vullen).
- **Geen float**: de speler werd ingedeeld binnen zijn eigen scoregroep.

De helper `floatAtRound(p, roundIdx)` in
`pairing/swisslib/criteria_optimization.go` haalt de floatrichting op
voor een specifieke eerdere ronde.

---

## Interactie van criteria met kantgewichten

De 21 criteria worden niet sequentieel toegepast. In plaats daarvan worden
de optimalisatiecriteria _gelijktijdig_ gecodeerd in elk kantgewicht via de
bitindeling beschreven in [Kantgewicht-codering](../edge-weights/). Het
Blossom-algoritme vindt vervolgens de maximum-gewichtmatching, die
automatisch alle criteria in de juiste prioriteitsvolgorde afhandelt.

Dit is het kernidee van de bbpPairings-architectuur (die deze implementatie
volgt): in plaats van criteria een voor een te verwerken met backtracking,
codeer ze allemaal in een enkel getal en laat het matchingalgoritme de
optimalisatie afhandelen.

De absolute criteria ($C_1$, $C_3$, verboden paren) worden anders behandeld:
zij bepalen welke kanten _bestaan_ in de graaf. Een paar dat een absoluut
criterium schendt, heeft simpelweg geen kant, zodat Blossom het niet kan
selecteren.

---

## Overzichtstabel criteria

| Criterium | Type          | Beschrijving                              | Kantgewichtveld |
| --------- | ------------- | ----------------------------------------- | --------------- |
| C1        | Absoluut      | Geen herparingen                          | Kantbestaan     |
| C2        | Absoluut      | Geen tweede PAB                           | Bye-selectie    |
| C3        | Absoluut      | Geen absoluut kleurconflict               | Kantbestaan     |
| C4        | Structureel   | Groepsvolledigheid                        | Groepsloop      |
| C5        | Optimalisatie | Maximaliseer binnen-groepsparen           | Veld 2          |
| C6        | Optimalisatie | Maximaliseer binnen-groepsscoresom        | Veld 3          |
| C7        | Optimalisatie | Maximaliseer volgende-groepsparen         | Veld 4          |
| C8        | Optimalisatie | Maximaliseer volgende-groepsscoresom      | Veld 5          |
| C9        | Optimalisatie | Minimaliseer ongespeelde partijen bye     | Velden 6--7     |
| C10       | Optimalisatie | Geen absoluut onbalansconflict            | Veld 8          |
| C11       | Optimalisatie | Geen absoluut voorkeurconflict            | Veld 9          |
| C12       | Optimalisatie | Kleurvoorkeuren compatibel                | Veld 10         |
| C13       | Optimalisatie | Geen sterk voorkeurconflict               | Veld 11         |
| C14       | Optimalisatie | Geen herhaald afdrijven ($R-1$)           | Veld 12         |
| C15       | Optimalisatie | Geen herhaald opdrijven ($R-1$)           | Veld 13         |
| C16       | Optimalisatie | Geen herhaald afdrijven ($R-2$)           | Veld 16         |
| C17       | Optimalisatie | Geen herhaald opdrijven ($R-2$)           | Veld 17         |
| C18       | Optimalisatie | Minimaliseer afdrijfscore ($R-1$)         | Veld 14         |
| C19       | Optimalisatie | Minimaliseer upfloat-tegenstander ($R-1$) | Veld 15         |
| C20       | Optimalisatie | Minimaliseer afdrijfscore ($R-2$)         | Veld 18         |
| C21       | Optimalisatie | Minimaliseer upfloat-tegenstander ($R-2$) | Veld 19         |

---

## Gerelateerde pagina's

- [Kantgewicht-codering](../edge-weights/) — hoe deze criteria
  Blossom-kantgewichten worden.
- [Completeerbaarheid](../completability/) — hoe de bye-ontvanger (C2/C9)
  wordt bepaald.
- [Kleurverdeling](../color-allocation/) — hoe kleurvoorkeuren (C10--C13)
  na de indeling worden opgelost.
- [Baku-acceleratie](../baku-acceleration/) — hoe virtuele punten de
  scoregroepen wijzigen waarop C5--C8 werken.
- [Nederlandse indeling](/docs/pairing-systems/dutch/) — het indelingssysteem
  dat door deze criteria wordt bestuurd.
