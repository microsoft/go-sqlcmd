package variables

import (
	"os"
	"strings"

	"github.com/microsoft/go-sqlcmd/errors"
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
	SQLCMDCOLDWIDTH         = "SQLCMDCOLDWIDTH"
	SQLCMDERRORLEVEL        = "SQLCMDERRORLEVEL"
	SQLCMDMAXVARTYPEWIDTH   = "SQLCMDMAXVARTYPEWIDTH"
	SQLCMDMAXFIXEDTYPEWIDTH = "SQLCMDMAXFIXEDTYPEWIDTH"
	SQLCMDEDITOR            = "SQLCMDEDITOR"
	SQLCMDUSEAAD            = "SQLCMDUSEAAD"
)

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
				return errors.ReadOnlyVariable(key)
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

func SqlCmdUser() string {
	return variables["SQLCMDUSER"]
}

// Initializes variables with default values.
// When fromEnvironment is true, then loads from the runtime environment
func InitializeVariables(fromEnvironment bool) *Variables {
	variables = Variables{
		SQLCMDUSER:              "",
		SQLCMDPASSWORD:          "",
		SQLCMDSERVER:            "",
		SQLCMDDBNAME:            "",
		SQLCMDLOGINTIMEOUT:      "8",
		SQLCMDSTATTIMEOUT:       "0",
		SQLCMDHEADERS:           "0",
		SQLCMDCOLSEP:            " ",
		SQLCMDCOLDWIDTH:         "0",
		SQLCMDPACKETSIZE:        "4096",
		SQLCMDERRORLEVEL:        "0",
		SQLCMDMAXVARTYPEWIDTH:   "256",
		SQLCMDMAXFIXEDTYPEWIDTH: "0",
		SQLCMDEDITOR:            "",
		SQLCMDINI:               "",
		SQLCMDUSEAAD:            "",
	}
	hostname, _ := os.Hostname()
	variables.Set(SQLCMDWORKSTATION, hostname)

	if fromEnvironment {
		for _, envVar := range os.Environ() {
			varParts := strings.Split(envVar, "=")
			err := ValidIdentifier(varParts[0])
			if err == nil {
				variables.Set(varParts[0], varParts[1])
			}
		}
	}
	return &variables
}

// Implements the :Setvar command
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
		return errors.InvalidCommandError(":setvar", 0)
	}
	return nil
}
