// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/http"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"github.com/spf13/viper"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
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
	encryptPassword        bool

	useCached              bool
	errorLogEntryToWaitFor string
	defaultContextName     string
	collation              string

	name     string
	hostname string

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
		Usage:         "Tag to use, use get-tags to see list of tags",
	})

	addFlag(cmdparser.FlagOptions{
		String:    &c.contextName,
		Name:      "context-name",
		Shorthand: "c",
		Usage:     "Context name (a default context name will be created if not provided)",
	})

	addFlag(cmdparser.FlagOptions{
		String:    &c.defaultDatabase,
		Name:      "user-database",
		Shorthand: "u",
		Usage:     "Create a user database and set it as the default for login",
	})

	addFlag(cmdparser.FlagOptions{
		Bool:  &c.acceptEula,
		Name:  "accept-eula",
		Usage: "Accept the SQL Server EULA",
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordLength,
		DefaultInt: 50,
		Name:       "password-length",
		Usage:      "Generated password length",
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinSpecial,
		DefaultInt: 10,
		Name:       "password-min-special",
		Usage:      "Minimum number of special characters",
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinNumber,
		DefaultInt: 10,
		Name:       "password-min-number",
		Usage:      "Minimum number of numeric characters",
	})

	addFlag(cmdparser.FlagOptions{
		Int:        &c.passwordMinUpper,
		DefaultInt: 10,
		Name:       "password-min-upper",
		Usage:      "Minimum number of upper characters",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.passwordSpecialCharSet,
		DefaultString: "!@#$%&*",
		Name:          "password-special-chars",
		Usage:         "Special character set to include in password",
	})

	c.encryptPasswordFlag(addFlag)

	addFlag(cmdparser.FlagOptions{
		Bool:  &c.useCached,
		Name:  "cached",
		Usage: "Don't download image.  Use already downloaded image",
	})

	// BUG(stuartpa): SQL Server bug: "SQL Server is now ready for client connections", oh no it isn't!!
	// Wait for "Server name is" instead!  Nope, that doesn't work on edge
	// Wait for "The default language" instead
	// BUG(stuartpa): This obviously doesn't work for non US LCIDs
	addFlag(cmdparser.FlagOptions{
		String:        &c.errorLogEntryToWaitFor,
		DefaultString: "The default language",
		Name:          "errorlog-wait-line",
		Usage:         "Line in errorlog to wait for before connecting to disable 'sa' account",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.name,
		DefaultString: "",
		Name:          "name",
		Usage:         "Specify a custom name for the container rather than a randomly generated one",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.hostname,
		DefaultString: "",
		Name:          "hostname",
		Usage:         "Explicitly set the container hostname, it defaults to the container ID",
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
		Usage:      "Port (next available port from 1433 upwards used by default)",
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.usingDatabaseUrl,
		DefaultString: "",
		Name:          "using",
		Usage:         "Download (into container) and attach database (.bak) from URL",
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
			[]string{"Either, add the --accept-eula flag to the command-line",
				fmt.Sprintf("Or, set the environment variable i.e. %s SQLCMD_ACCEPT_EULA=YES ", pal.CreateEnvVarKeyword())},
			"EULA not accepted")
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
			output.Fatalf("--user-database %q contains non-ASCII chars and/or quotes", c.defaultDatabase)
		}
	}

	controller := container.NewController()

	if !c.useCached {
		c.downloadImage(imageName, output, controller)
	}

	output.Infof("Starting %v", imageName)
	containerId := controller.ContainerRun(
		imageName,
		env,
		c.port,
		c.name,
		c.hostname,
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
		c.encryptPassword,
	)

	output.Infof(
		"Created context %q in \"%s\", configuring user account...",
		config.CurrentContextName(),
		config.GetConfigFileUsed())

	controller.ContainerWaitForLogEntry(
		containerId, c.errorLogEntryToWaitFor)

	output.Infof(
		"Disabled %q account (and rotated %q password). Creating user %q",
		"sa",
		"sa",
		userName)

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
			Username:          "sa",
			PasswordEncrypted: c.encryptPassword,
			Password:          secret.Encode(saPassword, c.encryptPassword)},
		Name: "sa"}

	c.sql.Connect(endpoint, saUser, sql.ConnectOptions{Interactive: false})

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
	if runtime.GOOS == "windows" {
		hints = append(hints, []string{"Open in Azure Data Studio", "sqlcmd open ads"})
	}

	hints = append(hints, []string{"Run a query", "sqlcmd query \"SELECT @@version\""})
	hints = append(hints, []string{"Start interactive session", "sqlcmd query"})

	if previousContextName != "" {
		hints = append(
			hints,
			[]string{"Change current context", fmt.Sprintf(
				"sqlcmd config use-context %v",
				previousContextName,
			)},
		)
	}

	hints = append(hints, []string{"View sqlcmd configuration", "sqlcmd config view"})
	hints = append(hints, []string{"See connection strings", "sqlcmd config connection-strings"})
	hints = append(hints, []string{"Remove", "sqlcmd delete"})

	output.InfofWithHintExamples(hints,
		"Now ready for client connections on port %d",
		c.port,
	)
}

