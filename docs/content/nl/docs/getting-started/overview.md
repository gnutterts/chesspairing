---
title: "Overzicht"
linkTitle: "Overzicht"
weight: 1
description: "Wat chesspairing is en hoe de onderdelen samenwerken."
---

## Wat is chesspairing?

Chesspairing is een pure Go-module (`github.com/gnutterts/chesspairing`) die drie kernbewerkingen van toernooien afhandelt: **indeling** (bepalen wie tegen wie speelt), **scoring** (uitslagen omzetten in een ranglijst) en **tiebreaking** (spelers met gelijke score rangschikken).

De module heeft nul externe afhankelijkheden -- alles draait uitsluitend op de Go-standaardbibliotheek. Er is geen I/O, geen database en geen netwerkcommunicatie. Elke engine werkt op datastructuren in het geheugen, waardoor het eenvoudig is om de module in te bouwen in een server, desktopapplicatie of geautomatiseerde pipeline. De module is veilig voor gelijktijdig gebruik zolang elke goroutine zijn eigen toernooistaat meelevert.

De module implementeert alle zes door de FIDE goedgekeurde Zwitserse indelingssystemen (zoals gedefinieerd in het FIDE Handbook, secties C.04.3 tot en met C.04.6), plus Keizerindeling en FIDE round-robin-indeling. Je kunt elk indelingssysteem combineren met elk scoringssysteem -- bijvoorbeeld Zwitserse indeling met Keizerscoring, of round-robin met voetbalscoring. Niets dwingt een bepaalde combinatie af.

## Architectuur

Drie interfaces in het rootpakket bepalen wat de engines doen:

```go
type Pairer interface {
    Pair(ctx context.Context, state *TournamentState) (*PairingResult, error)
}

type Scorer interface {
    Score(ctx context.Context, state *TournamentState) ([]PlayerScore, error)
    PointsForResult(result GameResult, rctx ResultContext) float64
}

type TieBreaker interface {
    ID() string
    Name() string
    Compute(ctx context.Context, state *TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

Een `TournamentState` is de enige invoer voor alle drie. Het is een alleen-lezen snapshot met de spelerslijst, alle voltooide rondes (met partijuitslagen en byes), het huidige rondenummer en de indelings-/scoringsconfiguratie. De aanroepende code bouwt deze snapshot op vanuit de eigen opslag, en de engines geven resultaten terug zonder neveneffecten.

### Gegevensstroom

```text
Caller builds TournamentState
  |
  +---> Pairer.Pair()       --> PairingResult (board pairings + byes)
  |
  +---> Scorer.Score()      --> []PlayerScore (score per player, ranked)
  |
  +---> TieBreaker.Compute()--> []TieBreakValue (one numeric value per player)
