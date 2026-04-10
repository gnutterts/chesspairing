---
title: "Team-Zwitsers systeem"
linkTitle: "Team-Zwitsers"
weight: 6
description: "FIDE C.04.6 — Zwitserse indeling voor teamcompetities met kleurverdeling op bordniveau."
---

Het Team-Zwitserse systeem deelt teams in in plaats van individuele spelers, met dezelfde lexicografische groepsindelingsinfrastructuur als het Dubbel-Zwitserse systeem. Elke `PlayerEntry` in `TournamentState` stelt een team voor. Het systeem heeft een negenstaps kleurverdelingsprocedure en twee configureerbare kleurvoorkeurstypes die bepalen hoe sterk de engine probeert de kleurbalans op teamniveau te respecteren. Goedgekeurd door de FIDE in oktober 2025 en van kracht sinds februari 2026.

## Wanneer gebruiken

Team-Zwitsers is geschikt wanneer:

- De competitie tussen teams is (bijv. clubkampioenschappen, Olympiade-achtige evenementen).
- Je een Zwitsers format nodig hebt dat teamindeling combineert met kleurverdeling op bordniveau.
- Het evenement wedstrijdpunten (team wint een wedstrijd) of partijpunten (som van individuele bordresultaten) als primair rangschikkingscriterium gebruikt.

Het is niet geschikt voor individuele toernooien. Gebruik voor individuele tweepartijwedstrijden het [Dubbel-Zwitserse](../double-swiss/) systeem.

## Configuratie

### CLI

```bash
chesspairing pair --team tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/team"

// Met getypeerde opties
p := team.New(team.Options{
    TopSeedColor:        chesspairing.StringPtr("white"),
    TotalRounds:         chesspairing.IntPtr(11),
    ColorPreferenceType: chesspairing.StringPtr("A"),
    PrimaryScore:        chesspairing.StringPtr("match"),
    ForbiddenPairs:      [][]string{{"team-1", "team-2"}},
})

// Vanuit een generieke map (JSON-configuratie)
p := team.NewFromMap(map[string]any{
    "topSeedColor":        "white",
    "totalRounds":         11,
    "colorPreferenceType": "B",
    "primaryScore":        "game",
})
```

### Opties

| Optie                 | Type         | Standaard | Beschrijving                                                                                                                                                                  |
| --------------------- | ------------ | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `topSeedColor`        | `string`     | `"white"` | Initiële kleur uit de indeling voor ronde 1 (Art. 4.1). Waarden: `"white"`, `"black"`. Let op: anders dan bij andere systemen is de standaardwaarde `"white"`, niet `"auto"`. |
| `totalRounds`         | `int`        | nil       | Totaal aantal rondes. Wordt gebruikt om de laatste twee rondes (C7/C10 versoepeling) en de laatste ronde (Type B milde voorkeur) te bepalen.                                  |
| `colorPreferenceType` | `string`     | `"A"`     | Kleurvoorkeursregels (Art. 1.7). Waarden: `"A"` (eenvoudig), `"B"` (sterk + mild), `"none"` (uitgeschakeld).                                                                  |
| `primaryScore`        | `string`     | `"match"` | Score voor de indeling (Art. 1.2). Waarden: `"match"` (wedstrijdpunten), `"game"` (partijpunten). De andere score wordt de secundaire score voor kleurverdeling (Art. 4.2.2). |
| `forbiddenPairs`      | `[][]string` | nil       | Paren van team-ID's die nooit tegen elkaar ingedeeld mogen worden.                                                                                                            |

## Hoe het werkt

Het indelingsalgoritme verloopt in vijf stappen en deelt het grootste deel van zijn infrastructuur met het Dubbel-Zwitserse systeem via het `pairing/lexswiss`-pakket.

### 1. Deelnemersstaten opbouwen

`lexswiss.BuildParticipantStates` bouwt de indelingsstaat op voor elk actief team. Elk team wordt als één deelnemer behandeld. Indelingsscores gebruiken standaard 1-0.5-0, ongeacht het scoresysteem van het toernooi. Teams worden gesorteerd op score (aflopend), dan op initiële ranking (oplopend), en krijgen een TPN toegewezen.

### 2. De bye toewijzen

Als het aantal teams oneven is, selecteert `assignTeamPAB` de ontvanger van de bye volgens Art. 3.4. Team-Zwitsers voegt een extra tiebreaker toe ten opzichte van de basis `lexswiss.AssignPAB`: bij teams met dezelfde laagste score krijgt het team met de meeste gespeelde wedstrijden (langste kleurgeschiedenis) de bye eerst, dan de grootste TPN.

### 3. Scoregroepen en criteria opbouwen

Teams worden per score gegroepeerd in aflopende scoregroepen. De criteriafunctie wordt opgebouwd op basis van het kleurvoorkeurstype en de rondecontext:

**C8 (kleurvoorkeuren):** Twee teams met dezelfde kleurvoorkeursrichting (beide wit willen, of beide zwart willen) kunnen niet allebei worden gehonoreerd in een indeling. Zulke indelingen worden afgewezen tijdens de lexicografische opsomming.

**C9 (alleen Type B):** Twee teams met dezelfde sterke kleurvoorkeur schenden C9. Omdat sterke voorkeuren dezelfde richting impliceren, wordt dit al door C8 afgevangen voor gevallen met dezelfde kleur.

**C7 en C10** worden versoepeld in de laatste twee rondes (als `TotalRounds` is ingesteld en `CurrentRound >= TotalRounds - 1`).

### 4. Lexicografische groepsindeling

Identiek aan Dubbel-Zwitsers: depth-first search die indelingen in lexicografische TPN-volgorde doorloopt, met backtracking. Oneven scoregroepen laten het laagst gerangschikte team naar de groep erboven doorstromen als floater.

