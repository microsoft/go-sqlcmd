package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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
		panic("Filename is required for git")
	}
	/*
		if databaseName == "" {
			panic("databaseName is required for git")
		}
	*/

	url := options.Filename
	dir := "."

	// If there are any files in the current directory then error
	// as we don't want to overwrite any files
	entries, err := os.ReadDir(dir)
	checkErr(err)

	alreadyCloned := false

	if len(entries) > 0 {
		if _, err := os.Stat(".git"); err == nil {
			var fs billy.Filesystem
			fs = osfs.New(".git")

			st := filesystem.NewStorage(fs, nil)
			c, err := st.Config()
			if err != nil {
				panic(err)
			}

			for _, remote := range c.Remotes {
				for _, u := range remote.URLs {
					if url == u {
						alreadyCloned = true
					}
				}
			}
		}

		if !alreadyCloned {
			fmt.Println("Current directory is not empty, cannot clone .git repo. Run sqlcmd again from an empty directory, or remove the --use switch.")

			os.Exit(1)
		}
	}

	if !alreadyCloned {
		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			URL: url,
		})

		if err != nil {
			fmt.Println("Error while cloning repository:", err)
			os.Exit(1)
		}

		fmt.Println("Repository cloned successfully")
	} else {
		fmt.Println("Repository already cloned, continuing...")
	}
}
