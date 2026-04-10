---
title: "Blossom Matching"
linkTitle: "Blossom"
weight: 1
description: "Het maximum weight matching-algoritme van Edmonds — O(n^3) voor algemene grafen met blossom-contractie."
---

## Probleemstelling

Gegeven een ongerichte gewogen graaf $G = (V, E)$ met kantgewichten
$w : E \to \mathbb{Z}$, vind een **matching** $M \subseteq E$ — een
verzameling kanten waarvan geen twee een eindpunt delen — die het totale
gewicht maximaliseert:

$$\max_{M} \sum_{e \in M} w(e)$$

In de **maximum cardinaliteit**-variant is het primaire doel om $|M|$ te
maximaliseren; onder alle matchings met maximale cardinaliteit maximaliseren
we vervolgens het totale gewicht.

De implementatie staat in `algorithm/blossom/`.

---

## Waarom Blossom?

Zwitserse indeling is _niet_ bipartiet. In een bipartiet matchingprobleem
behoort elk knooppunt tot een van twee vaste groepen en verbinden kanten
alleen knooppunten uit verschillende groepen. Bij Zwitserse indeling kunnen
twee willekeurige spelers die elkaar nog niet hebben ontmoet, worden ingedeeld —
de graaf is een **algemene** (niet-bipartiete) graaf.

Klassieke algoritmen zoals de Hongaarse methode of Hopcroft-Karp zijn
beperkt tot bipartiete grafen. Het maximum weight matching-probleem op
algemene grafen vereist Edmonds' **Blossom-algoritme** (1965), dat oneven
cycli afhandelt via een contractietechniek die geen enkel bipartiet algoritme
biedt.

---

## LP-relaxatie

Het matchingprobleem heeft een heldere geheeltallige programmeringsformulering.
Ken een binaire variabele $x_e \in \{0, 1\}$ toe aan elke kant $e \in E$:

$$\max \sum_{e \in E} w(e) \, x_e$$

met de voorwaarden:

$$\sum_{e \ni v} x_e \leq 1 \quad \text{for every } v \in V$$

Voor bipartiete grafen heeft deze LP-relaxatie gehele optima (door totale
unimodulariteit). Voor algemene grafen geldt dat niet — fractionele
halfgehele oplossingen kunnen optreden bij oneven cycli. De oplossing is de
**oneven-verzameling-voorwaarden** (Edmonds, 1965):

$$\sum_{e \subseteq B} x_e \leq \frac{|B| - 1}{2} \quad \text{for every odd subset } B \subseteq V, \; |B| \geq 3$$

waarbij $e \subseteq B$ betekent dat beide eindpunten van $e$ in $B$ liggen.
Het toevoegen van deze voorwaarden (een voor elke oneven deelverzameling)
herstelt de geheeltalligheid. Het Blossom-algoritme dwingt ze impliciet af
via duale variabelen op gecontracteerde blossoms.

---

## Duale variabelen

Het LP-duale associeert twee soorten variabelen met het primale probleem:

- **Knooppunt-duals** $u_v$ voor elk $v \in V$.
- **Blossom-duals** $z_B \geq 0$ voor elke niet-triviale blossom $B$ (oneven
  deelverzameling met $|B| \geq 3$).

De **complementaire slackheids**-voorwaarde voor een kant $(i, j)$ is:

$$\pi(i, j) = u_i + u_j + \sum_{\substack{B \ni i \\ B \ni j}} z_B - w(i, j) \geq 0$$

Een kant is **strak** wanneer $\pi(i, j) = 0$. Een primaal-duaal paar
$(M, u, z)$ is optimaal als:

1. Elke gematcht kant strak is.
2. Elke blossom $B$ met $z_B > 0$ "vol" is (gematcht op
   $\frac{|B|-1}{2}$ kanten).

### Opslagconventie

De implementatie slaat $2u_v$ op in `dualvar[v]` om breuken te vermijden
(alle rekenkunde blijft in gehele getallen). De slack van kant $k$ die
knooppunten $i$ en $j$ verbindt is daarom:

$$\text{slack}(k) = \text{dualvar}[i] + \text{dualvar}[j] - 2 \, w(k)$$

Initiële knooppunt-duals worden ingesteld op het maximale kantgewicht:

$$\text{dualvar}[v] = w_{\max} \quad \text{for all } v \in V$$

Initiële blossom-duals zijn nul: $z_B = 0$.

---

## Algoritmestructuur

Het algoritme verloopt in **fasen**. Elke fase probeert een **vergrotend pad**
te vinden — een pad tussen twee ongematchte (blootgestelde) knooppunten dat
afwisselt tussen ongematchte en gematchte kanten. Vergroten langs zo'n pad
verhoogt $|M|$ met een. Na maximaal $\lfloor n/2 \rfloor$ fasen is de
matching maximaal.

Binnen elke fase onderhoudt het algoritme een woud van alternerende bomen
geworteld in blootgestelde knooppunten. Knooppunten worden gelabeld:

