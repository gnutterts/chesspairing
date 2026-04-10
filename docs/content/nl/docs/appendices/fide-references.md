---
title: "FIDE Handbook-referenties"
linkTitle: "FIDE-referenties"
weight: 1
description: "Verwijzingen naar relevante FIDE Handbook-secties per indelings- en scoringssysteem."
---

Op deze pagina staan de FIDE Handbook-secties die betrekking hebben op elk indelingssysteem, elke scoringsregel en elke tiebreaker die in de chesspairing-module is geïmplementeerd. Links gebruiken het formaat `https://handbook.fide.com/chapter/CXXXX` waar van toepassing, maar exacte URL's kunnen in de loop der tijd veranderen. De sectienummers worden vermeld zodat je ze handmatig kunt opzoeken.

## Indelingssystemen

| Systeem           | FIDE-sectie  | Handbook-hoofdstuk                                        |
| ----------------- | ------------ | --------------------------------------------------------- |
| Zwitsers (Dutch)  | C.04.3       | Algemene regels, Dutch-systeem                            |
| Zwitsers Burstein | C.04.4.2     | Systemen gebaseerd op het Dutch-systeem, Burstein-systeem |
| Zwitsers Dubov    | C.04.4.1     | Systemen gebaseerd op het Dutch-systeem, Dubov-systeem    |
| Zwitsers Lim      | C.04.4.3     | Systemen gebaseerd op het Dutch-systeem, Lim-systeem      |
| Double-Swiss      | C.04.5       | Double-Swiss-systeem                                      |
| Team Zwitsers     | C.04.6       | Team Zwitsers systeem                                     |
| Bakoe-versnelling | C.04.7       | Versnelde indelingen voor het Zwitsers systeem            |
| Round-robin       | C.05         | Round-robin-systeem                                       |
| Berger-tabellen   | C.05 Annex 1 | Indelingstabellen voor round-robin-toernooien             |
| Varma-tabellen    | C.05 Annex 2 | Initiële nummertoewijzing voor round-robin-toernooien     |

### Zwitsers Dutch (C.04.3)

Het Dutch-systeem is de meest gebruikte Zwitserse indelingsmethode. Het definieert absolute criteria (C1--C4) waaraan voldaan moet worden, en optimalisatiecriteria (C5--C21) die het algoritme maximaliseert. De implementatie in chesspairing gebruikt een globale Blossom-matchingarchitectuur met een 7-fasen bracket-lus.

