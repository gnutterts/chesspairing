// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
	scoringTeam "github.com/gnutterts/chesspairing/scoring/team"
)

// ToTournamentState converts a Document to a TournamentState for engine use.
// Player IDs are set to the string representation of start numbers (e.g. "1", "2").
// RoundData is reconstructed by cross-referencing per-player round results.
func (doc *Document) ToTournamentState() (*chesspairing.TournamentState, error) {
	state := &chesspairing.TournamentState{}

	// Convert players.
	state.Players = make([]chesspairing.PlayerEntry, len(doc.Players))
	for i, pl := range doc.Players {
		state.Players[i] = chesspairing.PlayerEntry{
			ID:            strconv.Itoa(pl.StartNumber),
			DisplayName:   pl.Name,
			Rating:        pl.Rating,
			PairingNumber: pl.StartNumber,
			Federation:    pl.Federation,
			FideID:        pl.FideID,
			Title:         pl.Title,
			Sex:           pl.Sex,
			BirthDate:     pl.BirthDate,
		}
	}

	// Team records link board players to their team identifiers.
	for _, team := range doc.Teams {
		for _, member := range team.Members {
			for i := range state.Players {
				if state.Players[i].PairingNumber == member {
					state.Players[i].TeamID = strconv.Itoa(team.TeamNumber)
					break
				}
			}
		}
	}

	// Index players by start number so opponent round results can be
	// cross-referenced when building games.
	playerIdx := make(map[int]int, len(doc.Players))
	for i, pl := range doc.Players {
		playerIdx[pl.StartNumber] = i
	}

	teamSystem := inferPairingSystem(doc.TournamentType) == chesspairing.PairingTeam

	// Determine number of rounds from player data.
	maxRounds := 0
	for _, pl := range doc.Players {
		if len(pl.Rounds) > maxRounds {
			maxRounds = len(pl.Rounds)
		}
	}
	for _, record := range doc.DetailedTeamResults {
		if len(record.Rounds) > maxRounds {
			maxRounds = len(record.Rounds)
		}
	}
	for _, record := range doc.SimpleTeamResults {
		if len(record.Rounds) > maxRounds {
			maxRounds = len(record.Rounds)
		}
	}
	// A 320 PAB record only contributes rounds for team tournaments; a stray
	// 320 in an individual file must not extend the tournament.
	if teamSystem {
		for _, record := range doc.TeamPABs {
			if len(record.RoundTeams) > maxRounds {
				maxRounds = len(record.RoundTeams)
			}
		}
	}

	// Build rounds by cross-referencing player data.
	state.Rounds = make([]chesspairing.RoundData, maxRounds)
	for roundIdx := range maxRounds {
		rd := chesspairing.RoundData{Number: roundIdx + 1}
		seen := make(map[string]bool) // track processed games to avoid duplicates

		for _, pl := range doc.Players {
			if roundIdx >= len(pl.Rounds) {
				continue
			}
			rr := pl.Rounds[roundIdx]
			playerID := strconv.Itoa(pl.StartNumber)

			// Bye results -> ByeEntry
			if rr.Result.isByeResult() {
				var bt chesspairing.ByeType
				switch rr.Result {
				case ResultFullBye:
					bt = chesspairing.ByeFullPoint
				case ResultHalfBye:
					bt = chesspairing.ByeHalf
				case ResultZeroBye:
					bt = chesspairing.ByeZero
				case ResultUnpaired:
					bt = chesspairing.ByePAB
				}
				rd.Byes = append(rd.Byes, chesspairing.ByeEntry{
					PlayerID: playerID,
					Type:     bt,
				})
				continue
			}

			// Skip if not-yet-played.
			if rr.Result == ResultNotPlayed {
				continue
			}

			oppID := strconv.Itoa(rr.Opponent)

			// Avoid duplicate games: only process from one player's perspective.
			gameKey := playerID + "-" + oppID
			reverseKey := oppID + "-" + playerID
			if seen[gameKey] || seen[reverseKey] {
				continue
			}

			var whiteID, blackID string
			if rr.Color == ColorWhite {
				whiteID = playerID
				blackID = oppID
			} else {
				whiteID = oppID
				blackID = playerID
			}

			result := convertResultToGameResult(rr.Result, rr.Color)
			isForfeit := rr.Result == ResultForfeitWin || rr.Result == ResultForfeitLoss ||
				rr.Result == ResultWinByDefault || rr.Result == ResultDrawByDefault || rr.Result == ResultLossByDefault

			// When both opponents reference each other for the same round,
			// cross-check their results. A double forfeit is encoded as a
			// forfeit loss on both 001 lines; recognise it as such instead
			// of silently crediting one side with a forfeit win. Mutually
			// inconsistent results (for example two wins or two forfeit
			// wins) are reported as an error rather than resolved in favour
			// of the first player seen.
			if oppIdx, ok := playerIdx[rr.Opponent]; ok && roundIdx < len(doc.Players[oppIdx].Rounds) {
				oppRR := doc.Players[oppIdx].Rounds[roundIdx]
				if oppRR.Opponent == pl.StartNumber {
					switch {
					case rr.Result == ResultForfeitLoss && oppRR.Result == ResultForfeitLoss:
						result = chesspairing.ResultDoubleForfeit
						isForfeit = true
					case !areResultsConsistent(rr.Result, oppRR.Result):
						return nil, fmt.Errorf("trf: inconsistent results for players %d and %d in round %d",
							pl.StartNumber, rr.Opponent, roundIdx+1)
					}
				}
			}

			rd.Games = append(rd.Games, chesspairing.GameData{
				WhiteID:   whiteID,
				BlackID:   blackID,
				Result:    result,
				IsForfeit: isForfeit,
			})
			seen[gameKey] = true
			seen[reverseKey] = true
		}

		state.Rounds[roundIdx] = rd
	}
	if teamSystem {
		// Team 801/802 records use TeamByes. Do not remove the player-keyed
		// Byes: team and player start numbers share a namespace.
		buildTeamMatches(doc, state.Rounds)
		bridgeTeamPABs(doc.TeamPABs, state.Rounds)
	}

	state.CurrentRound = maxRounds + 1

	// Tournament info.
	state.Info = chesspairing.TournamentInfo{
		Name:          doc.Name,
		City:          doc.City,
		Federation:    doc.Federation,
		StartDate:     doc.StartDate,
		EndDate:       doc.EndDate,
		ChiefArbiter:  doc.ChiefArbiter,
		DeputyArbiter: doc.DeputyArbiter,
		TimeControl:   doc.TimeControl,
		RoundDates:    doc.RoundDates,
	}

	// Pairing config.
	state.PairingConfig = chesspairing.PairingConfig{
		System:  inferPairingSystem(doc.TournamentType),
		Options: make(map[string]any),
	}
	if tr := doc.EffectiveTotalRounds(); tr > 0 {
		state.PairingConfig.Options["totalRounds"] = tr
	}
	if ic := doc.EffectiveInitialColor(); ic != "" {
		topSeedColor, err := normalizeInitialColor(ic)
		if err != nil {
			return nil, fmt.Errorf("trf: %w", err)
		}
		state.PairingConfig.Options["topSeedColor"] = topSeedColor
	}
	// Legacy XXP forbidden pairs are permanent: they apply in every round.
	var forbiddenPairs [][2]int
	for _, fp := range doc.ForbiddenPairs {
		forbiddenPairs = append(forbiddenPairs, [2]int{fp.Player1, fp.Player2})
	}
	// TRF-2026 forbidden pair records (260) are round-scoped. They only
	// contribute pairs when the round about to be paired falls inside
	// FirstRound..LastRound; a zero bound means no limit on that side.
	// Each 260 record lists mutually forbidden players; generate all pairs.
	for _, fp := range doc.ForbiddenPairs26 {
		if fp.FirstRound != 0 && state.CurrentRound < fp.FirstRound {
			continue
		}
		if fp.LastRound != 0 && state.CurrentRound > fp.LastRound {
			continue
		}
		for i := 0; i < len(fp.Players); i++ {
			for j := i + 1; j < len(fp.Players); j++ {
				forbiddenPairs = append(forbiddenPairs, [2]int{fp.Players[i], fp.Players[j]})
			}
		}
	}
	if len(forbiddenPairs) > 0 {
		state.PairingConfig.Options["forbiddenPairs"] = forbiddenPairs
	}
	// Acceleration from XXS lines.
	if len(doc.Acceleration) > 0 || bakuCodedType(doc.CodedTournamentType) {
		state.PairingConfig.Options["acceleration"] = "baku"
	}
	// Round-Robin options.
	if doc.Cycles > 0 {
		state.PairingConfig.Options["cycles"] = doc.Cycles
	}
	if doc.ColorBalance != nil {
		state.PairingConfig.Options["colorBalance"] = *doc.ColorBalance
	}
	// Lim options.
	if doc.MaxiTournament != nil {
		state.PairingConfig.Options["maxiTournament"] = *doc.MaxiTournament
	}
	// Team options.
	if doc.ColorPreferenceType != "" {
		state.PairingConfig.Options["colorPreferenceType"] = doc.ColorPreferenceType
	}
	if doc.PrimaryScore != "" {
		state.PairingConfig.Options["primaryScore"] = doc.PrimaryScore
	}
	// Keizer options.
	if doc.AllowRepeatPairings != nil {
		state.PairingConfig.Options["allowRepeatPairings"] = *doc.AllowRepeatPairings
	}
	if doc.MinRoundsBetweenRepeats > 0 {
		state.PairingConfig.Options["minRoundsBetweenRepeats"] = doc.MinRoundsBetweenRepeats
	}

	// Scoring config: standard by default; a TRF-2026 162 record populates
	// the options with the scoring points the standard scorer reads.
	scoringSystem := chesspairing.ScoringStandard
	if state.PairingConfig.System == chesspairing.PairingTeam {
		scoringSystem = chesspairing.ScoringTeam
	}
	state.ScoringConfig = chesspairing.ScoringConfig{
		System:      scoringSystem,
		Tiebreakers: chesspairing.DefaultTiebreakers(state.PairingConfig.System),
	}
	if doc.ScoringSystem != nil || doc.PrimaryScore != "" {
		opts := make(map[string]any)
		if doc.ScoringSystem != nil && doc.ScoringSystem.W != nil {
			opts["pointWin"] = *doc.ScoringSystem.W
		}
		if doc.ScoringSystem != nil && doc.ScoringSystem.D != nil {
			opts["pointDraw"] = *doc.ScoringSystem.D
		}
		if doc.ScoringSystem != nil && doc.ScoringSystem.L != nil {
			opts["pointLoss"] = *doc.ScoringSystem.L
		}
		if doc.ScoringSystem != nil && doc.ScoringSystem.A != nil {
			opts["pointAbsent"] = *doc.ScoringSystem.A
		}
		// P is the full-point bye (PAB) value; scoring/standard reads it as
		// pointBye. X is an unknown result (for instance an adjourned game),
		// which TRF-2026 scores the same as a draw; standard has no dedicated
		// key for it, so X is intentionally left unmapped.
		if doc.ScoringSystem != nil && doc.ScoringSystem.P != nil {
			opts["pointBye"] = *doc.ScoringSystem.P
		}
		if doc.PrimaryScore != "" {
			opts["primaryScore"] = doc.PrimaryScore
		}
		if len(opts) > 0 {
			state.ScoringConfig.Options = opts
		}
	}

	// Bridge Section 240 absence records and chesspairing:bye directives
	// into PreAssignedByes for the upcoming round. Section 240 only carries
	// "F" (pairing-allocated bye in this section) and "H" (half) per the FIDE
	// spec; richer bye types arrive
	// via chesspairing directives. When both refer to the same player in the
	// same round the directive wins, since it is the more specific source.
	if err := bridgePreAssignedByes(doc, state); err != nil {
		return nil, err
	}
	if err := bridgeWithdrawnDirectives(doc, state); err != nil {
		return nil, err
	}

	return state, nil
}

