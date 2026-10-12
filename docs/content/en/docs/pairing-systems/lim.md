---
title: "Lim System"
linkTitle: "Lim"
weight: 4
description: "FIDE C.04.4.3 — score-group pairing, floaters, and exchange rules."
---

## Overview

Lim is the FIDE C.04.4.3 Swiss system. It pairs score groups by the prescribed
three-phase order and uses the Article 4 exchange procedure, not global
matching. Compatibility (Art. 2.1) requires a new pairing and a permissible
colour assignment: neither player may receive a third consecutive colour or a
three more of one colour than the other.

## When to Use

Use Lim when the event is run under C.04.4.3, including its explicit floater
and Maxi-tournament rules. It does not support acceleration. Set `TotalRounds`
when known so that the Article 6 last-round exception is applied.

## Configuration

### CLI

```bash
chesspairing pair --lim tournament.trf
```

### Go API

```go
import "github.com/gnutterts/chesspairing/pairing/lim"

p := lim.New(lim.Options{
    TopSeedColor:   chesspairing.StringPtr("auto"),
    MaxiTournament: chesspairing.BoolPtr(true),
    TotalRounds:    9,
    ForbiddenPairs: [][]string{{"P1", "P2"}},
})
```

### Options

| Option | Type | Default | Description |
| --- | --- | --- | --- |
| `TopSeedColor` | `string` | `"auto"` | Top seed colour in round 1: `"auto"`, `"white"`, or `"black"`. |
| `ForbiddenPairs` | `[][]string` | `nil` | Player-ID pairs that must not meet. |
| `MaxiTournament` | `bool` | `false` | Applies the 100-point constraints of Arts. 3.2.3, 3.8, and 5.7. |
| `TotalRounds` | `int` | `0` | Enables the Article 6 exception when the current round is the last round. |

## How It Works

Pairing scores and float history follow the tournament's configured point values. Under Art. 2.2 the median is half the maximum score after completed rounds (win points × rounds ÷ 2); for football scoring, a win counts as 3.

### Score-group order and transfers

After the pairing-allocated bye of Art. 1.1, Lim processes score groups under
Art. 2.2: from the highest group down to just before the median, then from the
lowest group up, and finally the median group. The median group is paired
downward.

Before a group is divided into halves, Arts. 2.3.1--2.3.3 transfer players who
have no suitable opponent: they have met everyone in the group, cannot receive
a permissible colour, or must receive the alternate colour after two equal
colours. Only then does Art. 2.3.4 make a group even. A transfer is a floater.

### Floater selection and pairing

Arts. 3.2--3.4 select a downfloater toward the next lower group and an
upfloater toward the next higher group. Due colours first tend to equalise the
two due-colour counts. The mandatory wording in 3.3 ("the player chosen is")
and 3.5 ("shall ... be exchanged") is applied; within that requirement,
3.9.2's "should" orders a choice by least disadvantage: D, B, C, A. The types
are, respectively: not previously floated with a compatible adjacent opponent;
previously floated with one; not previously floated without one; and previously
floated without one.

Article 3.10 prevents a player who floated in the immediately preceding round
from being selected under 2.3.4 when 3.10.1 produces no type A, B, or C floater
and 3.10.2 does not reduce the number of pairings. A previous-round floater is
only the player transferred between score groups; a bye is not a float. The
restriction is applied on the same side (down or up) of the score-group route.

Incoming floaters are paired under Arts. 3.6--3.8: downfloaters first in an
upper or median group, upfloaters first in a lower group, with the stated score
and pairing-number precedence. A downfloater takes the highest-numbered
available opponent due the alternate colour; an upfloater takes the lowest.
In a Maxi tournament, Arts. 3.8 and 5.7 compare the ratings of the two
opponents being exchanged, which must differ by at most 100 points.

### Exchange, cracking, and colours

The remaining players receive the Art. 2.4 half-against-half proposal and the
Article 4 exchange procedure. See [Lim Exchange Algorithm](../../algorithms/lim-exchange/)
for Arts. 4.1--4.4, including 4.4.1 and 4.4.2.

If the median group is blocked, Art. 2.6 cracks the next lower pairing when
there are more floaters from above, otherwise the next higher pairing, and
restarts the median pairing. The pairer returns `PairingImpossible` when no
Lim pairing can be completed, or `PairingNoPABCandidate` when no eligible
pairing-allocated-bye player exists; it does not return partial pairings with
extra byes.

Colours follow Article 5. The second scrutiny of Article 5.2 ("a second
scrutiny with exchanges where necessary") exchanges opponents between pairs of
the score group to give the players their due colours: a pair in conflict is
tried against the later pairs first and then against the earlier pairs, and an
exchange is made only when it lowers the number of colour conflicts and keeps
both new pairs compatible (in a Maxi tournament, the exchanged opponents must
differ by at most 100 points, Art. 5.7). Trying the later pairs first is the
library's choice; the text does not fix an order. In the last round, Article 6
gives same-score new pairings priority over alternation and equalisation; this
needs `TotalRounds`.

## Interpretations

The FIDE text does not decide the following implementation readings:

1. If nobody has the median score, floaters arriving from above and below form
the median group at that position. It is paired last under 2.2 and 3.6.3, and
2.6 applies to it.
2. Article 4.4.1 is applied to the failing player and, whenever a 2.3.4
floater was chosen in the score group, also when 2.3.1--2.3.3 transfers made
an originally even group odd.
3. In the median group, the 3.9 types are judged against the group below,
which has already been paired.
4. Under Article 3.8, a player without any played game has no due colour and
therefore is not due the alternate colour. The floater takes the first
available player due the alternate colour; only if there is none does it take
the first available compatible player. The text does not say what happens
then.

## FIDE Reference

The implementation follows FIDE Handbook C.04.4.3 Lim System (effective
1 February 2026).