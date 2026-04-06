package trf

import (
	"sort"
	"strconv"

	"github.com/gnutterts/chesspairing"
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
			ID:          strconv.Itoa(pl.StartNumber),
			DisplayName: pl.Name,
			Rating:      pl.Rating,
			Active:      true,
			Federation:  pl.Federation,
			FideID:      pl.FideID,
			Title:       pl.Title,
			Sex:         pl.Sex,
			BirthDate:   pl.BirthDate,
		}
	}

	// Determine number of rounds from player data.
	maxRounds := 0
	for _, pl := range doc.Players {
		if len(pl.Rounds) > maxRounds {
			maxRounds = len(pl.Rounds)
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
					bt = chesspairing.ByePAB
				case ResultHalfBye:
					bt = chesspairing.ByeHalf
				case ResultZeroBye:
					bt = chesspairing.ByeZero
				case ResultUnpaired:
					bt = chesspairing.ByeAbsent
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

	state.CurrentRound = maxRounds

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
		state.PairingConfig.Options["topSeedColor"] = ic
	}
	if len(doc.ForbiddenPairs) > 0 {
		pairs := make([][2]int, len(doc.ForbiddenPairs))
		for i, fp := range doc.ForbiddenPairs {
			pairs[i] = [2]int{fp.Player1, fp.Player2}
		}
		state.PairingConfig.Options["forbiddenPairs"] = pairs
	}
	// TRF-2026 forbidden pair records (260) — convert to the same format.
	// Each 260 record lists mutually forbidden players; generate all pairs.
	if len(doc.ForbiddenPairs26) > 0 && len(doc.ForbiddenPairs) == 0 {
		var pairs [][2]int
		for _, fp := range doc.ForbiddenPairs26 {
			for i := 0; i < len(fp.Players); i++ {
				for j := i + 1; j < len(fp.Players); j++ {
					pairs = append(pairs, [2]int{fp.Players[i], fp.Players[j]})
				}
			}
		}
		if len(pairs) > 0 {
			state.PairingConfig.Options["forbiddenPairs"] = pairs
		}
	}
	// Acceleration from XXS lines.
	if len(doc.Acceleration) > 0 {
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

	// Scoring config: defaults.
	state.ScoringConfig = chesspairing.ScoringConfig{
		System:      chesspairing.ScoringStandard,
		Tiebreakers: chesspairing.DefaultTiebreakers(state.PairingConfig.System),
	}

	return state, nil
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
// Players are assigned start numbers sorted by rating descending (ties broken
// by display name ascending, then by ID ascending for determinism).
// Returns the Document and a mapping from player ID to assigned start number.
func FromTournamentState(state *chesspairing.TournamentState) (*Document, map[string]int) {
	doc := &Document{}

	// Sort players by rating desc, name asc, ID asc.
	sorted := make([]chesspairing.PlayerEntry, len(state.Players))
	copy(sorted, state.Players)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Rating != sorted[j].Rating {
			return sorted[i].Rating > sorted[j].Rating
		}
		if sorted[i].DisplayName != sorted[j].DisplayName {
			return sorted[i].DisplayName < sorted[j].DisplayName
		}
		return sorted[i].ID < sorted[j].ID
	})

	// Assign start numbers and build lookup.
	playerMap := make(map[string]int, len(sorted))
	for i, p := range sorted {
		playerMap[p.ID] = i + 1
	}

	// Build player lines.
	doc.Players = make([]PlayerLine, len(sorted))
	for i, p := range sorted {
		sn := i + 1
		pl := PlayerLine{
			StartNumber: sn,
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
			doc.InitialColor = v
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

	return doc, playerMap
}

// buildRoundResultForPlayer builds a single RoundResult for a player in a round.
func buildRoundResultForPlayer(playerID string, round chesspairing.RoundData, playerMap map[string]int) RoundResult {
	// Check byes first.
	for _, bye := range round.Byes {
		if bye.PlayerID == playerID {
			rc := ResultFullBye
			switch bye.Type {
			case chesspairing.ByeHalf:
				rc = ResultHalfBye
			case chesspairing.ByeZero:
				rc = ResultZeroBye
			case chesspairing.ByeAbsent, chesspairing.ByeExcused, chesspairing.ByeClubCommitment:
				// TRF has no concept of excused/club commitment —
				// all absence types map to unpaired ("U").
				rc = ResultUnpaired
			}
			return RoundResult{
				Opponent: 0,
				Color:    ColorNone,
				Result:   rc,
			}
		}
	}

	// Check games.
	for _, game := range round.Games {
		if game.WhiteID == playerID {
			oppSN := playerMap[game.BlackID]
			rc := gameResultToTRFResult(game.Result, true)
			return RoundResult{
				Opponent: oppSN,
				Color:    ColorWhite,
				Result:   rc,
			}
		}
		if game.BlackID == playerID {
			oppSN := playerMap[game.WhiteID]
			rc := gameResultToTRFResult(game.Result, false)
			return RoundResult{
				Opponent: oppSN,
				Color:    ColorBlack,
				Result:   rc,
			}
		}
	}

	// Player didn't participate — absent.
	return RoundResult{
		Opponent: 0,
		Color:    ColorNone,
		Result:   ResultUnpaired,
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
	case ResultWin, ResultForfeitWin, ResultWinByDefault, ResultFullBye:
		return 1.0
	case ResultDraw, ResultDrawByDefault, ResultHalfBye:
		return 0.5
	default:
		return 0.0
	}
}
