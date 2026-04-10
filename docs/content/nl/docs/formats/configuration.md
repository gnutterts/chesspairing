---
title: "Configuratieformaat"
linkTitle: "Configuratie"
weight: 4
description: "Het configuratie-mapformaat voor engine-fabrieken en het generate-subcommando."
---

## Engine-optiemap

Alle indelings- en score-engines accepteren een `map[string]any` voor configuratie. Deze map kan afkomstig zijn uit:

- **TRF-systeemspecifieke velden** -- geparseerd door `ToTournamentState()` uit XX-codes en TRF-2026-headers
- **JSON-configuratie** -- gedeserialiseerd uit een JSON-object
- **CLI-flags** -- toegewezen aan optiesleutels door het `standings`-subcommando

Sleutels zijn strings die overeenkomen met de Go-structveldnamen (camelCase). Waarden worden geparseerd met typeconversie via de helperfuncties `GetFloat64()`, `GetInt()`, `GetBool()` en `GetString()` uit `options_helpers.go`. `GetFloat64` accepteert `float64`-, `int`- en `int64`-waarden; `GetInt` accepteert `int`, `int64` en `float64`; `GetBool` en `GetString` accepteren alleen hun eigen types.

### Indelingsoptiesleutels

Opties die gemeenschappelijk zijn voor meerdere Zwitserse indelingsengines:

| Sleutel          | Type       | Beschrijving                                          | Gebruikt door   |
| ---------------- | ---------- | ----------------------------------------------------- | --------------- |
| `totalRounds`    | int        | Totaal gepland aantal ronden                          | Alle Zwitsers   |
| `topSeedColor`   | string     | Beginkleur voor de topseed                            | Alle Zwitsers   |
| `forbiddenPairs` | `[][2]int` | Paren startnummers die niet tegen elkaar mogen spelen | Alle Zwitsers   |
| `acceleration`   | string     | Acceleratietype (`"baku"`)                            | Dutch, Burstein |

Round-Robin-opties:

| Sleutel             | Type | Beschrijving                                           |
| ------------------- | ---- | ------------------------------------------------------ |
| `cycles`            | int  | Aantal cycli (`1` = enkel, `2` = dubbel)               |
| `colorBalance`      | bool | Kleurbalancering inschakelen                           |
| `swapLastTwoRounds` | bool | Laatste twee ronden verwisselen bij dubbel round-robin |

Lim-opties:

| Sleutel          | Type | Beschrijving                  |
| ---------------- | ---- | ----------------------------- |
| `maxiTournament` | bool | Maxitoernooimodus inschakelen |

Team-Zwitserse opties:

| Sleutel               | Type   | Beschrijving                                     |
| --------------------- | ------ | ------------------------------------------------ |
| `colorPreferenceType` | string | Kleurvoorkeuralgoritme: `"A"`, `"B"` of `"none"` |
| `primaryScore`        | string | Primaire scoremaatstaf: `"match"` of `"game"`    |

Keizer-opties:

| Sleutel                   | Type | Beschrijving                              |
| ------------------------- | ---- | ----------------------------------------- |
| `allowRepeatPairings`     | bool | Herhaalde indelingen toestaan               |
| `minRoundsBetweenRepeats` | int  | Minimaal aantal ronden tussen herhalingen |

### Score-optiesleutels

Opties voor standaard- en voetbalscoring:

| Sleutel            | Type    | Standaard (Standard) | Standaard (Football) | Beschrijving               |
| ------------------ | ------- | -------------------- | -------------------- | -------------------------- |
| `pointWin`         | float64 | `1.0`                | `3.0`                | Punten voor winst          |
| `pointDraw`        | float64 | `0.5`                | `1.0`                | Punten voor remise         |
| `pointLoss`        | float64 | `0.0`                | `0.0`                | Punten voor verlies        |
| `pointBye`         | float64 | `1.0`                | `3.0`                | Punten voor PAB            |
| `pointForfeitWin`  | float64 | `1.0`                | `3.0`                | Punten voor forfaitwinst   |
| `pointForfeitLoss` | float64 | `0.0`                | `0.0`                | Punten voor forfaitverlies |
| `pointAbsent`      | float64 | `0.0`                | `0.0`                | Punten voor afwezigheid    |

