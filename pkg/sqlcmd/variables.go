// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

// Variables provides set and get of sqlcmd scripting variables
type Variables map[string]string

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
	SQLCMDFORMAT            = "SQLCMDFORMAT"
	SQLCMDMAXVARTYPEWIDTH   = "SQLCMDMAXVARTYPEWIDTH"
	SQLCMDMAXFIXEDTYPEWIDTH = "SQLCMDMAXFIXEDTYPEWIDTH"
	SQLCMDEDITOR            = "SQLCMDEDITOR"
	SQLCMDUSEAAD            = "SQLCMDUSEAAD"
)

// builtinVariables are the predefined SQLCMD variables. Their values are printed first by :listvar
var builtinVariables = []string{
	SQLCMDCOLSEP,
	SQLCMDCOLWIDTH,
	SQLCMDDBNAME,
	SQLCMDEDITOR,
	SQLCMDERRORLEVEL,
	SQLCMDFORMAT,
	SQLCMDHEADERS,
	SQLCMDINI,
	SQLCMDLOGINTIMEOUT,
	SQLCMDMAXFIXEDTYPEWIDTH,
	SQLCMDMAXVARTYPEWIDTH,
	SQLCMDPACKETSIZE,
	SQLCMDSERVER,
	SQLCMDSTATTIMEOUT,
	SQLCMDUSEAAD,
	SQLCMDUSER,
	SQLCMDWORKSTATION,
}

// readonlyVariables are variables that can't be changed via :setvar
var readOnlyVariables = []string{
	SQLCMDDBNAME,
	SQLCMDINI,
	SQLCMDPACKETSIZE,
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

// Get returns the value of the named variable
// To distinguish an empty value from an unset value use the bool return value
func (v Variables) Get(name string) (string, bool) {
	key := strings.ToUpper(name)
	s, ok := v[key]
	return s, ok
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
	return splitServer(serverName)
}

// SQLCmdDatabase returns the SQLCMDDBNAME variable value
func (v Variables) SQLCmdDatabase() string {
	return v[SQLCMDDBNAME]
}

// UseAad returns whether the SQLCMDUSEAAD variable value is set to "true"
func (v Variables) UseAad() bool {
	return strings.EqualFold(v[SQLCMDUSEAAD], "true")
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

// ErrorLevel controls the minimum level of errors that are printed
func (v Variables) ErrorLevel() int64 {
	return mustValue(v[SQLCMDERRORLEVEL])
}

// Format is the name of the results format
func (v Variables) Format() string {
	switch v[SQLCMDFORMAT] {
	case "vert", "vertical":
		return "vertical"
	}
	return "horizontal"
}

func mustValue(val string) int64 {
	var n int64
	_, err := fmt.Sscanf(val, "%d", &n)
	if err == nil {
		return n
	}
	panic(err)
}

// defaultVariables defines variables that cannot be removed from the map, only reset
// to their default values.
var defaultVariables = Variables{
	SQLCMDCOLSEP:            " ",
	SQLCMDCOLWIDTH:          "0",
	SQLCMDEDITOR:            "edit.com",
	SQLCMDERRORLEVEL:        "0",
	SQLCMDHEADERS:           "0",
	SQLCMDLOGINTIMEOUT:      "30",
	SQLCMDMAXFIXEDTYPEWIDTH: "0",
	SQLCMDMAXVARTYPEWIDTH:   "256",
	SQLCMDSTATTIMEOUT:       "0",
}

// InitializeVariables initializes variables with default values.
// When fromEnvironment is true, then loads from the runtime environment
func InitializeVariables(fromEnvironment bool) *Variables {
	variables := Variables{
		SQLCMDCOLSEP:            defaultVariables[SQLCMDCOLSEP],
		SQLCMDCOLWIDTH:          defaultVariables[SQLCMDCOLWIDTH],
		SQLCMDDBNAME:            "",
		SQLCMDEDITOR:            defaultVariables[SQLCMDEDITOR],
		SQLCMDERRORLEVEL:        defaultVariables[SQLCMDERRORLEVEL],
		SQLCMDHEADERS:           defaultVariables[SQLCMDHEADERS],
		SQLCMDINI:               "",
		SQLCMDLOGINTIMEOUT:      defaultVariables[SQLCMDLOGINTIMEOUT],
		SQLCMDMAXFIXEDTYPEWIDTH: defaultVariables[SQLCMDMAXFIXEDTYPEWIDTH],
		SQLCMDMAXVARTYPEWIDTH:   defaultVariables[SQLCMDMAXVARTYPEWIDTH],
		SQLCMDPACKETSIZE:        "4096",
		SQLCMDSERVER:            "",
		SQLCMDSTATTIMEOUT:       defaultVariables[SQLCMDSTATTIMEOUT],
		SQLCMDUSER:              "",
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
func (variables *Variables) Setvar(name, value string) error {
	err := ValidIdentifier(name)
	if err == nil {
		if err = variables.checkReadOnly(name); err != nil {
			err = ReadOnlyVariable(name)
		}
	}
	if err != nil {
		return err
	}
	if value == "" {
		if _, ok := variables.Get(name); !ok {
			return UndefinedVariable(name)
		}
		if def, ok := defaultVariables.Get(name); ok {
			value = def
		} else {
			variables.Unset(name)
			return nil
		}
	} else {
		value, err = ParseValue(value)
	}
	if err != nil {
		return err
	}
	variables.Set(name, value)
	return nil
}

const validVariableRunes = "_-"

// ValidIdentifier determines if a given string can be used as a variable name
func ValidIdentifier(name string) error {

	first := true
	for _, c := range name {
		if !unicode.IsLetter(c) && (first || (!unicode.IsDigit(c) && !strings.ContainsRune(validVariableRunes, c))) {
			return fmt.Errorf("Invalid variable identifier %s", name)
		}
		first = false
	}
	return nil
}

// ParseValue returns the string to use as the variable value
// If the string contains a space or a quote, it must be delimited by quotes and literal quotes
// within the value must be escaped by another quote
// "this has a quote "" in it" is valid
// "this has a quote" in it" is not valid
func ParseValue(val string) (string, error) {
	quoted := val[0] == '"'
	err := fmt.Errorf("Invalid variable value %s", val)
	if !quoted {
		if strings.ContainsAny(val, "\t\n\r ") {
			return "", err
		}
		return val, nil
	}
	if len(val) == 1 || val[len(val)-1] != '"' {
		return "", err
	}

	b := new(strings.Builder)
	quoted = false
	r := []rune(val)
loop:
	for i := 1; i < len(r)-1; i++ {
		switch {
		case quoted && r[i] == '"':
			b.WriteRune('"')
			quoted = false
		case quoted && r[i] != '"':
			break loop
		case !quoted && r[i] == '"':
			quoted = true
		default:
			b.WriteRune(r[i])
		}
	}
	if quoted {
		return "", err
	}
	return b.String(), nil
}
