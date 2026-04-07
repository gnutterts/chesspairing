---
title: "Floaters"
linkTitle: "Floaters"
weight: 6
description: "When a player must be paired outside their score group — upfloaters and downfloaters."
---

In a [Swiss tournament](/docs/concepts/swiss-system/), players are grouped by score and ideally paired against opponents in the same score group. But score groups do not always cooperate. When a group has an odd number of players, or when internal constraints prevent a complete matching, at least one player must leave their group to find an opponent elsewhere. That player is called a **floater**.

## Downfloaters and upfloaters

A floater moves in one of two directions:

- **Downfloater** -- a player who drops to a lower score group to find an opponent. The downfloater plays someone with fewer points, giving them an easier-than-expected game.
- **Upfloater** -- a player who rises to a higher score group. The upfloater faces a stronger opponent, making their game harder than expected.

These two always come in pairs: every downfloater from one score group produces an upfloater in the group that receives them. If the 3-point group has 7 players, one player floats down to the 2.5-point group, where a player from that group effectively floats up by being matched with the higher-scoring opponent.

## Why floating is necessary

Floating occurs for several reasons:

- **Odd group size.** A score group with an odd number of players cannot pair everyone internally. At least one player must be sent to an adjacent group.
- **Already-played opponents.** Two players in the same score group may have already faced each other in an earlier round. If no other valid pairings exist, someone must float.
- **Color constraints.** When too many players in a group need the same color and the absolute color rule cannot be satisfied, floating resolves the deadlock.
- **Forbidden pairs.** Players declared as forbidden pairs (e.g., from the same club or family) cannot meet, further constraining internal matching.

The pairing engine tries to minimize the number of floaters, because floating disrupts the competitive balance that Swiss pairings aim for.

## Float tracking

Every pairing system tracks float history to prevent the same player from floating round after round. The specifics differ by system.

### Dutch and Burstein systems

The Dutch and Burstein pairers record each player's float direction per round and track **consecutive same-direction floats**. The optimization criteria (C14 through C21 in the FIDE regulations) penalize pairings that would cause a player to float in the same direction they floated the previous round -- or even two rounds ago. These criteria are encoded as edge weights in the [Blossom matching](/docs/algorithms/blossom/) graph, so the algorithm naturally avoids repeated floating when better alternatives exist.

Specifically, the system tracks:

- Whether the player floated down or up in the most recent round.
- How many consecutive rounds the player has floated in the same direction.
- Whether the player floated to the same score group as the previous round (C14 penalizes this specifically).

### Lim system

The [Lim pairing system](/docs/pairing-systems/lim/) takes a different approach and classifies each potential floater into one of four types based on two factors: whether the player has already been floated into the current score group from a higher group, and whether they have a compatible opponent in the adjacent group.

| Type  | Already floated? | Compatible opponent in adjacent group? |
| ----- | ---------------- | -------------------------------------- |
| **A** | Yes              | No                                     |
| **B** | Yes              | Yes                                    |
| **C** | No               | No                                     |
| **D** | No               | Yes                                    |

Type A is the most disadvantaged (already floated once and has no compatible partner to move to), while Type D is the least disadvantaged (has not floated yet and has options available). When selecting which player to float, the Lim system prefers Type D candidates first, choosing the one that best equalizes color balance within the remaining group. When floating down, the lowest-numbered player is preferred; when floating up, the highest-numbered player is selected.

## Float direction and optimization

All Swiss pairing systems share the same goal: minimize the competitive impact of floating. In practice, this means:

1. **Minimize the total number of floaters.** Fewer floaters means more players face opponents at their own level.
2. **Avoid repeating a float for the same player.** A player who floated down last round should not float down again this round if it can be avoided.
3. **Distribute floats across players.** If floating must happen, spread it among different players rather than burdening the same person repeatedly.
4. **Prefer smaller score differences.** A player floating from 3 points to 2.5 points is less disruptive than one floating from 3 to 2.

These considerations are encoded differently in each system -- as optimization criteria weights in the Dutch and Burstein systems, as floater type classifications in the Lim system, and as selection rules in the Dubov system -- but the underlying principle is the same.

## See also

- [Swiss system overview](/docs/concepts/swiss-system/) -- how score groups are formed and processed
- [Lim pairing system](/docs/pairing-systems/lim/) -- the floater type classification in detail
