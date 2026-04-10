---
title: "Dubov-systeem"
linkTitle: "Dubov"
weight: 3
description: "FIDE C.04.4.1 — oplopende ARO-verwerking met transpositie-gebaseerde matching."
---

## Overzicht

Het Dubov-systeem is een Zwitserse variant die erop gericht is de Average Rating of Opponents (ARO) gelijk te trekken over spelers binnen dezelfde scoregroep. Waar het [Dutch-systeem](../dutch/) brackets splitst op rangnummer en globale Blossom-matching gebruikt, splitst Dubov op kleurvoorkeur (G1/G2) en sorteert de wit-zoekende helft op oplopende ARO. Matching gebeurt via transposities van de tegenstandergroep in plaats van gewogen grafenoptimalisatie.

Als je van het Dutch-systeem komt, is het meest zichtbare verschil als speler dat je tegenstanderselectie afhangt van wie je al hebt gespeeld (via ARO) in plaats van uitsluitend van je toernooi-rangnummer.

## Wanneer gebruiken

- Toernooien waar je sterke tegenstanders gelijkmatiger over het veld wilt spreiden, in plaats van ze op de topborden te concentreren.
- Evenementen waar ARO-gelijktrekking prioriteit heeft, zoals round-robin-vervangende formaten of normtoernooien.
- Situaties waar het eenvoudigere 10-criteriamodel (vergeleken met 21 in Dutch) de voorkeur heeft vanwege transparantie.

Het Dubov-systeem ondersteunt **geen** Baku-versnelling. Als je versnelde indelingen nodig hebt, gebruik dan het [Dutch](../dutch/)- of [Burstein](../burstein/)-systeem.

## Configuratie

### CLI

```bash
chesspairing pair --dubov tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dubov"

// Met getypeerde opties
p := dubov.New(dubov.Options{
    TopSeedColor: chesspairing.StringPtr("auto"),
    TotalRounds:  chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// Vanuit een generieke map (bijv. geparsed uit JSON-configuratie)
p := dubov.NewFromMap(map[string]any{
    "topSeedColor": "black",
    "totalRounds":  9.0,
})
```

### Opties

| Optie            | Type         | Standaard | Beschrijving                                                                                                   |
| ---------------- | ------------ | --------- | -------------------------------------------------------------------------------------------------------------- |
| `TopSeedColor`   | `string`     | `"auto"`  | Kleur voor de top-seed in ronde 1. Waarden: `"auto"`, `"white"`, `"black"`.                                    |
| `ForbiddenPairs` | `[][]string` | `nil`     | Paren speler-ID's die nooit tegen elkaar ingedeeld mogen worden.                                               |
| `TotalRounds`    | `int`        | afgeleid  | Gepland totaal aantal rondes. Gebruikt voor interne berekeningen; afgeleid uit de staat indien niet ingesteld. |

Het Dubov-systeem heeft geen `Acceleration`-optie. Baku-versnelling maakt geen deel uit van de C.04.4.1-specificatie.

## Hoe het werkt

### Stapsgewijs algoritme

1. **Spelersstatussen opbouwen** uit toernooihistorie (partijresultaten, kleurhistorie, float-historie, tegenstanders).

2. **Bye-selectie** (oneven spelersaantal): De `DubovByeSelector` kiest de bye-speler voordat de matching begint. Selectiecriteria, in volgorde: laagste score, meeste gespeelde partijen, hoogste TPN (laagst gerangschikt). De bye-speler wordt uit de pool verwijderd voordat brackets gevormd worden.

3. **Scoregroepen en brackets opbouwen** uit de overgebleven spelers.

4. **Ratingmap opbouwen** voor ARO-berekening.

5. **Brackets van boven naar beneden verwerken.** Voor elke bracket:
   - **Splitsen in G1 en G2.** In ronde 1 is G1 de eerste helft op TPN en G2 de rest. In latere rondes bevat G1 spelers die de voorkeur geven aan wit, G2 bevat spelers die de voorkeur geven aan zwart of geen voorkeur hebben. De groepen worden gebalanceerd zodat `|G1| = floor(n/2)`.
   - **G1 sorteren op oplopende ARO** (gelijken gebroken door oplopend TPN). Dit is het bepalende kenmerk van het Dubov-systeem: de speler met de laagste ARO in G1 wordt eerst ingedeeld.
   - **G2-transposities genereren** (tot 120 permutaties met het next-permutatie-algoritme van Narayana Pandita). Elke transpositie stelt opeenvolgende indelingen voor: G1[0] vs G2[0], G1[1] vs G2[1], enz.
   - **Elke transpositie evalueren** tegen absolute criteria (C1, C3, verboden paren). Als een paar een absoluut criterium schendt, wordt de hele transpositie verworpen. Geldige transposities worden gescoord op criteria C4-C10.
   - **De beste transpositie selecteren** door kandidaat-scores lexicografisch te vergelijken.

6. **Opwaarts samenvoegen bij falen.** Als een bracket alleen floaters oplevert en geen paren, wordt deze samengevoegd met de aangrenzende bracket en wordt de matching opnieuw geprobeerd.

