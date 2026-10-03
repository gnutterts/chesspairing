// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package trf_test

import (
	"fmt"
	"strings"

	"github.com/gnutterts/chesspairing/trf"
)

func ExampleRead() {
	doc, err := trf.Read(strings.NewReader("012 Club championship\nXXR 5\n"))
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Name, doc.TotalRounds)
	// Output:
	// Club championship 5
}

func ExampleDocument_ToTournamentState() {
	doc := &trf.Document{Players: []trf.PlayerLine{{StartNumber: 1, Name: "Ada", Rating: 2000}}}
	state, err := doc.ToTournamentState()
	if err != nil {
		panic(err)
	}
	fmt.Println(state.Players[0].ID, state.Players[0].DisplayName)
	// Output:
	// 1 Ada
}
