---
title: "Scoresystemen"
linkTitle: "Scoring"
weight: 3
description: "Hoe partijresultaten een stand worden -- van standaard 1-half-0 tot de iteratieve Keizerrangschikking."
---

## Wat scoring doet

Een scoresysteem zet ruwe partijresultaten om in een numerieke score per
speler en rangschikt spelers vervolgens op die score om een stand te
produceren. Elk toernooi heeft een scoresysteem nodig, en de keuze van
systeem beïnvloedt hoe de stand eruitziet, hoe remises worden
gewaardeerd en hoe afwezigheden worden bestraft.

chesspairing implementeert drie scoresystemen. Alle drie implementeren
dezelfde `Scorer`-interface, die twee methoden heeft:

- **`Score()`** -- neemt de volledige toernooisituatie en retourneert een
  gerangschikte lijst met spelerscores.
- **`PointsForResult()`** -- retourneert de punten die een specifiek
  resultaat waard is in een bepaalde context (handig om puntenwaarden
  aan spelers te tonen).

## Standaard scoring (1-half-0)

Het systeem dat door de FIDE wordt gebruikt voor vrijwel alle geratte
evenementen. Elk resultaat levert een vast aantal punten op:

| Resultaat                  | Standaardpunten |
| -------------------------- | --------------- |
| Winst                      | 1.0             |
| Remise                     | 0.5             |
| Verlies                    | 0.0             |
| PAB (indelingsvrij)        | 1.0             |
| Halve-punt bye             | 0.5             |
| Nulpunten-bye              | 0.0             |
| Forfaitwinst               | 1.0             |
| Forfaitverlies             | 0.0             |
| Afwezig (ongeoorloofd)     | 0.0             |
| Verontschuldigde afwezigheid | 0.0           |
| Clubverplichting           | 0.0             |

Al deze waarden zijn configureerbaar via de Options-struct. Sommige
organisatoren kennen bijvoorbeeld 0,5 toe voor een PAB in plaats van
1.0, of bestraffen ongeoorloofde afwezigheid met een negatieve score.

Standaard scoring is eenvoudig en voorspelbaar: je punten hangen alleen
af van je resultaten, niet van tegen wie je speelde. Dit maakt het ook
het scoresysteem dat intern door Zwitserse indelingen wordt gebruikt voor
het vormen van scoregroepen, zelfs wanneer de openbare toernooistand een
ander systeem hanteert.

Zie [Standaard scoring referentie](/docs/scoring/standard/) voor de
volledige optieslijst.

## Keizerscoring

Keizer is een iteratief scoresysteem dat populair is bij clubtoernooien,
met name in België en Nederland. Het centrale idee: een sterke
tegenstander verslaan levert meer op dan een zwakke.

### Hoe het werkt

1. **Waardenummers.** Elke speler krijgt een waardenummer op basis van
   de huidige ranking. De hoogst gerankte speler krijgt het hoogste
   waardenummer (standaard gelijk aan het aantal spelers); elke
   volgende rang krijgt er één minder.

2. **Partijpunten.** Als je een tegenstander verslaat, ontvang je diens
   waardenummer als punten. Remise levert de helft van het waardenummer
   op. Verlies levert nul op (standaard, hoewel een "toughness bonus"-
   variant een fractie toekent bij verlies).

3. **Niet-partijpunten.** Byes, afwezigheden en clubverplichtingen
   leveren een fractie van je eigen waardenummer op in plaats van dat
   van een tegenstander. Een PAB kan je bijvoorbeeld 50% van je eigen
   waardenummer opleveren.

4. **Zelfoverwinning.** Standaard wordt het eigen waardenummer van elke
   speler eenmaal bij het totaal opgeteld (niet per ronde). Dit beloont
   deelname en creëert scheiding tussen actieve en inactieve spelers.

5. **Iteratieve convergentie.** Hier zit de kern: omdat waardenummers
   afhangen van de ranglijst, en de ranglijst van de scores, en de
   scores van de waardenummers, is het systeem circulair. Keizer lost
   dit op door te itereren: bereken scores, herrangschik, herbereken
   scores met de nieuwe waardenummers, en herhaal tot de ranglijst
   stabiliseert. In de praktijk vindt convergentie binnen enkele
   iteraties plaats.

Keizerscoring heeft veel configureerbare opties: afwezigheidslimieten,
afwezigheidsverval, vaste-waarde-overschrijvingen voor byes,
clubverplichtingsfracties en diverse variantpresets (KeizerForClubs,
Classic KNSB, FreeKeizer).

Zie [Keizerscoring referentie](/docs/scoring/keizer/) voor de volledige
optieslijst en variantpresets.

## Voetbalscoring (3-1-0)

Voetbalscoring gebruikt het bekende voetbalpuntensysteem: 3 punten voor
winst, 1 voor remise, 0 voor verlies. Dit beloont beslissende resultaten
zwaarder dan standaard scoring -- een winst is drie remises waard in
plaats van twee.

| Resultaat                    | Standaardpunten |
| ---------------------------- | --------------- |
| Winst                        | 3.0             |
| Remise                       | 1.0             |
| Verlies                      | 0.0             |
| PAB                          | 3.0             |
| Forfaitwinst                 | 3.0             |
| Forfaitverlies               | 0.0             |
| Afwezig                      | 0.0             |
| Verontschuldigde afwezigheid | 0.0             |
| Clubverplichting             | 0.0             |

Voetbalscoring is geïmplementeerd als een dunne wrapper rond standaard
scoring met andere standaardwaarden. Alle puntenwaarden blijven
configureerbaar.

Zie [Voetbalscoring referentie](/docs/scoring/football/) voor details.

## Byes, forfaits en afwezigheden

Alle drie de scoresystemen verwerken speciale resultaattypen:

- **Indelingsvrij (PAB):** Het systeem kent een bye toe wanneer het
  aantal spelers oneven is. Royaal gescoord (een vol punt bij standaard,
  een fractie van het eigen waardenummer bij Keizer).
- **Halve-punt bye / nulpunten-bye:** Aangevraagde byes die verminderde
  of geen punten opleveren.
- **Forfaitwinst:** De winnaar krijgt punten, maar de partij wordt
  uitgesloten van de indelingshistorie (de spelers kunnen in een latere
  ronde opnieuw worden ingedeeld).
- **Dubbel forfait:** Geen van beide spelers krijgt punten. De partij
  wordt uitgesloten van zowel scoring als indelingshistorie -- hij wordt
  behandeld alsof hij nooit heeft plaatsgevonden.
- **Afwezig:** Een speler die noch heeft gespeeld noch een bye heeft
  ontvangen. Doorgaans gescoord als nul, hoewel Keizer een
  configureerbare fractie toekent.

## Scoring is onafhankelijk van de indeling

Een belangrijke ontwerpbeslissing in chesspairing: **scoring en indeling
zijn volledig onafhankelijk**. Elk scoresysteem werkt met elk
indelingssysteem. Je kunt een Zwitsers toernooi draaien met Keizerscoring,
of een round-robin met voetbalscoring. De indeling en scorer hoeven niets
van elkaar te weten -- ze opereren allebei op dezelfde `TournamentState`
en produceren onafhankelijke uitvoer.

De enige uitzondering is de Keizer-_indeling_, die intern de
Keizer-_scorer_ gebruikt om spelers te rangschikken voor de indeling. Maar
zelfs daar kan de openbare toernooistand een ander scoresysteem
gebruiken.

## Verder lezen

- [Standaard scoring](/docs/scoring/standard/)
- [Keizerscoring](/docs/scoring/keizer/)
- [Voetbalscoring](/docs/scoring/football/)
