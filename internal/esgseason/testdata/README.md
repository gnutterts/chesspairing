# ESG club season data

`season.json` is a deterministic, language-neutral 24-round club season with
26 anonymous players. `expected.json` records the Keizer standings after every
round. The generator and regression test are in the parent directory.

## Fields

- `players`: anonymous player code, rating, entry order, and optional joining
  and withdrawal rounds.
- `rounds.present`: players available for pairing in that round.
- `rounds.events`: club events; `bye`, `absent`, `external`, `forfeit`,
  `bothAbsent`, `newcomer`, and `withdrawn` identify their affected players.
- `rounds.pairings`: generated white player, black player, and result (`1-0`,
  `½-½`, `0-1`, or `0-0f`).
- `expected.rounds.standings`: the player score, value number, and position
  after the round.

## Club rules represented

Rounds 1--4 use Dutch pairing with 1-½-0 results while Keizer keeps the
ranking; later rounds use Keizer top-down. Initial order is rating then entry
order. The four periods each contain six rounds: opponents do not repeat in a
period and have at least one intervening round otherwise. A pairing-allocated
bye is worth 40, goes to the lowest eligible player, and is received at most
once. An ordinary absence earns 20 for its first five occurrences and zero
afterward; an external club game earns 40. A new player without a rating earns
15 for every round missed before joining. A withdrawn player remains ranked
and is treated as absent from then on.

For example, P05 is absent in rounds 2, 3, 4, 6, and 10 and earns 20 points
for each; its later absences in rounds 14, 18, 22, 23, and 24 earn zero. In round 7 the
`bothAbsent` pair has a `0-0f` result: both players are treated as absent, and
that pairing is not an encounter, so the same pair can be paired again in the
period. For a single forfeit, a forfeit loss scores 0 points; no absence points.

Value numbers are 60, 59, ..., 35 (`61 - position`). A win, draw, and loss
score 100%, 50%, and 0% respectively of the opponent's value number before
the round; a player's own value number is never added. The standings are not
recomputed from later rankings.