| Label | Naam       | Betekenis                                                                                        |
| ----- | ---------- | ------------------------------------------------------------------------------------------------ |
| S     | outer      | Blootgesteld knooppunt, of bereikt via een pad van even lengte vanaf een blootgesteld knooppunt. |
| T     | inner      | Bereikt via een pad van oneven lengte (partner van een S-knooppunt via een gematchte kant).      |
| free  | ongelabeld | Nog niet bereikt door enige alternerende boom.                                                   |

De fase scant herhaaldelijk kanten die incidenten aan S-knooppunten. Drie
gebeurtenissen kunnen optreden bij het verwerken van een strakke kant
$(v, w)$ met $v$ een S-knooppunt:

1. **Groei**: $w$ is vrij — label $w$ als T, label zijn partner als S. De
   alternerende boom groeit met twee knooppunten.
2. **Blossom**: $w$ is een S-knooppunt in _dezelfde_ boom — een oneven
   cyclus is gevonden. Contracteer deze tot een superknooppunt (zie hieronder).
3. **Vergroting**: $w$ is een S-knooppunt in een _andere_ boom — een
   vergrotend pad bestaat. Wissel gematcht/ongematcht langs het pad en
   beëindig de fase.

Wanneer er geen strakke S-kanten meer zijn, werkt het algoritme de duale
variabelen bij om nieuwe strakke kanten te creëren.

---

## De vier deltatypes

Bij elke duale-bijwerkstap berekent het algoritme vier kandidaat-stapgroottes
en neemt het minimum:

| Type       | Formule                                                                    | Resulterende actie                                                             |
| ---------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| $\delta_1$ | $\min_{v \in S} u_v$                                                       | Beëindigen (geen vergrotend pad bestaat — alleen bij `maxCardinality = false`) |
| $\delta_2$ | $\min_{\substack{v \text{ free} \\ (v,w) \in E,\, w \in S}} \pi(v, w)$     | Groei: een vrij knooppunt krijgt een strakke kant naar een S-knooppunt         |
| $\delta_3$ | $\min_{\substack{B_1, B_2 \in S \\ B_1 \neq B_2}} \frac{\pi(B_1, B_2)}{2}$ | Vergroting of blossom: een S-S-kant wordt strak                                |
| $\delta_4$ | $\min_{\substack{B \in T \\ B \text{ non-trivial}}} z_B$                   | Expansie: de duale van een T-blossom bereikt nul                               |

Stel $\delta = \min(\delta_1, \delta_2, \delta_3, \delta_4)$. Werk dan de
duals bij:

- S-knooppunt: $\text{dualvar}[v] \mathrel{-}= \delta$
- T-knooppunt: $\text{dualvar}[v] \mathrel{+}= \delta$
- S-blossom (niet-triviaal): $z_B \mathrel{+}= \delta$
- T-blossom (niet-triviaal): $z_B \mathrel{-}= \delta$

Dit behoudt de slack van kanten binnen S-T- of T-T-paren (hun duale
aanpassingen heffen elkaar op), terwijl de slack van S-vrije kanten strikt
daalt ($\delta_2$) en S-S-kanten ($\delta_3$, met factor 2 omdat beide
eindpunten dalen). Blossom-duals $z_B$ blijven niet-negatief omdat
$\delta \leq \delta_4$.

---

## Blossom-contractie

Wanneer twee S-knooppunten $v$ en $w$ in _verschillende_ toplevel-blossoms
een strakke kant delen, volgt het algoritme alternerende paden van beide
terug naar de boomwortels. Twee uitkomsten zijn mogelijk:

**Dezelfde wortel (oneven cyclus gevonden).** De paden ontmoeten elkaar bij
een gemeenschappelijke voorouder $b$. De knooppunten op de cyclus $v \to b
\to w$ (via de alternerende boom) plus de kant $(v, w)$ vormen een oneven
cyclus van lengte $2k + 1$. Het algoritme contracteert deze cyclus tot een
enkel superknooppunt — een **blossom** — met $b$ als **basis**:

1. Alle knooppunten in de cyclus worden samengevoegd tot blossom $B$.
2. $B$ erft het S-label en alle kanten die incident zijn aan zijn leden.
3. $B$ krijgt een duale variabele $z_B = 0$ (deze groeit bij volgende duale
   bijwerkingen zolang $B$ een S-blossom blijft).
4. Gematchte kanten binnen $B$ worden behouden; de matching op de rand van
   $B$ wordt bepaald door het basisknooppunt.

**Verschillende wortels (vergrotend pad gevonden).** Het pad van de wortel
van $v$ via de S-T-boom naar $v$, over kant $(v, w)$, en van $w$ via zijn
S-T-boom naar de wortel van $w$ vormt een vergrotend pad. Wissel
gematcht/ongematcht langs dit pad.

### Blossom-expansie

