// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// execCommand is a package-level seam so tests can stub out exec.Command.
var execCommand = exec.Command

// vswhereFind invokes vswhere.exe to locate a Visual Studio Installer product
// (for example SSMS 21+, which registers as a VS instance). When version is
// empty it returns the latest installed instance; when set (for example "21")
// it restricts the match to that major version line. Returns "" when vswhere is
// unavailable or no instance matches.
func vswhereFind(productID, version string) string {
	pf86 := os.Getenv("ProgramFiles(x86)")
	if pf86 == "" {
		return ""
	}

	vswhere := filepath.Join(pf86, "Microsoft Visual Studio", "Installer", "vswhere.exe")
	if _, err := os.Stat(vswhere); err != nil {
		return ""
	}

	args := []string{
		"-products", productID,
		"-property", "installationPath",
		"-format", "value",
		"-nologo",
		"-utf8",
	}
	if version == "" {
		args = append(args, "-latest")
	} else if major, err := strconv.Atoi(version); err == nil {
		// vswhere range syntax: "[21.0,22.0)" matches the 21.x line.
		args = append(args, "-version", fmt.Sprintf("[%d.0,%d.0)", major, major+1))
	}

	out, err := execCommand(vswhere, args...).Output()
	if err != nil {
		return ""
	}

	// vswhere may list multiple matching instances (one path per line); take
	// the first non-empty one.
	for _, line := range strings.Split(string(out), "\n") {
		if p := strings.TrimSpace(line); p != "" {
			return p
		}
	}
	return ""
}
