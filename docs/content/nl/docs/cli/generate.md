---
title: "generate"
linkTitle: "generate"
weight: 4
description: "Genereer een volledig toernooi-TRF-bestand met willekeurige uitslagen voor testdoeleinden."
---

Het `generate`-subcommando maakt een synthetisch toernooi-TRF-bestand door willekeurige spelers te genereren, elke ronde te indelen met de opgegeven engine en willekeurige uitslagen te simuleren. Dit is voornamelijk een test- en benchmarktool.

Het gegenereerde toernooi gebruikt realistische Elo-gebaseerde uitslagsimulatie: de winstkans volgt het logistische model (`1 / (1 + 10^(-diff/400))`), met instelbare remise- en forfaitpercentages.

In tegenstelling tot `pair` en `check` leest dit subcommando geen bestaand TRF-bestand. Het maakt een nieuw toernooi vanaf nul en vereist `-o` voor het uitvoerbestand.

## Synopsis

```text
chesspairing generate SYSTEM -o output-file [options]
```

De indelingssysteemvlag (bijv. `--dutch`) is vereist en mag overal in de argumentenlijst staan. De `-o`-vlag voor het uitvoer-TRF-bestand is ook vereist.

## Indelingssysteemvlaggen

Precies een systeemvlag is vereist:

| Vlag             | Systeem                         |
| ---------------- | ------------------------------- |
| `--dutch`        | Dutch (FIDE C.04.3)             |
| `--burstein`     | Burstein (FIDE C.04.4.2)        |
| `--dubov`        | Dubov (FIDE C.04.4.1)           |
| `--lim`          | Lim (FIDE C.04.4.3)             |
| `--double-swiss` | Double-Swiss (FIDE C.04.5)      |
| `--team`         | Team Zwitsers (FIDE C.04.6)     |
| `--keizer`       | Keizer                          |
| `--roundrobin`   | Round-Robin (FIDE C.05 Annex 1) |

## Opties

| Vlag       | Type   | Standaard | Beschrijving                      |
| ---------- | ------ | --------- | --------------------------------- |
| `-o`       | string | --        | Pad naar uitvoer-TRF (vereist)    |
| `--config` | string | --        | Pad naar RTG-configuratiebestand  |
| `-s`       | string | --        | PRNG-seed (geheel getal of tekst) |
| `--help`   | --     | --        | Hulp tonen                        |

## Configuratie

Wanneer `--config` wordt opgegeven, wordt het bestand geparsed als `key=value`-paren (een per regel, `#` voor commentaar, lege regels worden genegeerd).

| Sleutel                | Type    | Standaard | Beschrijving                                                               |
| ---------------------- | ------- | --------- | -------------------------------------------------------------------------- |
| `PlayersNumber`        | int     | 30        | Aantal spelers                                                             |
| `RoundsNumber`         | int     | 9         | Aantal rondes                                                              |
| `DrawPercentage`       | int     | 30        | Remisekans (%)                                                             |
| `ForfeitRate`          | int     | 20        | Inverse forfaitkans (hoger = minder forfaits)                              |
| `RetiredRate`          | int     | 100       | Terugtrekkingspercentage (gereserveerd, nog niet geïmplementeerd)          |
| `HalfPointByeRate`     | int     | 100       | Halve-punt-bye-aanvraagpercentage (gereserveerd, nog niet geïmplementeerd) |
| `HighestRating`        | int     | 2600      | Maximale rating                                                            |
| `LowestRating`         | int     | 1400      | Minimale rating                                                            |
| `PointsForWin`         | float64 | 1.0       | Punten voor winst (gereserveerd, nog niet geïmplementeerd)                 |
| `PointsForDraw`        | float64 | 0.5       | Punten voor remise (gereserveerd, nog niet geïmplementeerd)                |
| `PointsForLoss`        | float64 | 0.0       | Punten voor verlies (gereserveerd, nog niet geïmplementeerd)               |
| `PointsForZPB`         | float64 | 0.5       | Punten voor nulpunt-bye (gereserveerd, nog niet geïmplementeerd)           |
| `PointsForForfeitLoss` | float64 | 0.0       | Punten voor forfaitverlies (gereserveerd, nog niet geïmplementeerd)        |
| `PointsForPAB`         | float64 | 0.0       | Punten voor indelings-bye (gereserveerd, nog niet geïmplementeerd)         |

Voorbeeld configuratiebestand:

```text
# Large tournament
PlayersNumber=100
RoundsNumber=11
DrawPercentage=25
HighestRating=2700
LowestRating=1200
```

## Seed-gedrag

- Als `-s` een geheel getal is, wordt het direct als PRNG-seed gebruikt.
- Als `-s` een niet-numerieke tekst is, wordt deze met FNV-1a gehasht tot een 64-bit seed.
- Als `-s` wordt weggelaten, worden 8 willekeurige bytes van `crypto/rand` gebruikt en wordt de seed naar stderr geprint voor reproduceerbaarheid.

De PRNG is `math/rand/v2` met een PCG-bron.

## Voorbeelden

```bash
# Genereer een toernooi met 30 spelers en 9 rondes met standaardinstellingen
chesspairing generate --dutch -o tournament.trf

# Reproduceerbare generatie met een numerieke seed
chesspairing generate --dutch -o test.trf -s 42

# String-seed (gehasht via FNV-1a)
chesspairing generate --burstein -o large.trf -s myseed

# Met aangepaste configuratie
chesspairing generate --burstein -o large.trf --config rtg.conf -s myseed
```

## Spelergeneratie

Spelers krijgen willekeurige ratings, uniform verdeeld over [`LowestRating`, `HighestRating`], en worden gesorteerd op rating van hoog naar laag. De hoogst gewaardeerde speler krijgt startnummer 1. Namen worden gegenereerd als "Player N" waarbij N het startnummer is.

## Uitslagsimulatie

Voor elke ronde:

1. Converteer het huidige TRF-document naar een `TournamentState`.
2. Deel in met de opgegeven engine.
3. Controleer voor elke indeling eerst op forfaits -- de forfaitkans per speler is `sqrt(1 - 1/ForfeitRate)`. Als beide spelers forfait geven, is het een dubbel forfait; als er een forfait geeft, wint de ander door forfait.
4. Voor niet-forfaitpartijen wordt de uitslag bepaald met het logistische Elo-model. De remisekans wordt begrensd door `DrawPercentage`. De resterende kansmassa wordt verdeeld over wit-wint en zwart-wint, evenredig met de verwachte score.
5. Schrijf de uitslagen terug in het TRF-document.

## Exitcodes

| Code | Betekenis                                         |
| ---- | ------------------------------------------------- |
| 0    | Toernooi succesvol gegenereerd                    |
| 1    | Indeling mislukt voor een ronde                   |
| 2    | Onverwachte fout tijdens generatie                |
| 3    | Ongeldige invoer (foute config, ontbrekende `-o`) |
| 5    | Bestandsfout                                      |

## Zie ook

- [pair](../pair/) -- genereer indelingen voor een bestaand toernooi
- [Legacy-modus](../legacy/) -- bbpPairings/JaVaFo-compatibele interface
