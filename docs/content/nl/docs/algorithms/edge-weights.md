---
title: "Kantgewicht-codering"
linkTitle: "Kantgewichten"
weight: 2
description: "Hoe 16+ indelingscriteria worden gecodeerd in een enkel big.Int-kantgewicht voor Blossom matching."
---

## Het probleem

Zwitserse indelingssystemen definiëren een strikte prioriteitsvolgorde over veel
criteria. Het Nederlandse systeem (FIDE C.04.3) heeft bijvoorbeeld 21
optimaliseringscriteria genummerd $C_1$ tot en met $C_{21}$. Criterium $C_i$
heeft absolute voorrang boven $C_j$ wanneer $i < j$ -- geen enkele verbetering
in lagere-prioriteitscriteria kan opwegen tegen een enkele schending van een
hoger-prioriteitscriterium.

Het [Blossom-algoritme](../blossom/) maximaliseert een enkel numeriek gewicht
per kant. We hebben een manier nodig om alle criteria in één getal te coderen
zodat Blossoms gewichtsmaximalisatie de criteria automatisch in de juiste
prioriteitsvolgorde vervult.

---

## De oplossing: bitveldcodering

Elk criterium beslaat een aaneengesloten reeks bits in een `*big.Int`-waarde.
Criteria met hogere prioriteit bezetten meer-significante bits. Omdat de waarde
van een enkele bit op positie $p$ de gecombineerde waarde van alle onderliggende
bits overschrijdt:

$$2^p > \sum_{k=0}^{p-1} 2^k = 2^p - 1$$

vermindert een enkele schending in een hoog-prioriteitscriterium (het wissen
van diens bit) het gewicht met meer dan de maximaal mogelijke bijdrage van alle
lagere-prioriteitscriteria samen. Blossom, dat het totale gewicht maximaliseert,
zal daarom altijd eerst hogere-prioriteitscriteria oplossen.

**Positieve-logicaconventie:** een gezette bit (1) betekent "geen schending."
Een hoger gewicht betekent een betere indeling. Dit komt overeen met de
implementatie in `pairing/swisslib/criteria_pairs.go`.

---

## Bitbreedteparameters

Drie afgeleide waarden bepalen de indeling:

**$\text{sgBits}$** -- Scoregroepgrootte-bits.

$$\text{sgBits} = \lceil \log_2(\max_i |\text{SG}_i|) \rceil$$

waarbij $|\text{SG}_i|$ het aantal spelers in scoregroep $i$ is. Dit is de
breedte van elk booleaans veld (het veld kan een telling bevatten tot de
grootste scoregroepgrootte). Berekend door `bitsToRepresent(maxScoreGroupSize)`.

**$\text{sgsShift}$** -- Scoregroepenverschuiving (cumulatieve bitbreedte).

$$\text{sgsShift} = \sum_{i} \text{bitsToRepresent}(|\text{SG}_i|)$$

Elke scoregroep krijgt een subveld waarvan de breedte afhangt van zijn eigen
grootte. De totale breedte van een score-geïndexeerd veld is
$\text{sgsShift}$. Binnen zo'n veld begint het subveld van scoregroep $i$ op
offset:

$$\text{sgShifts}[\text{score}_i] = \sum_{j < i} \text{bitsToRepresent}(|\text{SG}_j|)$$

waarbij groepen zijn geordend van laagste score eerst (overeenkomend met de
laag-naar-hoog-iteratie van bbpPairings). Dit wordt opgeslagen in
`EdgeWeightParams.ScoreGroupShifts`.

**$\text{reserveBits}$** -- Reserve voor de paar-specifieke optelwaarde.

$$\text{reserveBits} = 3 \cdot \text{sgBits} + 1$$

---

## Bitindeling (hoog naar laag)

De volledige indeling van meest-significante naar minst-significante bits.
Velden staan van boven naar beneden in afnemende significantie. "Breedte" is
in bits.

