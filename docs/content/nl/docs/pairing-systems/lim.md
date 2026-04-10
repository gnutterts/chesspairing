---
title: "Lim-systeem"
linkTitle: "Lim"
weight: 4
description: "FIDE C.04.4.3 — mediaan-eerst verwerking met vier floatertypen en uitwisseling-matching."
---

## Overzicht

Het Lim-systeem is een Zwitserse variant die scoregroepen verwerkt vanaf de extremen richting de mediaan, in plaats van strikt van boven naar beneden. Het classificeert floaters in vier typen (A tot D) op basis van hun geschiedenis en compatibiliteit met aangrenzende groepen, en gebruikt uitwisseling-gebaseerde matching binnen elke scoregroep in plaats van Blossom of transpositie-matching.

Als speler is het meest merkbare verschil met het Dutch-systeem dat de middelste scoregroepen als laatste ingedeeld worden, waardoor de engine de meeste informatie over floaters heeft bij het verwerken van het meest bevolkte deel van het veld. Het systeem dwingt ook strikte compatibiliteitsregels af: geen drie opeenvolgende partijen met dezelfde kleur, en geen kleuronbalans van drie of meer.

## Wanneer gebruiken

- Toernooien waar je mediaan-eerst verwerking wilt voor betere floaterverdeling in het midden van de stand.
- Toernooien waar expliciete floaterclassificatie (typen A-D) transparantere indelingsbeslissingen oplevert.
- Maxiformaat-toernooien waar een 100-punts ratingbeperking op uitwisselingen en floaterselectie vereist is.
- Situaties waar uitwisseling-gebaseerde matching (in plaats van optimalisatie-gebaseerd) de voorkeur heeft voor controleerbaarheid.

Het Lim-systeem ondersteunt **geen** Baku-acceleratie of een instelbare `TotalRounds`-optie. Als je acceleratie nodig hebt, gebruik dan het [Dutch](../dutch/)- of [Burstein](../burstein/)-systeem.

## Configuratie

### CLI

```bash
chesspairing pair --lim tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/lim"

// Met getypeerde opties
p := lim.New(lim.Options{
    TopSeedColor:   chesspairing.StringPtr("auto"),
    MaxiTournament: chesspairing.BoolPtr(true),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// Vanuit een generieke map (bijv. geparsed uit JSON-configuratie)
p := lim.NewFromMap(map[string]any{
    "topSeedColor":   "white",
    "maxiTournament": true,
})
```

### Opties

| Optie            | Type         | Standaard | Beschrijving                                                                                            |
| ---------------- | ------------ | --------- | ------------------------------------------------------------------------------------------------------- |
| `TopSeedColor`   | `string`     | `"auto"`  | Kleur voor de topseed in ronde 1. Waarden: `"auto"`, `"white"`, `"black"`.                              |
| `ForbiddenPairs` | `[][]string` | `nil`     | Paren van speler-ID's die nooit tegen elkaar ingedeeld mogen worden.                                    |
| `MaxiTournament` | `bool`       | `false`   | Schakelt de 100-punts ratingbeperking in voor uitwisselingen en floaterselectie (Art. 3.2.3, 3.8, 5.7). |

Het Lim-systeem heeft geen `Acceleration`- of `TotalRounds`-opties.

## Hoe het werkt

### Stap-voor-stap algoritme

1. **Spelerstaten opbouwen** vanuit toernooigeschiedenis.

2. **Bye-selectie** (oneven spelaantal): De `LimByeSelector` kiest de laagst gerangschikte speler (hoogste TPN) in de laagste scoregroep die nog geen PAB heeft ontvangen. De bye-speler wordt uit de pool verwijderd.

3. **Scoregroepen opbouwen** uit de overgebleven spelers.

4. **Mediaanscore berekenen**: `roundsPlayed / 2.0`. Dit verdeelt het veld in boven-mediaan, onder-mediaan en mediaangroepen.

5. **Verwerkingsvolgorde bepalen** (Art. 2.2), het bepalende kenmerk van het Lim-systeem:
   - **Fase 1**: Hoogste scoregroep naar beneden tot net boven de mediaan.
   - **Fase 2**: Laagste scoregroep omhoog tot net onder de mediaan (omgekeerde richting).
   - **Fase 3**: Mediaangroep als laatste.

   Dit zorgt ervoor dat de mediaangroep -- doorgaans de grootste en meest beperkte -- verwerkt wordt met volledige kennis van alle floaters van boven en beneden.