```

De drie stappen zijn onafhankelijk. Je kunt `Pair` aanroepen zonder te scoren, scores berekenen zonder tiebreakers, of alle drie achter elkaar uitvoeren. Een typische toernooironde ziet er als volgt uit:

1. Bouw een `TournamentState` vanuit je databron.
2. Roep `Scorer.Score()` aan om de huidige stand te produceren (de indelingsengine gebruikt deze intern ook).
3. Roep `Pairer.Pair()` aan om de indeling voor de volgende ronde te genereren.
4. Nadat de ronde gespeeld is en de uitslagen zijn vastgelegd, roep je `Scorer.Score()` opnieuw aan en vervolgens een of meer `TieBreaker.Compute()`-aanroepen om de definitieve ranglijst op te stellen.

## Wat zit erin

### Indelingssystemen

| Systeem      | FIDE-ref | Pakket                | Opmerkingen                                             |
| ------------ | -------- | --------------------- | ------------------------------------------------------- |
| Dutch        | C.04.3   | `pairing/dutch`       | Globale Blossom-matching, Bakoe-acceleratie             |
| Burstein     | C.04.4.2 | `pairing/burstein`    | Seeding-/post-seedingrondes met oppositie-index         |
| Dubov        | C.04.4.1 | `pairing/dubov`       | ARO-gebaseerde rangschikking, 10 specifieke criteria    |
| Lim          | C.04.4.3 | `pairing/lim`         | Mediaan-eerst-verwerking, exchange-gebaseerde matching  |
| Double-Swiss | C.04.5   | `pairing/doubleswiss` | Lexicografische groepsindeling                          |
| Team Swiss   | C.04.6   | `pairing/team`        | Teamindeling met configureerbare kleurvoorkeur          |
| Keizer       | --       | `pairing/keizer`      | Top-down op Keizerscore, herhaling vermijden            |
| Round-Robin  | C.05     | `pairing/roundrobin`  | FIDE Berger-tabellen, ondersteuning voor meerdere cycli |

Elke indelingsengine implementeert de `Pairer`-interface en heeft een `Options`-struct voor systeemspecifieke configuratie. Zie de sectie [Indelingssystemen](/docs/pairing-systems/) voor gedetailleerde documentatie van elk systeem.

### Scoringssystemen

| Systeem  | Standaardpunten         | Pakket             | Opmerkingen                                                    |
| -------- | ----------------------- | ------------------ | -------------------------------------------------------------- |
| Standard | 1 -- 0,5 -- 0           | `scoring/standard` | Configureerbare puntwaarden voor winst, remise, byes, forfaits |
| Keizer   | Iteratieve convergentie | `scoring/keizer`   | Rangafhankelijke scoring met 24 configureerbare parameters     |
| Football | 3 -- 1 -- 0             | `scoring/football` | Dunne wrapper rond Standard met andere standaardwaarden        |

Elke scoringsengine implementeert de `Scorer`-interface. Zie de sectie [Scoringssystemen](/docs/scoring/) voor details over configuratie en gedrag.

### Tiebreakers

Chesspairing bevat 25 tiebreaker-implementaties die zichzelf registreren via een centraal register. Ze dekken het volledige scala aan door de FIDE erkende methoden:

| Categorie           | Tiebreakers                                                                                                            |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| Buchholz-familie    | Buchholz, Buchholz Cut-1, Buchholz Cut-2, Buchholz Median, Buchholz Median-2, Fore Buchholz, Average Opponent Buchholz |
| Onderling resultaat | Direct Encounter, Koya System                                                                                          |
| Resultaatgebaseerd  | Games Won, Rounds Won, Progressive Score, Standard Points                                                              |
| Prestatie           | Performance Rating (TPR), Performance Points (PTP), Average Opponent TPR, Average Opponent PTP, Player Rating          |
| Kleur en activiteit | Games with Black, Black Wins, Rounds Played, Games Played                                                              |
| Ratinggebaseerd     | Average Rating of Opponents (ARO)                                                                                      |
| Administratief      | Pairing Number (TPN)                                                                                                   |

Elke tiebreaker implementeert de `TieBreaker`-interface en kan worden aangeduid met zijn string-ID (bijv. `"buchholz-cut1"`, `"sonneborn-berger"`, `"direct-encounter"`). Zie de sectie [Tiebreakers](/docs/tiebreakers/) voor het volledige register en berekeningsdetails.

## Twee manieren om het te gebruiken

### Opdrachtregeltool

De `chesspairing`-CLI leest [TRF16-bestanden](/docs/formats/trf16/) (het standaard FIDE-uitwisselingsformaat voor toernooien) en produceert indelingen, ranglijsten en validatierapporten. De tool biedt acht subcommando's:

| Commando      | Doel                                                     |
| ------------- | -------------------------------------------------------- |
| `pair`        | Indeling genereren voor de volgende ronde                |
| `check`       | Bestaande indeling vergeleken met engine-uitvoer         |
| `generate`    | Bijgewerkt TRF-bestand uitvoeren met indeling toegevoegd |
| `validate`    | TRF-bestand valideren tegen meerdere profielen           |
| `standings`   | Huidige stand berekenen en tonen                         |
| `tiebreakers` | Beschikbare tiebreakers weergeven                        |
| `convert`     | TRF-bestand opnieuw serialiseren                         |
| `version`     | Versie-informatie tonen                                  |

Uitvoer is beschikbaar in vijf formaten: platte lijst (bbpPairings-compatibel), brede tabel, bordweergave, XML en JSON. Een legacy-modus biedt directe compatibiliteit met de opdrachtregelconventies van bbpPairings en JaVaFo.

Zie de [CLI-snelstart](/docs/getting-started/cli-quickstart/) om aan de slag te gaan, of de [CLI-referentie](/docs/cli/) voor volledige documentatie.

### Go-bibliotheek

Voeg de module toe aan je project:

```bash
go get github.com/gnutterts/chesspairing
```

Bouw vervolgens een `TournamentState`, kies je engines en roep ze aan:

```go
import (
    "context"
    "github.com/gnutterts/chesspairing"
    "github.com/gnutterts/chesspairing/pairing/dutch"
    "github.com/gnutterts/chesspairing/scoring/standard"
)

state := &chesspairing.TournamentState{
    Players:      players,
    Rounds:       rounds,
    CurrentRound: 3,
    PairingConfig: chesspairing.PairingConfig{
        System: chesspairing.PairingDutch,
    },
    ScoringConfig: chesspairing.ScoringConfig{
        System: chesspairing.ScoringStandard,
    },
}

pairer := dutch.NewFromMap(nil) // default options
result, err := pairer.Pair(context.Background(), state)
```

Elke engine heeft ook een `NewFromMap(map[string]any)`-constructor voor instantiatie vanuit configuratiemaps, wat het eenvoudig maakt om engines aan te sluiten vanuit JSON-configuratie of databaserecords.

Zie de [Go-bibliotheek-snelstart](/docs/getting-started/go-quickstart/) voor een volledig stappenplan, of de [API-referentie](/docs/api/) voor type- en pakketdocumentatie.

## Volgende stappen

- [CLI-snelstart](/docs/getting-started/cli-quickstart/) -- installeer de tool en deel een toernooi in in minder dan vijf minuten
- [Go-bibliotheek-snelstart](/docs/getting-started/go-quickstart/) -- voeg de module toe aan een Go-project en genereer indelingen programmatisch
- [Voor arbiters](/docs/getting-started/for-arbiters/) -- begrijp hoe chesspairing zich verhoudt tot FIDE-reglementen
- [Voor onderzoekers](/docs/getting-started/for-researchers/) -- verken de algoritme-implementaties en matchingstrategieën
- [Concepten](/docs/concepts/) -- achtergrond over Zwitserse systemen, scoring, tiebreaking, kleuren, byes en floaters