| #   | Veld                                 | Breedte                     | Beschrijving                                                                                                                                                                                  |
| --- | ------------------------------------ | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Bye-geschiktheid                     | 2                           | Waarde $1 + [\text{not bye candidate}_i] + [\text{not bye candidate}_j]$. Geeft de voorkeur aan het koppelen van niet-bye-geschikte spelers, zodat bye-kandidaten overblijven.                |
| 2   | Paren in huidige bracket             | $\text{sgBits}$             | 1 als beide spelers in de huidige scoregroep zitten, anders 0. Maximaliseert het aantal indelingen binnen de bracket (C5).                                                                      |
| 3   | Scores in huidige bracket            | $\text{sgsShift}$           | Zet een bit op de subveldpositie die overeenkomt met de score van de hogere speler. Maximaliseert de som van gelote scores binnen de bracket (C6).                                            |
| 4   | Paren in volgende bracket            | $\text{sgBits}$             | 1 als de lagere speler in de volgende scoregroep zit. Breidt de bracket naar beneden uit (C7).                                                                                                |
| 5   | Scores in volgende bracket           | $\text{sgsShift}$           | Analoog aan veld 3, maar voor de uitbreiding naar de volgende bracket.                                                                                                                        |
| 6   | Bye-ontvanger ongespeeld (lager)     | $\text{sgBits}$             | Aantal ongespeelde partijen van de lagere speler als die een bye-kandidaat is. C9: minimaliseer ongespeelde partijen van de uiteindelijke bye-ontvanger.                                      |
| 7   | Bye-ontvanger ongespeeld (hoger)     | $\text{sgBits}$             | Idem voor de hogere speler.                                                                                                                                                                   |
| 8   | Kleur: absoluut onevenwicht          | $\text{sgBits}$             | C10: 1 tenzij beide spelers absoluut kleuronevenwicht ($> 1$) hebben en dezelfde kleur prefereren.                                                                                            |
| 9   | Kleur: absolute voorkeur             | $\text{sgBits}$             | C11: complexe controle op onevenwichtsgrootte, herhaald-kleurgeschiedenis en voorkeurrichting.                                                                                                |
| 10  | Kleur: voorkeur compatibel           | $\text{sgBits}$             | C12: 1 als kleurvoorkeuren compatibel zijn (verschillende voorkeurskleuren, of ten minste één zonder voorkeur).                                                                               |
| 11  | Kleur: sterke voorkeur               | $\text{sgBits}$             | C13: 1 tenzij beide een sterke voorkeur voor dezelfde kleur hebben zonder absolute overschrijving.                                                                                            |
| 12  | C14: herhaalde neerwaartse float R-1 | $\text{sgBits}$             | Telling (0--2) van spelers die in de vorige ronde naar beneden floatten. Hoger = minder schendingen.                                                                                          |
| 13  | C15: herhaalde opwaartse float R-1   | $\text{sgBits}$             | 1 tenzij de lagere speler vorige ronde een opwaartse floater was en nu tegen een hoger scorende tegenstander is ingedeeld.                                                                       |
| 14  | C18: neerwaartse-floatscore R-1      | $\text{sgsShift}$           | Score-geïndexeerde optelwaarde voor elke speler die vorige ronde naar beneden floatte. Minimaliseert de score van neerwaartse floaters.                                                       |
| 15  | C19: opwaartse-float teg.score R-1   | $\text{sgsShift}$           | Score-geïndexeerde bit voor de score van de hogere speler, gezet wanneer de lagere speler vorige ronde geen opwaartse floater was. Minimaliseert de tegenstanderscore van opwaartse floaters. |
| 16  | C16: herhaalde neerwaartse float R-2 | $\text{sgBits}$             | Zoals C14, maar voor twee ronden geleden. Voorwaardelijk: alleen aanwezig als gespeelde ronden $> 1$.                                                                                         |
| 17  | C17: herhaalde opwaartse float R-2   | $\text{sgBits}$             | Zoals C15, maar voor twee ronden geleden. Voorwaardelijk.                                                                                                                                     |
| 18  | C20: neerwaartse-floatscore R-2      | $\text{sgsShift}$           | Zoals C18, maar voor twee ronden geleden. Voorwaardelijk.                                                                                                                                     |
| 19  | C21: opwaartse-float teg.score R-2   | $\text{sgsShift}$           | Zoals C19, maar voor twee ronden geleden. Voorwaardelijk.                                                                                                                                     |
| 20  | Reserve                              | $3 \cdot \text{sgBits} + 1$ | Gereserveerd voor de paar-specifieke optelwaarde die wordt ingevuld tijdens Fase 3 van de bracketverwerking (S1/S2-splitvoorkeur en BSN-afstand).                                             |

