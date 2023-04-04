package mssqlcontainer

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/http"
	output2 "github.com/microsoft/go-sqlcmd/internal/output"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func ValidateUsingUrlExists(
	usingDatabaseUrl string,
	output *output2.Output) {
	databaseUrl := extractUrl(usingDatabaseUrl)
	u, _ := url.Parse(databaseUrl)

	if len(u.Scheme) > 3 {
		if u.Scheme != "http" && u.Scheme != "https" {
			output.FatalfWithHints(
				[]string{
					"--using URL must be http or https",
				},
				"%q is not a valid URL for --using flag", usingDatabaseUrl)
		}
	}

	if len(u.Scheme) > 3 {
		if u.Path == "" {
			output.FatalfWithHints(
				[]string{
					"--using URL must have a path to .bak, .bacpac or .mdf (.7z) file",
				},
				"%q is not a valid URL for --using flag", usingDatabaseUrl)
		}
	}

	var f string
	if len(u.Scheme) > 3 {
		_, f = filepath.Split(u.Path)
	} else {
		_, f = filepath.Split(usingDatabaseUrl)
	}
	if filepath.Ext(f) != ".bak" && filepath.Ext(f) != ".bacpac" && filepath.Ext(f) != ".mdf" && filepath.Ext(f) != ".7z" {
		output.FatalfWithHints(
			[]string{
				"--using file URL must be a .bak, .bacpac, or .mdf (.7z) file",
			},
			"Invalid --using file type, extension %q is not supported", filepath.Ext(f))
	}

	if len(u.Scheme) > 3 {

		// Verify the url actually exists, and early exit if it doesn't
		urlExists(databaseUrl, output)
	}
}

