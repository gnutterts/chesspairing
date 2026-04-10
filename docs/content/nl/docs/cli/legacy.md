---
title: "Legacy-modus"
linkTitle: "Legacy-modus"
weight: 10
description: "Directe vervanging voor bbpPairings en JaVaFo — interface met positionele argumenten."
---

Legacy-modus biedt achterwaartse compatibiliteit met de opdrachtregelinterfaces van bbpPairings en JaVaFo. Het wordt automatisch geactiveerd wanneer het eerste argument geen herkend subcommando is (zoals `pair`, `check`, `generate`, `validate`, `standings`, `tiebreakers`, `convert` of `version`), waardoor `chesspairing` als directe vervanging kan dienen in bestaande scripts.

## Gebruikspatronen

```text
chesspairing SYSTEM input-file -p [output-file]    # indelen
chesspairing SYSTEM input-file -c                   # controleren
chesspairing SYSTEM -g [config-file] -o output      # genereren
```

De indelingssysteemvlag (bijv. `--dutch`) mag overal in de argumentenlijst staan. Vlaggen en positionele argumenten kunnen in willekeurige volgorde worden gemengd.

## Vlaggen

| Vlag | Beschrijving                                                                                                           |
| ---- | ---------------------------------------------------------------------------------------------------------------------- |
| `-p` | Indelingsmodus. Optioneel volgend argument is het uitvoerbestand.                                                      |
| `-c` | Controlemodus.                                                                                                         |
| `-g` | Generatiemodus. Optioneel volgend argument is een configuratiebestand.                                                 |
| `-o` | Uitvoerbestand (vereist voor generatiemodus).                                                                          |
| `-s` | PRNG-seed (voor generatiemodus).                                                                                       |
| `-r` | Toon versie. Kan alleen of in combinatie met een modusvlag worden gebruikt.                                            |
| `-w` | Breed uitvoerformaat (alleen indelingsmodus).                                                                          |
| `-q` | JaVaFo-compatibiliteitsvlag. Wordt geaccepteerd en genegeerd; een optioneel numeriek argument erna wordt ook verwerkt. |

Precies een indelingssysteemvlag is vereist (tenzij `-r` alleen wordt gebruikt):

| Vlag             | Systeem       |
| ---------------- | ------------- |
| `--dutch`        | Dutch         |
| `--burstein`     | Burstein      |
| `--dubov`        | Dubov         |
| `--lim`          | Lim           |
| `--double-swiss` | Double-Swiss  |
| `--team`         | Team Zwitsers |
| `--keizer`       | Keizer        |
| `--roundrobin`   | Round-Robin   |

## Modusafhandeling

### Indelen (`-p`)

Leest het invoer-TRF-bestand, voert de indelingsengine uit en schrijft het resultaat.

- Standaarduitvoer is `list`-formaat (bbpPairings-compatibel): aantal indelingen op de eerste regel, dan `wit zwart`-startnummerparen.
- Met `-w` wordt het `wide`-formaat gebruikt (tabel met namen en ratings).
- Schrijft naar het bestand opgegeven na `-p`, of naar stdout als er geen bestand volgt.

### Controleren (`-c`)

Verwijdert de laatste ronde uit het TRF, deelt opnieuw in en vergelijkt met de bestaande indelingen.

- Alleen tekstuitvoer: `OK: pairings match` of `MISMATCH: generated pairings differ from existing round`.
- Exitcode 0 bij overeenkomst, 1 bij verschil of mislukte indeling.

### Genereren (`-g`)

Delegeert intern naar het `generate`-subcommando. Het optionele configuratiebestand volgt na `-g`; het uitvoerbestand wordt opgegeven met `-o`.

### Versie (`-r`)

- `-r` alleen: toont versie-informatie en stopt.
- `-r` in combinatie met een modusvlag (bijv. `-r -p`): toont de versie, een lege regel en voert vervolgens de opgegeven modus uit.

## Verschillen met subcommando's

| Functie                  | Legacy-modus                  | Subcommando-equivalent                 |
| ------------------------ | ----------------------------- | -------------------------------------- |
| Indelingsuitvoerformaten | Alleen `list` en `wide`       | `list`, `wide`, `board`, `xml`, `json` |
| Check JSON-uitvoer       | Niet beschikbaar              | `--json`-vlag                          |
| Generate                 | Delegeert naar subcommando    | Volledige set vlaggen                  |
| Vlagverwerking           | Handmatige positionele parser | Go `flag`-pakket                       |

## Voorbeelden

```bash
# bbpPairings-compatibele indeling
chesspairing --dutch tournament.trf -p

# Schrijf indelingen naar bestand, breed formaat
chesspairing --dutch tournament.trf -p pairings.txt -w

# Controleer indelingen
chesspairing --dutch tournament.trf -c

# Genereer met seed
chesspairing --dutch -g config.txt -o output.trf -s 42

# Versie
chesspairing -r
```

## Exitcodes

| Code | Betekenis                                              |
| ---- | ------------------------------------------------------ |
| 0    | Gelukt (of indelingen komen overeen in controlemodus)  |
| 1    | Geen geldige indeling of indelingen komen niet overeen |
| 3    | Ongeldige invoer of ontbrekende vereiste vlaggen       |
| 5    | Bestandsfout                                           |

## Zie ook

- [pair](../pair/) -- volledig subcommando met alle uitvoerformaten
- [check](../check/) -- volledig subcommando met JSON-uitvoer
- [generate](../generate/) -- volledig subcommando voor TRF-generatie
