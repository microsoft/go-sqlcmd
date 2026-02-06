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

// copyPasswordToClipboard copies the password for the current context to the clipboard
// if the user is using SQL authentication. Returns true if a password was copied.
func copyPasswordToClipboard(user *sqlconfig.User, out *output.Output) bool {
	if user == nil || user.AuthenticationType != "basic" {
		return false
	}

	// Get the decrypted password from the current context
	_, _, password := config.GetCurrentContextInfo()

	if password == "" {
		return false
	}

	err := pal.CopyToClipboard(password)
	if err != nil {
		// Don't fail the command if clipboard copy fails, just warn the user
		out.Warn(localizer.Sprintf("Could not copy password to clipboard: %s", err.Error()))
		return false
	}

	out.Info(localizer.Sprintf("Password copied to clipboard - paste it when prompted"))
	return true
}
