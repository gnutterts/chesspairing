---
title: "Lim Exchange Matching"
linkTitle: "Lim Exchange"
weight: 12
description: "Exchange matching met scrutiny-volgorde en vier floatertypen in het Lim-systeem."
---

## Overzicht

Het Lim-systeem (FIDE C.04.4.3) deelt elke scoregroep in met behulp van een
**exchange-gebaseerd algoritme** dat spelers in een specifieke volgorde
verwerkt en systematisch alternatieve partners probeert wanneer de
natuurlijke indeling mislukt. Het systeem definieert ook vier **floatertypen**
(A--D) met eigen selectieregels om te bepalen welke spelers tussen
scoregroepen verschuiven.

De implementatie staat in `pairing/lim/`.

---

## Compatibiliteit (Art. 2.1)

Vóór elke indeling of wisseling moeten twee spelers **compatibel** zijn.
Het Lim-systeem definieert compatibiliteit strikter dan de meeste andere
systemen:

Twee spelers $a$ en $b$ zijn compatibel dan en slechts dan als:

1. Ze **nog niet tegen elkaar gespeeld** hebben (forfaits uitgezonderd).
2. Ze **geen verboden paar** zijn.
3. Er ten minste één geldige **kleurtoewijzing** voor het paar bestaat.
   Concreet moet er een kleurtoewijzing $(c_a, c_b)$ bestaan met
   $c_a \neq c_b$ zodat zowel `CanReceiveColor(a, c_a)` als
   `CanReceiveColor(b, c_b)` true retourneert.

Een speler **kan kleur** $c$ **ontvangen** als:

- De speler niet al 2 opeenvolgende partijen met kleur $c$ heeft gespeeld
  (geen 3 opeenvolgende partijen met dezelfde kleur).
- Het ontvangen van $c$ geen kleuronbalans van 3 of meer zou veroorzaken.

$$\text{CanReceiveColor}(p, c) = \begin{cases} \text{false} & \text{if last 2 games were color } c \\ \text{false} & \text{if } |\text{imbalance after } c| \geq 3 \\ \text{true} & \text{otherwise} \end{cases}$$

Implementatie: `pairing/lim/compatibility.go`.

---

## Indeling binnen Scoregroepen: S1/S2-splitsing

Elke scoregroep wordt in twee helften gesplitst:

- **S1** (bovenste helft): de hoger gerangschikte spelers (op TPN).
- **S2** (onderste helft): de lager gerangschikte spelers.

Als de groep een oneven aantal spelers bevat, krijgt S1 de extra speler.

De natuurlijke indeling is S1[1] tegen S2[1], S1[2] tegen S2[2], enzovoort.
Wanneer deze natuurlijke indeling mislukt door compatibiliteitsbeperkingen,
probeert het exchange-algoritme alternatieve partnertoewijzingen.

---

## Het Exchange-algoritme (Art. 4)

Het exchange-algoritme verwerkt S1-spelers in **scrutiny-volgorde** (oplopend
TPN binnen S1). Voor elke S1-speler wordt een reeks kandidaat-partners
gegenereerd en in volgorde geprobeerd.

### Startpunt

```text
function ExchangeMatch(players, pairingDownward, forbidden):
    split players into S1, S2
    result ← tryExchangePairing(S1, S2, forbidden, pairingDownward)
    if result == nil:
        result ← greedyPair(players, forbidden)
    return result
```

### Exchange-indeling

```text
function tryExchangePairing(S1, S2, forbidden, pairingDownward):
    unified ← [S1 | S2]        // S1 first, then S2
    pairs ← []

    for each player p in S1 (scrutiny order):
        if p is already paired:
            continue

        candidates ← generateExchangeOrder(p, unified, pairingDownward)

        for each candidate c in candidates:
            if c is already paired:
                continue
            if not IsCompatible(p, c, forbidden):
                continue
            pairs ← pairs + (p, c)
            break

    // Pair remaining unpaired S2 players among themselves
    pairRemainingS2(pairs)

    return pairs if valid, nil otherwise
```

### Kandidaatgeneratie (Art. 4.2)

