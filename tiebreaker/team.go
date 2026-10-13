// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package tiebreaker

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/team"
)

func init() {
	RegisterTeam("ede", func() chesspairing.TieBreaker { return &ExtendedDirectEncounter{} })
	RegisterTeam("sssc", func() chesspairing.TieBreaker { return &ScoresScheduleStrength{} })
	RegisterTeam("mpvgp", func() chesspairing.TieBreaker { return &MPvGP{} })
	RegisterTeam("emmsb", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamMatch}
	})
	RegisterTeam("emmsb-cut1", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamMatch, cut1: true}
	})
	RegisterTeam("emgsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamGame} })
	RegisterTeam("emgsb-cut1", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamMatch, scored: teamGame, cut1: true}
	})
	RegisterTeam("egmsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamMatch} })
	RegisterTeam("egmsb-cut1", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamMatch, cut1: true}
	})
	RegisterTeam("eggsb", func() chesspairing.TieBreaker { return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamGame} })
	RegisterTeam("eggsb-cut1", func() chesspairing.TieBreaker {
		return &ExtendedSonnebornBerger{opponent: teamGame, scored: teamGame, cut1: true}
	})
	RegisterTeam("buchholz-mp", func() chesspairing.TieBreaker { return &TeamBuchholz{} })
	RegisterTeam("buchholz-mp-cut1", func() chesspairing.TieBreaker { return &TeamBuchholz{cut1: true} })
	RegisterTeam("board-count", func() chesspairing.TieBreaker { return &BoardCount{} })
	RegisterTeam("top-board-results", func() chesspairing.TieBreaker { return &TopBoardResults{} })
	RegisterTeam("bottom-board-elimination", func() chesspairing.TieBreaker { return &BottomBoardElimination{} })
}

// ExtendedDirectEncounter ranks teams within each primary-score group using
// match points, then game points, restarting for every unresolved subset.
// The numeric value is an ordinal within the primary-score group.
type ExtendedDirectEncounter struct{}

func (*ExtendedDirectEncounter) ID() string   { return "ede" }
func (*ExtendedDirectEncounter) Name() string { return "Extended Direct Encounter" }
func (*ExtendedDirectEncounter) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	table := buildTeamOpponentRecords(state, scores)
	groups := make(map[float64][]string)
	for _, score := range scores {
		groups[score.Score] = append(groups[score.Score], score.PlayerID)
	}
	ranks := make(map[string]float64, len(scores))
	primary := teamMatch
	if team.ParseOptions(state.ScoringConfig.Options).WithDefaults().PrimaryScore == "game" {
		primary = teamGame
	}
	for _, ids := range groups {
		edeRank(ids, table, primary, float64(len(ids)), ranks)
	}
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		out[i] = chesspairing.TieBreakValue{PlayerID: score.PlayerID, Value: ranks[score.PlayerID]}
	}
	return out, nil
}