// buildTeamMatches reconstructs team matches from TRF-2026 801 and 802
// records. Detailed records provide board results; simple records preserve the
// reported game-point totals when boards are not available.
func buildTeamMatches(doc *Document, rounds []chesspairing.RoundData) {
	seen := make(map[string]bool)
	for _, record := range doc.DetailedTeamResults {
		for roundIndex, entry := range record.Rounds {
			if roundIndex >= len(rounds) {
				continue
			}
			if entry.Opponent == 0 {
				if byeType, ok := detailedByeType(entry.ByeType); ok {
					rounds[roundIndex].TeamByes = append(rounds[roundIndex].TeamByes, chesspairing.ByeEntry{PlayerID: strconv.Itoa(record.TeamNumber), Type: byeType})
					seen[teamByeKey(roundIndex, record.TeamNumber)] = true
				}
				continue
			}
			key := teamMatchKey(roundIndex, record.TeamNumber, entry.Opponent)
			if seen[key] {
				continue
			}
			home, away := record.TeamNumber, entry.Opponent
			if entry.Color != "w" && entry.Color != "W" {
				home, away = away, home
			}
			match := chesspairing.MatchData{HomeID: strconv.Itoa(home), AwayID: strconv.Itoa(away)}
			homeMembers := orderedTeamMembers(doc.Teams, home, entry.BoardOrder)
			awayMembers := orderedTeamMembers(doc.Teams, away, entry.BoardOrder)
			for board, result := range entry.Results {
				if entry.Color != "w" && entry.Color != "W" {
					switch result {
					case '1':
						result = '0'
					case '0':
						result = '1'
					}
				}
				homeWhite := board%2 == 0
				game := chesspairing.GameData{WhiteID: match.AwayID, BlackID: match.HomeID, Result: chesspairing.ResultPending}
				if homeWhite {
					game.WhiteID, game.BlackID = match.HomeID, match.AwayID
				}
				if board < len(homeMembers) && board < len(awayMembers) {
					if homeWhite {
						game.WhiteID, game.BlackID = strconv.Itoa(homeMembers[board]), strconv.Itoa(awayMembers[board])
					} else {
						game.WhiteID, game.BlackID = strconv.Itoa(awayMembers[board]), strconv.Itoa(homeMembers[board])
					}
				}
				switch result {
				case '1':
					if homeWhite {
						game.Result = chesspairing.ResultWhiteWins
					} else {
						game.Result = chesspairing.ResultBlackWins
					}
				case '0':
					if homeWhite {
						game.Result = chesspairing.ResultBlackWins
					} else {
						game.Result = chesspairing.ResultWhiteWins
					}
				case '=':
					game.Result = chesspairing.ResultDraw
				}
				match.Boards = append(match.Boards, game)
			}
			boardsComplete := len(match.Boards) > 0
			for _, board := range match.Boards {
				if board.Result == chesspairing.ResultPending || board.Result.IsDoubleForfeit() {
					boardsComplete = false
					break
				}
			}
			if !boardsComplete {
				homeGame, homeKnown := teamRoundGamePointsOK(doc.SimpleTeamResults, home, roundIndex)
				awayGame, awayKnown := teamRoundGamePointsOK(doc.SimpleTeamResults, away, roundIndex)
				if !homeKnown || !awayKnown {
					homeGame, awayGame = teamMatchPoints(match, scoringTeam.ParseOptions(nil).Options, teamPlayerTeams(doc.Teams))
				}
				match.Result = &chesspairing.TeamMatchResult{HomeGame: homeGame, AwayGame: awayGame}
			}
			rounds[roundIndex].Matches = append(rounds[roundIndex].Matches, match)
			seen[key] = true
		}
	}
	for _, record := range doc.SimpleTeamResults {
		for roundIndex, entry := range record.Rounds {
			if roundIndex >= len(rounds) {
				continue
			}
			if entry.Opponent == 0 {
				byeKey := teamByeKey(roundIndex, record.TeamNumber)
				if byeType, ok := simpleByeType(entry.ByeType); ok && !seen[byeKey] {
					rounds[roundIndex].TeamByes = append(rounds[roundIndex].TeamByes, chesspairing.ByeEntry{PlayerID: strconv.Itoa(record.TeamNumber), Type: byeType})
				}
				continue
			}
			key := teamMatchKey(roundIndex, record.TeamNumber, entry.Opponent)
			if seen[key] {
				continue
			}
			home, away := record.TeamNumber, entry.Opponent
			if entry.Color != "w" && entry.Color != "W" {
				home, away = away, home
			}
			homeGame := teamRoundGamePoints(doc.SimpleTeamResults, home, roundIndex)
			awayGame := teamRoundGamePoints(doc.SimpleTeamResults, away, roundIndex)
			rounds[roundIndex].Matches = append(rounds[roundIndex].Matches, chesspairing.MatchData{
				HomeID: strconv.Itoa(home),
				AwayID: strconv.Itoa(away),
				Result: &chesspairing.TeamMatchResult{HomeGame: homeGame, AwayGame: awayGame},
			})
			seen[key] = true
		}
	}
}

