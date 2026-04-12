---
title: "Kleurverdeling"
linkTitle: "Kleurverdeling"
weight: 14
description: "Zeven kleurverdelingsalgoritmen vergeleken — Nederlands, Keizer, Lim, Double-Swiss, Team en meer."
---

## Overzicht

Nadat de indeling bepaalt _wie_ tegen _wie_ speelt, bepaalt de
**kleurverdeling** _wie wit speelt en wie zwart_. Elk indelingssysteem
implementeert zijn eigen algoritme met verschillende prioriteitsregels, als
afspiegeling van de uiteenlopende filosofieën van de FIDE-reglementen.

Deze pagina vergelijkt de zeven kleurverdelingsalgoritmen in de codebase.

---

## Gemeenschappelijke begrippen

Alle algoritmen delen deze bouwstenen:

### Kleurvoorkeur

De kleurvoorkeur van een speler wordt afgeleid uit zijn partijgeschiedenis:

- **Onbalans**: aantal witpartijen minus aantal zwartpartijen.
- **Opeenvolgend aantal**: het aantal partijen met dezelfde kleur aan het
  einde van de geschiedenis.
- **Voorkeursrichting**: de kleur die de speler "nodig heeft" op basis van
  onbalans en opeenvolgende geschiedenis.

### Voorkeurssterkte

De meeste systemen classificeren voorkeuren op sterkte:

| Sterkte  | Voorwaarde                                              | Betekenis                                              |
| -------- | ------------------------------------------------------- | ------------------------------------------------------ |
| Absoluut | Onbalans $> 1$, of $\geq 2$ opeenvolgend dezelfde kleur | Speler _moet_ de tegenovergestelde kleur krijgen       |
| Sterk    | Onbalans $= 1$, of 1 opeenvolgend dezelfde kleur        | Speler _zou_ de tegenovergestelde kleur moeten krijgen |
| Mild     | Lichte voorkeur uit partijgeschiedenis                  | Mooi meegenomen, maar niet vereist                     |
| Geen     | Gebalanceerde geschiedenis                              | Geen voorkeur                                          |

### Beperking: geen 3 opeenvolgende

De hardste beperking die alle systemen delen: geen speler mag 3
opeenvolgende partijen met dezelfde kleur spelen. Dit wordt gecontroleerd
als voorwaarde (bij compatibiliteit voor Lim, bij kleurverdeling voor
anderen).

---

## Algoritme 1: Nederlands / Burstein (swisslib 6-stappen)

Gebruikt door het Nederlandse (C.04.3) en Burstein-systeem (C.04.4.2).
Implementatie in `pairing/swisslib/color.go`.

Het algoritme volgt de `choosePlayerColor` van bbpPairings:

### Stap 1: compatibele voorkeuren

Als de twee spelers verschillende kleuren prefereren (of ten minste een
geen voorkeur heeft), ken beide hun gewenste kleur toe. Dit lost de
meerderheid van de gevallen op.

### Stap 2: absoluut wint

Als een speler een **absolute** voorkeur heeft en de ander niet, wint de
absolute voorkeur. De andere speler krijgt de tegenovergestelde kleur.

### Stap 3: sterk verslaat niet-sterk

Als een speler een **sterke** voorkeur heeft en de ander slechts een milde
voorkeur of geen, wint de sterke voorkeur.

### Stap 4: eerste kleurverschil

Wanneer beide spelers dezelfde voorkeurssterkte hebben, loop **achterwaarts**
door hun partijgeschiedenissen tegelijk. Bij de eerste ronde waar de ene
speler wit had en de ander zwart (het "eerste kleurverschil"-punt), krijgt de
speler die daar de gewenste kleur had nu de tegenovergestelde kleur.

Bijvoorbeeld: als beiden wit prefereren en in ronde 3 speler A wit had en
speler B zwart, dan krijgt speler B wit in de huidige ronde (omdat speler A
het recentst had op het divergentiepunt).

### Stap 5: rang-tiebreak

