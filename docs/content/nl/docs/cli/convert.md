---
title: "convert"
linkTitle: "convert"
weight: 8
description: "Herserialiseer een TRF-bestand met genormaliseerde opmaak."
---

Het `convert`-subcommando leest een TRF16-bestand en schrijft het opnieuw weg, waarbij de veldvolgorde en opmaak worden genormaliseerd. Dit is handig voor het opschonen van handmatig bewerkte TRF-bestanden of het standaardiseren van uitvoer van andere tools.

## Synopsis

```text
chesspairing convert input-file -o output-file [options]
```

Geen indelingssysteemvlag vereist. Het invoerbestand kan een bestandspad zijn of `-` voor stdin.

## Opties

| Vlag           | Type   | Standaard | Beschrijving                              |
| -------------- | ------ | --------- | ----------------------------------------- |
| `-o`           | string | (vereist) | Pad naar uitvoerbestand                   |
| `--trf-format` | string | `trf2026` | Uitvoerformaat: `trf`, `trfbx`, `trf2026` |
| `--help`       | --     | --        | Hulp tonen                                |

Zowel `-o` als het invoerbestand zijn vereist. Als een van beide ontbreekt, wordt exitcode 3 gegeven.

## TRF-formaatvlag

De `--trf-format`-vlag accepteert drie waarden: `trf`, `trfbx` en `trf2026`. Onbekende waarden worden afgewezen met exitcode 3.

**Belangrijk:** Alleen `trf2026` wordt momenteel ondersteund. Het opgeven van `trf` of `trfbx` geeft een fout en stopt met exitcode 3:

```text
error: --trf-format FORMAT not yet supported
```

Deze formaatwaarden bestaan voor toekomstige compatibiliteit; alternatieve serializers worden in latere versies toegevoegd.

## Voorbeelden

```bash
# Normaliseer een TRF-bestand
chesspairing convert tournament.trf -o normalized.trf

# Lees van stdin
chesspairing convert - -o output.trf < tournament.trf

# Stel uitvoerformaat expliciet in (standaard)
chesspairing convert tournament.trf -o output.trf --trf-format trf2026
```

## Exitcodes

| Code | Betekenis                                                               |
| ---- | ----------------------------------------------------------------------- |
| 0    | Gelukt                                                                  |
| 2    | Schrijffout (TRF-serialisatie mislukt)                                  |
| 3    | Ongeldige invoer (ontbrekende argumenten, slecht TRF, onbekend formaat) |
| 5    | Bestandsfout (kan invoer niet openen of uitvoer niet aanmaken)          |

## Zie ook

- [validate](../validate/) -- valideer een TRF-bestand tegen een profiel
- [pair](../pair/) -- genereer indelingen vanuit een TRF-bestand
