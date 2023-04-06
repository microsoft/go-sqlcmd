package mechanism

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/container"
)

type restore struct {
}

func (m *restore) Initialize(controller *container.Controller) {
}

func (m *restore) CopyToLocation() string {
	return "/var/opt/mssql/backup"
}

func (m *restore) Name() string {
	return "restore"
}

func (m *restore) FileTypes() []string {
	return []string{"bak"}
}

func (m *restore) BringOnline(databaseName string, _ string, query func(string), options BringOnlineOptions) {
	if options.Filename == "" {
		panic("Filename is required for restore")
	}
	if databaseName == "" {
		panic("databaseName is required for restore")
	}

	query(fmt.Sprintf(
		m.restoreStatement(),
		m.CopyToLocation(),
		options.Filename,
		databaseName,
		m.CopyToLocation(),
		options.Filename,
	))
}

func (m *restore) restoreStatement() string {
	return `SET NOCOUNT ON;

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
}
