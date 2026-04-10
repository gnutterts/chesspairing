---
title: "Scoresystemen"
linkTitle: "Scores"
weight: 40
description: "Drie score-engines — Standaard (1-½-0), Keizer (iteratieve convergentie) en Voetbal (3-1-0)."
---

Chesspairing bevat drie score-engines. Elke engine implementeert de `Scorer`-interface en zet partijresultaten om in spelersscores. Scoren staat los van indelen -- elke score-engine is te combineren met elk indelingssysteem.

## In een oogopslag

| Systeem                | Puntenschema                        | Iteratie                  | Primair gebruik                                      |
| ---------------------- | ----------------------------------- | ------------------------- | ---------------------------------------------------- |
| [Standaard](standard/) | 1 - 0.5 - 0 (configureerbaar)       | Geen (enkele doorgang)    | FIDE-gewaarmerkte evenementen, Zwitsers, round-robin |
| [Keizer](keizer/)      | Dynamisch (gewogen op tegenstander) | Iteratief (tot 20 rondes) | Clubtoernooien, competitie                           |
| [Voetbal](football/)   | 3 - 1 - 0 (configureerbaar)         | Geen (enkele doorgang)    | Evenementen die winstprikkels willen                 |

## Waarin ze verschillen

**Standaardscoring** kent vaste puntwaarden toe: 1 voor winst, 0.5 voor remise, 0 voor verlies. Elke puntwaarde is configureerbaar, inclusief aparte waarden voor byes, forfaitwinst, forfaitverlies en afwezigheid. Omdat de waarde van een resultaat nooit afhangt van de tegenstander, levert één doorgang door de resultaten de eindstand op. Dit is het scoresysteem dat in vrijwel alle gewaarmerkte schaakevenementen wordt gebruikt.

**Keizerscoring** maakt de waarde van een resultaat afhankelijk van de huidige rangschikking van de tegenstander. De nummer 1 verslaan levert meer op dan de laatstgeplaatste verslaan. Omdat rangschikkingen afhangen van scores en scores van rangschikkingen, gebruikt Keizer iteratieve convergentie: bereken scores, herrangschik, herbereken, en herhaal tot de rangschikking stabiliseert (of tot maximaal 20 iteraties). De engine werkt intern met x2-gehele-getalrekenkunde en bevat 2-cyclusoscillatiedetectie met middeling om terminatie te garanderen. Keizerscoring is bedoeld voor clubevenementen waar het belonen van spelen tegen sterke tegenstanders competitieve prikkels creëert.

**Voetbalscoring** is een dunne wrapper rond standaardscoring met andere standaardwaarden: 3 voor winst, 1 voor remise, 0 voor verlies. De hogere winst/remiseverhouding ontmoedigt korte remises en beloont beslissende partijen. Alle puntwaarden blijven configureerbaar. Onder de motorkap delegeert voetbalscoring volledig aan de standaard-engine met aangepaste standaardparameters.

## Interactie tussen scoren en indelen

Indelen en scoren zijn bewust ontkoppeld. Een toernooi kan Zwitserse indeling met Keizerscoring gebruiken, round-robin met voetbalscoring, of elke andere combinatie. De Zwitserse indelingsmodules gebruiken intern standaard 1-0.5-0-scoring voor het vormen van scoregroepen, ongeacht het publieke scoresysteem van het toernooi -- dit is opzettelijk, aangezien FIDE-reglementen voor Zwitsers scoregroepen definiëren in termen van standaardpunten.

De enige uitzondering is de Keizer-indeling, die Keizerscores gebruikt om de indelingsvolgorde te bepalen. De Keizer-indeling gebruiken met een niet-Keizer-scoresysteem zou willekeurige indelingen opleveren, dus die combinatie is niet zinvol.

## Forfait- en bye-afhandeling

Alle drie de score-engines verwerken dezelfde set speciale resultaten:

| Resultaattype  | Standaard (default) | Keizer                        | Voetbal (default) |
| -------------- | ------------------- | ----------------------------- | ----------------- |
| Partijwinst    | 1.0                 | Dynamisch (op basis van rang) | 3.0               |
| Partijremise   | 0.5                 | Dynamisch                     | 1.0               |
| Partijverlies  | 0.0                 | Dynamisch                     | 0.0               |
| Forfaitwinst   | 1.0                 | Vast deel van zelfoverwinning | 3.0               |
| Forfaitverlies | 0.0                 | 0.0                           | 0.0               |
| PAB (Bye)      | 1.0                 | Vast deel van zelfoverwinning | 3.0               |
| Afwezig        | 0.0                 | 0.0                           | 0.0               |
| Dubbel forfait | 0.0 / 0.0           | 0.0 / 0.0                     | 0.0 / 0.0         |

Bij Keizerscoring worden forfaitwinsten en byes berekend als een configureerbaar deel van de "zelfoverwinningswaarde" -- de punten die een speler zou krijgen voor het verslaan van zichzelf op de huidige rangschikking. Dit houdt niet-partijresultaten evenredig aan de sterkte van een speler.

## Interface

Alle drie de engines implementeren dezelfde interface:

```go
type Scorer interface {
    Score(ctx context.Context, state TournamentState) ([]PlayerScore, error)
}
```

De geretourneerde `PlayerScore`-slice bevat één vermelding per speler met het totale puntenaantal. De scorer wijzigt de invoerstate nooit. Elke engine biedt ook `NewFromMap(map[string]any)` voor generieke instantiatie vanuit configuratiemaps.
