// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEndpointExists(t *testing.T) {
	defer func() { assert.NotNil(t, recover(), "The code did not panic as expected") }()
	EndpointExists("")
}