7. **MaxT upfloater-limiet**: `2 + floor(CompletedRounds / 5)`. Dit beperkt hoe vaak een speler opwaarts gefloat kan worden voordat de engine verdere floats penaliseert.

8. **Bordvolgorde**: maximale spelerscore aflopend, bracket-score aflopend, minimale TPN oplopend.

9. **Kleurverdeling** via Dubov's 5-regels-algoritme (Art. 5). Dit delegeert naar de gedeelde `swisslib.AllocateColor` zonder top-scorer-specifieke regels.

### Details bye-selectie

De Dubov bye-regel verschilt van de completability-gebaseerde aanpak van het Dutch-systeem. In plaats van te analyseren welke spelerverwijdering tot de beste totale matching leidt, selecteert Dubov deterministisch:

1. Laagste score
2. Meeste gespeelde partijen (bij gelijke scores)
3. Hoogste TPN / laagste rang (laatste tiebreak)

Spelers die al een PAB hebben ontvangen, worden uitgesloten.

### Criteria

Het Dubov-systeem gebruikt 10 criteria, vergeleken met 21 in het Dutch-systeem:

| Criterium | Type      | Beschrijving                                                                               |
| --------- | --------- | ------------------------------------------------------------------------------------------ |
| C1        | Absoluut  | Geen herpartijen (spelers mogen niet al eerder tegen elkaar gespeeld hebben)               |
| C3        | Absoluut  | Geen absolute kleurconflicten (beide spelers absoluut dezelfde kleur nodig)                |
| C4        | Kwaliteit | Minimaliseer het aantal upfloaters                                                         |
| C5        | Kwaliteit | Maximaliseer de scoresom van upfloaters (prefereer het floaten van hoger scorende spelers) |
| C6        | Kwaliteit | Minimaliseer kleurvoorkeur-schendingen                                                     |
| C7        | Kwaliteit | Minimaliseer upfloaters op of boven MaxT                                                   |
| C8        | Kwaliteit | Minimaliseer opeenvolgende-ronde upfloaters                                                |
| C9        | Kwaliteit | Minimaliseer upfloater-tegenstanders op of boven MaxT                                      |
| C10       | Kwaliteit | Minimaliseer opeenvolgende-ronde MaxT-schendingen                                          |

Vergelijkingsvolgorde van kandidaten: C4, dan C5 (omgekeerd -- hoger is beter), dan C6-C10 lexicografisch, dan transpositie-index als laatste tiebreak.

### Fouten

| Fout                   | Voorwaarde                                                  |
| ---------------------- | ----------------------------------------------------------- |
| `ErrTooFewPlayers`     | Minder dan 1 actieve speler                                 |
| `ErrNoPairingPossible` | Geen geldige indeling mogelijk voor de overgebleven spelers |

## Vergelijking met Dutch

| Aspect           | Dubov                                          | Dutch                                          |
| ---------------- | ---------------------------------------------- | ---------------------------------------------- |
| Groepssplitsing  | G1/G2 op kleurvoorkeur                         | S1/S2 op TPN (bovenste helft / onderste helft) |
| G1-sortering     | Oplopende ARO, dan TPN                         | Aflopende score, dan oplopend TPN              |
| Matching         | Alleen transposities (max 120)                 | Globale Blossom-matching                       |
| Uitwisselingen   | Geen                                           | S1/S2-uitwisselingen toegestaan                |
| Criteria         | 10 (C1-C10)                                    | 21 (C1-C4, C8-C21)                             |
| Bye-selectie     | Deterministisch (laagste ARO in laagste groep) | Completability pre-matching                    |
| Versnelling      | Niet ondersteund                               | Baku-versnelling (C.04.7)                      |
| Upfloater-limiet | MaxT = 2 + floor(Rnds/5)                       | Blossom edge-gewichten regelen float-kosten    |

## Wiskundige grondslagen

Het kernmechanisme van het Dubov-systeem is eenvoudiger dan de Blossom-gebaseerde aanpak van het Dutch-systeem. In plaats van optimale gewogen matchings te berekenen, somt het G2-permutaties op en evalueert elke tegen een vaste criteria-hiërarchie.

- **ARO-berekening**: rekenkundig gemiddelde van de ratings van tegenstanders, exclusief forfaits. Zie `pairing/dubov/aro.go`.
- **Transpositie-generatie**: het algoritme van Narayana Pandita voor lexicografische next-permutatie, gemaximeerd op 120 transposities per bracket.
- **Criteria-scoring**: het type `DubovCandidateScore` implementeert een totale ordening over C4-C10-schendingen met lexicografische vergelijking.
- **Dubov criteria details**: [Dubov-criteria](../../algorithms/dubov-criteria/) behandelt de volledige C1-C10-specificatie.
- **Kleurverdeling**: [Kleurverdeling](../../algorithms/color-allocation/) beschrijft de gedeelde verdelingsregels die over Zwitserse varianten heen worden gebruikt.

## FIDE-referentie

Het Dubov-systeem is gedefinieerd in FIDE Handbook C.04.4.1. Het specificeert de ARO-gelijktrekkingsaanpak, G1/G2 kleurvoorkeur-splitsing, transpositie-gebaseerde matching, het 10-criteria evaluatiemodel en de MaxT upfloater-limietformule.
