---
title: "Changelog"
linkTitle: "Changelog"
weight: 2
description: "Versiegeschiedenis en belangrijke wijzigingen."
---

## Huidige status

De chesspairing-module is in actieve ontwikkeling. Er zijn nog geen stabiele releases uitgebracht.

Tijdens de ontwikkeling is de versiestring standaard `"dev"`. Releaseversies worden bij het bouwen ingesteld via `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.0.0" ./cmd/chesspairing
```

Zodra er stabiele releases verschijnen, wordt de volledige versiegeschiedenis op deze pagina bijgehouden.

## Formaat

Toekomstige changelog-items volgen dit formaat:

```text
## v1.0.0 (YYYY-MM-DD)

### Added
- ...

### Changed
- ...

### Fixed
- ...
```
