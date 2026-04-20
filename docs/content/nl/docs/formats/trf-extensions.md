---
title: "TRF-2026-extensies"
linkTitle: "Extensies"
weight: 2
description: "Systeemspecifieke XX-velden en TRF-2026-recordtypes voor configuratie van indelingsengines."
---

Het `trf`-pakket ondersteunt zowel TRF16-legacy-extensies (XX-voorvoegsel) als de nieuwere TRF-2026-recordtypes. Deze pagina documenteert alle extensievelden en datarecords.

## Systeemspecifieke XX-velden (TRF16-legacy)

Deze velden bevatten per-engine-configuratie in het TRF16-formaat. Elk veld is een enkele regel met de XX-code, gevolgd door een spatie en de waarde.

| Code  | Veld                    | Type   | Gebruikt door | Beschrijving                                        |
| ----- | ----------------------- | ------ | ------------- | --------------------------------------------------- |
| `XXY` | Cycles                  | int    | Round-Robin   | Aantal cycli: `1` = enkel, `2` = dubbel round-robin |
| `XXB` | ColorBalance            | bool   | Round-Robin   | Kleurbalancering inschakelen (`true` of `false`)    |
| `XXM` | MaxiTournament          | bool   | Lim           | Maxitoernooimodus inschakelen (`true` of `false`)   |
| `XXT` | ColorPreferenceType     | string | Team          | Kleurvoorkeuralgoritme: `A`, `B` of `none`          |
| `XXG` | PrimaryScore            | string | Team          | Primaire scoremaatstaf: `match` of `game`           |
| `XXA` | AllowRepeatPairings     | bool   | Keizer        | Herhaalde indelingen toestaan (`true` of `false`)     |
| `XXK` | MinRoundsBetweenRepeats | int    | Keizer        | Minimaal aantal ronden tussen herhalingen           |

Voorbeelden:

```text
XXY 2
XXB true
XXM false
XXT A
XXG match
XXA true
XXK 3
```

Deze velden worden tijdens de `ToTournamentState()`-conversie gekoppeld aan indelingsoptiesleutels:

| XX-code | Optiesleutel              |
| ------- | ------------------------- |
| `XXY`   | `cycles`                  |
| `XXB`   | `colorBalance`            |
| `XXM`   | `maxiTournament`          |
| `XXT`   | `colorPreferenceType`     |
| `XXG`   | `primaryScore`            |
| `XXA`   | `allowRepeatPairings`     |
| `XXK`   | `minRoundsBetweenRepeats` |

## Gemeenschappelijke XX-velden

Naast de systeemspecifieke velden definieert TRF16 ook algemene extensiecodes:

| Code  | Veld           | Type     | Beschrijving                                                        |
| ----- | -------------- | -------- | ------------------------------------------------------------------- |
| `XXR` | TotalRounds    | int      | Totaal gepland aantal ronden                                        |
| `XXC` | InitialColor   | string   | Beginkeurtoewijzing (bijv. `white1`)                                |
| `XXS` | Acceleration   | string   | Baku-acceleratiegegevens (een regel per vermelding)                 |
| `XXP` | ForbiddenPairs | int-paar | Twee startnummers die niet ingedeeld mogen worden (een paar per regel) |

`XXR` wordt gekoppeld aan `totalRounds`, `XXC` aan `topSeedColor`, en `XXS` stelt `acceleration` in op `"baku"` in de indelingsopties. Meerdere `XXS`- en `XXP`-regels worden opgeteld.

## TRF-2026-headervelden

TRF-2026 introduceert nieuwe headercodes die TRF16-velden vervangen of uitbreiden:

| Code  | Veld                | Type   | Beschrijving                                                        |
| ----- | ------------------- | ------ | ------------------------------------------------------------------- |
| `142` | TotalRounds26       | int    | Totaal aantal ronden (vervangt `XXR`)                               |
| `152` | InitialColor26      | string | Beginkeurtoewijzing: `B` of `W` (vervangt `XXC`)                    |
| `162` | ScoringSystem       | string | Scorealgoritme (bijv. `W 1.0    D 0.5    L 0.0`)                    |
| `172` | StartingRankMethod  | string | Methode voor toewijzing startnummers (bijv. `IND FIDE`)             |
| `192` | CodedTournamentType | string | Machine-leesbaar toernooitype (bijv. `FIDE_TEAM_BAKU`)              |
| `202` | TieBreakDef         | string | Tiebreaker-configuratie (bijv. `EDET/P,EMGSB/C1/P,BH:MP/C1/P`)      |
| `222` | EncodedTimeControl  | string | Machine-leesbaar speeltempo (bijv. `40/6000+30:20/3000+30:1500+30`) |
| `352` | TeamInitialColor    | string | Teamkleurtoewijzingspatroon (bijv. `WBWB`)                          |
| `362` | TeamScoringSystem   | string | Teamscorealgoritme (bijv. `TW 2     TD 1     TL 0`)                 |

## TRF-2026-datarecords

### 240 -- Afwezigheidsrecords

Declareert afwezige spelers voor een ronde.

Formaat: `240 T RRR TOI1 TOI2 ...`

