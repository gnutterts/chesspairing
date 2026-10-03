---
title: "Burstein-systeem"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 Burstein Zwitsers indelingssysteem."
---

Burstein volgt FIDE C.04.4.2 (van kracht vanaf 1 februari 2026).

De eerste `min(floor(totalRounds / 2), 4)` rondes zijn Dutch-seedingrondes
(artikel 1.6), ingedeeld door de Dutch-pairer zelf, inclusief de
kleurtoewijzing daarvan. Daarna worden spelers op indelingsscore gegroepeerd en
worden brackets van hoog naar laag verwerkt (artikel 1.9). Binnen een bracket
is de volgorde Buchholz, Sonneborn-Berger en vast TPN (artikelen 1.7 en 1.8).

De indelingsbye wordt vóór de brackets toegekend: geschikte kandidaten worden
beoordeeld op laagste score, meeste partijen aan het bord en vervolgens laagste
rang; de overige spelers moeten volledig indeelbaar zijn (artikel 3.1).
`ForbiddenPairs` is een bibliotheekoptie, geen onderdeel van C1; deze paren
worden ook toegepast bij de test of een kandidaat een volledige indeling
mogelijk maakt. Brackets vermijden rematches en gelijke absolute
kleurvoorkeuren (C1 en C3). Zij kiezen floatersets volgens C5--C8 en daarna
paren in de volgorde van artikel 4.3 (artikelen 3.2 en 4); dit is
bracketindeling, geen globale Blossom-matching.

Voor C7 is de kwaliteit van de volgende bracket het aantal paren dat zij nog
kan maken en daarna de scores van de floaters die zij verder omlaag zou sturen,
met C5 en C6 in beide vergelijkingen voldaan. Brackets met hoogstens tien
spelers enumereren de volgorde van artikel 4.3. Grotere brackets enumereren
floatersets en leggen voor elke floaterset de partners volgens 4.3 vast; een
gelijke stand tussen floatersets wordt met diezelfde volgorde beslecht. Een
bracket waarvan de enumeratie van floatersets meer dan 200000 kandidaten heeft,
faalt met
`ErrBracketTooLarge` in plaats van C5/C7 stilzwijgend over te slaan.

De behandeling van artikel 1.7.2 is een interpretatie. Een niet-gespeelde ronde
(een bye, forfait of ronde zonder registratie) telt als gespeeld tegen de speler
zelf met de geregistreerde punten. Het voordeel van een reeks opeenvolgende
nulpuntbyes die eindigt bij de laatst voltooide ronde gaat uitsluitend naar de
werkelijke tegenstanders van de speler aan het bord. In hun Buchholz en
Sonneborn-Berger telt de score van die speler voor elke bye in de reeks 0,5
hoger. Rondes zonder registratie gelden in deze reeks als nulpuntbyes. In de
eigen index van de speler, waarin niet-gespeelde rondes als partijen tegen
zichzelf tellen, wordt de geregistreerde score zonder de extra halve punten
gebruikt. Virtuele versnellingspunten zijn van de index uitgesloten, maar
bepalen wel de scoregroepen.

Kleuren volgen de artikelen 5.2.1--5.2.5, niet de Dutch-kleurcascade. Voor twee
spelers zonder gespeelde partijen geeft 5.2.1 de hoger gerangschikte speler de
beginkleur; `TopSeedColor` stelt die kleur zowel na als tijdens de
seedingrondes in. De pariteit is de TPN-pariteit onder alle spelers die aan het
toernooi hebben deelgenomen, inclusief de speler die in deze ronde de bye
krijgt. De overige regels bekijken verenigbare voorkeuren, voorkeursterkte en
kleurverschil, de meest recente tegengestelde kleuren en rangorde. "Meest
recente keer" in artikel 5.2.4 wordt gelezen over alleen gespeelde partijen,
van de meest recente partij terug uitgelijnd, zoals in de Dutch-pairer en de
General Handling Rules 3.4; deze lezing is niet onafhankelijk geverifieerd.

## Configuratie

### CLI

```bash
chesspairing pair --burstein tournament.trf
```

### Go API

```go
p := burstein.New(burstein.Options{TotalRounds: chesspairing.IntPtr(9)})
result, err := p.Pair(ctx, &state)
```

### Opties

| Optie | Beschrijving |
| --- | --- |
| `TotalRounds` | Gepland aantal rondes; bepaalt de seedingfase van artikel 1.6. |
| `Acceleration` | Optionele versnelling `"baku"`. |
| `TopSeedColor` | Beginkleur van de topspeler in seedingrondes en daarna voor artikel 5.2.1. |
| `ForbiddenPairs` | Door de bibliotheek geconfigureerde speler-ID-paren die niet mogen spelen; geen C1-regel. |

### Fouten

| Fout | Voorwaarde |
| --- | --- |
| `ErrTooFewPlayers` | Het post-seedingveld is leeg (tenzij het alleen uit vooraf toegewezen byes bestaat). Eén resterende speler krijgt de bye. |
| `PairingNoPABCandidate` | Teruggegeven in een `chesspairing.PairingError` wanneer niemand de indelingsbye mag krijgen. |
| `ErrNoPairingPossible` | Teruggegeven in een `chesspairing.PairingError` met soort `PairingImpossible` wanneer geen volledige indeling bestaat. |
| `ErrBracketTooLarge` | Teruggegeven in een `chesspairing.PairingError` met soort `PairingImpossible` wanneer de floaterset-enumeratie van een bracket meer dan 200000 kandidaten heeft. |

De Burstein-modus van bbpPairings wordt niet als referentie gebruikt: die heeft
geen seedingrondes, gebruikt een andere indexvolgorde, behandelt niet-gespeelde
partijen als remises, slaat de stap met de meeste partijen bij de byekeuze over
en voegt uitgaande floaters samen; zij wijkt dus af van de gepubliceerde tekst.
Er is geen door FIDE onderschreven programma of openbare verzameling
geverifieerde Burstein-indelingen om mee te vergelijken.
