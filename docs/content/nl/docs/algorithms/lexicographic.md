---
title: "Lexicografische Indeling"
linkTitle: "Lexicografisch"
weight: 13
description: "Bracket-indeling in identifier-volgorde volgens artikel 3.6 voor Dubbel-Zwitsers en Team-Zwitsers."
---

## Overzicht

Dubbel-Zwitsers (FIDE C.04.5) en Team-Zwitsers (FIDE C.04.6) delen de
bracketindeler in `pairing/lexswiss/bracket.go`. Voor brackets met een even
aantal deelnemers kiest die de eerste geldige indeling in de identifier-volgorde
van artikelen 3.6.1--3.6.3. Brackets met een oneven aantal deelnemers behouden
de bestaande behandeling in reeksvolgorde.

## Identifier-volgorde

In elk paar is de deelnemer met het kleinere toernooirangnummer (TPN) het
toplid en de andere het onderlid. Een pairingidentifier bestaat uit de oplopende
reeks TPN's van de topleden, gevolgd door de bijbehorende TPN's van de
onderleden. Indelingen worden lexicografisch op die identifier vergeleken.

De implementatie genereert die volgorde lui:

1. Kies verzamelingen topleden in oplopende lexicografische volgorde.
2. Wijs voor elke verzameling onderleden toe in oplopende TPN-volgorde.
3. Wijs een kandidaatpaar af als het C1, een beperking voor verboden paren of
   een systeemspecifiek criterium schendt. Onuitvoerbare voorvoegsels van
   topleden worden met een matchingcontrole gesnoeid.

Voor zes deelnemers zonder beperkingen is de eerste identifier `1 2 3 4 5 6`;
die staat voor `1-4, 2-5, 3-6`.

## Criteriafunctie

```go
type CriteriaFunc func(a, b *ParticipantState) bool
```

De functie wordt voor elk kandidaatpaar aangeroepen na de controles op C1 en
verboden paren. Zij geeft aan of dat paar aan het systeemspecifieke criterium
voldoet. Dubbel-Zwitsers gebruikt haar voor het kleurvoorkeurscriterium;
Team-Zwitsers gebruikt haar voor kleur- en floatercriteria.

## Voorbeeld

Neem TPN's 3, 7, 12, 15, 22 en 28. Deelnemer 3 heeft al tegen 7 gespeeld en
12 heeft al tegen 15 gespeeld. De eerste mogelijke reeks topleden is `3 7 12`.
De eerste geldige toewijzing van onderleden is `15 22 28`, dus de indeler geeft
terug:

```
(3, 15), (7, 22), (12, 28)
```

De identifier is `3 7 12 15 22 28`.

## Aanvulling voor voltooiing

Als Dubbel-Zwitsers of Team-Zwitsers de afzonderlijke brackets niet volledig
kan indelen, probeert de aanroeper nu opnieuw één bracket met alle resterende
deelnemers. Dit is een aanvulling voor voltooiing buiten de bracketindeler en
moet worden herzien; zij vervangt de volgorde van artikel 3.6 binnen een
bracket niet.

## Complexiteit

Het aantal mogelijke identifiers is in het slechtste geval exponentieel. De
indeler stopt bij de eerste geldige volledige identifier en snoeit voorvoegsels
van topleden waarvoor de gekozen topleden geen verschillende geldige
onderleden kunnen krijgen. Daardoor blijven gewone brackets snel, met behoud
van de voorgeschreven volgorde.

## Gerelateerde pagina's

- [Dubbel-Zwitserse Indeling](/docs/pairing-systems/double-swiss/)
- [Team-Zwitserse Indeling](/docs/pairing-systems/team/)
- [Dutch-criteria](../dutch-criteria/)
