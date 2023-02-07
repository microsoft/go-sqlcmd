package tool

import (
	"testing"
)

func TestAds_Init(t *testing.T) {
	ads := &Ads{}
	ads.Init()

	if ads.Name() != "ads" {
		t.Errorf("ads.Init() = %v, want %v", ads.Name(), "ads")
	}

	// check Description
	if ads.toolDescription.Name != "ads" {
		t.Errorf("ads.Description().Name = %v, want %v", ads.toolDescription.Name, "ads")
	}
	if len(ads.toolDescription.Purpose) == 0 {
		t.Errorf("ads.Description().Description is empty")
	}
	if len(ads.toolDescription.InstallText.Windows) == 0 {
		t.Errorf("ads.Description().InstallText.Windows is empty")
	}
	if len(ads.toolDescription.InstallText.Linux) == 0 {
		t.Errorf("ads.Description().InstallText.Linux is empty")
	}
	if len(ads.toolDescription.InstallText.Mac) == 0 {
		t.Errorf("ads.Description().InstallText.Mac is empty")
	}
}
