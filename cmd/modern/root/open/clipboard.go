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

// copyPasswordToClipboard copies the SQL password to the system clipboard.
// The password remains on the clipboard until the user or another application
// clears it; callers rely on the user heeding the "clear your clipboard" message.
func copyPasswordToClipboard(user *sqlconfig.User, out *output.Output) bool {
	if user == nil || user.AuthenticationType != "basic" {
		return false
	}

	_, _, password := config.GetCurrentContextInfo()

	if password == "" {
		return false
	}

	err := pal.CopyToClipboard(password)
	if err != nil {
		out.Warn(localizer.Sprintf("Could not copy password to clipboard: %s", err.Error()))
		return false
	}

	out.Info(localizer.Sprintf("Password copied to clipboard - paste it when prompted, then clear your clipboard"))
	return true
}
