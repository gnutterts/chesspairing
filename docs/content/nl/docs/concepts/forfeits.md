---
title: "Forfaits en afwezigheden"
linkTitle: "Forfaits"
weight: 8
description: "Hoe forfaits en afwezigheden van invloed zijn op scoring, indelingshistorie en tiebreaker-berekeningen."
---

Niet elke partij in een schaaktoernooi eindigt doordat er stukken worden gezet. Soms komt een speler niet opdagen, of verschijnen beide spelers niet. Deze situaties leveren **forfait**-resultaten op die heel anders werken dan gewone partijuitslagen -- zowel voor de scoring als voor toekomstige indelingen.

## Forfait-partijresultaten

chesspairing kent drie forfait-resultaten:

| Resultaat              | Code   | Betekenis                                         |
| ---------------------- | ------ | ------------------------------------------------- |
| **Forfait wit wint**   | `1-0f` | Zwart is niet komen opdagen; wit krijgt de winst. |
| **Forfait zwart wint** | `0-1f` | Wit is niet komen opdagen; zwart krijgt de winst. |
| **Dubbel forfait**     | `0-0f` | Geen van beide spelers is komen opdagen.          |

Deze zijn te onderscheiden van de vier gewone partijresultaten (`1-0`, `0-1`, `0.5-0.5` en `*` voor lopend).

## Het cruciale verschil: indelingshistorie

Het belangrijkste om te begrijpen over forfaits is hoe ze de indelingshistorie beïnvloeden.

**Enkel forfait (één speler wint door forfait):** De winnaar ontvangt punten (standaard 1.0 bij standaard scoring), maar de partij wordt **uitgesloten van de indelingshistorie**. Omdat de spelers nooit daadwerkelijk tegenover elkaar aan het bord zaten, behandelt de indelingsengine hen alsof ze elkaar niet hebben ontmoet. Ze kunnen in een latere ronde opnieuw worden ingedeeld.

**Dubbel forfait:** De partij wordt uitgesloten van **zowel de scoring als de indelingshistorie**. Geen van beide spelers ontvangt punten, en de partij wordt behandeld alsof deze nooit heeft plaatsgevonden. De twee spelers kunnen opnieuw worden ingedeeld.

Dit betekent dat forfait-partijen niet meetellen als "gespeeld" voor het rematchverbod (het absolute criterium dat voorkomt dat twee spelers elkaar tweemaal ontmoeten). Een speler die in ronde 3 door forfait won van een tegenstander, kan diezelfde tegenstander in ronde 5 opnieuw treffen.

## Impact op kleurhistorie

Forfait-partijen worden ook uitgesloten van de kleurhistorie. Aangezien er geen partij is gespeeld, krijgt geen van beide spelers een kleurtoewijzing voor die ronde. Dit beïnvloedt:

- **Kleurvoorkeur-berekeningen.** De ronde draagt niet bij aan de wit/zwart-balans of de tracking van opeenvolgende dezelfde kleuren van de speler.
- **Kleurdifferentie.** Het kleuronevenwicht van de speler wordt alleen berekend over ronden waarin daadwerkelijk is gespeeld.

In de kleurhistorie van de speler wordt een forfait-ronde geregistreerd als "geen kleur" (hetzelfde als een bye), zodat het geen invloed heeft op toekomstige kleurtoewijzing.

## Impact op tiebreakers

Tiebreaker-berekeningen sluiten alle forfait-partijen systematisch uit. De `buildOpponentData`-functie die tiebreaker-berekeningen voedt, slaat elke partij met een forfait-resultaat over (enkel of dubbel). Dit betekent:

- **Buchholz** (alle varianten) telt de score van de forfait-tegenstander niet mee.
- **Sonneborn-Berger** neemt het resultaat-maal-tegenstander-score product van de forfait-partij niet op.
- **ARO** (Average Rating of Opponents) middelt alleen over tegenstanders uit daadwerkelijke partijen.
- **Direct Encounter** beschouwt alleen resultaten van partijen die aan het bord zijn gespeeld.
- **Performance Rating** en gerelateerde tiebreakers (PTP, APRO, APPO) sluiten forfait-partijen uit van hun berekeningen.

Alleen daadwerkelijk aan het bord gespeelde partijen -- waar beide spelers aanwezig waren en zetten deden -- dragen bij aan tegenstander-gebaseerde tiebreaker-waarden.

## Afwezigheidstypen

Naast forfaits die plaatsvinden binnen een geplande partij, kunnen spelers ook een volledige ronde missen. chesspairing onderscheidt drie afwezigheidstypen, elk geregistreerd als een [bye](/docs/concepts/byes/):

| Type                 | Beschrijving                                                                                | Standaardpunten |
| -------------------- | ------------------------------------------------------------------------------------------- | --------------- |
| **Afwezig**          | Ongeoorloofde afwezigheid. De speler is niet komen opdagen zonder de arbiter in te lichten. | 0.0             |
| **Verontschuldigd**  | De speler heeft de arbiter vooraf laten weten de ronde te missen.                           | 0.0             |
| **Clubverplichting** | De speler is afwezig omdat hij of zij speelt voor een clubteam in een interclubcompetitie.  | 0.0             |

Elk afwezigheidstype kan via de opties van het scoresysteem een andere puntwaarde krijgen. Een organisator kan er bijvoorbeeld voor kiezen om verontschuldigde afwezigheden een gedeeltelijk punt te geven terwijl ongeoorloofde afwezigheden nul punten opleveren.

## Forfait-status controleren in code

Het `GameResult`-type biedt twee methoden om forfaits te identificeren:

- `IsForfeit()` geeft `true` voor alle drie de forfait-resultaten (`1-0f`, `0-1f`, `0-0f`).
- `IsDoubleForfeit()` geeft `true` alleen voor het dubbel forfait (`0-0f`).

Dit onderscheid is belangrijk omdat enkel forfaits nog steeds punten toekennen aan de winnaar, terwijl dubbel forfaits niets toekennen. De `IsForfeit()`-controle wordt door de hele codebase gebruikt om forfait-partijen uit te sluiten van tegenstander-lijsten, kleurhistorie en tiebreaker-data.

## Zie ook

- [Scoresystemen](/docs/scoring/) -- hoe forfait-winsten en -verliezen worden gescoord
- [Tiebreakers](/docs/tiebreakers/) -- welke tiebreakers worden beïnvloed door forfait-uitsluiting
