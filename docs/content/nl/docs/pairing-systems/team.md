---
title: "Team-Zwitsers systeem"
linkTitle: "Team-Zwitsers"
weight: 6
description: "FIDE C.04.6 — Zwitserse indeling voor teamcompetities met kleurverdeling op bordniveau."
---

Het Team-Zwitserse systeem deelt teams in in plaats van individuele spelers, met dezelfde lexicografische groepsindelingsinfrastructuur als het Dubbel-Zwitserse systeem. Spelers koppelen aan hun team via `PlayerEntry.TeamID`; oudere states mogen één entry per team gebruiken. Het systeem heeft een negenstaps kleurverdelingsprocedure en twee configureerbare kleurvoorkeurstypes die bepalen hoe sterk de engine probeert de kleurbalans op teamniveau te respecteren. Goedgekeurd door de FIDE in oktober 2025 en van kracht sinds februari 2026.

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
| `primaryScore`        | `string`     | `"match"` | Primaire score: `"match"` of `"game"`. De andere score bepaalt de kleurverdeling (Art. 4.2.2). |
| `secondaryScoreForColour` | `bool` | `true` | `false` schakelt de secundaire score uit wanneer de competitiebepalingen dit voorschrijven (Art. 4.2.2). |
| `forbiddenPairs`      | `[][]string` | nil       | Paren van team-ID's die nooit tegen elkaar ingedeeld mogen worden.                                                                                                            |

## Hoe het werkt

Het indelingsalgoritme verloopt in vijf stappen en deelt het grootste deel van zijn infrastructuur met het Dubbel-Zwitserse systeem via het `pairing/lexswiss`-pakket.

### 1. Deelnemersstaten opbouwen

De teamindeler bouwt één indelingsstaat per actief team op. Indelingsscores gebruiken de geconfigureerde primaire score: standaard wedstrijdpunten, of partijpunten wanneer `primaryScore` `"game"` is. De onderliggende punten volgen de puntinstellingen van het toernooi: partijpunten volgen de geconfigureerde bordpunten, forfaitborden de geconfigureerde forfaitwaarden, en een nog niet afgeronde of afgebroken wedstrijd geeft aan beide teams de punten van een remisematch, alleen voor de indelingsscore. Nulpunt-byes en afwezigheden tellen de geconfigureerde waarde van hun soort bij de partijpunten op, zodat ze de wedstrijdpunten niet veranderen. Elk team krijgt een eigen vaste TPN van 1 tot en met het aantal teams (art. 1.1). Wanneer de competitiebepalingen geen volgorde voorschrijven, gebruikt de bibliotheek haar initiële teamvolgorde vóór ronde één; teams worden daarna voor de indeling op score aflopend en TPN oplopend gezet, zonder hernummering per ronde.

### 2. De bye toewijzen

Als het aantal teams oneven is, selecteert de gedeelde `lexswiss.AssignPAB` de ontvanger van de bye volgens art. 3.4. Van de volgens C2 in aanmerking komende teams (geen eerdere PAB, forfaitwinst of verouderde bye met een vol punt) blijven eerst alleen teams over van wie het weglaten een volledige geldige C1-indeling mogelijk maakt. Daarna gelden achtereenvolgens de laagste score, het grootste aantal werkelijk gespeelde teamwedstrijden en het grootste vaste TPN. Een teamwedstrijd die door forfait eindigde, is niet gespeeld. Een partij wordt als forfait behandeld wanneer het resultaat een forfait is of de vlag `IsForfeit` is ingesteld.

### 3. Scoregroepen opbouwen

Teams worden per score gegroepeerd in aflopende scoregroepen. Binnen elke groep zijn de teams oplopend geordend op TPN.

### 4. Lexicografische groepsindeling

Elke bracket wordt ingedeeld volgens de identifier-volgorde van artikel 3.6. De zoekopdracht kiest eerst de gesorteerde verzameling topleden en wijst daarna de onderleden toe in oplopende TPN-volgorde. Een identifier bestaat uit de TPN's van de topleden gevolgd door die van de bijbehorende onderleden. In een eerste ronde met zes teams zonder beperkingen is dit `1-4, 2-5, 3-6`.

**Absolute criteria (altijd gehandhaafd):**

- C1: Twee teams spelen niet vaker dan één keer tegen elkaar (art. 2.1.1).
- Verboden paren worden nooit ingedeeld.

**Kwaliteitscriteria (C8, C9, C10):**

Volgens art. 3.6.4 kiest het algoritme de eerste identifier onder de indelingen die de kwaliteitscriteria C8, C9 en C10 in aflopende prioriteit minimaliseren (art. 2.3):

