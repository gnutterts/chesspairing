---
title: "Rangnummers en plaatsing"
linkTitle: "Rangnummers"
weight: 9
description: "Tournament pairing numbers, initiële rangorde, en hoe de plaatsingsvolgorde indelingen beïnvloedt."
---

Elke speler in een schaaktoernooi krijgt een **rangnummer** toegewezen, formeel het Tournament Pairing Number (TPN) genoemd. Dit nummer dient als identiteit van de speler binnen de indelingsengine en bepaalt de plaatsingspositie ten opzichte van alle andere deelnemers.

## Hoe rangnummers worden toegewezen

Vóór de eerste ronde worden alle spelers gesorteerd op rating (hoogste eerst). Spelers met dezelfde rating worden alfabetisch op naam geordend. De gesorteerde positie wordt de **initiële rangorde** van de speler: positie 1 is de hoogst geratingde speler, positie 2 de op een na hoogste, enzovoort.

Aan het begin van elke ronde worden actieve spelers opnieuw gerangschikt op huidige score (aflopend), met bij gelijke score de initiële rangorde (oplopend) als tiebreak. Deze herrangschikking levert het TPN voor die ronde op. Een speler die als initiële rang 5 begon maar na drie ronden aan de leiding staat, kan in ronde 4 TPN 1 hebben.

Het belangrijke onderscheid:

- **Initiële rangorde** ligt vast voor het hele toernooi. Het weerspiegelt de pre-toernooi ratingvolgorde.
- **TPN** wordt elke ronde herberekend. Het weerspiegelt de huidige standvolgorde.

## Waar rangnummers een rol spelen

Het TPN beïnvloedt vrijwel elk aspect van het indelingsproces:

### Scoregroep-splitsing

In het [Dutch-systeem](/docs/pairing-systems/dutch/) wordt elke scoregroep in twee helften verdeeld -- S1 (de bovenste helft op basis van plaatsing) en S2 (de onderste helft). De verdeling wordt bepaald door TPN-volgorde: de bovenste helft van de groep op TPN vormt S1, de rest vormt S2. S1-spelers worden vervolgens ingedeeld tegen S2-spelers. Dit zorgt ervoor dat de hoogst gerangschikte spelers binnen een scoregroep tegenstanders uit de onderste helft treffen, wat gebalanceerde partijen oplevert.

### Bordvolgorde

Nadat indelingen zijn gegenereerd, worden partijen in een specifieke volgorde aan borden toegewezen. De primaire sortering is op de hoogste score in elke indeling (topborden bevatten de spelers met de hoogste scores). Binnen hetzelfde scoreniveau wordt de indeling met het laagste minimum-TPN op het hogere bord geplaatst. De partij met de toernooileider verschijnt dus op bord 1.

### Bye-toewijzing

Bij het selecteren van welke speler de [pairing-allocated bye](/docs/concepts/byes/) (PAB) krijgt, geven de meeste systemen de voorkeur aan de speler met het hoogste TPN (laagste rangorde) in de laagste scoregroep. Het hoogste TPN hoort doorgaans bij de laagst geratingde speler, wat hem of haar de natuurlijke bye-kandidaat maakt.

### Kleurtoewijzing

Wanneer twee spelers geen kleurhistorie hebben (typisch in ronde 1), worden kleuren toegewezen op bordnummer: op oneven borden krijgt de hoger geplaatste speler (lager TPN) wit, op even borden zwart. Dit afwisselende patroon garandeert een gebalanceerde kleurverdeling in de eerste ronde.

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

In het chesspairing datamodel heeft elke speler een `ID`-veld dat dient als unieke identificatie gedurende het hele toernooi. Dit ID wordt gebruikt in indelingsresultaten, partijrecords en bye-entries. Het TPN wordt elke ronde berekend vanuit de ID-geïndexeerde spelersgegevens -- het wordt niet opgeslagen als permanent attribuut maar afgeleid uit de huidige score en initiële rangorde.

## Zie ook

- [Varma-tabellen algoritme](/docs/algorithms/varma-tables/) -- federatie-bewuste nummertoewijzing voor round-robin
- [Indelingssystemen](/docs/pairing-systems/) -- hoe verschillende systemen de plaatsingsvolgorde gebruiken
