---
title: "Floaters"
linkTitle: "Floaters"
weight: 6
description: "Wanneer een speler buiten de eigen scoregroep moet worden ingedeeld — upfloaters en downfloaters."
---

In een [Zwitsers toernooi](/docs/concepts/swiss-system/) worden spelers gegroepeerd op score en idealiter ingedeeld tegen tegenstanders in dezelfde scoregroep. Maar scoregroepen werken niet altijd mee. Wanneer een groep een oneven aantal spelers heeft, of wanneer interne beperkingen een volledige matching verhinderen, moet minstens één speler de groep verlaten om elders een tegenstander te vinden. Die speler heet een **floater**.

## Downfloaters en upfloaters

Een floater beweegt in één van twee richtingen:

- **Downfloater** -- een speler die afdaalt naar een lagere scoregroep om een tegenstander te vinden. De downfloater speelt tegen iemand met minder punten, wat een makkelijkere partij dan verwacht oplevert.
- **Upfloater** -- een speler die opstijgt naar een hogere scoregroep. De upfloater speelt tegen een sterkere tegenstander, wat de partij moeilijker maakt dan verwacht.

Deze twee komen altijd in paren: elke downfloater uit de ene scoregroep levert een upfloater op in de groep die hem ontvangt. Als de 3-puntsgroep 7 spelers heeft, float er één speler naar beneden naar de 2,5-puntsgroep, waar een speler uit die groep in feite naar boven float door te worden gekoppeld aan de tegenstander met de hogere score.

## Waarom floating nodig is

Floating komt voor om meerdere redenen:

- **Oneven groepsgrootte.** Een scoregroep met een oneven aantal spelers kan niet iedereen intern indelen. Minstens één speler moet naar een aangrenzende groep worden gestuurd.
- **Eerder gespeelde tegenstanders.** Twee spelers in dezelfde scoregroep hebben elkaar mogelijk al eerder getroffen. Als er geen andere geldige indelingen zijn, moet iemand floaten.
- **Kleurbeperkingen.** Wanneer te veel spelers in een groep dezelfde kleur nodig hebben en de absolute kleurregel niet kan worden nageleefd, lost floating de impasse op.
- **Verboden paren.** Spelers die als verboden paar zijn aangemerkt (bijv. uit dezelfde club of familie) mogen niet tegen elkaar spelen, wat de interne matching verder beperkt.

De indelingsengine probeert het aantal floaters te minimaliseren, omdat floating het competitieve evenwicht verstoort dat Zwitserse indelingen nastreven.

## Floating-tracking

Elk indelingssysteem houdt de floating-historie bij om te voorkomen dat dezelfde speler ronde na ronde float. De details verschillen per systeem.

### Dutch- en Burstein-systeem

De Dutch- en Burstein-engines registreren per ronde de floating-richting van elke speler en houden **opeenvolgende floats in dezelfde richting** bij. De optimalisatiecriteria (C14 t/m C21 in de FIDE-reglementen) bestraffen indelingen die een speler in dezelfde richting zouden laten floaten als in de vorige ronde -- of zelfs twee ronden geleden. Deze criteria zijn gecodeerd als gewichten in de [Blossom matching](/docs/algorithms/blossom/)-graaf, zodat het algoritme herhaald floaten vanzelf vermijdt wanneer er betere alternatieven zijn.

Het systeem houdt specifiek bij:

- Of de speler in de meest recente ronde naar beneden of naar boven heeft gefloat.
- Hoeveel opeenvolgende ronden de speler in dezelfde richting heeft gefloat.
- Of de speler naar dezelfde scoregroep heeft gefloat als in de vorige ronde (C14 bestraft dit specifiek).

### Lim-systeem

Het [Lim-indelingssysteem](/docs/pairing-systems/lim/) hanteert een andere aanpak en classificeert elke potentiële floater in één van vier typen op basis van twee factoren: of de speler al vanuit een hogere groep naar de huidige scoregroep is gefloat, en of er een compatibele tegenstander in de aangrenzende groep beschikbaar is.

| Type  | Al gefloat? | Compatibele tegenstander in aangrenzende groep? |
| ----- | ----------- | ----------------------------------------------- |
| **A** | Ja          | Nee                                             |
| **B** | Ja          | Ja                                              |
| **C** | Nee         | Nee                                             |
| **D** | Nee         | Ja                                              |

Type A is het meest benadeeld (al één keer gefloat en geen compatibele partner om naartoe te gaan), terwijl Type D het minst benadeeld is (nog niet gefloat en opties beschikbaar). Bij het selecteren van welke speler moet floaten, geeft het Lim-systeem de voorkeur aan Type D-kandidaten, waarbij degene wordt gekozen die de kleurbalans binnen de resterende groep het beste gelijk trekt. Bij het naar beneden floaten heeft de speler met het laagste nummer de voorkeur; bij het naar boven floaten de speler met het hoogste nummer.

## Floating-richting en optimalisatie

Alle Zwitserse indelingssystemen delen hetzelfde doel: de competitieve impact van floating minimaliseren. In de praktijk betekent dit:

1. **Minimaliseer het totale aantal floaters.** Minder floaters betekent dat meer spelers tegen tegenstanders op hun eigen niveau spelen.
2. **Vermijd het herhalen van een float voor dezelfde speler.** Een speler die vorige ronde naar beneden floatte, zou deze ronde niet opnieuw naar beneden moeten floaten als het te vermijden is.
3. **Verdeel floats over spelers.** Als floating noodzakelijk is, spreid het dan over verschillende spelers in plaats van steeds dezelfde te belasten.
4. **Geef de voorkeur aan kleinere scoreverschillen.** Een speler die float van 3 punten naar 2,5 punten is minder verstorend dan een die float van 3 naar 2.

Deze overwegingen zijn in elk systeem anders gecodeerd -- als gewichten van optimalisatiecriteria in het Dutch- en Burstein-systeem, als floater-typeclassificaties in het Lim-systeem, en als selectieregels in het Dubov-systeem -- maar het onderliggende principe is hetzelfde.

## Zie ook

- [Overzicht Zwitsers systeem](/docs/concepts/swiss-system/) -- hoe scoregroepen worden gevormd en verwerkt
- [Lim-indelingssysteem](/docs/pairing-systems/lim/) -- de floater-typeclassificatie in detail