// bridgeTeamPABs adds the teams assigned a pairing-allocated bye to their rounds.
func bridgeTeamPABs(records []TeamPABRecord, rounds []chesspairing.RoundData) {
	for _, record := range records {
		for roundIndex, team := range record.RoundTeams {
			if team == 0 || roundIndex >= len(rounds) {
				continue
			}
			if teamBusyInRound(rounds[roundIndex], team) {
				continue
			}
			rounds[roundIndex].TeamByes = append(rounds[roundIndex].TeamByes, chesspairing.ByeEntry{
				PlayerID: strconv.Itoa(team),
				Type:     chesspairing.ByePAB,
			})
		}
	}
}

// teamBusyInRound reports whether the team already has a bye or a match in the
// round, so a 320 PAB record never invents a second result for it.
func teamBusyInRound(round chesspairing.RoundData, team int) bool {
	id := strconv.Itoa(team)
	for _, bye := range round.TeamByes {
		if bye.PlayerID == id {
			return true
		}
	}
	for _, match := range round.Matches {
		if match.HomeID == id || match.AwayID == id {
			return true
		}
	}
	return false
}

func teamMembers(teams []TeamLine, teamNumber int) []int {
	for _, team := range teams {
		if team.TeamNumber == teamNumber {
			return team.Members
		}
	}
	return nil
}

func orderedTeamMembers(teams []TeamLine, teamNumber int, boardOrder string) []int {
	members := teamMembers(teams, teamNumber)
	if boardOrder == "" {
		return members
	}
	ordered := make([]int, 0, len(boardOrder))
	for _, board := range boardOrder {
		index := int(board - '1')
		if index >= 0 && index < len(members) {
			ordered = append(ordered, members[index])
		}
	}
	if len(ordered) == len(members) {
		return ordered
	}
	return members
}

func simpleByeType(marker string) (chesspairing.ByeType, bool) {
	switch marker {
	case "FPB":
		return chesspairing.ByeFullPoint, true
	case "HPB":
		return chesspairing.ByeHalf, true
	case "ZPB":
		return chesspairing.ByeZero, true
	case "PAB":
		return chesspairing.ByePAB, true
	default:
		return 0, false
	}
}

func detailedByeType(marker string) (chesspairing.ByeType, bool) {
	switch marker {
	case "FFFF":
		return chesspairing.ByeFullPoint, true
	case "HHHH":
		return chesspairing.ByeHalf, true
	case "ZZZZ":
		return chesspairing.ByeZero, true
	case "UUUU":
		return chesspairing.ByePAB, true
	default:
		return 0, false
	}
}

func teamPlayerTeams(teams []TeamLine) map[string]string {
	result := make(map[string]string)
	for _, team := range teams {
		for _, member := range team.Members {
			result[strconv.Itoa(member)] = strconv.Itoa(team.TeamNumber)
		}
	}
	return result
}

func teamByeKey(round, team int) string {
	return "bye:" + strconv.Itoa(round) + ":" + strconv.Itoa(team)
}

func teamMatchKey(round, first, second int) string {
	if first > second {
		first, second = second, first
	}
	return strconv.Itoa(round) + ":" + strconv.Itoa(first) + ":" + strconv.Itoa(second)
}

