// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package factory_test

import (
	"fmt"

	"github.com/gnutterts/chesspairing/factory"
)

func ExampleNewPairer() {
	pairer, err := factory.NewPairer("dutch", nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%T\n", pairer)
	// Output:
	// *dutch.Pairer
}
