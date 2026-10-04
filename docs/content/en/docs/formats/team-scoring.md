---
title: "Team scoring"
weight: 35
---

Team events use two scores: match points (MP) and game points (GP). Set
`system` to `team` in `ScoringConfig`. Its default primary score is MP;
set `primaryScore` to `game` to make GP the scalar `PlayerScore.Score`.
Every team score also exposes both values through `PlayerScore.Team`.

`RoundData.Matches` records a team match. Its board results determine GP;
when boards are unavailable, `MatchData.Result` carries the reported totals.
The default match scale is 2 points for a win, 1 for a draw, and 0 for a
loss. `pointMatchWin`, `pointMatchDraw`, and `pointMatchLoss` configure it.
Board-point options use the standard `pointWin`, `pointDraw`, and `pointLoss`
keys. Team byes should be stored in `RoundData.TeamByes` (preferred over `Byes`); their game
points use the draw or win value on every board, with the board count taken
from the largest match in the event.

TRF-2026 `013` records link players to teams; their order supplies the team
number because record `013` has no team-number field. Detailed `801` records
become matches with boards and simple `802` records become matches with
explicit GP totals. Record `320` supplies pairing-allocated team byes and is
converted to `RoundData.TeamByes`. Record `162` configures board points. Team
names and `310`-only membership are not represented in `TournamentState`; use
`013` membership when converting team TRF files. A team without a match or a
bye receives no implicit absence penalty, because C.04.6 does not define one.

## Team tie-breaks

The team-only registry provides the C.07 variants `mpvgp`, `emmsb`,
`emmsb-cut1`, `emgsb`, and `egmsb`, plus `eggsb`, `buchholz-mp`,
`buchholz-mp-cut1`, `board-count`, `top-board-results`, and
`bottom-board-elimination`. They are available through `tiebreaker.Get` and
are included by the CLI's `tiebreakers` command.
