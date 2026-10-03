// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
)

func init() {
	RegisterTeam("mpvgp", func() chesspairing.TieBreaker { return &MPvGP{} })
	RegisterTeam("emmsb", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamMatch}
	})
	RegisterTeam("emmsb-cut1", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamMatch, cut1: true}
	})
	RegisterTeam("emgsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamGame} })
	RegisterTeam("egmsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamMatch} })
	RegisterTeam("eggsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamGame} })
	RegisterTeam("buchholz-mp", func() chesspairing.TieBreaker { return &TeamBuchholz{} })
	RegisterTeam("buchholz-mp-cut1", func() chesspairing.TieBreaker { return &TeamBuchholz{cut1: true} })
	RegisterTeam("board-count", func() chesspairing.TieBreaker { return &BoardCount{} })
	RegisterTeam("top-board-results", func() chesspairing.TieBreaker { return &TopBoardResults{} })
	RegisterTeam("bottom-board-elimination", func() chesspairing.TieBreaker { return &BottomBoardElimination{} })
}

type teamScore int

const (
	teamMatch teamScore = iota
	teamGame
)

// MPvGP computes the secondary team score: game points when match points are
// primary, and match points when game points are primary. It implements FIDE
// C.07:2026 Article 13.1.
type MPvGP struct{}

func (*MPvGP) ID() string   { return "mpvgp" }
func (*MPvGP) Name() string { return "Match Points or Game Points" }
func (*MPvGP) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	primary := team.ParseOptions(state.ScoringConfig.Options).WithDefaults().PrimaryScore
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		value := score.Score
		if score.Team != nil {
			if primary == "game" {
				value = score.Team.Match
			} else {
				value = score.Team.Game
			}
		}
		out[i] = chesspairing.TieBreakValue{PlayerID: score.PlayerID, Value: value}
	}
	return out, nil
}

// ExtendedSonnebornBerger computes an extended Sonneborn-Berger team value.
// The opponent and scored elements are MP or GP, selected independently, so
// the four variants of FIDE C.07:2026 Article 13.2 are expressible:
//
//	emmsb = opponent MP × own MP (13.2.1)
//	emgsb = opponent MP × own GP (13.2.2)
//	egmsb = opponent GP × own MP (13.2.3)
//	eggsb = opponent GP × own GP (13.2.4)
type ExtendedSonnebornBerger struct {
	opponent, scored teamScore
	cut1             bool
}

func (e *ExtendedSonnebornBerger) ID() string {
	id := [...]string{"emmsb", "emgsb", "egmsb", "eggsb"}[e.opponent*2+e.scored]
	if e.cut1 {
		return id + "-cut1"
	}
	return id
}

func (e *ExtendedSonnebornBerger) Name() string {
	names := [...]string{
		"Extended Sonneborn-Berger (Match/Match)",
		"Extended Sonneborn-Berger (Match/Game)",
		"Extended Sonneborn-Berger (Game/Match)",
		"Extended Sonneborn-Berger (Game/Game)",
	}
	name := names[e.opponent*2+e.scored]
	if e.cut1 {
		return name + " Cut 1"
	}
	return name
}

func (e *ExtendedSonnebornBerger) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildTeamOpponentRecords(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		contributions := teamSonnebornBergerContributions(score.PlayerID, table, e.opponent, e.scored)
		if e.cut1 {
			contributions = cutLeastSonnebornBerger(contributions)
		}
		for _, contribution := range contributions {
			out[i].Value += contribution.value
		}
		out[i].PlayerID = score.PlayerID
	}
	return out, nil
}

// TeamBuchholz computes Buchholz using opponents' match points, as used for
// team competitions by FIDE C.07:2026 Articles 13 and 16.
type TeamBuchholz struct{ cut1 bool }

func (b *TeamBuchholz) ID() string {
	if b.cut1 {
		return "buchholz-mp-cut1"
	}
	return "buchholz-mp"
}

func (b *TeamBuchholz) Name() string {
	if b.cut1 {
		return "Buchholz MP Cut-1"
	}
	return "Buchholz MP"
}

