---
title: "Standaardscoring"
linkTitle: "Standaard"
weight: 1
description: "Het klassieke 1-½-0-scoresysteem met configureerbare puntwaarden voor winst, remise, byes, forfaits en afwezigheid."
---

Standaardscoring kent een vast aantal punten toe per partijuitslag: 1 voor winst, 0.5 voor remise, 0 voor verlies. Elke puntwaarde is configureerbaar, maar de standaardwaarden volgen de FIDE-conventies die in vrijwel alle gewaarmerkte Zwitserse en round-robin-evenementen worden gehanteerd.

Omdat punten alleen afhangen van het resultaat -- niet van de tegenstander -- vereist standaardscoring geen iteratie en levert het een deterministische stand op na één enkele doorgang door de resultaten.

## Wanneer gebruiken

Standaardscoring is de juiste keuze voor de meeste toernooien:

- **FIDE-gewaarmerkte evenementen.** Vereist door de FIDE-reglementen voor alle officieel gewaarmerkte competities.
- **Zwitserse en round-robin-toernooien.** Het verwachte scoresysteem voor deze formats.
- **Elk evenement waar eenvoud belangrijk is.** Punten zijn makkelijk uit te leggen en te controleren: een overwinning is altijd evenveel waard, ongeacht tegen wie je speelt.

Zelfs wanneer een toernooi een ander publiek scoresysteem gebruikt (zoals Keizer), gebruiken de Zwitserse indelingsmodules intern standaardscoring voor het vormen van scoregroepen.

## Configuratie

### CLI

Geef scoreopties mee via het TRF `XXY`-veld of via `--config`:

