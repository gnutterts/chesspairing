---
title: "Optiepatroon"
linkTitle: "Opties"
weight: 6
description: "Het pointerveld-optiepatroon dat door alle engines wordt gebruikt — WithDefaults, ParseOptions, NewFromMap."
---

Elke engine in chesspairing (pairers en scorers) gebruikt hetzelfde optiepatroon voor configuratie. Deze pagina legt het ontwerp uit en de standaardmethoden die elke engine biedt.

## De nil-betekent-standaard-conventie

Elke engine definieert een `Options`-struct waarvan de configureerbare velden **pointers** zijn. Een nil-pointer betekent "gebruik de standaardwaarde van de engine." Hierdoor kunnen aanroepers specifieke velden overschrijven en voor de rest de standaardwaarden overnemen.

Hier is de Options-struct van de Dutch pairer als voorbeeld:

```go
type Options struct {
    Acceleration   *string    `json:"acceleration,omitempty"`
    TopSeedColor   *string    `json:"topSeedColor,omitempty"`
    ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`
}
```

`Acceleration` op `nil` zetten betekent dat de Dutch pairer zijn ingebouwde standaard gebruikt (`"none"`). Het instellen op een pointer naar `"baku"` schakelt Baku-acceleratie expliciet in.

### Waarom pointers?

Zonder pointers is er geen manier om "de aanroeper heeft dit op false gezet" te onderscheiden van "de aanroeper heeft dit helemaal niet ingesteld." Een `bool`-veld staat in Go standaard op `false` -- je kunt niet zien of de aanroeper het opzettelijk op `false` heeft gezet of het simpelweg heeft weggelaten.

Met pointers betekent `nil` ondubbelzinnig "niet opgegeven, gebruik standaard." Een niet-nil pointer vertegenwoordigt altijd een expliciete keuze van de aanroeper, zelfs als de waarde toevallig overeenkomt met de standaard.

## Standaardmethoden

Elke engine biedt drie standaardmethoden om met opties te werken.

### WithDefaults

```go
func (o Options) WithDefaults() Options
```

Retourneert een kopie van de Options met nil-velden vervangen door de standaardwaarden van de engine. Wijzigt de ontvanger **niet**. Dit wordt intern aangeroepen door de engine-constructor, dus aanroepers hoeven het zelden zelf aan te roepen.

**Uitzondering:** de `WithDefaults` van de Keizer-scorer neemt een `playerCount int`-parameter omdat het standaard-waardenummer afhankelijk is van het aantal spelers. Het wordt aangeroepen binnen `Score()` wanneer het aantal spelers bekend is, niet door de constructor.

Voorbeeld (Dutch):

```go
opts := dutch.Options{
    Acceleration: nil, // wordt "none"
    TopSeedColor: nil, // wordt "auto"
}
filled := opts.WithDefaults()
// *filled.Acceleration == "none"
// *filled.TopSeedColor == "auto"
```

### ParseOptions

```go
func ParseOptions(m map[string]any) Options
```

Pakket-niveau functie die een generieke `map[string]any` (doorgaans van JSON-configuratie, TRF-uitgebreide data of CLI-vlaggen) parst naar een getypeerde Options-struct. Retourneert een Options met alleen de herkende sleutels ingevuld; onbekende sleutels worden stilzwijgend genegeerd.

Typeconversie wordt afgehandeld met de helperfuncties uit het rootpakket (zie [Helperfuncties](#helperfuncties) hieronder).

### NewFromMap

```go
func NewFromMap(m map[string]any) *Pairer  // of *Scorer
```

Pakket-niveau constructor die een volledig ge-initialiseerde engine aanmaakt vanuit een generieke optiemap. Dit is het factory-ingangspunt dat de CLI en het `trf`-pakket gebruiken voor generieke, configuratiegedreven instantiatie.

Intern roept het `ParseOptions(m)` aan, gevolgd door `New(opts)`, dat `WithDefaults()` toepast.

```go
// Deze zijn equivalent:
p1 := dutch.NewFromMap(map[string]any{"acceleration": "baku"})

