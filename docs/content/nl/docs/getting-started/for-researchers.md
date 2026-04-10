---
title: "Voor onderzoekers"
linkTitle: "Voor onderzoekers"
weight: 5
description: "Startpunt voor wiskundigen en informatici die geinteresseerd zijn in de algoritmes achter de indeling van schaaktoernooien."
---

Het indelen van schaaktoernooien is een combinatorisch optimalisatieprobleem met randvoorwaarden. Gegeven een verzameling spelers met partijgeschiedenis, ratings, kleurhistorie en diverse toelaatbaarheidscriteria, is het doel om een set indelingen te produceren die voldoet aan harde beperkingen (geen herhaalde tegenstanders, niet drie keer dezelfde kleur op rij) en tegelijk een lexicografische doelfunctie optimaliseert over een dozijn of meer zachte criteria (homogeniteit van scoregroepen, kleuregalisatie, minimalisering van floaterafstand, behoud van ratingvolgorde).

Chesspairing lost dit probleem op voor alle huidige FIDE-indelingssystemen, drie scoresystemen en 25 tiebreakers. Alles is geïmplementeerd in pure Go zonder externe afhankelijkheden -- de broncode is de enige bron van waarheid voor elk hieronder beschreven algoritme.

Deze pagina geeft een overzicht van de belangrijkste algoritmische componenten en verwijst naar de gedetailleerde beschrijvingen in de sectie [Algoritmes](/docs/algorithms/).

## Maximum weight matching via Edmonds' Blossom-algoritme

De kern van de Dutch-, Burstein- en Dubov-indelingssystemen is een reductie naar maximum weight matching in een algemene (niet-bipartiete) graaf. Elk geldig spelerspaar wordt een kant, en de indelingscriteria worden gecodeerd in het kantgewicht zodanig dat de maximum weight matching overeenkomt met de optimale indeling.

Chesspairing bevat een volledige implementatie van Edmonds' Blossom-algoritme (O(n^3)), geporteerd vanuit de Python-referentie van Joris van Rantwijk. Er zijn twee varianten:

- **`MaxWeightMatching`** -- werkt met `int64`-kantgewichten.
- **`MaxWeightMatchingBig`** -- werkt met `*big.Int`-kantgewichten, wat in de praktijk noodzakelijk is omdat de kantgewichtcodering meer dan 64 bits vereist bij realistische toernooigroottes.

Zie [Blossom-algoritme](/docs/algorithms/blossom/).

## Kantgewichtcodering

Het FIDE Dutch-systeem definieert meer dan 16 indelingscriteria in strikte prioriteitsvolgorde (C1 tot en met C21, waarbij sommige criteria absolute beperkingen zijn en andere optimalisatiedoelen). In plaats van het Blossom-algoritme per criteriumlaag uit te voeren met iteratieve correcties, pakt chesspairing alle criteria in een enkel `*big.Int` per kant. Elk criterium beslaat een bitveldsegment met vaste breedte, en de segmenten zijn geordend van meest significant (hoogste prioriteit) tot minst significant (laagste prioriteit).

Dit reduceert de multi-objectieve lexicografische optimalisatie tot een enkele maximum weight matching-aanroep. De bitindeling is zo ontworpen dat het voldoen aan een criterium met hogere prioriteit altijd zwaarder weegt dan elke combinatie van criteria met lagere prioriteit.

Zie [Kantgewichtcodering](/docs/algorithms/edge-weights/).

## Completability pre-matching (Stage 0.5)

Wanneer een toernooironde een oneven aantal actieve spelers heeft, moet precies een speler een bye krijgen via de indeling. De keuze van de bye-ontvanger beïnvloedt de haalbaarheid en kwaliteit van de overige indelingen. De verkeerde speler selecteren kan het onmogelijk maken om de rest van het veld te indelen zonder absolute beperkingen te schenden.

Chesspairing gebruikt een completability pre-matching-fase (Stage 0.5, naar het model van bbpPairings) die een vereenvoudigde Blossom-matching uitvoert voor elke bye-kandidaat. Een kandidaat is alleen levensvatbaar als de overige spelers volledig ingedeeld kunnen worden. Onder de levensvatbare kandidaten wordt degene uit de laagste scoregroep met het hoogste rangnummer geselecteerd -- conform de FIDE-regels voor bye-toewijzing.

