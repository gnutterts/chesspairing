---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Versiegeschiedenis en belangrijke wijzigingen."
---

De canonieke changelog staat in [`CHANGELOG.md`](https://github.com/gnutterts/chesspairing/blob/main/CHANGELOG.md) in de hoofdmap van de repository. Deze volgt losjes het [Keep a Changelog](https://keepachangelog.com/)-formaat en Semantic Versioning.

## Laatste release: v0.2.0 (20-04-2026)

Deze release herziet het vocabulaire voor niet-gespeelde rondes in alle subsystemen, samen met een batch publieke-API-toevoegingen die al een tijdje op de plank lagen. De zes `ByeType`-waarden lopen nu consistent door scoring, tiebreaking, indeling, TRF-I/O en de standentabel. Speler-terugtrekkingen gaan van een per-ronde `Active`-boolean naar een eenmalige `WithdrawnAfterRound`-pointer. Vooraf toegewezen byes voor de komende ronde krijgen een eersteklas veld op `TournamentState`. `ResultContext` ontsluit het bye-type rechtstreeks in plaats van het samen te vatten tot een enkele "is een bye"-vlag.

De nieuwe `Parse*`-helpers, `PlayedPairs`, het `factory`-subpakket en het `standings`-subpakket completeren de publieke API voor afnemende tools die deze stukken voorheen zelf opnieuw implementeerden. Pre-1.0 breekwijzigingen staan in de hoofd-`CHANGELOG.md` onder Removed; er zijn geen shims.

## Versiestring bij het bouwen

Tijdens de ontwikkeling is de versiestring standaard `"dev"`. Releaseversies worden bij het bouwen ingesteld via `-ldflags`:

```bash
go build -ldflags "-X main.version=v0.2.0" ./cmd/chesspairing
```
