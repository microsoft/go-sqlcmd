// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package open

import (
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
)

// copyPasswordToClipboard returns true when the password was copied.
func copyPasswordToClipboard(user *sqlconfig.User, out *output.Output) bool {
	if out == nil || user == nil || user.AuthenticationType != "basic" || user.BasicAuth == nil {
		return false
	}

	_, _, password := config.GetCurrentContextInfo()
	if password == "" {
		return false
	}

	if err := pal.CopyToClipboard(password); err != nil {
		out.Warn(localizer.Sprintf("Could not copy password to clipboard: %s", err.Error()))
		return false
	}

	out.Info(localizer.Sprintf("Password copied to clipboard - paste it when prompted, then clear your clipboard"))
	return true
}
