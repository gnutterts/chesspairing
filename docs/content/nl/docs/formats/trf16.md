---
title: "TRF16-formaat"
linkTitle: "TRF16"
weight: 1
description: "Het FIDE Tournament Report File-formaat — regeltypes, spelersrecords en ronderesultaten."
---

## Overzicht

TRF16 is een tekstformaat met vaste breedte dat door de FIDE wordt gebruikt voor gegevensuitwisseling van toernooien. Elke regel begint met een code van 3 tekens die het type aangeeft, gevolgd door een spatie en de gegevens. Regels worden gescheiden door regeleinden. Het formaat is ontworpen voor overdraagbaarheid tussen indelingsengines.

Het `trf`-pakket (`trf/read.go`, `trf/write.go`) biedt een volledige lezer en schrijver die de documentstructuur getrouw bewaart, inclusief onbekende regelcodes voor round-trip-getrouwheid.

## Headerregels

Headerregels bevatten toernooi-metadata. Elke code correspondeert met een enkel veld:

| Code  | Veld                     | Voorbeeld                       |
| ----- | ------------------------ | ------------------------------- |
| `012` | Toernooiaam              | `012 World Championship 2024`   |
| `022` | Stad                     | `022 Budapest`                  |
| `032` | Federatie                | `032 HUN`                       |
| `042` | Startdatum               | `042 2024/11/01`                |
| `052` | Einddatum                | `052 2024/11/30`                |
| `062` | Aantal spelers           | `062 14`                        |
| `072` | Aantal gerankschikten    | `072 14`                        |
| `082` | Aantal teams             | `082 0`                         |
| `092` | Toernooitype             | `092 Swiss Dutch`               |
| `102` | Hoofdarbiter             | `102 IA FirstName LastName`     |
| `112` | Plaatsvervangend arbiter | `112 FA FirstName LastName`     |
| `122` | Speeltempo               | `122 90/40+30+30`               |
| `132` | Rondedata                | `132 2024/11/01 2024/11/02 ...` |

Meerdere `112`-regels worden ondersteund (TRF-2026 staat meerdere plaatsvervangende arbiters toe). De eerste `112`-regel vult het `DeputyArbiter`-veld; alle `112`-regels worden verzameld in `DeputyArbiters`.

Meerdere `132`-regels worden toegevoegd aan de `RoundDates`-slice.

Het `092`-toernooitype wordt door `ToTournamentState()` gebruikt om het indelingssysteem af te leiden. Herkende waarden:

| Toernooitype         | Indelingssysteem |
| -------------------- | -------------- |
| `Swiss Dutch`        | `dutch`        |
| `Swiss Burstein`     | `burstein`     |
| `Swiss Dubov`        | `dubov`        |
| `Swiss Lim`          | `lim`          |
| `Double Swiss`       | `doubleswiss`  |
| `Team Swiss`         | `team`         |
| `Round Robin`        | `roundrobin`   |
| `Double Round Robin` | `roundrobin`   |
| `Keizer`             | `keizer`       |

Niet-herkende types vallen terug op `dutch`.

## Spelersregels (001)

Spelersregels gebruiken een kolomindeling met vaste breedte. Byteposities zijn 0-gebaseerd:

```text
001 SSSS SEX TTT NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN RRRR FED FFFFFFFFFFF BBBBBBBBBB PPPP RRRR  round-results...
```

| Bytebereik | Veld            | Breedte      | Uitlijning | Beschrijving                  |
| ---------- | --------------- | ------------ | ---------- | ----------------------------- |
| 0-2        | Regelcode       | 3            |            | Altijd `001`                  |
| 4-7        | Startnummer     | 4            | Rechts     | Rangnummer in het toernooi |
| 9          | Geslacht        | 1            |            | `m` of `w`                    |
| 10-12      | Titel           | 3            | Rechts     | `GM`, `IM`, `FM`, `WGM`, etc. |
| 14-46      | Naam            | 33           | Links      | `Achternaam, Voornaam`        |
| 48-51      | Rating          | 4            | Rechts     | FIDE-rating                   |
| 53-55      | Federatie       | 3            | Links      | Drielettercode                |
| 57-67      | FIDE-ID         | 11           | Links      | FIDE-spelernummer             |
| 69-78      | Geboortedatum   | 10           | Links      | `JJJJ/MM/DD`                  |
| 80-83      | Punten          | 4            | Rechts     | Totaal punten (bijv. ` 5.5`)  |
| 85-88      | Rang            | 4            | Rechts     | Huidige rang                  |
| 89+        | Ronderesultaten | 10 per ronde |            | Zie hieronder                 |

De minimale regellengte is 84 tekens (tot en met het puntenveld). De parser vereist minimaal 84 bytes.

## Ronderesultaatformaat

Elke ronde beslaat precies 10 tekens binnen de spelersregel, startend bij byte 89:

```text
  OOOO C R
```

| Byte-offset | Veld         | Beschrijving                                                                 |
| ----------- | ------------ | ---------------------------------------------------------------------------- |
| 0-1         | Opvulling    | Twee spaties                                                                 |
| 2-5         | Tegenstander | Startnummer van 4 cijfers, met nullen aangevuld (`0000` = geen tegenstander) |
| 6           | Spatie       | Scheidingsteken                                                              |
| 7           | Kleur        | `w` (wit), `b` (zwart) of `-` (geen kleur / bye)                             |
| 8           | Spatie       | Scheidingsteken                                                              |
| 9           | Resultaat    | Resultaatcodeteken                                                           |

