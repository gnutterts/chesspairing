---
title: "Elo-waarschijnlijkheidsmodel"
linkTitle: "Elo-model"
weight: 8
description: "De verwachte-scorefunctie en haar rol in prestatieratingberekeningen."
---

## Het logistisch model

Het Elo-ratingsysteem modelleert de waarschijnlijkheid van een partijuitslag
als functie van het ratingverschil tussen twee spelers. Gegeven speler $A$ met
rating $R_A$ en speler $B$ met rating $R_B$, is de **verwachte score** voor
speler $A$:

$$E_A = \frac{1}{1 + 10^{-(R_A - R_B) / 400}}$$

Dit is een logistische functie met grondtal 10 en een schaalfactor van 400. De
belangrijkste eigenschappen:

- $E_A = 0.5$ wanneer $R_A = R_B$ (gelijke ratings impliceren gelijke kansen).
- $E_A \to 1$ als $(R_A - R_B) \to +\infty$.
- $E_A \to 0$ als $(R_A - R_B) \to -\infty$.
- $E_A + E_B = 1$ (de scores zijn complementair).

---

## Ratingverschil en winstkans

Met $d = R_A - R_B$ vereenvoudigt de verwachte score tot:

$$E(d) = \frac{1}{1 + 10^{-d/400}}$$

Enkele referentiewaarden:

| $d$ (ratingverschil) | $E(d)$ (verwachte score) | Interpretatie       |
| -------------------- | ------------------------ | ------------------- |
| 0                    | 0.50                     | Gelijke kansen      |
| 100                  | 0.64                     | Lichte favoriet     |
| 200                  | 0.76                     | Duidelijke favoriet |
| 400                  | 0.91                     | Sterke favoriet     |
| 800                  | 0.99                     | Vrijwel zeker       |
| $-100$               | 0.36                     | Lichte underdog     |
| $-200$               | 0.24                     | Duidelijke underdog |

De functie is monotoon stijgend: een groter ratingvoordeel betekent altijd een
hogere verwachte score. Ze is ook symmetrisch:

$$E(-d) = 1 - E(d)$$

---

## Inverse functie: ratingverschil uit score

De inverse van de verwachte-scorefunctie geeft het ratingverschil dat
overeenkomt met een waargenomen winstpercentage $p$:

$$d(p) = -400 \cdot \log_{10}\!\left(\frac{1}{p} - 1\right)$$

Dit wordt gebruikt door de [FIDE B.02-tabel](../fide-b02/) om
toernooi-prestatie (fractionele score) om te rekenen naar een ratingverschil.
Een speler die bijvoorbeeld 7/10 = 0,70 scoort tegen zijn tegenstanders heeft:

$$d(0.70) = -400 \cdot \log_{10}\!\left(\frac{1}{0.70} - 1\right) = -400 \cdot \log_{10}(0.4286) \approx 147$$

Dit betekent dat de speler ongeveer 147 ratingpunten boven de gemiddelde
rating van zijn tegenstanders presteerde.

### Randgedrag

De inverse functie heeft singulariteiten bij $p = 0$ en $p = 1$:

- $d(0) = -\infty$: een speler die elke partij verliest presteert oneindig
  onder zijn tegenstanders.
- $d(1) = +\infty$: een speler die elke partij wint presteert oneindig boven
  zijn tegenstanders.

FIDE behandelt dit door afkapping: $d$ wordt begrensd tot $[-800, +800]$ in de
B.02-tabel. Een speler met een perfecte score krijgt $d = +800$; een speler
met nul punten krijgt $d = -800$.

---

## Gebruik bij de prestatierating (TPR)

De **Tournament Performance Rating** (FIDE Art. 10.2) schat de rating die een
speler nodig zou hebben om zijn waargenomen resultaten tegen zijn specifieke
tegenstanders te behalen:

$$\text{TPR} = \text{ARO} + d(p)$$

waarbij:

