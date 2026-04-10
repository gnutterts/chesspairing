---
title: "Keizer-convergentie"
linkTitle: "Keizer-convergentie"
weight: 7
description: "Iteratieve scoring met oscillatiedetectie — hoe Keizer-scores in maximaal 20 iteraties convergeren."
---

## De circulaire afhankelijkheid

Keizer-scoring heeft een ongebruikelijke eigenschap: de score van een speler
hangt af van de **waarderingsgetallen** van diens tegenstanders, en
waarderingsgetallen worden afgeleid van de **ranglijst**, die op zijn beurt
weer bepaald wordt door de scores. Dit creëert een circulaire afhankelijkheid:

$$\text{scores} \to \text{ranking} \to \text{value numbers} \to \text{scores}$$

De implementatie lost dit op door middel van **fixed-point iteratie**: begin
met een initiële ranglijst (op rating), bereken scores, stel de ranglijst
opnieuw op, herbereken, en herhaal totdat de ranglijst stabiliseert. De
implementatie staat in `scoring/keizer/keizer.go`.

---

## Waarderingsgetallen

Het **waarderingsgetal** van elke speler wordt afgeleid van diens positie in
de huidige ranglijst. Voor $N$ actieve spelers met rang $1, 2, \ldots, N$:

$$\text{VN}(k) = \text{base} - (k - 1) \times \text{step}$$

waarbij $k$ de 1-geïndexeerde rang is, $\text{base}$ standaard gelijk is aan
$N$ (het aantal actieve spelers), en $\text{step}$ standaard 1 is.
De hoogstgerangschikte speler heeft het hoogste waarderingsgetal ($N$); de
laagstgerangschikte het kleinste ($1$).

De waarderingsgetallen zijn de "munteenheid" van het Keizer-systeem: een
hooggerangschikte tegenstander verslaan levert meer punten op dan een
lagergerangschikte verslaan.

---

## Scoreberekening

Voor elke speler $p$ is de Keizer-score de som van bijdragen van al diens
resultaten:

$$S(p) = \text{selfVN}(p) + \sum_{\text{games}} \text{gameValue}(p, g) + \sum_{\text{non-games}} \text{nonGameValue}(p, b)$$

### Zelfoverwinning

Elke speler ontvangt eenmaal het eigen waarderingsgetal:

$$\text{selfVN}(p) = \text{VN}(\text{rank}(p))$$

Dit is een standaard Keizer-conventie. Het zorgt ervoor dat zelfs een speler
zonder partijen een score heeft die niet nul is en evenredig is met diens
rang.

### Partijwaarden

Voor een partij waarin speler $p$ tegenover tegenstander $o$ met
waarderingsgetal $\text{VN}(o)$ stond:

| Resultaat | Waarde                                |
| --------- | ------------------------------------- |
| Winst     | $\text{VN}(o) \times \text{winFrac}$  |
| Remise    | $\text{VN}(o) \times \text{drawFrac}$ |
| Verlies   | $\text{VN}(o) \times \text{lossFrac}$ |

De standaardwaarden voor de fracties zijn: winst = 1.0, remise = 0.5,
verlies = 0.0. Deze zijn instelbaar via opties.

### Niet-partijwaarden (byes en afwezigheden)

Byes en afwezigheden hebben geen tegenstander. In plaats daarvan wordt het
**eigen** waarderingsgetal van de speler als basis gebruikt:

| Type                  | Waarde                                |
| --------------------- | ------------------------------------- |
| Pairing-allocated bye | $\text{VN}(p) \times \text{pabFrac}$  |
| Full-point bye        | $\text{VN}(p) \times \text{fpbFrac}$  |
| Half-point bye        | $\text{VN}(p) \times \text{hpbFrac}$  |
| Zero-point bye        | $0$                                   |
| Clubverplichting      | $\text{VN}(p) \times \text{clubFrac}$ |
| Afwezigheid           | zie hieronder                         |

Bij afwezigheden gebruikt de eerste afwezigheid
$\text{VN}(p) \times \text{absentFrac}$. Opeenvolgende afwezigheden worden
**afgebouwd** door halvering:

