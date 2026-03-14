package trf

import "fmt"

// ValidationProfile determines which checks to apply.
type ValidationProfile int

const (
	// ValidateGeneral checks structural integrity only.
	ValidateGeneral ValidationProfile = iota
	// ValidatePairingEngine checks fields needed by pairing programs (XXR, XXC, 092).
	ValidatePairingEngine
	// ValidateFIDE checks all fields required for FIDE rating submission.
	ValidateFIDE
)

// Severity indicates whether an issue is blocking or advisory.
type Severity int

const (
	// SeverityError indicates a field that must be present for the profile.
	SeverityError Severity = iota
	// SeverityWarning indicates a field that should be present but is not blocking.
	SeverityWarning
)

// String returns "error" or "warning".
func (s Severity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warning"
}

// ValidationIssue describes a single problem found during validation.
type ValidationIssue struct {
	Field    string   // e.g. "012", "XXR", "player.3.fideID"
	Severity Severity // SeverityError or SeverityWarning
	Message  string   // Human-readable description
}

// Validate checks the Document for completeness according to the given profile.
// Each profile is a superset of the previous one:
//   - ValidateGeneral: structural integrity (players exist, unique start numbers)
//   - ValidatePairingEngine: adds XXR, XXC, 092
//   - ValidateFIDE: adds tournament name, dates, time control, player data quality
func (doc *Document) Validate(profile ValidationProfile) []ValidationIssue {
	var issues []ValidationIssue

	// --- General checks (all profiles) ---

	if len(doc.Players) == 0 {
		issues = append(issues, ValidationIssue{
			Field:    "001",
			Severity: SeverityError,
			Message:  "no player lines",
		})
	}

	// Check unique start numbers.
	seen := make(map[int]bool, len(doc.Players))
	for _, p := range doc.Players {
		if seen[p.StartNumber] {
			issues = append(issues, ValidationIssue{
				Field:    "001.startNumber",
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate start number %d", p.StartNumber),
			})
		}
		seen[p.StartNumber] = true
	}

	// NumPlayers mismatch.
	if doc.NumPlayers > 0 && doc.NumPlayers != len(doc.Players) {
		issues = append(issues, ValidationIssue{
			Field:    "062",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("NumPlayers=%d but %d player lines", doc.NumPlayers, len(doc.Players)),
		})
	}

	if profile < ValidatePairingEngine {
		return issues
	}

	// --- Pairing engine checks ---

	if doc.TotalRounds == 0 {
		issues = append(issues, ValidationIssue{
			Field:    "XXR",
			Severity: SeverityError,
			Message:  "total rounds not set",
		})
	}

	if doc.InitialColor == "" {
		issues = append(issues, ValidationIssue{
			Field:    "XXC",
			Severity: SeverityError,
			Message:  "initial color not set",
		})
	}

	if doc.TournamentType == "" {
		issues = append(issues, ValidationIssue{
			Field:    "092",
			Severity: SeverityError,
			Message:  "tournament type not set",
		})
	}

	if profile < ValidateFIDE {
		return issues
	}

	// --- FIDE rating submission checks ---

	if doc.Name == "" {
		issues = append(issues, ValidationIssue{
			Field:    "012",
			Severity: SeverityError,
			Message:  "tournament name not set",
		})
	}

	if doc.StartDate == "" {
		issues = append(issues, ValidationIssue{
			Field:    "042",
			Severity: SeverityError,
			Message:  "start date not set",
		})
	}

	if doc.EndDate == "" {
		issues = append(issues, ValidationIssue{
			Field:    "052",
			Severity: SeverityError,
			Message:  "end date not set",
		})
	}

	if doc.TimeControl == "" {
		issues = append(issues, ValidationIssue{
			Field:    "122",
			Severity: SeverityError,
			Message:  "time control not set",
		})
	}

	// Per-player warnings.
	for _, p := range doc.Players {
		if p.Name == "" {
			issues = append(issues, ValidationIssue{
				Field:    fmt.Sprintf("player.%d.name", p.StartNumber),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("player %d has no name", p.StartNumber),
			})
		}
		if p.Federation == "" {
			issues = append(issues, ValidationIssue{
				Field:    fmt.Sprintf("player.%d.federation", p.StartNumber),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("player %d has no federation", p.StartNumber),
			})
		}
		// Rated player without FIDE ID — likely not submittable.
		if p.Rating > 0 && p.FideID == "" {
			issues = append(issues, ValidationIssue{
				Field:    fmt.Sprintf("player.%d.fideID", p.StartNumber),
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("player %d is rated (%d) but has no FIDE ID", p.StartNumber, p.Rating),
			})
		}
	}

	return issues
}
