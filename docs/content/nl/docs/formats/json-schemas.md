---
title: "JSON-schema's"
linkTitle: "JSON"
weight: 3
description: "JSON-uitvoerschema's voor indelingen, ranglijsten, validatie, versie en tiebreaker-overzicht."
---

De CLI produceert JSON-uitvoer voor verschillende subcommando's wanneer deze worden aangeroepen met `--json` of `--format json`. Alle JSON-uitvoer gebruikt inspringen met 2 spaties.

## Indelingen

Commando: `chesspairing pair SYSTEM input-file --format json`

```json
{
  "pairings": [
    { "board": 1, "white": 5, "black": 1 },
    { "board": 2, "white": 3, "black": 12 }
  ],
  "byes": [{ "player": 7, "type": "PAB" }]
}
```

### Velden

| Veld               | Type   | Beschrijving                           |
| ------------------ | ------ | -------------------------------------- |
| `pairings`         | array  | Bordtoewijzingen voor de ronde         |
| `pairings[].board` | int    | Bordnummer (1-gebaseerd)               |
| `pairings[].white` | int    | Startnummer van de witspeler           |
| `pairings[].black` | int    | Startnummer van de zwartspeler         |
| `byes`             | array  | Bye-toewijzingen (weggelaten als leeg) |
| `byes[].player`    | int    | Startnummer van de speler              |
| `byes[].type`      | string | Type bye als tekst                     |

Bye-typewaarden corresponderen met `ByeType.String()`:

| Waarde           | Betekenis                                     |
| ---------------- | --------------------------------------------- |
| `PAB`            | Indelings-bye (pairing-allocated bye, vol punt) |
| `Half`           | Halve-punt-bye                                |
| `Zero`           | Nulpunt-bye                                   |
| `Absent`         | Afwezig/niet ingedeeld                           |
| `Excused`        | Geexcuseerde afwezigheid                      |
| `ClubCommitment` | Afwezigheid wegens clubverplichting           |

## Ranglijst

Commando: `chesspairing standings SYSTEM input-file --json`

```json
{
  "standings": [
    {
      "rank": 1,
      "playerId": "5",
      "displayName": "GM Smith, John",
      "score": 7,
      "tieBreakers": [
        { "id": "buchholz-cut1", "name": "Buchholz Cut 1", "value": 28.5 },
        { "id": "buchholz", "name": "Buchholz", "value": 35 },
        { "id": "sonneborn-berger", "name": "Sonneborn-Berger", "value": 24.5 }
      ],
      "gamesPlayed": 9,
      "wins": 6,
      "draws": 2,
      "losses": 1
    }
  ],
  "scoring": "standard",
  "tiebreakers": ["buchholz-cut1", "buchholz", "sonneborn-berger"]
}
```

### Velden

| Veld                              | Type   | Beschrijving                                             |
| --------------------------------- | ------ | -------------------------------------------------------- |
| `standings`                       | array  | Gerangschikte spelersgegevens                            |
| `standings[].rank`                | int    | Rang (gedeelde rang bij gelijke stand)                   |
| `standings[].playerId`            | string | Speler-ID (startnummer als string)                       |
| `standings[].displayName`         | string | Weergavenaam van de speler                               |
| `standings[].score`               | float  | Totaalscore van de score-engine                          |
| `standings[].tieBreakers`         | array  | Tiebreaker-waarden in geconfigureerde volgorde           |
| `standings[].tieBreakers[].id`    | string | Tiebreaker-register-ID                                   |
| `standings[].tieBreakers[].name`  | string | Weergavenaam van de tiebreaker                           |
| `standings[].tieBreakers[].value` | float  | Berekende tiebreaker-waarde                              |
| `standings[].gamesPlayed`         | int    | Totaal gespeelde partijen                                |
| `standings[].wins`                | int    | Aantal overwinningen                                     |
| `standings[].draws`               | int    | Aantal remises                                           |
| `standings[].losses`              | int    | Aantal nederlagen                                        |
| `scoring`                         | string | Gebruikt scoresysteem (`standard`, `keizer`, `football`) |
| `tiebreakers`                     | array  | Geordende lijst van toegepaste tiebreaker-ID's           |

