---
title: "Dutch-systeem"
linkTitle: "Dutch"
weight: 1
description: "Het standaard FIDE Zwitserse indelingssysteem (C.04.3) — globale Blossom matching met 21 optimalisatiecriteria."
---

Als je ooit in een geratingd Zwitsers toernooi hebt gespeeld, heeft het Dutch-systeem vrijwel zeker bepaald tegenover wie je zat. Het is het standaard FIDE Zwitserse algoritme en het systeem dat de meeste arbiters en spelers als vanzelfsprekend beschouwen. De engine groepeert iedereen op score, splitst elke groep in een bovenste en onderste helft, en probeert over de helften heen te indelen met inachtneming van kleurgeschiedenis, rematchvermijding en zo eerlijk mogelijke floaterverdeling. Achter die eenvoudige beschrijving zit een globale Blossom matching-graaf, 21 gelaagde optimalisatiecriteria en kantgewichten die arbitrary-precision integers vereisen.

## Wanneer gebruiken

- **Standaard geratingde toernooien.** Elk FIDE-geratingd of nationaal geratingd Zwitsers toernooi dat geen alternatief systeem voorschrijft.
- **Open en gesloten Zwitserse toernooien** van elke grootte, van een weekendrapid met 20 spelers tot een open toernooi met 500+.
- **Wanneer je maximale indelingskwaliteit nodig hebt.** De 21 criteria en globale matching produceren indelingen die zo goed als wiskundig mogelijk aan de FIDE-regels voldoen.
- **Wanneer Baku-acceleratie gewenst is.** De Dutch-engine ondersteunt FIDE C.04.7-acceleratie voor diversiteit in vroege rondes.

Overweeg [Burstein](../burstein/) voor toernooien die herrangschikking op oppositie-index nodig hebben na een initiële fase. Zie [Dubov](../dubov/) voor ARO-gesorteerde verwerking. Zie [Lim](../lim/) voor mediaan-eerst floaterlogica.

## Configuratie

### CLI

```bash
chesspairing pair --dutch tournament.trf
```

De `--dutch`-vlag selecteert de Dutch-indelingsengine. Het uitvoerformaat wordt ingesteld met `--format` (list, wide, board, xml, json) of de `-w`-afkorting voor brede uitvoer.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dutch"

