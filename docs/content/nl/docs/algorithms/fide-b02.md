---
title: "FIDE B.02 Conversietabel"
linkTitle: "B.02 Tabel"
weight: 9
description: "De opzoektabel die winstpercentages omzet naar ratingverschillen en omgekeerd."
---

## Doel

Het [Elo-kansmodel](../elo-model/) definieert een continue functie die
ratingverschillen vertaalt naar verwachte scores. In de praktijk gebruikt
FIDE de continue formule niet rechtstreeks. In plaats daarvan biedt
**FIDE-reglement B.02 Tabel 8.1b** een discrete opzoektabel met 101
ingangen, en alle officiële berekeningen gebruiken deze tabel met
interpolatie.

De implementatie in `tiebreaker/ratingtable.go` slaat deze tabel op en
biedt twee opzoekfuncties: `dpFromP` (score naar ratingverschil) en
`expectedScore` (ratingverschil naar score).

---

## De tabel

De tabel koppelt fractionele scores $p \in [0.00, 1.00]$ aan
ratingverschillen $d_p \in [-800, +800]$. Enkele ingangen:

| $p$  | $d_p$  | $p$  | $d_p$ |
| ---- | ------ | ---- | ----- |
| 0.00 | $-800$ | 0.50 | $0$   |
| 0.01 | $-677$ | 0.55 | $36$  |
| 0.05 | $-470$ | 0.60 | $72$  |
| 0.10 | $-366$ | 0.65 | $110$ |
| 0.15 | $-296$ | 0.70 | $149$ |
| 0.20 | $-240$ | 0.75 | $193$ |
| 0.25 | $-193$ | 0.80 | $240$ |
| 0.30 | $-149$ | 0.85 | $296$ |
| 0.35 | $-110$ | 0.90 | $366$ |
| 0.40 | $-72$  | 0.95 | $470$ |
| 0.45 | $-36$  | 0.99 | $677$ |
|      |        | 1.00 | $800$ |

De tabel is symmetrisch rond $p = 0.50$: $d_p = -d_{1-p}$.

De volledige tabel van 101 ingangen wordt opgeslagen als een constante
array, geïndexeerd op $\lfloor 100 \cdot p \rfloor$.

---

## Voorwaartse opzoeking: dpFromP

Gegeven een fractionele score $p$ (score gedeeld door het aantal partijen),
vind het bijbehorende ratingverschil $d_p$.

### Algoritme

1. **Begrens** $p$ tot $[0, 1]$.
2. **Schaal** naar de tabelindex: $i = p \times 100$.
3. **Geheel deel**: $\lfloor i \rfloor$ geeft de onderste tabelindex.
4. **Fractioneel deel**: $f = i - \lfloor i \rfloor$.
5. **Interpoleer**:

$$d_p = d[\lfloor i \rfloor] + f \cdot \left(d[\lceil i \rceil] - d[\lfloor i \rfloor]\right)$$

Als $\lfloor i \rfloor = \lceil i \rceil$ (exacte tabelingang), is er
geen interpolatie nodig.

### Randgevallen

- $p = 0.00$: geeft $-800$ (minimaal ratingverschil).
- $p = 1.00$: geeft $+800$ (maximaal ratingverschil).

Dit zijn de conventionele begrenzingen van FIDE. Een speler die al zijn
partijen wint wordt behandeld alsof hij 800 punten boven de gemiddelde
tegenstander presteert; een speler die alles verliest alsof hij 800 punten
eronder zit.

---

## Inverse opzoeking: expectedScore

Gegeven een ratingverschil $d$, vind de verwachte fractionele score
$E(d)$. Dit is het omgekeerde van `dpFromP`.

### Algoritme

1. **Begrens** $d$ tot $[-800, +800]$.
2. **Binair zoeken** door de tabel (die gesorteerd is op $d_p$) om de twee
   ingangen te vinden die $d$ omsluiten:

   $$d[j] \leq d < d[j+1]$$

