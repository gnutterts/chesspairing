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
| `topSeedColor`   | `string`     | `"auto"`  | Kleur van de top-seed in ronde 1. Waarden: `"auto"`, `"white"`, `"black"`. Bij `"auto"` krijgt de hoger gerangschikte speler op oneven borden wit en op even borden zwart. |
| `totalRounds`    | `int`        | nil       | Totaal aantal rondes. Wordt gebruikt om de laatste ronde te detecteren voor C8-criteria-versoepeling.                                                                      |
| `forbiddenPairs` | `[][]string` | nil       | Paren deelnemer-ID's die nooit tegen elkaar ingedeeld mogen worden.                                                                                                        |

## Hoe het werkt

Het indelingsalgoritme verloopt in vijf fasen:

### 1. Deelnemersstatussen opbouwen

`lexswiss.BuildParticipantStates` haalt de indelingsstatus op voor elke actieve deelnemer uit `TournamentState`. Indelingsscores gebruiken standaard 1-0.5-0 per partij, ongeacht het scoresysteem van het toernooi. Deelnemers worden gesorteerd op score aflopend, dan initiële rang oplopend, en krijgen een Tournament Pairing Number (TPN) toegekend die hun huidige positie weerspiegelt.

Forfait-partijen worden uitgesloten van de tegenstander-historie, wat betekent dat twee spelers die eerder ingedeeld werden in een forfait-partij opnieuw ingedeeld kunnen worden.

### 2. Indelings-toegekende bye toewijzen

Als het deelnemersaantal oneven is, selecteert `lexswiss.AssignPAB` de bye-ontvanger conform Art. 3.4. De bye gaat naar de in aanmerking komende deelnemer (iemand die nog geen PAB heeft ontvangen) met de laagste score, met als tiebreak het hoogste TPN. De PAB kent 1,5 punt toe (equivalent aan remise in een match).

### 3. Scoregroepen opbouwen

Deelnemers worden op score gegroepeerd in scoregroepen, gesorteerd in aflopende scorevolgorde. Binnen elke groep zijn deelnemers geordend op TPN oplopend.

### 4. Lexicografische bracket-indeling

Elke scoregroep wordt ingedeeld met een diepte-eerst-zoekopdracht die indelingen in lexicografische volgorde opsomt. De deelnemer met het laagste TPN wordt ingedeeld met de laagst-beschikbare TPN-partner. Als dit tot een dood spoor leidt waar resterende deelnemers niet allemaal ingedeeld kunnen worden, gaat het algoritme terug en probeert de volgende partner.

**Absolute criteria (altijd afgedwongen):**

- C1: Twee deelnemers spelen niet meer dan één keer tegen elkaar.
- Verboden paren worden nooit ingedeeld.

**Kwaliteitscriteria (C8 -- kleurvoorkeuren):**

- C8 controleert of twee deelnemers met twee opeenvolgende dezelfde-kleur Partij 1-toewijzingen die allebei de volgende ronde dezelfde kleur nodig hebben, niet tegen elkaar ingedeeld worden, aangezien één van hen dan noodzakelijkerwijs de 3-opeenvolgend-regel zou schenden.
- In de laatste ronde (wanneer `TotalRounds` is ingesteld en `CurrentRound >= TotalRounds`) wordt C8 volledig versoepeld.

Als een scoregroep een oneven aantal deelnemers heeft, wordt de laagst gerangschikte deelnemer opwaarts gefloat naar de bovenliggende scoregroep. De upfloater moet minimaal één compatibele tegenstander in de doelgroep hebben.

### 5. Kleurverdeling

`AllocateColor` implementeert de vijfstaps kleurverdelingsprocedure (Art. 4):

1. **Harde beperking**: Geen deelnemer speelt Partij 1 als dezelfde kleur drie keer op rij. Als een deelnemer de laatste twee rondes wit had in Partij 1, moet hij zwart krijgen.
2. **Gelijktrekken**: De deelnemer met meer witte Partij 1-toewijzingen krijgt zwart.
3. **Afwisselen**: De deelnemer die vorige ronde wit had in Partij 1 krijgt zwart.
4. **Ronde 1 bordafwisseling**: Oneven genummerde borden geven de hoger gerangschikte speler wit; even genummerde borden geven hen zwart. De `topSeedColor`-optie kan dit patroon omkeren.
5. **Rang-tiebreak**: De hoger gerangschikte speler (lager TPN) krijgt wit.

Na kleurverdeling worden borden gesorteerd op maximale score in het paar (aflopend), dan minimale TPN in het paar (oplopend).

## Vergelijking

| Aspect                  | Double-Swiss          | Dutch/Burstein        | Dubov/Lim               |
| ----------------------- | --------------------- | --------------------- | ----------------------- |
| Partijen per ronde      | 2 (minimatch)         | 1                     | 1                       |
| Matchingalgoritme       | Lexicografische DFS   | Globale Blossom       | Transpositie / Exchange |
| Aantal criteria         | 2 (C1 + C8)           | 21 (C1-C21)           | 10 (C1-C10)             |
| Kleurbeperking          | 3-opeenvolgend verbod | 3-opeenvolgend verbod | 3-opeenvolgend verbod   |
| PAB-waarde              | 1,5 punt              | 1 punt                | 1 punt                  |
| Gedeelde infrastructuur | `pairing/lexswiss`    | `pairing/swisslib`    | `pairing/swisslib`      |

De lexicografische aanpak is eenvoudiger dan Blossom-matching: het vindt altijd de lexicografisch kleinste geldige indeling in plaats van een gewogen doel over alle brackets te optimaliseren. Dit maakt het algoritme makkelijker te verifiëren en deterministisch van nature, ten koste van het niet beschouwen van cross-bracket-optimalisatie.

## Wiskundige grondslagen

### Lexicografische opsomming

Gegeven n deelnemers gesorteerd op TPN, somt het algoritme indelingen op als een reeks paren `(p1, q1), (p2, q2), ...` waar `p_i < q_i` in TPN-volgorde en `p1 < p2 < ...`. De eerste geldige volledige indeling in deze lexicografische volgorde wordt geselecteerd.

De zoekopdracht is een diepte-eerst-doorloop met backtracking. Op elk niveau wordt de ongepaarde deelnemer met het laagste TPN vastgezet en wordt zijn partner in oplopende TPN-volgorde geprobeerd. Als geen partner tot een volledige indeling leidt, gaat het algoritme terug naar het vorige niveau.

### Complexiteit

Voor een scoregroep van grootte k is het worst-case aantal kandidaat-indelingen `(k-1)!! = (k-1) * (k-3) * ... * 1`. In de praktijk snoeien de C1-beperking (geen herhaalde tegenstanders) en verboden paren de zoekboom aanzienlijk. De DFS stopt bij de eerste geldige indeling, dus de gemiddelde prestatie is veel beter dan het worst-case scenario.

### Kleurverdelingsprioriteit

De vijfstaps prioriteitsketen vormt een strikte totale ordening over kleurtoewijzingen. Stap 1 (harde beperking) heeft vetokracht en kan elke voorkeur-gebaseerde stap overrulen. Stappen 2-5 worden als tiebreakers in volgorde toegepast.

## FIDE-referentie

- **Reglement**: FIDE C.04.5 (Double-Swiss-systeem)
- **Aangenomen**: oktober 2025
- **Van kracht**: februari 2026
- **Kernartikelen**: Art. 3.4 (PAB-toewijzing, 1,5 punt), Art. 3.5 (upfloater-selectie), Art. 3.6 (lexicografische bracket-indeling), Art. 4 (kleurverdeling)
