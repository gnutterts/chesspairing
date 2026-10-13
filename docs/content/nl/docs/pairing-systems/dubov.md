---
title: "Dubov-systeem"
linkTitle: "Dubov"
weight: 3
description: "FIDE C.04.4.1 Dubov-paringssysteem."
---

## Overzicht

Het Dubov-systeem implementeert FIDE C.04.4.1 (goedgekeurd door de Raad op
28 oktober 2025, van kracht vanaf 1 februari 2026). Het doel (artikel 0) is dat
een speler met meer punten dan een andere ook een hogere prestatierating heeft.
Dat wordt benaderd door de gemiddelde rating van de tegenstanders (ARO,
artikel 1.7) van alle spelers in een scoregroep gelijk te trekken: spelers met
een lage ARO worden gepaard tegen tegenstanders met een hoge rating.

Het pakket is `pairing/dubov`. Een ronde wordt in deze volgorde ingedeeld
(artikel 1.9.2):

1. Spelers die genoemd zijn in `state.PreAssignedByes` worden uit het veld
   gehaald en ongewijzigd teruggegeven in `PairingResult.Byes`.
2. Bij een oneven aantal resterende spelers wordt eerst de paringsbye (PAB)
   toegewezen (artikel 3.1).
3. De scoregroepen worden in aflopende scorevolgorde ingedeeld (artikel 3.2),
   telkens als een bracket van vaste spelers plus de opfloaters die uit lagere
   scoregroepen omhoog worden gehaald.
4. De paren worden in bordvolgorde gezet en krijgen een kleur (artikel 5.2).