func (c *MssqlBase) validateUsingUrlExists() {
	output := c.Cmd.Output()
	u, err := url.Parse(c.usingDatabaseUrl)
	c.CheckErr(err)

	if u.Scheme != "http" && u.Scheme != "https" {
		output.FatalfWithHints(
			[]string{
				"--using URL must be http or https",
			},
			"%q is not a valid URL for --using flag", c.usingDatabaseUrl)
	}

	// At the moment we only support attaching .bak files, but we should
	// support .bacpacs and .mdfs in the future
	if _, file := filepath.Split(c.usingDatabaseUrl); filepath.Ext(file) != ".bak" {
		output.FatalfWithHints(
			[]string{
				"--using file URL must be a .bak file",
			},
			"Invalid --using file type")
	}

	// Verify the url actually exists, and early exit if it doesn't
	urlExists(c.usingDatabaseUrl, output)
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
		output.Infof("Creating default database [%s]", defaultDatabase)
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

func (c *MssqlBase) downloadAndRestoreDb(
	controller *container.Controller,
	containerId string,
	userName string,
) {
	output := c.Cmd.Output()

	u, err := url.Parse(c.usingDatabaseUrl)
	c.CheckErr(err)
	_, file := filepath.Split(c.usingDatabaseUrl)
	fileNameWithNoExt := strings.TrimSuffix(file, filepath.Ext(file))

	// Download file from URL into container
	output.Infof("Downloading %s from %s", file, u.Hostname())

	temporaryFolder := "/var/opt/mssql/backup"

	controller.DownloadFile(
		containerId,
		c.usingDatabaseUrl,
		temporaryFolder,
	)

	// Restore database from file
	output.Infof("Restoring database %s", fileNameWithNoExt)

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

	c.query(fmt.Sprintf(text, temporaryFolder, file, fileNameWithNoExt, temporaryFolder, file))

	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		userName,
		fileNameWithNoExt)
	c.query(alterDefaultDb)
}

func (c *MssqlBase) downloadImage(
	imageName string,
	output *output.Output,
	controller *container.Controller,
) {
	output.Infof("Downloading %v", imageName)
	err := controller.EnsureImage(imageName)
	if err != nil || c.unitTesting {
		output.FatalfErrorWithHints(
			err,
			[]string{
				"Is a container runtime installed on this machine (e.g. Podman or Docker)?" + pal.LineBreak() +
					"\tIf not, download desktop engine from:" + pal.LineBreak() +
					"\t\thttps://podman-desktop.io/" + pal.LineBreak() +
					"\t\tor" + pal.LineBreak() +
					"\t\thttps://docs.docker.com/get-docker/",
				"Is a container runtime running. Try `podman ps` or `docker ps` (list containers), does it return without error?",
				fmt.Sprintf("If `podman ps` or `docker ps` works, try downloading the image with: `podman|docker pull %s`", imageName)},
			"Unable to download image %s", imageName)
	}
}

// Verify the file exists at the URL
func urlExists(url string, output *output.Output) {
	if !http.UrlExists(url) {
		output.FatalfWithHints(
			[]string{"File does not exist at URL"},
			"Unable to download file")
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
