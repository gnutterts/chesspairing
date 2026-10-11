---
title: "Lim-systeem"
linkTitle: "Lim"
weight: 4
description: "FIDE C.04.4.3 — indeling van scoregroepen, floaters en uitwisselingsregels."
---

## Overzicht

Lim is het Zwitserse systeem uit FIDE C.04.4.3. Het paart de scoregroepen in
de voorgeschreven driefasenvolgorde en gebruikt de uitwisselingsprocedure van
artikel 4, geen globale matching. Compatibiliteit (art. 2.1) vereist een nieuwe
paar en een toegestane kleurtoewijzing: geen speler mag voor de derde keer
opeenvolgend dezelfde kleur krijgen of drie meer van één kleur hebben dan van
de andere.

## Wanneer gebruiken

Gebruik Lim wanneer het toernooi volgens C.04.4.3 wordt gespeeld, inclusief de
expliciete floater- en maxitoernooiregels. Versnelling wordt niet
ondersteund. Stel `TotalRounds` in wanneer dit bekend is, zodat de uitzondering
van artikel 6 in de laatste ronde wordt toegepast.

## Configuratie

### CLI

```bash
chesspairing pair --lim tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/lim"

p := lim.New(lim.Options{
    TopSeedColor:   chesspairing.StringPtr("auto"),
    MaxiTournament: chesspairing.BoolPtr(true),
    TotalRounds:    9,
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})
```

### Opties

| Optie | Type | Standaard | Beschrijving |
| --- | --- | --- | --- |
| `TopSeedColor` | `string` | `"auto"` | Kleur van de topseed in ronde 1: `"auto"`, `"white"` of `"black"`. |
| `ForbiddenPairs` | `[][]string` | `nil` | Speler-ID-paren die niet tegen elkaar mogen spelen. |
| `MaxiTournament` | `bool` | `false` | Past de 100-puntsbeperkingen van art. 3.2.3, 3.8 en 5.7 toe. |
| `TotalRounds` | `int` | `0` | Schakelt artikel 6 in wanneer de huidige ronde de laatste is. |

## Hoe het werkt

### Volgorde van scoregroepen en transfers

Na de indelings-toegekende bye (PAB) van art. 1.1 verwerkt Lim de scoregroepen volgens
art. 2.2: vanaf de hoogste groep omlaag tot vlak voor de mediaan, dan vanaf de
laagste groep omhoog, en ten slotte de mediaangroep. De mediaangroep wordt
neerwaarts ingedeeld.

Voordat een groep in helften wordt verdeeld, verplaatsen art. 2.3.1--2.3.3
spelers die geen geschikte tegenstander hebben: zij hebben iedereen in de groep al
ontmoet, kunnen geen toegestane kleur ontvangen, of moeten na twee gelijke
kleuren de andere kleur krijgen. Pas daarna maakt art. 2.3.4 een groep even.
Een verplaatste speler is een floater.

### Floaterselectie en het indelen van floaters