Als de ronde niet voltooid kan worden, wordt helemaal geen paring teruggegeven;
zie [Wanneer er geen paring bestaat](#wanneer-er-geen-paring-bestaat).

## Configuratie

### CLI

```bash
chesspairing pair --dubov tournament.trf
```

De TRF-lezer kopieert het totale aantal ronden (record `142` van TRF-2026,
anders `XXR`), de beginkleur en de verboden paren naar de opties hieronder.

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/dubov"

// Met getypeerde opties
p := dubov.New(dubov.Options{
    TotalRounds:    chesspairing.IntPtr(9),
    TopSeedColor:   chesspairing.StringPtr("auto"),
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})

// Uit een gewone map (bijvoorbeeld uit JSON gelezen)
p := dubov.NewFromMap(map[string]any{
    "totalRounds":  9,
    "topSeedColor": "black",
})
result, err := p.Pair(ctx, &state)
```

### Opties

| Optie (JSON-key) | Go-veld | Type | Standaard | Betekenis |
| --- | --- | --- | --- | --- |
| `totalRounds` | `TotalRounds` | `*int` / geheel getal | geen; **verplicht** | Gepland aantal ronden `Rnds`. Bepaalt `MaxT = 2 + floor(Rnds/5)` (artikel 1.8.2) en of een ronde de laatste is (`CurrentRound == totalRounds`; C8–C10 worden dan niet toegepast). |
| `topSeedColor` | `TopSeedColor` | `*string` / string | `"auto"` | De beginkleur van artikel 5.1, gebruikt door artikel 5.2.1 voor elk paar waarin beide spelers nog niet speelden. `"auto"` en `"white"` beteken allebei wit; `"black"` betekent zwart; elke andere waarde wordt als wit behandeld. |
| `forbiddenPairs` | `ForbiddenPairs` | `[][]string` | geen | Paren van speler-ID's die niet tegen elkaar mogen worden ingedeeld. Ze tellen als een absoluut criterium naast C1 en C3: ze gelden in elk bracket, in de C4-completeerbaarheidstest en bij de byekeuze. Vermeldingen die niet precies twee ID's hebben, worden genegeerd. |

Details:

- **`totalRounds` is verplicht.** Wanneer hij ontbreekt of kleiner is dan 1,
  geeft `Pair` de gewone fout `dubov: total rounds required` terug voordat er
  iets wordt ingedeeld. In een map wordt hij gelezen als een `int`, een `int64`
  of een geheel-getal-`float64` (het type dat JSON-decodering oplevert); een
  breuk of een ander type telt als ontbrekend. De code leidt hem niet af uit het
  aantal tot nu toe gespeelde ronden, omdat `MaxT` en de laatste-ronde-
  versoepeling zijn gedefinieerd op de geplande lengte van het toernooi.
- **`forbiddenPairs` uit een map** moet een JSON-achtige `[]any` zijn waarvan de
  elementen `[]any`-paren van twee strings zijn; een Go `[][]string` die in de
  map is opgeslagen, wordt niet herkend. Bij de getypeerde `Options` is het een
  gewone `[][]string`.
- `WithDefaults` vult alleen `topSeedColor` in; de andere opties hebben geen
  standaard.

## Termen die de procedure gebruikt

| Term | Zoals geïmplementeerd |
| --- | --- |
| Score | De toernooiscore bij standaard- of voetbaltelling, inclusief ingestelde waarden voor partijen, forfaits, byes en afwezigheid. Afgebroken partijen tellen voor de indeling als remise. Andere puntentellingen blijven op 1 / ½ / 0. Een PAB levert standaard evenveel op als een winstpartij. |
| TPN (1.2) | Het vaste toernooiparingsnummer: `PlayerEntry.PairingNumber` wanneer die is ingesteld, anders toegewezen op rating, titel en naam (C.04.2 artikel 2). Elke TPN-sortering hieronder gebruikt dit vaste nummer, niet de rang in de huidige ronde. De code berekent TPN's niet opnieuw (1.2.2). |
| Rating (1.1) | `PlayerEntry.Rating`, gebruikt zoals gegeven. De code verzint geen voorlopige rating (1.1.2); een speler zonder rating heeft rating 0. |
| Gespeelde partijen | Ronden met een kleur. Een bye of een forfait is geen partij: geen kleur, geen tegenstander, niet meegeteld in de ARO, en de twee spelers van een forfaitpartij mogen elkaar opnieuw ontmoeten. |
| ARO (1.7) | Gemiddelde van de ratings van de aan het bord ontmoette tegenstanders, afgerond op het dichtstbijzijnde gehele getal waarbij .5 naar boven wordt afgerond (`ComputeARO`, geheel-getalrekening `(2·som + n) / (2·n)`); 0 voor een speler die nog niet speelde (1.7.3). Ook tegenstanders die later afhaakten, tellen mee. Voorbeeld: tegenstanders met rating 1000 en 1001 geven 1000,5, dus ARO 1001. |
| MaxT (1.8.2) | `MaxT(Rnds) = 2 + floor(Rnds/5)`: 2 voor 1–4 ronden, 3 voor 5–9, 4 voor 10–14, 5 voor 15–19. |
| Kleurvoorkeur (1.6) | Geclassificeerd door de gedeelde `swisslib.ComputeColorPreference` over alleen gespeelde partijen: absoluut (kleurverschil groter dan +1 of kleiner dan −1, of dezelfde kleur in de laatste twee gespeelde ronden), sterk (verschil ±1), licht (verschil 0: wisselen). Dubov voegt artikel 1.6.4 zelf toe: een speler zonder partij heeft een lichte voorkeur voor zwart. |
| Opfloat | Een eerdere ronde telt als een opfloat wanneer de score van de speler aan het begin van die ronde lager was dan die van de tegenstander. `UpfloatCount` telt die ronden; het laatste element van de float-historie geeft aan of de speler in de vorige ronde opfloatte. |

## Procedure

### 3.1 Paringsbye

Bij een even aantal spelers is er geen bye. Bij een oneven aantal zijn de
kandidaten de spelers die er nog een mogen ontvangen (3.1.1, criterium C2):
spelers die nog geen PAB hebben gekregen en niet zonder te spelen een vol punt
scoorden (een bye met volle punten of een forfaitwinst). Ze worden gesorteerd op

1. laagste score (3.1.3),
2. meeste gespeelde partijen (3.1.4),
3. grootste TPN (3.1.5).

Voor elke kandidaat, in die volgorde, wordt de kandidaat verwijderd en wordt de
volledige bracket-procedure van 3.2 op de andere spelers uitgevoerd. De eerste
kandidaat voor wie die procedure een volledige paring oplevert, krijgt de bye
(3.1.2, C4); de gevonden paring voor de anderen wordt gebruikt. De bye wordt
vastgelegd als een `ByePAB`-vermelding. Wanneer geen kandidaat voldoet,
mislukt de ronde met `PairingNoPABCandidate`.

### 3.2 Bracket voor een scoregroep

Scoregroepen worden van de hoogste score naar beneden verwerkt. De spelers met
de hoogste resterende score zijn de **vaste spelers** (1.3.2). Alle spelers met
een lagere score vormen de verzameling mogelijke opfloaters.

**3.2.1 Kleinste aantal opfloaters.** De code probeert 0 of 1 opfloater (het
aantal dat vaste spelers plus opfloaters even maakt), daarna 2 of 3, enzovoort in
stappen van 2. Voor elk aantal somt de code de mogelijke sets op (zie 4.2 en
4.1.3 hieronder). Een set is *toelaatbaar* wanneer

- de vaste spelers plus de set ingedeeld kunnen worden via de stappen 3.2.3–3.2.6
  hieronder, dus met inachtneming van C1, C3 en de verboden paren, en
- de spelers die in de lagere scoregroepen achterblijven nog een volledige paring
  toelaten onder C1, C3 en de verboden paren (C4, functie `completable`: een
  maximale koppeling waarin elke speler gekoppeld is).

Het eerste aantal waarvoor een toelaatbare set bestaat, is het minimum (3.2.1,
C5). Een vaste speler wordt nooit omlaag gestuurd: wanneer de vaste spelers niet
ingedeeld kunnen worden, worden er meer opfloaters omhoog gehaald, en als geen enkel aantal
werkt, mislukt de ronde.

**3.2.2 De set kiezen.** Sets met dit kleinste aantal worden gegroepeerd op hun
scoremultiset (4.2.3); de groepen komen aan bod in C6-volgorde. De eerste groep
met een toelaatbare set wordt gebruikt. Binnen die groep kiest de code de set
die C6–C10, in die volgorde, het best vervult (functie `betterC6ToC10`; zie
[Dubov-criteria](../../algorithms/dubov-criteria/)). Een latere set vervangt de
huidige beste alleen wanneer hij strikt beter is, dus bij gelijke kwaliteit
blijft de eerste set in de volgorde van artikel 4.1.3 staan ("de eerste set
opfloaters die … het beste voldoet"). C5 ligt al vast door het kleinste aantal.

Na de keuze verlaten de vaste spelers en de gekozen opfloaters het veld, de
volgende scoregroep wordt met de resterende vaste spelers verwerkt, en zo voort tot iedereen
ingedeeld is.

**3.2.3 G1 en G2.** Het bracket wordt in twee subgroepen gesplitst:

- Wanneer geen enkele speler van het bracket (opfloaters inbegrepen) een partij
  speelde, zoals in ronde 1, bevat G1 de eerste helft van het bracket op TPN en
  G2 de tweede helft. Dit hangt van de spelers af, niet van het rondenummer.
- Anders bevat G1 de spelers met een kleurvoorkeur voor wit en G2 alle overige
  spelers. Een speler zonder partij prefereert zwart (1.6.4) en hoort dus thuis
  in G2.

**3.2.4 Herindeling.** De code volgt eerst de eerste zin van het artikel. Voor
`s = 0, 1, 2, ...` verschuift hij `s` spelers uit de kleinste subgroep (G1 als
de subgroepen even groot zijn), in de verschuivervolgorde van 4.3, om
onvermijdelijke paren binnen een subgroep te voorkomen. Daarna maakt hij de
subgroepen even groot: spelers die nog de kleur van de grootste subgroep zoeken,
worden daaruit naar de kleinste subgroep verschoven, in de 4.3-volgorde. Een
speler die in de eerste stap is verschoven, houdt de kleurvoorkeur van de
subgroep die hij verliet en gaat dus niet terug. De uitzondering is een groep
waarin nog niemand heeft gespeeld (ronde 1): daar heeft iedereen voorkeur voor
zwart, het artikel zegt niet of een verschoven speler terug mag, en de code
staat het toe. G1 wordt tegen G2 ingedeeld, dus de resulterende subgroepen
moeten even groot zijn. Voor elke herindeling wordt de eerste legale
transpositie (3.2.6) gezocht; de **eerste legale herindeling** wordt gebruikt.

**3.2.5 S1.** De spelers van (het mogelijk nieuwe) G1 worden gesorteerd op
oplopende ARO en, bij gelijke ARO, op oplopende TPN. Het resultaat is S1. De ARO
is de afgeronde waarde van 1.7.1.

**3.2.6 T2.** G2 wordt in oplopende TPN-volgorde gezet en zijn transposities
worden in de volgorde van 4.4 geprobeerd. S1 en T2 worden positie voor positie
ingepaard: eerste met eerste, tweede met tweede. De **eerste** transpositie
waarin elk paar legaal is (C1, C3, verboden paren) wordt gebruikt. De code
vergelijkt transposities niet op kwaliteit, en er is geen grens van 120
transposities.

## Sorteercriteria

### 4.1 Sets op volgnummer

Een verzameling van P spelers krijgt de volgnummers #1 tot en met #P. De sets
van K spelers worden geordend op het kleinste verschillende volgnummer, wat
lexicografische volgorde op de gesorteerde nummers is: bij K = 2 komt {#1,#2}
vóór {#1,#3}, en {#1,#P} vóór {#2,#3}. `forEachCombination` somt precies in
deze volgorde op.

### 4.2 Opfloaters

De verzameling is elke speler met een score lager dan die van de vaste spelers. Het
volgnummer volgt op aflopende score en, bij gelijke scores, op oplopende TPN.

Voorbeeld: de verzameling is A (1,5 punt, TPN 7), B (1,0, TPN 4), C (1,0, TPN 9)
en D (0,5, TPN 2). De nummers zijn #1 A, #2 B, #3 C, #4 D. Sets van twee komen in
de volgorde {A,B}, {A,C}, {A,D}, {B,C}, {B,D}, {C,D}.

### 4.3 Verschuivers

Een speler met een kleurvoorkeur voor wit (zwart) is een mogelijke witte
(resp. zwarte) verschuiver (4.3.1). Spelers die wit zoeken worden gesorteerd op
oplopende ARO en, bij gelijke ARO, op oplopende TPN; spelers die zwart zoeken op oplopende TPN
(4.3.2). Vervolgens worden de volgnummers toegekend beginnend bij de speler in
het midden van de resterende lijst; wanneer twee spelers in het midden staan,
gaat degene die hoger in de lijst staat eerst (4.3.3). `shifterOrder` neemt het
element op index `(n−1)/2` van de resterende lijst, verwijdert het, en herhaalt.

Voorbeeld van 4.3.3: de gesorteerde lijst is A, B, C, D, E, F, G. #1 is D (midden
van zeven), #2 is C (hogere van C en E, het midden van de resterende zes), #3 is
E, #4 is B, #5 is F, #6 is A, #7 is G. Sets van twee verschuivers worden dus
geprobeerd als {D,C}, {D,E}, {D,B}, {D,F}, {D,A}, {D,G}, {C,E}, enzovoort.

### 4.4 Transposities van G2

De spelers die na de herindeling in G2 staan, vormen de verzameling en krijgen
volgnummers op oplopende TPN (4.4.1, 4.4.2). De code zoekt de eerste legale
transpositie met een diepte-eerstzoektocht in de lexicografische volgorde van
die permutaties.

Voorbeeld van 4.4.2: met A, B, C in G2 (oplopende TPN) is de volgorde {A,B,C},
{A,C,B}, {B,A,C}, {B,C,A}, {C,A,B}, {C,B,A}. Als S1 (x1, x2, x3) is en x2 al
tegen B speelde, mislukt de eerste transpositie op x2–B, de tweede geeft
x1–A, x2–C, x3–B, die wordt gebruikt wanneer hij legaal is.

Om te voorkomen dat een problematisch bracket eindeloos wordt doorzocht, geldt
één gedeeld budget van 1.000.000 zoekstappen voor de herindelingen en de
diepte-eerstzoektocht naar transposities; één opsomming van sets is bovendien
begrensd op 200.000 kandidaten. Bij overschrijding van een van beide grenzen
geeft `Pair` `ErrBracketTooLarge` terug, zonder een deel van de voorgeschreven
volgorde over te slaan. Een toernooi met 200 spelers en 11 ronden wordt in een
paar seconden ingedeeld.

## Criteria

De absolute criteria C1–C3 en het voltooiingscriterium C4 moeten altijd gelden.
De kwaliteitscriteria C5–C10 beslissen tussen kandidaten, in aflopende
prioriteit:

| Criterium | Artikel | Regel | Gemeten door |
| --- | --- | --- | --- |
| C1 | 2.1.1 | Geen rematch. | `C1NoRematches` |
| C2 | 2.1.2 | Geen tweede PAB, en geen PAB na een ongespeeld resultaat met volle punten. | `filterNoByeReceived` (in `orderedByeCandidates`) |
| C3 | 2.1.3 | Geen twee spelers met dezelfde absolute kleurvoorkeur. | `C3NoAbsoluteColorConflict` |
| C4 | 2.2.1 | Voor iedereen die nog niet ingedeeld is, moet een legale paring bestaan. | `completable`, en de bye-lus van 3.1 |
| C5 | 2.3.1 | Minimaliseer het aantal opfloaters. | `selectBracket` (kleinste aantal eerst), `candidateQuality.upfloaters` |
| C6 | 2.3.2 | Minimaliseer de scoreverschillen, in aflopende volgorde, in paren met opfloaters. | `candidateQuality.differences` |
| C7 | 2.3.3 | Minimaliseer het aantal spelers dat zijn kleurvoorkeur misloopt. | `candidateQuality.colorViolations` |
| C8 | 2.3.4 | Behalve in de laatste ronde: minimaliseer het aantal maximale opfloaters. | `candidateQuality.maximumUpfloaters` |
| C9 | 2.3.5 | Behalve in de laatste ronde: minimaliseer het aantal keren dat een maximale opfloater opfloatte. | `candidateQuality.maximumTimes` |
| C10 | 2.3.6 | Behalve in de laatste ronde: minimaliseer het aantal opfloaters dat in de vorige ronde opfloatte. | `candidateQuality.previousUpfloaters` |

Zie [Dubov-criteria](../../algorithms/dubov-criteria/) voor de definitie,
meting, vergelijkingsvolgorde en uitgewerkte voorbeelden van elk criterium.

## Kleuren

De kleur van elk paar volgt artikel 5.2, in deze volgorde (`AllocateColor` in
`pairing/dubov/color.go`). Dubov heeft een eigen procedure; hij gebruikt niet de
Nederlandse cascade van `swisslib`, alleen diens classificatie van voorkeuren.

1. **5.2.1** Wanneer beide spelers nog geen partij speelden, krijgt de
   hoger gerangschikte speler (meer punten, of bij gelijke punten de kleinste
   TPN) de beginkleur wanneer zijn TPN oneven is, en de tegenovergestelde kleur
   wanneer hij even is. De TPN is het vaste paringsnummer, niet het bordnummer.
   De beginkleur komt uit `topSeedColor` (wit tenzij hij `"black"` is).
2. **5.2.2** Anders, wanneer de voorkeuren voor verschillende kleuren zijn,
   worden beide gehonoreerd. Een speler zonder partij prefereert zwart (1.6.4).
3. **5.2.3** Wanneer beide dezelfde kleur willen, wint de sterkere voorkeur: een
   absolute voorkeur wint van een niet-absolute (bij twee absolute voorkeuren
   wint het grootste kleurverschil), daarna wint een sterke voorkeur van een
   lichte.
4. **5.2.4** Wanneer dat niet beslist, wisselen de kleuren ten opzichte van de
   meest recente keer dat de ene speler wit en de andere zwart had. De gespeelde
   partijen van beiden worden van de meest recente terug vergeleken; bij het
   eerste verschil krijgt elke speler de kleur die de ander toen had.
5. **5.2.5** Tenslotte wordt de voorkeur van de hoger gerangschikte speler (zoals
   in 5.2.1) gehonoreerd.

Voorbeeld van 5.2.1: in ronde 1 met de beginkleur wit krijgt de hoger
gerangschikte speler van elk paar wit wanneer zijn TPN oneven is en zwart wanneer
hij even is: TPN 1 speelt wit, TPN 2 speelt zwart, TPN 3 speelt wit, enzovoort.
Met `topSeedColor` op `"black"` wordt dit omgekeerd.

De kleuren die voor C7 tijdens het indelen van een bracket worden gemeten, zijn
niet deze uiteindelijke kleuren: daar telt de G1-speler als wit en de G2-speler
als zwart. Kleurtoewijzing loopt pas nadat de paren vaststaan en verandert nooit
wie tegen wie speelt.

## Bordvolgorde

Volgens C.04.2 artikel 3.6 worden de paren stabiel gesorteerd op de score
van de hoger gerangschikte speler (aflopend), daarna op de som van beide
scores (aflopend) en vervolgens op diens TPN (oplopend). Binnen een paar
gaat de hoogste score voor; bij gelijke score het laagste TPN.

Borden worden vanaf 1 in die volgorde genummerd. Een PAB en vooraf toegewezen
byes hebben geen bord; ze staan in `Byes`, vooraf toegewezen byes eerst.

## Wanneer er geen paring bestaat

Artikel 1.9.3 laat een onmogelijke ronde aan de hoofdscheidsrechter. De code
repareert die nooit met een noodparing: hij paart geen spelers die elkaar
al troffen, geeft geen tweede PAB, en deelt het veld niet opnieuw in om het te
doen passen. Hij geeft een fout terug en geen resultaat:

| Situatie | Fout |
| --- | --- |
| Even veld, een bracket kan niet ingedeeld worden | `*chesspairing.PairingError` van soort `PairingIncomplete`, `System` `"dubov"`. `Missing` noemt de ID's die ongepaard blijven, gesorteerd. `Partial` bevat de paren die al vaststaan voor de hogere brackets (met borden en kleuren) plus de vooraf toegewezen byes. Voorbeeldmelding: `dubov: pairing is incomplete: missing p1, p2, p3, p4`. |
| Oneven veld, geen speler kan de PAB nemen, of elke speler die het kan laat een veld achter dat 3.2 niet kan indelen | `*chesspairing.PairingError` van soort `PairingNoPABCandidate`, waarin `swisslib.ErrNoPABCandidate` zit; geen `Partial`. |
| Geen actieve speler en geen vooraf toegewezen bye | `ErrTooFewPlayers` |
| `totalRounds` ontbreekt | de gewone fout `dubov: total rounds required` |
| Gedeeld zoekbudget van een bracket of grens voor setopsomming overschreden | `ErrBracketTooLarge` |
| Context geannuleerd | de fout van de context |

Gebruik `errors.As` om bij de `*chesspairing.PairingError` te komen. De
geëxporteerde `ErrNoPairingPossible` is het interne signaal voor een bracket dat
mislukt; de aanroeper ziet hem als `PairingIncomplete`, niet als een
doorgelaten fout.

## Werkvoorbeeld

Acht spelers p1–p8 met rating 2400, 2300, … , 1700 (TPN 1–8), `totalRounds` 9.

**Ronde 1.** Niemand speelde, dus G1 = {p1..p4} en G2 = {p5..p8} (3.2.3). De
eerste transpositie is legaal: p1–p5, p2–p6, p3–p7, p4–p8. Volgens 5.2.1 krijgt
p1 (oneven TPN) de beginkleur wit; p2 (even) krijgt zwart, dus p6 speelt wit; p3
speelt wit; p4 krijgt zwart, dus p8 speelt wit. Borden: 1 p1–p5, 2 p6–p2, 3
p3–p7, 4 p8–p4. Stel dat de witte spelers p1, p6, p3 en p8 winnen.

**Ronde 2, scoregroep 1,0** = {p1, p3, p6, p8}. Allen speelden één keer wit, dus
allen prefereren zwart: G1 is leeg, G2 heeft vier spelers. Twee spelers moeten
verschoven worden (3.2.4). Zwartzoekers op TPN zijn p1, p3, p6, p8; vanuit het
midden genummerd zijn ze #1 p3, #2 p6, #3 p1, #4 p8. De eerste set is {p3, p6},
dus G1 = {p3, p6} en G2 = {p1, p8}. G1 op ARO: p3 (tegenstander p7, 1800) vóór
p6 (tegenstander p2, 2300). De eerste transpositie geeft p3–p1 en p6–p8, die
legaal is. In beide paren willen de twee spelers zwart met gelijke sterkte en
gelijke historie, dus geeft 5.2.5 zwart aan de hoger gerangschikte speler (p1,
p6); p3 en p8 spelen wit.

**Ronde 2, scoregroep 0** = {p2, p4, p5, p7}. Allen speelden zwart, allen
willen wit: G1 heeft vier spelers, G2 geen. De spelers die wit zoeken staan op ARO: p4 (1700),
p2 (1900), p7 (2200), p5 (2400). Vanuit het midden genummerd: #1 p2, #2 p7,
#3 p4, #4 p5. De eerste set {p2, p7} verhuist naar G2; G1 = {p4, p5}, G2 = {p2,
p7}. De eerste transpositie geeft p4–p2 en p5–p7. Volgens 5.2.5 spelen p2 en p5
(de hoger gerangschikten) wit.

Borden: 1 p3–p1, 2 p8–p6, 3 p2–p4, 4 p5–p7.

## Interpretaties

De tekst is niet expliciet over de punten hieronder. De code volgt de woorden zo
dicht mogelijk en doet het volgende; dit zijn interpretaties.

- **C7 tijdens het indelen.** Bij de vergelijking van opfloatersets telt een
  speler als iemand die zijn voorkeur misloopt wanneer de G1-speler (wit) of de
  G2-speler (zwart) de andere kleur prefereert. De uiteindelijke kleuren van 5.2
  kunnen verschillen.
- **3.1.2.** "Laat een volledige paring van alle resterende spelers toe" wordt
  getest door de procedure van 3.2 op de resterende spelers uit te voeren. Een
  kandidaat van wie de resterende spelers in theorie een legale paring hebben,
  maar die met die procedure niet voltooid kan worden, wordt overgeslagen.
- **3.2.1 en 3.2.2.** Het kleinste aantal opfloaters is het kleinste aantal
  waarvoor een toelaatbare set bestaat (inclusief de C4-test op de spelers die
  achterblijven); van de sets van dat aantal beslissen de kwaliteitscriteria.
- **C6.** De verschillen worden per paar dat een opfloater bevat genomen
  (|score van de ene speler − score van de andere|) en als aflopende vectoren
  vergeleken, element voor element. Wanneer de ene vector het begin van de andere
  is, wint de kortste.
- **C8–C10.** Een maximale opfloater is een speler met ten minste `MaxT`
  opfloats. C9 telt de eerdere opfloat-aantallen van de maximale opfloaters in de
  set op. C10 kijkt naar het laatste element van de float-historie.
- **Opfloat-historie.** Een eerdere ronde telt als een opfloat wanneer de speler
  aan het begin van die ronde minder punten had dan de tegenstander, op basis van de ingestelde toernooipunten (en virtuele punten bij versnelde indeling).
- **5.2.4.** "Meest recente keer" wordt gelezen over alleen gespeelde partijen,
  uitgelijnd vanaf de meest recente partij terug. Deze samenvoeging is de
  behandeling van niet-gespeelde ronden in artikel 3.4 van de General
  Handling Rules: een kleurhistorie van BWBuW wordt behandeld als uBWBW en
  BWWuBuW als uuBWWBW.
- **Bordvolgorde.** C.04.4.1 heeft geen afzonderlijke regel; de volgorde hierboven
  volgt C.04.2 3.6.
- **Zoekgrenzen.** De tekst kent geen grens; de code zoekt in de voorgeschreven
  volgorde en geeft `ErrBracketTooLarge` terug wanneer het gedeelde zoekbudget
  van een bracket of de grens voor een setopsomming wordt overschreden.

## FIDE-referentie

FIDE Handbook C.04.4.1 Dubov System, effective from 1 February 2026
([handbook.fide.com/chapter/C04041](https://handbook.fide.com/chapter/C04041)).
Geraadpleegde artikelen: 0 (preface), 1.1–1.9 (definities), 2.1–2.3 (criteria
C1–C10), 3.1–3.2 (bye- en bracket-procedure), 4.1–4.4 (sorteercriteria) en
5.1–5.2 (kleuren).
