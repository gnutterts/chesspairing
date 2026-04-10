---
title: "Bergertabel-rotatie"
linkTitle: "Bergertabellen"
weight: 4
description: "FIDE-Bergertabellen voor round-robin-planning — het rotatie-algoritme en de laatste-twee-rondenwissel."
---

## Het planningsprobleem

Een **round-robin**-toernooi vereist dat elke speler exact eenmaal tegen elke
andere speler speelt (enkel round-robin) of exact tweemaal (dubbel
round-robin). Voor $N$ spelers heeft een enkel round-robin
$\binom{N}{2} = \frac{N(N-1)}{2}$ partijen verdeeld over $N - 1$ ronden
(of $N$ ronden als $N$ oneven is, met een bye per ronde).

Johann Berger publiceerde in 1895 een systematische schema-opbouw die door
de FIDE als standaard is aangenomen (C.05 Annex 1). De methode houdt een
speler op een vaste positie en roteert alle anderen, wat een gebalanceerd
schema oplevert met goede kleurafwisselingseigenschappen.

De implementatie staat in `pairing/roundrobin/roundrobin.go`.

---

## Opzet

Laat $N$ het aantal spelers zijn. Als $N$ oneven is, voeg een **dummyspeler**
toe (genummerd $N$) om het aantal even te maken; elke speler die tegen de
dummy wordt ingedeeld krijgt een bye. Stel $n = N$ als even, $n = N + 1$ als
oneven.

Nummer de posities $0, 1, 2, \ldots, n - 1$. De speler op positie $n - 1$
is **vast** (de "draaispil"). De overige $n - 1$ spelers roteren.

---

## Rotatieformule

Voor ronde $r$ (0-geindexeerd) is de speler op positie $j$ in ronde $r$ de
speler die oorspronkelijk op de volgende positie stond:

$$\text{positions}[j] = (j - r \cdot s) \bmod (n - 1) \quad \text{for } j < n - 1$$

waarbij de **stap** is:

$$s = \frac{n}{2} - 1$$

De speler op positie $n - 1$ beweegt niet. Na $n - 1$ ronden is elke
niet-vaste positie door elke speler exact eenmaal bezet, en is elk
spelerspaar exact eenmaal gepland.

### Waarom deze stap?

De stap $s = n/2 - 1$ is zo gekozen dat elke rotatie spelers ongeveer een
halve tafel vooruit schuift. Dit maximaliseert de kleurafwisseling: een
speler die in de ene ronde wit had, heeft waarschijnlijk zwart in de
volgende, omdat hij naar de tegenovergestelde kant van de indelingstafel
verhuist.

De keuze van stap is niet uniek — elke waarde die onderling priem is met
$n - 1$ levert een geldig schema op. De Berger-stap $n/2 - 1$ is de
standaard omdat hij de kleurbalans-eigenschappen optimaliseert.

---

## Indelingsopbouw

In elke ronde $r$ worden de spelers op posities als volgt ingedeeld:

1. **Bord 1**: positie $0$ tegen positie $n - 1$ (de vaste speler).
2. **Bord $k$** (voor $k = 2, 3, \ldots, n/2$): positie $k - 1$ tegen
   positie $n - 1 - (k - 1) = n - k$.

Dit levert $n/2$ borden per ronde op. Als $N$ oneven was, krijgt de speler
die tegen de dummy is ingedeeld een bye in plaats van een partij.

---

## Kleurtoewijzing

Kleuren worden toegewezen volgens de FIDE-conventie:

- **Bord 1**: de roterende speler (positie $0$) wisselt elke ronde van
  kleur. In ronde 0 speelt hij wit; in ronde 1 zwart; enzovoort.
  Equivalent: de vaste speler (positie $n - 1$) speelt zwart in even ronden
  en wit in oneven ronden.
- **Overige borden**: de speler met de lagere positie-index speelt wit.

Formeel, voor bord $k > 1$ met spelers op posities $a < b$:

$$\text{White} = \text{player at position } a, \quad \text{Black} = \text{player at position } b$$

Voor bord 1 in ronde $r$:

$$\text{White} = \begin{cases} \text{rotating player} & \text{if } r \text{ is even} \\ \text{fixed player} & \text{if } r \text{ is odd} \end{cases}$$

---

## Dubbel round-robin

Een dubbel round-robin bestaat uit twee **cycli**. In cyclus 2 wordt elke
indeling uit cyclus 1 herhaald met omgekeerde kleuren:

- Als speler $A$ wit had tegen speler $B$ in cyclus 1, dan heeft $B$ wit
  tegen $A$ in cyclus 2.