func DownloadAndRestoreDb(
	controller *container.Controller,
	containerId string,
	usingDatabaseUrl string,
	userName string,
	password string,
	query func(commandText string),
	output *output2.Output,
) {
	parsed, _ := url.Parse(usingDatabaseUrl)

	databaseName := parseDbName(usingDatabaseUrl)
	if databaseName == "" {
		panic(fmt.Sprintf("databaseName is empty, failed to parse URL %q", usingDatabaseUrl))
	}

	databaseUrl := extractUrl(usingDatabaseUrl)

	var log_f string
	var log_file string

	_, file := filepath.Split(databaseUrl)

	// Download file from URL into container
	output.Infof("Downloading %s", file)

	var f string
	if len(parsed.Scheme) > 3 {
		_, f = filepath.Split(databaseUrl)
	} else {
		_, f = filepath.Split(usingDatabaseUrl)
	}

	var temporaryFolder string
	if filepath.Ext(f) == ".bak" || filepath.Ext(f) == ".7z" || filepath.Ext(f) == ".bacpac" {
		temporaryFolder = "/var/opt/mssql/backup"
	} else if filepath.Ext(f) == ".mdf" {
		temporaryFolder = "/var/opt/mssql/data"
	} else {
		panic(fmt.Sprintf("Unsupported file extension (%q)", filepath.Ext(f)))
	}

	if len(parsed.Scheme) > 3 {
		controller.DownloadFile(
			containerId,
			databaseUrl,
			temporaryFolder,
		)
	} else {
		controller.CopyFile(
			containerId,
			databaseUrl,
			temporaryFolder,
		)

		_, f := filepath.Split(databaseUrl)
		controller.RunCmdInContainer(containerId, []string{"chmod", "g+r", temporaryFolder + "/" + f})
	}

	if filepath.Ext(f) == ".7z" {
		controller.RunCmdInContainer(containerId, []string{
			"mkdir",
			"/opt/7-zip"})

		controller.RunCmdInContainer(containerId, []string{
			"wget",
			"-O",
			"/opt/7-zip/7-zip.tar",
			"https://7-zip.org/a/7z2201-linux-x64.tar.xz"})

		controller.RunCmdInContainer(containerId, []string{
			"tar",
			"xvf",
			"/opt/7-zip/7-zip.tar",
			"-C",
			"/opt/7-zip",
		})
		controller.RunCmdInContainer(containerId, []string{"chmod", "u+x", "/opt/7-zip/7zz"})
		controller.RunCmdInContainer(containerId, []string{
			"/opt/7-zip/7zz",
			"x",
			"-o/var/opt/mssql/data",
			temporaryFolder + "/" + file,
		})

		stdout, _ := controller.RunCmdInContainer(containerId, []string{
			"./opt/7-zip/7zz",
			"l",
			"-ba",
			"-slt",
			temporaryFolder + "/" + file,
		})

		databaseName = parseDbName(usingDatabaseUrl)

		temporaryFolder = "/var/opt/mssql/data"

		paths := extractPaths(string(stdout))
		for _, p := range paths {
			if filepath.Ext(p) == ".mdf" {
				f = p
				file = p
			}

			if filepath.Ext(p) == ".ldf" {
				log_f = p
				log_file = p
			}
		}
	}

	dbNameAsIdentifier := getDbNameAsIdentifier(databaseName)
	dbNameAsNonIdentifier := getDbNameAsNonIdentifier(databaseName)

	if filepath.Ext(f) == ".bak" {
		// Restore database from file
		output.Infof("Restoring database %s", databaseName)

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
SELECT @sql = @sql + char(13) + ' MOVE ''' + LogicalName + ''' TO ''/var/opt/mssql/data/' + LogicalName + '.' + RIGHT(PhysicalName,CHARINDEX('\',PhysicalName)) + ''','
FROM @fileListTable
WHERE IsPresent = 1
SET @sql = SUBSTRING(@sql, 1, LEN(@sql)-1)
EXEC(@sql)`

		query(fmt.Sprintf(text, temporaryFolder, file, dbNameAsIdentifier, temporaryFolder, file))
	} else if filepath.Ext(f) == ".mdf" {
		// Attach database
		output.Infof("Attaching database %s", databaseName)

		controller.RunCmdInContainer(containerId, []string{"chown", "mssql:root", temporaryFolder + "/" + file})
		controller.RunCmdInContainer(containerId, []string{"chmod", "-o-r", temporaryFolder + "/" + file})
		controller.RunCmdInContainer(containerId, []string{"chmod", "-u+rw", temporaryFolder + "/" + file})
		controller.RunCmdInContainer(containerId, []string{"chmod", "-g+r", temporaryFolder + "/" + file})

		text := `SET NOCOUNT ON;`
		if log_f == "" {
			text += `CREATE DATABASE [%s]   
    ON (FILENAME = '%s/%s')
    FOR ATTACH;`
			query(fmt.Sprintf(text, dbNameAsIdentifier, temporaryFolder, file))

		} else {
			controller.RunCmdInContainer(containerId, []string{"chown", "mssql:root", temporaryFolder + "/" + log_file})
			controller.RunCmdInContainer(containerId, []string{"chmod", "-o-r", temporaryFolder + "/" + log_file})
			controller.RunCmdInContainer(containerId, []string{"chmod", "-u+rw", temporaryFolder + "/" + log_file})
			controller.RunCmdInContainer(containerId, []string{"chmod", "-g+r", temporaryFolder + "/" + log_file})

			text += `CREATE DATABASE [%s]   
    ON (FILENAME = '%s/%s'), (FILENAME = '%s/%s') 
    FOR ATTACH;`

			query(fmt.Sprintf(text, dbNameAsIdentifier, temporaryFolder, file, temporaryFolder, log_file))
		}

	} else if filepath.Ext(f) == ".bacpac" {
		controller.DownloadFile(
			containerId,
			"https://aka.ms/sqlpackage-linux",
			"/tmp",
		)

		controller.RunCmdInContainer(containerId, []string{"apt-get", "update"})
		controller.RunCmdInContainer(containerId, []string{"apt-get", "install", "-y", "unzip"})
		controller.RunCmdInContainer(containerId, []string{"unzip", "/tmp/sqlpackage-linux", "-d", "/opt/sqlpackage"})
		controller.RunCmdInContainer(containerId, []string{"rm", "/tmp/sqlpackage-linux"})
		controller.RunCmdInContainer(containerId, []string{"chmod", "+x", "/opt/sqlpackage/sqlpackage"})

		alterDefaultDb := fmt.Sprintf(
			"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
			userName,
			"master")
		query(alterDefaultDb)

		controller.RunCmdInContainer(containerId, []string{
			"/opt/sqlpackage/sqlpackage",
			"/Diagnostics:true",
			"/Action:import",
			"/SourceFile:" + temporaryFolder + "/" + file,
			"/TargetServerName:localhost",
			"/TargetDatabaseName:" + dbNameAsIdentifier,
			"/TargetTrustServerCertificate:true",
			"/TargetUser:" + userName,
			"/TargetPassword:" + password,
		})
	}

	alterDefaultDb := fmt.Sprintf(
		"ALTER LOGIN [%s] WITH DEFAULT_DATABASE = [%s]",
		userName,
		dbNameAsNonIdentifier)
	query(alterDefaultDb)
}

func extractPaths(input string) []string {
	re := regexp.MustCompile(`Path\s*=\s*(\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	var paths []string
	for _, match := range matches {
		paths = append(paths, match[1])
	}
	return paths
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
		if lastIdx == -1 {
			lastIdx = strings.LastIndex(dbToken, ".mdf")
		}
		if lastIdx != -1 {
			//Get file name without extension
			fileName := dbToken[0:lastIdx]
			lastIdx += 5
			if lastIdx >= len(dbToken) {
				return fileName
			}
			//Return database name if it was specified
			return dbToken[lastIdx:]
		} else {
			lastIdx := strings.LastIndex(dbToken, ".bacpac")
			if lastIdx != -1 {
				//Get file name without extension
				fileName := dbToken[0:lastIdx]
				lastIdx += 8
				if lastIdx >= len(dbToken) {
					return fileName
				}
				//Return database name if it was specified
				return dbToken[lastIdx:]
			} else {
				lastIdx := strings.LastIndex(dbToken, ".7z")
				if lastIdx != -1 {
					//Get file name without extension
					fileName := dbToken[0:lastIdx]
					lastIdx += 4
					if lastIdx >= len(dbToken) {
						return fileName
					}
					//Return database name if it was specified
					return dbToken[lastIdx:]
				}
			}
		}

	}

	fileName := filepath.Base(usingDbUrl)
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

func extractUrl(usingArg string) string {
	urlEndIdx := strings.LastIndex(usingArg, ".bak")
	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(usingArg, ".mdf")
	}
	if urlEndIdx != -1 {
		return usingArg[0:(urlEndIdx + 4)]
	}

	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(usingArg, ".7z")
		if urlEndIdx != -1 {
			return usingArg[0:(urlEndIdx + 3)]
		}
	}

	if urlEndIdx == -1 {
		urlEndIdx = strings.LastIndex(usingArg, ".bacpac")
		if urlEndIdx != -1 {
			return usingArg[0:(urlEndIdx + 7)]
		}
	}

	return usingArg
}

// Verify the file exists at the URL
func urlExists(url string, output *output2.Output) {
	if !http.UrlExists(url) {
		output.FatalfWithHints(
			[]string{"File does not exist at URL"},
			"Unable to download file")
	}
}
