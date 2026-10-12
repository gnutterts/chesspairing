---
title: "Double-Swiss-systeem"
linkTitle: "Double-Swiss"
weight: 5
description: "FIDE C.04.5 — lexicografische indeling voor grote toernooien waar spelers twee keer tegen elkaar indelen."
---

Het Double-Swiss-systeem behandelt elke ronde als een tweepartijenmatch. Spelers worden ingedeeld met lexicografische bracket-indeling, en de kleuren wisselen binnen elke match zodat beide spelers één partij met wit en één met zwart spelen. Goedgekeurd door de FIDE in oktober 2025 en van kracht vanaf februari 2026, is het ontworpen voor grote open toernooien waar een Zwitsers formaat profiteert van de hogere statistische betrouwbaarheid van minimatches.

## Wanneer gebruiken

Double-Swiss is geschikt wanneer:

- Het toernooi genoeg rondes heeft ten opzichte van het aantal deelnemers voor een zinvolle Zwitserse indeling, maar je de variantie wilt verminderen die voortkomt uit enkelpartij-resultaten.
- Je wilt dat elke ronde vaker een beslissend matchresultaat oplevert, aangezien een tweepartijenmatch in winst kan eindigen zelfs als één partij remise wordt.
- Het evenement de langere speeltijd kan accommoderen die tweepartijmatches per ronde vereisen.

Het is minder geschikt wanneer de rondetijd beperkt is tot een enkele partij, of wanneer deelnemers het traditionele één-partij-per-ronde Zwitserse formaat verwachten.

## Configuratie

### CLI

```bash
chesspairing pair --double-swiss tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/doubleswiss"

// Met getypeerde opties
p := doubleswiss.New(doubleswiss.Options{
    TopSeedColor: chesspairing.StringPtr("auto"),
    TotalRounds:  chesspairing.IntPtr(9),
    ForbiddenPairs: [][]string{{"player-1", "player-2"}},
})

// Vanuit een generieke map (JSON-configuratie)
p := doubleswiss.NewFromMap(map[string]any{
    "topSeedColor": "auto",
    "totalRounds":  9,
})
```

### Opties-overzicht

| Optie            | Type         | Standaard | Beschrijving                                                                                                                                                               |
| ---------------- | ------------ | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `topSeedColor`   | `string`     | `"auto"`  | Beginkleur die voor ronde 1 wordt geloot. Waarden: `"auto"`, `"white"`, `"black"`; die krijgt een nieuwe hoger gerangschikte speler met een oneven TPN. |
| `totalRounds`    | `int`        | nil       | Totaal aantal rondes. Wordt gebruikt om de laatste ronde te detecteren voor C8-criteria-versoepeling.                                                                      |
| `forbiddenPairs` | `[][]string` | nil       | Paren deelnemer-ID's die nooit tegen elkaar ingedeeld mogen worden.                                                                                                        |
| `pabPoints`      | `float64`    | `1.5`     | Punten voor een indelings-bye; standaard de waarde van één winst en één remise (C.04.5 Art. 1.4).                                                                          |

## Hoe het werkt

Het indelingsalgoritme verloopt in vijf fasen:

### 1. Deelnemersstatussen opbouwen

`lexswiss.BuildParticipantStates` haalt de indelingsstatus op voor elke actieve deelnemer uit `TournamentState`. Indelingsscores volgen de ingestelde punten per partij, inclusief forfaits, byes en afwezigheden; afgebroken partijen tellen als remise. Het toernooi-indelingsnummer (TPN) wordt in de beginvolgorde vastgelegd (art. 1.1); voor de indeling worden deelnemers op score aflopend en daarna op TPN oplopend gezet (art. 1.2). Het nummer wordt niet per ronde opnieuw toegekend.