func (b *TeamBuchholz) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildTeamOpponentRecords(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		contributions := teamBuchholzContributions(score.PlayerID, table)
		if b.cut1 {
			contributions = cutLeast(contributions)
		}
		for _, contribution := range contributions {
			out[i].Value += contribution.value
		}
		out[i].PlayerID = score.PlayerID
	}
	return out, nil
}

// BoardCount computes the board count described in FIDE C.07:2026 Article
// 12.1: each board's game points are multiplied by the board number and the
// products are added. A lower value is better (wins on the top boards produce
// a smaller count); to sort correctly, the returned value is negated.
type BoardCount struct{}

func (*BoardCount) ID() string   { return "board-count" }
func (*BoardCount) Name() string { return "Board Count" }
func (*BoardCount) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	slices := teamBoardPointSlices(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		for _, matchPoints := range slices[score.PlayerID] {
			for board, points := range matchPoints {
				out[i].Value -= float64(board+1) * points
			}
		}
		out[i].PlayerID = score.PlayerID
	}
	return out, nil
}

// TopBoardResults computes the accumulated first-board result, as defined by
// FIDE C.07:2026 Article 12.2.
type TopBoardResults struct{}

func (*TopBoardResults) ID() string   { return "top-board-results" }
func (*TopBoardResults) Name() string { return "Top Board Results" }
func (*TopBoardResults) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	slices := teamBoardPointSlices(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		for _, matchPoints := range slices[score.PlayerID] {
			if len(matchPoints) > 0 {
				out[i].Value += matchPoints[0]
			}
		}
		out[i].PlayerID = score.PlayerID
	}
	return out, nil
}

// BottomBoardElimination sums board results after removing the lowest board,
// as defined by FIDE C.07:2026 Article 12.3.
type BottomBoardElimination struct{}

func (*BottomBoardElimination) ID() string   { return "bottom-board-elimination" }
func (*BottomBoardElimination) Name() string { return "Bottom Board Elimination" }
func (*BottomBoardElimination) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	slices := teamBoardPointSlices(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		for _, matchPoints := range slices[score.PlayerID] {
			if len(matchPoints) > 1 {
				for j, points := range matchPoints {
					if j != len(matchPoints)-1 {
						out[i].Value += points
					}
				}
			} else if len(matchPoints) == 1 {
				out[i].Value += matchPoints[0]
			}
		}
		out[i].PlayerID = score.PlayerID
	}
	return out, nil
}

// teamOpponentRecord is the per-round team equivalent of OpponentRecord.
type teamOpponentRecord struct {
	Round      int
	Played     bool
	OpponentID string
	Match      float64
	Game       float64
	Category   UnplayedCategory
	IsVUR      bool
}

func (r teamOpponentRecord) value(kind teamScore) float64 {
	if kind == teamGame {
		return r.Game
	}
	return r.Match
}

// teamOpponentTable is the team analogue of opponentTable.
type teamOpponentTable struct {
	records       map[string][]teamOpponentRecord
	matchPoints   map[string]float64
	gamePoints    map[string]float64
	adjustedMatch map[string]float64
	adjustedGame  map[string]float64
	totalRounds   int
	roundRobin    bool
}

