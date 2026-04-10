---
title: "Burstein-systeem"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 — een Zwitserse variant met seedingrondes en oppositie-index herranking."
---

Het Burstein-systeem splitst een Zwitsers toernooi in twee duidelijke fasen. In de openingsrondes -- de seedingrondes -- werkt de indeling precies zoals het Dutch-systeem, met initiële rangschikkingen om de indeling te maken. Nadat die fase eindigt, wordt elke speler opnieuw gerangschikt op basis van een oppositie-index die Buchholz- en Sonneborn-Berger-scores combineert. Het idee is eenvoudig: zodra er genoeg partijen gespeeld zijn, moet je rangschikking de sterkte weerspiegelen van de tegenstanders die je daadwerkelijk hebt getroffen en hoe goed je het tegen hen deed, niet alleen je rating van voor het toernooi. Vanaf dat punt stuurt de herrankte volgorde alle bracket-opbouw en indelingsbeslissingen aan.

Onder de motorkap gebruikt Burstein dezelfde globale Blossom-matchinginfrastructuur als het Dutch-systeem, maar het laat bewust meerdere lagen van optimalisatie achterwege. Er zijn geen top-scorer-regels, geen C8-vooruitkijken naar toekomstige brackets, en geen float-optimaliseringscriteria. Het resultaat is een eenvoudiger systeem dat een deel van de fijnmazige controle van de Dutch-engine inruilt voor een fundamenteel andere benadering van rangschikkings-eerlijkheid.

## Wanneer gebruiken

- **Toernooien waar vroege resultaten rangschikkingen moeten hervormen.** Als je wilt dat de tweede helft van het toernooi spelers indeelt op basis van wie ze werkelijk speelden in plaats van hun initiële seeding, is Burstein hier precies voor ontworpen.
- **Evenementen waar float-optimalisatie minder belangrijk is.** De vereenvoudigde criteria (alleen kleur, geen float-historie) maken het systeem lichter en de indelingen makkelijker uit te leggen.
- **Als FIDE-conform alternatief voor Dutch.** Burstein is een goedgekeurd FIDE-systeem (C.04.4.2) en kan overal worden gebruikt waar de reglementen een Zwitserse variant toestaan.

Gebruik [Dutch](../dutch/) voor de volledige 21-criteria behandeling. Zie [Dubov](../dubov/) voor oplopende-ARO-verwerking. Zie [Lim](../lim/) voor mediaan-eerst verwerking met floater-typen.

## Configuratie

### CLI

```bash
chesspairing pair --burstein tournament.trf
```

De `--burstein`-vlag selecteert de Burstein-indelingsengine. Het uitvoerformaat wordt geregeld met `--format` (list, wide, board, xml, json) of de `-w`-snelkoppeling voor brede uitvoer.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/burstein"

