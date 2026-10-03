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

De FIDE-reglementen bevelen specifieke tiebreaker-reeksen aan afhankelijk van het toernooiformat. De functie `DefaultTiebreakers()` in het root-pakket geeft de door FIDE aanbevolen reeks terug voor elk indelingssysteem. Gebruikelijke keuzes:

- **Zwitserse toernooien**: Buchholz Cut 1, Buchholz, Sonneborn-Berger, Progressief
- **Round-robin**: Sonneborn-Berger, Onderling resultaat, Winstpartijen, Partijen met zwart
- **Keizer**: De Keizerscore zelf is de primaire rangschikking; extra tiebreakers zijn zelden nodig

Bij evenementen met veel gelijk eindigende spelers zijn Buchholz-varianten het meest onderscheidend, omdat ze het volledige resultatennetwerk van het toernooi meenemen. Prestatietiebreakers (TPR, PTP) zijn nuttig in grote open toernooien waar ratinggebaseerde sterktemeting zinvol is. Onderlinge tiebreakers zoals Direct Encounter zijn doorslaggevend wanneer een kleine groep spelers gelijk staat.

## Forfait-uitsluiting

Alle tiebreakers die tegenstanders analyseren gebruiken de gedeelde functie `buildOpponentData`, die forfaitpartijen uitsluit van de tegenstanderlijst. Dit betekent:

- Een forfaitwinst voegt de afwezige tegenstander niet toe aan je Buchholz-berekening.
- Een dubbel forfait wordt volledig uitgesloten van de tiebreakberekeningen van beide spelers.
- Alleen daadwerkelijk gespeelde partijen (inclusief remises) tellen mee voor tiebreakers die op tegenstanders gebaseerd zijn.

Dit komt overeen met de FIDE-tiebreakreglementen, die forfaits voor tiebreakdoeleinden als niet-partijen beschouwen.

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
