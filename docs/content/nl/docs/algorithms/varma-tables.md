---
title: "Varma-tabeltoepassing"
linkTitle: "Varma-tabellen"
weight: 5
description: "Federatiebewuste toekenning van rangnummers voor round-robin toernooien (FIDE C.05 Annex 2)."
---

## Het probleem

In een [round-robin toernooi](../berger-tables/) bepaalt de Berger-rotatie
welke rangnummers elkaar in elke ronde tegenkomen. Als twee spelers van
dezelfde federatie toevallig opeenvolgende rangnummers hebben, worden ze
in een vroege ronde gepland. Toernooiorganisatoren geven er doorgaans de
voorkeur aan om ontmoetingen tussen federatiegenoten **te spreiden** over het
schema, zodat ze niet in vroege of late rondes samenkomen.

FIDE C.05 Annex 2 definieert het Varma-tabelsysteem: een verzameling
opzoektabellen die rangnummers verdelen over groepen, gecombineerd met een
toewijzingsalgoritme dat federaties over die groepen verdeelt.

De implementatie bevindt zich in `algorithm/varma/`.

---

## Varma-groepen

Voor $N$ spelers (even) worden de rangnummers $1, 2, \ldots, N$ verdeeld
in vier groepen met de labels **A**, **B**, **C** en **D**. De toewijzing
wordt bepaald door opzoektabellen voor elk even aantal spelers van 10 tot 24.

De belangrijkste eigenschap: spelers die aan dezelfde Varma-groep zijn
toegewezen, ontmoeten elkaar in rondes die maximaal gespreid zijn over het
schema. Spelers in verschillende groepen ontmoeten elkaar in tussenliggende
rondes. Door alle spelers van een bepaalde federatie in dezelfde groep te
plaatsen (indien mogelijk) worden hun onderlinge partijen optimaal gespreid.

### Tabelstructuur

Elke tabelingang koppelt een rangnummer aan zijn groep. Bijvoorbeeld bij
10 spelers:

| Pairing Number | 1   | 2   | 3   | 4   | 5   | 6   | 7   | 8   | 9   | 10  |
| -------------- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Group          | C   | D   | A   | A   | B   | C   | B   | A   | B   | D   |

De tabellen zijn opgeslagen als constante slices in `algorithm/varma/varma.go`.

### Kleine toernooien

Voor $N \leq 8$ spelers zijn de Varma-tabellen triviaal: alle spelers worden
in groep A geplaatst. Het voordeel van federatiescheiding treedt pas op bij
9 of meer spelers, waar de groepen genoeg leden hebben om zinvolle spreiding
te bieden.

### Oneven aantal spelers

Bij oneven $N$ rondt de implementatie op naar $N + 1$ (door een dummy toe te
voegen) en gebruikt de opzoektabel voor het even aantal. Het rangnummer
van de dummyspeler wordt vervolgens weggefilterd, waardoor $N$ toewijzingen
overblijven. De positie van de dummy wordt in feite de bye-positie.

---

## Het toewijzingsalgoritme

Gegeven de Varma-groepstabel en een lijst spelers met federatielabels,
verdeelt de `Assign`-functie spelers over rangnummers:

### Stap 1: Houd alle ingeschreven spelers aan

Verwijder teruggetrokken spelers niet. Elke ingeschreven speler behoudt zijn
plaats in de vaste Berger-tabel (C.05 6.6) en krijgt een rangnummer.

### Stap 2: Haal de groepstabel op

Zoek de Varma-groepstabel op (of bereken deze) voor het aantal spelers. Bij
meer dan 24 spelers geeft `Assign` een fout terug: de tabellen van C.05
Annex 2 dekken slechts 9 tot 24 spelers.

### Stap 3: Groepeer spelers per federatie

Verdeel spelers per federatie, gesorteerd op grootte (grootste eerst). Deze
greedy volgorde zorgt ervoor dat de grootste federaties als eerste een groep
kiezen, wat het spreidingsvoordeel maximaliseert.

### Stap 4: Best-fit-toewijzing

Voor elke federatie (grootste eerst):

1. **Vind de best passende groep.** Een geschikte groep is een groep die nog
   groot genoeg is om alle spelers van deze federatie te bevatten; de
   bibliotheek kiest de kleinste daarvan. Als geen resterende groep past --
   bij een contingent groter dan welke envelop dan ook -- wordt de federatie
   eerst verspreid over de groepen met de meeste beschikbare plaatsen; Annex 2
   kent daarvoor geen procedure, en de bibliotheek legt dit vast als eigen
   keuze.

