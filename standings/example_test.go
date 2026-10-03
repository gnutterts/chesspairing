// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package standings_test

import (
	"context"
	"fmt"

	"github.com/gnutterts/chesspairing"
	"github.com/gnutterts/chesspairing/scoring/standard"
	"github.com/gnutterts/chesspairing/standings"
)

func ExampleBuild() {
	state := &chesspairing.TournamentState{Players: []chesspairing.PlayerEntry{{ID: "1", DisplayName: "Ada"}}}
	table, err := standings.Build(context.Background(), state, standard.New(standard.Options{}), nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(table[0].DisplayName, table[0].Rank)
	// Output:
	// Ada 1
}
