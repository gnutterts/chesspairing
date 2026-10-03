---
title: "Testen"
weight: 5
---

Chesspairing gebruikt unit tests, FIDE-voorbeelden, gouden CLI-tests, fuzztests
voor TRF-verwerking en een opgeslagen Nederlandse differentiële corpus.

## Nederlandse differentiële harness

`internal/harness/testdata/corpus/` bevat 150 volledige toernooien (898
ronden), met forfaits, aangevraagde byes en terugtrekkingen. `oracle.json`
bevat de Nederlandse indelingen van bbpPairings 6.0.0; `baseline.json` legt
het geaccepteerde resultaat van chesspairing vast. De normale test vergelijkt
paren en kleuren met deze opgeslagen antwoorden en heeft daarom geen extern
programma nodig; de uitkomst is op elk platform reproduceerbaar:

```bash
go test ./internal/harness
```

De oracle is bbpPairings 6.0.0 omdat die de Nederlandse regels volgt die op 1
februari 2026 van kracht zijn, net als deze implementatie. JaVaFo 2.2 volgt de
regels van 2017 en is daarom geen oracle: de huidige regels verschillen onder
meer bij de downfloat na een forfaitverlies en bij het minimaliseren van de
score van de speler die de indelings-bye ontvangt.

Met bbpPairings 6.0.0 op `PATH` verifiëren deze optionele controles de
opgeslagen oracle of onderzoeken zij nieuw gegenereerde toernooien:

```bash
go test -tags harness ./internal/harness -run TestHarnessOracle
HARNESS_N=100 HARNESS_SEED=5000 go test -tags harness ./internal/harness -run TestHarnessExplore
```

Het corpus wordt opgeslagen in plaats van tijdens een test opnieuw gemaakt,
omdat de willekeurige generator verschillende toernooien oplevert op platforms
met verschillende C++-standaardbibliotheken. Zie
`internal/harness/testdata/README.md` voor de stappen om het te vernieuwen.
