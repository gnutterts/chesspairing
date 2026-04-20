---
title: "TRF16 I/O"
linkTitle: "TRF I/O"
weight: 7
description: "TRF16-bestanden lezen, schrijven en converteren met het trf-pakket."
---

Het `trf`-pakket (`github.com/gnutterts/chesspairing/trf`) verzorgt het lezen, schrijven, valideren en converteren van FIDE Tournament Report Files. Het ondersteunt zowel TRF16 (legacy) als TRF-2026.

## Kerntypen

### Document

Het hoofdtype dat een volledig TRF-bestand representeert. Belangrijke velden (verkort):

```go
type Document struct {
    // Toernooi-info (headerregels 012-132)
    Name           string   `json:"name,omitempty"`
    City           string   `json:"city,omitempty"`
    Federation     string   `json:"federation,omitempty"`
    StartDate      string   `json:"startDate,omitempty"`
    EndDate        string   `json:"endDate,omitempty"`
    NumPlayers     int      `json:"numPlayers,omitempty"`
    NumRated       int      `json:"numRated,omitempty"`
    TournamentType string   `json:"tournamentType,omitempty"`
    ChiefArbiter   string   `json:"chiefArbiter,omitempty"`
    TimeControl    string   `json:"timeControl,omitempty"`
    RoundDates     []string `json:"roundDates,omitempty"`

    // TRF-2026 headervelden
    TotalRounds26  int    `json:"totalRounds26,omitempty"`  // 142
    InitialColor26 string `json:"initialColor26,omitempty"` // 152

    // TRF16 uitgebreide data
    TotalRounds    int             `json:"totalRounds,omitempty"`    // XXR
    InitialColor   string          `json:"initialColor,omitempty"`   // XXC
    Acceleration   []string        `json:"acceleration,omitempty"`   // XXS
    ForbiddenPairs []ForbiddenPair `json:"forbiddenPairs,omitempty"` // XXP

    // Systeemspecifieke velden (zie hieronder)
    Cycles                  int    `json:"cycles,omitempty"`                  // XXY
    ColorBalance            *bool  `json:"colorBalance,omitempty"`            // XXB
    MaxiTournament          *bool  `json:"maxiTournament,omitempty"`          // XXM
    ColorPreferenceType     string `json:"colorPreferenceType,omitempty"`     // XXT
    PrimaryScore            string `json:"primaryScore,omitempty"`            // XXG
    AllowRepeatPairings     *bool  `json:"allowRepeatPairings,omitempty"`     // XXA
    MinRoundsBetweenRepeats int    `json:"minRoundsBetweenRepeats,omitempty"` // XXK

    // Spelers- en teamdata
    Players []PlayerLine `json:"players,omitempty"` // 001-regels
    Teams   []TeamLine   `json:"teams,omitempty"`   // 013-regels

    // Commentaarregio
    Comments               []string    `json:"comments,omitempty"`               // vrije ###-regels
    ChesspairingDirectives []Directive `json:"chesspairingDirectives,omitempty"` // ### chesspairing:<verb>

    // Onbekende regels bewaard voor round-trip-betrouwbaarheid
    Other []RawLine `json:"other,omitempty"`
}
```

Alle structvelden hebben JSON-tags voor serialisatie.

`ChesspairingDirectives` bevat getypeerde
`### chesspairing:<verb> k=v ...`-commentaarregels die gegevens
uitdrukken die buiten de FIDE-TRF-woordenschat vallen, zoals
`ByeExcused`/`ByeClubCommitment`-vooraf-toegewezen-byes en
spelersafmeldingen. Elk item is een `Directive{Verb string, Params
map[string]string}`. Zie [TRF-2026-extensies](/nl/docs/formats/trf-extensions/)
voor de directiefgrammatica en het koppelingsgedrag.

#### EffectiveTotalRounds

```go
func (doc *Document) EffectiveTotalRounds() int
```

Retourneert het totaal aantal ronden uit TRF-2026 veld 142 indien ingesteld, met terugval op TRF16 veld XXR. Retourneert 0 als geen van beide is ingesteld.