| Veld  | Beschrijving                                                     |
| ----- | ---------------------------------------------------------------- |
| `T`   | Afwezigheidstype: `F` (volledig forfait) of `H` (halve-punt-bye) |
| `RRR` | Rondenummer                                                      |
| `TOI` | Startnummers van afwezige spelers/teams                          |

Voorbeeld: `240 F 3 5 12 18`

Sectie 240 codeert alleen de twee door FIDE gedefinieerde
afwezigheidsletters. Rijkere byetypes (`zero`, `absent`, `excused`,
`clubcommitment`) en spelersafmeldingen reizen via
chesspairing-commentaardirectieven -- zie hieronder.

### chesspairing-commentaardirectieven

Regels die beginnen met `### chesspairing:` bevatten gegevens die de
FIDE-TRF-formaten niet rechtstreeks kunnen uitdrukken. Ze staan in het
commentaarblok, zodat parsers die ze niet herkennen de regel
ongewijzigd bewaren. Op dit moment zijn twee verba gedefinieerd.

`### chesspairing:bye round=N player=SN type=TYPE` declareert een
vooraf toegewezen bye voor de aankomende ronde. Geldige `type`-waarden
zijn de in kleine letters geschreven `ByeType.String()`-spellingen:

| Waarde           | ByeType             |
| ---------------- | ------------------- |
| `pab`            | `ByePAB`            |
| `half`           | `ByeHalf`           |
| `zero`           | `ByeZero`           |
| `absent`         | `ByeAbsent`         |
| `excused`        | `ByeExcused`        |
| `clubcommitment` | `ByeClubCommitment` |

`### chesspairing:withdrawn player=SN after-round=N` legt een
definitieve afmelding vast: de speler wordt uitgesloten van indeling
voor elke ronde strikt groter dan `N`. `N` moet een positief geheel
getal zijn.

Bij het lezen worden beide verba gekoppeld aan de `TournamentState`:
chesspairing:bye-items worden `PreAssignedByes` voor de huidige ronde,
en chesspairing:withdrawn-items vullen
`PlayerEntry.WithdrawnAfterRound`. Wanneer een Sectie-240-record en
een chesspairing:bye-directief dezelfde `(ronde, speler)` benoemen,
wint het directief; zo kunnen rijkere types de door FIDE gecodeerde
standaard overschrijven. Onbekende speler-ID's in beide verba
veroorzaken een validatiefout in plaats van stilzwijgend genegeerd te
worden, evenals niet-positieve `after-round`-waarden.

Bij het schrijven worden `PreAssignedByes` waarvan het type niet in
Sectie 240 uit te drukken is, geëmitteerd als
chesspairing:bye-directieven, en elke speler met een niet-nil
`WithdrawnAfterRound` levert een chesspairing:withdrawn-directief op.
Onbekende verba die bij het lezen worden aangetroffen, worden
ongewijzigd bewaard, zodat bestanden geschreven door een toekomstige
versie van de bibliotheek niet stilzwijgend door een oudere worden
herschreven.

Voorbeeld:

```text
### chesspairing:bye round=4 player=12 type=excused
### chesspairing:withdrawn player=18 after-round=3
```

Opgeslagen op `Document.ChesspairingDirectives` als een slice van
`Directive{Verb, Params}`.

### 250 -- Acceleratierecords

Baku-acceleratieparameters (vervangt `XXS`).

Formaat: `250 MMMM GGGG RRF RRL PPPF PPPL`

| Veld   | Beschrijving                                    |
| ------ | ----------------------------------------------- |
| `MMMM` | Matchpunten toe te voegen (voor teamtoernooien) |
| `GGGG` | Gamepunten toe te voegen                        |
| `RRF`  | Eerste ronde van acceleratie                    |
| `RRL`  | Laatste ronde van acceleratie                   |
| `PPPF` | Eerste speler-/teamnummer in bereik             |
| `PPPL` | Laatste speler-/teamnummer in bereik            |

Ruwe regelgegevens worden bewaard voor round-trip-getrouwheid.

### 260 -- Verboden-parrecords

Ronde-specifieke verboden indelingsrestricties (vervangt `XXP`).

Formaat: `260 RR1 RRL TOI1 TOI2 ...`

| Veld  | Beschrijving                             |
| ----- | ---------------------------------------- |
| `RR1` | Eerste ronde van de restrictie           |
| `RRL` | Laatste ronde van de restrictie          |
| `TOI` | Startnummers die onderling verboden zijn |

Anders dan `XXP`, dat slechts een enkel paar specificeert, bevat een `260`-record meerdere spelers die allen onderling verboden zijn. De `ToTournamentState()`-conversie genereert alle paarsgewijze combinaties.

### 300 -- Teamrondegegevens

Bordtoewijzingen voor teamwedstrijden.

Formaat: `300 RRR TT1 TT2 PPP1 PPP2 PPP3 PPP4`

| Veld  | Beschrijving                                        |
| ----- | --------------------------------------------------- |
| `RRR` | Rondenummer                                         |
| `TT1` | Eerste teamnummer                                   |
| `TT2` | Tweede teamnummer                                   |
| `PPP` | Startnummers van spelers per bord (`0` = leeg bord) |

