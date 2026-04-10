---
title: "Woordenlijst"
linkTitle: "Woordenlijst"
weight: 10
description: "Definities van termen die door deze documentatie worden gebruikt."
---

### ARO (Average Rating of Opponents)

Het rekenkundig gemiddelde van de ratings van alle tegenstanders die een speler heeft ontmoet in daadwerkelijk gespeelde partijen. Forfait-partijen worden uitgesloten. Gebruikt als [tiebreaker](/docs/tiebreakers/).

### Berger-tabel

Een vast rotatieschema dat in [round-robin toernooien](/docs/concepts/round-robin/) bepaalt wie tegen wie speelt in elke ronde. De laatste speler blijft op een vaste positie terwijl de overige spelers door de posities roteren. Gedefinieerd in FIDE C.05 Annex 1.

### Blossom matching

Een algoritme (Edmonds' maximum gewogen matching) dat de optimale set van indelingen vindt in een graaf waar spelers knooppunten zijn en potentiële indelingen gewogen verbindingen. Gebruikt door de [Dutch](/docs/pairing-systems/dutch/)- en [Burstein](/docs/pairing-systems/burstein/)-engines om globaal optimale Zwitserse indelingen te produceren. Zie [Blossom matching](/docs/algorithms/blossom/).

### Buchholz

Een tiebreaker die de eindscores van alle tegenstanders die een speler heeft ontmoet optelt. Er bestaan meerdere varianten: volledige Buchholz, cut-1 (laagste tegenstander-score weglaten), cut-2 (twee laagste weglaten), mediaan (hoogste en laagste weglaten), en mediaan-2 (twee hoogste en twee laagste weglaten). Zie [tiebreakers](/docs/tiebreakers/).

### Bye

Een ronde waarin een speler geen tegenstander heeft. Er zijn zes typen: PAB (pairing-allocated bye, standaard 1 punt waard), halve-punt bye (aangevraagd, 0,5 waard), nulpunten-bye (aangevraagd, 0 waard), afwezig (ongeoorloofd, 0 punten), verontschuldigd (vooraf gemeld), en clubverplichting (afwezig voor interclub-teamplicht). Zie [byes](/docs/concepts/byes/).

### Kleurvoorkeur

De berekende voorkeur van een speler voor wit of zwart op basis van de partijhistorie. Kent drie sterktes: **absoluut** (kleurbalans wijkt meer dan 1 partij af, of de speler heeft twee of meer keer dezelfde kleur achtereen gehad -- de voorkeur moet worden gehonoreerd), **sterk** (onevenwicht van precies 1, niet absoluut), en **mild** (geen onevenwicht, maar afwisseling ten opzichte van de laatste partij heeft de voorkeur). Zie [kleuren](/docs/concepts/colors/).

### Completability

Een pre-matching techniek die door de Dutch- en Burstein-systemen wordt gebruikt om te bepalen welke speler de [bye](/docs/concepts/byes/) moet krijgen wanneer er een oneven aantal spelers is. Het algoritme test of het verwijderen van een kandidaat nog steeds een volledige matching van alle overige spelers toelaat. Zie [completability](/docs/algorithms/completability/).

### Downfloater

Een speler die wordt ingedeeld tegen een tegenstander uit een lagere [scoregroep](/docs/concepts/floaters/). Dit gebeurt wanneer de eigen scoregroep een oneven aantal spelers heeft of intern niet volledig kan worden gematcht.

### Edge weight

Een numerieke waarde die wordt toegekend aan een potentiële indeling in de Blossom matching-graaf. Hogere gewichten vertegenwoordigen wenselijkere indelingen. Edge weights coderen alle optimalisatiecriteria (kleurbevrediging, floating-historie, scoregroep-afstand) als één multi-precisie integer met een bitveldindeling. Zie [Blossom matching](/docs/algorithms/blossom/).

### Exchange (Lim-systeem)

Een matching-techniek die in het [Lim-indelingssysteem](/docs/pairing-systems/lim/) wordt gebruikt waarbij spelers binnen een scoregroep worden herschikt om een geldige set indelingen te vinden. In tegenstelling tot transposities die alleen de onderste helft herordenen, kunnen exchanges spelers tussen de twee helften van de groep wisselen.

### Float / Floater

Wanneer een speler niet binnen de eigen scoregroep kan worden ingedeeld en moet worden gematcht tegen een tegenstander uit een hogere of lagere groep. Zie [floaters](/docs/concepts/floaters/).

### Forfait

Een partijresultaat waarbij één of beide spelers niet zijn verschenen. Enkel forfaits (`1-0f` of `0-1f`) kennen punten toe aan de winnaar maar sluiten de partij uit van indelingshistorie en tiebreaker-berekeningen. Dubbel forfaits (`0-0f`) sluiten de partij geheel uit van zowel scoring als indelingshistorie. Zie [forfaits](/docs/concepts/forfeits/).

### Partijresultaat

De uitkomst van een schaakpartij. Zeven waarden worden herkend: wit wint (`1-0`), zwart wint (`0-1`), remise (`0.5-0.5`), lopend (`*`), forfait wit wint (`1-0f`), forfait zwart wint (`0-1f`) en dubbel forfait (`0-0f`).

### Keizersysteem

Een indelings- en scoresysteem waarbij spelers worden gerangschikt op een iteratief berekende score die het spelen tegen hoger gerangschikte tegenstanders beloont. De indeling verloopt top-down op rang: 1 tegen 2, 3 tegen 4, enzovoort, met vermijding van herhalingen. Zie [Keizer](/docs/pairing-systems/keizer/) en [Keizers scoring](/docs/scoring/).

### Lexicografische indeling

Een matching-aanpak die door de [Double-Swiss](/docs/pairing-systems/double-swiss/)- en [Team Swiss](/docs/pairing-systems/team/)-systemen wordt gebruikt. In plaats van Blossom matching worden indelingen gegenereerd door combinaties in lexicografische (woordenboek-)volgorde uit te proberen en te toetsen aan kwaliteitscriteria, waarbij de eerste combinatie die aan alle eisen voldoet wordt geaccepteerd.

### Pairing-Allocated Bye (PAB)

De bye die aan één speler wordt gegeven wanneer een toernooironde een oneven aantal actieve deelnemers heeft. Standaard een vol punt waard. Een speler zou niet meer dan één PAB per toernooi moeten ontvangen. Zie [byes](/docs/concepts/byes/).

### Rangnummer (TPN)

Het Tournament Pairing Number dat aan elke speler wordt toegekend op basis van de huidige score en initiële rangorde. TPN 1 is de hoogst gerangschikte actieve speler. Wordt elke ronde opnieuw berekend. Zie [rangnummers](/docs/concepts/pairing-numbers/).

### Scoregroep

Alle actieve spelers met dezelfde score (of indelingsscore, wanneer acceleratie van kracht is) vormen een scoregroep. Binnen een scoregroep zijn spelers geordend op TPN. Scoregroepen worden verwerkt van hoog naar laag tijdens de indeling.

### Scoresysteem

De methode waarmee spelerscores worden berekend uit partijresultaten. chesspairing implementeert drie systemen: [standaard](/docs/scoring/) (1-0,5-0), [voetbal](/docs/scoring/) (3-1-0) en [Keizer](/docs/scoring/) (iteratieve waarde-gebaseerde scoring). Elk systeem heeft configureerbare puntwaarden voor winst, remise, verlies, byes, forfaits en afwezigheden.

### Sonneborn-Berger

Een [tiebreaker](/docs/tiebreakers/) die wordt berekend door voor elke tegenstander de eindscore van die tegenstander te vermenigvuldigen met het resultaat van de speler tegen die tegenstander (1 voor winst, 0,5 voor remise, 0 voor verlies). Beloont het verslaan van en remiseren tegen sterke tegenstanders boven het verslaan van zwakke.

### Zwitsers systeem

Een toernooiformaat waarbij spelers met vergelijkbare scores elke ronde tegen elkaar worden ingedeeld, zonder dat elke speler elke andere speler hoeft te ontmoeten. Er bestaan meerdere Zwitserse varianten, elk met eigen regels voor het oplossen van indelingsconflicten. Zie [Zwitsers systeem](/docs/concepts/swiss-system/).

### Tiebreaker

Een secundaire maatstaf die wordt gebruikt om spelers met dezelfde score te rangschikken. chesspairing implementeert [25 tiebreakers](/docs/tiebreakers/) waaronder Buchholz-varianten, Sonneborn-Berger, direct encounter, performance rating en vele anderen. Tiebreakers worden in een geconfigureerde volgorde toegepast om gelijke scores te onderscheiden. Zie [tiebreaking](/docs/concepts/tiebreaking/).

### Toernooi-status

De volledige momentopname van een toernooi op een bepaald tijdstip, meegegeven aan de indelings- en scoring-engines. Bevat de spelerslijst, alle afgeronde ronden met partijresultaten en byes, het huidige rondenummer, en configuratie voor de indelings- en scoresystemen. Engines werken op deze alleen-lezen structuur en wijzigen deze nooit rechtstreeks.

### Transpositie

Een herordening van spelers binnen de onderste helft (S2) van een scoregroep in het Dutch-indelingssysteem. Transposities worden systematisch uitgeprobeerd om een geldige set indelingen te vinden die voldoet aan de absolute criteria (geen rematches, geen absolute kleurconflicten). Te onderscheiden van exchanges, die spelers tussen de twee helften wisselen.

### TRF16

Het FIDE Tournament Report File formaat, versie 16. Een tekstgebaseerd bestandsformaat voor het uitwisselen van toernooigegevens. chesspairing kan TRF16-bestanden lezen, schrijven, valideren en converteren, en kan converteren tussen TRF-documenten en de interne `TournamentState`-representatie.

### Upfloater

Een speler die wordt ingedeeld tegen een tegenstander uit een hogere [scoregroep](/docs/concepts/floaters/). De ontvangende groep stuurt een downfloater; de speler die tegen de downfloater wordt gematcht, float in feite naar boven.

### Varma-tabel

Een opzoektabel die wordt gebruikt om federatie-bewuste [rangnummers](/docs/concepts/pairing-numbers/) toe te wijzen in round-robin toernooien. Spelers van dezelfde federatie worden verdeeld over vier groepen zodat ze elkaar in latere ronden treffen in plaats van vroege. Ondersteunt maximaal 24 spelers. Gedefinieerd in FIDE C.05 Annex 2. Zie [Varma-tabellen](/docs/algorithms/varma-tables/).

### Virtuele punten (Bakoe-acceleratie)

Tijdelijke bonuspunten die worden opgeteld bij de indelingsscore van een speler (niet de echte score) gedurende de vroege ronden van een groot Zwitsers toernooi. Onder Bakoe-acceleratie ontvangen topgerangschikte spelers (Groep A) extra virtuele punten die hen in hogere scoregroepen plaatsen, waardoor de eerste ronden gevarieerder worden. De bonus neemt af over de versnelde ronden en daalt uiteindelijk naar nul. Zie [Zwitsers systeem](/docs/concepts/swiss-system/).