Voorbeeld: `  0012 w 1` betekent tegenstander 12, speelt met wit, winst.

## Resultaatcodes

| Code | Constante             | Betekenis                       |
| ---- | --------------------- | ------------------------------- |
| `1`  | `ResultWin`           | Winst (gespeeld)                |
| `0`  | `ResultLoss`          | Verlies (gespeeld)              |
| `=`  | `ResultDraw`          | Remise                          |
| `+`  | `ResultForfeitWin`    | Winst door forfait              |
| `-`  | `ResultForfeitLoss`   | Verlies door forfait            |
| `H`  | `ResultHalfBye`       | Halve-punt-bye                  |
| `F`  | `ResultFullBye`       | Volle-punt-bye (PAB)            |
| `U`  | `ResultUnpaired`      | Niet ingedeeld (afwezig, 0 punten) |
| `Z`  | `ResultZeroBye`       | Nulpunt-bye                     |
| `*`  | `ResultNotPlayed`     | Nog niet gespeeld               |
| `W`  | `ResultWinByDefault`  | Winst, tegenstander afwezig     |
| `D`  | `ResultDrawByDefault` | Remise standaard                |
| `L`  | `ResultLossByDefault` | Verlies standaard               |

Bye-resultaten (`H`, `F`, `U`, `Z`) hebben tegenstander `0000` en kleur `-`.

Bij conversie naar `TournamentState` worden bye-resultaten omgezet in `ByeEntry`-records:

| TRF-code | ByeType     |
| -------- | ----------- |
| `F`      | `ByePAB`    |
| `H`      | `ByeHalf`   |
| `Z`      | `ByeZero`   |
| `U`      | `ByeAbsent` |

`ByeExcused` en `ByeClubCommitment` hebben geen rondekolomcode in Sectie 240. Zij worden gedragen door chesspairing-commentaardirectieven (zie [TRF-uitbreidingen](/docs/formats/trf-extensions/)) en bij het lezen overgebracht naar `TournamentState.PreAssignedByes`.

## Uitgebreide gegevensregels (XX-velden)

TRF16 definieert diverse extensiecodes voor gegevens die niet in de basisspecificatie vallen:

| Code  | Veld                 | Type     | Voorbeeld    |
| ----- | -------------------- | -------- | ------------ |
| `XXR` | Totaal aantal ronden | int      | `XXR 9`      |
| `XXC` | Beginkeurtoewijzing  | string   | `XXC white1` |
| `XXS` | Acceleratiegegevens  | string   | `XXS baku`   |
| `XXP` | Verboden paren       | int-paar | `XXP 5 12`   |

`XXR` en `XXC` hebben TRF-2026-vervangers (respectievelijk `142` en `152`). Het `trf`-pakket biedt de methoden `EffectiveTotalRounds()` en `EffectiveInitialColor()` die eerst de TRF-2026-velden controleren en terugvallen op de TRF16 XX-velden.

Meerdere `XXS`- en `XXP`-regels worden ondersteund. Elke `XXP`-regel specificeert een paar startnummers dat niet tegen elkaar ingedeeld mag worden.

Systeemspecifieke XX-velden (`XXY`, `XXB`, `XXM`, `XXT`, `XXG`, `XXA`, `XXK`) zijn gedocumenteerd in [TRF-2026-extensies](../trf-extensions/).

## Teamregels (013)

Teamregels definiëren de teamsamenstelling:

```text
013 SSSS NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN MMMM MMMM ...
```

| Bytebereik | Veld       | Breedte | Beschrijving                            |
| ---------- | ---------- | ------- | --------------------------------------- |
| 0-2        | Regelcode  | 3       | Altijd `013`                            |
| 4-7        | Teamnummer | 4       | Rechts uitgelijnd                       |
| 8-39       | Teamnaam   | 32      | Links uitgelijnd                        |
| 40+        | Leden      | 4 elk   | Startnummers, gescheiden door witruimte |

De minimale regellengte is 40 tekens.

## Commentaarregels

Regels die beginnen met `###` zijn commentaar en worden door de parser genegeerd:

```text
### This is a comment
```

Commentaarregels worden opgeslagen in de `Comments`-slice en bij serialisatie teruggeschreven.

## Onbekende regels

Regels met niet-herkende drieletterige codes worden bewaard als `RawLine`-items in de `Other`-slice. Elke `RawLine` slaat code en gegevens apart op. Bij het terugschrijven worden deze regels gereproduceerd als `CODE DATA`, zodat er geen gegevens verloren gaan tijdens lees-/schrijfcycli.

## National Rating System-records (NRS)

Regels die beginnen met een drieletterige code in hoofdletters die geen `XX`-voorvoegsel is en de `001`-kolomindeling volgen (minimaal 68 tekens met een numeriek startnummer op bytes 4-7), worden herkend als National Rating System-records. Deze bevatten nationale ratings en subfederatiegegevens naast de standaard spelervelden. NRS-records worden opgeslagen in `NRSRecords` en teruggeschreven met hun originele ruwe regelgegevens.

## Parsefouten

Parsefouten bevatten het 1-gebaseerde regelnummer en de regelcode voor diagnostische context:

```text
trf: line 42 (001): invalid rating: "XXXX"
```

De parser stopt bij de eerste fout die wordt aangetroffen.
