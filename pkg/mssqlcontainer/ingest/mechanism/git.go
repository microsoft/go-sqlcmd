package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"os"

	git "gopkg.in/src-d/go-git.v4"
)

type git2 struct {
}

func (m *git2) Initialize(controller *container.Controller) {
}

func (m *git2) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m *git2) Name() string {
	return "git"
}

func (m *git2) FileTypes() []string {
	return []string{"git"}
}

func (m *git2) BringOnline(databaseName string, _ string, query func(string), options BringOnlineOptions) {
	if options.Filename == "" {
		panic("Filename is required for restore")
	}
	if databaseName == "" {
		panic("databaseName is required for restore")
	}

	url := "https://github.com/stuartpa/DabBlazorSamplePages.git"
	dir := "."

	// If there are any files in the current directory then error
	// as we don't want to overwrite any files
	entries, err := os.ReadDir(dir)
	checkErr(err)

	if len(entries) > 0 {
		fmt.Println("Current directory is not empty, cannot clone .git repo. Run sqlcmd again from an empty directory, or remove the --use switch.")
		os.Exit(1)
	}

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		fmt.Println("Error while cloning repository:", err)
		os.Exit(1)
	}

	fmt.Println("Repository cloned successfully")
}
