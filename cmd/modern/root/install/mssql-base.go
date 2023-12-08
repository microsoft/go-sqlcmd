// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/http"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"github.com/spf13/viper"
)

// MssqlBase provide base support for installing SQL Server and all of its
// various flavors, e.g. SQL Server Edge.
type MssqlBase struct {
	cmdparser.Cmd

	tag             string
	registry        string
	repo            string
	acceptEula      bool
	contextName     string
	defaultDatabase string

	passwordLength         int
	passwordMinSpecial     int
	passwordMinNumber      int
	passwordMinUpper       int
	passwordSpecialCharSet string
	passwordEncryption     string

	useCached              bool
	errorLogEntryToWaitFor string
	defaultContextName     string
	collation              string

	name         string
	hostname     string
	architecture string
	os           string

	port int

	usingDatabaseUrl string

	unitTesting bool

	sql sql.Sql
}

func (c *MssqlBase) AddFlags(
	addFlag func(cmdparser.FlagOptions),
	repo string,
	defaultContextName string,
) {
	c.defaultContextName = defaultContextName

	addFlag(cmdparser.FlagOptions{
		String:        &c.registry,
		Name:          "registry",
		DefaultString: "mcr.microsoft.com",
		Usage:         "Container registry",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.repo,
		Name:          "repo",
		DefaultString: repo,
		Usage:         "Container repository",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.tag,
		Name:          "tag",
		DefaultString: "latest",
		Usage:         localizer.Sprintf("Tag to use, use get-tags to see list of tags"),
	})

	addFlag(cmdparser.FlagOptions{
		String:    &c.contextName,
		Name:      "context-name",
		Shorthand: "c",
		Usage:     localizer.Sprintf("Context name (a default context name will be created if not provided)"),
	})

	addFlag(cmdparser.FlagOptions{
		String:    &c.defaultDatabase,
		Name:      "user-database",
		Shorthand: "u",
		Usage:     localizer.Sprintf("Create a user database and set it as the default for login"),
	})

	addFlag(cmdparser.FlagOptions{
		Bool:  &c.acceptEula,
		Name:  "accept-eula",
		Usage: localizer.Sprintf("Accept the SQL Server EULA"),
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordLength,
		DefaultInt: 50,
		Name:       "password-length",
		Usage:      localizer.Sprintf("Generated password length"),
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinSpecial,
		DefaultInt: 10,
		Name:       "password-min-special",
		Usage:      localizer.Sprintf("Minimum number of special characters"),
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinNumber,
		DefaultInt: 10,
		Name:       "password-min-number",
		Usage:      localizer.Sprintf("Minimum number of numeric characters"),
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinUpper,
		DefaultInt: 10,
		Name:       "password-min-upper",
		Usage:      localizer.Sprintf("Minimum number of upper characters"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.passwordSpecialCharSet,
		DefaultString: "!@#$%&*",
		Name:          "password-special-chars",
		Usage:         localizer.Sprintf("Special character set to include in password"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.passwordEncryption,
		DefaultString: "none",
		Name:          "password-encryption",
		Usage: localizer.Sprintf("Password encryption method (%s) in sqlconfig file",
			secret.EncryptionMethodsForUsage()),
	})

	addFlag(cmdparser.FlagOptions{
		Bool:  &c.useCached,
		Name:  "cached",
		Usage: localizer.Sprintf("Don't download image.  Use already downloaded image"),
	})

	// BUG(stuartpa): SQL Server bug: "SQL Server is now ready for client connections", oh no it isn't!!
	// Wait for "Server name is" instead!  Nope, that doesn't work on edge
	// Wait for "The default language" instead
	// BUG(stuartpa): This obviously doesn't work for non US LCIDs
	addFlag(cmdparser.FlagOptions{
		String:        &c.errorLogEntryToWaitFor,
		DefaultString: "The default language",
		Name:          "errorlog-wait-line",
		Usage:         localizer.Sprintf("Line in errorlog to wait for before connecting"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.name,
		DefaultString: "",
		Name:          "name",
		Usage:         localizer.Sprintf("Specify a custom name for the container rather than a randomly generated one"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.hostname,
		DefaultString: "",
		Name:          "hostname",
		Usage:         localizer.Sprintf("Explicitly set the container hostname, it defaults to the container ID"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.architecture,
		DefaultString: "amd64",
		Name:          "architecture",
		Usage:         localizer.Sprintf("Specifies the image CPU architecture"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.os,
		DefaultString: "linux",
		Name:          "os",
		Usage:         localizer.Sprintf("Specifies the image operating system"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.collation,
		DefaultString: "SQL_Latin1_General_CP1_CI_AS",
		Name:          "collation",
		Usage:         "The SQL Server collation",
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.port,
		DefaultInt: 0,
		Name:       "port",
		Usage:      localizer.Sprintf("Port (next available port from 1433 upwards used by default)"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.usingDatabaseUrl,
		DefaultString: "",
		Name:          "using",
		Usage:         localizer.Sprintf("Download (into container) and attach database (.bak) from URL"),
	})
}

// Run checks that the end-user license agreement has been accepted,
// constructs the container image name from the provided registry, repository, and tag,
// and sets the context name to a default value if it is not provided.
// Then, it creates the image as a container and names it using the context name.
// Once the container is running, if a database backup file is provided, it is downloaded,
// restored attached.
// If the EULA has not been accepted, it prints an error message with suggestions for how to proceed,
// and exits the program.
func (c *MssqlBase) Run() {
	output := c.Cmd.Output()

	var imageName string

	if !c.acceptEula && viper.GetString("ACCEPT_EULA") == "" {
		output.FatalWithHints(
			[]string{localizer.Sprintf("Either, add the %s flag to the command-line", localizer.AcceptEulaFlag),
				localizer.Sprintf("Or, set the environment variable i.e. %s %s=YES ", pal.CreateEnvVarKeyword(), localizer.AcceptEulaEnvVar)},
			localizer.Sprintf("EULA not accepted"))
	}

	imageName = fmt.Sprintf(
		"%s/%s:%s",
		c.registry,
		c.repo,
		c.tag)

	if c.contextName == "" {
		c.contextName = c.defaultContextName
	}

	c.createContainer(imageName, c.contextName)
}

// createContainer installs an image for a SQL Server container. The image
// is specified by imageName, and the container will be given the name contextName.
// If the useCached flag is set, the function will skip downloading the image
// from the internet. The function outputs progress messages to the command-line
// as it runs. If any errors are encountered, they will be printed to the
// command-line and the program will exit.
func (c *MssqlBase) createContainer(imageName string, contextName string) {
	output := c.Cmd.Output()
	saPassword := c.generatePassword()

	env := []string{
		"ACCEPT_EULA=Y",
		fmt.Sprintf("MSSQL_SA_PASSWORD=%s", saPassword),
		fmt.Sprintf("MSSQL_COLLATION=%s", c.collation),
	}

	if c.port == 0 {
		c.port = config.FindFreePortForTds()
	}

	// Do an early exit if url doesn't exist
	if c.usingDatabaseUrl != "" {
		c.validateUsingUrlExists()
	}

	if c.defaultDatabase != "" {
		if !c.validateDbName(c.defaultDatabase) {
			output.Fatalf(localizer.Sprintf("--user-database %q contains non-ASCII chars and/or quotes", c.defaultDatabase))
		}
	}

	controller := container.NewController()

	if !c.useCached {
		c.downloadImage(imageName, output, controller)
	}

	output.Info(localizer.Sprintf("Starting %v", imageName))
	containerId := controller.ContainerRun(
		imageName,
		env,
		c.port,
		c.name,
		c.hostname,
		c.architecture,
		c.os,
		[]string{},
		false,
	)
	previousContextName := config.CurrentContextName()

	userName := pal.UserName()
	password := c.generatePassword()

	// Save the config now, so user can uninstall/delete, even if mssql in the container
	// fails to start
	config.AddContextWithContainer(
		contextName,
		imageName,
		c.port,
		containerId,
		userName,
		password,
		c.passwordEncryption,
	)

	output.Info(
		localizer.Sprintf("Created context %q in \"%s\", configuring user account...",
			config.CurrentContextName(),
			config.GetConfigFileUsed()))

	controller.ContainerWaitForLogEntry(
		containerId, c.errorLogEntryToWaitFor)

	output.Info(
		localizer.Sprintf("Disabled %q account (and rotated %q password). Creating user %q",
			"sa",
			"sa",
			userName))

	endpoint, _ := config.CurrentContext()

	// Connect to the instance as `sa` so we can create a new user
	//
	// For Unit Testing we use the Docker Hello World container, which
	// starts much faster than the SQL Server container!
	if c.errorLogEntryToWaitFor == "Hello from Docker!" {
		c.sql = sql.New(sql.SqlOptions{UnitTesting: true})
	} else {
		c.sql = sql.New(sql.SqlOptions{UnitTesting: false})
	}

	saUser := &sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: c.passwordEncryption,
			Password:           secret.Encode(saPassword, c.passwordEncryption)},
		Name: "sa"}

	c.sql.Connect(endpoint, saUser, sql.ConnectOptions{Database: "master", Interactive: false})

	c.createNonSaUser(userName, password)

	// Download and restore DB if asked
	if c.usingDatabaseUrl != "" {
		c.downloadAndRestoreDb(
			controller,
			containerId,
			userName,
		)
	}

	hints := [][]string{}

	// TODO: sqlcmd open ads only support on Windows right now, add Mac support
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		hints = append(hints, []string{localizer.Sprintf("Open in Azure Data Studio"), "sqlcmd open ads"})
	}

	hints = append(hints, []string{localizer.Sprintf("Run a query"), "sqlcmd query \"SELECT @@version\""})
	hints = append(hints, []string{localizer.Sprintf("Start interactive session"), "sqlcmd query"})

	if previousContextName != "" {
		hints = append(
			hints,
			[]string{localizer.Sprintf("Change current context"), fmt.Sprintf(
				"sqlcmd config use-context %v",
				previousContextName,
			)},
		)
	}

	hints = append(hints, []string{localizer.Sprintf("View sqlcmd configuration"), "sqlcmd config view"})
	hints = append(hints, []string{localizer.Sprintf("See connection strings"), "sqlcmd config connection-strings"})
	hints = append(hints, []string{localizer.Sprintf("Remove"), "sqlcmd delete"})

	output.InfoWithHintExamples(hints,
		localizer.Sprintf("Now ready for client connections on port %#v",
			c.port),
	)
}

func (c *MssqlBase) validateUsingUrlExists() {
	output := c.Cmd.Output()
	databaseUrl := extractUrl(c.usingDatabaseUrl)
	u, err := url.Parse(databaseUrl)
	c.CheckErr(err)

	if u.Scheme != "http" && u.Scheme != "https" {
		output.FatalWithHints(
			[]string{
				localizer.Sprintf("--using URL must be http or https"),
			},
			localizer.Sprintf("%q is not a valid URL for --using flag", c.usingDatabaseUrl))
	}

	if u.Path == "" {
		output.FatalWithHints(
			[]string{
				localizer.Sprintf("--using URL must have a path to .bak file"),
			},
			localizer.Sprintf("%q is not a valid URL for --using flag", c.usingDatabaseUrl))
	}

	// At the moment we only support attaching .bak files, but we should
	// support .bacpacs and .mdfs in the future
	if _, file := filepath.Split(u.Path); filepath.Ext(file) != ".bak" {
		output.FatalWithHints(
			[]string{
				localizer.Sprintf("--using file URL must be a .bak file"),
			},
			localizer.Sprintf("Invalid --using file type"))
	}

	// Verify the url actually exists, and early exit if it doesn't
	urlExists(databaseUrl, output)
}

func (c *MssqlBase) query(commandText string) {
	c.sql.Query(commandText)
}

// createNonSaUser creates a user (non-sa) and assigns the sysadmin role
// to the user. It also creates a default database with the provided name
// and assigns the default database to the user. Finally, it disables
// the sa account and rotates the sa password for security reasons.
func (c *MssqlBase) createNonSaUser(
	userName string,
	password string,
) {
	output := c.Cmd.Output()

	defaultDatabase := "master"

	if c.defaultDatabase != "" {
		defaultDatabase = c.defaultDatabase

		// Create the default database, if it isn't a downloaded database
		output.Info(localizer.Sprintf("Creating default database [%s]", defaultDatabase))
		c.query(fmt.Sprintf("CREATE DATABASE [%s]", defaultDatabase))
	}

	const createLogin = `CREATE LOGIN [%s]
WITH PASSWORD=N'%s',
DEFAULT_DATABASE=[%s],
CHECK_EXPIRATION=OFF,
CHECK_POLICY=OFF`
	const addSrvRoleMember = `EXEC master..sp_addsrvrolemember
@loginame = N'%s',
@rolename = N'sysadmin'`

	c.query(fmt.Sprintf(createLogin, userName, password, defaultDatabase))
	c.query(fmt.Sprintf(addSrvRoleMember, userName))

	// Correct safety protocol is to rotate the sa password, because the first
	// sa password has been in the docker environment (as SA_PASSWORD)
	c.query(fmt.Sprintf("ALTER LOGIN [sa] WITH PASSWORD = N'%s';",
		c.generatePassword()))
	c.query("ALTER LOGIN [sa] DISABLE")

	if c.defaultDatabase != "" {
		c.query(fmt.Sprintf("ALTER AUTHORIZATION ON DATABASE::[%s] TO %s",
			defaultDatabase, userName))
	}
}

func getDbNameAsIdentifier(dbName string) string {
	escapedDbNAme := strings.ReplaceAll(dbName, "'", "''")
	return strings.ReplaceAll(escapedDbNAme, "]", "]]")
}

func getDbNameAsNonIdentifier(dbName string) string {
	return strings.ReplaceAll(dbName, "]", "]]")
}

// parseDbName returns the databaseName from --using arg
// It sets database name to the specified database name
// or in absence of it, it is set to the filename without
// extension.
func parseDbName(usingDbUrl string) string {
	u, _ := url.Parse(usingDbUrl)
	dbToken := path.Base(u.Path)
	if dbToken != "." && dbToken != "/" {
		lastIdx := strings.LastIndex(dbToken, ".bak")
		if lastIdx != -1 {
			//Get file name without extension
			fileName := dbToken[0:lastIdx]
			lastIdx += 5
			if lastIdx >= len(dbToken) {
				return fileName
			}
			//Return database name if it was specified
			return dbToken[lastIdx:]
		}
	}
	return ""
}

func extractUrl(usingArg string) string {
	urlEndIdx := strings.LastIndex(usingArg, ".bak")
	if urlEndIdx != -1 {
		return usingArg[0:(urlEndIdx + 4)]
	}
	return usingArg
}

func (c *MssqlBase) downloadAndRestoreDb(
	controller *container.Controller,
	containerId string,
	userName string,
) {
	output := c.Cmd.Output()
	databaseName := parseDbName(c.usingDatabaseUrl)
	databaseUrl := extractUrl(c.usingDatabaseUrl)

	_, file := filepath.Split(databaseUrl)

	// Download file from URL into container
	output.Info(localizer.Sprintf("Downloading %s", file))

	temporaryFolder := "/var/opt/mssql/backup"

	controller.DownloadFile(
		containerId,
		databaseUrl,
		temporaryFolder,
	)

	// Restore database from file
	output.Info(localizer.Sprintf("Restoring database %s", databaseName))

	dbNameAsIdentifier := getDbNameAsIdentifier(databaseName)
	dbNameAsNonIdentifier := getDbNameAsNonIdentifier(databaseName)

	text := `SET NOCOUNT ON;

-- Build a SQL Statement to restore any .bak file to the Linux filesystem
DECLARE @sql NVARCHAR(max)

-- This table definition works since SQL Server 2017, therefore 
-- works for all SQL Server containers (which started in 2017)
DECLARE @fileListTable TABLE (
    [LogicalName]           NVARCHAR(128),
    [PhysicalName]          NVARCHAR(260),
    [Type]                  CHAR(1),
    [FileGroupName]         NVARCHAR(128),
    [Size]                  NUMERIC(20,0),
    [MaxSize]               NUMERIC(20,0),
    [FileID]                BIGINT,
    [CreateLSN]             NUMERIC(25,0),
    [DropLSN]               NUMERIC(25,0),
    [UniqueID]              UNIQUEIDENTIFIER,
    [ReadOnlyLSN]           NUMERIC(25,0),
    [ReadWriteLSN]          NUMERIC(25,0),
    [BackupSizeInBytes]     BIGINT,
    [SourceBlockSize]       INT,
    [FileGroupID]           INT,
    [LogGroupGUID]          UNIQUEIDENTIFIER,
    [DifferentialBaseLSN]   NUMERIC(25,0),
    [DifferentialBaseGUID]  UNIQUEIDENTIFIER,
    [IsReadOnly]            BIT,
    [IsPresent]             BIT,
    [TDEThumbprint]         VARBINARY(32),
    [SnapshotURL]           NVARCHAR(360)
)

INSERT INTO @fileListTable
EXEC('RESTORE FILELISTONLY FROM DISK = ''%s/%s''')
SET @sql = 'RESTORE DATABASE [%s] FROM DISK = ''%s/%s'' WITH '
SELECT @sql = @sql + char(13) + ' MOVE ''' + LogicalName + ''' TO ''/var/opt/mssql/' + LogicalName + '.' + RIGHT(PhysicalName,CHARINDEX('\',PhysicalName)) + ''','
FROM @fileListTable
WHERE IsPresent = 1
SET @sql = SUBSTRING(@sql, 1, LEN(@sql)-1)
EXEC(@sql)`

	c.query(fmt.Sprintf(text, temporaryFolder, file, dbNameAsIdentifier, temporaryFolder, file))

	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		userName,
		dbNameAsNonIdentifier)
	c.query(alterDefaultDb)
}

func (c *MssqlBase) downloadImage(
	imageName string,
	output *output.Output,
	controller *container.Controller,
) {
	output.Info(localizer.Sprintf("Downloading %v", imageName))
	err := controller.EnsureImage(imageName)
	if err != nil || c.unitTesting {
		output.FatalErrorWithHints(
			err,
			[]string{
				localizer.Sprintf("Is a container runtime installed on this machine (e.g. Podman or Docker)?") + pal.LineBreak() +
					localizer.Sprintf("\tIf not, download desktop engine from:") + pal.LineBreak() +
					"\t\thttps://podman-desktop.io/" + pal.LineBreak() +
					localizer.Sprintf("\t\tor") + pal.LineBreak() +
					"\t\thttps://docs.docker.com/get-docker/",
				localizer.Sprintf("Is a container runtime running?  (Try `%s` or `%s` (list containers), does it return without error?)", localizer.PodmanPsCommand, localizer.DockerPsCommand),
				localizer.Sprintf("If `podman ps` or `docker ps` works, try downloading the image with:"+pal.LineBreak()+
					"\t`podman|docker pull %s`", imageName)},
			localizer.Sprintf("Unable to download image %s", imageName))
	}
}

// Verify the file exists at the URL
func urlExists(url string, output *output.Output) {
	if !http.UrlExists(url) {
		output.FatalWithHints(
			[]string{localizer.Sprintf("File does not exist at URL")},
			localizer.Sprintf("Unable to download file"))
	}
}

func (c *MssqlBase) generatePassword() (password string) {
	password = secret.Generate(
		c.passwordLength,
		c.passwordMinSpecial,
		c.passwordMinNumber,
		c.passwordMinUpper,
		c.passwordSpecialCharSet)

	return
}

// validateDbName checks if the database name is something that is likely
// to work seamlessly through all tools, connection strings etc.
//
// TODO: Right now this is any ASCII char except for quotes,
// but this needs to be opened up for Kanji characters etc. with a full test suite
// to ensure the database name is valid in all the places it is passed to.
func (c *MssqlBase) validateDbName(s string) bool {
	for _, b := range []byte(s) {
		if b > 127 {
			return false
		}
	}
	return !strings.ContainsAny(s, "'\"`'")
}