#### EffectiveInitialColor

```go
func (doc *Document) EffectiveInitialColor() string
```

Retourneert de beginkleur uit TRF-2026 veld 152 indien ingesteld, met terugval op TRF16 veld XXC. Retourneert `""` als geen van beide is ingesteld.

### PlayerLine

```go
type PlayerLine struct {
    StartNumber int           `json:"startNumber"`
    Sex         string        `json:"sex,omitempty"`
    Title       string        `json:"title,omitempty"`
    Name        string        `json:"name,omitempty"`
    Rating      int           `json:"rating,omitempty"`
    Federation  string        `json:"federation,omitempty"`
    FideID      string        `json:"fideID,omitempty"`
    BirthDate   string        `json:"birthDate,omitempty"`
    Points      float64       `json:"points"`
    Rank        int           `json:"rank"`
    Rounds      []RoundResult `json:"rounds,omitempty"`
}
```

Representeert een enkele 001-spelersregel. De `Rounds`-slice bevat een `RoundResult` per ronde waarvoor de speler gegevens heeft.

### RoundResult

```go
type RoundResult struct {
    Opponent int        `json:"opponent"` // Startnummer (0 = bye/afwezig)
    Color    Color      `json:"color"`
    Result   ResultCode `json:"result"`
}
```

Een enkele rondevermelding uit de 001-regel van een speler.

### Color

```go
type Color int

const (
    ColorNone  Color = 0 // "-" (bye, afwezig, geen partij)
    ColorWhite Color = 1 // "w"
    ColorBlack Color = 2 // "b"
)
```

Methoden:

| Methode         | Signatuur                                          | Beschrijving                                 |
| --------------- | -------------------------------------------------- | -------------------------------------------- |
| `IsValid`       | `func (c Color) IsValid() bool`                    | True voor ColorNone, ColorWhite, ColorBlack  |
| `String`        | `func (c Color) String() string`                   | Retourneert `"None"`, `"White"` of `"Black"` |
| `Char`          | `func (c Color) Char() byte`                       | Retourneert `'-'`, `'w'` of `'b'`            |
| `MarshalJSON`   | `func (c Color) MarshalJSON() ([]byte, error)`     | JSON-serialisatie                            |
| `UnmarshalJSON` | `func (c *Color) UnmarshalJSON(data []byte) error` | JSON-deserialisatie                          |

### ResultCode

```go
type ResultCode int
```

13 constanten die alle mogelijke TRF-resultaattekens representeren:

| Constante             | Waarde | Teken | Betekenis                   |
| --------------------- | ------ | ----- | --------------------------- |
| `ResultWin`           | 0      | `1`   | Winst (gespeeld)            |
| `ResultLoss`          | 1      | `0`   | Verlies (gespeeld)          |
| `ResultDraw`          | 2      | `=`   | Remise                      |
| `ResultForfeitWin`    | 3      | `+`   | Winst door forfait          |
| `ResultForfeitLoss`   | 4      | `-`   | Verlies door forfait        |
| `ResultHalfBye`       | 5      | `H`   | Half-punt bye               |
| `ResultFullBye`       | 6      | `F`   | Vol-punt bye (PAB)          |
| `ResultUnpaired`      | 7      | `U`   | Niet ingedeeld (afwezig, 0 ptn)   |
| `ResultZeroBye`       | 8      | `Z`   | Nul-punt bye                |
| `ResultNotPlayed`     | 9      | `*`   | Nog niet gespeeld           |
| `ResultWinByDefault`  | 10     | `W`   | Winst, tegenstander afwezig |
| `ResultDrawByDefault` | 11     | `D`   | Remise bij verstek          |
| `ResultLossByDefault` | 12     | `L`   | Verlies bij verstek         |

Methoden: `IsValid() bool`, `String() string`, `Char() byte`, `MarshalJSON`, `UnmarshalJSON`.

## Lezen

```go
func Read(r io.Reader) (*Document, error)
```