Als de geschiedenissen identiek zijn over alle ronden, krijgt de speler met
de hogere rang (lager TPN) zijn gewenste kleur.

### Stap 6: bordafwisseling

Voor ronde 1-indelingen (geen partijgeschiedenis) geven oneven borden wit aan
de hoger gerangschikte speler en even borden aan de lager gerangschikte
(of andersom, afhankelijk van de `TopSeedColor`-optie). Dit zorgt voor
kleurafwisseling over de bordenlijst.

### Topscorer-regels

Wanneer beide spelers topscorers zijn (in de hoogste niet-lege scoregroep),
worden de absolute/sterke-onderscheidingen uit C3 versoepeld. Dit voorkomt
dat de leiders niet tegen elkaar ingedeeld kunnen worden vanwege kleurbeperkingen,
ten koste van het mogelijk geven van een derde opeenvolgende partij met
dezelfde kleur aan een speler.

---

## Algoritme 2: Dubov

Gebruikt door het Dubov-systeem (C.04.4.1). Delegeert naar het swisslib-
algoritme (dezelfde 6-stappenprocedure als Nederlands/Burstein) nadat de
Dubov-specifieke indelingsfase is afgerond. De kleurvoorkeuren die C6
(kleurvoorkeurschendingen) gebruikt tijdens de indeling zijn dezelfde als
die tijdens de verdeling.

---

## Algoritme 3: Keizer (swisslib-delegatie)

Gebruikt door het Keizer-indelingssysteem. Implementatie in
`pairing/keizer/keizer.go`, met delegatie naar `pairing/swisslib/color.go`.

De Keizer-indeling bouwt volledige kleurhistories op voor beide spelers
(forfaits worden uitgesloten; byes produceren `ColorNone`) en geeft deze
door aan de swisslib `AllocateColor`-functie. Dit betekent dat Keizer
dezelfde 6-stappencascade gebruikt als Nederlands en Burstein: compatibele
voorkeuren, absolute voorkeur wint, sterk verslaat niet-sterk, eerste
kleurverschil, rang-tiebreak en bordafwisseling.

TPN-waarden worden afgeleid uit de positie in de Keizer-rangschikking
(index + 1), en de topscorer-vlag is altijd `false` omdat het
Keizer-systeem niet de FIDE-topscorer-versoepelingen kent.

---

## Algoritme 4: Lim (Art. 5)

Gebruikt door het Lim-systeem (C.04.4.3). Implementatie in
`pairing/lim/color.go`.

Het Lim-algoritme is het meest onderscheidend, met **rondepariteitsbewustzijn**
en **mediaan-tiebreaking**:

### Ronde 1

Oneven TPN krijgt de beginkleur (standaard wit); even TPN krijgt het
omgekeerde.

### Art. 5.3: moet-afwisselen

Als een speler 2 opeenvolgende partijen met dezelfde kleur heeft, _moet_
hij de tegenovergestelde kleur krijgen. Als beide spelers deze regel
activeren en dezelfde kleur nodig hebben, signaleert het algoritme een
conflict (dit had tijdens de compatibiliteitscontrole bij de indeling moeten
worden ondervangen).

### Art. 5.2/5.6: even-ronde-egalisering

In even ronden krijgt de speler met meer partijen van een kleur de
tegenovergestelde kleur. Dit egaliseert actief de kleurbalans.

### Art. 5.5: oneven-ronde-afwisseling

In oneven ronden krijgt elke speler het tegenovergestelde van zijn laatst
gespeelde kleur. Dit creëert een natuurlijk afwisselingspatroon.

### Art. 5.4: geschiedenis-tiebreak met mediaan

Wanneer bovenstaande regels de toewijzing niet oplossen (beide spelers
hebben identieke beperkingen), loop achterwaarts door de partijgeschiedenis:

1. Vind de eerste ronde waar de twee spelers verschillende kleuren hadden.
2. De speler wiens positie **boven de mediaan** van de huidige scoregroep
   ligt krijgt prioriteit voor zijn gewenste kleur.

