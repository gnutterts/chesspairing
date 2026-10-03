// Copyright 2026 Gert Nutterts
// SPDX-License-Identifier: Apache-2.0

package keizer

import "fmt"

func ExampleESGOptions() {
	options := ESGOptions()
	fmt.Println(*options.ValueNumberBase, *options.ByeFixedValue)
	// Output:
	// 60 40
}
