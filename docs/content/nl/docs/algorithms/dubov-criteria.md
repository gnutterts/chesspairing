---
title: "Dubov-criteria"
linkTitle: "Dubov-criteria"
weight: 11
description: "FIDE C.04.4.1-criteria C1–C10 voor Dubov-paringen: definities, meting, vergelijkingsvolgorde en voorbeelden."
---

Het Dubov-systeem gebruikt de criteria van C.04.4.1 (van kracht vanaf 1 februari
2026), artikel 2. Er zijn er tien: drie absolute criteria (C1–C3), één
voltooiingscriterium (C4) en zes kwaliteitscriteria (C5–C10). Deze pagina geeft
voor elk de tekst van het artikel, de functie die hem in `pairing/dubov` meet, en
uitgewerkte voorbeelden. De procedure die ze gebruikt, staat op de pagina
[Dubov-systeem](../../pairing-systems/dubov/).

## Waar de criteria worden gebruikt

| Stap | Criteria | Artikel |
| --- | --- | --- |
| Het kiezen van de paringsbye (PAB) | C2, en C4 via de volledige bracket-procedure | 3.1 |
| Het kiezen van het aantal opfloaters | C1, C3, C4, C5 | 3.2.1 |
| Het kiezen van de set opfloaters | eerst scoremultiset (4.2.3), dan C6–C10 | 3.2.2 |
| Het kiezen van een herindeling | eerste legale herindeling in de volgorde van 3.2.4 en 4.3 | 3.2.4 |
| Het kiezen van de transpositie van G2 | C1, C3 alleen; de eerste legale wint | 3.2.6 |

Verboden paren (optie `forbiddenPairs`) maken geen deel uit van C.04.4.1. De
code behandelt ze als nog een absoluut criterium: `SatisfiesAbsoluteCriteria`
toetst C1, C3 en de verboden paren samen.

## Absolute criteria

### C1 -- geen rematch (2.1.1)

*Tekst:* twee deelnemers spelen niet meer dan één keer tegen elkaar.

*Gemeten door* `C1NoRematches` in `criteria.go`, die waar is wanneer
`swisslib.HasPlayed` geen eerdere partij tussen de twee vindt. Alleen aan het
bord gespeelde partijen tellen: een forfaitpartij staat niet in de
tegenstanderslijst, dus die twee spelers mogen elkaar opnieuw ontmoeten.

*Voorbeeld:* p2 en p5 troffen elkaar in ronde 1. In ronde 2 wordt de
transpositie die p2 tegen p5 zet verworpen en de volgende transpositie in de
volgorde van artikel 4.4 wordt geprobeerd.

### C2 -- geen tweede paringsbye (2.1.2)

*Tekst:* een deelnemer die al een paringsbye ontving, of die al in één enkele
ronde zonder te spelen zoveel punten scoorde als voor een winst worden beloond,
ontvangt geen paringsbye.

*Gemeten door* `filterNoByeReceived` in `bye.go`, die elke speler weglaat van
wie de `PABIneligible`-vlag gezet is. De vlag wordt gezet door een eerdere PAB,
een eerdere bye met volle punten en een eerdere forfaitwinst. `orderedByeCandidates`
in `dubov.go` gebruikt hem wanneer `Pair` de bye kiest.

*Voorbeeld:* een speler die de PAB in ronde 3 kreeg, is in ronde 6 geen
kandidaat. Een speler met een bye met halve punten is nog wel kandidaat; een
speler die met forfait won, is dat niet.

### C3 -- geen twee absolute kleurvoorkeuren voor dezelfde kleur (2.1.3)

*Tekst:* twee spelers met dezelfde absolute kleurvoorkeur ontmoeten elkaar niet.
Artikel 1.6.1: een voorkeur is absoluut wanneer het kleurverschil groter is dan
+1 of kleiner dan −1, of wanneer de speler dezelfde kleur had in de twee laatst
gespeelde ronden.

*Gemeten door* `C3NoAbsoluteColorConflict`, die beide voorkeuren leest via de
lokale `preference`-functie (artikel 1.6.4 inbegrepen) en het paar alleen
verwerpt wanneer beide absoluut zijn en voor dezelfde kleur. Er is geen
uitzondering voor toppuntenaars. Een speler die nog niet speelde, heeft een lichte
voorkeur voor zwart en is nooit absoluut.

*Voorbeeld:* een speler van wie de partijen zwart, zwart waren, heeft een
absolute voorkeur voor wit; dat heeft ook een speler met zwart, zwart, wit,
zwart (drie zwart, één wit: verschil boven +1 voor zwart). Die twee kunnen niet
samen worden ingedeeld. Een speler met één zwarte partij heeft alleen een sterke
voorkeur voor wit en kan tegen elk van hen worden ingepaard (C7 telt dan de
voorkeur die niet vervuld wordt).