"Boven de mediaan" betekent dat de rang van de speler in de bovenste helft
van de scoregroep valt. Dit is een bewust voordeel voor hoger gerangschikte
spelers in de Lim-filosofie.

---

## Algoritme 5: Double-Swiss (Art. 4)

Gebruikt door het Double-Swiss-systeem (C.04.5). Implementatie in
`pairing/doubleswiss/color.go`.

Double-Swiss wijst kleuren toe voor **Partij 1** van elke 2-partijmatch
(Partij 2 keert automatisch de kleuren om). Het algoritme heeft 5 stappen:

### Stap 1: harde beperking (geen 3 opeenvolgende)

Als het toekennen van een specifieke kleur aan een speler 3 opeenvolgende
partijen met die kleur zou opleveren, ken de tegenovergestelde toe. Dit is
de enige absolute beperking.

### Stap 2: egaliseren

De speler met meer partijen van een kleur krijgt de tegenovergestelde kleur.
Dit balanceert de algehele kleurverdeling.

### Stap 3: afwisselen

Elke speler krijgt het tegenovergestelde van zijn laatst gespeelde kleur.

### Stap 4: ronde 1-bordafwisseling

In ronde 1 geven oneven borden wit aan de hoger gerangschikte speler; even
borden keren om. De `TopSeedColor`-optie bepaalt de beginkleur.

### Stap 5: rang-tiebreak

De speler met de hogere rang (lager TPN) krijgt wit.

Het verbod op 3 opeenvolgende (stap 1) is uniek voor Double-Swiss: het wordt
gecontroleerd als harde beperking tijdens kleurverdeling in plaats van
tijdens compatibiliteit (zoals bij Lim).

---

## Algoritme 6: Team Swiss (Art. 4, 9 stappen)

Gebruikt door het Team Swiss-systeem (C.04.6). Implementatie in
`pairing/team/color.go` en `pairing/team/color_pref.go`.

Dit is het meest complexe kleuralgoritme, met 9 stappen en het concept van
een **eerste team**:

### Kleurvoorkeurtypes

Team Swiss ondersteunt twee voorkeurberekeningswijzen:

| Type   | Beschrijving                                                                                                                      |
| ------ | --------------------------------------------------------------------------------------------------------------------------------- |
| Type A | Eenvoudig: kleurverschil $< -1$ of laatste 2 partijen zwart impliceert witvoorkeur. Symmetrisch voor zwart.                       |
| Type B | Sterk + mild: dezelfde condities geven "sterke" voorkeur; extra condities (kleurverschil $= \pm 1$, enz.) geven "milde" voorkeur. |

### Eerste team

Het **eerste team** in een indeling wordt bepaald door:

1. Hogere primaire score, of
2. Bij gelijkstand, hogere secundaire score, of
3. Als nog steeds gelijk, lager TPN (hogere plaatsing).

Het eerste-teamconcept geeft een team lichte prioriteit in ambigue gevallen.

### De 9 stappen

1. **Geen geschiedenis**: als geen van beide teams partijgeschiedenis heeft,
   wijs toe op basis van TPN-pariteit en initiële kleurinstelling.
2. **Een voorkeur**: als slechts een team een kleurvoorkeur heeft, honoreer
   die.
3. **Tegengestelde voorkeuren**: als de voorkeuren verschillen, honoreer
   beide.
4. **Sterk verslaat niet-sterk** (alleen Type B): als het ene team een
   sterke voorkeur heeft en het andere slechts een milde, wint de sterke.
5. **Lager kleurverschil**: het team met het lagere kleurverschil (minder
   witpartijen ten opzichte van zwartpartijen) krijgt wit.
6. **Afwisseling uit geschiedenis**: loop achterwaarts door de
   partijgeschiedenissen om het meest recente divergentiepunt te vinden.
   Het team dat aan de beurt is voor een wisseling krijgt die.
7. **Voorkeur eerste team**: honoreer de voorkeur van het eerste team.
8. **Afwisseling eerste team**: geef het eerste team het tegenovergestelde
   van hun laatste kleur.