```bash
chesspairing pair --config '{"scoring": {"pointWin": 1.0, "pointDraw": 0.5, "pointBye": 0.5}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/standard"

// Met expliciete opties (nil-velden gebruiken standaardwaarden).
scorer := standard.New(standard.Options{
    PointBye: chesspairing.Float64Ptr(0.5),
})

// Vanuit een generieke map (bijv. geparsed uit JSON-configuratie).
scorer := standard.NewFromMap(map[string]any{
    "pointBye": 0.5,
})

// De Scorer-interface gebruiken.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

Het type `Scorer` voldoet aan de `chesspairing.Scorer`-interface bij compilatie:

```go
var _ chesspairing.Scorer = (*standard.Scorer)(nil)
```

### Optiereferentie

Alle velden zijn `*float64`. Een `nil`-waarde betekent "gebruik de standaardwaarde." Dit maakt onderscheid tussen "niet geconfigureerd" en "expliciet op nul gezet."

| Veld                  | JSON-sleutel          | Default | Omschrijving                                                                              |
| --------------------- | --------------------- | ------- | ----------------------------------------------------------------------------------------- |
| `PointWin`            | `pointWin`            | 1.0     | Punten voor een partijwinst.                                                              |
| `PointDraw`           | `pointDraw`           | 0.5     | Punten voor remise. Ook gebruikt voor `ByeHalf`.                                          |
| `PointLoss`           | `pointLoss`           | 0.0     | Punten voor verlies. Ook gebruikt voor `ByeZero`.                                         |
| `PointBye`            | `pointBye`            | 1.0     | Punten voor een indelings-bye (`ByePAB`).                                                 |
| `PointForfeitWin`     | `pointForfeitWin`     | 1.0     | Punten voor winst door forfait.                                                           |
| `PointForfeitLoss`    | `pointForfeitLoss`    | 0.0     | Punten voor verlies door forfait.                                                         |
| `PointAbsent`         | `pointAbsent`         | 0.0     | Punten bij ongeoorloofde afwezigheid (`ByeAbsent`, of geen partij en geen bye).           |
| `PointExcused`        | `pointExcused`        | 0.0     | Punten bij verontschuldigde afwezigheid (`ByeExcused`).                                   |
| `PointClubCommitment` | `pointClubCommitment` | 0.0     | Punten bij afwezigheid door clubverplichting (`ByeClubCommitment`).                       |

## Hoe het werkt

### Score()

De `Score()`-methode maakt één doorgang door alle rondes om punten op te tellen:

1. **Initialiseer.** Maak een nulscore aan voor elke actieve speler.

2. **Verwerk elke ronde.** Voor iedere ronde in het toernooi:
   - **Partijen.** Elke partij wordt gescoord volgens het resultaat:
     - _Dubbel forfait_ -- beide spelers ontvangen 0 (de partij wordt behandeld alsof die nooit heeft plaatsgevonden).
     - _Enkel forfait_ -- de winnaar ontvangt `PointForfeitWin`, de verliezer ontvangt `PointForfeitLoss`.
     - _Regulier resultaat_ -- `PointWin`/`PointDraw`/`PointLoss` naar gelang van toepassing. Lopende partijen (`*`) leveren niets op.

   - **Byes.** Elke bye wordt gescoord per type:
     - `ByePAB` -- `PointBye`
     - `ByeHalf` -- `PointDraw`
     - `ByeZero` -- `PointLoss`
     - `ByeAbsent` -- `PointAbsent`
     - `ByeExcused` -- `PointExcused`
     - `ByeClubCommitment` -- `PointClubCommitment`

   - **Afwezigheidsdetectie.** Elke actieve speler die in een ronde noch een partij heeft gespeeld noch een bye heeft ontvangen, wordt als afwezig beschouwd en krijgt `PointAbsent`.

3. **Rangschik.** Spelers worden gesorteerd op score aflopend, dan rating aflopend, dan weergavenaam oplopend.

### PointsForResult()

Retourneert de puntwaarde voor een enkel resultaat. De methode werkt in deze volgorde:

1. Als `rctx.ByeType` niet-nil is -- retourneer de bijbehorende `PointBye` / `PointDraw` / `PointLoss` / `PointAbsent` / `PointExcused` / `PointClubCommitment`.
2. Anders, als `result.IsForfeit()` -- retourneer `PointForfeitWin` voor een forfaitwinst, anders `PointForfeitLoss`.
3. Anders, voor een reguliere partij -- retourneer `PointWin`, `PointDraw`, of `0` (verlies / lopend).

De `ByeType`-pointer heeft voorrang: bij een bye wordt het `Result` van een gespeelde partij genegeerd.

## Voorbeelden

### Standaard FIDE-scoring

```go
scorer := standard.New(standard.Options{})
scores, _ := scorer.Score(ctx, &state)
// Speler met 3 winst, 1 remise, 1 verlies: 3×1.0 + 1×0.5 + 1×0.0 = 3.5
```

### Halve-punt-PAB

Sommige organisatoren geven de voorkeur aan een halve-punt-bye in plaats van een heel punt:

```go
scorer := standard.New(standard.Options{
    PointBye: chesspairing.Float64Ptr(0.5),
})
```

### Afwezigheidsstraffen

Ken een negatieve score toe voor ongeoorloofde afwezigheid:

```go
scorer := standard.New(standard.Options{
    PointAbsent: chesspairing.Float64Ptr(-1.0),
})
```

## Gerelateerd

- [Scoreconcepten](/docs/concepts/scoring/) -- overzicht van alle drie de scoresystemen en hun interactie met indelen
- [Voetbalscoring](/docs/scoring/football/) -- de 3-1-0-variant gebouwd bovenop standaardscoring
- [Keizerscoring](/docs/scoring/keizer/) -- iteratief, op rangschikking gebaseerd alternatief
- [Byes](/docs/concepts/byes/) -- bye-typen en hoe ze worden gescoord
- [Forfaits en afwezigheid](/docs/concepts/forfeits/) -- hoe forfaitresultaten scoring en indelingshistorie beïnvloeden
- [Scorer-interface](/docs/api/scorer/) -- API-referentie voor de `Scorer`-interface
