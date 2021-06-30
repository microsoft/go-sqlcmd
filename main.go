package main

import (
	//"database/sql"

	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/microsoft/go-sqlcmd/errors"
)

// The exhaustive list is at https://docs.microsoft.com/sql/tools/sqlcmd-utility?view=sql-server-ver15
type SqlCmdArguments struct {
	BatchTerminator        string   `short:"c" default:"GO" arghelp:"Specifies the batch terminator. The default value is GO."`
	TrustServerCertificate bool     `short:"C" help:"Implicitly trust the server certificate without validation."`
	DatabaseName           string   `short:"d" help:"This option sets the sqlcmd scripting variable SQLCMDDBNAME. This parameter specifies the initial database. The default is your login's default-database property. If the database does not exist, an error message is generated and sqlcmd exits."`
	UseTrustedConnection   bool     `short:"E" xor:"uid" help:"Uses a trusted connection instead of using a user name and password to sign in to SQL Server, ignoring any any environment variables that define user name and password."`
	Password               string   `short:"P" xor:"pwd" env:"SQLCMDPASSWORD" help:"User-specified password."`
	UserName               string   `short:"U" xor:"uid" env:"SQLCMDUSER"  help:"The login name or contained database user name.  For contained database users, you must provide the database name option"`
	InputFile              []string `short:"i" type:"existingFile" help:"Identifies one or more files that contain batches of SQL statements. If one or more files do not exist, sqlcmd will exit. Mutually exclusive with -Q/-q."`
	OutputFile             string   `short:"o" type:"path" help:"Identifies the file that receives output from sqlcmd."`
	InitialQuery           string   `short:"q" help:"Executes a query when sqlcmd starts, but does not exit sqlcmd when the query has finished running. Multiple-semicolon-delimited queries can be executed."`
	Query                  string   `short:"Q" help:"Executes a query when sqlcmd starts and then immediately exits sqlcmd. Multiple-semicolon-delimited queries can be executed."`
	Server                 string   `short:"S" env:"SQLCMDSERVER" default:"." help:"[tcp:]server[\\instance_name][,port]Specifies the instance of SQL Server to which to connect. It sets the sqlcmd scripting variable SQLCMDSERVER."`
	Port                   uint64   `kong:"-"`
	Instance               string   `kong:"-"`
}

var Args SqlCmdArguments

// Constructs an URL-style connection string from the SqlCmdArguments structure.
// If the input structure has an instance or port in the server name, it modifies
// the structure on output to remove those decorations from the Server property
// and updates the Port and Instance fields appropriately.
// The URL connection string format supports the entirety of allowed characters and
// is easily encoded/decoded, unlike the ADO or odbc strings.
// go-mssqldb doesn't support quoted values or values with semi-colons in the ADO style strings
func connectionString(args *SqlCmdArguments) (connectionString string, err error) {
	if err = validate(args); err != nil {
		return "", err
	}

	serverName := "."
	if args.Server != "" {
		serverName = args.Server
	}

	query := url.Values{}
	connectionUrl := &url.URL{
		Scheme: "sqlserver",
		Path:   args.Instance,
	}
	if !args.UseTrustedConnection {
		connectionUrl.User = url.UserPassword(args.UserName, args.Password)
	}
	if args.Port > 0 {
		connectionUrl.Host = fmt.Sprintf("%s:%d", serverName, args.Port)
	} else {
		connectionUrl.Host = serverName
	}
	if args.DatabaseName != "" {
		query.Add("database", args.DatabaseName)
	}
	if args.TrustServerCertificate {
		query.Add("trustservercertificate", "true")
	}
	connectionUrl.RawQuery = query.Encode()
	return connectionUrl.String(), nil
}

// Validates combinations not covered by kong.
// Parses the port number from the server name and replaces the server name with the minimal version
// Processing of environment variables and default values must have occurred before this is called.
func validate(args *SqlCmdArguments) error {
	if args.Server != "" {
		serverName := args.Server
		if strings.HasPrefix(serverName, "tcp:") {
			if len(args.Server) == 4 {
				return &errors.InvalidServerName
			}
			serverName = serverName[4:]
		}
		serverNameParts := strings.Split(serverName, ",")
		if len(serverNameParts) > 2 {
			return &errors.InvalidServerName
		}
		if len(serverNameParts) == 2 {
			var err error
			args.Port, err = strconv.ParseUint(serverNameParts[1], 10, 16)
			if err != nil {
				return &errors.InvalidServerName
			}
			serverName = serverNameParts[0]
		} else {
			serverNameParts = strings.Split(serverName, "/")
			if len(serverNameParts) > 2 {
				return &errors.InvalidServerName
			}
			if len(serverNameParts) == 2 {
				args.Instance = serverNameParts[1]
				serverName = serverNameParts[0]
			}
		}
		args.Server = serverName
	}

	if !args.UseTrustedConnection && args.UserName == "" {
		args.UseTrustedConnection = true
	}
	return nil
}

func main() {
	kong.Parse(&Args)
	connectionString, err := connectionString(&Args)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(connectionString)
	}
}