## Voltooiingscriterium

### C4 -- een legale paring moet bestaan (2.2.1)

*Tekst:* een paring die aan alle absolute criteria voldoet, moet altijd bestaan
voor alle spelers die nog niet ingedeeld zijn.

*Gemeten door* `completable` in `dubov.go` (niet in `criteria.go`). Hij bouwt een
graaf waarin twee spelers verbonden zijn wanneer `SatisfiesAbsoluteCriteria`
geldt en vraagt het Blossom-algoritme om een maximale koppeling; het antwoord is
alleen ja wanneer elke speler gekoppeld is. Hij wordt tweemaal
gebruikt:

- Bij het kiezen van een set opfloaters (3.2.1/3.2.2) moeten de spelers die in de
  lagere scoregroepen achterblijven de test nog steeds doorstaan; anders wordt
  de set overgeslagen.
- Voor de PAB (3.1.2) verwijdert de code elke kandidaat op zijn beurt en voert de
  hele bracket-procedure uit op de rest (die voor elk bracket de test hierboven
  bevat); de eerste kandidaat voor wie dat lukt, krijgt de bye.

*Voorbeeld:* een bracket van drie vaste spelers heeft één opfloater nodig uit de
spelers daaronder. Als kiezen van p4 p5 en p6 achter zou laten en die twee
elkaar al troffen, wordt p4 overgeslagen, hoe goed p4 ook in het bracket zou
passen.

## Kwaliteitscriteria

Kwaliteit wordt bewaard in `candidateQuality` en berekend door `qualityFor(pairs,
upfloaterIDs, lastRound, maxT)` voor het volledig ingedeelde bracket, waarbij
`upfloaterIDs` de opfloaters van de beoordeelde set zijn.

### C5 -- aantal opfloaters (2.3.1)

*Tekst:* minimaliseer het aantal opfloaters.

*Gemeten door* `candidateQuality.upfloaters`, de grootte van de set. Omdat
`selectBracket` begint met het kleinste aantal dat bij de pariteit van het bracket
past en alleen omhoog gaat wanneer geen toelaatbare set bestaat, hebben de sets
die vergeleken worden altijd dezelfde grootte; C5 is daarmee bij constructie
vervuld.

*Voorbeeld:* in een scoregroep van drie vaste spelers is ten minste één opfloater nodig
(drie spelers kunnen niet ingedeeld worden). De code probeert eerst elke set van
één. Pas wanneer geen enkele een legale paring geeft, probeert hij sets van drie.

### C6 -- scoreverschillen in paren met opfloaters (2.3.2)

*Tekst:* minimaliseer de scoreverschillen (in aflopende volgorde) in de paren
met opfloaters, d.w.z. maximaliseer de scores (in oplopende volgorde) van de
opfloaters.

*Gemeten door* `candidateQuality.differences`: voor elk paar dat een opfloater
bevat, het absolute scoreverschil van de twee spelers; de lijst wordt in
aflopende volgorde gesorteerd. Lijsten worden element voor element vergeleken, de
kleinste waarde wint bij het eerste verschil. Wanneer de ene lijst het begin van
de andere is, wint de kortste (minder paren met opfloaters, bijvoorbeeld twee
opfloaters tegen elkaar). Dit is de vectorvergelijking van de tekst, niet een som.

*Voorbeeld:* de vaste spelers hebben 4 punten en er zijn twee opfloaters nodig uit
spelers met 3, 2, 2 en 1 punt.

| Opfloaters | Verschillen, aflopend | Rang |
| --- | --- | --- |
| 3 en 2 | (2, 1) | beste |
| 2 en 2 | (2, 2) | |
| 3 en 1 | (3, 1) | |
| 2 en 1 | (3, 2) | slechtste |

`{2,2}` verslaat `{3,1}` hoewel beide sommen 4 zijn, omdat het grootste
verschil kleiner is.

### C7 -- spelers die hun kleurvoorkeur mislopen (2.3.3)

*Tekst:* minimaliseer het aantal spelers dat zijn kleurvoorkeur niet krijgt.

*Gemeten door* `candidateQuality.colorViolations` (`CriterionC7` is de
geëxporteerde vorm). De G1-speler van elk paar telt als wit en de G2-speler als
zwart; elke speler die de andere kleur prefereert, telt 1. De voorkeur is die van
`preference`, dus een speler zonder partij prefereert zwart. De uiteindelijke
kleuren van artikel 5.2 worden hier niet gebruikt; ze worden vastgesteld na het
indelen.

Artikel 3.2.4 gebruikt de eerste legale herindeling; herindelingen worden niet
op C7 vergeleken. C7 onderscheidt toelaatbare opfloatersets binnen een groep met
hetzelfde scoremultiset, na C6.

