---
title: "Kerntypen"
linkTitle: "Kerntypen"
weight: 2
description: "TournamentState, PlayerEntry, RoundData, GameData en andere fundamentele typen."
---

Alle kerntypen zijn gedefinieerd in het root-`chesspairing`-pakket (`result.go`). Ze vormen het gedeelde datamodel waarop alle indelings-, scorings- en tiebreaker-engines werken.

## GameResult

`GameResult` is een `string`-type dat de uitslag van een schaakpartij representeert.

### Constanten

| Constante                | Waarde      | Betekenis               |
| ------------------------ | ----------- | ----------------------- |
| `ResultWhiteWins`        | `"1-0"`     | Wit wint                |
| `ResultBlackWins`        | `"0-1"`     | Zwart wint              |
| `ResultDraw`             | `"0.5-0.5"` | Remise                  |
| `ResultPending`          | `"*"`       | Nog niet gespeeld       |
| `ResultForfeitWhiteWins` | `"1-0f"`    | Wit wint door forfait   |
| `ResultForfeitBlackWins` | `"0-1f"`    | Zwart wint door forfait |
| `ResultDoubleForfeit`    | `"0-0f"`    | Beide forfait           |

### Methoden

| Methode             | Retourneert | Beschrijving                                                                                                                                                                                                                      |
| ------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `IsValid()`         | `bool`      | True als de waarde een van de 7 erkende constanten is.                                                                                                                                                                            |
| `IsRecordable()`    | `bool`      | True als het resultaat door een gebruiker kan worden vastgelegd. Alle geldige resultaten behalve `ResultPending` zijn vastlegbaar. `ResultPending` is de beginstatus die het systeem instelt wanneer een partij wordt aangemaakt. |
| `IsForfeit()`       | `bool`      | True als het resultaat een forfait is (enkel of dubbel).                                                                                                                                                                          |
| `IsDoubleForfeit()` | `bool`      | True alleen voor `ResultDoubleForfeit`.                                                                                                                                                                                           |

### Forfait-semantiek

**Enkel forfait** (`ResultForfeitWhiteWins`, `ResultForfeitBlackWins`): De winnaar ontvangt punten. De partij wordt uitgesloten van de indelingsgeschiedenis, wat betekent dat de twee spelers in een latere ronde opnieuw tegen elkaar ingedeeld kunnen worden alsof ze nooit gespeeld hebben.

**Dubbel forfait** (`ResultDoubleForfeit`): De partij wordt uitgesloten van zowel scoring als indeling. Geen van beide spelers ontvangt punten, en de partij wordt behandeld alsof deze nooit heeft plaatsgevonden.

## ByeType

`ByeType` is een `int`-type (iota-gebaseerd) dat classificeert hoe een bye wordt gescoord.

### Constanten

| Constante           | Waarde | Beschrijving                                                |
| ------------------- | ------ | ----------------------------------------------------------- |
| `ByePAB`            | `0`    | Indelings-toegewezen bye. Vol punt. TRF-code `"F"`.           |
| `ByeHalf`           | `1`    | Half-punt bye. TRF-code `"H"`.                              |
| `ByeZero`           | `2`    | Nul-punt bye. TRF-code `"Z"`.                               |
| `ByeAbsent`         | `3`    | Afwezig/niet ingedeeld, zonder bericht. TRF-code `"U"`.           |
| `ByeExcused`        | `4`    | Verontschuldigde afwezigheid (vooraf gemeld).               |
| `ByeClubCommitment` | `5`    | Clubverplichting (afwezig voor intercompetitie-teamplicht). |

### Methoden

| Methode     | Retourneert | Beschrijving                                                                                             |
| ----------- | ----------- | -------------------------------------------------------------------------------------------------------- |
| `IsValid()` | `bool`      | True als de waarde in het bereik `ByePAB` tot en met `ByeClubCommitment` valt.                           |
| `String()`  | `string`    | Leesbare naam: `"PAB"`, `"Half"`, `"Zero"`, `"Absent"`, `"Excused"`, `"ClubCommitment"`, of `"Unknown"`. |

## TournamentState

De centrale datastructuur. Alle engines ontvangen een pointer naar `TournamentState` en behandelen deze als alleen-lezen.

```go
type TournamentState struct {
    Players       []PlayerEntry
    Rounds        []RoundData
    CurrentRound  int
    PairingConfig PairingConfig
    ScoringConfig ScoringConfig
    Info          TournamentInfo
}
```

