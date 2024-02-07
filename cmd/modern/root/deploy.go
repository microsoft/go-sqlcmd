// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/dotsqlcmdconfig"
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/sql"

	"github.com/rdegges/go-ipify"

	"os"
	"os/exec"
	"strings"
)

// Open defines the `sqlcmd open` sub-commands
type Deploy struct {
	cmdparser.Cmd

	target      string
	environment string
	notFree     bool
	force       bool
}

func (c *Deploy) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "deploy",
		Short: localizer.Sprintf("Deploy current current context to a target environment"),
		Run:   c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.target,
		DefaultString: "azure",
		Name:          "target",
		Shorthand:     "t",
		Usage:         localizer.Sprintf("Target cloud platform (azure, fabric)")})

	c.AddFlag(cmdparser.FlagOptions{
		String:        &c.environment,
		DefaultString: "",
		Name:          "environment",
		Shorthand:     "e",
		Usage:         localizer.Sprintf("Target environment name (default {username}-{sqlcmd-context-name})")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.notFree,
		Name:  "not-free",
		Usage: localizer.Sprintf("Use not free SKUs")})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.force,
		Name:  "force",
		Usage: localizer.Sprintf("Remove existing azure.yaml and infra folder")})
}

func (c *Deploy) run() {
	output := c.Output()

	current_contextName := config.CurrentContextName()
	if current_contextName == "" {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To view available contexts"), "sqlcmd config get-contexts"},
		}, localizer.Sprintf("No current context"))
	}

	if c.force {
		if file.Exists("azure.yaml") {
			file.Remove("azure.yaml")
		}
		folder.RemoveAll("infra")
		folder.RemoveAll(".azure")
		folder.RemoveAll("DataApiBuilder")
	}

	// Check to see if we're already logged in (this will return 1 if not
	cmd := exec.Command("azd", "auth", "token", "--output", "json")
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.Fatal(localizer.Sprintf("Error checking if logged in to azd"))
	}

	// If we're not logged in, log in
	if cmd.ProcessState.ExitCode() == 1 {
		cmd := exec.Command("azd", "auth", "login")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		cmd.Wait()
		if cmd.ProcessState.ExitCode() != 0 {
			output.Fatal(localizer.Sprintf("Error logging in to azd"))
		}
	}

	var stdout bytes.Buffer

	cmd = exec.Command("azd", "auth", "token", "--output", "json")
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.Fatal(localizer.Sprintf("Error getting token from azd auth token"))
	}

	tokenBlob := stdout.String()

	var token string
	{
		var payloadJson map[string]interface{}
		json.Unmarshal([]byte(tokenBlob), &payloadJson)
		token = payloadJson["token"].(string)
	}

	split := strings.Split(token, ".")

	// base64 decode the payload
	payload, _ := base64.StdEncoding.DecodeString(split[1])
	payloadstring := string(payload)
	if payloadstring[len(payloadstring)-1:] != "}" {
		payloadstring += "}" // BUGBUG: Why do I need to do this!
	}

	var email string
	{
		var payloadJson map[string]interface{}
		json.Unmarshal([]byte(payloadstring), &payloadJson)

		// test is the key in the json
		if payloadJson["email"] != nil {
			email = payloadJson["email"].(string)
		} else if payloadJson["upn"] != nil {
			email = payloadJson["upn"].(string)
		} else {
			panic("Unable to get principal name from token")
		}
	}

	output.Info("Using principal name: " + email)

	// get the string up to the @ from email
	parts := strings.Split(email, "@")
	username := parts[0]
	username = strings.ToLower(username)
	domain := parts[1]
	domain = strings.ToLower(domain)

	if domain == "hotmail.com" || domain == "live.com" || domain == "outlook.com" {
		output.FatalWithHints([]string{
			localizer.Sprintf("TEMP: The go-mssqldb driver is unable to auth using consumer accounts.  Use a corporate account (non-consumer, e.g. not hotmail.com etc.).  Run `azd auth logout` & `azd auth login alias@microsoft.com`"),
			localizer.Sprintf("azd logged in using a consumer domain: %q", domain),
		})
	}

	dotsqlcmdconfig.SetFileName(dotsqlcmdconfig.DefaultFileName())
	dotsqlcmdconfig.Load()

	// If the file azure.yaml does not exist in current directory
	if _, err := os.Stat("azure.yaml"); os.IsNotExist(err) {
		addons := dotsqlcmdconfig.AddonTypes()

		for i, addon := range addons {
			if addon == "dab" {
				folder.MkdirAll("DataApiBuilder")

				f := file.OpenFile("DataApiBuilder\\DataApiBuilder.csproj")
				f.WriteString(dataApiBuilderCsProj)
				f.Close()

				f = file.OpenFile("DataApiBuilder\\Dockerfile")
				f.WriteString(dockerfile)
				f.Close()

				f = file.OpenFile("DataApiBuilder\\Program.cs")
				f.Close()

				var dabConfigJson map[string]interface{}

				files := dotsqlcmdconfig.AddonFiles(i)

				if len(files) == 1 {
					dabConfigString := file.GetContents(files[0])
					json.Unmarshal([]byte(dabConfigString), &dabConfigJson)

					dataSource := dabConfigJson["data-source"]
					dataSource.(map[string]interface{})["connection-string"] = "@env('CONN_STRING')"

					newData, err := json.Marshal(dabConfigJson)
					if err != nil {
						panic(err)
					}

					var prettyJSON bytes.Buffer
					json.Indent(&prettyJSON, newData, "", "    ")

					f = file.OpenFile("DataApiBuilder\\dab-config.json")
					f.WriteString(prettyJSON.String())
					f.Close()
				} else {
					panic("There should be exactly one dab-config.json file specified as a 'use' in the .sqlcmd file")
				}
			}
		}

		if c.environment == "" {
			c.environment = username + "-" + current_contextName
		}

		cmd = exec.Command("azd", "init", "--environment", c.environment)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Start()
		cmd.Wait()
	}

	if file.Exists("next-steps.md") {
		file.Remove("next-steps.md")
	}

	{
		f := file.OpenFile("infra\\app\\db.bicep")
		f.WriteString(dbBicep)
		f.Close()

		folder.MkdirAll("infra\\core\\database\\sqlserver")
		f = file.OpenFile("infra\\core\\database\\sqlserver\\sqlserver.bicep")
		f.WriteString(sqlserverBicep)
		f.Close()
	}

	{
		mainBicep := file.GetContents("infra\\main.bicep")
		mainBicep = strings.Replace(mainBicep, "module monitoring",
			mainBicepDbCall+"\n\nmodule monitoring", 1)

		// If windows then replace \r\n with \n
		keyvaultBicep = strings.Replace(keyvaultBicep, "\n", "\r\n", -1)
		mainBicep = strings.Replace(mainBicep, keyvaultBicep, "/*\n"+keyvaultBicep+"*/\n", 1)

		mainBicep = strings.Replace(mainBicep, "output AZURE_KEY_VAULT_NAME",
			"// output AZURE_KEY_VAULT_NAME", 1)

		mainBicep = strings.Replace(mainBicep, "output AZURE_KEY_VAULT_ENDPOINT",
			"// output AZURE_KEY_VAULT_ENDPOINT", 1)

		mainBicep += "\noutput DATA_API_BUILDER_ENDPOINT string = dataApiBuilder.outputs.uri\noutput AZURE_CLIENT_ID string = dataApiBuilder.outputs.managedUserIdentity\n"

		f := file.OpenFile("infra\\main.bicep")
		f.WriteString(string(mainBicep))
		f.Close()
	}

	{
		dabBicep := file.GetContents("infra\\app\\DataApiBuilder.bicep")
		dabBicep = strings.Replace(dabBicep, "cpu: json('1.0')", "cpu: json('0.25')", 1)
		dabBicep = strings.Replace(dabBicep, "memory: '2.0Gi'", "memory: '0.5Gi'", 1)

		dabBicep += "\noutput managedUserIdentity string = identity.properties.clientId\n"

		f := file.OpenFile("infra\\app\\DataApiBuilder.bicep")
		f.WriteString(string(dabBicep))
		f.Close()
	}

	databases := dotsqlcmdconfig.DatabaseNames()
	if len(databases) == 0 {
		panic("No databases found in .sqlcmd file")
	}
	databaseName := databases[0]

	{
		var mainParamsJson map[string]interface{}
		mainParamsString := file.GetContents("infra\\main.parameters.json")
		json.Unmarshal([]byte(mainParamsString), &mainParamsJson)

		// Append a parameter sqlAdminLoginName to the parameters object
		mainParamsJson["parameters"].(map[string]interface{})["sqlAdminLoginName"] = map[string]interface{}{
			"value": "${AZURE_PRINCIPAL_NAME}",
		}

		mainParamsJson["parameters"].(map[string]interface{})["sqlDatabaseName"] = map[string]interface{}{
			"value": databaseName,
		}

		mainParamsJson["parameters"].(map[string]interface{})["sqlClientIpAddress"] = map[string]interface{}{
			"value": "${MY_IP}",
		}

		mainParamsJson["parameters"].(map[string]interface{})["useFreeLimit"] = map[string]interface{}{
			"value": "${USE_FREE_LIMIT}",
		}

		var arr []interface{}
		arr = append(arr, map[string]interface {
		}{
			"name":  "ASPNETCORE_ENVIRONMENT",
			"value": "Development",
		})

		mainParamsJson["parameters"].(map[string]interface {
		})["dataApiBuilderDefinition"].(map[string]interface {
		})["value"].(map[string]interface {
		})["settings"] = arr

		newData, err := json.Marshal(mainParamsJson)
		if err != nil {
			panic(err)
		}

		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, newData, "", "  ")

		f := file.OpenFile("infra\\main.parameters.json")
		f.WriteString(prettyJSON.String())
		f.Close()
	}

	// BUGBUG: Do this using a microsoft bless method (SSMS/ADS must do this in the connection dialogs)
	output.Infof("Discovering IP address for this client, to allow firewall access to the Azure SQL server")

	ip, err := ipify.GetIp()
	output.FatalErr(err)

	output.Infof("Setting local Address to %q to have access to the Azure SQL server", ip)

	cmd = exec.Command("azd", "env", "set", "MY_IP", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.Fatal(localizer.Sprintf("Error setting environment variable MY_IP"))
	}

	cmd = exec.Command("azd", "env", "set", "AZURE_PRINCIPAL_NAME", email)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.Fatal(localizer.Sprintf("Error setting environment variable AZURE_PRINCIPAL_NAME"))
	}

	cmd = exec.Command("azd", "env", "set", "USE_FREE_LIMIT", fmt.Sprintf("%t", !c.notFree))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.Fatal(localizer.Sprintf("Error setting environment variable USE_FREE_LIMIT"))
	}

	cmd = exec.Command("azd", "provision")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()
	if cmd.ProcessState.ExitCode() != 0 {
		output.FatalWithHintExamples([][]string{
			{localizer.Sprintf("To clean up any resources created"), "azd down --force"},
			{localizer.Sprintf("To not create an Azure SQL 'Spinnaker' run again with --not-free"), "sqlcmd deploy --not-free"},
			{localizer.Sprintf("If failed with 'Invalid value given for parameter ExternalAdministratorLoginName'"), "sqlcmd deploy (but `azd auth login` with a corp (non-consumer, e.g. not hotmail.com etc.) account)"},
		}, localizer.Sprintf("Error provisioning infrastructure"))
	}

	var defaultEnvironment string
	{
		var payloadJson map[string]interface{}
		configJson := file.GetContents(".azure\\config.json")
		json.Unmarshal([]byte(configJson), &payloadJson)
		if payloadJson["defaultEnvironment"] != nil {
			defaultEnvironment = payloadJson["defaultEnvironment"].(string)
		} else {
			panic("Unable to get defaultEnvironment from .azure\\config.json")
		}
	}

	// Run the SQL scripts
	filename := ".azure\\" + defaultEnvironment + "\\.env"

	envFile, err := godotenv.Read(".azure\\" + defaultEnvironment + "\\.env")
	if err != nil {
		panic("Unable to read .env file: " + filename)
	}

	random, ok := envFile["AZURE_CONTAINER_REGISTRY_ENDPOINT"]
	if !ok {
		panic("AZURE_CONTAINER_REGISTRY_ENDPOINT is not set in .env")
	}
	if random == "" {
		panic("AZURE_CONTAINER_REGISTRY_ENDPOINT is not set in .env")
	}

	// Get the word up to the first .
	random = strings.Split(random, ".")[0]

	if len(random) < 2 {
		panic("Random is too short, the value is '" + random + "'")
	}

	// Remove the first two characters from random (the 'cr' which stands for Container Registry)
	random = random[2:]

	azureClientId, ok := envFile["AZURE_CLIENT_ID"]
	if !ok {
		panic("AZURE_CLIENT_ID is not set in .env")
	}
	if azureClientId == "" {
		panic("AZURE_CLIENT_ID is not set in .env")
	}

	f := file.OpenFile("DataApiBuilder\\Dockerfile")
	f.WriteString(
		fmt.Sprintf(dockerfile,
			fmt.Sprintf("Server=sql-%s.database.windows.net; Database=%s; Authentication=Active Directory Default; Encrypt=True",
				random,
				databaseName),
			azureClientId))
	f.Close()

	endpoint := sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "sql-" + random + ".database.windows.net",
			Port:    1433,
		},
	}

	user := sqlconfig.User{
		Name:               email,
		AuthenticationType: "ActiveDirectoryDefault",
	}

	// options := sql.ConnectOptions{Database: databaseName, Interactive: false}
	// options.LogLevel = 255

	s := sql.NewSql(sql.SqlOptions{})
	s.Connect(endpoint, &user, sql.ConnectOptions{Database: databaseName, Interactive: false})
	s.Query("DROP USER IF EXISTS [id-dataapibuild-" + random + "]")
	s.Query("CREATE USER [id-dataapibuild-" + random + "] FROM EXTERNAL PROVIDER")
	s.Query("ALTER ROLE db_datareader ADD MEMBER [id-dataapibuild-" + random + "]")
	s.Query("ALTER ROLE db_datawriter ADD MEMBER [id-dataapibuild-" + random + "]")

	// If folder .sqlcmd exists
	if file.Exists(".sqlcmd") {

		dotsqlcmdconfig.SetFileName(dotsqlcmdconfig.DefaultFileName())
		dotsqlcmdconfig.Load()
		files := dotsqlcmdconfig.DatabaseFiles(0)

		//for each file in folder .sqlcmd
		for _, fi := range files {
			//if file is .sql
			if strings.HasSuffix(fi, ".sql") {
				//run sql file
				output.Infof("Running %q", fi)
				s.ExecuteSqlFile(fi)
			} else {
				panic(fmt.Sprintf("File %q is not supported", fi))
			}
		}

		cmd = exec.Command("azd", "package")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Start()
		cmd.Wait()
		if cmd.ProcessState.ExitCode() != 0 {
			output.Fatal(localizer.Sprintf("Error packaging application"))
		}

		cmd = exec.Command("azd", "deploy")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Start()
		cmd.Wait()
		if cmd.ProcessState.ExitCode() != 0 {
			output.Fatal(localizer.Sprintf("Error deploying application"))
		}

		output.InfofWithHintExamples([][]string{
			{localizer.Sprintf("To view the deployed resources"), "azd show"},
			{localizer.Sprintf("To setup a deployment pipeline"), "azd pipeline config --help"},
			{localizer.Sprintf("To delete all resource in Azure"), "azd down --force"},
		}, localizer.Sprintf("Successfully deployed application to %q", c.target))

	}
}

