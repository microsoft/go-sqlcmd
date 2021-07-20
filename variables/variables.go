package variables

import (
	"os"
	"strconv"
	"strings"

	"github.com/microsoft/go-sqlcmd/sqlcmderrors"
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
	if strings.HasPrefix(serverName, "tcp:") {
		if len(serverName) == 4 {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		serverName = serverName[4:]
	}
	serverNameParts := strings.Split(serverName, ",")
	if len(serverNameParts) > 2 {
		return "", "", 0, &sqlcmderrors.InvalidServerName
	}
	if len(serverNameParts) == 2 {
		var err error
		port, err = strconv.ParseUint(serverNameParts[1], 10, 16)
		if err != nil {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		serverName = serverNameParts[0]
	} else {
		serverNameParts = strings.Split(serverName, "/")
		if len(serverNameParts) > 2 {
			return "", "", 0, &sqlcmderrors.InvalidServerName
		}
		if len(serverNameParts) == 2 {
			instance = serverNameParts[1]
			serverName = serverNameParts[0]
		}
	}
	return serverName, instance, port, nil
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

// Initializes variables with default values.
// When fromEnvironment is true, then loads from the runtime environment
func InitializeVariables(fromEnvironment bool) *Variables {
	variables = Variables{
		SQLCMDCOLSEP:            " ",
		SQLCMDCOLDWIDTH:         "0",
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