| Veld            | Type             | Beschrijving                                                               |
| --------------- | ---------------- | -------------------------------------------------------------------------- |
| `Players`       | `[]PlayerEntry`  | Alle spelers die voor het toernooi zijn ingeschreven.                      |
| `Rounds`        | `[]RoundData`    | Voltooide ronden met partijenresultaten en byes.                           |
| `CurrentRound`  | `int`            | De volgende te indelen ronde (1-gebaseerd).                                  |
| `PairingConfig` | `PairingConfig`  | Selectie van het indelingssysteem en engine-specifieke opties.               |
| `ScoringConfig` | `ScoringConfig`  | Selectie van het scoringssysteem, tiebreaker-lijst en scoringsopties.      |
| `Info`          | `TournamentInfo` | Toernooi-metadata. Nulwaarde als niet ingesteld. Engines negeren dit veld. |

### Validate()

```go
func (s *TournamentState) Validate() error
```

Controleert structurele invarianten en retourneert een fout die het eerste gevonden probleem beschrijft, of `nil` als alles geldig is. De controles zijn:

- Er bestaat minstens een speler.
- Geen speler heeft een leeg `ID`.
- Geen dubbele speler-ID's.
- `CurrentRound` overschrijdt niet `len(Rounds)`.

## PlayerEntry

Representeert een enkele speler voor engine-doeleinden.

```go
type PlayerEntry struct {
    ID          string
    DisplayName string
    Rating      int
    Active      bool
    Federation  string
    FideID      string
    Title       string
    Sex         string
    BirthDate   string
}
```

| Veld          | Type     | Beschrijving                                                                                                                      |
| ------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `ID`          | `string` | Unieke spelersidentificatie. Mag niet leeg zijn.                                                                                  |
| `DisplayName` | `string` | Spelersnaam voor weergavedoeleinden.                                                                                              |
| `Rating`      | `int`    | Spelersrating (bijv. FIDE Elo). Gebruikt voor seeding en tiebreakers.                                                             |
| `Active`      | `bool`   | Of de speler actief is. `false` betekent dat de speler zich heeft teruggetrokken en niet meer ingedeeld wordt in toekomstige ronden. |
| `Federation`  | `string` | FIDE-federatiecode (bijv. `"NED"`, `"USA"`, `"IND"`). Leeg als onbekend. Gebruikt door Varma-toewijzing voor round-robin.         |
| `FideID`      | `string` | FIDE-spelernummer. Leeg als onbekend.                                                                                             |
| `Title`       | `string` | FIDE-titelcode (`"GM"`, `"IM"`, `"FM"`, `"WGM"`, `"WIM"`, `"WFM"`, `"CM"`, `"WCM"`). Leeg als zonder titel.                       |
| `Sex`         | `string` | `"m"` of `"w"`. Leeg als onbekend.                                                                                                |
| `BirthDate`   | `string` | Geboortedatum als `YYYY/MM/DD`. Leeg als onbekend.                                                                                |

## RoundData

Bevat alle partijen en byes van een voltooide ronde.

```go
type RoundData struct {
    Number int
    Games  []GameData
    Byes   []ByeEntry
}
```

| Veld     | Type         | Beschrijving                        |
| -------- | ------------ | ----------------------------------- |
| `Number` | `int`        | Rondenummer, 1-gebaseerd.           |
| `Games`  | `[]GameData` | Alle partijen in deze ronde.        |
| `Byes`   | `[]ByeEntry` | Alle byes toegewezen in deze ronde. |

## GameData

Een enkel partijresultaat voor gebruik door engines.

```go
type GameData struct {
    WhiteID   string
    BlackID   string
    Result    GameResult
    IsForfeit bool
}
```

| Veld        | Type         | Beschrijving                                                                                                                      |
| ----------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------- |
| `WhiteID`   | `string`     | Speler-ID van de witspeler.                                                                                                       |
| `BlackID`   | `string`     | Speler-ID van de zwartspeler.                                                                                                     |
| `Result`    | `GameResult` | De uitslag van de partij.                                                                                                         |
| `IsForfeit` | `bool`       | Redundant met `Result.IsForfeit()` maar voor het gemak aanwezig, zodat aanroepers niet de resultaat-string hoeven te controleren. |

## ByeEntry

