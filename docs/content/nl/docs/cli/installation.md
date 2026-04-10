---
title: "Installatie"
linkTitle: "Installatie"
weight: 1
description: "Hoe je de chesspairing CLI-tool installeert."
---

## Vanuit broncode (go install)

Met Go 1.26 of nieuwer geinstalleerd:

```bash
go install github.com/gnutterts/chesspairing/cmd/chesspairing@latest
```

Dit bouwt het programma en plaatst het in `$GOBIN` (meestal `$HOME/go/bin`). Zorg ervoor dat deze map in je `PATH` staat.

## Handmatig bouwen vanuit broncode

Kloon de repository en bouw handmatig:

```bash
git clone https://github.com/gnutterts/chesspairing.git
cd chesspairing
go build -o chesspairing ./cmd/chesspairing
```

Om een versiestring mee te compileren:

```bash
go build -ldflags "-X main.version=1.0.0" -o chesspairing ./cmd/chesspairing
```

Zonder `-ldflags` is de versie standaard `dev`.

## Installatie controleren

```bash
chesspairing version
```

Verwachte uitvoer:

```text
chesspairing dev

Pairing systems:  dutch, burstein, dubov, lim, doubleswiss, team, keizer, roundrobin
Scoring systems:  standard, keizer, football
Tiebreakers:      25 available
```

## Vereisten

- **Go 1.26.1 of nieuwer** om vanuit broncode te bouwen.
- **Geen externe afhankelijkheden.** De volledige module gebruikt uitsluitend de Go-standaardbibliotheek.
- Het programma is statisch gelinkt en heeft geen runtime-afhankelijkheden.

## Als bibliotheek

Als je chesspairing als Go-bibliotheek wilt gebruiken in plaats van als CLI-tool, voeg het dan toe aan je module:

```bash
go get github.com/gnutterts/chesspairing@latest
```

Zie de [Go-snelstart](/docs/getting-started/go-quickstart/) voor bibliotheekgebruik.