func buildTeamOpponentRecords(state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) teamOpponentTable {
	opts := team.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	boardCount := team.BoardCount(state, opts)
	playerTeams := team.PlayerTeams(state)

	table := teamOpponentTable{
		records:       make(map[string][]teamOpponentRecord, len(scores)),
		matchPoints:   make(map[string]float64, len(scores)),
		gamePoints:    make(map[string]float64, len(scores)),
		adjustedMatch: make(map[string]float64, len(scores)),
		adjustedGame:  make(map[string]float64, len(scores)),
		totalRounds:   len(state.Rounds),
		roundRobin:    state.PairingConfig.System == chesspairing.PairingRoundRobin,
	}

	teamIDs := make([]string, 0, len(scores))
	for _, score := range scores {
		teamIDs = append(teamIDs, score.PlayerID)
		table.records[score.PlayerID] = make([]teamOpponentRecord, 0, len(state.Rounds))
		if score.Team != nil {
			table.matchPoints[score.PlayerID] = score.Team.Match
			table.gamePoints[score.PlayerID] = score.Team.Game
		} else {
			table.matchPoints[score.PlayerID] = score.Score
			table.gamePoints[score.PlayerID] = score.Score
		}
	}

	for _, round := range state.Rounds {
		roundRecords := make(map[string]teamOpponentRecord, len(teamIDs))
		for _, id := range teamIDs {
			roundRecords[id] = teamOpponentRecord{Round: round.Number, Category: RequestedByeFinal, IsVUR: true}
		}
		for _, match := range round.Matches {
			homeGame, awayGame := team.MatchGamePoints(match, opts.Options, playerTeams)
			homeMatch, awayMatch := team.MatchPoints(match, homeGame, awayGame, opts)
			home := teamOpponentRecord{Round: round.Number, OpponentID: match.AwayID, Match: homeMatch, Game: homeGame}
			away := teamOpponentRecord{Round: round.Number, OpponentID: match.HomeID, Match: awayMatch, Game: awayGame}
			switch {
			case teamMatchForfeit(match):
				if homeMatch > awayMatch {
					home.Category, away.Category, away.IsVUR = ForfeitWin, ForfeitLoss, true
				} else if awayMatch > homeMatch {
					home.Category, home.IsVUR, away.Category = ForfeitLoss, true, ForfeitWin
				} else {
					home.Played, away.Played = true, true
				}
			case teamMatchUndecided(match):
				home.Category, away.Category = None, None
			default:
				home.Played, away.Played = true, true
			}
			if _, ok := roundRecords[match.HomeID]; ok {
				roundRecords[match.HomeID] = home
			}
			if _, ok := roundRecords[match.AwayID]; ok {
				roundRecords[match.AwayID] = away
			}
		}
		for _, bye := range round.TeamByes {
			record := teamOpponentRecord{Round: round.Number}
			switch bye.Type {
			case chesspairing.ByePAB:
				record.Match, record.Game, record.Category = *opts.PointMatchDraw, *opts.PointDraw*float64(boardCount), PABOrFullPoint
			case chesspairing.ByeFullPoint:
				record.Match, record.Game, record.Category = *opts.PointMatchWin, *opts.PointWin*float64(boardCount), PABOrFullPoint
			case chesspairing.ByeHalf:
				record.Match, record.Game, record.Category, record.IsVUR = *opts.PointMatchDraw, *opts.PointDraw*float64(boardCount), RequestedByeFinal, true
			case chesspairing.ByeZero:
				record.Game, record.Category, record.IsVUR = *opts.PointLoss, RequestedByeFinal, true
			case chesspairing.ByeAbsent:
				record.Game, record.Category, record.IsVUR = *opts.PointAbsent, RequestedByeFinal, true
			case chesspairing.ByeExcused:
				record.Game, record.Category, record.IsVUR = *opts.PointExcused, RequestedByeFinal, true
			case chesspairing.ByeClubCommitment:
				record.Game, record.Category, record.IsVUR = *opts.PointClubCommitment, RequestedByeFinal, true
			}
			if _, ok := roundRecords[bye.PlayerID]; ok {
				roundRecords[bye.PlayerID] = record
			}
		}
		for _, id := range teamIDs {
			table.records[id] = append(table.records[id], roundRecords[id])
		}
	}

	for id, records := range table.records {
		for i := range records {
			if records[i].Category != RequestedByeFinal {
				continue
			}
			for j := i + 1; j < len(records); j++ {
				if teamIsNonVUR(records[j]) {
					records[i].Category = RequestedByeFollowedByPlay
					break
				}
			}
		}
		table.records[id] = records
		table.adjustedMatch[id] = table.matchPoints[id]
		table.adjustedGame[id] = table.gamePoints[id]
		for _, record := range records {
			if record.Category == RequestedByeFinal {
				table.adjustedMatch[id] += *opts.PointMatchDraw - record.Match
				table.adjustedGame[id] += *opts.PointDraw*float64(boardCount) - record.Game
			}
		}
	}
	return table
}

func teamIsNonVUR(record teamOpponentRecord) bool {
	if record.Played {
		return true
	}
	switch record.Category {
	case PABOrFullPoint, ForfeitWin:
		return true
	default:
		return false
	}
}

