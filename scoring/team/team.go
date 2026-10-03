// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

// Package team implements match-point and game-point scoring for team events.
//
// Limitations:
// - It does not merge the two primaryScore option sources (from options map vs explicit).
package team

import (
	"context"
	"fmt"
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
)

// Ensure Scorer implements chesspairing.Scorer.
var _ chesspairing.Scorer = (*Scorer)(nil)

// Options configures team scoring. Match points default to the C.04.6
// 2-1-0 scale. Game points use the standard scoring options.
type Options struct {
	PointMatchWin  *float64
	PointMatchDraw *float64
	PointMatchLoss *float64
	PrimaryScore   string
	BoardCount     *int
	standard.Options
}

// WithDefaults returns options with the C.04.6 match-point defaults.
func (o Options) WithDefaults() Options {
	if o.PointMatchWin == nil {
		o.PointMatchWin = chesspairing.Float64Ptr(2)
	}
	if o.PointMatchDraw == nil {
		o.PointMatchDraw = chesspairing.Float64Ptr(1)
	}
	if o.PointMatchLoss == nil {
		o.PointMatchLoss = chesspairing.Float64Ptr(0)
	}
	if o.PrimaryScore != "game" {
		o.PrimaryScore = "match"
	}
	o.Options = o.Options.WithDefaults()
	return o
}

// ParseOptions converts scoring configuration options into typed options.
func ParseOptions(m map[string]any) Options {
	o := Options{Options: standard.ParseOptions(m)}
	if v, ok := chesspairing.GetFloat64(m, "pointMatchWin"); ok {
		o.PointMatchWin = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "pointMatchDraw"); ok {
		o.PointMatchDraw = &v
	}
	if v, ok := chesspairing.GetFloat64(m, "pointMatchLoss"); ok {
		o.PointMatchLoss = &v
	}
	if v, ok := m["primaryScore"].(string); ok {
		o.PrimaryScore = v
	}
	if v, ok := chesspairing.GetFloat64(m, "boardCount"); ok {
		intV := int(v)
		o.BoardCount = &intV
	}
	return o
}

// Scorer calculates team match and game points.
type Scorer struct {
	opts Options
}

// New creates a team scorer.
func New(opts Options) *Scorer {
	return &Scorer{opts: opts}
}

// NewFromMap creates a team scorer from scoring configuration options.
func NewFromMap(m map[string]any) *Scorer {
	return New(ParseOptions(m))
}

