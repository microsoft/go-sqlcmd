// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// searchLocations resolves the VS Code executable across the requested builds,
// trying three tiers per build:
//
//  1. PATH (the installer's optional "Add to PATH" shim, <install>\bin\code.cmd)
//  2. The Inno Setup uninstall registry key (catches custom-drive installs that
//     were never added to PATH)
//  3. The standard user and system default install directories
func (t *VSCode) searchLocations() []string {
	var locations []string
	for _, build := range t.buildsToSearch() {
		locations = append(locations, vscodeWindowsLocations(build)...)
	}
	return locations
}

func vscodeWindowsLocations(build string) []string {
	cliName, exeName, userDir, systemDir := vscodeWindowsBuildInfo(build)

	var locations []string

	// Tier 1: PATH. The shim normally lives at <install>\bin\<cli>.cmd, so the
	// launchable exe is two directories up. Some setups (enterprise images,
	// portable installs added to PATH directly) instead expose Code.exe itself,
	// in which case the install root is just one directory up. Detect the `bin`
	// shim before walking up two.
	if resolved, err := exec.LookPath(cliName); err == nil {
		parent := filepath.Dir(resolved)
		var install string
		if strings.EqualFold(filepath.Base(parent), "bin") {
			install = filepath.Dir(parent)
		} else {
			install = parent
		}
		locations = append(locations, filepath.Join(install, exeName))
	}

	// Tier 2: Inno Setup uninstall registry key.
	if install := vscodeRegistryInstallLocation(build); install != "" {
		locations = append(locations, filepath.Join(install, exeName))
	}

	// Tier 3: standard default install directories.
	locations = append(locations,
		filepath.Join(userDir, exeName),
		filepath.Join(systemDir, exeName),
	)

	return locations
}

func vscodeWindowsBuildInfo(build string) (cliName, exeName, userDir, systemDir string) {
	userProfile := os.Getenv("USERPROFILE")
	programFiles := os.Getenv("ProgramFiles")

	if build == "insiders" {
		return "code-insiders",
			"Code - Insiders.exe",
			filepath.Join(userProfile, "AppData", "Local", "Programs", "Microsoft VS Code Insiders"),
			filepath.Join(programFiles, "Microsoft VS Code Insiders")
	}
	return "code",
		"Code.exe",
		filepath.Join(userProfile, "AppData", "Local", "Programs", "Microsoft VS Code"),
		filepath.Join(programFiles, "Microsoft VS Code")
}

// vscodeRegistryInstallLocation reads InstallLocation from the Inno Setup
// uninstall keys VS Code writes on Windows. The GUIDs below are the published
// x64 product codes for the per-user and system-wide installers; arm64 and
// 32-bit installs are not probed here and fall through to PATH and the standard
// directories. Returns "" when no key is present (portable installs, missing
// builds).
func vscodeRegistryInstallLocation(build string) string {
	var guids []string
	if build == "insiders" {
		guids = []string{
			"{217B4C08-948D-4276-BFBB-BEE930AE5A2C}_is1", // user
			"{1287CAD5-7C8D-410D-88B9-0D1EE4A83FF2}_is1", // system
		}
	} else {
		guids = []string{
			"{771FD6B0-FA20-440A-A002-3B3BAC16DC50}_is1", // user
			"{EA457B21-F73E-494C-ACAB-524FDE069978}_is1", // system
		}
	}

	roots := []registry.Key{registry.CURRENT_USER, registry.LOCAL_MACHINE}
	for _, root := range roots {
		for _, guid := range guids {
			path := `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\` + guid
			k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			location, _, err := k.GetStringValue("InstallLocation")
			k.Close()
			if err == nil && location != "" {
				return location
			}
		}
	}
	return ""
}

func (t *VSCode) installText() string {
	return `Install using a package manager:

    winget install Microsoft.VisualStudioCode
    # or
    choco install vscode

Or download the latest version from:

    https://code.visualstudio.com/download

The MSSQL extension is installed on first use: when sqlcmd opens VS Code via
the vscode:// URL, VS Code prompts to install the extension if it is missing.`
}