6. **Voor elke scoregroep in verwerkingsvolgorde**:
   - **Inkomende floaters samenvoegen** in de groep. Floaters worden gesorteerd volgens Art. 3.6/3.7-prioriteit: down-floaters voor up-floaters in bovenste-helft-groepen, omgekeerd in onderste-helft-groepen.
   - **Bij oneven aantal, een floater selecteren** om door te geven aan de volgende groep. Selectie gebruikt `SelectDownFloater` (boven mediaan) of `SelectUpFloater` (onder mediaan), rekening houdend met floatertype, kleurgelijktrekking en compatibiliteit met de aangrenzende groep.
   - **Uitwisseling-matching** van de resterende even-aantal groep (Art. 4). Spelers worden gesplitst in bovenste helft (S1) en onderste helft (S2) op TPN, met voorgestelde indelingen S1[i] vs S2[i]. Incompatibele paren worden opgelost door de S2-partner uit te wisselen volgens de Art. 4.2-toetsingsvolgorde.
   - **Kleuruitwisselingspass** (Art. 5.2/5.7): na matching worden tegenstanders tussen paren gewisseld om kleurconflicten te verminderen, met inachtneming van compatibiliteitsbeperkingen. In maxitoernooien moeten de ratings van gewisselde spelers maximaal 100 punten verschillen.

7. **Resterende floaters indelen** over scoregroepgrenzen heen met greedy matching en herstelstrategieën (same-pair swap en chain swap).

8. **Bordvolgorde**: maximumscore van het paar aflopend, dan minimum-TPN oplopend.

9. **Kleurverdeling** via Lim-specifieke regels (Art. 5) met mediaan-bewuste tiebreaking: boven de mediaan wint de hoger gerangschikte speler kleurgelijkspelen; onder de mediaan wint de lager gerangschikte speler.

### Vier floatertypen

Het Lim-systeem classificeert elke floaterkandidaat op basis van twee factoren: of de speler al in de huidige groep is gevloat, en of de speler een compatibele tegenstander in de aangrenzende groep heeft.

| Type | Al gevloat? | Compatibele tegenstander in aangrenzende groep? | Prioriteit                 |
| ---- | ----------- | ----------------------------------------------- | -------------------------- |
| A    | Ja          | Nee                                             | Slechtst (meest benadeeld) |
| B    | Ja          | Ja                                              |                            |
| C    | Nee         | Nee                                             |                            |
| D    | Nee         | Ja                                              | Best (minst benadeeld)     |

Bij het selecteren van een floater prefereert de engine type D (minst benadeeld) als eerste. Dit minimaliseert de schade aan de speler die moet floaten, omdat deze de beste kans heeft om in de volgende groep ingedeeld te worden.

### Compatibiliteitsregels

Twee spelers zijn compatibel (Art. 2.1) als aan alle volgende voorwaarden voldaan wordt:

1. Ze hebben nog niet eerder tegen elkaar gespeeld.
2. Ze zijn geen verboden paar.
3. Er bestaat minstens een legale kleurtoewijzing waarbij geen van beide spelers:
   - Dezelfde kleur in drie opeenvolgende rondes zou hebben.
   - Een kleuronbalans van drie of meer zou hebben (bijv. 5 keer wit en 2 keer zwart).

### Details uitwisseling-matching

Het uitwisselingsalgoritme (Art. 4) werkt binnen een enkele scoregroep:

1. Splits spelers in S1 (lagere TPN's) en S2 (hogere TPN's).
2. Stel initiële indelingen voor: S1[0] vs S2[0], S1[1] vs S2[1], etc.
3. Toets elk paar. Bij indeling naar beneden begint de toetsing bij de hoogstgenummerde speler in S1. Bij indeling naar boven begint de toetsing bij de laagstgenummerde.
4. Probeer bij elke incompatibele indeling eerst S2-uitwisselingen (voorgestelde partner, dan resterende S2-spelers in uitwisselingsvolgorde), dan S1 cross-half partners.
5. Als volledige indeling mislukt, val terug op greedy matching.