- **C8 (kleurvoorkeuren):** Minimaliseert het aantal teams van wie de eventuele kleurvoorkeur niet wordt gehonoreerd (art. 2.3.5). In een paar waarin beide teams dezelfde kleur willen, blijft precies één team onvervuld (kosten 1); paren met tegengestelde of enkele voorkeuren kunnen beide worden gehonoreerd (kosten 0).
- **C9 (sterke kleurvoorkeuren, alleen Type B):** Minimaliseert het aantal teams van wie de sterke kleurvoorkeur niet wordt gehonoreerd (art. 2.3.6). Wanneer beide teams een sterke voorkeur voor dezelfde kleur hebben, blijft er één onvervuld (kosten 1).
- **C10 (tegenstanders van upfloaters):** Met uitzondering van de laatste twee rondes (als `TotalRounds` is ingesteld en `CurrentRound >= TotalRounds - 1`), minimaliseert het aantal upfloaters van wie de tegenstander in de vorige ronde een floater was (art. 2.3.7).

C8 domineert C9 strikt, en C9 domineert C10 strikt. Deze kwaliteitscriteria filteren kandidaatparen niet vooraf uit; ze worden exact geminimaliseerd via gewogen matching op de gekozen bracket. Kleurvoorkeuren vormen geen absolute belemmering (alleen C1 en C2 zijn absoluut, volgens art. 2.1 en het voorwoord).

Brackets worden opgebouwd van de hoogste nog niet ingedeelde scoregroep naar beneden volgens C.04.6 art. 3.3.2, waarbij alle eigen teams in hun bracket blijven. Sets upfloaters worden gekozen volgens C4, C5, C6, C7 en de volgorde van artikel 3.5.4; de setkeuze gebruikt C8-C10 niet (minimalisatie geldt alleen voor de gekozen set). Elke set moet een C1/C2-voltooiing overlaten voor de overblijvende teams. Behalve in de laatste twee rondes geeft C7 de voorkeur aan teams die in de voorgaande ronde geen floater waren. Mislukte brackets worden nooit samengevoegd; art. 3.3.3 meldt een onmogelijke ronde.

### 5. Kleurverdeling (9 stappen)

`AllocateColor` implementeert de negenstaps kleurverdelingsprocedure van Art. 4:

1. **Bepaal het eerste team** (Art. 4.2): Hogere primaire score, dan hogere secundaire score (indien beschikbaar), dan kleinere TPN.
2. **Geen geschiedenis** (Art. 4.3.1): Als geen van beide teams een wedstrijd heeft gespeeld, wijs toe op basis van TPN-pariteit. Oneven TPN krijgt de initiële kleur; even TPN krijgt de tegenovergestelde.
3. **Eén voorkeur** (Art. 4.3.2): Als slechts één team een kleurvoorkeur heeft, honoreer die.
4. **Tegengestelde voorkeuren** (Art. 4.3.3): Als beide teams tegengestelde voorkeuren hebben, honoreer beide.
5. **Sterk vs. niet-sterk** (Art. 4.3.4, alleen Type B): Als slechts één team een sterke voorkeur heeft, honoreer die.
6. **Kleurverschil** (Art. 4.3.5): Het team met het lagere kleurverschil (minder wit minus zwart) krijgt wit.
7. **Afwisseling** (Art. 4.3.6): Verwijder ongespeelde wedstrijden, lijn beide geschiedenissen uit op hun meest recente gespeelde wedstrijd, zoek dan de meest recente tegengestelde kleuren en wissel af.
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
| Matching-algoritme       | Identifiervolgorde + gewogen matching voor C8-C10 | Identifiervolgorde + gewogen matching voor C8 | Globale Blossom            |
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

De identifier-opsomming van artikel 3.6 is identiek aan die van Dubbel-Zwitsers. Zie de [wiskundige grondslagen van Dubbel-Zwitsers](../double-swiss/#lexicographic-enumeration) voor de formele beschrijving.

## FIDE-referentie

- **Reglement**: FIDE C.04.6 (Team-Zwitsers systeem)
- **Aangenomen**: oktober 2025
- **Van kracht**: februari 2026
- **Belangrijke artikelen**: Art. 1.2 (primaire/secundaire score), Art. 1.6.1 (kleur per eerste bord), Art. 1.7 (kleurvoorkeurtypes), Art. 3.4 (bye-toewijzing), Art. 3.5 (floater-selectie), Art. 3.6 (lexicografische groepsindeling), Art. 4 (negenstaps kleurverdeling)

## Team-byes vastleggen

Kopieer `PairingResult.TeamByes` naar `RoundData.TeamByes`. Dit is het
voorkeurskanaal voor team-byes. Voor compatibiliteit herkent de teamindeler ook
een `RoundData.Byes`-entry als een team-bye wanneer de ID een team aanduidt.

## Beperkingen

- De pairer leidt de teamset niet af uit `Matches`.
- De TRF-omzetter leidt de teamset niet af uit `Matches`; teams die alleen in Matches voorkomen krijgen geen 013/801/802-records.
- De TRF-omzetter schrijft geen echte bord-1-kleuren naar 801/802-records. Wedstrijden met alleen totalen kunnen geen individuele 001-tegenstanders leveren.
- De optie `primaryScore` wordt niet automatisch samengevoegd tussen de instellingen van de pairer en de scorer.
- Tie-break PAB-afhandeling voor Art. 12 TBR/BBE is incompleet.
- De scorer leest `Games` niet voor staten zonder `Matches`.