### 5. Kleurverdeling (9 stappen)

`AllocateColor` implementeert de negenstaps kleurverdelingsprocedure van Art. 4:

1. **Bepaal het eerste team** (Art. 4.2): Hogere primaire score, dan hogere secundaire score (indien beschikbaar), dan kleinere TPN.
2. **Geen geschiedenis** (Art. 4.3.1): Als geen van beide teams een wedstrijd heeft gespeeld, wijs toe op basis van TPN-pariteit. Oneven TPN krijgt de initiële kleur; even TPN krijgt de tegenovergestelde.
3. **Eén voorkeur** (Art. 4.3.2): Als slechts één team een kleurvoorkeur heeft, honoreer die.
4. **Tegengestelde voorkeuren** (Art. 4.3.3): Als beide teams tegengestelde voorkeuren hebben, honoreer beide.
5. **Sterk vs. niet-sterk** (Art. 4.3.4, alleen Type B): Als slechts één team een sterke voorkeur heeft, honoreer die.
6. **Kleurverschil** (Art. 4.3.5): Het team met het lagere kleurverschil (minder wit minus zwart) krijgt wit.
7. **Afwisseling** (Art. 4.3.6): Zoek de meest recente ronde waarin het ene team wit had en het andere zwart, en wissel dan af.
8. **Voorkeur van het eerste team** (Art. 4.3.7): Honoreer de kleurvoorkeur van het eerste team.
9. **Afwisseling van laatste kleur** (Art. 4.3.8-9): Wissel af ten opzichte van de laatst gespeelde kleur van het eerste team; als het nog steeds gelijk is, wissel af ten opzichte van de laatst gespeelde kleur van het andere team.

De kleur wordt bepaald door de bordtoewijzing op bord 1 (Art. 1.6.1): welk team wit krijgt op bord 1 wordt beschouwd als "Wit" voor de wedstrijd.

## Kleurvoorkeurtypes

### Type A (eenvoudig)

Een team heeft voorkeur voor wit als het kleurverschil lager is dan -1, of als CD gelijk is aan 0 of -1 en de laatste twee gespeelde wedstrijden beide zwart waren. De symmetrische regel geldt voor zwart. Anders geen voorkeur. Dit is een binair systeem: het team heeft wel of geen voorkeur.

### Type B (sterk + mild)

Type B voegt een tweede niveau toe. Sterke voorkeuren gebruiken dezelfde voorwaarden als Type A. Milde voorkeuren gelden wanneer:

- CD is -1: milde voorkeur voor wit.
- CD is 0 en het is niet de laatste ronde en de laatst gespeelde wedstrijd was zwart: milde voorkeur voor wit.
- Symmetrische regels gelden voor zwart.

In de laatste ronde vervallen milde voorkeuren (CD = 0 geeft geen voorkeur). Het onderscheid sterk/mild beïnvloedt stap 5 van de kleurverdeling: een sterke voorkeur heeft voorrang op een niet-sterke.

## Vergelijking

| Aspect                   | Team-Zwitsers              | Dubbel-Zwitsers     | Nederlands                 |
| ------------------------ | -------------------------- | ------------------- | -------------------------- |
| Deelnemers               | Teams                      | Individuele spelers | Individuele spelers        |
| Matching-algoritme       | Lexicografische DFS        | Lexicografische DFS | Globale Blossom            |
| Kleurverdelingstappen    | 9                          | 5                   | FIDE C.04.3 regels         |
| Kleurvoorkeurtypes       | 3 (A, B, None)             | N.v.t.              | N.v.t.                     |
| Primaire score-opties    | Wedstrijd- of partijpunten | Partijpunten        | Standaardpunten            |
| Criteriaversoepeling     | Laatste 2 rondes (C7/C10)  | Laatste ronde (C8)  | Geen speciale versoepeling |
| Standaard `topSeedColor` | `"white"`                  | `"auto"`            | `"auto"`                   |
| Gedeelde infrastructuur  | `pairing/lexswiss`         | `pairing/lexswiss`  | `pairing/swisslib`         |

## Wiskundige grondslagen

### Kleurverschil

Het kleurverschil (CD) voor een team is gedefinieerd als:

```text
CD = (number of White assignments) - (number of Black assignments)
```

Rondes waarin het team geen partij had (bye, afwezigheid) worden uitgesloten. CD stuurt de voorkeursberekening: een negatief CD duwt richting wit, een positief CD duwt richting zwart.

### Eerste-teamvolgorde

De bepaling van het eerste team in Art. 4.2 creëert een strikte totale ordening voor elk paar:

```text
first(a, b) = a  if  score(a) > score(b)
            = a  if  score(a) == score(b) and secondary(a) > secondary(b)
            = a  if  scores equal and TPN(a) < TPN(b)
```

Deze ordening garandeert een deterministische kleurverdeling wanneer alle andere tiebreakers zijn uitgeput.

### Lexicografische indeling

De lexicografische opsomming is identiek aan die van Dubbel-Zwitsers. Zie de [wiskundige grondslagen van Dubbel-Zwitsers](../double-swiss/#lexicographic-enumeration) voor de formele beschrijving.

## FIDE-referentie

- **Reglement**: FIDE C.04.6 (Team-Zwitsers systeem)
- **Aangenomen**: oktober 2025
- **Van kracht**: februari 2026
- **Belangrijke artikelen**: Art. 1.2 (primaire/secundaire score), Art. 1.6.1 (kleur per eerste bord), Art. 1.7 (kleurvoorkeurtypes), Art. 3.4 (bye-toewijzing), Art. 3.5 (floater-selectie), Art. 3.6 (lexicografische groepsindeling), Art. 4 (negenstaps kleurverdeling)
