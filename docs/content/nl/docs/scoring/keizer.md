---
title: "Keizerscoring"
linkTitle: "Keizer"
weight: 2
description: "Iteratieve, op rangschikking gebaseerde scoring waarbij punten afhangen van de sterkte van de tegenstander — convergeert in maximaal 20 iteraties."
---

Keizerscoring is een op rangschikking gebaseerd systeem dat populair is bij clubtoernooien in België en Nederland. Het centrale idee: een sterke tegenstander verslaan levert meer punten op dan een zwakke verslaan. Elke speler krijgt een waarderingsgetal op basis van de huidige rang, en partijenscores worden berekend als fracties van het waarderingsgetal van de tegenstander. Omdat waarderingsgetallen afhangen van rangschikkingen en rangschikkingen van scores, moet het systeem itereren tot het convergeert.

De engine werkt intern met x2-gehele-getalrekenkunde om drijvende-kommafouten te elimineren en tegelijkertijd halve-puntresolutie te behouden. Convergentie treedt doorgaans op binnen 3-5 iteraties, met een harde limiet van 20 en oscillatiedetectie om terminatie te garanderen.

## Wanneer gebruiken

- **Interne clubcompetities.** Keizer is ontworpen voor doorlopende clubcompetities waar spelers regelmatig rondes missen. Het afwezigheidssysteem (fracties, limieten, verval, clubverplichtingen) is veel rijker dan bij standaardscoring.
- **Toernooien waar de sterkte van de tegenstander ertoe moet doen.** Bij standaardscoring is een overwinning altijd 1 punt waard, ongeacht de tegenstander. Bij Keizer levert het verslaan van de nummer 1 aanzienlijk meer op dan het verslaan van de laatstgeplaatste.
- **Gecombineerde indeling en scoring.** De Keizer-indeling gebruikt de Keizerscorer intern om spelers te rangschikken vóór de indeling, wat een hecht geïntegreerd systeem oplevert. De scorer is echter ook onafhankelijk te gebruiken met elk indelingssysteem.

Keizerscoring is niet geschikt voor FIDE-gewaarmerkte evenementen, die standaard 1-half-0-scoring vereisen.

## Configuratie

### CLI

Geef Keizer-scoreopties mee via het TRF `XXY`-veld of via `--config`:

```bash
chesspairing pair --config '{"scoring": {"winFraction": 1.0, "drawFraction": 0.5, "selfVictory": true}}' tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/scoring/keizer"

// Met expliciete opties (nil-velden gebruiken standaardwaarden).
scorer := keizer.New(keizer.Options{
    LossFraction: chesspairing.Float64Ptr(1.0 / 6.0),
    AbsenceLimit: chesspairing.IntPtr(0),
})

// Vanuit een generieke map (bijv. geparsed uit JSON-configuratie).
scorer := keizer.NewFromMap(map[string]any{
    "lossFraction": 1.0 / 6.0,
    "absenceLimit": 0,
})

// De Scorer-interface gebruiken.
scores, err := scorer.Score(ctx, &state)
points := scorer.PointsForResult(result, rctx)
```

Het type `Scorer` voldoet aan de `chesspairing.Scorer`-interface bij compilatie:

```go
var _ chesspairing.Scorer = (*keizer.Scorer)(nil)
```

### Optiereferentie

De opties zijn georganiseerd in vijf groepen. Alle pointervelden gebruiken `nil` als "gebruik de standaardwaarde".

#### Waarderingsgetallen

Deze bepalen hoe waarderingsgetallen worden toegewezen vanuit rangschikkingen. De speler op rang _r_ krijgt: `ValueNumberBase - (r-1) * ValueNumberStep`.

| Veld              | Type   | JSON-sleutel      | Default | Omschrijving                                                              |
| ----------------- | ------ | ----------------- | ------- | ------------------------------------------------------------------------- |
| `ValueNumberBase` | `*int` | `valueNumberBase` | N       | Waarderingsgetal voor de hoogstgerangschikte. N = aantal actieve spelers. |
| `ValueNumberStep` | `*int` | `valueNumberStep` | 1       | Afname per rangpositie.                                                   |

#### Partijfracties

Fracties van het waarderingsgetal van de **tegenstander** die worden toegekend voor partijresultaten.

| Veld                    | Type       | JSON-sleutel            | Default | Omschrijving                                  |
| ----------------------- | ---------- | ----------------------- | ------- | --------------------------------------------- |
| `WinFraction`           | `*float64` | `winFraction`           | 1.0     | Fractie voor winst.                           |
| `DrawFraction`          | `*float64` | `drawFraction`          | 0.5     | Fractie voor remise.                          |
| `LossFraction`          | `*float64` | `lossFraction`          | 0.0     | Fractie voor verlies.                         |
| `ForfeitWinFraction`    | `*float64` | `forfeitWinFraction`    | 1.0     | Fractie voor winst door forfait.              |
| `ForfeitLossFraction`   | `*float64` | `forfeitLossFraction`   | 0.0     | Fractie voor verlies door forfait.            |
| `DoubleForfeitFraction` | `*float64` | `doubleForfeitFraction` | 0.0     | Fractie voor een dubbel forfait (per speler). |

