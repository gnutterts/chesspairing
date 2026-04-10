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

Gegeven een toernooi met $R$ ronden totaal en $N$ actieve spelers, definieert
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
ontvangt. Groep A bestaat uit de hoogst gerangschikte spelers (met
initiële rang $\leq \text{gaSize}$). De formule zorgt ervoor dat Groep A
altijd een even aantal spelers bevat.

---

## Virtuele-puntfunctie

Voor speler $p$ in ronde $r$ (1-geïndexeerd):

$$\text{VP}(p, r) = \begin{cases} 1.0 & \text{if } \text{rank}(p) \leq \text{gaSize} \text{ and } r \leq \text{fullVP} \\ 0.5 & \text{if } \text{rank}(p) \leq \text{gaSize} \text{ and } \text{fullVP} < r \leq \text{accelerated} \\ 0.0 & \text{otherwise} \end{cases}$$

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
   hun `PairingScore`. Dit wordt gedaan door `ApplyBakuAcceleration` in het
   swisslib-pakket.
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

**Even Groep A.** De formule $2 \cdot \lceil N/4 \rceil$ zorgt ervoor dat
Groep A altijd een even aantal spelers heeft, waardoor er geen bye nodig is
binnen de versnelde groep.

---

## Ondersteunde systemen

Baku-acceleratie wordt ondersteund door de Nederlandse en Burstein-
indelingssystemen (in te schakelen via de `Acceleration`-optie). De Dubov-,
Lim-, Double-Swiss- en Team Swiss-systemen implementeren momenteel geen
acceleratie.

---

## Gerelateerde pagina's

- [Nederlandse indeling](/docs/pairing-systems/dutch/) — het primaire systeem
  dat Baku-acceleratie gebruikt.
- [Nederlandse criteria](../dutch-criteria/) — de criteria die van toepassing
  zijn nadat acceleratie de scoregroepen heeft aangepast.
- [Completeerbaarheid](../completability/) — Stage 0.5 werkt op de versnelde
  scoregroepen.