func teamRoundGamePoints(records []SimpleTeamResult, teamNumber, roundIndex int) float64 {
	points, _ := teamRoundGamePointsOK(records, teamNumber, roundIndex)
	return points
}

func teamRoundGamePointsOK(records []SimpleTeamResult, teamNumber, roundIndex int) (float64, bool) {
	for _, record := range records {
		if record.TeamNumber == teamNumber && roundIndex < len(record.Rounds) {
			return record.Rounds[roundIndex].GamePoints, true
		}
	}
	return 0, false
}

// bridgePreAssignedByes populates state.PreAssignedByes from doc.Absences and
// doc.ChesspairingDirectives, using state.CurrentRound to identify the
// upcoming round. Unknown player IDs in either source are reported as a
// validation error rather than silently dropped.
func bridgePreAssignedByes(doc *Document, state *chesspairing.TournamentState) error {
	if state.CurrentRound == 0 {
		return nil
	}
	known := make(map[string]bool, len(state.Players))
	for _, p := range state.Players {
		known[p.ID] = true
	}
	// Index by player ID so directive entries can override Section 240.
	byPlayer := make(map[string]chesspairing.ByeType)
	order := make([]string, 0)

	add := func(playerID string, bt chesspairing.ByeType, source string) error {
		if !known[playerID] {
			return fmt.Errorf("trf: %s references unknown player %q for round %d",
				source, playerID, state.CurrentRound)
		}
		if _, seen := byPlayer[playerID]; !seen {
			order = append(order, playerID)
		}
		byPlayer[playerID] = bt
		return nil
	}

	for _, a := range doc.Absences {
		if a.Round != state.CurrentRound {
			continue
		}
		bt, ok := byeTypeFromAbsenceCode(a.Type)
		if !ok {
			continue
		}
		for _, sn := range a.Players {
			if err := add(strconv.Itoa(sn), bt, "Section 240 record"); err != nil {
				return err
			}
		}
	}

	for _, d := range doc.ChesspairingDirectives {
		if d.Verb != "bye" {
			continue
		}
		roundStr := d.Params["round"]
		round, err := strconv.Atoi(roundStr)
		if err != nil || round != state.CurrentRound {
			continue
		}
		playerID := d.Params["player"]
		if playerID == "" {
			continue
		}
		bt, ok := byeTypeFromDirectiveString(d.Params["type"])
		if !ok {
			continue
		}
		if err := add(playerID, bt, "chesspairing:bye directive"); err != nil {
			return err
		}
	}

	for _, id := range order {
		state.PreAssignedByes = append(state.PreAssignedByes, chesspairing.ByeEntry{
			PlayerID: id,
			Type:     byPlayer[id],
		})
	}
	return nil
}

// bridgeWithdrawnDirectives populates PlayerEntry.WithdrawnAfterRound from
// `### chesspairing:withdrawn player=<sn> after-round=<N>` directives. Unknown
// player IDs and malformed after-round values are reported as validation
// errors. If the same player appears in multiple directives the latest one
// wins, mirroring the bye bridge's last-write semantics.
func bridgeWithdrawnDirectives(doc *Document, state *chesspairing.TournamentState) error {
	idx := make(map[string]int, len(state.Players))
	for i, p := range state.Players {
		idx[p.ID] = i
	}
	for _, d := range doc.ChesspairingDirectives {
		if d.Verb != "withdrawn" {
			continue
		}
		playerID := d.Params["player"]
		if playerID == "" {
			continue
		}
		i, ok := idx[playerID]
		if !ok {
			return fmt.Errorf("trf: chesspairing:withdrawn directive references unknown player %q", playerID)
		}
		afterStr := d.Params["after-round"]
		after, err := strconv.Atoi(afterStr)
		if err != nil {
			return fmt.Errorf("trf: chesspairing:withdrawn directive for player %q has invalid after-round %q", playerID, afterStr)
		}
		if after <= 0 {
			return fmt.Errorf("trf: chesspairing:withdrawn directive for player %q has non-positive after-round %d", playerID, after)
		}
		v := after
		state.Players[i].WithdrawnAfterRound = &v
	}
	return nil
}

// byeTypeFromAbsenceCode maps a Section 240 type letter to a ByeType. Only
// "F", "H" and "Z" are FIDE-defined; richer types travel via chesspairing
// directives instead.
func byeTypeFromAbsenceCode(code string) (chesspairing.ByeType, bool) {
	switch code {
	case "F":
		return chesspairing.ByePAB, true
	case "H":
		return chesspairing.ByeHalf, true
	case "Z":
		return chesspairing.ByeZero, true
	default:
		return 0, false
	}
}

// byeTypeFromDirectiveString parses the lowercased ByeType.String() spelling
// used in chesspairing:bye directives.
func byeTypeFromDirectiveString(s string) (chesspairing.ByeType, bool) {
	switch strings.ToLower(s) {
	case "pab":
		return chesspairing.ByePAB, true
	case "fullpoint":
		return chesspairing.ByeFullPoint, true
	case "half":
		return chesspairing.ByeHalf, true
	case "zero":
		return chesspairing.ByeZero, true
	case "absent":
		return chesspairing.ByeAbsent, true
	case "excused":
		return chesspairing.ByeExcused, true
	case "clubcommitment":
		return chesspairing.ByeClubCommitment, true
	default:
		return 0, false
	}
}

// byeTypeToDirectiveString returns the lowercased ByeType spelling used on
// chesspairing:bye directives. The empty string signals an unknown type.
func byeTypeToDirectiveString(bt chesspairing.ByeType) string {
	if !bt.IsValid() {
		return ""
	}
	return strings.ToLower(bt.String())
}

// convertResultToGameResult converts a TRF ResultCode + Color to a chesspairing.GameResult.
func convertResultToGameResult(rc ResultCode, color Color) chesspairing.GameResult {
	switch rc {
	case ResultWin, ResultWinByDefault:
		if color == ColorWhite {
			return chesspairing.ResultWhiteWins
		}
		return chesspairing.ResultBlackWins
	case ResultLoss, ResultLossByDefault:
		if color == ColorWhite {
			return chesspairing.ResultBlackWins
		}
		return chesspairing.ResultWhiteWins
	case ResultDraw, ResultDrawByDefault:
		return chesspairing.ResultDraw
	case ResultForfeitWin:
		if color == ColorWhite {
			return chesspairing.ResultForfeitWhiteWins
		}
		return chesspairing.ResultForfeitBlackWins
	case ResultForfeitLoss:
		if color == ColorWhite {
			return chesspairing.ResultForfeitBlackWins
		}
		return chesspairing.ResultForfeitWhiteWins
	default:
		return chesspairing.ResultPending
	}
}

