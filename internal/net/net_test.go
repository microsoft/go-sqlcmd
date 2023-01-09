// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package net

import (
	"testing"
)

// TestIsLocalPortAvailable verified the function for both available and unavailable
// code (this function expects a local SQL Server instance listening on port 1433
func TestIsLocalPortAvailable(t *testing.T) {
	var testedPortAvailable bool
	var testedNotPortAvailable bool

	for i := 1432; i <= 1434; i++ {
		isPortAvailable := IsLocalPortAvailable(i)
		if isPortAvailable {
			testedPortAvailable = true
			t.Logf("Port %d is available", i)
		} else {
			testedNotPortAvailable = true
			t.Logf("Port %d is not available", i)
		}
		if testedPortAvailable && testedNotPortAvailable {
			return
		}
	}

	t.Log("Didn't find both an available port and unavailable port")
	t.Fail()

}