*Voorbeeld:* zes spelers, ronde 1 paart p1–p4, p5–p2 en p3–p6; p1, p2 en p3
winnen. In ronde 2 hebben de vaste spelers p1, p2, p3 één opfloater nodig uit p4, p5,
p6 (4.2-volgorde: p4, p5, p6 bij gelijke score en oplopende TPN). p1 en p3
prefereren zwart, p2 prefereert wit (hij speelde zwart).

| Opfloater | G1 (wit) / G2 (zwart) | Schendingen |
| --- | --- | --- |
| p4 (speelde zwart, prefereert wit) | G1 {p2, p4}, G2 {p1, p3} | 0 |
| p5 (speelde wit, prefereert zwart) | G1 {p2, p3}, G2 {p1, p5} na het verschuiven van p3 | 1 |
| p6 (speelde zwart, prefereert wit) | G1 {p2, p6}, G2 {p1, p3} | 0 |

C5 en C6 zijn gelijk voor de drie sets (één opfloater, verschil 1). p5 verliest op
C7. p4 en p6 gelijken op C7 en op C8–C10, dus de eerdere set, p4, wordt gebruikt.
De paring is p2–p1 en p4–p3, en p6–p5 in het volgende bracket.

### C8 -- maximale opfloaters (2.3.4)

*Tekst:* behalve in de laatste ronde, minimaliseer het aantal opfloaters dat een
maximale opfloater is (artikel 1.8: een speler is een maximale opfloater wanneer
hij al `MaxT` keer opfloatte).

*Gemeten door* `candidateQuality.maximumUpfloaters`: het aantal opfloaters van de
set waarvan `UpfloatCount` ten minste `MaxT` is. `UpfloatCount` telt de eerdere
ronden waarin de score van de speler aan het begin van de ronde lager was dan die
van de tegenstander. `CriterionC8(players, maxT)` is de geëxporteerde vorm; hij
telt over de spelers die hij krijgt.

*Voorbeeld:* `totalRounds` is 9, dus `MaxT = 2 + floor(9/5) = 3`. X floatte drie
keer op, Y twee keer. Een set met X heeft C8 = 1, een set met Y heeft C8 = 0; als
C5–C7 gelijktrekken, wordt de set met Y gebruikt.

### C9 -- aantal keren dat een maximale opfloater opfloatte (2.3.5)

*Tekst:* behalve in de laatste ronde, minimaliseer het aantal keren dat een
maximale opfloater opfloatte.

*Gemeten door* `candidateQuality.maximumTimes`: voor de maximale opfloaters van
de set, de som van hun `UpfloatCount`. Een speler die al vaker opfloatte, weegt
zwaarder mee.

*Voorbeeld:* met `MaxT` 3 heeft een set met X (3 eerdere opfloats) C9 = 3 en een
set met Z (4 eerdere opfloats) C9 = 4. Bij gelijke C5–C8 wordt de set met X
gebruikt.

### C10 -- opfloaters uit de vorige ronde (2.3.6)

*Tekst:* behalve in de laatste ronde, minimaliseer het aantal opfloaters dat in
de vorige ronde opfloatte.

*Gemeten door* `candidateQuality.previousUpfloaters`: het aantal opfloaters van
de set waarvan het laatste element in de float-historie een opfloat is
(`CriterionC10` is de geëxporteerde vorm).

*Voorbeeld:* U floatte in ronde 4, V niet. Bij het indelen van ronde 5 heeft een
set met U C10 = 1, een set met V C10 = 0.

## Volgorde van vergelijking

`betterC6ToC10(left, right)` vergelijkt volledig ingedeelde brackets binnen één
groep met hetzelfde scoremultiset lexicografisch, en geeft aan of `left` beter
is:

1. de C6-vectoren zoals hierboven beschreven,
2. minder kleurschendingen (C7),
3. minder maximale opfloaters (C8),
4. een kleinere som van eerdere opfloats van maximale opfloaters (C9),
5. minder opfloaters uit de vorige ronde (C10).

Het kleinste aantal opfloaters (C5) ligt vast vóór de groepering; groepen worden
in scoremultisetvolgorde behandeld. Een criterium wordt alleen bekeken wanneer
alle voorgaande criteria gelijk zijn. Een kandidaat vervangt de huidige beste
alleen wanneer hij strikt beter is, dus bij gelijke kwaliteit blijft de eerste
set in de volgorde van artikel 4.1.3 staan.

*Voorbeeld:* een set die C6 = (2, 1) en vier kleurschendingen geeft, verslaat een
set met C6 = (2, 2) en geen: C6 heeft prioriteit boven C7.

