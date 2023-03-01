// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEndpointExists(t *testing.T) {
	assert.Panics(t, func() { EndpointExists("") })

}