Parst een TRF-bestand vanuit elke `io.Reader`. Ondersteunt zowel TRF16 als TRF-2026 (automatisch gedetecteerd aan de hand van de aanwezige regelcodes).

Retourneert een `*ParseError` bij parseerfouten. ParseError bevat het 1-gebaseerde regelnummer en de regelcode voor diagnostiek:

```go
type ParseError struct {
    Line    int    `json:"line"`
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

Voorbeeld:

```go
f, err := os.Open("tournament.trf")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

doc, err := trf.Read(f)
if err != nil {
    var pe *trf.ParseError
    if errors.As(err, &pe) {
        log.Printf("parseerfout op regel %d (%s): %s", pe.Line, pe.Code, pe.Message)
    }
    log.Fatal(err)
}
```

## Schrijven

```go
func Write(w io.Writer, doc *Document) error
```

Serialiseert een Document naar TRF-formaat, schrijvend naar elke `io.Writer`. Behandelt zowel TRF16- als TRF-2026-velden -- het schrijft de velden die in het Document zijn ingevuld.

```go
var buf bytes.Buffer
if err := trf.Write(&buf, doc); err != nil {
    log.Fatal(err)
}
// buf.String() bevat de TRF-bestandsinhoud
```

## Converteren

Bidirectionele conversie tussen `Document` en `chesspairing.TournamentState`.

### ToTournamentState

```go
func (doc *Document) ToTournamentState() (*chesspairing.TournamentState, error)
```

Converteert een TRF-Document naar een `TournamentState` geschikt voor engine-gebruik.

Belangrijk gedrag:

- **Speler-ID's** zijn startnummers als strings (`"1"`, `"2"`, ...).
- **Ronden** worden gereconstrueerd door per-speler resultaten te kruisverwijzen. Partijen worden ontdubbeld (elke partij verschijnt eenmaal, niet tweemaal).
- **Indelingssysteem** wordt afgeleid uit het TRF-toernooi-typeveld (bijv. `"Swiss Dutch"` mapt naar `PairingDutch`).
- **Systeemspecifieke opties** uit XX-velden (XXR, XXC, XXS, XXP, XXY, XXB, XXM, XXT, XXG, XXA, XXK) en TRF-2026-equivalenten worden in `PairingConfig.Options` geplaatst voor round-tripping.
- **Scoringsconfiguratie** valt standaard terug op `ScoringStandard` met `DefaultTiebreakers()` voor het afgeleide indelingssysteem.

```go
doc, err := trf.Read(reader)
if err != nil {
    log.Fatal(err)
}