9. **Afwisseling ander team**: geef het niet-eerste team het
   tegenovergestelde van hun laatste kleur.

Stappen 7--9 zijn progressieve terugvalopties voor wanneer alle eerdere
regels onbepaald zijn.

---

## Vergelijkingstabel

| Eigenschap             | Nederlands/Burstein               | Keizer                            | Lim                                         | Double-Swiss      | Team Swiss                                |
| ---------------------- | --------------------------------- | --------------------------------- | ------------------------------------------- | ----------------- | ----------------------------------------- |
| Stappen                | 6                                 | 6 (swisslib)                      | 5 + mediaan                                 | 5                 | 9                                         |
| Voorkeursniveaus       | Absoluut, Sterk, Mild, Geen       | Absoluut, Sterk, Mild, Geen       | Binair + moet-afwisselen                    | Binair            | Type A (eenvoudig) of Type B (sterk/mild) |
| Geschiedenisloop       | Achterwaarts naar eerste verschil | Achterwaarts naar eerste verschil | Achterwaarts + mediaan                      | N.v.t.            | Achterwaarts naar recente divergentie     |
| Rondepariteit          | Nee                               | Nee                               | Ja (even = egaliseren, oneven = afwisselen) | Nee               | Nee                                       |
| Mediaan-tiebreak       | Nee                               | Nee                               | Ja                                          | Nee               | Nee                                       |
| Eerste-entiteitbegrip  | Nee                               | Nee                               | Nee                                         | Nee               | Ja (eerste team)                          |
| 3-opeenvolgendcontrole | Tijdens compatibiliteit           | Tijdens verdeling                 | Tijdens compatibiliteit                     | Tijdens verdeling | N.v.t. (inherent aan voorkeuren)          |
| Bordafwisseling        | Ronde 1                           | Ronde 1                           | Ronde 1 op TPN-pariteit                     | Ronde 1           | N.v.t.                                    |
| Topscorer-uitzondering | Ja                                | Nee                               | Nee                                         | Nee               | Nee                                       |

---

## Ontwerprationale

De verschillende algoritmen weerspiegelen verschillende filosofieën:

- **Nederlands/Burstein**: maximaliseert kleurtevredenheid over het hele
  toernooi via op-geschiedenis-gebaseerde tiebreaking. De achterwaartse loop
  zorgt ervoor dat langetermijn-kleurpatronen worden meegewogen, niet alleen
  recente partijen.

- **Keizer**: delegeert naar hetzelfde swisslib-algoritme als
  Nederlands/Burstein. Het Keizer-systeem heeft geen FIDE-reglementen om aan
  te voldoen, maar het gebruik van de volledige cascade biedt dezelfde
  kwaliteit van kleurbalans als bij de Zwitserse systemen.

- **Lim**: benadrukt eerlijkheid op rondeniveau. Even ronden egaliseren
  actief; oneven ronden wisselen af. De mediaan-tiebreak voegt een subtiel
  rangvoordeel toe dat betere toernooi-prestaties beloont.

- **Double-Swiss**: geeft prioriteit aan de matchervaring. Aangezien elke
  match 2 partijen met omgekeerde kleuren omvat, is alleen de kleur van
  Partij 1 van belang voor de reeks. Het algoritme is eenvoudiger omdat de
  matchstructuur inherent kleurbalans biedt.

- **Team Swiss**: voegt organisatorische complexiteit toe voor
  teamwedstrijden. Het eerste-teamconcept zorgt ervoor dat het team met
  betere toernooi-prestaties lichte prioriteit krijgt bij ambigue
  kleurbeslissingen, als weerspiegeling van de competitieve hiërarchie.

---

## Gerelateerde pagina's

- [Nederlandse criteria](../dutch-criteria/) — C10--C13 kleuroptimalisatie
  in de kantgewichten.
- [Lim Exchange Matching](../lim-exchange/) — compatibiliteitscontroles
  inclusief kleurhaalbaarheid.
- [Concepten: kleuren en balans](/docs/concepts/colors/) — algemene
  introductie tot kleurverdeling voor schaakspelers.
