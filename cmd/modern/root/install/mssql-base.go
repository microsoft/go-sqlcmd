// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package install

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/microsoft/go-sqlcmd/cmd/modern/root/open"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser/dependency"
	"github.com/microsoft/go-sqlcmd/internal/dotsqlcmdconfig"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"github.com/microsoft/go-sqlcmd/internal/tools"
	"path/filepath"
	"runtime"
	"strconv"
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

	useUrl       []string
	useMechanism []string

	openTool string
	openFile string

	network  string
	addOn    []string
	addOnUse []string

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
		StringArray: &c.useUrl,
		Name:        "using",
		Hidden:      true,
		Usage:       localizer.Sprintf("[DEPRECATED use --use] Download %q and use database", ingest.ValidFileExtensions()),
	})

	addFlag(cmdparser.FlagOptions{
		StringArray: &c.useUrl,
		Name:        "use",
		Usage:       localizer.Sprintf("Download %q and use database", ingest.ValidFileExtensions()),
	})

	addFlag(cmdparser.FlagOptions{
		StringArray: &c.useMechanism,
		Name:        "use-mechanism",
		Usage:       localizer.Sprintf("Mechanism to use to bring database online (%s)", strings.Join(mechanism.Mechanisms(), ",")),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.openTool,
		DefaultString: "",
		Name:          "open",
		Usage:         localizer.Sprintf("Open tool e.g. ads"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.openFile,
		DefaultString: "",
		Name:          "open-file",
		Usage:         localizer.Sprintf("Open file in tool e.g. https://aks.ms/adventureworks-demo.sql"),
	})

	addFlag(cmdparser.FlagOptions{
		String:        &c.network,
		DefaultString: "",
		Name:          "network",
		Usage:         localizer.Sprintf("Container network name (defaults to 'container-network' if --add-on specified)"),
	})

	addFlag(cmdparser.FlagOptions{
		StringArray:   &c.addOn,
		DefaultString: "",
		Name:          "add-on",
		Usage:         localizer.Sprintf("Create add-on container (i.e. dab, fleet-manager)"),
	})

	addFlag(cmdparser.FlagOptions{
		StringArray:   &c.addOnUse,
		DefaultString: "",
		Name:          "add-on-use",
		Usage:         localizer.Sprintf("File to use for add-on container"),
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

	c.GetValuesFromDotSqlcmd()

	c.createContainer(imageName, c.contextName)
}

func (c *MssqlBase) GetValuesFromDotSqlcmd() {
	if file.Exists(filepath.Join(".sqlcmd", "sqlcmd.yaml")) {
		dotsqlcmdconfig.SetFileName(dotsqlcmdconfig.DefaultFileName())

		// If there is a .sqlcmd/sqlcmd.yaml file, then load that up and use it for any values not provided
		dotsqlcmdconfig.Load()
		dbs := dotsqlcmdconfig.DatabaseNames()

		if len(dbs) > 1 {
			panic("Only a single database is supported at this time")
		}

		if len(dbs) > 0 {
			if c.defaultDatabase == "" {
				c.defaultDatabase = dbs[0]
			}

			if len(c.useUrl) == 0 {
				c.useUrl = append(c.useUrl, dotsqlcmdconfig.DatabaseFiles(0)...)
			}
		}

		addons := dotsqlcmdconfig.AddonTypes()

		if len(addons) > 0 {
			c.addOn = append(c.addOn, addons...)
		}

		for i, _ := range c.addOn {
			if len(c.addOnUse) < i+1 {
				c.addOnUse = append(c.addOnUse, dotsqlcmdconfig.AddonFiles(i)...)
			}
		}
	}
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
	userName := pal.UserName()
	password := c.generatePassword()

	contextName = config.FindUniqueContextName(contextName, userName)

	if c.port == 0 {
		c.port = config.FindFreePort(1433)
	}

	// Do an early exit if url doesn't exist
	var useUrls []ingest.Ingest
	if len(c.useUrl) > 0 {
		useUrls = c.verifyUseSourceFileExists(controller, output)
	}

	if len(useUrls) == 1 {
		if useUrls[0].UserProvidedFileExt() == "git" {
			useUrls[0].BringOnline(nil, "", "")
			useUrls = nil
			c.useUrl = nil // Blank this out, because we now will get more useUrls from the .sqlcmd/sqlcmd.yaml file
			c.useMechanism = nil
		}
	}

	// Now that we have any remote repo cloned local, now is the time to
	// go look for the .sqlcmd/sqlcmd.yaml settings
	c.GetValuesFromDotSqlcmd()

	if len(c.useUrl) > 0 {
		useUrls = c.verifyUseSourceFileExists(controller, output)
	}

	if c.defaultDatabase != "" {
		if !c.validateDbName(c.defaultDatabase) {
			output.Fatalf(localizer.Sprintf("--database %q contains non-ASCII chars and/or quotes", c.defaultDatabase))
		}
	}

	// If an add-on is specified, and no network name, then set a default network name
	if len(c.addOn) > 0 && c.network == "" {
		c.network = "sqlcmd-" + contextName + "-network"
	}

	if c.network != "" {
		// Create a docker network
		if !controller.NetworkExists(c.network) {
			output.Info(localizer.Sprintf("Creating %q, for add-on cross container communication", c.network))
			controller.NetworkCreate(c.network)
		}
	}

	// Very strange issue that we need to work here.  If we are using add-on containers
	// we have to specify the name of the mssql container!
	// Details in this bug/DCR here:
	//		https://github.com/moby/moby/issues/45183
	if c.name == "" {
		c.name = contextName + "-container"
	}

	if !c.useCached {
		c.downloadImage(imageName, output, controller)
	}

	runOptions := container.RunOptions{
		PortInternal: 1433,
		Port:         c.port,
		Name:         c.name,
		Hostname:     c.hostname,
		Architecture: c.architecture,
		Os:           c.os,
		Network:      c.network,
	}

	runOptions.Env = []string{
		"ACCEPT_EULA=Y",
		fmt.Sprintf("MSSQL_SA_PASSWORD=%s", saPassword),
		fmt.Sprintf("MSSQL_COLLATION=%s", c.collation)}

	output.Info(localizer.Sprintf("Starting %q", imageName))

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
		Password:           password,
		PasswordEncryption: c.passwordEncryption,
		Network:            c.network}
	config.AddContextWithContainer(contextName, contextOptions)

	output.Infof(
		localizer.Sprintf("Created context %q in \"%s\", configuring user account",
			config.CurrentContextName(),
			config.GetConfigFileUsed()))

	controller.ContainerWaitForLogEntry(containerId, c.errorLogEntryToWaitFor)

	output.Info(
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
	c.sql.Connect(
		endpoint,
		saUser,
		sql.ConnectOptions{Database: "master", Interactive: false},
	)

	// Create a new (non-sa) SQL Server user
	c.createUser(contextOptions.Username, contextOptions.Password)

	// Download and restore/attach etc. DB if asked
	if len(useUrls) > 0 {
		for i, useUrl := range useUrls {
			if useUrl.IsRemoteUrl() {
				if useUrl.UserProvidedFileExt() != "git" {
					output.Infof("Downloading %q to container", useUrl.UrlFilename())
				}
			} else {
				output.Infof("Copying %q to container", useUrl.UrlFilename())
			}
			useUrl.CopyToContainer(containerId)

			if useUrl.IsExtractionNeeded() {
				output.Infof("Extracting files from %q archive", useUrl.UrlFilename())
				useUrl.Extract()
			}

			output.Infof("Bringing database %q online (%s)", useUrl.DatabaseName(), useUrl.OnlineMethod())

			// Connect to master, unless a default database was specified (at this point the default database
			// has not been set yet, so we need to specify it in the -d statement
			databaseToConnectTo := "master"
			if c.defaultDatabase != "" {
				databaseToConnectTo = c.defaultDatabase
			}
			if useUrl.OnlineMethod() == "script" {
				runSqlcmdInContainer := func(query string) {
					cmd := []string{
						"/opt/mssql-tools/bin/sqlcmd",
						"-S",
						"localhost",
						"-U",
						contextOptions.Username,
						"-P",
						contextOptions.Password,
						"-d",
						databaseToConnectTo,
						"-i",
						"/var/opt/mssql/backup/" + useUrl.UrlFilename(),
					}

					controller.RunCmdInContainer(containerId, cmd, container.ExecOptions{})
				}
				useUrl.BringOnline(runSqlcmdInContainer, contextOptions.Username, contextOptions.Password)
			} else {
				useUrl.BringOnline(c.sql.Query, contextOptions.Username, contextOptions.Password)
			}

			for _, f := range dotsqlcmdconfig.DatabaseFiles(i) {
				//if file is .sql
				if strings.HasSuffix(f, ".sql") {
					//run sql file
					output.Infof("Running %q", f)
					c.sql.ExecuteSqlFile(f)
				}
			}
		}

	}

	dabPort := 0
	fleetManagerPort := 0
	for i, addOn := range c.addOn {

		if addOn == "fleet-manager" {
			dabImageName := "fleet-manager"

			if !c.useCached {
				c.downloadImage(dabImageName, output, controller)
			}

			fleetManagerPort = config.FindFreePort(8080)

			fleetManagerRunOptions := container.RunOptions{
				PortInternal: 80,
				Port:         fleetManagerPort,
				Architecture: c.architecture,
				Os:           c.os,
				Network:      c.network,
			}

			addOnContainerId := controller.ContainerRun(
				dabImageName,
				fleetManagerRunOptions,
			)

			contextName := config.CurrentContextName()

			// Save add-on details to config file now, so it can be deleted even
			// if something below fails
			config.AddAddOn(
				contextName,
				"fleet-manager",
				addOnContainerId,
				dabImageName,
				"127.0.0.1",
				fleetManagerPort)
		}

		if addOn == "dab" {
			dabImageName := "mcr.microsoft.com/azure-databases/data-api-builder"

			if !c.useCached {
				c.downloadImage(dabImageName, output, controller)
			}

			dabPort = config.FindFreePort(5000)

			dabRunOptions := container.RunOptions{
				PortInternal: 5000,
				Port:         dabPort,
				Architecture: c.architecture,
				Os:           c.os,
				Network:      c.network,
			}

			addOnContainerId := controller.ContainerRun(
				dabImageName,
				dabRunOptions,
			)

			contextName := config.CurrentContextName()

			// Save add-on details to config file now, so it can be deleted even
			// if something below fails
			config.AddAddOn(
				contextName,
				"dab",
				addOnContainerId,
				dabImageName,
				"127.0.0.1",
				dabPort)

			if len(c.addOnUse) >= i+1 {

				var dabConfigJson map[string]interface{}
				dabConfigString := file.GetContents(c.addOnUse[i])
				json.Unmarshal([]byte(dabConfigString), &dabConfigJson)

				dataSource := dabConfigJson["data-source"]
				dataSource.(map[string]interface{})["connection-string"] = fmt.Sprintf("Server=%s;Database=%s;User ID=%s;Password=%s;TrustServerCertificate=true",
					c.name,
					c.defaultDatabase,
					userName,
					password)

				newData, err := json.Marshal(dabConfigJson)
				if err != nil {
					panic(err)
				}

				var prettyJSON bytes.Buffer
				json.Indent(&prettyJSON, newData, "", "    ")

				folder.RemoveAll("tmp-dab-config.json")
				folder.MkdirAll("tmp-dab-config.json")

				f := file.OpenFile(filepath.Join("tmp-dab-config.json", "dab-config.json"))
				f.WriteString(prettyJSON.String())
				f.Close()

				// Download the dab-config file to the container
				controller.CopyFile(
					addOnContainerId,
					filepath.Join("tmp-dab-config.json", "dab-config.json"),
					"/App",
				)

				folder.RemoveAll("tmp-dab-config.json")
			}

			// Restart the container, now that the dab-config file is there
			controller.ContainerStop(addOnContainerId)
			controller.ContainerStart(addOnContainerId)
		}
	}

	hints := [][]string{}

	// TODO: sqlcmd open ads only supported on Windows and Mac right now, add Linux support
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

	for _, addOn := range c.addOn {
		if addOn == "fleet-manager" {
			hints = append(hints, []string{
				localizer.Sprintf("Fleet Manager (Renzo) API UI"),
				fmt.Sprintf("http://localhost:%d/swagger/index.html", fleetManagerPort),
			})
		}

		if addOn == "dab" {
			hints = append(hints, []string{
				localizer.Sprintf("Data API Builder (DAB) Health Status"),
				fmt.Sprintf("curl -s http://localhost:%d", dabPort),
			})
		}

		if addOn == "fleet-manager" {
			output.Info(localizer.Sprintf("Now ready for Fleet Manager connections on port %v",
				strconv.Itoa(fleetManagerPort)),
			)
		}

		if addOn == "dab" {
			output.Info(localizer.Sprintf("Now ready for DAB connections on port %v",
				strconv.Itoa(dabPort)),
			)
		}
	}

	output.InfofWithHintExamples(hints,
		localizer.Sprintf("Now ready for SQL connections on port %v",
			strconv.Itoa(c.port)),
	)

	if c.openTool == "ads" {
		ads := open.Ads{}
		ads.SetCrossCuttingConcerns(dependency.Options{
			EndOfLine: pal.LineBreak(),
			Output:    c.Output(),
		})

		user := &sqlconfig.User{
			AuthenticationType: "basic",
			BasicAuth: &sqlconfig.BasicAuthDetails{
				Username:           contextOptions.Username,
				PasswordEncryption: c.passwordEncryption,
				Password:           secret.Encode(contextOptions.Password, c.passwordEncryption)},
			Name: contextOptions.Username}

		ads.PersistCredentialForAds(endpoint.EndpointDetails.Address, endpoint, user)

		output := c.Output()
		args := []string{"-r", fmt.Sprintf("--server=%s", fmt.Sprintf(
			"%s,%d",
			"127.0.0.1",
			c.port))}

		args = append(args, fmt.Sprintf("--user=%s",
			strings.Replace(contextOptions.Username, `"`, `\"`, -1)))

		tool := tools.NewTool("ads")
		if !tool.IsInstalled() {
			output.Fatalf(tool.HowToInstall())
		}

		// BUGBUG: This should come from: displayPreLaunchInfo
		output.Info(localizer.Sprintf("Press Ctrl+C to exit this process..."))

		_, err := tool.Run(args)
		c.CheckErr(err)
	}

	if c.openTool == "vscode" {
		tool := tools.NewTool("vscode")
		if !tool.IsInstalled() {
			output.Fatalf(tool.HowToInstall())
		}

		// BUGBUG: This should come from: displayPreLaunchInfo
		output.Info(localizer.Sprintf("Launching Visual Studio Code..."))

		_, err := tool.Run([]string{"."}) // []string{"."}) BUGBUG
		c.CheckErr(err)
	}
}

func (c *MssqlBase) verifyUseSourceFileExists(
	controller *container.Controller,
	output *output.Output) (useUrls []ingest.Ingest) {

	for i, url := range c.useUrl {

		mechanism := ""
		if len(c.useMechanism) > i {
			mechanism = c.useMechanism[i]
		}

		useUrls = append(useUrls, ingest.NewIngest(url, controller, ingest.IngestOptions{
			Mechanism:    mechanism,
			DatabaseName: c.defaultDatabase,
		}))

		if !useUrls[i].IsValidFileExtension() {
			output.FatalWithHints(
				[]string{
					fmt.Sprintf(
						localizer.Sprintf("--use must be a path to a file with a %q extension"),
						ingest.ValidFileExtensions(),
					),
				},
				localizer.Sprintf("%q is not a valid file extension for --use flag"), useUrls[i].UserProvidedFileExt())
		}

		if useUrls[i].IsRemoteUrl() && !useUrls[i].IsValidScheme() {
			output.FatalfWithHints(
				[]string{
					fmt.Sprintf(
						localizer.Sprintf("--use URL must one of %q"),
						strings.Join(useUrls[i].ValidSchemes(), ", "),
					),
				},
				localizer.Sprintf("%q is not a valid URL for --use flag", useUrls[i].UrlFilename()))
		}

		if !useUrls[i].SourceFileExists() {
			output.FatalfWithHints(
				[]string{localizer.Sprintf("File does not exist at URL %q", useUrls[i].UrlFilename())},
				"Unable to download file")
		}
	}

	return useUrls
}

// createUser creates a user (non-sa) and assigns the sysadmin role
// to the user. It also creates a default database with the provided name
// and assigns the default database to the user. Finally, it disables
// the sa account and rotates the sa password for security reasons.
func (c *MssqlBase) createUser(
	userName string,
	password string,
) {
	const createLogin = `CREATE LOGIN [%s]
WITH PASSWORD=N'%s',
DEFAULT_DATABASE=[%s],
CHECK_EXPIRATION=OFF,
CHECK_POLICY=OFF`
	const addSrvRoleMember = `EXEC master..sp_addsrvrolemember
@loginame = N'%s',
@rolename = N'sysadmin'`

	output := c.Cmd.Output()
	defaultDatabase := "master"

	if c.defaultDatabase != "" {
		defaultDatabase = c.defaultDatabase

		// Create the default database, if it isn't a downloaded database
		output.Infof(localizer.Sprintf("Creating default database %q", defaultDatabase))
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
	output.Info(localizer.Sprintf("Downloading %q", imageName))
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
