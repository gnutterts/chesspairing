---
title: "Tiebreaking"
linkTitle: "Tiebreaking"
weight: 4
description: "Wanneer spelers dezelfde score delen, bepalen tiebreakers wie hoger wordt gerangschikt."
---

## Waarom tiebreakers bestaan

In een Zwitsers toernooi met 40 spelers en 7 ronden is het gebruikelijk
dat meerdere spelers op dezelfde score eindigen. Twee spelers met 5,5/7
moeten worden gerangschikt -- wie krijgt de trofee? Tiebreakers
beantwoorden die vraag door voor elke speler een secundaire waarde te
berekenen die verder identieke scores onderscheidt.

## Hoe tiebreakers worden toegepast

Tiebreakers worden geconfigureerd als een geordende lijst. De
scoringsengine berekent elke tiebreaker voor elke speler en past ze
vervolgens in volgorde toe:

1. Spelers worden eerst gerangschikt op score.
2. Onder spelers met gelijke score wordt de eerste tiebreaker vergeleken.
3. Als de eerste tiebreaker ook gelijk is, wordt de tweede tiebreaker
   vergeleken.
4. Dit gaat door totdat een verschil is gevonden of alle tiebreakers
   zijn uitgeput.

De volgorde doet ertoe. Buchholz als eerste plaatsen betekent dat
"sterkte van tegenstanders" boven alles wordt gewaardeerd; Direct
Encounter als eerste plaatsen betekent dat onderlinge resultaten
voorrang krijgen.

## Categorieën tiebreakers

chesspairing biedt 25 tiebreakers, die in zes categorieën vallen:

### Tegenstander-gebaseerd

Deze meten hoe sterk je tegenstanders waren. De logica: tegenstanders
verslaan die zelf goed scoorden is indrukwekkender dan tegenstanders
verslaan die slecht scoorden.

- **Buchholz** -- som van de eindscores van alle tegenstanders. De meest
  gebruikte Zwitserse tiebreaker.
- **Buchholz Cut-1** -- Buchholz min de laagste tegenstander-score.
  Vermindert de straf voor één zwakke tegenstander.
- **Buchholz Cut-2** -- Buchholz min de twee laagste
  tegenstander-scores.
- **Buchholz Mediaan** -- Buchholz min de hoogste en laagste
  tegenstander-score. Verwijdert uitschieters in beide richtingen.
- **Buchholz Mediaan-2** -- Buchholz min de twee hoogste en twee
  laagste.
- **Fore Buchholz** -- Buchholz waarbij nog te spelen partijen als
  remise worden behandeld. Handig voor de tussenstand.
- **Average Opponent Buchholz** -- Buchholz gedeeld door het aantal
  gespeelde partijen.

### Prestatie-gebaseerd

Deze schatten hoe goed je presteerde ten opzichte van je rating.

- **Performance Rating (TPR)** -- de rating waarbij je resultaat
  verwacht zou zijn, berekend uit de FIDE B.02-conversietabel.
- **Performance Points (PTP)** -- verwachte score op basis van
  ratingverschillen, met behulp van FIDE-verwachtingsscoretabellen.
- **Average Rating of Opponents (ARO)** -- het gemiddelde van de
  ratings van je tegenstanders.
- **Average Opponent TPR (APRO)** -- het gemiddelde van de performance
  ratings van je tegenstanders.
- **Average Opponent PTP (APPO)** -- het gemiddelde van de performance
  points van je tegenstanders.

### Resultaat-gebaseerd

Deze richten zich op de kwaliteit van je individuele resultaten.

- **Games Won** -- aantal overwinningen aan het bord (exclusief
  forfaits).
- **Rounds Won** -- aantal overwinningen inclusief forfaitwinstpartijen
  en PAB-byes.
- **Progressive Score** -- cumulatieve (lopende) score na elke ronde.
  Beloont vroege overwinningen meer dan late.

### Onderling resultaat

Deze kijken naar de resultaten tussen de specifieke spelers die gelijk
staan.

- **Direct Encounter** -- de score tussen de gelijkstaande spelers in
  hun onderlinge partijen. Alleen zinvol wanneer de gelijkstaande
  spelers daadwerkelijk hebben gespeeld.
- **Sonneborn-Berger** -- vermenigvuldig voor elke tegenstander je
  resultaat tegen hem met zijn eindscore, en tel op. Beloont
  overwinningen op hoog scorende tegenstanders en bestraft remises tegen
  laag scorende tegenstanders.
- **Koya-systeem** -- je score tegen tegenstanders die in de bovenste
  helft van de stand eindigden. Gebruikelijk bij round-robin-evenementen.

### Activiteit-gebaseerd

Deze weerspiegelen het deelnamenniveau.

- **Games Played** -- totaal aantal gespeelde partijen (exclusief
  forfaits).
- **Games with Black** -- aantal partijen gespeeld als zwart (exclusief
  forfaits). Geeft een tiebreak-voordeel aan spelers die een moeilijker
  kleurschema hadden.

### Ordening

Deze bieden een deterministische laatste tiebreak wanneer alle andere
tiebreakers gelijk zijn.

- **Pairing Number** -- het rangnummer (TPN) van de speler. Lager
  nummer = hogere rang.
- **Player Rating** -- de rating van de speler. Hoger = beter.

## Standaard tiebreakers per systeem

Elk indelingssysteem heeft een aanbevolen standaard tiebreakervolgorde.
Deze standaarden worden geretourneerd door `DefaultTiebreakers()` en
gebruikt wanneer geen expliciete tiebreakerlijst is geconfigureerd:

**Zwitserse systemen** (Dutch, Burstein, Dubov, Lim, Double-Swiss, Team):

1. Buchholz Cut-1
2. Buchholz
3. Sonneborn-Berger
4. Direct Encounter

**Round-robin:**

1. Sonneborn-Berger
2. Direct Encounter
3. Games Won
4. Koya

**Keizer:**

1. Games Played
2. Direct Encounter
3. Games Won

Deze standaarden volgen de FIDE-aanbevelingen. Je kunt ze overschrijven
met elke combinatie van de 25 beschikbare tiebreakers.

## Het tiebreakerregister

Alle tiebreakers registreren zichzelf bij het opstarten via Go's
`init()`-mechanisme. Je zoekt ze op via een string-ID:

```go
tb, err := tiebreaker.Get("buchholz-cut1")
values, err := tb.Compute(ctx, state, scores)
```

De 25 geregistreerde ID's zijn: `buchholz`, `buchholz-cut1`, `buchholz-cut2`,
`buchholz-median`, `buchholz-median2`, `sonneborn-berger`,
`direct-encounter`, `wins`, `win`, `black-games`, `black-wins`,
`rounds-played`, `standard-points`, `pairing-number`, `koya`,
`progressive`, `aro`, `fore-buchholz`, `avg-opponent-buchholz`,
`performance-rating`, `performance-points`, `avg-opponent-tpr`,
`avg-opponent-ptp`, `player-rating`, `games-played`.

## Verder lezen

- [Buchholz-tiebreakers](/docs/tiebreakers/buchholz/)
- [Prestatie-gebaseerde tiebreakers](/docs/tiebreakers/performance/)
- [Resultaat-gebaseerde tiebreakers](/docs/tiebreakers/results/)
- [Onderlinge tiebreakers](/docs/tiebreakers/head-to-head/)
- [Kleur-, activiteits- en ordeningstiebreakers](/docs/tiebreakers/color-activity/)
