---
title: "Round-robin toernooien"
linkTitle: "Round-Robin"
weight: 2
description: "Elke speler ontmoet elke andere speler — planning met Berger-tabellen en afhandeling van oneven aantallen."
---

## Iedereen speelt tegen iedereen

In een round-robin toernooi speelt elke speler precies één keer tegen
elke andere speler. Er is geen matching-algoritme, geen scoregroepen en
geen floaters -- het volledige schema staat vast vóór de eerste zet. Een
toernooi met N spelers heeft N-1 ronden nodig (of N ronden bij een
oneven N, omdat er dan elke ronde één speler overslaat).

Round-robin is de gouden standaard voor het bepalen van de sterkste
speler wanneer het deelnemersveld klein genoeg is. Het is het formaat
bij uitstek voor Kandidatentoernooien, nationale kampioenschappen en
gesloten invitationals waar het aantal deelnemers beheersbaar is.

## Enkel en dubbel round-robin

Een **enkel round-robin** (één cyclus) geeft elk paar spelers één
partij. De Cycles-optie bepaalt hoe vaak het volledige schema wordt
herhaald:

- **Cycles: 1** -- enkel round-robin. N-1 ronden voor N spelers.
- **Cycles: 2** -- dubbel round-robin. Elk paar speelt twee keer met
  omgekeerde kleuren. 2(N-1) ronden in totaal.

Dubbel round-robin is gebruikelijk wanneer het deelnemersveld erg klein
is (4-8 spelers) en de organisator meer beslissende resultaten wil. In
de tweede cyclus worden de kleuren omgedraaid zodat elke speler één
partij als wit en één als zwart speelt tegen elke tegenstander.

## Berger-tabellen

Het schema wordt opgebouwd met FIDE Berger-tabellen (C.05 Annex 1). Het
algoritme werkt als volgt:

1. Houd de laatste speler (speler N) op een vaste positie.
2. Roteer de overige N-1 spelers door de andere posities met een
   vaste stap.
3. Elke rotatie levert één ronde van indelingen op: de speler op
   positie 0 speelt tegen de speler op positie N-1, positie 1 speelt
   tegen positie N-2, enzovoort.

Dit levert een schema op waarin elke speler precies één keer elke andere
speler ontmoet, en kleurtoewijzingen volgen rechtstreeks uit de
tabelposities. Bord 1 (het bord van de vaste speler) wisselt elke ronde
van kleur; de andere borden wijzen wit toe aan de speler met de lagere
positie-index.

Zie [Berger-tabellen](/docs/algorithms/berger-tables/) voor een
diepgaandere kijk op hoe de tabellen worden opgebouwd.

## Oneven spelersaantallen

Bij een oneven spelersaantal wordt een virtuele "dummy"-positie
toegevoegd om het aantal even te maken. Elke ronde ontvangt de speler
die tegen de dummy is ingedeeld een pairing-allocated bye (PAB) in
plaats van te spelen. De Berger-tabelrotatie zorgt ervoor dat elke
speler precies één bye krijgt gedurende de cyclus.

## Kleurbalancering

Binnen een enkele cyclus verdeelt de structuur van de Berger-tabel
kleuren van nature eerlijk. Bij een dubbel round-robin draait de tweede
cyclus alle kleurtoewijzingen om (wanneer de ColorBalance-optie is
ingeschakeld, wat standaard het geval is), zodat elke speler één partij
als wit en één als zwart speelt tegen elke tegenstander.

### Wisseling van de laatste twee ronden

Bij een dubbel round-robin kan de overgang tussen cyclus 1 en cyclus 2
ertoe leiden dat een speler drie keer achtereen dezelfde kleur krijgt
(laatste ronde van cyclus 1, plus de eerste twee ronden van cyclus 2
met omgekeerde kleuren). Om dit te voorkomen wisselt chesspairing
standaard de laatste twee ronden van de eerste cyclus om (de
SwapLastTwoRounds-optie). Dit is de standaard FIDE-aanbeveling voor
dubbel round-robin evenementen.

## Varma-tabellen

Wanneer een round-robin spelers uit veel federaties bevat, beïnvloedt
de initiële spelernummering wie tegen wie speelt in welke ronde. FIDE's
Varma-tabellen (C.05 Annex 2) bieden een federatie-bewuste methode voor
het toewijzen van rangnummers, zodat spelers van dezelfde federatie
elkaar zo laat mogelijk in het toernooi treffen en de "interne"
ontmoetingen worden gespreid.

chesspairing implementeert Varma-tabeltoewijzing via het `algorithm/varma`-pakket.
Zie [Varma-tabellen](/docs/algorithms/varma-tables/) voor details.

## Verder lezen

- [Round-Robin indelingssysteem referentie](/docs/pairing-systems/round-robin/)
- [Berger-tabellen algoritme](/docs/algorithms/berger-tables/)
- [Varma-tabellen algoritme](/docs/algorithms/varma-tables/)
