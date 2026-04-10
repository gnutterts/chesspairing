---
title: "Kleuren en balans"
linkTitle: "Kleuren & balans"
weight: 5
description: "Waarom kleurtoewijzing belangrijk is en hoe indelingssystemen wit- en zwart-toewijzingen balanceren."
---

## Waarom kleur ertoe doet

Wit zet als eerste in schaken. Statistisch gezien wint wit vaker dan
zwart -- rond de 55% op meesterniveau. Hoewel dit verschil klein is in
één partij, staat een speler die herhaaldelijk dezelfde kleur krijgt
systematisch in voor- of nadeel over het hele toernooi. Een eerlijke
kleurverdeling is een kernverantwoordelijkheid van elk indelingssysteem.

## Kleurvoorkeuren

Na elke ronde bouwt een speler een kleurhistorie op. Uit die historie
berekent het indelingssysteem een **kleurvoorkeur** -- de kleur die de
speler idealiter in de volgende ronde zou moeten krijgen. Voorkeuren
kennen drie sterktes:

- **Absolute voorkeur.** De speler _moet_ deze kleur krijgen.
  Treedt op wanneer de kleurbalans meer dan 1 afwijkt (bijv. drie keer
  wit en één keer zwart) of wanneer de speler de laatste twee
  opeenvolgende ronden dezelfde kleur heeft gehad. Het schenden van een
  absolute voorkeur zou drie partijen achtereen met dezelfde kleur
  opleveren, wat in de meeste systemen verboden is.

- **Sterke voorkeur.** De speler _zou_ deze kleur moeten krijgen. De
  kleuraantallen zijn ongelijk (bijv. twee keer wit en één keer zwart)
  maar niet kritiek. De indeling probeert hieraan te voldoen, maar kan
  het opzijzetten indien nodig.

- **Milde voorkeur.** De speler _heeft liever_ deze kleur, gebaseerd
  op afwisseling (het tegenovergestelde van de kleur in de vorige
  ronde). Dit is de zwakste voorkeur en wordt als eerste opgeofferd
  wanneer beperkingen botsen.

Een speler zonder eerdere partijen heeft geen kleurvoorkeur.

## Het doel

Het kleurtoewijzingssysteem streeft twee doelen tegelijk na:

1. **Kleuren afwisselen** van ronde tot ronde. Speelde je vorige ronde
   wit, dan zou je deze ronde zwart moeten spelen.
2. **Kleuraantallen gelijktrekken** over het toernooi. Je totale
   aantal wit-partijen en zwart-partijen moet zo dicht mogelijk bij
   elkaar liggen.

Deze twee doelen zijn meestal in lijn maar kunnen botsen. In dat geval
heeft het vermijden van drie opeenvolgende partijen met dezelfde kleur
voorrang boven het gelijktrekken.

## Kleurtoewijzing vindt plaats na de indeling

Een belangrijk architecturaal punt: **kleurtoewijzing is een aparte stap
die wordt uitgevoerd nadat de indeling heeft bepaald wie tegen wie
speelt.** De indelingsalgoritmen houden rekening met kleurbeperkingen bij
het bepalen of twee spelers _ingedeeld_ kunnen worden (een absoluut
kleurconflict maakt een indeling ongeldig), maar de daadwerkelijke
wit/zwart-toewijzing per bord gebeurt daarna.

Deze scheiding houdt de indelingslogica gericht op het
constraint-satisfaction probleem (wie speelt tegen wie) terwijl het
kleurtoewijzingsprobleem wordt gedelegeerd aan een apart algoritme.

## Hoe elk systeem kleur afhandelt

### Dutch, Burstein en Dubov

Deze drie systemen delen dezelfde kleurtoewijzingscode in het
`swisslib`-pakket. Het algoritme volgt een 6-staps prioriteit:

1. **Compatibele voorkeuren.** Als de ene speler wit wil en de andere
   zwart (of geen voorkeur heeft), worden beiden tevreden gesteld.
