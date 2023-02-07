package tool

import (
	"os"
	"testing"
)

func TestName(t *testing.T) {
	base := Base{name: "test_tool"}
	if base.Name() != "test_tool" {
		t.Errorf("expected 'test_tool', but got %v", base.Name())
	}
}

func TestSetName(t *testing.T) {
	base := Base{}
	base.SetName("test_tool")
	if base.Name() != "test_tool" {
		t.Errorf("expected 'test_tool', but got %v", base.Name())
	}
}

func TestExeName(t *testing.T) {
	base := Base{exeName: "test_exe"}
	if base.ExeName() != "test_exe" {
		t.Errorf("expected 'test_exe', but got %v", base.ExeName())
	}
}

func TestSetExeName(t *testing.T) {
	base := Base{}
	base.SetExeName("test_exe")
	if base.ExeName() != "test_exe" {
		t.Errorf("expected 'test_exe', but got %v", base.ExeName())
	}
}

func TestWhere(t *testing.T) {
	base := Base{exeFullPath: "test/path"}
	if base.Where() != "test/path" {
		t.Errorf("expected 'test/path', but got %v", base.Where())
	}
}

func TestIsInstalled(t *testing.T) {
	// Test when exeName is not set
	base := Base{}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected a panic, but got nil")
		}
	}()
	base.IsInstalled()

	// Test when exeName is set, but LookPath returns an error
	base = Base{exeName: "test_exe"}
	base.isInstalledCalled = true
	base.lookPathError = os.ErrNotExist
	if base.IsInstalled() {
		t.Errorf("expected false, but got true")
	}

	// Test when exeName is set, and LookPath returns no error
	base = Base{exeName: "test_exe"}
	base.isInstalledCalled = true
	base.lookPathError = nil
	if !base.IsInstalled() {
		t.Errorf("expected true, but got false")
	}
}

func TestHowToInstall(t *testing.T) {
	t.Run("windows", func(t *testing.T) {
		b := &Base{
			name: "test",
			toolDescription: Description{
				Purpose: "test purpose",
				InstallText: InstallText{
					Windows: "windows install",
					Mac:     "mac install",
					Linux:   "linux install",
				},
			},
		}
		b.HowToInstall()
	})
}