Voor een S1-speler op positie $i$ genereert de functie `generateExchangeOrder`
kandidaten in deze prioriteit:

1. **Voorgestelde S2-partner**: S2[$i$] (de "natuurlijke" partner).
2. **Overige S2-spelers**: andere S2-spelers in exchange-volgorde (op afstand
   tot de voorgestelde positie).
3. **S1-partners uit de andere helft**: als er geen S2-partner beschikbaar is,
   probeer te koppelen met een andere S1-speler. Dit gebeurt alleen wanneer
   S1 groter is dan S2 of wanneer alle S2-partners onverenigbaar zijn.

De exchange-volgorde binnen S2 volgt de door FIDE vastgestelde reeks: probeer
eerst de dichtstbijzijnde S2-spelers, daarna geleidelijk verder verwijderde.
De vlag `pairingDownward` bepaalt of de exchange neerwaarts zoekt (normaal)
of opwaarts (wanneer de groep upfloaters verwerkt).

---

## Floatertypen (Art. 3.9)

Wanneer een scoregroep niet volledig ingedeeld kan worden, moeten sommige spelers
doorschuiven naar een aangrenzende groep. Het Lim-systeem classificeert
potentiële floaters in vier typen op basis van hun geschiedenis en
compatibiliteit:

| Type | Al eerder gefloat? | Compatibel met aangrenzende groep? | Prioriteit                      |
| ---- | ------------------ | ---------------------------------- | ------------------------------- |
| D    | Nee                | Ja                                 | Beste (als eerste gekozen)      |
| C    | Nee                | Nee                                |                                 |
| B    | Ja                 | Ja                                 |                                 |
| A    | Ja                 | Nee                                | Slechtste (als laatste gekozen) |

De classificatie houdt rekening met of de speler in een vorige ronde al
is doorgeschoven en of er compatibele tegenstanders in de aangrenzende
scoregroep bestaan.

**Selectievoorkeur:** Type D-floaters hebben de voorkeur omdat ze nog niet
eerder zijn doorgeschoven (minimaliseert herhaald floaten) en compatibele
tegenstanders in de doelgroep hebben (zodat ze daar daadwerkelijk ingedeeld
kunnen worden). Type A-floaters zijn een laatste redmiddel: ze zijn al
eerder doorgeschoven en missen compatibele tegenstanders.

Implementatie: `ClassifyFloater` in `pairing/lim/floater.go`.

---

## Selectie van Downfloaters (Art. 3.2--3.4)

Wanneer een speler uit een scoregroep moet doorschuiven naar beneden,
werkt het selectie-algoritme als volgt:

1. **Classificeer** alle ongepaarde spelers in de scoregroep naar floatertype.
2. **Geef voorkeur aan type D**, dan C, dan B, dan A.
3. **Binnen hetzelfde type** gelden tiebreakers:
   - **Kleuregalisatie**: geef voorkeur aan spelers wier kleurbalans dichter
     bij gelijk ligt, of wier doorschuiving zou helpen bij het egaliseren
     van kleuren in de doelgroep.
   - **Laagste TPN**: bij gelijke kleurbalans, kies de speler met het laagste
     rangnummer (hoogste rang).
4. **Compatibiliteitscontrole**: verifieer dat de geselecteerde speler
   ten minste één compatibele tegenstander in de aangrenzende groep heeft.
   Zo niet, probeer de volgende kandidaat.

### Maxitoernooi-uitzondering

Bij maxitoernooien (grote open evenementen) geldt een extra beperking:
de rating van de downfloater mag niet meer dan 100 punten afwijken van de
hoogst gewaardeerde speler in de doelgroep. Dit voorkomt extreme
ratingverschillen bij het doorschuiven naar beneden. Wanneer ingeschakeld
via de optie `MaxiTournament`, overschrijft deze 100-puntengrens de
normale floaterselectie.

Implementatie: `SelectDownFloater` in `pairing/lim/floater.go`.

---

## Selectie van Upfloaters (Art. 3.2.4)