// edeRank only splits a complete round-robin subset. Repeated encounters
// contribute their average (Article 6.1.2), not their sum.
func edeRank(ids []string, table teamOpponentTable, primary teamScore, rank float64, ranks map[string]float64) {
	for _, id := range ids {
		ranks[id] = rank
	}
	if len(ids) < 2 {
		return
	}
	for _, kind := range []teamScore{primary, 1 - primary} {
		values := make(map[string]float64, len(ids))
		for _, id := range ids {
			for _, opponent := range ids {
				if id == opponent {
					continue
				}
				var sum float64
				count := 0
				for _, record := range table.records[id] {
					if record.OpponentID == opponent && (record.Played || table.roundRobin && (record.Category == ForfeitWin || record.Category == ForfeitLoss)) {
						sum += record.value(kind)
						count++
					}
				}
				if count == 0 {
					return
				}
				values[id] += sum / float64(count)
			}
		}
		slices.SortFunc(ids, func(a, b string) int { return -cmpFloat(values[a], values[b]) })
		if values[ids[0]] == values[ids[len(ids)-1]] {
			continue
		}
		for start := 0; start < len(ids); {
			end := start + 1
			for end < len(ids) && values[ids[end]] == values[ids[start]] {
				end++
			}
			if end-start == len(ids) {
				break
			}
			edeRank(ids[start:end], table, primary, rank-float64(start), ranks)
			start = end
		}
		return
	}
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

// ScoresScheduleStrength implements C.07:2026 Article 13.4.
type ScoresScheduleStrength struct{}

func (*ScoresScheduleStrength) ID() string   { return "sssc" }
func (*ScoresScheduleStrength) Name() string { return "Scores and Schedule Strength Combination" }
func (*ScoresScheduleStrength) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	opts := team.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	boards := team.BoardCount(state, opts)
	primaryMax, secondaryMatchMax := *opts.PointMatchWin, *opts.PointWin*float64(boards)
	if opts.PrimaryScore == "game" {
		primaryMax, secondaryMatchMax = secondaryMatchMax, primaryMax
	}
	if secondaryMatchMax <= 0 {
		return nil, fmt.Errorf("sssc requires a positive maximum secondary score per match")
	}
	divisor := math.Trunc(primaryMax * float64(len(state.Rounds)) / secondaryMatchMax)
	if divisor < 1 {
		return nil, fmt.Errorf("sssc requires a positive normalising factor")
	}
	table := buildTeamOpponentRecords(state, scores)
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		secondary := score.Score
		if score.Team != nil {
			secondary = score.Team.Game
			if opts.PrimaryScore == "game" {
				secondary = score.Team.Match
			}
		}
		var buchholz float64
		kind := teamMatch
		if opts.PrimaryScore == "game" {
			kind = teamGame
		}
		for _, contribution := range teamBuchholzContributionsFor(score.PlayerID, table, kind) {
			buchholz += contribution.value
		}
		out[i] = chesspairing.TieBreakValue{PlayerID: score.PlayerID, Value: secondary + buchholz/divisor}
	}
	return out, nil
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
	// Standings passes the whole field; check GP within each primary-score group.
	gamePointsByScore := make(map[float64]float64, len(scores))
	for _, score := range scores {
		if score.Team == nil {
			continue
		}
		if gamePoints, ok := gamePointsByScore[score.Score]; ok && gamePoints != score.Team.Game {
			return nil, fmt.Errorf("board count requires equal game points for teams tied at %g primary points", score.Score)
		}
		gamePointsByScore[score.Score] = score.Team.Game
	}
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

// TopBoardResults computes the accumulated board results in board order, as
// defined by FIDE C.07:2026 Article 12.2. The per-board totals are encoded
// into one number (see encodeBoardDigits); the encoding only stays
// order-preserving while base^boards < 2^53, which is far beyond a realistic
// match board count.
type TopBoardResults struct{}

func (*TopBoardResults) ID() string   { return "top-board-results" }
func (*TopBoardResults) Name() string { return "Top Board Results" }
func (*TopBoardResults) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	totals := teamBoardTotals(state, scores)
	digits := make(map[string][]float64, len(scores))
	for _, score := range scores {
		digits[score.PlayerID] = totals[score.PlayerID]
	}
	return encodeBoardDigits(scores, digits), nil
}

// BottomBoardElimination sums board results after repeatedly removing the
// bottom-most board, as defined by FIDE C.07:2026 Article 12.3. The
// per-stage totals are encoded into one number with the same base and the
// same 2^53 precision limit as TopBoardResults.
type BottomBoardElimination struct{}

func (*BottomBoardElimination) ID() string   { return "bottom-board-elimination" }
func (*BottomBoardElimination) Name() string { return "Bottom Board Elimination" }
func (*BottomBoardElimination) Compute(_ context.Context, state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) ([]chesspairing.TieBreakValue, error) {
	totals := teamBoardTotals(state, scores)
	digits := make(map[string][]float64, len(scores))
	for _, score := range scores {
		boards := totals[score.PlayerID]
		for board := len(boards) - 1; board > 0; board-- {
			var points float64
			for _, boardPoints := range boards[:board] {
				points += boardPoints
			}
			digits[score.PlayerID] = append(digits[score.PlayerID], points)
		}
	}
	return encodeBoardDigits(scores, digits), nil
}

