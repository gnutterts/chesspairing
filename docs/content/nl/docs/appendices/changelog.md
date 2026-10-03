---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Versiegeschiedenis en belangrijke wijzigingen."
---

De canonieke changelog staat in [`CHANGELOG.md`](https://github.com/gnutterts/chesspairing/blob/main/CHANGELOG.md) in de hoofdmap van de repository. Deze volgt losjes het [Keep a Changelog](https://keepachangelog.com/)-formaat en Semantic Versioning.

De onderstaande vermeldingen worden uit dat bestand opgebouwd, zodat deze pagina altijd de actuele geschiedenis toont, met de nieuwste release bovenaan. De changelog is in het Engels geschreven.

{{< changelog >}}

## Versiestring bij het bouwen

Tijdens de ontwikkeling is de versiestring standaard `"dev"`. Releaseversies worden bij het bouwen ingesteld via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.2.3" ./cmd/chesspairing
```
