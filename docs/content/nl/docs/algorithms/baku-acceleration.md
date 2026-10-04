---
title: "Baku-acceleratie"
linkTitle: "Acceleratie"
weight: 6
description: "Virtuele punten in vroege ronden om voorspelbare indelingen te voorkomen (FIDE C.04.7)."
---

## Motivatie

In een standaard Zwitsers toernooi deelt de eerste ronde speler 1 in tegen speler
$\lceil N/2 \rceil + 1$, speler 2 tegen $\lceil N/2 \rceil + 2$, enzovoort.
Na ronde 1 hebben alle winnaars uit de bovenste helft 1 punt en worden ze in
ronde 2 tegen elkaar ingedeeld. Dit creëert een voorspelbaar patroon waarin de
sterkste spelers elkaar heel vroeg tegenkomen, met beslissende resultaten die
latere ronden minder competitief maken.

**Baku-acceleratie** (FIDE C.04.7, vernoemd naar de Schaakolympiade van 2016
in Bakoe waar het voor het eerst bij een groot evenement werd toegepast) doorbreekt
dit patroon door **virtuele punten** toe te kennen aan een subset van spelers
in de vroege ronden. Deze virtuele punten blazen de scoregroepen op, waardoor
spelers uit verschillende ratinglagen in dezelfde groep terechtkomen. Na de
acceleratiefase worden de virtuele punten verwijderd en nemen de echte scores
het over.

De implementatie staat in `pairing/swisslib/acceleration.go`.

---

## Definities

Gegeven een toernooi met $R$ ronden totaal en $N$ deelnemers, definieert
Baku-acceleratie vier parameters:

### Versnelde ronden

$$\text{accelerated} = \left\lceil \frac{R}{2} \right\rceil$$

Het totale aantal ronden waarin acceleratie actief is.

### Volle virtuele-puntronden

$$\text{fullVP} = \left\lceil \frac{\text{accelerated}}{2} \right\rceil$$

Ronden $1, 2, \ldots, \text{fullVP}$ kennen 1,0 virtuele punten toe aan
spelers die daarvoor in aanmerking komen.

### Halve virtuele-puntronden

$$\text{halfVP} = \text{accelerated} - \text{fullVP}$$

Ronden $\text{fullVP} + 1, \ldots, \text{accelerated}$ kennen 0,5 virtuele
punten toe aan spelers die daarvoor in aanmerking komen.

### Groep A-grootte

$$\text{gaSize} = 2 \cdot \left\lceil \frac{N}{4} \right\rceil$$

Het aantal spelers in "Groep A" — de set spelers die virtuele punten
ontvangt. $N$ telt de deelnemers vóór de eerste ronde: spelers die vanaf ronde
1 aanwezig zijn, ook met een aangevraagde bye in ronde 1 of die later zijn
teruggetrokken. Dat een aangevraagde bye in ronde 1 meetelt, is een
interpretatie. Late inschrijvingen tellen niet mee (C.04.7 1.2). Een late
inschrijving is een speler met een `JoinedRound` van 2 of hoger; een TRF-bestand
kan dat niet vastleggen, dus bij TRF-invoer (en in bbpPairings) telt elk
spelersrecord mee.

Late inschrijvingen komen volgens hun indelingsnummer op de indelingslijst. De
laatste speler van Groep A blijft dezelfde speler; een late inschrijving met
een lager indelingsnummer dan die speler komt dus in Groep A. Groep A kan daardoor oneven
worden (1.3.2, noot 2). Als de bibliotheek alle indelingsnummers tegelijk
toekent (geen nummers opgegeven), wordt elke deelnemer volgens C.04.2 2.2
geordend en komt een late inschrijving boven de laatste speler van Groep A ook
in Groep A. Als bestaande spelers al nummers hebben en een late inschrijving
niet, geeft deze implementatie die na hen een nummer, zodat die in Groep B
komt.

