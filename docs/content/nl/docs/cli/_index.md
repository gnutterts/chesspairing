---
title: "CLI-referentie"
linkTitle: "CLI-referentie"
weight: 60
description: "Volledige referentie voor de chesspairing-opdrachtregel — alle subcommando's, vlaggen en uitvoerformaten."
---

## Overzicht

Het `chesspairing`-programma verwerkt [TRF16](https://www.fide.com)-toernooibestanden en biedt functies voor indelen, validatie, scoring en andere hulpmiddelen. Alle subcommando's volgen een vast patroon: ze accepteren een TRF-bestand als invoer, voeren een bewerking uit, en schrijven het resultaat naar stdout (of naar een uitvoerbestand via `-o`).

Het algemene aanroeppatroon is:

```text
chesspairing COMMAND [SYSTEM] input.trf [options]
```

Hierin is `COMMAND` een van de onderstaande subcommando's, `SYSTEM` een indelingssysteemvlag (vereist bij sommige commando's) en `input.trf` het pad naar een TRF16-toernooibestand.

## Subcommando's

| Commando                        | Beschrijving                                         | Systeemvlag vereist |
| ------------------------------- | ---------------------------------------------------- | ------------------- |
| [pair](pair/)                   | Genereer indelingen voor de volgende ronde             | Ja                  |
| [check](check/)                 | Controleer of bestaande indelingen overeenkomen        | Ja                  |
| [generate](generate/)           | Genereer een volledig TRF met willekeurige uitslagen | Ja                  |
| [validate](validate/)           | Valideer de structuur van een TRF-bestand            | Nee                 |
| [standings](standings/)         | Bereken en toon de stand                             | Ja                  |
| [tiebreakers](tiebreakers-cmd/) | Toon beschikbare tiebreakers                         | Nee                 |
| [convert](convert/)             | Herserialiseer een TRF-bestand                       | Nee                 |
| [version](version/)             | Toon versie-informatie                               | Nee                 |

Elk subcommando heeft een eigen pagina met volledige gebruiksvoorbeelden en beschrijving van de vlaggen. Gebruik `chesspairing <command> --help` voor inline-hulp.

## Systeemvlaggen

Commando's die een indelingssysteem vereisen accepteren een van deze vlaggen voor het invoerbestand:

| Vlag             | Indelingssysteem                   |
| ---------------- | ------------------------------- |
| `--dutch`        | Dutch (FIDE C.04.3)             |
| `--burstein`     | Burstein (FIDE C.04.4.2)        |
| `--dubov`        | Dubov (FIDE C.04.4.1)           |
| `--lim`          | Lim (FIDE C.04.4.3)             |
| `--double-swiss` | Double-Swiss (FIDE C.04.5)      |
| `--team`         | Team Zwitsers (FIDE C.04.6)     |
| `--keizer`       | Keizer                          |
| `--roundrobin`   | Round-robin (FIDE C.05 Annex 1) |

De systeemvlag bepaalt welke indelingsengine (en bijbehorende standaard-scorer) wordt gebruikt. Wanneer een TRF-bestand systeemspecifieke `XX`-velden bevat, worden die opties automatisch doorgegeven aan de engine.

## Invoerverwerking

Invoerbestanden kunnen worden opgegeven als:

- Een bestandspad: `chesspairing pair --dutch tournament.trf`
- Een streepje (`-`) voor stdin: `cat tournament.trf | chesspairing pair --dutch -`

Als er geen invoerbestand wordt opgegeven, meldt de tool een fout en stopt met exitcode 3 (`ExitInvalidInput`).

## Uitvoer

De meeste commando's schrijven standaard naar stdout. Waar ondersteund kun je `-o` gebruiken om de uitvoer naar een bestand te sturen. De `--json`-vlag is beschikbaar bij de meeste commando's voor machineleesbare uitvoer. Het `pair`-subcommando ondersteunt daarnaast `--format` met de waarden `list`, `wide`, `board`, `xml` en `json`. Zie [Uitvoerformaten en exitcodes](output-formats/) voor details.

## Legacy-modus

Wanneer het programma wordt aangeroepen zonder een herkend subcommando, schakelt het over naar legacy-modus -- een interface met positionele argumenten die compatibel is met bbpPairings/JaVaFo. Hierdoor kan `chesspairing` als directe vervanging dienen in bestaande toolchains:

```bash
chesspairing --dutch input.trf -p
chesspairing --dutch input.trf -c
```

Zie [Legacy-modus](legacy/) voor de volledige interface met positionele argumenten.

## Exitcodes

| Code | Constante          | Betekenis                                              |
| ---- | ------------------ | ------------------------------------------------------ |
| 0    | `ExitSuccess`      | Bewerking succesvol afgerond                           |
| 1    | `ExitNoPairing`    | Geen geldige indeling mogelijk                           |
| 2    | `ExitUnexpected`   | Onverwachte fout tijdens uitvoering                    |
| 3    | `ExitInvalidInput` | Ongeldige of misvormde invoer                          |
| 4    | `ExitSizeOverflow` | Toernooiomvang overschrijdt limieten                   |
| 5    | `ExitFileAccess`   | Bestand kon niet worden geopend, gelezen of geschreven |

Zie [Uitvoerformaten en exitcodes](output-formats/) voor gedetailleerde beschrijvingen.