2. **Wijs spelers toe aan plaatsen.** Binnen elke federatie worden spelers
   alfabetisch op weergavenaam geordend en toegewezen aan de beschikbare
   plaatsen in hun aangewezen groep(en).

De best-fit-strategie is een bin-packing-heuristiek. Ze garandeert geen
globaal optimale federatiescheiding, maar werkt in de praktijk goed omdat:

- De grootste federaties als eerste worden geplaatst en de gunstigste groepen
  krijgen.
- Kleinere federaties de resterende gaten opvullen.
- De structuur van de Varma-tabel ervoor zorgt dat zelfs suboptimale
  groepskeuzes een redelijke rondespreiding opleveren.

### Stap 5: Geef geordende spelers terug

De uitvoer is de spelerslijst geordend op toegewezen rangnummer, waarbij elke
speler zijn `PairingNumber` (1..n) meekrijgt. De [Berger-rotatie](../berger-tables/)
gebruikt die rangnummers vervolgens om het rondeschema op te stellen.

---

## Voorbeeld

Neem een toernooi met 12 spelers en drie federaties:

- Federatie X: 5 spelers
- Federatie Y: 4 spelers
- Federatie Z: 3 spelers

De Varma-tabel voor 12 spelers heeft groep A met 4 plaatsen, B met 3,
C met 3 en D met 2 (A: (4,5,9,10); B: (1,2,7); C: (6,8,12); D: (3,11)).

1. **Federatie X** (5 spelers, grootste): geen enkele groep past voor alle 5.
   Verspreid over de groepen met de meeste beschikbare plaatsen: 4 naar
   groep A, 1 naar groep B.
2. **Federatie Y** (4 spelers): de groepen A, B, C en D hebben nu 0, 2, 3 en
   2 vrije plaatsen; geen enkele groep past voor alle 4. Verspreiding: 3 naar
   groep C, 1 naar groep B.
3. **Federatie Z** (3 spelers): de vrije plaatsen zijn 0, 1, 0 en 2; geen
   enkele groep past voor alle 3. Verspreiding: 1 naar groep B, 2 naar groep D.

Resultaat: groep A krijgt 4 van X; groep B krijgt 1 van X, 1 van Y en 1 van
Z; groep C krijgt 3 van Y; groep D krijgt 2 van Z. Met contingenten van 5,
4 en 3 is er telkens geen envelop die het hele contingent nog kan bevatten,
dus beslaat elke federatie twee groepen.

---

## Wiskundige onderbouwing

De Varma-groepsstructuur maakt gebruik van een eigenschap van de
Berger-rotatie. In een schema met stap $s = n/2 - 1$ ontmoeten twee spelers
op posities $p$ en $q$ elkaar in ronde:

$$r = \frac{(q - p) \cdot s^{-1} \bmod (n - 1)}{1}$$

waarbij $s^{-1}$ de modulaire inverse van $s$ modulo $n - 1$ is (die bestaat
omdat $\gcd(s, n - 1) = 1$ voor de Berger-stap). Spelers waarvan het
positieverschil $|p - q|$ naar een ronde dicht bij $\lfloor (n-1)/2 \rfloor$
leidt, ontmoeten elkaar in het midden van het schema.

De Varma-tabellen zijn zo geconstrueerd dat spelers binnen dezelfde groep
positieverschillen hebben die rondes nabij het middelpunt opleveren -- zodat
de "afstand" tot ronde 1 en ronde $n - 1$ maximaal is.

---

## Beperkingen

- **Bereik van het spelersaantal.** Expliciete opzoektabellen bestaan alleen
  voor 9--24 spelers; bij meer dan 24 geeft `Assign` een fout terug in plaats
  van terug te vallen.
- **Onvolmaakte scheiding.** Wanneer een federatie groter is dan welke groep
  dan ook, moeten haar spelers over meerdere groepen worden verdeeld, wat het
  spreidingsvoordeel vermindert.
- **Alfabetische tiebreak.** Binnen een federatie worden spelers geordend op
  weergavenaam. Dit is een conventionele keuze zonder wiskundige betekenis.

---

## Complexiteit

Het toewijzingsalgoritme is $O(F \cdot G + N)$ waarbij $F$ het aantal
federaties is, $G = 4$ het aantal groepen, en $N$ het aantal spelers. Het
sorteren van federaties is $O(F \log F)$. Voor realistische toernooigroottes
is de gehele procedure in feite $O(N)$.

---

## Gerelateerde pagina's

- [Berger-tabelrotatie](../berger-tables/) -- het schema waarin de
  Varma-toewijzing invoer levert.
- [Round-robin indeling](/docs/pairing-systems/round-robin/) -- het
  indelingssysteem dat zowel Varma-tabellen als Berger-rotatie gebruikt.