De selectie van upfloaters spiegelt die van downfloaters, met één belangrijk
verschil: in plaats van het laagste TPN te verkiezen, geven upfloaters
de voorkeur aan het **hoogste TPN** (laagste rang). Dit zorgt ervoor dat de
zwakste speler uit de lagere groep omhoog schuift, wat het competitieve
evenwicht bewaart.

De voorkeur voor floatertype en de kleuregalisatie-tiebreakers zijn hetzelfde.

Implementatie: `SelectUpFloater` in `pairing/lim/floater.go`.

---

## Kleurtoewijzing (Art. 5)

Na het indelen wijst het Lim-systeem kleuren toe met een apart algoritme
met **mediaanbewuste tiebreaking**. Zie [Kleurtoewijzing](../color-allocation/)
voor de volledige vergelijking met andere systemen. Belangrijkste kenmerken:

- **Ronde 1**: oneven TPN krijgt de beginkleur (instelbaar, standaard wit).
- **Art. 5.3**: een speler met 2 opeenvolgende partijen in dezelfde kleur
  _moet_ de tegenovergestelde kleur krijgen.
- **Even rondes**: egaliseer het aantal kleuren (Art. 5.2/5.6).
- **Oneven rondes**: wissel af ten opzichte van de laatst gespeelde kleur
  (Art. 5.5).
- **Geschiedenis-tiebreak** (Art. 5.4): loop achterwaarts door de
  partijgeschiedenis tot het eerste afwijkingspunt. De speler boven de
  mediaan krijgt voorrang.

De mediaan-tiebreak is uniek voor het Lim-systeem. "Boven de mediaan"
betekent dat de rang van de speler in de bovenste helft van de huidige
scoregroep valt, wat de filosofie weerspiegelt dat hoger gerangschikte
spelers een licht kleurvoordeel behoren te krijgen.

---

## Greedy Fallback

Als het exchange-algoritme geen geldige indeling kan vinden (alle compatibele
partners zijn uitgeput), deelt een greedy fallback de spelers in TPN-volgorde in:

```text
function greedyPair(players, forbidden):
    for each unpaired player p (ascending TPN):
        for each unpaired candidate c (ascending TPN, c ≠ p):
            if IsCompatible(p, c, forbidden):
                pair (p, c)
                break
    return pairs
```

De greedy fallback kan ongepaarde spelers overlaten (die floaters worden).
Het garandeert dat het algoritme altijd eindigt met _een_ indeling.

---

## Vergelijking met Andere Matchingmethoden

| Eigenschap        | Lim Exchange                           | Dutch Blossom           | Dubov Transpositie                 | Lexicografische DFS                    |
| ----------------- | -------------------------------------- | ----------------------- | ---------------------------------- | -------------------------------------- |
| Bereik            | Eén scoregroep                         | Globaal                 | Eén scoregroep                     | Eén scoregroep                         |
| Zoekstrategie     | Sequentieel + exchange                 | Gewichtsmaximalisatie   | Begrensde transposities            | Depth-first backtracking               |
| Floaterselectie   | 4-typen classificatie                  | Impliciet in gewichten  | ARO-gebaseerd                      | Overgebleven na DFS                    |
| Kleur in indeling | Onderdeel van compatibiliteit          | Onderdeel van gewicht   | Onderdeel van absolute controle    | Na de indeling                         |
| Garantie          | Alle compatibelen ingedeeld of gefloat | Maximum weight matching | Beste binnen MaxT                  | Lexicografisch eerste                  |
| Complexiteit      | $O(n^2)$ per groep                     | $O(n^3)$ globaal        | $O(n \cdot \text{MaxT})$ per groep | Exponentieel slechtste, snel praktisch |

---

## Gerelateerde Pagina's

- [Lim-indeling](/docs/pairing-systems/lim/) — het indelingssysteem dat
  exchange matching gebruikt.
- [Kleurtoewijzing](../color-allocation/) — het Lim-kleuralgoritme met
  mediaan-tiebreaking.
- [Dutch-criteria](../dutch-criteria/) — het Blossom-gebaseerde alternatief.
- [Lexicografische Indeling](../lexicographic/) — het DFS-gebaseerde alternatief
  dat Dubbel-Zwitsers en Team-Zwitsers gebruiken.
