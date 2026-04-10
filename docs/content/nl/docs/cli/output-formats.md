---
title: "Uitvoerformaten en exitcodes"
linkTitle: "Uitvoer & exits"
weight: 11
description: "De vijf uitvoerformaten (list, wide, board, xml, json) en zes exitcodes."
---

Deze pagina documenteert de vijf uitvoerformaten die het `pair`-subcommando ondersteunt en de exitcodes die de CLI teruggeeft.

## Uitvoerformaten

Het `pair`-subcommando accepteert `--format` met een van vijf waarden: `list`, `wide`, `board`, `xml`, `json`. Het formaat kan ook worden geselecteerd met de `-w`-afkorting (wide) of de `--json`-afkorting.

### list (standaard)

Machineleesbaar formaat, compatibel met bbpPairings en JaVaFo.

- Eerste regel: het aantal bordindelingen (exclusief byes).
- Volgende regels: `startnr_wit startnr_zwart`, een paar per regel.
- Byes worden na alle indelingen toegevoegd als `startnr 0`.

```text
5
5 1
3 12
8 2
6 9
4 11
7 0
```

Dit formaat bevat alleen startnummers -- geen namen, ratings of bordnummers.

### wide

Leesbaar tabelformaat, gerenderd met Go's `tabwriter`.

Kolommen: `Board`, `White`, `Rtg`, `-`, `Black`, `Rtg`.

Formaat voor spelerweergave: `TPN Titel Naam`, waarbij `TPN` het rangnummer (startnummer) is. De titel wordt weggelaten als de speler geen titel heeft. De ratingkolom is leeg als de rating van de speler nil of 0 is.

Byes worden na de laatste bordindeling vermeld, zonder bordnummer.

```text
Board  White              Rtg   -  Black              Rtg
-----  -----              ---      -----              ---
1      5 GM Smith          2600  -  1 IM Jones         2500
2      3 WGM Lee           2400  -  12 FM Petrov       2350
       7 Brown             1800     Bye (PAB)
```

### board

Compact bordweergave met dynamische veldbreedtes.

Formaat: `Board N: W - B`, waarbij veldbreedtes voor bord- en spelernummers worden aangepast op basis van de grootste aanwezige waarden. Byes worden weergegeven als `Bye: N`.

```text
Board  1:  5 -  1
Board  2:  3 - 12
Board  3:  8 -  2
Bye:  7
```

De veldbreedte voor spelernummers wordt bepaald door het grootste startnummer over alle indelingen en byes. De bordnummerbreedte wordt bepaald door het totale aantal borden.

### json

Gestructureerde JSON-uitvoer met 2-spatie-inspringing.

```json
{
  "pairings": [
    {
      "board": 1,
      "white": 5,
      "black": 1
    },
    {
      "board": 2,
      "white": 3,
      "black": 12
    }
  ],
  "byes": [
    {
      "player": 7,
      "type": "PAB"
    }
  ]
}
```

Structuur:

| Veld               | Type   | Beschrijving                                                           |
| ------------------ | ------ | ---------------------------------------------------------------------- |
| `pairings`         | array  | Bordindelingen                                                           |
| `pairings[].board` | int    | 1-geindexeerd bordnummer                                               |
| `pairings[].white` | int    | Startnummer van de witspeler                                           |
| `pairings[].black` | int    | Startnummer van de zwartspeler                                         |
| `byes`             | array  | Bye-toewijzingen (weggelaten indien leeg)                              |
| `byes[].player`    | int    | Startnummer van de speler                                              |
| `byes[].type`      | string | Bye-type: `PAB`, `Half`, `Zero`, `Absent`, `Excused`, `ClubCommitment` |

De `byes`-sleutel gebruikt `omitempty` en is afwezig in de uitvoer wanneer er geen byes zijn.

### xml

Volledig XML-document inclusief `xml.Header` (`<?xml version="1.0" encoding="UTF-8"?>`).

```xml
<?xml version="1.0" encoding="UTF-8"?>
<pairings round="4" boards="3" byes="1">
  <board number="1">
    <white number="5" name="Smith" rating="2600" title="GM"></white>
    <black number="1" name="Jones" rating="2500" title="IM"></black>
  </board>
  <board number="2">
    <white number="3" name="Lee" rating="2400" title="WGM"></white>
    <black number="12" name="Petrov" rating="2350" title="FM"></black>
  </board>
  <bye number="7" name="Brown" type="PAB"></bye>
</pairings>
```

Attributen van het root-element:

| Attribuut | Type | Beschrijving                    |
| --------- | ---- | ------------------------------- |
| `round`   | int  | Rondenummer (huidige ronde + 1) |
| `boards`  | int  | Aantal bordindelingen             |
| `byes`    | int  | Aantal byes                     |

Children van `<board>` (`<white>`, `<black>`) en `<bye>`-elementen delen deze attributen:

| Attribuut | Type   | Beschrijving                                 |
| --------- | ------ | -------------------------------------------- |
| `number`  | int    | Startnummer speler (altijd aanwezig)         |
| `name`    | string | Weergavenaam speler (weggelaten indien leeg) |
| `rating`  | int    | Rating speler (weggelaten indien 0)          |
| `title`   | string | Titel speler (weggelaten indien leeg)        |

`<bye>`-elementen hebben ook een `type`-attribuut met het bye-type als string.

## Exitcodes

De CLI gebruikt zes exitcodes, gedefinieerd als constanten in `exitcodes.go`:

| Code | Constante          | Betekenis                                          | Gebruikt door                                       |
| ---- | ------------------ | -------------------------------------------------- | --------------------------------------------------- |
| 0    | `ExitSuccess`      | Bewerking succesvol afgerond                       | Alle subcommando's                                  |
| 1    | `ExitNoPairing`    | Geen geldige indeling of indelingen komen niet overeen | pair, check, generate                               |
| 2    | `ExitUnexpected`   | Onverwachte fout (JSON-encoding, schrijffout)      | Alle subcommando's                                  |
| 3    | `ExitInvalidInput` | Foute invoer, onbekende vlaggen, misvormd TRF      | Alle subcommando's                                  |
| 4    | `ExitSizeOverflow` | Toernooi te groot voor de implementatie            | Gedefinieerd maar momenteel niet in gebruik         |
| 5    | `ExitFileAccess`   | Bestands-I/O-fout (openen, lezen of schrijven)     | pair, check, generate, validate, standings, convert |

## Scripting-richtlijnen

Exitcodes maken betrouwbare foutafhandeling in shell-scripts mogelijk:

```bash
chesspairing pair --dutch tournament.trf -o pairings.txt
case $? in
  0) echo "Pairings generated" ;;
  1) echo "No valid pairing found" ;;
  3) echo "Invalid input file" ;;
  5) echo "File error" ;;
  *) echo "Unexpected error" ;;
esac
```

Voor JSON-uitvoer, controleer de exitcode voordat je gaat parsen:

```bash
if output=$(chesspairing pair --dutch tournament.trf --format json); then
  echo "$output" | jq '.pairings | length'
else
  echo "Pairing failed with exit code $?" >&2
fi
```

## Zie ook

- [pair](../pair/) -- het primaire indelingssubcommando
- [Legacy-modus](../legacy/) -- achterwaarts compatibele interface (ondersteunt alleen list- en wide-formaat)