### Maxitoernooi-modus

Wanneer `MaxiTournament` is ingeschakeld:

- **Floaterselectie** (Art. 3.2.3): als de rating van de geselecteerde floater meer dan 100 punten verschilt van de referentiespeler, wordt de referentiespeler (laagste TPN bij naar beneden floaten, hoogste TPN bij naar boven floaten) in plaats daarvan gekozen, waarbij de floatertype-prioriteit overschreven wordt.
- **Floater-tegenstanderselectie** (Art. 3.8): kandidaten wier rating meer dan 100 punten verschilt van de floater worden uitgesloten.
- **Kleuruitwisseling** (Art. 5.7): tegenstanders wisselen tussen paren is alleen toegestaan als de ratings van de gewisselde spelers maximaal 100 punten verschillen.

### Foutafhandeling

De Lim-engine retourneert geen sentinel errors. Wanneer indeling gedeeltelijk onmogelijk is, retourneert deze het beste gedeeltelijke resultaat met extra byes toegewezen aan spelers die niet ingedeeld konden worden.

## Vergelijking met Dutch

| Aspect               | Lim                                                                  | Dutch                                                  |
| -------------------- | -------------------------------------------------------------------- | ------------------------------------------------------ |
| Verwerkingsvolgorde  | Mediaan-eerst (hoog-naar-mediaan, laag-naar-mediaan, mediaan laatst) | Strikt top-down                                        |
| Matching             | Uitwisseling-gebaseerd (S1/S2 met Art. 4-toetsing)                   | Globale Blossom matching                               |
| Floaterclassificatie | 4 typen (A-D) met prioriteitsvolgorde                                | Geen expliciete typen; Blossom behandelt floaterkosten |
| Compatibiliteit      | Expliciete 3-opeenvolgend / 3-onbalans regels                        | Absolute kleurcriteria (C3)                            |
| Kleurverdeling       | Mediaan-bewuste tiebreaking (Art. 5.4)                               | Standaard swisslib-verdeling                           |
| Maxitoernooi         | 100-punts ratingbeperking op uitwisselingen                          | Niet ondersteund                                       |
| Acceleratie          | Niet ondersteund                                                     | Baku-acceleratie (C.04.7)                              |
| Bye-selectie         | Laagste rang in laagste groep                                        | Completability pre-matching                            |
| Aantal criteria      | Compatibiliteit-gebaseerd (geen genummerde kwaliteitscriteria)       | 21 criteria (C1-C4, C8-C21)                            |

## Wiskundige grondslagen

Het Lim-systeem gebruikt een deterministisch uitwisselingsalgoritme in plaats van optimalisatie-gebaseerde matching. De complexiteit per bracket is lager, maar de driefasen-verwerkingsvolgorde en floatertype-classificatie voegen structurele complexiteit toe.

- **Uitwisseling-matching**: [Lim Exchange-algoritme](../../algorithms/lim-exchange/) behandelt de Art. 4-uitwisselingsprocedure, toetsingsvolgorde en cross-half matching.
- **Kleurverdeling**: [Kleurverdeling](../../algorithms/color-allocation/) beschrijft de gedeelde verdelingsregels. Het Lim-systeem voegt daar mediaan-bewuste tiebreaking aan toe.
- **Floaterclassificatie**: de vier typen (A-D) zijn gedefinieerd in `pairing/lim/floater.go`. Classificatie hangt af van floatergeschiedenis en compatibiliteit met leden van de aangrenzende groep.
- **Compatibiliteitscontrole**: `pairing/lim/compatibility.go` implementeert de drie-opeenvolgend en drie-onbalans beperkingen met kleurgeschiedenisanalyse.

## FIDE-referentie

Het Lim-systeem is gedefinieerd in FIDE Handbook C.04.4.3. Het specificeert de mediaan-eerst verwerkingsvolgorde, vier floatertypen met prioriteitsselectie, uitwisseling-gebaseerde matching binnen scoregroepen, compatibiliteitsbeperkingen op kleurreeksen en onbalans, en de optionele maxitoernooi-ratingbeperking.
