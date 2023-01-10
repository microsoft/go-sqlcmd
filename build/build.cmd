REM Generates sqlcmd.exe in the root dir of the repo
go build -o %~dp0..\sqlcmd.exe %~dp0..\cmd\modern