// Score calculates both components for every active team. Score is the
// configured primary component, preserving the scalar Scorer interface.
func (s *Scorer) Score(_ context.Context, state *chesspairing.TournamentState) ([]chesspairing.PlayerScore, error) {
	opts := s.opts.WithDefaults()
	ids := state.ActivePlayerIDs(state.CurrentRound)
	playerTeams := PlayerTeams(state)
	teamRatings := make(map[string]int)
	teams := make(map[string]bool)
	for _, player := range state.Players {
		tid, ok := playerTeams[player.ID]
		if !ok {
			continue
		}
		if player.Rating > teamRatings[tid] {
			teamRatings[tid] = player.Rating
		}
		if state.IsActiveInRound(player.ID, state.CurrentRound) {
			teams[tid] = true
		}
	}
	for _, round := range state.Rounds {
		for _, match := range round.Matches {
			if match.HomeID != "" && match.AwayID != "" &&
				teamIsActive(match.HomeID, state, playerTeams) && teamIsActive(match.AwayID, state, playerTeams) {
				teams[match.HomeID] = true
				teams[match.AwayID] = true
			}
		}
		for _, bye := range round.TeamByes {
			if teamIsActive(bye.PlayerID, state, playerTeams) {
				teams[bye.PlayerID] = true
			}
		}
	}
	if len(teams) > 0 {
		ids = ids[:0]
		for id := range teams {
			ids = append(ids, id)
		}
	}
	points := make(map[string]chesspairing.TeamPoints, len(ids))
	active := make(map[string]bool, len(ids))
	for _, id := range ids {
		active[id] = true
	}
	for _, round := range state.Rounds {
		for _, match := range round.Matches {
			if !active[match.HomeID] || !active[match.AwayID] {
				continue
			}
			homeGame, awayGame := MatchGamePoints(match, opts.Options, playerTeams)
			home := points[match.HomeID]
			away := points[match.AwayID]
			home.Game += homeGame
			away.Game += awayGame
			homeMatch, awayMatch := MatchPoints(match, homeGame, awayGame, opts)
			home.Match += homeMatch
			away.Match += awayMatch
			points[match.HomeID] = home
			points[match.AwayID] = away
		}
		boardCount := BoardCount(state, s.opts)
		if boardCount == 0 && len(round.TeamByes) > 0 {
			return nil, fmt.Errorf("cannot infer board count for team byes")
		}
		for _, bye := range round.TeamByes {
			if !active[bye.PlayerID] {
				continue
			}
			p := points[bye.PlayerID]
			switch bye.Type {
			case chesspairing.ByePAB, chesspairing.ByeHalf:
				p.Match += *opts.PointMatchDraw
				p.Game += *opts.PointDraw * float64(boardCount)
			case chesspairing.ByeFullPoint:
				p.Match += *opts.PointMatchWin
				p.Game += *opts.PointWin * float64(boardCount)
			case chesspairing.ByeZero:
				p.Game += *opts.PointLoss
			case chesspairing.ByeAbsent:
				p.Game += *opts.PointAbsent
			case chesspairing.ByeExcused:
				p.Game += *opts.PointExcused
			case chesspairing.ByeClubCommitment:
				p.Game += *opts.PointClubCommitment
			}
			points[bye.PlayerID] = p
		}
	}
	sort.SliceStable(ids, func(i, j int) bool {
		left, right := points[ids[i]], points[ids[j]]
		if opts.PrimaryScore == "game" && left.Game != right.Game {
			return left.Game > right.Game
		}
		if opts.PrimaryScore == "match" && left.Match != right.Match {
			return left.Match > right.Match
		}
		// Teams with equal primary scores are ordered by rating, then by their
		// identifier, matching the stable ordering used by other scorers.
		if teamRatings[ids[i]] != teamRatings[ids[j]] {
			return teamRatings[ids[i]] > teamRatings[ids[j]]
		}
		return ids[i] < ids[j]
	})
	out := make([]chesspairing.PlayerScore, len(ids))
	for i, id := range ids {
		p := points[id]
		score := p.Match
		if opts.PrimaryScore == "game" {
			score = p.Game
		}
		out[i] = chesspairing.PlayerScore{PlayerID: id, Score: score, Rank: i + 1, Team: &p}
	}
	return out, nil
}

// PointsForResult returns game points for one board result.
func (s *Scorer) PointsForResult(result chesspairing.GameResult, rctx chesspairing.ResultContext) float64 {
	return standard.New(s.opts.WithDefaults().Options).PointsForResult(result, rctx)
}

// MatchPoints returns the match points for a decided match. Undecided boards
// receive no match points.
func MatchPoints(match chesspairing.MatchData, homeGame, awayGame float64, opts Options) (float64, float64) {
	if MatchIsUndecided(match) {
		return 0, 0
	}
	switch {
	case homeGame > awayGame:
		return *opts.PointMatchWin, *opts.PointMatchLoss
	case homeGame < awayGame:
		return *opts.PointMatchLoss, *opts.PointMatchWin
	default:
		return *opts.PointMatchDraw, *opts.PointMatchDraw
	}
}

// BoardCount returns the board count for PAB calculations. It checks the explicit
// BoardCount option, then played match boards, then team membership, then BoardOrder.
// If it cannot be inferred, it returns 0.
func BoardCount(state *chesspairing.TournamentState, opts Options) int {
	if opts.BoardCount != nil {
		return *opts.BoardCount
	}
	count := 0
	for _, round := range state.Rounds {
		for _, match := range round.Matches {
			if len(match.Boards) > count {
				count = len(match.Boards)
			}
		}
	}
	if count > 0 {
		return count
	}

	// Derive from team membership (max players in a team)
	teamCounts := make(map[string]int)
	for _, p := range state.Players {
		if p.TeamID != "" {
			teamCounts[p.TeamID]++
			if teamCounts[p.TeamID] > count {
				count = teamCounts[p.TeamID]
			}
		}
	}
	if count > 0 {
		return count
	}

	// Derive from BoardOrder if present (not implemented in PlayerEntry directly, but max team count handles it if players are registered).
	// If no clue at all, return 0 (refuse to score).
	return 0
}

