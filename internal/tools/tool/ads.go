package tool

import (
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"os"
	"path/filepath"
)

type Ads struct {
	Base
}

func (t *Ads) Init() {
	t.Base.SetName("ads")
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")

	// Search in this order
	//   User Insiders Install
	//   System Insiders Install
	//   User non-Insiders install
	//   System non-Insiders install
	searchLocations := []string{
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Azure Data Studio - Insiders\\azuredatastudio-insiders.exe"),
		filepath.Join(programFiles, "Azure Data Studio - Insiders\\azuredatastudio-insiders.exe"),
		filepath.Join(userProfile, "AppData\\Local\\Programs\\Azure Data Studio\\azuredatastudio.exe"),
		filepath.Join(programFiles, "Azure Data Studio\\azuredatastudio.exe"),
	}

	t.Base.SetExeName("azuredatastudio-insiders")
	for _, location := range searchLocations {
		if file.Exists(location) {
			t.Base.SetExeName(location)
			break
		}
	}

	t.Base.SetToolYaml(ToolDescription{
		t.Name(),
		"Azure Data Studio provides a User Interface for working with SQL Server, Azure SQL Database, and Azure SQL Data Warehouse.",
		InstallText{
			Windows: `Download the latest 'User Installer' .msi from:

    https://go.microsoft.com/fwlink/?linkid=2150927

More information can be found here:

    https://docs.microsoft.com/en-us/sql/azure-data-studio/download-azure-data-studio#get-azure-data-studio-for-windows`,
			Linux: `Follow the instructions here:

   https://docs.microsoft.com/en-us/sql/azure-data-studio/download-azure-data-studio?#get-azure-data-studio-for-linux`,
			Mac: `Download the latest .zip from:

    https://go.microsoft.com/fwlink/?linkid=2151311

More information can be found here:

    https://docs.microsoft.com/en-us/sql/azure-data-studio/download-azure-data-studio?#get-azure-data-studio-for-macos`,
		}})
}

func (t *Ads) Run(args []string) (int, error, string, string) {
	return t.Base.Run(args)
}