opts := dutch.ParseOptions(map[string]any{"acceleration": "baku"})
p2 := dutch.New(opts)
```

## Helperfuncties

Het rootpakket (`chesspairing`) biedt pointerconstructors en type-veilige map-extractiefuncties in `options_helpers.go`. Deze worden gebruikt door `ParseOptions`-implementaties en zijn beschikbaar voor aanroepers die Options-structs direct willen construeren.

### Pointerconstructors

```go
func Float64Ptr(v float64) *float64
func IntPtr(v int) *int
func BoolPtr(v bool) *bool
func StringPtr(v string) *string
```

Maken een pointer naar de opgegeven waarde. Gebruikt bij het direct construeren van Options-structs:

```go
opts := keizer.Options{
    WinFraction:  chesspairing.Float64Ptr(1.0),
    SelfVictory:  chesspairing.BoolPtr(false),
}
```

### Map-extractie

```go
func GetFloat64(m map[string]any, key string) (float64, bool)
func GetInt(m map[string]any, key string) (int, bool)
func GetBool(m map[string]any, key string) (bool, bool)
func GetString(m map[string]any, key string) (string, bool)
```

Elke functie extraheert een getypeerde waarde uit een generieke map. De tweede retourwaarde geeft aan of de sleutel is gevonden met een compatibel type.

Typeconversieregels:

- **GetFloat64**: accepteert `float64`-, `int`- en `int64`-waarden.
- **GetInt**: accepteert `int`-, `int64`- en `float64`-waarden (afgekapt tot int).
- **GetBool**: accepteert alleen `bool`-waarden.
- **GetString**: accepteert alleen `string`-waarden.

Alle drie retourneren de nulwaarde en `false` als de sleutel ontbreekt of een incompatibel type heeft.

## Gebruikspatronen

### Directe constructie met specifieke overschrijvingen

```go
p := dutch.New(dutch.Options{
    Acceleration: chesspairing.StringPtr("baku"),
    // TopSeedColor: nil -- gebruikt standaard ("auto")
})
```

### Vanuit een configuratiemap

```go
opts := map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
}
p := dutch.NewFromMap(opts)
```

### Scorer-voorbeeld (Keizer)

```go
s := keizer.NewFromMap(map[string]any{
    "winFraction":  1.0,
    "drawFraction": 0.5,
    "selfVictory":  false,
})
```

## Engines die dit patroon implementeren

Elk pairer- en scorer-pakket volgt deze conventie:

| Pakket                | Type   | Belangrijke opties                                                                       |
| --------------------- | ------ | ---------------------------------------------------------------------------------------- |
| `pairing/dutch`       | Pairer | Acceleration, TopSeedColor, ForbiddenPairs                                               |
| `pairing/burstein`    | Pairer | Acceleration, TopSeedColor, ForbiddenPairs, TotalRounds                                  |
| `pairing/dubov`       | Pairer | TopSeedColor, ForbiddenPairs, TotalRounds                                                |
| `pairing/lim`         | Pairer | TopSeedColor, ForbiddenPairs, MaxiTournament                                             |
| `pairing/doubleswiss` | Pairer | TopSeedColor, ForbiddenPairs, TotalRounds                                                |
| `pairing/team`        | Pairer | TopSeedColor, ForbiddenPairs, TotalRounds, ColorPreferenceType, PrimaryScore             |
| `pairing/keizer`      | Pairer | AllowRepeatPairings, MinRoundsBetweenRepeats, ScoringOptions                             |
| `pairing/roundrobin`  | Pairer | Cycles, ColorBalance, SwapLastTwoRounds                                                  |
| `scoring/standard`    | Scorer | PointWin, PointDraw, PointLoss, PointBye, PointForfeitWin, PointForfeitLoss, PointAbsent |
| `scoring/keizer`      | Scorer | 24 velden voor waardenummers, fracties, vaste waarden, verval en limieten                |
| `scoring/football`    | Scorer | Zelfde als standaard (andere standaardwaarden: 3-1-0)                                    |
