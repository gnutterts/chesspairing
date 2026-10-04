---
title: "Tiebreakers"
linkTitle: "Tiebreakers"
weight: 50
description: "25 tiebreaker-implementaties ingedeeld per categorie — van Buchholz-varianten tot prestatieratings."
---

Chesspairing biedt 25 tiebreakers in zeven categorieën. Elke tiebreaker implementeert de `TieBreaker`-interface en berekent per speler een numerieke waarde op basis van de toernooistatus en stand. Tiebreakers registreren zichzelf via een centraal register en kunnen bij naam worden geselecteerd tijdens runtime.

## categorieën

| Categorie                                   | Tiebreakers                                                                | Wat ze meten                                                 |
| ------------------------------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------------ |
| [Buchholz](buchholz/)                       | buchholz, buchholz-cut1, buchholz-cut2, buchholz-median, buchholz-median2  | Som van scores van tegenstanders (met cut/mediaan-varianten) |
| [Prestatie](performance/)                   | performance-rating, performance-points, avg-opponent-tpr, avg-opponent-ptp | Sterkte van het spel ten opzichte van de tegenstand          |
| [Resultaten](results/)                      | wins, win, standard-points, progressive, rounds-played, games-played       | Directe maatstaven van partijuitslagen                       |
| [Onderling resultaat](head-to-head/)        | direct-encounter, sonneborn-berger, koya                                   | Resultaten tussen gelijke of bovenste-helft-tegenstanders    |
| [Kleur & Activiteit](color-activity/)       | black-games, black-wins                                                    | Kleurverdelingsstatistieken                                  |
| [Volgorde](ordering/)                       | pairing-number, player-rating                                              | Statische spelereigenschappen voor definitieve volgorde      |
| [Tegenstander-Buchholz](opponent-buchholz/) | fore-buchholz, avg-opponent-buchholz                                       | Buchholz-afgeleide maatstaven van tegenstanders              |

## Tiebreakers kiezen

De hoofdorganisator kiest de tiebreakvolgorde in het toernooireglement (FIDE C.07:2026 artikel 4.1). `DefaultTiebreakers()` levert bibliotheekstandaarden, die in de configuratie kunnen worden vervangen:

- **Zwitserse toernooien**: Buchholz Cut-1, Buchholz, Sonneborn-Berger, Onderling resultaat
- **Team-Zwitsers**: Buchholz MP Cut-1, Buchholz MP, Uitgebreide Sonneborn-Berger (MP/MP), Matchpunten of partijpunten
- **Round-robin**: Sonneborn-Berger, Onderling resultaat, Gewonnen partijen, Koya
- **Keizer**: Gespeelde partijen, Onderling resultaat, Gewonnen partijen

Bij evenementen met veel gelijk eindigende spelers zijn Buchholz-varianten het meest onderscheidend, omdat ze het volledige resultatennetwerk van het toernooi meenemen. Prestatietiebreakers (TPR, PTP) zijn nuttig in grote open toernooien waar ratinggebaseerde sterktemeting zinvol is. Onderlinge tiebreakers zoals Direct Encounter zijn doorslaggevend wanneer een kleine groep spelers gelijk staat.

## Niet-gespeelde ronden

Tiebreakers op basis van tegenstanders bouwen voor elke speler en ronde één record op. In Zwitserse toernooien deelt FIDE C.07:2026 artikel 16 niet-gespeelde ronden in, past het waar nodig tegenstanderscores aan en gebruikt de begrensde dummy-tegenstanders voor de eigen berekening. In round-robins en andere vooraf vastgelegde indelingen behandelt artikel 15.2 forfaits juist als gewone ontmoetingen, behalve bij ratingtiebreakers en Type-B-forfaitverliezen. Hangende partijen tellen niet als partijen aan het bord.

## Register

Tiebreakers registreren zichzelf via `init()`-functies:

```go
func init() {
    Register("buchholz", func() chesspairing.TieBreaker {
        return &Buchholz{variant: buchholzFull}
    })
}
```

Haal tijdens runtime een tiebreaker op via zijn geregistreerde naam:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
if err != nil {
    // unknown tiebreaker name
}
values, err := tb.Compute(ctx, state, scores)
```

De functie `tiebreaker.All()` geeft de geregistreerde individuele namen terug,
`tiebreaker.TeamAll()` de teamnamen, en het CLI-subcommando `tiebreakers`
toont beide met beschrijvingen.

## Teamcompetities

Bij teamscoring levert `mpvgp` de secundaire MP- of GP-score. De uitgebreide
Sonneborn-Berger-varianten zijn `emmsb`, `emgsb`, `egmsb` en `eggsb`; gebruik
`emmsb-cut1` voor de Cut-1-variant. `buchholz-mp` en `buchholz-mp-cut1`
gebruiken de MP van tegenstanders. Matches met bordresultaten ondersteunen
ook `board-count`, `top-board-results` en `bottom-board-elimination`. Zie
FIDE C.07:2026 artikelen 12 en 13.


## Interface

Elke tiebreaker implementeert:

```go
type TieBreaker interface {
    Compute(ctx context.Context, state TournamentState, scores []PlayerScore) ([]TieBreakValue, error)
}
```

De parameter `scores` bevat de huidige stand (van een willekeurige scoring-engine). De teruggegeven `TieBreakValue`-slice bevat per speler een item met de berekende tiebreakwaarde. Tiebreakers wijzigen nooit de invoerstatus of scores.

## Geregistreerde ID's

`aro`, `aro-cut1`, `avg-opponent-buchholz`, `avg-opponent-fore-buchholz`, `avg-opponent-ptp`, `avg-opponent-tpr`, `black-games`, `black-wins`, `board-count`, `bottom-board-elimination`, `buchholz`, `buchholz-cut1`, `buchholz-cut2`, `buchholz-median`, `buchholz-median2`, `buchholz-mp`, `buchholz-mp-cut1`, `direct-encounter`, `eggsb`, `egmsb`, `emgsb`, `emmsb`, `emmsb-cut1`, `fore-buchholz`, `games-played`, `ge`, `koya`, `mpvgp`, `pairing-number`, `performance-points`, `performance-rating`, `player-rating`, `progressive`, `progressive-cut1`, `rounds-played`, `sonneborn-berger`, `sonneborn-berger-cut1`, `standard-points`, `top-board-results`, `win`, `wins`