Art. 3.2--3.4 kiezen een downfloater naar de volgende lagere groep en een
upfloater naar de volgende hogere groep. Verschuldigde kleuren zorgen er eerst
voor dat de aantallen van beide kleuren gelijk zijn. De dwingende formulering in 3.3 ("the player chosen
is") en 3.5 ("shall ... be exchanged") wordt toegepast; binnen die eis ordent
het "should" van 3.9.2 een keuze van minst naar meest nadelig: D, B, C, A. De
typen zijn respectievelijk: niet eerder gefloat met een compatibele
tegenstander in de aangrenzende groep; eerder gefloat met zo'n tegenstander;
niet eerder gefloat zonder zo'n tegenstander; en eerder gefloat zonder zo'n
tegenstander.

Artikel 3.10 verhindert dat een speler die in de direct voorafgaande ronde
floater was, onder 2.3.4 opnieuw wordt gekozen wanneer 3.10.1 geen floater van
type A, B of C veroorzaakt en 3.10.2 het aantal paren niet vermindert. Een
floater uit de vorige ronde is alleen de speler die tussen scoregroepen is
verplaatst; een bye is geen float. De beperking geldt aan dezelfde kant
(neerwaarts of opwaarts) van de scoregroeproute.

Inkomende floaters worden volgens art. 3.6--3.8 gepaard: downfloaters eerst in
een bovenste groep of de mediaangroep, upfloaters eerst in een onderste groep,
met de genoemde voorrangsregels voor score en startnummer. Een downfloater krijgt de hoogst
genummerde beschikbare tegenstander die de andere kleur verschuldigd is; een
upfloater de laagst genummerde. In een maxitoernooi vergelijken art. 3.8 en
5.7 de ratings van de twee tegenstanders die worden uitgewisseld; het verschil
mag hoogstens 100 punten zijn.

### Uitwisseling, cracking en kleuren

De overige spelers worden gepaard volgens het voorstel van art. 2.4 — de ene
helft tegen de andere — en daarna volgens de uitwisselingsprocedure van
artikel 4. Zie [Lim-uitwisselingsalgoritme](../../algorithms/lim-exchange/) voor art.
4.1--4.4, met 4.4.1 en 4.4.2.

Als de mediaangroep geblokkeerd is, breekt art. 2.6 — als er meer floaters van
boven zijn — het volgende paar in de lagere groep open, anders het volgende
paar in de hogere groep, en begint de mediaanindeling opnieuw. Als geen
Lim-indeling mogelijk is, retourneert de pairer `PairingImpossible`, of
`PairingNoPABCandidate` wanneer geen speler voor de indelings-toegekende bye in
aanmerking komt; hij retourneert geen gedeeltelijke indelingen met extra byes.

Kleuren volgen artikel 5. De tweede toetsing van artikel 5.2 ("a second
scrutiny with exchanges where necessary") wisselt tegenstanders uit tussen
paren van de scoregroep om de spelers hun verschuldigde kleuren te geven: een
paar met een kleurconflict wordt eerst tegen de latere paren getoetst en
daarna tegen de eerdere paren, en er wordt alleen gewisseld als dat het aantal
kleurconflicten verlaagt en beide nieuwe paren compatibel houdt (in een
maxitoernooi mogen de uitgewisselde tegenstanders hoogstens 100 punten
verschillen, art. 5.7). Dat de latere paren eerst worden getoetst, is een
keuze van de bibliotheek; de tekst legt geen volgorde vast. In de laatste
ronde geeft artikel 6 een nieuw paar binnen dezelfde score voorrang boven
afwisseling en egalisatie; dit vereist `TotalRounds`.

## Interpretaties

De FIDE-tekst beslist niet over de volgende lezingen van de implementatie:

1. Als niemand de mediaanscore heeft, vormen de floaters die van boven en van
onderen binnenkomen de mediaangroep op die positie. Die groep wordt volgens
2.2 en 3.6.3 als laatste gepaard, en voor die groep geldt 2.6.
2. Artikel 4.4.1 wordt toegepast op de speler voor wie het indelen mislukt en,
zodra in de scoregroep een 2.3.4-floater is gekozen, ook wanneer 2.3.1--2.3.3
een oorspronkelijk even groep oneven maakten.
3. In de mediaangroep worden de typen van 3.9 beoordeeld tegen de groep eronder,
die al is gepaard.
4. Onder artikel 3.8 heeft een speler zonder gespeelde partij geen
verschuldigde kleur en dus ook geen andere kleur verschuldigd. De
floater neemt de eerste beschikbare speler die de andere kleur verschuldigd
is; alleen als die er niet is, neemt die de eerste beschikbare compatibele
speler. De tekst zegt niet wat er dan gebeurt.

## FIDE-referentie

De implementatie volgt het FIDE Handbook C.04.4.3 Lim System (van kracht vanaf
1 februari 2026).