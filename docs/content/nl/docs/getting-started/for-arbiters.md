---
title: "Voor Arbiters"
linkTitle: "Voor Arbiters"
weight: 4
description: "Een praktische gids voor schaakarbiters die willen begrijpen hoe chesspairing de FIDE-reglementen implementeert."
---

Chesspairing is een indelingsengine voor schaaktoernooien. Op basis van een spelerslijst en de behaalde resultaten produceert het de indeling voor de volgende ronde, berekent het de stand en lost het gelijke scores op -- allemaal volgens de FIDE-reglementen die u al kent.

Deze pagina koppelt de belangrijkste aandachtspunten van een arbiter aan de relevante onderdelen van de documentatie. U hoeft geen programmeur te zijn om het te volgen, al is de [CLI Snelstart](../cli-quickstart/) de snelste manier om de tool in actie te zien.

## Wat het doet

Chesspairing voert drie taken uit:

1. **Indeling** -- bepalen wie tegen wie speelt in de volgende ronde, inclusief de toewijzing van een bye bij een oneven aantal spelers.
2. **Scoren** -- spelresultaten omzetten in punttotalen, met instelbare puntwaarden voor winst, remise, byes, forfait en afwezigheid.
3. **Tiebreaking** -- tiebreakwaarden berekenen om een definitieve ranglijst op te stellen wanneer spelers gelijk staan in score.

Deze drie taken zijn onafhankelijk van elkaar. U kunt een toernooi indelen met het Zwitsers systeem (Dutch) en scoren met Keizer-punten, of een round-robin spelen met standaard 1-half-0 scoring. De engine dwingt geen vaste combinatie af.

## Ondersteunde FIDE-indelingssystemen

Chesspairing implementeert alle huidige FIDE-goedgekeurde Zwitserse indelingssystemen, plus round-robin en Keizer:

| Systeem      | FIDE-reglement                 | Documentatie                                        |
| ------------ | ------------------------------ | --------------------------------------------------- |
| Dutch        | C.04.3                         | [Dutch](/docs/pairing-systems/dutch/)               |
| Burstein     | C.04.4.2                       | [Burstein](/docs/pairing-systems/burstein/)         |
| Dubov        | C.04.4.1                       | [Dubov](/docs/pairing-systems/dubov/)               |
| Lim          | C.04.4.3                       | [Lim](/docs/pairing-systems/lim/)                   |
| Double-Swiss | C.04.5                         | [Double-Swiss](/docs/pairing-systems/double-swiss/) |
| Team Swiss   | C.04.6                         | [Team Swiss](/docs/pairing-systems/team/)           |
| Round-Robin  | C.05 (Bergertabellen, Annex 1) | [Round-Robin](/docs/pairing-systems/round-robin/)   |
| Keizer       | (niet FIDE-gereguleerd)        | [Keizer](/docs/pairing-systems/keizer/)             |

Elk indelingssysteem heeft een eigen pagina in het onderdeel [Indelingssystemen](/docs/pairing-systems/), met uitleg over hoe de engine de regelementcriteria toepast, randgevallen afhandelt en kleuren toewijst.

### Drop-in vervanging

De CLI kan als drop-in vervanging voor **bbpPairings** en **JaVaFo** fungeren. Als u momenteel een van beide tools gebruikt, kunt u overstappen op chesspairing zonder uw werkwijze aan te passen. Zie [Legacy-modus](/docs/cli/legacy/) voor details.

## Praktische werkwijze

Een typische ronde verloopt in drie stappen:

1. **Maak een TRF-bestand aan.** Het FIDE Tournament Report File (TRF16) is het standaard uitwisselingsformaat voor toernooigegevens. Uw toernooibeheersoftware kan het waarschijnlijk exporteren. Als u het formaat wilt begrijpen, zie [TRF16](/docs/formats/trf16/).

2. **Voer de indeling uit.** Geef het TRF-bestand aan de CLI:

   ```
   chesspairing pair --system dutch tournament.trf
   ```

   De engine leest de spelerslijst en alle voorgaande ronden, en geeft vervolgens de indeling voor de volgende ronde uit. Er zijn meerdere [uitvoerformaten](/docs/cli/output-formats/) beschikbaar (tabel, JSON, XML, bordweergave).

3. **Bekijk de indeling en de stand.** De uitvoer toont elk bord met de wit- en zwartspeler. Voor de stand:
   ```
   chesspairing standings tournament.trf
   ```
   Dit berekent de scores en past tiebreakers toe om een gerangschikte stand te produceren.

Zie de [CLI Snelstart](../cli-quickstart/) voor een stapsgewijze uitleg.

## Veelgestelde vragen van arbiters

De documentatie is zo opgezet dat u antwoorden kunt vinden op de vragen die tijdens een toernooi opkomen:

| Vraag                                                                  | Waar te vinden                                                                      |
| ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Hoe wordt een bye toegewezen bij een oneven aantal spelers?            | [Concepten: Byes](/docs/concepts/byes/)                                             |
| Waarom is speler X tegen speler Y ingedeeld?                           | De pagina van uw indelingssysteem onder [Indelingssystemen](/docs/pairing-systems/) |
| Wat gebeurt er als een speler forfait krijgt?                          | [Concepten: Forfait](/docs/concepts/forfeits/)                                      |
| Hoe worden kleuren toegewezen?                                         | [Concepten: Kleuren](/docs/concepts/colors/)                                        |
| Wat is een floater en waarom is dat van belang?                        | [Concepten: Floaters](/docs/concepts/floaters/)                                     |
| Hoe wordt de stand berekend?                                           | [Scoresystemen](/docs/scoring/)                                                     |
| Welke tiebreakers zijn beschikbaar?                                    | Zie hieronder, en het onderdeel [Tiebreakers](/docs/tiebreakers/)                   |
| Hoe controleer ik of de engine correcte indelingen heeft geproduceerd? | [CLI: check](/docs/cli/check/)                                                      |
| Hoe valideer ik mijn TRF-bestand?                                      | [CLI: validate](/docs/cli/validate/)                                                |