De ranglijst is gesorteerd op aflopende score, gevolgd door tiebreaker-waarden in volgorde. Spelers met identieke scores en tiebreaker-waarden delen dezelfde rang.

## Validatie

Commando: `chesspairing validate input-file --json`

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

### Velden

| Veld                  | Type   | Beschrijving                                               |
| --------------------- | ------ | ---------------------------------------------------------- |
| `valid`               | bool   | `true` als er geen fouten zijn (waarschuwingen zijn prima) |
| `errors`              | array  | Problemen op foutniveau (null als er geen zijn)            |
| `warnings`            | array  | Problemen op waarschuwingsniveau (null als er geen zijn)   |
| `errors[].field`      | string | Punt-gescheiden veldpad                                    |
| `errors[].severity`   | string | Altijd `"error"`                                           |
| `errors[].message`    | string | Leesbare beschrijving                                      |
| `warnings[].field`    | string | Punt-gescheiden veldpad                                    |
| `warnings[].severity` | string | Altijd `"warning"`                                         |
| `warnings[].message`  | string | Leesbare beschrijving                                      |
| `profile`             | string | Gebruikt validatieprofiel                                  |
| `format`              | string | Gedetecteerd of opgegeven formaat                          |

## Controle

Commando: `chesspairing check SYSTEM input-file --json`

```json
{
  "match": true,
  "system": "dutch",
  "round": 5
}
```

### Velden

| Veld     | Type   | Beschrijving                                                                  |
| -------- | ------ | ----------------------------------------------------------------------------- |
| `match`  | bool   | `true` als de opnieuw gegenereerde indelingen overeenkomen met de laatste ronde |
| `system` | string | Gebruikt indelingssysteem voor herindeling                                        |
| `round`  | int    | Rondenummer dat gecontroleerd is                                              |

Het `check`-subcommando verwijdert de laatste ronde, genereert de indelingen opnieuw en vergelijkt ze. Exitcode `0` betekent een overeenkomst; exitcode `1` een verschil.

## Versie

Commando: `chesspairing version --json`

```json
{
  "version": "dev",
  "pairingSystems": [
    "dutch",
    "burstein",
    "dubov",
    "lim",
    "doubleswiss",
    "team",
    "keizer",
    "roundrobin"
  ],
  "scoringSystems": ["standard", "keizer", "football"],
  "tiebreakers": [
    "aro",
    "avg-opponent-buchholz",
    "avg-opponent-ptp",
    "avg-opponent-tpr",
    "black-games",
    "black-wins",
    "buchholz",
    "buchholz-cut1",
    "buchholz-cut2",
    "buchholz-median",
    "buchholz-median2",
    "direct-encounter",
    "fore-buchholz",
    "games-played",
    "koya",
    "pairing-number",
    "performance-points",
    "performance-rating",
    "player-rating",
    "progressive",
    "rounds-played",
    "sonneborn-berger",
    "standard-points",
    "win",
    "wins"
  ]
}
```

### Velden

| Veld             | Type   | Beschrijving                                                 |
| ---------------- | ------ | ------------------------------------------------------------ |
| `version`        | string | Buildversiestring                                            |
| `pairingSystems` | array  | Alle ondersteunde indelingssysteem-identifiers                 |
| `scoringSystems` | array  | Alle ondersteunde scoresysteem-identifiers                   |
| `tiebreakers`    | array  | Alle geregistreerde tiebreaker-ID's (alfabetisch gesorteerd) |

## Tiebreakers

Commando: `chesspairing tiebreakers --json`

```json
[
  { "id": "aro", "name": "Average Rating of Opponents" },
  { "id": "buchholz", "name": "Buchholz" },
  { "id": "buchholz-cut1", "name": "Buchholz Cut 1" }
]
```

### Velden

De uitvoer is een JSON-array (niet gewrapped in een object).

| Veld      | Type   | Beschrijving                                                 |
| --------- | ------ | ------------------------------------------------------------ |
| `[].id`   | string | Tiebreaker-register-ID (gebruikt in de `--tiebreakers`-flag) |
| `[].name` | string | Leesbare weergavenaam                                        |

Items zijn alfabetisch gesorteerd op ID.
