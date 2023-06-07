// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest"
	"github.com/microsoft/go-sqlcmd/pkg/mssqlcontainer/ingest/mechanism"
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

	useDatabaseUrl string
	useMechanism   string

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
		Hidden:    true,
		Usage:     localizer.Sprintf("[DEPRECATED use --database] Create a user database and set it as the default for login"),
	})

	addFlag(cmdparser.FlagOptions{
		String:    &c.defaultDatabase,
		Name:      "database",
		Shorthand: "d",
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
		String:        &c.useDatabaseUrl,
		DefaultString: "",
		Name:          "using",
		Hidden:        true,
		Usage:         localizer.Sprintf("[DEPRECATED use --use] Download %q and use database", ingest.ValidFileExtensions()),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.useDatabaseUrl,
		DefaultString: "",
		Name:          "use",
		Usage:         localizer.Sprintf("Download %q and use database", ingest.ValidFileExtensions()),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.useMechanism,
		DefaultString: "",
		Name:          "use-mechanism",
		Usage:         localizer.Sprintf("Mechanism to use to bring database online (%s)", strings.Join(mechanism.Mechanisms(), ",")),
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
	var imageName string

	output := c.Cmd.Output()

	if !c.acceptEula && viper.GetString("ACCEPT_EULA") == "" {
		output.FatalWithHints(
			[]string{localizer.Sprintf("Either, add the %s flag to the command-line", localizer.AcceptEulaFlag),
				localizer.Sprintf("Or, set the environment variable i.e. %s %s=YES ", pal.CreateEnvVarKeyword(), localizer.AcceptEulaEnvVar)},
			localizer.Sprintf("EULA not accepted"))
	}

	imageName = fmt.Sprintf("%s/%s:%s", c.registry, c.repo, c.tag)

	// If no context name provided, set it to the default (e.g. mssql or edge)
	if c.contextName == "" {
		c.contextName = c.defaultContextName
	}

	c.createContainer(imageName, c.contextName)
}

// createContainer creates a SQL Server container for an image. The image
// is specified by imageName, and the container will be given the name contextName.
// If the useCached flag is set, the function will skip downloading the image
// from the internet. The function outputs progress messages to the command-line
// as it runs. If any errors are encountered, they will be printed to the
// command-line and the program will exit.
func (c *MssqlBase) createContainer(imageName string, contextName string) {
	output := c.Cmd.Output()
	controller := container.NewController()
	saPassword := c.generatePassword()

	if c.port == 0 {
		c.port = config.FindFreePortForTds()
	}

	// Do an early exit if url doesn't exist
	var useDatabase ingest.Ingest
	if c.useDatabaseUrl != "" {
		useDatabase = c.verifyUseSourceFileExists(controller, output)
	}

	if c.defaultDatabase != "" {
		if !c.validateDbName(c.defaultDatabase) {
			output.Fatalf(localizer.Sprintf("--database %q contains non-ASCII chars and/or quotes", c.defaultDatabase))
		}
	}

	if !c.useCached {
		c.downloadImage(imageName, output, controller)
	}

	runOptions := container.RunOptions{
		Port:         c.port,
		Name:         c.name,
		Hostname:     c.hostname,
		Architecture: c.architecture,
		Os:           c.os}

	runOptions.Env = []string{
		"ACCEPT_EULA=Y",
		fmt.Sprintf("MSSQL_SA_PASSWORD=%s", saPassword),
		fmt.Sprintf("MSSQL_COLLATION=%s", c.collation)}

	output.Infof(localizer.Sprintf("Starting %v", imageName))
	containerId := controller.ContainerRun(
		imageName,
		runOptions,
	)
	previousContextName := config.CurrentContextName()

	// Save the config now, so user can uninstall/delete, even if mssql in the container
	// fails to start
	contextOptions := config.ContextOptions{
		ImageName:          imageName,
		PortNumber:         c.port,
		ContainerId:        containerId,
		Username:           pal.UserName(),
		Password:           c.generatePassword(),
		PasswordEncryption: c.passwordEncryption}
	config.AddContextWithContainer(contextName, contextOptions)

	output.Infof(
		localizer.Sprintf("Created context %q in \"%s\", configuring user account",
			config.CurrentContextName(),
			config.GetConfigFileUsed()))

	controller.ContainerWaitForLogEntry(containerId, c.errorLogEntryToWaitFor)

	output.Infof(
		localizer.Sprintf("Disabled %q account (and rotated %q password). Creating user %q",
			"sa",
			"sa",
			contextOptions.Username))

	endpoint, _ := config.CurrentContext()

	// Connect to the instance as `sa` so we can create a new user
	//
	// For Unit Testing we use the Docker Hello World container, which
	// starts much faster than the SQL Server container!
	sqlOptions := sql.SqlOptions{}
	if c.errorLogEntryToWaitFor == "Hello from Docker!" {
		sqlOptions.UnitTesting = true
	}
	c.sql = sql.NewSql(sqlOptions)

	saUser := &sqlconfig.User{
		AuthenticationType: "basic",
		BasicAuth: &sqlconfig.BasicAuthDetails{
			Username:           "sa",
			PasswordEncryption: c.passwordEncryption,
			Password:           secret.Encode(saPassword, c.passwordEncryption)},
		Name: "sa"}

	// Connect to master database on SQL Server in the container as `sa`
	c.sql.Connect(endpoint, saUser, sql.ConnectOptions{Database: "master"})

	// Create a new (non-sa) SQL Server user
	c.createUser(contextOptions.Username, contextOptions.Password)

	// Download and restore/attach etc. DB if asked
	if useDatabase != nil {
		if useDatabase.IsRemoteUrl() {
			output.Infof("Downloading %q to container", useDatabase.UrlFilename())
		} else {
			output.Infof("Copying %q to container", useDatabase.UrlFilename())
		}
		useDatabase.CopyToContainer(containerId)

		output.Infof("Bringing database %q online", useDatabase.DatabaseName())
		useDatabase.BringOnline(c.sql.Query, contextOptions.Username, contextOptions.Password)
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

	output.InfofWithHintExamples(hints,
		localizer.Sprintf("Now ready for client connections on port %d",
			c.port),
	)
}

func (c *MssqlBase) verifyUseSourceFileExists(
	controller *container.Controller,
	output *output.Output,
) (useDatabase ingest.Ingest) {
	useDatabase = ingest.NewIngest(c.useDatabaseUrl, controller, ingest.IngestOptions{
		Mechanism: c.useMechanism,
	})

	if !useDatabase.IsValidFileExtension() {
		output.FatalfWithHints(
			[]string{
				fmt.Sprintf(
					localizer.Sprintf("--use must be a path to a file with a %q extension"),
					ingest.ValidFileExtensions(),
				),
			},
			localizer.Sprintf("%q is not a valid file extension for --use flag"), useDatabase.UserProvidedFileExt())
	}

	if useDatabase.IsRemoteUrl() && !useDatabase.IsValidScheme() {
		output.FatalfWithHints(
			[]string{
				localizer.Sprintf("--use URL must one of %q"),
				strings.Join(useDatabase.ValidSchemes(), ", "),
			},
			localizer.Sprintf("%q is not a valid URL for --use flag", c.useDatabaseUrl))
	}

	if !useDatabase.SourceFileExists() {
		output.FatalfWithHints(
			[]string{localizer.Sprintf("File does not exist at URL %q", c.useDatabaseUrl)},
			"Unable to download file")
	}
	return
}

// createUser creates a user (non-sa) and assigns the sysadmin role
// to the user. It also creates a default database with the provided name
// and assigns the default database to the user. Finally, it disables
// the sa account and rotates the sa password for security reasons.
func (c *MssqlBase) createUser(
	userName string,
	password string,
) {
	output := c.Cmd.Output()

	defaultDatabase := "master"

	if c.defaultDatabase != "" {
		defaultDatabase = c.defaultDatabase

		// Create the default database, if it isn't a downloaded database
		output.Infof(localizer.Sprintf("Creating default database [%s]", defaultDatabase))
		c.sql.Query(fmt.Sprintf("CREATE DATABASE [%s]", defaultDatabase))
	}

	const createLogin = `CREATE LOGIN [%s]
WITH PASSWORD=N'%s',
DEFAULT_DATABASE=[%s],
CHECK_EXPIRATION=OFF,
CHECK_POLICY=OFF`
	const addSrvRoleMember = `EXEC master..sp_addsrvrolemember
@loginame = N'%s',
@rolename = N'sysadmin'`

	if c.defaultDatabase != "" {
		defaultDatabase = c.defaultDatabase

		// Create the default database, if it isn't a downloaded database
		output.Infof("Creating default database [%s]", defaultDatabase)
		c.sql.Query(fmt.Sprintf("CREATE DATABASE [%s]", defaultDatabase))
	}

	c.sql.Query(fmt.Sprintf(createLogin, userName, password, defaultDatabase))
	c.sql.Query(fmt.Sprintf(addSrvRoleMember, userName))

	// Correct safety protocol is to rotate the sa password, because the first
	// sa password has been in the docker environment (as SA_PASSWORD)
	c.sql.Query(fmt.Sprintf("ALTER LOGIN [sa] WITH PASSWORD = N'%s';",
		c.generatePassword()))
	c.sql.Query("ALTER LOGIN [sa] DISABLE")

	if c.defaultDatabase != "" {
		c.sql.Query(fmt.Sprintf("ALTER AUTHORIZATION ON DATABASE::[%s] TO %s",
			defaultDatabase, userName))
	}
}

func (c *MssqlBase) downloadImage(
	imageName string,
	output *output.Output,
	controller *container.Controller,
) {
	output.Infof(localizer.Sprintf("Downloading %v", imageName))
	err := controller.EnsureImage(imageName)
	if err != nil || c.unitTesting {
		output.FatalfErrorWithHints(
			err,
			[]string{
				localizer.Sprintf("Is a container runtime installed on this machine (e.g. Podman or Docker)?") + pal.LineBreak() +
					localizer.Sprintf("\tIf not, download desktop engine from:") + pal.LineBreak() +
					"\t\thttps://podman-desktop.io/" + pal.LineBreak() +
					localizer.Sprintf("\t\tor") + pal.LineBreak() +
					"\t\thttps://docs.docker.com/get-docker/",
				localizer.Sprintf("Is a container runtime running?  (Try `%s` or `%s` (list containers), does it return without error?)", localizer.PodmanPsCommand, localizer.DockerPsCommand),
				fmt.Sprintf("If `podman ps` or `docker ps` works, try downloading the image with:"+pal.LineBreak()+
					"\t`podman|docker pull %s`", imageName)},
			localizer.Sprintf("Unable to download image %s", imageName))
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