2. **Absolute voorkeur wint.** Als slechts één speler een absolute
   voorkeur heeft, of één een sterker onevenwicht heeft, krijgt die
   speler de gewenste kleur.
3. **Sterke voorkeur wint.** Als één speler een sterke voorkeur heeft
   en de ander niet, wordt de sterke voorkeur gehonoreerd.
4. **Kleurhistorie-tiebreak.** Loop achteruit door de kleurhistorie van
   beide spelers en zoek de meest recente ronde waarin ze een
   verschillende kleur hadden. Wissel op basis van dat verschil.
5. **Rang-tiebreak.** Als beide spelers dezelfde kleur willen met
   gelijke sterkte en identieke historie, krijgt de hoger gerangschikte
   speler de voorkeur.
6. **Bord-afwisseling.** Als geen van beide spelers een voorkeur heeft
   (bijv. ronde 1), wissel per bordnummer: de hoger gerangschikte
   speler krijgt wit op oneven borden, zwart op even borden. De
   TopSeedColor-optie kan dit patroon omdraaien.

### Lim

Het Lim-systeem gebruikt een rondepariteitsaanpak:

- **Even ronden** streven naar het _gelijktrekken_ van kleuraantallen
  (als je meer wit hebt gespeeld, krijg je zwart).
- **Oneven ronden** streven naar _afwisseling_ (het tegenovergestelde
  van je laatste kleur).
- Wanneer beide spelers dezelfde kleur willen, beslist een
  **mediaan-tiebreak**: in de bovenste helft van de stand krijgt de
  hoger gerangschikte speler de voorkeur; in de onderste helft de
  lager gerangschikte speler.
- De verplichte regel (niet drie keer dezelfde kleur achtereen) heeft
  altijd voorrang.

### Double-Swiss

Bij Double-Swiss is elke ronde een tweekamp waarbij kleuren automatisch
afwisselen tussen de partijen. "Kleur" verwijst hier naar wie wit krijgt
in partij 1. Het systeem hanteert een harde beperking: **geen speler mag
drie ronden achtereen dezelfde kleur in partij 1 hebben.** Daarbuiten
volgt het een 5-staps prioriteit: gelijktrekken, afwisselen,
rang-tiebreak en bord-afwisseling.

### Team Swiss

Team Swiss gebruikt een 9-staps kleurtoewijzingsproces, het meest
complexe van alle systemen. Het introduceert het **eerste-team concept**:
het team met de hogere score (of hogere secundaire score, of lager
rangnummer) is het "eerste team" en krijgt voorrang bij tiebreaks.

De 9 stappen behandelen: initiële rondetoewijzing, toekenning van
enkele voorkeuren, bevrediging van tegengestelde voorkeuren, sterk vs.
mild (voor Type B voorkeursmodus), kleurdifferentie-vergelijking,
afwisseling vanaf de meest recente afwijkende ronde,
eerste-team-voorkeur, eerste-team-afwisseling en ander-team-afwisseling.

### Keizer

Keizer gebruikt de eenvoudigste kleurtoewijzing: de hoger gerangschikte
speler krijgt wit, tenzij die in de laatste partij wit had, in welk
geval de kleuren worden omgedraaid. Er is geen meertraps-prioriteit of
complexe historie-analyse -- alleen simpele afwisseling op basis van de
laatste partij.

### Round-Robin

Bij round-robin worden kleuren bepaald door de Berger-tabel zelf.
Bord 1 (met de vaste speler) wisselt elke ronde. Alle andere borden
wijzen wit toe aan de speler met de lagere tabelpositie. Bij een dubbele
round-robin met kleurbalancering ingeschakeld, draait de tweede cyclus
alle kleurtoewijzingen om zodat elk paar één partij per kleur speelt.

## Verder lezen

Voor de wiskundige details achter het kleuralgoritme van elk systeem,
zie [Kleurtoewijzing](/docs/algorithms/color-allocation/).
