---
title: "Byes"
linkTitle: "Byes"
weight: 7
description: "Indelings-byes, halve-punt byes, en hoe oneven spelersaantallen worden afgehandeld."
---

Een **bye** is een ronde waarin een speler geen tegenstander heeft. Byes ontstaan om verschillende redenen -- een oneven aantal deelnemers, een speler die vrij vraagt, of een speler die simpelweg niet komt opdagen -- en elke reden heeft andere puntconsequenties.

## Bye-types

chesspairing implementeert zes bye-types, elk geïdentificeerd door een code in TRF16-toernooibestanden:

| Bye-type                        | TRF-code | Standaardpunten | Beschrijving                                                                          |
| ------------------------------- | -------- | --------------- | ------------------------------------------------------------------------------------- |
| **PAB** (Pairing-Allocated Bye) | `F`      | 1.0             | Automatisch toegekend bij een oneven aantal actieve spelers.                          |
| **Halve-punt bye**              | `H`      | 0.5             | Vooraf aangevraagd door de speler. De speler slaat een ronde over voor een half punt. |
| **Nulpunten-bye**               | `Z`      | 0.0             | Aangevraagd door de speler. Geen punten.                                              |
| **Afwezig**                     | `U`      | 0.0             | De speler is niet komen opdagen en heeft de arbiter niet vooraf ingelicht.            |
| **Verontschuldigd**             | --       | 0.0             | De speler heeft de arbiter vooraf laten weten de ronde te missen.                     |
| **Clubverplichting**            | --       | 0.0             | De speler is afwezig vanwege interclub-teamplicht.                                    |

De getoonde puntwaarden zijn standaardwaarden voor [standaard scoring](/docs/scoring/). Elk scoresysteem kan deze waarden via opties anders configureren.

## De Pairing-Allocated Bye (PAB)

Het belangrijkste bye-type is de PAB. Als een toernooi een oneven aantal actieve spelers heeft, moet er elke ronde één speler overslaan. De PAB is standaard een vol punt waard, als compensatie voor de partij die de speler niet kon spelen.

Een fundamentele regel in alle indelingssystemen: **een speler mag niet meer dan één keer een PAB ontvangen** in een toernooi. De engine filtert spelers die er al een hebben gehad voordat de volgende PAB-ontvanger wordt gekozen.

### Hoe PAB-toewijzing werkt

Elk indelingssysteem gebruikt een andere methode om te bepalen wie de PAB ontvangt:

**Dutch en Burstein** -- Deze systemen gebruiken een completability-gebaseerde aanpak. Voordat de eigenlijke indeling begint, test een pre-matching fase (Stage 0.5 genoemd) welke speler, wanneer verwijderd uit de pool, het nog steeds mogelijk maakt om de overige spelers volledig te indelen. Dit garandeert dat de bye gaat naar een speler wiens verwijdering de indeling niet verstoort. Onder de geschikte kandidaten wordt de speler met de laagste score, de meeste gespeelde partijen en de laagste rangorde (hoogste rangnummer) verkozen. Zie [completability](/docs/algorithms/completability/) voor details.

**Dubov** -- De bye gaat naar de laagst gerangschikte speler (hoogste rangnummer) in de laagste scoregroep die nog geen PAB heeft ontvangen. Bij gelijke spelers wordt degene met de meeste gespeelde partijen het eerst geselecteerd.

**Lim** -- De bye wordt toegewezen aan de laagst gerangschikte speler in de laagste scoregroep, mits deze nog geen PAB heeft ontvangen.

**Double-Swiss en Team Swiss** -- Deze lexicografische indelingssystemen wijzen de bye toe aan de speler met de laagste score, met als tiebreak de laagste rangorde (hoogste rangnummer).

**Keizer** -- De laagst gerangschikte speler (op basis van de huidige Keizerscore, of op basis van rating als er nog geen ronden zijn gespeeld) ontvangt de bye.

**Round-Robin** -- Oneven spelersaantallen worden afgehandeld door een virtuele "dummy"-speler aan de rotatie toe te voegen. Elke ronde ontvangt de echte speler die tegen de dummy is ingedeeld de bye. Dit roteert vanzelf door de Berger-tabel, zodat iedere speler precies één bye krijgt gedurende de cyclus.

## Bye-scoring

Hoeveel punten een bye waard is, hangt af van het gebruikte [scoresysteem](/docs/scoring/):

- **Standaard scoring**: PAB = 1.0, halve-punt bye = 0.5, alle overige = 0.0 standaard. Elke waarde is configureerbaar via de opties `pointBye`, `pointAbsent` en gerelateerde instellingen.
- **Football scoring**: volgt dezelfde standaardwaarden als standaard scoring maar dan op de voetbalpuntenschaal (winst = 3, remise = 1, verlies = 0).
- **Keizers scoring**: byes worden gescoord met instelbare fracties van het eigen waardenummer van de speler, met aparte instellingen voor PAB, halve-punt byes en afwezigheden.

## Byes en tiebreakers

Bye-ronden zijn geen "echte" partijen. Omdat er geen tegenstander is:

- Tegenstander-gebaseerde tiebreakers (Buchholz, Sonneborn-Berger, ARO) hebben geen tegenstander-data voor die ronde. De tiebreaker-implementaties lossen dit op door alleen over daadwerkelijke partij-entries te sommeren.
- De bye-ronde zelf telt niet mee voor partijen-gespeeld of winsttellingen die door bepaalde tiebreakers worden gebruikt.

Het aantal bye-ronden dat een speler heeft gehad wordt apart bijgehouden en kan invloed hebben op hoe virtuele tegenstanders worden berekend in tiebreakers zoals Fore Buchholz.

## Zie ook

- [Overzicht indelingssystemen](/docs/pairing-systems/) -- hoe elk systeem de PAB-ontvanger selecteert
- [Scoresystemen](/docs/scoring/) -- bye-puntwaarden configureren
- [Completability-algoritme](/docs/algorithms/completability/) -- de Dutch/Burstein-methode voor het vinden van de optimale bye-kandidaat