#### Niet-partijfracties

Fracties van het **eigen** waarderingsgetal die worden toegekend bij niet-partijsituaties.

| Veld                     | Type       | JSON-sleutel             | Default | Omschrijving                                     |
| ------------------------ | ---------- | ------------------------ | ------- | ------------------------------------------------ |
| `ByeValueFraction`       | `*float64` | `byeValueFraction`       | 0.50    | Fractie voor een indelings-bye (PAB).            |
| `HalfByeFraction`        | `*float64` | `halfByeFraction`        | 0.50    | Fractie voor een halve-punt-bye.                 |
| `ZeroByeFraction`        | `*float64` | `zeroByeFraction`        | 0.0     | Fractie voor een nulpunt-bye.                    |
| `AbsentPenaltyFraction`  | `*float64` | `absentPenaltyFraction`  | 0.35    | Fractie voor een ongeoorloofde afwezigheid.      |
| `ExcusedAbsentFraction`  | `*float64` | `excusedAbsentFraction`  | 0.35    | Fractie voor een geoorloofde afwezigheid.        |
| `ClubCommitmentFraction` | `*float64` | `clubCommitmentFraction` | 0.70    | Fractie voor afwezigheid wegens interclubplicht. |

#### Vaste-waarde-overschrijvingen

Wanneer ingesteld (niet-nil), vervangen deze de bijbehorende fractieberekening door een vaste score. Waarden zijn in reële eenheden (niet x2). Laat op `nil` staan om de fractieberekening te gebruiken.

| Veld                       | Type   | JSON-sleutel               | Default | Omschrijving                                |
| -------------------------- | ------ | -------------------------- | ------- | ------------------------------------------- |
| `ByeFixedValue`            | `*int` | `byeFixedValue`            | nil     | Vaste score voor PAB.                       |
| `HalfByeFixedValue`        | `*int` | `halfByeFixedValue`        | nil     | Vaste score voor halve-punt-bye.            |
| `ZeroByeFixedValue`        | `*int` | `zeroByeFixedValue`        | nil     | Vaste score voor nulpunt-bye.               |
| `AbsentFixedValue`         | `*int` | `absentFixedValue`         | nil     | Vaste score voor ongeoorloofde afwezigheid. |
| `ExcusedAbsentFixedValue`  | `*int` | `excusedAbsentFixedValue`  | nil     | Vaste score voor geoorloofde afwezigheid.   |
| `ClubCommitmentFixedValue` | `*int` | `clubCommitmentFixedValue` | nil     | Vaste score voor clubverplichting.          |

#### Gedragsopties

| Veld               | Type       | JSON-sleutel       | Default | Omschrijving                                                                                                                                                 |
| ------------------ | ---------- | ------------------ | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `SelfVictory`      | `*bool`    | `selfVictory`      | true    | Het eigen waarderingsgetal eenmalig bij het totaal optellen (niet per ronde).                                                                                |
| `AbsenceLimit`     | `*int`     | `absenceLimit`     | 5       | Maximaal aantal afwezigheden dat punten oplevert. Daarboven scoren afwezigheden 0. Clubverplichtingen zijn vrijgesteld. 0 = onbeperkt.                       |
| `AbsenceDecay`     | `*bool`    | `absenceDecay`     | false   | Halveer de afwezigheidsscore bij elke volgende afwezigheid (1e = volledig, 2e = helft, 3e = kwart, ...). Clubverplichtingen zijn vrijgesteld.                |
| `Frozen`           | `*bool`    | `frozen`           | false   | Schakel iteratieve convergentie uit. Elke ronde wordt eenmalig gescoord met de rangschikking op dat moment, en eerdere rondes worden nooit opnieuw berekend. |
| `LateJoinHandicap` | `*float64` | `lateJoinHandicap` | 0       | Vaste score per gemiste ronde voor toetreding. Vereist `PlayerEntry.JoinedRound`. Niet onderhevig aan `AbsenceLimit` of `AbsenceDecay`.                      |

## Hoe het werkt

### Score()

Keizerscoring is iteratief vanwege een circulaire afhankelijkheid: scores hangen af van waarderingsgetallen, waarderingsgetallen van rangschikkingen, en rangschikkingen van scores. Het algoritme lost dit op door herhaald herberekenen:

1. **Initiële rangschikking.** Rangschik alle actieve spelers op rating aflopend, dan op weergavenaam oplopend.

