---
title: "Options Pattern"
linkTitle: "Options"
weight: 6
description: "The pointer-field Options pattern used by all engines — WithDefaults, ParseOptions, NewFromMap."
---

Every engine in chesspairing (pairers and scorers) uses the same Options pattern for configuration. This page explains the design and the standard methods that every engine provides.

## The nil-means-default convention

Each engine defines an `Options` struct where configurable fields are **pointers**. A nil pointer means "use the engine's default value." This lets callers override specific fields while inheriting defaults for everything else.

Here is the Dutch pairer's Options struct as an example:

```go
type Options struct {
    Acceleration   *string    `json:"acceleration,omitempty"`
    TopSeedColor   *string    `json:"topSeedColor,omitempty"`
    ForbiddenPairs [][]string `json:"forbiddenPairs,omitempty"`
}
```

Setting `Acceleration` to `nil` means the Dutch pairer uses its built-in default (`"none"`). Setting it to a pointer to `"baku"` explicitly enables Baku acceleration.

### Why pointers?

Without pointers, there is no way to distinguish "the caller set this to false" from "the caller didn't set this at all." For example, a `bool` field defaults to `false` in Go -- you cannot tell whether the caller intentionally set it to `false` or simply omitted it.

With pointers, `nil` unambiguously means "not specified, use default." A non-nil pointer always represents an explicit caller choice, even when the value happens to match the default.

## Standard methods

Every engine provides three standard methods for working with options.

### WithDefaults

```go
func (o Options) WithDefaults() Options
```

Returns a copy of the Options with nil fields replaced by the engine's defaults. Does **not** modify the receiver. This is called internally by the engine constructor, so callers rarely need to call it directly.

**Exception:** the Keizer scorer's `WithDefaults` takes a `playerCount int` parameter because the default value number base depends on the number of players. It is called inside `Score()` when the player count is known, not by the constructor.

Example (Dutch):

```go
opts := dutch.Options{
    Acceleration: nil, // will become "none"
    TopSeedColor: nil, // will become "auto"
}
filled := opts.WithDefaults()
// *filled.Acceleration == "none"
// *filled.TopSeedColor == "auto"
```

### ParseOptions

```go
func ParseOptions(m map[string]any) Options
```

Package-level function that parses a generic `map[string]any` (typically from JSON config, TRF extended data, or CLI flags) into a typed Options struct. Returns an Options with only the recognized keys populated; unrecognized keys are silently ignored.

Type coercion is handled using the root package helper functions (see [Helper functions](#helper-functions) below).

### NewFromMap

```go
func NewFromMap(m map[string]any) *Pairer  // or *Scorer
```

Package-level constructor that creates a fully initialized engine from a generic options map. This is the factory entry point used by the CLI and the `trf` package for generic configuration-driven instantiation.

Internally it calls `ParseOptions(m)` followed by `New(opts)`, which applies `WithDefaults()`.

```go
// These are equivalent:
p1 := dutch.NewFromMap(map[string]any{"acceleration": "baku"})

opts := dutch.ParseOptions(map[string]any{"acceleration": "baku"})
p2 := dutch.New(opts)
```

## Helper functions

The root package (`chesspairing`) provides pointer constructors and type-safe map extraction helpers in `options_helpers.go`. These are used by `ParseOptions` implementations and are available to callers for direct Options construction.

### Pointer constructors

```go
func Float64Ptr(v float64) *float64
func IntPtr(v int) *int
func BoolPtr(v bool) *bool
func StringPtr(v string) *string
```

Create a pointer to the given value. Used when constructing Options structs directly:

```go
opts := keizer.Options{
    WinFraction:  chesspairing.Float64Ptr(1.0),
    SelfVictory:  chesspairing.BoolPtr(false),
}
```

### Map extraction

```go
func GetFloat64(m map[string]any, key string) (float64, bool)
func GetInt(m map[string]any, key string) (int, bool)
func GetBool(m map[string]any, key string) (bool, bool)
func GetString(m map[string]any, key string) (string, bool)
```

Each function extracts a typed value from a generic map. The second return value indicates whether the key was found with a compatible type.

Type coercion rules:

- **GetFloat64**: accepts `float64`, `int`, and `int64` values.
- **GetInt**: accepts `int`, `int64`, and `float64` values (truncates to int).
- **GetBool**: accepts `bool` values only.
- **GetString**: accepts `string` values only.

All three return the zero value and `false` if the key is missing or has an incompatible type.

## Usage patterns

### Direct construction with specific overrides

```go
p := dutch.New(dutch.Options{
    Acceleration: chesspairing.StringPtr("baku"),
    // TopSeedColor: nil -- uses default ("auto")
})
```

### From a configuration map

```go
opts := map[string]any{
    "acceleration": "baku",
    "topSeedColor": "white",
}
p := dutch.NewFromMap(opts)
```

### Scorer example (Keizer)

```go
s := keizer.NewFromMap(map[string]any{
    "winFraction":  1.0,
    "drawFraction": 0.5,
    "selfVictory":  false,
})
```

## Engines implementing this pattern

Every pairer and scorer package follows this convention:

| Package               | Type   | Key options                                                                              |
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
| `scoring/keizer`      | Scorer | 24 fields covering value numbers, fractions, fixed values, decay, and limits             |
| `scoring/football`    | Scorer | Same as standard (different defaults: 3-1-0)                                             |