// inferPairingSystem maps a TRF tournament type string to a PairingSystem.
func inferPairingSystem(tournamentType string) chesspairing.PairingSystem {
	switch tournamentType {
	case "Swiss Dutch":
		return chesspairing.PairingDutch
	case "Swiss Burstein":
		return chesspairing.PairingBurstein
	case "Swiss Dubov":
		return chesspairing.PairingDubov
	case "Swiss Lim":
		return chesspairing.PairingLim
	case "Double Swiss":
		return chesspairing.PairingDoubleSwiss
	case "Team Swiss":
		return chesspairing.PairingTeam
	case "Round Robin", "Double Round Robin":
		return chesspairing.PairingRoundRobin
	case "Keizer":
		return chesspairing.PairingKeizer
	default:
		return chesspairing.PairingDutch
	}
}

// FromTournamentState creates a Document from a TournamentState.
// Players are written in PairingNumber order and retain that number as their
// TRF start number. If no pairing numbers have been assigned yet, they are
// assigned once according to FIDE C.04.2 article 2.
//
// Limitations for team tournaments:
// - TRF team names (013/801/802 records) are omitted.
// - Real board-1 colours are not written to 801/802 records.
// - The team set is derived only from PlayerEntry.TeamID; teams appearing only in Matches get no 013/801/802 records.
//
// Returns the Document and a mapping from player ID to start number.
func FromTournamentState(state *chesspairing.TournamentState) (*Document, map[string]int) {
	doc := &Document{}

	// Work on a copy so conversion never mutates the tournament state.
	sorted, err := chesspairing.AssignPairingNumbers(state.Players)
	if err != nil {
		// Invalid numbers (duplicates or partly set): derive fresh, unique
		// numbers from rating, title and name so the TRF stays valid.
		fresh := append([]chesspairing.PlayerEntry(nil), state.Players...)
		for i := range fresh {
			fresh[i].PairingNumber = 0
		}
		sorted, _ = chesspairing.AssignPairingNumbers(fresh)
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].PairingNumber < sorted[j].PairingNumber
	})

	// Preserve pairing numbers as TRF start numbers and build the lookup.
	playerMap := make(map[string]int, len(sorted))
	for _, p := range sorted {
		playerMap[p.ID] = p.PairingNumber
	}

	// Build player lines.
	doc.Players = make([]PlayerLine, len(sorted))
	for i, p := range sorted {
		pl := PlayerLine{
			StartNumber: p.PairingNumber,
			Name:        p.DisplayName,
			Rating:      p.Rating,
			Federation:  p.Federation,
			FideID:      p.FideID,
			Title:       p.Title,
			Sex:         p.Sex,
			BirthDate:   p.BirthDate,
		}

		// Build round results.
		for _, round := range state.Rounds {
			rr := buildRoundResultForPlayer(p.ID, round, playerMap)
			pl.Rounds = append(pl.Rounds, rr)
		}

		// Calculate total points.
		for _, rr := range pl.Rounds {
			pl.Points += pointsForTRFResult(rr.Result)
		}

		doc.Players[i] = pl
	}

	// Assign ranks by points descending.
	type rankEntry struct {
		idx    int
		points float64
	}
	entries := make([]rankEntry, len(doc.Players))
	for i, p := range doc.Players {
		entries[i] = rankEntry{idx: i, points: p.Points}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].points != entries[j].points {
			return entries[i].points > entries[j].points
		}
		return doc.Players[entries[i].idx].StartNumber < doc.Players[entries[j].idx].StartNumber
	})
	for rank, e := range entries {
		doc.Players[e.idx].Rank = rank + 1
	}

	// Tournament info.
	doc.Name = state.Info.Name
	doc.City = state.Info.City
	doc.Federation = state.Info.Federation
	doc.StartDate = state.Info.StartDate
	doc.EndDate = state.Info.EndDate
	doc.ChiefArbiter = state.Info.ChiefArbiter
	doc.DeputyArbiter = state.Info.DeputyArbiter
	doc.TimeControl = state.Info.TimeControl
	doc.RoundDates = state.Info.RoundDates
	doc.NumPlayers = len(state.Players)
	numRated := 0
	for _, p := range state.Players {
		if p.Rating > 0 {
			numRated++
		}
	}
	doc.NumRated = numRated

	// Tournament type from pairing config.
	switch state.PairingConfig.System {
	case chesspairing.PairingDutch:
		doc.TournamentType = "Swiss Dutch"
	case chesspairing.PairingBurstein:
		doc.TournamentType = "Swiss Burstein"
	case chesspairing.PairingDubov:
		doc.TournamentType = "Swiss Dubov"
	case chesspairing.PairingLim:
		doc.TournamentType = "Swiss Lim"
	case chesspairing.PairingDoubleSwiss:
		doc.TournamentType = "Double Swiss"
	case chesspairing.PairingTeam:
		doc.TournamentType = "Team Swiss"
	case chesspairing.PairingRoundRobin:
		cycles := 1
		if opts := state.PairingConfig.Options; opts != nil {
			if v, ok := opts["cycles"]; ok {
				switch c := v.(type) {
				case int:
					cycles = c
				case float64:
					cycles = int(c)
				}
			}
		}
		if cycles >= 2 {
			doc.TournamentType = "Double Round Robin"
		} else {
			doc.TournamentType = "Round Robin"
		}
	case chesspairing.PairingKeizer:
		doc.TournamentType = "Keizer"
	}

	// XX lines from pairing config options.
	if opts := state.PairingConfig.Options; opts != nil {
		if v, ok := opts["totalRounds"]; ok {
			switch tr := v.(type) {
			case int:
				doc.TotalRounds = tr
			case float64:
				doc.TotalRounds = int(tr)
			}
		}
		if v, ok := opts["topSeedColor"].(string); ok {
			switch v {
			case "white":
				doc.InitialColor = "white1"
			case "black":
				doc.InitialColor = "black1"
			}
		}
		if v, ok := opts["forbiddenPairs"]; ok {
			if pairs, ok := v.([][2]int); ok {
				for _, pair := range pairs {
					doc.ForbiddenPairs = append(doc.ForbiddenPairs, ForbiddenPair{
						Player1: pair[0],
						Player2: pair[1],
					})
				}
			}
		}
		// Acceleration (Dutch/Burstein): if "baku", write a marker XXS line.
		if v, ok := opts["acceleration"].(string); ok && v == "baku" {
			if len(doc.Acceleration) == 0 {
				doc.Acceleration = []string{"baku"}
			}
		}
		// Round-Robin options.
		if v, ok := opts["cycles"]; ok {
			switch c := v.(type) {
			case int:
				doc.Cycles = c
			case float64:
				doc.Cycles = int(c)
			}
		}
		if v, ok := opts["colorBalance"]; ok {
			if b, ok := v.(bool); ok {
				doc.ColorBalance = &b
			}
		}
		// Lim options.
		if v, ok := opts["maxiTournament"]; ok {
			if b, ok := v.(bool); ok {
				doc.MaxiTournament = &b
			}
		}
		// Team options.
		if v, ok := opts["colorPreferenceType"].(string); ok {
			doc.ColorPreferenceType = v
		}
		if v, ok := opts["primaryScore"].(string); ok {
			doc.PrimaryScore = v
		}
		// Keizer options.
		if v, ok := opts["allowRepeatPairings"]; ok {
			if b, ok := v.(bool); ok {
				doc.AllowRepeatPairings = &b
			}
		}
		if v, ok := opts["minRoundsBetweenRepeats"]; ok {
			switch n := v.(type) {
			case int:
				doc.MinRoundsBetweenRepeats = n
			case float64:
				doc.MinRoundsBetweenRepeats = int(n)
			}
		}
	}

	// Fallback: if TotalRounds was not set from options, use len(state.Rounds).
	if doc.TotalRounds == 0 && len(state.Rounds) > 0 {
		doc.TotalRounds = len(state.Rounds)
	}

	// Bridge PreAssignedByes back into Section 240 records and chesspairing
	// directives. ByePAB / ByeHalf travel via Section 240; richer types ride
	// along on a typed comment directive so a future round-trip recovers the
	// original ByeType.
	emitPreAssignedByes(doc, state, playerMap)
	emitWithdrawnDirectives(doc, state, playerMap)
	emitTeamRecords(doc, state, playerMap)

	return doc, playerMap
}

