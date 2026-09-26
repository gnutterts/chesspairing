package swisslib

import "testing"

// C.04.3 art. 5.2.5 says: "If the higher ranked player has an odd TPN ...,
// give them the initial-colour; otherwise, give them the opposite colour."
// A (TPN 2) and B (TPN 3) have played no games, so neither has a preference;
// with White as initial colour FIDE gives A Black.
func TestColourParity_D10_NoPreferenceUsesFixedNumberParity(t *testing.T) {
	a := &PlayerState{ID: "a", PairingNumber: 2, TPN: 1}
	b := &PlayerState{ID: "b", PairingNumber: 3, TPN: 2}
	white, black := AllocateColor(a, b, false, 1, nil, FixedNumberParity)
	if white != "b" || black != "a" {
		t.Fatalf("expected fixed-number allocation B-white, got white=%s black=%s", white, black)
	}
}
