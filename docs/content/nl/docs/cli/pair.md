---
title: "pair"
linkTitle: "pair"
weight: 2
description: "Lees een TRF-bestand en genereer indelingen voor de volgende ronde."
---

Het `pair`-subcommando leest een TRF16-bestand, voert de opgegeven indelingsengine uit en geeft de indelingen voor de volgende ronde als uitvoer. Dit is het primaire subcommando voor de meeste werkstromen.

## Synopsis

```text
chesspairing pair SYSTEM input-file [options]
```

De indelingssysteemvlag (bijv. `--dutch`) is vereist en mag overal in de argumentenlijst staan -- voor of na het invoerbestand. Het invoerbestand kan een bestandspad zijn of `-` voor stdin.

Vlaggen en positionele argumenten kunnen in willekeurige volgorde worden gemengd. Een kaal `--` beeindigt de vlagverwerking; alles erna wordt als positioneel behandeld.

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

De systeemvlag wordt verwerkt voordat andere vlaggen worden geparsed, en kan dus overal in de argumentenlijst staan.

## Opties

| Vlag       | Type   | Standaard | Beschrijving                                             |
| ---------- | ------ | --------- | -------------------------------------------------------- |
| `--format` | string | `list`    | Uitvoerformaat: `list`, `wide`, `board`, `xml`, `json`   |
| `-w`       | bool   | `false`   | Afkorting voor `--format wide`                           |
| `--json`   | bool   | `false`   | Afkorting voor `--format json` (achterwaarts compatibel) |
| `-o`       | string | stdout    | Schrijf uitvoer naar bestand in plaats van stdout        |
| `--help`   | --     | --        | Hulp tonen                                               |

**Prioriteit formaatkeuze:** `--format` > `-w` > `--json` > standaard `list`.

Als `--format` expliciet is ingesteld, worden `-w` en `--json` genegeerd. Als meerdere afkortingen worden opgegeven zonder `--format`, heeft `-w` voorrang op `--json`.

## Voorbeelden

```bash
# Standaard list-formaat (bbpPairings-compatibel)
chesspairing pair --dutch tournament.trf

# Breed tabelformaat met namen en ratings
chesspairing pair --dutch tournament.trf -w

# JSON-uitvoer naar bestand
chesspairing pair --dutch tournament.trf --format json -o pairings.json

# Bordformaat met het Lim-indelingssysteem
chesspairing pair --lim tournament.trf --format board

# XML-formaat
chesspairing pair --dutch tournament.trf --format xml

# Lees van stdin
cat tournament.trf | chesspairing pair --dutch -

# Vlaggen en invoerbestand kunnen worden gemengd
chesspairing pair tournament.trf --burstein -o result.txt --format wide
```

## Uitvoerformaten

### list (standaard)

Compact, machineleesbaar formaat compatibel met bbpPairings en JaVaFo. De eerste regel is het aantal bordindelingen. Elke volgende regel is een `wit zwart`-paar van startnummers. Byes worden na de indelingen vermeld als `speler 0`.

```text
5
5 1
3 12
8 2
6 9
4 11
7 0
```

### wide

Leesbare tabel met bordnummers, startnummers, titels, namen en ratings. Byes worden na het laatste bord vermeld.

```text
Board  White              Rtg   -  Black              Rtg
-----  -----              ---      -----              ---
1      5 GM Smith          2600  -  1 IM Jones         2500
2      3 WGM Lee           2400  -  12 FM Petrov       2350
       7 Brown             1800     Bye (PAB)
```

### board

Compacte genummerde bordlijst met alleen startnummers. Byes volgen na het laatste bord.

```text
Board  1:  5 -  1
Board  2:  3 - 12
Board  3:  8 -  2
Bye:  7
```

### json

Gestructureerde JSON met een `pairings`-array en een optionele `byes`-array. Bordnummers zijn 1-geindexeerd. Bye-types gebruiken hun stringrepresentatie (`PAB`, `Half`, `Zero`, `Absent`, `Excused`, `ClubCommitment`).

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 12 },
    { "board": 3, "white": 8, "black": 2 }
  ],
  "byes": [{ "player": 7, "type": "PAB" }]
}
```

De `byes`-sleutel wordt weggelaten wanneer er geen byes zijn.

### xml

XML-document met spelersmetadata (naam, rating, titel) op elk bordelement. Bevat een root-element `<pairings>` met attributen `round`, `boards` en `byes`.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<pairings round="4" boards="3" byes="1">
  <board number="1">
    <white number="5" name="Smith" rating="2600" title="GM"></white>
    <black number="1" name="Jones" rating="2500" title="IM"></black>
  </board>
  <bye number="7" name="Brown" type="PAB"></bye>
</pairings>
```

Zie [Uitvoerformaten en exitcodes](../output-formats/) voor volledige formaatspecificaties.

## Exitcodes

| Code | Betekenis                                                                  |
| ---- | -------------------------------------------------------------------------- |
| 0    | Indelingen succesvol gegenereerd                                             |
| 1    | Geen geldige indeling mogelijk                                               |
| 3    | Ongeldige invoer (misvormd TRF, onbekend formaat, ontbrekende systeemvlag) |
| 5    | Bestand kon niet worden geopend of geschreven                              |

## Opties van de indelingsengine

Elk indelingssysteem accepteert engine-specifieke opties via het `XXY`-extensieveld van het TRF-bestand. Deze opties regelen gedrag zoals Baku-acceleratie, kleur van de topgeplaatste speler, verboden paren en totaal aantal rondes. Raadpleeg de documentatie van elk [indelingssysteem](../../pairing-systems/) voor beschikbare opties.

## Zie ook

- [check](../check/) -- controleer bestaande indelingen tegen de engine-uitvoer
- [generate](../generate/) -- genereer indelingen en produceer een bijgewerkt TRF
- [Legacy-modus](../legacy/) -- bbpPairings/JaVaFo-compatibele interface
- [Uitvoerformaten en exitcodes](../output-formats/) -- gedetailleerde formaatspecificaties en alle exitcodes
