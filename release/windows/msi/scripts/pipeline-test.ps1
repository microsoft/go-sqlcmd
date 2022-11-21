#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
#------------------------------------------------------------------------------

# Description:
#
# Instructions to be invoked under the build CI pipeline in AzureDevOps.
#
# Kickoff MSI package install test.
#
# Usage:
#
# set SYSTEM_ARTIFACTSDIRECTORY=\path\to\msi\sqlcmd-<CL_VERSION>.msi
#
# $ pipeline-test.ps1

if (-not (Test-Path env:CLI_VERSION)) { $env:CLI_VERSION = '0.0.1' }
if (-not (Test-Path env:CLI_VERSION_REVISION)) { $env:CLI_VERSION_REVISION = '1' }

tree /A /F $env:SYSTEM_ARTIFACTSDIRECTORY

$msiPath = Join-Path $env:SYSTEM_ARTIFACTSDIRECTORY ("sqlcmd_" + $env:CLI_VERSION + "-" + $env:CLI_VERSION_REVISION + ".msi")

$msiPath

$InstallArgs = @(
    "/I"
    $msiPath
    "/norestart"
    "/L*v"
    ".\install-logs.txt"
    "/qn"
)
        
Write-Output "Starting msi install $msiPath..."
Start-Process "msiexec.exe" -ArgumentList $InstallArgs -Wait -NoNewWindow
Get-Content .\install-logs.txt

Write-Output "Done installing msi, checking PATH setup..."
Write-Output "$env:path"
$env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine")
Write-Output "$env:path"

$sqlcmd = $env:ProgramFiles + "\SqlCmd\Tools\sqlcmd"
& $sqlcmd --help