$$\text{absenceValue}(p, k) = \frac{\text{VN}(p) \times \text{absentFrac}}{2^{k-1}}$$

waarbij $k$ het aantal opeenvolgende afwezigheden is. Er is ook een
instelbaar `AbsenceLimit` (standaard: 3) -- afwezigheden boven deze limiet
leveren niets op.

---

## x2 gehele-getallenrekenkunde

Om drijvende-kommadrift over iteraties te voorkomen, gebruikt de
implementatie **verdubbelde gehele-getallenrekenkunde**. Alle
waarderingsgetallen worden opgeslagen als $2 \times \text{VN}$, en alle
fracties worden toegepast via gehele vermenigvuldiging en deling:

$$\text{internal score} = 2 \times S(p)$$

Dit behoudt halve-puntgranulariteit (remises, half-point byes) zonder enige
drijvende-kommarepresentatie. De uiteindelijke geëxporteerde scores worden
gedeeld door 2 om de oorspronkelijke schaal te herstellen.

---

## De iteratielus

```text
function KeizerScore(state):
    ranking ← initial ranking (by rating)
    prevRanking ← nil
    prevPrevRanking ← nil

    for iteration = 1 to 20:
        VN ← computeValueNumbers(ranking)
        scores ← computeScores(state, VN)
        newRanking ← sortByScore(scores)

        if newRanking == ranking:
            return scores                  // Converged

        if newRanking == prevPrevRanking:
            // 2-cycle oscillation detected
            scores ← average(scores, prevScores)
            return scores

        prevPrevRanking ← prevRanking
        prevRanking ← ranking
        prevScores ← scores
        ranking ← newRanking

    return scores                          // Max iterations reached
```

### Convergentiecontrole

Na elke iteratie wordt de nieuwe ranglijst vergeleken met de vorige. Als ze
identiek zijn (dezelfde spelervolgorde), zijn de scores gestabiliseerd en
eindigt de lus.

### Oscillatiedetectie

Een **2-cyclus oscillatie** treedt op wanneer de ranglijst afwisselt tussen
twee toestanden:

$$R_k \to R_{k+1} \to R_k \to R_{k+1} \to \cdots$$

Dit gebeurt wanneer twee spelers met zeer dichtbijgelegen scores steeds van
rangpositie wisselen, waardoor hun waarderingsgetallen net genoeg veranderen
om ze weer terug te wisselen. De implementatie detecteert dit door de huidige
ranglijst te vergelijken met de ranglijst van _twee_ iteraties geleden
(`prevPrevRanking`). Als ze overeenkomen, oscilleert het systeem tussen twee
vaste punten.

De oplossing is **middeling**: de scores van de laatste twee iteraties worden
per element gemiddeld. Dit levert een stabiele tussenliggende score op die de
oscillerende spelers op het midden van hun twee afwisselende posities plaatst.

---

## Convergentieanalyse

### Waarom convergeert het?

De Keizer-scorefunctie $f : \text{rankings} \to \text{rankings}$ beeldt een
ranglijst af op een nieuwe ranglijst via scoreberekening en hersortering. Het
domein is eindig (er zijn $N!$ mogelijke ranglijsten van $N$ spelers). Dus:

1. De reeks $R_0, R_1, R_2, \ldots$ moet uiteindelijk in een cyclus
   terechtkomen (volgens het duiventilprincipe).
2. Als de cycluslengte 1 is, is de ranglijst geconvergeerd naar een vast
   punt.
3. Als de cycluslengte 2 is, slaat de oscillatiedetector aan en lost het
   op door middeling.
4. Langere cycli zijn theoretisch mogelijk maar in de praktijk niet
   waargenomen. De limiet van 20 iteraties garandeert hoe dan ook
   terminatie.

### Praktische convergentiesnelheid

In de praktijk convergeert Keizer-scoring binnen 2--5 iteraties voor
typische clubtoernooien (20--60 spelers, 7--11 ronden). De initiële
ranglijst (op rating) ligt doorgaans dicht bij de uiteindelijke ranglijst
(op Keizer-score), zodat slechts een paar spelers van positie wisselen
voordat het stabiliseert.

