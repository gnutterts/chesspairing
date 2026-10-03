// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package team

import (
	"sort"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/pairing/lexswiss"
	scoringteam "github.com/gnutterts/chesspairing/scoring/team"
)

// buildParticipantStates applies C.04.6 Article 1.2: teams are grouped by
// the configured primary score. Games remain accepted for legacy team states;
// each distinct pair of teams in a round is one match.
func buildParticipantStates(state *chesspairing.TournamentState, primaryScore string) ([]lexswiss.ParticipantState, error) {
	participants, err := lexswiss.BuildParticipantStates(state)
	if err != nil {
		return nil, err
	}
	playerTeams := make(map[string]string, len(state.Players))
	teamParticipants := make(map[string]lexswiss.ParticipantState)
	for _, participant := range participants {
		teamID := participant.ID
		for _, player := range state.Players {
			if player.ID == participant.ID && player.TeamID != "" {
				teamID = player.TeamID
			}
		}
		playerTeams[participant.ID] = teamID
		team, found := teamParticipants[teamID]
		if !found || participant.Rating > team.Rating {
			participant.ID = teamID
			teamParticipants[teamID] = participant
		}
	}
	participants = participants[:0]
	teamIDs := make([]string, 0, len(teamParticipants))
	for teamID := range teamParticipants {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Strings(teamIDs)
	for _, teamID := range teamIDs {
		participants = append(participants, teamParticipants[teamID])
	}
	points := make(map[string]chesspairing.TeamPoints, len(participants))
	beforeLast := make(map[string]float64, len(participants))
	lastOpponents := make(map[string][]string, len(participants))
	pabIneligibility := make(map[string]lexswiss.PABIneligibility, len(participants))
	for _, p := range participants {
		pabIneligibility[p.ID] = p.PABIneligible
	}
	options := scoringteam.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	historyEnd := state.CurrentRound - 1
	if historyEnd < 0 || historyEnd > len(state.Rounds) {
		historyEnd = len(state.Rounds)
	}
	for ri := 0; ri < historyEnd; ri++ {
		if ri == historyEnd-1 {
			for id, value := range points {
				beforeLast[id] = primary(value, primaryScore)
			}
		}
		round := state.Rounds[ri]
		lastOpponents = make(map[string][]string, len(participants))
		if len(round.Matches) > 0 {
			for _, match := range round.Matches {
				home, away := scoringteam.MatchGamePoints(match, options.Options, playerTeams)
				addMatchPoints(points, match, home, away, options)
				if !matchIsForfeit(match) {
					lastOpponents[match.HomeID] = []string{match.AwayID}
					lastOpponents[match.AwayID] = []string{match.HomeID}
				} else {
					homeMatch, awayMatch := scoringteam.MatchPoints(match, home, away, options)
					if homeMatch > awayMatch {
						ineligible := pabIneligibility[match.HomeID]
						ineligible.FullPointUnplayed = true
						pabIneligibility[match.HomeID] = ineligible
					} else if awayMatch > homeMatch {
						ineligible := pabIneligibility[match.AwayID]
						ineligible.FullPointUnplayed = true
						pabIneligibility[match.AwayID] = ineligible
					}
				}
			}
		} else {
			for _, match := range gameMatches(round.Games, playerTeams) {
				home, away := scoringteam.MatchGamePoints(match, options.Options, playerTeams)
				addMatchPoints(points, match, home, away, options)
				if !matchIsForfeit(match) {
					lastOpponents[match.HomeID] = []string{match.AwayID}
					lastOpponents[match.AwayID] = []string{match.HomeID}
				} else {
					homeMatch, awayMatch := scoringteam.MatchPoints(match, home, away, options)
					if homeMatch > awayMatch {
						ineligible := pabIneligibility[match.HomeID]
						ineligible.FullPointUnplayed = true
						pabIneligibility[match.HomeID] = ineligible
					} else if awayMatch > homeMatch {
						ineligible := pabIneligibility[match.AwayID]
						ineligible.FullPointUnplayed = true
						pabIneligibility[match.AwayID] = ineligible
					}
				}
			}
		}
		var teamByes []chesspairing.ByeEntry
		teamByes = append(teamByes, round.TeamByes...)
		for _, bye := range round.Byes {
			if _, ok := teamParticipants[bye.PlayerID]; ok {
				teamByes = append(teamByes, bye)
			}
		}
		for _, bye := range teamByes {
			p := points[bye.PlayerID]
			switch bye.Type {
			case chesspairing.ByePAB, chesspairing.ByeHalf:
				p.Match += *options.PointMatchDraw
				p.Game += *options.PointDraw * float64(scoringteam.BoardCount(state, options))
			case chesspairing.ByeFullPoint:
				p.Match += *options.PointMatchWin
				p.Game += *options.PointWin * float64(scoringteam.BoardCount(state, options))
			}
			points[bye.PlayerID] = p
			ineligible := pabIneligibility[bye.PlayerID]
			switch bye.Type {
			case chesspairing.ByePAB:
				ineligible.PriorPAB = true
			case chesspairing.ByeFullPoint:
				ineligible.FullPointUnplayed = true
			}
			pabIneligibility[bye.PlayerID] = ineligible
		}
	}
	for i := range participants {
		participants[i].Score = primary(points[participants[i].ID], primaryScore)
		participants[i].WasFloater = wasFloater(participants[i].ID, lastOpponents, beforeLast)
		participants[i].PABIneligible = pabIneligibility[participants[i].ID]
	}
	addMatchHistory(participants, state, historyEnd, playerTeams)
	// Score and TPN must be recalculated after replacing the standard score.
	return lexswiss.SortParticipants(participants), nil
}

func gameMatches(games []chesspairing.GameData, playerTeams map[string]string) []chesspairing.MatchData {
	matches := make(map[[2]string]chesspairing.MatchData)
	order := make([][2]string, 0, len(games))
	for _, game := range games {
		wTeam := playerTeams[game.WhiteID]
		if wTeam == "" {
			wTeam = game.WhiteID
		}
		bTeam := playerTeams[game.BlackID]
		if bTeam == "" {
			bTeam = game.BlackID
		}
		key := [2]string{wTeam, bTeam}
		if bTeam < wTeam {
			key = [2]string{bTeam, wTeam}
		}
		match, found := matches[key]
		if !found {
			match = chesspairing.MatchData{HomeID: wTeam, AwayID: bTeam}
			order = append(order, key)
		}
		match.Boards = append(match.Boards, game)
		matches[key] = match
	}
	out := make([]chesspairing.MatchData, 0, len(order))
	for _, key := range order {
		out = append(out, matches[key])
	}
	return out
}

func addMatchPoints(points map[string]chesspairing.TeamPoints, match chesspairing.MatchData, homeGame, awayGame float64, options scoringteam.Options) {
	home, away := points[match.HomeID], points[match.AwayID]
	home.Game += homeGame
	away.Game += awayGame
	homeMatch, awayMatch := scoringteam.MatchPoints(match, homeGame, awayGame, options)
	home.Match += homeMatch
	away.Match += awayMatch
	points[match.HomeID], points[match.AwayID] = home, away
}

func matchIsForfeit(match chesspairing.MatchData) bool {
	return len(match.Boards) > 0 && allForfeit(match.Boards)
}

func allForfeit(boards []chesspairing.GameData) bool {
	for _, board := range boards {
		if !board.IsForfeit && !board.Result.IsForfeit() {
			return false
		}
	}
	return true
}

func primary(points chesspairing.TeamPoints, primaryScore string) float64 {
	if primaryScore == "game" {
		return points.Game
	}
	return points.Match
}

func addMatchHistory(participants []lexswiss.ParticipantState, state *chesspairing.TournamentState, historyEnd int, playerTeams map[string]string) {
	byID := make(map[string]*lexswiss.ParticipantState, len(participants))
	for i := range participants {
		// Clear per-board history added by lexswiss.BuildParticipantStates
		participants[i].Opponents = nil
		participants[i].ColorHistory = nil
		byID[participants[i].ID] = &participants[i]
	}
	for ri := 0; ri < historyEnd; ri++ {
		matches := state.Rounds[ri].Matches
		if len(matches) == 0 {
			matches = gameMatches(state.Rounds[ri].Games, playerTeams)
		}
		for _, match := range matches {
			if matchIsForfeit(match) {
				continue
			}
			home, homeOK := byID[match.HomeID]
			away, awayOK := byID[match.AwayID]
			if homeOK {
				home.Opponents = append(home.Opponents, match.AwayID)
				home.ColorHistory = append(home.ColorHistory, lexswiss.ColorWhite)
			}
			if awayOK {
				away.Opponents = append(away.Opponents, match.HomeID)
				away.ColorHistory = append(away.ColorHistory, lexswiss.ColorBlack)
			}
		}
	}
}

func wasFloater(id string, opponents map[string][]string, scores map[string]float64) bool {
	for _, opponent := range opponents[id] {
		if scores[id] != scores[opponent] {
			return true
		}
	}
	return false
}
