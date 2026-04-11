---
title: "standings"
linkTitle: "standings"
weight: 6
description: "Bereken en toon de toernooistand vanuit een TRF-bestand."
---

Het `standings`-subcommando berekent scores, tiebreakers en de eindstand vanuit een TRF-bestand. Een indelingssysteemvlag bepaalt de standaard tiebreaker-volgorde via `DefaultTiebreakers(system)`. Wanneer `--tiebreakers` expliciet wordt meegegeven, is de systeemvlag optioneel. Het scoringssysteem, puntwaarden en de tiebreaker-selectie zijn allemaal instelbaar.

## Synopsis

```text
chesspairing standings [SYSTEM] input-file [options]
```

De indelingssysteemvlag (bijv. `--dutch`) mag overal in de argumentenlijst staan -- voor of na het invoerbestand. De vlag is vereist, tenzij `--tiebreakers` expliciet wordt meegegeven. Het invoerbestand kan een bestandspad zijn of `-` voor stdin.

Vlaggen en positionele argumenten kunnen in willekeurige volgorde worden gemengd.

## Indelingssysteemvlaggen

Een systeemvlag wordt verwacht (optioneel wanneer `--tiebreakers` wordt meegegeven):

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

De systeemvlag wordt verwerkt voordat andere vlaggen worden geparsed, en kan dus overal in de argumentenlijst staan. De vlag selecteert geen indelingsengine -- hij wordt alleen gebruikt om de standaard tiebreaker-volgorde te bepalen.

## Vlaggen

| Vlag             | Type    | Standaard  | Beschrijving                                                      |
| ---------------- | ------- | ---------- | ----------------------------------------------------------------- |
| `-o`             | string  | --         | Pad naar uitvoerbestand (stdout indien weggelaten)                |
| `--scoring`      | string  | `standard` | Scoringssysteem: `standard`, `keizer`, `football`                 |
| `--tiebreakers`  | string  | --         | Kommagescheiden tiebreaker-ID's (overschrijft systeemstandaarden) |
| `--win`          | float64 | --         | Overschrijf punten voor winst                                     |
| `--draw`         | float64 | --         | Overschrijf punten voor remise                                    |
| `--loss`         | float64 | --         | Overschrijf punten voor verlies                                   |
| `--forfeit-win`  | float64 | --         | Overschrijf punten voor forfaitwinst                              |
| `--bye`          | float64 | --         | Overschrijf punten voor bye                                       |
| `--forfeit-loss` | float64 | --         | Overschrijf punten voor forfaitverlies                            |
| `--json`         | bool    | `false`    | Uitvoer als JSON                                                  |
| `--help`         | --      | --         | Hulp tonen                                                        |

Vlaggen voor puntoverschrijving werken alleen wanneer de waarde >= 0 is. Ze corresponderen met scoringsoptiesleutels: `pointWin`, `pointDraw`, `pointLoss`, `pointForfeitWin`, `pointBye`, `pointForfeitLoss`. Deze overschrijvingen worden toegepast bovenop eventuele scoringsopties die al in het TRF-bestand aanwezig zijn.

## Voorbeelden

```bash
# Standaardstand (standaard scoring, systeem-standaard tiebreakers)
chesspairing standings --dutch tournament.trf

# Football-scoring met aangepaste tiebreakers
chesspairing standings --dutch tournament.trf --scoring football --tiebreakers buchholz,wins,progressive

# Aangepaste puntwaarden
chesspairing standings --dutch tournament.trf --win 3 --draw 1 --loss 0

# Keizer-scoring
chesspairing standings --keizer tournament.trf --scoring keizer

# JSON-uitvoer
chesspairing standings --dutch tournament.trf --json

# Schrijf stand naar bestand
chesspairing standings --dutch tournament.trf -o standings.txt

# Zonder systeemvlag (vereist expliciete tiebreakers)
chesspairing standings tournament.trf --tiebreakers buchholz,wins

# Lees van stdin
cat tournament.trf | chesspairing standings --dutch -
```

## Uitvoer

### Tekstformaat (standaard)

Tab-uitgelijnde tabel met rang, speler-ID, naam, score en een kolom per tiebreaker. De kolomkoppen van de tiebreakers gebruiken de weergavenaam van elke tiebreaker.

