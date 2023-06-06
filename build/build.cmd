@echo off

REM We get the value of the escape character by using PROMPT $E
for /F "tokens=1,2 delims=#" %%a in ('"prompt #$H#$E# & echo on & for %%b in (1) do rem"') do (
  set "DEL=%%a"
  set "ESC=%%b"
)
setlocal
SET     RED=%ESC%[1;31m
echo %RED%
REM run the custom sqlcmd linter for code style enforcement
REM using for/do instead of running it directly so the status code isn't checked by the shell.
REM Once we are prepared to block the build with the linter we will move this step into a pipeline
for /F  "usebackq"  %%l in (`go run cmd\sqlcmd-linter\main.go -test %~dp0../...`) DO echo %%l
echo %ESC%[0m
endlocal
REM Get Version Tag
for /f %%i in ('"git describe --tags --abbrev=0"') do set sqlcmdVersion=%%i

if not exist %gopath%\bin\go-winres.exe (
    go install github.com/tc-hib/go-winres@latest
)
if not exist %gopath%\bin\gotext.exe (
    go install golang.org/x/text/cmd/gotext@latest
)

REM go-winres likes to append instead of overwrite so delete existing resource file
del %~dp0..\cmd\modern\*.syso

REM generates translations file and resources
go generate %~dp0../... 2> %~dp0generate.txt
echo Fix any conflicting localizable strings:
echo %RED%
findstr conflicting "%~dp0generate.txt"
echo %ESC%[0m
if not %errorlevel% == 0 goto :end
REM Generates sqlcmd.exe in the root dir of the repo
go build -o %~dp0..\sqlcmd.exe -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern

REM Generate NOTICE
if not exist %gopath%\bin\go-licenses.exe (
    go install github.com/google/go-licenses@latest
)
go-licenses report github.com/microsoft/go-sqlcmd/cmd/modern --template build\NOTICE.tpl --ignore github.com/microsoft > %~dp0notice.txt 2>nul
copy %~dp0NOTICE.header + %~dp0notice.txt %~dp0..\NOTICE.md
del %~dp0notice.txt

REM Generates all versions of sqlcmd in platform-specific folder
setlocal

for /F "tokens=1-3 delims=," %%i in (%~dp0arch.txt) do set GOOS=%%i&set GOARCH=%%j&go build -o %~dp0..\%%i-%%j\%%k -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern
endlocal

:end
del %~dp0generate.txt
