// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"os"
	"strings"
)

// Variables provides set and get of sqlcmd scripting variables
type Variables map[string]string

var variables Variables

// Built-in scripting variables
const (
	SQLCMDDBNAME            = "SQLCMDDBNAME"
	SQLCMDINI               = "SQLCMDINI"
	SQLCMDPACKETSIZE        = "SQLCMDPACKETSIZE"
	SQLCMDPASSWORD          = "SQLCMDPASSWORD"
	SQLCMDSERVER            = "SQLCMDSERVER"
	SQLCMDUSER              = "SQLCMDUSER"
	SQLCMDWORKSTATION       = "SQLCMDWORKSTATION"
	SQLCMDLOGINTIMEOUT      = "SQLCMDLOGINTIMEOUT"
	SQLCMDSTATTIMEOUT       = "SQLCMDSTATTIMEOUT"
	SQLCMDHEADERS           = "SQLCMDHEADERS"
	SQLCMDCOLSEP            = "SQLCMDCOLSEP"
	SQLCMDCOLWIDTH          = "SQLCMDCOLWIDTH"
	SQLCMDERRORLEVEL        = "SQLCMDERRORLEVEL"
	SQLCMDMAXVARTYPEWIDTH   = "SQLCMDMAXVARTYPEWIDTH"
	SQLCMDMAXFIXEDTYPEWIDTH = "SQLCMDMAXFIXEDTYPEWIDTH"
	SQLCMDEDITOR            = "SQLCMDEDITOR"
	SQLCMDUSEAAD            = "SQLCMDUSEAAD"
)

// Variables that can only be set at startup
var readOnlyVariables = []string{
	SQLCMDDBNAME,
	SQLCMDINI,
	SQLCMDPACKETSIZE,
	SQLCMDPASSWORD,
	SQLCMDSERVER,
	SQLCMDUSER,
	SQLCMDWORKSTATION,
}

func (v Variables) checkReadOnly(key string) error {
	currentValue, hasValue := v[key]
	if hasValue {
		for _, variable := range readOnlyVariables {
			if variable == key && currentValue != "" {
				return ReadOnlyVariable(key)
			}
		}
	}
	return nil
}

// Set sets or adds the value in the map.
func (v Variables) Set(name, value string) {
	key := strings.ToUpper(name)
	v[key] = value
}

// Unset removes the value from the map
func (v Variables) Unset(name string) {
	key := strings.ToUpper(name)
	delete(v, key)
}

// All returns a copy of the current variables
func (v Variables) All() map[string]string {
	return map[string]string(v)
}

// SQLCmdUser returns the SQLCMDUSER variable value
func (v Variables) SQLCmdUser() string {
	return v[SQLCMDUSER]
}

// SQLCmdServer returns the server connection parameters derived from the SQLCMDSERVER variable value
func (v Variables) SQLCmdServer() (serverName string, instance string, port uint64, err error) {
	serverName = v[SQLCMDSERVER]
	return SplitServer(serverName)
}

// SQLCmdDatabase returns the SQLCMDDBNAME variable value
func (v Variables) SQLCmdDatabase() string {
	return v[SQLCMDDBNAME]
}

// UseAad returns whether the SQLCMDUSEAAD variable value is set to "true"
func (v Variables) UseAad() bool {
	return strings.EqualFold(v[SQLCMDUSEAAD], "true")
}

// Password returns the password used for connections as specified by SQLCMDPASSWORD variable
func (v Variables) Password() string {
	return v[SQLCMDPASSWORD]
}

// ColumnSeparator is the value of SQLCMDCOLSEP variable. It can have 0 or 1 characters
func (v Variables) ColumnSeparator() string {
	sep := v[SQLCMDCOLSEP]
	if len(sep) > 1 {
		return sep[:1]
	}
	return sep
}

// MaxFixedColumnWidth is the value of SQLCMDMAXFIXEDTYPEWIDTH variable.
// When non-zero, it limits the width of columns for types CHAR, NCHAR, NVARCHAR, VARCHAR, VARBINARY, VARIANT
func (v Variables) MaxFixedColumnWidth() int64 {
	w := v[SQLCMDMAXFIXEDTYPEWIDTH]
	return mustValue(w)
}

// MaxVarColumnWidth is the value of SQLCMDMAXVARTYPEWIDTH variable.
// When non-zero, it limits the width of columns for (max) versions of CHAR, NCHAR, VARBINARY.
// It also limits the width of xml, UDT, text, ntext, and image
func (v Variables) MaxVarColumnWidth() int64 {
	w := v[SQLCMDMAXVARTYPEWIDTH]
	return mustValue(w)
}

// ScreenWidth is the value of SQLCMDCOLWIDTH variable.
// It tells the formatter how many characters wide to limit all screen output.
func (v Variables) ScreenWidth() int64 {
	w := v[SQLCMDCOLWIDTH]
	return mustValue(w)
}

// RowsBetweenHeaders is the value of SQLCMDHEADERS variable.
// When MaxVarColumnWidth() is 0, it returns -1
func (v Variables) RowsBetweenHeaders() int64 {
	if v.MaxVarColumnWidth() == 0 {
		return -1
	}
	h := mustValue(v[SQLCMDHEADERS])
	return h
}

func mustValue(val string) int64 {
	var n int64
	_, err := fmt.Sscanf(val, "%d", &n)
	if err == nil {
		return n
	}
	panic(err)
}

// InitializeVariables initializes variables with default values.
// When fromEnvironment is true, then loads from the runtime environment
func InitializeVariables(fromEnvironment bool) *Variables {
	variables = Variables{
		SQLCMDCOLSEP:            " ",
		SQLCMDCOLWIDTH:          "0",
		SQLCMDDBNAME:            "",
		SQLCMDEDITOR:            "edit.com",
		SQLCMDERRORLEVEL:        "0",
		SQLCMDHEADERS:           "0",
		SQLCMDINI:               "",
		SQLCMDLOGINTIMEOUT:      "30",
		SQLCMDMAXFIXEDTYPEWIDTH: "0",
		SQLCMDMAXVARTYPEWIDTH:   "256",
		SQLCMDPACKETSIZE:        "4096",
		SQLCMDSERVER:            "",
		SQLCMDSTATTIMEOUT:       "0",
		SQLCMDUSER:              "",
		SQLCMDPASSWORD:          "",
		SQLCMDUSEAAD:            "",
	}
	hostname, _ := os.Hostname()
	variables.Set(SQLCMDWORKSTATION, hostname)

	if fromEnvironment {
		for v := range variables.All() {
			envVar, ok := os.LookupEnv(v)
			if ok {
				variables.Set(v, envVar)
			}
		}
	}
	return &variables
}

// Setvar implements the :Setvar command
// TODO: Add validation functions for the variables.
func Setvar(name, value string) error {
	err := ValidIdentifier(name)
	if err == nil {
		err = variables.checkReadOnly(name)
	}
	if err != nil {
		return err
	}
	variables.Set(name, value)
	return nil
}

// ValidIdentifier determines if a given string can be used as a variable name
func ValidIdentifier(name string) error {
	if strings.HasPrefix(name, "$(") || strings.ContainsAny(name, "'\"\t\n\r ") {
		return InvalidCommandError(":setvar", 0)
	}
	return nil
}