De aanroeper moet een match als twee `GameData`-items in dezelfde ronde vastleggen: partij 1 met de indelingskleuren en partij 2 met omgekeerde kleuren. Eén item blijft ondersteund als een match van één partij. De kleurhistorie bevat één item per match: de geplande kleur van partij 1, alleen als ten minste één partij is gespeeld. Een match is alleen forfait als één speler alle vastgelegde partijen forfait geeft; één forfaitpartij maakt de match voor C1, C2 en gespeelde matches dus gewoon gespeeld. Een partij wordt als forfait behandeld wanneer het resultaat een forfait is of de vlag `IsForfeit` is ingesteld.

### 2. Indelings-toegekende bye toewijzen

Als het deelnemersaantal oneven is, selecteert `lexswiss.AssignPAB` de bye-ontvanger conform art. 3.4. Van de volgens C2 in aanmerking komende deelnemers (geen eerdere PAB, forfaitwinst of verouderde bye met een vol punt) blijven eerst alleen degenen over van wie het weglaten een volledige geldige C1-indeling mogelijk maakt. Daarna gelden achtereenvolgens de laagste score, het grootste aantal werkelijk gespeelde matches en het grootste vaste TPN. Een forfaitmatch is niet gespeeld; een Dubbel-Zwitserse match met één forfaitpartij is wel gespeeld. De PAB kent één winst plus één remise toe (standaard 1,5 punt, bij voetbalscoring 4), tenzij `pabPoints` is ingesteld.

### 3. Scoregroepen opbouwen

Deelnemers worden op score gegroepeerd in scoregroepen, gesorteerd in aflopende scorevolgorde. Binnen elke groep zijn deelnemers geordend op TPN oplopend.

### 4. Lexicografische bracket-indeling

Elke scoregroep wordt ingedeeld in de identifier-volgorde van artikel 3.6. De zoekopdracht kiest eerst de gesorteerde verzameling topleden en wijst daarna de onderleden in oplopende TPN-volgorde toe. Een identifier bestaat uit de TPN's van de topleden, gevolgd door die van de bijbehorende onderleden; de eerste geldige volledige identifier wordt gekozen. In de eerste ronde met zes spelers is dat `1-4, 2-5, 3-6`.

**Absolute criteria (altijd afgedwongen):**

- C1: Twee deelnemers spelen niet meer dan één keer tegen elkaar.
- Verboden paren worden nooit ingedeeld.

**Kwaliteitscriterium (C8):**

- Behalve in de laatste ronde minimaliseert C8 het aantal upfloaters waarvan de tegenstander in de vorige ronde een floater was.
- Indelingen worden in identifier-volgorde bekeken; de eerste indeling met de minimale C8-telling wordt gekozen. C8 geldt niet in de laatste ronde (wanneer `TotalRounds` is ingesteld en `CurrentRound >= TotalRounds`).
- Kleur is geen indelingscriterium. Die wordt pas na de volledige ronde-indeling toegewezen.

Brackets worden volgens C.04.5 Art. 3.3.2 van de hoogste nog niet ingedeelde scoregroep naar beneden opgebouwd. Alle eigen deelnemers blijven in hun bracket. Sets upfloaters worden geprobeerd met eerst het kleinste werkende aantal (C4), daarna het beste scoreprofiel (C5), daarna C6, C7 en ten slotte de TPN-volgorde van artikel 3.5.4. Een reeks moet een geldige bracket volgens artikel 3.6 opleveren en voor de overige deelnemers een volledige indeling volgens C1/C2 overlaten. Mislukte brackets worden niet samengevoegd; een onmogelijke ronde wordt volgens artikel 3.3.3 gemeld.

### 5. Kleurverdeling

`AllocateColor` past artikel 4.3 na de indeling op elk paar toe:

1. Als beide spelers nog geen match hebben gespeeld, krijgt de hoger gerangschikte speler bij een oneven TPN de beginkleur en bij een even TPN de andere kleur.
2. Anders krijgt de speler met minder eerdere witpartijen wit.
3. Bij gelijkstand wisselen de kleuren vanaf de meest recente match waarin de spelers tegengestelde kleuren hadden; ongespeelde matches worden vóór de vergelijking verwijderd.
4. Bij verdere gelijkstand wisselt de kleur van de hoger gerangschikte speler ten opzichte van diens laatst gespeelde ronde.
5. Ten slotte wisselt de kleur van de tegenstander ten opzichte van diens laatst gespeelde ronde.