Registreert een bye-toewijzing met het bijbehorende type.

```go
type ByeEntry struct {
    PlayerID string
    Type     ByeType
}
```

| Veld       | Type      | Beschrijving                          |
| ---------- | --------- | ------------------------------------- |
| `PlayerID` | `string`  | De speler die de bye heeft ontvangen. |
| `Type`     | `ByeType` | Hoe de bye wordt gescoord.            |

## ResultContext

Biedt aanvullende context aan `Scorer.PointsForResult()` bij het berekenen van punten voor een specifiek partijresultaat.

```go
type ResultContext struct {
    OpponentRank        int
    OpponentValueNumber int
    PlayerRank          int
    PlayerValueNumber   int
    IsBye               bool
    IsAbsent            bool
    IsForfeit           bool
}
```

| Veld                  | Type   | Beschrijving                                            |
| --------------------- | ------ | ------------------------------------------------------- |
| `OpponentRank`        | `int`  | Huidige rang van de tegenstander (1-gebaseerd).         |
| `OpponentValueNumber` | `int`  | Keizer-waardenummer van de tegenstander (rangafgeleid). |
| `PlayerRank`          | `int`  | Huidige rang van de speler.                             |
| `PlayerValueNumber`   | `int`  | Keizer-waardenummer van de speler.                      |
| `IsBye`               | `bool` | True als dit resultaat een bye is (geen tegenstander).  |
| `IsAbsent`            | `bool` | True als de speler afwezig was.                         |
| `IsForfeit`           | `bool` | True als de partij een forfait was.                     |

Deze struct wordt voornamelijk gebruikt door het Keizer-scoringssysteem, waar puntwaarden afhangen van de rang en het waardenummer van de tegenstander. Standaard- en football-scoring negeren de rang-/waardevelden en gebruiken alleen `IsBye`, `IsAbsent` en `IsForfeit`.

## PairingResult

Uitvoer van `Pairer.Pair()`. Bevat de bordtoewijzingen en eventuele byes voor de ronde.

```go
type PairingResult struct {
    Pairings []GamePairing
    Byes     []ByeEntry
    Notes    []string
}
```

| Veld       | Type            | Beschrijving                                                                            |
| ---------- | --------------- | --------------------------------------------------------------------------------------- |
| `Pairings` | `[]GamePairing` | Bordtoewijzingen voor de ronde.                                                         |
| `Byes`     | `[]ByeEntry`    | Byes toegewezen door de engine (doorgaans hoogstens een PAB bij Zwitserse systemen).    |
| `Notes`    | `[]string`      | Diagnostische berichten van de engine (bijv. waarschuwingen over criteriaversoepeling). |

## GamePairing

Een enkele bordtoewijzing binnen een `PairingResult`.

```go
type GamePairing struct {
    Board   int
    WhiteID string
    BlackID string
}
```

| Veld      | Type     | Beschrijving                                      |
| --------- | -------- | ------------------------------------------------- |
| `Board`   | `int`    | Bordnummer, 1-geindexeerd. Bord 1 is het topbord. |
| `WhiteID` | `string` | Speler-ID toegewezen aan de witte stukken.        |
| `BlackID` | `string` | Speler-ID toegewezen aan de zwarte stukken.       |

## PlayerScore

Uitvoer van `Scorer.Score()`. Een vermelding per speler.

```go
type PlayerScore struct {
    PlayerID string
    Score    float64
    Rank     int
}
```

| Veld       | Type      | Beschrijving                                                                      |
| ---------- | --------- | --------------------------------------------------------------------------------- |
| `PlayerID` | `string`  | De unieke identificatie van de speler.                                            |
| `Score`    | `float64` | De totale score van de speler onder het actieve scoringssysteem.                  |
| `Rank`     | `int`     | De rang van de speler op basis van score (1-gebaseerd, gelijke standen mogelijk). |

## TieBreakValue

Uitvoer van `TieBreaker.Compute()`. Een vermelding per speler.

```go
type TieBreakValue struct {
    PlayerID string
    Value    float64
}
```

| Veld       | Type      | Beschrijving                                  |
| ---------- | --------- | --------------------------------------------- |
| `PlayerID` | `string`  | De unieke identificatie van de speler.        |
| `Value`    | `float64` | De berekende tiebreakwaarde voor deze speler. |

## Standing