### 310 -- Teamdefinitie

TRF-2026-teamrecords (vervangt `013`). Kolomindeling met vaste breedte:

| Bytebereik | Veld              | Breedte | Beschrijving       |
| ---------- | ----------------- | ------- | ------------------ |
| 0-2        | Code              | 3       | Altijd `310`       |
| 4-6        | Teamnummer        | 3       | Rechts uitgelijnd  |
| 8-40       | Teamnaam          | 33      | Links uitgelijnd   |
| 41-45      | Federatie         | 5       | Links uitgelijnd   |
| 46-52      | Gemiddelde rating | 7       | Rechts uitgelijnd  |
| 53-58      | Matchpunten       | 6       | Rechts uitgelijnd  |
| 59-66      | Gamepunten        | 8       | Rechts uitgelijnd  |
| 67-70      | Rang              | 4       | Rechts uitgelijnd  |
| 72+        | Leden             | 4 elk   | Startnummers leden |

### 320 -- Teamrondescores

Scores per ronde per team.

Formaat: `320 TTT GGGG RRR1 RRR2 ...`

| Veld   | Beschrijving           |
| ------ | ---------------------- |
| `TTT`  | Teamnummer             |
| `GGGG` | Totaal gamepunten      |
| `RRR`  | Scorestrings per ronde |

Ruwe regelgegevens worden bewaard voor round-trip-getrouwheid.

### 330 -- Oude afwezigheids-/forfaitrecords

Legacy afwezigheids-/forfaitrecords voor teamtoernooien.

Formaat: `330 TT RRR WWW BBB`

| Veld  | Beschrijving                      |
| ----- | --------------------------------- |
| `TT`  | Resultaatcode: `+-`, `-+` of `--` |
| `RRR` | Rondenummer                       |
| `WWW` | Teamnummer wit                    |
| `BBB` | Teamnummer zwart                  |

### 801 -- Gedetailleerde teamresultaten

Gedetailleerde resultaten per bord per teamwedstrijd. Bevat teamnummer, teamnaam, matchpunten, gamepunten en gegevens per ronde met tegenstander, kleur, individuele bordresultaten en bordvolgorde.

Bye-ronden gebruiken markeringen: `FFFF`, `HHHH`, `ZZZZ`, `UUUU`.

Ruwe regelgegevens worden bewaard voor round-trip-getrouwheid vanwege de complexe opmaak met variabele breedte.

### 802 -- Vereenvoudigde teamresultaten

Vereenvoudigde teamronderesultaten. Bevat teamnummer, teamnaam, matchpunten, gamepunten en per ronde gegevens met tegenstander, kleur en gamepunten.

Bye-ronden gebruiken markeringen: `FPB`, `HPB`, `ZPB`, `PAB` gevolgd door gamepunten.

Een afsluitende `f` bij gamepunten geeft aan dat er een forfait bij betrokken was.

Ruwe regelgegevens worden bewaard voor round-trip-getrouwheid.

### NRS -- Nationale ratingrecords

Regels die beginnen met een drieletterige federatiecode in hoofdletters (zonder `XX`-voorvoegsel) en de `001`-kolomindeling volgen, worden geparseerd als National Rating System-records. Deze bevatten nationale ratings, subfederatiecodes en nationale ID's naast de standaard spelervelden.

NRS-records worden opgeslagen met hun ruwe regelgegevens en ongewijzigd teruggeschreven.

## Terugvalmethoden

Het `trf.Document`-type biedt terugvalaccessors die eerst TRF-2026-velden controleren en daarna terugvallen op TRF16-legacy-velden:

- `EffectiveTotalRounds()` -- retourneert `TotalRounds26` (code `142`) indien ingesteld, anders `TotalRounds` (code `XXR`). Retourneert `0` als geen van beide is ingesteld.
- `EffectiveInitialColor()` -- retourneert `InitialColor26` (code `152`) indien ingesteld, anders `InitialColor` (code `XXC`). Retourneert `""` als geen van beide is ingesteld.

Deze methoden worden gebruikt door `ToTournamentState()` en hebben de voorkeur boven directe toegang tot de velden wanneer zowel TRF16- als TRF-2026-gegevens aanwezig kunnen zijn.

## Round-trip-getrouwheid

Het `trf`-pakket bewaart alle gegevens tijdens lees-/schrijfcycli:

- Onbekende regelcodes worden opgeslagen als `RawLine`-items in de `Other`-slice en teruggeschreven als `CODE DATA`.
- TRF-2026-records met complexe opmaak (`250`, `260`, `320`, `801`, `802`) slaan de ruwe regelgegevens op en gebruiken die bij serialisatie wanneer beschikbaar.
- NRS-records worden opgeslagen en teruggeschreven met hun originele ruwe regels.
- Commentaarregels (`###`) worden opgeslagen in de `Comments`-slice en bij het schrijven gereproduceerd.

Dit zorgt ervoor dat een lees-schrijfcyclus een identiek document produceert voor alle gegevens die de parser niet structureel wijzigt.