Zie [Completeerbaarheid](/docs/algorithms/completability/).

## Lexicografische groepsindeling

De Double-Swiss- (C.04.5) en Team Swiss-systemen (C.04.6) gebruiken een andere aanpak dan Blossom-matching. Binnen elke scoregroep worden spelers verdeeld in een bovenhelft (S1) en een onderhelft (S2). Het algoritme probeert S1[1] te indelen tegen S2[1], S1[2] tegen S2[2], enzovoort. Wanneer een indeling niet haalbaar is, wordt er teruggestapt en wordt de lexicografisch kleinste geldige indeling geproduceerd door S2-permutaties op volgorde te proberen.

Kwaliteitscriteria (kleuregalisatie, floaterminimalisatie) worden geëvalueerd voor elke kandidaatindeling en gebruikt om te kiezen tussen haalbare alternatieven. Het terugstappen is begrensd door de groepsgrootte, wat het praktisch houdt voor invoer op toernooischaal.

Zie [Lexicografische indeling](/docs/algorithms/lexicographic/).

## Lim exchange matching

Het Lim-systeem (C.04.4.3) hanteert weer een andere aanpak. Het classificeert spelers in vier floatertypes (A tot en met D) op basis van hun floatergeschiedenis en verwerkt scoregroepen vanuit de mediaan naar buiten. Binnen elke scoregroep probeert een exchange-gebaseerde matchingprocedure systematisch transposities van de lager gerangschikte subgroep, en accepteert de eerste indeling die voldoet aan de compatibiliteitsbeperkingen.

De floaterselectie en exchange-volgorde zijn deterministisch, wat reproduceerbare indelingen garandeert. Dit is een fundamenteel andere algoritmische structuur dan zowel de Blossom-gebaseerde als de lexicografische aanpak.

Zie [Lim Exchange Matching](/docs/algorithms/lim-exchange/).

## Berger-tabelrotatie voor round-robin

Round-robin-indeling volgt de FIDE Berger-tabellen (C.05, Annex 1). Voor n spelers wordt ronde k gegenereerd door een rotatie van rangnummers waarbij speler n vaststaat. Bij oneven n wordt de vaste positie de bye-plek. De implementatie ondersteunt meerdere cycli (dubbel round-robin, etc.) en een optionele verwisseling van de laatste twee rondes om de kleurbalans te verbeteren.

Zie [Bergertabellen](/docs/algorithms/berger-tables/).

## Varma-tabellen voor toewijzing van rangnummers

In round-robintoernooien beïnvloedt de initiële toewijzing van rangnummers aan spelers de kleurverdeling over het toernooi. De Varma-tabellen (FIDE C.05, Annex 2) bieden een federatiebewuste toewijzing die voorkomt dat spelers van dezelfde federatie elkaar in de vroege rondes tegenkomen. De implementatie bevat de volledige opzoektabellen en een federatiebewust toewijzingsalgoritme.

Zie [Varma-tabellen](/docs/algorithms/varma-tables/).

## Baku-versnelling

In de openingsrondes van een groot Zwitsers toernooi bevatten de bovenste scoregroepen veel spelers met identieke scores, waardoor de indelingen binnen die groepen enigszins willekeurig zijn. Baku-versnelling (FIDE C.04.7) kent virtuele punten toe aan de hoogst geplaatste spelers in vroege rondes, wat meer gedifferentieerde scoregroepen oplevert en vanaf het begin zinvollere indelingen produceert. De virtuele punten worden verwijderd nadat de versnellingsfase eindigt.

Zie [Baku-acceleratie](/docs/algorithms/baku-acceleration/).

## Keizer-scoreconvergentie

Keizer-scoring is een iteratief algoritme. De score van elke speler hangt af van de scores van diens tegenstanders (sterkere tegenstanders leveren meer punten op), die op hun beurt afhangen van de scores van hun tegenstanders, enzovoort. De implementatie lost deze circulariteit op door te itereren: bereken scores, rangschik opnieuw, herbereken, en herhaal totdat de rangschikking stabiliseert.

Alle rekenkunde gebruikt verdubbelde gehele getallen om floating-point-problemen te vermijden. Convergentie is in de praktijk gegarandeerd, maar de implementatie bevat oscillatiedetectie en begrenst het aantal iteraties op 20. In de meeste toernooien stabiliseert de rangschikking binnen 3-5 iteraties.