// Met getypeerde opties
p := dutch.New(dutch.Options{
    Acceleration:   chesspairing.StringPtr("baku"),
    TopSeedColor:   chesspairing.StringPtr("white"),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// Vanuit een generieke map (bijv. geparsede JSON-configuratie)
p := dutch.NewFromMap(map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
    "forbiddenPairs": []any{[]any{"P1", "P2"}},
})

result, err := p.Pair(ctx, &state)
```

### Opties

| Optie            | Type         | Standaard | Beschrijving                                                                                                                       |
| ---------------- | ------------ | --------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `Acceleration`   | `*string`    | `"none"`  | `"none"` of `"baku"`. Baku-acceleratie (FIDE C.04.7) voegt virtuele punten toe in vroege rondes om ratingniveaus eerder te mengen. |
| `TopSeedColor`   | `*string`    | `"auto"`  | `"auto"`, `"white"` of `"black"`. Bepaalt welke kleur de topseed in ronde 1 krijgt. Volgende borden wisselen af.                   |
| `ForbiddenPairs` | `[][]string` | `nil`     | Paren van speler-ID's die nooit tegen elkaar ingedeeld mogen worden, afgedwongen als absoluut criterium naast C1 en C3.            |

Alle velden volgen het pointer-nil-patroon: een nil-waarde betekent "gebruik de standaard". Roep `WithDefaults()` expliciet aan of laat `New()` het voor je doen.

### Fouten

| Fout                   | Voorwaarde                                                    |
| ---------------------- | ------------------------------------------------------------- |
| `ErrTooFewPlayers`     | Minder dan 2 actieve spelers in de toernooi-status.           |
| `ErrNoPairingPossible` | Geen geldige indeling mogelijk gezien de huidige beperkingen. |

## Hoe het werkt

De Dutch-engine volgt zeven stappen, overeenkomend met de architectuur van bbpPairings:

### 1. Spelerstaten opbouwen

De volledige toernooigeschiedenis van elke actieve speler wordt gecompileerd tot een `PlayerState`: score, kleurgeschiedenis, tegenstanders, floatergeschiedenis, bye-status en rangnummer (TPN).

### 2. Baku-acceleratie toepassen (optioneel)

Wanneer `Acceleration` op `"baku"` staat, worden virtuele punten toegevoegd aan de indelingsscore van elke speler volgens FIDE C.04.7:

- **Groep A-grootte** = 2 \* ceil(N / 4), waarbij N het aantal deelnemers vóór de eerste ronde is (late inschrijvingen tellen niet mee).
- **Versnelde rondes** = ceil(totalRounds / 2). De eerste helft hiervan (naar boven afgerond) geeft evenveel virtuele punten als een winstpartij; de rest de helft daarvan.
- Alleen Groep A-spelers krijgen virtuele punten. De laatste speler van Groep A blijft in elke ronde dezelfde speler, dus een late inschrijving die hoger in de lijst staat dan die speler komt in Groep A (zie [Baku-acceleratie](/docs/algorithms/baku-acceleration/)).
- De puntentelling moet een winstpartij evenveel laten opleveren als twee remises en een verliespartij niets (C.04.7 1.1), en het totale aantal rondes moet bekend zijn.

Dit duwt topgeratingde spelers in vroege rondes naar verschillende scorebrackets, waardoor ze niet allemaal direct bovenaan clusteren.

### 3. Scoregroepen opbouwen

Spelers worden verdeeld in scoregroepen, geordend van hoogste naar laagste score. Binnen elke groep worden spelers gesorteerd op TPN oplopend (sterkste eerst).

### 4. Globale Blossom matching

Dit is de kern van het algoritme. In plaats van elke bracket afzonderlijk te indelen (wat suboptimale resultaten kan opleveren), bouwt de Dutch-engine een enkele globale matching-graaf met alle spelers en voert Edmonds' maximum weight Blossom-algoritme uit om de optimale indeling te vinden.

Het proces heeft twee fasen:

**Fase 0.5 -- Completability pre-matching** (alleen bij oneven spelaantallen). Een vereenvoudigde Blossom matching bepaalt welke speler de indelings-toegekende bye (PAB) krijgt. De vereenvoudigde kantgewichten coderen bye-geschiktheid, scoremaximalisatie en topscorer-bescherming. De score van de ongepaarde speler wordt meegenomen in de echte kantgewichten.

**Hoofd-matching -- 7-fasen bracketloop.** Scoregroepen worden van boven naar beneden verwerkt. Voor elke bracket worden kanten in de globale graaf ingevoegd met gewichten die alle 21 criteria coderen. Het Blossom-algoritme draait incrementeel en legt paren uit de huidige bracket vast voordat het naar de volgende gaat. Deze incrementele aanpak weerspiegelt de `computeMatching`-procedure van bbpPairings.

### 5. Bordvolgorde

Vastgelegde paren worden gesorteerd voor bordtoewijzing:

1. **Maximumscore aflopend** -- het paar met de hogerscoerende speler komt eerst.
2. **Bracketscore aflopend** -- onder paren met dezelfde maximale spelerscore komen homogene paren (beide spelers inheems aan de bracket) voor heterogene paren (een speler ingevloat).
3. **Minimum-TPN oplopend** -- gelijkspel wordt verbroken door het TPN van de sterkste speler (lager nummer = hoger bord).

### 6. Kleurverdeling

Kleuren worden toegewezen met een zes-prioriteiten-algoritme dat overeenkomt met `choosePlayerNeutralColor` en `choosePlayerColor` van bbpPairings:

1. Compatibele voorkeuren -- de voorkeuren van beide spelers kunnen tegelijk bevredigd worden.
2. Absolute voorkeur wint -- een speler met kleuronbalans > 1 of 2+ opeenvolgende keren dezelfde kleur krijgt voorrang.
3. Sterke voorkeur gaat boven niet-sterk -- onbalans > 0 (maar niet absoluut) overtreft een milde voorkeur.
4. Eerste kleurverschil -- loop terug door beide spelers' kleurgeschiedenissen en wissel vanaf de meest recente ronde waarin ze verschilden.
5. Gelijke-kleurconflict -- wanneer beiden dezelfde kleur willen met gelijke sterkte, krijgt de hoger gerangschikte speler zijn voorkeur.
6. Geen voorkeur -- pas de pariteitsregel van 5.2.5 toe (interpretatie): als
   de hoger gerangschikte speler een oneven rangnummer heeft onder de spelers
   die aan het toernooi zijn toegetreden, krijgt die de initiële kleur; anders
   de tegenovergestelde kleur (instelbaar via `TopSeedColor`). De letterlijke
   tekst van 5.2.5 spreekt van "oneven TPN", maar de Geannoteerde indelingsregels
   voor het FIDE (Nederlandse) systeem (editie 2025/2026, artikel 5.2.5)
   beschrijven precies dit resultaat met pariteit over toegetreden spelers,
   zodat deze implementatie alleen spelers meetelt die zijn toegetreden.

In de laatste ronde gelden topscorer-regels: spelers met meer dan 50% van de maximaal mogelijke score krijgen speciale aandacht om kleurgebaseerd competitief nadeel te voorkomen.

### 7. Bye-toewijzing

Als het aantal spelers oneven is, krijgt de enkele ongepaarde speler uit de Blossom matching een indelings-toegekende bye (PAB).

## Vergelijking met andere systemen

| Aspect                   | Dutch                        | [Burstein](../burstein/)               | [Dubov](../dubov/)     | [Lim](../lim/)                  |
| ------------------------ | ---------------------------- | -------------------------------------- | ---------------------- | ------------------------------- |
| **Matchingalgoritme**    | Globale Blossom              | Globale Blossom                        | Transpositie-gebaseerd | Uitwisseling-gebaseerd          |
| **Aantal criteria**      | 21 (C1-C21)                  | C1-C4, C10-C13 alleen                  | 10 (C1-C10)            | Compatibiliteit + floatertypen  |
| **S1/S2-splitsingen**    | Ja (bovenste/onderste helft) | Ja                                     | G1/G2-splitsing        | Nee (uitwisseling binnen groep) |
| **C8 vooruitblik**       | Ja (MatchBracketFeasible)    | Nee                                    | Nee                    | Nee                             |
| **Floatercriteria**      | C14-C21 (volledig)           | Geen                                   | Geen floatercriteria   | Floatertypen A-D                |
| **Topscorer-regels**     | Ja (laatste ronde)           | Nee                                    | Nee                    | Nee                             |
| **Spelersrangschikking** | TPN doorlopend               | TPN in seeding, oppositie-index daarna | ARO-gebaseerd          | TPN doorlopend                  |
| **Verwerkingsvolgorde**  | Top-down op score            | Top-down op score                      | Oplopende ARO          | Mediaan-eerst                   |

Het Dutch-systeem is het meest uitgebreid qua optimalisatiecriteria. Burstein laat bewust floatercriteria en topscorer-regels weg ten gunste van herrangschikking op oppositie-index. Dubov en Lim gebruiken fundamenteel andere matchingstrategieën (respectievelijk transpositie en uitwisseling) in plaats van Blossom matching.

## Wiskundige grondslagen

De Dutch-engine steunt op verschillende algoritmen die gedocumenteerd zijn in het [Algoritmen](/docs/algorithms/)-gedeelte:

- **[Blossom Matching](/docs/algorithms/blossom/)** -- Edmonds' O(n^3) maximum weight matching voor algemene grafen. Het `algorithm/blossom/`-pakket biedt zowel `int64`- als `*big.Int`-varianten.
- **[Kantgewichtcodering](/docs/algorithms/edge-weights/)** -- De 16+ criteriavelden worden in een enkel `*big.Int`-kantgewicht verpakt met positionele bitcodering. Hogerprioritaire criteria bezetten meer significante bits, zodat het Blossom-algoritme van nature indelingen prefereert die aan de belangrijkste criteria voldoen.
- **[Completeerbaarheid pre-matching](/docs/algorithms/completability/)** -- Fase 0.5 gebruikt een vereenvoudigde Blossom-run met gereduceerde kantgewichten om de bye-ontvanger te bepalen voor de hoofd-matching.
- **[Nederlandse criteria](/docs/algorithms/dutch-criteria/)** -- Gedetailleerde uiteenzetting van alle 21 criteria: C1-C4 (absoluut), C5 (PAB-score), C6-C21 (kwaliteit: afdrijvers, vooruitblik naar volgende bracket, ongespeelde partijen van de bye-ontvanger, kleur en floatergeschiedenis).
- **[Baku-acceleratie](/docs/algorithms/baku-acceleration/)** -- Berekening van virtuele punten, Groep A-grootte en rondeclassificatie.
- **[Kleurverdeling](/docs/algorithms/color-allocation/)** -- De zes-prioriteiten-kleurverdelingsprocedure.

### Waarom big.Int?

Elk kantgewicht codeert 16+ velden over bitbereiken ter grootte van scoregroepen. Bij een toernooi met veel scoregroepen overschrijdt de totale bitbreedte gemakkelijk 64 bits. Het `algorithm/blossom/`-pakket biedt `MaxWeightMatchingBig` specifiek voor dit geval. De int64-variant wordt alleen gebruikt in completability pre-matching waar de vereenvoudigde gewichten in 64 bits passen.

## FIDE-referentie

Het Dutch-systeem is gedefinieerd in FIDE-reglement C.04.3. De implementatie dekt:

- **C.04.3 Artikel 1** -- Definities: scoregroepen, indelingsbrackets (homogeen en heterogeen), floaters (afdrijven/opdrijven), PAB, kleurverschil en kleurvoorkeur, topscorers en de rondindelingsverwachting.
- **C.04.3 Artikel 2.1--2.2** -- Absolute criteria C1-C4 (geen rematches, geen tweede PAB, geen absoluut kleurconflict voor niet-topscorers, groepsvolledigheid).
- **C.04.3 Artikel 2.3** -- PAB-criterium C5 (minimaliseer de score van de PAB-ontvanger).
- **C.04.3 Artikel 2.4** -- Kwaliteitscriteria C6-C21 (minimaliseer afdrijvers, minimaliseer afdrijf-scores, vooruitblik naar C1-C7 in de volgende bracket, ongespeelde partijen van de bye-ontvanger, kleur- en floatergeschiedeniscriteria).
- **C.04.3 Artikel 3** -- Bracketindelingsproces: S1/S2-subgroepen, kandidaatopbouw, transpositie- en uitwisselingswijzigingen.
- **C.04.3 Artikel 4** -- Regels voor de volgorde waarin kandidaten worden gegenereerd: bracket-reeksnummers (BSN), transpositievolgorde, uitwisselingsvolgorde en de volgorde van inzetbare verzamelingen ingevloeide spelers.
- **C.04.3 Artikel 5** -- Kleurverdelingsregels 5.2.1--5.2.5 (oplossing van kleurvoorkeuren en de pariteitsregel voor de initiële kleur).
- **C.04.7** -- Baku-acceleratie (virtuele punten, Groep A, versneld ronde-aantal).

De S1/S2-helftsplitsing, Narayana Pandita-transpositievolgorde en combinatie-gebaseerde uitwisselingsopsomming volgen de procedures beschreven in het FIDE-handboek voor deterministische doorloop van kandidaatindelingen binnen elke bracket.