Uiteindelijke gerangschikte uitvoer die score en tiebreakers combineert. Alle velden hebben JSON-structtags voor serialisatie.

```go
type Standing struct {
    Rank        int          `json:"rank"`
    PlayerID    string       `json:"playerId"`
    DisplayName string       `json:"displayName"`
    Score       float64      `json:"score"`
    TieBreakers []NamedValue `json:"tieBreakers"`
    GamesPlayed int          `json:"gamesPlayed"`
    Wins        int          `json:"wins"`
    Draws       int          `json:"draws"`
    Losses      int          `json:"losses"`
}
```

| Veld          | JSON-sleutel  | Type           | Beschrijving                                |
| ------------- | ------------- | -------------- | ------------------------------------------- |
| `Rank`        | `rank`        | `int`          | Eindrangschikking (1-gebaseerd).            |
| `PlayerID`    | `playerId`    | `string`       | Spelersidentificatie.                       |
| `DisplayName` | `displayName` | `string`       | Spelersnaam.                                |
| `Score`       | `score`       | `float64`      | Totale score.                               |
| `TieBreakers` | `tieBreakers` | `[]NamedValue` | Geordende tiebreakwaarden.                  |
| `GamesPlayed` | `gamesPlayed` | `int`          | Totaal gespeelde partijen (exclusief byes). |
| `Wins`        | `wins`        | `int`          | Totaal gewonnen partijen.                   |
| `Draws`       | `draws`       | `int`          | Totaal remises.                             |
| `Losses`      | `losses`      | `int`          | Totaal verloren partijen.                   |

## NamedValue

Koppelt een tiebreaker-identificatie aan de berekende waarde. Gebruikt binnen `Standing.TieBreakers`.

```go
type NamedValue struct {
    ID    string  `json:"id"`
    Name  string  `json:"name"`
    Value float64 `json:"value"`
}
```

| Veld    | JSON-sleutel | Type      | Beschrijving                                                 |
| ------- | ------------ | --------- | ------------------------------------------------------------ |
| `ID`    | `id`         | `string`  | Tiebreaker-register-identificatie (bijv. `"buchholz-cut1"`). |
| `Name`  | `name`       | `string`  | Leesbare tiebreaker-naam (bijv. `"Buchholz Cut 1"`).         |
| `Value` | `value`      | `float64` | Berekende tiebreakwaarde.                                    |

## TournamentInfo

Metadatastruct voor weergave en TRF round-trip-betrouwbaarheid. Engines negeren deze struct volledig. Hij wordt gevuld vanuit TRF-headerregels bij het parsen en teruggeschreven bij serialisatie naar TRF.

```go
type TournamentInfo struct {
    Name          string
    City          string
    Federation    string
    StartDate     string
    EndDate       string
    ChiefArbiter  string
    DeputyArbiter string
    TimeControl   string
    RoundDates    []string
}
```

| Veld            | Type       | Beschrijving                           |
| --------------- | ---------- | -------------------------------------- |
| `Name`          | `string`   | Toernooinama.                          |
| `City`          | `string`   | Stad waar het toernooi wordt gehouden. |
| `Federation`    | `string`   | Code van de organiserende federatie.   |
| `StartDate`     | `string`   | Startdatum als `YYYY/MM/DD`.           |
| `EndDate`       | `string`   | Einddatum als `YYYY/MM/DD`.            |
| `ChiefArbiter`  | `string`   | Naam van de hoofdarbiter.              |
| `DeputyArbiter` | `string`   | Naam van de plaatsvervangend arbiter.  |
| `TimeControl`   | `string`   | Beschrijving van de bedenktijd.        |
| `RoundDates`    | `[]string` | Data per ronde als `YYYY/MM/DD`.       |

## Configuratietypen

### PairingConfig

```go
type PairingConfig struct {
    System  PairingSystem
    Options map[string]any
}
```

Selecteert het indelingsalgoritme en geeft engine-specifieke opties door. De `Options`-map wordt door de `ParseOptions()`-functie van elke engine geparsed. Zie [Optiepatroon](../options/) voor details.

### ScoringConfig

```go
type ScoringConfig struct {
    System      ScoringSystem
    Tiebreakers []string
    Options     map[string]any
}
```

Selecteert het scoringsalgoritme, specificeert de geordende tiebreaker-lijst en geeft scoringsopties door. Tiebreaker-ID's corresponderen met het [tiebreaker-register](../tiebreaker/).