- $\text{ARO} = \frac{1}{n} \sum_{i=1}^{n} R_i$ de Average Rating of
  Opponents is (berekend over daadwerkelijke partijen, exclusief forfaits en
  byes).
- $p = \frac{\text{score}}{n}$ de fractionele score is (winst telt als 1,
  remise als 0,5).
- $d(p)$ wordt opgezocht uit de [FIDE B.02-tabel](../fide-b02/) (niet direct
  berekend uit de formule, voor consistentie met FIDE's officiële
  getabelleerde waarden).

Het resultaat wordt afgerond op het dichtstbijzijnde gehele getal.

---

## Gebruik bij prestatiepunten (PTP)

De **Performance Points** tiebreaker (FIDE Art. 10.3) werkt omgekeerd. In
plaats van een rating te berekenen uit een score, zoekt het de minimale rating
$R^*$ zodanig dat de verwachte totaalscore tegen de daadwerkelijke
tegenstanders de werkelijke score evenaart of overtreft:

$$R^* = \min \left\{ R \;\middle|\; \sum_{i=1}^{n} E(R - R_i) \geq S \right\}$$

waarbij $R_i$ de tegenstanderratings zijn en $S$ de werkelijke score.

Dit wordt opgelost door **binair zoeken** over het ratingbereik
$[\min_i R_i - 800, \max_i R_i + 800]$ met een precisie van 0,5 ratingpunt.
De verwachte-scorefunctie $E(d)$ wordt bij elke stap geëvalueerd met behulp
van de [FIDE B.02-tabel](../fide-b02/).

Speciale gevallen:

| Voorwaarde               | Resultaat                |
| ------------------------ | ------------------------ |
| $S = 0$ (nul punten)     | $R^* = \min_i R_i - 800$ |
| $S = n$ (perfecte score) | $R^* = \max_i R_i + 800$ |
| Geen partijen            | $R^* = 0$                |

---

## Waarom grondtal 10 en factor 400?

Het oorspronkelijke Elo-systeem (jaren 60) gebruikte een
normaalverdelingsbendering. FIDE heeft later het logistisch model met grondtal
10 en schaalfactor 400 overgenomen omdat:

1. **400 punten verschil** overeenkomt met een oddsverhouding van
   ongeveer 10:1 ($E(400) \approx 0.91$), wat een intuïtieve schaal biedt
   voor schakers.
2. **Grondtal 10** de rekenkunde eenvoudig houdt: een verschil van 400 punten
   betekent een factor $10^1 = 10$ in de odds.
3. De logistische functie eenvoudiger te berekenen is en zwaardere staarten
   heeft dan de normaalverdeling, wat beter aansluit bij waargenomen
   schaakresultaten waar verrassingen vaker voorkomen dan een normaal model
   voorspelt.

---

## Relatie met andere modellen

| Model                 | Formule                      | Gebruikt door                                              |
| --------------------- | ---------------------------- | ---------------------------------------------------------- |
| FIDE Elo (logistisch) | $E = (1 + 10^{-d/400})^{-1}$ | FIDE-ratings, deze implementatie                           |
| USCF Elo (logistisch) | $E = (1 + 10^{-d/400})^{-1}$ | Dezelfde formule, andere K-factoren                        |
| Glicko                | Logistisch met onzekerheid   | Glicko-2-ratingsysteem                                     |
| Bradley-Terry         | $E = e^{d} / (1 + e^{d})$    | Algemene paarsgewijze vergelijking (natuurlijke logaritme) |

Het FIDE logistisch model is een speciaal geval van het Bradley-Terry-model
met de substitutie $d_{\text{BT}} = \frac{d \cdot \ln(10)}{400}$.

---

## Gerelateerde pagina's

- [FIDE B.02-conversietabel](../fide-b02/) -- de gediscretiseerde opzoektabel
  die wordt gebruikt in plaats van de continue formule.
- [Prestatiegerating tiebreakers](/docs/tiebreakers/performance/) -- TPR- en
  PTP-tiebreakers die het Elo-model gebruiken.
