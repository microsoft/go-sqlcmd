-- go-sqlcmd Development Database Setup
-- This script runs automatically when the devcontainer starts

USE master;
GO

-- Create a test database for development
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = 'SqlCmdTest')
BEGIN
    CREATE DATABASE SqlCmdTest;
    PRINT 'Created database: SqlCmdTest';
END
GO

-- Enable contained database authentication for testing
-- Use TRY/CATCH in case this fails on certain SQL Server configurations
BEGIN TRY
    EXEC sp_configure 'contained database authentication', 1;
    RECONFIGURE;
END TRY
BEGIN CATCH
    PRINT 'Note: Could not enable contained database authentication (may already be enabled or not supported)';
END CATCH;
GO

-- Make SqlCmdTest a contained database for testing
BEGIN TRY
    ALTER DATABASE SqlCmdTest SET CONTAINMENT = PARTIAL;
END TRY
BEGIN CATCH
    PRINT 'Note: Could not set database containment (may not be supported)';
END CATCH;
GO

USE SqlCmdTest;
GO

-- Create a sample table for quick testing
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'TestTable')
BEGIN
    CREATE TABLE TestTable (
        ID INT IDENTITY(1,1) PRIMARY KEY,
        Name NVARCHAR(100) NOT NULL,
        Value NVARCHAR(MAX),
        CreatedAt DATETIME2 DEFAULT GETUTCDATE()
    );
    
    INSERT INTO TestTable (Name, Value) VALUES 
        ('Test1', 'Sample value 1'),
        ('Test2', 'Sample value 2'),
        ('Test3', 'Sample value 3');
    
    PRINT 'Created table: TestTable with sample data';
END
GO

-- Create a view for testing
IF NOT EXISTS (SELECT * FROM sys.views WHERE name = 'TestView')
BEGIN
    EXEC('CREATE VIEW TestView AS SELECT ID, Name, CreatedAt FROM TestTable');
    PRINT 'Created view: TestView';
END
GO

PRINT 'go-sqlcmd development database setup complete!';
PRINT 'Test database: SqlCmdTest';
PRINT 'Sample table: TestTable (3 rows)';
PRINT 'Sample view: TestView';
GO