var dataApiBuilderCsProj = `<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>
`

var dockerfile = `FROM mcr.microsoft.com/azure-databases/data-api-builder:latest

COPY dab-config.json /App
WORKDIR /App
ENV CONN_STRING='%s'
ENV AZURE_CLIENT_ID=%s
ENV ASPNETCORE_URLS=http://+:5000
EXPOSE 5000
ENTRYPOINT ["dotnet", "Azure.DataApiBuilder.Service.dll"]`

var mainBicepDbCall = `param sqlDatabaseName string = ''
param sqlServerName string = ''

param sqlAdminLoginName string
param sqlClientIpAddress string
param useFreeLimit bool

// The application database
module sqlServer './app/db.bicep' = {
  name: 'sql'
  scope: rg
  params: {
    name: !empty(sqlServerName) ? sqlServerName : '${abbrs.sqlServers}${resourceToken}'
    databaseName: sqlDatabaseName
    location: location
    tags: tags
    sqlAdminLoginName: sqlAdminLoginName
    sqlAdminLoginObjectId: principalId
    sqlClientIpAddress: sqlClientIpAddress
    useFreeLimit: useFreeLimit
  }
}
`

var dbBicep = `param name string
param location string = resourceGroup().location
param tags object = {}

param databaseName string = ''
param sqlAdminLoginName string
param sqlAdminLoginObjectId string
param sqlClientIpAddress string
param useFreeLimit bool

module sqlServer '../core/database/sqlserver/sqlserver.bicep' = {
  name: 'sqlserver'
  params: {
    name: name
    location: location
    tags: tags
    databaseName: databaseName
    sqlAdminLoginName: sqlAdminLoginName
    sqlAdminLoginObjectId: sqlAdminLoginObjectId
    sqlClientIpAddress: sqlClientIpAddress
    useFreeLimit: useFreeLimit
  }
}

output connectionString string = sqlServer.outputs.connectionString
`