Wanneer de duale $z_B$ van een niet-triviale T-blossom nul bereikt
($\delta_4$), wordt de blossom terug uitgebreid naar zijn samenstellende
sub-blossoms. Sub-blossoms worden opnieuw gelabeld (sommige worden S,
sommige T) zodat de alternerende boomstructuur behouden blijft. Aan het
einde van een fase worden S-blossoms met $z_B = 0$ ook uitgebreid.

---

## Vergrotend pad

Wanneer een vergrotend pad is gevonden, wisselt de `augmentMatching`-functie
gematcht/ongematcht langs het pad:

1. Traceer vanaf beide eindpunten van de ontdekkende kant terug door de
   alternerende bomen naar hun respectieve wortels (blootgestelde
   knooppunten).
2. Wissel langs elk spoor de gematcht/ongematcht-status van elke kant.
3. Als het pad door een niet-triviale blossom loopt, roteert
   `augmentBlossom` recursief de interne kinderlijst van de blossom zodat het
   ingangsknooppunt de nieuwe basis wordt.

Na vergroting neemt $|M|$ met een toe en eindigt de fase.

---

## Twee implementatievarianten

Het `algorithm/blossom/`-pakket biedt twee functies:

| Functie                                                             | Gewichtstype | Toepassing                                                                    |
| ------------------------------------------------------------------- | ------------ | ----------------------------------------------------------------------------- |
| `MaxWeightMatching(edges []BlossomEdge, maxCardinality bool) []int` | `int64`      | Kleine problemen of wanneer 63 bruikbare bits volstaan                        |
| `MaxWeightMatchingBig(edges []BigEdge, maxCardinality bool) []int`  | `*big.Int`   | Zwitserse indelingskantgewichten (zie [Kantgewicht-codering](../edge-weights/)) |

Beide retourneren een slice `m` waarbij `m[i]` het knooppunt is dat gematcht
is met `i`, of `-1` als het ongematcht is.

De `*big.Int`-variant bestaat omdat Zwitserse indelingskantgewichten 16+
criteriavelden coderen in een enkel geheel getal. Voor een toernooi met 100
spelers en 9 ronden kan de totale bitbreedte ongeveer 294 bits bereiken —
ruim boven de 63 bruikbare bits van `int64`. De algoritmestructuur is
identiek in beide varianten; alleen de rekenkunde verschilt.

---

## Complexiteit

**Tijd:** $O(n^3)$ waarbij $n = |V|$.

Elke fase voert maximaal $O(n)$ duale bijwerkingen uit (omdat elke
bijwerking ofwel de boom laat groeien, een blossom ontdekt of er een
uitbreidt). Elke duale bijwerking scant $O(n)$ knooppunten/blossoms om de
minimale delta te vinden. Er zijn maximaal $\lfloor n/2 \rfloor$ fasen (een
vergroting per fase). Dit geeft:

$$O\!\left(\frac{n}{2}\right) \times O(n) \times O(n) = O(n^3)$$

**Ruimte:** $O(n + m)$ waarbij $m = |E|$. De dominante structuren zijn de
buurlijsten ($O(m)$) en de per-knooppunt/blossom-arrays ($O(n)$ elk, met
tot $2n$ plaatsen om blossoms te accommoderen).

---

## Implementatienotities

De Go-implementatie is een directe port van Joris van Rantwijks Python-
referentie-implementatie (`mwmatching.py`, publiek domein). Belangrijke
overeenkomsten:

| Python             | Go (`blossom.go`)                                                    |
| ------------------ | -------------------------------------------------------------------- |
| `mate[v]`          | `mate[v]` — remote endpoint-index, of $-1$                           |
| `label[b]`         | `label[b]` — `0` = ongelabeld, `1` = S, `2` = T                      |
| `inblossom[v]`     | `inblossom[v]` — toplevel-blossom die $v$ bevat                      |
| `dualvar[v]`       | `dualvar[v]` — slaat $2u_v$ op voor knooppunten, $z_B$ voor blossoms |
| `blossomchilds[b]` | `blossomchilds[b]` — geordende kinderlijst van blossom $b$           |
| `blossombase[b]`   | `blossombase[b]` — basisknooppunt van blossom $b$                    |
| `bestedge[b]`      | `bestedge[b]` — kant met minste slack naar een andere S-blossom      |

Knooppunten zijn genummerd $0, 1, \ldots, n-1$. Niet-triviale blossoms zijn
genummerd $n, n+1, \ldots, 2n-1$ en worden toegewezen uit een vrije pool.
Kanten worden aangeduid met index $k$; hun twee eindpunten zijn opgeslagen op
posities $2k$ en $2k+1$ in de `endpoint`-array.

---

## Referenties

- J. Edmonds, "Paths, trees, and flowers," _Canadian Journal of Mathematics_,
  vol. 17, pp. 449--467, 1965.
- J. van Rantwijk, `mwmatching.py` — Python-referentie-implementatie
  (publiek domein).
- Z. Galil, "Efficient algorithms for finding maximum matching in graphs,"
  _ACM Computing Surveys_, vol. 18, no. 1, pp. 23--38, 1986.
