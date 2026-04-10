---
title: "Het Zwitsers systeem"
linkTitle: "Zwitsers systeem"
weight: 1
description: "Hoe het Zwitsers indelingssysteem werkt: spelers met vergelijkbare scores indelen en herhalingen voorkomen."
---

## Het probleem dat het Zwitsers systeem oplost

Een round-robintoernooi met 40 spelers heeft 39 ronden nodig. De meeste
evenementen kunnen er 7 of 9 aan. Het Zwitsers systeem is bedacht om met
veel minder ronden dan een volledige round-robin een betrouwbare
rangschikking te produceren, door elke ronde spelers van vergelijkbare
sterkte tegen elkaar te indelen in plaats van elke mogelijke combinatie af
te werken.

Het idee stamt uit 1895 in Zürich, en de FIDE verfijnt de regels
sindsdien. Tegenwoordig is het Zwitsers systeem het standaardformaat
voor de meeste geratte schaakevenementen wereldwijd.

## Kernprincipes

Elk Zwitsers systeem, ongeacht de variant, volgt dezelfde drie
principes:

1. **Vergelijkbare scores spelen tegen elkaar.** Spelers met gelijke
   (of dichtbijliggende) scores worden tegen elkaar ingedeeld. Dit
   concentreert de beslissende partijen naarmate het toernooi vordert
   bij de top van de stand.

2. **Geen herhaalde indelingen.** Twee spelers mogen niet meer dan één
   keer tegen elkaar spelen in hetzelfde toernooi. Dit dwingt de indeling
   om elke ronde verse tegenstanders te vinden.

3. **Kleurbalans.** Elke speler zou van ronde tot ronde moeten
   afwisselen tussen wit en zwart, en het totale aantal partijen per
   kleur moet zo gelijk mogelijk blijven.

Deze drie principes creëren een constraint-satisfaction-probleem. Het
systeem moet indelingen vinden die tegelijkertijd aan alle drie voldoen,
waarbij zwakkere beperkingen pas worden losgelaten wanneer de sterkere
geen andere optie overlaten.

## Scoregroepen en brackets

Na elke ronde worden spelers gegroepeerd op basis van hun huidige score.
Een **scoregroep** is de verzameling van alle spelers met hetzelfde
puntentotaal -- bijvoorbeeld iedereen met 3/4 na vier ronden.

Wanneer een scoregroep een oneven aantal spelers heeft, of wanneer de
herhalings- en kleurbeperkingen het onmogelijk maken om iedereen binnen
de groep te indelen, moeten een of meer spelers worden ingedeeld tegen iemand
uit een aangrenzende groep. Dit creëert **brackets** -- werkeenheden die
twee naburige scoregroepen kunnen omvatten. De speler die tussen groepen
beweegt heet een **floater**: een upfloater gaat naar een hogere groep,
een downfloater naar een lagere.

## Hoe een Zwitserse ronde wordt ingedeeld

Op hoog niveau volgt elke Zwitserse indeling dezelfde stroom:

1. **Rangschik alle actieve spelers** op score, daarna op initiële
   ranking (rangnummer) bij gelijke score.
2. **Vorm scoregroepen** uit de huidige stand.
3. **Wijs de bye toe.** Als het aantal spelers oneven is, ontvangt één
   speler een indelingsvrij (PAB). De bye gaat doorgaans naar de laagst
   gerankte speler in de laagste scoregroep die nog geen bye heeft
   gehad.
4. **Deel elke scoregroep in** van de hoogste score omlaag, onder de
   absolute criteria: geen herhaalde tegenstanders, respecteer
   kleurbeperkingen die niet mogen worden geschonden, en zorg dat de
   resterende spelers nog ingedeeld kunnen worden (completeerbaarheid).
5. **Optimaliseer.** Binnen de ruimte van geldige indelingen, pas
   kwaliteitscriteria toe om de beste indeling te kiezen -- bijvoorbeeld
   scoreverschillen tussen tegenstanders minimaliseren, floaters
   minimaliseren, of kleurvoorkeurvoldoening maximaliseren.
6. **Wijs kleuren toe.** Zodra de indeling vaststaat, bepaal wie wit en
   wie zwart speelt op basis van de kleurhistorie.
7. **Orden de borden.** Indelingen met hogere scores komen op de
   topborden.

De details van stappen 3-5 zijn waar de varianten verschillen. Elk
systeem definieert een eigen criteriahiërarchie, eigen matchingstrategie
en eigen tiebreakregels voor randgevallen.

## Zes Zwitserse varianten

chesspairing implementeert alle zes door de FIDE goedgekeurde Zwitserse
indelingssystemen, plus twee aanvullende systemen:

| Systeem                                             | Beschrijving                                                                                                                           |
| --------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| [Dutch](/docs/pairing-systems/dutch/)               | Het standaard FIDE-systeem (C.04.3). Gebruikt globale Blossom-matching om over alle scoregroepen tegelijk te optimaliseren.            |
| [Burstein](/docs/pairing-systems/burstein/)         | FIDE C.04.4.2. Scheidt het toernooi in plaatsingsronden en post-plaatsingsronden, met oppositie-indices om spelers te herrangschikken. |
| [Dubov](/docs/pairing-systems/dubov/)               | FIDE C.04.4.1. Gebruikt ARO voor bracketordening en kent een eigen set van tien indelingscriteria.                                     |
| [Lim](/docs/pairing-systems/lim/)                   | FIDE C.04.4.3. Verwerkt scoregroepen van de mediaan naar buiten en gebruikt exchange-gebaseerde matching binnen elke groep.            |
| [Double-Swiss](/docs/pairing-systems/double-swiss/) | FIDE C.04.5. Elke ronde bestaat uit een tweepartijmatch. Gebruikt lexicografische bracket-indeling.                                    |
| [Team Swiss](/docs/pairing-systems/team/)           | FIDE C.04.6. Zwitserse indeling voor teamcompetities, met 9-staps kleurtoewijzing en eerste-teamconcept.                               |

Alle zes delen dezelfde `Pairer`-interface, accepteren dezelfde
`TournamentState`-invoer en retourneren dezelfde `PairingResult`-uitvoer.
Je kunt het ene voor het andere wisselen zonder de rest van je code aan
te passen.

## Matchingalgoritmen

De Dutch- en Burstein-indelingen gebruiken Edmonds' maximum weight matching
(Blossom-algoritme) om optimale indelingen over alle brackets tegelijk te
vinden. De Dubov- en Lim-systemen gebruiken transpositie- en
exchange-gebaseerde matching binnen individuele scoregroepen. De
Double-Swiss- en Team Swiss-systemen gebruiken lexicografische
bracket-indeling.

Zie het gedeelte [Algoritmen](/docs/algorithms/) voor meer informatie
over deze algoritmen.
