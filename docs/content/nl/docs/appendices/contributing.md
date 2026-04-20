---
title: "Bijdragen"
linkTitle: "Bijdragen"
weight: 3
description: "Hoe je kunt bijdragen aan het chesspairing-project."
---

## Vereisten

- Go 1.24 of nieuwer
- Er worden geen externe afhankelijkheden gebruikt -- de module gebruikt uitsluitend de Go-standaardbibliotheek, en dat is bewust zo

## Ontwikkelworkflow

```bash
# Clone
git clone https://github.com/gnutterts/chesspairing.git
cd chesspairing

# Run tests
go test -race -count=1 ./...

# Lint
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 run ./...

# Vet
go vet ./...
```

Alle drie de controles moeten slagen voordat je wijzigingen indient.

## Codestijl

- Standaard Go-opmaak (`gofmt`). Gebruik geen alternatieve formatters.
- Geef errors terug in plaats van te panicken.
- Alle engine-methoden accepteren `context.Context` als eerste parameter.
- Engine-configuratie gebruikt het Options-patroon met pointer-velden: nil-velden betekenen "gebruik standaard". Elke Options-struct biedt `WithDefaults()` en `ParseOptions(map[string]any)`.
- Compile-time interface-controles in elk engine-pakket (bijv. `var _ chesspairing.Pairer = (*Pairer)(nil)`).
- Geen externe afhankelijkheden. Als je functionaliteit nodig hebt die niet in de standaardbibliotheek zit, implementeer het dan binnen de module.

## Commitberichten

Gebruik natuurlijke, beschrijvende commitberichten. Geen conventional commit-prefixen (geen `feat:`, `fix:`, etc.). Voorbeelden:

- "Add Dutch pairer with global Blossom matching"
- "Wire up Keizer scoring with iterative convergence"
- "Fix color allocation edge case in odd player counts"

## Testen

- Alle wijzigingen moeten slagen voor de bestaande testsuite (~1325 tests verdeeld over 19 pakketten).
- Nieuwe functionaliteit hoort tests te bevatten.
- White-box tests (zelfde pakket) zijn de norm. Het root-pakket gebruikt black-box tests (`chesspairing_test`).
- De Dutch pairer gebruikt golden file tests met zelfgegenereerde, JaVaFo 2.2- en bbpPairings-referentieindelingen.
- Fuzz testing is beschikbaar voor de TRF-parser (`trf/fuzz_test.go`).

Draai de volledige testsuite met de race detector ingeschakeld:

```bash
go test -race -count=1 ./...
```

## Licentie voor bijdragen

Door een pull request of patch in te dienen, ga je ermee akkoord dat je bijdrage
gelicenseerd is onder de [Apache License 2.0](https://github.com/gnutterts/chesspairing/blob/main/LICENSE),
dezelfde licentie als de rest van dit project. Zie Sectie 5 van de licentie
voor details.

## Interessegebieden

De volgende gebieden kunnen baat hebben bij bijdragen:

- **Extra indelingssystemen** -- Nieuwe FIDE- of niet-FIDE-indelingssystemen.
- **Prestatieoptimalisatie** -- Profilering en verbetering van de Blossom-matching of iteratieve scoringsalgoritmen.
- **Documentatie** -- Correcties, verduidelijkingen en extra voorbeelden.
- **Bugrapporten** -- Meldingen uit daadwerkelijk toernooigebruik zijn bijzonder waardevol voor het valideren van randgevallen.

## Problemen melden

Maak issues aan op [https://github.com/gnutterts/chesspairing/issues](https://github.com/gnutterts/chesspairing/issues). Vermeld bij bugs de invoerdata (TRF-bestand of `TournamentState`-constructie) en de verwachte versus daadwerkelijke uitvoer.