bbpPairings maakt Groep A juist de eerste $\lceil N/2 \rceil$ spelers. De twee
komen overeen als $N \bmod 4$ gelijk is aan 0 of 3 en verschillen anders (161
deelnemers: 82 in de FIDE-tekst, 81 in bbpPairings). bbpPairings telt elk
spelersrecord in het TRF-bestand (`src/fileformats/trf.cpp`, rond regel 724),
dus ook late inschrijvingen. Deze implementatie volgt de tekst.

---

## Virtuele-puntfunctie

Voor speler $p$ in ronde $r$ (1-geïndexeerd):

$$\text{VP}(p, r) = \begin{cases} 1.0 & \text{als } p \in \text{GA} \\ & \text{en } r \leq \text{fullVP} \\ 0.5 & \text{als } p \in \text{GA} \\ & \text{en } \text{fullVP} < r \leq \text{accelerated} \\ 0.0 & \text{anders} \end{cases}$$

De virtuele punten worden opgeteld bij de **indelingsscore** van de speler
(de score die gebruikt wordt voor groepsindeling), niet bij de werkelijke
toernooiscore. Dit betekent:

- Tijdens versnelde ronden lijken Groep A-spelers hogere scores te hebben
  dan ze werkelijk hebben, waardoor ze in hogere groepen terechtkomen.
- Tiebreakers en de eindstand gebruiken de echte scores, niet de opgeblazen
  indelingsscores.
- Na ronde $\text{accelerated}$ zijn alle virtuele punten nul en verloopt de
  indeling normaal.

---

## Effect op groepen

### Zonder acceleratie

In een toernooi met 100 spelers en 9 ronden na ronde 1:

- Scoregroep 1,0: ~50 spelers (alle winnaars)
- Scoregroep 0,5: ~0 spelers (uitgaande van geen remises voor de eenvoud)
- Scoregroep 0,0: ~50 spelers (alle verliezers)

Ronde 2 deelt de 50 winnaars tegen elkaar in: speler 1 tegen ~speler 25, speler
2 tegen ~speler 26, enz. De topgeplaatsten ontmoeten direct sterke
tegenstanders.

### Met acceleratie

Hetzelfde toernooi heeft $\text{gaSize} = 2 \cdot \lceil 100/4 \rceil = 50$
en $\text{fullVP} = \lceil \lceil 9/2 \rceil / 2 \rceil = 3$. In ronde 1:

- Groep A-spelers (rang 1--50) hebben indelingsscore $0{,}0 + 1{,}0 = 1{,}0$.
- Groep B-spelers (rang 51--100) hebben indelingsscore $0{,}0$.

Ronde 1 deelt in binnen deze opgeblazen groepen. Groep A's groep van 50 spelers
levert indelingen op als speler 1 tegen speler 26 (vergelijkbaar met zonder
acceleratie).

Na ronde 1 heeft een Groep A-winnaar indelingsscore $1{,}0 + 1{,}0 = 2{,}0$
voor ronde 2. Een Groep B-winnaar heeft $1{,}0 + 0{,}0 = 1{,}0$. De
groepsstructuur is nu:

- Indelingsscore 2,0: ~25 Groep A-winnaars
- Indelingsscore 1,0: ~25 Groep A-verliezers + ~25 Groep B-winnaars
- Indelingsscore 0,0: ~25 Groep B-verliezers

Ronde 2 deelt de 25 Groep A-winnaars tegen elkaar in, maar de interessante groep
is score 1,0, die Groep A-verliezers mengt met Groep B-winnaars — spelers
uit verschillende ratinglagen die elkaar zonder acceleratie zo vroeg niet
zouden ontmoeten.

---

## Uitgewerkt voorbeeld

Toernooi: 20 spelers, 7 ronden.

Parameters:

$$\text{accelerated} = \lceil 7/2 \rceil = 4$$
$$\text{fullVP} = \lceil 4/2 \rceil = 2$$
$$\text{halfVP} = 4 - 2 = 2$$
$$\text{gaSize} = 2 \cdot \lceil 20/4 \rceil = 10$$

Schema virtuele punten:

| Ronde | VP voor Groep A (rang 1--10) | VP voor Groep B (rang 11--20) |
| ----- | ---------------------------- | ----------------------------- |
| 1     | 1,0                          | 0,0                           |
| 2     | 1,0                          | 0,0                           |
| 3     | 0,5                          | 0,0                           |
| 4     | 0,5                          | 0,0                           |
| 5--7  | 0,0                          | 0,0                           |

