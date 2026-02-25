package swisslib

import "sort"

// NumViolations is the number of optimization criteria tracked (C8, C10-C21).
const NumViolations = 13

// Violation array indices — maps to FIDE C.04.3 criteria.
const (
	IdxC8  = 0  // look-ahead (0=OK, 1=fails)
	IdxC10 = 1  // topscorer |color diff| > 2 (count)
	IdxC11 = 2  // topscorer 3+ same streak (count)
	IdxC12 = 3  // color pref not granted (count)
	IdxC13 = 4  // strong color pref not granted (count)
	IdxC14 = 5  // downfloat prev round (count)
	IdxC15 = 6  // upfloat prev round (count)
	IdxC16 = 7  // downfloat 2 rounds ago (count)
	IdxC17 = 8  // upfloat 2 rounds ago (count)
	IdxC18 = 9  // max diff downfloat prev (int, score×2)
	IdxC19 = 10 // max diff upfloat prev (int, score×2)
	IdxC20 = 11 // max diff downfloat prev-2 (int, score×2)
	IdxC21 = 12 // max diff upfloat prev-2 (int, score×2)
)

// Candidate represents a complete bracket pairing attempt to be scored.
type Candidate struct {
	Pairs          []ProposedPairing // paired players
	Floaters       []*PlayerState    // S1 players floating down
	Residuals      []*PlayerState    // unmatched S2 players for sub-bracket pairing
	DownfloaterIDs map[string]bool   // S1 player IDs (for C14-C21 float criteria)
	BracketScore   float64           // native bracket score (for C18-C21 score diff)
}

// CandidateScore holds the quality metrics for a Candidate.
// Compared lexicographically: FloaterScores first (C7), then Violations
// (C8-C21), then FloaterTPNs (C7 tiebreaker: higher-ranked floater preferred),
// then TranspositionOrder (FIDE C.04.3 B.3: prefer identity over later
// transpositions when all quality metrics are equal).
type CandidateScore struct {
	FloaterScores      []float64          // C7: scores of downfloaters, sorted descending
	FloaterTPNs        []int              // C7 tiebreaker: TPNs of floaters, sorted ascending (lower TPN preferred)
	Violations         [NumViolations]int // C8-C21: violation counts per criterion
	TranspositionOrder int                // FIDE B.3: lower = closer to identity transposition = preferred
}

// Compare returns -1 if s is better than other, +1 if worse, 0 if equal.
// Comparison order:
//  1. FloaterScores (C7): fewer floaters and lower scores are better
//  2. Violations (C8-C21): lower violation counts are better
//  3. FloaterTPNs (C7 tiebreaker): lower TPN (higher-ranked) preferred
//  4. TranspositionOrder (FIDE B.3): earlier transposition is preferred
//
// FloaterScores are compared lexicographically after sorting descending.
// Fewer floaters is always better. For equal-length lists, lower values win.
//
// Violations are compared lexicographically by index (C8 first, C21 last).
// Lower values are better at each position.
//
// FloaterTPNs break ties after violations. Lower TPN (higher-ranked player)
// is preferred. This ensures that when two transpositions produce identical
// violation scores, the one that floats the higher-ranked player is chosen,
// matching FIDE convention that quality criteria determine preference over
// transposition order.
//
// TranspositionOrder (FIDE C.04.3 B.3) is the final tiebreaker: when two
// candidates have identical quality (floater scores, violations, and floater
// TPNs), the one closer to the identity transposition is preferred.
func (s *CandidateScore) Compare(other *CandidateScore) int {
	// C7: compare floater scores lexicographically.
	minLen := len(s.FloaterScores)
	if len(other.FloaterScores) < minLen {
		minLen = len(other.FloaterScores)
	}
	for i := 0; i < minLen; i++ {
		if s.FloaterScores[i] < other.FloaterScores[i] {
			return -1
		}
		if s.FloaterScores[i] > other.FloaterScores[i] {
			return 1
		}
	}
	// More floaters = worse.
	if len(s.FloaterScores) < len(other.FloaterScores) {
		return -1
	}
	if len(s.FloaterScores) > len(other.FloaterScores) {
		return 1
	}

	// C8-C21: compare violations lexicographically.
	for i := 0; i < NumViolations; i++ {
		if s.Violations[i] < other.Violations[i] {
			return -1
		}
		if s.Violations[i] > other.Violations[i] {
			return 1
		}
	}

	// C7 tiebreaker: prefer lower floater TPN (higher-ranked player floats).
	minTPNLen := len(s.FloaterTPNs)
	if len(other.FloaterTPNs) < minTPNLen {
		minTPNLen = len(other.FloaterTPNs)
	}
	for i := 0; i < minTPNLen; i++ {
		if s.FloaterTPNs[i] < other.FloaterTPNs[i] {
			return -1
		}
		if s.FloaterTPNs[i] > other.FloaterTPNs[i] {
			return 1
		}
	}

	// FIDE B.3: prefer earlier transposition (closer to identity).
	if s.TranspositionOrder < other.TranspositionOrder {
		return -1
	}
	if s.TranspositionOrder > other.TranspositionOrder {
		return 1
	}

	return 0
}

