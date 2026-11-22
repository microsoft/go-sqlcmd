package install

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/go-sqlcmd/cmd/modern/sqlconfig"
	"github.com/microsoft/go-sqlcmd/internal/secret"
	"github.com/microsoft/go-sqlcmd/internal/sql"

	"github.com/stretchr/testify/assert"
)

func TestGetDbNameIfExists(t *testing.T) {

	type test struct {
		input                   string
		expectedNonIdentifierOp string
	}

	tests := []test{
		// Positive Testcases
		// Database name specified
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDbName", "myDbName"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDb Name", "myDb Name"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,myDb Na,me", "myDb Na,me"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb Na,me]", "[myDb Na,me]]"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb Na'me]", "[myDb Na'me]]"},
		{"https://example.com/my%20random%20bac%27kp%5Dack.bak,[myDb ,Nam,e]", "[myDb ,Nam,e]]"},

		// Delimiter between filename and databaseName is part of the filename
		// Decoded filename: my random .bak bac'kp]ack.bak
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak,[myDb ,Nam,e]", "[myDb ,Nam,e]]"},

		// Database name not specified
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak", "my random .bak bac'kp]]ack"},
		{"https://example.com/my%20random%20.bak%20bac%27kp%5Dack.bak,", "my random .bak bac'kp]]ack"},

		//Negative Testcases
		{"https://example.com,myDbName", ""},
	}

	for _, testcase := range tests {
		dbname := parseDbName(testcase.input)
		dbnameAsNonIdentifier := getDbNameAsNonIdentifier(dbname)
		assert.Equal(t, testcase.expectedNonIdentifierOp, dbnameAsNonIdentifier, "Unexpected database name as non-identifier")
	}
}

func TestExtractUrl(t *testing.T) {
	type test struct {
		inputURL    string
		expectedURL string
	}

	tests := []test{
		{"https://example.com/testdb.bak,myDbName", "https://example.com/testdb.bak"},
		{"https://example.com/testdb.bak", "https://example.com/testdb.bak"},
		{"https://example.com,", "https://example.com,"},
	}

	for _, testcase := range tests {
		assert.Equal(t, testcase.expectedURL, extractUrl(testcase.inputURL), "Extracted URL does not match expected URL")
	}
}

func TestRestoreFileStem(t *testing.T) {
	tests := []struct {
		name         string
		databaseName string
		expected     string
	}{
		{"letters, digits, underscore and hyphen are kept", "AdventureWorks_2022-v1", "AdventureWorks_2022-v1"},
		{"apostrophe and space are replaced", "Sales O'Brien", "Sales_O_Brien"},
		{"path separators are replaced", "../../etc/x", "______etc_x"},
		{"non-ASCII letters are replaced", "Données", "Donn_es"},
		{"empty name gets a default", "", "database"},
		{"long name is truncated", strings.Repeat("a", 100), strings.Repeat("a", 64)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, restoreFileStem(tt.databaseName))
		})
	}
}

func TestRestoreDatabaseQueryQuotesValues(t *testing.T) {
	query := restoreDatabaseQuery("/var/opt/mssql/backup/O'Brien.bak", "Sales O'Brien")

	assert.Contains(t, query, "DECLARE @backupFile NVARCHAR(4000) = N'/var/opt/mssql/backup/O''Brien.bak';")
	assert.Contains(t, query, "DECLARE @databaseName NVARCHAR(MAX) = N'Sales O''Brien';")
	assert.Contains(t, query, "DECLARE @filePrefix NVARCHAR(260) = N'/var/opt/mssql/Sales_O_Brien';")
}