// Met getypeerde opties
p := burstein.New(burstein.Options{
    Acceleration:   chesspairing.StringPtr("baku"),
    TopSeedColor:   chesspairing.StringPtr("auto"),
    TotalRounds:    chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// Vanuit een generieke map (bijv. geparsede JSON-configuratie)
p := burstein.NewFromMap(map[string]any{
    "acceleration": "none",
    "topSeedColor": "auto",
    "totalRounds":  9.0,
    "forbiddenPairs": []any{[]any{"P1", "P2"}},
})

result, err := p.Pair(ctx, &state)
```

### Opties

| Optie            | Type         | Standaard | Beschrijving                                                                                                                              |
| ---------------- | ------------ | --------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `Acceleration`   | `*string`    | `"none"`  | `"none"` of `"baku"`. Baku-versnelling (FIDE C.04.7) voegt virtuele punten toe in vroege rondes.                                          |
| `TopSeedColor`   | `*string`    | `"auto"`  | `"auto"`, `"white"` of `"black"`. Bepaalt welke kleur de top-seed ontvangt in ronde 1.                                                    |
| `ForbiddenPairs` | `[][]string` | `nil`     | Paren speler-ID's die nooit tegen elkaar ingedeeld mogen worden, afgedwongen als absoluut criterium.                                      |
| `TotalRounds`    | `*int`       | afgeleid  | Gepland totaal aantal rondes in het toernooi. Wordt gebruikt om het aantal seedingrondes te berekenen. Indien nil, afgeleid uit de staat. |

Alle velden volgen het pointer-nil-patroon: een nil-waarde betekent "gebruik de standaard". Roep `WithDefaults()` expliciet aan of laat `New()` het voor je doen.

### Fouten

| Fout                   | Voorwaarde                                                     |
| ---------------------- | -------------------------------------------------------------- |
| `ErrTooFewPlayers`     | Minder dan 2 actieve spelers in de toernooi-staat.             |
| `ErrNoPairingPossible` | Geen geldige indeling mogelijk gegeven de huidige beperkingen. |

## Hoe het werkt

### 1. Spelersstatussen opbouwen

De toernooihistorie van elke actieve speler wordt samengesteld in een `PlayerState`: score, kleurhistorie, tegenstander-lijst, float-historie, bye-status en toernooi-rangnummer (TPN).

### 2. De huidige fase bepalen

Het toernooi wordt verdeeld in seedingrondes en post-seedingrondes:

```text
SeedingRounds = min(floor(TotalRounds / 2), 4)
```

Voor een 9-ronden toernooi zijn de eerste 4 rondes seedingrondes. Voor een 7-ronden toernooi zijn dat de eerste 3. De formule is gemaximeerd op 4 ongeacht de toernooilengte.

| Totaal rondes | Seedingrondes |
| ------------- | ------------- |
| 3             | 1             |
| 5             | 2             |
| 7             | 3             |
| 9+            | 4             |

### 3. Seedingrondes: TPN-gebaseerde rangschikking

Tijdens seedingrondes behouden spelers hun oorspronkelijke TPN-volgorde. Bracket-opbouw en matching werken hetzelfde als in het Dutch-systeem. Deze fase produceert een basis aan resultaten waarmee de oppositie-index later kan werken.

### 4. Post-seedingrondes: oppositie-index herranking

Na de seedingfase worden spelers opnieuw gerangschikt met `RankByOppositionIndex()`. De herranking gebruikt drie tiebreakers in volgorde:

1. **Score** (aflopend) -- dezelfde primaire sortering als standaard rangschikking.
2. **Buchholz** (aflopend) -- som van alle tegenstanders-scores. Hogere Buchholz betekent dat je over het geheel sterkere oppositie had.
3. **Sonneborn-Berger** (aflopend) -- som van (je resultaat tegen elke tegenstander vermenigvuldigd met de score van die tegenstander). Beloont winst tegen sterke tegenstanders meer dan winst tegen zwakkere.
4. **Oorspronkelijk TPN** (oplopend) -- breekt eventuele resterende gelijken.

Na het sorteren worden nieuwe TPN-waarden opeenvolgend toegekend (1, 2, 3, ...). Deze opnieuw toegekende TPN's sturen alle verdere bracket-opbouw en S1/S2-splitsingen aan.

Forfait-partijen worden uitgesloten van de Sonneborn-Berger-berekening. Buchholz omvat alle tegenstanders (inclusief inactieve spelers) om te voorkomen dat een speler benadeeld wordt wiens tegenstander zich terugtrok.

### 5. Baku-versnelling toepassen (optioneel)

Hetzelfde als het Dutch-systeem: wanneer `Acceleration` op `"baku"` staat, worden virtuele punten toegevoegd aan de indelingsscores van Groep A-spelers. Zie [Baku-versnelling](/docs/algorithms/baku-acceleration/) voor de berekeningsdetails.

### 6. Scoregroepen opbouwen

Spelers worden verdeeld in scoregroepen gerangschikt van hoogste naar laagste. Binnen elke groep worden spelers gesorteerd op TPN oplopend -- wat in post-seedingrondes betekent: gesorteerd op oppositie-index.

### 7. Globale Blossom-matching

De matching gebruikt dezelfde `PairBracketsGlobal`-infrastructuur als het Dutch-systeem (completability pre-matching voor oneven aantallen, incrementele bracket-verwerking), maar met cruciale verschillen in de criteria-context:

- **Geen top-scorer-regels.** De `TopScorers`-map is leeg, dus kleur-bescherming in de laatste ronde voor leidende spelers geldt niet.
- **Geen C8-vooruitkijken.** De `LookAhead`-functie is niet ingesteld, dus er wordt niet gecontroleerd of floaters toelaten dat de volgende bracket indelbaar is. Dit vereenvoudigt de matching ten koste van af en toe suboptimale float-verdelingen.
- **Alleen kleurcriteria (C10-C13).** De edge-gewichten coderen absolute kleurschendingen, sterke kleurvoorkeur-tevredenheid, milde kleurvoorkeur-tevredenheid en kleur-onbalans minimalisatie. Float-criteria C14-C21 worden niet gebruikt.

De Blossom-matching produceert nog steeds globaal optimale indelingen binnen de gereduceerde criteriaset.

### 8. Kleurverdeling

Kleuren worden verdeeld met hetzelfde zes-prioriteiten-algoritme als het Dutch-systeem, maar zonder top-scorer-regels. De `topScorerRules`-parameter staat op `false`, zodat de speciale behandeling in de laatste ronde voor leidende spelers wordt overgeslagen.

### 9. Bye-toewijzing

Als het spelersaantal oneven is, ontvangt de enige ongematchte speler uit de Blossom-matching een indelings-toegekende bye (PAB).

## Vergelijking met andere systemen

| Aspect                      | Burstein                               | [Dutch](../dutch/) | [Dubov](../dubov/)     | [Lim](../lim/)                  |
| --------------------------- | -------------------------------------- | ------------------ | ---------------------- | ------------------------------- |
| **Rangschikkingsmethode**   | TPN in seeding, oppositie-index daarna | TPN doorlopend     | ARO-gebaseerd          | TPN doorlopend                  |
| **Seedingrondes**           | Ja (tot 4)                             | Nee                | Nee                    | Nee                             |
| **Matchingalgoritme**       | Globale Blossom                        | Globale Blossom    | Transpositie-gebaseerd | Exchange-gebaseerd              |
| **Optimaliseringscriteria** | C10-C13 (alleen kleur)                 | C1-C21 (volledig)  | C1-C10                 | Compatibiliteit + floater-typen |
| **C8-vooruitkijken**        | Nee                                    | Ja                 | Nee                    | Nee                             |
| **Float-criteria**          | Geen                                   | C14-C21            | Geen                   | Floater-typen A-D               |
| **Top-scorer-regels**       | Nee                                    | Ja (laatste ronde) | Nee                    | Nee                             |
| **Extra opties**            | `TotalRounds`                          | --                 | `TotalRounds`          | `MaxiTournament`                |

Het Burstein-systeem neemt een middenpositie in: het gebruikt dezelfde krachtige Blossom-matchingengine als Dutch maar past een kleinere, meer gerichte set criteria toe. De oppositie-index herranking is het onderscheidende kenmerk -- geen enkele andere Zwitserse variant in deze bibliotheek herordent spelers op basis van tegenstanders-sterkte na de openingsfase.

## Wiskundige grondslagen

De Burstein-indeler deelt het merendeel van zijn algoritmische infrastructuur met het Dutch-systeem:

- **[Blossom Matching](/docs/algorithms/blossom/)** -- Edmonds' O(n^3) maximum weight matching, met `*big.Int`-edge-gewichten.
- **[Kantgewichtcodering](/docs/algorithms/edge-weights/)** -- Dezelfde bit-packed codering als Dutch, maar alleen de kleurgerelateerde velden (C10-C13) dragen betekenisvol gewicht; float-velden (C14-C21) zijn nul.
- **[Completeerbaarheid pre-matching](/docs/algorithms/completability/)** -- Stage 0.5 bye-bepaling, identiek aan Dutch.
- **[Baku-versnelling](/docs/algorithms/baku-acceleration/)** -- Hetzelfde virtuele-puntensysteem indien ingeschakeld.
- **[Kleurverdeling](/docs/algorithms/color-allocation/)** -- Dezelfde zes-prioriteiten procedure, zonder top-scorer-regels.

### Oppositie-index

De oppositie-index wordt per speler berekend als een tuple (Buchholz, Sonneborn-Berger, TPN). Spelers worden eerst op score gesorteerd, daarna op dit tuple.

**Buchholz** is de som van de standaard indelingsscores (1-0.5-0) van alle tegenstanders:

```text
Buchholz(i) = sum over all opponents j of Score(j)
```

**Sonneborn-Berger** weegt de score van elke tegenstander met het behaalde resultaat daartegen:

```text
SB(i) = sum over all games g of Result(i, g) * Score(opponent(i, g))
```

waarbij `Result(i, g)` 1 is voor winst, 0.5 voor remise en 0 voor verlies. Forfait-partijen worden uitgesloten.

De combinatie zorgt ervoor dat onder spelers met gelijke scores, degenen die tegen sterkere oppositie speelden en wonnen hoger gerangschikt worden. Dit levert zinvollere bracket-samenstellingen op in de post-seedingfase dan de onbewerkte TPN-volgorde zou doen.

## FIDE-referentie

Het Burstein-systeem is gedefinieerd in FIDE-reglement C.04.4.2. De implementatie dekt:

- **C.04.4.2 Artikel 1** -- Definitie van seedingrondes en de seedingronde-formule.
- **C.04.4.2 Artikel 2** -- Oppositie-index berekening (Buchholz + Sonneborn-Berger) en herrangschikkingsprocedure voor post-seedingrondes.
- **C.04.4.2 Artikel 3** -- Indelingsprocedure met score-brackets met S1/S2-splitsingen, die de absolute criteria (C1-C4) deelt met het Dutch-systeem.
- **C.04.4.2 Artikel 4** -- Optimalisatie beperkt tot kleurcriteria (C10-C13); float-optimaliseringscriteria (C14-C21) worden niet toegepast.
- **C.04.7** -- Baku-versnelling (gedeeld met Dutch, optioneel).

Het systeem delegeert naar dezelfde Blossom-matchinginfrastructuur als de Dutch-engine, met de criteria-context geconfigureerd om de Burstein-specifieke regels weer te geven: lege top-scorer-map, geen vooruitkijk-functie en alleen kleur-optimaliseringsgewichten.