// IsPerfect returns true if this score has no floaters and no violations.
func (s *CandidateScore) IsPerfect() bool {
	if len(s.FloaterScores) > 0 {
		return false
	}
	for i := 0; i < NumViolations; i++ {
		if s.Violations[i] != 0 {
			return false
		}
	}
	return true
}

// RelaxationOrder defines the order in which optimization criteria are
// relaxed per FIDE C.04.3: lowest-priority criterion first (C21), working
// upward to C8. Each entry is a violation index (IdxC8..IdxC21).
// When a relaxation level N is active, criteria at positions 0..N-1 in
// this slice are ignored (allowed to have any violation count).
var RelaxationOrder = [NumViolations]int{
	IdxC21, // first to relax (lowest priority)
	IdxC20,
	IdxC19,
	IdxC18,
	IdxC17,
	IdxC16,
	IdxC15,
	IdxC14,
	IdxC13,
	IdxC12,
	IdxC11,
	IdxC10,
	IdxC8, // last to relax (highest priority among optimization criteria)
}

// MeetsThreshold checks if this score's violations are acceptable at the
// given relaxation level. At level 0, all violations must be 0 (strictest).
// At level N, the first N criteria in RelaxationOrder are ignored.
// At level NumViolations, all criteria are ignored.
//
// This implements FIDE C.04.3's iterative relaxation: try strict first,
// then progressively allow violations from the lowest-priority criterion
// upward until a valid pairing is found.
func (s *CandidateScore) MeetsThreshold(relaxLevel int) bool {
	// Build the set of ignored criteria at this relaxation level.
	ignored := [NumViolations]bool{}
	for i := 0; i < relaxLevel && i < NumViolations; i++ {
		ignored[RelaxationOrder[i]] = true
	}

	// All non-ignored criteria must have 0 violations.
	for i := 0; i < NumViolations; i++ {
		if !ignored[i] && s.Violations[i] != 0 {
			return false
		}
	}
	return true
}

// OptimizationCriterion evaluates a single optimization criterion for a
// candidate. Returns the violation count (0 = no violations = perfect).
type OptimizationCriterion func(c *Candidate, ctx *CriteriaContext) int

// ScoreCandidate evaluates all optimization criteria for a candidate and
// returns its composite score. Lower scores are better.
//
// criteria is the list of OptimizationCriterion functions to evaluate
// (one per violation index). If criteria is nil, only C7 (floater scores)
// is computed.
func ScoreCandidate(cand *Candidate, ctx *CriteriaContext, criteria []OptimizationCriterion) CandidateScore {
	var score CandidateScore

	// C7: Collect floater scores, sorted descending.
	if len(cand.Floaters) > 0 {
		score.FloaterScores = make([]float64, len(cand.Floaters))
		for i, f := range cand.Floaters {
			score.FloaterScores[i] = f.Score
		}
		sort.Float64s(score.FloaterScores)
		// Reverse to descending.
		for i, j := 0, len(score.FloaterScores)-1; i < j; i, j = i+1, j-1 {
			score.FloaterScores[i], score.FloaterScores[j] = score.FloaterScores[j], score.FloaterScores[i]
		}

		// C7 tiebreaker: floater TPNs sorted ascending (lower TPN = higher-ranked = preferred).
		score.FloaterTPNs = make([]int, len(cand.Floaters))
		for i, f := range cand.Floaters {
			score.FloaterTPNs[i] = f.TPN
		}
		sort.Ints(score.FloaterTPNs)
	}

	// C8-C21: evaluate each optimization criterion.
	for i, criterion := range criteria {
		if criterion != nil && i < NumViolations {
			score.Violations[i] = criterion(cand, ctx)
		}
	}

	return score
}
