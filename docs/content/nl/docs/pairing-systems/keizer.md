---
title: "Keizer-systeem"
linkTitle: "Keizer"
weight: 7
description: "Een op rangschikking gebaseerd indelingssysteem populair in clubverband — de hoogst gerangschikte spelers spelen tegen elkaar."
---

Het Keizer-systeem deelt spelers van boven naar beneden in op hun huidige rangschikking: rang 1 tegen rang 2, rang 3 tegen rang 4, enzovoort. In ronde 1 wordt de rangschikking bepaald door rating. Vanaf ronde 2 wordt de rangschikking bepaald door de Keizer-score, berekend door de interne Keizer-scorer. Deze nauwe koppeling tussen indeling en scoring is een bepalend kenmerk van het Keizer-systeem. Het is geen FIDE-systeem maar wordt veel gebruikt in clubschaak in Belgie en Nederland.

## Wanneer gebruiken

Keizer is geschikt wanneer:

- Het toernooi een clubcompetitie is die over vele weken loopt met onregelmatige opkomst (het Keizer-scoresysteem gaat goed om met afwezigheden).
- Je wilt dat de sterkste actieve spelers elke ronde tegen elkaar spelen, wat competitieve topbordpartijen oplevert.
- Herhaalde indelingen acceptabel (of zelfs wenselijk) zijn in langlopende toernooien.
- Het toernooi Keizer-scoring gebruikt, omdat de indeling afhankelijk is van de scorer voor de rangschikking.

Het is niet geschikt voor FIDE-geratingde toernooien die een officieel Zwitsers systeem vereisen, of voor korte toernooien waar spelers verwachten elke ronde andere tegenstanders te treffen.

## Configuratie

### CLI

```bash
chesspairing pair --keizer tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/keizer"

// Met getypeerde opties
p := keizer.New(keizer.Options{
    AllowRepeatPairings:     chesspairing.BoolPtr(true),
    MinRoundsBetweenRepeats: chesspairing.IntPtr(3),
})

// Vanuit een generieke map (JSON-configuratie)
p := keizer.NewFromMap(map[string]any{
    "allowRepeatPairings":     true,
    "minRoundsBetweenRepeats": 5,
})
```

### Opties

| Optie                     | Type                     | Standaard | Beschrijving                                                                                                                                                                  |
| ------------------------- | ------------------------ | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `allowRepeatPairings`     | `bool`                   | `true`    | Of spelers opnieuw tegen dezelfde tegenstander ingedeeld mogen worden. Bij `false` zijn herhaalde indelingen nooit toegestaan.                                                |
| `minRoundsBetweenRepeats` | `int`                    | `3`       | Minimaal aantal rondes dat moet verstrijken voordat twee spelers opnieuw tegen elkaar ingedeeld kunnen worden. Alleen van toepassing wanneer `allowRepeatPairings` `true` is. |
| `scoringOptions`          | `scoring/keizer.Options` | nil       | Configuratie voor de interne Keizer-scorer die voor rangschikking wordt gebruikt. Bij nil gebruikt de scorer zijn eigen standaardwaarden (24 instelbare parameters).          |

Het `scoringOptions`-veld accepteert de volledige set Keizer-scoringsparameters. Elke Keizer-scoringsoptie kan ook op het hoogste niveau van de optiemap worden ingesteld -- `ParseOptions` stuurt onbekende sleutels door naar de scoringsparser.

## Hoe het werkt

### 1. Spelers rangschikken

In ronde 1 worden spelers gerangschikt op rating aflopend (alfabetische naam als tiebreaker). Vanaf ronde 2 instantieert de engine een interne Keizer-scorer, voert `Score()` uit op de huidige toernooi-status en rangschikt spelers op hun Keizer-score aflopend. Rating is de secundaire tiebreaker en weergavenaam de tertiaire tiebreaker.

Als scoring om welke reden dan ook mislukt, valt de engine terug op ratinggebaseerde rangschikking.

### 2. Indelingsgeschiedenis opbouwen

De engine scant alle afgeronde rondes om een kaart op te bouwen van in welke ronde elk paar spelers voor het laatst tegen elkaar speelde. Forfait-partijen worden uitgesloten van deze geschiedenis -- conform de projectconventie telt een forfait-partij niet als een echte ontmoeting, dus de twee spelers kunnen opnieuw tegen elkaar ingedeeld worden.

### 3. Van boven naar beneden indelen

Spelers worden sequentieel ingedeeld vanaf de top van de rangschikking:

- Rang 1 vs rang 2
- Rang 3 vs rang 4
- Rang 5 vs rang 6
- ...enzovoort

Als het aantal spelers oneven is, krijgt de laagst gerangschikte speler een indelings-toegekende bye.