var sqlserverBicep = `metadata description = 'Creates an Azure SQL Server instance.'
param name string
param location string = resourceGroup().location
param tags object = {}

param sqlAdminLoginName string
param sqlAdminLoginObjectId string
param sqlClientIpAddress string
param databaseName string
param useFreeLimit bool

resource sqlServer 'Microsoft.Sql/servers@2023-05-01-preview' = {
  name: name
  location: location
  tags: tags
  properties: {
    administrators: {
      azureADOnlyAuthentication: true
      administratorType: 'ActiveDirectory'
      tenantId: subscription().tenantId
      principalType: 'User'
      login: sqlAdminLoginName
      sid: sqlAdminLoginObjectId
    }
    version: '12.0'
    minimalTlsVersion: '1.2'
    publicNetworkAccess: 'Enabled'
  }
  identity: {
    type: 'SystemAssigned'
  }

  resource database 'databases' = {
    name: databaseName
    location: location
    sku: {
      name: 'GP_S_Gen5_2'
    }
    properties: {
      useFreeLimit: useFreeLimit
    }
  }

  resource firewall 'firewallRules' = {
    name: 'Azure Services'
    properties: {
      // Note: range [0.0.0.0-0.0.0.0] means "allow all Azure-hosted clients only".
      startIpAddress: '0.0.0.0'
      endIpAddress: '0.0.0.0'
    }
  }
}

resource clientFirewallRules 'Microsoft.Sql/servers/firewallRules@2023-05-01-preview' = {
  name: 'AllowClientIp'
  parent: sqlServer
  properties: {
    startIpAddress: sqlClientIpAddress
    endIpAddress: sqlClientIpAddress
  }
}

var connectionString = 'Server=${sqlServer.properties.fullyQualifiedDomainName}; Database=${sqlServer::database.name}; Authentication=Active Directory Default; Encrypt=True'
output connectionString string = connectionString
output databaseName string = sqlServer::database.name
`

var keyvaultBicep = `module keyVault './shared/keyvault.bicep' = {
  name: 'keyvault'
  params: {
    location: location
    tags: tags
    name: '${abbrs.keyVaultVaults}${resourceToken}'
    principalId: principalId
  }
  scope: rg
}`
