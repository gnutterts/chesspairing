---
title: "CLI Snelstart"
linkTitle: "CLI Snelstart"
weight: 2
description: "Installeer de chesspairing CLI en maak je eerste indeling in minder dan vijf minuten."
---

Op deze pagina doorlopen we het installeren van de `chesspairing`-opdrachtregeltool, het genereren van je eerste indeling vanuit een TRF16-bestand, en de beschikbare uitvoerformaten.

## Vereisten

Je hebt [Go](https://go.dev/dl/) 1.24 of hoger nodig. Controleer dit met:

```bash
go version
```

## Installatie

```bash
go install github.com/gnutterts/chesspairing/cmd/chesspairing@latest
```

Zorg dat `$GOPATH/bin` (of `$HOME/go/bin`) in je `PATH` staat. Bevestig de installatie:

```bash
chesspairing version
```

## Je eerste indeling

Het `pair`-subcommando leest een TRF16-toernooibestand en genereert de indeling voor de volgende ronde. Je moet een indelingssysteem opgeven met een vlag zoals `--dutch`, `--burstein`, `--dubov`, `--lim`, `--double-swiss`, `--team`, `--keizer` of `--roundrobin`.

Stel dat je een bestand `tournament.trf` hebt met drie voltooide ronden van een Zwitsers toernooi (Dutch). Genereer de indeling voor ronde 4:

```bash
chesspairing pair --dutch tournament.trf
```

De standaarduitvoer is het **list**-formaat -- een compact, machineleesbaar formaat dat compatibel is met bbpPairings en JaVaFo. De eerste regel bevat het aantal partijen, gevolgd door een `wit zwart`-paar per regel (startnummers):

```text
6
5 1
3 8
7 4
2 10
11 6
9 0
```

Een `0` rechts betekent een bye.

## Uitvoerformaten

Het `pair`-commando ondersteunt vijf uitvoerformaten via de `--format`-vlag.

### list (standaard)

```bash
chesspairing pair --dutch tournament.trf --format list
```

Compacte paarlijst zoals hierboven getoond. Dit is wat indelingssoftware verwacht.

### wide

```bash
chesspairing pair --dutch tournament.trf --format wide
```

Verkorte vorm: `-w`

Leesbare tabel met bordnummers, spelernamen, titels en ratings:

```text
Board  White                Rtg     Black                Rtg
-----  -----                ---     -----                ---
1      5 GM Carlsen, Magnus  2830  -  1 GM Caruana, Fabiano  2786
2      3 IM Doe, Jane        2412  -  8 FM Smith, John       2350
...
```

### board

```bash
chesspairing pair --dutch tournament.trf --format board
```

Genummerde bordweergave:

```text
Board 1:  5 -  1
Board 2:  3 -  8
Board 3:  7 -  4
```

### json

```bash
chesspairing pair --dutch tournament.trf --format json
```

Verkorte vorm: `--json`

Gestructureerde JSON met bordnummers en bye-details:

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 8 }
  ],
  "byes": [{ "player": 9, "type": "PAB" }]
}
```

### xml

```bash
chesspairing pair --dutch tournament.trf --format xml
```

XML met spelersgegevens (namen, ratings, titels) op elk bordelement.

## Uitvoer naar een bestand schrijven

Gebruik `-o` om de indeling naar een bestand te schrijven in plaats van stdout:

```bash
chesspairing pair --dutch tournament.trf -o round4.txt
```

## Lezen vanaf stdin

Geef `-` als bestandsnaam om het TRF-bestand via stdin in te lezen:

```bash
cat tournament.trf | chesspairing pair --dutch -
```

## Overige commando's

Naast `pair` biedt de CLI diverse andere subcommando's. Voer elk uit met `--help` voor de volledige gebruiksaanwijzing.

| Commando      | Wat het doet                                                                       |
| ------------- | ---------------------------------------------------------------------------------- |
| `check`       | Maakt de indeling van de laatste ronde opnieuw en vergelijkt met de bestaande indeling |
| `standings`   | Berekent en toont de stand met configureerbare scoring en tiebreakers              |
| `validate`    | Valideert een TRF16-bestand tegen een profiel (minimal, standard of strict)        |
| `generate`    | Genereert een willekeurig toernooi (bbpPairings RTG-compatibel)                    |
| `convert`     | Converteert tussen TRF-bestandsformaten                                            |
| `tiebreakers` | Toont alle 25 beschikbare tiebreaker-algoritmes                                    |
| `version`     | Toont de versie en ondersteunde indelingssystemen                                    |

Snelle voorbeelden:

```bash
# Verifieer de indeling van de laatste ronde
chesspairing check --dutch tournament.trf

# Toon de stand met Buchholz en partijwinsten als tiebreakers
chesspairing standings --dutch tournament.trf --tiebreakers buchholz,wins

# Valideer een TRF-bestand met strikte FIDE-controles
chesspairing validate tournament.trf --profile strict
```

## Legacy-modus

Als je migreert vanuit bbpPairings of JaVaFo, ondersteunt `chesspairing` ook hun positionele-argumentinterface:

```bash
# Indelen (legacy)
chesspairing --dutch tournament.trf -p

# Controleren (legacy)
chesspairing --dutch tournament.trf -c
```

Legacy-modus wordt automatisch geactiveerd als het eerste argument een systeemvlag is in plaats van een subcommando. Zie de [Legacy-modus referentie](/docs/cli/legacy/) voor alle details.

## Exitcodes

Bij het scripten met `chesspairing` kun je de exitcode controleren om het resultaat te bepalen:

| Code | Betekenis                                              |
| ---- | ------------------------------------------------------ |
| 0    | Succes                                                 |
| 1    | Geen geldige indeling mogelijk                           |
| 2    | Onverwachte fout tijdens uitvoering                    |
| 3    | Ongeldige of verkeerd geformatteerde invoer            |
| 4    | Toernooiomvang overschrijdt implementatielimiet        |
| 5    | Bestand kon niet worden geopend, gelezen of geschreven |

## Volgende stappen

- [CLI Referentie](/docs/cli/) -- volledige documentatie voor elk subcommando en elke vlag
- [Uitvoerformaten & Exitcodes](/docs/cli/output-formats/) -- gedetailleerde formaatspecificaties
- [Go Library Snelstart](/docs/getting-started/go-quickstart/) -- gebruik chesspairing als Go-bibliotheek
