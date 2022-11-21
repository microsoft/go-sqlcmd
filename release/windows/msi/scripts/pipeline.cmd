@echo off
SetLocal EnableDelayedExpansion
echo build a msi installer. You need to have curl.exe, unzip.exe and msbuild.exe available under PATH
echo.

set "PATH=%PATH%;%ProgramFiles%\Git\bin;%ProgramFiles%\Git\usr\bin;C:\Program Files (x86)\Git\bin;C:\Program Files (x86)\MSBuild\14.0\Bin;C:\Program Files (x86)\Microsoft Visual Studio\2017\Enterprise\MSBuild\15.0\Bin;C:\Program Files (x86)\Windows Kits\10;"
echo %PATH%

if "%CLI_VERSION%"=="" (
    set CLI_VERSION=0.0.1
)

if "%WIX_DOWNLOAD_URL%"=="" (
    echo Please set the WIX_DOWNLOAD_URL environment variable, e.g. https://host/wix314-binaries-mirror.zip
    goto ERROR
)

:: Set up the output directory and temp. directories
echo Cleaning previous build artifacts...
set OUTPUT_DIR=%~dp0..\..\..\..\output\msi
if exist %OUTPUT_DIR% rmdir /s /q %OUTPUT_DIR%
mkdir %OUTPUT_DIR%

set ARTIFACTS_DIR=%~dp0..\..\..\..\output\msi\artifacts
mkdir %ARTIFACTS_DIR%

set WIX_DIR=%ARTIFACTS_DIR%\wix
set REPO_ROOT=%~dp0..\..\..\..

set PIPELINE_WORKSPACE=%ARTIFACTS_DIR%\workspace

mkdir %PIPELINE_WORKSPACE%\SqlcmdWindowsAmd64

copy /y %REPO_ROOT%\sqlcmd.exe %PIPELINE_WORKSPACE%\SqlcmdWindowsAmd64\sqlcmd.exe

::ensure wix is available
if exist %WIX_DIR% (
    echo Using existing Wix at %WIX_DIR%
)
if not exist %WIX_DIR% (
    mkdir %WIX_DIR%
    pushd %WIX_DIR%
    echo Downloading Wix.
    curl -o wix-archive.zip %WIX_DOWNLOAD_URL% -k
    unzip -q wix-archive.zip
    if %errorlevel% neq 0 goto ERROR
    del wix-archive.zip
    echo Wix downloaded and extracted successfully.
    popd
)

if %errorlevel% neq 0 goto ERROR

set PATH=%PATH%;%WIX_DIR%

@echo off

:: During pipeline we want to skip msbuild here and use the AzureDevOps Task instead
if "%1"=="--skip-msbuild" (
    echo Skipping inline MSI Build...
) else (
    echo Building MSI...
    cd %OUTPUT_DIR%
    msbuild /t:rebuild /p:Configuration=Release %REPO_ROOT%\release\windows\msi\sqlcmd.wixproj -p:OutDir=%OUTPUT_DIR%\
    start %OUTPUT_DIR%
)

goto END

:ERROR
echo Error occurred, please check the output for details.
exit /b 1

:END
exit /b 0
popd