// MatchGamePoints returns the board-point totals for a match. playerTeams maps
// board players to teams and lets board colours vary independently of team colour.
func MatchGamePoints(match chesspairing.MatchData, options standard.Options, playerTeams map[string]string) (float64, float64) {
	if match.Result != nil && !matchBoardsDecided(match) {
		return match.Result.HomeGame, match.Result.AwayGame
	}
	opts := options.WithDefaults()
	var home, away float64
	for _, board := range match.Boards {
		white, black := BoardPoints(board, opts)
		whiteHome := BelongsToTeam(board.WhiteID, match.HomeID, playerTeams)
		blackHome := BelongsToTeam(board.BlackID, match.HomeID, playerTeams)
		if whiteHome || !blackHome {
			home += white
			away += black
		} else {
			home += black
			away += white
		}
	}
	return home, away
}

// matchBoardsDecided reports whether a match has at least one board and every
// board is decided. Pending and double-forfeit boards leave its totals incomplete.
func matchBoardsDecided(match chesspairing.MatchData) bool {
	if len(match.Boards) == 0 {
		return false
	}
	for _, board := range match.Boards {
		if board.Result == chesspairing.ResultPending || board.Result.IsDoubleForfeit() {
			return false
		}
	}
	return true
}

// BelongsToTeam reports whether playerID belongs to teamID.
func BelongsToTeam(playerID, teamID string, playerTeams map[string]string) bool {
	if playerTeam, ok := playerTeams[playerID]; ok {
		return playerTeam == teamID
	}
	return playerID == teamID
}

func teamIsActive(teamID string, state *chesspairing.TournamentState, playerTeams map[string]string) bool {
	for _, player := range state.Players {
		if (player.ID == teamID || playerTeams[player.ID] == teamID) && state.IsActiveInRound(player.ID, state.CurrentRound) {
			return true
		}
	}
	return false
}

// MatchIsUndecided reports whether a match has neither reported totals nor a
// decided board. Double forfeits do not decide a match.
func MatchIsUndecided(match chesspairing.MatchData) bool {
	if match.Result != nil {
		return false
	}
	if len(match.Boards) == 0 {
		return true
	}
	hasDecided := false
	for _, board := range match.Boards {
		if board.Result != chesspairing.ResultPending && !board.Result.IsDoubleForfeit() {
			hasDecided = true
		}
	}
	return !hasDecided
}

// BoardPoints returns the points awarded to White and Black for one board.
func BoardPoints(board chesspairing.GameData, opts standard.Options) (float64, float64) {
	if board.Result.IsDoubleForfeit() || board.Result == chesspairing.ResultPending {
		return 0, 0
	}
	if board.IsForfeit || board.Result.IsForfeit() {
		switch board.Result {
		case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
			return *opts.PointForfeitWin, *opts.PointForfeitLoss
		case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
			return *opts.PointForfeitLoss, *opts.PointForfeitWin
		default:
			return 0, 0
		}
	}
	switch board.Result {
	case chesspairing.ResultWhiteWins:
		return *opts.PointWin, *opts.PointLoss
	case chesspairing.ResultBlackWins:
		return *opts.PointLoss, *opts.PointWin
	case chesspairing.ResultDraw:
		return *opts.PointDraw, *opts.PointDraw
	default:
		return 0, 0
	}
}

// PlayerTeams maps player IDs to team IDs. When no player has a TeamID, each
// player represents a team with its own ID. In a roster-based team event,
// players without a TeamID are not teams.
func PlayerTeams(state *chesspairing.TournamentState) map[string]string {
	playerTeams := make(map[string]string, len(state.Players))
	hasTeamIDs := false
	for _, player := range state.Players {
		if player.TeamID != "" {
			hasTeamIDs = true
			break
		}
	}
	for _, player := range state.Players {
		if player.TeamID != "" {
			playerTeams[player.ID] = player.TeamID
		} else if !hasTeamIDs {
			playerTeams[player.ID] = player.ID
		}
	}
	return playerTeams
}
