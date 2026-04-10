---
title: "Dubov-criteria"
linkTitle: "Dubov-criteria"
weight: 11
description: "De 10 criteria die Dubov-indelingen besturen — MaxT-tracking en transpositielimieten."
---

## Overzicht

Het Dubov-systeem (FIDE C.04.4.1) definieert 10 criteria ($C_1$--$C_{10}$)
die de indeling binnen elke scoregroep besturen. Anders dan het Nederlandse
systeem, dat alle optimalisatiecriteria codeert in Blossom-kantgewichten
voor globale matching, verwerkt het Dubov-systeem elke scoregroep
afzonderlijk met behulp van **transpositie-gebaseerde matching** met een
lexicografische kandidaatvergelijking.

De criteria zijn geïmplementeerd in `pairing/dubov/criteria.go`.

---

## Absolute criteria

### C1: geen herparingen

Identiek aan Nederlands C1. Twee spelers die al tegen elkaar gespeeld hebben
(forfaits uitgezonderd) worden niet opnieuw ingedeeld.

Implementatie delegeert naar `swisslib.C1NoRematches`.

### C3: geen absoluut kleurconflict

Twee spelers die beide een absolute kleurvoorkeur voor dezelfde kleur hebben,
worden niet ingedeeld. Een speler heeft een absolute kleurvoorkeur wanneer zijn
kleuronbalans groter is dan 1 of hij 2+ opeenvolgende partijen met dezelfde
kleur heeft gespeeld.

Anders dan bij Nederlands C3 bestaat er in het Dubov-systeem geen
topscorer-uitzondering.

Implementatie: `C3NoAbsoluteColorConflict` in `pairing/dubov/criteria.go`.

### Verboden paren

Paren die expliciet door de toernooiorganisator zijn verboden, worden
uitgesloten. Dit wordt samen met C1 en C3 gecontroleerd in
`SatisfiesAbsoluteCriteria`.

---

## De MaxT-parameter

Een onderscheidend kenmerk van het Dubov-systeem is de **MaxT**-limiet op
transposities. Voor elke scoregroep is het aantal transposities dat het
algoritme mag overwegen voordat het een kandidaatindeling accepteert, beperkt
tot:

$$\text{MaxT} = 2 + \left\lfloor \frac{R}{5} \right\rfloor$$

waarbij $R$ het aantal gespeelde ronden is.

| Gespeelde ronden | MaxT |
| ---------------- | ---- |
| 0--4             | 2    |
| 5--9             | 3    |
| 10--14           | 4    |
| 15--19           | 5    |

MaxT bepaalt de afweging tussen indelingskwaliteit en rekenkosten. Vroeg in
het toernooi zijn minder transposities nodig omdat de meeste indelingen
eenvoudig zijn. Naarmate het toernooi vordert en beperkingen zich opstapelen,
wordt meer flexibiliteit toegestaan.

Implementatie: functie `MaxT` in `pairing/dubov/criteria.go`.

---

## Optimalisatiecriteria (C4--C10)

De optimalisatiecriteria worden geëvalueerd op kandidaatindelingen en
lexicografisch vergeleken. Een `DubovCandidateScore` registreert de
schendingsaantallen voor C4--C10 plus een transpositie-index. De
`Compare`-methode voert een strikte lexicografische vergelijking uit:
C4-schendingen worden eerst vergeleken; bij gelijkheid C5; enzovoort.

### C4: minimaliseer upfloater-aantal

Minimaliseer het aantal spelers in de groep dat is opgedreven vanuit een
lagere scoregroep. Een upfloater is een speler wiens floatgeschiedenis
`FloatUp` bevat.

$$\text{C4 violations} = |\{p \in \text{bracket} : \text{FloatUp} \in \text{history}(p)\}|$$

Implementatie: `UpfloatCount` telt `FloatUp`-vermeldingen in de
floatgeschiedenis van de speler.

### C5: minimaliseer upfloater-scoresom

Minimaliseer onder de in C4 getelde upfloaters de som van hun scores. Dit
geeft de voorkeur aan het opdrijven van lager scorende spelers boven hoger
scorende.

$$\text{C5 violations} = \sum_{\substack{p \in \text{bracket} \\ \text{upfloater}(p)}} \text{score}(p)$$

### C6: minimaliseer kleurvoorkeurschendingen

Minimaliseer het aantal paren waarbij beide spelers een kleurvoorkeur
(sterk of absoluut) voor dezelfde kleur hebben. Anders dan Nederlands
C10--C13, dat vier niveaus onderscheidt, behandelt Dubov C6 alle
voorkeurconflicten gelijk.

### C7: minimaliseer MaxT-upfloater-schendingen

Tel het aantal upfloaters wiens upfloat-aantal MaxT overschrijdt. Een
speler die te vaak is opgedreven (meer dan MaxT keer) vormt een
C7-schending.

$$\text{C7 violations} = |\{p : \text{upfloatCount}(p) > \text{MaxT}\}|$$

### C8: minimaliseer opeenvolgende upfloaters

Tel het aantal spelers dat zowel in de huidige ronde als in de direct
voorgaande ronde is opgedreven. Opeenvolgende upfloats zijn storender dan
geïsoleerde.

### C9: minimaliseer MaxT-tegenstanderschendingen

Tel het aantal indelingen waarbij de tegenstander van een speler meer dan
MaxT keer is opgedreven. Dit spreidt de last van het tegenkomen van
upfloaters.

### C10: minimaliseer opeenvolgende MaxT-schendingen

