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
	"github.com/microsoft/go-sqlcmd/internal/io/file"
	"github.com/microsoft/go-sqlcmd/internal/io/folder"
	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"github.com/microsoft/go-sqlcmd/internal/sql"

	"github.com/rdegges/go-ipify"

	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Open defines the `sqlcmd open` sub-commands
type Deploy struct {
	cmdparser.Cmd

	target  string
	notFree bool
	force   bool
}

func (c *Deploy) DefineCommand(...cmdparser.CommandOptions) {
	options := cmdparser.CommandOptions{
		Use:   "deploy",
		Short: localizer.Sprintf("Deploy current current context to a target environment"),
		Run:   c.run,
	}

	c.Cmd.DefineCommand(options)

	c.AddFlag(cmdparser.FlagOptions{
		String:    &c.target,
		Name:      "target",
		Shorthand: "t",
		Usage:     localizer.Sprintf("Target environment (e.g. azure, fabric")})

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

	// If we're not logged in, log in
	if cmd.ProcessState.ExitCode() == 1 {
		cmd := exec.Command("azd", "auth", "login")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		cmd.Wait()
	}

	var stdout bytes.Buffer

	cmd = exec.Command("azd", "auth", "token", "--output", "json")
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()

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
	username := strings.Split(email, "@")[0]
	username = strings.ToLower(username)

	// If the file azure.yaml does not exist in current directory
	if _, err := os.Stat("azure.yaml"); os.IsNotExist(err) {
		{
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
			dabConfigString := file.GetContents("dab-config.json")
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
		}

		cmd = exec.Command("azd", "init", "--environment", username+"-"+current_contextName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Start()
		cmd.Wait()

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
			mainBicep = strings.Replace(mainBicep, "module monitoring", mainBicepDbCall+"\n\nmodule monitoring", 1)

			mainBicep = strings.Replace(mainBicep, "module samplePages", "output DATA_API_BUILDER_ENDPOINT string = dataApiBuilder.outputs.uri\n\nmodule samplePages", 1)

			// If windows then replace \r\n with \n
			keyvaultBicep = strings.Replace(keyvaultBicep, "\n", "\r\n", -1)
			mainBicep = strings.Replace(mainBicep, keyvaultBicep, "/*\n"+keyvaultBicep+"*/\n", 1)

			mainBicep = strings.Replace(mainBicep, "output AZURE_KEY_VAULT_NAME",
				"// output AZURE_KEY_VAULT_NAME", 1)

			mainBicep = strings.Replace(mainBicep, "output AZURE_KEY_VAULT_ENDPOINT",
				"// output AZURE_KEY_VAULT_ENDPOINT", 1)

			//mainBicep += "\noutput DATA_API_BUILDER_ENDPOINT string = dataApiBuilder.outputs.uri\n"
			mainBicep += "\noutput AZURE_CLIENT_ID string = dataApiBuilder.outputs.managedUserIdentity\n"

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

		{
			var mainParamsJson map[string]interface{}
			mainParamsString := file.GetContents("infra\\main.parameters.json")
			json.Unmarshal([]byte(mainParamsString), &mainParamsJson)

			// Append a parameter sqlAdminLoginName to the parameters object
			mainParamsJson["parameters"].(map[string]interface{})["sqlAdminLoginName"] = map[string]interface{}{
				"value": "${AZURE_PRINCIPAL_NAME}",
			}

			mainParamsJson["parameters"].(map[string]interface{})["sqlClientIpAddress"] = map[string]interface{}{
				"value": "${MY_IP}",
			}

			mainParamsJson["parameters"].(map[string]interface{})["useFreeLimit"] = map[string]interface{}{
				"value": "${USE_FREE_LIMIT}",
			}

			var arr []interface{}
			/*arr = append(arr, map[string]interface {
			}{
				"name":  "AZURE_CLIENT_ID",
				"value": "${AZURE_PRINCIPAL_ID}",
			})

			arr = append(arr, map[string]interface {
			}{
				"name":  "CONN_STRING",
				"value": "${CONN_STRING}",
			})
			*/

			arr = append(arr, map[string]interface {
			}{
				"name":  "ASPNETCORE_ENVIRONMENT",
				"value": "Development",
			})

			/*
				arr = append(arr, map[string]interface {
				}{
					"name":  "DAB_MEMORY",
					"value": "0.5Gi",
				})
			*/

			mainParamsJson["parameters"].(map[string]interface {
			})["dataApiBuilderDefinition"].(map[string]interface {
			})["value"].(map[string]interface {
			})["settings"] = arr

			var arr2 []interface{}

			/*arr2 = append(arr2, map[string]interface {
			}{
				"name":  "AZURE_CLIENT_ID",
				"value": "${AZURE_PRINCIPAL_ID}",
			})
			*/
			/*
				arr2 = append(arr2, map[string]interface {
				}{
					"name":  "GRAPHQL_ENDPOINT",
					"value": "${DATA_API_BUILDER_ENDPOINT}/graphql",
				})
			*/

			arr2 = append(arr2, map[string]interface {
			}{
				"name":  "ASPNETCORE_ENVIRONMENT",
				"value": "Development",
			})

			mainParamsJson["parameters"].(map[string]interface {
			})["samplePagesDefinition"].(map[string]interface {
			})["value"].(map[string]interface {
			})["settings"] = arr2

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
	}

	// BUGBUG: Do this using a microsoft bless method (SSMS/ADS must do this in the connection dialogs)
	output.Infof("Discovering IP address for this client, to allow firewall access to the Azure SQL server")

	ip, err := ipify.GetIp()
	if err != nil {
		panic(err)
	}

	output.Infof("Setting local Address to %q to have access to the Azure SQL server", ip)

	cmd = exec.Command("azd", "env", "set", "MY_IP", ip)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	cmd = exec.Command("azd", "env", "set", "AZURE_PRINCIPAL_NAME", email)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	cmd = exec.Command("azd", "env", "set", "USE_FREE_LIMIT", fmt.Sprintf("%t", !c.notFree))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	cmd = exec.Command("azd", "provision")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

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

	cmd = exec.Command("azd", "env", "set", "CONN_STRING",
		fmt.Sprintf("Server=sql-%s.database.windows.net; Database=sample; Authentication=Active Directory Default; Encrypt=True", random))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	// Read the file back in (now that CONN_STRING) is there
	envFile, err = godotenv.Read(".azure\\" + defaultEnvironment + "\\.env")
	if err != nil {
		panic("Unable to read .env file: " + filename)
	}

	// Save the .env to the DataApiBuilder so it is included in the docker image
	godotenv.Write(envFile, "DataApiBuilder\\.env")

	endpoint := sqlconfig.Endpoint{
		EndpointDetails: sqlconfig.EndpointDetails{
			Address: "sql-" + random + ".database.windows.net",
			Port:    1433,
		},
	}

	user := sqlconfig.User{
		Name:               email,
		AuthenticationType: "ActiveDirectoryDefault",
		//AuthenticationType: "ActiveDirectoryInteractive",
	}

	options := sql.ConnectOptions{Database: "sample", Interactive: false}

	options.LogLevel = 255

	s := sql.NewSql(sql.SqlOptions{})
	s.Connect(endpoint, &user, sql.ConnectOptions{Database: "sample", Interactive: false})
	s.Query("use [sample]")
	s.Query("DROP USER IF EXISTS [id-dataapibuild-" + random + "]")
	s.Query("CREATE USER [id-dataapibuild-" + random + "] FROM EXTERNAL PROVIDER")
	s.Query("ALTER ROLE db_datareader ADD MEMBER [id-dataapibuild-" + random + "]")
	s.Query("ALTER ROLE db_datawriter ADD MEMBER [id-dataapibuild-" + random + "]")

	// If folder .sqlcmd exists
	if file.Exists(".sqlcmd") {

		//for each file in folder .sqlcmd
		files, err := os.ReadDir(".sqlcmd")
		if err != nil {
			output.Fatalf("Error reading .sqlcmd folder: %v", err)
		}
		for _, f := range files {
			//if file is .sql
			if strings.HasSuffix(f.Name(), ".sql") {
				//run sql file
				output.Infof("Running %q", f.Name())
				s.ExecuteSqlFile(".sqlcmd/" + f.Name())
			}
		}
	}

	cmd = exec.Command("azd", "package")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	cmd = exec.Command("azd", "deploy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Start()
	cmd.Wait()

	output.InfofWithHintExamples([][]string{
		{localizer.Sprintf("To view the deployed resources"), "azd show"},
		{localizer.Sprintf("To setup a deployment pipeline"), "azd pipeline config --help"},
		{localizer.Sprintf("To delete all resource in Azure"), "azd down"},
	}, localizer.Sprintf("Successfully deployed application to %q", c.target))

}

// LocalIP get the host machine local IP address
func GetOutboundIP() (string, error) {
	resp, err := http.Get("https://ifconfig.me")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var dataApiBuilderCsProj = `<Project Sdk="Microsoft.NET.Sdk.Web">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
    <UserSecretsId>69229d98-6579-4a11-a8ad-cc511e35d465</UserSecretsId>
  </PropertyGroup>
</Project>
`

var dockerfile = `FROM mcr.microsoft.com/azure-databases/data-api-builder:latest

COPY dab-config.json /App
COPY .env /App
WORKDIR /App
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

// output CONN_STRING string = sqlServer.outputs.connectionString
`

var dbBicep = `param name string
param location string = resourceGroup().location
param tags object = {}

param databaseName string = ''
param sqlAdminLoginName string
param sqlAdminLoginObjectId string
param sqlClientIpAddress string
param useFreeLimit bool

// Because databaseName is optional in main.bicep, we make sure the database name is set here.
var defaultDatabaseName = 'sample'
var actualDatabaseName = !empty(databaseName) ? databaseName : defaultDatabaseName

module sqlServer '../core/database/sqlserver/sqlserver.bicep' = {
  name: 'sqlserver'
  params: {
    name: name
    location: location
    tags: tags
    databaseName: actualDatabaseName
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
// output CONN_STRING string = connectionString
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