## Laatste-ronde-versoepeling

In de laatste ronde worden C8, C9 en C10 niet toegepast (artikelen 2.3.4–2.3.6).
`qualityFor` geeft terug direct na C5–C7 wanneer `lastRound` waar is, dus die
velden blijven 0 voor elke kandidaat. De laatste ronde is de ronde waarin
`CurrentRound == totalRounds`. C5, C6 en C7 blijven van kracht.

*Voorbeeld:* in ronde 9 van 9 telt een set met een maximale opfloater en een set
zonder bij gelijke C5–C7 als gelijk, en wordt de eerste set in de 4.1.3-volgorde
gebruikt.

## MaxT

`MaxT(totalRounds) = 2 + floor(totalRounds / 5)` (artikel 1.8.2), met
`totalRounds` de geplande lengte van het toernooi, niet het aantal gespeelde
ronden.

| Geplande ronden | MaxT |
| --- | --- |
| 1–4 | 2 |
| 5–9 | 3 |
| 10–14 | 4 |
| 15–19 | 5 |

## Sorteervolgordes bij de criteria

| Artikel | Wat gesorteerd wordt | Volgorde | Code |
| --- | --- | --- | --- |
| 4.1.3 | sets van K spelers uit een genummerde verzameling | kleinste verschillende volgnummer eerst | `forEachCombination` |
| 4.2 | mogelijke opfloaters | aflopende score, dan oplopende TPN | `groupPlayers`, `lowerPlayers` |
| 4.3 | verschuivers | spelers die wit zoeken op oplopende ARO en TPN, spelers die zwart zoeken op oplopende TPN; nummers vanuit het midden toegekend | `shifterOrder` |
| 3.2.5 | G1 (S1) | oplopende ARO, dan oplopende TPN | `SortByAROAscending` |
| 4.4 | G2 (T2) | diepte-eerstzoektocht in lexicografische volgorde van oplopende TPN; de eerste legale wordt gebruikt | `firstLegalTransposition` |

De ARO wordt vóór de vergelijking afgerond op een geheel getal, .5 naar boven
(`ComputeARO`, artikel 1.7.1). TPN betekent altijd het vaste toernooiparingsnummer.

*Voorbeeld van 4.3.3:* de gesorteerde lijst A, B, C, D, E, F, G krijgt de nummers
D, C, E, B, F, A, G. *Voorbeeld van 4.4.2:* G2 met A, B, C geeft {A,B,C},
{A,C,B}, {B,A,C}, {B,C,A}, {C,A,B}, {C,B,A}.

## Gerelateerde code

`DubovCandidateScore` en `NumDubovViolations` zijn nog geëxporteerd voor
aanroepers van de vroegere criteriuminspecteur. De paringsprocedure gebruikt ze
niet en hun zeven schendingsplekken volgen niet de nummering van 2026. De
criteria van 2026 worden geëvalueerd via `candidateQuality` zoals hierboven
beschreven.

## Interpretaties

- **C6 bij gelijke beginreeks.** Wanneer de ene verschillenreeks het begin van
  de andere is, is de kortste beter. De tekst zegt niet wat er dan gebeurt.
- **C9.** "Het aantal keren dat een maximale opfloater opfloatte" wordt gelezen
  als de som van de eerdere opfloat-aantallen van de maximale opfloaters in de
  set.
- **C8–C10 op de set.** Ze worden gemeten over de opfloaters van de beoordeelde
  set, niet over alle spelers van het bracket.
- **Opfloat-historie.** Een ronde telt als een opfloat wanneer de speler aan het
  begin van die ronde minder punten had dan de tegenstander, met 1 / ½ / 0 als
  scores. C10 gebruikt het laatste element in de float-historie van een speler.
- **C4 en de PAB.** Artikel 3.1.2 wordt getest door de bracket-procedure op de
  resterende spelers uit te voeren; een kandidaat voor wie die procedure niet
  voltooid kan worden, wordt overgeslagen.
- **Zoekgrenzen.** De tekst kent geen numerieke limiet. De code houdt de
  volgorde van artikel 4.4 en 4.1.3 aan en geeft `ErrBracketTooLarge` terug
  wanneer het gedeelde budget van 1.000.000 zoekstappen per bracket
  (herindelingen en diepte-eerstzoektocht naar transposities) of de grens van
  200.000 kandidaten voor één setopsomming wordt overschreden.

## Gerelateerde pagina's

- [Dubov-systeem](../../pairing-systems/dubov/) -- de volledige procedure, opties
  en kleurregels.
- [Completeerbaarheid](../completability/) -- de koppelingstest achter C4.
- [Kleurtoewijzing](../color-allocation/) -- hoe de kleuren van artikel 5.2 na het
  indelen worden gegeven.
