---
title: "Voetbalscoring"
linkTitle: "Voetbal"
weight: 3
description: "Een 3-1-0-scoresysteem dat winst zwaarder beloont dan remise."
---

Voetbalscoring leent van het voetbal: 3 punten voor winst, 1 voor remise, 0 voor verlies. Dit verschuift de prikkelstructuur ten opzichte van standaard 1-half-0-scoring -- een overwinning is nu drie remises waard in plaats van twee, wat passief spel ontmoedigt en beslissende resultaten beloont.

De implementatie is een dunne wrapper rond de [standaardscorer](/docs/scoring/standard/). Hij past voetbalspecifieke standaardwaarden toe en delegeert vervolgens alle scorelogica aan de standaard-engine. Dit betekent dat het algoritme, de rangschikkingsregels en de resultaatverwerking identiek zijn aan standaardscoring; alleen de puntwaarden verschillen.

## Wanneer gebruiken

- **Clubevenementen die remises willen ontmoedigen.** Bij standaardscoring zijn twee remises gelijk aan één overwinning. Bij voetbalscoring zijn drie remises gelijk aan één overwinning, waardoor remises aanzienlijk minder aantrekkelijk worden.
- **Informele of snelschaaktoernooien.** Voetbalscoring is populair bij recreatieve en snelschaakevenementen waar beslissende partijen een dynamischere toernooisfeer creëren.
- **Elk format.** Voetbalscoring werkt met Zwitsers, round-robin of elk ander indelingssysteem -- het is puur een scoreverandering.

Voetbalscoring wordt niet gebruikt bij FIDE-gewaarmerkte evenementen.

## Configuratie

### CLI

Geef opties mee via het TRF `XXY`-veld of via `--config`:

```bash
chesspairing pair --config '{"scoring": {"pointWin": 3.0, "pointDraw": 1.0}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/football"

// Met alle standaardwaarden (3-1-0).
scorer := football.New(standard.Options{})

// Specifieke waarden overschrijven.
scorer := football.New(standard.Options{
    PointDraw: chesspairing.Float64Ptr(0.5),
})

// Vanuit een generieke map.
scorer := football.NewFromMap(map[string]any{
    "pointDraw": 0.5,
})

// De Scorer-interface gebruiken.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

Het type `Scorer` voldoet aan de `chesspairing.Scorer`-interface bij compilatie:

```go
var _ chesspairing.Scorer = (*football.Scorer)(nil)
```

Merk op dat `football.New()` en `football.NewFromMap()` beide `standard.Options` accepteren -- dezelfde options-struct die de standaardscorer gebruikt. Nil-velden krijgen voetbalstandaardwaarden in plaats van standaardwaarden.

### Optiereferentie

Voetbalscoring gebruikt dezelfde `standard.Options`-struct met andere standaardwaarden:

| Veld               | JSON-sleutel       | Voetbal-default | Standaard-default | Omschrijving                       |
| ------------------ | ------------------ | --------------- | ----------------- | ---------------------------------- |
| `PointWin`         | `pointWin`         | 3.0             | 1.0               | Punten voor een partijwinst.       |
| `PointDraw`        | `pointDraw`        | 1.0             | 0.5               | Punten voor remise.                |
| `PointLoss`        | `pointLoss`        | 0.0             | 0.0               | Punten voor verlies.               |
| `PointBye`         | `pointBye`         | 3.0             | 1.0               | Punten voor een indelings-bye (PAB). |
| `PointForfeitWin`  | `pointForfeitWin`  | 3.0             | 1.0               | Punten voor winst door forfait.    |
| `PointForfeitLoss` | `pointForfeitLoss` | 0.0             | 0.0               | Punten voor verlies door forfait.  |
| `PointAbsent`      | `pointAbsent`      | 0.0             | 0.0               | Punten bij afwezigheid.            |

Alle velden zijn `*float64`. Als je een waarde expliciet instelt, heeft die voorrang op de voetbalstandaardwaarde. Zo geeft het instellen van `PointDraw` op 0.5 een 3-0.5-0-systeem.

## Hoe het werkt

Voetbalscoring delegeert volledig aan de standaard-score-engine. De methoden `Score()` en `PointsForResult()` roepen de onderliggende `standard.Scorer`-instantie aan.

Het enige verschil zit in de standaardinitialisatie: waar `standard.Options.WithDefaults()` nil-velden vult met 1-half-0-waarden, vult de voetbalscorer nil-velden met 3-1-0-waarden voordat de opties aan de standaard-engine worden doorgegeven. Zodra de standaardwaarden zijn toegepast, regelt de standaard-engine alles -- partijverwerking, bye-afhandeling, afwezigheidsdetectie en rangschikking.

Het algoritme is dezelfde enkele-doorgang-aanpak als beschreven bij [standaardscoring](/docs/scoring/standard/):

1. Initialiseer nulscores voor alle actieve spelers.
2. Verwerk elke ronde: score partijen, score byes, detecteer afwezigheid.
3. Rangschik op score aflopend, rating aflopend, weergavenaam oplopend.

## Voorbeelden

### Standaard voetbalscoring

```go
scorer := football.New(standard.Options{})
scores, _ := scorer.Score(ctx, &state)
// Speler met 3 winst, 1 remise, 1 verlies: 3×3.0 + 1×1.0 + 1×0.0 = 10.0
```

### Aangepaste remisewaarde

Verlaag de remisewaarde om overwinningen nog dominanter te maken:

```go
scorer := football.New(standard.Options{
    PointDraw: chesspairing.Float64Ptr(0.5),
})
// Een winst (3.0) is nu zes remises waard (6×0.5 = 3.0).
```

### Vergelijking met standaardscoring

Neem een speler met 5 overwinningen, 3 remises en 2 nederlagen:

| Systeem   | Berekening               | Totaal |
| --------- | ------------------------ | ------ |
| Standaard | 5(1.0) + 3(0.5) + 2(0.0) | 6.5    |
| Voetbal   | 5(3.0) + 3(1.0) + 2(0.0) | 18.0   |

De onderlinge stand tussen spelers blijft gelijk wanneer alle spelers evenveel partijen hebben gespeeld. Waar voetbalscoring de uitkomst verandert, is wanneer spelers een verschillende winst/remiseverhouding hebben. Een speler met 4 overwinningen en 4 remises (4 winst en 2 verlies in standaard: 5.0 vs 6.0) krijgt 16.0 in voetbal, terwijl een speler met 6 remises en 2 overwinningen 8.0 krijgt -- een veel groter verschil dan bij standaardscoring.

## Gerelateerd

- [Standaardscoring](/docs/scoring/standard/) -- de onderliggende engine en de volledige optiereferentie
- [Scoreconcepten](/docs/concepts/scoring/) -- overzicht van alle drie de scoresystemen
- [Keizerscoring](/docs/scoring/keizer/) -- het op rangschikking gebaseerde alternatief
- [Scorer-interface](/docs/api/scorer/) -- API-referentie voor de `Scorer`-interface