```text
Rank  ID  Name              Score  Buchholz Cut 1  Buchholz  Wins
----  --  ----              -----  --------------  --------  ----
1     5   GM Smith          7.0    28.5            35.0      6
2     1   IM Jones          7.0    27.0            34.5      5
3     12  FM Brown          6.5    26.0            33.0      5
```

Gelijkgeplaatste spelers delen dezelfde rang. Scores en tiebreaker-waarden worden als gehele getallen weergegeven wanneer het hele getallen zijn, anders met een decimaal.

Wanneer er geen stand beschikbaar is, is de uitvoer `(no standings)`.

### JSON-formaat

```json
{
  "standings": [
    {
      "rank": 1,
      "playerId": "5",
      "displayName": "GM Smith",
      "score": 7,
      "tieBreakers": [
        { "id": "buchholz-cut1", "name": "Buchholz Cut 1", "value": 28.5 },
        { "id": "buchholz", "name": "Buchholz", "value": 35 },
        { "id": "wins", "name": "Wins", "value": 6 }
      ],
      "gamesPlayed": 9,
      "wins": 6,
      "draws": 2,
      "losses": 1
    }
  ],
  "scoring": "standard",
  "tiebreakers": ["buchholz-cut1", "buchholz", "wins"]
}
```

Elke standvermelding bevat partijstatistieken (`gamesPlayed`, `wins`, `draws`, `losses`), berekend uit de rondegegevens. De `tiebreakers`-array op het hoogste niveau toont de aangevraagde tiebreaker-ID's (inclusief eventuele die zijn mislukt en overgeslagen).

## Tiebreaker-selectie

Als `--tiebreakers` niet wordt opgegeven, wordt de standaard tiebreaker-volgorde voor het opgegeven indelingssysteem gebruikt (een systeemvlag is dan vereist). De standaarden vanuit `DefaultTiebreakers()` zijn:

| Systeem                                         | Standaard tiebreakers                                               |
| ----------------------------------------------- | ------------------------------------------------------------------- |
| Dutch, Burstein, Dubov, Lim, Double-Swiss, Team | `buchholz-cut1`, `buchholz`, `sonneborn-berger`, `direct-encounter` |
| Round-Robin                                     | `sonneborn-berger`, `direct-encounter`, `wins`, `koya`              |
| Keizer                                          | `games-played`, `direct-encounter`, `wins`                          |

Onbekende tiebreaker-ID's geven een waarschuwing naar stderr en worden overgeslagen. Mislukte tiebreaker-berekeningen geven ook een waarschuwing en worden overgeslagen. In geen van beide gevallen mislukt het commando.

Beschikbare tiebreaker-ID's kunnen worden opgevraagd met het [tiebreakers](../tiebreakers-cmd/)-subcommando.

## Standberekening

1. Parseer het TRF-bestand en converteer naar `TournamentState`
2. Voer de geselecteerde scorer uit (`standard`, `keizer` of `football`) om `PlayerScore`-waarden te produceren
3. Bereken elke tiebreaker op volgorde, sla mislukte over
4. Sorteer op score aflopend, daarna op elke tiebreaker-waarde op volgorde (aflopend)
5. Ken gedeelde rangen toe -- spelers met identieke scores en identieke tiebreaker-waarden krijgen dezelfde rang; de volgende rang is de positie in de gesorteerde lijst (bijv. twee spelers gelijk op rang 1 worden gevolgd door rang 3)

## Exitcodes

| Code | Betekenis                                                                                         |
| ---- | ------------------------------------------------------------------------------------------------- |
| 0    | Stand succesvol berekend                                                                          |
| 2    | Onverwachte fout (scoringsfout, JSON-encodingfout)                                                |
| 3    | Ongeldige invoer (ontbrekende systeemvlag zonder `--tiebreakers`, misvormd TRF, onbekende scorer) |
| 5    | Bestandsfout                                                                                      |

## Zie ook

- [tiebreakers](../tiebreakers-cmd/) -- toon beschikbare tiebreaker-ID's
- [pair](../pair/) -- genereer indelingen voor de volgende ronde
- [Uitvoerformaten en exitcodes](../output-formats/) -- gedetailleerde formaatspecificaties en alle exitcodes