Zie [Keizer-convergentie](/docs/algorithms/keizer-convergence/).

## Kleurtoewijzing

Nadat de indelingen bepaald zijn, moet aan elk bord een kleur worden toegewezen. Elk FIDE-indelingssysteem specificeert zijn eigen kleurtoewijzingsprocedure met verschillende prioriteitsregels. De algemene aanpak balanceert kleurgeschiedenis, respecteert kleurvoorkeuren en kleurrecht, en voorkomt dat dezelfde kleur drie keer op rij wordt gegeven. Het Dutch-systeem gebruikt een meertraps-prioriteit met alternatie als tiebreaker; het Team Swiss-systeem heeft een 9-stapsprocedure; en het Double-Swiss-systeem gebruikt een 5-stapsprioriteit.

Zie [Kleurverdeling](/docs/algorithms/color-allocation/).

## Dutch- en Dubov-optimalisatiecriteria

Het Dutch-systeem definieert criteria C1 tot en met C21 (absolute beperkingen C1-C6, optimalisatiecriteria C8-C21). Het Dubov-systeem definieert een eigen set van tien criteria (C1-C10) met andere prioriteiten. Beide sets zijn geïmplementeerd als functies die een voorgestelde indeling evalueren ten opzichte van de toernooistatus en bijdragen aan kantgewichten.

Zie [Nederlandse criteria](/docs/algorithms/dutch-criteria/) en [Dubov-criteria](/docs/algorithms/dubov-criteria/).

## FIDE B.02-conversietabel

Verschillende prestatiegebaseerde tiebreakers (TPR, PTP, APRO, APPO) vereisen conversie tussen verwachte scores en ratingverschillen. De FIDE B.02-conversietabel biedt deze afbeelding. Chesspairing bevat de volledige tabel en interpolatielogica.

Zie [FIDE B.02](/docs/algorithms/fide-b02/) en [Elo-model](/docs/algorithms/elo-model/).

## De code lezen

De gehele codebase is opgezet voor leesbaarheid. Belangrijkste ingangspunten:

| Algoritme                      | Package               | Ingangsfunctie                              |
| ------------------------------ | --------------------- | ------------------------------------------- |
| Blossom-matching               | `algorithm/blossom`   | `MaxWeightMatching`, `MaxWeightMatchingBig` |
| Kantgewichtberekening          | `pairing/swisslib`    | `ComputeBaseEdgeWeight`                     |
| Completability                 | `pairing/swisslib`    | `PairBracketsGlobal` (Stage 0.5)            |
| Dutch-indeling                 | `pairing/dutch`       | `Pair`                                      |
| Burstein-indeling              | `pairing/burstein`    | `Pair`                                      |
| Dubov-indeling                 | `pairing/dubov`       | `Pair`                                      |
| Lim-indeling                   | `pairing/lim`         | `Pair`                                      |
| Lexicografische groepsindeling | `pairing/lexswiss`    | `PairBracket`                               |
| Double-Swiss-indeling          | `pairing/doubleswiss` | `Pair`                                      |
| Team Swiss-indeling            | `pairing/team`        | `Pair`                                      |
| Berger-tabelrotatie            | `pairing/roundrobin`  | `Pair`                                      |
| Varma-toewijzing               | `algorithm/varma`     | `Groups`, `Assign`                          |
| Baku-versnelling               | `pairing/swisslib`    | `AdjustScoreGroups`                         |
| Keizer-scoring                 | `scoring/keizer`      | `Score`                                     |
| Tiebreaker-register            | `tiebreaker`          | `Get`, `All`                                |

Zie het [API-overzicht](/docs/api/overview/) en [Kerntypen](/docs/api/core-types/) voor een rondleiding door het typesysteem en de interfaces.

## Volgende stappen

- [Algoritmes](/docs/algorithms/) -- wiskundige verdiepingen met formules, pseudocode en bewijsschetsen
- [API-referentie](/docs/api/) -- het Go-typesysteem, interfaces en packagestructuur
- [Go-snelstart](../go-quickstart/) -- gebruik de bibliotheek rechtstreeks in Go-code
- [Indelingssystemen](/docs/pairing-systems/) -- documentatie op regelementniveau voor elke indelingsengine
