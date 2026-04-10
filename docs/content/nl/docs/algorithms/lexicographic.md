---
title: "Lexicografische Indeling"
linkTitle: "Lexicografisch"
weight: 13
description: "DFS-backtracking voor de lexicografisch kleinste geldige indeling — gebruikt door Dubbel-Zwitsers en Team-Zwitsers."
---

## Overzicht

De Dubbel-Zwitserse (FIDE C.04.5) en Team-Zwitserse (FIDE C.04.6) indelingssystemen
delen een algoritme voor het indelen binnen scoregroepen: zoek de
**lexicografisch kleinste** geldige indeling met behulp van depth-first search
met backtracking.

"Lexicografisch kleinst" betekent: van alle geldige indelingen, kies de indeling
waarbij het eerste paar (op volgorde van rangnummer) zo klein mogelijk is,
vervolgens het tweede paar zo klein mogelijk gegeven het eerste, enzovoort.
Dit levert een deterministische, reproduceerbare indeling op die de voorkeur
geeft aan het eerst koppelen van laaggenummerde spelers (hoger geplaatst).

De implementatie staat in `pairing/lexswiss/bracket.go`.

---

## Definities

Gegeven $n$ deelnemers in een scoregroep, gesorteerd op oplopend
rangnummer (TPN): $p_1, p_2, \ldots, p_n$.

Een **geldige indeling** is een verzameling paren $\{(p_{a_1}, p_{b_1}), (p_{a_2},
p_{b_2}), \ldots\}$ waarbij:

1. Elke deelnemer in ten hoogste één paar voorkomt.
2. Geen enkel paar de absolute criteria schendt:
   - **C1**: De twee spelers hebben nog niet tegen elkaar gespeeld (forfaits
     uitgezonderd).
   - **Verboden paren**: Het paar staat niet op de verbodenlijst.
   - **Systeemcriteria**: Het paar voldoet aan de systeemspecifieke
     criteriafunctie.
3. Als $n$ oneven is, blijft precies één deelnemer ongekoppeld (deze zakt
   door naar de volgende groep).

Een **lexicografische ordening** op indelingen: indeling $A$ is lexicografisch
kleiner dan indeling $B$ als, op de eerste positie waar ze verschillen, de
deelnemer in $A$ een kleiner TPN heeft.

---

## De Criteriafunctie

Het algoritme accepteert een **criteriafunctie** die systeemspecifieke
kwaliteitseisen codeert bovenop de basis C1/verbodencontroles:

```go
type CriteriaFunc func(pairs []Pair, remaining []Participant) bool
```

Deze functie ontvangt de tot dusver gevormde paren en de overgebleven
niet-gekoppelde deelnemers, en retourneert `true` als de huidige gedeeltelijke
indeling acceptabel is.

| Systeem                  | Gecontroleerde criteria                                                                                                  |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| Dubbel-Zwitsers (C.04.5) | C8: Minimaliseer het aantal upfloaters                                                                                   |
| Team-Zwitsers (C.04.6)   | C8: Minimaliseer upfloaters, C9: Minimaliseer scoreverschil van gekoppelde teams, C10: Minimaliseer score van upfloaters |

De criteriafunctie wordt aangeroepen bij elk knooppunt van de DFS-boom,
waardoor vroegtijdig snoeien mogelijk is van takken die niet aan de
kwaliteitseisen kunnen voldoen.

---

## Algoritme: pairRecursive

De kern is een recursieve DFS die indelingen opbouwt, één paar tegelijk:

```text
function pairRecursive(participants, forbidden, criteriaFn, pairs):
    if no unpaired participants remain:
        return pairs                  // Complete valid pairing found

    first ← smallest-TPN unpaired participant

    for each candidate in remaining participants (ascending TPN):
        if first == candidate:
            continue
        if alreadyPlayed(first, candidate):
            continue                  // C1 violation
        if isForbidden(first, candidate):
            continue

        newPairs ← pairs + (first, candidate)
        remaining ← participants - {first, candidate}

        if criteriaFn(newPairs, remaining) == false:
            continue                  // System criteria violated

        result ← pairRecursive(remaining, forbidden, criteriaFn, newPairs)
        if result != nil:
            return result             // Success — propagate up

    // If n is odd and first is the last unpaired, allow leaving them unpaired
    if only one participant remains:
        return pairs                  // first floats

    return nil                        // Backtrack — no valid partner for first
```

### Belangrijke Eigenschappen

1. **Eerste ongebruikte, kleinste TPN.** Op elk recursieniveau kiest het
   algoritme de ongekoppelde deelnemer met het kleinste TPN. Dit garandeert
   de lexicografische eigenschap: het eerste paar wordt bepaald door de
   best beschikbare partner van de speler met het laagste TPN.

2. **Partnerzoektocht in TPN-volgorde.** Kandidaten worden in oplopende
   TPN-volgorde geprobeerd. De eerste geldige partner die gevonden wordt,
   levert de lexicografisch kleinste indeling voor deze positie.

3. **Backtracking.** Als er geen geldige partner bestaat voor de huidige
   speler, keert het algoritme terug naar het vorige niveau en probeert
   daar de volgende kandidaat. Dit vangt situaties op waarin een lokaal
   geldige keuze verderop in de boom tot een dood spoor leidt.

