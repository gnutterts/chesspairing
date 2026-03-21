package chesspairing

// ScoringSystem identifies which scoring algorithm to use.
type ScoringSystem string

const (
	ScoringStandard ScoringSystem = "standard"
	ScoringKeizer   ScoringSystem = "keizer"
	ScoringFootball ScoringSystem = "football"
)

// IsValid returns true if the scoring system is a recognized value.
func (s ScoringSystem) IsValid() bool {
	switch s {
	case ScoringStandard, ScoringKeizer, ScoringFootball:
		return true
	}
	return false
}

// PairingSystem identifies which pairing algorithm to use.
type PairingSystem string

const (
	PairingDutch       PairingSystem = "dutch"
	PairingBurstein    PairingSystem = "burstein"
	PairingDubov       PairingSystem = "dubov"
	PairingLim         PairingSystem = "lim"
	PairingDoubleSwiss PairingSystem = "doubleswiss"
	PairingKeizer      PairingSystem = "keizer"
	PairingRoundRobin  PairingSystem = "roundrobin"
)

// IsValid returns true if the pairing system is a recognized value.
func (p PairingSystem) IsValid() bool {
	switch p {
	case PairingDutch, PairingBurstein, PairingDubov, PairingLim, PairingDoubleSwiss, PairingKeizer, PairingRoundRobin:
		return true
	}
	return false
}

// ScoringConfig holds tournament-wide scoring settings.
type ScoringConfig struct {
	System      ScoringSystem
	Tiebreakers []string
	Options     map[string]any
}

// PairingConfig holds per-period pairing settings.
type PairingConfig struct {
	System  PairingSystem
	Options map[string]any
}

// DefaultTiebreakers returns the FIDE-recommended tiebreaker order
// for the given pairing system.
func DefaultTiebreakers(system PairingSystem) []string {
	switch system {
	case PairingDutch, PairingBurstein, PairingDubov, PairingLim, PairingDoubleSwiss:
		return []string{"buchholz-cut1", "buchholz", "sonneborn-berger", "direct-encounter"}
	case PairingRoundRobin:
		return []string{"sonneborn-berger", "direct-encounter", "wins", "koya"}
	case PairingKeizer:
		return []string{"games-played", "direct-encounter", "wins"}
	default:
		return []string{"direct-encounter", "wins"}
	}
}