Toernooien waarin veel spelers vergelijkbare ratings en scores hebben,
kunnen meer iteraties vereisen, doordat kleine wijzigingen in
waarderingsgetallen door de ranglijst heen kunnen doorwerken. Het slechtste
waargenomen geval bij het testen was ongeveer 12 iteraties.

### De limiet van 20 iteraties

De implementatie hanteert een harde limiet van 20 iteraties. Als
convergentie dan nog niet bereikt is, worden de laatst berekende scores
ongewijzigd teruggegeven. Dit is een veiligheidsmaatregel; in de praktijk
wordt deze voor realistische toernooigegevens nooit bereikt.

---

## Uitgewerkt voorbeeld

Een toernooi met 5 spelers na 3 ronden. Initiële ranglijst op rating:

| Rang | Speler | Rating |
| ---- | ------ | ------ |
| 1    | Alice  | 2100   |
| 2    | Bob    | 2050   |
| 3    | Carol  | 1950   |
| 4    | Dave   | 1900   |
| 5    | Eve    | 1800   |

Met standaardwaarden (base = 5, step = 1) zijn de initiële
waarderingsgetallen: Alice = 5, Bob = 4, Carol = 3, Dave = 2, Eve = 1.

**Iteratie 1.** Bereken scores met deze waarderingsgetallen. Stel:

- Alice versloeg Bob (VN 4) en Carol (VN 3), verloor van Dave (VN 2): partijscore = $4 + 3 + 0 = 7$, zelf = 5, totaal = 12.
- Bob versloeg Dave (VN 2) en Eve (VN 1), verloor van Alice (VN 5): partijscore = $2 + 1 + 0 = 3$, zelf = 4, totaal = 7.
- Carol versloeg Eve (VN 1), remise tegen Dave (VN 2), verloor van Alice (VN 5): partijscore = $1 + 1 + 0 = 2$, zelf = 3, totaal = 5.
- Dave versloeg Alice (VN 5), remise tegen Carol (VN 3), verloor van Bob (VN 4): partijscore = $5 + 1.5 + 0 = 6.5$, zelf = 2, totaal = 8.5.
- Eve verloor van Bob (VN 4) en Carol (VN 3): partijscore = 0, zelf = 1, totaal = 1.

Nieuwe ranglijst: Alice (12), Dave (8.5), Bob (7), Carol (5), Eve (1).

**Iteratie 2.** Dave en Bob zijn van rang gewisseld (2 en 3). Herbereken
waarderingsgetallen: Alice = 5, Dave = 4, Bob = 3, Carol = 2, Eve = 1.

Herbereken de scores met de nieuwe waarderingsgetallen. Als de nieuwe
ranglijst overeenkomt -- Alice, Dave, Bob, Carol, Eve -- convergeert de
iteratie.

---

## Vaste-waarde-overrides

De opties maken het mogelijk om de via formule afgeleide waarderingsgetallen
te overschrijven met vaste waarden voor specifieke resultaattypen.
Bijvoorbeeld: `FixedWinValue` omzeilt het waarderingsgetal van de
tegenstander volledig en kent voor elke winst een constante toe. Indien
ingesteld wordt de partijwaarde:

$$\text{gameValue} = \text{fixedValue}$$

ongeacht de rang van de tegenstander. Dit maakt van Keizer een eenvoudiger
puntengebaseerd systeem, terwijl het iteratieve raamwerk behouden blijft voor
niet-overschreven resultaattypen.

---

## Complexiteit

Elke iteratie kost $O(N \cdot G)$ waarbij $N$ het aantal spelers is en $G$
het maximale aantal partijen per speler. Hersortering kost $O(N \log N)$.
Met maximaal 20 iteraties:

$$O(20 \cdot (N \cdot G + N \log N)) = O(N \cdot G)$$

aangezien $G$ begrensd is door het aantal ronden (maximaal $N - 1$ bij
round-robin, doorgaans 7--11 bij Zwitsers).

---

## Gerelateerde pagina's

- [Keizer-scoring](/docs/scoring/keizer/) -- configuratie en gebruik van het
  Keizer-scoringssysteem.
- [Keizer-indeling](/docs/pairing-systems/keizer/) -- het indelingssysteem dat
  Keizer-scores gebruikt voor ranglijstgebaseerde indeling.