func teamMatchForfeit(match chesspairing.MatchData) bool {
	if len(match.Boards) == 0 {
		return false
	}
	for _, board := range match.Boards {
		if !board.IsForfeit && !board.Result.IsForfeit() {
			return false
		}
	}
	return true
}

func teamMatchUndecided(match chesspairing.MatchData) bool {
	return team.MatchIsUndecided(match)
}

func teamDummyValue(id string, kind teamScore, table teamOpponentTable) float64 {
	if kind == teamGame {
		return table.gamePoints[id]
	}
	return table.matchPoints[id]
}

func teamAdjustedValue(id string, kind teamScore, table teamOpponentTable) float64 {
	if kind == teamGame {
		return table.adjustedGame[id]
	}
	return table.adjustedMatch[id]
}

func teamBuchholzContributions(id string, table teamOpponentTable) []tieBreakContribution {
	contributions := make([]tieBreakContribution, 0, table.totalRounds)
	for _, record := range table.records[id] {
		var opponent float64
		switch {
		case record.Played:
			opponent = table.adjustedMatch[record.OpponentID]
		case record.Category == ForfeitWin || record.Category == ForfeitLoss:
			if table.roundRobin {
				opponent = table.adjustedMatch[record.OpponentID]
				contributions = append(contributions, tieBreakContribution{value: opponent, significance: opponent})
				continue
			}
			opponent = teamDummyValue(id, teamMatch, table)
		case record.Category != None:
			opponent = teamDummyValue(id, teamMatch, table)
		default:
			continue
		}
		contributions = append(contributions, tieBreakContribution{value: opponent, significance: opponent, vur: record.IsVUR})
	}
	return contributions
}

func teamSonnebornBergerContributions(id string, table teamOpponentTable, opponent, scored teamScore) []tieBreakContribution {
	contributions := make([]tieBreakContribution, 0, table.totalRounds)
	for _, record := range table.records[id] {
		var opp, own float64
		switch {
		case record.Played:
			opp = teamAdjustedValue(record.OpponentID, opponent, table)
			own = record.value(scored)
		case record.Category == ForfeitWin || record.Category == ForfeitLoss:
			if table.roundRobin {
				opp = teamAdjustedValue(record.OpponentID, opponent, table)
				own = record.value(scored)
				contributions = append(contributions, tieBreakContribution{value: opp * own, significance: opp, result: own})
				continue
			}
			opp = teamDummyValue(id, opponent, table)
			own = record.value(scored)
		case record.Category != None:
			opp = teamDummyValue(id, opponent, table)
			own = record.value(scored)
		default:
			continue
		}
		contributions = append(contributions, tieBreakContribution{value: opp * own, significance: opp, result: own, vur: record.IsVUR})
	}
	return contributions
}

func teamBoardPointSlices(state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) map[string][][]float64 {
	opts := team.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	playerTeams := team.PlayerTeams(state)
	out := make(map[string][][]float64, len(scores))
	for _, score := range scores {
		out[score.PlayerID] = [][]float64{}
	}
	for _, round := range state.Rounds {
		for _, match := range round.Matches {
			var homePoints []float64
			var awayPoints []float64
			for _, board := range match.Boards {
				white, black := team.BoardPoints(board, opts.Options)
				whiteHome := team.BelongsToTeam(board.WhiteID, match.HomeID, playerTeams)
				blackHome := team.BelongsToTeam(board.BlackID, match.HomeID, playerTeams)
				home, away := white, black
				if !whiteHome && blackHome {
					home, away = black, white
				}
				homePoints = append(homePoints, home)
				awayPoints = append(awayPoints, away)
			}
			if len(homePoints) > 0 {
				if _, ok := out[match.HomeID]; ok {
					out[match.HomeID] = append(out[match.HomeID], homePoints)
				}
			}
			if len(awayPoints) > 0 {
				if _, ok := out[match.AwayID]; ok {
					out[match.AwayID] = append(out[match.AwayID], awayPoints)
				}
			}
		}
	}
	return out
}

// (teamBoardCount removed, using team.BoardCount instead)
