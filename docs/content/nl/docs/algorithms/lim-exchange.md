---
title: "Lim-uitwisselingsindeling"
linkTitle: "Lim-uitwisseling"
weight: 12
description: "De uitwisselingsprocedure van artikel 4 uit FIDE C.04.4.3."
---

## Overzicht

Deze pagina beschrijft de uitwisselingsengine van de Lim-pairer. Dit is de
procedure van artikel 4, toegepast nadat artikel 2.3 spelers heeft verplaatst
en de scoregroep even heeft gemaakt. Het is geen greedy matching en ook geen globale
matching.

## Compatibiliteit (art. 2.1)

Een paar moet nieuw zijn en kleuren toelaten zonder drie opeenvolgende gelijke
kleuren of een verschil van drie in de kleuraantallen. Verboden paren zijn ook
niet compatibel. In de laatste ronde versoepelt de uitzondering van artikel 6
het kleurgedeelte, zodat een beschikbaar nieuw paar binnen dezelfde score
voorrang heeft.

## Voorstel voor paren (art. 2.4)

Spelers worden op startnummer geordend en in een bovenste en onderste helft
verdeeld. Het voorstel paart de eerste speler van de bovenste helft tegen de
eerste van de onderste helft, de tweede tegen de tweede, enzovoort. De groep
heeft al de verplichte verplaatsingen van 2.3.1--2.3.3 gehad; alleen als ze
daarna nog oneven is, wordt een 2.3.4-floater gekozen.

## Uitwisselingsregels (art. 4.1--4.4)

De voorgestelde paren worden op volgorde getoetst. Bij neerwaarts indelen
begint dat met #1; bij opwaarts indelen begint dat met #n. Bij een mislukt neerwaarts
voorstel wordt de laagst genummerde speler uitgewisseld tot een compatibele
paar is gevonden; opwaarts gebruikt de overeenkomstige hoogst genummerde
speler. De implementatie volgt de uitwisselingsreeksen in de tabellen van 4.2
en 4.3 en behoudt een eerder compatibel paar terwijl het volgende voorstel
wordt onderzocht.

Artikel 4.4 behoudt het eerdere paar wanneer de volgende speler geen
compatibele tegenstander kan krijgen. Onder 4.4.1 wordt de speler voor wie dat
mislukt uitgewisseld met de floater die een oorspronkelijk oneven groep even
maakte. Onder 4.4.2, in een oorspronkelijk even groep, wordt die speler
floater en wordt hij gepaard met de laagst genummerde overblijvende speler,
zodat de groep even blijft.

Er is geen willekeurige terugval. De ontstane floaters worden verder verwerkt
volgens artikel 3 en 2.6; een onmogelijke ronde levert `PairingImpossible` op,
geen onvolledige resultaten met extra byes.

## Floaterregels van de engine

Floaterselectie volgt art. 3.2--3.5 en 3.9: de dwingende regels van 3.3/3.5
kiezen of wisselen een compatibele floater, en 3.9.2 ordent alternatieven D,
B, C, A. Artikel 3.10 beschermt een floater die in de vorige ronde aan
de dezelfde kant is verplaatst (geen bye) tegen een 2.3.4-float als 3.10.1 en
3.10.2 dat toelaten. De indeling van floaters volgt 3.6--3.8, met de Maxi-vergelijking
van de twee uitgewisselde tegenstanders.

## Mediaangroep en cracking

Lim verwerkt groepen boven de mediaan neerwaarts, groepen eronder opwaarts en
de mediaangroep als laatste. Een geblokkeerde mediaangroep wordt via cracking
van art. 2.6 uitgebreid. Als geen speler de mediaanscore heeft, interpreteert
deze implementatie de binnenkomende floaters als de mediaangroep op die
positie; 2.6 geldt dan. In de mediaangroep interpreteert het de typen van 3.9
tegen de al gepaarde groep eronder. Dit zijn interpretaties van punten
waarover de FIDE-tekst niet beslist.

Ook interpreteert het 4.4.1 voor de speler voor wie het indelen mislukt, zodra
de groep een 2.3.4-floater heeft gekozen, ook als 2.3.1--2.3.3 een
oorspronkelijk even groep oneven maakten.

## Gerelateerde pagina's

- [Lim-indeling](/docs/pairing-systems/lim/) beschrijft het volledige systeem.
- [Kleurtoewijzing](../color-allocation/) beschrijft gedeelde kleurbegrippen.