// FromTournamentStateWithError creates a TRF document and rejects team IDs
// that cannot be represented by the numeric TRF team-number field.
func FromTournamentStateWithError(state *chesspairing.TournamentState) (*Document, map[string]int, error) {
	for _, player := range state.Players {
		if player.TeamID != "" {
			if _, err := strconv.Atoi(player.TeamID); err != nil {
				return nil, nil, fmt.Errorf("trf: team ID %q is not a numeric team number", player.TeamID)
			}
		}
	}
	for _, round := range state.Rounds {
		for _, match := range round.Matches {
			for _, teamID := range []string{match.HomeID, match.AwayID} {
				if _, err := strconv.Atoi(teamID); err != nil {
					return nil, nil, fmt.Errorf("trf: team ID %q is not a numeric team number", teamID)
				}
			}
		}
	}
	doc, playerMap := FromTournamentState(state)
	return doc, playerMap, nil
}

// emitPreAssignedByes serialises state.PreAssignedByes into doc.Absences
// (for ByePAB / ByeHalf, ByeZero) and doc.ChesspairingDirectives (for the richer
// types FIDE TRF cannot express). It is the inverse of bridgePreAssignedByes.
// When state.CurrentRound is zero there is no upcoming round to anchor the
// records to, so nothing is emitted.
func emitPreAssignedByes(doc *Document, state *chesspairing.TournamentState, playerMap map[string]int) {
	if state.CurrentRound == 0 || len(state.PreAssignedByes) == 0 {
		return
	}
	round := state.CurrentRound

	// Group PAB and Half players for compact Section 240 records. The order
	// of players within a record follows the iteration order of
	// PreAssignedByes so a round-trip is stable.
	var pabPlayers, halfPlayers, zeroPlayers []int
	for _, b := range state.PreAssignedByes {
		sn, ok := playerMap[b.PlayerID]
		if !ok {
			// Player not in the document — skip rather than emit a record
			// that would fail validation on a subsequent read.
			continue
		}
		switch b.Type {
		case chesspairing.ByePAB:
			pabPlayers = append(pabPlayers, sn)
		case chesspairing.ByeHalf:
			halfPlayers = append(halfPlayers, sn)
		case chesspairing.ByeZero:
			zeroPlayers = append(zeroPlayers, sn)
		default:
			s := byeTypeToDirectiveString(b.Type)
			if s == "" {
				continue
			}
			doc.ChesspairingDirectives = append(doc.ChesspairingDirectives, Directive{
				Verb: "bye",
				Params: map[string]string{
					"round":  strconv.Itoa(round),
					"player": strconv.Itoa(sn),
					"type":   s,
				},
			})
		}
	}
	if len(pabPlayers) > 0 {
		doc.Absences = append(doc.Absences, AbsenceRecord{
			Type:    "F",
			Round:   round,
			Players: pabPlayers,
		})
	}
	if len(halfPlayers) > 0 {
		doc.Absences = append(doc.Absences, AbsenceRecord{
			Type:    "H",
			Round:   round,
			Players: halfPlayers,
		})
	}
	if len(zeroPlayers) > 0 {
		doc.Absences = append(doc.Absences, AbsenceRecord{
			Type:    "Z",
			Round:   round,
			Players: zeroPlayers,
		})
	}
}

// emitWithdrawnDirectives serialises every PlayerEntry.WithdrawnAfterRound
// into a `### chesspairing:withdrawn` directive. It is the inverse of
// bridgeWithdrawnDirectives. Players whose start number is unknown to the
// document (shouldn't happen in practice) are skipped silently.
func emitWithdrawnDirectives(doc *Document, state *chesspairing.TournamentState, playerMap map[string]int) {
	for _, p := range state.Players {
		if p.WithdrawnAfterRound == nil {
			continue
		}
		sn, ok := playerMap[p.ID]
		if !ok {
			continue
		}
		doc.ChesspairingDirectives = append(doc.ChesspairingDirectives, Directive{
			Verb: "withdrawn",
			Params: map[string]string{
				"player":      strconv.Itoa(sn),
				"after-round": strconv.Itoa(*p.WithdrawnAfterRound),
			},
		})
	}
}