// encodeBoardDigits turns each participant's sequence of board values into one
// number that orders like the sequences compared element by element (art. 12).
// Values are counted in half points, so every digit is a whole number, and the
// base is one more than the largest digit of any participant. Sequences of
// different length are padded with zeros at the end.
func encodeBoardDigits(scores []chesspairing.PlayerScore, digits map[string][]float64) []chesspairing.TieBreakValue {
	var largest float64
	length := 0
	for _, values := range digits {
		length = max(length, len(values))
		for _, value := range values {
			largest = max(largest, math.Round(2*value))
		}
	}
	base := largest + 1
	out := make([]chesspairing.TieBreakValue, len(scores))
	for i, score := range scores {
		values := digits[score.PlayerID]
		for position := range length {
			var digit float64
			if position < len(values) {
				digit = math.Round(2 * values[position])
			}
			out[i].Value = out[i].Value*base + digit
		}
		out[i].PlayerID = score.PlayerID
	}
	return out
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
	matchDraw     float64
	gameDraw      float64
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
		matchDraw:     *opts.PointMatchDraw,
		gameDraw:      *opts.PointDraw * float64(boardCount),
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

func teamDummyValue(id string, record teamOpponentRecord, kind teamScore, table teamOpponentTable) float64 {
	value := table.matchPoints[id]
	ceiling := table.matchDraw * float64(table.totalRounds)
	if kind == teamGame {
		value = table.gamePoints[id]
		ceiling = table.gameDraw * float64(table.totalRounds)
	}
	if record.Category == ForfeitWin || record.Category == ForfeitLoss {
		ceiling = teamAdjustedValue(record.OpponentID, kind, table)
	}
	if value > ceiling {
		return ceiling
	}
	return value
}

func teamAdjustedValue(id string, kind teamScore, table teamOpponentTable) float64 {
	if kind == teamGame {
		return table.adjustedGame[id]
	}
	return table.adjustedMatch[id]
}

func teamBuchholzContributions(id string, table teamOpponentTable) []tieBreakContribution {
	return teamBuchholzContributionsFor(id, table, teamMatch)
}

func teamBuchholzContributionsFor(id string, table teamOpponentTable, kind teamScore) []tieBreakContribution {
	contributions := make([]tieBreakContribution, 0, table.totalRounds)
	records := table.records[id]
	for _, record := range records {
		var opponent float64
		switch {
		case record.Played:
			opponent = teamAdjustedValue(record.OpponentID, kind, table)
		case record.Category == ForfeitWin || record.Category == ForfeitLoss:
			if table.roundRobin {
				opponent = teamAdjustedValue(record.OpponentID, kind, table)
				contributions = append(contributions, tieBreakContribution{value: opponent, significance: opponent})
				continue
			}
			opponent = teamDummyValue(id, record, kind, table)
		case record.Category != None:
			opponent = teamDummyValue(id, record, kind, table)
		default:
			continue
		}
		contributions = append(contributions, tieBreakContribution{value: opponent, significance: opponent, vur: record.IsVUR})
	}
	return contributions
}

func teamSonnebornBergerContributions(id string, table teamOpponentTable, opponent, scored teamScore) []tieBreakContribution {
	contributions := make([]tieBreakContribution, 0, table.totalRounds)
	records := table.records[id]
	for _, record := range records {
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
			opp = teamDummyValue(id, record, opponent, table)
			own = record.value(scored)
		case record.Category != None:
			opp = teamDummyValue(id, record, opponent, table)
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
		for _, bye := range round.TeamByes {
			if _, ok := out[bye.PlayerID]; !ok {
				continue
			}
			if bye.Type == chesspairing.ByePAB || bye.Type == chesspairing.ByeFullPoint {
				points := make([]float64, team.BoardCount(state, opts))
				for i := range points {
					points[i] = *opts.PointWin
				}
				out[bye.PlayerID] = append(out[bye.PlayerID], points)
			}
		}
	}
	return out
}

func teamBoardTotals(state *chesspairing.TournamentState, scores []chesspairing.PlayerScore) map[string][]float64 {
	slices := teamBoardPointSlices(state, scores)
	totals := make(map[string][]float64, len(slices))

	for id, matches := range slices {
		for _, match := range matches {
			if len(totals[id]) < len(match) {
				totals[id] = append(totals[id], make([]float64, len(match)-len(totals[id]))...)
			}
			for board, points := range match {
				totals[id][board] += points
			}
		}
	}
	return totals
}

// (teamBoardCount removed, using team.BoardCount instead)