De overgang van 1,0 naar 0,5 virtuele punten in ronde 3 zorgt voor een
geleidelijke "landing" in plaats van een abrupte verwijdering. Vanaf ronde 5
spelen alle spelers op basis van hun echte scores.

---

## Toepassing in de indelingspijplijn

Baku-acceleratie integreert in de Zwitserse indelingspijplijn bij de stap waar
scoregroepen worden samengesteld:

1. **Bouw spelerstaten** op uit de toernooi-historie.
2. **Pas acceleratie toe.** Voeg voor elke speler $\text{VP}(p, r)$ toe aan
   de score. Dit wordt gedaan door `AddVirtualPoints` in het swisslib-pakket,
   dat de spelers ook ordent op de resulterende indelingsscore. Vanaf hier is
   de indelingsscore de score voor alles wat de indeling bepaalt: de
   scoregroepen, de criteria, de keuze van de indelingsbye en de bordvolgorde
   (C.04.7 1.5). Wie in een eerdere ronde is gefloat, wordt ook met de
   indelingsscore van die ronde beoordeeld, zodat de virtuele punten van die
   ronde daar meetellen. De topscorerregel van de laatste ronde blijft de echte
   scores gebruiken.
3. **Bouw scoregroepen** met de aangepaste indelingsscores.
4. **Ga verder met de normale indeling** (groepsopbouw, Blossom matching,
   enz.).

De acceleratie is transparant voor de rest van de indelingslogica. Scoregroepen
en groepen werken met de opgeblazen scores zonder speciale behandeling.

---

## Eigenschappen

**Transitiviteit.** De toevoeging van virtuele punten behoudt de relatieve
volgorde binnen Groep A en binnen Groep B. Het verandert alleen de
kruislingse volgorde door Groep A boven Groep B te tillen.

**Convergentie.** Naarmate de ronden vorderen, domineren echte
scoreverschillen over de virtuele punten. Na de acceleratiefase is de indeling
volledig scoregestuurd. De eindstand van het toernooi wordt niet beïnvloed.

**Even Groep A.** De formule $2 \cdot \lceil N/4 \rceil$ zorgt ervoor
dat Groep A bij de start een even aantal spelers heeft. Na de eerste ronde kan Groep A door late
inschrijvingen oneven zijn.

---

## Inschakelen

Zet de indelingsoptie `acceleration` op `"baku"`. bbpPairings leest alleen
record `192` met een `FIDE_*_BAKU`-code en negeert `XXS`; deze implementatie
leest beide. De implementatie van het Dutch-systeem is vergeleken met bbpPairings op
tienduizenden gegenereerde rondes met en zonder Baku-acceleratie (zie
[Testen](/docs/appendices/testing/)), behalve voor de configuratie met 9
spelers, waar de twee verschillen zoals hierboven beschreven.

De puntentelling moet een winstpartij evenveel laten opleveren als twee
remises en een verliespartij niets (1.1); anders geeft de indeling een fout. De indelingsscores staan op de schaal
1-½-0, dus de virtuele punten worden daar als 1 en ½ opgeteld; bij een
telling van 2-1-0 komt dat overeen met 2 en 1.
Dutch en Burstein hebben ook het totale aantal rondes nodig.

---

## Ondersteunde systemen

Baku-acceleratie wordt ondersteund door het Dutch- en het Burstein-systeem
(in te schakelen via de `Acceleration`-optie). De Dubov-,
Lim-, Double-Swiss- en Team Swiss-systemen implementeren momenteel geen
acceleratie.

---

## Gerelateerde pagina's

- [Dutch-systeem](/docs/pairing-systems/dutch/) — het primaire systeem
  dat Baku-acceleratie gebruikt.
- [Dutch-criteria](../dutch-criteria/) — de criteria die van toepassing
  zijn nadat acceleratie de scoregroepen heeft aangepast.
- [Completeerbaarheid](../completability/) — Stage 0.5 werkt op de versnelde
  scoregroepen.
