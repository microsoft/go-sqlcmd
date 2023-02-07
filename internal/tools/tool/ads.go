package tool

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

// Run for ADS deals with the special case of launching from WSL, which
// requires staging the files from WSL to Windows %temp%
func (t *Ads) Run(args []string) (int, error, string, string) {
	if file.Exists(filepath.Join("/", "proc", "version")) {
		version := file.GetContents(filepath.Join("/", "proc", "version"))
		reWsl, _ := regexp.Compile(".*microsoft-standard.*")

		// Are we in WSL?
		if reWsl.MatchString(version) {
			// Get the Windows %TEMP% dir as a WSL path:
			//   1. Use cmd.exe to get the Windows %TEMP%, and prefix it with /mnt  (i.e. /mnt/C:\Users\username\AppData\Local\Temp)
			//   2. Replace \\ with / (i.e. /mnt/C:/Users/username/AppData/Local/Temp)
			//   3. Finally, replace the : with '' (i.e. /mnt/C/Users/username/AppData/Local/Temp)
			//   4. Convert the path to lower (because that is what wsl needs)

			out, err := exec.Command("cmd.exe", "/c echo %temp%").Output()
			if err != nil {
				panic(err)
			}

			// Make it Linux style
			// TODO: Why on earth do I have to remove the last 3 chars!!
			linuxPath := strings.ReplaceAll(string(out)[:len(string(out))-3], "\\", "/")
			linuxPath = strings.ReplaceAll(linuxPath, "C:", "/mnt/C")
			linuxPath = strings.ToLower(linuxPath)
			fmt.Println(linuxPath)

			currentFolder := filepath.Base(folder.Getwd())

			src := args[0]
			dest := filepath.Join(linuxPath, currentFolder)

			folder.MkdirAll(dest)

			// Copy the book/testresults from Linux (WSL) to Windows %temp%
			fmt.Printf("NOTE: Staging files from WSL ('%v') to Windows ('%v').\n", src, dest)

			// /mnt/c/Windows/System32/cmd.exe
			f := file.OpenFile(filepath.Join(dest, "run.cmd"))
			file.WriteString(
				f,
				t.ExeName()+` & exit`,
			)
			file.CloseFile(f)

			cmd := exec.Command(
				"cmd.exe",
				`/c start %temp%\`+currentFolder+`\run.cmd`,
			)

			out, err = cmd.Output()

			return cmd.ProcessState.ExitCode(), err, string(out), ""

		} else {
			panic("This is not Windows, nor is it WSL, can't open ADS")
		}

	} else {
		return t.Base.Run(args)
	}
}