2. **Itereer** (tot 20 keer):

   a. **Reset scores.** Zet alle x2-scores op nul.

   b. **Bereken waarderingsgetallen.** Wijs vanuit de huidige rangschikking elke speler een waarderingsgetal toe: `ValueNumberBase - (rang-1) * ValueNumberStep`. De hoogstgerangschikte krijgt het hoogste getal.

   c. **Score alle rondes.** Verwerk per ronde partijen, byes, afwezigheden en nagevorderde rondes:
   - _Partijen:_ punten = `round(tegenstander_waarde * fractie * 2)` met x2-gehele-getalrekenkunde. Bijvoorbeeld, winst tegen een speler met waarde 20 bij WinFraction=1.0 levert `20 * 1.0 * 2 = 40` op in x2-eenheden.
   - _Byes:_ punten = `round(eigen_waarde * fractie * 2)`, of `vaste_waarde * 2` wanneer een vaste-waarde-overschrijving is ingesteld. Clubverplichtingen zijn vrijgesteld van de afwezigheidslimiet en het verval.
   - _Afwezigheden:_ hetzelfde als byes, met `AbsentPenaltyFraction` of `AbsentFixedValue`, onderhevig aan de afwezigheidslimiet en het verval. Geoorloofde afwezigheden tellen mee voor de limiet; clubverplichtingen niet.
   - _Nagevorderde rondes:_ voor spelers met `JoinedRound > 1` scoren rondes voor het toetredingsmoment `LateJoinHandicap` als vaste waarde, in plaats van de afwezigheidsberekening. Deze rondes tellen niet mee voor de afwezigheidslimiet en worden niet beïnvloed door verval.

   d. **Zelfoverwinning.** Indien ingeschakeld, tel `eigen_waarde * 2` eenmalig op bij het x2-totaal van elke speler (niet per ronde).

   e. **Herrangschik.** Sorteer spelers op x2-score aflopend, rating aflopend, weergavenaam oplopend.

   f. **Controleer convergentie.** Als de rangschikking ongewijzigd is ten opzichte van de vorige iteratie, stop.

   g. **Controleer op oscillatie.** Als de rangschikking overeenkomt met die van twee iteraties terug (een 2-cyclus), middel de x2-scores van de laatste twee iteraties, herrangschik en stop. Dit vangt gevallen op waarin twee spelers met zeer vergelijkbare scores voortdurend van positie wisselen.

3. **Converteer naar reële scores.** Deel alle x2-scores door 2 voor de eindwaarden.

### Bevroren modus

Wanneer `Frozen` op `true` staat, wordt de iteratieve lus vervangen door een sequentiële doorgang door de rondes. Elke ronde wordt eenmalig gescoord met de rangschikking zoals die op dat moment was, waarna de rangschikking wordt bijgewerkt. Eerdere rondes worden nooit opnieuw berekend wanneer latere resultaten de stand verschuiven.

De volgorde:

1. Begin met de initiële op rating gebaseerde rangschikking.
2. Bereken per ronde de waarderingsgetallen vanuit de huidige rangschikking, score partijen/byes/afwezigheden voor die ronde, en herrangschik.
3. Tel na alle rondes de zelfoverwinning op (indien ingeschakeld) op basis van de eindrangschikking.

Dit levert andere resultaten op dan de standaard iteratieve modus. In de standaardmodus worden alle rondes achteraf opnieuw gescoord met de geconvergeerde rangschikking, zodat een ronde-1-overwinning de eindwaarde van de tegenstander waard is. In de bevroren modus is diezelfde overwinning de waarde waard die de tegenstander op dat moment had -- die hoger of lager kan zijn geweest voordat latere rondes de stand verschoven.

De bevroren modus is nuttig voor clubs die willen dat scores het verloop van het seizoen weerspiegelen, in plaats van de geschiedenis achteraf te herschrijven vanuit het eindpunt.

### Waarom x2-gehele-getalrekenkunde

Keizerscores zijn sommen van producten van gehele getallen en fracties. Herhaalde drijvende-kommaoptelling zou afrondingsfouten ophopen die de rangschikking kunnen beïnvloeden. De x2-aanpak werkt intern met verdubbelde gehele getallen (dus 0.5 reële punten = 1 in x2-eenheden), waardoor drift wordt geëlimineerd terwijl halve-puntresolutie behouden blijft. De conversie naar reële getallen vindt pas plaats bij de uiteindelijke uitvoer.

### Afwezigheidsregels