## Beschikbare tiebreakers

Chesspairing biedt 25 tiebreakers. De onderstaande tabel groepeert ze per categorie, volgens de structuur van FIDE-handboek sectie C.07. Elke tiebreaker wordt aangeduid met een korte ID die u in de configuratie gebruikt.

### Buchholz-familie

Som van de scores van de tegenstanders, met varianten die extreme waarden uitsluiten.

| Tiebreaker                | ID                      |
| ------------------------- | ----------------------- |
| Buchholz (volledig)       | `buchholz`              |
| Buchholz Cut-1            | `buchholz-cut1`         |
| Buchholz Cut-2            | `buchholz-cut2`         |
| Buchholz Median           | `buchholz-median`       |
| Buchholz Median-2         | `buchholz-median2`      |
| Fore Buchholz             | `fore-buchholz`         |
| Average Opponent Buchholz | `avg-opponent-buchholz` |

Zie [Buchholz](/docs/tiebreakers/buchholz/) en [Opponent Buchholz](/docs/tiebreakers/opponent-buchholz/) voor details.

### Onderlinge ontmoeting

Resultaten tussen specifieke tegenstanders, of resultaten gewogen naar de score van de tegenstander.

| Tiebreaker       | ID                 |
| ---------------- | ------------------ |
| Direct Encounter | `direct-encounter` |
| Sonneborn-Berger | `sonneborn-berger` |

Zie [Onderlinge ontmoeting](/docs/tiebreakers/head-to-head/).

### Resultaatgebaseerd

Direct afgeleid van partijuitslagen.

| Tiebreaker                                           | ID                |
| ---------------------------------------------------- | ----------------- |
| Partijen gewonnen (alleen OTB-winsten)               | `wins`            |
| Ronden gewonnen (OTB-winst + forfaitwinst + PAB)     | `win`             |
| Standaardpunten (1-half-0 ongeacht het scoresysteem) | `standard-points` |
| Progressieve (cumulatieve) score                     | `progressive`     |
| Koya-systeem                                         | `koya`            |

Zie [Resultaten](/docs/tiebreakers/results/).

### Prestatiegebaseerd

Afgeleid van spelersratings en de FIDE B.02-conversietabel.

| Tiebreaker                          | ID                   |
| ----------------------------------- | -------------------- |
| Gemiddelde rating van tegenstanders | `aro`                |
| Tournament Performance Rating       | `performance-rating` |
| Performance Points                  | `performance-points` |
| Average Opponent TPR                | `avg-opponent-tpr`   |
| Average Opponent PTP                | `avg-opponent-ptp`   |

Zie [Prestatie](/docs/tiebreakers/performance/).

### Kleur en activiteit

Deelname- en kleurverdelingsmaatstaven.

| Tiebreaker         | ID              |
| ------------------ | --------------- |
| Partijen met zwart | `black-games`   |
| Zwartwinsten       | `black-wins`    |
| Ronden gespeeld    | `rounds-played` |
| Partijen gespeeld  | `games-played`  |

Zie [Kleur & Activiteit](/docs/tiebreakers/color-activity/).

### Ordening

Deterministische eindtiebreakers wanneer al het andere gelijk is.

| Tiebreaker       | ID               |
| ---------------- | ---------------- |
| Rangnummer (TPN) | `pairing-number` |
| Spelerrating     | `player-rating`  |

Zie [Ordening](/docs/tiebreakers/ordering/).

### FIDE-standaardwaarden

Wanneer u geen tiebreakers expliciet opgeeft, past chesspairing de door FIDE aanbevolen standaardwaarden toe per indelingssysteem. Voor Zwitserse systemen zijn dat Buchholz Cut-1, Buchholz, Sonneborn-Berger en Direct Encounter. Round-robin gebruikt standaard Sonneborn-Berger, Direct Encounter, Wins en Koya. U kunt deze overschrijven in de [configuratie](/docs/formats/configuration/).

## Scoresystemen

Er zijn drie score-engines beschikbaar, elk instelbaar met eigen puntwaarden:

| Systeem  | Standaardpunten                | Documentatie                         |
| -------- | ------------------------------ | ------------------------------------ |
| Standard | Winst 1, Remise 0.5, Verlies 0 | [Standaard](/docs/scoring/standard/) |
| Football | Winst 3, Remise 1, Verlies 0   | [Voetbal](/docs/scoring/football/)   |
| Keizer   | Iteratieve convergentie        | [Keizer](/docs/scoring/keizer/)      |

Alle drie verwerken byes, forfait en afwezigheid met instelbare puntwaarden. Zie het onderdeel [Scoresystemen](/docs/scoring/) voor alle details.

## Volgende stappen

- [CLI Snelstart](../cli-quickstart/) -- deel een toernooi in vanuit de commandoregel in vijf minuten
- [Indelingssystemen](/docs/pairing-systems/) -- gedetailleerde documentatie voor elk indelingsalgoritme
- [Concepten](/docs/concepts/) -- basisprincipes van het Zwitsers systeem, byes, kleuren, floaters, forfait
- [CLI Referentie](/docs/cli/) -- alle beschikbare commando's en uitvoerformaten