### Koppeling van CLI-flags aan optiesleutels

Het `standings`-subcommando koppelt zijn flags aan score-optiesleutels:

| CLI-flag         | Optiesleutel       |
| ---------------- | ------------------ |
| `--win`          | `pointWin`         |
| `--draw`         | `pointDraw`        |
| `--loss`         | `pointLoss`        |
| `--forfeit-win`  | `pointForfeitWin`  |
| `--bye`          | `pointBye`         |
| `--forfeit-loss` | `pointForfeitLoss` |

Flagwaarden van `-1` (de standaardwaarde) worden als niet-ingesteld beschouwd en overschrijven de engine-standaardwaarden niet.

## RTG-configuratiebestand

Het `generate`-subcommando accepteert een configuratiebestand via `--config` voor de Random Tournament Generator. Het bestand gebruikt een eenvoudig `key=value`-formaat met `#`-commentaar:

```text
# RTG configuration
PlayersNumber=100
RoundsNumber=11
DrawPercentage=25
HighestRating=2800
LowestRating=1200
```

### RTG-configuratiesleutels

| Sleutel                | Type    | Standaard | Beschrijving                                                                   |
| ---------------------- | ------- | --------- | ------------------------------------------------------------------------------ |
| `PlayersNumber`        | int     | `30`      | Aantal te genereren spelers                                                    |
| `RoundsNumber`         | int     | `9`       | Aantal te simuleren ronden                                                     |
| `DrawPercentage`       | int     | `30`      | Remisekans als percentage (0-100)                                              |
| `ForfeitRate`          | int     | `20`      | Forfaitzeldzaamheidsfactor (hoger = zeldzamer; kans is `1 - sqrt(1 - 1/rate)`) |
| `RetiredRate`          | int     | `100`     | Zeldzaamheidsfactor voor terugtrekking                                         |
| `HalfPointByeRate`     | int     | `100`     | Zeldzaamheidsfactor voor halve-punt-bye                                        |
| `HighestRating`        | int     | `2600`    | Bovengrens voor willekeurige ratings                                           |
| `LowestRating`         | int     | `1400`    | Ondergrens voor willekeurige ratings                                           |
| `PointsForWin`         | float64 | `1.0`     | Toegekende punten voor winst                                                   |
| `PointsForDraw`        | float64 | `0.5`     | Toegekende punten voor remise                                                  |
| `PointsForLoss`        | float64 | `0.0`     | Toegekende punten voor verlies                                                 |
| `PointsForZPB`         | float64 | `0.5`     | Punten voor nulpunt-bye                                                        |
| `PointsForForfeitLoss` | float64 | `0.0`     | Punten voor forfaitverlies                                                     |
| `PointsForPAB`         | float64 | `0.0`     | Punten voor indelings-bye (pairing-allocated bye)                                |

Onbekende sleutels worden stilzwijgend genegeerd. Lege regels en regels die beginnen met `#` worden overgeslagen.

### Resultaatsimulatie

De RTG gebruikt een logistisch Elo-model voor resultaatsimulatie. Gegeven een ratingverschil `d = witRating - zwartRating`, is de verwachte score voor wit:

```text
E(white) = 1 / (1 + 10^(-d/400))
```

De remisekans wordt geschaald op basis van het geconfigureerde `DrawPercentage` en begrensd om consistent te blijven met de verwachte score. Forfaits worden per speler onafhankelijk bepaald via de `ForfeitRate`-parameter, voorafgaand aan de resultaatsimulatie.

### Seeding

De RTG zaait de PRNG via de `-s`-flag. Gehele getallen als string worden direct geparseerd als `int64`-seeds. Niet-gehele strings worden gehasht via FNV-1a om een deterministische seed te produceren. Als er geen seed wordt opgegeven, wordt een cryptografisch willekeurige seed gegenereerd en naar stderr geschreven voor reproduceerbaarheid.