Velden 12--19 zijn voorwaardelijk op het aantal gespeelde ronden. Velden 12--15
vereisen ten minste 1 gespeelde ronde ($R > 0$); velden 16--19 vereisen ten
minste 2 gespeelde ronden ($R > 1$). Wanneer ze afwezig zijn, worden die bits
simpelweg niet toegekend, wat de totale breedte vermindert.

---

## Formule voor de totale bitbreedte

Laat $R$ het aantal reeds gespeelde ronden aanduiden. De totale breedte $W$ is:

$$W = 2 + 2\,\text{sgBits} + 2\,\text{sgsShift} + 2\,\text{sgBits} + 4\,\text{sgBits}$$

plus voorwaardelijke velden:

$$+ \; [R > 0] \cdot (2\,\text{sgBits} + 2\,\text{sgsShift})$$

$$+ \; [R > 1] \cdot (2\,\text{sgBits} + 2\,\text{sgsShift})$$

plus de reserve:

$$+ \; 3\,\text{sgBits} + 1$$

Termen verzameld voor het gebruikelijke geval $R > 1$:

$$W = 3 + 15\,\text{sgBits} + 6\,\text{sgsShift}$$

---

## Rekenvoorbeeld

Beschouw een Zwitsers toernooi met 100 spelers en 9 ronden, momenteel bezig
met het indelen van ronde 5 ($R = 4$ gespeelde ronden). Stel dat de grootste
scoregroep 30 spelers telt en er 9 verschillende scoregroepen van wisselende
grootte zijn.

- $\text{sgBits} = \lceil \log_2(30) \rceil = 5$
- $\text{sgsShift} = \sum_{i=1}^{9} \text{bitsToRepresent}(|\text{SG}_i|)$

  Als de scoregroepgroottes ongeveer 2, 5, 10, 15, 30, 20, 10, 5, 3 zijn:

  $= 1 + 3 + 4 + 4 + 5 + 5 + 4 + 3 + 2 = 31$

- Totaal: $W = 3 + 15(5) + 6(31) = 3 + 75 + 186 = 264$ bits

Voor grotere toernooien of toernooien met meer gedetailleerde scoregroepen
(bijv. remises die halvepuntsverschillen creëren) groeit $\text{sgsShift}$
verder. In de praktijk kunnen waarden van 294 bits of meer voorkomen.

Dit is waarom `int64` (63 bruikbare bits) onvoldoende is en de
`*big.Int`-variant van het Blossom-algoritme nodig is.

---

## Reservebits (paar-specifieke optelwaarde)

De onderste $3 \cdot \text{sgBits} + 1$ bits zijn gereserveerd voor de
**paar-specifieke optelwaarde**, die wordt ingevuld tijdens Fase 3 van de
bracketverwerkingslus in `PairBracketsGlobal`. Deze optelwaarde codeert
optimalisatie binnen de bracket:

- **S1/S2-splitvoorkeur.** Binnen een scoregroep worden spelers verdeeld in een
  bovenste helft (S1) en een onderste helft (S2). De optelwaarde beloont het
  koppelen van S1-spelers met S2-spelers boven S1-S1- of S2-S2-koppelingen.