func TestRestoreDatabaseQueryRestoresLogicalNamesWithApostrophes(t *testing.T) {
	s := connectToTestServer(t)
	failOnQueryError(t)

	if platform := s.ScalarString("DECLARE @platform NVARCHAR(256); SELECT @platform = host_platform FROM sys.dm_os_host_info; PRINT @platform"); platform != "Linux" {
		t.Skipf("restore paths are for SQL Server on Linux, server runs on %q", platform)
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	sourceDb := "restore_source_" + suffix
	restoredDb := "restored O'Brien " + suffix
	backupFile := "/var/opt/mssql/data/" + sourceDb + " O'Brien.bak"

	s.Query(fmt.Sprintf(`CREATE DATABASE [%[1]s]
ON (NAME = N'Sales O''Brien', FILENAME = N'/var/opt/mssql/data/%[1]s.mdf')
LOG ON (NAME = N'Sales O''Brien log', FILENAME = N'/var/opt/mssql/data/%[1]s.ldf')`, sourceDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", sourceDb))
	s.Query(fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = %s WITH INIT", sourceDb, sqlStringLiteral(backupFile)))

	s.Query(restoreDatabaseQuery(backupFile, restoredDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", restoredDb))

	files := s.ScalarString(fmt.Sprintf(`DECLARE @files NVARCHAR(MAX);
SELECT @files = STRING_AGG(CAST(name + N'|' + physical_name AS NVARCHAR(MAX)), N';') WITHIN GROUP (ORDER BY file_id)
FROM sys.master_files
WHERE database_id = DB_ID(%s);
PRINT @files`, sqlStringLiteral(restoredDb)))

	assert.Equal(t,
		fmt.Sprintf("Sales O'Brien|/var/opt/mssql/%[1]s_1.mdf;Sales O'Brien log|/var/opt/mssql/%[1]s_2.ldf", restoreFileStem(restoredDb)),
		files)
}

func TestRestoreDatabaseQueryDoesNotCollideWithUserDatabase(t *testing.T) {
	s := connectToTestServer(t)
	failOnQueryError(t)

	if platform := s.ScalarString("DECLARE @platform NVARCHAR(256); SELECT @platform = host_platform FROM sys.dm_os_host_info; PRINT @platform"); platform != "Linux" {
		t.Skipf("restore paths are for SQL Server on Linux, server runs on %q", platform)
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	sourceDb := "restore_source_" + suffix
	restoredDb := "Sales_" + suffix
	userDb := restoredDb + "_1"
	backupFile := "/var/opt/mssql/data/" + sourceDb + ".bak"

	s.Query(fmt.Sprintf("CREATE DATABASE [%s]", sourceDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", sourceDb))
	s.Query(fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = %s WITH INIT", sourceDb, sqlStringLiteral(backupFile)))

	// --user-database is created with default file paths before the restore runs.
	s.Query(fmt.Sprintf("CREATE DATABASE [%s]", userDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", userDb))
	assert.Equal(t, "/var/opt/mssql/data/"+userDb+".mdf",
		s.ScalarString(fmt.Sprintf("DECLARE @f NVARCHAR(260); SELECT @f = physical_name FROM sys.master_files WHERE database_id = DB_ID(N'%s') AND file_id = 1; PRINT @f", userDb)))

	s.Query(restoreDatabaseQuery(backupFile, restoredDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", restoredDb))

	assert.Equal(t, "ONLINE",
		s.ScalarString(fmt.Sprintf("DECLARE @s NVARCHAR(60); SELECT @s = state_desc FROM sys.databases WHERE name = N'%s'; PRINT @s", restoredDb)))
}

func TestRestoreDatabaseQueryRejectsOverlongName(t *testing.T) {
	s := connectToTestServer(t)
	failOnQueryError(t)

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	sourceDb := "restore_source_" + suffix
	backupFile := "/var/opt/mssql/data/" + sourceDb + ".bak"
	overlongName := suffix + strings.Repeat("x", 129-len(suffix))
	truncatedName := overlongName[:128]

	s.Query(fmt.Sprintf("CREATE DATABASE [%s]", sourceDb))
	defer s.Query(fmt.Sprintf("DROP DATABASE [%s]", sourceDb))
	s.Query(fmt.Sprintf("BACKUP DATABASE [%s] TO DISK = %s WITH INIT", sourceDb, sqlStringLiteral(backupFile)))

	// Restore errors are printed, not returned, so check the outcome instead.
	s.Query(restoreDatabaseQuery(backupFile, overlongName))
	defer s.Query(fmt.Sprintf("IF DB_ID(N'%[1]s') IS NOT NULL DROP DATABASE [%[1]s]", truncatedName))

	assert.Equal(t, "0",
		s.ScalarString(fmt.Sprintf("DECLARE @n INT; SELECT @n = COUNT(*) FROM sys.databases WHERE name = N'%s'; PRINT @n", truncatedName)))
}

// connectToTestServer connects to the server named by the SQLCMD* environment
// variables, or skips the test when no password is set.
func connectToTestServer(t *testing.T) sql.Sql {
	password := os.Getenv("SQLCMDPASSWORD")
	if password == "" {
		t.Skip("SQLCMDPASSWORD is not set")
	}
	server := os.Getenv("SQLCMDSERVER")
	if server == "" {
		server = "localhost"
	}
	port := 1433
	if p := os.Getenv("SQLCMDPORT"); p != "" {
		var err error
		port, err = strconv.Atoi(p)
		if err != nil {
			t.Fatalf("SQLCMDPORT %q is not a number", p)
		}
	}
	userName := os.Getenv("SQLCMDUSER")
	if userName == "" {
		userName = "sa"
	}

	failOnQueryError(t)
	s := sql.New(sql.SqlOptions{})
	s.Connect(
		sqlconfig.Endpoint{EndpointDetails: sqlconfig.EndpointDetails{Address: server, Port: port}},
		&sqlconfig.User{
			AuthenticationType: "basic",
			BasicAuth: &sqlconfig.BasicAuthDetails{
				Username:           userName,
				PasswordEncryption: "none",
				Password:           secret.Encode(password, "none"),
			},
		},
		sql.ConnectOptions{Database: "master"})
	return s
}

func failOnQueryError(t *testing.T) {
	sql.Initialize(func(err error) {
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
	}, func(string, ...any) {}, secret.Decode)
}