4. **Vroegtijdige beëindiging.** De criteriafunctie maakt snoeien mogelijk.
   Als een gedeeltelijke indeling al de kwaliteitscriteria schendt, wordt
   de tak verlaten zonder de onderliggende kinderen te verkennen.

---

## Greedy Fallback

Als de DFS geen volledige geldige indeling vindt (alle takken worden gesnoeid
door de criteriafunctie), valt het algoritme terug op een greedy gedeeltelijke
indeling:

```text
function greedyPartialPair(participants, forbidden):
    pairs ← []
    for each unpaired participant p (ascending TPN):
        for each unpaired candidate c (ascending TPN, c ≠ p):
            if not alreadyPlayed(p, c) and not isForbidden(p, c):
                pairs ← pairs + (p, c)
                mark p and c as paired
                break
    return pairs
```

De greedy fallback controleert de criteriafunctie niet — alleen C1 en
verboden paren worden afgedwongen. Het kan voorkomen dat sommige deelnemers
ongekoppeld blijven (als floaters). Deze fallback garandeert dat het
algoritme altijd _een_ indeling produceert, zelfs wanneer de criteriafunctie
te beperkend is.

---

## Complexiteit

### Slechtste Geval

De DFS doorzoekt een zoekboom met diepte $n/2$ (één niveau per paar) met
een vertakkingsfactor van maximaal $n - 1$ op het eerste niveau, $n - 3$
op het tweede, enzovoort:

$$\text{nodes} \leq \prod_{k=0}^{n/2 - 1} (n - 2k - 1) = (n-1)!! \quad \text{(double factorial)}$$

Voor $n = 20$ is dit $19!! = 654{,}729{,}075$ — te groot voor brute force.
De criteriafunctie snoeit echter agressief, en door de eigenschap van
vroegtijdige beëindiging wordt de eerste geldige indeling gevonden zonder
de volledige boom te doorzoeken.

### Prestaties in de Praktijk

In de praktijk stopt de DFS snel, omdat:

- **De meeste paren compatibel zijn.** In een typische scoregroep hebben
  weinig spelers al tegen elkaar gespeeld, dus de eerst geprobeerde
  kandidaat is meestal geldig.
- **Criteriasnoeien.** De criteriafunctie elimineert ongeldige takken
  vroegtijdig.
- **Kleine scoregroepen.** Scoregroepen in Zwitserse toernooien bevatten
  zelden meer dan 20--30 spelers (en zijn vaak veel kleiner), waardoor
  de zoekruimte beheersbaar blijft.

Voor typische toernooigroottes is de DFS binnen microseconden afgerond.

---

## Voorbeeld

Scoregroep met 6 deelnemers: TPN 3, 7, 12, 15, 22, 28.

Eerder gespeelde partijen: 3 heeft tegen 7 gespeeld; 12 heeft tegen 15 gespeeld.

**DFS-uitvoering:**

1. Kies TPN 3 (kleinste). Probeer TPN 7 — al gespeeld (C1). Probeer TPN 12 — geldig.
   Paar (3, 12).
2. Kies TPN 7 (volgende kleinste). Probeer TPN 15 — geldig. Paar (7, 15).
3. Kies TPN 22 (volgende kleinste). Probeer TPN 28 — geldig. Paar (22, 28).
4. Geen ongekoppelde deelnemers meer. Resultaat: {(3, 12), (7, 15), (22, 28)}.

Dit is de lexicografisch kleinste geldige indeling. Als TPN 12 ook
onverenigbaar was geweest met TPN 3, zou de DFS TPN 15 als partner voor 3
proberen, wat een andere indeling zou opleveren.

---

## Vergelijking met Blossom-gebaseerde Systemen

| Eigenschap       | Lexicografische DFS                            | Blossom matching          |
| ---------------- | ---------------------------------------------- | ------------------------- |
| Gebruikt door    | Dubbel-Zwitsers, Team-Zwitsers                 | Dutch, Burstein           |
| Optimaliteit     | Lexicografisch eerste                          | Maximaal gewicht          |
| Criteria         | Per paar gecontroleerd, met backtracking       | Gecodeerd in gewichten    |
| Complexiteit     | Exponentieel slechtste geval, snel in praktijk | $O(n^3)$ gegarandeerd     |
| Scoregroepbereik | Eén groep tegelijk                             | Globaal over alle groepen |

De Blossom-aanpak is theoretisch krachtiger: het vindt de globaal optimale
matching over alle criteria tegelijk. De lexicografische aanpak is
eenvoudiger, deterministisch en goed geschikt voor de Dubbel-Zwitserse
en Team-Zwitserse reglementen waar de criteria beperkter zijn en de
definitie van "eerste geldige indeling" expliciet in de regels staat.

---

## Gerelateerde Pagina's

- [Dubbel-Zwitserse Indeling](/docs/pairing-systems/double-swiss/) — gebruikt
  lexicografische indeling met C8-criteria.
- [Team-Zwitserse Indeling](/docs/pairing-systems/team/) — gebruikt
  lexicografische indeling met C8--C10-criteria.
- [Dutch-criteria](../dutch-criteria/) — het Blossom-gebaseerde alternatief
  met 21 optimalisatiecriteria.
