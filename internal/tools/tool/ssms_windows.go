// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

const ssmsPrefix = "Microsoft SQL Server Management Studio "

// SSMS 21+ moved the exe under Release\; older versions use Common7\IDE directly.
var ssmsExeSubPaths = []string{
	`Release\Common7\IDE\Ssms.exe`,
	`Common7\IDE\Ssms.exe`,
}

func (t *SSMS) searchLocations() []string {
	programFiles := os.Getenv("ProgramFiles")
	programFilesX86 := os.Getenv("ProgramFiles(x86)")

	roots := []string{programFiles, programFilesX86}
	dirs := discoverSSMSDirs(roots)

	var paths []string
	for _, dir := range dirs {
		for _, sub := range ssmsExeSubPaths {
			paths = append(paths, filepath.Join(dir, sub))
		}
	}
	return paths
}

// discoverSSMSDirs globs for SSMS install directories under the given root
// folders and returns them sorted by version number descending (newest first).
func discoverSSMSDirs(roots []string) []string {
	versionRe := regexp.MustCompile(`(\d+)$`)
	type entry struct {
		path    string
		version int
	}

	var entries []entry
	for _, root := range roots {
		if root == "" {
			continue
		}
		matches, err := filepath.Glob(filepath.Join(root, ssmsPrefix+"*"))
		if err != nil {
			continue
		}
		for _, m := range matches {
			base := filepath.Base(m)
			if sub := versionRe.FindString(base); sub != "" {
				if v, err := strconv.Atoi(sub); err == nil {
					entries = append(entries, entry{m, v})
				}
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].version > entries[j].version
	})

	dirs := make([]string, len(entries))
	for i, e := range entries {
		dirs[i] = e.path
	}
	return dirs
}

func (t *SSMS) installText() string {
	return `Install using a package manager:

    winget install Microsoft.SQLServerManagementStudio
    # or
    choco install sql-server-management-studio

Or download the latest version from:

    https://aka.ms/ssmsfullsetup

Note: SSMS is only available on Windows.`
}