De rondenummering loopt door: cyclus 1 gebruikt ronden $0, 1, \ldots, n - 2$
en cyclus 2 gebruikt ronden $n - 1, n, \ldots, 2(n - 1) - 1$.

### De laatste-twee-rondenwissel

Er ontstaat een probleem bij de cyclusgrens. In de laatste ronde van cyclus 1
en de eerste ronde van cyclus 2 komen dezelfde spelersparen voor met
omgekeerde kleuren. Een speler die wit had in ronde $n - 2$ zou zwart hebben
in ronde $n - 1$ tegen dezelfde tegenstander, en dan onmiddellijk een andere
tegenstander tegenkomen in ronde $n$ met de kleur die doorloopt vanaf ronde
$n - 2$. Dit kan sequenties van drie opeenvolgende partijen met dezelfde
kleur veroorzaken.

De door de FIDE aanbevolen oplossing is om **de laatste twee ronden van
cyclus 1 om te wisselen** (niet cyclus 2). Wanneer de optie
`SwapLastTwoRounds` is ingeschakeld:

- Ronde $n - 3$ krijgt de indelingen die oorspronkelijk voor ronde $n - 2$
  gepland waren.
- Ronde $n - 2$ krijgt de indelingen die oorspronkelijk voor ronde $n - 3$
  gepland waren.

Dit doorbreekt het drie-opeenvolgende-kleurenpatroon ten koste van een
kleine schema-onregelmatigheid in de laatste twee ronden van de eerste
cyclus.

---

## Uitgewerkt voorbeeld: 6 spelers

Met $N = 6$ (even), $n = 6$, stap $s = 6/2 - 1 = 2$.

Posities in ronde 0: spelers 0, 1, 2, 3, 4 roteren; speler 5 is vast.

| Ronde | Posities na rotatie | Bord 1 | Bord 2 | Bord 3 |
| ----- | ------------------- | ------ | ------ | ------ |
| 0     | 0 1 2 3 4 **5**     | 0 vs 5 | 1 vs 4 | 2 vs 3 |
| 1     | 3 4 0 1 2 **5**     | 3 vs 5 | 4 vs 2 | 0 vs 1 |
| 2     | 1 2 3 4 0 **5**     | 1 vs 5 | 2 vs 0 | 3 vs 4 |
| 3     | 4 0 1 2 3 **5**     | 4 vs 5 | 0 vs 3 | 1 vs 2 |
| 4     | 2 3 4 0 1 **5**     | 2 vs 5 | 3 vs 1 | 4 vs 0 |

Elk paar komt exact eenmaal voor in 5 ronden. Bord 1 wisselt de kleur van de
roterende speler: wit in ronden 0, 2, 4; zwart in ronden 1, 3.

---

## Oneven aantal spelers

Voor $N = 5$ wordt dummyspeler 5 toegevoegd zodat $n = 6$. Het schema is
identiek aan het bovenstaande voorbeeld, maar elke partij met speler 5 wordt
een bye voor de tegenstander:

- Ronde 0: speler 0 heeft een bye (was ingedeeld tegen dummy 5).
- Ronde 1: speler 2 heeft een bye.
- Enzovoort.

Elke speler krijgt exact een bye gedurende het toernooi.

---

## Eigenschappen

**Volledigheid.** Elk paar echte spelers wordt exact eenmaal per cyclus
gepland. Dit volgt uit het feit dat de rotatie een cyclische permutatie van
orde $n - 1$ is op de niet-vaste posities.

**Kleurbalans.** Elke speler speelt maximaal $\lceil (n-1)/2 \rceil$
partijen met een kleur en minimaal $\lfloor (n-1)/2 \rfloor$ met de andere.
De vaste speler wisselt perfect af. Roterende spelers hebben bijna perfecte
afwisseling dankzij de stapkeuze.

**Rondeaantal.** Enkel round-robin: $n - 1$ ronden. Dubbel round-robin:
$2(n - 1)$ ronden.

---

## Complexiteit

De schema-opbouw is $O(n^2)$ — elk van de $n - 1$ ronden levert $n/2$
indelingen op, wat $O(n^2)$ totale bewerkingen geeft. De rotatie zelf is
$O(1)$ per speler per ronde (een modulaire rekenbewerking).

---

## Gerelateerde pagina's

- [Round-robin-indeling](/docs/pairing-systems/round-robin/) — het
  indelingssysteem dat Bergertabellen gebruikt.
- [Varma-tabellen](../varma-tables/) — federatiebewuste toewijzing van
  rangnummers die bepaalt _welke_ speler op _welke_ positie zit voordat
  de Berger-rotatie begint.
