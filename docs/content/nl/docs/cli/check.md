---
title: "check"
linkTitle: "check"
weight: 3
description: "Controleer of bestaande indelingen overeenkomen met de uitvoer van de engine."
---

Het `check`-subcommando controleert of de indelingen in de laatste ronde van een TRF-bestand overeenkomen met wat de opgegeven indelingsengine zou produceren. Het verwijdert de laatste ronde uit de toernooistatus, deelt opnieuw in met het opgegeven systeem en vergelijkt het resultaat met de bestaande indelingen.

Dit is handig om te valideren dat een toernooi correct is ingedeeld, of om engine-implementaties te testen tegen referentiebestanden.

## Synopsis

```text
chesspairing check SYSTEM input-file [options]
```

De indelingssysteemvlag (bijv. `--dutch`) is vereist en mag overal in de argumentenlijst staan. Het invoerbestand kan een bestandspad zijn of `-` voor stdin.

## Vergelijkingslogica

De vergelijking is volgorde-onafhankelijk (op basis van verzamelingen). Er wordt gecontroleerd dat:

- Dezelfde spelersparen voorkomen (ongeacht bordtoewijzing)
- Dezelfde bye-toewijzingen aanwezig zijn
- Het aantal indelingen overeenkomt

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

| Vlag     | Type | Standaard | Beschrijving             |
| -------- | ---- | --------- | ------------------------ |
| `--json` | bool | `false`   | Resultaat als JSON tonen |
| `--help` | --   | --        | Hulp tonen               |

## Voorbeelden

```bash
# Tekstuitvoer
chesspairing check --dutch tournament.trf

# Lees van stdin
chesspairing check --dutch - < tournament.trf

# JSON-uitvoer
chesspairing check --dutch tournament.trf --json
```

## Uitvoer

**Tekstformaat** (standaard):

- `OK: pairings match` wanneer het herhaalde indelingsresultaat overeenkomt.
- `MISMATCH: generated pairings differ from existing round` wanneer ze verschillen.

**JSON-formaat:**

```json
{
  "match": true,
  "system": "dutch",
  "round": 5
}
```

Het `round`-veld geeft aan welke ronde is gecontroleerd (de laatste ronde van het invoerbestand).

## Exitcodes

| Code | Betekenis                                                             |
| ---- | --------------------------------------------------------------------- |
| 0    | Indelingen komen overeen                                              |
| 1    | Indelingen komen niet overeen of indeling is mislukt                  |
| 3    | Ongeldige invoer (geen rondes, misvormd TRF, ontbrekende systeemvlag) |
| 5    | Bestand kon niet worden geopend                                       |

## Hoe het werkt

1. Parseer het TRF-bestand en converteer naar een `TournamentState`.
2. Sla de indelingen en byes van de laatste ronde op.
3. Verwijder de laatste ronde uit de toernooistatus en verlaag het huidige rondenummer.
4. Deel opnieuw in met de opgegeven engine en de indelingsopties van het toernooi.
5. Vergelijk de opnieuw gegenereerde indelingen met de opgeslagen laatste ronde (vergelijking op verzamelingsbasis van wit/zwart-paren en bye-toewijzingen).

## Zie ook

- [pair](../pair/) -- genereer indelingen voor de volgende ronde
- [generate](../generate/) -- genereer indelingen en produceer een bijgewerkt TRF
- [Legacy-modus](../legacy/) -- bbpPairings/JaVaFo-compatibele interface