Na kleurverdeling volgen de borden C.04.2 artikel 3.6: de score van de hoger gerangschikte deelnemer (aflopend), de som van beide scores (aflopend) en diens TPN (oplopend). Binnen een paar bepalen de hoogste score en daarna het laagste TPN wie hoger gerangschikt is.

## Vergelijking

| Aspect                  | Double-Swiss          | Dutch/Burstein        | Dubov/Lim               |
| ----------------------- | --------------------- | --------------------- | ----------------------- |
| Partijen per ronde      | 2 (minimatch)         | 1                     | 1                       |
| Matchingalgoritme       | Identifiervolgorde + gewogen matching voor C8 | Globale Blossom       | Transpositie / Exchange |
| Aantal criteria         | C1, C2 en C8             | 21 (C1-C21)           | 10 (C1-C10)             |
| Kleurverdeling          | Na de indeling, art. 4   | 3-opeenvolgend verbod | 3-opeenvolgend verbod   |
| PAB-waarde              | 1,5 punt              | 1 punt                | 1 punt                  |
| Gedeelde infrastructuur | `pairing/lexswiss`    | `pairing/swisslib`    | `pairing/swisslib`      |

De lexicografische aanpak is eenvoudiger dan Blossom-matching: zij kiest de eerste identifier van artikel 3.6 onder de indelingen met de minimale C8-telling, in plaats van een gewogen doel over alle brackets te optimaliseren. Daardoor is het algoritme deterministisch, ten koste van het niet beschouwen van optimalisatie over brackets heen.

## Wiskundige grondslagen

### Lexicografische opsomming

In elk paar is het lagere TPN het toplid. De identifier bestaat uit de oplopende reeks TPN's van de topleden, gevolgd door het bijbehorende onderlid voor elk toplid. De zoekopdracht somt eerst verzamelingen topleden op in oplopende lexicografische volgorde en daarna toewijzingen van onderleden in oplopende lexicografische volgorde. Van de geldige volledige identifiers wordt de eerste met de minimale C8-telling gekozen; beginreeksen die niet tot een volledige indeling leiden, worden weggesnoeid.

### Complexiteit

Een bracket van grootte k heeft tot `(k-1)!! = (k-1) * (k-3) * ... * 1` indelingen; die worden dus nooit allemaal opgesomd. Zonder C8-kosten loopt het zoeken de identifiers af en snoeit het elke tak die onder C1 en de verboden paren niet af te maken is. Met C8 bepaalt eerst één gewogen matching het minimale C8-aantal; daarna wordt de indeling in identifiervolgorde opgebouwd en wordt elke stap gecontroleerd met een gewogen matching op de resterende spelers, zodat het werk polynomiaal blijft. Elke controle telt mee voor een zoeklimiet; is die op, dan mislukt de ronde met de gedeeltelijke indeling.

### Kleurverdelingsprioriteit

Artikel 4.3 is een volgorde met afnemende prioriteit: beginkleur voor twee nieuwe spelers, minder witpartijen, de meest recente tegengestelde kleuren, de laatste kleur van de hoger gerangschikte speler en ten slotte die van de tegenstander. De kleurverdeling volgt na de indeling en kan een indeling niet blokkeren.

## FIDE-referentie

- **Reglement**: FIDE C.04.5 (Double-Swiss-systeem)
- **Aangenomen**: oktober 2025
- **Van kracht**: februari 2026
- **Kernartikelen**: Art. 3.4 (PAB-toewijzing, 1,5 punt), Art. 3.5 (upfloater-selectie), Art. 3.6 (lexicografische bracket-indeling), Art. 4 (kleurverdeling)