// emitTeamRecords serialises team membership and matches into TRF-2026 team
// records. Board results are emitted in 801; 802 always carries the game
// point totals, including matches for which only aggregate totals are known.
func emitTeamRecords(doc *Document, state *chesspairing.TournamentState, playerMap map[string]int) {
	if len(state.Rounds) == 0 {
		return
	}
	members := make(map[string][]int)
	for _, player := range state.Players {
		if player.TeamID == "" {
			continue
		}
		if number, ok := playerMap[player.ID]; ok {
			members[player.TeamID] = append(members[player.TeamID], number)
		}
	}
	teamIDs := make([]string, 0, len(members))
	for teamID := range members {
		teamIDs = append(teamIDs, teamID)
	}
	sort.Slice(teamIDs, func(i, j int) bool {
		left, leftErr := strconv.Atoi(teamIDs[i])
		right, rightErr := strconv.Atoi(teamIDs[j])
		return leftErr == nil && (rightErr != nil || left < right)
	})
	// Record 013 has no team-number field: the team number is the record
	// order. Renumber the state's team IDs to 1..n so every emitted team
	// record (013, 801, 802) refers to the same team.
	renumber := make(map[string]int, len(teamIDs))
	for i, teamID := range teamIDs {
		renumber[teamID] = i + 1
	}
	originalPlayerTeams := make(map[string]string)
	for teamID, ms := range members {
		for _, m := range ms {
			originalPlayerTeams[strconv.Itoa(m)] = teamID
		}
	}
	for _, teamID := range teamIDs {
		doc.Teams = append(doc.Teams, TeamLine{
			TeamNumber: renumber[teamID],
			TeamName:   "",
			Members:    members[teamID],
		})
	}

	detailed := make(map[string]*DetailedTeamResult)
	simple := make(map[string]*SimpleTeamResult)
	for roundIndex, round := range state.Rounds {
		for _, match := range round.Matches {
			home, ok := renumber[match.HomeID]
			if !ok {
				continue
			}
			away, ok := renumber[match.AwayID]
			if !ok {
				continue
			}
			homeKey := strconv.Itoa(home)
			awayKey := strconv.Itoa(away)
			homeGame, awayGame := teamMatchPoints(match, scoringTeam.ParseOptions(state.ScoringConfig.Options).Options, originalPlayerTeams)
			homeSimple := teamSimpleRecord(simple, homeKey, home, len(state.Rounds))
			awaySimple := teamSimpleRecord(simple, awayKey, away, len(state.Rounds))
			homeSimple.Rounds[roundIndex] = SimpleTeamRound{Opponent: away, Color: "w", GamePoints: homeGame}
			awaySimple.Rounds[roundIndex] = SimpleTeamRound{Opponent: home, Color: "b", GamePoints: awayGame}
			homeDetailed := teamDetailedRecord(detailed, homeKey, home, len(state.Rounds))
			awayDetailed := teamDetailedRecord(detailed, awayKey, away, len(state.Rounds))
			homeDetailed.Rounds[roundIndex].Opponent = away
			homeDetailed.Rounds[roundIndex].Color = "w"
			awayDetailed.Rounds[roundIndex].Opponent = home
			awayDetailed.Rounds[roundIndex].Color = "b"
			if len(match.Boards) == 0 {
				continue
			}
			homeResults := teamBoardResults(match, originalPlayerTeams)
			homeDetailed.Rounds[roundIndex] = DetailedTeamRound{Opponent: away, Color: "w", Results: homeResults, BoardOrder: teamBoardOrder(match)}
			awayDetailed.Rounds[roundIndex] = DetailedTeamRound{Opponent: home, Color: "b", Results: invertTeamBoardResults(homeResults), BoardOrder: teamBoardOrder(match)}
		}
	}
	boardOpts := scoringTeam.ParseOptions(state.ScoringConfig.Options).WithDefaults().Options
	opts := scoringTeam.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	boardCount := scoringTeam.BoardCount(state, opts)
	for roundIndex, round := range state.Rounds {
		for _, bye := range round.TeamByes {
			number, ok := renumber[bye.PlayerID]
			if !ok {
				continue
			}
			key := strconv.Itoa(number)
			detailedRecord := teamDetailedRecord(detailed, key, number, len(state.Rounds))
			simpleRecord := teamSimpleRecord(simple, key, number, len(state.Rounds))
			detailedRecord.Rounds[roundIndex].ByeType = detailedByeMarker(bye.Type)
			simpleRecord.Rounds[roundIndex] = SimpleTeamRound{
				ByeType:    simpleByeMarker(bye.Type),
				GamePoints: teamByeGamePoints(bye.Type, boardOpts, boardCount),
			}
		}
	}
	for _, record := range detailed {
		for roundIndex := range record.Rounds {
			if record.Rounds[roundIndex].Opponent == 0 && record.Rounds[roundIndex].ByeType == "" {
				record.Rounds[roundIndex].ByeType = "ZZZZ"
			}
		}
	}
	for _, record := range simple {
		for roundIndex := range record.Rounds {
			if record.Rounds[roundIndex].Opponent == 0 && record.Rounds[roundIndex].ByeType == "" {
				record.Rounds[roundIndex].ByeType = "ZPB"
			}
		}
	}
	matchOpts := scoringTeam.ParseOptions(state.ScoringConfig.Options).WithDefaults()
	for teamID, record := range simple {
		for roundIndex, entry := range record.Rounds {
			record.GamePoints += entry.GamePoints
			if entry.Opponent == 0 {
				switch entry.ByeType {
				case "PAB", "HPB":
					record.MatchPoints += *matchOpts.PointMatchDraw
				case "FPB":
					record.MatchPoints += *matchOpts.PointMatchWin
				}
				continue
			}
			opponent := teamRoundGamePointsByID(simple, strconv.Itoa(entry.Opponent), roundIndex)
			switch {
			case entry.GamePoints > opponent:
				record.MatchPoints += *matchOpts.PointMatchWin
			case entry.GamePoints < opponent:
				record.MatchPoints += *matchOpts.PointMatchLoss
			default:
				record.MatchPoints += *matchOpts.PointMatchDraw
			}
		}
		if detailedRecord := detailed[teamID]; detailedRecord != nil {
			detailedRecord.MatchPoints = record.MatchPoints
			detailedRecord.GamePoints = record.GamePoints
		}
	}
	for _, teamID := range teamIDs {
		key := strconv.Itoa(renumber[teamID])
		if record := detailed[key]; record != nil {
			doc.DetailedTeamResults = append(doc.DetailedTeamResults, *record)
		}
		if record := simple[key]; record != nil {
			doc.SimpleTeamResults = append(doc.SimpleTeamResults, *record)
		}
	}
}

func detailedByeMarker(byeType chesspairing.ByeType) string {
	switch byeType {
	case chesspairing.ByeFullPoint:
		return "FFFF"
	case chesspairing.ByeHalf:
		return "HHHH"
	case chesspairing.ByePAB:
		return "UUUU"
	default:
		return "ZZZZ"
	}
}

func simpleByeMarker(byeType chesspairing.ByeType) string {
	switch byeType {
	case chesspairing.ByeFullPoint:
		return "FPB"
	case chesspairing.ByeHalf:
		return "HPB"
	case chesspairing.ByePAB:
		return "PAB"
	default:
		return "ZPB"
	}
}

// (largestTeamBoardCount removed, using team.BoardCount instead)

func teamByeGamePoints(byeType chesspairing.ByeType, options standard.Options, boardCount int) float64 {
	switch byeType {
	case chesspairing.ByePAB, chesspairing.ByeHalf:
		return *options.PointDraw * float64(boardCount)
	case chesspairing.ByeFullPoint:
		return *options.PointWin * float64(boardCount)
	case chesspairing.ByeAbsent:
		return *options.PointAbsent
	case chesspairing.ByeExcused:
		return *options.PointExcused
	case chesspairing.ByeClubCommitment:
		return *options.PointClubCommitment
	default:
		return *options.PointLoss
	}
}