3. **Interpoleer**:

$$E = \frac{j}{100} + \frac{d - d[j]}{d[j+1] - d[j]} \cdot \frac{1}{100}$$

Dit zet de tabelindex terug om naar een fractionele score, met interpolatie
binnen het omsluitende interval.

### Randgevallen

- $d \leq -800$: geeft $0.00$.
- $d \geq +800$: geeft $1.00$.

---

## Interpolatienauwkeurigheid

De tabelingangen liggen op intervallen van 1% in $p$, dus de maximale
interpolatiefout hangt af van de kromming van de Elo-functie binnen elk
interval. Rond $p = 0.50$ (waar de functie vrijwel lineair is) is de fout
verwaarloosbaar. Bij de extremen ($p$ dicht bij 0 of 1) buigt de functie
scherp en is de interpolatie minder nauwkeurig.

Voor de waarden die in FIDE-tiebreaker-berekeningen worden gebruikt, valt
de interpolatiefout ruim binnen de afrondingstolerantie van 1 ratingpunt.
De uiteindelijke TPR- en PTP-waarden worden afgerond op gehele getallen,
waarmee sub-eenheidsartefacten van interpolatie worden geabsorbeerd.

---

## Gebruik in tiebreakers

### Tournament Performance Rating (TPR)

De `performancerating`-tiebreaker berekent:

$$\text{TPR} = \text{ARO} + d_p\!\left(\frac{S}{n}\right)$$

waarbij $S$ de score van de speler is en $n$ het aantal gerate partijen
(exclusief forfaits en byes). De $d_p$-opzoeking gebruikt deze tabel.

### Performance Points (PTP)

De `performancepoints`-tiebreaker gebruikt `expectedScore` iteratief
tijdens binair zoeken:

$$\sum_{i=1}^{n} E(R - R_i) \stackrel{?}{\geq} S$$

Elke $E(R - R_i)$-aanroep doorloopt de inverse opzoeking in deze tabel.

### Average Opponent TPR (APRO)

Berekent de TPR voor elke tegenstander en neemt vervolgens het gemiddelde.
De TPR van elke tegenstander gebruikt de $d_p$-opzoeking.

### Average Opponent PTP (APPO)

Berekent de PTP voor elke tegenstander en neemt vervolgens het gemiddelde.
De PTP van elke tegenstander gebruikt de $E(d)$-opzoeking.

---

## Tabel versus formule

De waarden in de FIDE B.02-tabel liggen _dicht bij_ maar zijn _niet
identiek aan_ de continue logistische formule:

$$d_{\text{formula}}(p) = -400 \cdot \log_{10}\!\left(\frac{1}{p} - 1\right)$$

Bijvoorbeeld:

| $p$  | Tabel $d_p$ | Formule $d_{\text{formula}}$ | Verschil |
| ---- | ----------- | ---------------------------- | -------- |
| 0.50 | 0           | 0.0                          | 0        |
| 0.60 | 72          | 72.2                         | 0.2      |
| 0.70 | 149         | 146.8                        | 2.2      |
| 0.80 | 240         | 240.8                        | 0.8      |
| 0.90 | 366         | 366.0                        | 0.0      |
| 0.95 | 470         | 476.2                        | 6.2      |

De verschillen zijn klein (binnen 7 ratingpunten) maar bestaan omdat de
FIDE-tabel oorspronkelijk is afgeleid van normale-verdelingbenaderingen
en om historische redenen behouden is gebleven. De implementatie gebruikt
de tabel, niet de formule, om overeen te komen met de officiële
FIDE-berekeningen.

---

## Gerelateerde pagina's

- [Elo-kansmodel](../elo-model/) -- het continue model dat de tabel
  discretiseert.
- [Prestatietiebreakers](/docs/tiebreakers/performance/) -- de
  tiebreakers die deze opzoekfuncties gebruiken.
