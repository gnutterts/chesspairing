// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import "testing"

func TestESGOptions(t *testing.T) {
	o := ESGOptions()
	checks := []struct {
		name string
		ok   bool
	}{
		{"InitialOrder=rating-entry", o.InitialOrder != nil && *o.InitialOrder == "rating-entry"},
		{"ByePolicy=lowest-without-bye", o.ByePolicy != nil && *o.ByePolicy == byePolicyLowestWithoutBye},
		{"PeriodLength=6", o.PeriodLength != nil && *o.PeriodLength == 6},
		{"NoRepeatWithinPeriod=true", o.NoRepeatWithinPeriod != nil && *o.NoRepeatWithinPeriod},
		{"MinRoundsBetweenRepeats=2", o.MinRoundsBetweenRepeats != nil && *o.MinRoundsBetweenRepeats == 2},
		{"ForfeitCountsAsMet=false", o.ForfeitCountsAsMet != nil && !*o.ForfeitCountsAsMet},
		{"ScoringOptions=scoring/keizer.ESGOptions", o.ScoringOptions != nil &&
			o.ScoringOptions.Method != nil && *o.ScoringOptions.Method == "frozen" &&
			o.ScoringOptions.ForfeitCountsAsMet != nil && !*o.ScoringOptions.ForfeitCountsAsMet},
	}
	for _, check := range checks {
		if !check.ok {
			t.Errorf("ESGOptions %s not satisfied", check.name)
		}
	}
}