func teamDetailedRecord(records map[string]*DetailedTeamResult, teamID string, number, rounds int) *DetailedTeamResult {
	if records[teamID] == nil {
		records[teamID] = &DetailedTeamResult{TeamNumber: number, TeamName: "", Rounds: make([]DetailedTeamRound, rounds)}
	}
	return records[teamID]
}

func teamRoundGamePointsByID(records map[string]*SimpleTeamResult, teamID string, roundIndex int) float64 {
	record := records[teamID]
	if record == nil || roundIndex >= len(record.Rounds) {
		return 0
	}
	return record.Rounds[roundIndex].GamePoints
}

func teamSimpleRecord(records map[string]*SimpleTeamResult, teamID string, number, rounds int) *SimpleTeamResult {
	if records[teamID] == nil {
		records[teamID] = &SimpleTeamResult{TeamNumber: number, TeamName: "", Rounds: make([]SimpleTeamRound, rounds)}
	}
	return records[teamID]
}

func teamMatchPoints(match chesspairing.MatchData, options standard.Options, playerTeams map[string]string) (float64, float64) {
	return scoringTeam.MatchGamePoints(match, options, playerTeams)
}

func teamBoardResults(match chesspairing.MatchData, playerTeams map[string]string) string {
	var b strings.Builder
	for _, board := range match.Boards {
		whiteHome := belongsToTeam(board.WhiteID, match.HomeID, playerTeams)
		switch board.Result {
		case chesspairing.ResultWhiteWins, chesspairing.ResultForfeitWhiteWins:
			if whiteHome {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		case chesspairing.ResultBlackWins, chesspairing.ResultForfeitBlackWins:
			if whiteHome {
				b.WriteByte('0')
			} else {
				b.WriteByte('1')
			}
		case chesspairing.ResultDraw:
			b.WriteByte('=')
		default:
			b.WriteByte('?')
		}
	}
	return b.String()
}

func belongsToTeam(playerID, teamID string, playerTeams map[string]string) bool {
	if playerTeam, ok := playerTeams[playerID]; ok {
		return playerTeam == teamID
	}
	return playerID == teamID
}

func teamBoardOrder(match chesspairing.MatchData) string {
	var b strings.Builder
	for index := range match.Boards {
		if index < 9 {
			b.WriteByte(byte('1' + index))
		}
	}
	return b.String()
}

func invertTeamBoardResults(results string) string {
	return strings.NewReplacer("1", "0", "0", "1").Replace(results)
}

// buildRoundResultForPlayer builds a single RoundResult for a player in a round.
func buildRoundResultForPlayer(playerID string, round chesspairing.RoundData, playerMap map[string]int) RoundResult {
	// Check byes first.
	for _, bye := range round.Byes {
		if bye.PlayerID == playerID {
			rc := ResultZeroBye
			switch bye.Type {
			case chesspairing.ByePAB:
				rc = ResultUnpaired
			case chesspairing.ByeFullPoint:
				rc = ResultFullBye
			case chesspairing.ByeHalf:
				rc = ResultHalfBye
			case chesspairing.ByeZero, chesspairing.ByeAbsent,
				chesspairing.ByeExcused, chesspairing.ByeClubCommitment:
				// TRF has no absence reason, so all absence types become Z.
			}
			return RoundResult{
				Opponent: 0,
				Color:    ColorNone,
				Result:   rc,
			}
		}
	}

	checkGame := func(game chesspairing.GameData) *RoundResult {
		if game.Result == chesspairing.ResultDoubleForfeit {
			if game.WhiteID == playerID {
				return &RoundResult{
					Opponent: playerMap[game.BlackID],
					Color:    ColorNone,
					Result:   ResultForfeitLoss,
				}
			}
			if game.BlackID == playerID {
				return &RoundResult{
					Opponent: playerMap[game.WhiteID],
					Color:    ColorNone,
					Result:   ResultForfeitLoss,
				}
			}
		}
		if game.WhiteID == playerID {
			oppSN := playerMap[game.BlackID]
			rc := gameResultToTRFResult(game.Result, true)
			return &RoundResult{
				Opponent: oppSN,
				Color:    ColorWhite,
				Result:   rc,
			}
		}
		if game.BlackID == playerID {
			oppSN := playerMap[game.WhiteID]
			rc := gameResultToTRFResult(game.Result, false)
			return &RoundResult{
				Opponent: oppSN,
				Color:    ColorBlack,
				Result:   rc,
			}
		}
		return nil
	}

	// Check games.
	for _, game := range round.Games {
		if res := checkGame(game); res != nil {
			return *res
		}
	}
	for _, match := range round.Matches {
		for _, game := range match.Boards {
			if res := checkGame(game); res != nil {
				return *res
			}
		}
	}

	// Player didn't participate — absent. TRF code U is the
	// pairing-allocated bye, so an absence is written as a zero-point bye.
	return RoundResult{
		Opponent: 0,
		Color:    ColorNone,
		Result:   ResultZeroBye,
	}
}

// gameResultToTRFResult converts a chesspairing.GameResult to a TRF ResultCode
// from the perspective of the player with the given color (isWhite).
func gameResultToTRFResult(gr chesspairing.GameResult, isWhite bool) ResultCode {
	switch gr {
	case chesspairing.ResultWhiteWins:
		if isWhite {
			return ResultWin
		}
		return ResultLoss
	case chesspairing.ResultBlackWins:
		if isWhite {
			return ResultLoss
		}
		return ResultWin
	case chesspairing.ResultDraw:
		return ResultDraw
	case chesspairing.ResultForfeitWhiteWins:
		if isWhite {
			return ResultForfeitWin
		}
		return ResultForfeitLoss
	case chesspairing.ResultForfeitBlackWins:
		if isWhite {
			return ResultForfeitLoss
		}
		return ResultForfeitWin
	case chesspairing.ResultDoubleForfeit:
		return ResultForfeitLoss
	case chesspairing.ResultPending:
		return ResultNotPlayed
	default:
		return ResultNotPlayed
	}
}

// pointsForTRFResult returns the standard points for a TRF result code.
func pointsForTRFResult(rc ResultCode) float64 {
	switch rc {
	case ResultWin, ResultForfeitWin, ResultWinByDefault, ResultFullBye, ResultUnpaired:
		return 1.0
	case ResultDraw, ResultDrawByDefault, ResultHalfBye:
		return 0.5
	default:
		return 0.0
	}
}

// bakuCodedType reports whether the coded tournament type of record 192 asks
// for Baku acceleration, as bbpPairings reads it.
func bakuCodedType(code string) bool {
	switch strings.TrimSpace(code) {
	case "FIDE_DUTCH_2025_BAKU", "FIDE_DUTCH_BAKU", "FIDE_BURSTEIN_BAKU":
		return true
	}
	return false
}
