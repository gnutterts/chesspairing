---
title: "Rangnummers en plaatsing"
linkTitle: "Rangnummers"
weight: 9
description: "Tournament pairing numbers, initiële rangorde, en hoe de plaatsingsvolgorde indelingen beïnvloedt."
---

Elke speler in een schaaktoernooi krijgt een **rangnummer** toegewezen, formeel het Tournament Pairing Number (TPN) genoemd. Het wordt los van de spelersidentiteit opgeslagen en bepaalt de vaste plaatsingspositie ten opzichte van alle andere deelnemers.

## Hoe rangnummers worden toegewezen

Vóór de eerste ronde worden alle spelers gerangschikt op:
1. Rating (hoogste eerst)
2. FIDE-titel (GM > IM > WGM > FM > WIM > CM > WFM > WCM > geen)
3. Alfabetisch op naam
4. Oorspronkelijke invoervolgorde

De gesorteerde positie wordt het **Tournament Pairing Number (TPN)** van de speler: positie 1 is de hoogst gerangschikte speler, positie 2 de op een na hoogste, enzovoort. Late instromers (spelers die na ronde 1 meedoen) worden aan het eind van de lijst geplaatst en krijgen de daaropvolgende nummers.

Het belangrijke onderscheid met historische systemen:
- Het **TPN** is een vaste eigenschap voor het hele toernooi, expliciet toegewezen vóór de eerste ronde. Het weerspiegelt de pre-toernooi rangorde.
- De **plaatsingspositie** binnen een scoregroep wordt bepaald door spelers te sorteren op hun TPN.

## Waar rangnummers een rol spelen

Het TPN beïnvloedt vrijwel elk aspect van het indelingsproces:

### Scoregroep-splitsing

In het [Dutch-systeem](/docs/pairing-systems/dutch/) wordt elke scoregroep in twee helften verdeeld -- S1 (de bovenste helft op basis van plaatsing) en S2 (de onderste helft). De verdeling wordt bepaald door TPN-volgorde: de bovenste helft van de groep op TPN vormt S1, de rest vormt S2. S1-spelers worden vervolgens ingedeeld tegen S2-spelers. Dit zorgt ervoor dat de hoogst gerangschikte spelers binnen een scoregroep tegenstanders uit de onderste helft treffen, wat gebalanceerde partijen oplevert.

### Bordvolgorde

Nadat indelingen zijn gegenereerd, worden partijen in een specifieke volgorde aan borden toegewezen. De primaire sortering is op de hoogste score in elke indeling (topborden bevatten de spelers met de hoogste scores). Binnen hetzelfde scoreniveau wordt de indeling met het laagste minimum-TPN op het hogere bord geplaatst. De partij met de toernooileider verschijnt dus op bord 1.

### Bye-toewijzing

Bij het selecteren van welke speler de [pairing-allocated bye](/docs/concepts/byes/) (PAB) krijgt, geven de meeste systemen de voorkeur aan de speler met het hoogste TPN (laagste rangorde) in de laagste scoregroep. Het hoogste TPN hoort doorgaans bij de laagst geratingde speler, wat hem of haar de natuurlijke bye-kandidaat maakt.

### Kleurtoewijzing

In het Dutch-systeem krijgt bij twee spelers zonder kleurvoorkeur de hoger gerangschikte speler de beginkleur als diens vaste TPN oneven is, en de tegengestelde kleur als het TPN even is. Andere systemen behouden hun systeemspecifieke kleurregel, waaronder afwisseling per bord waar die is voorgeschreven.

Wanneer beide spelers kleurvoorkeuren van gelijke sterkte hebben, krijgt de hoger gerangschikte speler (lager TPN) de gewenste kleur.

### Floater-selectie

Het TPN beïnvloedt welke speler [float](/docs/concepts/floaters/) wanneer een scoregroep niet intern kan worden ingedeeld. In het Lim-systeem worden downfloaters geselecteerd vanaf het laagste TPN (sterkste speler), terwijl upfloaters worden geselecteerd vanaf het hoogste TPN (zwakste speler). Dit ontwerpprincipe houdt de sterkste spelers waar mogelijk in hun natuurlijke scoregroep.

### Tiebreaking

De "pairing number"-tiebreaker gebruikt het TPN rechtstreeks als tiebreaker-waarde. Aangezien lagere TPN's overeenkomen met hoger geratingde spelers, begunstigt deze tiebreaker de hoger geratingde speler wanneer alle andere tiebreakers gelijk zijn.

## Round-robin: Varma-tabellen

In [round-robin toernooien](/docs/concepts/round-robin/) krijgen rangnummers een bijzondere betekenis omdat ze direct het speelschema bepalen via de Berger-tabellen. De volgorde waarin spelers worden genummerd, bepaalt wie tegen wie speelt in welke ronde.

Wanneer spelers uit meerdere federaties komen, is het wenselijk om ontmoetingen tussen spelers van dezelfde federatie in de vroege ronden te vermijden. De [Varma-tabellen](/docs/algorithms/varma-tables/) (gedefinieerd in FIDE C.05 Annex 2) bieden een federatie-bewuste methode voor het toewijzen van rangnummers, zodat spelers van hetzelfde land op posities in de Berger-tabel komen waar ze elkaar pas in latere ronden treffen.

Het Varma-toewijzingsalgoritme:

1. Groepeert spelers per federatie, gesorteerd op federatie-grootte (grootste eerst).
2. Wijst de spelers van elke federatie toe aan Varma-groepen, waarbij de groep met de meeste beschikbare plaatsen wordt gekozen.
3. Als een federatie te groot is voor één groep, worden spelers verspreid over meerdere groepen.

Dit ondersteunt toernooien met maximaal 24 spelers en werkt met de standaard Berger-tabelrotatie.

## Het PlayerEntry.ID-veld

In het chesspairing datamodel heeft elke speler een `ID`-veld dat dient als unieke identificatie gedurende het hele toernooi. Dit ID wordt gebruikt in indelingsresultaten, partijrecords en bye-entries. Het TPN wordt opgeslagen als het `PairingNumber`-veld in de `PlayerEntry` en blijft behouden in TRF-bestanden.

## Zie ook

- [Varma-tabellen algoritme](/docs/algorithms/varma-tables/) -- federatie-bewuste nummertoewijzing voor round-robin
- [Indelingssystemen](/docs/pairing-systems/) -- hoe verschillende systemen de plaatsingsvolgorde gebruiken
