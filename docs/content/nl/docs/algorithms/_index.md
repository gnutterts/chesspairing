---
title: "Algoritmen"
linkTitle: "Algoritmen"
weight: 90
description: "Wiskundige verdiepingen in de algoritmen achter toernooi-indelingen — met formules, bewijsschetsen en pseudocode."
---

Deze sectie verkent de wiskunde achter de indelings-, scorings- en
tiebreaker-engines van chesspairing. De pagina's die volgen presenteren
formules, bewijsschetsen, complexiteitsanalyses en pseudocode — samen met
verwijzingen naar de Go-broncode waar elk algoritme is geïmplementeerd.

**Doelgroep.** Onderzoekers, wiskundigen en ontwikkelaars die het _waarom_
achter de code willen begrijpen — niet alleen het API-oppervlak. Als je op
zoek bent naar gebruiksvoorbeelden of configuratieopties, kijk dan bij de
secties [Aan de slag](../getting-started/) en [Formaten](../formats/).

**Reikwijdte en nauwkeurigheid.** Dit zijn bewijsschetsen en uitgewerkte
intuities, geen formele bewijzen. Waar een resultaat algemeen bekend is
(bijv. de LP-relaxatie van maximum weight matching), vermelden we de stelling
en verwijzen naar een leerboekbewijs. Waar de redenering specifiek is voor
deze codebase (bijv. de bitindeling van kantgewichten), geven we een
volledige afleiding.

**FIDE-reglementen.** Reglementstekst in deze sectie is geparafraseerd, niet
letterlijk geciteerd. Raadpleeg voor de gezaghebbende bron het
[FIDE Handboek — Schaakreglementen](https://handbook.fide.com/chapter/C0403).

---

## Pagina's per categorie

### Kernalgoritmen

| Pagina                                | Samenvatting                                                                                                                               |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| [Blossom Matching](blossom/)          | Het algoritme van Edmonds voor maximum weight matching in algemene grafen — $O(n^3)$ met blossom-contractie.                               |
| [Kantgewicht-codering](edge-weights/) | Hoe 16+ indelingscriteria worden ingepakt in een enkel `*big.Int`-kantgewicht zodat Blossom-maximalisatie de criteriaprioriteit respecteert. |
| [Completeerbaarheid](completability/) | Stage 0.5 pre-matching die de bye-ontvanger bepaalt voordat de echte matching begint.                                                      |

### Indelingssysteemspecifiek

| Pagina                                   | Samenvatting                                                                                                                                     |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| [Bergertabellen](berger-tables/)         | FIDE round-robin-rotatie (C.05 Annex 1) — het $n-1$ rondenschema opbouwen uit Bergers tabel van 1895.                                            |
| [Varma-tabellen](varma-tables/)          | Federatiebewuste toewijzing van rangnummers voor round-robin-toernooien (C.05 Annex 2).                                                       |
| [Baku-acceleratie](baku-acceleration/)   | Virtuele punten in vroege ronden (C.04.7) — het aantal remises onder topgeplaatsten verminderen door aanvankelijke scoregroepen op te blazen.    |
| [Nederlandse criteria](dutch-criteria/)  | De 21 optimalisatiecriteria van het Nederlandse systeem (C.04.3) — van absolute beperkingen $C_1$--$C_4$ tot kwaliteitscriteria $C_8$--$C_{21}$. |
| [Dubov-criteria](dubov-criteria/)        | De 10 criteria van het Dubov-systeem (C.04.4.1) met MaxT-tracking en oplopende ARO-sortering.                                                    |
| [Lim Exchange Matching](lim-exchange/)   | Exchange-gebaseerde matching (C.04.4.3) met vier floater-typen (A--D) en mediaan-tiebreaking.                                                    |
| [Lexicografische indeling](lexicographic/) | DFS-backtracking over criteriafuncties, gedeeld door Double-Swiss (C.04.5) en Team Swiss (C.04.6).                                               |

### Scoring en tiebreaking

| Pagina                                     | Samenvatting                                                                                                  |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------- |
| [Keizer-convergentie](keizer-convergence/) | Iteratieve scoring met oscillatiedetectie — bewijsschets voor de convergentie van de vaste-puntiteratie.      |
| [Elo-model](elo-model/)                    | De verwachte-scorefunctie $E = \frac{1}{1 + 10^{-d/400}}$ en het gebruik ervan in prestatieratingtiebreakers. |
| [FIDE B.02-tabel](fide-b02/)               | De ratingverschil $\leftrightarrow$ verwachte score opzoektabel, inclusief interpolatie en randgevallen.      |
| [Kleurverdeling](color-allocation/)        | Zes kleurverdelingsalgoritmen vergeleken: Nederlands, Burstein, Dubov, Lim, Double-Swiss en Team Swiss.       |