Link: [https://handbook.fide.com/chapter/C0403](https://handbook.fide.com/chapter/C0403)

### Zwitsers Burstein (C.04.4.2)

Het Burstein-systeem is een Dutch-variant die onderscheid maakt tussen seedingronden en post-seedingronden. In post-seedingronden worden spelers opnieuw gerangschikt op basis van een oppositie-index afgeleid van Buchholz- en Sonneborn-Berger-waarden.

Link: [https://handbook.fide.com/chapter/C04042](https://handbook.fide.com/chapter/C04042)

### Zwitsers Dubov (C.04.4.1)

Het Dubov-systeem gebruikt Average Rating of Opponents (ARO) voor de sortering binnen scoregroepen en definieert eigen criteria (C1--C10). Het gebruikt oplopende ARO-sortering en een transpositiegebaseerde matchingaanpak.

Link: [https://handbook.fide.com/chapter/C04041](https://handbook.fide.com/chapter/C04041)

### Zwitsers Lim (C.04.4.3)

Het Lim-systeem verwerkt scoregroepen in mediaan-eerst-volgorde en gebruikt uitwisselingsgebaseerde matching. Het classificeert floaters in vier typen (A--D) en definieert specifieke compatibiliteitsvoorwaarden.

Link: [https://handbook.fide.com/chapter/C04043](https://handbook.fide.com/chapter/C04043)

### Double-Swiss (C.04.5)

Het Double-Swiss-systeem gebruikt lexicografische bracket-indeling en een 5-staps kleurtoewijzingsprioriteit. Het is ontworpen voor toernooien waarin spelers twee partijen per ronde spelen tegen verschillende tegenstanders.

Link: [https://handbook.fide.com/chapter/C0405](https://handbook.fide.com/chapter/C0405)

### Team Zwitsers (C.04.6)

Het Team Zwitsers systeem past Zwitserse indeling aan voor teamcompetities. Het ondersteunt configureerbare kleurvoorkeurtypen (A, B of Geen) en gebruikt een 9-staps kleurtoewijzingsproces.

Link: [https://handbook.fide.com/chapter/C0406](https://handbook.fide.com/chapter/C0406)

### Bakoe-versnelling (C.04.7)

Bakoe-versnelling kent virtuele punten toe in de eerste ronden om topgeplaatste spelers te scheiden, waardoor het aantal beslissende partijen tussen topspelers in de openingsronden afneemt. De chesspairing-module implementeert dit als optie voor de Dutch- en Burstein-indelingen.

Link: [https://handbook.fide.com/chapter/C0407](https://handbook.fide.com/chapter/C0407)

### Round-robin (C.05)

Het round-robin-systeem deelt elke speler in tegen elke andere speler. De chesspairing-module implementeert FIDE Berger-tabellen voor de planning en ondersteunt enkelvoudige en dubbele round-robin-formaten met configureerbare kleurbalancering.

Link: [https://handbook.fide.com/chapter/C05](https://handbook.fide.com/chapter/C05)

### Berger-tabellen (C.05 Annex 1)

Berger-tabellen definiëren het standaard indelingsschema voor round-robin-toernooien. De tabellen specificeren welke spelers in elke ronde tegen elkaar spelen en welke speler wit heeft.

### Varma-tabellen (C.05 Annex 2)

Varma-tabellen bieden een federatiebewuste initiële nummertoewijzing voor round-robin-toernooien. Dit zorgt ervoor dat spelers van dezelfde federatie zo gelijkmatig mogelijk over de ronden worden verdeeld.

## Scoring en tiebreakers

| Onderwerp               | FIDE-sectie | Omschrijving                                                         |
| ----------------------- | ----------- | -------------------------------------------------------------------- |
| Algemene scoringsregels | C.02        | Definieert de standaard punttoekenning voor winst, remise en verlies |
| Tiebreak-procedures     | B.02        | Definieert goedgekeurde tiebreaker-methoden en hun berekening        |

### Standaard scoring (C.02)

FIDE C.02 behandelt de algemene regels voor scoring in schaakcompetities. De chesspairing-module implementeert standaard scoring (1--0,5--0) met configureerbare puntwaarden voor winst, remise, verlies, bye, forfait en afwezigheid.

### Tiebreakers (B.02)

FIDE B.02 definieert goedgekeurde tiebreak-procedures. De chesspairing-module implementeert 25 tiebreakers, waaronder alle gangbare FIDE-methoden. Zie de [Tiebreakers](/docs/tiebreakers/)-documentatie voor de volledige lijst.

Link: [https://handbook.fide.com/chapter/B02](https://handbook.fide.com/chapter/B02)

## Gerelateerde reglementen

- **C.01 -- Algemene regels voor competities**: Omvat de overkoepelende regels die gelden voor alle FIDE-gewaarmerkte competities, inclusief definities en algemene procedures.
- **C.02 -- Scoring**: Definieert hoe partijresultaten worden omgezet in punten.
- **FIDE-ratingreglementen**: Los van de indelings- en scoringsregels regelen deze hoe individuele speelsterkte wordt berekend en bijgewerkt. De chesspairing-module berekent geen ratings, maar gebruikt ze als invoer voor indeling en tiebreaking.

## FIDE-reglementen vinden

Het FIDE Handbook is beschikbaar op [https://handbook.fide.com](https://handbook.fide.com). Reglementen zijn georganiseerd per hoofdstuknummer. Als een directe link hierboven niet werkt, ga dan naar het handbook en zoek op sectienummer (bijv. C.04.3).
