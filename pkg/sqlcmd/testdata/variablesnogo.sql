set nocount on
:setvar hundred 100

-- verify fix for https://github.com/microsoft/go-sqlcmd/issues/197

-- Correctly handle the first line of a batch having a variable after an empty line

GO

select $(hundred)

GO