### 4. Rematchvermijding

Wanneer een voorgestelde indeling de herhalingsregels zou schenden, wisselt de engine de lager gerangschikte speler in het paar met de dichtstbijzijnde beschikbare lager gerangschikte speler die een legale tegenstander is:

- Als `allowRepeatPairings` `false` is, vormt elke eerdere ontmoeting een conflict.
- Als `allowRepeatPairings` `true` is, wordt de indeling alleen geblokkeerd als de laatste ontmoeting minder dan `minRoundsBetweenRepeats` rondes geleden was.

De wisselzoektocht gaat naar beneden door de rangschikking totdat een compatibele partner is gevonden. Als geen enkele wissel mogelijk is, blijft de herhaalde indeling staan en wordt een notitie aan het resultaat toegevoegd.

### 5. Kleurtoewijzing

De hoger gerangschikte speler in elk paar krijgt wit, tenzij deze in de meest recente partij wit had, in welk geval de kleuren worden gewisseld. Deze eenvoudige afwisselingsregel houdt de kleurbalans redelijk zonder de uitgebreide meerstapsprocedures van FIDE Zwitserse systemen.

Forfait-partijen dragen niet bij aan de kleurgeschiedenis.

## Vergelijking

| Aspect                  | Keizer                         | Dutch Zwitsers           | Round-Robin                     |
| ----------------------- | ------------------------------ | ------------------------ | ------------------------------- |
| Indelingsmethode        | Top-down op score              | Globale Blossom matching | Berger-tabelrotatie             |
| Herhaalde indelingen    | Toegestaan (instelbaar)        | Nooit                    | Elk paar speelt precies eenmaal |
| Scoringsafhankelijkheid | Nauw gekoppeld (Keizer-scorer) | Onafhankelijk            | Onafhankelijk                   |
| Kleurverdeling          | Eenvoudige afwisseling         | 5+ staps FIDE-regels     | Berger-tabelconventie           |
| FIDE-gereguleerd        | Nee                            | Ja (C.04.3)              | Ja (C.05 Annex 1)               |
| Typisch gebruik         | Clubverband, lange toernooien  | Open toernooien          | Kleine gesloten toernooien      |
| Bye-toewijzing          | Laagst gerangschikte speler    | Completability-gebaseerd | Berger-tabel dummy              |

Het Keizer-systeem verschilt fundamenteel van Zwitserse systemen. Zwitserse systemen proberen spelers met vergelijkbare scores te indelen en herhaalde ontmoetingen te vermijden. Keizer probeert de hoogst gerangschikte spelers elke ronde tegen elkaar te indelen, wat bewust topzware paren creëert. De rangschikking evolueert elke ronde naarmate Keizer-scores veranderen, dus een speler die verliest zakt in de rangschikking en treft de volgende ronde zwakkere tegenstanders.

## Wiskundige grondslagen

### Rangschikkingsfunctie

De rangschikking voor ronde r is gedefinieerd als:

```text
rank(p, 1) = sort by rating descending
rank(p, r) = sort by keizerScore(p, r-1) descending    for r > 1
```

waarbij `keizerScore` wordt berekend door de Keizer-scorer met iteratieve convergentie (zie de [Keizer-scoringsdocumentatie](/docs/scoring/keizer/) voor details over het scoringsalgoritme).

### Rematchvermijdingsbeperking

Voor twee spelers a en b met laatste ontmoeting in ronde L, is de indeling toegestaan in ronde R als:

```text
allowRepeatPairings = false:  (a, b) never played
allowRepeatPairings = true:   R - L >= minRoundsBetweenRepeats
```

### Wisselafstand

Wanneer een conflict optreedt op positie i (rang i vs rang i+1), zoekt de engine naar de kleinste j > i+1 zodanig dat `canPair(ranked[i], ranked[j])` waar is. De wissel vervangt ranked[i+1] door ranked[j] in de indeling, waarbij de relatieve volgorde van alle andere spelers behouden blijft. Deze greedy-aanpak minimaliseert de verstoring van het rangschikking-gebaseerde indelingsideaal.

## FIDE-referentie

Het Keizer-systeem valt niet onder FIDE-reglementen. Het is ontstaan in Nederland en wordt voornamelijk gebruikt in clubcompetities in Belgie en Nederland. Er is geen FIDE-handboek-artikel voor Keizer-indeling of -scoring.

Het systeem wordt in Nederlandstalige schaak-literatuur soms "Keizer-Sonneborn" genoemd, hoewel het niet verward moet worden met de Sonneborn-Berger tiebreaker. De scoringsmethode is beschreven in de [Keizer-scoringsdocumentatie](/docs/scoring/keizer/).
