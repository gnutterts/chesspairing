---
title: "Round-Robin"
linkTitle: "Round-Robin"
weight: 8
description: "Elke speler ontmoet elke andere speler — FIDE Berger-tabelplanning met optionele dubbele round-robin."
---

Het Round-Robin-systeem plant elke speler om elke andere speler precies één keer te ontmoeten (enkele round-robin) of twee keer met omgekeerde kleuren (dubbele round-robin). Indelingen worden deterministisch gegenereerd uit de FIDE Berger-tabellen met een rotatie-algoritme. Er is geen matching-optimalisatie en geen criteria-evaluatie -- het schema staat volledig vast vóór het toernooi begint op basis van het aantal spelers en de cyclusconfiguratie.

## Wanneer gebruiken

Round-Robin is geschikt wanneer:

- Het toernooi klein genoeg is dat elke speler elke andere speler kan ontmoeten (doorgaans 6-16 spelers).
- Eerlijkheid vereist dat alle spelers elkaar ontmoeten, waardoor de mogelijkheid wordt geëlimineerd dat twee sterke spelers elkaar ontlopen via Zwitserse indeling.
- Het evenement een gesloten kampioenschap, kwalificatietoernooi of competitiewedstrijd is.
- Dubbele round-robin wordt gebruikt voor evenementen waarbij elk paar met beide kleuren moet spelen.

Het is niet geschikt voor grote open toernooien waar het aantal rondes onpraktisch zou zijn (n-1 rondes voor n spelers in een enkele cyclus).

## Configuratie

### CLI

```bash
chesspairing pair --roundrobin tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/roundrobin"

// Met getypeerde opties
p := roundrobin.New(roundrobin.Options{
    Cycles:            chesspairing.IntPtr(2),
    ColorBalance:      chesspairing.BoolPtr(true),
    SwapLastTwoRounds: chesspairing.BoolPtr(true),
})

// Vanuit een generieke map (JSON-configuratie)
p := roundrobin.NewFromMap(map[string]any{
    "cycles":            2,
    "colorBalance":      true,
    "swapLastTwoRounds": true,
})
```

### Opties-overzicht

