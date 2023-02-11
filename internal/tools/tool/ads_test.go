// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAds_Init(t *testing.T) {
	ads := &AzureDataStudio{}
	ads.Init()
	assert.Equal(t, ads.Name(), "ads", "ads.Init() = %v, want %v", ads.Name(), "ads")
	assert.Equal(t, ads.description.Name, "ads", "ads.Description().Name = %v, want %v", ads.description.Name, "ads")
	assert.NotEqual(t, len(ads.description.Purpose), 0, "ads.Description().Description is empty")
	assert.NotEqual(t, len(ads.description.InstallText.Windows), 0, "ads.Description().InstallText.Windows is empty")
	assert.NotEqual(t, len(ads.description.InstallText.Linux), 0, "ads.Description().InstallText.Linux is empty")
	assert.NotEqual(t, len(ads.description.InstallText.Mac), 0, "ads.Description().InstallText.Mac is empty")
}