- **BSN-afstandsminimalisatie.** Onder S1-S2-koppelingen geeft de optelwaarde
  de voorkeur aan het koppelen van de eerste speler in S1 met de eerste in S2,
  de tweede met de tweede, enzovoort. Dit minimaliseert de
  "board seeding number"-afstand.

De reservebreedte van $3 \cdot \text{sgBits} + 1$ biedt voldoende ruimte voor
deze waarden zonder over te lopen in de C20/C21-velden erboven.

---

## Codering van kleurcriteria

Vier bitvelden (velden 8--11) coderen de kleur-gerelateerde
optimaliseringscriteria. Elk is $\text{sgBits}$ breed en vertegenwoordigt een
booleaanse voorwaarde over de kleurvoorkeurcompatibiliteit van het paar. Van
hoogste naar laagste prioriteit:

1. **Absoluut onevenwicht** (C10). Gezet tenzij beide spelers een
   kleuronevenwicht $> 1$ hebben en dezelfde kleur prefereren. Een schending
   betekent dat de indeling een speler zou dwingen een al extreem onevenwicht
   verder te vergroten.

2. **Absolute voorkeur** (C11). Een genuanceerdere controle. Wanneer beide
   spelers absolute kleurvoorkeuren hebben (door onevenwicht of opeenvolgende
   partijen met dezelfde kleur) voor dezelfde kleur, controleert het algoritme
   of het conflict kan worden opgelost door de onevenwichtsgrootte en
   herhaald-kleurgeschiedenis te inspecteren.

3. **Voorkeur compatibel** (C12). Gezet wanneer de voorkeurskleuren van de
   twee spelers verschillen, of wanneer ten minste één speler geen
   kleurvoorkeur heeft. Dit is de standaard "kunnen we kleuren toekennen die
   beide spelers tevreden stellen?"-controle.

4. **Sterke voorkeur** (C13). Gezet tenzij beide spelers sterke (maar niet
   absolute) voorkeuren voor dezelfde kleur hebben en geen van beiden een
   absolute voorkeur heeft die zou overschrijven. Dit is het zwakste
   kleurcriterium.

De implementatie berekent deze uit de `ColorHistory` van elke speler via de
functie `ComputeColorPreference` in `pairing/swisslib/color.go`.

---

## Waarom big.Int?

Een snelle ondergrens: zelfs een bescheiden toernooi produceert totale
bitbreedtes die `int64` overschrijden:

| Parameter         | Klein (20 spelers) | Middel (60 spelers) | Groot (200 spelers) |
| ----------------- | ------------------ | ------------------- | ------------------- |
| $\text{sgBits}$   | 3                  | 5                   | 7                   |
| $\text{sgsShift}$ | 12                 | 30                  | 55                  |
| $W$ ($R > 1$)     | 114                | 248                 | 424                 |

Het type `int64` biedt slechts 63 bruikbare bits (het tekenbit is niet
beschikbaar omdat kantgewichten niet-negatief moeten zijn). Elk toernooi met
meer dan ongeveer 10 spelers en 2+ ronden zal waarschijnlijk kantgewichten
produceren die 63 bits overschrijden.

De functie `MaxWeightMatchingBig` in `algorithm/blossom/blossom_big.go`
gebruikt overal `*big.Int`-rekenkunde, wat geen precisieverlies garandeert
ongeacht de toernooigrootte. De prestatieoverhead van `*big.Int` ten opzichte
van `int64` is ongeveer 3--5x voor typische toernooigroottes, wat verwaarloosbaar
blijft vergeleken met de $O(n^3)$ algoritmekosten.

---

## Gerelateerde pagina's

- [Blossom Matching](../blossom/) -- het algoritme dat deze kantgewichten
  verwerkt.
- [Nederlandse criteria](../dutch-criteria/) -- de 21 criteria die deze
  codering vertegenwoordigt.