- **Afwezigheidslimiet.** Na `AbsenceLimit` afwezigheden scoren alle volgende afwezigheden 0. Dit voorkomt dat spelers die nauwelijks deelnemen toch betekenisvolle scores opbouwen.
- **Afwezigheidsverval.** Wanneer ingeschakeld, levert elke volgende afwezigheid de helft van de vorige op: 1e = volledige fractie, 2e = fractie/2, 3e = fractie/4, enzovoort (geïmplementeerd als een rechtse bitshift op de x2-waarde).
- **Clubverplichtingen** zijn altijd vrijgesteld van zowel de limiet als het verval. Een speler die rondes mist vanwege interclubplicht wordt niet gestraft zoals bij een gewone afwezigheid.
- **Geoorloofde afwezigheden** ontvangen hun eigen fractie (`ExcusedAbsentFraction`) maar tellen wel mee voor de afwezigheidslimiet en het verval.
- **Nagevorderde spelers.** Wanneer een speler `JoinedRound > 1` heeft, worden rondes voor het toetredingsmoment gescoord met `LateJoinHandicap` als vaste waarde in plaats van de normale afwezigheidslogica. Deze rondes omzeilen de afwezigheidslimiet en het verval volledig, zodat de werkelijke afwezigheden van een nagevorderde speler (na toetreding) vanaf nul worden geteld.

## Variantpresets

Verschillende bekende Keizer-varianten kunnen worden geconfigureerd door specifieke opties in te stellen:

| Variant                      | Belangrijkste verschillen met standaardwaarden                                                                                                                                                     |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **KeizerForClubs**           | Alle standaardwaarden. De meest gebruikte variant.                                                                                                                                                 |
| **Classic KNSB**             | `ByeValueFraction`=4/6, `AbsentPenaltyFraction`=2/6, `ClubCommitmentFraction`=2/3, `ExcusedAbsentFraction`=2/6, `AbsenceLimit`=5. Gebaseerd op het traditionele systeem van de KNSB.               |
| **FreeKeizer**               | `LossFraction`=1/6, `ByeValueFraction`=4/6, `AbsentPenaltyFraction`=2/6, `AbsenceLimit`=5. Voegt een "taaiheidsbonus" toe waarbij verlies tegen sterke tegenstanders een kleine beloning oplevert. |
| **Geen zelfoverwinning**     | `SelfVictory`=false. Verwijdert de deelnamebonus.                                                                                                                                                  |
| **Vaste afwezigheden**       | `AbsentFixedValue`=15, `ExcusedAbsentFixedValue`=15, `ClubCommitmentFixedValue`=25. Gebruikt absolute waarden in plaats van fracties.                                                              |
| **Vervallende afwezigheden** | `AbsenceDecay`=true, `AbsenceLimit`=0. Elke afwezigheid levert minder op dan de vorige, zonder harde limiet.                                                                                       |

## Voorbeelden

### Standaard KeizerForClubs

```go
scorer := keizer.New(keizer.Options{})
scores, _ := scorer.Score(ctx, &state)
```

Met 10 spelers heeft de hoogstgerangschikte waarderingsgetal 10. Als die speler de nummer 3 verslaat (waarde 8): punten = 8 _ 1.0 = 8.0. Bij remise: 8 _ 0.5 = 4.0. Zelfoverwinning telt het eigen getal (10) eenmalig op.

### FreeKeizer met taaiheidsbonus

```go
scorer := keizer.New(keizer.Options{
    LossFraction:          chesspairing.Float64Ptr(1.0 / 6.0),
    ByeValueFraction:      chesspairing.Float64Ptr(4.0 / 6.0),
    AbsentPenaltyFraction: chesspairing.Float64Ptr(2.0 / 6.0),
    AbsenceLimit:          chesspairing.IntPtr(5),
})
```

Verlies tegen de nummer 1 (waarde 10) levert nu 10 \* 1/6 = 1.5 punten op (afgerond op het x2-raster) in plaats van 0.

### Vaste afwezigheidswaarden

```go
scorer := keizer.New(keizer.Options{
    AbsentFixedValue:         chesspairing.IntPtr(15),
    ExcusedAbsentFixedValue:  chesspairing.IntPtr(15),
    ClubCommitmentFixedValue: chesspairing.IntPtr(25),
})
```

Alle spelers ontvangen dezelfde vaste score voor afwezigheid, ongeacht hun rang.

## Gerelateerd

- [Scoreconcepten](/docs/concepts/scoring/) -- overzicht van alle drie de scoresystemen en hun interactie met indelen
- [Keizer-convergentiealgoritme](/docs/algorithms/keizer-convergence/) -- gedetailleerde analyse van de iteratieve convergentie en oscillatiedetectie
- [Keizer-indelingssysteem](/docs/pairing-systems/keizer/) -- de indeling die intern Keizerscoring gebruikt voor rangschikking
- [Standaardscoring](/docs/scoring/standard/) -- het alternatief met vaste punten
- [Byes](/docs/concepts/byes/) -- bye-typen en hun scoring in alle systemen
- [Scorer-interface](/docs/api/scorer/) -- API-referentie voor de `Scorer`-interface
