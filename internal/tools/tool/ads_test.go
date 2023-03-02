// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestAzureDataStudio_Init(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Not yet implemented on Linux")
	}
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		ads := &AzureDataStudio{}
		ads.Init()

		filepath.Base(ads.exeName)
	})

}

func TestAzureDataStudio_Run(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Not yet implemented on Linux")
	}
	t.Parallel()

	ads := &AzureDataStudio{}
	ads.Init()
	ads.IsInstalled()
	_, _ = ads.Run(nil)
}

func TestAzureDataStudio_searchLocations(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Not yet implemented on Linux")
	}
	t.Parallel()

	got := (&AzureDataStudio{}).searchLocations()

	assert.GreaterOrEqual(t, len(got), 1, "expecting 1 or  search locations for Azure Data Studio on Windows, got %d", len(got))
}

func TestAzureDataStudio_installText(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Not yet implemented on Linux")
	}

	t.Parallel()

	got := (&AzureDataStudio{}).installText()

	assert.GreaterOrEqual(t, len(got), 1, "no install text provided")
}