| Optie               | Type   | Standaard | Beschrijving                                                                                                                                                         |
| ------------------- | ------ | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cycles`            | `int`  | `1`       | Aantal volledige round-robins. `1` = enkel (elk paar speelt eenmaal). `2` = dubbel (elk paar speelt tweemaal met omgekeerde kleuren).                                |
| `colorBalance`      | `bool` | `true`    | Of alle kleuren omgekeerd worden in even cycli (cyclus 2, 4, ...) van een meervoudige-cyclus round-robin.                                                            |
| `swapLastTwoRounds` | `bool` | `true`    | Of de voorlaatste en laatste rondes van cyclus 1 gewisseld worden bij dubbele round-robin. Geldt alleen als `cycles` `2` is en er minstens 2 rondes per cyclus zijn. |

## Hoe het werkt

### 1. Tabelopzet

De engine bepaalt de tabelgrootte n uit het aantal actieve spelers. Als het aantal spelers oneven is, wordt een dummy "BYE"-speler toegevoegd om n even te maken. De speler die in een ronde tegen de dummy wordt ingedeeld, ontvangt een indelings-bye.

Belangrijke waarden:

- **Rondes per cyclus** = n - 1
- **Totaal rondes** = rondes per cyclus vermenigvuldigd met het aantal cycli

### 2. Berger-tabelrotatie

De FIDE Berger-tabel wordt gegenereerd met een vaste-punt-rotatie-algoritme:

1. Fixeer de laatste speler (index n-1) op positie n-1 voor alle rondes. Bij een oneven aantal spelers is deze vaste positie de bye-dummy.
2. Roteer de overige n-1 spelers door posities 0 tot n-2. De rotatie gebruikt een stapgrootte van n/2 - 1 per ronde.
3. Voor ronde r binnen een cyclus is de positie van speler j: `positions[j] = ((j - r * stride) mod m + m) mod m`, waarbij m = n - 1.

### 3. Koppelen vanuit posities

Koppelingen worden gevormd door posities symmetrisch te matchen: positie 0 met positie n-1, positie 1 met positie n-2, enzovoort tot positie n/2 - 1 met positie n/2.

Als een van beide posities in een paar overeenkomt met de bye-dummy (oneven spelersaantal), ontvangt de echte speler een bye in plaats van een partij.

### 4. Kleurverdeling

Kleuren volgen de FIDE Berger-tabelconventies:

- **Bord 1** (de vaste speler tegen de roterende speler op positie 0): In even rondes (0-gebaseerd) krijgt de roterende speler wit. In oneven rondes krijgt de vaste speler wit.
- **Overige borden**: De speler met de lagere positie-index (de "bovenste-rij" speler) krijgt wit.

### 5. Cyclus-kleuromkering

Bij meervoudige-cyclus-toernooien met `colorBalance` ingeschakeld worden alle kleuren omgekeerd in even cycli (0-gebaseerde cyclus-index 1, 3, ...). Dit zorgt ervoor dat in een dubbele round-robin elk paar eenmaal met elke kleurverdeling speelt.

### 6. Laatste-twee-rondes-wisseling

Bij een dubbele round-robin (`cycles` = 2) met `swapLastTwoRounds` ingeschakeld en minstens 2 rondes per cyclus, wisselt de engine de voorlaatste en laatste rondes van cyclus 1. Dit voorkomt drie of meer opeenvolgende partijen met dezelfde kleur op de grens tussen cyclus 1 en cyclus 2.

Zonder deze wisseling zou een speler die wit heeft in de laatste ronde van cyclus 1 ook wit hebben in de eerste ronde van cyclus 2 (aangezien cyclus 2 kleuren omkeert ten opzichte van cyclus 1, maar de eerste ronde van cyclus 2 mapt naar ronde 1 van de Berger-tabel, niet de laatste). De wisseling doorbreekt dit patroon.

## Vergelijking

| Aspect           | Round-Robin                                | Dutch Zwitsers                   | Keizer                 |
| ---------------- | ------------------------------------------ | -------------------------------- | ---------------------- |
| Schema           | Volledig vooraf bepaald                    | Dynamisch per ronde              | Dynamisch per ronde    |
| Herpartijen      | Elk paar speelt precies één keer (of twee) | Nooit                            | Configureerbaar        |
| Benodigde rondes | n - 1 per cyclus                           | Doorgaans log2(n) tot 2\*log2(n) | Onbeperkt              |
| Spelercapaciteit | Klein (6-16 typisch)                       | Groot (elke omvang)              | Middel (clubformaat)   |
| Kleurverdeling   | Berger-tabelconventie                      | FIDE meerstaps                   | Eenvoudige afwisseling |
| Bye-afhandeling  | Dummy-spelerrotatie                        | Completability-gebaseerd         | Laagstgerangschikte    |
| FIDE-gereguleerd | Ja (C.05 Annex 1)                          | Ja (C.04.3)                      | Nee                    |

## Wiskundige grondslagen

### Berger-tabelformule

Voor n spelers (inclusief dummy bij oneven), m = n - 1 roterende spelers, stapgrootte s = n/2 - 1:

```text
position(j, r) = ((j - r * s) mod m + m) mod m    for j in [0, m-1]
position(m, r) = m                                  (fixed player)
```

Koppelingen voor ronde r:

```text
pair(i, r) = (player[position(i, r)], player[position(n-1-i, r)])    for i in [0, n/2 - 1]
```

### Aantal rondes

| Configuratie        | Rondes   |
| ------------------- | -------- |
| Enkel RR, n even    | n - 1    |
| Enkel RR, n oneven  | n        |
| Dubbel RR, n even   | 2(n - 1) |
| Dubbel RR, n oneven | 2n       |

Het oneven-spelers-geval telt 1 op bij het effectieve spelersaantal (de dummy), dus rondes per cyclus = n.

### Kleurbalanseigenschap

Bij een enkele round-robin met n spelers:

- Wanneer n even is: elke speler krijgt (n-2)/2 partijen als wit en (n-2)/2 als zwart, plus één partij waarvan de kleur afhangt van de positie (niet perfect gebalanceerd voor alle spelers).
- De vaste speler wisselt elke ronde van kleur.

Bij een dubbele round-robin met `colorBalance` ingeschakeld: elk paar speelt precies eenmaal met elke kleurverdeling, wat perfecte paarsgewijze kleurbalans oplevert.

### Correctheid laatste-twee-rondes-wisseling

De wisseling wordt alleen toegepast op cyclus 1. Laat R = rondes per cyclus. In cyclus 1 wordt de logische rondeafbeelding:

```text
actual round R-2 -> Berger round R-1
actual round R-1 -> Berger round R-2
```

Dit wijzigt de kleur van de laatste twee rondes in cyclus 1, waardoor het patroon van opeenvolgende dezelfde kleur wordt doorbroken dat anders zou optreden op de grens tussen cyclus 1 en cyclus 2. De wisseling is alleen zinvol voor dubbele round-robin en wordt uitgeschakeld voor enkele round-robin of wanneer rondes per cyclus minder dan 2 is.

## FIDE-referentie

- **Reglement**: FIDE C.05 Annex 1 (Berger-tabellen)
- **Belangrijke regels**: Berger-tabelrotatie voor planning, kleurverdeling op bordpositie, kleuromkering bij dubbele round-robin-cycli
- **Voorverwerking**: Voor toernooien met FIDE-rangnummertoewijzing kunnen de [Varma-tabellen](/docs/algorithms/varma-tables/) (FIDE C.05 Annex 2) worden gebruikt voor federatie-bewuste nummertoewijzing vóór de round-robin-indeling begint
