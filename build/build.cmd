@echo off
REM Get Version Tag
for /f %%i in ('"git describe --tags --abbrev=0"') do set sqlcmdVersion=%%i

REM Generates sqlcmd.exe in the root dir of the repo
go build -o %~dp0..\sqlcmd.exe -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern

REM Generate NOTICE
if not exist %gopath%\bin\go-licenses.exe (
    go install github.com/google/go-licenses@latest
)
go-licenses report github.com/microsoft/go-sqlcmd/cmd/modern --template build\NOTICE.tpl --ignore github.com/microsoft > %~dp0notice.txt
copy %~dp0NOTICE.header + %~dp0notice.txt %~dp0..\NOTICE.md
del %~dp0notice.txt

REM Generates all versions of sqlcmd in platform-specific folder
setlocal

for /F "tokens=1-3 delims=," %%i in (%~dp0arch.txt) do set GOOS=%%i&set GOARCH=%%j&go build -o %~dp0..\%%i-%%j\%%k -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern


