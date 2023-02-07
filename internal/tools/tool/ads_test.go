package tool

import (
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

func TestAds_Init(t *testing.T) {
	ads := &Ads{}
	ads.Init()

	if ads.Name() != "ads" {
		t.Errorf("ads.Init() = %v, want %v", ads.Name(), "ads")
	}

	// check ToolDescription
	if ads.toolYaml.Name != "ads" {
		t.Errorf("ads.ToolDescription().Name = %v, want %v", ads.toolYaml.Name, "ads")
	}
	if len(ads.toolYaml.Purpose) == 0 {
		t.Errorf("ads.ToolDescription().Description is empty")
	}
	if len(ads.toolYaml.InstallText.Windows) == 0 {
		t.Errorf("ads.ToolDescription().InstallText.Windows is empty")
	}
	if len(ads.toolYaml.InstallText.Linux) == 0 {
		t.Errorf("ads.ToolDescription().InstallText.Linux is empty")
	}
	if len(ads.toolYaml.InstallText.Mac) == 0 {
		t.Errorf("ads.ToolDescription().InstallText.Mac is empty")
	}
}

type MockFile struct {
	mock.Mock
}

func (m *MockFile) Exists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

func (m *MockFile) GetContents(path string) string {
	args := m.Called(path)
	return args.String(0)
}

func (m *MockFile) OpenFile(path string) *os.File {
	args := m.Called(path)
	return args.Get(0).(*os.File)
}

func (m *MockFile) WriteString(file *os.File, contents string) {
	m.Called(file, contents)
}

func (m *MockFile) CloseFile(file *os.File) {
	m.Called(file)
}

type MockFolder struct {
	mock.Mock
}

func (m *MockFolder) Getwd() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFolder) MkdirAll(path string) {
	m.Called(path)
}
