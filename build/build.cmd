REM Get Version Tag
for /f %%i in ('"git describe --tags --abbrev=0"') do set sqlcmdVersion=%%i

REM Generates sqlcmd.exe in the root dir of the repo
go build -o %~dp0..\sqlcmd.exe -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern