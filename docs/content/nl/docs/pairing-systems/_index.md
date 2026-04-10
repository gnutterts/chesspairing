---
title: "Indelingssystemen"
linkTitle: "Indelingssystemen"
weight: 30
description: "Acht indelingsalgoritmen — van FIDE Zwitserse varianten tot Round-Robin en Keizer."
---

Chesspairing implementeert acht indelingsengines. Elke engine voldoet aan de `Pairer`-interface, wat betekent dat je elke engine kunt combineren met elk scoresysteem en elke set tiebreakers. De keuze van indelingssysteem hangt af van het toernooiformaat, de geldende reglementen en de doelen van de organisator.

## Zwitserse systemen

Vijf engines implementeren de FIDE-reglementen voor Zwitserse indelingen. Ze delen allemaal hetzelfde basisprincipe -- groepeer spelers op score en deel binnen groepen in -- maar ze verschillen in matchingstrategie, behandeling van floaters, kleurverdeling en optimaliseringscriteria.

| Systeem                       | FIDE-reglement | Matchingstrategie                            | Meest geschikt voor                                                |
| ----------------------------- | -------------- | -------------------------------------------- | ------------------------------------------------------------------ |
| [Dutch](dutch/)               | C.04.3         | Globale Blossom (21 criteria)                | Standaard gewaardeeerde toernooien, elke omvang                    |
| [Burstein](burstein/)         | C.04.4.2       | Globale Blossom + oppositie-index herranking | Evenementen met een seedingfase gevolgd door competitieve indeling |
| [Dubov](dubov/)               | C.04.4.1       | Transpositie-gebaseerd, ARO-geordend         | Evenementen die balans in tegenstanders-sterkte prioriteren        |
| [Lim](lim/)                   | C.04.4.3       | Exchange-gebaseerd, mediaan-eerst            | Evenementen die expliciete floater-controle willen                 |
| [Double-Swiss](double-swiss/) | C.04.5         | Lexicografische bracket-indeling             | Grote evenementen die snellere indelingsberekening nodig hebben    |

Alle vijf Zwitserse engines behandelen bye-toewijzing, kleurbalans, rematch-vermijding en forfait-uitsluiting. Ze verschillen voornamelijk in hoe ze conflicten oplossen wanneer een perfecte indeling niet mogelijk is.

### Kiezen tussen Zwitserse varianten

**Dutch** is de standaard. Het codeert 21 criteria in één enkel Blossom-matchingprobleem, wat een globaal optimale oplossing garandeert binnen de FIDE-beperkingen. Tenzij reglementen of toernooikenmerken iets anders vereisen, is Dutch de juiste keuze.

**Burstein** breidt Dutch uit met een oppositie-indexmechanisme. Tijdens de vroege "seeding"-rondes volgen de indelingen de standaard Dutch-regels. Na de seedingfase worden spelers opnieuw gerangschikt op basis van Buchholz- en Sonneborn-Berger-indices, wat meer gebalanceerde oppositie in latere rondes oplevert. Dit past bij evenementen waar de vroege rondes het veld sorteren en latere rondes gelijkwaardig presterende spelers moeten koppelen.

**Dubov** verwerkt scoregroepen in oplopende ARO-volgorde (Average Rating of Opponents) in plaats van aflopend rangnummer. Dit spreidt sterke oppositie gelijkmatiger over de indeling. Het gebruikt transpositie-gebaseerde matching binnen scoregroepen, wat eenvoudiger is dan Blossom maar de meeste praktische gevallen goed afhandelt.

**Lim** gebruikt een mediaan-eerst verwerkingsvolgorde (middelste scoregroepen eerst, dan naar buiten) en expliciete floater-classificatie (typen A tot en met D). De exchange-gebaseerde matching binnen scoregroepen geeft de arbiter een voorspelbaarder indelingsproces, ten koste van enige optimaliteit vergeleken met globale Blossom-matching.

**Double-Swiss** gebruikt lexicografische bracket-indeling uit de gedeelde `lexswiss`-bibliotheek. Het deelt sneller in dan Blossom-gebaseerde systemen en bevat een expliciet verbod op drie opeenvolgende partijen met dezelfde kleur. Het richt zich op grote open evenementen waar rekensnelheid ertoe doet.

## Teamzwitsers

[Teamzwitsers](team/) (FIDE C.04.6) deelt teams in in plaats van individuen. Het deelt de lexicografische bracket-indelingsinfrastructuur met Double-Swiss, maar voegt teamspecifieke kleurvoorkeur toe (typen A, B of Geen op basis van bord-1-geschiedenis) en een 9-staps kleurverdelingsprocedure. De primaire score kan matchpunten of partijpunten zijn, configureerbaar via opties.

## Niet-Zwitserse systemen

| Systeem                     | Matchingstrategie                   | Meest geschikt voor                             |
| --------------------------- | ----------------------------------- | ----------------------------------------------- |
| [Round-Robin](round-robin/) | FIDE Berger-tabellen (C.05 Annex 1) | Evenementen met vast deelnemersveld, competitie |
| [Keizer](keizer/)           | Top-down op Keizer-score            | Clubevenementen met competitieve prikkels       |

**Round-Robin** genereert indelingen op basis van FIDE Berger-rotatietabellen. Elke speler speelt precies één keer tegen elke andere speler per cyclus, met configureerbare meervoudige cycli en optionele laatste-twee-rondes-wissel voor dubbel round-robin-evenementen. Er is geen scoregebaseerde matching -- het schema staat volledig vast voordat het toernooi begint.

**Keizer** rangschikt spelers op Keizer-score (berekend door de Keizer-scoringsengine) en deelt van boven naar beneden in: eerste tegen tweede, derde tegen vierde, enzovoort. Herhaling-vermijding duwt tegenstanders uit elkaar wanneer ze al eerder tegen elkaar speelden. Keizer-indeling heeft alleen zin in combinatie met Keizer-scoring, omdat de rangschikking die de indeling aanstuurt afhangt van de iteratieve Keizer-scoreberekening.

## Interface

Alle acht engines implementeren dezelfde interface:

```go
type Pairer interface {
    Pair(ctx context.Context, state TournamentState) (PairingResult, error)
}
```

De `TournamentState` bevat alle spelersgegevens, rondehistorie, partijresultaten en configuratie. Het geretourneerde `PairingResult` bevat de lijst van partijtoewijzingen (bordtoewijzingen met wit/zwart) en eventuele bye-vermeldingen. De indelingsengine wijzigt nooit de invoerstaat.

Elke engine biedt ook een `NewFromMap(map[string]any)`-constructor voor instantiatie vanuit generieke configuratie (JSON, TRF-opties, CLI-vlaggen). Engine-specifieke opties staan gedocumenteerd op de pagina van elke engine.