state, err := doc.ToTournamentState()
if err != nil {
    log.Fatal(err)
}
// state.PairingConfig.System == chesspairing.PairingDutch
// state.Players[0].ID == "1"
```

### FromTournamentState

```go
func FromTournamentState(state *chesspairing.TournamentState) (*Document, map[string]int)
```

Maakt een Document aan vanuit een TournamentState. Retourneert het Document en een mapping van speler-ID naar toegewezen startnummer.

Belangrijk gedrag:

- **Startnummers** worden toegewezen door spelers te sorteren op rating aflopend, dan op naam oplopend, dan op ID oplopend (voor determinisme).
- **Spelersmap** laat aanroepers opzoeken welk startnummer aan elke speler-ID is toegewezen.
- **Ronderesultaten** worden gereconstrueerd uit `state.Rounds`-partijgegevens en byes.
- **Toernooitype** en **XX-velden** worden ingesteld vanuit `PairingConfig.System` en `PairingConfig.Options`.

```go
doc, playerMap := trf.FromTournamentState(&state)
// playerMap["player-uuid-1"] == 1  (hoogst geratet)
// playerMap["player-uuid-2"] == 2
err := trf.Write(os.Stdout, doc)
```

## Validatie

```go
func (doc *Document) Validate(profile ValidationProfile) []ValidationIssue
```

Controleert een Document op volledigheid volgens een validatieprofiel. Elk profiel is een superset van het vorige.

### Profielen

| Profiel       | Constante                   | Controles                                                                                                                                      |
| ------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Algemeen      | `ValidateGeneral` (0)       | Structurele integriteit: spelers bestaan, unieke startnummers, NumPlayers-consistentie                                                         |
| Indelingsengine | `ValidatePairingEngine` (1) | Algemeen + XXR/142 (totaal ronden), XXC/152 (beginkleur), 092 (toernooitype), tegenstander-symmetrie, kleurconsistentie, resultaatconsistentie |
| FIDE          | `ValidateFIDE` (2)          | Indelingsengine + toernooinama, data, bedenktijd, per-speler naam/federatie/FIDE-ID                                                              |

### ValidationIssue

```go
type ValidationIssue struct {
    Field    string   // bijv. "012", "XXR", "player.3.fideID"
    Severity Severity // SeverityError of SeverityWarning
    Message  string   // Leesbare beschrijving
}
```

- **SeverityError**: moet opgelost worden om aan het profiel te voldoen.
- **SeverityWarning**: informatief, niet blokkerend.

Voorbeeld:

```go
issues := doc.Validate(trf.ValidatePairingEngine)
for _, issue := range issues {
    fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Field, issue.Message)
}
```

## Systeemspecifieke uitgebreide velden

TRF16 gebruikt XX-voorvoegsels als regelcodes voor engine-specifieke configuratie. Deze worden opgeslagen in het Document en maken een round-trip via `ToTournamentState()` / `FromTournamentState()`.

### Gemeenschappelijke velden

| Veld             | Regelcode | Type              | Beschrijving                                                       |
| ---------------- | --------- | ----------------- | ------------------------------------------------------------------ |
| `TotalRounds`    | XXR       | `int`             | Totaal aantal ronden in het toernooi                               |
| `InitialColor`   | XXC       | `string`          | Beginkleur-toewijzing voor de hoogst geplaatste (bijv. `"white1"`) |
| `Acceleration`   | XXS       | `[]string`        | Baku-acceleratiedata                                               |
| `ForbiddenPairs` | XXP       | `[]ForbiddenPair` | Paren startnummers die niet ingedeeld mogen worden                    |

### Round-robin-velden

| Veld           | Regelcode | Type    | Beschrijving                                      |
| -------------- | --------- | ------- | ------------------------------------------------- |
| `Cycles`       | XXY       | `int`   | Aantal cycli (1 = enkel, 2 = dubbele round-robin) |
| `ColorBalance` | XXB       | `*bool` | FIDE-kleurbalancering inschakelen                 |

### Lim-veld

| Veld             | Regelcode | Type    | Beschrijving            |
| ---------------- | --------- | ------- | ----------------------- |
| `MaxiTournament` | XXM       | `*bool` | Lim maxi-toernooi-modus |

### Team-Zwitserse velden

| Veld                  | Regelcode | Type     | Beschrijving                                |
| --------------------- | --------- | -------- | ------------------------------------------- |
| `ColorPreferenceType` | XXT       | `string` | Kleurvoorkeurtype: `"A"`, `"B"` of `"none"` |
| `PrimaryScore`        | XXG       | `string` | Primair scoretype: `"match"` of `"game"`    |

### Keizer-velden

| Veld                      | Regelcode | Type    | Beschrijving                                     |
| ------------------------- | --------- | ------- | ------------------------------------------------ |
| `AllowRepeatPairings`     | XXA       | `*bool` | Of herhaalde indelingen zijn toegestaan            |
| `MinRoundsBetweenRepeats` | XXK       | `int`   | Minimaal aantal ronden tussen herhaalde indelingen |

### TRF-2026-equivalenten

TRF-2026 introduceert nieuwe regelcodes die sommige XX-velden vervangen. Het Document slaat beide op, en de `Effective*`-methoden bieden uniforme toegang:

- **142** (`TotalRounds26`) vervangt XXR -- gebruik `EffectiveTotalRounds()`.
- **152** (`InitialColor26`) vervangt XXC -- gebruik `EffectiveInitialColor()`.
- **250** (`Accelerations26`) vervangt XXS.
- **260** (`ForbiddenPairs26`) vervangt XXP.
