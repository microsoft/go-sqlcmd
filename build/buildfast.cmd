@echo off

REM We get the value of the escape character by using PROMPT $E
for /F "tokens=1,2 delims=#" %%a in ('"prompt #$H#$E# & echo on & for %%b in (1) do rem"') do (
  set "DEL=%%a"
  set "ESC=%%b"
)

REM Get Version Tag
for /f %%i in ('"git describe --tags --abbrev=0"') do set sqlcmdVersion=%%i

REM Generates sqlcmd.exe in the root dir of the repo
go build -o %~dp0..\sqlcmd.exe -ldflags="-X main.version=%sqlcmdVersion%" %~dp0..\cmd\modern

:end
