// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package variables

import (
	"fmt"
	"os"
	"strings"

	"github.com/microsoft/go-sqlcmd/sqlcmderrors"
	"github.com/microsoft/go-sqlcmd/util"
)

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
var readOnlyVariables []string = []string{
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
				return sqlcmderrors.ReadOnlyVariable(key)
			}
		}
	}
	return nil
}

// Sets or adds the value in the map.
func (v Variables) Set(name, value string) {
	key := strings.ToUpper(name)
	v[key] = value
}

// Removes the value from the map
func (v Variables) Unset(name string) {
	key := strings.ToUpper(name)
	delete(v, key)
}

func (v Variables) All() map[string]string {
	return map[string]string(v)
}

func (v Variables) SqlCmdUser() string {
	return v[SQLCMDUSER]
}

func (v Variables) SqlCmdServer() (serverName string, instance string, port uint64, err error) {
	serverName = v[SQLCMDSERVER]
	return util.SplitServer(serverName)
}
func (v Variables) SqlCmdDatabase() string {
	return v[SQLCMDDBNAME]
}

func (v Variables) UseAad() bool {
	return strings.EqualFold(v[SQLCMDUSEAAD], "true")
}

func (v Variables) Password() string {
	return v[SQLCMDPASSWORD]
}

// ColumnSeparator can have 0 or 1 characters
func (v Variables) ColumnSeparator() string {
	sep := v[SQLCMDCOLSEP]
	if len(sep) > 1 {
		return sep[:1]
	}
	return sep
}

func (v Variables) MaxFixedColumnWidth() int64 {
	w := v[SQLCMDMAXFIXEDTYPEWIDTH]
	return mustValue(w)
}

func (v Variables) MaxVarColumnWidth() int64 {
	w := v[SQLCMDMAXVARTYPEWIDTH]
	return mustValue(w)
}

func (v Variables) ScreenWidth() int64 {
	w := v[SQLCMDCOLWIDTH]
	return mustValue(w)
}

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

// Initializes variables with default values.
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
		SQLCMDLOGINTIMEOUT:      "8",
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

// Implements the :Setvar command
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

func ValidIdentifier(name string) error {
	if strings.HasPrefix(name, "$(") || strings.ContainsAny(name, "'\"\t\n\r ") {
		return sqlcmderrors.InvalidCommandError(":setvar", 0)
	}
	return nil
}
