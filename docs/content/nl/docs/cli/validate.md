---
title: "validate"
linkTitle: "validate"
weight: 5
description: "Valideer een TRF-bestand tegen een van drie profielen."
---

Het `validate`-subcommando controleert een TRF16-bestand op structurele en semantische fouten. In tegenstelling tot de meeste andere subcommando's vereist `validate` geen indelingssysteemvlag -- het werkt puur op de TRF-documentstructuur.

Er zijn drie validatieprofielen beschikbaar, die elk strengere controles toevoegen bovenop het vorige niveau.

## Synopsis

```text
chesspairing validate input-file [options]
```

Het invoerbestand kan een bestandspad zijn of `-` voor stdin.

## Vlaggen

| Vlag        | Type   | Standaard  | Beschrijving                                      |
| ----------- | ------ | ---------- | ------------------------------------------------- |
| `--profile` | string | `standard` | Validatieprofiel: `minimal`, `standard`, `strict` |
| `--json`    | bool   | `false`    | Uitvoer als JSON                                  |
| `--help`    | --     | --         | Hulp tonen                                        |

## Profielen

**minimal** (`ValidateGeneral`): Basisstructuurvalidatie. Controleert of het TRF correct wordt geparsed en of verplichte velden aanwezig zijn.

**standard** (`ValidatePairingEngine`): Standaard. Voegt controles toe die nodig zijn voor indelingsengines -- volledigheid van spelersgegevens, consistentie van ronde-uitslagen, kruisverwijzingen tussen spelers en rondes.

**strict** (`ValidateFIDE`): Voegt FIDE-specifieke vereisten toe voor officiele toernooirapportage.

## Voorbeelden

```bash
# Standaard (standard-profiel, tekstuitvoer)
chesspairing validate tournament.trf

# Minimale validatie
chesspairing validate tournament.trf --profile minimal

# Strikte FIDE-validatie met JSON-uitvoer
chesspairing validate tournament.trf --profile strict --json

# Lees van stdin
cat tournament.trf | chesspairing validate -
```

## Uitvoer

### Tekstformaat (standaard)

Wanneer er problemen worden gevonden:

```text
tournament.trf: 2 errors, 1 warning

Errors:
  player.3.rating: rating must be a positive integer
  round.2.result: unknown result code

Warnings:
  player.5.title: title field is empty
```

Wanneer het bestand in orde is:

```text
tournament.trf: 0 errors, 0 warnings
```

### JSON-formaat

```json
{
  "valid": false,
  "errors": [
    {
      "field": "player.3.rating",
      "severity": "error",
      "message": "rating must be a positive integer"
    }
  ],
  "warnings": [
    {
      "field": "player.5.title",
      "severity": "warning",
      "message": "title field is empty"
    }
  ],
  "profile": "standard",
  "format": "auto"
}
```

Het `valid`-veld is alleen `true` wanneer er nul fouten zijn. Waarschuwingen hebben geen invloed op de geldigheid.

## Exitcodes

| Code | Betekenis                                                |
| ---- | -------------------------------------------------------- |
| 0    | Bestand is geldig (geen fouten; waarschuwingen mogelijk) |
| 2    | JSON-encodingfout                                        |
| 3    | Validatiefouten gevonden, of ongeldige invoer            |
| 5    | Bestandsfout                                             |

## Zie ook

- [convert](../convert/) -- herserialiseer een TRF-bestand
- [Uitvoerformaten en exitcodes](../output-formats/) -- gedetailleerde formaatspecificaties en alle exitcodes