Tel het aantal spelers dat zowel in de huidige ronde als in de voorgaande
ronde de MaxT-upfloatlimiet heeft overschreden.

---

## Kandidaat-scoring

Elke kandidaatindeling voor een scoregroep ontvangt een
`DubovCandidateScore` met:

```text
DubovCandidateScore {
    C4Violations    int  // upfloater count
    C5Violations    int  // upfloater score sum
    C6Violations    int  // color preference violations
    C7Violations    int  // MaxT upfloater violations
    C8Violations    int  // consecutive upfloaters
    C9Violations    int  // MaxT opponent violations
    C10Violations   int  // consecutive MaxT violations
    Transposition   int  // transposition index (lower = better)
}
```

De `Compare`-methode retourneert $-1$, $0$ of $+1$ door velden in volgorde
van C4 tot C10 te vergelijken. Als alle schendingsaantallen gelijk zijn,
doorbreekt de transpositie-index de gelijkstand (lager is beter,
overeenkomend met de meer "natuurlijke" indelingsvolgorde).

Een score is **perfect** wanneer alle schendingsaantallen nul zijn en de
transpositie-index nul is. Een perfecte score betekent dat de natuurlijke
indelingsvolgorde alle optimalisatiecriteria voldoet.

---

## Verwerkingsvolgorde: oplopende ARO

Anders dan het Nederlandse systeem, dat scoregroepen van hoog naar laag
verwerkt, verwerkt het Dubov-systeem spelers binnen elke scoregroep in
**oplopende ARO**-volgorde (Average Rating of Opponents).

De ARO wordt berekend uit de daadwerkelijke partijgeschiedenis van de
speler:

$$\text{ARO}(p) = \frac{1}{|G(p)|} \sum_{g \in G(p)} \text{rating}(\text{opponent}(g))$$

waarbij $G(p)$ de verzameling gespeelde partijen is van $p$ (forfaits
uitgezonderd).

Oplopende ARO-verwerking betekent dat spelers die tot dusver zwakkere
tegenstanders hebben gehad, eerst worden ingedeeld. Dit levert doorgaans
indelingen op die de gemiddelde tegenstandersterkte over het toernooi
gelijkmatiger verdelen.

Implementatie: `pairing/dubov/aro.go`.

---

## Matchingalgoritme

Het Dubov-systeem gebruikt **transpositie-gebaseerde matching** binnen elke
scoregroep:

1. **Sorteer** spelers op oplopende ARO.
2. **Splits** in twee helften: G1 (bovenste, hogere ARO) en G2 (onderste,
   lagere ARO).
3. **Genereer transposities** van G2 (permutaties die de indelingspartners
   herschikken).
4. Voor elke transpositie (tot MaxT):
   - Deel G1[i] in met G2[i] voor elke positie $i$.
   - Controleer absolute criteria (C1, C3, verboden).
   - Score de kandidaatindeling (C4--C10).
   - Als perfect, accepteer direct.
   - Anders, registreer als beste kandidaat als deze verbetert ten opzichte
     van de vorige beste (via lexicografische vergelijking).
5. Accepteer de beste kandidaat gevonden binnen MaxT transposities.

Dit verschilt fundamenteel van de Nederlandse aanpak (globale Blossom
matching) en de Lim-aanpak (exchange matching). De transpositelimiet MaxT
beperkt de zoekruimte en ruilt optimaliteit in voor voorspelbaarheid.

Implementatie: `pairing/dubov/matching.go`.

---

## Vergelijking met Nederlandse criteria

| Aspect              | Nederlands ($C_1$--$C_{21}$)        | Dubov ($C_1$--$C_{10}$)                  |
| ------------------- | ----------------------------------- | ---------------------------------------- |
| Aantal criteria     | 21                                  | 10                                       |
| Absolute criteria   | C1, C3 (met topscorer-uitzondering) | C1, C3 (geen uitzondering)               |
| Kleurcriteria       | 4 niveaus (C10--C13)                | 1 niveau (C6)                            |
| Floatcriteria       | 8 (C14--C21, ronde $R-1$ en $R-2$)  | 4 (C7--C10, MaxT-gebaseerd)              |
| Matchingmethode     | Globale Blossom                     | Transpositie met MaxT-limiet             |
| Verwerkingsvolgorde | Aflopende score                     | Oplopende ARO binnen groep               |
| Scoregroepen        | Globaal over alle groepen           | Een groep tegelijk                       |
| Rekenmodel          | Polynomiaal (Blossom $O(n^3)$)      | Begrensde zoektocht (MaxT transposities) |

Het Dubov-systeem is eenvoudiger en voorspelbaarder. De MaxT-limiet zorgt
ervoor dat het gedrag van het algoritme door arbiters begrepen kan worden:
er worden maximaal $\text{MaxT}$ alternatieve indelingen overwogen voordat de
beste wordt geaccepteerd. De Blossom matching van het Nederlandse systeem
overweegt alle mogelijke indelingen gelijktijdig, wat krachtiger is maar
moeilijker handmatig na te volgen.

---

## Gerelateerde pagina's

- [Dubov-indeling](/docs/pairing-systems/dubov/) — het indelingssysteem dat
  door deze criteria wordt bestuurd.
- [Nederlandse criteria](../dutch-criteria/) — het Nederlandse alternatief
  met 21 criteria.
- [Kleurverdeling](../color-allocation/) — hoe Dubov kleurvoorkeuren na
  de indeling oplost.
