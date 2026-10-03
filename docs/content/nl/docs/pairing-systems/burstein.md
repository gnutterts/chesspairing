---
title: "Burstein-systeem"
linkTitle: "Burstein"
weight: 2
description: "FIDE C.04.4.2 Burstein Zwitsers indelingssysteem."
---

Burstein volgt FIDE C.04.4.2 (van kracht vanaf 1 februari 2026).

De eerste `min(floor(totalRounds / 2), 4)` rondes zijn Dutch-seedingrondes
(artikel 1.6), met precies de Dutch-indeling en kleurverdeling. Daarna worden
spelers op indelingsscore gegroepeerd en worden brackets van hoog naar laag
verwerkt (artikel 1.9). Binnen een bracket is de volgorde Buchholz,
Sonneborn-Berger en vast TPN (artikelen 1.7 en 1.8).

De indelingsbye wordt vóór de brackets toegekend: geschikte kandidaten worden
beoordeeld op laagste score, meeste partijen aan het bord en vervolgens laagste
rang; de overige spelers moeten volledig indeelbaar zijn (artikel 3.1).
Brackets vermijden rematches en gelijke absolute kleurvoorkeuren (C1 en C3),
maximaliseren paren en kiezen floaters en paren in de volgorde van artikel 4.3
met C5--C8 (artikelen 3.2 en 4).

Niet-gespeelde rondes, inclusief forfaits, tellen voor de index als een partij
tegen de speler zelf (artikel 1.7.2). Virtuele versnellingspunten bepalen de
scoregroepen maar niet de index. Kleuren volgen artikel 5.2.

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
| `TopSeedColor` | Beginkleur van de topspeler in seedingrondes. |
| `ForbiddenPairs` | Speler-ID-paren die door C1 verboden zijn. |

### Fouten

| Fout | Voorwaarde |
| --- | --- |
| `ErrTooFewPlayers` | Minder dan twee actieve spelers zonder vooraf toegewezen bye. |
| `ErrNoPairingPossible` | De absolute criteria laten geen volledige indeling toe. |

De Burstein-modus van bbpPairings wordt niet als referentie gebruikt: die
beschrijft zichzelf als een gebrekkige implementatie van een eerdere versie en
wijkt van dit reglement af.
